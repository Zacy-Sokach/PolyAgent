package tui

import (
	"github.com/Zacy-Sokach/PolyAgent/internal/api"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
)

// RefactoredModelState 重构后的模型状态，使用组合模式
type RefactoredModelState struct {
	uiManager       *UIStateManager
	messageManager  *MessageManager
	streamManager   *StreamManager
	toolManager     *ToolManagerState
	// perfManager     *PerformanceManager // 暂时禁用
}

// NewRefactoredModelState 创建新的重构后模型状态
func NewRefactoredModelState(apiKey string, toolManager interface{}, commandParser *CommandParser) *RefactoredModelState {
	// 创建各个管理器
	uiManager := NewUIStateManager()
	messageManager := NewMessageManager(50) // 限制最多显示50条消息
	streamManager := NewStreamManager()
	toolManagerState := NewToolManagerState(toolManager, commandParser)
// 创建性能管理器（暂时禁用）
	// viewport := uiManager.GetViewport()
	// perfManager := NewPerformanceManager(&viewport)

	return &RefactoredModelState{
		uiManager:      uiManager,
		messageManager: messageManager,
		streamManager:  streamManager,
		toolManager:    toolManagerState,
		// perfManager:    perfManager, // 暂时禁用
	}
}

// GetUIManager 获取UI管理器
func (s *RefactoredModelState) GetUIManager() *UIStateManager {
	return s.uiManager
}

// GetMessageManager 获取消息管理器
func (s *RefactoredModelState) GetMessageManager() *MessageManager {
	return s.messageManager
}

// GetStreamManager 获取流式管理器
func (s *RefactoredModelState) GetStreamManager() *StreamManager {
	return s.streamManager
}

// GetToolManager 获取工具管理器
func (s *RefactoredModelState) GetToolManager() *ToolManagerState {
	return s.toolManager
}

// GetPerformanceManager 获取性能管理器（暂时禁用）
// func (s *RefactoredModelState) GetPerformanceManager() *PerformanceManager {
// 	return s.perfManager
// }

// 为了兼容现有代码，提供一些便捷方法
func (s *RefactoredModelState) GetViewport() viewport.Model {
	return s.uiManager.GetViewport()
}

func (s *RefactoredModelState) GetTextarea() textarea.Model {
	return s.uiManager.GetTextarea()
}

func (s *RefactoredModelState) SetViewport(vp viewport.Model) {
	s.uiManager.SetViewport(vp)
}

func (s *RefactoredModelState) SetTextarea(ta textarea.Model) {
	s.uiManager.SetTextarea(ta)
}

func (s *RefactoredModelState) IsReady() bool {
	return s.uiManager.IsReady()
}

func (s *RefactoredModelState) SetReady(ready bool) {
	s.uiManager.SetReady(ready)
}

func (s *RefactoredModelState) GetMessages() []Message {
	return s.messageManager.GetMessages()
}

func (s *RefactoredModelState) AddMessage(role, content string) {
	s.messageManager.AddMessage(role, content)
}

func (s *RefactoredModelState) GetAPIMessages() []api.Message {
	return s.messageManager.GetAPIMessages()
}

func (s *RefactoredModelState) AddAPIMessage(msg api.Message) {
	s.messageManager.AddAPIMessage(msg)
}

func (s *RefactoredModelState) IsThinking() bool {
	return s.streamManager.IsThinking()
}

func (s *RefactoredModelState) SetThinking(thinking bool) {
	s.streamManager.SetThinking(thinking)
}

func (s *RefactoredModelState) GetCurrentResponse() string {
	return s.streamManager.GetCurrentResponse()
}

func (s *RefactoredModelState) SetCurrentResponse(resp string) {
	s.streamManager.SetCurrentResponse(resp)
}

func (s *RefactoredModelState) GetCurrentThinking() string {
	return s.streamManager.GetCurrentThinking()
}

func (s *RefactoredModelState) SetCurrentThinking(think string) {
	s.streamManager.SetCurrentThinking(think)
}

func (s *RefactoredModelState) SaveHistory() {
	s.messageManager.SaveHistory()
	// 清理异步渲染器（暂时禁用）
	// if s.perfManager.asyncRenderer != nil {
	// 	s.perfManager.asyncRenderer.ClearCache()
	// }
}