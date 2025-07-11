package handlers

// import (
// 	"bytes"
// 	"context"
// 	"encoding/hex"
// 	"encoding/json"
// 	"fmt"
// 	"io"
// 	"net/http"
// 	"net/http/httptest"
// 	"testing"

// 	"github.com/gin-gonic/gin"
// 	"github.com/stretchr/testify/assert"
// 	"github.com/stretchr/testify/mock"

// 	"github.com/trigg3rX/triggerx-backend-imua/internal/keeper/config"
// 	execution "github.com/trigg3rX/triggerx-backend-imua/internal/keeper/core/execution"
// 	validation "github.com/trigg3rX/triggerx-backend-imua/internal/keeper/core/validation"
// 	"github.com/trigg3rX/triggerx-backend-imua/pkg/logging"
// 	"github.com/trigg3rX/triggerx-backend-imua/pkg/types"
// )

// // MockLogger is a mock implementation of the Logger interface
// type MockLogger struct {
// 	mock.Mock
// }

// func (m *MockLogger) Debug(msg string, keysAndValues ...interface{}) {
// 	m.Called(msg, keysAndValues)
// }

// func (m *MockLogger) Info(msg string, keysAndValues ...interface{}) {
// 	m.Called(msg, keysAndValues)
// }

// func (m *MockLogger) Warn(msg string, keysAndValues ...interface{}) {
// 	m.Called(msg, keysAndValues)
// }

// func (m *MockLogger) Error(msg string, keysAndValues ...interface{}) {
// 	m.Called(msg, keysAndValues)
// }

// func (m *MockLogger) Fatal(msg string, keysAndValues ...interface{}) {
// 	m.Called(msg, keysAndValues)
// }

// func (m *MockLogger) Debugf(template string, args ...interface{}) {
// 	m.Called(template, args)
// }

// func (m *MockLogger) Infof(template string, args ...interface{}) {
// 	m.Called(template, args)
// }

// func (m *MockLogger) Warnf(template string, args ...interface{}) {
// 	m.Called(template, args)
// }

// func (m *MockLogger) Errorf(template string, args ...interface{}) {
// 	m.Called(template, args)
// }

// func (m *MockLogger) Fatalf(template string, args ...interface{}) {
// 	m.Called(template, args)
// }

// func (m *MockLogger) With(keysAndValues ...interface{}) logging.Logger {
// 	args := m.Called(keysAndValues)
// 	return args.Get(0).(logging.Logger)
// }

// // MockTaskExecutor is a mock implementation of the TaskExecutorInterface
// type MockTaskExecutor struct {
// 	executionResult bool
// 	executionError  error
// }

// func (m *MockTaskExecutor) ExecuteTask(ctx context.Context, task *types.SendTaskDataToKeeper, traceID string) (bool, error) {
// 	fmt.Println("[DEBUG] MockTaskExecutor.ExecuteTask called")
// 	return m.executionResult, m.executionError
// }

// // Update setupTestHandler to use the mock executor for specific test cases
// func setupTestHandler(useMockExecutor bool, mockResult bool, mockError error) (*TaskHandler, TaskExecutorInterface, *validation.TaskValidator, *MockLogger) {
// 	mockLogger := &MockLogger{}

// 	var executor TaskExecutorInterface
// 	var validator *validation.TaskValidator
// 	if useMockExecutor {
// 		mockExec := &MockTaskExecutor{
// 			executionResult: mockResult,
// 			executionError:  mockError,
// 		}
// 		executor = mockExec
// 		validator = validation.NewTaskValidator(mockLogger)
// 		handler := NewTaskHandler(mockLogger, executor, validator)
// 		fmt.Printf("[DEBUG] handler.executor (mock): %T, %v\n", handler.executor, handler.executor)
// 		if handler.executor == nil {
// 			panic("executor is nil in handler setup (mock)")
// 		}
// 		return handler, executor, validator, mockLogger
// 	}

