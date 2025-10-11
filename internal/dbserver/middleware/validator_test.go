package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

func setupTestRouter() (*gin.Engine, *Validator) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	loggerConfig := logging.LoggerConfig{
		ProcessName:   logging.DatabaseProcess,
		IsDevelopment: true,
	}
	logger, err := logging.NewZapLogger(loggerConfig)
	if err != nil {
		panic("failed to initialize logger: " + err.Error())
	}
	validator := NewValidator(logger)
	return router, validator
}

func TestTimeBasedJobValidation(t *testing.T) {
	router, validator := setupTestRouter()
	router.POST("/api/jobs", validator.GinMiddleware(), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	// Test valid time-based job
	validTimeJob := map[string]interface{}{
		"job_id":                  "test-time-job-1",
		"user_address":            "0x1234567890123456789012345678901234567890",
		"created_chain_id":        "1",
		"timezone":                "UTC",
		"task_definition_id":      1,
		"job_title":               "Test Time Job",
		"time_frame":              3600,
		"recurring":               true,
		"schedule_type":           "interval",
		"time_interval":           1800,
		"job_cost_prediction":     "0.1",
		"target_chain_id":         "1",
		"target_contract_address": "0x1234567890123456789012345678901234567890",
		"target_function":         "execute",
		"abi":                     "[{\"inputs\":[],\"name\":\"execute\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
		"arg_type":                1,
		"arguments":               []string{},
		"job_type":                "sdk",
	}

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/jobs", createJSONBody(t, []interface{}{validTimeJob}))
	router.ServeHTTP(w, req)

	// Print response body for debugging
	body, _ := io.ReadAll(w.Body)
	t.Logf("Time-based job response: %s", string(body))

	assert.Equal(t, http.StatusOK, w.Code)

	// Create a new response recorder for the next request
	w = httptest.NewRecorder()

	// Test invalid time-based job (invalid user address)
	invalidTimeJob := make(map[string]interface{})
	for k, v := range validTimeJob {
		invalidTimeJob[k] = v
	}
	invalidTimeJob["user_address"] = "invalid_address" // Not a valid ethereum address
	req, _ = http.NewRequest("POST", "/api/jobs", createJSONBody(t, []interface{}{invalidTimeJob}))
	router.ServeHTTP(w, req)

	// Print response body for debugging
	body, _ = io.ReadAll(w.Body)
	t.Logf("Invalid time-based job response: %s", string(body))

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestEventBasedJobValidation(t *testing.T) {
	router, validator := setupTestRouter()
	router.POST("/api/jobs", validator.GinMiddleware(), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	// Test valid event-based job
	validEventJob := map[string]interface{}{
		"job_id":                   "test-event-job-1",
		"user_address":             "0x1234567890123456789012345678901234567890",
		"created_chain_id":         "1",
		"timezone":                 "UTC",
		"task_definition_id":       3,
		"job_title":                "Test Event Job",
		"time_frame":               3600,
		"recurring":                true,
		"trigger_chain_id":         "1",
		"trigger_contract_address": "0x1234567890123456789012345678901234567890",
		"trigger_event":            "Transfer",
		"job_cost_prediction":      "0.1",
		"target_chain_id":          "1",
		"target_contract_address":  "0x1234567890123456789012345678901234567890",
		"target_function":          "execute",
		"abi":                      "[{\"inputs\":[],\"name\":\"execute\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
		"arg_type":                 1,
		"arguments":                []string{},
		"job_type":                 "frontend",
	}

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/jobs", createJSONBody(t, []interface{}{validEventJob}))
	router.ServeHTTP(w, req)

	// Print response body for debugging
	body, _ := io.ReadAll(w.Body)
	t.Logf("Event-based job response: %s", string(body))

	assert.Equal(t, http.StatusOK, w.Code)

	// Create a new response recorder for the next request
	w = httptest.NewRecorder()

	// Test invalid event-based job (invalid target chain ID)
	invalidEventJob := make(map[string]interface{})
	for k, v := range validEventJob {
		invalidEventJob[k] = v
	}
	invalidEventJob["target_chain_id"] = "" // Empty chain ID violates required,chain_id
	req, _ = http.NewRequest("POST", "/api/jobs", createJSONBody(t, []interface{}{invalidEventJob}))
	router.ServeHTTP(w, req)

	// Print response body for debugging
	body, _ = io.ReadAll(w.Body)
	t.Logf("Invalid event-based job response: %s", string(body))

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestConditionBasedJobValidation(t *testing.T) {
	router, validator := setupTestRouter()
	router.POST("/api/jobs", validator.GinMiddleware(), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	// Test valid condition-based job
	validConditionJob := map[string]interface{}{
		"job_id":                  "test-condition-job-1",
		"user_address":            "0x1234567890123456789012345678901234567890",
		"created_chain_id":        "1",
		"timezone":                "UTC",
		"task_definition_id":      5,
		"job_title":               "Test Condition Job",
		"time_frame":              3600,
		"recurring":               true,
		"condition_type":          "price",
		"upper_limit":             100.0,
		"lower_limit":             50.0,
		"value_source_type":       "api",
		"value_source_url":        "https://api.example.com/price",
		"job_cost_prediction":     "0.1",
		"target_chain_id":         "1",
		"target_contract_address": "0x1234567890123456789012345678901234567890",
		"target_function":         "execute",
		"abi":                     "[{\"inputs\":[],\"name\":\"execute\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
		"arg_type":                1,
		"arguments":               []string{},
		"job_type":                "contract",
	}

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/jobs", createJSONBody(t, []interface{}{validConditionJob}))
	router.ServeHTTP(w, req)

	// Print response body for debugging
	body, _ := io.ReadAll(w.Body)
	t.Logf("Condition-based job response: %s", string(body))

	assert.Equal(t, http.StatusOK, w.Code)

	// Create a new response recorder for the next request
	w = httptest.NewRecorder()

	// Test invalid condition-based job (job_title too short)
	invalidConditionJob := make(map[string]interface{})
	for k, v := range validConditionJob {
		invalidConditionJob[k] = v
	}
	invalidConditionJob["job_title"] = "ab" // Too short (min=3)
	req, _ = http.NewRequest("POST", "/api/jobs", createJSONBody(t, []interface{}{invalidConditionJob}))
	router.ServeHTTP(w, req)

	// Print response body for debugging
	body, _ = io.ReadAll(w.Body)
	t.Logf("Invalid condition-based job response: %s", string(body))

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestInvalidTaskDefinitionID(t *testing.T) {
	router, validator := setupTestRouter()
	router.POST("/api/jobs", validator.GinMiddleware(), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	// Test invalid task definition ID
	invalidJob := map[string]interface{}{
		"job_id":                  "test-invalid-job-1",
		"user_address":            "0x1234567890123456789012345678901234567890",
		"created_chain_id":        "1",
		"timezone":                "UTC",
		"task_definition_id":      7, // Invalid ID (max is 6)
		"job_title":               "Test Invalid Job",
		"time_frame":              3600,
		"recurring":               true,
		"job_cost_prediction":     "0.1",
		"target_chain_id":         "1",
		"target_contract_address": "0x1234567890123456789012345678901234567890",
		"target_function":         "execute",
		"abi":                     "[{\"inputs\":[],\"name\":\"execute\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
		"arg_type":                1,
		"arguments":               []string{},
		"job_type":                "template",
	}

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/jobs", createJSONBody(t, []interface{}{invalidJob}))
	router.ServeHTTP(w, req)

	// Print response body for debugging
	body, _ := io.ReadAll(w.Body)
	t.Logf("Invalid task definition ID response: %s", string(body))

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// Helper function to create JSON request body
func createJSONBody(t *testing.T, data interface{}) *bytes.Buffer {
	jsonData, err := json.Marshal(data)
	assert.NoError(t, err)
	return bytes.NewBuffer(jsonData)
}
