package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"
	"unsafe"

	"github.com/gin-gonic/gin"
	"github.com/gocql/gocql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/trigg3rX/triggerx-backend/pkg/database"
)

type MockConnection struct {
	mock.Mock
	session database.Sessioner
}

func (m *MockConnection) Session() database.Sessioner {
	return m.session
}

func (m *MockConnection) Close() {
	m.Called()
}

type MockSession struct {
	mock.Mock
	scanFunc func(dest ...interface{}) error
}

func (m *MockSession) Query(query string, args ...interface{}) *gocql.Query {
	m.Called(query, args)
	return &gocql.Query{}
}

func (m *MockSession) ExecuteBatch(batch *gocql.Batch) error {
	args := m.Called(batch)
	return args.Error(0)
}

func (m *MockSession) Close() {
	m.Called()
}

type MockHTTPClient struct {
	mock.Mock
}

func (m *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	args := m.Called(req)
	return args.Get(0).(*http.Response), args.Error(1)
}

func (m *MockHTTPClient) DoWithRetry(req *http.Request) (*http.Response, error) {
	args := m.Called(req)
	return args.Get(0).(*http.Response), args.Error(1)
}

func setupTestHandlerForHealthCheck() (*Handler, *MockSession) {
	mockSession := new(MockSession)
	conn := &database.Connection{}
	// Use reflection to set the unexported session field
	connValue := reflect.ValueOf(conn).Elem()
	sessionField := connValue.FieldByName("session")
	sessionField = reflect.NewAt(sessionField.Type(), unsafe.Pointer(sessionField.UnsafeAddr())).Elem()
	sessionField.Set(reflect.ValueOf(mockSession))

	handler := &Handler{
		db:     conn,
		logger: &MockLogger{},
		config: NotificationConfig{
			EmailFrom:     "test@example.com",
			EmailPassword: "password",
			BotToken:      "token",
		},
	}
	handler.scanNowQuery = func(ts *time.Time) error {
		return mockSession.scanFunc(ts)
	}
	return handler, mockSession
}

