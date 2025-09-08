package aggregator

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// jsonRPCRequest is a helper struct to unmarshal JSON-RPC requests for validation.
type jsonRPCRequest struct {
	Method string        `json:"method"`
	Params []interface{} `json:"params"`
	ID     int           `json:"id"`
}

// TestSendTaskToPerformerWithMock validates the logic using mocks (unit tests)
func TestSendTaskToPerformerWithMock(t *testing.T) {
	taskData := &types.BroadcastDataForPerformer{
		TaskID:           123,
		PerformerAddress: "0x1234567890123456789012345678901234567890",
		Data:             []byte("some custom task data"),
		TaskDefinitionID: 42,
	}

	t.Run("Success", func(t *testing.T) {
		// Use the mock builder to set up expectations
		client := NewMockAggregatorClientBuilder().
			ExpectSendTaskToPerformer(context.Background(), taskData, true, nil).
			Build()

		success, err := client.SendTaskToPerformer(context.Background(), taskData)
		assert.NoError(t, err)
		assert.True(t, success)

		// Assert expectations were met
		client.AssertExpectations(t)
	})

	t.Run("Failure: Network Error", func(t *testing.T) {
		expectedErr := fmt.Errorf("network error")
		client := NewMockAggregatorClientBuilder().
			ExpectSendTaskToPerformer(context.Background(), taskData, false, expectedErr).
			Build()

		success, err := client.SendTaskToPerformer(context.Background(), taskData)
		assert.Error(t, err)
		assert.False(t, success)
		assert.Equal(t, expectedErr, err)

		client.AssertExpectations(t)
	})

	t.Run("Failure: RPC Error", func(t *testing.T) {
		expectedErr := fmt.Errorf("RPC call failed")
		client := NewMockAggregatorClientBuilder().
			ExpectSendTaskToPerformer(context.Background(), taskData, false, expectedErr).
			Build()

		success, err := client.SendTaskToPerformer(context.Background(), taskData)
		assert.Error(t, err)
		assert.False(t, success)
		assert.Equal(t, expectedErr, err)

		client.AssertExpectations(t)
	})
}

// TestSendTaskToPerformerWithHTTP validates the logic using real HTTP servers (integration tests)
func TestSendTaskToPerformerWithHTTP(t *testing.T) {
	logger := logging.NewNoOpLogger()
	_, privateKeyHex := generateTestPrivateKey(t)

	taskData := &types.BroadcastDataForPerformer{
		TaskID:           123,
		PerformerAddress: "0x1234567890123456789012345678901234567890",
		Data:             []byte("some custom task data"),
		TaskDefinitionID: 42,
	}

	t.Run("Success", func(t *testing.T) {
		// Mock a JSON-RPC server that validates the incoming request.
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			bodyBytes, err := io.ReadAll(r.Body)
			require.NoError(t, err)

			var req jsonRPCRequest
			err = json.Unmarshal(bodyBytes, &req)
			require.NoError(t, err)

			// 1. Validate the method name.
			assert.Equal(t, "sendCustomMessage", req.Method)
			require.Len(t, req.Params, 2)

			// 2. Validate the parameters in the correct order: Data, TaskDefinitionID.
			expectedData := "0x" + hex.EncodeToString(taskData.Data)
			assert.Equal(t, expectedData, req.Params[0])
			// JSON unmarshals numbers into float64 by default, so we cast for comparison.
			assert.Equal(t, float64(taskData.TaskDefinitionID), req.Params[1])

			// 3. Send a successful response.
			w.WriteHeader(http.StatusOK)
			_, err = fmt.Fprintln(w, `{"jsonrpc":"2.0","id":1,"result":true}`)
			require.NoError(t, err)
		}))
		defer server.Close()

		cfg := AggregatorClientConfig{
			AggregatorRPCUrl: server.URL,
			SenderPrivateKey: privateKeyHex,
		}
		// Use the real client for HTTP testing
		client, err := NewAggregatorClient(logger, cfg)
		require.NoError(t, err)

		success, err := client.SendTaskToPerformer(context.Background(), taskData)
		assert.NoError(t, err)
		assert.True(t, success)
	})

	t.Run("Failure: RPC Call Fails", func(t *testing.T) {
		// Mock a server that returns an internal server error.
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

		success, err := client.SendTaskToPerformer(context.Background(), taskData)
		assert.Error(t, err)
		assert.False(t, success)
		assert.Contains(t, err.Error(), "failed to send custom task")
	})

	t.Run("Failure: Network Dial Error", func(t *testing.T) {
		cfg := AggregatorClientConfig{
			AggregatorRPCUrl: "http://localhost:9999", // Unreachable port
			SenderPrivateKey: privateKeyHex,
		}
		client, err := NewAggregatorClient(logger, cfg)
		require.NoError(t, err)

		success, err := client.SendTaskToPerformer(context.Background(), taskData)
		assert.Error(t, err)
		assert.False(t, success)
		assert.Contains(t, err.Error(), "failed to send custom task")
		assert.Contains(t, err.Error(), "failed to dial aggregator RPC")
	})

	t.Run("Failure: RPC Method Not Found", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, err := fmt.Fprintln(w, `{"jsonrpc":"2.0","id":1,"error":{"code":-32601,"message":"Method not found"}}`)
			require.NoError(t, err)
		}))
		defer server.Close()

		cfg := AggregatorClientConfig{
			AggregatorRPCUrl: server.URL,
			SenderPrivateKey: privateKeyHex,
		}
		client, err := NewAggregatorClient(logger, cfg)
		require.NoError(t, err)

		success, err := client.SendTaskToPerformer(context.Background(), taskData)
		assert.Error(t, err)
		assert.False(t, success)
		assert.Contains(t, err.Error(), "failed to send custom task")
		assert.Contains(t, err.Error(), "RPC call failed")
	})
}

