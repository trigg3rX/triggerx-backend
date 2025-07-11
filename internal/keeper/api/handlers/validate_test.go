package handlers

// import (
// 	"bytes"
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

// 	"github.com/trigg3rX/triggerx-backend-imua/internal/keeper/core/validation"
// 	"github.com/trigg3rX/triggerx-backend-imua/pkg/types"
// )

// // MockTaskValidator is a mock implementation of the TaskValidatorInterface
// type MockTaskValidator struct {
// 	mock.Mock
// }

// func (m *MockTaskValidator) ValidateTask(task *types.SendTaskDataToKeeper, traceID string) (bool, error) {
// 	args := m.Called(task, traceID)
// 	return args.Bool(0), args.Error(1)
// }

// func (m *MockTaskValidator) ValidateSchedulerSignature(task *types.SendTaskDataToKeeper, traceID string) (bool, error) {
// 	args := m.Called(task, traceID)
// 	return args.Bool(0), args.Error(1)
// }

// func (m *MockTaskValidator) ValidateTrigger(triggerData *types.TaskTriggerData, traceID string) (bool, error) {
// 	args := m.Called(triggerData, traceID)
// 	return args.Bool(0), args.Error(1)
// }

// func (m *MockTaskValidator) ValidateAction(targetData *types.TaskTargetData, actionData *types.PerformerActionData, client validation.EthClientInterface, traceID string) (bool, error) {
// 	args := m.Called(targetData, actionData, client, traceID)
// 	return args.Bool(0), args.Error(1)
// }

// func (m *MockTaskValidator) ValidateProof(ipfsData types.IPFSData, traceID string) (bool, error) {
// 	args := m.Called(ipfsData, traceID)
// 	return args.Bool(0), args.Error(1)
// }

// func (m *MockTaskValidator) ValidatePerformerSignature(ipfsData types.IPFSData, traceID string) (bool, error) {
// 	args := m.Called(ipfsData, traceID)
// 	return args.Bool(0), args.Error(1)
// }

// // MockIPFSFetcher is a mock implementation of the IPFS content fetcher
// type MockIPFSFetcher struct {
// 	mock.Mock
// }

// func (m *MockIPFSFetcher) FetchContent(cid string) (types.IPFSData, error) {
// 	args := m.Called(cid)
// 	return args.Get(0).(types.IPFSData), args.Error(1)
// }

// func setupValidateTestHandler() (*TaskHandler, *MockTaskValidator, *MockLogger, *MockIPFSFetcher) {
// 	mockLogger := &MockLogger{}
// 	mockValidator := &MockTaskValidator{}
// 	mockIPFSFetcher := &MockIPFSFetcher{}
// 	handler := NewTaskHandler(mockLogger, nil, mockValidator)
// 	handler.ipfsFetcher = mockIPFSFetcher
// 	return handler, mockValidator, mockLogger, mockIPFSFetcher
// }

