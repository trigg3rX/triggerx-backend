package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/trigg3rX/triggerx-backend/pkg/datastore/mocks"
	"github.com/trigg3rX/triggerx-backend/pkg/errors"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

func setupLeaderboardTestHandler() (*Handler, *mocks.MockGenericRepository[types.KeeperDataEntity], *mocks.MockGenericRepository[types.UserDataEntity], *logging.MockLogger) {
	mockKeeperRepo := new(mocks.MockGenericRepository[types.KeeperDataEntity])
	mockUserRepo := new(mocks.MockGenericRepository[types.UserDataEntity])
	mockLogger := new(logging.MockLogger)
	mockLogger.SetupDefaultExpectations()

	handler := &Handler{
		logger:           mockLogger,
		keeperRepository: mockKeeperRepo,
		userRepository:   mockUserRepo,
	}

	return handler, mockKeeperRepo, mockUserRepo, mockLogger
}

func TestGetKeeperLeaderboard(t *testing.T) {
	handler, mockKeeperRepo, _, mockLogger := setupLeaderboardTestHandler()
	defer mockLogger.AssertExpectations(t)
	defer mockKeeperRepo.AssertExpectations(t)

	tests := []struct {
		name           string
		host           string
		mockSetup      func()
		expectedCode   int
		expectedError  string
		validateResult func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "Success - All Keepers (Default Host)",
			host: "localhost:8080",
			mockSetup: func() {
				keepers := []*types.KeeperDataEntity{
					{
						KeeperAddress:   "0xkeeper1",
						OperatorID:      1,
						KeeperName:      "Keeper One",
						KeeperPoints:    "1000",
						NoExecutedTasks: 50,
						NoAttestedTasks: 45,
						OnImua:          false,
					},
					{
						KeeperAddress:   "0xkeeper2",
						OperatorID:      2,
						KeeperName:      "Keeper Two",
						KeeperPoints:    "2000",
						NoExecutedTasks: 100,
						NoAttestedTasks: 95,
						OnImua:          true,
					},
					{
						KeeperAddress:   "0xkeeper3",
						OperatorID:      3,
						KeeperName:      "Keeper Three",
						KeeperPoints:    "500",
						NoExecutedTasks: 25,
						NoAttestedTasks: 20,
						OnImua:          false,
					},
				}
				mockKeeperRepo.On("List", mock.Anything).Return(keepers, nil).Once()
			},
			expectedCode:  http.StatusOK,
			expectedError: "",
			validateResult: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response []types.KeeperLeaderboardEntry
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Len(t, response, 3)
				// Should be sorted by points (descending, string comparison: "500" > "2000" > "1000")
				assert.Equal(t, "500", response[0].KeeperPoints)
				assert.Equal(t, "0xkeeper3", response[0].KeeperAddress)
				assert.Equal(t, "2000", response[1].KeeperPoints)
				assert.Equal(t, "1000", response[2].KeeperPoints)
			},
		},
		{
			name: "Success - Filter TriggerX Keepers Only",
			host: "app.triggerx.network",
			mockSetup: func() {
				keepers := []*types.KeeperDataEntity{
					{
						KeeperAddress:   "0xkeeper1",
						OperatorID:      1,
						KeeperName:      "Keeper One",
						KeeperPoints:    "1000",
						NoExecutedTasks: 50,
						NoAttestedTasks: 45,
						OnImua:          false,
					},
					{
						KeeperAddress:   "0xkeeper2",
						OperatorID:      2,
						KeeperName:      "Keeper Two",
						KeeperPoints:    "2000",
						NoExecutedTasks: 100,
						NoAttestedTasks: 95,
						OnImua:          true,
					},
					{
						KeeperAddress:   "0xkeeper3",
						OperatorID:      3,
						KeeperName:      "Keeper Three",
						KeeperPoints:    "500",
						NoExecutedTasks: 25,
						NoAttestedTasks: 20,
						OnImua:          false,
					},
				}
				mockKeeperRepo.On("List", mock.Anything).Return(keepers, nil).Once()
			},
			expectedCode:  http.StatusOK,
			expectedError: "",
			validateResult: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response []types.KeeperLeaderboardEntry
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				// Should only include non-Imua keepers (2 keepers), sorted by points (string comparison: "500" > "1000")
				assert.Len(t, response, 2)
				assert.Equal(t, "0xkeeper3", response[0].KeeperAddress)
				assert.Equal(t, "0xkeeper1", response[1].KeeperAddress)
				assert.False(t, response[0].OnImua)
				assert.False(t, response[1].OnImua)
			},
		},
		{
			name: "Success - Filter Imua Keepers Only",
			host: "imua.triggerx.network",
			mockSetup: func() {
				keepers := []*types.KeeperDataEntity{
					{
						KeeperAddress:   "0xkeeper1",
						OperatorID:      1,
						KeeperName:      "Keeper One",
						KeeperPoints:    "1000",
						NoExecutedTasks: 50,
						NoAttestedTasks: 45,
						OnImua:          false,
					},
					{
						KeeperAddress:   "0xkeeper2",
						OperatorID:      2,
						KeeperName:      "Keeper Two",
						KeeperPoints:    "2000",
						NoExecutedTasks: 100,
						NoAttestedTasks: 95,
						OnImua:          true,
					},
				}
				mockKeeperRepo.On("List", mock.Anything).Return(keepers, nil).Once()
			},
			expectedCode:  http.StatusOK,
			expectedError: "",
			validateResult: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response []types.KeeperLeaderboardEntry
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				// Should only include Imua keepers (1 keeper)
				assert.Len(t, response, 1)
				assert.Equal(t, "0xkeeper2", response[0].KeeperAddress)
				assert.True(t, response[0].OnImua)
			},
		},
		{
			name: "Success - Empty Leaderboard",
			host: "localhost:8080",
			mockSetup: func() {
				mockKeeperRepo.On("List", mock.Anything).Return([]*types.KeeperDataEntity{}, nil).Once()
			},
			expectedCode:  http.StatusOK,
			expectedError: "",
			validateResult: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response []types.KeeperLeaderboardEntry
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Len(t, response, 0)
			},
		},
		{
			name: "Error - Database Error",
			host: "localhost:8080",
			mockSetup: func() {
				mockKeeperRepo.On("List", mock.Anything).Return(nil, assert.AnError).Once()
			},
			expectedCode:  http.StatusInternalServerError,
			expectedError: errors.ErrDBOperationFailed,
			validateResult: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, errors.ErrDBOperationFailed, response["error"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mocks
			mockKeeperRepo.ExpectedCalls = nil
			mockKeeperRepo.Calls = nil

			// Setup
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("GET", "/leaderboard/keepers", nil)
			c.Request.Host = tt.host

			// Setup mocks
			tt.mockSetup()

			// Execute
			handler.GetKeeperLeaderboard(c)

			// Assert
			assert.Equal(t, tt.expectedCode, w.Code)
			if tt.validateResult != nil {
				tt.validateResult(t, w)
			}

			mockKeeperRepo.AssertExpectations(t)
		})
	}
}

