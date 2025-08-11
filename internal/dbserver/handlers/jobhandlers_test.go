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
	"github.com/stretchr/testify/mock"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/types"
	pkgtypes "github.com/trigg3rX/triggerx-backend/pkg/types"
)

// Mock repositories
type MockUserRepository struct {
	mock.Mock
}

type MockJobRepository struct {
	mock.Mock
}

type MockTimeJobRepository struct {
	mock.Mock
}

type MockEventJobRepository struct {
	mock.Mock
}

// Add missing method to satisfy the interface
func (m *MockEventJobRepository) GetActiveEventJobs() ([]types.EventJobData, error) {
	args := m.Called()
	return args.Get(0).([]types.EventJobData), args.Error(1)
}

type MockConditionJobRepository struct {
	mock.Mock
}

// Add missing method to satisfy the interface
func (m *MockConditionJobRepository) GetActiveConditionJobs() ([]types.ConditionJobData, error) {
	args := m.Called()
	return args.Get(0).([]types.ConditionJobData), args.Error(1)
}

// Mock implementations
func (m *MockUserRepository) GetUserDataByAddress(address string) (int64, types.UserData, error) {
	args := m.Called(address)
	return args.Get(0).(int64), args.Get(1).(types.UserData), args.Error(2)
}

func (m *MockUserRepository) CreateNewUser(user *types.CreateUserDataRequest) (types.UserData, error) {
	args := m.Called(user)
	return args.Get(0).(types.UserData), args.Error(1)
}

func (m *MockUserRepository) UpdateUserTasksAndPoints(userID int64, tasks int64, points float64) error {
	args := m.Called(userID, tasks, points)
	return args.Error(0)
}

func (m *MockUserRepository) UpdateUserJobIDs(userID int64, jobIDs []*big.Int) error {
	args := m.Called(userID, jobIDs)
	return args.Error(0)
}

func (m *MockUserRepository) GetUserJobIDsByAddress(address string) (int64, []*big.Int, error) {
	args := m.Called(address)
	return args.Get(0).(int64), args.Get(1).([]*big.Int), args.Error(2)
}

func (m *MockUserRepository) CheckUserExists(address string) (int64, error) {
	args := m.Called(address)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockUserRepository) GetUserLeaderboard() ([]types.UserLeaderboardEntry, error) {
	args := m.Called()
	return args.Get(0).([]types.UserLeaderboardEntry), args.Error(1)
}

func (m *MockUserRepository) GetUserLeaderboardByAddress(address string) (types.UserLeaderboardEntry, error) {
	args := m.Called(address)
	return args.Get(0).(types.UserLeaderboardEntry), args.Error(1)
}

func (m *MockUserRepository) GetUserPointsByAddress(address string) (float64, error) {
	args := m.Called(address)
	return args.Get(0).(float64), args.Error(1)
}

func (m *MockUserRepository) GetUserPointsByID(userID int64) (float64, error) {
	args := m.Called(userID)
	return args.Get(0).(float64), args.Error(1)
}

func (m *MockUserRepository) UpdateUserBalance(updateData *types.UpdateUserBalanceRequest) error {
	args := m.Called(updateData)
	return args.Error(0)
}

func (m *MockUserRepository) UpdateUserEmail(address, email string) error {
	args := m.Called(address, email)
	return args.Error(0)
}

func (m *MockJobRepository) CreateNewJob(job *types.JobData) (*big.Int, error) {
	args := m.Called(job)
	return args.Get(0).(*big.Int), args.Error(1)
}

func (m *MockJobRepository) UpdateJobFromUserInDB(jobID *big.Int, updateData *types.UpdateJobDataFromUserRequest) error {
	args := m.Called(jobID, updateData)
	return args.Error(0)
}

func (m *MockJobRepository) UpdateJobStatus(jobID *big.Int, status string) error {
	args := m.Called(jobID, status)
	return args.Error(0)
}

func (m *MockJobRepository) UpdateJobLastExecutedAt(jobID *big.Int, taskID int64, jobCostActual float64, lastExecutedAt time.Time) error {
	args := m.Called(jobID, taskID, jobCostActual, lastExecutedAt)
	return args.Error(0)
}

func (m *MockJobRepository) GetJobByID(jobID *big.Int) (*types.JobData, error) {
	args := m.Called(jobID)
	return args.Get(0).(*types.JobData), args.Error(1)
}

