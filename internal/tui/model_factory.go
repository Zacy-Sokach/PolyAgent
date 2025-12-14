package tui

import (
	"fmt"
	
	"github.com/Zacy-Sokach/PolyAgent/internal/mcp"
)

// ModelFactory 模型工厂接口
type ModelFactory interface {
	CreateModel(apiKey string, toolRegistry *mcp.ToolRegistry) Model
	CreateModelWithContainer(container Container) Model
}

// DefaultModelFactory 默认模型工厂实现
type DefaultModelFactory struct{}

// NewModelFactory 创建新的模型工厂
func NewModelFactory() ModelFactory {
	return &DefaultModelFactory{}
}

// CreateModel 创建模型（使用内部容器）
func (f *DefaultModelFactory) CreateModel(apiKey string, toolRegistry *mcp.ToolRegistry) Model {
	container := NewDIContainer(apiKey, toolRegistry)
	return f.CreateModelWithContainer(container)
}

// CreateModelWithContainer 使用容器创建模型
func (f *DefaultModelFactory) CreateModelWithContainer(container Container) Model {
	model := Model{
		state:     container.ResolveModelState(),
		stream:    container.ResolveStreamHandler(),
		render:    container.ResolveRenderManager(),
		command:   container.ResolveCommandProcessor(),
		eventBus:  GetGlobalEventBus(),
	}
	
	// 延迟初始化消息处理器链
	model.initializeHandlerChain()
	
	// 初始化事件处理器
	model.initializeEventHandlers()
	
	return model
}

// ModelBuilder 模型构建器，提供更灵活的构建方式
type ModelBuilder struct {
	apiKey       string
	toolRegistry *mcp.ToolRegistry
	container    Container
}

// NewModelBuilder 创建新的模型构建器
func NewModelBuilder() *ModelBuilder {
	return &ModelBuilder{}
}

// WithAPIKey 设置API密钥
func (b *ModelBuilder) WithAPIKey(apiKey string) *ModelBuilder {
	b.apiKey = apiKey
	return b
}

// WithToolRegistry 设置工具注册表
func (b *ModelBuilder) WithToolRegistry(toolRegistry *mcp.ToolRegistry) *ModelBuilder {
	b.toolRegistry = toolRegistry
	return b
}

// WithContainer 设置容器
func (b *ModelBuilder) WithContainer(container Container) *ModelBuilder {
	b.container = container
	return b
}

// Build 构建模型
func (b *ModelBuilder) Build() (Model, error) {
	// 验证必要参数
	if b.apiKey == "" {
		return Model{}, fmt.Errorf("API key is required")
	}
	
	var container Container
	if b.container != nil {
		container = b.container
	} else {
		if b.toolRegistry == nil {
			b.toolRegistry = mcp.DefaultToolRegistry(nil)
		}
		container = NewDIContainer(b.apiKey, b.toolRegistry)
	}
	
	factory := NewModelFactory()
	return factory.CreateModelWithContainer(container), nil
}

// ModelConfig 模型配置
type ModelConfig struct {
	APIKey       string
	ToolRegistry *mcp.ToolRegistry
	Container    Container
	
	// 扩展配置
	MaxMessages  int
	EnableCoT    bool
	ShowCoT      bool
	MaxRetries   int
}

// DefaultModelConfig 默认模型配置
func DefaultModelConfig(apiKey string) ModelConfig {
	return ModelConfig{
		APIKey:      apiKey,
		MaxMessages: 50,
		EnableCoT:   true,
		ShowCoT:     true,
		MaxRetries:  3,
	}
}

// ConfigurableModelFactory 可配置的模型工厂
type ConfigurableModelFactory struct{}

// NewConfigurableModelFactory 创建可配置的模型工厂
func NewConfigurableModelFactory() *ConfigurableModelFactory {
	return &ConfigurableModelFactory{}
}

// CreateModelFromConfig 从配置创建模型
func (f *ConfigurableModelFactory) CreateModelFromConfig(config ModelConfig) (Model, error) {
	// 验证配置
	if config.APIKey == "" {
		return Model{}, fmt.Errorf("API key is required")
	}
	
	// 创建容器
	var container Container
	if config.Container != nil {
		container = config.Container
	} else {
		if config.ToolRegistry == nil {
			config.ToolRegistry = mcp.DefaultToolRegistry(nil)
		}
		container = NewDIContainer(config.APIKey, config.ToolRegistry)
	}
	
	// 应用配置到模型状态
	modelState := container.ResolveModelState()
	streamManager := modelState.GetStreamManager()
	
	// 配置消息管理器
	if config.MaxMessages > 0 {
		// TODO: 应用配置到消息管理器
	}
	
	// 配置流式管理器
	streamManager.SetCoTEnabled(config.EnableCoT)
	streamManager.SetCoTVisible(config.ShowCoT)
	if config.MaxRetries > 0 {
		streamManager.SetMaxRetries(config.MaxRetries)
	}
	
	// 创建模型
	factory := NewModelFactory()
	return factory.CreateModelWithContainer(container), nil
}