package aggregator

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// TestMockAggregatorClientBuilder validates that the builder correctly sets up mock expectations.
func TestMockAggregatorClientBuilder(t *testing.T) {
	ctx := context.Background()
	taskData := &types.BroadcastDataForPerformer{TaskID: 123}
	taskResult := &types.BroadcastDataForValidators{ProofOfTask: "proof-1"}
	expectedErr := errors.New("something went wrong")

	// 1. Setup builder with all expectations
	builder := NewMockAggregatorClientBuilder().
		ExpectSendTaskToPerformer(ctx, taskData, true, nil).
		ExpectSendTaskToValidators(ctx, taskResult, false, expectedErr).
		ExpectClose()

	// 2. Build the mock client
	mockClient := builder.Build()
	require.NotNil(t, mockClient)

	// 3. Call the methods and assert the results match expectations
	success, err := mockClient.SendTaskToPerformer(ctx, taskData)
	assert.True(t, success)
	assert.NoError(t, err)

	success, err = mockClient.SendTaskToValidators(ctx, taskResult)
	assert.False(t, success)
	assert.Equal(t, expectedErr, err)

	mockClient.Close()

	// 4. Verify that all expected calls were made
	builder.AssertExpectations(t)
	builder.AssertNumberOfCalls(t, "SendTaskToPerformer", 1)
	builder.AssertNumberOfCalls(t, "SendTaskToValidators", 1)
	builder.AssertNumberOfCalls(t, "Close", 1)
}

// TestNewNoOpAggregatorClient checks that the client does nothing and returns success.
func TestNewNoOpAggregatorClient(t *testing.T) {
	client := NewNoOpAggregatorClient()
	require.NotNil(t, client)

	// All methods should return success (true, nil) without any setup.
	success, err := client.SendTaskToPerformer(context.Background(), nil)
	assert.True(t, success)
	assert.NoError(t, err)

	success, err = client.SendTaskToValidators(context.Background(), nil)
	assert.True(t, success)
	assert.NoError(t, err)

	// Close should not panic
	assert.NotPanics(t, func() { client.Close() })
}

// TestNewFailingAggregatorClient checks that the client always returns a specified error.
func TestNewFailingAggregatorClient(t *testing.T) {
	// Define a specific error to check against.
	failErr := errors.New("simulated network failure")
	client := NewFailingAggregatorClient(failErr)
	require.NotNil(t, client)

	// All methods should return failure (false, failErr).
	success, err := client.SendTaskToPerformer(context.Background(), nil)
	assert.False(t, success)
	assert.Equal(t, failErr, err)

	success, err = client.SendTaskToValidators(context.Background(), nil)
	assert.False(t, success)
	assert.Equal(t, failErr, err)

	// Close should not panic
	assert.NotPanics(t, func() { client.Close() })
}

// TestMockAggregatorClientFactory validates the mock factory for creating clients.
func TestMockAggregatorClientFactory(t *testing.T) {
	logger := logging.NewNoOpLogger()
	// As recommended, use the actual config struct for testing purposes.
	cfg := AggregatorClientConfig{}

	t.Run("Success: Factory returns a mock client", func(t *testing.T) {
		factory := NewMockAggregatorClientFactory()
		expectedClient := NewNoOpAggregatorClient()

		// Expect a call to CreateAggregatorClient and tell it what to return.
		factory.On("CreateAggregatorClient", logger, cfg).Return(expectedClient, nil)

		// Call the factory method.
		client, err := factory.CreateAggregatorClient(logger, cfg)

		// Assert that we got back what we expected.
		assert.NoError(t, err)
		assert.Equal(t, expectedClient, client)
		factory.AssertExpectations(t)
	})

	t.Run("Failure: Factory returns an error", func(t *testing.T) {
		factory := NewMockAggregatorClientFactory()
		expectedErr := errors.New("config validation failed")

		// Expect a call and tell it to return an error.
		factory.On("CreateAggregatorClient", logger, cfg).Return(nil, expectedErr)

		// Call the factory method.
		client, err := factory.CreateAggregatorClient(logger, cfg)

		// Assert that we got back the expected error.
		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		assert.Nil(t, client)
		factory.AssertExpectations(t)
	})
}

func TestNewMockAggregatorClientConfig(t *testing.T) {
	cfg := NewMockAggregatorClientConfig()
	require.NotNil(t, cfg)
	assert.Equal(t, "http://localhost:9007", cfg.AggregatorRPCUrl)
	assert.Equal(t, "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef", cfg.SenderPrivateKey)
	assert.Equal(t, "0x742d35Cc6634C0532925a3b8D4C9db96C4b4d8b6", cfg.SenderAddress)
	assert.Equal(t, 3, cfg.RetryAttempts)
	assert.Equal(t, 1000*time.Millisecond, cfg.RetryDelay)
	assert.Equal(t, 10000*time.Millisecond, cfg.RequestTimeout)
}