func (m *MockJobRepository) GetTaskDefinitionIDByJobID(jobID *big.Int) (int, error) {
	args := m.Called(jobID)
	return args.Get(0).(int), args.Error(1)
}

func (m *MockJobRepository) GetTaskFeesByJobID(jobID *big.Int) ([]types.TaskFeeResponse, error) {
	args := m.Called(jobID)
	return args.Get(0).([]types.TaskFeeResponse), args.Error(1)
}

func (m *MockTimeJobRepository) CreateTimeJob(job *types.TimeJobData) error {
	args := m.Called(job)
	return args.Error(0)
}

func (m *MockTimeJobRepository) GetTimeJobByJobID(jobID *big.Int) (types.TimeJobData, error) {
	args := m.Called(jobID)
	return args.Get(0).(types.TimeJobData), args.Error(1)
}

func (m *MockTimeJobRepository) UpdateTimeJobStatus(jobID *big.Int, isActive bool) error {
	args := m.Called(jobID, isActive)
	return args.Error(0)
}

func (m *MockTimeJobRepository) CompleteTimeJob(jobID *big.Int) error {
	args := m.Called(jobID)
	return args.Error(0)
}

func (m *MockTimeJobRepository) GetTimeJobsByNextExecutionTimestamp(timestamp time.Time) ([]pkgtypes.ScheduleTimeTaskData, error) {
	args := m.Called(timestamp)
	return args.Get(0).([]pkgtypes.ScheduleTimeTaskData), args.Error(1)
}

func (m *MockTimeJobRepository) UpdateTimeJobNextExecutionTimestamp(jobID *big.Int, nextExecutionTimestamp time.Time) error {
	args := m.Called(jobID, nextExecutionTimestamp)
	return args.Error(0)
}

func (m *MockTimeJobRepository) UpdateTimeJobInterval(jobID *big.Int, timeInterval int64) error {
	args := m.Called(jobID, timeInterval)
	return args.Error(0)
}

func (m *MockTimeJobRepository) GetActiveTimeJobs() ([]types.TimeJobData, error) {
	args := m.Called()
	return args.Get(0).([]types.TimeJobData), args.Error(1)
}

func (m *MockEventJobRepository) CreateEventJob(job *types.EventJobData) error {
	args := m.Called(job)
	return args.Error(0)
}

func (m *MockEventJobRepository) GetEventJobByJobID(jobID *big.Int) (types.EventJobData, error) {
	args := m.Called(jobID)
	return args.Get(0).(types.EventJobData), args.Error(1)
}

func (m *MockEventJobRepository) UpdateEventJobStatus(jobID *big.Int, isActive bool) error {
	args := m.Called(jobID, isActive)
	return args.Error(0)
}

func (m *MockEventJobRepository) CompleteEventJob(jobID *big.Int) error {
	args := m.Called(jobID)
	return args.Error(0)
}

func (m *MockConditionJobRepository) CreateConditionJob(job *types.ConditionJobData) error {
	args := m.Called(job)
	return args.Error(0)
}

func (m *MockConditionJobRepository) GetConditionJobByJobID(jobID *big.Int) (types.ConditionJobData, error) {
	args := m.Called(jobID)
	return args.Get(0).(types.ConditionJobData), args.Error(1)
}

func (m *MockConditionJobRepository) UpdateConditionJobStatus(jobID *big.Int, isActive bool) error {
	args := m.Called(jobID, isActive)
	return args.Error(0)
}

func (m *MockConditionJobRepository) CompleteConditionJob(jobID *big.Int) error {
	args := m.Called(jobID)
	return args.Error(0)
}

// Test setup helper
func setupTestHandler() (*Handler, *MockUserRepository, *MockJobRepository, *MockTimeJobRepository, *MockEventJobRepository, *MockConditionJobRepository) {
	mockUserRepo := new(MockUserRepository)
	mockJobRepo := new(MockJobRepository)
	mockTimeJobRepo := new(MockTimeJobRepository)
	mockEventJobRepo := new(MockEventJobRepository)
	mockConditionJobRepo := new(MockConditionJobRepository)

	handler := &Handler{
		userRepository:         mockUserRepo,
		jobRepository:          mockJobRepo,
		timeJobRepository:      mockTimeJobRepo,
		eventJobRepository:     mockEventJobRepo,
		conditionJobRepository: mockConditionJobRepo,
		logger:                 &MockLogger{},
	}

	return handler, mockUserRepo, mockJobRepo, mockTimeJobRepo, mockEventJobRepo, mockConditionJobRepo
}

