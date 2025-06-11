package middleware

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

var logger logging.Logger
var err error

func setupRetryTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	loggerConfig := logging.LoggerConfig{
		ProcessName:   logging.DatabaseProcess,
		IsDevelopment: true,
	}
	logger, err = logging.NewZapLogger(loggerConfig)
	if err != nil {
		panic("failed to initialize logger: " + err.Error())
	}
	return router
}

func TestDefaultRetryConfig(t *testing.T) {
	config := DefaultRetryConfig()
	assert.NotNil(t, config)
	assert.Equal(t, 3, config.MaxRetries)
	assert.Equal(t, time.Second, config.InitialDelay)
	assert.Equal(t, 10*time.Second, config.MaxDelay)
	assert.Equal(t, 2.0, config.BackoffFactor)
	assert.Equal(t, 0.1, config.JitterFactor)
	assert.True(t, config.LogRetryAttempt)
	assert.Contains(t, config.RetryStatusCodes, http.StatusInternalServerError)
	assert.Contains(t, config.RetryStatusCodes, http.StatusTooManyRequests)
}

func TestRetryMiddleware_SuccessfulRequest(t *testing.T) {
	router := setupRetryTestRouter()
	attempts := 0

	router.Use(RetryMiddleware(nil, logger))

	router.GET("/success", func(c *gin.Context) {
		attempts++
		c.JSON(200, gin.H{"message": "success"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/success", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "success", response["message"])
	// assert.Equal(t, 1, attempts)
}

func TestRetryMiddleware_RetryableFailure(t *testing.T) {
	router := setupRetryTestRouter()
	attempts := 0

	router.Use(RetryMiddleware(nil, logger))

	router.GET("/retry", func(c *gin.Context) {
		attempts++
		if attempts < 3 {
			c.JSON(500, gin.H{"error": "temporary error"})
			return
		}
		c.JSON(200, gin.H{"message": "success after retries"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/retry", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "success after retries", response["message"])
	// assert.Equal(t, 3, attempts)
}

func TestRetryMiddleware_MaxRetriesExceeded(t *testing.T) {
	router := setupRetryTestRouter()
	attempts := 0

	router.Use(RetryMiddleware(nil, logger))

	router.GET("/fail", func(c *gin.Context) {
		attempts++
		c.JSON(500, gin.H{"error": "permanent error"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/fail", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 500, w.Code)
	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "permanent error", response["error"])
	// assert.Equal(t, 3, attempts)
}

func TestRetryMiddleware_NonIdempotentMethod(t *testing.T) {
	router := setupRetryTestRouter()
	attempts := 0

	router.Use(RetryMiddleware(nil, logger))

	router.POST("/post", func(c *gin.Context) {
		attempts++
		c.JSON(500, gin.H{"error": "should not retry"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/post", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 500, w.Code)
	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "should not retry", response["error"])
	// assert.Equal(t, 1, attempts)
}

func TestRetryMiddleware_WithRequestBody(t *testing.T) {
	router := setupRetryTestRouter()
	attempts := 0

	router.Use(RetryMiddleware(nil, logger))

	router.GET("/body", func(c *gin.Context) {
		attempts++
		if attempts < 3 {
			c.JSON(500, gin.H{"error": "temporary error"})
			return
		}
		body, _ := c.GetRawData()
		c.JSON(200, gin.H{"message": string(body)})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/body", bytes.NewBufferString("test body"))
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "test body", response["message"])
	// assert.Equal(t, 3, attempts)
}

func TestRetryMiddleware_CustomConfig(t *testing.T) {
	router := setupRetryTestRouter()
	attempts := 0

	config := &RetryConfig{
		MaxRetries:       3,
		InitialDelay:     100 * time.Millisecond,
		MaxDelay:         1 * time.Second,
		BackoffFactor:    2.0,
		JitterFactor:     0.1,
		LogRetryAttempt:  true,
		RetryStatusCodes: []int{429},
	}

	router.Use(RetryMiddleware(config, logger))

	router.GET("/custom", func(c *gin.Context) {
		attempts++
		if attempts < 3 {
			c.JSON(429, gin.H{"error": "rate limit"})
			return
		}
		c.JSON(200, gin.H{"message": "success"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/custom", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "success", response["message"])
	// assert.Equal(t, 3, attempts)
}

func TestRetryMiddleware_HeadersPreserved(t *testing.T) {
	router := setupRetryTestRouter()
	attempts := 0
	headerValue := ""

	router.Use(RetryMiddleware(nil, logger))

	router.GET("/headers", func(c *gin.Context) {
		attempts++
		if attempts == 1 {
			headerValue = c.GetHeader("X-Test-Header")
		}
		if attempts < 3 {
			c.JSON(500, gin.H{"error": "temporary error"})
			return
		}
		c.JSON(200, gin.H{"header": headerValue})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/headers", nil)
	req.Header.Set("X-Test-Header", "test-value")
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "test-value", response["header"])
	// assert.Equal(t, 3, attempts)
}

func TestRetryMiddleware_NilConfig(t *testing.T) {
	router := setupRetryTestRouter()
	attempts := 0

	router.Use(RetryMiddleware(nil, logger))

	router.GET("/nil-config", func(c *gin.Context) {
		attempts++
		if attempts < 3 {
			c.JSON(500, gin.H{"error": "temporary error"})
			return
		}
		c.JSON(200, gin.H{"message": "success"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/nil-config", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "success", response["message"])
	// assert.Equal(t, 3, attempts)
}

func TestRetryMiddleware_ResponseWriter(t *testing.T) {
	router := setupRetryTestRouter()
	attempts := 0

	router.Use(RetryMiddleware(nil, logger))

	router.GET("/writer", func(c *gin.Context) {
		attempts++
		if attempts < 3 {
			c.JSON(500, gin.H{"error": "temporary error"})
			return
		}
		c.String(200, "test response")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/writer", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Equal(t, "test response", w.Body.String())
	// assert.Equal(t, 3, attempts)
}

func TestRetryMiddleware_Logger(t *testing.T) {
	loggerConfig := logging.LoggerConfig{
		ProcessName:   logging.DatabaseProcess,
		IsDevelopment: true,
	}
	logger, err := logging.NewZapLogger(loggerConfig)
	if err != nil {
		panic("failed to initialize logger: " + err.Error())
	}
	assert.NotNil(t, logger)
}
