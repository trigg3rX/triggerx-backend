package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	datastoreMocks "github.com/trigg3rX/triggerx-backend/pkg/datastore/mocks"
	pkgErrors "github.com/trigg3rX/triggerx-backend/pkg/errors"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

func setupTaskReadTestHandler() (*Handler, *datastoreMocks.MockGenericRepository[types.TaskDataEntity], *datastoreMocks.MockGenericRepository[types.JobDataEntity], *logging.MockLogger) {
	mockTaskRepo := new(datastoreMocks.MockGenericRepository[types.TaskDataEntity])
	mockJobRepo := new(datastoreMocks.MockGenericRepository[types.JobDataEntity])
	mockLogger := new(logging.MockLogger)
	mockLogger.SetupDefaultExpectations()

	handler := &Handler{
		logger:         mockLogger,
		taskRepository: mockTaskRepo,
		jobRepository:  mockJobRepo,
	}

	return handler, mockTaskRepo, mockJobRepo, mockLogger
}

func TestGetTaskDataByTaskID(t *testing.T) {
	tests := []struct {
		name           string
		taskID         string
		mockSetup      func(*datastoreMocks.MockGenericRepository[types.TaskDataEntity])
		expectedCode   int
		expectedError  string
		validateResult func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:   "Success - Valid Task ID",
			taskID: "12345",
			mockSetup: func(mockRepo *datastoreMocks.MockGenericRepository[types.TaskDataEntity]) {
				taskData := &types.TaskDataEntity{
					TaskID:               12345,
					TaskNumber:           1,
					JobID:                "job-123",
					TaskDefinitionID:     1,
					CreatedAt:            time.Now(),
					TaskOpxPredictedCost: "1000000000000000000",
					TaskOpxActualCost:    "900000000000000000",
					ExecutionTimestamp:   time.Now(),
					ExecutionTxHash:      "0xabc123",
					TaskPerformerID:      1,
					TaskAttesterIDs:      []int64{2, 3},
					ConvertedArguments:   "arg1,arg2",
					ProofOfTask:          "proof123",
					SubmissionTxHash:     "0xdef456",
					IsSuccessful:         true,
					IsAccepted:           true,
					IsImua:               false,
				}
				mockRepo.On("GetByID", mock.Anything, int64(12345)).Return(taskData, nil).Once()
			},
			expectedCode: http.StatusOK,
			validateResult: func(t *testing.T, w *httptest.ResponseRecorder) {
				var taskData types.TaskDataEntity
				err := json.Unmarshal(w.Body.Bytes(), &taskData)
				assert.NoError(t, err)
				assert.Equal(t, int64(12345), taskData.TaskID)
				assert.Equal(t, "job-123", taskData.JobID)
				assert.True(t, taskData.IsSuccessful)
				assert.True(t, taskData.IsAccepted)
			},
		},
		{
			name:   "Error - Missing Task ID",
			taskID: "",
			mockSetup: func(mockRepo *datastoreMocks.MockGenericRepository[types.TaskDataEntity]) {
				// No mock setup needed
			},
			expectedCode:  http.StatusBadRequest,
			expectedError: pkgErrors.ErrInvalidRequestBody,
			validateResult: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, pkgErrors.ErrInvalidRequestBody, response["error"])
			},
		},
		{
			name:   "Error - Invalid Task ID (non-numeric)",
			taskID: "invalid",
			mockSetup: func(mockRepo *datastoreMocks.MockGenericRepository[types.TaskDataEntity]) {
				// No mock setup needed
			},
			expectedCode:  http.StatusBadRequest,
			expectedError: pkgErrors.ErrInvalidRequestBody,
			validateResult: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, pkgErrors.ErrInvalidRequestBody, response["error"])
			},
		},
		{
			name:   "Error - Task Not Found",
			taskID: "99999",
			mockSetup: func(mockRepo *datastoreMocks.MockGenericRepository[types.TaskDataEntity]) {
				mockRepo.On("GetByID", mock.Anything, int64(99999)).Return(nil, errors.New("not found")).Once()
			},
			expectedCode:  http.StatusNotFound,
			expectedError: pkgErrors.ErrDBRecordNotFound,
			validateResult: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, pkgErrors.ErrDBRecordNotFound, response["error"])
			},
		},
		{
			name:   "Error - Task Data is Nil",
			taskID: "12345",
			mockSetup: func(mockRepo *datastoreMocks.MockGenericRepository[types.TaskDataEntity]) {
				mockRepo.On("GetByID", mock.Anything, int64(12345)).Return(nil, nil).Once()
			},
			expectedCode:  http.StatusNotFound,
			expectedError: pkgErrors.ErrDBRecordNotFound,
			validateResult: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, pkgErrors.ErrDBRecordNotFound, response["error"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockTaskRepo, _, mockLogger := setupTaskReadTestHandler()
			defer mockLogger.AssertExpectations(t)
			defer mockTaskRepo.AssertExpectations(t)

			// Setup
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("GET", "/tasks/"+tt.taskID, nil)
			c.Params = gin.Params{gin.Param{Key: "id", Value: tt.taskID}}

			// Setup mocks
			tt.mockSetup(mockTaskRepo)

			// Execute
			handler.GetTaskDataByTaskID(c)

			// Assert
			assert.Equal(t, tt.expectedCode, w.Code)
			if tt.validateResult != nil {
				tt.validateResult(t, w)
			}
		})
	}
}

