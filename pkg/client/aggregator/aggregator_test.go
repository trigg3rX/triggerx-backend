package aggregator

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

// generateTestPrivateKey is a helper function to create a valid private key for testing.
func generateTestPrivateKey(t *testing.T) (*ecdsa.PrivateKey, string) {
	privateKey, err := crypto.GenerateKey()
	require.NoError(t, err, "should be able to generate a private key")
	privateKeyBytes := crypto.FromECDSA(privateKey)
	return privateKey, fmt.Sprintf("%x", privateKeyBytes)
}

// TestNewAggregatorClient covers all scenarios for the client constructor.
func TestNewAggregatorClient(t *testing.T) {
	_, privateKeyHex := generateTestPrivateKey(t)
	logger := logging.NewNoOpLogger()

	t.Run("Success", func(t *testing.T) {
		cfg := AggregatorClientConfig{
			AggregatorRPCUrl: "http://localhost:8545",
			SenderPrivateKey: privateKeyHex,
		}

		client, err := NewAggregatorClient(logger, cfg)
		require.NoError(t, err)
		require.NotNil(t, client)

		assert.Equal(t, logger, client.logger)
		assert.Equal(t, cfg, client.config)
		assert.NotNil(t, client.privateKey)
		assert.NotNil(t, client.publicKey)
		assert.NotNil(t, client.httpClient)

		// Ensure the Close method can be called without panicking.
		client.Close()
	})

	t.Run("Failure: Nil Logger", func(t *testing.T) {
		cfg := AggregatorClientConfig{
			AggregatorRPCUrl: "http://localhost:8545",
			SenderPrivateKey: privateKeyHex,
		}
		client, err := NewAggregatorClient(nil, cfg)
		assert.Error(t, err)
		assert.Nil(t, client)
		assert.EqualError(t, err, "logger cannot be nil")
	})

	t.Run("Failure: Empty RPC Address", func(t *testing.T) {
		cfg := AggregatorClientConfig{
			AggregatorRPCUrl: "", // Invalid RPC URL
			SenderPrivateKey: privateKeyHex,
		}
		client, err := NewAggregatorClient(logger, cfg)
		assert.Error(t, err)
		assert.Nil(t, client)
		assert.EqualError(t, err, "RPC address cannot be empty")
	})

	t.Run("Failure: Empty Private Key", func(t *testing.T) {
		cfg := AggregatorClientConfig{
			AggregatorRPCUrl: "http://localhost:8545",
			SenderPrivateKey: "", // Invalid private key
		}
		client, err := NewAggregatorClient(logger, cfg)
		assert.Error(t, err)
		assert.Nil(t, client)
		assert.EqualError(t, err, "sender private key cannot be empty")
	})

	t.Run("Failure: Invalid Private Key Format", func(t *testing.T) {
		cfg := AggregatorClientConfig{
			AggregatorRPCUrl: "http://localhost:8545",
			SenderPrivateKey: "this-is-not-a-valid-hex-key",
		}
		client, err := NewAggregatorClient(logger, cfg)
		require.Error(t, err)
		assert.Nil(t, client)
		assert.ErrorIs(t, err, ErrInvalidKey, "error should wrap ErrInvalidKey")
		assert.Contains(t, err.Error(), "failed to convert private key", "error message should be descriptive")
	})

	t.Run("Failure: Private Key Too Short", func(t *testing.T) {
		cfg := AggregatorClientConfig{
			AggregatorRPCUrl: "http://localhost:8545",
			SenderPrivateKey: "0x123", // Too short for a valid private key
		}
		client, err := NewAggregatorClient(logger, cfg)
		require.Error(t, err)
		assert.Nil(t, client)
		assert.ErrorIs(t, err, ErrInvalidKey, "error should wrap ErrInvalidKey")
		assert.Contains(t, err.Error(), "failed to convert private key", "error message should be descriptive")
	})

	t.Run("Failure: Private Key Too Long", func(t *testing.T) {
		cfg := AggregatorClientConfig{
			AggregatorRPCUrl: "http://localhost:8545",
			SenderPrivateKey: "0x" + strings.Repeat("a", 100), // Too long for a valid private key
		}
		client, err := NewAggregatorClient(logger, cfg)
		require.Error(t, err)
		assert.Nil(t, client)
		assert.ErrorIs(t, err, ErrInvalidKey, "error should wrap ErrInvalidKey")
		assert.Contains(t, err.Error(), "failed to convert private key", "error message should be descriptive")
	})

	t.Run("Failure: Whitespace in Private Key", func(t *testing.T) {
		cfg := AggregatorClientConfig{
			AggregatorRPCUrl: "http://localhost:8545",
			SenderPrivateKey: " " + privateKeyHex + " ", // Private key with whitespace
		}
		client, err := NewAggregatorClient(logger, cfg)
		require.Error(t, err)
		assert.Nil(t, client)
		assert.ErrorIs(t, err, ErrInvalidKey, "error should wrap ErrInvalidKey")
		assert.Contains(t, err.Error(), "failed to convert private key", "error message should be descriptive")
	})
}

