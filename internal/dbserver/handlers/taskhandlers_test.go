package handlers

import (
	"bytes"
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/types"
)

// Test setup helper
func setupTestTaskHandler() (*Handler, *MockTaskRepository) {
	mockTaskRepo := new(MockTaskRepository)
	handler := &Handler{
		taskRepository: mockTaskRepo,
		logger:         &MockLogger{},
	}
	return handler, mockTaskRepo
}

func TestCreateTaskData(t *testing.T) {
	handler, mockTaskRepo := setupTestTaskHandler()

	tests := []struct {
		name          string
		requestBody   types.CreateTaskDataRequest
		setupMocks    func()
		expectedCode  int
		expectedError string
	}{
		{
			name: "Success - Create Task",
			requestBody: types.CreateTaskDataRequest{
				JobID:            big.NewInt(1),
				TaskDefinitionID: 1,
			},
			setupMocks: func() {
				mockTaskRepo.On("CreateTaskDataInDB", &types.CreateTaskDataRequest{
					JobID:            big.NewInt(1),
					TaskDefinitionID: 1,
				}).Return(int64(1), nil)
			},
			expectedCode: http.StatusCreated,
		},
		{
			name:        "Error - Invalid Request Body",
			requestBody: types.CreateTaskDataRequest{},
			setupMocks: func() {
				// No mock setup needed as the handler should return before calling the repository
			},
			expectedCode:  http.StatusBadRequest,
			expectedError: "invalid request body",
		},
		{
			name: "Error - Database Error",
			requestBody: types.CreateTaskDataRequest{
				JobID:            big.NewInt(1),
				TaskDefinitionID: 1,
			},
			setupMocks: func() {
				mockTaskRepo.On("CreateTaskDataInDB", &types.CreateTaskDataRequest{
					JobID:            big.NewInt(1),
					TaskDefinitionID: 1,
				}).Return(int64(0), assert.AnError)
			},
			expectedCode:  http.StatusInternalServerError,
			expectedError: assert.AnError.Error(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mocks for isolation
			mockTaskRepo.ExpectedCalls = nil

			// Setup
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			// Setup request
			body, _ := json.Marshal(tt.requestBody)
			c.Request = httptest.NewRequest("POST", "/", bytes.NewBuffer(body))
			c.Request.Header.Set("Content-Type", "application/json")

			// Setup mocks
			tt.setupMocks()

			// Execute
			handler.CreateTaskData(c)

			// Assert
			assert.Equal(t, tt.expectedCode, w.Code)
			if tt.expectedError != "" {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response["error"], tt.expectedError)
			} else {
				var response map[string]int64
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, int64(1), response["task_id"])
			}
		})
	}
}

