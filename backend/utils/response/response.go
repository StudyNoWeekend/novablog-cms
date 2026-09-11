// Package response 提供统一的 HTTP 响应封装。
package response

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"novablog/enum"
)

// Response 统一响应结构体。
type Response struct {
	Data    interface{} `json:"data,omitempty"`     // 响应数据
	Code    int         `json:"code"`               // 业务状态码，0 表示成功
	Msg     string      `json:"msg"`                // 响应信息
	TraceID string      `json:"trace_id,omitempty"` // 链路追踪 ID
}

// ErrorLogger 错误响应日志记录器（由启动时注入）。
//
// 调用方可在 cmd/api/main.go 中将全局 logger 赋给此变量。
// 留空时不记录日志，保持向后兼容。
var ErrorLogger *zap.Logger

// Success 返回成功响应（Code=0, Msg="success"）。
func Success(c *gin.Context, data interface{}) {
	traceID, _ := c.Get("trace_id")
	tid := ""
	if traceID != nil {
		tid = traceID.(string)
	}
	c.JSON(http.StatusOK, Response{
		Data:    data,
		Code:    0,
		Msg:     "success",
		TraceID: tid,
	})
}

// Error 返回简单错误响应（HTTP 400）。
func Error(c *gin.Context, msg string) {
	traceID, _ := c.Get("trace_id")
	tid := ""
	if traceID != nil {
		tid = traceID.(string)
	}
	c.JSON(http.StatusBadRequest, Response{
		Code:    400,
		Msg:     msg,
		TraceID: tid,
	})
}

// Fail 返回失败响应，使用指定的错误码、错误信息和 HTTP 状态码。
func Fail(c *gin.Context, code int, msg string, httpCode int) {
	traceID, _ := c.Get("trace_id")
	tid := ""
	if traceID != nil {
		tid = traceID.(string)
	}
	c.JSON(httpCode, Response{
		Code:    code,
		Msg:     msg,
		TraceID: tid,
	})
}

// HandleError 统一处理控制器错误响应。
// 优先检查 BizError，其次检查 GORM RecordNotFound，最后默认 500。
// 同时将错误记录到 zap 日志（若 ErrorLogger 已注入）。
func HandleError(c *gin.Context, err error) {
	traceID, _ := c.Get("trace_id")
	tid := ""
	if traceID != nil {
		tid = traceID.(string)
	}

	logFields := []zap.Field{
		zap.String("path", c.Request.URL.Path),
		zap.String("method", c.Request.Method),
		zap.String("trace_id", tid),
		zap.Error(err),
	}

	var bizErr *enum.BizError
	if errors.As(err, &bizErr) {
		if ErrorLogger != nil {
			// 业务错误通常是预期内的，按 warn 记录。
			ErrorLogger.Warn("业务错误", append(logFields,
				zap.Int("code", bizErr.Code),
				zap.Int("http_code", bizErr.HttpCode),
			)...)
		}
		c.JSON(bizErr.HttpCode, Response{
			Code:    bizErr.Code,
			Msg:     bizErr.Msg,
			TraceID: tid,
		})
		return
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		if ErrorLogger != nil {
			ErrorLogger.Warn("资源不存在", logFields...)
		}
		c.JSON(http.StatusNotFound, Response{
			Code:    enum.ErrNotFound.Code,
			Msg:     enum.ErrNotFound.Msg,
			TraceID: tid,
		})
		return
	}

	if ErrorLogger != nil {
		ErrorLogger.Error("系统内部错误", logFields...)
	}
	c.JSON(http.StatusInternalServerError, Response{
		Code:    enum.ErrInternalServer.Code,
		Msg:     err.Error(),
		TraceID: tid,
	})
}
