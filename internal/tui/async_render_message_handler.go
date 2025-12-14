package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

// AsyncRenderMessageHandlerImpl 异步渲染消息处理器实现
type AsyncRenderMessageHandlerImpl struct {
	priority int
}

// NewAsyncRenderMessageHandler 创建新的异步渲染消息处理器
func NewAsyncRenderMessageHandler() *AsyncRenderMessageHandlerImpl {
	return &AsyncRenderMessageHandlerImpl{
		priority: 5, // 最低优先级
	}
}

// CanHandle 检查是否可以处理该消息
func (h *AsyncRenderMessageHandlerImpl) CanHandle(msg tea.Msg) bool {
	// 这个处理器已经包含在StreamMessageHandler中
	// 这里只是为了演示扩展性
	return false
}

// Handle 处理消息
func (h *AsyncRenderMessageHandlerImpl) Handle(msg tea.Msg, model *Model) (tea.Model, tea.Cmd) {
	return model, nil
}

// Priority 处理优先级
func (h *AsyncRenderMessageHandlerImpl) Priority() int {
	return h.priority
}