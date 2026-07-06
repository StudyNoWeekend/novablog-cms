// Package res 定义响应 DTO（数据传输对象）。
package res

// StatusRes 初始化状态响应。
type StatusRes struct {
	Initialized bool `json:"initialized"` // 是否已初始化
}

// InitRes 初始化结果响应。
type InitRes struct {
	Success bool   `json:"success"` // 是否成功
	Message string `json:"message"` // 响应消息
}