package api

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const (
	baseURL = "https://open.bigmodel.cn/api/paas/v4"
)

type Client struct {
	apiKey string
	client *http.Client
}

// NewClient 创建新的GLM-4.5 API客户端
// apiKey: GLM-4.5 API密钥
// 返回配置好的API客户端实例
func NewClient(apiKey string) *Client {
	return &Client{
		apiKey: apiKey,
		client: &http.Client{},
	}
}

// ChatCompletion 发送聊天补全请求到GLM-4.5 API
// messages: 消息历史数组
// stream: 是否使用流式响应
// tools: 可用的工具列表
// 返回聊天响应或错误
func (c *Client) ChatCompletion(messages []Message, stream bool, tools []Tool) (*ChatResponse, error) {
	req := ChatRequest{
		Model:       "glm-4.5",
		Messages:    messages,
		Stream:      stream,
		MaxTokens:   4096,
		Temperature: 0.6,
		Thinking: &Thinking{
			Type: "enabled",
		},
	}

	if len(tools) > 0 {
		req.Tools = tools
		// 设置为自动选择工具
		autoChoice, _ := json.Marshal("auto")
		req.ToolChoice = autoChoice
	}

	if stream {
		return c.chatStream(req)
	}
	return c.chatNonStream(req)
}

func (c *Client) chatNonStream(req ChatRequest) (*ChatResponse, error) {
	url := fmt.Sprintf("%s/chat/completions", baseURL)

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %w", err)
	}

	httpReq, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API请求失败 (状态码: %d): %s", resp.StatusCode, string(bodyBytes))
	}

	var chatResp ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	return &chatResp, nil
}

func (c *Client) chatStream(req ChatRequest) (*ChatResponse, error) {
	url := fmt.Sprintf("%s/chat/completions", baseURL)

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %w", err)
	}

	httpReq, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))
	httpReq.Header.Set("Accept", "text/event-stream")
	httpReq.Header.Set("Cache-Control", "no-cache")
	httpReq.Header.Set("Connection", "keep-alive")

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("API请求失败 (状态码: %d): %s", resp.StatusCode, string(bodyBytes))
	}

	var fullResponse ChatResponse
	var contentBuilder strings.Builder
	var reasoningBuilder strings.Builder

	reader := bufio.NewReader(resp.Body)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			resp.Body.Close()
			return nil, fmt.Errorf("reading stream response failed: %w", err)
		}

		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "data: ") {
			data := strings.TrimPrefix(line, "data: ")
			if data == "[DONE]" {
				break
			}

			var chunk StreamChunk
			if err := json.Unmarshal([]byte(data), &chunk); err != nil {
				continue
			}

			if len(chunk.Choices) > 0 && chunk.Choices[0].Delta != nil {
				delta := chunk.Choices[0].Delta
				if delta.Content != "" {
					contentBuilder.WriteString(delta.Content)
				}
				if delta.ReasoningContent != "" {
					reasoningBuilder.WriteString(delta.ReasoningContent)
				}
			}

			if fullResponse.ID == "" {
				fullResponse = ChatResponse{
					ID:      chunk.ID,
					Object:  chunk.Object,
					Created: chunk.Created,
					Model:   chunk.Model,
				}
			}
		}
	}
	resp.Body.Close()

	contentBytes, _ := json.Marshal(contentBuilder.String())
	fullResponse.Choices = []Choice{
		{
			Index: 0,
			Message: &Message{
				Role:    "assistant",
				Content: contentBytes,
			},
			FinishReason: "stop",
		},
	}

	return &fullResponse, nil
}

// StreamChat 执行流式聊天请求，支持工具调用
func (c *Client) StreamChat(messages []Message, tools []Tool, onChunk func(string, string, []ToolCall)) error {
	req := ChatRequest{
		Model:       "glm-4.5",
		Messages:    messages,
		Stream:      true,
		MaxTokens:   4096,
		Temperature: 0.6,
		Thinking: &Thinking{
			Type: "enabled",
		},
	}

	if len(tools) > 0 {
		req.Tools = tools
		// 设置为自动选择工具
		autoChoice, _ := json.Marshal("auto")
		req.ToolChoice = autoChoice
	}

	url := fmt.Sprintf("%s/chat/completions", baseURL)

	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("序列化请求失败: %w", err)
	}

	httpReq, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))
	httpReq.Header.Set("Accept", "text/event-stream")
	httpReq.Header.Set("Cache-Control", "no-cache")
	httpReq.Header.Set("Connection", "keep-alive")

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API请求失败 (状态码: %d): %s", resp.StatusCode, string(bodyBytes))
	}

	reader := bufio.NewReader(resp.Body)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("reading stream response failed: %w", err)
		}

		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "data: ") {
			data := strings.TrimPrefix(line, "data: ")
			if data == "[DONE]" {
				break
			}

			var chunk StreamChunk
			if err := json.Unmarshal([]byte(data), &chunk); err != nil {
				continue
			}

			if len(chunk.Choices) > 0 && chunk.Choices[0].Delta != nil {
				delta := chunk.Choices[0].Delta
				onChunk(delta.Content, delta.ReasoningContent, delta.ToolCalls)
			}
		}
	}

	return nil
}

// StreamChatWithChannel 执行流式聊天请求并返回通道
func (c *Client) StreamChatWithChannel(messages []Message, tools []Tool) (<-chan string, <-chan string, <-chan []ToolCall, <-chan error) {
	chunkCh := make(chan string)
	reasoningCh := make(chan string)
	toolCallCh := make(chan []ToolCall)
	errCh := make(chan error, 1)

	go func() {
		defer close(chunkCh)
		defer close(reasoningCh)
		defer close(toolCallCh)
		defer close(errCh)

		err := c.StreamChat(messages, tools, func(content, reasoning string, toolCalls []ToolCall) {
			if content != "" {
				chunkCh <- content
			}
			if reasoning != "" {
				reasoningCh <- reasoning
			}
			if len(toolCalls) > 0 {
				toolCallCh <- toolCalls
			}
		})

		if err != nil {
			errCh <- err
		} else {
			// 流正常结束时发送空字符串表示结束
			chunkCh <- ""
		}
	}()

	return chunkCh, reasoningCh, toolCallCh, errCh
}
