package tui

import (
	"context"
	"strings"
	"time"

	"github.com/Zacy-Sokach/PolyAgent/internal/api"
)

// StreamManager 管理流式响应状态
type StreamManager struct {
	thinking            bool
	currentResp         string
	currentThink        string
	streamCh            <-chan string
	reasoningCh         <-chan string
	toolCallCh          <-chan []api.ToolCall
	streamErrCh         <-chan error
	pendingToolCalls    []api.ToolCall
	streamBuffer        *strings.Builder
	lastChunkAt         time.Time
	pendingRender       string
	
	// 上下文和重试控制
	ctx                 context.Context
	cancel              context.CancelFunc
	retryCount          int
	maxRetries          int
	originalMessages    []api.Message
	
	// CoT 相关
	cotEnabled          bool
	cotVisible          bool
	cotHistory          []string
}

// NewStreamManager 创建新的流式管理器
func NewStreamManager() *StreamManager {
	ctx, cancel := context.WithCancel(context.Background())
	
	return &StreamManager{
		thinking:         false,
		currentResp:      "",
		currentThink:     "",
		pendingToolCalls: []api.ToolCall{},
		streamBuffer:     &strings.Builder{},
		lastChunkAt:      time.Now(),
		pendingRender:    "",
		ctx:              ctx,
		cancel:           cancel,
		retryCount:       0,
		maxRetries:       3,
		originalMessages: []api.Message{},
		cotEnabled:       true, // 默认启用CoT
		cotVisible:       true, // 默认显示思考过程
		cotHistory:       []string{},
	}
}

// IsThinking 检查是否正在思考
func (m *StreamManager) IsThinking() bool {
	return m.thinking
}

// SetThinking 设置思考状态
func (m *StreamManager) SetThinking(thinking bool) {
	m.thinking = thinking
}

// GetCurrentResponse 获取当前响应
func (m *StreamManager) GetCurrentResponse() string {
	return m.currentResp
}

// SetCurrentResponse 设置当前响应
func (m *StreamManager) SetCurrentResponse(resp string) {
	m.currentResp = resp
}

// AppendToCurrentResponse 追加内容到当前响应
func (m *StreamManager) AppendToCurrentResponse(chunk string) {
	m.currentResp += chunk
	m.lastChunkAt = time.Now()
}

// GetCurrentThinking 获取当前思考内容
func (m *StreamManager) GetCurrentThinking() string {
	return m.currentThink
}

// SetCurrentThinking 设置当前思考内容
func (m *StreamManager) SetCurrentThinking(think string) {
	m.currentThink = think
}

// AppendToCurrentThinking 追加内容到当前思考
func (m *StreamManager) AppendToCurrentThinking(chunk string) {
	m.currentThink += chunk
	
	// 记录思考历史（优化：限制历史记录数量）
	if len(m.cotHistory) == 0 || m.cotHistory[len(m.cotHistory)-1] != m.currentThink {
		// 限制历史记录最多20条，避免内存无限增长
		if len(m.cotHistory) >= 20 {
			// 移除最旧的记录，保持切片长度
			copy(m.cotHistory, m.cotHistory[1:])
			m.cotHistory = m.cotHistory[:19]
		}
		m.cotHistory = append(m.cotHistory, m.currentThink)
	}
}

// GetContext 获取上下文
func (m *StreamManager) GetContext() context.Context {
	return m.ctx
}

// GetCancelFunc 获取取消函数
func (m *StreamManager) GetCancelFunc() context.CancelFunc {
	return m.cancel
}

// ResetContext 重置上下文
func (m *StreamManager) ResetContext() {
	m.ctx, m.cancel = context.WithCancel(context.Background())
}

// GetRetryCount 获取重试次数
func (m *StreamManager) GetRetryCount() int {
	return m.retryCount
}

// SetRetryCount 设置重试次数
func (m *StreamManager) SetRetryCount(count int) {
	m.retryCount = count
}

// IncrementRetryCount 增加重试次数
func (m *StreamManager) IncrementRetryCount() {
	m.retryCount++
}

// GetMaxRetries 获取最大重试次数
func (m *StreamManager) GetMaxRetries() int {
	return m.maxRetries
}

// SetMaxRetries 设置最大重试次数
func (m *StreamManager) SetMaxRetries(maxRetries int) {
	m.maxRetries = maxRetries
}

// GetOriginalMessages 获取原始消息
func (m *StreamManager) GetOriginalMessages() []api.Message {
	return m.originalMessages
}

// SetOriginalMessages 设置原始消息
func (m *StreamManager) SetOriginalMessages(messages []api.Message) {
	m.originalMessages = messages
}

// IsCoTEnabled 检查是否启用CoT
func (m *StreamManager) IsCoTEnabled() bool {
	return m.cotEnabled
}

// SetCoTEnabled 设置CoT启用状态
func (m *StreamManager) SetCoTEnabled(enabled bool) {
	m.cotEnabled = enabled
}

// IsCoTVisible 检查是否显示CoT
func (m *StreamManager) IsCoTVisible() bool {
	return m.cotVisible
}

// SetCoTVisible 设置CoT可见状态
func (m *StreamManager) SetCoTVisible(visible bool) {
	m.cotVisible = visible
}

// ToggleCoTVisible 切换CoT可见状态
func (m *StreamManager) ToggleCoTVisible() {
	m.cotVisible = !m.cotVisible
}

// GetCoTHistory 获取CoT历史
func (m *StreamManager) GetCoTHistory() []string {
	return m.cotHistory
}

// GetPendingToolCalls 获取挂起的工具调用
func (m *StreamManager) GetPendingToolCalls() []api.ToolCall {
	return m.pendingToolCalls
}

// SetPendingToolCalls 设置挂起的工具调用
func (m *StreamManager) SetPendingToolCalls(calls []api.ToolCall) {
	m.pendingToolCalls = calls
}

// AddPendingToolCalls 添加挂起的工具调用
func (m *StreamManager) AddPendingToolCalls(calls []api.ToolCall) {
	m.pendingToolCalls = append(m.pendingToolCalls, calls...)
}

// ClearPendingToolCalls 清空挂起的工具调用
func (m *StreamManager) ClearPendingToolCalls() {
	m.pendingToolCalls = []api.ToolCall{}
}

// GetStreamBuffer 获取流式缓冲区
func (m *StreamManager) GetStreamBuffer() *strings.Builder {
	return m.streamBuffer
}

// ResetStreamBuffer 重置流式缓冲区
func (m *StreamManager) ResetStreamBuffer() {
	m.streamBuffer.Reset()
}

// GetLastChunkAt 获取最后一块数据的时间
func (m *StreamManager) GetLastChunkAt() time.Time {
	return m.lastChunkAt
}

// GetPendingRender 获取待渲染内容
func (m *StreamManager) GetPendingRender() string {
	return m.pendingRender
}

// SetPendingRender 设置待渲染内容
func (m *StreamManager) SetPendingRender(content string) {
	m.pendingRender = content
}

// ClearStreamData 清空流式数据
func (m *StreamManager) ClearStreamData() {
	m.currentResp = ""
	m.currentThink = ""
	m.cotHistory = []string{}
	m.retryCount = 0
	m.pendingToolCalls = []api.ToolCall{}
	m.streamBuffer.Reset()
}