package handlers

import (
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

func setupTestUserHandler() (*Handler, *MockUserRepository) {
	mockUserRepo := new(MockUserRepository)
	handler := &Handler{
		userRepository: mockUserRepo,
		logger:         &MockLogger{},
	}
	return handler, mockUserRepo
}

func TestGetUserDataByAddress(t *testing.T) {
	handler, mockUserRepo := setupTestUserHandler()

	// Create fixed timestamps for testing
	fixedTime := time.Date(2025, time.June, 2, 17, 41, 39, 0, time.Local)

	tests := []struct {
		name           string
		userAddress    string
		setupMocks     func()
		expectedCode   int
		expectedError  string
		expectedUserID int64
		expectedData   types.UserData
	}{
		{
			name:        "Success - Get User Data",
			userAddress: "0x123",
			setupMocks: func() {
				mockUserRepo.On("GetUserDataByAddress", "0x123").Return(int64(1), types.UserData{
					UserID:        1,
					UserAddress:   "0x123",
					JobIDs:        []*big.Int{big.NewInt(1), big.NewInt(2), big.NewInt(3)},
					EtherBalance:  big.NewInt(1000000000000000000), // 1 ETH
					TokenBalance:  big.NewInt(1000000),             // 1M tokens
					UserPoints:    100.0,
					TotalJobs:     3,
					TotalTasks:    10,
					CreatedAt:     fixedTime,
					LastUpdatedAt: fixedTime,
				}, nil)
			},
			expectedCode:   http.StatusOK,
			expectedUserID: 1,
			expectedData: types.UserData{
				UserID:        1,
				UserAddress:   "0x123",
				JobIDs:        []*big.Int{big.NewInt(1), big.NewInt(2), big.NewInt(3)},
				EtherBalance:  big.NewInt(1000000000000000000), // 1 ETH
				TokenBalance:  big.NewInt(1000000),             // 1M tokens
				UserPoints:    100.0,
				TotalJobs:     3,
				TotalTasks:    10,
				CreatedAt:     fixedTime,
				LastUpdatedAt: fixedTime,
			},
		},
		{
			name:          "Error - Invalid User Address",
			userAddress:   "",
			setupMocks:    func() {},
			expectedCode:  http.StatusBadRequest,
			expectedError: "Invalid user ID",
		},
		{
			name:        "Error - Database Error",
			userAddress: "0x123",
			setupMocks: func() {
				mockUserRepo.On("GetUserDataByAddress", "0x123").Return(int64(0), types.UserData{}, assert.AnError)
			},
			expectedCode:  http.StatusInternalServerError,
			expectedError: assert.AnError.Error(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mocks for isolation
			mockUserRepo.ExpectedCalls = nil

			// Setup
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			// Setup request
			c.Request = httptest.NewRequest("GET", "/", nil)
			c.Params = []gin.Param{{Key: "address", Value: tt.userAddress}}

			// Setup mocks
			tt.setupMocks()

			// Execute
			handler.GetUserDataByAddress(c)

			// Assert
			assert.Equal(t, tt.expectedCode, w.Code)
			if tt.expectedError != "" {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response["error"], tt.expectedError)
			} else {
				var response types.UserData
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedData, response)
			}
		})
	}
}

func TestGetWalletPoints(t *testing.T) {
	handler, mockUserRepo := setupTestUserHandler()

	tests := []struct {
		name           string
		walletAddress  string
		setupMocks     func()
		expectedCode   int
		expectedError  string
		expectedPoints float64
	}{
		{
			name:          "Success - Get Wallet Points",
			walletAddress: "0x123",
			setupMocks: func() {
				mockUserRepo.On("GetUserPointsByAddress", "0x123").Return(100.0, nil)
			},
			expectedCode:   http.StatusOK,
			expectedPoints: 100.0,
		},
		{
			name:          "Success - No Points Found",
			walletAddress: "0x123",
			setupMocks: func() {
				mockUserRepo.On("GetUserPointsByAddress", "0x123").Return(0.0, assert.AnError)
			},
			expectedCode:   http.StatusOK,
			expectedPoints: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mocks for isolation
			mockUserRepo.ExpectedCalls = nil

			// Setup
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			// Setup request
			c.Request = httptest.NewRequest("GET", "/", nil)
			c.Params = []gin.Param{{Key: "address", Value: tt.walletAddress}}

			// Setup mocks
			tt.setupMocks()

			// Execute
			handler.GetWalletPoints(c)

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
				assert.Equal(t, tt.expectedPoints, response["total_points"])
			}
		})
	}
}
