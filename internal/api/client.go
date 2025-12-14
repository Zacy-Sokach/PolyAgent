package api

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

const (
	baseURL = "https://open.bigmodel.cn/api/paas/v4"
)

// APIError 表示 API 请求错误，包含状态码和错误信息
type APIError struct {
	StatusCode int
	Message    string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("API请求失败 (状态码: %d): %s", e.StatusCode, e.Message)
}

// 全局共享的HTTP客户端，实现连接池化
var (
	sharedHTTPClient *http.Client
	httpClientOnce   sync.Once
)

// getSharedHTTPClient 返回共享的HTTP客户端实例
func getSharedHTTPClient() *http.Client {
	httpClientOnce.Do(func() {
		sharedHTTPClient = &http.Client{
			// 设置合理的超时，避免无限等待
			Timeout: 60 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 20,        // 优化：减少到20，避免资源浪费
				IdleConnTimeout:     90 * time.Second,
				DisableCompression:  false,      // 启用压缩，减少传输数据量
				MaxConnsPerHost:     50,         // 优化：减少到50，平衡性能和资源使用
				// 新增：响应头超时和TLS握手超时
				ResponseHeaderTimeout: 30 * time.Second,
				TLSHandshakeTimeout:   10 * time.Second,
			},
		}
	})
	return sharedHTTPClient
}

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
		client: getSharedHTTPClient(),
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
	// 预分配容量，减少内存重分配
	var contentBuilder strings.Builder
	contentBuilder.Grow(4096) // 预估响应大小
	var reasoningBuilder strings.Builder
	reasoningBuilder.Grow(2048) // 预估推理内容大小

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

	// 优化：直接使用字符串而非 JSON 编码，减少转换开销
	contentStr := contentBuilder.String()
	fullResponse.Choices = []Choice{
		{
			Index: 0,
			Message: &Message{
				Role:    "assistant",
				Content: []byte(contentStr),
			},
			FinishReason: "stop",
		},
	}

	return &fullResponse, nil
}

// StreamChat 执行流式聊天请求，支持工具调用
func (c *Client) StreamChat(messages []Message, tools []Tool, onChunk func(string, string, []ToolCall)) error {
	return c.StreamChatWithCoT(messages, tools, onChunk, true)
}

// StreamChatWithCoT 执行流式聊天请求，支持工具调用和CoT配置
func (c *Client) StreamChatWithCoT(messages []Message, tools []Tool, onChunk func(string, string, []ToolCall), enableCoT bool) error {
	const maxRetries = 3

	retryDelays := []time.Duration{5 * time.Second, 15 * time.Second, 30 * time.Second}
	var lastErr error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		if attempt > 1 {
			time.Sleep(retryDelays[attempt-2])
		}

		err := c.streamChat(messages, tools, onChunk, enableCoT)
		if err == nil {
			return nil
		}

		lastErr = err
		
		// 对所有 API 错误（非200状态码）重试
		if apiErr, ok := err.(*APIError); ok && apiErr.StatusCode != 200 {
			if attempt < maxRetries {
				continue
			}
		}
		
		// 其他错误立即返回
		return err
	}

	return fmt.Errorf("stream chat failed after %d attempts: %w", maxRetries, lastErr)
}

func (c *Client) streamChat(messages []Message, tools []Tool, onChunk func(string, string, []ToolCall), enableCoT bool) error {
	req := ChatRequest{
		Model:       "glm-4.5",
		Messages:    messages,
		Stream:      true,
		MaxTokens:   4096,
		Temperature: 0.6,
	}
	
	// 根据CoT配置决定是否启用thinking
	if enableCoT {
		req.Thinking = &Thinking{
			Type: "enabled",
		}
	}

	if len(tools) > 0 {
		req.Tools = tools
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
		return &APIError{
			StatusCode: resp.StatusCode,
			Message:    string(bodyBytes),
		}
	}

	reader := bufio.NewReader(resp.Body)
	
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			// 如果是超时错误，返回特定错误以便上层重试
			if err.Error() == "read timeout" {
				return fmt.Errorf("reading stream response failed: context deadline exceeded")
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
func (c *Client) StreamChatWithChannel(ctx context.Context, messages []Message, tools []Tool) (<-chan string, <-chan string, <-chan []ToolCall, <-chan error) {
	return c.StreamChatWithChannelAndCoT(ctx, messages, tools, true)
}

// StreamChatWithChannelAndCoT 执行流式聊天请求并返回通道，支持CoT配置
func (c *Client) StreamChatWithChannelAndCoT(ctx context.Context, messages []Message, tools []Tool, enableCoT bool) (<-chan string, <-chan string, <-chan []ToolCall, <-chan error) {
	chunkCh := make(chan string, 10)
	reasoningCh := make(chan string, 10)
	toolCallCh := make(chan []ToolCall, 5)
	errCh := make(chan error, 1)

	go func() {
		defer func() {
			close(chunkCh)
			close(reasoningCh)
			close(toolCallCh)
			close(errCh)
		}()

		streamCtx, cancel := context.WithCancel(ctx)
		defer cancel()

		done := make(chan struct{})
		go func() {
			<-streamCtx.Done()
			close(done)
		}()

		err := c.StreamChatWithCoT(messages, tools, func(content, reasoning string, toolCalls []ToolCall) {
			select {
			case <-done:
				return
			default:
				if content != "" {
					select {
					case chunkCh <- content:
					case <-time.After(100 * time.Millisecond):
					}
				}
				if reasoning != "" {
					select {
					case reasoningCh <- reasoning:
					case <-time.After(100 * time.Millisecond):
					}
				}
				if len(toolCalls) > 0 {
					select {
					case toolCallCh <- toolCalls:
					case <-time.After(100 * time.Millisecond):
					}
				}
			}
		}, enableCoT)

		if err != nil {
			select {
			case errCh <- err:
			case <-done:
			}
		} else {
			select {
			case chunkCh <- "":
			case <-done:
			}
		}
	}()

	return chunkCh, reasoningCh, toolCallCh, errCh
}
