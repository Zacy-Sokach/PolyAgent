package utils

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"net/http"
	"time"
)

// RetryConfig 配置重试参数
type RetryConfig struct {
	// MaxRetries 最大重试次数
	MaxRetries int
	// InitialDelay 初始延迟时间
	InitialDelay time.Duration
	// MaxDelay 最大延迟时间
	MaxDelay time.Duration
	// BackoffMultiplier 退避倍数
	BackoffMultiplier float64
	// RetryableStatusCodes 需要重试的HTTP状态码
	RetryableStatusCodes []int
	// RetryableErrors 需要重试的错误类型判断函数
	RetryableErrors func(error) bool
}

// DefaultRetryConfig 返回默认的重试配置
func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
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
			// 默认重试所有网络错误
			return true
		},
	}
}

// RetryableHTTPClient 带重试机制的HTTP客户端
type RetryableHTTPClient struct {
	client *http.Client
	config *RetryConfig
}

// NewRetryableHTTPClient 创建新的带重试机制的HTTP客户端
func NewRetryableHTTPClient(client *http.Client, config *RetryConfig) *RetryableHTTPClient {
	if config == nil {
		config = DefaultRetryConfig()
	}
	return &RetryableHTTPClient{
		client: client,
		config: config,
	}
}

// Do 执行HTTP请求，支持重试
func (r *RetryableHTTPClient) Do(req *http.Request) (*http.Response, error) {
	var lastErr error
	var lastResp *http.Response

	for attempt := 0; attempt <= r.config.MaxRetries; attempt++ {
		// 检查上下文是否已取消
		if req.Context() != nil {
			select {
			case <-req.Context().Done():
				return nil, req.Context().Err()
			default:
			}
		}

		if attempt > 0 {
			// 计算延迟时间（指数退避）
			delay := r.calculateDelay(attempt)
			
			// 使用可取消的sleep，支持上下文取消
			if req.Context() != nil {
				timer := time.NewTimer(delay)
				select {
				case <-req.Context().Done():
					timer.Stop()
					return nil, req.Context().Err()
				case <-timer.C:
					timer.Stop()
				}
			} else {
				time.Sleep(delay)
			}
		}

		// 每次重试都需要克隆请求，因为请求体只能读取一次
		var clonedReq *http.Request
		if req.Body != nil && req.Body != http.NoBody {
			// 如果有请求体，需要重新创建请求
			clonedReq = r.cloneRequestWithBody(req)
		} else {
			// 如果没有请求体，可以直接克隆
			clonedReq = req.Clone(req.Context())
		}

		resp, err := r.client.Do(clonedReq)
		if err != nil {
			lastErr = err
			if !r.shouldRetryError(err) {
				break
			}
			continue
		}

		// 检查状态码
		if !r.shouldRetryStatus(resp.StatusCode) {
			return resp, nil
		}

		// 需要重试，关闭响应体
		if resp.Body != nil {
			resp.Body.Close()
		}
		lastResp = resp
		lastErr = fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	// 所有重试都失败了
	if lastErr != nil {
		return nil, fmt.Errorf("after %d retries: %w", r.config.MaxRetries, lastErr)
	}
	return lastResp, lastErr
}

// calculateDelay 计算延迟时间
func (r *RetryableHTTPClient) calculateDelay(attempt int) time.Duration {
	// 指数退避：delay = initialDelay * (backoffMultiplier ^ (attempt - 1))
	delay := float64(r.config.InitialDelay) * math.Pow(r.config.BackoffMultiplier, float64(attempt-1))
	
	// 限制最大延迟
	if delay > float64(r.config.MaxDelay) {
		delay = float64(r.config.MaxDelay)
	}
	
	return time.Duration(delay)
}

// shouldRetryStatus 判断是否应该重试某个状态码
func (r *RetryableHTTPClient) shouldRetryStatus(statusCode int) bool {
	for _, code := range r.config.RetryableStatusCodes {
		if statusCode == code {
			return true
		}
	}
	return false
}

// shouldRetryError 判断是否应该重试某个错误
func (r *RetryableHTTPClient) shouldRetryError(err error) bool {
	if r.config.RetryableErrors == nil {
		return false
	}
	return r.config.RetryableErrors(err)
}

// cloneRequestWithBody 克隆带请求体的HTTP请求
func (r *RetryableHTTPClient) cloneRequestWithBody(req *http.Request) *http.Request {
	// 读取请求体
	bodyBytes, err := io.ReadAll(req.Body)
	if err != nil {
		return req.Clone(req.Context())
	}
	
	// 重置原始请求体
	req.Body.Close()
	req.Body = io.NopCloser(bytes.NewReader(bodyBytes))
	
	// 创建克隆请求
	clonedReq := req.Clone(req.Context())
	// 为克隆请求设置新的请求体
	clonedReq.Body = io.NopCloser(bytes.NewReader(bodyBytes))
	
	return clonedReq
}

// WithRetry 为函数添加重试机制
func WithRetry(fn func() error, config *RetryConfig) error {
	if config == nil {
		config = DefaultRetryConfig()
	}

	var lastErr error
	for attempt := 0; attempt <= config.MaxRetries; attempt++ {
		if attempt > 0 {
			delay := time.Duration(float64(config.InitialDelay) * math.Pow(config.BackoffMultiplier, float64(attempt-1)))
			if delay > config.MaxDelay {
				delay = config.MaxDelay
			}
			time.Sleep(delay)
		}

		err := fn()
		if err == nil {
			return nil
		}

		lastErr = err
		if config.RetryableErrors != nil && !config.RetryableErrors(err) {
			break
		}
	}

	return fmt.Errorf("after %d retries: %w", config.MaxRetries, lastErr)
}