func TestGetTasksByJobID(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name           string
		jobID          string
		mockSetup      func(*datastoreMocks.MockGenericRepository[types.JobDataEntity], *datastoreMocks.MockGenericRepository[types.TaskDataEntity])
		expectedCode   int
		expectedError  string
		validateResult func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:  "Success - Job with Multiple Tasks",
			jobID: "job-123",
			mockSetup: func(
				mockJobRepo *datastoreMocks.MockGenericRepository[types.JobDataEntity],
				mockTaskRepo *datastoreMocks.MockGenericRepository[types.TaskDataEntity],
			) {
				jobData := &types.JobDataEntity{
					JobID:             "job-123",
					JobTitle:          "Test Job",
					TaskDefinitionID:  1,
					CreatedChainID:    "11155111",
					UserAddress:       "0x123",
					TaskIDs:           []int64{1, 2, 3},
					Status:            "active",
					JobCostPrediction: "1000000000000000000",
					CreatedAt:         now,
				}
				mockJobRepo.On("GetByID", mock.Anything, "job-123").Return(jobData, nil).Once()

				task1 := &types.TaskDataEntity{
					TaskID:               1,
					TaskNumber:           1,
					JobID:                "job-123",
					TaskOpxPredictedCost: "1000000000000000000",
					TaskOpxActualCost:    "900000000000000000",
					ExecutionTimestamp:   now,
					ExecutionTxHash:      "0xabc123",
					TaskPerformerID:      1,
					TaskAttesterIDs:      []int64{2, 3},
					ConvertedArguments:   "arg1",
					IsSuccessful:         true,
					IsAccepted:           true,
				}
				task2 := &types.TaskDataEntity{
					TaskID:               2,
					TaskNumber:           2,
					JobID:                "job-123",
					TaskOpxPredictedCost: "1000000000000000000",
					TaskOpxActualCost:    "950000000000000000",
					ExecutionTimestamp:   now.Add(time.Hour),
					ExecutionTxHash:      "0xdef456",
					TaskPerformerID:      2,
					TaskAttesterIDs:      []int64{1, 3},
					ConvertedArguments:   "arg2",
					IsSuccessful:         true,
					IsAccepted:           false,
				}
				task3 := &types.TaskDataEntity{
					TaskID:               3,
					TaskNumber:           3,
					JobID:                "job-123",
					TaskOpxPredictedCost: "1000000000000000000",
					TaskOpxActualCost:    "1100000000000000000",
					ExecutionTimestamp:   now.Add(2 * time.Hour),
					ExecutionTxHash:      "0xghi789",
					TaskPerformerID:      3,
					TaskAttesterIDs:      []int64{1, 2},
					ConvertedArguments:   "arg3",
					IsSuccessful:         false,
					IsAccepted:           false,
				}

				mockTaskRepo.On("GetByID", mock.Anything, int64(1)).Return(task1, nil).Once()
				mockTaskRepo.On("GetByID", mock.Anything, int64(2)).Return(task2, nil).Once()
				mockTaskRepo.On("GetByID", mock.Anything, int64(3)).Return(task3, nil).Once()
			},
			expectedCode: http.StatusOK,
			validateResult: func(t *testing.T, w *httptest.ResponseRecorder) {
				var tasks []types.GetTasksByJobIDResponse
				err := json.Unmarshal(w.Body.Bytes(), &tasks)
				assert.NoError(t, err)
				assert.Len(t, tasks, 3)

				// Verify first task
				assert.Equal(t, int64(1), tasks[0].TaskNumber)
				assert.Equal(t, "1000000000000000000", tasks[0].TaskOpXPredictedCost)
				assert.Equal(t, "900000000000000000", tasks[0].TaskOpXActualCost)
				assert.Equal(t, "0xabc123", tasks[0].ExecutionTxHash)
				assert.True(t, tasks[0].IsSuccessful)
				assert.True(t, tasks[0].IsAccepted)
				assert.Contains(t, tasks[0].TxURL, "0xabc123")

				// Verify second task
				assert.Equal(t, int64(2), tasks[1].TaskNumber)
				assert.True(t, tasks[1].IsSuccessful)
				assert.False(t, tasks[1].IsAccepted)

				// Verify third task
				assert.Equal(t, int64(3), tasks[2].TaskNumber)
				assert.False(t, tasks[2].IsSuccessful)
			},
		},
		{
			name:  "Success - Job with No Tasks",
			jobID: "job-empty",
			mockSetup: func(
				mockJobRepo *datastoreMocks.MockGenericRepository[types.JobDataEntity],
				mockTaskRepo *datastoreMocks.MockGenericRepository[types.TaskDataEntity],
			) {
				jobData := &types.JobDataEntity{
					JobID:             "job-empty",
					JobTitle:          "Empty Job",
					CreatedChainID:    "1",
					TaskIDs:           []int64{},
					Status:            "pending",
					JobCostPrediction: "0",
					CreatedAt:         now,
				}
				mockJobRepo.On("GetByID", mock.Anything, "job-empty").Return(jobData, nil).Once()
			},
			expectedCode: http.StatusOK,
			validateResult: func(t *testing.T, w *httptest.ResponseRecorder) {
				var tasks []types.GetTasksByJobIDResponse
				err := json.Unmarshal(w.Body.Bytes(), &tasks)
				assert.NoError(t, err)
				assert.Len(t, tasks, 0)
			},
		},
		{
			name:  "Success - Job with Some Failed Task Lookups",
			jobID: "job-partial",
			mockSetup: func(
				mockJobRepo *datastoreMocks.MockGenericRepository[types.JobDataEntity],
				mockTaskRepo *datastoreMocks.MockGenericRepository[types.TaskDataEntity],
			) {
				jobData := &types.JobDataEntity{
					JobID:             "job-partial",
					JobTitle:          "Partial Job",
					CreatedChainID:    "10",
					TaskIDs:           []int64{1, 2, 3},
					Status:            "active",
					JobCostPrediction: "1000000000000000000",
					CreatedAt:         now,
				}
				mockJobRepo.On("GetByID", mock.Anything, "job-partial").Return(jobData, nil).Once()

				task1 := &types.TaskDataEntity{
					TaskID:               1,
					TaskNumber:           1,
					JobID:                "job-partial",
					TaskOpxPredictedCost: "1000000000000000000",
					TaskOpxActualCost:    "900000000000000000",
					ExecutionTimestamp:   now,
					ExecutionTxHash:      "0xabc123",
					TaskPerformerID:      1,
					TaskAttesterIDs:      []int64{2, 3},
					ConvertedArguments:   "arg1",
					IsSuccessful:         true,
					IsAccepted:           true,
				}

				mockTaskRepo.On("GetByID", mock.Anything, int64(1)).Return(task1, nil).Once()
				mockTaskRepo.On("GetByID", mock.Anything, int64(2)).Return(nil, errors.New("not found")).Once()
				mockTaskRepo.On("GetByID", mock.Anything, int64(3)).Return(nil, nil).Once()
			},
			expectedCode: http.StatusOK,
			validateResult: func(t *testing.T, w *httptest.ResponseRecorder) {
				var tasks []types.GetTasksByJobIDResponse
				err := json.Unmarshal(w.Body.Bytes(), &tasks)
				assert.NoError(t, err)
				// Only one task should be returned since the other two failed
				assert.Len(t, tasks, 1)
				assert.Equal(t, int64(1), tasks[0].TaskNumber)
			},
		},
		{
			name:  "Error - Missing Job ID",
			jobID: "",
			mockSetup: func(
				mockJobRepo *datastoreMocks.MockGenericRepository[types.JobDataEntity],
				mockTaskRepo *datastoreMocks.MockGenericRepository[types.TaskDataEntity],
			) {
				// No mock setup needed
			},
			expectedCode:  http.StatusBadRequest,
			expectedError: pkgErrors.ErrInvalidRequestBody,
			validateResult: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, pkgErrors.ErrInvalidRequestBody, response["error"])
			},
		},
		{
			name:  "Error - Job Not Found",
			jobID: "job-nonexistent",
			mockSetup: func(
				mockJobRepo *datastoreMocks.MockGenericRepository[types.JobDataEntity],
				mockTaskRepo *datastoreMocks.MockGenericRepository[types.TaskDataEntity],
			) {
				mockJobRepo.On("GetByID", mock.Anything, "job-nonexistent").Return(nil, errors.New("not found")).Once()
			},
			expectedCode:  http.StatusNotFound,
			expectedError: pkgErrors.ErrDBRecordNotFound,
			validateResult: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, pkgErrors.ErrDBRecordNotFound, response["error"])
			},
		},
		{
			name:  "Error - Job Data is Nil",
			jobID: "job-nil",
			mockSetup: func(
				mockJobRepo *datastoreMocks.MockGenericRepository[types.JobDataEntity],
				mockTaskRepo *datastoreMocks.MockGenericRepository[types.TaskDataEntity],
			) {
				mockJobRepo.On("GetByID", mock.Anything, "job-nil").Return(nil, nil).Once()
			},
			expectedCode:  http.StatusNotFound,
			expectedError: pkgErrors.ErrDBRecordNotFound,
			validateResult: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, pkgErrors.ErrDBRecordNotFound, response["error"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockTaskRepo, mockJobRepo, mockLogger := setupTaskReadTestHandler()
			defer mockLogger.AssertExpectations(t)
			defer mockTaskRepo.AssertExpectations(t)
			defer mockJobRepo.AssertExpectations(t)

			// Setup
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("GET", "/jobs/"+tt.jobID+"/tasks", nil)
			c.Params = gin.Params{gin.Param{Key: "job_id", Value: tt.jobID}}

			// Setup mocks
			tt.mockSetup(mockJobRepo, mockTaskRepo)

			// Execute
			handler.GetTasksByJobID(c)

			// Assert
			assert.Equal(t, tt.expectedCode, w.Code)
			if tt.validateResult != nil {
				tt.validateResult(t, w)
			}
		})
	}
}