// TestAggregatorClient_executeWithRetry tests the core RPC execution logic.
func TestAggregatorClient_executeWithRetry(t *testing.T) {
	logger := logging.NewNoOpLogger()
	_, privateKeyHex := generateTestPrivateKey(t)

	t.Run("Success: RPC Call Succeeds on First Try", func(t *testing.T) {
		// Mock a JSON-RPC server that returns a successful response.
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			// A valid JSON-RPC response is required for the client to succeed.
			_, err := fmt.Fprintln(w, `{"jsonrpc":"2.0","id":1,"result":"0xSuccess"}`)
			require.NoError(t, err)
		}))
		defer server.Close()

		cfg := AggregatorClientConfig{
			AggregatorRPCUrl: server.URL,
			SenderPrivateKey: privateKeyHex,
		}
		client, err := NewAggregatorClient(logger, cfg)
		require.NoError(t, err)

		var result string
		params := CallParams{Data: "0x123", TaskDefinitionID: 1}
		err = client.executeWithRetry(context.Background(), "sendCustomMessage", &result, params)

		require.NoError(t, err, "executeWithRetry should not return an error on success")
	})

	t.Run("Failure: RPC Call Returns Error", func(t *testing.T) {
		// Mock a server that returns a JSON-RPC error.
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK) // The HTTP request itself is ok.
			_, err := fmt.Fprintln(w, `{"jsonrpc":"2.0","id":1,"error":{"code":-32000,"message":"execution reverted"}}`)
			require.NoError(t, err)
		}))
		defer server.Close()

		cfg := AggregatorClientConfig{
			AggregatorRPCUrl: server.URL,
			SenderPrivateKey: privateKeyHex,
		}
		client, err := NewAggregatorClient(logger, cfg)
		require.NoError(t, err)
		client.httpClient.HTTPConfig.RetryConfig.MaxRetries = 2
		client.httpClient.HTTPConfig.RetryConfig.InitialDelay = 10 * time.Millisecond

		var result interface{}
		err = client.executeWithRetry(context.Background(), "sendTask", &result, CallParams{})

		require.Error(t, err)
		assert.Contains(t, err.Error(), "RPC call failed", "error message should indicate an RPC call failure")
		assert.Contains(t, err.Error(), "execution reverted", "error should contain the message from the server")
	})

	t.Run("Failure: Retry Exhausted", func(t *testing.T) {
		// Mock a server that always returns an error
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		cfg := AggregatorClientConfig{
			AggregatorRPCUrl: server.URL,
			SenderPrivateKey: privateKeyHex,
		}
		client, err := NewAggregatorClient(logger, cfg)
		require.NoError(t, err)
		client.httpClient.HTTPConfig.RetryConfig.MaxRetries = 3
		client.httpClient.HTTPConfig.RetryConfig.InitialDelay = 10 * time.Millisecond

		var result interface{}
		err = client.executeWithRetry(context.Background(), "sendTask", &result, CallParams{})

		require.Error(t, err)
		assert.Contains(t, err.Error(), "RPC call failed", "error message should indicate an RPC call failure")
	})

	t.Run("Success: Retry Eventually Succeeds", func(t *testing.T) {
		attemptCount := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			attemptCount++
			if attemptCount < 3 {
				// Fail first two attempts
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			// Succeed on third attempt
			w.WriteHeader(http.StatusOK)
			_, err := fmt.Fprintln(w, `{"jsonrpc":"2.0","id":1,"result":"0xSuccess"}`)
			require.NoError(t, err)
		}))
		defer server.Close()

		cfg := AggregatorClientConfig{
			AggregatorRPCUrl: server.URL,
			SenderPrivateKey: privateKeyHex,
		}
		client, err := NewAggregatorClient(logger, cfg)
		require.NoError(t, err)
		client.httpClient.HTTPConfig.RetryConfig.MaxRetries = 5
		client.httpClient.HTTPConfig.RetryConfig.InitialDelay = 10 * time.Millisecond

		var result string
		params := CallParams{Data: "0x123", TaskDefinitionID: 1}
		err = client.executeWithRetry(context.Background(), "sendCustomMessage", &result, params)

		require.NoError(t, err, "executeWithRetry should succeed after retries")
		assert.Equal(t, "0xSuccess", result, "should get the expected result")
		assert.Equal(t, 3, attemptCount, "should have made exactly 3 attempts")
	})

	t.Run("Failure: DNS Resolution Error", func(t *testing.T) {
		cfg := AggregatorClientConfig{
			AggregatorRPCUrl: "http://nonexistent-domain-that-will-never-exist-12345.com:8545",
			SenderPrivateKey: privateKeyHex,
		}
		client, err := NewAggregatorClient(logger, cfg)
		require.NoError(t, err)
		client.httpClient.HTTPConfig.RetryConfig.MaxRetries = 2
		client.httpClient.HTTPConfig.RetryConfig.InitialDelay = 10 * time.Millisecond

		var result interface{}
		err = client.executeWithRetry(context.Background(), "sendTask", &result, CallParams{})

		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to dial aggregator RPC", "error message should indicate a dialing failure")
	})
}
