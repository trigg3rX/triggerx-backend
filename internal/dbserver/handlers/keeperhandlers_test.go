package handlers

import (
	"bytes"
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/types"
	commonTypes "github.com/trigg3rX/triggerx-backend/pkg/types"
)

// MockKeeperRepository is a mock implementation of the KeeperRepository interface
type MockKeeperRepository struct {
	mock.Mock
}

// MockTaskRepository is a mock implementation of the TaskRepository interface
type MockTaskRepository struct {
	mock.Mock
}

// Mock implementations for KeeperRepository
func (m *MockKeeperRepository) CheckKeeperExists(address string) (int64, error) {
	args := m.Called(address)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockKeeperRepository) CreateKeeper(keeperData types.CreateKeeperData) (int64, error) {
	args := m.Called(keeperData)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockKeeperRepository) GetKeeperAsPerformer() ([]types.GetPerformerData, error) {
	args := m.Called()
	return args.Get(0).([]types.GetPerformerData), args.Error(1)
}

func (m *MockKeeperRepository) GetKeeperDataByID(id int64) (commonTypes.KeeperData, error) {
	args := m.Called(id)
	return args.Get(0).(commonTypes.KeeperData), args.Error(1)
}

func (m *MockKeeperRepository) IncrementKeeperTaskCount(id int64) (int64, error) {
	args := m.Called(id)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockKeeperRepository) GetKeeperTaskCount(id int64) (int64, error) {
	args := m.Called(id)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockKeeperRepository) UpdateKeeperPoints(id int64, taskFee float64) (float64, error) {
	args := m.Called(id, taskFee)
	return args.Get(0).(float64), args.Error(1)
}

func (m *MockKeeperRepository) GetKeeperPointsByIDInDB(id int64) (float64, error) {
	args := m.Called(id)
	return args.Get(0).(float64), args.Error(1)
}

func (m *MockKeeperRepository) UpdateKeeperChatID(address string, chatID int64) error {
	args := m.Called(address, chatID)
	return args.Error(0)
}

func (m *MockKeeperRepository) GetKeeperCommunicationInfo(id int64) (types.KeeperCommunicationInfo, error) {
	args := m.Called(id)
	return args.Get(0).(types.KeeperCommunicationInfo), args.Error(1)
}

func (m *MockKeeperRepository) GetKeeperLeaderboard() ([]types.KeeperLeaderboardEntry, error) {
	args := m.Called()
	return args.Get(0).([]types.KeeperLeaderboardEntry), args.Error(1)
}

func (m *MockKeeperRepository) GetKeeperLeaderboardByOnImua(onImua bool) ([]types.KeeperLeaderboardEntry, error) {
	args := m.Called(onImua)
	return args.Get(0).([]types.KeeperLeaderboardEntry), args.Error(1)
}

func (m *MockKeeperRepository) GetKeeperLeaderboardByIdentifierInDB(address string, name string) (types.KeeperLeaderboardEntry, error) {
	args := m.Called(address, name)
	return args.Get(0).(types.KeeperLeaderboardEntry), args.Error(1)
}

// Mock implementation for TaskRepository
func (m *MockTaskRepository) GetTaskFee(taskID int64) (float64, error) {
	args := m.Called(taskID)
	return args.Get(0).(float64), args.Error(1)
}

func (m *MockTaskRepository) CreateTaskDataInDB(taskData *types.CreateTaskDataRequest) (int64, error) {
	args := m.Called(taskData)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockTaskRepository) GetTaskDataByID(taskID int64) (commonTypes.TaskData, error) {
	args := m.Called(taskID)
	return args.Get(0).(commonTypes.TaskData), args.Error(1)
}

// Update the return type of GetTasksByJobID to match the interface
func (m *MockTaskRepository) GetTasksByJobID(jobID *big.Int) ([]types.GetTasksByJobID, error) {
	args := m.Called(jobID)
	return args.Get(0).([]types.GetTasksByJobID), args.Error(1)
}

func (m *MockTaskRepository) UpdateTaskAttestationDataInDB(task *types.UpdateTaskAttestationDataRequest) error {
	args := m.Called(task)
	return args.Error(0)
}

func (m *MockTaskRepository) UpdateTaskExecutionDataInDB(task *types.UpdateTaskExecutionDataRequest) error {
	args := m.Called(task)
	return args.Error(0)
}

func (m *MockTaskRepository) UpdateTaskFee(taskID int64, fee float64) error {
	args := m.Called(taskID, fee)
	return args.Error(0)
}

func (m *MockTaskRepository) AddTaskIDToJob(jobID *big.Int, taskID int64) error {
	args := m.Called(jobID, taskID)
	return args.Error(0)
}

