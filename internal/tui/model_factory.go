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
	// For now, just return a basic model since the refactored structure is not yet implemented
	// TODO: Implement proper container-based model creation
	return Model{}
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
	
	// For now, return a basic model since the refactored structure is not yet implemented
	// TODO: Implement proper container-based model creation
	return Model{}, nil
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
	
	// For now, return a basic model since the refactored structure is not yet implemented
	// TODO: Implement proper container-based model creation with configuration
	return Model{}, nil
}