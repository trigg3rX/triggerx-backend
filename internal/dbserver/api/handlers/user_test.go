package handlers

import (
	"bytes"
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

func setupUserTestHandler() (*Handler, *mocks.MockGenericRepository[types.UserDataEntity], *logging.MockLogger) {
	mockUserRepo := new(mocks.MockGenericRepository[types.UserDataEntity])
	mockLogger := new(logging.MockLogger)
	mockLogger.SetupDefaultExpectations()

	handler := &Handler{
		logger:         mockLogger,
		userRepository: mockUserRepo,
	}

	return handler, mockUserRepo, mockLogger
}

func TestGetUserDataByAddress(t *testing.T) {
	handler, mockUserRepo, mockLogger := setupUserTestHandler()
	defer mockLogger.AssertExpectations(t)
	defer mockUserRepo.AssertExpectations(t)

	validAddress := "0x1234567890123456789012345678901234567890"
	validAddressLower := "0x1234567890123456789012345678901234567890"
	invalidAddress := "invalid_address"

	tests := []struct {
		name           string
		address        string
		mockSetup      func()
		expectedCode   int
		expectedError  string
		validateResult func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:    "Success - Valid Address with User Data",
			address: validAddress,
			mockSetup: func() {
				userData := &types.UserDataEntity{
					UserAddress:   validAddressLower,
					EmailID:       "test@example.com",
					JobIDs:        []string{"1", "2", "3"},
					UserPoints:    "1000",
					TotalJobs:     3,
					TotalTasks:    10,
					CreatedAt:     time.Now().UTC(),
					LastUpdatedAt: time.Now().UTC(),
				}
				mockUserRepo.On("GetByNonID", mock.Anything, "user_address", validAddressLower).Return(userData, nil).Once()
			},
			expectedCode:  http.StatusOK,
			expectedError: "",
			validateResult: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response types.UserDataEntity
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, validAddressLower, response.UserAddress)
				assert.Equal(t, "test@example.com", response.EmailID)
				assert.Equal(t, int64(3), response.TotalJobs)
			},
		},
		{
			name:    "Error - Invalid Ethereum Address",
			address: invalidAddress,
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
			name:    "Error - User Not Found",
			address: validAddress,
			mockSetup: func() {
				mockUserRepo.On("GetByNonID", mock.Anything, "user_address", validAddressLower).Return(nil, assert.AnError).Once()
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
			name:    "Error - Repository Returns Nil User",
			address: validAddress,
			mockSetup: func() {
				mockUserRepo.On("GetByNonID", mock.Anything, "user_address", validAddressLower).Return(nil, nil).Once()
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
			c.Request = httptest.NewRequest("GET", "/users/"+tt.address, nil)
			c.Params = gin.Params{gin.Param{Key: "address", Value: tt.address}}

			// Setup mocks
			tt.mockSetup()

			// Execute
			handler.GetUserDataByAddress(c)

			// Assert
			assert.Equal(t, tt.expectedCode, w.Code)
			if tt.validateResult != nil {
				tt.validateResult(t, w)
			}

			mockUserRepo.AssertExpectations(t)
		})
	}
}