func TestHealthCheck(t *testing.T) {
	handler, mockSession := setupTestHandlerForHealthCheck()

	tests := []struct {
		name           string
		setupMocks     func()
		expectedCode   int
		expectedStatus string
		expectedDB     map[string]interface{}
	}{
		{
			name: "Success - Healthy Database",
			setupMocks: func() {
				mockSession.scanFunc = func(dest ...interface{}) error {
					arg := dest[0].(*time.Time)
					*arg = time.Now()
					return nil
				}
				mockSession.On("Query", "SELECT now() FROM system.local", mock.Anything).Return(&gocql.Query{})
			},
			expectedCode:   http.StatusOK,
			expectedStatus: "ok",
			expectedDB: map[string]interface{}{
				"status": "healthy",
				"error":  "",
			},
		},
		{
			name: "Error - Unhealthy Database",
			setupMocks: func() {
				mockSession.scanFunc = func(dest ...interface{}) error {
					return assert.AnError
				}
				mockSession.On("Query", "SELECT now() FROM system.local", mock.Anything).Return(&gocql.Query{})
			},
			expectedCode:   http.StatusServiceUnavailable,
			expectedStatus: "degraded",
			expectedDB: map[string]interface{}{
				"status": "unhealthy",
				"error":  assert.AnError.Error(),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("GET", "/health", nil)

			// Setup mocks
			tt.setupMocks()

			// Execute
			handler.HealthCheck(c)

			// Assert
			assert.Equal(t, tt.expectedCode, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			assert.Equal(t, tt.expectedStatus, response["status"])
			assert.NotEmpty(t, response["timestamp"])
			assert.Equal(t, "dbserver", response["service"])
			assert.Equal(t, "1.0.0", response["version"])
			assert.NotEmpty(t, response["uptime"])

			db, ok := response["database"].(map[string]interface{})
			assert.True(t, ok)
			assert.Equal(t, tt.expectedDB["status"], db["status"])
			assert.Equal(t, tt.expectedDB["error"], db["error"])

			checks, ok := response["checks"].(map[string]interface{})
			assert.True(t, ok)
			assert.Equal(t, tt.expectedDB["status"] == "healthy", checks["database_connection"])
		})
	}
}

func setupTestHandlerForScheduler() (*Handler, *MockHTTPClient) {
	mockSession := new(MockSession)
	conn := &database.Connection{}
	// Use reflection to set the unexported session field
	connValue := reflect.ValueOf(conn).Elem()
	sessionField := connValue.FieldByName("session")
	sessionField = reflect.NewAt(sessionField.Type(), unsafe.Pointer(sessionField.UnsafeAddr())).Elem()
	sessionField.Set(reflect.ValueOf(mockSession))

	mockHTTPClient := new(MockHTTPClient)

	handler := &Handler{
		db:     conn,
		logger: &MockLogger{},
		config: NotificationConfig{
			EmailFrom:     "test@example.com",
			EmailPassword: "password",
			BotToken:      "token",
		},
	}
	handler.scanNowQuery = func(ts *time.Time) error {
		return mockSession.scanFunc(ts)
	}

	return handler, mockHTTPClient
}

func TestSendDataToScheduler(t *testing.T) {
	handler, mockHTTPClient := setupTestHandlerForScheduler()

	// Create a test server to handle the requests
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	tests := []struct {
		name           string
		apiURL         string
		data           interface{}
		schedulerName  string
		expectedResult bool
		expectedError  string
		setupMock      func()
	}{
		{
			name:           "Success - Event Scheduler",
			apiURL:         server.URL + "/test",
			data:           map[string]string{"key": "value"},
			schedulerName:  "event scheduler",
			expectedResult: true,
			setupMock: func() {
				mockHTTPClient.On("DoWithRetry", mock.Anything).Return(&http.Response{
					StatusCode: http.StatusOK,
					Body:       http.NoBody,
				}, nil)
			},
		},
		{
			name:           "Success - Condition Scheduler",
			apiURL:         server.URL + "/test",
			data:           map[string]string{"key": "value"},
			schedulerName:  "condition scheduler",
			expectedResult: true,
			setupMock: func() {
				mockHTTPClient.On("DoWithRetry", mock.Anything).Return(&http.Response{
					StatusCode: http.StatusOK,
					Body:       http.NoBody,
				}, nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock
			tt.setupMock()

			// Execute
			result, err := handler.sendDataToScheduler(tt.apiURL, tt.data, tt.schedulerName)

			// Assert
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult, result)
			}
		})
	}
}

func TestSendPauseToScheduler(t *testing.T) {
	handler, mockHTTPClient := setupTestHandlerForScheduler()

	// Create a test server to handle the requests
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	tests := []struct {
		name           string
		apiURL         string
		schedulerName  string
		expectedResult bool
		expectedError  string
		setupMock      func()
	}{
		{
			name:           "Success - Event Scheduler",
			apiURL:         server.URL + "/test",
			schedulerName:  "event scheduler",
			expectedResult: true,
			setupMock: func() {
				mockHTTPClient.On("DoWithRetry", mock.Anything).Return(&http.Response{
					StatusCode: http.StatusOK,
					Body:       http.NoBody,
				}, nil)
			},
		},
		{
			name:           "Success - Condition Scheduler",
			apiURL:         server.URL + "/test",
			schedulerName:  "condition scheduler",
			expectedResult: true,
			setupMock: func() {
				mockHTTPClient.On("DoWithRetry", mock.Anything).Return(&http.Response{
					StatusCode: http.StatusOK,
					Body:       http.NoBody,
				}, nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock
			tt.setupMock()

			// Execute
			result, err := handler.notifyPauseToEventScheduler(1)

			// Assert
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult, result)
			}
		})
	}
}
