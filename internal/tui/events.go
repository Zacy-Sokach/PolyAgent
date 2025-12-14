package tui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// UIUpdateEvent UI更新事件
type UIUpdateEvent struct {
	*BaseEvent
}

// NewUIUpdateEvent 创建UI更新事件
func NewUIUpdateEvent() *UIUpdateEvent {
	return &UIUpdateEvent{
		BaseEvent: NewBaseEvent(EventTypeUIUpdate, nil),
	}
}

// UIResizeEvent UI调整大小事件
type UIResizeEvent struct {
	*BaseEvent
	Width  int
	Height int
}

// NewUIResizeEvent 创建UI调整大小事件
func NewUIResizeEvent(width, height int) *UIResizeEvent {
	event := &UIResizeEvent{
		BaseEvent: NewBaseEvent(EventTypeUIResize, nil),
		Width:     width,
		Height:    height,
	}
	event.data = map[string]interface{}{
		"width":  width,
		"height": height,
	}
	return event
}

// MessageAddedEvent 消息添加事件
type MessageAddedEvent struct {
	*BaseEvent
	Role    string
	Content string
}

// NewMessageAddedEvent 创建消息添加事件
func NewMessageAddedEvent(role, content string) *MessageAddedEvent {
	event := &MessageAddedEvent{
		BaseEvent: NewBaseEvent(EventTypeMessageAdded, nil),
		Role:      role,
		Content:   content,
	}
	event.data = map[string]interface{}{
		"role":    role,
		"content": content,
	}
	return event
}

// MessageUpdatedEvent 消息更新事件
type MessageUpdatedEvent struct {
	*BaseEvent
	MessageID string
	Role      string
	Content   string
}

// NewMessageUpdatedEvent 创建消息更新事件
func NewMessageUpdatedEvent(messageID, role, content string) *MessageUpdatedEvent {
	event := &MessageUpdatedEvent{
		BaseEvent: NewBaseEvent(EventTypeMessageUpdated, nil),
		MessageID: messageID,
		Role:      role,
		Content:   content,
	}
	event.data = map[string]interface{}{
		"message_id": messageID,
		"role":       role,
		"content":    content,
	}
	return event
}

// StreamStartedEvent 流式开始事件
type StreamStartedEvent struct {
	*BaseEvent
	Query string
}

// NewStreamStartedEvent 创建流式开始事件
func NewStreamStartedEvent(query string) *StreamStartedEvent {
	event := &StreamStartedEvent{
		BaseEvent: NewBaseEvent(EventTypeStreamStarted, nil),
		Query:     query,
	}
	event.data = map[string]interface{}{
		"query": query,
	}
	return event
}

// StreamChunkEvent 流式数据块事件
type StreamChunkEvent struct {
	*BaseEvent
	Content   string
	Reasoning string
}

// NewStreamChunkEvent 创建流式数据块事件
func NewStreamChunkEvent(content, reasoning string) *StreamChunkEvent {
	event := &StreamChunkEvent{
		BaseEvent: NewBaseEvent(EventTypeStreamChunk, nil),
		Content:   content,
		Reasoning: reasoning,
	}
	event.data = map[string]interface{}{
		"content":   content,
		"reasoning": reasoning,
	}
	return event
}

// StreamFinishedEvent 流式完成事件
type StreamFinishedEvent struct {
	*BaseEvent
	TotalChunks int
	Duration    time.Duration
}

// NewStreamFinishedEvent 创建流式完成事件
func NewStreamFinishedEvent(totalChunks int, duration time.Duration) *StreamFinishedEvent {
	event := &StreamFinishedEvent{
		BaseEvent:  NewBaseEvent(EventTypeStreamFinished, nil),
		TotalChunks: totalChunks,
		Duration:    duration,
	}
	event.data = map[string]interface{}{
		"total_chunks": totalChunks,
		"duration":     duration,
	}
	return event
}

// StreamErrorEvent 流式错误事件
type StreamErrorEvent struct {
	*BaseEvent
	Error  error
	Retry  bool
	Attempt int
}

// NewStreamErrorEvent 创建流式错误事件
func NewStreamErrorEvent(err error, retry bool, attempt int) *StreamErrorEvent {
	event := &StreamErrorEvent{
		BaseEvent: NewBaseEvent(EventTypeStreamError, nil),
		Error:     err,
		Retry:     retry,
		Attempt:   attempt,
	}
	event.data = map[string]interface{}{
		"error":   err,
		"retry":   retry,
		"attempt": attempt,
	}
	return event
}

// ToolCalledEvent 工具调用事件
type ToolCalledEvent struct {
	*BaseEvent
	ToolName string
	Args     map[string]interface{}
}

// NewToolCalledEvent 创建工具调用事件
func NewToolCalledEvent(toolName string, args map[string]interface{}) *ToolCalledEvent {
	event := &ToolCalledEvent{
		BaseEvent: NewBaseEvent(EventTypeToolCalled, nil),
		ToolName:  toolName,
		Args:      args,
	}
	event.data = map[string]interface{}{
		"tool_name": toolName,
		"args":      args,
	}
	return event
}

// ToolCompletedEvent 工具完成事件
type ToolCompletedEvent struct {
	*BaseEvent
	ToolName string
	Result   interface{}
	Duration time.Duration
}

// NewToolCompletedEvent 创建工具完成事件
func NewToolCompletedEvent(toolName string, result interface{}, duration time.Duration) *ToolCompletedEvent {
	event := &ToolCompletedEvent{
		BaseEvent: NewBaseEvent(EventTypeToolCompleted, nil),
		ToolName:  toolName,
		Result:    result,
		Duration:  duration,
	}
	event.data = map[string]interface{}{
		"tool_name": toolName,
		"result":    result,
		"duration":  duration,
	}
	return event
}

