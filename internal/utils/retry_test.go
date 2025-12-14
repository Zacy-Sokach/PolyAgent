package utils

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRetryableHTTPClient_Success(t *testing.T) {
	// 创建一个测试服务器，总是返回200
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	}))
	defer server.Close()

	baseClient := &http.Client{Timeout: 5 * time.Second}
	retryClient := NewRetryableHTTPClient(baseClient, DefaultRetryConfig())

	req, err := http.NewRequest("GET", server.URL, nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	resp, err := retryClient.Do(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestRetryableHTTPClient_RetryOn500(t *testing.T) {
	// 记录请求次数
	requestCount := 0
	
	// 创建一个测试服务器，前2次返回500，第3次返回200
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		if requestCount <= 2 {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("internal server error"))
		} else {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("success after retry"))
		}
	}))
	defer server.Close()

	baseClient := &http.Client{Timeout: 5 * time.Second}
	config := &RetryConfig{
		MaxRetries:         3,
		InitialDelay:       10 * time.Millisecond, // 使用短延迟加速测试
		MaxDelay:           100 * time.Millisecond,
		BackoffMultiplier:  2.0,
		RetryableStatusCodes: []int{http.StatusInternalServerError},
	}
	retryClient := NewRetryableHTTPClient(baseClient, config)

	req, err := http.NewRequest("GET", server.URL, nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	start := time.Now()
	resp, err := retryClient.Do(req)
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	if requestCount != 3 {
		t.Errorf("Expected 3 requests, got %d", requestCount)
	}

	// 验证延迟时间（至少应该有10ms + 20ms的延迟）
	if elapsed < 25*time.Millisecond {
		t.Errorf("Expected at least 25ms delay due to retries, got %v", elapsed)
	}
}

func TestRetryableHTTPClient_FailAfterMaxRetries(t *testing.T) {
	// 记录请求次数
	requestCount := 0
	
	// 创建一个测试服务器，总是返回500
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal server error"))
	}))
	defer server.Close()

	baseClient := &http.Client{Timeout: 5 * time.Second}
	config := &RetryConfig{
		MaxRetries:         2,
		InitialDelay:       10 * time.Millisecond,
		MaxDelay:           100 * time.Millisecond,
		BackoffMultiplier:  2.0,
		RetryableStatusCodes: []int{http.StatusInternalServerError},
	}
	retryClient := NewRetryableHTTPClient(baseClient, config)

	req, err := http.NewRequest("GET", server.URL, nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	resp, err := retryClient.Do(req)
	
	// 应该返回错误
	if err == nil {
		t.Fatal("Expected error after max retries")
	}
	if resp != nil {
		resp.Body.Close()
	}

	// 应该尝试了3次（1次初始请求 + 2次重试）
	if requestCount != 3 {
		t.Errorf("Expected 3 requests (1 initial + 2 retries), got %d", requestCount)
	}

	// 验证错误消息
	expectedErrMsg := "after 2 retries"
	if err.Error()[:len(expectedErrMsg)] != expectedErrMsg {
		t.Errorf("Expected error message to start with %q, got %q", expectedErrMsg, err.Error())
	}
}

func TestWithRetry_Function(t *testing.T) {
	// 记录调用次数
	callCount := 0
	
	err := WithRetry(func() error {
		callCount++
		if callCount <= 2 {
			return fmt.Errorf("temporary error")
		}
		return nil
	}, &RetryConfig{
		MaxRetries:         3,
		InitialDelay:       10 * time.Millisecond,
		MaxDelay:           100 * time.Millisecond,
		BackoffMultiplier:  2.0,
	})

	if err != nil {
		t.Fatalf("Expected success after retries, got error: %v", err)
	}

	if callCount != 3 {
		t.Errorf("Expected 3 calls, got %d", callCount)
	}
}

func TestWithRetry_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	callCount := 0
	
	err := WithRetry(func() error {
		callCount++
		// 检查context是否已取消
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		
		if callCount <= 5 {
			return fmt.Errorf("temporary error")
		}
		return nil
	}, &RetryConfig{
		MaxRetries:         10,
		InitialDelay:       20 * time.Millisecond,
		MaxDelay:           100 * time.Millisecond,
		BackoffMultiplier:  2.0,
	})

	// 应该因为context取消而失败
	if err == nil {
		t.Fatal("Expected context cancellation error")
	}

	// 检查错误是否包含context deadline exceeded
	if err.Error() != "after 10 retries: context deadline exceeded" {
		t.Errorf("Expected 'after 10 retries: context deadline exceeded', got %q", err.Error())
	}
}