func (m *MockTaskRepository) AddTaskPerformerID(taskID int64, performerID int64) error {
	args := m.Called(taskID, performerID)
	return args.Error(0)
}

func (m *MockTaskRepository) UpdateTaskNumberAndStatus(taskID int64, taskNumber int64, status string, txHash string) error {
	args := m.Called(taskID, taskNumber, status, txHash)
	return args.Error(0)
}

func (m *MockTaskRepository) GetCreatedChainIDByJobID(jobID *big.Int) (string, error) {
	args := m.Called(jobID)
	return args.String(0), args.Error(1)
}

func (m *MockTaskRepository) GetRecentTasks(limit int) ([]types.RecentTaskResponse, error) {
	args := m.Called(limit)
	return args.Get(0).([]types.RecentTaskResponse), args.Error(1)
}

// Test setup helper
func setupTestKeeperHandler() (*Handler, *MockKeeperRepository, *MockTaskRepository) {
	mockKeeperRepo := new(MockKeeperRepository)
	mockTaskRepo := new(MockTaskRepository)

	handler := &Handler{
		keeperRepository: mockKeeperRepo, // This will cause a compile error if MockKeeperRepository does not implement all methods of KeeperRepository
		taskRepository:   mockTaskRepo,
		logger:           &MockLogger{},
	}

	return handler, mockKeeperRepo, mockTaskRepo
}

func TestCreateKeeperData(t *testing.T) {
	handler, mockKeeperRepo, _ := setupTestKeeperHandler()

	tests := []struct {
		name          string
		input         types.CreateKeeperData
		setupMocks    func()
		expectedCode  int
		expectedError string
	}{
		{
			name: "Success - Create New Keeper",
			input: types.CreateKeeperData{
				KeeperAddress: "0x123",
				KeeperName:    "Test Keeper",
				EmailID:       "test@example.com",
			},
			setupMocks: func() {
				mockKeeperRepo.On("CheckKeeperExists", "0x123").Return(int64(-1), nil)
				mockKeeperRepo.On("CreateKeeper", mock.Anything).Return(int64(1), nil)
			},
			expectedCode: http.StatusCreated,
		},
		{
			name: "Success - Keeper Already Exists",
			input: types.CreateKeeperData{
				KeeperAddress: "0x123",
				KeeperName:    "Test Keeper",
				EmailID:       "test@example.com",
			},
			setupMocks: func() {
				mockKeeperRepo.On("CheckKeeperExists", "0x123").Return(int64(1), nil)
			},
			expectedCode: http.StatusCreated,
		},
		{
			name: "Error - Invalid Input",
			input: types.CreateKeeperData{
				KeeperAddress: "", // Invalid empty address
			},
			setupMocks: func() {
				// No mock setup needed as it should fail before repository call
			},
			expectedCode:  http.StatusBadRequest,
			expectedError: "Invalid keeper address",
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
			handler.CreateKeeperData(c)

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

func TestGetPerformers(t *testing.T) {
	handler, mockKeeperRepo, _ := setupTestKeeperHandler()

	tests := []struct {
		name          string
		setupMocks    func()
		expectedCode  int
		expectedError string
	}{
		{
			name: "Success - Get Performers",
			setupMocks: func() {
				mockKeeperRepo.On("GetKeeperAsPerformer").Return([]types.GetPerformerData{
					{KeeperID: 1, KeeperAddress: "0x123"},
					{KeeperID: 2, KeeperAddress: "0x456"},
				}, nil)
			},
			expectedCode: http.StatusOK,
		},
		{
			name: "Error - Database Error",
			setupMocks: func() {
				mockKeeperRepo.On("GetKeeperAsPerformer").Return([]types.GetPerformerData{}, assert.AnError)
			},
			expectedCode:  http.StatusInternalServerError,
			expectedError: assert.AnError.Error(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mocks for isolation
			mockKeeperRepo.ExpectedCalls = nil

			// Setup
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			// Setup mocks
			tt.setupMocks()

			// Execute
			handler.GetPerformers(c)

			// Assert
			assert.Equal(t, tt.expectedCode, w.Code)
			if tt.expectedError != "" {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response["error"], tt.expectedError)
			} else {
				var response []types.GetPerformerData
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Len(t, response, 2)
			}
		})
	}
}

func TestGetKeeperData(t *testing.T) {
	handler, mockKeeperRepo, _ := setupTestKeeperHandler()

	tests := []struct {
		name          string
		keeperID      string
		setupMocks    func()
		expectedCode  int
		expectedError string
	}{
		{
			name:     "Success - Get Keeper Data",
			keeperID: "1",
			setupMocks: func() {
				mockKeeperRepo.On("GetKeeperDataByID", int64(1)).Return(commonTypes.KeeperData{
					KeeperID:        1,
					KeeperAddress:   "0x123",
					KeeperName:      "Test Keeper",
					EmailID:         "test@example.com",
					NoExecutedTasks: 5,
					KeeperPoints:    100,
				}, nil)
			},
			expectedCode: http.StatusOK,
		},
		{
			name:          "Error - Invalid Keeper ID",
			keeperID:      "invalid",
			setupMocks:    func() {},
			expectedCode:  http.StatusBadRequest,
			expectedError: "Invalid keeper ID format",
		},
		{
			name:     "Error - Keeper Not Found",
			keeperID: "999",
			setupMocks: func() {
				mockKeeperRepo.On("GetKeeperDataByID", int64(999)).Return(commonTypes.KeeperData{}, assert.AnError)
			},
			expectedCode:  http.StatusInternalServerError,
			expectedError: assert.AnError.Error(),
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
				{Key: "id", Value: tt.keeperID},
			}

			// Setup mocks
			tt.setupMocks()

			// Execute
			handler.GetKeeperData(c)

			// Assert
			assert.Equal(t, tt.expectedCode, w.Code)
			if tt.expectedError != "" {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response["error"], tt.expectedError)
			} else {
				var response commonTypes.KeeperData
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, int64(1), response.KeeperID)
			}
		})
	}
}