// 	executor = execution.NewTaskExecutor("", nil, nil, nil, mockLogger)
// 	validator = validation.NewTaskValidator(mockLogger)
// 	handler := NewTaskHandler(mockLogger, executor, validator)
// 	if handler.executor == nil {
// 		panic("executor is nil in handler setup (real)")
// 	}
// 	return handler, executor, validator, mockLogger
// }

// func TestExecuteTask(t *testing.T) {
// 	tests := []struct {
// 		name            string
// 		method          string
// 		requestBody     interface{}
// 		keeperAddress   string
// 		expectedStatus  int
// 		expectedBody    map[string]interface{}
// 		setupMock       func(TaskExecutorInterface, *validation.TaskValidator, *MockLogger)
// 		useMockExecutor bool
// 		mockResult      bool
// 		mockError       error
// 	}{
// 		{
// 			name:   "Invalid HTTP Method",
// 			method: http.MethodGet,
// 			requestBody: map[string]string{
// 				"data": "0x123",
// 			},
// 			expectedStatus: http.StatusMethodNotAllowed,
// 			expectedBody: map[string]interface{}{
// 				"error": "Invalid method",
// 			},
// 			setupMock: func(executor TaskExecutorInterface, validator *validation.TaskValidator, logger *MockLogger) {
// 				logger.On("Info", "Executing task ...", []interface{}{"trace_id", ""}).Return()
// 			},
// 		},
// 		{
// 			name:           "Invalid JSON Body",
// 			method:         http.MethodPost,
// 			requestBody:    "invalid json",
// 			expectedStatus: http.StatusBadRequest,
// 			expectedBody: map[string]interface{}{
// 				"error": "Invalid JSON body",
// 			},
// 			setupMock: func(executor TaskExecutorInterface, validator *validation.TaskValidator, logger *MockLogger) {
// 				logger.On("Info", "Executing task ...", []interface{}{"trace_id", ""}).Return()
// 			},
// 		},
// 		{
// 			name:   "Invalid Hex Data",
// 			method: http.MethodPost,
// 			requestBody: map[string]string{
// 				"data": "invalid hex",
// 			},
// 			expectedStatus: http.StatusBadRequest,
// 			expectedBody: map[string]interface{}{
// 				"error": "Invalid hex data",
// 			},
// 			setupMock: func(executor TaskExecutorInterface, validator *validation.TaskValidator, logger *MockLogger) {
// 				logger.On("Info", "Executing task ...", []interface{}{"trace_id", ""}).Return()
// 			},
// 		},
// 		{
// 			name:   "Not the Performer",
// 			method: http.MethodPost,
// 			requestBody: func() map[string]string {
// 				data := types.SendTaskDataToKeeper{
// 					PerformerData: types.GetPerformerData{
// 						KeeperAddress: "0xDifferentAddress",
// 					},
// 				}
// 				jsonData, _ := json.Marshal(data)
// 				return map[string]string{
// 					"data": "0x" + hex.EncodeToString(jsonData),
// 				}
// 			}(),
// 			keeperAddress:  "0xCurrentAddress",
// 			expectedStatus: http.StatusOK,
// 			expectedBody: map[string]interface{}{
// 				"message": "I am not the performer",
// 			},
// 			setupMock: func(executor TaskExecutorInterface, validator *validation.TaskValidator, logger *MockLogger) {
// 				logger.On("Info", "Executing task ...", []interface{}{"trace_id", ""}).Return()
// 				logger.On("Infof", "I am not the performer: %s", []interface{}{"0xDifferentAddress"}).Return()
// 			},
// 		},
// 		{
// 			name:   "Successful Task Execution",
// 			method: http.MethodPost,
// 			requestBody: func() map[string]string {
// 				data := types.SendTaskDataToKeeper{
// 					PerformerData: types.GetPerformerData{
// 						KeeperAddress: "0xCurrentAddress",
// 					},
// 					TargetData: &types.TaskTargetData{
// 						TaskID:           1,
// 						TaskDefinitionID: 1,
// 						TargetChainID:    "chain-1",
// 					},
// 					TriggerData: &types.TaskTriggerData{},
// 				}
// 				jsonData, _ := json.Marshal(data)
// 				return map[string]string{
// 					"data": "0x" + hex.EncodeToString(jsonData),
// 				}
// 			}(),
// 			keeperAddress:  "0xCurrentAddress",
// 			expectedStatus: http.StatusOK,
// 			expectedBody: map[string]interface{}{
// 				"success": "true",
// 			},
// 			useMockExecutor: true,
// 			mockResult:      true,
// 			mockError:       nil,
// 			setupMock: func(executor TaskExecutorInterface, validator *validation.TaskValidator, logger *MockLogger) {
// 				logger.On("Info", "Executing task ...", []interface{}{"trace_id", ""}).Return()
// 				logger.On("Infof", "I am the performer: %s", []interface{}{"0xCurrentAddress"}).Return()
// 				logger.On("Info", "Execution starts for task ID: ", []interface{}{"task_id", int64(1), "trace_id", ""}).Return()
// 				logger.On("Infof", "Task Definition ID: %d | Target Chain ID: %s", []interface{}{int(1), "chain-1"}).Return()
// 				logger.On("Info", "Task execution completed", []interface{}{"success", true, "trace_id", ""}).Return()
// 			},
// 		},
// 		{
// 			name:   "Execution Failure",
// 			method: http.MethodPost,
// 			requestBody: func() map[string]string {
// 				data := types.SendTaskDataToKeeper{
// 					PerformerData: types.GetPerformerData{
// 						KeeperAddress: "0xCurrentAddress",
// 					},
// 					TargetData: &types.TaskTargetData{
// 						TaskID:           1,
// 						TaskDefinitionID: 1,
// 						TargetChainID:    "chain-1",
// 					},
// 					TriggerData: &types.TaskTriggerData{},
// 				}
// 				jsonData, _ := json.Marshal(data)
// 				return map[string]string{
// 					"data": "0x" + hex.EncodeToString(jsonData),
// 				}
// 			}(),
// 			keeperAddress:  "0xCurrentAddress",
// 			expectedStatus: http.StatusInternalServerError,
// 			expectedBody: map[string]interface{}{
// 				"error": "Task execution failed",
// 			},
// 			useMockExecutor: true,
// 			mockResult:      false,
// 			mockError:       fmt.Errorf("execution failed"),
// 			setupMock: func(executor TaskExecutorInterface, validator *validation.TaskValidator, logger *MockLogger) {
// 				logger.On("Info", "Executing task ...", []interface{}{"trace_id", ""}).Return()
// 				logger.On("Infof", "I am the performer: %s", []interface{}{"0xCurrentAddress"}).Return()
// 				logger.On("Info", "Execution starts for task ID: ", []interface{}{"task_id", int64(1), "trace_id", ""}).Return()
// 				logger.On("Infof", "Task Definition ID: %d | Target Chain ID: %s", []interface{}{int(1), "chain-1"}).Return()
// 				logger.On("Error", "Task execution failed", mock.Anything).Return()
// 			},
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			if tt.keeperAddress != "" {
// 				config.SetKeeperAddress(tt.keeperAddress)
// 			}
// 			handler, executor, validator, mockLogger := setupTestHandler(tt.useMockExecutor, tt.mockResult, tt.mockError)
// 			if tt.setupMock != nil {
// 				tt.setupMock(executor, validator, mockLogger)
// 			}

// 			// Create a new Gin context
// 			w := httptest.NewRecorder()
// 			c, _ := gin.CreateTestContext(w)
// 			c.Request = httptest.NewRequest(tt.method, "/", nil)

// 			// Set the request body
// 			if tt.requestBody != nil {
// 				jsonBody, _ := json.Marshal(tt.requestBody)
// 				c.Request.Body = io.NopCloser(bytes.NewBuffer(jsonBody))
// 			}

// 			// Call the handler
// 			handler.ExecuteTask(c)

// 			// Assert the response
// 			assert.Equal(t, tt.expectedStatus, w.Code)
// 			var response map[string]interface{}
// 			err := json.Unmarshal(w.Body.Bytes(), &response)
// 			assert.NoError(t, err)
// 			assert.Equal(t, tt.expectedBody, response)
// 		})
// 	}
// }
