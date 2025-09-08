package http

import (
	"context"
	"io"
	"net/http"
)

// HTTPClientInterface defines the interface for HTTP operations
type HTTPClientInterface interface {
	DoWithRetry(ctx context.Context, req *http.Request) (*http.Response, error)
	Get(ctx context.Context, url string) (*http.Response, error)
	Post(ctx context.Context, url, contentType string, body io.Reader) (*http.Response, error)
	Put(ctx context.Context, url, contentType string, body io.Reader) (*http.Response, error)
	Delete(ctx context.Context, url string) (*http.Response, error)
	GetClient() *http.Client
	Close()
}
