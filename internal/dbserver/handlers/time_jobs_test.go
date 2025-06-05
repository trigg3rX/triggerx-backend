package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/types"
)

// Mock time.Now() for testing
var mockTime = time.Date(2025, time.June, 2, 17, 41, 39, 0, time.Local)

func setupTestTimeJobHandler() (*Handler, *MockTimeJobRepository) {
	mockTimeJobRepo := new(MockTimeJobRepository)
	handler := &Handler{
		timeJobRepository: mockTimeJobRepo,
		logger:            &MockLogger{},
	}
	return handler, mockTimeJobRepo
}

func TestGetTimeBasedJobs(t *testing.T) {
	// Save original time.Now and restore it after the test
	originalTimeNow := timeNow
	defer func() { timeNow = originalTimeNow }()

	// Set our mock time
	timeNow = func() time.Time { return mockTime }

	handler, mockTimeJobRepo := setupTestTimeJobHandler()

	tests := []struct {
		name          string
		pollInterval  string
		setupMocks    func()
		expectedCode  int
		expectedError string
		expectedJobs  []types.TimeJobData
	}{
		{
			name:         "Success - Get Time Based Jobs",
			pollInterval: "60",
			setupMocks: func() {
				nextExecutionTime := mockTime.Add(60 * time.Second)
				mockTimeJobRepo.On("GetTimeJobsByNextExecutionTimestamp", nextExecutionTime).Return([]types.TimeJobData{
					{
						JobID:                     1,
						TimeFrame:                 3600,
						Recurring:                 true,
						TimeInterval:              60,
						ScheduleType:              "interval",
						CronExpression:            "",
						SpecificSchedule:          "",
						NextExecutionTimestamp:    nextExecutionTime,
						TargetChainID:             "1",
						TargetContractAddress:     "0x123",
						TargetFunction:            "execute",
						ABI:                       "[]",
						ArgType:                   1,
						Arguments:                 []string{"arg1", "arg2"},
						DynamicArgumentsScriptUrl: "",
						IsCompleted:               false,
						IsActive:                  true,
					},
					{
						JobID:                     2,
						TimeFrame:                 7200,
						Recurring:                 false,
						TimeInterval:              120,
						ScheduleType:              "specific",
						CronExpression:            "",
						SpecificSchedule:          "2025-06-02T18:00:00Z",
						NextExecutionTimestamp:    nextExecutionTime,
						TargetChainID:             "1",
						TargetContractAddress:     "0x456",
						TargetFunction:            "execute",
						ABI:                       "[]",
						ArgType:                   1,
						Arguments:                 []string{"arg3", "arg4"},
						DynamicArgumentsScriptUrl: "",
						IsCompleted:               false,
						IsActive:                  true,
					},
				}, nil)
			},
			expectedCode: http.StatusOK,
			expectedJobs: []types.TimeJobData{
				{
					JobID:                     1,
					TimeFrame:                 3600,
					Recurring:                 true,
					TimeInterval:              60,
					ScheduleType:              "interval",
					CronExpression:            "",
					SpecificSchedule:          "",
					NextExecutionTimestamp:    mockTime.Add(60 * time.Second),
					TargetChainID:             "1",
					TargetContractAddress:     "0x123",
					TargetFunction:            "execute",
					ABI:                       "[]",
					ArgType:                   1,
					Arguments:                 []string{"arg1", "arg2"},
					DynamicArgumentsScriptUrl: "",
					IsCompleted:               false,
					IsActive:                  true,
				},
				{
					JobID:                     2,
					TimeFrame:                 7200,
					Recurring:                 false,
					TimeInterval:              120,
					ScheduleType:              "specific",
					CronExpression:            "",
					SpecificSchedule:          "2025-06-02T18:00:00Z",
					NextExecutionTimestamp:    mockTime.Add(60 * time.Second),
					TargetChainID:             "1",
					TargetContractAddress:     "0x456",
					TargetFunction:            "execute",
					ABI:                       "[]",
					ArgType:                   1,
					Arguments:                 []string{"arg3", "arg4"},
					DynamicArgumentsScriptUrl: "",
					IsCompleted:               false,
					IsActive:                  true,
				},
			},
		},
		{
			name:          "Error - Invalid Poll Interval",
			pollInterval:  "invalid",
			setupMocks:    func() {},
			expectedCode:  http.StatusBadRequest,
			expectedError: "invalid poll interval",
		},
		{
			name:         "Error - Database Error",
			pollInterval: "60",
			setupMocks: func() {
				nextExecutionTime := mockTime.Add(60 * time.Second)
				mockTimeJobRepo.On("GetTimeJobsByNextExecutionTimestamp", nextExecutionTime).Return([]types.TimeJobData{}, assert.AnError)
			},
			expectedCode:  http.StatusInternalServerError,
			expectedError: assert.AnError.Error(),
		},
		{
			name:         "Success - No Jobs Found",
			pollInterval: "60",
			setupMocks: func() {
				nextExecutionTime := mockTime.Add(60 * time.Second)
				mockTimeJobRepo.On("GetTimeJobsByNextExecutionTimestamp", nextExecutionTime).Return([]types.TimeJobData{}, nil)
			},
			expectedCode: http.StatusOK,
			expectedJobs: []types.TimeJobData{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mocks for isolation
			mockTimeJobRepo.ExpectedCalls = nil

			// Setup
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			// Setup request
			c.Request = httptest.NewRequest("GET", "/", nil)
			q := c.Request.URL.Query()
			q.Add("pollInterval", tt.pollInterval)
			c.Request.URL.RawQuery = q.Encode()

			// Setup mocks
			tt.setupMocks()

			// Execute
			handler.GetTimeBasedJobs(c)

			// Assert
			assert.Equal(t, tt.expectedCode, w.Code)
			if tt.expectedError != "" {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response["error"], tt.expectedError)
			} else {
				var response []types.TimeJobData
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedJobs, response)
			}
		})
	}
}
