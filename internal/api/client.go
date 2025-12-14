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
	
	"github.com/Zacy-Sokach/PolyAgent/internal/utils"
)

const (
	baseURL = "https://open.bigmodel.cn/api/paas/v4"
)

// 全局共享的HTTP客户端，实现连接池化
var (
	sharedHTTPClient utils.Doer
	httpClientOnce   sync.Once
)

// getSharedHTTPClient 返回共享的HTTP客户端实例
func getSharedHTTPClient() utils.Doer {
	httpClientOnce.Do(func() {
		baseClient := &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 50,        // 从10增加到50，提高并发性能
				IdleConnTimeout:     90 * time.Second,
				DisableCompression:  false,      // 启用压缩，减少传输数据量
				MaxConnsPerHost:     100,        // 新增：限制每个主机的最大连接数
			},
		}
		// 包装为带重试机制的客户端
		retryConfig := &utils.RetryConfig{
			MaxRetries:         3,
			InitialDelay:       1 * time.Second,
			MaxDelay:           30 * time.Second,
			BackoffMultiplier:  2.0,
			RetryableStatusCodes: []int{
				http.StatusRequestTimeout,      // 408
				http.StatusTooManyRequests,     // 429
				http.StatusInternalServerError, // 500
				http.StatusBadGateway,          // 502
				http.StatusServiceUnavailable,  // 503
				http.StatusGatewayTimeout,      // 504
			},
			RetryableErrors: func(err error) bool {
				// 重试网络错误和超时
				return true
			},
		}
		sharedHTTPClient = utils.NewRetryableHTTPClient(baseClient, retryConfig)
	})
	return sharedHTTPClient
}

type Client struct {
	apiKey string
	client utils.Doer
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

	// 只在最后需要时调用String()，避免中间转换
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
func (c *Client) StreamChatWithChannel(ctx context.Context, messages []Message, tools []Tool) (<-chan string, <-chan string, <-chan []ToolCall, <-chan error) {
	chunkCh := make(chan string, 10)  // 添加缓冲区，提高吞吐量
	reasoningCh := make(chan string, 10)
	toolCallCh := make(chan []ToolCall, 5)
	errCh := make(chan error, 1)

	go func() {
		// 确保所有channel在goroutine退出时关闭
		defer func() {
			close(chunkCh)
			close(reasoningCh)
			close(toolCallCh)
			close(errCh)
		}()

		// 创建可取消的子context，关联到StreamChat调用
		streamCtx, cancel := context.WithCancel(ctx)
		defer cancel()

		// 使用channel监听context取消信号
		done := make(chan struct{})
		go func() {
			<-streamCtx.Done()
			close(done)
		}()

		// 执行流式请求
		err := c.StreamChat(messages, tools, func(content, reasoning string, toolCalls []ToolCall) {
			select {
			case <-done:
				// context已取消，停止发送
				return
			default:
				// 发送数据到channel，带超时避免阻塞
				if content != "" {
					select {
					case chunkCh <- content:
					case <-time.After(100 * time.Millisecond):
						// 发送超时，跳过
					}
				}
				if reasoning != "" {
					select {
					case reasoningCh <- reasoning:
					case <-time.After(100 * time.Millisecond):
						// 发送超时，跳过
					}
				}
				if len(toolCalls) > 0 {
					select {
					case toolCallCh <- toolCalls:
					case <-time.After(100 * time.Millisecond):
						// 发送超时，跳过
					}
				}
			}
		})

		if err != nil {
			select {
			case errCh <- err:
			case <-done:
				// context已取消
			}
		} else {
			// 流正常结束时发送空字符串表示结束
			select {
			case chunkCh <- "":
			case <-done:
				// context已取消
			}
		}
	}()

	return chunkCh, reasoningCh, toolCallCh, errCh
}
