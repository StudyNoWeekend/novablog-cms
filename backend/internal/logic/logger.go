package logic

import "go.uber.org/zap"

// LogicLogger logic 层通用日志记录器（bootstrap 注入，未注入时静默）。
// 供无 logger 依赖的 logic 结构体在 best-effort 操作失败时记录日志使用。
var LogicLogger *zap.Logger

// logicLog 获取 logic 层通用日志记录器。
func logicLog() *zap.Logger {
	if LogicLogger != nil {
		return LogicLogger
	}
	return zap.NewNop()
}
