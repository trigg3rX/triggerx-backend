package execution

// import (
// 	"context"
// 	"fmt"
// 	"math/big"
// 	"os"
// 	"testing"
// 	"time"

// 	"github.com/ethereum/go-ethereum"
// 	"github.com/ethereum/go-ethereum/accounts/abi"
// 	ethcommon "github.com/ethereum/go-ethereum/common"
// 	"github.com/ethereum/go-ethereum/core/types"
// 	"github.com/ethereum/go-ethereum/ethclient"
// 	"github.com/stretchr/testify/assert"
// 	"github.com/stretchr/testify/mock"
// 	jobtypes "github.com/trigg3rX/triggerx-backend/pkg/types"
// 	"github.com/trigg3rX/triggerx-backend/internal/keeper/interfaces"
// )

// var (
// 	_ interfaces.EthClientInterface = (*MockEthClient)(nil)
// 	_ interfaces.Logger             = (*MockLogger)(nil)
// 	_ interfaces.ValidatorInterface = (*MockValidator)(nil)
// )

// type MockEthClient struct {
// 	mock.Mock
// }

// type MockLogger struct {
// 	mock.Mock
// }

// type MockValidator struct {
// 	mock.Mock
// }

// func (m *MockEthClient) PendingNonceAt(ctx context.Context, account ethcommon.Address) (uint64, error) {
// 	args := m.Called(ctx, account)
// 	return uint64(args.Int(0)), args.Error(1)
// }

// func (m *MockEthClient) SuggestGasPrice(ctx context.Context) (*big.Int, error) {
// 	args := m.Called(ctx)
// 	return args.Get(0).(*big.Int), args.Error(1)
// }

// func (m *MockEthClient) SendTransaction(ctx context.Context, tx *types.Transaction) error {
// 	args := m.Called(ctx, tx)
// 	return args.Error(0)
// }

// func (m *MockEthClient) CodeAt(ctx context.Context, contract ethcommon.Address, blockNumber *big.Int) ([]byte, error) {
// 	args := m.Called(ctx, contract, blockNumber)
// 	return args.Get(0).([]byte), args.Error(1)
// }

// func (m *MockEthClient) CallContract(ctx context.Context, call ethereum.CallMsg, blockNumber *big.Int) ([]byte, error) {
// 	args := m.Called(ctx, call, blockNumber)
// 	return args.Get(0).([]byte), args.Error(1)
// }

// func (m *MockEthClient) HeaderByNumber(ctx context.Context, number *big.Int) (*types.Header, error) {
// 	args := m.Called(ctx, number)
// 	return args.Get(0).(*types.Header), args.Error(1)
// }

// func (m *MockEthClient) TransactionReceipt(ctx context.Context, txHash ethcommon.Hash) (*types.Receipt, error) {
// 	args := m.Called(ctx, txHash)
// 	return args.Get(0).(*types.Receipt), args.Error(1)
// }

// func (m *MockLogger) Infof(format string, args ...interface{}) {
// 	m.Called(format, args)
// }

// func (m *MockLogger) Errorf(format string, args ...interface{}) {
// 	m.Called(format, args)
// }

// func (m *MockLogger) Warnf(format string, args ...interface{}) {
// 	m.Called(format, args)
// }

// func (m *MockValidator) ValidateTimeBasedJob(job *jobtypes.HandleCreateJobData) (bool, error) {
// 	args := m.Called(job)
// 	return args.Bool(0), args.Error(1)
// }

// func (m *MockValidator) ValidateEventBasedJob(job *jobtypes.HandleCreateJobData, ipfsData *jobtypes.IPFSData) (bool, error) {
// 	args := m.Called(job, ipfsData)
// 	return args.Bool(0), args.Error(1)
// }

// func (m *MockValidator) ValidateAndPrepareJob(job *jobtypes.HandleCreateJobData, triggerData *jobtypes.TriggerData) (bool, error) {
// 	args := m.Called(job, triggerData)
// 	return args.Bool(0), args.Error(1)
// }

// func TestExecuteActionWithDynamicArgs(t *testing.T) {
// 	originalKey := os.Getenv("PRIVATE_KEY_CONTROLLER")
// 	if originalKey == "" {
// 		t.Skip("PRIVATE_KEY_CONTROLLER environment variable not set")
// 	}

