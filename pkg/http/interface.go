package http

import (
	"io"
	"net/http"
)

// HTTPClientInterface defines the interface for HTTP operations
type HTTPClientInterface interface {
	DoWithRetry(req *http.Request) (*http.Response, error)
	Get(url string) (*http.Response, error)
	Post(url, contentType string, body io.Reader) (*http.Response, error)
	Put(url, contentType string, body io.Reader) (*http.Response, error)
	Delete(url string) (*http.Response, error)
	GetClient() *http.Client
	Close()
}
