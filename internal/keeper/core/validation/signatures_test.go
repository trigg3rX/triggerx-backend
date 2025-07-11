package validation

// import (
// 	"testing"

// 	"github.com/stretchr/testify/mock"
// 	"github.com/trigg3rX/triggerx-backend-imua/pkg/types"
// )

// // MockCryptography is a mock for the cryptography package
// type MockCryptography struct {
// 	mock.Mock
// }

// func (m *MockCryptography) VerifySignatureFromJSON(jsonData interface{}, signature string, signerAddress string) (bool, error) {
// 	args := m.Called(jsonData, signature, signerAddress)
// 	return args.Bool(0), args.Error(1)
// }

// func TestValidateSchedulerSignature(t *testing.T) {
// 	tests := []struct {
// 		name           string
// 		task           *types.SendTaskDataToKeeper
// 		setupMock      func(*MockLogger, *MockCryptography)
// 		expectedResult bool
// 		expectedError  string
// 	}{
// 		{
// 			name: "Valid_Scheduler_Signature",
// 			task: &types.SendTaskDataToKeeper{
// 				TaskID: 1,
// 				PerformerData: types.GetPerformerData{
// 					KeeperID:      1,
// 					KeeperAddress: "0x1234567890123456789012345678901234567890",
// 				},
// 				TargetData: &types.TaskTargetData{
// 					TaskID:           1,
// 					TaskDefinitionID: 1,
// 				},
// 				TriggerData: &types.TaskTriggerData{
// 					TaskID:           1,
// 					TaskDefinitionID: 1,
// 				},
// 				SchedulerSignature: &types.SchedulerSignatureData{
// 					TaskID:                  1,
// 					SchedulerSigningAddress: "0xabcdef1234567890abcdef1234567890abcdef12",
// 					SchedulerSignature:      "valid_signature_here",
// 				},
// 			},
// 			setupMock: func(mockLogger *MockLogger, mockCrypto *MockCryptography) {
// 				mockLogger.On("With", mock.Anything).Return(mockLogger)
// 				mockLogger.On("Info", mock.Anything, mock.Anything).Return()
// 				mockCrypto.On("VerifySignatureFromJSON", mock.Anything, "valid_signature_here", "0xabcdef1234567890abcdef1234567890abcdef12").Return(true, nil)
// 			},
// 			expectedResult: true,
// 			expectedError:  "",
// 		},
// 		{
// 			name: "Missing_Scheduler_Signature",
// 			task: &types.SendTaskDataToKeeper{
// 				TaskID: 1,
// 				PerformerData: types.GetPerformerData{
// 					KeeperID:      1,
// 					KeeperAddress: "0x1234567890123456789012345678901234567890",
// 				},
// 				TargetData: &types.TaskTargetData{
// 					TaskID:           1,
// 					TaskDefinitionID: 1,
// 				},
// 				TriggerData: &types.TaskTriggerData{
// 					TaskID:           1,
// 					TaskDefinitionID: 1,
// 				},
// 				SchedulerSignature: nil,
// 			},
// 			setupMock: func(mockLogger *MockLogger, mockCrypto *MockCryptography) {
// 				mockLogger.On("With", mock.Anything).Return(mockLogger)
// 				mockLogger.On("Error", mock.Anything, mock.Anything).Return()
// 			},
// 			expectedResult: false,
// 			expectedError:  "scheduler signature data is missing",
// 		},
// 		{
// 			name: "Empty_Scheduler_Signature",
// 			task: &types.SendTaskDataToKeeper{
// 				TaskID: 1,
// 				PerformerData: types.GetPerformerData{
// 					KeeperID:      1,
// 					KeeperAddress: "0x1234567890123456789012345678901234567890",
// 				},
// 				TargetData: &types.TaskTargetData{
// 					TaskID:           1,
// 					TaskDefinitionID: 1,
// 				},
// 				TriggerData: &types.TaskTriggerData{
// 					TaskID:           1,
// 					TaskDefinitionID: 1,
// 				},
// 				SchedulerSignature: &types.SchedulerSignatureData{
// 					TaskID:                  1,
// 					SchedulerSigningAddress: "0xabcdef1234567890abcdef1234567890abcdef12",
// 					SchedulerSignature:      "",
// 				},
// 			},
// 			setupMock: func(mockLogger *MockLogger, mockCrypto *MockCryptography) {
// 				mockLogger.On("With", mock.Anything).Return(mockLogger)
// 				mockLogger.On("Error", mock.Anything, mock.Anything).Return()
// 			},
// 			expectedResult: false,
// 			expectedError:  "scheduler signature is empty",
// 		},
// 		{
// 			name: "Empty_Scheduler_Signing_Address",
// 			task: &types.SendTaskDataToKeeper{
// 				TaskID: 1,
// 				PerformerData: types.GetPerformerData{
// 					KeeperID:      1,
// 					KeeperAddress: "0x1234567890123456789012345678901234567890",
// 				},
// 				TargetData: &types.TaskTargetData{
// 					TaskID:           1,
// 					TaskDefinitionID: 1,
// 				},
// 				TriggerData: &types.TaskTriggerData{
// 					TaskID:           1,
// 					TaskDefinitionID: 1,
// 				},
// 				SchedulerSignature: &types.SchedulerSignatureData{
// 					TaskID:                  1,
// 					SchedulerSigningAddress: "",
// 					SchedulerSignature:      "valid_signature_here",
// 				},
// 			},
// 			setupMock: func(mockLogger *MockLogger, mockCrypto *MockCryptography) {
// 				mockLogger.On("With", mock.Anything).Return(mockLogger)
// 				mockLogger.On("Error", mock.Anything, mock.Anything).Return()
// 			},
// 			expectedResult: false,
// 			expectedError:  "scheduler signing address is empty",
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			// Setup mock logger and cryptography
// 			mockLogger := new(MockLogger)
// 			mockCrypto := new(MockCryptography)
// 			tt.setupMock(mockLogger, mockCrypto)