// func TestValidateTask(t *testing.T) {
// 	tests := []struct {
// 		name           string
// 		method         string
// 		requestBody    interface{}
// 		expectedStatus int
// 		expectedBody   map[string]interface{}
// 		setupMock      func(*MockTaskValidator, *MockLogger, *MockIPFSFetcher)
// 	}{
// 		{
// 			name:   "Invalid HTTP Method",
// 			method: http.MethodGet,
// 			requestBody: map[string]string{
// 				"data": "0x123",
// 			},
// 			expectedStatus: http.StatusBadRequest,
// 			expectedBody: map[string]interface{}{
// 				"error":   true,
// 				"data":    false,
// 				"message": "Failed to decode hex data",
// 			},
// 			setupMock: func(validator *MockTaskValidator, logger *MockLogger, ipfsFetcher *MockIPFSFetcher) {
// 				logger.On("Info", "Validating task ...", []interface{}{"trace_id", ""}).Return()
// 				logger.On("Errorf", "Failed to hex-decode data: %v", mock.Anything).Return()
// 			},
// 		},
// 		{
// 			name:           "Invalid JSON Body",
// 			method:         http.MethodPost,
// 			requestBody:    "invalid json",
// 			expectedStatus: http.StatusBadRequest,
// 			expectedBody: map[string]interface{}{
// 				"error":   true,
// 				"data":    false,
// 				"message": "cannot unmarshal string into Go value",
// 			},
// 			setupMock: func(validator *MockTaskValidator, logger *MockLogger, ipfsFetcher *MockIPFSFetcher) {
// 				logger.On("Info", "Validating task ...", []interface{}{"trace_id", ""}).Return()
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
// 				"error":   true,
// 				"data":    false,
// 				"message": "Failed to decode hex data",
// 			},
// 			setupMock: func(validator *MockTaskValidator, logger *MockLogger, ipfsFetcher *MockIPFSFetcher) {
// 				logger.On("Info", "Validating task ...", []interface{}{"trace_id", ""}).Return()
// 				logger.On("Errorf", "Failed to hex-decode data: %v", mock.Anything).Return()
// 			},
// 		},
// 		{
// 			name:   "Successful Task Validation",
// 			method: http.MethodPost,
// 			requestBody: map[string]interface{}{
// 				"proofOfTask":      "0x123",
// 				"data":             "0x" + hex.EncodeToString([]byte("valid-ipfs-cid")),
// 				"taskDefinitionId": 1,
// 				"performer":        "0x123",
// 			},
// 			expectedStatus: http.StatusOK,
// 			expectedBody: map[string]interface{}{
// 				"data":    true,
// 				"error":   false,
// 				"message": nil,
// 			},
// 			setupMock: func(validator *MockTaskValidator, logger *MockLogger, ipfsFetcher *MockIPFSFetcher) {
// 				logger.On("Info", "Validating task ...", []interface{}{"trace_id", ""}).Return()
// 				logger.On("Infof", "Decoded Data CID: %s", []interface{}{"valid-ipfs-cid"}).Return()
// 				ipfsFetcher.On("FetchContent", "valid-ipfs-cid").Return(types.IPFSData{}, nil)
// 				validator.On("ValidateTask", mock.Anything, mock.Anything).Return(true, nil)
// 				logger.On("Info", "Task validation completed", []interface{}{"trace_id", ""}).Return()
// 			},
// 		},
// 		{
// 			name:   "Validation Failure",
// 			method: http.MethodPost,
// 			requestBody: map[string]interface{}{
// 				"proofOfTask":      "0x123",
// 				"data":             "0x" + hex.EncodeToString([]byte("valid-ipfs-cid")),
// 				"taskDefinitionId": 1,
// 				"performer":        "0x123",
// 			},
// 			expectedStatus: http.StatusOK,
// 			expectedBody: map[string]interface{}{
// 				"data":    false,
// 				"error":   true,
// 				"message": "validation failed",
// 			},
// 			setupMock: func(validator *MockTaskValidator, logger *MockLogger, ipfsFetcher *MockIPFSFetcher) {
// 				logger.On("Info", "Validating task ...", []interface{}{"trace_id", ""}).Return()
// 				logger.On("Infof", "Decoded Data CID: %s", []interface{}{"valid-ipfs-cid"}).Return()
// 				ipfsFetcher.On("FetchContent", "valid-ipfs-cid").Return(types.IPFSData{}, nil)
// 				validator.On("ValidateTask", mock.Anything, mock.Anything).Return(false, fmt.Errorf("validation failed"))
// 				logger.On("Error", "Validation error", mock.Anything).Return()
// 			},
// 		},
// 		{
// 			name:   "IPFS Content Fetch Failure",
// 			method: http.MethodPost,
// 			requestBody: map[string]interface{}{
// 				"proofOfTask":      "0x123",
// 				"data":             "0x" + hex.EncodeToString([]byte("invalid-ipfs-cid")),
// 				"taskDefinitionId": 1,
// 				"performer":        "0x123",
// 			},
// 			expectedStatus: http.StatusInternalServerError,
// 			expectedBody: map[string]interface{}{
// 				"data":    false,
// 				"error":   true,
// 				"message": "Failed to fetch IPFS content: failed to fetch IPFS content",
// 			},
// 			setupMock: func(validator *MockTaskValidator, logger *MockLogger, ipfsFetcher *MockIPFSFetcher) {
// 				logger.On("Info", "Validating task ...", []interface{}{"trace_id", ""}).Return()
// 				logger.On("Infof", "Decoded Data CID: %s", []interface{}{"invalid-ipfs-cid"}).Return()
// 				ipfsFetcher.On("FetchContent", "invalid-ipfs-cid").Return(types.IPFSData{}, fmt.Errorf("failed to fetch IPFS content"))
// 				logger.On("Errorf", "Failed to fetch IPFS content: %v", mock.Anything).Return()
// 			},
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			handler, validator, logger, ipfsFetcher := setupValidateTestHandler()
// 			tt.setupMock(validator, logger, ipfsFetcher)

// 			gin.SetMode(gin.TestMode)
// 			router := gin.New()
// 			router.POST("/validate", handler.ValidateTask)
// 			router.GET("/validate", handler.ValidateTask)

// 			var body io.Reader
// 			if tt.requestBody != nil {
// 				jsonBody, err := json.Marshal(tt.requestBody)
// 				assert.NoError(t, err)
// 				body = bytes.NewBuffer(jsonBody)
// 			}

// 			req := httptest.NewRequest(tt.method, "/validate", body)
// 			req.Header.Set("Content-Type", "application/json")
// 			w := httptest.NewRecorder()

// 			router.ServeHTTP(w, req)

// 			assert.Equal(t, tt.expectedStatus, w.Code)

// 			var response map[string]interface{}
// 			err := json.Unmarshal(w.Body.Bytes(), &response)
// 			assert.NoError(t, err)

// 			for key, expectedValue := range tt.expectedBody {
// 				if key == "message" && expectedValue != nil {
// 					assert.Contains(t, response[key], expectedValue, "Response field %s does not contain expected substring", key)
// 				} else {
// 					assert.Equal(t, expectedValue, response[key], "Response field %s does not match", key)
// 				}
// 			}

// 			validator.AssertExpectations(t)
// 			logger.AssertExpectations(t)
// 			ipfsFetcher.AssertExpectations(t)
// 		})
// 	}
// }