// 	tests := []struct {
// 		name           string
// 		job            *jobtypes.HandleCreateJobData
// 		setupMocks     func(*MockEthClient, *MockLogger, *MockValidator)
// 		expectedError  bool
// 		expectedResult jobtypes.ActionData
// 		errorContains  string
// 	}{
// 		{
// 			name: "Successful execution with script trigger function",
// 			job: &jobtypes.HandleCreateJobData{
// 				JobID:                 1,
// 				TaskDefinitionID:      6,
// 				ScriptTriggerFunction: "https://gateway.lighthouse.storage/ipfs/bafkreiaeuy3fyzaecbh2zolndnebccpnrkpwobigtmugzntnyew5oprb4a",
// 				TargetContractAddress: "0x98a170b9b24aD4f42B6B3630A54517fd7Ff3Ac6d",
// 				TargetFunction:        "addTaskId",
// 				Arguments:             []string{`{"jobId": "2", "taskId": "5"}`},
// 			},
// 			setupMocks: func(mockEth *MockEthClient, mockLogger *MockLogger, mockValidator *MockValidator) {
// 				mockEth.On("PendingNonceAt", mock.Anything, mock.Anything).Return(uint64(1), nil)
// 				mockEth.On("SuggestGasPrice", mock.Anything).Return(big.NewInt(20000000000), nil)
// 				mockEth.On("SendTransaction", mock.Anything, mock.Anything).Return(nil)
// 				mockEth.On("CodeAt", mock.Anything, mock.Anything, mock.Anything).Return([]byte("contract-code"), nil)
// 				mockEth.On("CallContract", mock.Anything, mock.Anything, mock.Anything).Return([]byte(`{"jobId": 2, "taskId": 5}`), nil)
// 				mockEth.On("HeaderByNumber", mock.Anything, mock.Anything).Return(&types.Header{}, nil)
// 				mockEth.On("TransactionReceipt", mock.Anything, mock.Anything).Return(&types.Receipt{
// 					Status:  1,
// 					GasUsed: 21000,
// 				}, nil)

// 				mockLogger.On("Infof", mock.Anything, mock.Anything).Return()
// 				mockValidator.On("ValidateAndPrepareJob", mock.Anything, mock.Anything).Return(true, nil)
// 			},
// 			expectedError: false,
// 			expectedResult: jobtypes.ActionData{
// 				Status:    true,
// 				GasUsed:   "21000",
// 				TaskID:    0,
// 				Timestamp: time.Now().UTC(),
// 			},
// 		},
// 		{
// 			name: "Successful execution with script IPFS URL",
// 			job: &jobtypes.HandleCreateJobData{
// 				JobID:                 2,
// 				TaskDefinitionID:      4,
// 				ScriptIPFSUrl:         "https://gateway.lighthouse.storage/ipfs/bafkreiaeuy3fyzaecbh2zolndnebccpnrkpwobigtmugzntnyew5oprb4a",
// 				TargetContractAddress: "0x98a170b9b24aD4f42B6B3630A54517fd7Ff3Ac6d",
// 				TargetFunction:        "addTaskId",
// 				Arguments:             []string{`{"jobId": "2", "taskId": "5"}`},
// 			},
// 			setupMocks: func(mockEth *MockEthClient, mockLogger *MockLogger, mockValidator *MockValidator) {
// 				mockEth.On("PendingNonceAt", mock.Anything, mock.Anything).Return(uint64(1), nil)
// 				mockEth.On("SuggestGasPrice", mock.Anything).Return(big.NewInt(20000000000), nil)
// 				mockEth.On("SendTransaction", mock.Anything, mock.Anything).Return(nil)
// 				mockEth.On("CodeAt", mock.Anything, mock.Anything, mock.Anything).Return([]byte("contract-code"), nil)
// 				mockEth.On("CallContract", mock.Anything, mock.Anything, mock.Anything).Return([]byte(`{"arguments": ["2", "5"]}`), nil)
// 				mockEth.On("HeaderByNumber", mock.Anything, mock.Anything).Return(&types.Header{}, nil)
// 				mockEth.On("TransactionReceipt", mock.Anything, mock.Anything).Return(&types.Receipt{
// 					Status:  1,
// 					GasUsed: 21000,
// 				}, nil)

