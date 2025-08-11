package aggregator

import (
	"context"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// MockAggregatorClient is a mock implementation of the AggregatorClient
type MockAggregatorClient struct {
	mock.Mock
}

func NewMockAggregatorClient(logger logging.Logger, cfg AggregatorClientConfig) (*MockAggregatorClient, error) {
	return &MockAggregatorClient{}, nil
}

// SendTaskToPerformer mocks the SendTaskToPerformer method
func (m *MockAggregatorClient) SendTaskToPerformer(ctx context.Context, taskData *types.BroadcastDataForPerformer) (bool, error) {
	args := m.Called(ctx, taskData)
	return args.Bool(0), args.Error(1)
}

// SendTaskToValidators mocks the SendTaskToValidators method
func (m *MockAggregatorClient) SendTaskToValidators(ctx context.Context, taskResult *types.BroadcastDataForValidators) (bool, error) {
	args := m.Called(ctx, taskResult)
	return args.Bool(0), args.Error(1)
}

// Close mocks the Close method
func (m *MockAggregatorClient) Close() {
	m.Called()
}

// MockAggregatorClientBuilder provides a fluent interface for building mock aggregator clients
type MockAggregatorClientBuilder struct {
	client *MockAggregatorClient
}

// NewMockAggregatorClientBuilder creates a new mock aggregator client builder
func NewMockAggregatorClientBuilder() *MockAggregatorClientBuilder {
	return &MockAggregatorClientBuilder{
		client: &MockAggregatorClient{},
	}
}

// ExpectSendTaskToPerformer sets up an expectation for a SendTaskToPerformer call
func (b *MockAggregatorClientBuilder) ExpectSendTaskToPerformer(ctx context.Context, taskData *types.BroadcastDataForPerformer, success bool, err error) *MockAggregatorClientBuilder {
	b.client.On("SendTaskToPerformer", ctx, taskData).Return(success, err)
	return b
}

// ExpectSendTaskToPerformerAny sets up an expectation for any SendTaskToPerformer call
func (b *MockAggregatorClientBuilder) ExpectSendTaskToPerformerAny(success bool, err error) *MockAggregatorClientBuilder {
	b.client.On("SendTaskToPerformer", mock.Anything, mock.Anything).Return(success, err)
	return b
}

// ExpectSendTaskToValidators sets up an expectation for a SendTaskToValidators call
func (b *MockAggregatorClientBuilder) ExpectSendTaskToValidators(ctx context.Context, taskResult *types.BroadcastDataForValidators, success bool, err error) *MockAggregatorClientBuilder {
	b.client.On("SendTaskToValidators", ctx, taskResult).Return(success, err)
	return b
}

// ExpectSendTaskToValidatorsAny sets up an expectation for any SendTaskToValidators call
func (b *MockAggregatorClientBuilder) ExpectSendTaskToValidatorsAny(success bool, err error) *MockAggregatorClientBuilder {
	b.client.On("SendTaskToValidators", mock.Anything, mock.Anything).Return(success, err)
	return b
}

// ExpectClose sets up an expectation for a Close call
func (b *MockAggregatorClientBuilder) ExpectClose() *MockAggregatorClientBuilder {
	b.client.On("Close").Return()
	return b
}

// Build returns the configured mock aggregator client
func (b *MockAggregatorClientBuilder) Build() *MockAggregatorClient {
	return b.client
}

// AssertExpectations asserts that all expected calls were made
func (b *MockAggregatorClientBuilder) AssertExpectations(t mock.TestingT) bool {
	return b.client.AssertExpectations(t)
}

// AssertNumberOfCalls asserts the number of calls to a specific method
func (b *MockAggregatorClientBuilder) AssertNumberOfCalls(t mock.TestingT, methodName string, expectedCalls int) bool {
	return b.client.AssertNumberOfCalls(t, methodName, expectedCalls)
}

// MockAggregatorClientFactory is a factory for creating mock aggregator clients
type MockAggregatorClientFactory struct {
	mock.Mock
}

// NewMockAggregatorClientFactory creates a new mock aggregator client factory
func NewMockAggregatorClientFactory() *MockAggregatorClientFactory {
	return &MockAggregatorClientFactory{}
}

// CreateAggregatorClient mocks the aggregator client creation process
func (f *MockAggregatorClientFactory) CreateAggregatorClient(logger logging.Logger, cfg AggregatorClientConfig) (*MockAggregatorClient, error) {
	args := f.Called(logger, cfg)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*MockAggregatorClient), args.Error(1)
}

// Test utilities for common mock setups

// NewNoOpAggregatorClient creates an aggregator client that does nothing (useful for tests that don't care about aggregator operations)
func NewNoOpAggregatorClient() *MockAggregatorClient {
	client := &MockAggregatorClient{}

	// Set up default no-op behavior
	client.On("SendTaskToPerformer", mock.Anything, mock.Anything).Return(true, nil)
	client.On("SendTaskToValidators", mock.Anything, mock.Anything).Return(true, nil)
	client.On("Close").Return()

	return client
}

// NewFailingAggregatorClient creates an aggregator client that always fails (useful for testing error scenarios)
func NewFailingAggregatorClient(err error) *MockAggregatorClient {
	client := &MockAggregatorClient{}

	// Set up default failing behavior
	client.On("SendTaskToPerformer", mock.Anything, mock.Anything).Return(false, err)
	client.On("SendTaskToValidators", mock.Anything, mock.Anything).Return(false, err)
	client.On("Close").Return()

	return client
}

// NewMockAggregatorClientConfig creates a new mock aggregator client config
func NewMockAggregatorClientConfig() AggregatorClientConfig {
	return AggregatorClientConfig{
		AggregatorRPCUrl: "http://localhost:9007",
		SenderPrivateKey: "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
		SenderAddress:    "0x742d35Cc6634C0532925a3b8D4C9db96C4b4d8b6",
		RetryAttempts:    3,
		RetryDelay:       1000 * time.Millisecond,
		RequestTimeout:   10000 * time.Millisecond,
	}
}

// Mock utilities for testing with real HTTP servers

// NewMockAggregatorClientWithHTTP creates a mock client that can be used with real HTTP servers
// This is useful for integration tests where you want to test the actual HTTP interaction
func NewMockAggregatorClientWithHTTP(logger logging.Logger, cfg AggregatorClientConfig) (*MockAggregatorClient, error) {
	client := &MockAggregatorClient{}

	// Set up expectations that will be called by the real client
	client.On("SendTaskToPerformer", mock.Anything, mock.Anything).Return(true, nil)
	client.On("SendTaskToValidators", mock.Anything, mock.Anything).Return(true, nil)
	client.On("Close").Return()

	return client, nil
}

// NewMockAggregatorClientForHTTPTest creates a mock client specifically for HTTP testing
// This sets up the mock to work with real HTTP servers in tests
func NewMockAggregatorClientForHTTPTest(logger logging.Logger, cfg AggregatorClientConfig, expectedSuccess bool, expectedError error) (*MockAggregatorClient, error) {
	client := &MockAggregatorClient{}

	// Set up expectations for HTTP testing
	client.On("SendTaskToPerformer", mock.Anything, mock.Anything).Return(expectedSuccess, expectedError)
	client.On("SendTaskToValidators", mock.Anything, mock.Anything).Return(expectedSuccess, expectedError)
	client.On("Close").Return()

	return client, nil
}
