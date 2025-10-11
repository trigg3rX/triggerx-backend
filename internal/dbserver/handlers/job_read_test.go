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

func setupJobReadTestHandler() (
	*Handler,
	*datastoreMocks.MockGenericRepository[types.JobDataEntity],
	*datastoreMocks.MockGenericRepository[types.TimeJobDataEntity],
	*datastoreMocks.MockGenericRepository[types.EventJobDataEntity],
	*datastoreMocks.MockGenericRepository[types.ConditionJobDataEntity],
	*datastoreMocks.MockGenericRepository[types.UserDataEntity],
	*logging.MockLogger,
) {
	mockJobRepo := new(datastoreMocks.MockGenericRepository[types.JobDataEntity])
	mockTimeJobRepo := new(datastoreMocks.MockGenericRepository[types.TimeJobDataEntity])
	mockEventJobRepo := new(datastoreMocks.MockGenericRepository[types.EventJobDataEntity])
	mockConditionJobRepo := new(datastoreMocks.MockGenericRepository[types.ConditionJobDataEntity])
	mockUserRepo := new(datastoreMocks.MockGenericRepository[types.UserDataEntity])
	mockLogger := new(logging.MockLogger)
	mockLogger.SetupDefaultExpectations()

	handler := &Handler{
		logger:                 mockLogger,
		jobRepository:          mockJobRepo,
		timeJobRepository:      mockTimeJobRepo,
		eventJobRepository:     mockEventJobRepo,
		conditionJobRepository: mockConditionJobRepo,
		userRepository:         mockUserRepo,
	}

	return handler, mockJobRepo, mockTimeJobRepo, mockEventJobRepo, mockConditionJobRepo, mockUserRepo, mockLogger
}

