package utils

import "net/http"

// Doer 接口，支持http.Client和RetryableHTTPClient
type Doer interface {
	Do(*http.Request) (*http.Response, error)
}