func TestIncrementKeeperTaskCount(t *testing.T) {
	handler, mockKeeperRepo, _ := setupTestKeeperHandler()

	tests := []struct {
		name          string
		keeperID      string
		setupMocks    func()
		expectedCode  int
		expectedError string
	}{
		{
			name:     "Success - Increment Task Count",
			keeperID: "1",
			setupMocks: func() {
				mockKeeperRepo.On("IncrementKeeperTaskCount", int64(1)).Return(int64(6), nil)
			},
			expectedCode: http.StatusOK,
		},
		{
			name:          "Error - Invalid Keeper ID",
			keeperID:      "invalid",
			setupMocks:    func() {},
			expectedCode:  http.StatusBadRequest,
			expectedError: "Invalid keeper ID format",
		},
		{
			name:     "Error - Database Error",
			keeperID: "1",
			setupMocks: func() {
				mockKeeperRepo.On("IncrementKeeperTaskCount", int64(1)).Return(int64(0), assert.AnError)
			},
			expectedCode:  http.StatusInternalServerError,
			expectedError: assert.AnError.Error(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mocks for isolation
			mockKeeperRepo.ExpectedCalls = nil

			// Setup
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			// Setup request
			c.Params = []gin.Param{
				{Key: "id", Value: tt.keeperID},
			}

			// Setup mocks
			tt.setupMocks()

			// Execute
			handler.IncrementKeeperTaskCount(c)

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
				assert.Equal(t, int64(6), response["no_executed_tasks"])
			}
		})
	}
}

func TestAddTaskFeeToKeeperPoints(t *testing.T) {
	handler, mockKeeperRepo, mockTaskRepo := setupTestKeeperHandler()

	tests := []struct {
		name          string
		keeperID      string
		taskID        int64
		setupMocks    func()
		expectedCode  int
		expectedError string
	}{
		{
			name:     "Success - Add Task Fee",
			keeperID: "1",
			taskID:   100,
			setupMocks: func() {
				mockTaskRepo.On("GetTaskFee", int64(100)).Return(float64(50), nil)
				mockKeeperRepo.On("UpdateKeeperPoints", int64(1), float64(50)).Return(float64(150), nil)
			},
			expectedCode: http.StatusOK,
		},
		{
			name:          "Error - Invalid Keeper ID",
			keeperID:      "invalid",
			taskID:        100,
			setupMocks:    func() {},
			expectedCode:  http.StatusBadRequest,
			expectedError: "Invalid keeper ID format",
		},
		{
			name:     "Error - Task Fee Not Found",
			keeperID: "1",
			taskID:   999,
			setupMocks: func() {
				mockTaskRepo.On("GetTaskFee", int64(999)).Return(float64(0), assert.AnError)
			},
			expectedCode:  http.StatusInternalServerError,
			expectedError: assert.AnError.Error(),
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
				{Key: "id", Value: tt.keeperID},
			}
			body, _ := json.Marshal(map[string]int64{"task_id": tt.taskID})
			c.Request = httptest.NewRequest("POST", "/", bytes.NewBuffer(body))
			c.Request.Header.Set("Content-Type", "application/json")

			// Setup mocks
			tt.setupMocks()

			// Execute
			handler.AddTaskFeeToKeeperPoints(c)

			// Assert
			assert.Equal(t, tt.expectedCode, w.Code)
			if tt.expectedError != "" {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response["error"], tt.expectedError)
			} else {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, float64(tt.taskID), response["task_id"])
				assert.Equal(t, float64(50), response["task_fee"])
				assert.Equal(t, float64(150), response["keeper_points"])
			}
		})
	}
}