func TestGetExplorerBaseURL(t *testing.T) {
	tests := []struct {
		name        string
		chainID     string
		expectedURL string
	}{
		// Testnets
		{
			name:        "Ethereum Sepolia",
			chainID:     "11155111",
			expectedURL: "https://eth-sepolia.blockscout.com/tx/",
		},
		{
			name:        "Optimism Sepolia",
			chainID:     "11155420",
			expectedURL: "https://testnet-explorer.optimism.io/tx/",
		},
		{
			name:        "Base Sepolia",
			chainID:     "84532",
			expectedURL: "https:/base-sepolia.blockscout.com/tx/",
		},
		{
			name:        "Arbitrum Sepolia",
			chainID:     "421614",
			expectedURL: "https://arbitrum-sepolia.blockscout.com/tx/",
		},
		// Mainnets
		{
			name:        "Ethereum Mainnet",
			chainID:     "1",
			expectedURL: "https://eth.blockscout.com/tx/",
		},
		{
			name:        "Optimism Mainnet",
			chainID:     "10",
			expectedURL: "https://explorer.optimism.io/tx/",
		},
		{
			name:        "Base Mainnet",
			chainID:     "8453",
			expectedURL: "https://base.blockscout.com/tx/",
		},
		{
			name:        "Arbitrum Mainnet",
			chainID:     "42161",
			expectedURL: "https:/arbitrum.blockscout.com/tx/",
		},
		// Default case
		{
			name:        "Unknown Chain - Default",
			chainID:     "99999",
			expectedURL: "https://sepolia.etherscan.io/tx/",
		},
		{
			name:        "Empty Chain ID - Default",
			chainID:     "",
			expectedURL: "https://sepolia.etherscan.io/tx/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getExplorerBaseURL(tt.chainID)
			assert.Equal(t, tt.expectedURL, result)
		})
	}
}

