package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	datastoreMocks "github.com/trigg3rX/triggerx-backend/pkg/datastore/mocks"
	"github.com/trigg3rX/triggerx-backend/pkg/errors"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

func setupHealthCheckTestHandler() (*Handler, *datastoreMocks.MockDatastoreService, *logging.MockLogger) {
	mockDatastore := new(datastoreMocks.MockDatastoreService)
	mockLogger := new(logging.MockLogger)
	mockLogger.SetupDefaultExpectations()

	handler := &Handler{
		logger:    mockLogger,
		datastore: mockDatastore,
	}

	return handler, mockDatastore, mockLogger
}

func TestHealthCheck(t *testing.T) {
	handler, mockDatastore, mockLogger := setupHealthCheckTestHandler()
	defer mockLogger.AssertExpectations(t)
	defer mockDatastore.AssertExpectations(t)

	tests := []struct {
		name           string
		mockSetup      func()
		expectedCode   int
		expectedStatus string
		expectedError  string
		expectDBError  bool
	}{
		{
			name: "Success - Healthy Database",
			mockSetup: func() {
				mockDatastore.On("HealthCheck", mock.Anything).Return(nil).Once()
			},
			expectedCode:   http.StatusOK,
			expectedStatus: "healthy",
			expectedError:  "",
			expectDBError:  false,
		},
		{
			name: "Error - Unhealthy Database",
			mockSetup: func() {
				mockDatastore.On("HealthCheck", mock.Anything).Return(assert.AnError).Once()
			},
			expectedCode:   http.StatusServiceUnavailable,
			expectedStatus: "unhealthy",
			expectedError:  errors.ErrDBOperationFailed,
			expectDBError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mocks
			mockDatastore.ExpectedCalls = nil
			mockDatastore.Calls = nil

			// Setup
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("GET", "/health", nil)

			// Setup mocks
			tt.mockSetup()

			// Execute
			handler.HealthCheck(c)

			// Assert
			assert.Equal(t, tt.expectedCode, w.Code)

			if tt.expectedCode == http.StatusOK {
				var response types.HealthCheckResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)

				assert.Equal(t, tt.expectedStatus, response.Status)
				assert.NotEmpty(t, response.Timestamp)
				assert.Equal(t, "dbserver", response.Service)
				assert.Equal(t, "1.0.0", response.Version)
				assert.Empty(t, response.Error)
			} else {
				var response types.HealthCheckResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedStatus, response.Status)
				assert.NotEmpty(t, response.Error)
			}

			mockDatastore.AssertExpectations(t)
		})
	}
}
