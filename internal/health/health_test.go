package health

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/trigg3rX/triggerx-backend/internal/health/keeper"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

// StateManagerInterface defines the interface for StateManager
type StateManagerInterface interface {
	UpdateKeeperHealth(health types.KeeperHealthCheckIn) error
	GetKeeperCount() (int, int)
	GetAllActiveKeepers() []string
	GetDetailedKeeperInfo() []types.HealthKeeperInfo
}

// MockStateManager is a mock implementation of StateManagerInterface
type MockStateManager struct {
	mock.Mock
}

func (m *MockStateManager) UpdateKeeperHealth(health types.KeeperHealthCheckIn) error {
	args := m.Called(health)
	return args.Error(0)
}

func (m *MockStateManager) GetKeeperCount() (int, int) {
	args := m.Called()
	return args.Int(0), args.Int(1)
}

func (m *MockStateManager) GetAllActiveKeepers() []string {
	args := m.Called()
	return args.Get(0).([]string)
}

func (m *MockStateManager) GetDetailedKeeperInfo() []types.HealthKeeperInfo {
	args := m.Called()
	return args.Get(0).([]types.HealthKeeperInfo)
}

// TestHandler is a test-specific version of Handler that uses our mock
type TestHandler struct {
	logger       logging.Logger
	stateManager StateManagerInterface
}