// Test cases
func TestCreateJobData(t *testing.T) {
	handler, mockUserRepo, mockJobRepo, mockTimeJobRepo, _, _ := setupTestHandler()

	tests := []struct {
		name          string
		input         []types.CreateJobData
		setupMocks    func()
		expectedCode  int
		expectedError string
	}{
		{
			name: "Success - Create Time-based Job",
			input: []types.CreateJobData{
				{
					UserAddress:      "0x123",
					EtherBalance:     big.NewInt(1),
					TokenBalance:     big.NewInt(100),
					JobTitle:         "Test Job",
					TaskDefinitionID: 1,
					TimeFrame:        3600,
					Recurring:        true,
					Custom:           false,
				},
			},
			setupMocks: func() {
				mockUserRepo.On("GetUserDataByAddress", "0x123").Return(int64(1), types.UserData{
					UserID:       1,
					UserAddress:  "0x123",
					EtherBalance: big.NewInt(1),
					TokenBalance: big.NewInt(100),
					UserPoints:   0.0,
					JobIDs:       []*big.Int{},
				}, nil)

				mockJobRepo.On("CreateNewJob", mock.Anything).Return(big.NewInt(1), nil)
				mockTimeJobRepo.On("CreateTimeJob", mock.Anything).Return(nil)
				mockUserRepo.On("UpdateUserTasksAndPoints", int64(1), int64(0), 10.0).Return(nil)
				mockUserRepo.On("UpdateUserJobIDs", int64(1), []*big.Int{big.NewInt(1)}).Return(nil)
			},
			expectedCode: http.StatusOK,
		},
		{
			name:          "Error - Empty Job List",
			input:         []types.CreateJobData{},
			setupMocks:    func() {},
			expectedCode:  http.StatusBadRequest,
			expectedError: "No jobs provided",
		},
		{
			name: "Error - Invalid User Address",
			input: []types.CreateJobData{
				{
					UserAddress: "0x123",
				},
			},
			setupMocks: func() {
				mockUserRepo.On("GetUserDataByAddress", "0x123").Return(int64(0), types.UserData{}, assert.AnError)
			},
			expectedCode:  http.StatusBadRequest,
			expectedError: "Invalid task definition ID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			// Setup request
			body, _ := json.Marshal(tt.input)
			c.Request = httptest.NewRequest("POST", "/", bytes.NewBuffer(body))
			c.Request.Header.Set("Content-Type", "application/json")

			// Setup mocks
			tt.setupMocks()

			// Execute
			handler.CreateJobData(c)

			// Assert
			assert.Equal(t, tt.expectedCode, w.Code)
			if tt.expectedError != "" {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response["error"], tt.expectedError)
			}
		})
	}
}

func TestUpdateJobStatus(t *testing.T) {
	handler, _, mockJobRepo, _, _, _ := setupTestHandler()

	tests := []struct {
		name          string
		jobID         string
		status        string
		setupMocks    func()
		expectedCode  int
		expectedError string
	}{
		{
			name:   "Success - Update Status to Running",
			jobID:  "1",
			status: "running",
			setupMocks: func() {
				mockJobRepo.On("UpdateJobStatus", big.NewInt(1), "running").Return(nil)
			},
			expectedCode: http.StatusOK,
		},
		{
			name:          "Error - Invalid Status",
			jobID:         "1",
			status:        "invalid",
			setupMocks:    func() {},
			expectedCode:  http.StatusBadRequest,
			expectedError: "Invalid status",
		},
		{
			name:          "Error - Invalid Job ID",
			jobID:         "invalid",
			status:        "running",
			setupMocks:    func() {},
			expectedCode:  http.StatusBadRequest,
			expectedError: "Invalid job ID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			// Setup request
			c.Params = []gin.Param{
				{Key: "job_id", Value: tt.jobID},
				{Key: "status", Value: tt.status},
			}

			// Setup mocks
			tt.setupMocks()

			// Execute
			handler.UpdateJobStatus(c)

			// Assert
			assert.Equal(t, tt.expectedCode, w.Code)
			if tt.expectedError != "" {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response["error"], tt.expectedError)
			}
		})
	}
}

