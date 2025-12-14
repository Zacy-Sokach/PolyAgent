package tui

import "github.com/Zacy-Sokach/PolyAgent/internal/mcp"

// Container 依赖注入容器接口
type Container interface {
	// 解析各种处理器
	ResolveStreamHandler() *TempStreamHandler
	ResolveRenderManager() *TempRenderManager
	ResolveCommandProcessor() *TempCommandProcessor
	ResolveToolManager() *mcp.ToolRegistry
	
	// 解析各种管理器
	ResolveUIStateManager() *UIStateManager
	ResolveMessageManager() *MessageManager
	ResolveStreamManager() *StreamManager
	ResolveToolManagerState() *ToolManagerState
	// ResolvePerformanceManager() *PerformanceManager // 暂时禁用
	
	// 解析模型状态
	ResolveModelState() *RefactoredModelState
}

// DIContainer 依赖注入容器实现
type DIContainer struct {
	streamHandler   *TempStreamHandler
	renderManager   *TempRenderManager
	commandProcessor *TempCommandProcessor
	toolRegistry    *mcp.ToolRegistry
	
	uiStateManager   *UIStateManager
	messageManager  *MessageManager
	streamManager   *StreamManager
	toolManagerState *ToolManagerState
	// perfManager     *PerformanceManager // 暂时禁用
	
	modelState      *RefactoredModelState
}

// NewDIContainer 创建新的依赖注入容器
func NewDIContainer(apiKey string, toolRegistry *mcp.ToolRegistry) *DIContainer {
	// 创建管理器
	uiStateManager := NewUIStateManager()
	messageManager := NewMessageManager(50)
	streamManager := NewStreamManager()
	
	// 创建命令解析器
	commandParser := NewCommandParser()
	toolManagerState := NewToolManagerState(toolRegistry, commandParser)
	
	// 创建UI管理器后获取viewport用于性能管理器
	
	// 创建模型状态
	modelState := NewRefactoredModelState(apiKey, toolRegistry, commandParser)
	
	// 创建临时处理器（稍后替换为真正的处理器）
	streamHandler := NewTempStreamHandler()
	renderManager := NewTempRenderManager()
	commandProcessor := NewTempCommandProcessor()
	
	return &DIContainer{
		streamHandler:    streamHandler,
		renderManager:    renderManager,
		commandProcessor: commandProcessor,
		toolRegistry:     toolRegistry,
		
		uiStateManager:   uiStateManager,
		messageManager:  messageManager,
		streamManager:   streamManager,
		toolManagerState: toolManagerState,
		// perfManager:     perfManager, // 暂时禁用
		
		modelState:      modelState,
	}
}

// ResolveStreamHandler 解析流式处理器
func (c *DIContainer) ResolveStreamHandler() *TempStreamHandler {
	return c.streamHandler
}

// ResolveRenderManager 解析渲染管理器
func (c *DIContainer) ResolveRenderManager() *TempRenderManager {
	return c.renderManager
}

// ResolveCommandProcessor 解析命令处理器
func (c *DIContainer) ResolveCommandProcessor() *TempCommandProcessor {
	return c.commandProcessor
}

// ResolveToolManager 解析工具管理器
func (c *DIContainer) ResolveToolManager() *mcp.ToolRegistry {
	return c.toolRegistry
}

// ResolveUIStateManager 解析UI状态管理器
func (c *DIContainer) ResolveUIStateManager() *UIStateManager {
	return c.uiStateManager
}

// ResolveMessageManager 解析消息管理器
func (c *DIContainer) ResolveMessageManager() *MessageManager {
	return c.messageManager
}

// ResolveStreamManager 解析流式管理器
func (c *DIContainer) ResolveStreamManager() *StreamManager {
	return c.streamManager
}

// ResolveToolManagerState 解析工具管理器状态
func (c *DIContainer) ResolveToolManagerState() *ToolManagerState {
	return c.toolManagerState
}

// ResolvePerformanceManager 解析性能管理器（暂时禁用）
// func (c *DIContainer) ResolvePerformanceManager() *PerformanceManager {
// 	return c.perfManager
// }

// ResolveModelState 解析模型状态
func (c *DIContainer) ResolveModelState() *RefactoredModelState {
	return c.modelState
}