// 				mockLogger.On("Infof", mock.Anything, mock.Anything).Return()
// 				mockValidator.On("ValidateAndPrepareJob", mock.Anything, mock.Anything).Return(true, nil)
// 			},
// 			expectedError: false,
// 			expectedResult: jobtypes.ActionData{
// 				Status:    true,
// 				GasUsed:   "21000",
// 				TaskID:    0,
// 				Timestamp: time.Now().UTC(),
// 			},
// 		},
// 		{
// 			name: "Failed execution - Invalid script URL",
// 			job: &jobtypes.HandleCreateJobData{
// 				JobID:                 3,
// 				TaskDefinitionID:      4,
// 				ScriptIPFSUrl:         "invalid-url",
// 				TargetContractAddress: "0x98a170b9b24aD4f42B6B3630A54517fd7Ff3Ac6d",
// 				TargetFunction:        "addTaskId",
// 			},
// 			setupMocks: func(mockEth *MockEthClient, mockLogger *MockLogger, mockValidator *MockValidator) {
// 				mockLogger.On("Errorf", mock.Anything, mock.Anything).Return()
// 				mockValidator.On("ValidateAndPrepareJob", mock.Anything, mock.Anything).Return(true, nil)
// 			},
// 			expectedError: true,
// 			errorContains: "failed to download script",
// 		},
// 		{
// 			name: "Failed execution - Empty arguments",
// 			job: &jobtypes.HandleCreateJobData{
// 				JobID:                 4,
// 				TaskDefinitionID:      4,
// 				TargetContractAddress: "0x98a170b9b24aD4f42B6B3630A54517fd7Ff3Ac6d",
// 				TargetFunction:        "addTaskId",
// 			},
// 			setupMocks: func(mockEth *MockEthClient, mockLogger *MockLogger, mockValidator *MockValidator) {
// 				mockLogger.On("Errorf", mock.Anything, mock.Anything).Return()
// 				mockValidator.On("ValidateAndPrepareJob", mock.Anything, mock.Anything).Return(true, nil)
// 			},
// 			expectedError: true,
// 			errorContains: "no script URL or arguments provided",
// 		},
// 		{
// 			name: "Failed execution - Transaction error",
// 			job: &jobtypes.HandleCreateJobData{
// 				JobID:                 5,
// 				TaskDefinitionID:      4,
// 				ScriptIPFSUrl:         "https://gateway.lighthouse.storage/ipfs/bafkreiaeuy3fyzaecbh2zolndnebccpnrkpwobigtmugzntnyew5oprb4a",
// 				TargetContractAddress: "0x98a170b9b24aD4f42B6B3630A54517fd7Ff3Ac6d",
// 				TargetFunction:        "addTaskId",
// 				Arguments:             []string{`{"jobId": "2", "taskId": "5"}`},
// 			},
// 			setupMocks: func(mockEth *MockEthClient, mockLogger *MockLogger, mockValidator *MockValidator) {
// 				mockEth.On("PendingNonceAt", mock.Anything, mock.Anything).Return(uint64(1), nil)
// 				mockEth.On("SuggestGasPrice", mock.Anything).Return(big.NewInt(20000000000), nil)
// 				mockEth.On("SendTransaction", mock.Anything, mock.Anything).Return(fmt.Errorf("transaction failed"))
// 				mockEth.On("CodeAt", mock.Anything, mock.Anything, mock.Anything).Return([]byte("contract-code"), nil)
// 				mockEth.On("CallContract", mock.Anything, mock.Anything, mock.Anything).Return([]byte(`{"arguments": ["2", "5"]}`), nil)

