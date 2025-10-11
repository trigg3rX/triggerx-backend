package handlers

import (
	"bytes"
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

func setupJobUpdateTestHandler() (
	*Handler,
	*datastoreMocks.MockGenericRepository[types.JobDataEntity],
	*datastoreMocks.MockGenericRepository[types.TimeJobDataEntity],
	*datastoreMocks.MockGenericRepository[types.EventJobDataEntity],
	*datastoreMocks.MockGenericRepository[types.ConditionJobDataEntity],
	*logging.MockLogger,
) {
	mockJobRepo := new(datastoreMocks.MockGenericRepository[types.JobDataEntity])
	mockTimeJobRepo := new(datastoreMocks.MockGenericRepository[types.TimeJobDataEntity])
	mockEventJobRepo := new(datastoreMocks.MockGenericRepository[types.EventJobDataEntity])
	mockConditionJobRepo := new(datastoreMocks.MockGenericRepository[types.ConditionJobDataEntity])
	mockLogger := new(logging.MockLogger)
	mockLogger.SetupDefaultExpectations()

	handler := &Handler{
		logger:                 mockLogger,
		jobRepository:          mockJobRepo,
		timeJobRepository:      mockTimeJobRepo,
		eventJobRepository:     mockEventJobRepo,
		conditionJobRepository: mockConditionJobRepo,
	}

	return handler, mockJobRepo, mockTimeJobRepo, mockEventJobRepo, mockConditionJobRepo, mockLogger
}

func TestDeleteJobData(t *testing.T) {
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
			name:  "Success - Delete Time Job (TaskDefinitionID 1)",
			jobID: "12345",
			mockSetup: func(
				mockJobRepo *datastoreMocks.MockGenericRepository[types.JobDataEntity],
				mockTimeJobRepo *datastoreMocks.MockGenericRepository[types.TimeJobDataEntity],
				mockEventJobRepo *datastoreMocks.MockGenericRepository[types.EventJobDataEntity],
				mockConditionJobRepo *datastoreMocks.MockGenericRepository[types.ConditionJobDataEntity],
			) {
				job := &types.JobDataEntity{
					JobID:            "12345",
					TaskDefinitionID: 1,
					Status:           string(types.JobStatusRunning),
					CreatedAt:        now,
					UpdatedAt:        now,
				}
				mockJobRepo.On("GetByID", mock.Anything, "12345").Return(job, nil).Once()
				mockJobRepo.On("Update", mock.Anything, mock.MatchedBy(func(j *types.JobDataEntity) bool {
					return j.JobID == "12345" && j.Status == string(types.JobStatusDeleted)
				})).Return(nil).Once()

				timeJob := &types.TimeJobDataEntity{
					JobID:       "12345",
					IsCompleted: false,
				}
				mockTimeJobRepo.On("GetByNonID", mock.Anything, "job_id", "12345").Return(timeJob, nil).Once()
				mockTimeJobRepo.On("Update", mock.Anything, mock.MatchedBy(func(tj *types.TimeJobDataEntity) bool {
					return tj.JobID == "12345" && tj.IsCompleted == true
				})).Return(nil).Once()
			},
			expectedCode: http.StatusOK,
			validateResult: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "Job deleted successfully", response["message"])
			},
		},
		{
			name:  "Success - Delete Time Job (TaskDefinitionID 2)",
			jobID: "12346",
			mockSetup: func(
				mockJobRepo *datastoreMocks.MockGenericRepository[types.JobDataEntity],
				mockTimeJobRepo *datastoreMocks.MockGenericRepository[types.TimeJobDataEntity],
				mockEventJobRepo *datastoreMocks.MockGenericRepository[types.EventJobDataEntity],
				mockConditionJobRepo *datastoreMocks.MockGenericRepository[types.ConditionJobDataEntity],
			) {
				job := &types.JobDataEntity{
					JobID:            "12346",
					TaskDefinitionID: 2,
					Status:           string(types.JobStatusRunning),
					CreatedAt:        now,
					UpdatedAt:        now,
				}
				mockJobRepo.On("GetByID", mock.Anything, "12346").Return(job, nil).Once()
				mockJobRepo.On("Update", mock.Anything, mock.MatchedBy(func(j *types.JobDataEntity) bool {
					return j.Status == string(types.JobStatusDeleted)
				})).Return(nil).Once()

				timeJob := &types.TimeJobDataEntity{
					JobID:       "12346",
					IsCompleted: false,
				}
				mockTimeJobRepo.On("GetByNonID", mock.Anything, "job_id", "12346").Return(timeJob, nil).Once()
				mockTimeJobRepo.On("Update", mock.Anything, mock.MatchedBy(func(tj *types.TimeJobDataEntity) bool {
					return tj.IsCompleted == true
				})).Return(nil).Once()
			},
			expectedCode: http.StatusOK,
			validateResult: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "Job deleted successfully", response["message"])
			},
		},
		{
			name:  "Success - Delete Event Job (TaskDefinitionID 3) - Note: notifyPauseToConditionScheduler will fail in test",
			jobID: "12347",
			mockSetup: func(
				mockJobRepo *datastoreMocks.MockGenericRepository[types.JobDataEntity],
				mockTimeJobRepo *datastoreMocks.MockGenericRepository[types.TimeJobDataEntity],
				mockEventJobRepo *datastoreMocks.MockGenericRepository[types.EventJobDataEntity],
				mockConditionJobRepo *datastoreMocks.MockGenericRepository[types.ConditionJobDataEntity],
			) {
				job := &types.JobDataEntity{
					JobID:            "12347",
					TaskDefinitionID: 3,
					Status:           string(types.JobStatusRunning),
					CreatedAt:        now,
					UpdatedAt:        now,
				}
				mockJobRepo.On("GetByID", mock.Anything, "12347").Return(job, nil).Once()
				mockJobRepo.On("Update", mock.Anything, mock.MatchedBy(func(j *types.JobDataEntity) bool {
					return j.Status == string(types.JobStatusDeleted)
				})).Return(nil).Once()

				eventJob := &types.EventJobDataEntity{
					JobID:       "12347",
					IsCompleted: false,
				}
				mockEventJobRepo.On("GetByNonID", mock.Anything, "job_id", "12347").Return(eventJob, nil).Once()
				mockEventJobRepo.On("Update", mock.Anything, mock.MatchedBy(func(ej *types.EventJobDataEntity) bool {
					return ej.IsCompleted == true
				})).Return(nil).Once()
			},
			expectedCode: http.StatusInternalServerError,
			validateResult: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				// In test environment, notifyPauseToConditionScheduler will fail because there's no actual server
				assert.Contains(t, response["error"], "Error sending pause to event scheduler")
			},
		},
		{
			name:  "Success - Delete Event Job (TaskDefinitionID 4) - Note: notifyPauseToConditionScheduler will fail in test",
			jobID: "12348",
			mockSetup: func(
				mockJobRepo *datastoreMocks.MockGenericRepository[types.JobDataEntity],
				mockTimeJobRepo *datastoreMocks.MockGenericRepository[types.TimeJobDataEntity],
				mockEventJobRepo *datastoreMocks.MockGenericRepository[types.EventJobDataEntity],
				mockConditionJobRepo *datastoreMocks.MockGenericRepository[types.ConditionJobDataEntity],
			) {
				job := &types.JobDataEntity{
					JobID:            "12348",
					TaskDefinitionID: 4,
					Status:           string(types.JobStatusRunning),
					CreatedAt:        now,
					UpdatedAt:        now,
				}
				mockJobRepo.On("GetByID", mock.Anything, "12348").Return(job, nil).Once()
				mockJobRepo.On("Update", mock.Anything, mock.MatchedBy(func(j *types.JobDataEntity) bool {
					return j.Status == string(types.JobStatusDeleted)
				})).Return(nil).Once()

				eventJob := &types.EventJobDataEntity{
					JobID:       "12348",
					IsCompleted: false,
				}
				mockEventJobRepo.On("GetByNonID", mock.Anything, "job_id", "12348").Return(eventJob, nil).Once()
				mockEventJobRepo.On("Update", mock.Anything, mock.MatchedBy(func(ej *types.EventJobDataEntity) bool {
					return ej.IsCompleted == true
				})).Return(nil).Once()
			},
			expectedCode: http.StatusInternalServerError,
			validateResult: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response["error"], "Error sending pause to event scheduler")
			},
		},
		{
			name:  "Success - Delete Condition Job (TaskDefinitionID 5) - Note: notifyPauseToConditionScheduler will fail in test",
			jobID: "12349",
			mockSetup: func(
				mockJobRepo *datastoreMocks.MockGenericRepository[types.JobDataEntity],
				mockTimeJobRepo *datastoreMocks.MockGenericRepository[types.TimeJobDataEntity],
				mockEventJobRepo *datastoreMocks.MockGenericRepository[types.EventJobDataEntity],
				mockConditionJobRepo *datastoreMocks.MockGenericRepository[types.ConditionJobDataEntity],
			) {
				job := &types.JobDataEntity{
					JobID:            "12349",
					TaskDefinitionID: 5,
					Status:           string(types.JobStatusRunning),
					CreatedAt:        now,
					UpdatedAt:        now,
				}
				mockJobRepo.On("GetByID", mock.Anything, "12349").Return(job, nil).Once()
				mockJobRepo.On("Update", mock.Anything, mock.MatchedBy(func(j *types.JobDataEntity) bool {
					return j.Status == string(types.JobStatusDeleted)
				})).Return(nil).Once()

				conditionJob := &types.ConditionJobDataEntity{
					JobID:       "12349",
					IsCompleted: false,
				}
				mockConditionJobRepo.On("GetByNonID", mock.Anything, "job_id", "12349").Return(conditionJob, nil).Once()
				mockConditionJobRepo.On("Update", mock.Anything, mock.MatchedBy(func(cj *types.ConditionJobDataEntity) bool {
					return cj.IsCompleted == true
				})).Return(nil).Once()
			},
			expectedCode: http.StatusInternalServerError,
			validateResult: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response["error"], "Error sending pause to condition scheduler")
			},
		},
		{
			name:  "Success - Delete Condition Job (TaskDefinitionID 6) - Note: notifyPauseToConditionScheduler will fail in test",
			jobID: "12350",
			mockSetup: func(
				mockJobRepo *datastoreMocks.MockGenericRepository[types.JobDataEntity],
				mockTimeJobRepo *datastoreMocks.MockGenericRepository[types.TimeJobDataEntity],
				mockEventJobRepo *datastoreMocks.MockGenericRepository[types.EventJobDataEntity],
				mockConditionJobRepo *datastoreMocks.MockGenericRepository[types.ConditionJobDataEntity],
			) {
				job := &types.JobDataEntity{
					JobID:            "12350",
					TaskDefinitionID: 6,
					Status:           string(types.JobStatusRunning),
					CreatedAt:        now,
					UpdatedAt:        now,
				}
				mockJobRepo.On("GetByID", mock.Anything, "12350").Return(job, nil).Once()
				mockJobRepo.On("Update", mock.Anything, mock.MatchedBy(func(j *types.JobDataEntity) bool {
					return j.Status == string(types.JobStatusDeleted)
				})).Return(nil).Once()

				conditionJob := &types.ConditionJobDataEntity{
					JobID:       "12350",
					IsCompleted: false,
				}
				mockConditionJobRepo.On("GetByNonID", mock.Anything, "job_id", "12350").Return(conditionJob, nil).Once()
				mockConditionJobRepo.On("Update", mock.Anything, mock.MatchedBy(func(cj *types.ConditionJobDataEntity) bool {
					return cj.IsCompleted == true
				})).Return(nil).Once()
			},
			expectedCode: http.StatusInternalServerError,
			validateResult: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response["error"], "Error sending pause to condition scheduler")
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
			expectedError: pkgErrors.ErrDBRecordNotFound,
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
			expectedError: pkgErrors.ErrDBRecordNotFound,
		},
		{
			name:  "Error - Job Update Failed",
			jobID: "77777",
			mockSetup: func(
				mockJobRepo *datastoreMocks.MockGenericRepository[types.JobDataEntity],
				mockTimeJobRepo *datastoreMocks.MockGenericRepository[types.TimeJobDataEntity],
				mockEventJobRepo *datastoreMocks.MockGenericRepository[types.EventJobDataEntity],
				mockConditionJobRepo *datastoreMocks.MockGenericRepository[types.ConditionJobDataEntity],
			) {
				job := &types.JobDataEntity{
					JobID:            "77777",
					TaskDefinitionID: 1,
					Status:           string(types.JobStatusRunning),
					CreatedAt:        now,
					UpdatedAt:        now,
				}
				mockJobRepo.On("GetByID", mock.Anything, "77777").Return(job, nil).Once()
				mockJobRepo.On("Update", mock.Anything, mock.Anything).Return(errors.New("update failed")).Once()
			},
			expectedCode: http.StatusInternalServerError,
			validateResult: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response["error"], "Error updating job status")
			},
		},
		{
			name:  "Error - Time Job Update Failed",
			jobID: "66666",
			mockSetup: func(
				mockJobRepo *datastoreMocks.MockGenericRepository[types.JobDataEntity],
				mockTimeJobRepo *datastoreMocks.MockGenericRepository[types.TimeJobDataEntity],
				mockEventJobRepo *datastoreMocks.MockGenericRepository[types.EventJobDataEntity],
				mockConditionJobRepo *datastoreMocks.MockGenericRepository[types.ConditionJobDataEntity],
			) {
				job := &types.JobDataEntity{
					JobID:            "66666",
					TaskDefinitionID: 1,
					Status:           string(types.JobStatusRunning),
					CreatedAt:        now,
					UpdatedAt:        now,
				}
				mockJobRepo.On("GetByID", mock.Anything, "66666").Return(job, nil).Once()
				mockJobRepo.On("Update", mock.Anything, mock.Anything).Return(nil).Once()

				timeJob := &types.TimeJobDataEntity{
					JobID:       "66666",
					IsCompleted: false,
				}
				mockTimeJobRepo.On("GetByNonID", mock.Anything, "job_id", "66666").Return(timeJob, nil).Once()
				mockTimeJobRepo.On("Update", mock.Anything, mock.Anything).Return(errors.New("update failed")).Once()
			},
			expectedCode: http.StatusInternalServerError,
			validateResult: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response["error"], "Error updating time job status")
			},
		},
		{
			name:  "Error - Event Job Update Failed",
			jobID: "55555",
			mockSetup: func(
				mockJobRepo *datastoreMocks.MockGenericRepository[types.JobDataEntity],
				mockTimeJobRepo *datastoreMocks.MockGenericRepository[types.TimeJobDataEntity],
				mockEventJobRepo *datastoreMocks.MockGenericRepository[types.EventJobDataEntity],
				mockConditionJobRepo *datastoreMocks.MockGenericRepository[types.ConditionJobDataEntity],
			) {
				job := &types.JobDataEntity{
					JobID:            "55555",
					TaskDefinitionID: 3,
					Status:           string(types.JobStatusRunning),
					CreatedAt:        now,
					UpdatedAt:        now,
				}
				mockJobRepo.On("GetByID", mock.Anything, "55555").Return(job, nil).Once()
				mockJobRepo.On("Update", mock.Anything, mock.Anything).Return(nil).Once()

				eventJob := &types.EventJobDataEntity{
					JobID:       "55555",
					IsCompleted: false,
				}
				mockEventJobRepo.On("GetByNonID", mock.Anything, "job_id", "55555").Return(eventJob, nil).Once()
				mockEventJobRepo.On("Update", mock.Anything, mock.Anything).Return(errors.New("update failed")).Once()
			},
			expectedCode: http.StatusInternalServerError,
			validateResult: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response["error"], "Error updating event job status")
			},
		},
		{
			name:  "Error - Condition Job Update Failed",
			jobID: "44444",
			mockSetup: func(
				mockJobRepo *datastoreMocks.MockGenericRepository[types.JobDataEntity],
				mockTimeJobRepo *datastoreMocks.MockGenericRepository[types.TimeJobDataEntity],
				mockEventJobRepo *datastoreMocks.MockGenericRepository[types.EventJobDataEntity],
				mockConditionJobRepo *datastoreMocks.MockGenericRepository[types.ConditionJobDataEntity],
			) {
				job := &types.JobDataEntity{
					JobID:            "44444",
					TaskDefinitionID: 5,
					Status:           string(types.JobStatusRunning),
					CreatedAt:        now,
					UpdatedAt:        now,
				}
				mockJobRepo.On("GetByID", mock.Anything, "44444").Return(job, nil).Once()
				mockJobRepo.On("Update", mock.Anything, mock.Anything).Return(nil).Once()

				conditionJob := &types.ConditionJobDataEntity{
					JobID:       "44444",
					IsCompleted: false,
				}
				mockConditionJobRepo.On("GetByNonID", mock.Anything, "job_id", "44444").Return(conditionJob, nil).Once()
				mockConditionJobRepo.On("Update", mock.Anything, mock.Anything).Return(errors.New("update failed")).Once()
			},
			expectedCode: http.StatusInternalServerError,
			validateResult: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response["error"], "Error updating condition job status")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockJobRepo, mockTimeJobRepo, mockEventJobRepo, mockConditionJobRepo, mockLogger := setupJobUpdateTestHandler()
			defer mockLogger.AssertExpectations(t)
			defer mockJobRepo.AssertExpectations(t)
			defer mockTimeJobRepo.AssertExpectations(t)
			defer mockEventJobRepo.AssertExpectations(t)
			defer mockConditionJobRepo.AssertExpectations(t)

			// Setup
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("PUT", "/jobs/"+tt.jobID+"/delete", nil)
			c.Params = gin.Params{gin.Param{Key: "id", Value: tt.jobID}}

			// Setup mocks
			tt.mockSetup(mockJobRepo, mockTimeJobRepo, mockEventJobRepo, mockConditionJobRepo)

			// Execute
			handler.DeleteJobData(c)

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