func TestUpdateTaskExecutionData(t *testing.T) {
	handler, mockTaskRepo := setupTestTaskHandler()

	// Create a fixed timestamp for testing
	fixedTime := time.Date(2025, time.June, 2, 17, 41, 39, 0, time.Local)

	tests := []struct {
		name          string
		taskID        string
		requestBody   types.UpdateTaskExecutionDataRequest
		setupMocks    func()
		expectedCode  int
		expectedError string
	}{
		{
			name:   "Success - Update Task Execution Data",
			taskID: "1",
			requestBody: types.UpdateTaskExecutionDataRequest{
				TaskID:             1,
				ExecutionTimestamp: fixedTime,
				ExecutionTxHash:    "0x123",
				ProofOfTask:        "proof",
				TaskOpXCost:        10.5,
			},
			setupMocks: func() {
				mockTaskRepo.On("UpdateTaskExecutionDataInDB", &types.UpdateTaskExecutionDataRequest{
					TaskID:             1,
					ExecutionTimestamp: fixedTime,
					ExecutionTxHash:    "0x123",
					ProofOfTask:        "proof",
					TaskOpXCost:        10.5,
				}).Return(nil)
			},
			expectedCode: http.StatusOK,
		},
		{
			name:          "Error - Invalid Request Body",
			taskID:        "1",
			requestBody:   types.UpdateTaskExecutionDataRequest{},
			setupMocks:    func() {},
			expectedCode:  http.StatusBadRequest,
			expectedError: "missing required fields",
		},
		{
			name:   "Error - Missing Required Fields",
			taskID: "1",
			requestBody: types.UpdateTaskExecutionDataRequest{
				TaskID:             1,
				ExecutionTimestamp: time.Time{}, // Zero time
				ExecutionTxHash:    "",
				ProofOfTask:        "proof",
				TaskOpXCost:        10.5,
			},
			setupMocks:    func() {},
			expectedCode:  http.StatusBadRequest,
			expectedError: "missing required fields",
		},
		{
			name:   "Error - Database Error",
			taskID: "1",
			requestBody: types.UpdateTaskExecutionDataRequest{
				TaskID:             1,
				ExecutionTimestamp: fixedTime,
				ExecutionTxHash:    "0x123",
				ProofOfTask:        "proof",
				TaskOpXCost:        10.5,
			},
			setupMocks: func() {
				mockTaskRepo.On("UpdateTaskExecutionDataInDB", &types.UpdateTaskExecutionDataRequest{
					TaskID:             1,
					ExecutionTimestamp: fixedTime,
					ExecutionTxHash:    "0x123",
					ProofOfTask:        "proof",
					TaskOpXCost:        10.5,
				}).Return(assert.AnError)
			},
			expectedCode:  http.StatusInternalServerError,
			expectedError: assert.AnError.Error(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mocks for isolation
			mockTaskRepo.ExpectedCalls = nil

			// Setup
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			// Setup request
			body, _ := json.Marshal(tt.requestBody)
			c.Request = httptest.NewRequest("PUT", "/", bytes.NewBuffer(body))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Params = []gin.Param{{Key: "id", Value: tt.taskID}}

			// Setup mocks
			tt.setupMocks()

			// Execute
			handler.UpdateTaskExecutionData(c)

			// Assert
			assert.Equal(t, tt.expectedCode, w.Code)
			if tt.expectedError != "" {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response["error"], tt.expectedError)
			} else {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "Task execution data updated successfully", response["message"])
			}
		})
	}
}

func TestUpdateTaskAttestationData(t *testing.T) {
	handler, mockTaskRepo := setupTestTaskHandler()

	tests := []struct {
		name          string
		taskID        string
		requestBody   types.UpdateTaskAttestationDataRequest
		setupMocks    func()
		expectedCode  int
		expectedError string
	}{
		{
			name:   "Success - Update Task Attestation Data",
			taskID: "1",
			requestBody: types.UpdateTaskAttestationDataRequest{
				TaskID:               1,
				TaskNumber:           1,
				TaskAttesterIDs:      []int64{1, 2},
				TpSignature:          []byte("tp"),
				TaSignature:          []byte("ta"),
				TaskSubmissionTxHash: "0x123",
				IsSuccessful:         true,
			},
			setupMocks: func() {
				mockTaskRepo.On("UpdateTaskAttestationDataInDB", &types.UpdateTaskAttestationDataRequest{
					TaskID:               1,
					TaskNumber:           1,
					TaskAttesterIDs:      []int64{1, 2},
					TpSignature:          []byte("tp"),
					TaSignature:          []byte("ta"),
					TaskSubmissionTxHash: "0x123",
					IsSuccessful:         true,
				}).Return(nil)
			},
			expectedCode: http.StatusOK,
		},
		{
			name:          "Error - Invalid Request Body",
			taskID:        "1",
			requestBody:   types.UpdateTaskAttestationDataRequest{},
			setupMocks:    func() {},
			expectedCode:  http.StatusBadRequest,
			expectedError: "missing required fields",
		},
		{
			name:   "Error - Database Error",
			taskID: "1",
			requestBody: types.UpdateTaskAttestationDataRequest{
				TaskID:               1,
				TaskNumber:           1,
				TaskAttesterIDs:      []int64{1, 2},
				TpSignature:          []byte("tp"),
				TaSignature:          []byte("ta"),
				TaskSubmissionTxHash: "0x123",
				IsSuccessful:         true,
			},
			setupMocks: func() {
				mockTaskRepo.On("UpdateTaskAttestationDataInDB", &types.UpdateTaskAttestationDataRequest{
					TaskID:               1,
					TaskNumber:           1,
					TaskAttesterIDs:      []int64{1, 2},
					TpSignature:          []byte("tp"),
					TaSignature:          []byte("ta"),
					TaskSubmissionTxHash: "0x123",
					IsSuccessful:         true,
				}).Return(assert.AnError)
			},
			expectedCode:  http.StatusInternalServerError,
			expectedError: assert.AnError.Error(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mocks for isolation
			mockTaskRepo.ExpectedCalls = nil

			// Setup
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			// Setup request
			body, _ := json.Marshal(tt.requestBody)
			c.Request = httptest.NewRequest("PUT", "/", bytes.NewBuffer(body))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Params = []gin.Param{{Key: "id", Value: tt.taskID}}

			// Setup mocks
			tt.setupMocks()

			// Execute
			handler.UpdateTaskAttestationData(c)

			// Assert
			assert.Equal(t, tt.expectedCode, w.Code)
			if tt.expectedError != "" {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response["error"], tt.expectedError)
			} else {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "Task attestation data updated successfully", response["message"])
			}
		})
	}
}