func TestGetUserLeaderboard(t *testing.T) {
	handler, _, mockUserRepo, mockLogger := setupLeaderboardTestHandler()
	defer mockLogger.AssertExpectations(t)
	defer mockUserRepo.AssertExpectations(t)

	tests := []struct {
		name           string
		mockSetup      func()
		expectedCode   int
		expectedError  string
		validateResult func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "Success - User Leaderboard",
			mockSetup: func() {
				users := []*types.UserDataEntity{
					{
						UserAddress: "0xuser1",
						UserPoints:  "5000",
						TotalJobs:   10,
						TotalTasks:  50,
					},
					{
						UserAddress: "0xuser2",
						UserPoints:  "3000",
						TotalJobs:   5,
						TotalTasks:  25,
					},
					{
						UserAddress: "0xuser3",
						UserPoints:  "8000",
						TotalJobs:   15,
						TotalTasks:  75,
					},
				}
				mockUserRepo.On("List", mock.Anything).Return(users, nil).Once()
			},
			expectedCode:  http.StatusOK,
			expectedError: "",
			validateResult: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response []types.UserLeaderboardEntry
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Len(t, response, 3)
				// Should be sorted by points (descending)
				assert.Equal(t, "8000", response[0].UserPoints)
				assert.Equal(t, "0xuser3", response[0].UserAddress)
				assert.Equal(t, "5000", response[1].UserPoints)
				assert.Equal(t, "3000", response[2].UserPoints)
			},
		},
		{
			name: "Success - Empty Leaderboard",
			mockSetup: func() {
				mockUserRepo.On("List", mock.Anything).Return([]*types.UserDataEntity{}, nil).Once()
			},
			expectedCode:  http.StatusOK,
			expectedError: "",
			validateResult: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response []types.UserLeaderboardEntry
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Len(t, response, 0)
			},
		},
		{
			name: "Error - Database Error",
			mockSetup: func() {
				mockUserRepo.On("List", mock.Anything).Return(nil, assert.AnError).Once()
			},
			expectedCode:  http.StatusInternalServerError,
			expectedError: errors.ErrDBOperationFailed,
			validateResult: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, errors.ErrDBOperationFailed, response["error"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mocks
			mockUserRepo.ExpectedCalls = nil
			mockUserRepo.Calls = nil

			// Setup
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("GET", "/leaderboard/users", nil)

			// Setup mocks
			tt.mockSetup()

			// Execute
			handler.GetUserLeaderboard(c)

			// Assert
			assert.Equal(t, tt.expectedCode, w.Code)
			if tt.validateResult != nil {
				tt.validateResult(t, w)
			}

			mockUserRepo.AssertExpectations(t)
		})
	}
}

