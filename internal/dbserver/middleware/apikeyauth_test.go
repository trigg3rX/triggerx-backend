package middleware

// import (
// 	"encoding/json"
// 	"net/http"
// 	"net/http/httptest"
// 	"testing"
// 	"time"

// 	"github.com/gin-gonic/gin"
// 	"github.com/gocql/gocql"
// 	"github.com/stretchr/testify/assert"
// 	"github.com/stretchr/testify/mock"
// 	"github.com/trigg3rX/triggerx-backend/pkg/logging"
// 	"github.com/trigg3rX/triggerx-backend/pkg/types"
// )

// // MockLogger is a mock implementation of the logging.Logger interface
// type MockLogger struct {
// 	mock.Mock
// }

// func (m *MockLogger) Debug(msg string, tags ...any) {
// 	m.Called(msg, tags)
// }

// func (m *MockLogger) Info(msg string, tags ...any) {
// 	m.Called(msg, tags)
// }

// func (m *MockLogger) Warn(msg string, tags ...any) {
// 	m.Called(msg, tags)
// }

// func (m *MockLogger) Error(msg string, tags ...any) {
// 	m.Called(msg, tags)
// }

// func (m *MockLogger) Fatal(msg string, tags ...any) {
// 	m.Called(msg, tags)
// }

// func (m *MockLogger) Debugf(template string, args ...interface{}) {
// 	m.Called(template, args)
// }

// func (m *MockLogger) Infof(template string, args ...interface{}) {
// 	m.Called(template, args)
// }

// func (m *MockLogger) Warnf(template string, args ...interface{}) {
// 	m.Called(template, args)
// }

// func (m *MockLogger) Errorf(template string, args ...interface{}) {
// 	m.Called(template, args)
// }

// func (m *MockLogger) Fatalf(template string, args ...interface{}) {
// 	m.Called(template, args)
// }

// func (m *MockLogger) With(tags ...any) logging.Logger {
// 	args := m.Called(tags)
// 	return args.Get(0).(logging.Logger)
// }

// // MockSession is a mock implementation of the database session
// type MockSession struct {
// 	mock.Mock
// }

// func (m *MockSession) Query(query string, values ...interface{}) *gocql.Query {
// 	args := m.Called(query, values)
// 	return args.Get(0).(*gocql.Query)
// }

// func (m *MockSession) ExecuteBatch(batch *gocql.Batch) error {
// 	args := m.Called(batch)
// 	return args.Error(0)
// }

// func (m *MockSession) Close() {
// 	m.Called()
// }

// // MockQuery is a mock implementation of the database query
// type MockQuery struct {
// 	mock.Mock
// }

// func (m *MockQuery) Scan(dest ...interface{}) error {
// 	args := m.Called(dest)
// 	return args.Error(0)
// }

// func (m *MockQuery) Exec() error {
// 	args := m.Called()
// 	return args.Error(0)
// }

// func TestApiKeyAuth_Middleware(t *testing.T) {
// 	// Setup
// 	gin.SetMode(gin.TestMode)

// 	// Create mock objects
// 	mockLogger := new(MockLogger)
// 	mockSession := new(MockSession)
// 	mockQuery := new(MockQuery)

// 	// Create test cases
// 	tests := []struct {
// 		name           string
// 		apiKey         string
// 		mockSetup      func()
// 		expectedStatus int
// 		expectedBody   map[string]interface{}
// 	}{
// 		{
// 			name:   "Missing API Key",
// 			apiKey: "",
// 			mockSetup: func() {
// 				// No mock setup needed
// 			},
// 			expectedStatus: http.StatusUnauthorized,
// 			expectedBody: map[string]interface{}{
// 				"error": "API key is required",
// 			},
// 		},
// 		{
// 			name:   "Valid API Key",
// 			apiKey: "valid-key",
// 			mockSetup: func() {
// 				mockSession.On("Query", mock.Anything, []interface{}{"valid-key", true}).Return(mockQuery)
// 				mockQuery.On("Scan", mock.Anything).Run(func(args mock.Arguments) {
// 					apiKey := args.Get(0).(*types.ApiKey)
// 					apiKey.Key = "valid-key"
// 					apiKey.Owner = "test-owner"
// 					apiKey.IsActive = true
// 					apiKey.RateLimit = 100
// 					apiKey.LastUsed = time.Now()
// 					apiKey.CreatedAt = time.Now()
// 				}).Return(nil)
// 				mockSession.On("Query", mock.Anything, mock.Anything, mock.Anything).Return(mockQuery)
// 				mockQuery.On("Exec").Return(nil)
// 			},
// 			expectedStatus: http.StatusOK,
// 			expectedBody:   nil,
// 		},
// 		{
// 			name:   "Inactive API Key",
// 			apiKey: "inactive-key",
// 			mockSetup: func() {
// 				mockSession.On("Query", mock.Anything, []interface{}{"inactive-key", true}).Return(mockQuery)
// 				mockQuery.On("Scan", mock.Anything).Run(func(args mock.Arguments) {
// 					apiKey := args.Get(0).(*types.ApiKey)
// 					apiKey.IsActive = false
// 				}).Return(nil)
// 			},
// 			expectedStatus: http.StatusForbidden,
// 			expectedBody: map[string]interface{}{
// 				"error": "API key is inactive",
// 			},
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			// Setup test
// 			tt.mockSetup()