func TestGetJobDataByJobID(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name           string
		jobID          string
		mockSetup      func(*datastoreMocks.MockGenericRepository[types.JobDataEntity], *datastoreMocks.MockGenericRepository[types.TimeJobDataEntity], *datastoreMocks.MockGenericRepository[types.EventJobDataEntity], *datastoreMocks.MockGenericRepository[types.ConditionJobDataEntity])
		expectedCode   int
		expectedError  string
		validateResult func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:  "Success - Time Job (TaskDefinitionID 1)",
			jobID: "12345",
			mockSetup: func(
				mockJobRepo *datastoreMocks.MockGenericRepository[types.JobDataEntity],
				mockTimeJobRepo *datastoreMocks.MockGenericRepository[types.TimeJobDataEntity],
				mockEventJobRepo *datastoreMocks.MockGenericRepository[types.EventJobDataEntity],
				mockConditionJobRepo *datastoreMocks.MockGenericRepository[types.ConditionJobDataEntity],
			) {
				jobData := &types.JobDataEntity{
					JobID:             "12345",
					JobTitle:          "Test Time Job",
					TaskDefinitionID:  1,
					CreatedChainID:    "11155111",
					UserAddress:       "0x1234567890123456789012345678901234567890",
					Status:            string(types.JobStatusRunning),
					JobCostPrediction: "1000000000000000000",
					CreatedAt:         now,
					UpdatedAt:         now,
				}
				mockJobRepo.On("GetByID", mock.Anything, "12345").Return(jobData, nil).Once()

				timeJobData := &types.TimeJobDataEntity{
					JobID:                  "12345",
					TaskDefinitionID:       1,
					TimeInterval:           3600,
					ScheduleType:           "interval",
					NextExecutionTimestamp: now.Add(time.Hour),
					TargetChainID:          "11155111",
					TargetContractAddress:  "0x1234567890123456789012345678901234567890",
					TargetFunction:         "testFunction",
					ABI:                    `[{"name":"testFunction","type":"function"}]`,
					ArgType:                1,
					Arguments:              []string{"arg1", "arg2"},
					IsCompleted:            false,
					ExpirationTime:         now.Add(24 * time.Hour),
				}
				mockTimeJobRepo.On("GetByID", mock.Anything, "12345").Return(timeJobData, nil).Once()
			},
			expectedCode: http.StatusOK,
			validateResult: func(t *testing.T, w *httptest.ResponseRecorder) {
				var result types.CompleteJobDataDTO
				err := json.Unmarshal(w.Body.Bytes(), &result)
				assert.NoError(t, err)
				assert.Equal(t, "12345", result.JobDataDTO.JobID)
				assert.Equal(t, "Test Time Job", result.JobDataDTO.JobTitle)
				assert.Equal(t, 1, result.JobDataDTO.TaskDefinitionID)
				assert.NotNil(t, result.TimeJobDataDTO)
				assert.Equal(t, int64(3600), result.TimeJobDataDTO.TimeInterval)
				assert.Nil(t, result.EventJobDataDTO)
				assert.Nil(t, result.ConditionJobDataDTO)
			},
		},
		{
			name:  "Success - Time Job (TaskDefinitionID 2)",
			jobID: "12346",
			mockSetup: func(
				mockJobRepo *datastoreMocks.MockGenericRepository[types.JobDataEntity],
				mockTimeJobRepo *datastoreMocks.MockGenericRepository[types.TimeJobDataEntity],
				mockEventJobRepo *datastoreMocks.MockGenericRepository[types.EventJobDataEntity],
				mockConditionJobRepo *datastoreMocks.MockGenericRepository[types.ConditionJobDataEntity],
			) {
				jobData := &types.JobDataEntity{
					JobID:            "12346",
					TaskDefinitionID: 2,
					JobTitle:         "Test Time Job 2",
					Status:           string(types.JobStatusRunning),
					CreatedAt:        now,
				}
				mockJobRepo.On("GetByID", mock.Anything, "12346").Return(jobData, nil).Once()

				timeJobData := &types.TimeJobDataEntity{
					JobID:            "12346",
					TaskDefinitionID: 2,
					TimeInterval:     7200,
					ScheduleType:     "cron",
					CronExpression:   "0 * * * *",
				}
				mockTimeJobRepo.On("GetByID", mock.Anything, "12346").Return(timeJobData, nil).Once()
			},
			expectedCode: http.StatusOK,
			validateResult: func(t *testing.T, w *httptest.ResponseRecorder) {
				var result types.CompleteJobDataDTO
				err := json.Unmarshal(w.Body.Bytes(), &result)
				assert.NoError(t, err)
				assert.Equal(t, "12346", result.JobDataDTO.JobID)
				assert.NotNil(t, result.TimeJobDataDTO)
			},
		},
		{
			name:  "Success - Event Job (TaskDefinitionID 3)",
			jobID: "12347",
			mockSetup: func(
				mockJobRepo *datastoreMocks.MockGenericRepository[types.JobDataEntity],
				mockTimeJobRepo *datastoreMocks.MockGenericRepository[types.TimeJobDataEntity],
				mockEventJobRepo *datastoreMocks.MockGenericRepository[types.EventJobDataEntity],
				mockConditionJobRepo *datastoreMocks.MockGenericRepository[types.ConditionJobDataEntity],
			) {
				jobData := &types.JobDataEntity{
					JobID:            "12347",
					TaskDefinitionID: 3,
					JobTitle:         "Test Event Job",
					Status:           string(types.JobStatusRunning),
					CreatedAt:        now,
				}
				mockJobRepo.On("GetByID", mock.Anything, "12347").Return(jobData, nil).Once()

				eventJobData := &types.EventJobDataEntity{
					JobID:                  "12347",
					TaskDefinitionID:       3,
					TriggerChainID:         "11155111",
					TriggerContractAddress: "0x1234567890123456789012345678901234567890",
					TriggerEvent:           "Transfer",
					Recurring:              true,
					TargetChainID:          "11155111",
					TargetContractAddress:  "0x0987654321098765432109876543210987654321",
					TargetFunction:         "handleTransfer",
					ExpirationTime:         now.Add(24 * time.Hour),
				}
				mockEventJobRepo.On("GetByID", mock.Anything, "12347").Return(eventJobData, nil).Once()
			},
			expectedCode: http.StatusOK,
			validateResult: func(t *testing.T, w *httptest.ResponseRecorder) {
				var result types.CompleteJobDataDTO
				err := json.Unmarshal(w.Body.Bytes(), &result)
				assert.NoError(t, err)
				assert.Equal(t, "12347", result.JobDataDTO.JobID)
				assert.NotNil(t, result.EventJobDataDTO)
				assert.Equal(t, "Transfer", result.EventJobDataDTO.TriggerEvent)
				assert.True(t, result.EventJobDataDTO.Recurring)
				assert.Nil(t, result.TimeJobDataDTO)
				assert.Nil(t, result.ConditionJobDataDTO)
			},
		},
		{
			name:  "Success - Event Job (TaskDefinitionID 4)",
			jobID: "12348",
			mockSetup: func(
				mockJobRepo *datastoreMocks.MockGenericRepository[types.JobDataEntity],
				mockTimeJobRepo *datastoreMocks.MockGenericRepository[types.TimeJobDataEntity],
				mockEventJobRepo *datastoreMocks.MockGenericRepository[types.EventJobDataEntity],
				mockConditionJobRepo *datastoreMocks.MockGenericRepository[types.ConditionJobDataEntity],
			) {
				jobData := &types.JobDataEntity{
					JobID:            "12348",
					TaskDefinitionID: 4,
					JobTitle:         "Test Event Job 4",
					Status:           string(types.JobStatusRunning),
					CreatedAt:        now,
				}
				mockJobRepo.On("GetByID", mock.Anything, "12348").Return(jobData, nil).Once()

				eventJobData := &types.EventJobDataEntity{
					JobID:            "12348",
					TaskDefinitionID: 4,
					TriggerEvent:     "Approval",
				}
				mockEventJobRepo.On("GetByID", mock.Anything, "12348").Return(eventJobData, nil).Once()
			},
			expectedCode: http.StatusOK,
			validateResult: func(t *testing.T, w *httptest.ResponseRecorder) {
				var result types.CompleteJobDataDTO
				err := json.Unmarshal(w.Body.Bytes(), &result)
				assert.NoError(t, err)
				assert.NotNil(t, result.EventJobDataDTO)
			},
		},
		{
			name:  "Success - Condition Job (TaskDefinitionID 5)",
			jobID: "12349",
			mockSetup: func(
				mockJobRepo *datastoreMocks.MockGenericRepository[types.JobDataEntity],
				mockTimeJobRepo *datastoreMocks.MockGenericRepository[types.TimeJobDataEntity],
				mockEventJobRepo *datastoreMocks.MockGenericRepository[types.EventJobDataEntity],
				mockConditionJobRepo *datastoreMocks.MockGenericRepository[types.ConditionJobDataEntity],
			) {
				jobData := &types.JobDataEntity{
					JobID:            "12349",
					TaskDefinitionID: 5,
					JobTitle:         "Test Condition Job",
					Status:           string(types.JobStatusRunning),
					CreatedAt:        now,
				}
				mockJobRepo.On("GetByID", mock.Anything, "12349").Return(jobData, nil).Once()

				conditionJobData := &types.ConditionJobDataEntity{
					JobID:            "12349",
					TaskDefinitionID: 5,
					ConditionType:    "price_threshold",
					UpperLimit:       100.0,
					LowerLimit:       50.0,
					ValueSourceType:  "api",
					ValueSourceURL:   "https://api.example.com/price",
					Recurring:        true,
					ExpirationTime:   now.Add(24 * time.Hour),
				}
				mockConditionJobRepo.On("GetByID", mock.Anything, "12349").Return(conditionJobData, nil).Once()
			},
			expectedCode: http.StatusOK,
			validateResult: func(t *testing.T, w *httptest.ResponseRecorder) {
				var result types.CompleteJobDataDTO
				err := json.Unmarshal(w.Body.Bytes(), &result)
				assert.NoError(t, err)
				assert.Equal(t, "12349", result.JobDataDTO.JobID)
				assert.NotNil(t, result.ConditionJobDataDTO)
				assert.Equal(t, "price_threshold", result.ConditionJobDataDTO.ConditionType)
				assert.Equal(t, 100.0, result.ConditionJobDataDTO.UpperLimit)
				assert.Equal(t, 50.0, result.ConditionJobDataDTO.LowerLimit)
				assert.Nil(t, result.TimeJobDataDTO)
				assert.Nil(t, result.EventJobDataDTO)
			},
		},
		{
			name:  "Success - Condition Job (TaskDefinitionID 6)",
			jobID: "12350",
			mockSetup: func(
				mockJobRepo *datastoreMocks.MockGenericRepository[types.JobDataEntity],
				mockTimeJobRepo *datastoreMocks.MockGenericRepository[types.TimeJobDataEntity],
				mockEventJobRepo *datastoreMocks.MockGenericRepository[types.EventJobDataEntity],
				mockConditionJobRepo *datastoreMocks.MockGenericRepository[types.ConditionJobDataEntity],
			) {
				jobData := &types.JobDataEntity{
					JobID:            "12350",
					TaskDefinitionID: 6,
					JobTitle:         "Test Condition Job 6",
					Status:           string(types.JobStatusRunning),
					CreatedAt:        now,
				}
				mockJobRepo.On("GetByID", mock.Anything, "12350").Return(jobData, nil).Once()

				conditionJobData := &types.ConditionJobDataEntity{
					JobID:            "12350",
					TaskDefinitionID: 6,
					ConditionType:    "balance_check",
				}
				mockConditionJobRepo.On("GetByID", mock.Anything, "12350").Return(conditionJobData, nil).Once()
			},
			expectedCode: http.StatusOK,
			validateResult: func(t *testing.T, w *httptest.ResponseRecorder) {
				var result types.CompleteJobDataDTO
				err := json.Unmarshal(w.Body.Bytes(), &result)
				assert.NoError(t, err)
				assert.NotNil(t, result.ConditionJobDataDTO)
			},
		},
		{
			name:  "Error - Invalid Job ID (empty)",
			jobID: "",
			mockSetup: func(
				mockJobRepo *datastoreMocks.MockGenericRepository[types.JobDataEntity],
				mockTimeJobRepo *datastoreMocks.MockGenericRepository[types.TimeJobDataEntity],
				mockEventJobRepo *datastoreMocks.MockGenericRepository[types.EventJobDataEntity],
				mockConditionJobRepo *datastoreMocks.MockGenericRepository[types.ConditionJobDataEntity],
			) {
				// No mock setup needed
			},
			expectedCode:  http.StatusBadRequest,
			expectedError: pkgErrors.ErrInvalidRequestBody,
		},
		{
			name:  "Error - Job Not Found",
			jobID: "99999",
			mockSetup: func(
				mockJobRepo *datastoreMocks.MockGenericRepository[types.JobDataEntity],
				mockTimeJobRepo *datastoreMocks.MockGenericRepository[types.TimeJobDataEntity],
				mockEventJobRepo *datastoreMocks.MockGenericRepository[types.EventJobDataEntity],
				mockConditionJobRepo *datastoreMocks.MockGenericRepository[types.ConditionJobDataEntity],
			) {
				mockJobRepo.On("GetByID", mock.Anything, "99999").Return(nil, errors.New("not found")).Once()
			},
			expectedCode:  http.StatusNotFound,
			expectedError: "job not found",
		},
		{
			name:  "Error - Job Data is Nil",
			jobID: "88888",
			mockSetup: func(
				mockJobRepo *datastoreMocks.MockGenericRepository[types.JobDataEntity],
				mockTimeJobRepo *datastoreMocks.MockGenericRepository[types.TimeJobDataEntity],
				mockEventJobRepo *datastoreMocks.MockGenericRepository[types.EventJobDataEntity],
				mockConditionJobRepo *datastoreMocks.MockGenericRepository[types.ConditionJobDataEntity],
			) {
				mockJobRepo.On("GetByID", mock.Anything, "88888").Return(nil, nil).Once()
			},
			expectedCode:  http.StatusNotFound,
			expectedError: "job not found",
		},
		{
			name:  "Error - Time Job Not Found",
			jobID: "77777",
			mockSetup: func(
				mockJobRepo *datastoreMocks.MockGenericRepository[types.JobDataEntity],
				mockTimeJobRepo *datastoreMocks.MockGenericRepository[types.TimeJobDataEntity],
				mockEventJobRepo *datastoreMocks.MockGenericRepository[types.EventJobDataEntity],
				mockConditionJobRepo *datastoreMocks.MockGenericRepository[types.ConditionJobDataEntity],
			) {
				jobData := &types.JobDataEntity{
					JobID:            "77777",
					TaskDefinitionID: 1,
					JobTitle:         "Test Job",
					CreatedAt:        now,
				}
				mockJobRepo.On("GetByID", mock.Anything, "77777").Return(jobData, nil).Once()
				mockTimeJobRepo.On("GetByID", mock.Anything, "77777").Return(nil, errors.New("not found")).Once()
			},
			expectedCode:  http.StatusNotFound,
			expectedError: "time job not found",
		},
		{
			name:  "Error - Event Job Not Found",
			jobID: "66666",
			mockSetup: func(
				mockJobRepo *datastoreMocks.MockGenericRepository[types.JobDataEntity],
				mockTimeJobRepo *datastoreMocks.MockGenericRepository[types.TimeJobDataEntity],
				mockEventJobRepo *datastoreMocks.MockGenericRepository[types.EventJobDataEntity],
				mockConditionJobRepo *datastoreMocks.MockGenericRepository[types.ConditionJobDataEntity],
			) {
				jobData := &types.JobDataEntity{
					JobID:            "66666",
					TaskDefinitionID: 3,
					JobTitle:         "Test Job",
					CreatedAt:        now,
				}
				mockJobRepo.On("GetByID", mock.Anything, "66666").Return(jobData, nil).Once()
				mockEventJobRepo.On("GetByID", mock.Anything, "66666").Return(nil, errors.New("not found")).Once()
			},
			expectedCode:  http.StatusNotFound,
			expectedError: "event job not found",
		},
		{
			name:  "Error - Condition Job Not Found",
			jobID: "55555",
			mockSetup: func(
				mockJobRepo *datastoreMocks.MockGenericRepository[types.JobDataEntity],
				mockTimeJobRepo *datastoreMocks.MockGenericRepository[types.TimeJobDataEntity],
				mockEventJobRepo *datastoreMocks.MockGenericRepository[types.EventJobDataEntity],
				mockConditionJobRepo *datastoreMocks.MockGenericRepository[types.ConditionJobDataEntity],
			) {
				jobData := &types.JobDataEntity{
					JobID:            "55555",
					TaskDefinitionID: 5,
					JobTitle:         "Test Job",
					CreatedAt:        now,
				}
				mockJobRepo.On("GetByID", mock.Anything, "55555").Return(jobData, nil).Once()
				mockConditionJobRepo.On("GetByID", mock.Anything, "55555").Return(nil, errors.New("not found")).Once()
			},
			expectedCode:  http.StatusNotFound,
			expectedError: "condition job not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockJobRepo, mockTimeJobRepo, mockEventJobRepo, mockConditionJobRepo, _, mockLogger := setupJobReadTestHandler()
			defer mockLogger.AssertExpectations(t)
			defer mockJobRepo.AssertExpectations(t)
			defer mockTimeJobRepo.AssertExpectations(t)
			defer mockEventJobRepo.AssertExpectations(t)
			defer mockConditionJobRepo.AssertExpectations(t)

			// Setup
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("GET", "/jobs/"+tt.jobID, nil)
			c.Params = gin.Params{gin.Param{Key: "job_id", Value: tt.jobID}}

			// Setup mocks
			tt.mockSetup(mockJobRepo, mockTimeJobRepo, mockEventJobRepo, mockConditionJobRepo)

			// Execute
			handler.GetJobDataByJobID(c)

			// Assert
			assert.Equal(t, tt.expectedCode, w.Code)
			if tt.validateResult != nil {
				tt.validateResult(t, w)
			}
			if tt.expectedError != "" {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response["error"], tt.expectedError)
			}
		})
	}
}