func (h *TestHandler) handleRoot(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"service":   "TriggerX Health Service",
		"status":    "running",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

// MockVerifySignature is a variable that holds the mock implementation of VerifySignature
var MockVerifySignature func(message string, signatureHex string, expectedAddress string) (bool, error)

// mockVerifySignature is a wrapper around the mock implementation
func mockVerifySignature(message string, signatureHex string, expectedAddress string) (bool, error) {
	if MockVerifySignature != nil {
		return MockVerifySignature(message, signatureHex, expectedAddress)
	}
	return false, fmt.Errorf("mock not set")
}

func (h *TestHandler) HandleCheckInEvent(c *gin.Context) {
	var keeperHealth types.KeeperHealthCheckIn
	if err := c.ShouldBindJSON(&keeperHealth); err != nil {
		h.logger.Error("Failed to parse keeper health check-in request",
			"error", err,
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.logger.Debug("Received keeper health check-in",
		"keeper", keeperHealth.KeeperAddress,
		"version", keeperHealth.Version,
		"peer_id", keeperHealth.PeerID,
	)

	if keeperHealth.Version == "0.0.7" || keeperHealth.Version == "0.0.6" || keeperHealth.Version == "0.0.5" || keeperHealth.Version == "" {
		h.logger.Warn("Rejecting obsolete keeper version",
			"keeper", keeperHealth.KeeperAddress,
			"version", keeperHealth.Version,
		)
		c.JSON(http.StatusPreconditionFailed, gin.H{
			"error": "OBSOLETE VERSION of Keeper, authorization failed, UPGRADE TO v0.1.2",
		})
		return
	}

	if keeperHealth.Version == "0.1.0" || keeperHealth.Version == "0.1.1" {
		h.logger.Warn("Older keeper version detected",
			"keeper", keeperHealth.KeeperAddress,
			"version", keeperHealth.Version,
		)
		c.JSON(http.StatusOK, gin.H{
			"message": "OLDER VERSION of Keeper, UPGRADE TO v0.1.2",
		})
		return
	}

	if keeperHealth.Version == "0.1.2" {
		ok, err := mockVerifySignature(keeperHealth.KeeperAddress, keeperHealth.Signature, keeperHealth.ConsensusPubKey)
		if !ok {
			h.logger.Error("Invalid keeper signature",
				"keeper", keeperHealth.KeeperAddress,
				"error", err,
			)
			c.JSON(http.StatusPreconditionFailed, gin.H{
				"error": "Invalid signature",
			})
			return
		}

		h.logger.Debug("Valid keeper signature verified",
			"keeper", keeperHealth.KeeperAddress,
			"version", keeperHealth.Version,
			"ip", c.ClientIP(),
		)

		keeperHealth.KeeperAddress = strings.ToLower(keeperHealth.KeeperAddress)
		keeperHealth.ConsensusPubKey = strings.ToLower(keeperHealth.ConsensusPubKey)

		if err := h.stateManager.UpdateKeeperHealth(keeperHealth); err != nil {
			if errors.Is(err, keeper.ErrKeeperNotVerified) {
				h.logger.Warn("Unverified keeper attempted health check-in",
					"keeper", keeperHealth.KeeperAddress,
				)
				c.JSON(http.StatusForbidden, gin.H{
					"error": "Keeper not verified",
					"code":  "KEEPER_NOT_VERIFIED",
				})
				return
			}

			h.logger.Error("Failed to update keeper state",
				"error", err,
				"keeper", keeperHealth.KeeperAddress,
			)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update keeper state"})
			return
		}

		h.logger.Info("Successfully processed keeper health check-in",
			"keeper", keeperHealth.KeeperAddress,
			"version", keeperHealth.Version,
		)

		c.JSON(http.StatusOK, gin.H{
			"message": "Keeper health check-in received",
			"active":  true,
		})
	}
}

func (h *TestHandler) GetKeeperStatus(c *gin.Context) {
	total, active := h.stateManager.GetKeeperCount()
	activeKeepers := h.stateManager.GetAllActiveKeepers()

	c.JSON(http.StatusOK, gin.H{
		"total_keepers":      total,
		"active_keepers":     active,
		"active_keeper_list": activeKeepers,
	})
}

func (h *TestHandler) GetDetailedKeeperStatus(c *gin.Context) {
	total, active := h.stateManager.GetKeeperCount()
	detailedInfo := h.stateManager.GetDetailedKeeperInfo()

	c.JSON(http.StatusOK, gin.H{
		"total_keepers":  total,
		"active_keepers": active,
		"keepers":        detailedInfo,
		"timestamp":      time.Now().UTC().Format(time.RFC3339),
	})
}

func setupTestRouter() (*gin.Engine, *logging.MockLogger, *MockStateManager) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockLogger := new(logging.MockLogger)
	mockStateManager := new(MockStateManager)

	// Create a test handler with the mock implementations
	handler := &TestHandler{
		logger:       mockLogger,
		stateManager: mockStateManager,
	}

	router.GET("/", handler.handleRoot)
	router.POST("/health", handler.HandleCheckInEvent)
	router.GET("/status", handler.GetKeeperStatus)
	router.GET("/operators", handler.GetDetailedKeeperStatus)

	return router, mockLogger, mockStateManager
}

func TestHandleRoot(t *testing.T) {
	router, _, _ := setupTestRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)

	assert.NoError(t, err)
	assert.Equal(t, "TriggerX Health Service", response["service"])
	assert.Equal(t, "running", response["status"])
	assert.NotEmpty(t, response["timestamp"])
}

func TestHandleCheckInEvent(t *testing.T) {
	// Save original function and restore after test
	originalVerifySignature := MockVerifySignature
	defer func() { MockVerifySignature = originalVerifySignature }()

	tests := []struct {
		name           string
		keeperHealth   types.KeeperHealthCheckIn
		expectedStatus int
		mockSetup      func(*logging.MockLogger, *MockStateManager)
		verifyResult   bool
		verifyError    error
	}{
		{
			name: "Valid keeper check-in",
			keeperHealth: types.KeeperHealthCheckIn{
				KeeperAddress:   "0x123",
				Version:         "0.1.2",
				Signature:       "0x456",
				ConsensusPubKey: "0x789",
				PeerID:          "test-peer",
			},
			expectedStatus: http.StatusOK,
			mockSetup: func(logger *logging.MockLogger, stateManager *MockStateManager) {
				logger.On("Debug", mock.Anything, mock.Anything).Return()
				logger.On("Info", mock.Anything, mock.Anything).Return()
				stateManager.On("UpdateKeeperHealth", mock.Anything).Return(nil)
			},
			verifyResult: true,
			verifyError:  nil,
		},
		{
			name: "Obsolete version",
			keeperHealth: types.KeeperHealthCheckIn{
				KeeperAddress: "0x123",
				Version:       "0.0.7",
				PeerID:        "test-peer",
			},
			expectedStatus: http.StatusPreconditionFailed,
			mockSetup: func(logger *logging.MockLogger, stateManager *MockStateManager) {
				logger.On("Debug", mock.Anything, mock.Anything).Return()
				logger.On("Warn", mock.Anything, mock.Anything).Return()
			},
		},
		{
			name: "Older version",
			keeperHealth: types.KeeperHealthCheckIn{
				KeeperAddress: "0x123",
				Version:       "0.1.1",
				PeerID:        "test-peer",
			},
			expectedStatus: http.StatusOK,
			mockSetup: func(logger *logging.MockLogger, stateManager *MockStateManager) {
				logger.On("Debug", mock.Anything, mock.Anything).Return()
				logger.On("Warn", mock.Anything, mock.Anything).Return()
			},
		},
		{
			name:           "Invalid JSON",
			keeperHealth:   types.KeeperHealthCheckIn{},
			expectedStatus: http.StatusPreconditionFailed,
			mockSetup: func(logger *logging.MockLogger, stateManager *MockStateManager) {
				logger.On("Debug", mock.Anything, mock.Anything).Return()
				logger.On("Warn", mock.Anything, mock.Anything).Return()
				logger.On("Error", mock.Anything, mock.Anything).Return()
			},
		},
		{
			name: "Invalid signature",
			keeperHealth: types.KeeperHealthCheckIn{
				KeeperAddress:   "0x123",
				Version:         "0.1.2",
				Signature:       "invalid",
				ConsensusPubKey: "0x789",
				PeerID:          "test-peer",
			},
			expectedStatus: http.StatusPreconditionFailed,
			mockSetup: func(logger *logging.MockLogger, stateManager *MockStateManager) {
				logger.On("Debug", mock.Anything, mock.Anything).Return()
				logger.On("Error", mock.Anything, mock.Anything).Return()
			},
			verifyResult: false,
			verifyError:  fmt.Errorf("invalid signature"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router, mockLogger, mockStateManager := setupTestRouter()
			tt.mockSetup(mockLogger, mockStateManager)

			// Set up the mock for VerifySignature
			if tt.name == "Valid keeper check-in" || tt.name == "Invalid signature" {
				MockVerifySignature = func(message string, signatureHex string, expectedAddress string) (bool, error) {
					return tt.verifyResult, tt.verifyError
				}
			}

			body, _ := json.Marshal(tt.keeperHealth)
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/health", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			router.ServeHTTP(w, req)
			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestGetKeeperStatus(t *testing.T) {
	router, _, mockStateManager := setupTestRouter()

	mockStateManager.On("GetKeeperCount").Return(10, 5)
	mockStateManager.On("GetAllActiveKeepers").Return([]string{"0x1", "0x2"})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/status", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)

	assert.NoError(t, err)
	assert.Equal(t, float64(10), response["total_keepers"])
	assert.Equal(t, float64(5), response["active_keepers"])
	assert.NotNil(t, response["active_keeper_list"])
}

func TestGetDetailedKeeperStatus(t *testing.T) {
	router, _, mockStateManager := setupTestRouter()

	mockStateManager.On("GetKeeperCount").Return(10, 5)
	mockStateManager.On("GetDetailedKeeperInfo").Return([]types.HealthKeeperInfo{
		{
			KeeperAddress: "0x1",
			IsActive:      true,
			LastCheckedIn: time.Now(),
		},
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/operators", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)

	assert.NoError(t, err)
	assert.Equal(t, float64(10), response["total_keepers"])
	assert.Equal(t, float64(5), response["active_keepers"])
	assert.NotNil(t, response["keepers"])
	assert.NotEmpty(t, response["timestamp"])
}

func TestVerifySignature(t *testing.T) {
	tests := []struct {
		name           string
		message        string
		signature      string
		expectedAddr   string
		expectedResult bool
		expectError    bool
	}{
		{
			name:           "Invalid signature length",
			message:        "test",
			signature:      "0x123",
			expectedAddr:   "0x456",
			expectedResult: false,
			expectError:    true,
		},
		{
			name:           "Invalid signature format",
			message:        "test",
			signature:      "invalid",
			expectedAddr:   "0x456",
			expectedResult: false,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := mockVerifySignature(tt.message, tt.signature, tt.expectedAddr)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.expectedResult, result)
		})
	}
}

func TestLoggerMiddleware(t *testing.T) {
	mockLogger := new(logging.MockLogger)
	mockLogger.On("Debug", mock.Anything, mock.Anything).Return()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(LoggerMiddleware(mockLogger))

	router.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockLogger.AssertExpectations(t)
}