// 			// Create router and middleware
// 			router := gin.New()
// 			auth := NewApiKeyAuth(mockSession, nil, mockLogger)
// 			router.Use(auth.GinMiddleware())

// 			// Add test endpoint
// 			router.GET("/test", func(c *gin.Context) {
// 				c.Status(http.StatusOK)
// 			})

// 			// Create test request
// 			w := httptest.NewRecorder()
// 			req, _ := http.NewRequest("GET", "/test", nil)
// 			if tt.apiKey != "" {
// 				req.Header.Set("X-Api-Key", tt.apiKey)
// 			}

// 			// Perform request
// 			router.ServeHTTP(w, req)

// 			// Assertions
// 			assert.Equal(t, tt.expectedStatus, w.Code)
// 			if tt.expectedBody != nil {
// 				var response map[string]interface{}
// 				err := json.Unmarshal(w.Body.Bytes(), &response)
// 				assert.NoError(t, err)
// 				assert.Equal(t, tt.expectedBody, response)
// 			}
// 		})
// 	}
// }

// func TestApiKeyAuth_GetApiKey(t *testing.T) {
// 	// Setup
// 	mockLogger := new(MockLogger)
// 	mockSession := new(MockSession)
// 	mockQuery := new(MockQuery)

// 	tests := []struct {
// 		name      string
// 		key       string
// 		mockSetup func()
// 		wantErr   bool
// 	}{
// 		{
// 			name: "Valid API Key",
// 			key:  "valid-key",
// 			mockSetup: func() {
// 				mockSession.On("Query", mock.Anything, []interface{}{"valid-key", true}).Return(mockQuery)
// 				mockQuery.On("Scan", mock.Anything).Run(func(args mock.Arguments) {
// 					apiKey := args.Get(0).(*types.ApiKey)
// 					apiKey.Key = "valid-key"
// 					apiKey.Owner = "test-owner"
// 					apiKey.IsActive = true
// 					apiKey.RateLimit = 100
// 					apiKey.LastUsed = time.Now()
// 					apiKey.CreatedAt = time.Now()
// 				}).Return(nil)
// 			},
// 			wantErr: false,
// 		},
// 		{
// 			name: "Database Error",
// 			key:  "error-key",
// 			mockSetup: func() {
// 				mockSession.On("Query", mock.Anything, []interface{}{"error-key", true}).Return(mockQuery)
// 				mockQuery.On("Scan", mock.Anything).Return(assert.AnError)
// 			},
// 			wantErr: true,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			tt.mockSetup()

// 			auth := NewApiKeyAuth(mockSession, nil, mockLogger)
// 			apiKey, err := auth.getApiKey(tt.key)

// 			if tt.wantErr {
// 				assert.Error(t, err)
// 				assert.Nil(t, apiKey)
// 			} else {
// 				assert.NoError(t, err)
// 				assert.NotNil(t, apiKey)
// 				assert.Equal(t, tt.key, apiKey.Key)
// 			}
// 		})
// 	}
// }

// func TestApiKeyAuth_UpdateLastUsed(t *testing.T) {
// 	// Setup
// 	mockLogger := new(MockLogger)
// 	mockSession := new(MockSession)
// 	mockQuery := new(MockQuery)

// 	tests := []struct {
// 		name      string
// 		key       string
// 		mockSetup func()
// 		wantErr   bool
// 	}{
// 		{
// 			name: "Successful Update",
// 			key:  "valid-key",
// 			mockSetup: func() {
// 				mockSession.On("Query", mock.Anything, mock.Anything, mock.Anything).Return(mockQuery)
// 				mockQuery.On("Exec").Return(nil)
// 			},
// 			wantErr: false,
// 		},
// 		{
// 			name: "Update Error",
// 			key:  "error-key",
// 			mockSetup: func() {
// 				mockSession.On("Query", mock.Anything, mock.Anything, mock.Anything).Return(mockQuery)
// 				mockQuery.On("Exec").Return(assert.AnError)
// 			},
// 			wantErr: true,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			tt.mockSetup()

// 			auth := NewApiKeyAuth(mockSession, nil, mockLogger)
// 			auth.updateLastUsed(tt.key)

// 			// Since updateLastUsed is a goroutine, we can't directly test the error
// 			// We can only verify that the mock was called as expected
// 			mockSession.AssertExpectations(t)
// 			mockQuery.AssertExpectations(t)
// 		})
// 	}
// }