func TestGetKeeperByIdentifier(t *testing.T) {
	handler, mockKeeperRepo, _, mockLogger := setupLeaderboardTestHandler()
	defer mockLogger.AssertExpectations(t)
	defer mockKeeperRepo.AssertExpectations(t)

	validAddress := "0x1234567890123456789012345678901234567890"

	tests := []struct {
		name           string
		queryParams    map[string]string
		mockSetup      func()
		expectedCode   int
		expectedError  string
		validateResult func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "Success - Get Keeper by Address",
			queryParams: map[string]string{
				"keeper_address": validAddress,
			},
			mockSetup: func() {
				keeper := &types.KeeperDataEntity{
					KeeperAddress:   validAddress,
					OperatorID:      1,
					KeeperName:      "Test Keeper",
					KeeperPoints:    "1000",
					NoExecutedTasks: 50,
					NoAttestedTasks: 45,
					OnImua:          false,
				}
				mockKeeperRepo.On("GetByID", mock.Anything, validAddress).Return(keeper, nil).Once()
			},
			expectedCode:  http.StatusOK,
			expectedError: "",
			validateResult: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response types.KeeperLeaderboardEntry
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, validAddress, response.KeeperAddress)
				assert.Equal(t, "Test Keeper", response.KeeperName)
				assert.Equal(t, "1000", response.KeeperPoints)
			},
		},
		{
			name: "Success - Get Keeper by Name",
			queryParams: map[string]string{
				"keeper_name": "Test Keeper",
			},
			mockSetup: func() {
				keeper := &types.KeeperDataEntity{
					KeeperAddress:   validAddress,
					OperatorID:      1,
					KeeperName:      "Test Keeper",
					KeeperPoints:    "1000",
					NoExecutedTasks: 50,
					NoAttestedTasks: 45,
					OnImua:          false,
				}
				mockKeeperRepo.On("GetByNonID", mock.Anything, "keeper_name", "Test Keeper").Return(keeper, nil).Once()
			},
			expectedCode:  http.StatusOK,
			expectedError: "",
			validateResult: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response types.KeeperLeaderboardEntry
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "Test Keeper", response.KeeperName)
				assert.Equal(t, validAddress, response.KeeperAddress)
			},
		},
		{
			name:        "Error - No Identifier Provided",
			queryParams: map[string]string{},
			mockSetup: func() {
				// No mock setup needed
			},
			expectedCode:  http.StatusBadRequest,
			expectedError: errors.ErrInvalidRequestBody,
			validateResult: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, errors.ErrInvalidRequestBody, response["error"])
			},
		},
		{
			name: "Error - Keeper Not Found by Address",
			queryParams: map[string]string{
				"keeper_address": validAddress,
			},
			mockSetup: func() {
				mockKeeperRepo.On("GetByID", mock.Anything, validAddress).Return(nil, assert.AnError).Once()
			},
			expectedCode:  http.StatusNotFound,
			expectedError: errors.ErrDBRecordNotFound,
			validateResult: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, errors.ErrDBRecordNotFound, response["error"])
			},
		},
		{
			name: "Error - Repository Returns Nil Keeper",
			queryParams: map[string]string{
				"keeper_name": "Nonexistent Keeper",
			},
			mockSetup: func() {
				mockKeeperRepo.On("GetByNonID", mock.Anything, "keeper_name", "Nonexistent Keeper").Return(nil, nil).Once()
			},
			expectedCode:  http.StatusNotFound,
			expectedError: errors.ErrDBRecordNotFound,
			validateResult: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, errors.ErrDBRecordNotFound, response["error"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mocks
			mockKeeperRepo.ExpectedCalls = nil
			mockKeeperRepo.Calls = nil

			// Setup
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			// Build URL with query params
			c.Request = httptest.NewRequest("GET", "/keeper", nil)
			q := c.Request.URL.Query()
			for key, value := range tt.queryParams {
				q.Add(key, value)
			}
			c.Request.URL.RawQuery = q.Encode()

			// Setup mocks
			tt.mockSetup()

			// Execute
			handler.GetKeeperByIdentifier(c)

			// Assert
			assert.Equal(t, tt.expectedCode, w.Code)
			if tt.validateResult != nil {
				tt.validateResult(t, w)
			}

			mockKeeperRepo.AssertExpectations(t)
		})
	}
}