// TestSendTaskToPerformer validates the logic for sending a custom message.
// This is the original test function, now using the real client for HTTP testing
func TestSendTaskToPerformer(t *testing.T) {
	logger := logging.NewNoOpLogger()
	_, privateKeyHex := generateTestPrivateKey(t)

	taskData := &types.BroadcastDataForPerformer{
		TaskID:           123,
		PerformerAddress: "0x1234567890123456789012345678901234567890",
		Data:             []byte("some custom task data"),
		TaskDefinitionID: 42,
	}

	t.Run("Success", func(t *testing.T) {
		// Mock a JSON-RPC server that validates the incoming request.
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			bodyBytes, err := io.ReadAll(r.Body)
			require.NoError(t, err)

			var req jsonRPCRequest
			err = json.Unmarshal(bodyBytes, &req)
			require.NoError(t, err)

			// 1. Validate the method name.
			assert.Equal(t, "sendCustomMessage", req.Method)
			require.Len(t, req.Params, 2)

			// 2. Validate the parameters in the correct order: Data, TaskDefinitionID.
			expectedData := "0x" + hex.EncodeToString(taskData.Data)
			assert.Equal(t, expectedData, req.Params[0])
			// JSON unmarshals numbers into float64 by default, so we cast for comparison.
			assert.Equal(t, float64(taskData.TaskDefinitionID), req.Params[1])

			// 3. Send a successful response.
			w.WriteHeader(http.StatusOK)
			_, err = fmt.Fprintln(w, `{"jsonrpc":"2.0","id":1,"result":true}`)
			require.NoError(t, err)
		}))
		defer server.Close()

		cfg := AggregatorClientConfig{
			AggregatorRPCUrl: server.URL,
			SenderPrivateKey: privateKeyHex,
		}
		// Use the real client for HTTP testing
		client, err := NewAggregatorClient(logger, cfg)
		require.NoError(t, err)

		success, err := client.SendTaskToPerformer(context.Background(), taskData)
		assert.NoError(t, err)
		assert.True(t, success)
	})

	t.Run("Failure: RPC Call Fails", func(t *testing.T) {
		// Mock a server that returns an internal server error.
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

		success, err := client.SendTaskToPerformer(context.Background(), taskData)
		assert.Error(t, err)
		assert.False(t, success)
		assert.Contains(t, err.Error(), "failed to send custom task")
	})

	t.Run("Failure: Network Dial Error", func(t *testing.T) {
		cfg := AggregatorClientConfig{
			AggregatorRPCUrl: "http://localhost:9999", // Unreachable port
			SenderPrivateKey: privateKeyHex,
		}
		client, err := NewAggregatorClient(logger, cfg)
		require.NoError(t, err)

		success, err := client.SendTaskToPerformer(context.Background(), taskData)
		assert.Error(t, err)
		assert.False(t, success)
		assert.Contains(t, err.Error(), "failed to send custom task")
		assert.Contains(t, err.Error(), "failed to dial aggregator RPC")
	})

	t.Run("Failure: RPC Method Not Found", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, err := fmt.Fprintln(w, `{"jsonrpc":"2.0","id":1,"error":{"code":-32601,"message":"Method not found"}}`)
			require.NoError(t, err)
		}))
		defer server.Close()

		cfg := AggregatorClientConfig{
			AggregatorRPCUrl: server.URL,
			SenderPrivateKey: privateKeyHex,
		}
		client, err := NewAggregatorClient(logger, cfg)
		require.NoError(t, err)

		success, err := client.SendTaskToPerformer(context.Background(), taskData)
		assert.Error(t, err)
		assert.False(t, success)
		assert.Contains(t, err.Error(), "failed to send custom task")
		assert.Contains(t, err.Error(), "RPC call failed")
	})
}