func TestGetTaskDataByID(t *testing.T) {
	handler, mockTaskRepo := setupTestTaskHandler()

	// Create fixed timestamps for testing
	fixedTime := time.Date(2025, time.June, 2, 17, 41, 39, 0, time.Local)

	tests := []struct {
		name          string
		taskID        string
		setupMocks    func()
		expectedCode  int
		expectedError string
	}{
		{
			name:   "Success - Get Task Data",
			taskID: "1",
			setupMocks: func() {
				mockTaskRepo.On("GetTaskDataByID", int64(1)).Return(types.TaskData{
					TaskID:               1,
					TaskNumber:           1,
					JobID:                big.NewInt(1),
					TaskDefinitionID:     1,
					CreatedAt:            fixedTime,
					TaskOpXCost:          10.5,
					ExecutionTimestamp:   fixedTime,
					ExecutionTxHash:      "0x123",
					TaskPerformerID:      1,
					ProofOfTask:          "proof",
					TaskAttesterIDs:      []int64{1, 2},
					TpSignature:          []byte("tp"),
					TaSignature:          []byte("ta"),
					TaskSubmissionTxHash: "0x123",
					IsSuccessful:         true,
				}, nil)
			},
			expectedCode: http.StatusOK,
		},
		{
			name:          "Error - Invalid Task ID",
			taskID:        "invalid",
			setupMocks:    func() {},
			expectedCode:  http.StatusBadRequest,
			expectedError: "Invalid task ID",
		},
		{
			name:   "Error - Task Not Found",
			taskID: "1",
			setupMocks: func() {
				mockTaskRepo.On("GetTaskDataByID", int64(1)).Return(types.TaskData{}, assert.AnError)
			},
			expectedCode:  http.StatusInternalServerError,
			expectedError: assert.AnError.Error(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mocks for isolation
			mockTaskRepo.ExpectedCalls = nil

			// Setup
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			// Setup request
			c.Request = httptest.NewRequest("GET", "/", nil)
			c.Params = []gin.Param{{Key: "id", Value: tt.taskID}}

			// Setup mocks
			tt.setupMocks()

			// Execute
			handler.GetTaskDataByID(c)

			// Assert
			assert.Equal(t, tt.expectedCode, w.Code)
			if tt.expectedError != "" {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response["error"], tt.expectedError)
			} else {
				var response types.TaskData
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, int64(1), response.TaskID)
				assert.Equal(t, int64(1), response.TaskNumber)
				assert.Equal(t, int64(1), response.JobID)
				assert.Equal(t, 1, response.TaskDefinitionID)
				assert.Equal(t, 10.5, response.TaskOpXCost)
				assert.Equal(t, "0x123", response.ExecutionTxHash)
				assert.Equal(t, int64(1), response.TaskPerformerID)
				assert.Equal(t, "proof", response.ProofOfTask)
				assert.Equal(t, []int64{1, 2}, response.TaskAttesterIDs)
				assert.Equal(t, []byte("tp"), response.TpSignature)
				assert.Equal(t, []byte("ta"), response.TaSignature)
				assert.Equal(t, "0x123", response.TaskSubmissionTxHash)
				assert.True(t, response.IsSuccessful)
			}
		})
	}
}