func TestGetJobsByUserAddress(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name           string
		userAddress    string
		chainIDFilter  string
		mockSetup      func(*datastoreMocks.MockGenericRepository[types.UserDataEntity], *datastoreMocks.MockGenericRepository[types.JobDataEntity], *datastoreMocks.MockGenericRepository[types.TimeJobDataEntity], *datastoreMocks.MockGenericRepository[types.EventJobDataEntity], *datastoreMocks.MockGenericRepository[types.ConditionJobDataEntity])
		expectedCode   int
		expectedError  string
		validateResult func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:        "Success - User with Multiple Jobs",
			userAddress: "0x1234567890123456789012345678901234567890",
			mockSetup: func(
				mockUserRepo *datastoreMocks.MockGenericRepository[types.UserDataEntity],
				mockJobRepo *datastoreMocks.MockGenericRepository[types.JobDataEntity],
				mockTimeJobRepo *datastoreMocks.MockGenericRepository[types.TimeJobDataEntity],
				mockEventJobRepo *datastoreMocks.MockGenericRepository[types.EventJobDataEntity],
				mockConditionJobRepo *datastoreMocks.MockGenericRepository[types.ConditionJobDataEntity],
			) {
				userData := &types.UserDataEntity{
					UserAddress: "0x1234567890123456789012345678901234567890",
					JobIDs:      []string{"job1", "job2", "job3"},
					TotalJobs:   3,
				}
				mockUserRepo.On("GetByID", mock.Anything, "0x1234567890123456789012345678901234567890").Return(userData, nil).Once()

				// Job 1 - Time job
				job1 := &types.JobDataEntity{
					JobID:            "job1",
					TaskDefinitionID: 1,
					JobTitle:         "Time Job",
					CreatedChainID:   "11155111",
					Status:           string(types.JobStatusRunning),
					CreatedAt:        now,
				}
				mockJobRepo.On("GetByID", mock.Anything, "job1").Return(job1, nil).Once()
				timeJob1 := &types.TimeJobDataEntity{
					JobID:        "job1",
					TimeInterval: 3600,
				}
				mockTimeJobRepo.On("GetByID", mock.Anything, "job1").Return(timeJob1, nil).Once()

				// Job 2 - Event job
				job2 := &types.JobDataEntity{
					JobID:            "job2",
					TaskDefinitionID: 3,
					JobTitle:         "Event Job",
					CreatedChainID:   "11155111",
					Status:           string(types.JobStatusRunning),
					CreatedAt:        now,
				}
				mockJobRepo.On("GetByID", mock.Anything, "job2").Return(job2, nil).Once()
				eventJob2 := &types.EventJobDataEntity{
					JobID:        "job2",
					TriggerEvent: "Transfer",
				}
				mockEventJobRepo.On("GetByID", mock.Anything, "job2").Return(eventJob2, nil).Once()

				// Job 3 - Condition job
				job3 := &types.JobDataEntity{
					JobID:            "job3",
					TaskDefinitionID: 5,
					JobTitle:         "Condition Job",
					CreatedChainID:   "11155111",
					Status:           string(types.JobStatusRunning),
					CreatedAt:        now,
				}
				mockJobRepo.On("GetByID", mock.Anything, "job3").Return(job3, nil).Once()
				conditionJob3 := &types.ConditionJobDataEntity{
					JobID:         "job3",
					ConditionType: "price",
				}
				mockConditionJobRepo.On("GetByID", mock.Anything, "job3").Return(conditionJob3, nil).Once()
			},
			expectedCode: http.StatusOK,
			validateResult: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				jobs := response["jobs"].([]interface{})
				assert.Len(t, jobs, 3)
			},
		},
		{
			name:          "Success - User with Jobs Filtered by Chain ID",
			userAddress:   "0x1234567890123456789012345678901234567890",
			chainIDFilter: "1",
			mockSetup: func(
				mockUserRepo *datastoreMocks.MockGenericRepository[types.UserDataEntity],
				mockJobRepo *datastoreMocks.MockGenericRepository[types.JobDataEntity],
				mockTimeJobRepo *datastoreMocks.MockGenericRepository[types.TimeJobDataEntity],
				mockEventJobRepo *datastoreMocks.MockGenericRepository[types.EventJobDataEntity],
				mockConditionJobRepo *datastoreMocks.MockGenericRepository[types.ConditionJobDataEntity],
			) {
				userData := &types.UserDataEntity{
					UserAddress: "0x1234567890123456789012345678901234567890",
					JobIDs:      []string{"job1", "job2"},
					TotalJobs:   2,
				}
				mockUserRepo.On("GetByID", mock.Anything, "0x1234567890123456789012345678901234567890").Return(userData, nil).Once()

				// Job 1 - matches filter (chain_id = 1)
				job1 := &types.JobDataEntity{
					JobID:            "job1",
					TaskDefinitionID: 1,
					JobTitle:         "Time Job",
					CreatedChainID:   "1",
					Status:           string(types.JobStatusRunning),
					CreatedAt:        now,
				}
				mockJobRepo.On("GetByID", mock.Anything, "job1").Return(job1, nil).Once()
				timeJob1 := &types.TimeJobDataEntity{
					JobID:        "job1",
					TimeInterval: 3600,
				}
				mockTimeJobRepo.On("GetByID", mock.Anything, "job1").Return(timeJob1, nil).Once()

				// Job 2 - doesn't match filter (chain_id = 11155111)
				job2 := &types.JobDataEntity{
					JobID:            "job2",
					TaskDefinitionID: 1,
					JobTitle:         "Another Time Job",
					CreatedChainID:   "11155111",
					Status:           string(types.JobStatusRunning),
					CreatedAt:        now,
				}
				mockJobRepo.On("GetByID", mock.Anything, "job2").Return(job2, nil).Once()
			},
			expectedCode: http.StatusOK,
			validateResult: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				jobs := response["jobs"].([]interface{})
				// Only job1 should be returned (chain_id = 1)
				assert.Len(t, jobs, 1)
			},
		},
		{
			name:        "Success - User with No Jobs",
			userAddress: "0x1234567890123456789012345678901234567890",
			mockSetup: func(
				mockUserRepo *datastoreMocks.MockGenericRepository[types.UserDataEntity],
				mockJobRepo *datastoreMocks.MockGenericRepository[types.JobDataEntity],
				mockTimeJobRepo *datastoreMocks.MockGenericRepository[types.TimeJobDataEntity],
				mockEventJobRepo *datastoreMocks.MockGenericRepository[types.EventJobDataEntity],
				mockConditionJobRepo *datastoreMocks.MockGenericRepository[types.ConditionJobDataEntity],
			) {
				userData := &types.UserDataEntity{
					UserAddress: "0x1234567890123456789012345678901234567890",
					JobIDs:      []string{},
					TotalJobs:   0,
				}
				mockUserRepo.On("GetByID", mock.Anything, "0x1234567890123456789012345678901234567890").Return(userData, nil).Once()
			},
			expectedCode: http.StatusOK,
			validateResult: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, pkgErrors.ErrDBRecordNotFound, response["message"])
				jobs := response["jobs"].([]interface{})
				assert.Len(t, jobs, 0)
			},
		},
		{
			name:        "Success - User Not Found",
			userAddress: "0x0987654321098765432109876543210987654321",
			mockSetup: func(
				mockUserRepo *datastoreMocks.MockGenericRepository[types.UserDataEntity],
				mockJobRepo *datastoreMocks.MockGenericRepository[types.JobDataEntity],
				mockTimeJobRepo *datastoreMocks.MockGenericRepository[types.TimeJobDataEntity],
				mockEventJobRepo *datastoreMocks.MockGenericRepository[types.EventJobDataEntity],
				mockConditionJobRepo *datastoreMocks.MockGenericRepository[types.ConditionJobDataEntity],
			) {
				mockUserRepo.On("GetByID", mock.Anything, "0x0987654321098765432109876543210987654321").Return(nil, nil).Once()
			},
			expectedCode: http.StatusOK,
			validateResult: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, pkgErrors.ErrDBRecordNotFound, response["message"])
				jobs := response["jobs"].([]interface{})
				assert.Len(t, jobs, 0)
			},
		},
		{
			name:        "Partial Success - Some Jobs Failed to Load",
			userAddress: "0x1234567890123456789012345678901234567890",
			mockSetup: func(
				mockUserRepo *datastoreMocks.MockGenericRepository[types.UserDataEntity],
				mockJobRepo *datastoreMocks.MockGenericRepository[types.JobDataEntity],
				mockTimeJobRepo *datastoreMocks.MockGenericRepository[types.TimeJobDataEntity],
				mockEventJobRepo *datastoreMocks.MockGenericRepository[types.EventJobDataEntity],
				mockConditionJobRepo *datastoreMocks.MockGenericRepository[types.ConditionJobDataEntity],
			) {
				userData := &types.UserDataEntity{
					UserAddress: "0x1234567890123456789012345678901234567890",
					JobIDs:      []string{"job1", "job2", "job3"},
					TotalJobs:   3,
				}
				mockUserRepo.On("GetByID", mock.Anything, "0x1234567890123456789012345678901234567890").Return(userData, nil).Once()

				// Job 1 - succeeds
				job1 := &types.JobDataEntity{
					JobID:            "job1",
					TaskDefinitionID: 1,
					JobTitle:         "Time Job",
					CreatedChainID:   "11155111",
					Status:           string(types.JobStatusRunning),
					CreatedAt:        now,
				}
				mockJobRepo.On("GetByID", mock.Anything, "job1").Return(job1, nil).Once()
				timeJob1 := &types.TimeJobDataEntity{
					JobID:        "job1",
					TimeInterval: 3600,
				}
				mockTimeJobRepo.On("GetByID", mock.Anything, "job1").Return(timeJob1, nil).Once()

				// Job 2 - fails
				mockJobRepo.On("GetByID", mock.Anything, "job2").Return(nil, errors.New("not found")).Once()

				// Job 3 - fails with nil
				mockJobRepo.On("GetByID", mock.Anything, "job3").Return(nil, nil).Once()
			},
			expectedCode: http.StatusPartialContent,
			validateResult: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				jobs := response["jobs"].([]interface{})
				assert.Len(t, jobs, 1)
				assert.Contains(t, response["message"], "Some jobs were retrieved")
			},
		},
		{
			name:        "Error - All Jobs Failed to Load",
			userAddress: "0x1234567890123456789012345678901234567890",
			mockSetup: func(
				mockUserRepo *datastoreMocks.MockGenericRepository[types.UserDataEntity],
				mockJobRepo *datastoreMocks.MockGenericRepository[types.JobDataEntity],
				mockTimeJobRepo *datastoreMocks.MockGenericRepository[types.TimeJobDataEntity],
				mockEventJobRepo *datastoreMocks.MockGenericRepository[types.EventJobDataEntity],
				mockConditionJobRepo *datastoreMocks.MockGenericRepository[types.ConditionJobDataEntity],
			) {
				userData := &types.UserDataEntity{
					UserAddress: "0x1234567890123456789012345678901234567890",
					JobIDs:      []string{"job1", "job2"},
					TotalJobs:   2,
				}
				mockUserRepo.On("GetByID", mock.Anything, "0x1234567890123456789012345678901234567890").Return(userData, nil).Once()

				// Job 1 - fails
				mockJobRepo.On("GetByID", mock.Anything, "job1").Return(nil, errors.New("not found")).Once()

				// Job 2 - fails
				mockJobRepo.On("GetByID", mock.Anything, "job2").Return(nil, errors.New("not found")).Once()
			},
			expectedCode:  http.StatusInternalServerError,
			expectedError: pkgErrors.ErrDBOperationFailed,
		},
		{
			name:        "Error - Invalid User Address (empty)",
			userAddress: "",
			mockSetup: func(
				mockUserRepo *datastoreMocks.MockGenericRepository[types.UserDataEntity],
				mockJobRepo *datastoreMocks.MockGenericRepository[types.JobDataEntity],
				mockTimeJobRepo *datastoreMocks.MockGenericRepository[types.TimeJobDataEntity],
				mockEventJobRepo *datastoreMocks.MockGenericRepository[types.EventJobDataEntity],
				mockConditionJobRepo *datastoreMocks.MockGenericRepository[types.ConditionJobDataEntity],
			) {
				// No mock setup needed
			},
			expectedCode:  http.StatusBadRequest,
			expectedError: pkgErrors.ErrInvalidRequestBody,
		},
		{
			name:        "Error - User Repository Error",
			userAddress: "0x1234567890123456789012345678901234567890",
			mockSetup: func(
				mockUserRepo *datastoreMocks.MockGenericRepository[types.UserDataEntity],
				mockJobRepo *datastoreMocks.MockGenericRepository[types.JobDataEntity],
				mockTimeJobRepo *datastoreMocks.MockGenericRepository[types.TimeJobDataEntity],
				mockEventJobRepo *datastoreMocks.MockGenericRepository[types.EventJobDataEntity],
				mockConditionJobRepo *datastoreMocks.MockGenericRepository[types.ConditionJobDataEntity],
			) {
				mockUserRepo.On("GetByID", mock.Anything, "0x1234567890123456789012345678901234567890").Return(nil, errors.New("database error")).Once()
			},
			expectedCode:  http.StatusOK,
			expectedError: pkgErrors.ErrDBOperationFailed,
		},
		{
			name:        "Success - Unknown Task Definition ID",
			userAddress: "0x1234567890123456789012345678901234567890",
			mockSetup: func(
				mockUserRepo *datastoreMocks.MockGenericRepository[types.UserDataEntity],
				mockJobRepo *datastoreMocks.MockGenericRepository[types.JobDataEntity],
				mockTimeJobRepo *datastoreMocks.MockGenericRepository[types.TimeJobDataEntity],
				mockEventJobRepo *datastoreMocks.MockGenericRepository[types.EventJobDataEntity],
				mockConditionJobRepo *datastoreMocks.MockGenericRepository[types.ConditionJobDataEntity],
			) {
				userData := &types.UserDataEntity{
					UserAddress: "0x1234567890123456789012345678901234567890",
					JobIDs:      []string{"job1"},
					TotalJobs:   1,
				}
				mockUserRepo.On("GetByID", mock.Anything, "0x1234567890123456789012345678901234567890").Return(userData, nil).Once()

				// Job with unknown task definition ID
				job1 := &types.JobDataEntity{
					JobID:            "job1",
					TaskDefinitionID: 99, // Unknown task definition
					JobTitle:         "Unknown Job",
					CreatedChainID:   "11155111",
					Status:           string(types.JobStatusRunning),
					CreatedAt:        now,
				}
				mockJobRepo.On("GetByID", mock.Anything, "job1").Return(job1, nil).Once()
			},
			expectedCode: http.StatusInternalServerError,
			validateResult: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, pkgErrors.ErrDBOperationFailed, response["error"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockJobRepo, mockTimeJobRepo, mockEventJobRepo, mockConditionJobRepo, mockUserRepo, mockLogger := setupJobReadTestHandler()
			defer mockLogger.AssertExpectations(t)
			defer mockJobRepo.AssertExpectations(t)
			defer mockTimeJobRepo.AssertExpectations(t)
			defer mockEventJobRepo.AssertExpectations(t)
			defer mockConditionJobRepo.AssertExpectations(t)
			defer mockUserRepo.AssertExpectations(t)

			// Setup
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			url := "/users/" + tt.userAddress + "/jobs"
			if tt.chainIDFilter != "" {
				url += "?chain_id=" + tt.chainIDFilter
			}
			c.Request = httptest.NewRequest("GET", url, nil)
			c.Params = gin.Params{gin.Param{Key: "user_address", Value: tt.userAddress}}

			// Setup mocks
			tt.mockSetup(mockUserRepo, mockJobRepo, mockTimeJobRepo, mockEventJobRepo, mockConditionJobRepo)

			// Execute
			handler.GetJobsByUserAddress(c)

			// Assert
			assert.Equal(t, tt.expectedCode, w.Code)
			if tt.validateResult != nil {
				tt.validateResult(t, w)
			}
			if tt.expectedError != "" {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				// Check either 'error' or 'message' field
				if errMsg, ok := response["error"]; ok {
					assert.Contains(t, errMsg, tt.expectedError)
				} else if msg, ok := response["message"]; ok {
					assert.Contains(t, msg, tt.expectedError)
				}
			}
		})
	}
}
