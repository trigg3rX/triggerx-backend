package utils

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestFetchDataFromUrl_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello world"))
	}))
	defer server.Close()

	data, err := FetchDataFromUrl(server.URL)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if data != "hello world" {
		t.Errorf("expected 'hello world', got %q", data)
	}
}

func TestFetchDataFromUrl_Non200(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("not found"))
	}))
	defer server.Close()

	data, err := FetchDataFromUrl(server.URL)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !strings.Contains(data, "not found") {
		t.Errorf("expected 'not found' in response, got %q", data)
	}
}

func TestFetchDataFromUrl_InvalidURL(t *testing.T) {
	_, err := FetchDataFromUrl(":badurl:")
	if err == nil {
		t.Error("expected error for invalid URL, got nil")
	}
}

func TestFetchDataFromUrl_EmptyURL(t *testing.T) {
	_, err := FetchDataFromUrl("")
	if err == nil {
		t.Error("expected error for empty URL, got nil")
	}
}