func TestUpdateUserEmail(t *testing.T) {
	handler, mockUserRepo, mockLogger := setupUserTestHandler()
	defer mockLogger.AssertExpectations(t)
	defer mockUserRepo.AssertExpectations(t)

	validAddress := "0x1234567890123456789012345678901234567890"
	validAddressLower := "0x1234567890123456789012345678901234567890"

	tests := []struct {
		name           string
		requestBody    interface{}
		mockSetup      func()
		expectedCode   int
		expectedError  string
		validateResult func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "Success - Update Email",
			requestBody: types.UpdateUserEmailRequest{
				UserAddress: validAddress,
				Email:       "newemail@example.com",
			},
			mockSetup: func() {
				existingUser := &types.UserDataEntity{
					UserAddress:   validAddressLower,
					EmailID:       "oldemail@example.com",
					JobIDs:        []string{"1"},
					UserPoints:    "500",
					TotalJobs:     1,
					TotalTasks:    5,
					CreatedAt:     time.Now().UTC(),
					LastUpdatedAt: time.Now().UTC(),
				}
				mockUserRepo.On("GetByID", mock.Anything, validAddressLower).Return(existingUser, nil).Once()
				mockUserRepo.On("Update", mock.Anything, mock.MatchedBy(func(user *types.UserDataEntity) bool {
					return user.UserAddress == validAddressLower && user.EmailID == "newemail@example.com"
				})).Return(nil).Once()
			},
			expectedCode:  http.StatusOK,
			expectedError: "",
			validateResult: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "Email updated successfully", response["message"])
			},
		},
		{
			name: "Success - Unsubscribe (Empty Email)",
			requestBody: types.UpdateUserEmailRequest{
				UserAddress: validAddress,
				Email:       "",
			},
			mockSetup: func() {
				existingUser := &types.UserDataEntity{
					UserAddress:   validAddressLower,
					EmailID:       "oldemail@example.com",
					JobIDs:        []string{"1"},
					UserPoints:    "500",
					TotalJobs:     1,
					TotalTasks:    5,
					CreatedAt:     time.Now().UTC(),
					LastUpdatedAt: time.Now().UTC(),
				}
				mockUserRepo.On("GetByID", mock.Anything, validAddressLower).Return(existingUser, nil).Once()
				mockUserRepo.On("Update", mock.Anything, mock.MatchedBy(func(user *types.UserDataEntity) bool {
					return user.UserAddress == validAddressLower && user.EmailID == ""
				})).Return(nil).Once()
			},
			expectedCode:  http.StatusOK,
			expectedError: "",
			validateResult: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "Email updated successfully", response["message"])
			},
		},
		{
			name:        "Error - Invalid Request Body",
			requestBody: "invalid json",
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
			name: "Error - User Not Found",
			requestBody: types.UpdateUserEmailRequest{
				UserAddress: validAddress,
				Email:       "newemail@example.com",
			},
			mockSetup: func() {
				mockUserRepo.On("GetByID", mock.Anything, validAddressLower).Return(nil, assert.AnError).Once()
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
			name: "Error - Repository Returns Nil User",
			requestBody: types.UpdateUserEmailRequest{
				UserAddress: validAddress,
				Email:       "newemail@example.com",
			},
			mockSetup: func() {
				mockUserRepo.On("GetByID", mock.Anything, validAddressLower).Return(nil, nil).Once()
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
			name: "Error - Update Fails",
			requestBody: types.UpdateUserEmailRequest{
				UserAddress: validAddress,
				Email:       "newemail@example.com",
			},
			mockSetup: func() {
				existingUser := &types.UserDataEntity{
					UserAddress:   validAddressLower,
					EmailID:       "oldemail@example.com",
					JobIDs:        []string{"1"},
					UserPoints:    "500",
					TotalJobs:     1,
					TotalTasks:    5,
					CreatedAt:     time.Now().UTC(),
					LastUpdatedAt: time.Now().UTC(),
				}
				mockUserRepo.On("GetByID", mock.Anything, validAddressLower).Return(existingUser, nil).Once()
				mockUserRepo.On("Update", mock.Anything, mock.Anything).Return(assert.AnError).Once()
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

			// Create request body
			var reqBody []byte
			var err error
			if _, ok := tt.requestBody.(string); ok {
				reqBody = []byte(tt.requestBody.(string))
			} else {
				reqBody, err = json.Marshal(tt.requestBody)
				assert.NoError(t, err)
			}

			c.Request = httptest.NewRequest("PUT", "/users/email", bytes.NewBuffer(reqBody))
			c.Request.Header.Set("Content-Type", "application/json")

			// Setup mocks
			tt.mockSetup()

			// Execute
			handler.UpdateUserEmail(c)

			// Assert
			assert.Equal(t, tt.expectedCode, w.Code)
			if tt.validateResult != nil {
				tt.validateResult(t, w)
			}

			mockUserRepo.AssertExpectations(t)
		})
	}
}