// 			// Create validator with mocked cryptography
// 			validator := NewTaskValidator(mockLogger)
// 			validator.crypto = mockCrypto

// 			// Execute test
// 			result, err := validator.ValidateSchedulerSignature(tt.task, "test-trace-id")

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
// 			mockLogger.AssertExpectations(t)
// 			mockCrypto.AssertExpectations(t)
// 		})
// 	}
// }

// func TestValidatePerformerSignature(t *testing.T) {
// 	tests := []struct {
// 		name           string
// 		ipfsData       types.IPFSData
// 		setupMock      func(*MockLogger, *MockCryptography)
// 		expectedResult bool
// 		expectedError  string
// 	}{
// 		{
// 			name: "Valid_Performer_Signature",
// 			ipfsData: types.IPFSData{
// 				TaskData: &types.SendTaskDataToKeeper{
// 					TaskID: 1,
// 					PerformerData: types.GetPerformerData{
// 						KeeperID:      1,
// 						KeeperAddress: "0x1234567890123456789012345678901234567890",
// 					},
// 				},
// 				ActionData: &types.PerformerActionData{
// 					TaskID: 1,
// 				},
// 				ProofData: &types.ProofData{
// 					TaskID: 1,
// 				},
// 				PerformerSignature: &types.PerformerSignatureData{
// 					PerformerSigningAddress: "0x1234567890123456789012345678901234567890",
// 					PerformerSignature:      "valid_signature_here",
// 				},
// 			},
// 			setupMock: func(mockLogger *MockLogger, mockCrypto *MockCryptography) {
// 				mockLogger.On("With", mock.Anything).Return(mockLogger)
// 				mockLogger.On("Info", mock.Anything, mock.Anything).Return()
// 				mockCrypto.On("VerifySignatureFromJSON", mock.Anything, "valid_signature_here", "0x1234567890123456789012345678901234567890").Return(true, nil)
// 			},
// 			expectedResult: true,
// 			expectedError:  "",
// 		},
// 		{
// 			name: "Missing_Performer_Signature",
// 			ipfsData: types.IPFSData{
// 				TaskData: &types.SendTaskDataToKeeper{
// 					TaskID: 1,
// 					PerformerData: types.GetPerformerData{
// 						KeeperID:      1,
// 						KeeperAddress: "0x1234567890123456789012345678901234567890",
// 					},
// 				},
// 				ActionData: &types.PerformerActionData{
// 					TaskID: 1,
// 				},
// 				ProofData: &types.ProofData{
// 					TaskID: 1,
// 				},
// 				PerformerSignature: nil,
// 			},
// 			setupMock: func(mockLogger *MockLogger, mockCrypto *MockCryptography) {
// 				mockLogger.On("With", mock.Anything).Return(mockLogger)
// 				mockLogger.On("Error", mock.Anything, mock.Anything).Return()
// 			},
// 			expectedResult: false,
// 			expectedError:  "performer signature data is missing",
// 		},
// 		{
// 			name: "Empty_Performer_Signature",
// 			ipfsData: types.IPFSData{
// 				TaskData: &types.SendTaskDataToKeeper{
// 					TaskID: 1,
// 					PerformerData: types.GetPerformerData{
// 						KeeperID:      1,
// 						KeeperAddress: "0x1234567890123456789012345678901234567890",
// 					},
// 				},
// 				ActionData: &types.PerformerActionData{
// 					TaskID: 1,
// 				},
// 				ProofData: &types.ProofData{
// 					TaskID: 1,
// 				},
// 				PerformerSignature: &types.PerformerSignatureData{
// 					PerformerSigningAddress: "0x1234567890123456789012345678901234567890",
// 					PerformerSignature:      "",
// 				},
// 			},
// 			setupMock: func(mockLogger *MockLogger, mockCrypto *MockCryptography) {
// 				mockLogger.On("With", mock.Anything).Return(mockLogger)
// 				mockLogger.On("Error", mock.Anything, mock.Anything).Return()
// 			},
// 			expectedResult: false,
// 			expectedError:  "performer signature is empty",
// 		},
// 		{
// 			name: "Empty_Performer_Signing_Address",
// 			ipfsData: types.IPFSData{
// 				TaskData: &types.SendTaskDataToKeeper{
// 					TaskID: 1,
// 					PerformerData: types.GetPerformerData{
// 						KeeperID:      1,
// 						KeeperAddress: "0x1234567890123456789012345678901234567890",
// 					},
// 				},
// 				ActionData: &types.PerformerActionData{
// 					TaskID: 1,
// 				},
// 				ProofData: &types.ProofData{
// 					TaskID: 1,
// 				},
// 				PerformerSignature: &types.PerformerSignatureData{
// 					PerformerSigningAddress: "",
// 					PerformerSignature:      "valid_signature_here",
// 				},
// 			},
// 			setupMock: func(mockLogger *MockLogger, mockCrypto *MockCryptography) {
// 				mockLogger.On("With", mock.Anything).Return(mockLogger)
// 				mockLogger.On("Error", mock.Anything, mock.Anything).Return()
// 			},
// 			expectedResult: false,
// 			expectedError:  "performer signing address is empty",
// 		},
// 		{
// 			name: "Mismatched_Performer_Address",
// 			ipfsData: types.IPFSData{
// 				TaskData: &types.SendTaskDataToKeeper{
// 					TaskID: 1,
// 					PerformerData: types.GetPerformerData{
// 						KeeperID:      1,
// 						KeeperAddress: "0x1234567890123456789012345678901234567890",
// 					},
// 				},
// 				ActionData: &types.PerformerActionData{
// 					TaskID: 1,
// 				},
// 				ProofData: &types.ProofData{
// 					TaskID: 1,
// 				},
// 				PerformerSignature: &types.PerformerSignatureData{
// 					PerformerSigningAddress: "0xabcdef1234567890abcdef1234567890abcdef12",
// 					PerformerSignature:      "valid_signature_here",
// 				},
// 			},
// 			setupMock: func(mockLogger *MockLogger, mockCrypto *MockCryptography) {
// 				mockLogger.On("With", mock.Anything).Return(mockLogger)
// 				mockLogger.On("Error", mock.Anything, mock.Anything).Return()
// 			},
// 			expectedResult: false,
// 			expectedError:  "performer signing address does not match the assigned performer",
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			// Setup mock logger and cryptography
// 			mockLogger := new(MockLogger)
// 			mockCrypto := new(MockCryptography)
// 			tt.setupMock(mockLogger, mockCrypto)

// 			// Create validator with mocked cryptography
// 			validator := NewTaskValidator(mockLogger)
// 			validator.crypto = mockCrypto

// 			// Execute test
// 			result, err := validator.ValidatePerformerSignature(tt.ipfsData, "test-trace-id")

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
// 			mockLogger.AssertExpectations(t)
// 			mockCrypto.AssertExpectations(t)
// 		})
// 	}
// }