func TestUpdateJobDataFromUser(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name           string
		request        types.UpdateJobDataFromUserRequest
		mockSetup      func(*datastoreMocks.MockGenericRepository[types.JobDataEntity], *datastoreMocks.MockGenericRepository[types.TimeJobDataEntity], *datastoreMocks.MockGenericRepository[types.EventJobDataEntity], *datastoreMocks.MockGenericRepository[types.ConditionJobDataEntity])
		expectedCode   int
		expectedError  string
		validateResult func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "Success - Update Time Job (TaskDefinitionID 1)",
			request: types.UpdateJobDataFromUserRequest{
				JobID:             "12345",
				JobTitle:          "Updated Time Job",
				Recurring:         true,
				TimeFrame:         86400, // 24 hours
				JobCostPrediction: "2000000000000000000",
				TimeInterval:      7200, // 2 hours
			},
			mockSetup: func(
				mockJobRepo *datastoreMocks.MockGenericRepository[types.JobDataEntity],
				mockTimeJobRepo *datastoreMocks.MockGenericRepository[types.TimeJobDataEntity],
				mockEventJobRepo *datastoreMocks.MockGenericRepository[types.EventJobDataEntity],
				mockConditionJobRepo *datastoreMocks.MockGenericRepository[types.ConditionJobDataEntity],
			) {
				job := &types.JobDataEntity{
					JobID:             "12345",
					TaskDefinitionID:  1,
					JobTitle:          "Old Time Job",
					Recurring:         true,
					TimeFrame:         43200,
					JobCostPrediction: "1000000000000000000",
					Status:            string(types.JobStatusRunning),
					CreatedAt:         now,
					UpdatedAt:         now,
				}
				mockJobRepo.On("GetByID", mock.Anything, "12345").Return(job, nil).Once()
				mockJobRepo.On("Update", mock.Anything, mock.MatchedBy(func(j *types.JobDataEntity) bool {
					return j.JobID == "12345" &&
						j.JobTitle == "Updated Time Job" &&
						j.TimeFrame == int64(86400) &&
						j.JobCostPrediction == "2000000000000000000"
				})).Return(nil).Once()

				mockTimeJobRepo.On("Update", mock.Anything, mock.MatchedBy(func(tj *types.TimeJobDataEntity) bool {
					return tj.JobID == "12345" && tj.TimeInterval == int64(7200)
				})).Return(nil).Once()
			},
			expectedCode: http.StatusOK,
			validateResult: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "Job updated successfully", response["message"])
				assert.Equal(t, "12345", response["job_id"])
			},
		},
		{
			name: "Success - Update Time Job (TaskDefinitionID 2)",
			request: types.UpdateJobDataFromUserRequest{
				JobID:             "12346",
				JobTitle:          "Updated Time Job 2",
				Recurring:         false,
				TimeFrame:         43200,
				JobCostPrediction: "1500000000000000000",
				TimeInterval:      3600,
			},
			mockSetup: func(
				mockJobRepo *datastoreMocks.MockGenericRepository[types.JobDataEntity],
				mockTimeJobRepo *datastoreMocks.MockGenericRepository[types.TimeJobDataEntity],
				mockEventJobRepo *datastoreMocks.MockGenericRepository[types.EventJobDataEntity],
				mockConditionJobRepo *datastoreMocks.MockGenericRepository[types.ConditionJobDataEntity],
			) {
				job := &types.JobDataEntity{
					JobID:            "12346",
					TaskDefinitionID: 2,
					Recurring:        false,
					CreatedAt:        now,
					UpdatedAt:        now,
				}
				mockJobRepo.On("GetByID", mock.Anything, "12346").Return(job, nil).Once()
				mockJobRepo.On("Update", mock.Anything, mock.Anything).Return(nil).Once()
				mockTimeJobRepo.On("Update", mock.Anything, mock.Anything).Return(nil).Once()
			},
			expectedCode: http.StatusOK,
			validateResult: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "Job updated successfully", response["message"])
			},
		},
		{
			name: "Success - Update Event Job (TaskDefinitionID 3) without recurring change",
			request: types.UpdateJobDataFromUserRequest{
				JobID:             "12347",
				JobTitle:          "Updated Event Job",
				Recurring:         true,
				TimeFrame:         86400,
				JobCostPrediction: "2000000000000000000",
				TimeInterval:      0, // Not used for event jobs
			},
			mockSetup: func(
				mockJobRepo *datastoreMocks.MockGenericRepository[types.JobDataEntity],
				mockTimeJobRepo *datastoreMocks.MockGenericRepository[types.TimeJobDataEntity],
				mockEventJobRepo *datastoreMocks.MockGenericRepository[types.EventJobDataEntity],
				mockConditionJobRepo *datastoreMocks.MockGenericRepository[types.ConditionJobDataEntity],
			) {
				job := &types.JobDataEntity{
					JobID:            "12347",
					TaskDefinitionID: 3,
					Recurring:        true, // Same as request
					CreatedAt:        now,
					UpdatedAt:        now,
				}
				mockJobRepo.On("GetByID", mock.Anything, "12347").Return(job, nil).Once()
				mockJobRepo.On("Update", mock.Anything, mock.Anything).Return(nil).Once()
				mockEventJobRepo.On("Update", mock.Anything, mock.Anything).Return(nil).Once()
			},
			expectedCode: http.StatusOK,
			validateResult: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "Job updated successfully", response["message"])
			},
		},
		{
			name: "Success - Update Event Job (TaskDefinitionID 4) with recurring change - Note: notifyUpdateToConditionScheduler will fail in test",
			request: types.UpdateJobDataFromUserRequest{
				JobID:             "12348",
				JobTitle:          "Updated Event Job 4",
				Recurring:         false, // Changed from true
				TimeFrame:         86400,
				JobCostPrediction: "2000000000000000000",
				TimeInterval:      0,
			},
			mockSetup: func(
				mockJobRepo *datastoreMocks.MockGenericRepository[types.JobDataEntity],
				mockTimeJobRepo *datastoreMocks.MockGenericRepository[types.TimeJobDataEntity],
				mockEventJobRepo *datastoreMocks.MockGenericRepository[types.EventJobDataEntity],
				mockConditionJobRepo *datastoreMocks.MockGenericRepository[types.ConditionJobDataEntity],
			) {
				job := &types.JobDataEntity{
					JobID:            "12348",
					TaskDefinitionID: 4,
					Recurring:        true, // Different from request
					CreatedAt:        now,
					UpdatedAt:        now,
				}
				mockJobRepo.On("GetByID", mock.Anything, "12348").Return(job, nil).Once()
				mockJobRepo.On("Update", mock.Anything, mock.Anything).Return(nil).Once()
				mockEventJobRepo.On("Update", mock.Anything, mock.Anything).Return(nil).Once()
			},
			expectedCode: http.StatusOK,
			validateResult: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "Job updated successfully", response["message"])
			},
		},
		{
			name: "Success - Update Condition Job (TaskDefinitionID 5) without recurring change",
			request: types.UpdateJobDataFromUserRequest{
				JobID:             "12349",
				JobTitle:          "Updated Condition Job",
				Recurring:         true,
				TimeFrame:         86400,
				JobCostPrediction: "2000000000000000000",
				TimeInterval:      0,
			},
			mockSetup: func(
				mockJobRepo *datastoreMocks.MockGenericRepository[types.JobDataEntity],
				mockTimeJobRepo *datastoreMocks.MockGenericRepository[types.TimeJobDataEntity],
				mockEventJobRepo *datastoreMocks.MockGenericRepository[types.EventJobDataEntity],
				mockConditionJobRepo *datastoreMocks.MockGenericRepository[types.ConditionJobDataEntity],
			) {
				job := &types.JobDataEntity{
					JobID:            "12349",
					TaskDefinitionID: 5,
					Recurring:        true,
					CreatedAt:        now,
					UpdatedAt:        now,
				}
				mockJobRepo.On("GetByID", mock.Anything, "12349").Return(job, nil).Once()
				mockJobRepo.On("Update", mock.Anything, mock.Anything).Return(nil).Once()
				mockConditionJobRepo.On("Update", mock.Anything, mock.Anything).Return(nil).Once()
			},
			expectedCode: http.StatusOK,
			validateResult: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "Job updated successfully", response["message"])
			},
		},
		{
			name: "Success - Update Condition Job (TaskDefinitionID 6) with recurring change - Note: notifyUpdateToConditionScheduler will fail in test",
			request: types.UpdateJobDataFromUserRequest{
				JobID:             "12350",
				JobTitle:          "Updated Condition Job 6",
				Recurring:         false,
				TimeFrame:         86400,
				JobCostPrediction: "2000000000000000000",
				TimeInterval:      0,
			},
			mockSetup: func(
				mockJobRepo *datastoreMocks.MockGenericRepository[types.JobDataEntity],
				mockTimeJobRepo *datastoreMocks.MockGenericRepository[types.TimeJobDataEntity],
				mockEventJobRepo *datastoreMocks.MockGenericRepository[types.EventJobDataEntity],
				mockConditionJobRepo *datastoreMocks.MockGenericRepository[types.ConditionJobDataEntity],
			) {
				job := &types.JobDataEntity{
					JobID:            "12350",
					TaskDefinitionID: 6,
					Recurring:        true,
					CreatedAt:        now,
					UpdatedAt:        now,
				}
				mockJobRepo.On("GetByID", mock.Anything, "12350").Return(job, nil).Once()
				mockJobRepo.On("Update", mock.Anything, mock.Anything).Return(nil).Once()
				mockConditionJobRepo.On("Update", mock.Anything, mock.Anything).Return(nil).Once()
			},
			expectedCode: http.StatusOK,
			validateResult: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "Job updated successfully", response["message"])
			},
		},
		{
			name:    "Error - Invalid Request Body",
			request: types.UpdateJobDataFromUserRequest{
				// Empty request
			},
			mockSetup: func(
				mockJobRepo *datastoreMocks.MockGenericRepository[types.JobDataEntity],
				mockTimeJobRepo *datastoreMocks.MockGenericRepository[types.TimeJobDataEntity],
				mockEventJobRepo *datastoreMocks.MockGenericRepository[types.EventJobDataEntity],
				mockConditionJobRepo *datastoreMocks.MockGenericRepository[types.ConditionJobDataEntity],
			) {
				mockJobRepo.On("GetByID", mock.Anything, "").Return(nil, nil).Once()
			},
			expectedCode:  http.StatusNotFound,
			expectedError: pkgErrors.ErrDBRecordNotFound,
		},
		{
			name: "Error - Job Not Found",
			request: types.UpdateJobDataFromUserRequest{
				JobID:             "99999",
				JobTitle:          "Non-existent Job",
				Recurring:         true,
				TimeFrame:         86400,
				JobCostPrediction: "1000000000000000000",
				TimeInterval:      3600,
			},
			mockSetup: func(
				mockJobRepo *datastoreMocks.MockGenericRepository[types.JobDataEntity],
				mockTimeJobRepo *datastoreMocks.MockGenericRepository[types.TimeJobDataEntity],
				mockEventJobRepo *datastoreMocks.MockGenericRepository[types.EventJobDataEntity],
				mockConditionJobRepo *datastoreMocks.MockGenericRepository[types.ConditionJobDataEntity],
			) {
				mockJobRepo.On("GetByID", mock.Anything, "99999").Return(nil, errors.New("not found")).Once()
			},
			expectedCode:  http.StatusNotFound,
			expectedError: pkgErrors.ErrDBRecordNotFound,
		},
		{
			name: "Error - Job Data is Nil",
			request: types.UpdateJobDataFromUserRequest{
				JobID:             "88888",
				JobTitle:          "Nil Job",
				Recurring:         true,
				TimeFrame:         86400,
				JobCostPrediction: "1000000000000000000",
				TimeInterval:      3600,
			},
			mockSetup: func(
				mockJobRepo *datastoreMocks.MockGenericRepository[types.JobDataEntity],
				mockTimeJobRepo *datastoreMocks.MockGenericRepository[types.TimeJobDataEntity],
				mockEventJobRepo *datastoreMocks.MockGenericRepository[types.EventJobDataEntity],
				mockConditionJobRepo *datastoreMocks.MockGenericRepository[types.ConditionJobDataEntity],
			) {
				mockJobRepo.On("GetByID", mock.Anything, "88888").Return(nil, nil).Once()
			},
			expectedCode:  http.StatusNotFound,
			expectedError: pkgErrors.ErrDBRecordNotFound,
		},
		{
			name: "Error - Job Update Failed",
			request: types.UpdateJobDataFromUserRequest{
				JobID:             "77777",
				JobTitle:          "Failed Job",
				Recurring:         true,
				TimeFrame:         86400,
				JobCostPrediction: "1000000000000000000",
				TimeInterval:      3600,
			},
			mockSetup: func(
				mockJobRepo *datastoreMocks.MockGenericRepository[types.JobDataEntity],
				mockTimeJobRepo *datastoreMocks.MockGenericRepository[types.TimeJobDataEntity],
				mockEventJobRepo *datastoreMocks.MockGenericRepository[types.EventJobDataEntity],
				mockConditionJobRepo *datastoreMocks.MockGenericRepository[types.ConditionJobDataEntity],
			) {
				job := &types.JobDataEntity{
					JobID:            "77777",
					TaskDefinitionID: 1,
					Recurring:        true,
					CreatedAt:        now,
					UpdatedAt:        now,
				}
				mockJobRepo.On("GetByID", mock.Anything, "77777").Return(job, nil).Once()
				mockJobRepo.On("Update", mock.Anything, mock.Anything).Return(errors.New("update failed")).Once()
			},
			expectedCode:  http.StatusInternalServerError,
			expectedError: pkgErrors.ErrDBOperationFailed,
		},
		{
			name: "Error - Time Job Update Failed",
			request: types.UpdateJobDataFromUserRequest{
				JobID:             "66666",
				JobTitle:          "Failed Time Job",
				Recurring:         true,
				TimeFrame:         86400,
				JobCostPrediction: "1000000000000000000",
				TimeInterval:      3600,
			},
			mockSetup: func(
				mockJobRepo *datastoreMocks.MockGenericRepository[types.JobDataEntity],
				mockTimeJobRepo *datastoreMocks.MockGenericRepository[types.TimeJobDataEntity],
				mockEventJobRepo *datastoreMocks.MockGenericRepository[types.EventJobDataEntity],
				mockConditionJobRepo *datastoreMocks.MockGenericRepository[types.ConditionJobDataEntity],
			) {
				job := &types.JobDataEntity{
					JobID:            "66666",
					TaskDefinitionID: 1,
					Recurring:        true,
					CreatedAt:        now,
					UpdatedAt:        now,
				}
				mockJobRepo.On("GetByID", mock.Anything, "66666").Return(job, nil).Once()
				mockJobRepo.On("Update", mock.Anything, mock.Anything).Return(nil).Once()
				mockTimeJobRepo.On("Update", mock.Anything, mock.Anything).Return(errors.New("update failed")).Once()
			},
			expectedCode:  http.StatusInternalServerError,
			expectedError: pkgErrors.ErrDBOperationFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockJobRepo, mockTimeJobRepo, mockEventJobRepo, mockConditionJobRepo, mockLogger := setupJobUpdateTestHandler()
			defer mockLogger.AssertExpectations(t)
			defer mockJobRepo.AssertExpectations(t)
			defer mockTimeJobRepo.AssertExpectations(t)
			defer mockEventJobRepo.AssertExpectations(t)
			defer mockConditionJobRepo.AssertExpectations(t)

			// Setup
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			// Create request body
			reqBody, _ := json.Marshal(tt.request)
			c.Request = httptest.NewRequest("PUT", "/jobs/update", bytes.NewBuffer(reqBody))
			c.Request.Header.Set("Content-Type", "application/json")

			// Setup mocks
			tt.mockSetup(mockJobRepo, mockTimeJobRepo, mockEventJobRepo, mockConditionJobRepo)

			// Execute
			handler.UpdateJobDataFromUser(c)

			// Assert
			assert.Equal(t, tt.expectedCode, w.Code)
			if tt.validateResult != nil {
				tt.validateResult(t, w)
			}
			if tt.expectedError != "" {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				if errMsg, ok := response["error"]; ok {
					assert.Contains(t, errMsg, tt.expectedError)
				}
			}
		})
	}
}
