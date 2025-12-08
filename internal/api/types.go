package api

import (
	"encoding/json"
)

type Message struct {
	Role       string          `json:"role"`
	Content    json.RawMessage `json:"content"`
	ToolCalls  []ToolCall      `json:"tool_calls,omitempty"`
	ToolCallID string          `json:"tool_call_id,omitempty"`
	Name       string          `json:"name,omitempty"`
}

type ChatRequest struct {
	Model       string          `json:"model"`
	Messages    []Message       `json:"messages"`
	Stream      bool            `json:"stream"`
	MaxTokens   int             `json:"max_tokens,omitempty"`
	Temperature float64         `json:"temperature,omitempty"`
	Thinking    *Thinking       `json:"thinking,omitempty"`
	Tools       []Tool          `json:"tools,omitempty"`
	ToolChoice  json.RawMessage `json:"tool_choice,omitempty"`
}

type Thinking struct {
	Type string `json:"type"`
}

type ChatResponse struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
}

type Choice struct {
	Index        int        `json:"index"`
	Delta        *Delta     `json:"delta,omitempty"`
	Message      *Message   `json:"message,omitempty"`
	FinishReason string     `json:"finish_reason"`
	ToolCalls    []ToolCall `json:"tool_calls,omitempty"`
}

type Delta struct {
	Role             string     `json:"role,omitempty"`
	Content          string     `json:"content,omitempty"`
	ReasoningContent string     `json:"reasoning_content,omitempty"`
	ToolCalls        []ToolCall `json:"tool_calls,omitempty"`
}

type StreamChunk struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
}

// 工具相关类型
type Tool struct {
	Type     string       `json:"type"`
	Function ToolFunction `json:"function"`
}

type ToolFunction struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	Parameters  map[string]interface{} `json:"parameters"`
}

type ToolCall struct {
	ID       string           `json:"id"`
	Type     string           `json:"type"`
	Function ToolCallFunction `json:"function"`
}

type ToolCallFunction struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}

// 工具调用结果
type ToolResult struct {
	ToolCallID string          `json:"tool_call_id"`
	Content    json.RawMessage `json:"content"`
}

// 创建文本消息
func TextMessage(role, content string) Message {
	contentBytes, _ := json.Marshal(content)
	return Message{
		Role:    role,
		Content: contentBytes,
	}
}

// 创建工具调用消息
func ToolCallMessage(toolCalls []ToolCall) Message {
	// 根据 OpenAI 格式，工具调用消息的 content 应该为 null，tool_calls 在顶层
	return Message{
		Role:      "assistant",
		Content:   json.RawMessage("null"),
		ToolCalls: toolCalls,
	}
}

// 创建工具结果消息
func ToolResultMessage(toolCallID string, result interface{}) Message {
	resultBytes, _ := json.Marshal(result)

	// 根据 OpenAI 格式，工具结果消息直接使用结果JSON，不要双重编码
	return Message{
		Role:       "tool",
		Content:    resultBytes,
		ToolCallID: toolCallID,
		// 注意：OpenAI 示例中有 name 字段，但可能不是必需的
	}
}

// 创建带名称的工具结果消息
func ToolResultMessageWithName(toolCallID, name string, result interface{}) Message {
	resultBytes, _ := json.Marshal(result)

	return Message{
		Role:       "tool",
		Content:    resultBytes,
		ToolCallID: toolCallID,
		Name:       name,
	}
}
