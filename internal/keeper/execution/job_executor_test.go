package execution

import (
	"context"
	"math/big"
	"os"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	jobtypes "github.com/trigg3rX/triggerx-backend/pkg/types"
)

// Mock EthClient
type MockEthClient struct {
	mock.Mock
}

func (m *MockEthClient) PendingNonceAt(ctx context.Context, account common.Address) (uint64, error) {
	args := m.Called(ctx, account)
	return uint64(args.Int(0)), args.Error(1)
}

func (m *MockEthClient) SuggestGasPrice(ctx context.Context) (*big.Int, error) {
	args := m.Called(ctx)
	return args.Get(0).(*big.Int), args.Error(1)
}

func (m *MockEthClient) SendTransaction(ctx context.Context, tx *types.Transaction) error {
	args := m.Called(ctx, tx)
	return args.Error(0)
}

func (m *MockEthClient) CodeAt(ctx context.Context, contract common.Address, blockNumber *big.Int) ([]byte, error) {
	args := m.Called(ctx, contract, blockNumber)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockEthClient) CallContract(ctx context.Context, call ethereum.CallMsg, blockNumber *big.Int) ([]byte, error) {
	args := m.Called(ctx, call, blockNumber)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockEthClient) HeaderByNumber(ctx context.Context, number *big.Int) (*types.Header, error) {
	args := m.Called(ctx, number)
	return args.Get(0).(*types.Header), args.Error(1)
}

func (m *MockEthClient) TransactionReceipt(ctx context.Context, txHash common.Hash) (*types.Receipt, error) {
	args := m.Called(ctx, txHash)
	return args.Get(0).(*types.Receipt), args.Error(1)
}

// Mock Logger
type MockLogger struct {
	mock.Mock
}

func (m *MockLogger) Infof(format string, args ...interface{}) {
	m.Called(format, args)
}

func (m *MockLogger) Errorf(format string, args ...interface{}) {
	m.Called(format, args)
}

func (m *MockLogger) Warnf(format string, args ...interface{}) {
	m.Called(format, args)
}

