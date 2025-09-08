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

// TestNewMockAggregatorClient ensures the basic constructor works as expected.
func TestNewMockAggregatorClient(t *testing.T) {
	logger := logging.NewNoOpLogger()
	cfg := NewMockAggregatorClientConfig()

	client, err := NewMockAggregatorClient(logger, cfg)
	require.NoError(t, err)
	require.NotNil(t, client)
}

// TestMockAggregatorClientBuilder validates that the builder correctly sets up specific mock expectations.
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

// TestMockAggregatorClientBuilder_AnyMatchers validates the builder methods using mock.Anything.
func TestMockAggregatorClientBuilder_AnyMatchers(t *testing.T) {
	expectedErr := errors.New("any call fails")

	builder := NewMockAggregatorClientBuilder().
		ExpectSendTaskToPerformerAny(true, nil).
		ExpectSendTaskToValidatorsAny(false, expectedErr)

	mockClient := builder.Build()
	require.NotNil(t, mockClient)

	// Call methods with arbitrary arguments to ensure "Anything" matcher works
	success, err := mockClient.SendTaskToPerformer(context.TODO(), &types.BroadcastDataForPerformer{TaskID: 999})
	assert.True(t, success)
	assert.NoError(t, err)

	success, err = mockClient.SendTaskToValidators(context.Background(), nil)
	assert.False(t, success)
	assert.Equal(t, expectedErr, err)

	builder.AssertExpectations(t)
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

	// Verify the mock calls were made
	client.AssertExpectations(t)
}

// TestNewFailingAggregatorClient checks that the client always returns a specified error.
func TestNewFailingAggregatorClient(t *testing.T) {
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

	// Verify the mock calls were made
	client.AssertExpectations(t)
}

// TestMockAggregatorClientFactory validates the mock factory for creating clients.
func TestMockAggregatorClientFactory(t *testing.T) {
	logger := logging.NewNoOpLogger()
	cfg := AggregatorClientConfig{} // Use an empty config as the value doesn't matter here.

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

// TestNewMockAggregatorClientConfig verifies the content of the mock config.
func TestNewMockAggregatorClientConfig(t *testing.T) {
	cfg := NewMockAggregatorClientConfig()

	assert.Equal(t, "http://localhost:9007", cfg.AggregatorRPCUrl)
	assert.Equal(t, "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef", cfg.SenderPrivateKey)
	assert.Equal(t, "0x742d35Cc6634C0532925a3b8D4C9db96C4b4d8b6", cfg.SenderAddress)
	assert.Equal(t, 3, cfg.RetryAttempts)
	assert.Equal(t, 1000*time.Millisecond, cfg.RetryDelay)
	assert.Equal(t, 10000*time.Millisecond, cfg.RequestTimeout)
}

// TestHTTPMockHelpers validates the constructors intended for HTTP integration tests.
func TestHTTPMockHelpers(t *testing.T) {
	logger := logging.NewNoOpLogger()
	cfg := NewMockAggregatorClientConfig()

	t.Run("NewMockAggregatorClientWithHTTP", func(t *testing.T) {
		client, err := NewMockAggregatorClientWithHTTP(logger, cfg)
		require.NoError(t, err)
		require.NotNil(t, client)

		// This client should default to success.
		success, err := client.SendTaskToPerformer(context.Background(), nil)
		assert.True(t, success)
		assert.NoError(t, err)

		success, err = client.SendTaskToValidators(context.Background(), nil)
		assert.True(t, success)
		assert.NoError(t, err)

		// Call Close() as expected by the mock
		client.Close()

		client.AssertExpectations(t)
	})

	t.Run("NewMockAggregatorClientForHTTPTest - Success", func(t *testing.T) {
		client, err := NewMockAggregatorClientForHTTPTest(logger, cfg, true, nil)
		require.NoError(t, err)
		require.NotNil(t, client)

		success, err := client.SendTaskToPerformer(context.Background(), nil)
		assert.True(t, success)
		assert.NoError(t, err)

		success, err = client.SendTaskToValidators(context.Background(), nil)
		assert.True(t, success)
		assert.NoError(t, err)

		// Call Close() as expected by the mock
		client.Close()

		client.AssertExpectations(t)
	})

	t.Run("NewMockAggregatorClientForHTTPTest - Failure", func(t *testing.T) {
		expectedErr := errors.New("http test failure")
		client, err := NewMockAggregatorClientForHTTPTest(logger, cfg, false, expectedErr)
		require.NoError(t, err)
		require.NotNil(t, client)

		success, err := client.SendTaskToPerformer(context.Background(), nil)
		assert.False(t, success)
		assert.Equal(t, expectedErr, err)

		success, err = client.SendTaskToValidators(context.Background(), nil)
		assert.False(t, success)
		assert.Equal(t, expectedErr, err)

		// Call Close() as expected by the mock
		client.Close()

		client.AssertExpectations(t)
	})
}