// ToolFailedEvent 工具失败事件
type ToolFailedEvent struct {
	*BaseEvent
	ToolName string
	Error    error
	Duration time.Duration
}

// NewToolFailedEvent 创建工具失败事件
func NewToolFailedEvent(toolName string, err error, duration time.Duration) *ToolFailedEvent {
	event := &ToolFailedEvent{
		BaseEvent: NewBaseEvent(EventTypeToolFailed, nil),
		ToolName:  toolName,
		Error:     err,
		Duration:  duration,
	}
	event.data = map[string]interface{}{
		"tool_name": toolName,
		"error":     err,
		"duration":  duration,
	}
	return event
}

// PerformanceWarningEvent 性能警告事件
type PerformanceWarningEvent struct {
	*BaseEvent
	Metric    string
	Value     float64
	Threshold float64
	Message   string
}

// NewPerformanceWarningEvent 创建性能警告事件
func NewPerformanceWarningEvent(metric string, value, threshold float64, message string) *PerformanceWarningEvent {
	event := &PerformanceWarningEvent{
		BaseEvent: NewBaseEvent(EventTypePerformanceWarning, nil),
		Metric:    metric,
		Value:     value,
		Threshold: threshold,
		Message:   message,
	}
	event.data = map[string]interface{}{
		"metric":    metric,
		"value":     value,
		"threshold": threshold,
		"message":   message,
	}
	return event
}

// RenderStartedEvent 渲染开始事件
type RenderStartedEvent struct {
	*BaseEvent
	ContentType string
	Size        int
}

// NewRenderStartedEvent 创建渲染开始事件
func NewRenderStartedEvent(contentType string, size int) *RenderStartedEvent {
	event := &RenderStartedEvent{
		BaseEvent:  NewBaseEvent(EventTypeRenderStarted, nil),
		ContentType: contentType,
		Size:       size,
	}
	event.data = map[string]interface{}{
		"content_type": contentType,
		"size":        size,
	}
	return event
}

// RenderCompletedEvent 渲染完成事件
type RenderCompletedEvent struct {
	*BaseEvent
	ContentType string
	Size        int
	Duration    time.Duration
}

// NewRenderCompletedEvent 创建渲染完成事件
func NewRenderCompletedEvent(contentType string, size int, duration time.Duration) *RenderCompletedEvent {
	event := &RenderCompletedEvent{
		BaseEvent:  NewBaseEvent(EventTypeRenderCompleted, nil),
		ContentType: contentType,
		Size:       size,
		Duration:   duration,
	}
	event.data = map[string]interface{}{
		"content_type": contentType,
		"size":        size,
		"duration":    duration,
	}
	return event
}

// SystemErrorEvent 系统错误事件
type SystemErrorEvent struct {
	*BaseEvent
	Error     error
	Component string
	Context   map[string]interface{}
}

// NewSystemErrorEvent 创建系统错误事件
func NewSystemErrorEvent(err error, component string, context map[string]interface{}) *SystemErrorEvent {
	event := &SystemErrorEvent{
		BaseEvent: NewBaseEvent(EventTypeSystemError, nil),
		Error:     err,
		Component: component,
		Context:   context,
	}
	event.data = map[string]interface{}{
		"error":     err,
		"component": component,
		"context":   context,
	}
	return event
}

// SystemWarningEvent 系统警告事件
type SystemWarningEvent struct {
	*BaseEvent
	Message   string
	Component string
	Context   map[string]interface{}
}

// NewSystemWarningEvent 创建系统警告事件
func NewSystemWarningEvent(message, component string, context map[string]interface{}) *SystemWarningEvent {
	event := &SystemWarningEvent{
		BaseEvent: NewBaseEvent(EventTypeSystemWarning, nil),
		Message:   message,
		Component: component,
		Context:   context,
	}
	event.data = map[string]interface{}{
		"message":   message,
		"component": component,
		"context":   context,
	}
	return event
}

// SystemInfoEvent 系统信息事件
type SystemInfoEvent struct {
	*BaseEvent
	Message   string
	Component string
	Context   map[string]interface{}
}

// NewSystemInfoEvent 创建系统信息事件
func NewSystemInfoEvent(message, component string, context map[string]interface{}) *SystemInfoEvent {
	event := &SystemInfoEvent{
		BaseEvent: NewBaseEvent(EventTypeSystemInfo, nil),
		Message:   message,
		Component: component,
		Context:   context,
	}
	event.data = map[string]interface{}{
		"message":   message,
		"component": component,
		"context":   context,
	}
	return event
}

// KeyEvent 键盘事件包装
type KeyEvent struct {
	*BaseEvent
	Key tea.KeyMsg
}

// NewKeyEvent 创建键盘事件
func NewKeyEvent(key tea.KeyMsg) *KeyEvent {
	event := &KeyEvent{
		BaseEvent: NewBaseEvent("key.pressed", nil),
		Key:       key,
	}
	event.data = map[string]interface{}{
		"key": key,
	}
	return event
}

// WindowEvent 窗口事件包装
type WindowEvent struct {
	*BaseEvent
	Size tea.WindowSizeMsg
}

// NewWindowEvent 创建窗口事件
func NewWindowEvent(size tea.WindowSizeMsg) *WindowEvent {
	event := &WindowEvent{
		BaseEvent: NewBaseEvent("window.resized", nil),
		Size:      size,
	}
	event.data = map[string]interface{}{
		"size": size,
	}
	return event
}