func TestExecuteActionWithDynamicArgs(t *testing.T) {
	// Setup test cases
	tests := []struct {
		name           string
		job            *jobtypes.HandleCreateJobData
		setupMocks     func(*MockEthClient)
		expectedError  bool
		expectedResult jobtypes.ActionData
	}{
		{
			name: "Test with script trigger function",
			job: &jobtypes.HandleCreateJobData{
				JobID:                 1,
				TaskDefinitionID:      6,
				ScriptTriggerFunction: "https://gateway.lighthouse.storage/ipfs/bafkreiaeuy3fyzaecbh2zolndnebccpnrkpwobigtmugzntnyew5oprb4a",
				TargetContractAddress: "0x98a170b9b24aD4f42B6B3630A54517fd7Ff3Ac6d",
				TargetFunction:        "addTaskId",
				Arguments:             []string{`{"jobId": "2", "taskId": "5"}`},
			},
			setupMocks: func(mockEth *MockEthClient) {
				mockEth.On("PendingNonceAt", mock.Anything, mock.Anything).Return(uint64(1), nil)
				mockEth.On("SuggestGasPrice", mock.Anything).Return(big.NewInt(20000000000), nil)
				mockEth.On("SendTransaction", mock.Anything, mock.Anything).Return(nil)

				// Add contract related mocks
				mockEth.On("CodeAt", mock.Anything, mock.Anything, mock.Anything).Return([]byte("contract-code"), nil)
				mockEth.On("CallContract", mock.Anything, mock.Anything, mock.Anything).Return([]byte(`{"jobId": 2, "taskId": 5}`), nil)
				mockEth.On("HeaderByNumber", mock.Anything, mock.Anything).Return(&types.Header{}, nil)
				mockEth.On("TransactionReceipt", mock.Anything, mock.Anything).Return(&types.Receipt{
					Status:  1,
					GasUsed: 21000,
				}, nil)
			},
			expectedError: false,
			expectedResult: jobtypes.ActionData{
				Status:    true,
				GasUsed:   "21000",
				TaskID:    0,
				Timestamp: time.Now().UTC(),
			},
		},
		{
			name: "Test with script IPFS URL",
			job: &jobtypes.HandleCreateJobData{
				JobID:                 2,
				TaskDefinitionID:      4,
				ScriptIPFSUrl:         "https://gateway.lighthouse.storage/ipfs/bafkreiaeuy3fyzaecbh2zolndnebccpnrkpwobigtmugzntnyew5oprb4a",
				TargetContractAddress: "0x98a170b9b24aD4f42B6B3630A54517fd7Ff3Ac6d",
				TargetFunction:        "addTaskId",
				Arguments:             []string{`{"jobId": "2", "taskId": "5"}`},
			},
			setupMocks: func(mockEth *MockEthClient) {
				mockEth.On("PendingNonceAt", mock.Anything, mock.Anything).Return(uint64(1), nil)
				mockEth.On("SuggestGasPrice", mock.Anything).Return(big.NewInt(20000000000), nil)
				mockEth.On("SendTransaction", mock.Anything, mock.Anything).Return(nil)

				// Add contract related mocks
				mockEth.On("CodeAt", mock.Anything, mock.Anything, mock.Anything).Return([]byte("contract-code"), nil)
				mockEth.On("CallContract", mock.Anything, mock.Anything, mock.Anything).Return([]byte(`{"jobId": 2, "taskId": 5}`), nil)
				mockEth.On("HeaderByNumber", mock.Anything, mock.Anything).Return(&types.Header{}, nil)
				mockEth.On("TransactionReceipt", mock.Anything, mock.Anything).Return(&types.Receipt{
					Status:  1,
					GasUsed: 21000,
				}, nil)
			},
			expectedError: false,
			expectedResult: jobtypes.ActionData{
				Status:    true,
				GasUsed:   "21000",
				TaskID:    0,
				Timestamp: time.Now().UTC(),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mocks
			mockEth := new(MockEthClient)
			mockLogger := new(MockLogger)

			// Setup mocks
			tt.setupMocks(mockEth)
			mockLogger.On("Infof", mock.Anything, mock.Anything).Return()
			mockLogger.On("Errorf", mock.Anything, mock.Anything).Return()
			mockLogger.On("Warnf", mock.Anything, mock.Anything).Return()

			// Create executor
			executor := &JobExecutor{
				ethClient:       mockEth,
				etherscanAPIKey: "test-key",
				argConverter:    &ArgumentConverter{},
				logger:          mockLogger,
			}

			// Execute test
			result, err := executor.executeActionWithDynamicArgs(tt.job)

			// Assert results
			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult.Status, result.Status)
				assert.NotEmpty(t, result.ActionTxHash)
			}

			// Verify mocks
			mockEth.AssertExpectations(t)
			mockLogger.AssertExpectations(t)
		})
	}
}

// Test helper functions
func TestProcessArguments(t *testing.T) {
	executor := &JobExecutor{
		argConverter: &ArgumentConverter{},
	}

	tests := []struct {
		name          string
		args          interface{}
		methodInputs  []abi.Argument
		expectedError bool
	}{
		{
			name: "Test with simple string array",
			args: []string{"123", "test"},
			methodInputs: []abi.Argument{
				{Type: abi.Type{T: abi.UintTy, Size: 256}},
				{Type: abi.Type{T: abi.StringTy}},
			},
			expectedError: false,
		},
		// Add more test cases for different argument types
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := executor.processArguments(tt.args, tt.methodInputs, nil)
			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestIntegration(t *testing.T) {
	// Skip if not in integration test environment
	if os.Getenv("INTEGRATION_TEST") != "true" {
		t.Skip("Skipping integration test")
	}

	// Setup real ethclient
	client, err := ethclient.Dial("https://eth-holesky.g.alchemy.com/v2/")
	if err != nil {
		t.Fatal(err)
	}

	executor := NewJobExecutor(client, "ETHERSCAN_API_KEY")

	// Create test job
	job := &jobtypes.HandleCreateJobData{
		JobID:                 1,
		TaskDefinitionID:      4,
		ScriptIPFSUrl:         "https://gateway.lighthouse.storage/ipfs/bafkreiaeuy3fyzaecbh2zolndnebccpnrkpwobigtmugzntnyew5oprb4a",
		TargetContractAddress: "0x98a170b9b24aD4f42B6B3630A54517fd7Ff3Ac6d",
		TargetFunction:        "addTaskId",
	}

	// Execute test
	result, err := executor.executeActionWithDynamicArgs(job)
	assert.NoError(t, err)
	assert.True(t, result.Status)
}