func TestDeleteJobData(t *testing.T) {
	handler, _, mockJobRepo, mockTimeJobRepo, mockEventJobRepo, _ := setupTestHandler()

	tests := []struct {
		name          string
		jobID         string
		taskDefID     int
		setupMocks    func()
		expectedCode  int
		expectedError string
	}{
		{
			name:      "Success - Delete Time-based Job",
			jobID:     "1",
			taskDefID: 1,
			setupMocks: func() {
				mockJobRepo.On("GetTaskDefinitionIDByJobID", big.NewInt(1)).Return(1, nil)
				mockJobRepo.On("UpdateJobStatus", big.NewInt(1), "deleted").Return(nil)
				mockTimeJobRepo.On("UpdateTimeJobStatus", big.NewInt(1), false).Return(nil)
			},
			expectedCode: http.StatusOK,
		},
		{
			name:      "Success - Delete Event-based Job",
			jobID:     "2",
			taskDefID: 3,
			setupMocks: func() {
				mockJobRepo.On("GetTaskDefinitionIDByJobID", big.NewInt(2)).Return(3, nil)
				mockJobRepo.On("UpdateJobStatus", big.NewInt(2), "deleted").Return(nil)
				mockEventJobRepo.On("UpdateEventJobStatus", big.NewInt(2), false).Return(nil)
			},
			expectedCode: http.StatusOK,
		},
		{
			name:          "Error - Invalid Job ID",
			jobID:         "invalid",
			setupMocks:    func() {},
			expectedCode:  http.StatusBadRequest,
			expectedError: "Invalid job ID format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			// Setup request
			c.Params = []gin.Param{
				{Key: "id", Value: tt.jobID},
			}

			// Setup mocks
			tt.setupMocks()

			// Execute
			handler.DeleteJobData(c)

			// Assert
			assert.Equal(t, tt.expectedCode, w.Code)
			if tt.expectedError != "" {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response["error"], tt.expectedError)
			}
		})
	}
}

func TestGetJobsByUserAddress(t *testing.T) {
	handler, mockUserRepo, mockJobRepo, mockTimeJobRepo, mockEventJobRepo, _ := setupTestHandler()

	tests := []struct {
		name          string
		userAddress   string
		setupMocks    func()
		expectedCode  int
		expectedError string
	}{
		{
			name:        "Success - Get User Jobs",
			userAddress: "0x123",
			setupMocks: func() {
				mockUserRepo.On("GetUserJobIDsByAddress", "0x123").Return(int64(1), []*big.Int{big.NewInt(1), big.NewInt(2)}, nil)
				mockJobRepo.On("GetJobByID", big.NewInt(1)).Return(&types.JobData{
					JobID:            big.NewInt(1),
					TaskDefinitionID: 1,
				}, nil)
				mockJobRepo.On("GetJobByID", big.NewInt(2)).Return(&types.JobData{
					JobID:            big.NewInt(2),
					TaskDefinitionID: 3,
				}, nil)
				mockTimeJobRepo.On("GetTimeJobByJobID", big.NewInt(1)).Return(types.TimeJobData{}, nil)
				mockEventJobRepo.On("GetEventJobByJobID", big.NewInt(2)).Return(types.EventJobData{}, nil)
			},
			expectedCode: http.StatusOK,
		},
		{
			name:          "Error - Empty User Address",
			userAddress:   "",
			setupMocks:    func() {},
			expectedCode:  http.StatusBadRequest,
			expectedError: "No user address provided",
		},
		{
			name:        "Error - User Not Found",
			userAddress: "0x123",
			setupMocks: func() {
				mockUserRepo.On("GetUserJobIDsByAddress", "0x123").Return(int64(0), []*big.Int{}, assert.AnError)
			},
			expectedCode:  http.StatusOK,
			expectedError: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			// Setup request
			c.Params = []gin.Param{
				{Key: "user_address", Value: tt.userAddress},
			}

			// Setup mocks
			tt.setupMocks()

			// Execute
			handler.GetJobsByUserAddress(c)

			// Assert
			assert.Equal(t, tt.expectedCode, w.Code)
			if tt.expectedError != "" {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response["error"], tt.expectedError)
			} else if tt.expectedCode == http.StatusOK {
				var response []types.JobResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
			}
		})
	}
}
