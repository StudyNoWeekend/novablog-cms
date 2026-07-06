// Package response 提供统一的 HTTP 响应封装。
package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response 统一响应结构体。
type Response struct {
	Data    interface{} `json:"data,omitempty"`     // 响应数据
	Code    int         `json:"code"`               // 业务状态码，0 表示成功
	Msg     string      `json:"msg"`                // 响应信息
	TraceID string      `json:"trace_id,omitempty"` // 链路追踪 ID
}

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
