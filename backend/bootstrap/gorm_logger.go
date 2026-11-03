package bootstrap

import (
	"context"
	"errors"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// gormZapLogger 将 GORM SQL 日志接入统一 zap 日志体系。
// 级别策略：应用日志为 debug 时记录全部 SQL（gorm Info）；
// 其余级别仅记录错误与慢查询（gorm Warn），启动期 AutoMigrate 的元数据查询不再刷屏。
type gormZapLogger struct {
	logger        *zap.Logger
	level         gormlogger.LogLevel
	slowThreshold time.Duration
}

// newGormZapLogger 按应用日志级别构造 GORM 日志适配器；logger 为空时丢弃全部 SQL 日志。
func newGormZapLogger(logger *zap.Logger) gormlogger.Interface {
	if logger == nil {
		return gormlogger.Discard
	}
	level := gormlogger.Warn
	if logger.Level() == zapcore.DebugLevel {
		level = gormlogger.Info
	}
	return &gormZapLogger{
		logger:        logger,
		level:         level,
		slowThreshold: 200 * time.Millisecond,
	}
}

// LogMode 实现gorm logger.Interface，返回调整级别后的副本。
func (l *gormZapLogger) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	clone := *l
	clone.level = level
	return &clone
}

// Info 实现gorm logger.Interface。
func (l *gormZapLogger) Info(_ context.Context, msg string, args ...interface{}) {
	if l.level < gormlogger.Info {
		return
	}
	l.logger.Sugar().Infof(msg, args...)
}

// Warn 实现gorm logger.Interface。
func (l *gormZapLogger) Warn(_ context.Context, msg string, args ...interface{}) {
	if l.level < gormlogger.Warn {
		return
	}
	l.logger.Sugar().Warnf(msg, args...)
}

// Error 实现gorm logger.Interface。
func (l *gormZapLogger) Error(_ context.Context, msg string, args ...interface{}) {
	if l.level < gormlogger.Error {
		return
	}
	l.logger.Sugar().Errorf(msg, args...)
}

// Trace 实现gorm logger.Interface：错误、慢查询与（debug 级别下的）常规 SQL。
func (l *gormZapLogger) Trace(_ context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.level <= gormlogger.Silent {
		return
	}
	elapsed := time.Since(begin)
	switch {
	case err != nil && l.level >= gormlogger.Error && !errors.Is(err, gorm.ErrRecordNotFound):
		sql, rows := fc()
		l.logger.Error("GORM SQL 执行失败",
			zap.Error(err), zap.Duration("elapsed", elapsed), zap.Int64("rows", rows), zap.String("sql", sql))
	case l.slowThreshold != 0 && elapsed > l.slowThreshold && l.level >= gormlogger.Warn:
		sql, rows := fc()
		l.logger.Warn("GORM 慢查询",
			zap.Duration("elapsed", elapsed), zap.Int64("rows", rows), zap.String("sql", sql))
	case l.level >= gormlogger.Info:
		sql, rows := fc()
		l.logger.Debug("GORM SQL",
			zap.Duration("elapsed", elapsed), zap.Int64("rows", rows), zap.String("sql", sql))
	}
}
