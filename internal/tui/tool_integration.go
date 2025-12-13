package tui

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/Zacy-Sokach/PolyAgent/internal/api"
	"github.com/Zacy-Sokach/PolyAgent/internal/mcp"
)

// StreamResult 流式结果
type StreamResult struct {
	Content   string
	Reasoning string
	ToolCalls []api.ToolCall
	Error     error
}

// ToolManager 工具管理器
type ToolManager struct {
	tools       []api.Tool
	mcpRegistry *mcp.ToolRegistry
}

// NewToolManager 创建工具管理器
func NewToolManager(fileEngineConfig ...*mcp.FileEngineConfig) *ToolManager {
	// 使用提供的配置或默认配置
	var engineConfig *mcp.FileEngineConfig
	if len(fileEngineConfig) > 0 && fileEngineConfig[0] != nil {
		engineConfig = fileEngineConfig[0]
	} else {
		// 创建默认配置（当前工作目录）
		wd, _ := os.Getwd()
		defaultConfig := mcp.FileEngineConfig{
			AllowedRoots:    []string{wd},
			BlacklistedExts: []string{".exe", ".dll", ".so", ".dylib", ".bin"},
			MaxFileSize:     10 * 1024 * 1024,
			EnableCache:     true,
			BackupDir:       ".polyagent-backups",
		}
		engineConfig = &defaultConfig
	}
	
	// 创建 MCP 工具注册表
	mcpRegistry := mcp.DefaultToolRegistry(engineConfig)

	// 获取 MCP 工具并转换为 API 工具格式
	var apiTools []api.Tool
	if mcpRegistry != nil {
		for _, mcpTool := range mcpRegistry.ListTools() {
			handler, ok := mcpRegistry.GetTool(mcpTool.Name)
			if !ok {
				continue
			}
			// 假设 handler.GetSchema() 返回 map[string]interface{}
			apiTools = append(apiTools, api.Tool{
				Type: "function",
				Function: api.ToolFunction{
					Name:        mcpTool.Name,
					Description: mcpTool.Description,
					Parameters:  handler.GetSchema(),
				},
			})
		}
	}

	return &ToolManager{
		tools:       apiTools,
		mcpRegistry: mcpRegistry,
	}
}

// GetToolsForAPI 获取API工具列表
func (tm *ToolManager) GetToolsForAPI() []api.Tool {
	return tm.tools
}

// FormatToolCallForDisplay 格式化工具调用用于显示
func (tm *ToolManager) FormatToolCallForDisplay(toolCall api.ToolCall) string {
	// 为了兼容Arguments可能是json.RawMessage或string，这里尝试进行一次解码以美化输出
	var args interface{}
	if err := json.Unmarshal(toolCall.Function.Arguments, &args); err != nil {
		args = string(toolCall.Function.Arguments)
	}

	argsFormatted, _ := json.MarshalIndent(args, "", "  ")
	return fmt.Sprintf("调用工具: %s\n参数:\n%s", toolCall.Function.Name, argsFormatted)
}

// HandleToolCalls 处理工具调用
func (tm *ToolManager) HandleToolCalls(toolCalls []api.ToolCall) ([]api.Message, error) {
	var messages []api.Message

	for _, toolCall := range toolCalls {
		// fmt.Printf("[ToolManager] 处理工具调用: %s\n", toolCall.Function.Name)
		result, err := tm.executeToolCall(toolCall)
		if err != nil {
			// 将错误也作为工具结果返回给 LLM
			// 必须确保结果是有效的 JSON 格式，否则 json.RawMessage 再次序列化时会报错
			errorStr := fmt.Sprintf("Error: 工具 '%s' 执行失败: %v", toolCall.Function.Name, err)
			errorJSON, _ := json.Marshal(errorStr)
			result = string(errorJSON)
		}

		// fmt.Printf("[ToolManager] 工具结果长度: %d\n", len(result))
		/*
			if len(result) > 0 {
				preview := result
				if len(preview) > 100 {
					preview = preview[:100]
				}
				fmt.Printf("[ToolManager] 工具结果前100字符: %s...\n", preview)
			}
		*/

		// 必须将结果字符串序列化为 JSON 格式，因为 Content 是 json.RawMessage
		// 如果直接使用原始字符串，序列化时会因为不是有效 JSON 而报错
		resultBytes, err := json.Marshal(result)
		if err != nil {
			// 如果序列化失败，使用错误信息
			resultBytes, _ = json.Marshal(fmt.Sprintf("序列化结果失败: %v", err))
		}

		messages = append(messages, api.Message{
			Role:       "tool",
			Content:    resultBytes,
			ToolCallID: toolCall.ID,
			Name:       toolCall.Function.Name,
		})
	}

	return messages, nil
}

