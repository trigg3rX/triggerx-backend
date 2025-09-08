package aggregator

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// TestSendTaskToValidators validates the ECDSA signing and sending logic.
func TestSendTaskToValidators(t *testing.T) {
	logger := logging.NewNoOpLogger()
	privateKey, privateKeyHex := generateTestPrivateKey(t)
	senderAddress := crypto.PubkeyToAddress(privateKey.PublicKey).Hex()

	taskResult := &types.BroadcastDataForValidators{
		ProofOfTask:      "ecdsa-proof-123",
		Data:             []byte("ecdsa result data"),
		TaskDefinitionID: 101,
	}

	t.Run("Success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, err := io.ReadAll(r.Body)
			require.NoError(t, err)

			var req jsonRPCRequest
			err = json.Unmarshal(body, &req)
			require.NoError(t, err)

			assert.Equal(t, "sendTask", req.Method)
			require.Len(t, req.Params, 7, "expected 7 parameters for sendTask")

			// Validate parameters in the order they are passed to rpcClient.Call
			assert.Equal(t, taskResult.ProofOfTask, req.Params[0])
			assert.Equal(t, "0x"+hex.EncodeToString(taskResult.Data), req.Params[1])
			assert.Equal(t, float64(taskResult.TaskDefinitionID), req.Params[2])
			assert.Equal(t, strings.ToLower(senderAddress), strings.ToLower(req.Params[3].(string)))
			assert.NotEmpty(t, req.Params[4], "signature should not be empty")
			assert.Equal(t, "ecdsa", req.Params[5])
			assert.Equal(t, float64(84532), req.Params[6]) // TargetChainID

			w.WriteHeader(http.StatusOK)
			_, err = fmt.Fprintln(w, `{"jsonrpc":"2.0","id":1,"result":true}`)
			require.NoError(t, err)
		}))
		defer server.Close()

		cfg := AggregatorClientConfig{
			AggregatorRPCUrl: server.URL,
			SenderPrivateKey: privateKeyHex,
			SenderAddress:    senderAddress,
		}
		client, err := NewAggregatorClient(logger, cfg)
		require.NoError(t, err)

		success, err := client.SendTaskToValidators(context.Background(), taskResult)
		assert.NoError(t, err)
		assert.True(t, success)
	})

	t.Run("Failure on RPC Call", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		cfg := AggregatorClientConfig{
			AggregatorRPCUrl: server.URL,
			SenderPrivateKey: privateKeyHex,
			SenderAddress:    senderAddress,
		}
		client, err := NewAggregatorClient(logger, cfg)
		require.NoError(t, err)

		// Override the retry config to fail fast for this test
		client.httpClient.HTTPConfig.RetryConfig.MaxRetries = 1
		client.httpClient.HTTPConfig.RetryConfig.InitialDelay = 10 * time.Millisecond
		client.httpClient.HTTPConfig.RetryConfig.MaxDelay = 10 * time.Millisecond

		success, err := client.SendTaskToValidators(context.Background(), taskResult)
		assert.Error(t, err)
		assert.False(t, success)
		assert.Contains(t, err.Error(), "failed to send task result")
	})

	t.Run("Failure: Invalid Private Key", func(t *testing.T) {
		cfg := AggregatorClientConfig{
			AggregatorRPCUrl: "http://localhost:8080", // Not used since client creation fails
			SenderPrivateKey: "invalid-private-key",
			SenderAddress:    senderAddress,
		}
		client, err := NewAggregatorClient(logger, cfg)
		assert.Error(t, err)
		assert.Nil(t, client)
		assert.Contains(t, err.Error(), "invalid hex character 'i'")
	})

	t.Run("Success: Different Sender Address", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, err := fmt.Fprintln(w, `{"jsonrpc":"2.0","id":1,"result":true}`)
			require.NoError(t, err)
		}))
		defer server.Close()

		cfg := AggregatorClientConfig{
			AggregatorRPCUrl: server.URL,
			SenderPrivateKey: privateKeyHex,
			SenderAddress:    "0x1234567890123456789012345678901234567890", // Valid hex but different from derived address
		}
		client, err := NewAggregatorClient(logger, cfg)
		require.NoError(t, err)

		// This should succeed since the address is valid hex
		success, err := client.SendTaskToValidators(context.Background(), taskResult)
		assert.NoError(t, err)
		assert.True(t, success)
	})

	t.Run("Failure: Network Dial Error", func(t *testing.T) {
		cfg := AggregatorClientConfig{
			AggregatorRPCUrl: "http://localhost:9999", // Unreachable port
			SenderPrivateKey: privateKeyHex,
			SenderAddress:    senderAddress,
		}
		client, err := NewAggregatorClient(logger, cfg)
		require.NoError(t, err)

		success, err := client.SendTaskToValidators(context.Background(), taskResult)
		assert.Error(t, err)
		assert.False(t, success)
		assert.Contains(t, err.Error(), "failed to send task result")
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
			SenderAddress:    senderAddress,
		}
		client, err := NewAggregatorClient(logger, cfg)
		require.NoError(t, err)

		success, err := client.SendTaskToValidators(context.Background(), taskResult)
		assert.Error(t, err)
		assert.False(t, success)
		assert.Contains(t, err.Error(), "failed to send task result")
		assert.Contains(t, err.Error(), "RPC call failed")
	})
}
