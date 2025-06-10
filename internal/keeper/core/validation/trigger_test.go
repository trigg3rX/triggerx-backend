package validation

import (
	"context"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	ethTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// MockLogger is a mock implementation of logging.Logger
type MockLogger struct {
	mock.Mock
}

func (m *MockLogger) Info(msg string, keysAndValues ...interface{}) {
	m.Called(msg, keysAndValues)
}

func (m *MockLogger) Error(msg string, keysAndValues ...interface{}) {
	m.Called(msg, keysAndValues)
}

func (m *MockLogger) Debug(msg string, keysAndValues ...interface{}) {
	m.Called(msg, keysAndValues)
}

func (m *MockLogger) Warn(msg string, keysAndValues ...interface{}) {
	m.Called(msg, keysAndValues)
}

func (m *MockLogger) Fatal(msg string, keysAndValues ...interface{}) {
	m.Called(msg, keysAndValues)
}

func (m *MockLogger) Debugf(template string, args ...interface{}) {
	m.Called(template, args)
}

func (m *MockLogger) Infof(template string, args ...interface{}) {
	m.Called(template, args)
}

func (m *MockLogger) Warnf(template string, args ...interface{}) {
	m.Called(template, args)
}

func (m *MockLogger) Errorf(template string, args ...interface{}) {
	m.Called(template, args)
}

func (m *MockLogger) Fatalf(template string, args ...interface{}) {
	m.Called(template, args)
}

func (m *MockLogger) With(tags ...interface{}) logging.Logger {
	args := m.Called(tags)
	return args.Get(0).(logging.Logger)
}

// MockEthClient is a mock implementation of EthClientInterface
type MockEthClient struct {
	mock.Mock
}

func (m *MockEthClient) CodeAt(ctx context.Context, account common.Address, blockNumber *big.Int) ([]byte, error) {
	args := m.Called(ctx, account, blockNumber)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockEthClient) TransactionReceipt(ctx context.Context, txHash common.Hash) (*ethTypes.Receipt, error) {
	args := m.Called(ctx, txHash)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	receipt, ok := args.Get(0).(*ethTypes.Receipt)
	if !ok {
		return nil, args.Error(1)
	}
	return receipt, args.Error(1)
}

func (m *MockEthClient) TransactionByHash(ctx context.Context, txHash common.Hash) (*ethTypes.Transaction, bool, error) {
	args := m.Called(ctx, txHash)
	if args.Get(0) == nil {
		return nil, args.Bool(1), args.Error(2)
	}
	tx, ok := args.Get(0).(*ethTypes.Transaction)
	if !ok {
		return nil, args.Bool(1), args.Error(2)
	}
	return tx, args.Bool(1), args.Error(2)
}

func (m *MockEthClient) BlockByHash(ctx context.Context, hash common.Hash) (*ethTypes.Block, error) {
	args := m.Called(ctx, hash)
	return args.Get(0).(*ethTypes.Block), args.Error(1)
}

func (m *MockEthClient) BlockByNumber(ctx context.Context, number *big.Int) (*ethTypes.Block, error) {
	args := m.Called(ctx, number)
	return args.Get(0).(*ethTypes.Block), args.Error(1)
}

func (m *MockEthClient) Close() {
	m.Called()
}

// MockUtils is a mock implementation of the utils package
type MockUtils struct {
	mock.Mock
}

func (m *MockUtils) GetChainRpcUrl(chainId string) string {
	args := m.Called(chainId)
	return args.String(0)
}

func setupTestValidator(mockClient EthClientInterface, mockUtils *MockUtils) *TaskValidator {
	mockLogger := new(MockLogger)
	// Setup default expectations for the mock logger
	mockLogger.On("Info", mock.Anything, mock.Anything).Return()
	mockLogger.On("Error", mock.Anything, mock.Anything).Return()
	mockLogger.On("Debug", mock.Anything, mock.Anything).Return()
	mockLogger.On("Warn", mock.Anything, mock.Anything).Return()
	mockLogger.On("With", mock.Anything).Return(mockLogger)

	// Setup default expectations for the mock utils
	mockUtils.On("GetChainRpcUrl", mock.Anything).Return("http://localhost:8545")

	validator := NewTaskValidator(mockLogger)
	validator.ethClientMaker = func(url string) (EthClientInterface, error) {
		return mockClient, nil
	}

	return validator
}

func setupMockEthClient() *MockEthClient {
	mockClient := new(MockEthClient)

	// Setup default responses for the mock client
	mockClient.On("CodeAt", mock.Anything, mock.Anything, mock.Anything).Return([]byte{1}, nil)

	receipt := &ethTypes.Receipt{
		Status:          1,
		ContractAddress: common.HexToAddress("0x1234567890123456789012345678901234567890"),
		BlockHash:       common.HexToHash("0xabcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890"),
		BlockNumber:     big.NewInt(1),
		Logs: []*ethTypes.Log{
			{
				Topics: []common.Hash{
					common.HexToHash("0x1234567890123456789012345678901234567890123456789012345678901234"),
				},
			},
		},
	}
	mockClient.On("TransactionReceipt", mock.Anything, mock.Anything).Return(receipt, nil)

	// Create a block with proper initialization
	header := &ethTypes.Header{
		Time: uint64(time.Now().Unix()),
	}
	body := &ethTypes.Body{
		Transactions: []*ethTypes.Transaction{},
		Uncles:       []*ethTypes.Header{},
	}
	block := ethTypes.NewBlock(header, body, nil, nil)
	mockClient.On("BlockByHash", mock.Anything, mock.Anything).Return(block, nil)
	mockClient.On("BlockByNumber", mock.Anything, mock.Anything).Return(block, nil)
	mockClient.On("Close").Return()

	return mockClient
}

func TestValidateTrigger(t *testing.T) {
	tests := []struct {
		name        string
		triggerData *types.TaskTriggerData
		want        bool
		wantErr     bool
	}{
		{
			name: "Valid Time Based Trigger",
			triggerData: &types.TaskTriggerData{
				TaskID:           1,
				TaskDefinitionID: 1,
				ExpirationTime:   time.Now().Add(1 * time.Hour),
				TriggerTimestamp: time.Now(),
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "Invalid Time Based Trigger - Expiration Before Trigger",
			triggerData: &types.TaskTriggerData{
				TaskID:           2,
				TaskDefinitionID: 1,
				ExpirationTime:   time.Now().Add(-1 * time.Hour),
				TriggerTimestamp: time.Now(),
			},
			want:    false,
			wantErr: true,
		},
		{
			name: "Valid Event Based Trigger",
			triggerData: &types.TaskTriggerData{
				TaskID:                      3,
				TaskDefinitionID:            3,
				EventChainId:                "1",
				EventTriggerContractAddress: "0x1234567890123456789012345678901234567890",
				EventTxHash:                 "0xabcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
				EventTriggerName:            "0x1234567890123456789012345678901234567890123456789012345678901234",
				ExpirationTime:              time.Now().Add(1 * time.Hour),
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "Valid Condition Based Trigger - Equals",
			triggerData: &types.TaskTriggerData{
				TaskID:                  4,
				TaskDefinitionID:        5,
				ConditionSourceType:     ConditionEquals,
				ConditionSatisfiedValue: 100,
				ConditionUpperLimit:     100,
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "Valid Condition Based Trigger - Not Equals",
			triggerData: &types.TaskTriggerData{
				TaskID:                  5,
				TaskDefinitionID:        5,
				ConditionSourceType:     ConditionNotEquals,
				ConditionSatisfiedValue: 100,
				ConditionUpperLimit:     200,
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "Valid Condition Based Trigger - Greater Than",
			triggerData: &types.TaskTriggerData{
				TaskID:                  6,
				TaskDefinitionID:        5,
				ConditionSourceType:     ConditionGreaterThan,
				ConditionSatisfiedValue: 200,
				ConditionUpperLimit:     100,
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "Valid Condition Based Trigger - Less Than",
			triggerData: &types.TaskTriggerData{
				TaskID:                  7,
				TaskDefinitionID:        5,
				ConditionSourceType:     ConditionLessThan,
				ConditionSatisfiedValue: 50,
				ConditionUpperLimit:     100,
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "Valid Condition Based Trigger - Between",
			triggerData: &types.TaskTriggerData{
				TaskID:                  8,
				TaskDefinitionID:        5,
				ConditionSourceType:     ConditionBetween,
				ConditionSatisfiedValue: 150,
				ConditionLowerLimit:     100,
				ConditionUpperLimit:     200,
			},
			want:    true,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := setupMockEthClient()
			mockUtils := new(MockUtils)
			validator := setupTestValidator(mockClient, mockUtils)

			got, err := validator.ValidateTrigger(tt.triggerData, "test-trace-id")
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestIsValidTimeBasedTrigger(t *testing.T) {
	tests := []struct {
		name        string
		triggerData *types.TaskTriggerData
		want        bool
		wantErr     bool
	}{
		{
			name: "Valid Time Based Trigger",
			triggerData: &types.TaskTriggerData{
				TaskID:           1,
				ExpirationTime:   time.Now().Add(1 * time.Hour),
				TriggerTimestamp: time.Now(),
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "Invalid Time Based Trigger - Expiration Before Trigger",
			triggerData: &types.TaskTriggerData{
				TaskID:           2,
				ExpirationTime:   time.Now().Add(-1 * time.Hour),
				TriggerTimestamp: time.Now(),
			},
			want:    false,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := setupTestValidator(nil, new(MockUtils))
			got, err := validator.IsValidTimeBasedTrigger(tt.triggerData)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestIsValidConditionBasedTrigger(t *testing.T) {
	tests := []struct {
		name        string
		triggerData *types.TaskTriggerData
		want        bool
		wantErr     bool
	}{
		{
			name: "Valid Condition - Equals",
			triggerData: &types.TaskTriggerData{
				TaskID:                  1,
				ConditionSourceType:     ConditionEquals,
				ConditionSatisfiedValue: 100,
				ConditionUpperLimit:     100,
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "Invalid Condition - Equals",
			triggerData: &types.TaskTriggerData{
				TaskID:                  2,
				ConditionSourceType:     ConditionEquals,
				ConditionSatisfiedValue: 100,
				ConditionUpperLimit:     200,
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "Valid Condition - Not Equals",
			triggerData: &types.TaskTriggerData{
				TaskID:                  3,
				ConditionSourceType:     ConditionNotEquals,
				ConditionSatisfiedValue: 100,
				ConditionUpperLimit:     200,
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "Valid Condition - Greater Than",
			triggerData: &types.TaskTriggerData{
				TaskID:                  4,
				ConditionSourceType:     ConditionGreaterThan,
				ConditionSatisfiedValue: 200,
				ConditionUpperLimit:     100,
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "Valid Condition - Less Than",
			triggerData: &types.TaskTriggerData{
				TaskID:                  5,
				ConditionSourceType:     ConditionLessThan,
				ConditionSatisfiedValue: 50,
				ConditionUpperLimit:     100,
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "Valid Condition - Between",
			triggerData: &types.TaskTriggerData{
				TaskID:                  6,
				ConditionSourceType:     ConditionBetween,
				ConditionSatisfiedValue: 150,
				ConditionLowerLimit:     100,
				ConditionUpperLimit:     200,
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "Invalid Condition - Between",
			triggerData: &types.TaskTriggerData{
				TaskID:                  7,
				ConditionSourceType:     ConditionBetween,
				ConditionSatisfiedValue: 300,
				ConditionLowerLimit:     100,
				ConditionUpperLimit:     200,
			},
			want:    false,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := setupTestValidator(nil, new(MockUtils))
			got, err := validator.IsValidConditionBasedTrigger(tt.triggerData)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}
