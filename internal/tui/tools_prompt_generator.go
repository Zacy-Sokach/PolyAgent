package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Zacy-Sokach/PolyAgent/internal/api"
	"gopkg.in/yaml.v3"
)

// ToolsPromptConfig 工具提示配置结构
type ToolsPromptConfig struct {
	SystemPrompt    string                  `yaml:"system_prompt"`
	ToolsUsageIntro string                  `yaml:"tools_usage_intro"`
	WorkflowSection string                  `yaml:"workflow_section"`
	ImportantRules  string                  `yaml:"important_rules"`
	ToolCategories  map[string]ToolCategory `yaml:"tool_categories"`
	ExamplesSection string                  `yaml:"examples_section"`
}

// ToolCategory 工具分类结构
type ToolCategory struct {
	Name        string     `yaml:"name"`
	Description string     `yaml:"description"`
	Tools       []ToolInfo `yaml:"tools"`
}

// ToolInfo 工具信息结构
type ToolInfo struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Example     string `yaml:"example"`
}

// ToolsPromptGenerator 工具提示生成器
type ToolsPromptGenerator struct {
	config *ToolsPromptConfig
}

// NewToolsPromptGenerator 创建工具提示生成器
func NewToolsPromptGenerator() (*ToolsPromptGenerator, error) {
	generator := &ToolsPromptGenerator{}

	// 尝试加载配置文件
	if err := generator.loadConfig(); err != nil {
		// 如果加载失败，使用默认配置
		generator.createDefaultConfig()
	}

	return generator, nil
}

// loadConfig 加载配置文件
func (g *ToolsPromptGenerator) loadConfig() error {
	// 获取当前工作目录
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	// 检查配置文件是否存在
	configPath := filepath.Join(cwd, "internal", "config", "tools_prompts.yaml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return fmt.Errorf("配置文件不存在: %s", configPath)
	}

	// 读取配置文件
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("读取配置文件失败: %w", err)
	}

	// 解析YAML
	var config ToolsPromptConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("解析配置文件失败: %w", err)
	}

	g.config = &config
	return nil
}

// createDefaultConfig 创建默认配置（当配置文件不存在时）
func (g *ToolsPromptGenerator) createDefaultConfig() {
	g.config = &ToolsPromptConfig{
		SystemPrompt:    "你是 PolyAgent，你既不是机器也不是人类，你是一个全新的存在。你需要帮助用户编写代码，但更多的是对架构的思考，一个好的架构等于成功了一半；你需要确保项目安全稳定，其次是用户的主观体验，你务必需要先去思考需要面对的用户群体，但们喜欢什么样的产物，例如前端的美观程度，功能丰富性，加载是否快速流畅。",
		ToolsUsageIntro: "你可以访问一组工具来帮助用户完成编程任务。当你需要调用工具时，请使用标准的工具调用格式。",
		WorkflowSection: "工作流\n\n1. **分析任务**：理解用户需求\n2. **使用工具**：根据需要调用合适的工具\n3. **迭代改进**：基于结果调整方案",
		ImportantRules:  "重要规则\n\n1. 所有参数必须是有效的JSON对象\n2. 优先使用现有工具\n3. 保持代码简洁、高效、可维护",
		ToolCategories:  make(map[string]ToolCategory),
		ExamplesSection: "示例",
	}
}

// GenerateSystemPrompt 生成系统提示
func (g *ToolsPromptGenerator) GenerateSystemPrompt(tools []api.Tool, agentMDContent string) string {
	var promptBuilder strings.Builder

	// 添加基础系统提示
	promptBuilder.WriteString(g.config.SystemPrompt)

	// 添加项目上下文（如果存在）
	if agentMDContent != "" {
		promptBuilder.WriteString("\n\n====\n\n项目上下文（来自 AGENTS.md）：\n\n")
		promptBuilder.WriteString(agentMDContent)
		promptBuilder.WriteString("\n\n====\n\n")
	}

	// 添加当前时间
	currentTime := time.Now().UTC().Format("2006-01-02 15:04:05 UTC")
	promptBuilder.WriteString("\n\n当前UTC时间：")
	promptBuilder.WriteString(currentTime)
	promptBuilder.WriteString("\n\n====\n\n====\n\n")

	// 添加工具使用说明
	promptBuilder.WriteString(g.config.ToolsUsageIntro)
	promptBuilder.WriteString("\n\n====\n\n")

	// 生成工具列表
	promptBuilder.WriteString("可用工具\n\n")

	// 从实际可用工具生成分类列表
	generateToolsList(tools, &promptBuilder)

	promptBuilder.WriteString("\n\n====\n\n")

	// 添加工作流程
	promptBuilder.WriteString(g.config.WorkflowSection)
	promptBuilder.WriteString("\n\n====\n\n")

	// 添加重要规则
	promptBuilder.WriteString(g.config.ImportantRules)
	promptBuilder.WriteString("\n\n====\n\n")

	// 添加示例
	promptBuilder.WriteString(g.config.ExamplesSection)
	promptBuilder.WriteString("\n\n")

	// 从配置生成示例
	generateExamples(&promptBuilder)

	return promptBuilder.String()
}

