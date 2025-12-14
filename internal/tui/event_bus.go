package tui

import (
	"sync"
	"time"
)

// Event 事件接口
type Event interface {
	// Type 事件类型
	Type() string
	// Data 事件数据
	Data() interface{}
	// Timestamp 事件时间戳
	Timestamp() time.Time
}

// EventHandler 事件处理器接口
type EventHandler interface {
	// CanHandle 检查是否可以处理该事件
	CanHandle(event Event) bool
	
	// Handle 处理事件
	Handle(event Event) error
	
	// Priority 处理优先级，数值越小优先级越高
	Priority() int
}

// EventBus 事件总线接口
type EventBus interface {
	// Subscribe 订阅事件
	Subscribe(eventType string, handler EventHandler)
	
	// Unsubscribe 取消订阅事件
	Unsubscribe(eventType string, handler EventHandler)
	
	// Publish 发布事件
	Publish(event Event)
	
	// PublishAsync 异步发布事件
	PublishAsync(event Event)
	
	// Clear 清空所有订阅
	Clear()
}

// BaseEvent 基础事件实现
type BaseEvent struct {
	eventType string
	data      interface{}
	timestamp time.Time
}

// NewBaseEvent 创建基础事件
func NewBaseEvent(eventType string, data interface{}) *BaseEvent {
	return &BaseEvent{
		eventType: eventType,
		data:      data,
		timestamp: time.Now(),
	}
}

// Type 事件类型
func (e *BaseEvent) Type() string {
	return e.eventType
}

// Data 事件数据
func (e *BaseEvent) Data() interface{} {
	return e.data
}

// Timestamp 事件时间戳
func (e *BaseEvent) Timestamp() time.Time {
	return e.timestamp
}

// MemoryEventBus 内存事件总线实现
type MemoryEventBus struct {
	handlers map[string][]EventHandler
	mutex    sync.RWMutex
}

// NewMemoryEventBus 创建内存事件总线
func NewMemoryEventBus() *MemoryEventBus {
	return &MemoryEventBus{
		handlers: make(map[string][]EventHandler),
	}
}

// Subscribe 订阅事件
func (bus *MemoryEventBus) Subscribe(eventType string, handler EventHandler) {
	bus.mutex.Lock()
	defer bus.mutex.Unlock()
	
	if bus.handlers[eventType] == nil {
		bus.handlers[eventType] = []EventHandler{}
	}
	
	bus.handlers[eventType] = append(bus.handlers[eventType], handler)
	
	// 按优先级排序
	handlers := bus.handlers[eventType]
	for i := 0; i < len(handlers)-1; i++ {
		for j := i + 1; j < len(handlers); j++ {
			if handlers[i].Priority() > handlers[j].Priority() {
				handlers[i], handlers[j] = handlers[j], handlers[i]
			}
		}
	}
}

// Unsubscribe 取消订阅事件
func (bus *MemoryEventBus) Unsubscribe(eventType string, handler EventHandler) {
	bus.mutex.Lock()
	defer bus.mutex.Unlock()
	
	handlers := bus.handlers[eventType]
	for i, h := range handlers {
		if h == handler {
			bus.handlers[eventType] = append(handlers[:i], handlers[i+1:]...)
			break
		}
	}
}

// Publish 发布事件
func (bus *MemoryEventBus) Publish(event Event) {
	bus.mutex.RLock()
	handlers := bus.handlers[event.Type()]
	bus.mutex.RUnlock()
	
	for _, handler := range handlers {
		if handler.CanHandle(event) {
			handler.Handle(event) // 忽略错误，保持简单
		}
	}
}

// PublishAsync 异步发布事件
func (bus *MemoryEventBus) PublishAsync(event Event) {
	go bus.Publish(event)
}

// Clear 清空所有订阅
func (bus *MemoryEventBus) Clear() {
	bus.mutex.Lock()
	defer bus.mutex.Unlock()
	
	bus.handlers = make(map[string][]EventHandler)
}

// 全局事件总线实例
var (
	globalEventBus EventBus
	eventBusOnce   sync.Once
)

// GetGlobalEventBus 获取全局事件总线
func GetGlobalEventBus() EventBus {
	eventBusOnce.Do(func() {
		globalEventBus = NewMemoryEventBus()
	})
	return globalEventBus
}

// 事件类型常量
const (
	// UI事件
	EventTypeUIUpdate       = "ui.update"
	EventTypeUIResize       = "ui.resize"
	EventTypeUIFocusChanged = "ui.focus_changed"
	
	// 消息事件
	EventTypeMessageAdded   = "message.added"
	EventTypeMessageUpdated = "message.updated"
	EventTypeMessageCleared = "message.cleared"
	
	// 流式事件
	EventTypeStreamStarted   = "stream.started"
	EventTypeStreamChunk     = "stream.chunk"
	EventTypeStreamFinished  = "stream.finished"
	EventTypeStreamError     = "stream.error"
	
	// 工具事件
	EventTypeToolCalled    = "tool.called"
	EventTypeToolCompleted = "tool.completed"
	EventTypeToolFailed    = "tool.failed"
	
	// 性能事件
	EventTypePerformanceWarning = "performance.warning"
	EventTypeRenderStarted      = "render.started"
	EventTypeRenderCompleted    = "render.completed"
	
	// 系统事件
	EventTypeSystemError   = "system.error"
	EventTypeSystemWarning = "system.warning"
	EventTypeSystemInfo    = "system.info"
)