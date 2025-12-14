package tui

import (
	"fmt"

	"github.com/Zacy-Sokach/PolyAgent/internal/mcp"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/textarea"
)

// Model 是 Bubble Tea 的主模型，现在作为各个模块的协调器
type Model struct {
	state    *RefactoredModelState
	stream   *TempStreamHandler
	render   *TempRenderManager
	command  *TempCommandProcessor
	// handlerChain *MessageHandlerChain // 暂时禁用
	eventBus    EventBus
}

// InitialModel 创建初始模型（保持向后兼容）
func InitialModel(apiKey string, toolManager interface{}) Model {
	// 将ToolManager转换为ToolRegistry
	var toolRegistry *mcp.ToolRegistry
	if toolManager != nil {
		// 这里需要适配，暂时使用默认工具注册表
		toolRegistry = mcp.DefaultToolRegistry(nil)
	}
	factory := NewModelFactory()
	return factory.CreateModel(apiKey, toolRegistry)
}

// NewModel 使用工厂模式创建模型
func NewModel(apiKey string, toolRegistry *mcp.ToolRegistry) Model {
	factory := NewModelFactory()
	return factory.CreateModel(apiKey, toolRegistry)
}

// NewModelWithContainer 使用容器创建模型
func NewModelWithContainer(container Container) Model {
	factory := NewModelFactory()
	return factory.CreateModelWithContainer(container)
}

// NewModelFromConfig 从配置创建模型
func NewModelFromConfig(config ModelConfig) (Model, error) {
	factory := NewConfigurableModelFactory()
	return factory.CreateModelFromConfig(config)
}

// Init 初始化模型
func (m *Model) Init() tea.Cmd {
	// 初始化消息处理器链
	m.initializeHandlerChain()
	
	return tea.Batch(
		tea.WindowSize(),
		textarea.Blink,
	)
}

// initializeHandlerChain 初始化消息处理器链（暂时禁用）
func (m *Model) initializeHandlerChain() {
	// TODO: 重新启用消息处理器链
	// 当前为了编译通过，暂时禁用
}

// initializeEventHandlers 初始化事件处理器（暂时禁用）
func (m *Model) initializeEventHandlers() {
	// TODO: 重新启用事件处理器
	// 当前为了编译通过，暂时禁用
}

// GetEventBus 获取事件总线
func (m *Model) GetEventBus() EventBus {
	return m.eventBus
}

// Update 更新模型状态（简化版本）
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	
	// 更新UI组件
	uiManager := m.state.GetUIManager()
	var uiCmd tea.Cmd
	textarea, uiCmd := uiManager.GetTextarea().Update(msg)
	uiManager.SetTextarea(textarea)
	if uiCmd != nil {
		cmds = append(cmds, uiCmd)
	}
	
	viewport, uiCmd := uiManager.GetViewport().Update(msg)
	uiManager.SetViewport(viewport)
	if uiCmd != nil {
		cmds = append(cmds, uiCmd)
	}
	
	return m, tea.Batch(cmds...)
}

// View 渲染视图
func (m *Model) View() string {
	uiManager := m.state.GetUIManager()
	if !uiManager.IsReady() {
		return "初始化中..."
	}

	return fmt.Sprintf(
		"%s\n\n%s\n%s",
		uiManager.GetViewport().View(),
		uiManager.GetTextarea().View(),
		"帮助信息", // 临时替换，因为TempRenderManager没有HelpView方法
	)
}