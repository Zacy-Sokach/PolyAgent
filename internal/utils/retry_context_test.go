package utils

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRetryableHTTPClient_ContextCancellation(t *testing.T) {
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
		MaxRetries:         10,
		InitialDelay:       100 * time.Millisecond,
		MaxDelay:           1 * time.Second,
		BackoffMultiplier:  2.0,
		RetryableStatusCodes: []int{http.StatusInternalServerError},
	}
	retryClient := NewRetryableHTTPClient(baseClient, config)

	// 创建一个会在50ms后取消的上下文
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", server.URL, nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	start := time.Now()
	resp, err := retryClient.Do(req)
	elapsed := time.Since(start)

	// 应该返回上下文取消错误
	if err == nil {
		t.Fatal("Expected context cancellation error")
	}
	if resp != nil {
		resp.Body.Close()
	}

	if err != context.DeadlineExceeded {
		t.Errorf("Expected context.DeadlineExceeded, got %v", err)
	}

	// 验证请求次数（由于延迟，应该只有1次请求）
	if requestCount != 1 {
		t.Errorf("Expected 1 request before context cancellation, got %d", requestCount)
	}

	// 验证总时间（应该小于100ms，因为上下文在50ms后取消）
	if elapsed > 100*time.Millisecond {
		t.Errorf("Expected request to complete within 100ms due to context cancellation, took %v", elapsed)
	}
}

func TestRetryableHTTPClient_ContextCancellationDuringDelay(t *testing.T) {
	// 记录请求次数
	requestCount := 0
	
	// 创建一个测试服务器，第一次返回500，第二次返回200
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		if requestCount == 1 {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("internal server error"))
		} else {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("success"))
		}
	}))
	defer server.Close()

	baseClient := &http.Client{Timeout: 5 * time.Second}
	config := &RetryConfig{
		MaxRetries:         3,
		InitialDelay:       200 * time.Millisecond, // 较长的延迟
		MaxDelay:           1 * time.Second,
		BackoffMultiplier:  2.0,
		RetryableStatusCodes: []int{http.StatusInternalServerError},
	}
	retryClient := NewRetryableHTTPClient(baseClient, config)

	// 创建一个会在100ms后取消的上下文（在第一次重试延迟期间取消）
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", server.URL, nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	start := time.Now()
	resp, err := retryClient.Do(req)
	elapsed := time.Since(start)

	// 应该返回上下文取消错误
	if err == nil {
		t.Fatal("Expected context cancellation error")
	}
	if resp != nil {
		resp.Body.Close()
	}

	if err != context.DeadlineExceeded {
		t.Errorf("Expected context.DeadlineExceeded, got %v", err)
	}

	// 验证请求次数（应该只有1次请求，因为第二次请求在延迟期间被取消）
	if requestCount != 1 {
		t.Errorf("Expected 1 request before context cancellation, got %d", requestCount)
	}

	// 验证总时间（应该大约100ms，上下文取消的时间）
	if elapsed > 150*time.Millisecond {
		t.Errorf("Expected request to complete within 150ms due to context cancellation, took %v", elapsed)
	}
}