// 				mockLogger.On("Infof", mock.Anything, mock.Anything).Return()
// 				mockLogger.On("Errorf", mock.Anything, mock.Anything).Return()
// 				mockValidator.On("ValidateAndPrepareJob", mock.Anything, mock.Anything).Return(true, nil)
// 			},
// 			expectedError: true,
// 			errorContains: "transaction failed",
// 		},
// 		{
// 			name: "Failed execution - Invalid JSON arguments",
// 			job: &jobtypes.HandleCreateJobData{
// 				JobID:                 6,
// 				TaskDefinitionID:      4,
// 				ScriptIPFSUrl:         "https://gateway.lighthouse.storage/ipfs/bafkreiaeuy3fyzaecbh2zolndnebccpnrkpwobigtmugzntnyew5oprb4a",
// 				TargetContractAddress: "0x98a170b9b24aD4f42B6B3630A54517fd7Ff3Ac6d",
// 				TargetFunction:        "addTaskId",
// 				Arguments:             []string{`invalid-json`},
// 			},
// 			setupMocks: func(mockEth *MockEthClient, mockLogger *MockLogger, mockValidator *MockValidator) {
// 				mockLogger.On("Errorf", mock.Anything, mock.Anything).Return()
// 				mockValidator.On("ValidateAndPrepareJob", mock.Anything, mock.Anything).Return(true, nil)
// 			},
// 			expectedError: true,
// 			errorContains: "failed to parse argument",
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			mockEth := new(MockEthClient)
// 			mockLogger := new(MockLogger)
// 			mockValidator := new(MockValidator)

// 			tt.setupMocks(mockEth, mockLogger, mockValidator)

// 			executor := &JobExecutor{
// 				ethClient:       mockEth,
// 				etherscanAPIKey: "test-key",
// 				argConverter:    &ArgumentConverter{},
// 				logger:          mockLogger,
// 				validator:       mockValidator,
// 			}

// 			result, err := executor.executeActionWithDynamicArgs(tt.job)

// 			if tt.expectedError {
// 				assert.Error(t, err)
// 				if tt.errorContains != "" {
// 					assert.Contains(t, err.Error(), tt.errorContains)
// 				}
// 			} else {
// 				assert.NoError(t, err)
// 				assert.Equal(t, tt.expectedResult.Status, result.Status)
// 				assert.Equal(t, tt.expectedResult.GasUsed, result.GasUsed)
// 				assert.NotEmpty(t, result.ActionTxHash)
// 			}

// 			mockEth.AssertExpectations(t)
// 			mockLogger.AssertExpectations(t)
// 			mockValidator.AssertExpectations(t)
// 		})
// 	}

// 	os.Unsetenv("PRIVATE_KEY_CONTROLLER")
// }

// func TestProcessArguments(t *testing.T) {
// 	executor := &JobExecutor{
// 		argConverter: &ArgumentConverter{},
// 	}

// 	tests := []struct {
// 		name          string
// 		args          interface{}
// 		methodInputs  []abi.Argument
// 		expectedError bool
// 	}{
// 		{
// 			name: "Test with simple string array",
// 			args: []string{"123", "test"},
// 			methodInputs: []abi.Argument{
// 				{Type: abi.Type{T: abi.UintTy, Size: 256}},
// 				{Type: abi.Type{T: abi.StringTy}},
// 			},
// 			expectedError: false,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			_, err := executor.processArguments(tt.args, tt.methodInputs, nil)
// 			if tt.expectedError {
// 				assert.Error(t, err)
// 			} else {
// 				assert.NoError(t, err)
// 			}
// 		})
// 	}
// }

// func TestIntegration(t *testing.T) {
// 	if os.Getenv("INTEGRATION_TEST") != "true" {
// 		t.Skip("Skipping integration test")
// 	}

// 	client, err := ethclient.Dial("https://eth-holesky.g.alchemy.com/v2/E3OSaENxCMNoRBi_quYcmTNPGfRitxQa")
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	executor := NewJobExecutor(client, os.Getenv("ETHERSCAN_API_KEY"))

// 	job := &jobtypes.HandleCreateJobData{
// 		JobID:                 1,
// 		TaskDefinitionID:      4,
// 		ScriptIPFSUrl:         "https://gateway.lighthouse.storage/ipfs/bafkreiaeuy3fyzaecbh2zolndnebccpnrkpwobigtmugzntnyew5oprb4a",
// 		TargetContractAddress: "0x98a170b9b24aD4f42B6B3630A54517fd7Ff3Ac6d",
// 		TargetFunction:        "addTaskId",
// 	}

// 	result, err := executor.executeActionWithDynamicArgs(job)
// 	assert.NoError(t, err)
// 	assert.True(t, result.Status)
// }
