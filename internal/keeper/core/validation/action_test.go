package validation

// import (
// 	"math/big"
// 	"testing"
// 	"time"

// 	"github.com/ethereum/go-ethereum/common"
// 	ethTypes "github.com/ethereum/go-ethereum/core/types"
// 	"github.com/stretchr/testify/mock"
// 	"github.com/trigg3rX/triggerx-backend-imua/pkg/types"
// )

// func TestValidateAction(t *testing.T) {
// 	tests := []struct {
// 		name           string
// 		targetData     *types.TaskTargetData
// 		actionData     *types.PerformerActionData
// 		setupMock      func(*MockEthClient)
// 		expectedResult bool
// 		expectedError  string
// 	}{
// 		{
// 			name: "Valid_Time_Based_Action",
// 			targetData: &types.TaskTargetData{
// 				TaskDefinitionID:       1,
// 				NextExecutionTimestamp: time.Now(),
// 			},
// 			actionData: &types.PerformerActionData{
// 				ActionTxHash: "0x1234567890123456789012345678901234567890123456789012345678901234",
// 			},
// 			setupMock: func(mockClient *MockEthClient) {
// 				receipt := &ethTypes.Receipt{
// 					Status:      1,
// 					BlockNumber: big.NewInt(1),
// 					BlockHash:   common.HexToHash("0xabcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890"),
// 				}
// 				mockClient.On("TransactionReceipt", mock.Anything, mock.Anything).Return(receipt, nil)

// 				header := &ethTypes.Header{
// 					Time: uint64(time.Now().Unix()),
// 				}
// 				body := &ethTypes.Body{
// 					Transactions: []*ethTypes.Transaction{},
// 					Uncles:       []*ethTypes.Header{},
// 				}
// 				block := ethTypes.NewBlock(header, body, nil, nil)
// 				mockClient.On("BlockByHash", mock.Anything, mock.Anything).Return(block, nil)
// 			},
// 			expectedResult: true,
// 			expectedError:  "",
// 		},
// 		{
// 			name: "Pending_Transaction",
// 			targetData: &types.TaskTargetData{
// 				TaskDefinitionID:       1,
// 				NextExecutionTimestamp: time.Now(),
// 			},
// 			actionData: &types.PerformerActionData{
// 				ActionTxHash: "0x1234567890123456789012345678901234567890123456789012345678901234",
// 			},
// 			setupMock: func(mockClient *MockEthClient) {
// 				mockClient.On("TransactionReceipt", mock.Anything, mock.Anything).Return(nil, nil)
// 				mockClient.On("TransactionByHash", mock.Anything, mock.Anything).Return(nil, true, nil)
// 			},
// 			expectedResult: false,
// 			expectedError:  "transaction is pending",
// 		},
// 		{
// 			name: "Failed_Transaction",
// 			targetData: &types.TaskTargetData{
// 				TaskDefinitionID:       1,
// 				NextExecutionTimestamp: time.Now(),
// 			},
// 			actionData: &types.PerformerActionData{
// 				ActionTxHash: "0x1234567890123456789012345678901234567890123456789012345678901234",
// 			},
// 			setupMock: func(mockClient *MockEthClient) {
// 				receipt := &ethTypes.Receipt{
// 					Status:      0,
// 					BlockNumber: big.NewInt(1),
// 					BlockHash:   common.HexToHash("0xabcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890"),
// 				}
// 				mockClient.On("TransactionReceipt", mock.Anything, mock.Anything).Return(receipt, nil)
// 			},
// 			expectedResult: false,
// 			expectedError:  "transaction is not successful",
// 		},
// 		{
// 			name: "Late_Execution",
// 			targetData: &types.TaskTargetData{
// 				TaskDefinitionID:       1,
// 				NextExecutionTimestamp: time.Now().Add(-2 * time.Second),
// 			},
// 			actionData: &types.PerformerActionData{
// 				ActionTxHash: "0x1234567890123456789012345678901234567890123456789012345678901234",
// 			},
// 			setupMock: func(mockClient *MockEthClient) {
// 				receipt := &ethTypes.Receipt{
// 					Status:      1,
// 					BlockNumber: big.NewInt(1),
// 					BlockHash:   common.HexToHash("0xabcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890"),
// 				}
// 				mockClient.On("TransactionReceipt", mock.Anything, mock.Anything).Return(receipt, nil)

// 				header := &ethTypes.Header{
// 					Time: uint64(time.Now().Unix()),
// 				}
// 				body := &ethTypes.Body{
// 					Transactions: []*ethTypes.Transaction{},
// 					Uncles:       []*ethTypes.Header{},
// 				}
// 				block := ethTypes.NewBlock(header, body, nil, nil)
// 				mockClient.On("BlockByHash", mock.Anything, mock.Anything).Return(block, nil)
// 			},
// 			expectedResult: false,
// 			expectedError:  "transaction was made after the next execution timestamp",
// 		},
// 		{
// 			name: "Early_Execution",
// 			targetData: &types.TaskTargetData{
// 				TaskDefinitionID:       1,
// 				NextExecutionTimestamp: time.Now().Add(2 * time.Second),
// 			},
// 			actionData: &types.PerformerActionData{
// 				ActionTxHash: "0x1234567890123456789012345678901234567890123456789012345678901234",
// 			},
// 			setupMock: func(mockClient *MockEthClient) {
// 				receipt := &ethTypes.Receipt{
// 					Status:      1,
// 					BlockNumber: big.NewInt(1),
// 					BlockHash:   common.HexToHash("0xabcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890"),
// 				}
// 				mockClient.On("TransactionReceipt", mock.Anything, mock.Anything).Return(receipt, nil)

// 				header := &ethTypes.Header{
// 					Time: uint64(time.Now().Unix()),
// 				}
// 				body := &ethTypes.Body{
// 					Transactions: []*ethTypes.Transaction{},
// 					Uncles:       []*ethTypes.Header{},
// 				}
// 				block := ethTypes.NewBlock(header, body, nil, nil)
// 				mockClient.On("BlockByHash", mock.Anything, mock.Anything).Return(block, nil)
// 			},
// 			expectedResult: false,
// 			expectedError:  "transaction was made before the next execution timestamp",
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			// Setup mock client
// 			mockClient := new(MockEthClient)
// 			tt.setupMock(mockClient)

// 			// Setup mock logger
// 			mockLogger := new(MockLogger)

// 			// Create validator
// 			validator := NewTaskValidator(mockLogger)

// 			// Execute test
// 			result, err := validator.ValidateAction(tt.targetData, tt.actionData, mockClient, "test-trace-id")

// 			// Verify results
// 			if tt.expectedError != "" {
// 				if err == nil {
// 					t.Errorf("expected error %q, got nil", tt.expectedError)
// 				} else if err.Error() != tt.expectedError {
// 					t.Errorf("expected error %q, got %q", tt.expectedError, err.Error())
// 				}
// 			} else if err != nil {
// 				t.Errorf("unexpected error: %v", err)
// 			}

// 			if result != tt.expectedResult {
// 				t.Errorf("expected result %v, got %v", tt.expectedResult, result)
// 			}

// 			// Verify all mock expectations were met
// 			mockClient.AssertExpectations(t)
// 			mockLogger.AssertExpectations(t)
// 		})
// 	}
// }