func TestGetUserLeaderboardByAddress(t *testing.T) {
	handler, _, mockUserRepo, mockLogger := setupLeaderboardTestHandler()
	defer mockLogger.AssertExpectations(t)
	defer mockUserRepo.AssertExpectations(t)

	validAddress := "0x1234567890123456789012345678901234567890"
	invalidAddress := "invalid_address"

	tests := []struct {
		name           string
		userAddress    string
		mockSetup      func()
		expectedCode   int
		expectedError  string
		validateResult func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:        "Success - Get User by Address",
			userAddress: validAddress,
			mockSetup: func() {
				user := &types.UserDataEntity{
					UserAddress:   validAddress,
					EmailID:       "test@example.com",
					UserPoints:    "5000",
					TotalJobs:     10,
					TotalTasks:    50,
					CreatedAt:     time.Now().UTC(),
					LastUpdatedAt: time.Now().UTC(),
				}
				mockUserRepo.On("GetByID", mock.Anything, validAddress).Return(user, nil).Once()
			},
			expectedCode:  http.StatusOK,
			expectedError: "",
			validateResult: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response types.UserLeaderboardEntry
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, validAddress, response.UserAddress)
				assert.Equal(t, "5000", response.UserPoints)
				assert.Equal(t, int64(10), response.TotalJobs)
				assert.Equal(t, int64(50), response.TotalTasks)
			},
		},
		{
			name:        "Error - Invalid Address",
			userAddress: invalidAddress,
			mockSetup: func() {
				// No mock setup needed - validation fails before repository call
			},
			expectedCode:  http.StatusBadRequest,
			expectedError: errors.ErrInvalidRequestBody,
			validateResult: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, errors.ErrInvalidRequestBody, response["error"])
			},
		},
		{
			name:        "Error - User Not Found",
			userAddress: validAddress,
			mockSetup: func() {
				mockUserRepo.On("GetByID", mock.Anything, validAddress).Return(nil, assert.AnError).Once()
			},
			expectedCode:  http.StatusNotFound,
			expectedError: errors.ErrDBRecordNotFound,
			validateResult: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, errors.ErrDBRecordNotFound, response["error"])
			},
		},
		{
			name:        "Error - Repository Returns Nil User",
			userAddress: validAddress,
			mockSetup: func() {
				mockUserRepo.On("GetByID", mock.Anything, validAddress).Return(nil, nil).Once()
			},
			expectedCode:  http.StatusNotFound,
			expectedError: errors.ErrDBRecordNotFound,
			validateResult: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, errors.ErrDBRecordNotFound, response["error"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mocks
			mockUserRepo.ExpectedCalls = nil
			mockUserRepo.Calls = nil

			// Setup
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("GET", "/user?user_address="+tt.userAddress, nil)

			// Setup mocks
			tt.mockSetup()

			// Execute
			handler.GetUserLeaderboardByAddress(c)

			// Assert
			assert.Equal(t, tt.expectedCode, w.Code)
			if tt.validateResult != nil {
				tt.validateResult(t, w)
			}

			mockUserRepo.AssertExpectations(t)
		})
	}
}