func TestGetTasksByJobID(t *testing.T) {
	handler, mockTaskRepo := setupTestTaskHandler()

	// Create fixed timestamps for testing
	fixedTime := time.Date(2025, time.June, 2, 17, 41, 39, 0, time.Local)

	tests := []struct {
		name          string
		jobID         string
		setupMocks    func()
		expectedCode  int
		expectedError string
	}{
		{
			name:  "Success - Get Tasks by Job ID",
			jobID: "1",
			setupMocks: func() {
				mockTaskRepo.On("GetTasksByJobID", int64(1)).Return([]types.TasksByJobIDResponse{
					{
						TaskID:             1,
						TaskNumber:         1,
						TaskOpXCost:        10.5,
						ExecutionTimestamp: fixedTime,
						ExecutionTxHash:    "0x123",
						TaskPerformerID:    1,
						TaskAttesterIDs:    []int64{1, 2},
						IsSuccessful:       true,
					},
					{
						TaskID:             2,
						TaskNumber:         2,
						TaskOpXCost:        20.5,
						ExecutionTimestamp: fixedTime,
						ExecutionTxHash:    "0x456",
						TaskPerformerID:    2,
						TaskAttesterIDs:    []int64{3, 4},
						IsSuccessful:       true,
					},
				}, nil)
			},
			expectedCode: http.StatusOK,
		},
		{
			name:          "Error - Invalid Job ID",
			jobID:         "invalid",
			setupMocks:    func() {},
			expectedCode:  http.StatusBadRequest,
			expectedError: "Invalid job ID",
		},
		{
			name:  "Error - Tasks Not Found",
			jobID: "1",
			setupMocks: func() {
				mockTaskRepo.On("GetTasksByJobID", int64(1)).Return([]types.TasksByJobIDResponse{}, assert.AnError)
			},
			expectedCode:  http.StatusInternalServerError,
			expectedError: assert.AnError.Error(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mocks for isolation
			mockTaskRepo.ExpectedCalls = nil

			// Setup
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			// Setup request
			c.Request = httptest.NewRequest("GET", "/", nil)
			c.Params = []gin.Param{{Key: "id", Value: tt.jobID}}

			// Setup mocks
			tt.setupMocks()

			// Execute
			handler.GetTasksByJobID(c)

			// Assert
			assert.Equal(t, tt.expectedCode, w.Code)
			if tt.expectedError != "" {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response["error"], tt.expectedError)
			} else {
				var response []types.TasksByJobIDResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Len(t, response, 2)
				assert.Equal(t, int64(1), response[0].TaskID)
				assert.Equal(t, int64(1), response[0].TaskNumber)
				assert.Equal(t, 10.5, response[0].TaskOpXCost)
				assert.Equal(t, "0x123", response[0].ExecutionTxHash)
				assert.Equal(t, int64(1), response[0].TaskPerformerID)
				assert.Equal(t, []int64{1, 2}, response[0].TaskAttesterIDs)
				assert.True(t, response[0].IsSuccessful)
				assert.Equal(t, int64(2), response[1].TaskID)
				assert.Equal(t, int64(2), response[1].TaskNumber)
				assert.Equal(t, 20.5, response[1].TaskOpXCost)
				assert.Equal(t, "0x456", response[1].ExecutionTxHash)
				assert.Equal(t, int64(2), response[1].TaskPerformerID)
				assert.Equal(t, []int64{3, 4}, response[1].TaskAttesterIDs)
				assert.True(t, response[1].IsSuccessful)
			}
		})
	}
}

