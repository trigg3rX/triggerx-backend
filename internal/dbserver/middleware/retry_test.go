package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

func init() {
	// Initialize logger for all tests
	config := logging.NewDefaultConfig(logging.ManagerProcess)
	config.UseColors = false // Disable colors in tests
	err := logging.InitServiceLogger(config)
	if err != nil {
		panic("failed to initialize logger: " + err.Error())
	}
}

func setupRetryTestRouter() (*gin.Engine, *RetryConfig) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	config := DefaultRetryConfig()
	// Reduce delays for testing
	config.InitialDelay = 10 * time.Millisecond
	config.MaxDelay = 50 * time.Millisecond
	return router, config
}

func TestRetryMiddleware_SuccessfulRequest(t *testing.T) {
	router, config := setupRetryTestRouter()
	attempts := 0

	router.GET("/test", RetryMiddleware(config), func(c *gin.Context) {
		attempts++
		t.Logf("attempt: %d", attempts)
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, 1, attempts, "Should not retry successful request")
}

func TestRetryMiddleware_RetryableFailure(t *testing.T) {
	router, config := setupRetryTestRouter()
	attempts := 0

	router.GET("/test", RetryMiddleware(config), func(c *gin.Context) {
		attempts++
		t.Logf("attempt: %d", attempts)
		if attempts < 2 {
			c.Status(http.StatusInternalServerError)
		} else {
			c.Status(http.StatusOK)
		}
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, 2, attempts, "Should retry failed request until success")
}

func TestRetryMiddleware_MaxRetriesExceeded(t *testing.T) {
	router, config := setupRetryTestRouter()
	attempts := 0

	router.GET("/test", RetryMiddleware(config), func(c *gin.Context) {
		attempts++
		t.Logf("attempt: %d", attempts)
		c.Status(http.StatusInternalServerError)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Equal(t, config.MaxRetries+1, attempts, "Should retry up to MaxRetries times")
}

func TestRetryMiddleware_NonIdempotentMethod(t *testing.T) {
	router, config := setupRetryTestRouter()
	attempts := 0

	router.POST("/test", RetryMiddleware(config), func(c *gin.Context) {
		attempts++
		t.Logf("attempt: %d", attempts)
		c.Status(http.StatusInternalServerError)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Equal(t, 1, attempts, "Should not retry non-idempotent methods")
}

func TestRetryMiddleware_RequestBodyPreservation(t *testing.T) {
	router, config := setupRetryTestRouter()
	attempts := 0
	expectedBody := "test body"

	router.GET("/test", RetryMiddleware(config), func(c *gin.Context) {
		attempts++
		t.Logf("attempt: %d", attempts)
		body, _ := c.GetRawData()
		assert.Equal(t, expectedBody, string(body), "Request body should be preserved across retries")

		if attempts < 2 {
			c.Status(http.StatusInternalServerError)
		} else {
			c.Status(http.StatusOK)
		}
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", strings.NewReader(expectedBody))
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, 2, attempts, "Should retry with preserved request body")
}

func TestRetryMiddleware_CustomConfig(t *testing.T) {
	router, _ := setupRetryTestRouter()
	attempts := 0

	customConfig := &RetryConfig{
		MaxRetries:      1,
		InitialDelay:    10 * time.Millisecond,
		MaxDelay:        50 * time.Millisecond,
		BackoffFactor:   2.0,
		JitterFactor:    0.1,
		LogRetryAttempt: true,
		RetryStatusCodes: []int{
			http.StatusBadGateway,
		},
	}

	router.GET("/test", RetryMiddleware(customConfig), func(c *gin.Context) {
		attempts++
		t.Logf("attempt: %d", attempts)
		if attempts < 2 {
			c.Status(http.StatusBadGateway)
		} else {
			c.Status(http.StatusOK)
		}
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, 2, attempts, "Should retry with custom config")
}