// executeToolCall 执行单个工具调用
func (tm *ToolManager) executeToolCall(toolCall api.ToolCall) (string, error) {
	// 添加恢复机制防止panic
	defer func() {
		if r := recover(); r != nil {
			// fmt.Printf("[ToolManager] executeToolCall 恢复panic: %v\n", r)
		}
	}()

	var args map[string]interface{}
	var firstErr error

	// 检查参数是否为空
	if len(toolCall.Function.Arguments) == 0 {
		args = make(map[string]interface{})
		return tm.executeWithArgs(toolCall.Function.Name, args)
	}

	// 1. 优先尝试直接解析为 JSON 对象 (Standard approach)
	// 例如: {"path": "./file.txt"}
	if err := json.Unmarshal(toolCall.Function.Arguments, &args); err == nil {
		return tm.executeWithArgs(toolCall.Function.Name, args)
	} else {
		firstErr = err // 记录第一次解析失败的错误
	}

	// 2. 如果失败，尝试解析为 JSON 字符串 (Double encoded JSON)
	// 例如: "{\"path\": \"./file.txt\"}"
	var jsonStr string
	if err2 := json.Unmarshal(toolCall.Function.Arguments, &jsonStr); err2 != nil {
		// 真正的解析失败：既不是 JSON 对象，也不是包含 JSON 字符串的 JSON 字符串
		return "", fmt.Errorf("解析参数失败: 既不是有效的JSON对象也不是JSON字符串: %w (原始错误: %v)", firstErr, err2)
	}

	// 3. 解析 JSON 字符串中的对象
	if err3 := json.Unmarshal([]byte(jsonStr), &args); err3 != nil {
		return "", fmt.Errorf("解析二次编码的JSON参数失败: %w", err3)
	}

	return tm.executeWithArgs(toolCall.Function.Name, args)
}

// executeWithArgs 使用解析后的参数执行工具
func (tm *ToolManager) executeWithArgs(toolName string, args map[string]interface{}) (string, error) {
	// 添加恢复机制防止panic
	defer func() {
		if r := recover(); r != nil {
			// fmt.Printf("[ToolManager] executeWithArgs 恢复panic: %v\n", r)
		}
	}()

	// 检查工具注册表是否存在
	if tm.mcpRegistry == nil {
		return "", fmt.Errorf("工具注册表未初始化")
	}

	// 检查 MCP 中是否存在该工具
	_, exists := tm.mcpRegistry.GetTool(toolName)
	if !exists {
		return "", fmt.Errorf("未知工具: %s", toolName)
	}

	// 创建 MCP 请求
	mcpRequest := mcp.CallToolRequest{
		Name:      toolName,
		Arguments: args,
	}

	// 执行工具调用（添加错误处理）
	result, err := tm.mcpRegistry.HandleCallTool(mcpRequest)
	if err != nil {
		return "", fmt.Errorf("MCP工具执行失败: %w", err)
	}

	// 检查结果
	if result == nil {
		return "", fmt.Errorf("工具返回空结果")
	}

	if len(result.Content) > 0 {
		return result.Content[0].Text, nil
	}

	return "success", nil
}