func TestGetTasksByJobID_ExplorerURLGeneration(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name              string
		chainID           string
		expectedURLPrefix string
	}{
		{
			name:              "Ethereum Sepolia Explorer URL",
			chainID:           "11155111",
			expectedURLPrefix: "https://eth-sepolia.blockscout.com/tx/",
		},
		{
			name:              "Ethereum Mainnet Explorer URL",
			chainID:           "1",
			expectedURLPrefix: "https://eth.blockscout.com/tx/",
		},
		{
			name:              "Optimism Mainnet Explorer URL",
			chainID:           "10",
			expectedURLPrefix: "https://explorer.optimism.io/tx/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockTaskRepo, mockJobRepo, mockLogger := setupTaskReadTestHandler()
			defer mockLogger.AssertExpectations(t)
			defer mockTaskRepo.AssertExpectations(t)
			defer mockJobRepo.AssertExpectations(t)

			jobData := &types.JobDataEntity{
				JobID:             "job-explorer-test",
				JobTitle:          "Explorer Test Job",
				CreatedChainID:    tt.chainID,
				TaskIDs:           []int64{1},
				Status:            "active",
				JobCostPrediction: "1000000000000000000",
				CreatedAt:         now,
			}
			mockJobRepo.On("GetByID", mock.Anything, "job-explorer-test").Return(jobData, nil).Once()

			taskData := &types.TaskDataEntity{
				TaskID:               1,
				TaskNumber:           1,
				JobID:                "job-explorer-test",
				TaskOpxPredictedCost: "1000000000000000000",
				TaskOpxActualCost:    "900000000000000000",
				ExecutionTimestamp:   now,
				ExecutionTxHash:      "0xTestTxHash",
				TaskPerformerID:      1,
				TaskAttesterIDs:      []int64{2, 3},
				ConvertedArguments:   "arg1",
				IsSuccessful:         true,
				IsAccepted:           true,
			}
			mockTaskRepo.On("GetByID", mock.Anything, int64(1)).Return(taskData, nil).Once()

			// Setup
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("GET", "/jobs/job-explorer-test/tasks", nil)
			c.Params = gin.Params{gin.Param{Key: "job_id", Value: "job-explorer-test"}}

			// Execute
			handler.GetTasksByJobID(c)

			// Assert
			assert.Equal(t, http.StatusOK, w.Code)

			var tasks []types.GetTasksByJobIDResponse
			err := json.Unmarshal(w.Body.Bytes(), &tasks)
			assert.NoError(t, err)
			assert.Len(t, tasks, 1)

			// Verify the TxURL is correctly generated
			expectedURL := tt.expectedURLPrefix + "0xTestTxHash"
			assert.Equal(t, expectedURL, tasks[0].TxURL)
		})
	}
}