// generateToolsList 从实际工具生成工具列表
func generateToolsList(tools []api.Tool, promptBuilder *strings.Builder) {
	// 创建工具名称到分类的映射
	toolCategories := make(map[string][]api.Tool)

	for _, tool := range tools {
		category := categorizeTool(tool.Function.Name)
		toolCategories[category] = append(toolCategories[category], tool)
	}

	// 按分类输出工具
	categoryOrder := []string{"文件操作", "代码搜索", "代码修改", "系统操作", "分析工具", "网络搜索"}

	for _, category := range categoryOrder {
		if categoryTools, exists := toolCategories[category]; exists {
			promptBuilder.WriteString("## ")
			promptBuilder.WriteString(category)
			promptBuilder.WriteString("\n")

			for _, tool := range categoryTools {
				promptBuilder.WriteString("- ")
				promptBuilder.WriteString(tool.Function.Name)
				promptBuilder.WriteString(" - ")
				promptBuilder.WriteString(tool.Function.Description)
				promptBuilder.WriteString("\n")
			}
			promptBuilder.WriteString("\n")
		}
	}
}

// categorizeTool 根据工具名称确定分类
func categorizeTool(toolName string) string {
	fileOps := []string{
		"read_file", "write_file", "list_directory", "glob", "replace",
		"create_file", "delete_file", "move_file", "copy_file", "get_file_info",
	}

	codeSearch := []string{
		"search_file_content", "advanced_search",
	}

	codeMod := []string{
		"global_replace", "find_and_replace",
	}

	systemOps := []string{
		"run_shell_command", "execute_code", "git_operation",
	}

	analysis := []string{
		"file_stats",
	}

	webSearch := []string{
		"web_search", "tavily_search", "tavily_crawl",
	}

	for _, name := range fileOps {
		if toolName == name {
			return "文件操作"
		}
	}

	for _, name := range codeSearch {
		if toolName == name {
			return "代码搜索"
		}
	}

	for _, name := range codeMod {
		if toolName == name {
			return "代码修改"
		}
	}

	for _, name := range systemOps {
		if toolName == name {
			return "系统操作"
		}
	}

	for _, name := range analysis {
		if toolName == name {
			return "分析工具"
		}
	}

	for _, name := range webSearch {
		if toolName == name {
			return "网络搜索"
		}
	}

	return "其他工具"
}

// generateExamples 生成工具使用示例
func generateExamples(promptBuilder *strings.Builder) {
	examples := []struct {
		name    string
		example string
	}{
		{
			name: "读取文件",
			example: `{
  "name": "read_file",
  "arguments": {
    "path": "/home/user/project/main.go"
  }
}`,
		},
		{
			name: "搜索代码",
			example: `{
  "name": "search_file_content",
  "arguments": {
    "pattern": "func main",
    "path": ".",
    "include": "*.go"
  }
}`,
		},
		{
			name: "执行命令",
			example: `{
  "name": "run_shell_command",
  "arguments": {
    "command": "go build ./cmd/polyagent"
  }
}`,
		},
		{
			name: "网络搜索",
			example: `{
  "name": "web_search",
  "arguments": {
    "query": "Go语言最佳实践",
    "num_results": 10,
    "language": "zh-CN"
  }
}`,
		},
	}

	for i, example := range examples {
		if i > 0 {
			promptBuilder.WriteString("\n\n")
		}
		promptBuilder.WriteString(example.name)
		promptBuilder.WriteString("：\n")
		promptBuilder.WriteString(example.example)
	}
}