func TestGetTaskFees(t *testing.T) {
	handler, _ := setupTestTaskHandler()

	tests := []struct {
		name          string
		ipfsURLs      string
		setupMocks    func()
		expectedCode  int
		expectedError string
	}{
		{
			name:     "Success - Get Task Fees",
			ipfsURLs: "ipfs://QmTest1,ipfs://QmTest2",
			setupMocks: func() {
				// Mock Docker client and resource monitoring if needed
			},
			expectedCode: http.StatusOK,
		},
		{
			name:          "Error - Missing IPFS URLs",
			ipfsURLs:      "",
			setupMocks:    func() {},
			expectedCode:  http.StatusInternalServerError,
			expectedError: "missing IPFS URLs",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			// Setup request
			c.Request = httptest.NewRequest("GET", "/", nil)
			q := c.Request.URL.Query()
			if tt.ipfsURLs != "" {
				q.Add("ipfs_url", tt.ipfsURLs)
			}
			c.Request.URL.RawQuery = q.Encode()

			// Setup mocks
			tt.setupMocks()

			// Execute
			handler.GetTaskFees(c)

			// Assert
			assert.Equal(t, tt.expectedCode, w.Code)
			if tt.expectedError != "" {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response["error"], tt.expectedError)
			} else {
				var response map[string]float64
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.GreaterOrEqual(t, response["total_fee"], 0.0)
			}
		})
	}
}

func TestUpdateTaskFee(t *testing.T) {
	handler, mockTaskRepo := setupTestTaskHandler()

	tests := []struct {
		name          string
		taskID        string
		requestBody   struct{ Fee float64 }
		setupMocks    func()
		expectedCode  int
		expectedError string
	}{
		{
			name:        "Success - Update Task Fee",
			taskID:      "1",
			requestBody: struct{ Fee float64 }{Fee: 10.5},
			setupMocks: func() {
				mockTaskRepo.On("UpdateTaskFee", int64(1), 10.5).Return(nil)
			},
			expectedCode: http.StatusOK,
		},
		{
			name:          "Error - Invalid Task ID",
			taskID:        "invalid",
			requestBody:   struct{ Fee float64 }{Fee: 10.5},
			setupMocks:    func() {},
			expectedCode:  http.StatusBadRequest,
			expectedError: "Invalid task ID",
		},
		{
			name:        "Error - Database Error",
			taskID:      "1",
			requestBody: struct{ Fee float64 }{Fee: 10.5},
			setupMocks: func() {
				mockTaskRepo.On("UpdateTaskFee", int64(1), 10.5).Return(assert.AnError)
			},
			expectedCode:  http.StatusInternalServerError,
			expectedError: assert.AnError.Error(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mocks for isolation
			mockTaskRepo.ExpectedCalls = nil

			// Setup
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			// Setup request
			body, _ := json.Marshal(tt.requestBody)
			c.Request = httptest.NewRequest("PUT", "/", bytes.NewBuffer(body))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Params = []gin.Param{{Key: "id", Value: tt.taskID}}

			// Setup mocks
			tt.setupMocks()

			// Execute
			handler.UpdateTaskFee(c)

			// Assert
			assert.Equal(t, tt.expectedCode, w.Code)
			if tt.expectedError != "" {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response["error"], tt.expectedError)
			} else {
				var response struct{ Fee float64 }
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, 10.5, response.Fee)
			}
		})
	}
}
