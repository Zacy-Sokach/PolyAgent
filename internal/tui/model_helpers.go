package tui

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/Zacy-Sokach/PolyAgent/internal/api"
)

// addSystemPromptIfNeeded 如果需要，添加系统提示
func addSystemPromptIfNeeded(messages []api.Message) []api.Message {
	// 检查是否已经有 PolyAgent 系统消息（避免重复添加）
	hasPolyAgentSystemMessage := false
	for _, msg := range messages {
		if msg.Role == "system" && strings.Contains(string(msg.Content), "你是 PolyAgent") {
			hasPolyAgentSystemMessage = true
			break
		}
	}

	// 如果没有 PolyAgent 系统消息，添加一个
	if !hasPolyAgentSystemMessage {
		// 读取 AGENT.md 文件（如果存在）
		agentMDContent := ""
		if content, err := readAgentMDFile(); err == nil && content != "" {
			agentMDContent = content
		}

		// 创建工具提示生成器
		promptGenerator, err := NewToolsPromptGenerator()
		if err != nil {
			// 如果创建生成器失败，使用简单的默认提示
			systemMessage := api.TextMessage("system", "你是 PolyAgent，一个AI编程助手。")
			return append([]api.Message{systemMessage}, messages...)
		}

		// 创建工具管理器获取可用工具
		toolManager := NewToolManager()
		tools := toolManager.GetToolsForAPI()

		// 生成动态系统提示
		systemPrompt := promptGenerator.GenerateSystemPrompt(tools, agentMDContent)
		systemMessage := api.TextMessage("system", systemPrompt)
		
		return append([]api.Message{systemMessage}, messages...)
	}

	return messages
}

// readAgentMDFile 读取 AGENT.md 文件内容
func readAgentMDFile() (string, error) {
	// 获取当前工作目录
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	// 检查 AGENT.md 文件是否存在
	agentMDPath := filepath.Join(cwd, "AGENT.md")
	if _, err := os.Stat(agentMDPath); os.IsNotExist(err) {
		return "", nil // 文件不存在，返回空字符串
	}

	// 读取文件内容
	content, err := os.ReadFile(agentMDPath)
	if err != nil {
		return "", err
	}

	// 返回完整内容，不再截断
	contentStr := string(content)
	return contentStr, nil
}
