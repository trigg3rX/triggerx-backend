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
	"github.com/trigg3rX/triggerx-backend-imua/internal/dbserver/types"
)

// MockApiKeysRepository is a mock implementation of the API keys repository
type MockApiKeysRepository struct {
	mock.Mock
}

func setupApiKeyTestRouter() (*gin.Engine, *MockApiKeysRepository) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	mockRepo := new(MockApiKeysRepository)
	handler := &Handler{
		logger:            &MockLogger{},
		apiKeysRepository: mockRepo,
	}
	router.POST("/api-keys", handler.CreateApiKey)
	router.PUT("/api-keys/:key", handler.UpdateApiKey)
	router.DELETE("/api-keys/:key", handler.DeleteApiKey)
	return router, mockRepo
}

func (m *MockApiKeysRepository) GetApiKeyDataByOwner(owner string) ([]*types.ApiKeyData, error) {
	args := m.Called(owner)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*types.ApiKeyData), args.Error(1)
}

func (m *MockApiKeysRepository) GetApiKeyDataByKey(key string) (*types.ApiKeyData, error) {
	args := m.Called(key)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.ApiKeyData), args.Error(1)
}

func (m *MockApiKeysRepository) GetApiKeyCounters(key string) (*types.ApiKeyCounters, error) {
	args := m.Called(key)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.ApiKeyCounters), args.Error(1)
}

func (m *MockApiKeysRepository) GetApiKeyByOwner(owner string) (string, error) {
	args := m.Called(owner)
	return args.String(0), args.Error(1)
}

func (m *MockApiKeysRepository) GetApiOwnerByApiKey(key string) (string, error) {
	args := m.Called(key)
	return args.String(0), args.Error(1)
}

func (m *MockApiKeysRepository) CreateApiKey(apiKey *types.ApiKeyData) error {
	args := m.Called(apiKey)
	return args.Error(0)
}

func (m *MockApiKeysRepository) UpdateApiKey(req *types.UpdateApiKeyRequest) error {
	args := m.Called(req)
	return args.Error(0)
}

func (m *MockApiKeysRepository) UpdateApiKeyStatus(req *types.UpdateApiKeyStatusRequest) error {
	args := m.Called(req)
	return args.Error(0)
}

func (m *MockApiKeysRepository) UpdateApiKeyLastUsed(key string, isSuccess bool) error {
	args := m.Called(key, isSuccess)
	return args.Error(0)
}

func (m *MockApiKeysRepository) DeleteApiKey(key string) error {
	args := m.Called(key)
	return args.Error(0)
}

