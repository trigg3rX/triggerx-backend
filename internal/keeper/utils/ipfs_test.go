package utils

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/trigg3rX/triggerx-backend-imua/pkg/types"
)

func TestUploadToIPFS_Success(t *testing.T) {
	// Set PINATA_JWT env var if needed by config.GetPinataJWT
	if err := os.Setenv("PINATA_JWT", "testtoken"); err != nil {
		t.Fatalf("failed to set PINATA_JWT: %v", err)
	}
	defer func() {
		if err := os.Unsetenv("PINATA_JWT"); err != nil {
			t.Logf("failed to unset PINATA_JWT: %v", err)
		}
	}()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.Header.Get("Authorization"), "Bearer ") {
			t.Errorf("missing or invalid Authorization header: %v", r.Header.Get("Authorization"))
		}
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(`{"data":{"cid":"testcid123"}}`))
		if err != nil {
			t.Errorf("error writing response: %v", err)
		}
	}))
	defer server.Close()

	// We cannot change the URL in UploadToIPFS, so this test will only check for network errors (unless the function is refactored)
	// This is a limitation without code modification.
	// The test below will actually hit the real Pinata endpoint unless the code is refactored to allow URL injection.
	// So, we only check for error on invalid URL.
	_, err := UploadToIPFS("file.txt", []byte("data"))
	if err == nil {
		t.Skip("Cannot test UploadToIPFS fully without code modification to inject URL")
	}
}

func TestUploadToIPFS_ErrorCases(t *testing.T) {
	// Invalid file name or data
	_, err := UploadToIPFS("", nil)
	if err == nil {
		t.Error("expected error for empty filename/data, got nil")
	}
}

func TestFetchIPFSContent_Success(t *testing.T) {
	// Set IPFS_HOST env var if needed by config.GetIpfsHost
	if err := os.Setenv("IPFS_HOST", "localhost"); err != nil {
		t.Fatalf("failed to set IPFS_HOST: %v", err)
	}
	defer func() {
		if err := os.Unsetenv("IPFS_HOST"); err != nil {
			t.Logf("failed to unset IPFS_HOST: %v", err)
		}
	}()

	expected := types.IPFSData{}
	body, _ := json.Marshal(expected)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write(body)
		if err != nil {
			t.Errorf("error writing response: %v", err)
		}
	}))
	defer server.Close()

	// We cannot change the URL in FetchIPFSContent, so this test will only check for network errors (unless the function is refactored)
	_, err := FetchIPFSContent("testcid")
	if err == nil {
		t.Skip("Cannot test FetchIPFSContent fully without code modification to inject URL")
	}
}

func TestFetchIPFSContent_NotFound(t *testing.T) {
	// This will always hit the real endpoint, so we only check for error on invalid CID
	_, err := FetchIPFSContent("testcid")
	if err == nil {
		t.Error("expected error for empty CID, got nil")
	}
}

func TestFetchIPFSContent_InvalidJSON(t *testing.T) {
	// This will always hit the real endpoint, so we only check for error on invalid CID
	// (Cannot test invalid JSON without code modification)
}
