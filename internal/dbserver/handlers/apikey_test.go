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
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

func setupApiKeyTestRouter() (*gin.Engine, *mocks.MockGenericRepository[types.ApiKeyDataEntity], *logging.MockLogger) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockRepo := new(mocks.MockGenericRepository[types.ApiKeyDataEntity])
	mockLogger := new(logging.MockLogger)
	mockLogger.SetupDefaultExpectations()

	handler := &Handler{
		logger:            mockLogger,
		apiKeysRepository: mockRepo,
	}

	router.POST("/api-keys", handler.CreateApiKey)
	router.PUT("/api-keys/:key", handler.DeleteApiKey)
	router.GET("/api-keys/owner/:owner", handler.GetApiKeysByOwner)

	return router, mockRepo, mockLogger
}

func TestCreateApiKey(t *testing.T) {
	router, mockRepo, mockLogger := setupApiKeyTestRouter()
	defer mockLogger.AssertExpectations(t)

	tests := []struct {
		name           string
		request        interface{}
		mockSetup      func()
		expectedStatus int
		checkResponse  func(*testing.T, map[string]interface{})
	}{
		{
			name: "Success - Create new API key",
			request: types.CreateApiKeyRequest{
				Owner:     "test-owner",
				RateLimit: 100,
			},
			mockSetup: func() {
				mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*types.ApiKeyDataEntity")).Return(nil).Once()
			},
			expectedStatus: http.StatusCreated,
			checkResponse: func(t *testing.T, response map[string]interface{}) {
				assert.Equal(t, "test-owner", response["owner"])
				assert.Equal(t, true, response["is_active"])
				assert.NotEmpty(t, response["key"])
				// Key should start with "TGRX-"
				assert.Contains(t, response["key"].(string), "TGRX-")
			},
		},
		{
			name:           "Error - Invalid request body",
			request:        "invalid json",
			mockSetup:      func() {},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, response map[string]interface{}) {
				assert.Equal(t, "Invalid request body", response["error"])
			},
		},
		{
			name: "Error - Database error on create",
			request: types.CreateApiKeyRequest{
				Owner:     "test-owner",
				RateLimit: 100,
			},
			mockSetup: func() {
				mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*types.ApiKeyDataEntity")).Return(assert.AnError).Once()
			},
			expectedStatus: http.StatusInternalServerError,
			checkResponse: func(t *testing.T, response map[string]interface{}) {
				assert.Equal(t, "Database operation failed", response["error"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mock
			mockRepo.ExpectedCalls = nil
			mockRepo.Calls = nil
			tt.mockSetup()

			var body []byte
			var err error
			if str, ok := tt.request.(string); ok {
				body = []byte(str)
			} else {
				body, err = json.Marshal(tt.request)
				assert.NoError(t, err)
			}

			req := httptest.NewRequest(http.MethodPost, "/api-keys", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.checkResponse != nil {
				var response map[string]interface{}
				err = json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				tt.checkResponse(t, response)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestDeleteApiKey(t *testing.T) {
	router, mockRepo, mockLogger := setupApiKeyTestRouter()
	defer mockLogger.AssertExpectations(t)

	tests := []struct {
		name           string
		requestBody    types.DeleteApiKeyRequest
		mockSetup      func()
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "Success - Delete API key by direct key",
			requestBody: types.DeleteApiKeyRequest{
				Key:   "TGRX-test-key-123",
				Owner: "test-owner",
			},
			mockSetup: func() {
				existingKey := &types.ApiKeyDataEntity{
					Key:       "TGRX-test-key-123",
					Owner:     "test-owner",
					IsActive:  true,
					RateLimit: 100,
				}
				// Mock GetByID since the key is not masked
				mockRepo.On("GetByID", mock.Anything, "TGRX-test-key-123").Return(existingKey, nil).Once()
				mockRepo.On("Update", mock.Anything, mock.MatchedBy(func(data *types.ApiKeyDataEntity) bool {
					return data.Key == "TGRX-test-key-123" && !data.IsActive
				})).Return(nil).Once()
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusOK, w.Code)
			},
		},
		{
			name: "Success - Delete API key by masked key",
			requestBody: types.DeleteApiKeyRequest{
				// TGRX-1234567890123456 (21 chars) -> TGRX*************3456
				Key:   "TGRX*************3456",
				Owner: "test-owner",
			},
			mockSetup: func() {
				existingKeys := []*types.ApiKeyDataEntity{
					{
						Key:       "TGRX-1234567890123456",
						Owner:     "test-owner",
						IsActive:  true,
						RateLimit: 100,
					},
				}
				mockRepo.On("GetByField", mock.Anything, "owner", "test-owner").Return(existingKeys, nil).Once()
				mockRepo.On("Update", mock.Anything, mock.MatchedBy(func(data *types.ApiKeyDataEntity) bool {
					return data.Key == "TGRX-1234567890123456" && !data.IsActive
				})).Return(nil).Once()
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusOK, w.Code)
			},
		},
		{
			name: "Error - Database error on update",
			requestBody: types.DeleteApiKeyRequest{
				Key:   "TGRX-test-key-123",
				Owner: "test-owner",
			},
			mockSetup: func() {
				existingKey := &types.ApiKeyDataEntity{
					Key:       "TGRX-test-key-123",
					Owner:     "test-owner",
					IsActive:  true,
					RateLimit: 100,
				}
				mockRepo.On("GetByID", mock.Anything, "TGRX-test-key-123").Return(existingKey, nil).Once()
				mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*types.ApiKeyDataEntity")).Return(assert.AnError).Once()
			},
			expectedStatus: http.StatusInternalServerError,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "Database operation failed", response["error"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mock
			mockRepo.ExpectedCalls = nil
			mockRepo.Calls = nil
			tt.mockSetup()

			body, err := json.Marshal(tt.requestBody)
			assert.NoError(t, err)

			req := httptest.NewRequest(http.MethodPut, "/api-keys/"+tt.requestBody.Key, bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.checkResponse != nil {
				tt.checkResponse(t, w)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestGetApiKeysByOwner(t *testing.T) {
	router, mockRepo, mockLogger := setupApiKeyTestRouter()
	defer mockLogger.AssertExpectations(t)

	tests := []struct {
		name           string
		owner          string
		mockSetup      func()
		expectedStatus int
		checkResponse  func(*testing.T, map[string]interface{})
	}{
		{
			name:  "Success - Get API keys for owner",
			owner: "test-owner",
			mockSetup: func() {
				existingKeys := []*types.ApiKeyDataEntity{
					{
						Key:          "TGRX-test-key-123",
						Owner:        "test-owner",
						IsActive:     true,
						RateLimit:    100,
						SuccessCount: 10,
						FailedCount:  2,
						LastUsed:     time.Now().UTC(),
						CreatedAt:    time.Now().UTC(),
					},
				}
				mockRepo.On("GetByField", mock.Anything, "owner", "test-owner").Return(existingKeys, nil).Once()
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, response map[string]interface{}) {
				// Response should be an array
				// Since the response is an array, we need to check differently
				assert.NotNil(t, response)
			},
		},
		{
			name:  "Error - No API keys found",
			owner: "non-existent-owner",
			mockSetup: func() {
				mockRepo.On("GetByField", mock.Anything, "owner", "non-existent-owner").Return([]*types.ApiKeyDataEntity{}, nil).Once()
			},
			expectedStatus: http.StatusNotFound,
			checkResponse: func(t *testing.T, response map[string]interface{}) {
				assert.Equal(t, "Database record not found", response["error"])
			},
		},
		{
			name:  "Error - Database error",
			owner: "test-owner",
			mockSetup: func() {
				mockRepo.On("GetByField", mock.Anything, "owner", "test-owner").Return(nil, assert.AnError).Once()
			},
			expectedStatus: http.StatusNotFound,
			checkResponse: func(t *testing.T, response map[string]interface{}) {
				assert.Equal(t, "Database record not found", response["error"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mock
			mockRepo.ExpectedCalls = nil
			mockRepo.Calls = nil
			tt.mockSetup()

			req := httptest.NewRequest(http.MethodGet, "/api-keys/owner/"+tt.owner, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.checkResponse != nil {
				var response map[string]interface{}
				// Try to unmarshal as map first, if it fails it might be an array
				if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
					// If it's an array response for success case
					if tt.expectedStatus == http.StatusOK {
						var arrayResponse []map[string]interface{}
						err = json.Unmarshal(w.Body.Bytes(), &arrayResponse)
						assert.NoError(t, err)
						assert.NotEmpty(t, arrayResponse)
						return
					}
				}
				tt.checkResponse(t, response)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}