func TestCreateApiKey(t *testing.T) {
	router, mockRepo := setupApiKeyTestRouter()

	tests := []struct {
		name           string
		request        interface{}
		mockSetup      func()
		expectedStatus int
		expectedBody   interface{}
	}{
		{
			name: "Success - Create new API key",
			request: types.CreateApiKeyRequest{
				Owner:     "test-owner",
				RateLimit: 100,
			},
			mockSetup: func() {
				mockRepo.On("GetApiKeyDataByOwner", "test-owner").Return([]*types.ApiKeyData{}, nil)
				mockRepo.On("CreateApiKey", mock.AnythingOfType("*types.ApiKeyData")).Return(nil)
			},
			expectedStatus: http.StatusCreated,
			expectedBody: func() interface{} {
				return map[string]interface{}{
					"owner":      "test-owner",
					"is_active":  true,
					"rate_limit": float64(100),
				}
			},
		},
		{
			name: "Success - Default rate limit",
			request: types.CreateApiKeyRequest{
				Owner: "test-owner",
			},
			mockSetup: func() {
				mockRepo.On("GetApiKeyDataByOwner", "test-owner").Return([]*types.ApiKeyData{}, nil)
				mockRepo.On("CreateApiKey", mock.AnythingOfType("*types.ApiKeyData")).Return(nil)
			},
			expectedStatus: http.StatusCreated,
			expectedBody: func() interface{} {
				return map[string]interface{}{
					"owner":      "test-owner",
					"is_active":  true,
					"rate_limit": float64(60),
				}
			},
		},
		{
			name: "Error - Missing owner",
			request: types.CreateApiKeyRequest{
				RateLimit: 100,
			},
			mockSetup:      func() {},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"error": "Owner is required",
			},
		},
		{
			name:           "Error - Invalid request body",
			request:        "invalid json",
			mockSetup:      func() {},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"error": "Invalid request body",
			},
		},
		{
			name: "Error - API key already exists",
			request: types.CreateApiKeyRequest{
				Owner:     "existing-owner",
				RateLimit: 100,
			},
			mockSetup: func() {
				existingKey := &types.ApiKeyData{
					Key:       "existing-key",
					Owner:     "existing-owner",
					IsActive:  true,
					RateLimit: 100,
					LastUsed:  time.Now().UTC(),
					CreatedAt: time.Now().UTC(),
				}
				mockRepo.On("GetApiKeyDataByOwner", "existing-owner").Return([]*types.ApiKeyData{existingKey}, nil)
			},
			expectedStatus: http.StatusConflict,
			expectedBody: map[string]interface{}{
				"error": "API key already exists for this owner",
			},
		},
		{
			name: "Error - Database error on check",
			request: types.CreateApiKeyRequest{
				Owner:     "test-owner",
				RateLimit: 100,
			},
			mockSetup: func() {
				mockRepo.On("GetApiKeyDataByOwner", "test-owner").Return(nil, assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody: map[string]interface{}{
				"error": "Internal server error",
			},
		},
		{
			name: "Error - Database error on create",
			request: types.CreateApiKeyRequest{
				Owner:     "test-owner",
				RateLimit: 100,
			},
			mockSetup: func() {
				mockRepo.On("GetApiKeyDataByOwner", "test-owner").Return([]*types.ApiKeyData{}, nil)
				mockRepo.On("CreateApiKey", mock.AnythingOfType("*types.ApiKeyData")).Return(assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody: map[string]interface{}{
				"error": "Failed to create API key",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mock
			mockRepo.ExpectedCalls = nil
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
			var response map[string]interface{}
			err = json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			if expectedBody, ok := tt.expectedBody.(func() interface{}); ok {
				expected := expectedBody()
				for k, v := range expected.(map[string]interface{}) {
					assert.Equal(t, v, response[k])
				}
			} else {
				assert.Equal(t, tt.expectedBody, response)
			}
		})
	}
}

func TestUpdateApiKey(t *testing.T) {
	router, mockRepo := setupApiKeyTestRouter()

	tests := []struct {
		name           string
		key            string
		request        interface{}
		mockSetup      func()
		expectedStatus int
		expectedBody   interface{}
	}{
		{
			name: "Success - Update rate limit",
			key:  "test-key",
			request: types.UpdateApiKeyRequest{
				RateLimit: ptr(int(200)),
			},
			mockSetup: func() {
				existingKey := &types.ApiKeyData{
					Key:       "test-key",
					Owner:     "test-owner",
					IsActive:  true,
					RateLimit: 100,
				}
				mockRepo.On("GetApiKeyDataByKey", "test-key").Return(existingKey, nil)
				mockRepo.On("UpdateApiKey", mock.AnythingOfType("*types.UpdateApiKeyRequest")).Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"key":           "test-key",
				"owner":         "test-owner",
				"is_active":     true,
				"rate_limit":    float64(200),
				"created_at":    "0001-01-01T00:00:00Z",
				"last_used":     "0001-01-01T00:00:00Z",
				"success_count": float64(0),
				"failed_count":  float64(0),
			},
		},
		{
			name:           "Error - Invalid request body",
			key:            "test-key",
			request:        "invalid json",
			mockSetup:      func() {},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"error": "Invalid request body",
			},
		},
		{
			name: "Error - API key not found",
			key:  "non-existent-key",
			request: types.UpdateApiKeyRequest{
				RateLimit: ptr(int(200)),
			},
			mockSetup: func() {
				mockRepo.On("GetApiKeyDataByKey", "non-existent-key").Return(nil, assert.AnError)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody: map[string]interface{}{
				"error": "API key not found",
			},
		},
		{
			name: "Error - Database error on update",
			key:  "test-key",
			request: types.UpdateApiKeyRequest{
				RateLimit: ptr(int(200)),
			},
			mockSetup: func() {
				existingKey := &types.ApiKeyData{
					Key:       "test-key",
					Owner:     "test-owner",
					IsActive:  true,
					RateLimit: 100,
				}
				mockRepo.On("GetApiKeyDataByKey", "test-key").Return(existingKey, nil)
				mockRepo.On("UpdateApiKey", mock.AnythingOfType("*types.UpdateApiKeyRequest")).Return(assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody: map[string]interface{}{
				"error": "Failed to update API key",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mock
			mockRepo.ExpectedCalls = nil
			tt.mockSetup()

			var body []byte
			var err error
			if str, ok := tt.request.(string); ok {
				body = []byte(str)
			} else {
				body, err = json.Marshal(tt.request)
				assert.NoError(t, err)
			}

			req := httptest.NewRequest(http.MethodPut, "/api-keys/"+tt.key, bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			var response map[string]interface{}
			err = json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedBody, response)
		})
	}
}

func TestDeleteApiKey(t *testing.T) {
	router, mockRepo := setupApiKeyTestRouter()

	tests := []struct {
		name           string
		key            string
		mockSetup      func()
		expectedStatus int
		expectedBody   interface{}
	}{
		{
			name: "Success - Delete API key",
			key:  "test-key",
			mockSetup: func() {
				mockRepo.On("UpdateApiKeyStatus", &types.UpdateApiKeyStatusRequest{
					Key:      "test-key",
					IsActive: false,
				}).Return(nil)
			},
			expectedStatus: http.StatusNoContent,
			expectedBody:   nil,
		},
		{
			name: "Error - Database error on delete",
			key:  "test-key",
			mockSetup: func() {
				mockRepo.On("UpdateApiKeyStatus", &types.UpdateApiKeyStatusRequest{
					Key:      "test-key",
					IsActive: false,
				}).Return(assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody: map[string]interface{}{
				"error": "Failed to delete API key",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mock
			mockRepo.ExpectedCalls = nil
			tt.mockSetup()

			req := httptest.NewRequest(http.MethodDelete, "/api-keys/"+tt.key, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedBody != nil {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedBody, response)
			} else {
				assert.Empty(t, w.Body.String())
			}
		})
	}
}

// Helper function to create pointer to int
func ptr(i int) *int {
	return &i
}
