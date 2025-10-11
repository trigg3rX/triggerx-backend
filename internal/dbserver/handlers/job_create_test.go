package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	datastoreMocks "github.com/trigg3rX/triggerx-backend/pkg/datastore/mocks"
	pkgErrors "github.com/trigg3rX/triggerx-backend/pkg/errors"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

func setupJobCreateTestHandler() (
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

func TestCreateJobData(t *testing.T) {
	tests := []struct {
		name           string
		request        []types.CreateJobDataRequest
		mockSetup      func(*datastoreMocks.MockGenericRepository[types.UserDataEntity], *datastoreMocks.MockGenericRepository[types.JobDataEntity], *datastoreMocks.MockGenericRepository[types.TimeJobDataEntity], *datastoreMocks.MockGenericRepository[types.EventJobDataEntity], *datastoreMocks.MockGenericRepository[types.ConditionJobDataEntity])
		expectedCode   int
		expectedError  string
		validateResult func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "Success - Create Time Job (TaskDefinitionID 1) for New User",
			request: []types.CreateJobDataRequest{
				{
					JobID:                 "job123",
					JobTitle:              "Test Time Job",
					TaskDefinitionID:      1,
					CreatedChainID:        "11155111",
					UserAddress:           "0x1234567890123456789012345678901234567890",
					Timezone:              "UTC",
					IsImua:                false,
					JobType:               "frontend",
					TimeFrame:             86400,
					Recurring:             true,
					JobCostPrediction:     "1000000000000000000",
					ScheduleType:          "interval",
					TimeInterval:          3600,
					TargetChainID:         "11155111",
					TargetContractAddress: "0x1234567890123456789012345678901234567890",
					TargetFunction:        "testFunction",
					ABI:                   `[{"name":"testFunction","type":"function"}]`,
					ArgType:               1,
					Arguments:             []string{"arg1", "arg2"},
				},
			},
			mockSetup: func(
				mockUserRepo *datastoreMocks.MockGenericRepository[types.UserDataEntity],
				mockJobRepo *datastoreMocks.MockGenericRepository[types.JobDataEntity],
				mockTimeJobRepo *datastoreMocks.MockGenericRepository[types.TimeJobDataEntity],
				mockEventJobRepo *datastoreMocks.MockGenericRepository[types.EventJobDataEntity],
				mockConditionJobRepo *datastoreMocks.MockGenericRepository[types.ConditionJobDataEntity],
			) {
				// User doesn't exist
				mockUserRepo.On("GetByID", mock.Anything, "0x1234567890123456789012345678901234567890").Return(nil, nil).Once()
				// Create new user
				mockUserRepo.On("Create", mock.Anything, mock.MatchedBy(func(u *types.UserDataEntity) bool {
					return u.UserAddress == "0x1234567890123456789012345678901234567890"
				})).Return(nil).Once()
				// Create job
				mockJobRepo.On("Create", mock.Anything, mock.MatchedBy(func(j *types.JobDataEntity) bool {
					return j.JobID == "job123" && j.TaskDefinitionID == 1 && j.Status == string(types.JobStatusRunning)
				})).Return(nil).Once()
				// Create time job
				mockTimeJobRepo.On("Create", mock.Anything, mock.MatchedBy(func(tj *types.TimeJobDataEntity) bool {
					return tj.JobID == "job123" && tj.TimeInterval == 3600
				})).Return(nil).Once()
				// Update user with job IDs
				mockUserRepo.On("Update", mock.Anything, mock.MatchedBy(func(u *types.UserDataEntity) bool {
					return len(u.JobIDs) == 1 && u.JobIDs[0] == "job123" && u.TotalJobs == 1
				})).Return(nil).Once()
			},
			expectedCode: http.StatusOK,
			validateResult: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response types.CreateJobResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Len(t, response.JobIDs, 1)
				assert.Equal(t, "job123", response.JobIDs[0])
				assert.Equal(t, 1, response.TaskDefinitionIDs[0])
				assert.Equal(t, int64(86400), response.TimeFrames[0])
			},
		},
		{
			name: "Success - Create Time Job (TaskDefinitionID 2) for Existing User",
			request: []types.CreateJobDataRequest{
				{
					JobID:                 "job456",
					JobTitle:              "Test Time Job 2",
					TaskDefinitionID:      2,
					CreatedChainID:        "11155111",
					UserAddress:           "0x1234567890123456789012345678901234567890",
					Timezone:              "UTC",
					JobType:               "sdk",
					TimeFrame:             43200,
					Recurring:             false,
					JobCostPrediction:     "2000000000000000000",
					ScheduleType:          "cron",
					CronExpression:        "0 * * * *",
					TimeInterval:          3600,
					TargetChainID:         "11155111",
					TargetContractAddress: "0x1234567890123456789012345678901234567890",
					TargetFunction:        "testFunction",
					ABI:                   `[{"name":"testFunction","type":"function"}]`,
					ArgType:               1,
				},
			},
			mockSetup: func(
				mockUserRepo *datastoreMocks.MockGenericRepository[types.UserDataEntity],
				mockJobRepo *datastoreMocks.MockGenericRepository[types.JobDataEntity],
				mockTimeJobRepo *datastoreMocks.MockGenericRepository[types.TimeJobDataEntity],
				mockEventJobRepo *datastoreMocks.MockGenericRepository[types.EventJobDataEntity],
				mockConditionJobRepo *datastoreMocks.MockGenericRepository[types.ConditionJobDataEntity],
			) {
				// Existing user
				existingUser := &types.UserDataEntity{
					UserAddress: "0x1234567890123456789012345678901234567890",
					JobIDs:      []string{"oldJob1", "oldJob2"},
					TotalJobs:   2,
				}
				mockUserRepo.On("GetByID", mock.Anything, "0x1234567890123456789012345678901234567890").Return(existingUser, nil).Once()
				mockJobRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Once()
				mockTimeJobRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Once()
				mockUserRepo.On("Update", mock.Anything, mock.MatchedBy(func(u *types.UserDataEntity) bool {
					return len(u.JobIDs) == 3 && u.TotalJobs == 3
				})).Return(nil).Once()
			},
			expectedCode: http.StatusOK,
			validateResult: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response types.CreateJobResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Len(t, response.JobIDs, 1)
				assert.Equal(t, "job456", response.JobIDs[0])
			},
		},
		{
			name: "Success - Create Event Job (TaskDefinitionID 3) - Note: notifyConditionScheduler will fail in test",
			request: []types.CreateJobDataRequest{
				{
					JobID:                  "job789",
					JobTitle:               "Test Event Job",
					TaskDefinitionID:       3,
					CreatedChainID:         "11155111",
					UserAddress:            "0x1234567890123456789012345678901234567890",
					Timezone:               "UTC",
					JobType:                "frontend",
					TimeFrame:              86400,
					Recurring:              true,
					JobCostPrediction:      "1000000000000000000",
					TriggerChainID:         "11155111",
					TriggerContractAddress: "0x1234567890123456789012345678901234567890",
					TriggerEvent:           "Transfer",
					TargetChainID:          "11155111",
					TargetContractAddress:  "0x0987654321098765432109876543210987654321",
					TargetFunction:         "handleTransfer",
					ABI:                    `[{"name":"handleTransfer","type":"function"}]`,
					ArgType:                1,
				},
			},
			mockSetup: func(
				mockUserRepo *datastoreMocks.MockGenericRepository[types.UserDataEntity],
				mockJobRepo *datastoreMocks.MockGenericRepository[types.JobDataEntity],
				mockTimeJobRepo *datastoreMocks.MockGenericRepository[types.TimeJobDataEntity],
				mockEventJobRepo *datastoreMocks.MockGenericRepository[types.EventJobDataEntity],
				mockConditionJobRepo *datastoreMocks.MockGenericRepository[types.ConditionJobDataEntity],
			) {
				mockUserRepo.On("GetByID", mock.Anything, "0x1234567890123456789012345678901234567890").Return(nil, nil).Once()
				mockUserRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Once()
				mockJobRepo.On("Create", mock.Anything, mock.MatchedBy(func(j *types.JobDataEntity) bool {
					return j.TaskDefinitionID == 3 && j.Status == string(types.JobStatusCreated)
				})).Return(nil).Once()
				mockEventJobRepo.On("Create", mock.Anything, mock.MatchedBy(func(ej *types.EventJobDataEntity) bool {
					return ej.JobID == "job789" && ej.TriggerEvent == "Transfer"
				})).Return(nil).Once()
				// Job status will be updated to running after successful scheduler notification (which fails in test)
				mockJobRepo.On("Update", mock.Anything, mock.Anything).Return(nil).Maybe()
				mockUserRepo.On("Update", mock.Anything, mock.Anything).Return(nil).Once()
			},
			expectedCode: http.StatusOK,
			validateResult: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response types.CreateJobResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "job789", response.JobIDs[0])
				assert.Equal(t, 3, response.TaskDefinitionIDs[0])
			},
		},
		{
			name: "Success - Create Event Job (TaskDefinitionID 4)",
			request: []types.CreateJobDataRequest{
				{
					JobID:                  "job101",
					JobTitle:               "Test Event Job 4",
					TaskDefinitionID:       4,
					CreatedChainID:         "11155111",
					UserAddress:            "0x1234567890123456789012345678901234567890",
					Timezone:               "UTC",
					JobType:                "contract",
					TimeFrame:              86400,
					Recurring:              false,
					JobCostPrediction:      "1000000000000000000",
					TriggerChainID:         "11155111",
					TriggerContractAddress: "0x1234567890123456789012345678901234567890",
					TriggerEvent:           "Approval",
					TargetChainID:          "11155111",
					TargetContractAddress:  "0x0987654321098765432109876543210987654321",
					TargetFunction:         "handleApproval",
					ABI:                    `[{"name":"handleApproval","type":"function"}]`,
					ArgType:                1,
				},
			},
			mockSetup: func(
				mockUserRepo *datastoreMocks.MockGenericRepository[types.UserDataEntity],
				mockJobRepo *datastoreMocks.MockGenericRepository[types.JobDataEntity],
				mockTimeJobRepo *datastoreMocks.MockGenericRepository[types.TimeJobDataEntity],
				mockEventJobRepo *datastoreMocks.MockGenericRepository[types.EventJobDataEntity],
				mockConditionJobRepo *datastoreMocks.MockGenericRepository[types.ConditionJobDataEntity],
			) {
				mockUserRepo.On("GetByID", mock.Anything, "0x1234567890123456789012345678901234567890").Return(nil, nil).Once()
				mockUserRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Once()
				mockJobRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Once()
				mockEventJobRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Once()
				mockJobRepo.On("Update", mock.Anything, mock.Anything).Return(nil).Maybe()
				mockUserRepo.On("Update", mock.Anything, mock.Anything).Return(nil).Once()
			},
			expectedCode: http.StatusOK,
			validateResult: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response types.CreateJobResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "job101", response.JobIDs[0])
			},
		},
		{
			name: "Success - Create Condition Job (TaskDefinitionID 5)",
			request: []types.CreateJobDataRequest{
				{
					JobID:                 "job202",
					JobTitle:              "Test Condition Job",
					TaskDefinitionID:      5,
					CreatedChainID:        "11155111",
					UserAddress:           "0x1234567890123456789012345678901234567890",
					Timezone:              "UTC",
					JobType:               "frontend",
					TimeFrame:             86400,
					Recurring:             true,
					JobCostPrediction:     "1000000000000000000",
					ConditionType:         "price_threshold",
					UpperLimit:            100.0,
					LowerLimit:            50.0,
					ValueSourceType:       "api",
					ValueSourceUrl:        "https://api.example.com/price",
					SelectedKeyRoute:      "price",
					TargetChainID:         "11155111",
					TargetContractAddress: "0x1234567890123456789012345678901234567890",
					TargetFunction:        "executeOnCondition",
					ABI:                   `[{"name":"executeOnCondition","type":"function"}]`,
					ArgType:               1,
				},
			},
			mockSetup: func(
				mockUserRepo *datastoreMocks.MockGenericRepository[types.UserDataEntity],
				mockJobRepo *datastoreMocks.MockGenericRepository[types.JobDataEntity],
				mockTimeJobRepo *datastoreMocks.MockGenericRepository[types.TimeJobDataEntity],
				mockEventJobRepo *datastoreMocks.MockGenericRepository[types.EventJobDataEntity],
				mockConditionJobRepo *datastoreMocks.MockGenericRepository[types.ConditionJobDataEntity],
			) {
				mockUserRepo.On("GetByID", mock.Anything, "0x1234567890123456789012345678901234567890").Return(nil, nil).Once()
				mockUserRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Once()
				mockJobRepo.On("Create", mock.Anything, mock.MatchedBy(func(j *types.JobDataEntity) bool {
					return j.TaskDefinitionID == 5 && j.Status == string(types.JobStatusCreated)
				})).Return(nil).Once()
				mockConditionJobRepo.On("Create", mock.Anything, mock.MatchedBy(func(cj *types.ConditionJobDataEntity) bool {
					return cj.JobID == "job202" && cj.ConditionType == "price_threshold"
				})).Return(nil).Once()
				mockJobRepo.On("Update", mock.Anything, mock.Anything).Return(nil).Maybe()
				mockUserRepo.On("Update", mock.Anything, mock.Anything).Return(nil).Once()
			},
			expectedCode: http.StatusOK,
			validateResult: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response types.CreateJobResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "job202", response.JobIDs[0])
				assert.Equal(t, 5, response.TaskDefinitionIDs[0])
			},
		},
		{
			name: "Success - Create Condition Job (TaskDefinitionID 6)",
			request: []types.CreateJobDataRequest{
				{
					JobID:                 "job303",
					JobTitle:              "Test Condition Job 6",
					TaskDefinitionID:      6,
					CreatedChainID:        "11155111",
					UserAddress:           "0x1234567890123456789012345678901234567890",
					Timezone:              "UTC",
					JobType:               "sdk",
					TimeFrame:             86400,
					Recurring:             false,
					JobCostPrediction:     "1000000000000000000",
					ConditionType:         "balance_check",
					UpperLimit:            1000.0,
					LowerLimit:            100.0,
					ValueSourceType:       "blockchain",
					ValueSourceUrl:        "https://api.example.com/balance",
					TargetChainID:         "11155111",
					TargetContractAddress: "0x1234567890123456789012345678901234567890",
					TargetFunction:        "executeOnBalance",
					ABI:                   `[{"name":"executeOnBalance","type":"function"}]`,
					ArgType:               1,
				},
			},
			mockSetup: func(
				mockUserRepo *datastoreMocks.MockGenericRepository[types.UserDataEntity],
				mockJobRepo *datastoreMocks.MockGenericRepository[types.JobDataEntity],
				mockTimeJobRepo *datastoreMocks.MockGenericRepository[types.TimeJobDataEntity],
				mockEventJobRepo *datastoreMocks.MockGenericRepository[types.EventJobDataEntity],
				mockConditionJobRepo *datastoreMocks.MockGenericRepository[types.ConditionJobDataEntity],
			) {
				mockUserRepo.On("GetByID", mock.Anything, "0x1234567890123456789012345678901234567890").Return(nil, nil).Once()
				mockUserRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Once()
				mockJobRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Once()
				mockConditionJobRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Once()
				mockJobRepo.On("Update", mock.Anything, mock.Anything).Return(nil).Maybe()
				mockUserRepo.On("Update", mock.Anything, mock.Anything).Return(nil).Once()
			},
			expectedCode: http.StatusOK,
			validateResult: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response types.CreateJobResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "job303", response.JobIDs[0])
			},
		},
		{
			name: "Success - Create Multiple Jobs with Chaining",
			request: []types.CreateJobDataRequest{
				{
					JobID:                 "job1",
					JobTitle:              "First Job",
					TaskDefinitionID:      1,
					CreatedChainID:        "11155111",
					UserAddress:           "0x1234567890123456789012345678901234567890",
					Timezone:              "UTC",
					JobType:               "frontend",
					TimeFrame:             86400,
					Recurring:             true,
					JobCostPrediction:     "1000000000000000000",
					ScheduleType:          "interval",
					TimeInterval:          3600,
					TargetChainID:         "11155111",
					TargetContractAddress: "0x1234567890123456789012345678901234567890",
					TargetFunction:        "func1",
					ABI:                   `[{"name":"func1","type":"function"}]`,
					ArgType:               1,
				},
				{
					JobID:                 "job2",
					JobTitle:              "Second Job",
					TaskDefinitionID:      2,
					CreatedChainID:        "11155111",
					UserAddress:           "0x1234567890123456789012345678901234567890",
					Timezone:              "UTC",
					JobType:               "frontend",
					TimeFrame:             86400,
					Recurring:             true,
					JobCostPrediction:     "1000000000000000000",
					ScheduleType:          "interval",
					TimeInterval:          7200,
					TargetChainID:         "11155111",
					TargetContractAddress: "0x1234567890123456789012345678901234567890",
					TargetFunction:        "func2",
					ABI:                   `[{"name":"func2","type":"function"}]`,
					ArgType:               1,
				},
			},
			mockSetup: func(
				mockUserRepo *datastoreMocks.MockGenericRepository[types.UserDataEntity],
				mockJobRepo *datastoreMocks.MockGenericRepository[types.JobDataEntity],
				mockTimeJobRepo *datastoreMocks.MockGenericRepository[types.TimeJobDataEntity],
				mockEventJobRepo *datastoreMocks.MockGenericRepository[types.EventJobDataEntity],
				mockConditionJobRepo *datastoreMocks.MockGenericRepository[types.ConditionJobDataEntity],
			) {
				mockUserRepo.On("GetByID", mock.Anything, "0x1234567890123456789012345678901234567890").Return(nil, nil).Once()
				mockUserRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Once()

				// Jobs are created in reverse order
				// First iteration: job2 (i=1, chainStatus=1, linkJobID="")
				mockJobRepo.On("Create", mock.Anything, mock.MatchedBy(func(j *types.JobDataEntity) bool {
					return j.JobID == "job2" && j.ChainStatus == 1 && j.LinkJobID == ""
				})).Return(nil).Once()
				mockTimeJobRepo.On("Create", mock.Anything, mock.MatchedBy(func(tj *types.TimeJobDataEntity) bool {
					return tj.JobID == "job2"
				})).Return(nil).Once()

				// Second iteration: job1 (i=0, chainStatus=0, linkJobID="job2")
				mockJobRepo.On("Create", mock.Anything, mock.MatchedBy(func(j *types.JobDataEntity) bool {
					return j.JobID == "job1" && j.ChainStatus == 0 && j.LinkJobID == "job2"
				})).Return(nil).Once()
				mockTimeJobRepo.On("Create", mock.Anything, mock.MatchedBy(func(tj *types.TimeJobDataEntity) bool {
					return tj.JobID == "job1"
				})).Return(nil).Once()

				mockUserRepo.On("Update", mock.Anything, mock.MatchedBy(func(u *types.UserDataEntity) bool {
					return len(u.JobIDs) == 2 && u.TotalJobs == 2
				})).Return(nil).Once()
			},
			expectedCode: http.StatusOK,
			validateResult: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response types.CreateJobResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Len(t, response.JobIDs, 2)
				assert.Equal(t, "job1", response.JobIDs[0])
				assert.Equal(t, "job2", response.JobIDs[1])
			},
		},
		{
			name:    "Error - Invalid Request Body (malformed JSON)",
			request: nil, // Will send invalid JSON
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
			name:    "Error - Empty Jobs Array",
			request: []types.CreateJobDataRequest{},
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
			name: "Error - User Repository Error on GetByID",
			request: []types.CreateJobDataRequest{
				{
					JobID:                 "job999",
					JobTitle:              "Test Job",
					TaskDefinitionID:      1,
					CreatedChainID:        "11155111",
					UserAddress:           "0x1234567890123456789012345678901234567890",
					Timezone:              "UTC",
					JobType:               "frontend",
					TimeFrame:             86400,
					Recurring:             true,
					JobCostPrediction:     "1000000000000000000",
					ScheduleType:          "interval",
					TimeInterval:          3600,
					TargetChainID:         "11155111",
					TargetContractAddress: "0x1234567890123456789012345678901234567890",
					TargetFunction:        "testFunc",
					ABI:                   `[]`,
					ArgType:               1,
				},
			},
			mockSetup: func(
				mockUserRepo *datastoreMocks.MockGenericRepository[types.UserDataEntity],
				mockJobRepo *datastoreMocks.MockGenericRepository[types.JobDataEntity],
				mockTimeJobRepo *datastoreMocks.MockGenericRepository[types.TimeJobDataEntity],
				mockEventJobRepo *datastoreMocks.MockGenericRepository[types.EventJobDataEntity],
				mockConditionJobRepo *datastoreMocks.MockGenericRepository[types.ConditionJobDataEntity],
			) {
				mockUserRepo.On("GetByID", mock.Anything, "0x1234567890123456789012345678901234567890").Return(nil, errors.New("database error")).Once()
			},
			expectedCode:  http.StatusInternalServerError,
			expectedError: "Internal server error",
		},
		{
			name: "Error - User Creation Failed",
			request: []types.CreateJobDataRequest{
				{
					JobID:                 "job888",
					JobTitle:              "Test Job",
					TaskDefinitionID:      1,
					CreatedChainID:        "11155111",
					UserAddress:           "0x1234567890123456789012345678901234567890",
					Timezone:              "UTC",
					JobType:               "frontend",
					TimeFrame:             86400,
					Recurring:             true,
					JobCostPrediction:     "1000000000000000000",
					ScheduleType:          "interval",
					TimeInterval:          3600,
					TargetChainID:         "11155111",
					TargetContractAddress: "0x1234567890123456789012345678901234567890",
					TargetFunction:        "testFunc",
					ABI:                   `[]`,
					ArgType:               1,
				},
			},
			mockSetup: func(
				mockUserRepo *datastoreMocks.MockGenericRepository[types.UserDataEntity],
				mockJobRepo *datastoreMocks.MockGenericRepository[types.JobDataEntity],
				mockTimeJobRepo *datastoreMocks.MockGenericRepository[types.TimeJobDataEntity],
				mockEventJobRepo *datastoreMocks.MockGenericRepository[types.EventJobDataEntity],
				mockConditionJobRepo *datastoreMocks.MockGenericRepository[types.ConditionJobDataEntity],
			) {
				mockUserRepo.On("GetByID", mock.Anything, "0x1234567890123456789012345678901234567890").Return(nil, nil).Once()
				mockUserRepo.On("Create", mock.Anything, mock.Anything).Return(errors.New("creation failed")).Once()
			},
			expectedCode:  http.StatusInternalServerError,
			expectedError: "Internal server error",
		},
		{
			name: "Error - Job Creation Failed",
			request: []types.CreateJobDataRequest{
				{
					JobID:                 "job777",
					JobTitle:              "Test Job",
					TaskDefinitionID:      1,
					CreatedChainID:        "11155111",
					UserAddress:           "0x1234567890123456789012345678901234567890",
					Timezone:              "UTC",
					JobType:               "frontend",
					TimeFrame:             86400,
					Recurring:             true,
					JobCostPrediction:     "1000000000000000000",
					ScheduleType:          "interval",
					TimeInterval:          3600,
					TargetChainID:         "11155111",
					TargetContractAddress: "0x1234567890123456789012345678901234567890",
					TargetFunction:        "testFunc",
					ABI:                   `[]`,
					ArgType:               1,
				},
			},
			mockSetup: func(
				mockUserRepo *datastoreMocks.MockGenericRepository[types.UserDataEntity],
				mockJobRepo *datastoreMocks.MockGenericRepository[types.JobDataEntity],
				mockTimeJobRepo *datastoreMocks.MockGenericRepository[types.TimeJobDataEntity],
				mockEventJobRepo *datastoreMocks.MockGenericRepository[types.EventJobDataEntity],
				mockConditionJobRepo *datastoreMocks.MockGenericRepository[types.ConditionJobDataEntity],
			) {
				mockUserRepo.On("GetByID", mock.Anything, "0x1234567890123456789012345678901234567890").Return(nil, nil).Once()
				mockUserRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Once()
				mockJobRepo.On("Create", mock.Anything, mock.Anything).Return(errors.New("job creation failed")).Once()
			},
			expectedCode:  http.StatusInternalServerError,
			expectedError: "Internal server error",
		},
		{
			name: "Error - Time Job Creation Failed",
			request: []types.CreateJobDataRequest{
				{
					JobID:                 "job666",
					JobTitle:              "Test Job",
					TaskDefinitionID:      1,
					CreatedChainID:        "11155111",
					UserAddress:           "0x1234567890123456789012345678901234567890",
					Timezone:              "UTC",
					JobType:               "frontend",
					TimeFrame:             86400,
					Recurring:             true,
					JobCostPrediction:     "1000000000000000000",
					ScheduleType:          "interval",
					TimeInterval:          3600,
					TargetChainID:         "11155111",
					TargetContractAddress: "0x1234567890123456789012345678901234567890",
					TargetFunction:        "testFunc",
					ABI:                   `[]`,
					ArgType:               1,
				},
			},
			mockSetup: func(
				mockUserRepo *datastoreMocks.MockGenericRepository[types.UserDataEntity],
				mockJobRepo *datastoreMocks.MockGenericRepository[types.JobDataEntity],
				mockTimeJobRepo *datastoreMocks.MockGenericRepository[types.TimeJobDataEntity],
				mockEventJobRepo *datastoreMocks.MockGenericRepository[types.EventJobDataEntity],
				mockConditionJobRepo *datastoreMocks.MockGenericRepository[types.ConditionJobDataEntity],
			) {
				mockUserRepo.On("GetByID", mock.Anything, "0x1234567890123456789012345678901234567890").Return(nil, nil).Once()
				mockUserRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Once()
				mockJobRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Once()
				mockTimeJobRepo.On("Create", mock.Anything, mock.Anything).Return(errors.New("time job creation failed")).Once()
			},
			expectedCode:  http.StatusInternalServerError,
			expectedError: "Internal server error",
		},
		{
			name: "Error - Event Job Creation Failed",
			request: []types.CreateJobDataRequest{
				{
					JobID:                  "job555",
					JobTitle:               "Test Job",
					TaskDefinitionID:       3,
					CreatedChainID:         "11155111",
					UserAddress:            "0x1234567890123456789012345678901234567890",
					Timezone:               "UTC",
					JobType:                "frontend",
					TimeFrame:              86400,
					Recurring:              true,
					JobCostPrediction:      "1000000000000000000",
					TriggerChainID:         "11155111",
					TriggerContractAddress: "0x1234567890123456789012345678901234567890",
					TriggerEvent:           "Transfer",
					TargetChainID:          "11155111",
					TargetContractAddress:  "0x0987654321098765432109876543210987654321",
					TargetFunction:         "handleTransfer",
					ABI:                    `[]`,
					ArgType:                1,
				},
			},
			mockSetup: func(
				mockUserRepo *datastoreMocks.MockGenericRepository[types.UserDataEntity],
				mockJobRepo *datastoreMocks.MockGenericRepository[types.JobDataEntity],
				mockTimeJobRepo *datastoreMocks.MockGenericRepository[types.TimeJobDataEntity],
				mockEventJobRepo *datastoreMocks.MockGenericRepository[types.EventJobDataEntity],
				mockConditionJobRepo *datastoreMocks.MockGenericRepository[types.ConditionJobDataEntity],
			) {
				mockUserRepo.On("GetByID", mock.Anything, "0x1234567890123456789012345678901234567890").Return(nil, nil).Once()
				mockUserRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Once()
				mockJobRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Once()
				mockEventJobRepo.On("Create", mock.Anything, mock.Anything).Return(errors.New("event job creation failed")).Once()
			},
			expectedCode:  http.StatusInternalServerError,
			expectedError: "Internal server error",
		},
		{
			name: "Error - Condition Job Creation Failed",
			request: []types.CreateJobDataRequest{
				{
					JobID:                 "job444",
					JobTitle:              "Test Job",
					TaskDefinitionID:      5,
					CreatedChainID:        "11155111",
					UserAddress:           "0x1234567890123456789012345678901234567890",
					Timezone:              "UTC",
					JobType:               "frontend",
					TimeFrame:             86400,
					Recurring:             true,
					JobCostPrediction:     "1000000000000000000",
					ConditionType:         "price",
					UpperLimit:            100.0,
					LowerLimit:            50.0,
					ValueSourceType:       "api",
					ValueSourceUrl:        "https://api.example.com",
					TargetChainID:         "11155111",
					TargetContractAddress: "0x1234567890123456789012345678901234567890",
					TargetFunction:        "execute",
					ABI:                   `[]`,
					ArgType:               1,
				},
			},
			mockSetup: func(
				mockUserRepo *datastoreMocks.MockGenericRepository[types.UserDataEntity],
				mockJobRepo *datastoreMocks.MockGenericRepository[types.JobDataEntity],
				mockTimeJobRepo *datastoreMocks.MockGenericRepository[types.TimeJobDataEntity],
				mockEventJobRepo *datastoreMocks.MockGenericRepository[types.EventJobDataEntity],
				mockConditionJobRepo *datastoreMocks.MockGenericRepository[types.ConditionJobDataEntity],
			) {
				mockUserRepo.On("GetByID", mock.Anything, "0x1234567890123456789012345678901234567890").Return(nil, nil).Once()
				mockUserRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Once()
				mockJobRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Once()
				mockConditionJobRepo.On("Create", mock.Anything, mock.Anything).Return(errors.New("condition job creation failed")).Once()
			},
			expectedCode:  http.StatusInternalServerError,
			expectedError: "Internal server error",
		},
		{
			name: "Error - Invalid Task Definition ID",
			request: []types.CreateJobDataRequest{
				{
					JobID:                 "job333",
					JobTitle:              "Test Job",
					TaskDefinitionID:      99, // Invalid
					CreatedChainID:        "11155111",
					UserAddress:           "0x1234567890123456789012345678901234567890",
					Timezone:              "UTC",
					JobType:               "frontend",
					TimeFrame:             86400,
					Recurring:             true,
					JobCostPrediction:     "1000000000000000000",
					TargetChainID:         "11155111",
					TargetContractAddress: "0x1234567890123456789012345678901234567890",
					TargetFunction:        "test",
					ABI:                   `[]`,
					ArgType:               1,
				},
			},
			mockSetup: func(
				mockUserRepo *datastoreMocks.MockGenericRepository[types.UserDataEntity],
				mockJobRepo *datastoreMocks.MockGenericRepository[types.JobDataEntity],
				mockTimeJobRepo *datastoreMocks.MockGenericRepository[types.TimeJobDataEntity],
				mockEventJobRepo *datastoreMocks.MockGenericRepository[types.EventJobDataEntity],
				mockConditionJobRepo *datastoreMocks.MockGenericRepository[types.ConditionJobDataEntity],
			) {
				mockUserRepo.On("GetByID", mock.Anything, "0x1234567890123456789012345678901234567890").Return(nil, nil).Once()
				mockUserRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Once()
				mockJobRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Once()
			},
			expectedCode:  http.StatusBadRequest,
			expectedError: "Invalid task definition ID",
		},
		{
			name: "Error - User Update Failed",
			request: []types.CreateJobDataRequest{
				{
					JobID:                 "job222",
					JobTitle:              "Test Job",
					TaskDefinitionID:      1,
					CreatedChainID:        "11155111",
					UserAddress:           "0x1234567890123456789012345678901234567890",
					Timezone:              "UTC",
					JobType:               "frontend",
					TimeFrame:             86400,
					Recurring:             true,
					JobCostPrediction:     "1000000000000000000",
					ScheduleType:          "interval",
					TimeInterval:          3600,
					TargetChainID:         "11155111",
					TargetContractAddress: "0x1234567890123456789012345678901234567890",
					TargetFunction:        "test",
					ABI:                   `[]`,
					ArgType:               1,
				},
			},
			mockSetup: func(
				mockUserRepo *datastoreMocks.MockGenericRepository[types.UserDataEntity],
				mockJobRepo *datastoreMocks.MockGenericRepository[types.JobDataEntity],
				mockTimeJobRepo *datastoreMocks.MockGenericRepository[types.TimeJobDataEntity],
				mockEventJobRepo *datastoreMocks.MockGenericRepository[types.EventJobDataEntity],
				mockConditionJobRepo *datastoreMocks.MockGenericRepository[types.ConditionJobDataEntity],
			) {
				mockUserRepo.On("GetByID", mock.Anything, "0x1234567890123456789012345678901234567890").Return(nil, nil).Once()
				mockUserRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Once()
				mockJobRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Once()
				mockTimeJobRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Once()
				mockUserRepo.On("Update", mock.Anything, mock.Anything).Return(errors.New("user update failed")).Once()
			},
			expectedCode:  http.StatusInternalServerError,
			expectedError: "Internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockJobRepo, mockTimeJobRepo, mockEventJobRepo, mockConditionJobRepo, mockUserRepo, mockLogger := setupJobCreateTestHandler()
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

			// Create request body
			var reqBody []byte
			if tt.request == nil && tt.name == "Error - Invalid Request Body (malformed JSON)" {
				reqBody = []byte("{invalid json")
			} else {
				reqBody, _ = json.Marshal(tt.request)
			}
			c.Request = httptest.NewRequest("POST", "/jobs", bytes.NewBuffer(reqBody))
			c.Request.Header.Set("Content-Type", "application/json")

			// Setup mocks
			tt.mockSetup(mockUserRepo, mockJobRepo, mockTimeJobRepo, mockEventJobRepo, mockConditionJobRepo)

			// Execute
			handler.CreateJobData(c)

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