func TestUpdateKeeperChatID(t *testing.T) {
	handler, mockKeeperRepo, _ := setupTestKeeperHandler()

	tests := []struct {
		name          string
		input         types.UpdateKeeperChatIDRequest
		setupMocks    func()
		expectedCode  int
		expectedError string
	}{
		{
			name: "Success - Update Chat ID",
			input: types.UpdateKeeperChatIDRequest{
				KeeperAddress: "0x123",
				ChatID:        12345,
			},
			setupMocks: func() {
				mockKeeperRepo.On("UpdateKeeperChatID", "0x123", int64(12345)).Return(nil)
			},
			expectedCode: http.StatusOK,
		},
		{
			name: "Error - Invalid Input",
			input: types.UpdateKeeperChatIDRequest{
				KeeperAddress: "", // Invalid empty address
				ChatID:        12345,
			},
			setupMocks:    func() {},
			expectedCode:  http.StatusBadRequest,
			expectedError: "Invalid keeper address",
		},
		{
			name: "Error - Database Error",
			input: types.UpdateKeeperChatIDRequest{
				KeeperAddress: "0x123",
				ChatID:        12345,
			},
			setupMocks: func() {
				mockKeeperRepo.On("UpdateKeeperChatID", "0x123", int64(12345)).Return(assert.AnError)
			},
			expectedCode:  http.StatusInternalServerError,
			expectedError: assert.AnError.Error(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mocks for isolation
			mockKeeperRepo.ExpectedCalls = nil

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
			handler.UpdateKeeperChatID(c)

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
				assert.Equal(t, "Chat ID updated successfully", response["message"])
			}
		})
	}
}

func TestGetKeeperCommunicationInfo(t *testing.T) {
	handler, mockKeeperRepo, _ := setupTestKeeperHandler()

	tests := []struct {
		name          string
		keeperID      string
		setupMocks    func()
		expectedCode  int
		expectedError string
	}{
		{
			name:     "Success - Get Communication Info",
			keeperID: "1",
			setupMocks: func() {
				mockKeeperRepo.On("GetKeeperCommunicationInfo", int64(1)).Return(types.KeeperCommunicationInfo{
					KeeperName: "Test Keeper",
					EmailID:    "test@example.com",
					ChatID:     12345,
				}, nil)
			},
			expectedCode: http.StatusOK,
		},
		{
			name:          "Error - Invalid Keeper ID",
			keeperID:      "invalid",
			setupMocks:    func() {},
			expectedCode:  http.StatusBadRequest,
			expectedError: "Invalid keeper ID format",
		},
		{
			name:     "Error - Keeper Not Found",
			keeperID: "999",
			setupMocks: func() {
				mockKeeperRepo.On("GetKeeperCommunicationInfo", int64(999)).Return(types.KeeperCommunicationInfo{}, assert.AnError)
			},
			expectedCode:  http.StatusInternalServerError,
			expectedError: assert.AnError.Error(),
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
				{Key: "id", Value: tt.keeperID},
			}

			// Setup mocks
			tt.setupMocks()

			// Execute
			handler.GetKeeperCommunicationInfo(c)

			// Assert
			assert.Equal(t, tt.expectedCode, w.Code)
			if tt.expectedError != "" {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response["error"], tt.expectedError)
			} else {
				var response types.KeeperCommunicationInfo
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "Test Keeper", response.KeeperName)
				assert.Equal(t, "test@example.com", response.EmailID)
				assert.Equal(t, int64(12345), response.ChatID)
			}
		})
	}
}

// Add missing methods to MockKeeperRepository
func (m *MockKeeperRepository) CheckKeeperExistsByAddress(address string) (int64, error) {
	args := m.Called(address)
	var defaultReturnInt64 int64 = 0
	if args.Get(0) == nil {
		return defaultReturnInt64, args.Error(1)
	}
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockKeeperRepository) CreateOrUpdateKeeperFromGoogleForm(keeperData types.GoogleFormCreateKeeperData) (int64, error) {
	args := m.Called(keeperData)
	var defaultReturnInt64 int64 = 0
	if args.Get(0) == nil {
		return defaultReturnInt64, args.Error(1)
	}
	return args.Get(0).(int64), args.Error(1)
}
