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

func init() {
	// Initialize logger for all tests
	config := logging.NewDefaultConfig(logging.ManagerProcess)
	config.UseColors = false // Disable colors in tests
	err := logging.InitServiceLogger(config)
	if err != nil {
		panic("failed to initialize logger: " + err.Error())
	}
}

func setupTestRouter() (*gin.Engine, *Validator) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	logger := logging.GetServiceLogger()
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
		"user_address":            "0x1234567890123456789012345678901234567890",
		"stake_amount":            1000000000000000000,
		"token_amount":            1000000000000000000,
		"task_definition_id":      1,
		"priority":                5,
		"security":                5,
		"custom":                  false,
		"job_title":               "Test Time Job",
		"time_frame":              3600,
		"recurring":               true,
		"time_interval":           1800,
		"job_cost_prediction":     0.1,
		"target_chain_id":         "1",
		"target_contract_address": "0x1234567890123456789012345678901234567890",
		"target_function":         "execute",
		"abi":                     "[{\"inputs\":[],\"name\":\"execute\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
		"arg_type":                1,
		"arguments":               []string{},
		"job_type":                "time",
	}

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/jobs", createJSONBody(t, validTimeJob))
	router.ServeHTTP(w, req)

	// Print response body for debugging
	body, _ := io.ReadAll(w.Body)
	t.Logf("Time-based job response: %s", string(body))

	assert.Equal(t, http.StatusOK, w.Code)

	// Create a new response recorder for the next request
	w = httptest.NewRecorder()

	// Test invalid time-based job (missing required time_interval)
	invalidTimeJob := validTimeJob
	invalidTimeJob["time_interval"] = 0
	req, _ = http.NewRequest("POST", "/api/jobs", createJSONBody(t, invalidTimeJob))
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
		"user_address":             "0x1234567890123456789012345678901234567890",
		"stake_amount":             1000000000000000000,
		"token_amount":             1000000000000000000,
		"task_definition_id":       3,
		"priority":                 5,
		"security":                 5,
		"custom":                   false,
		"job_title":                "Test Event Job",
		"time_frame":               3600,
		"recurring":                true,
		"trigger_chain_id":         "1",
		"trigger_contract_address": "0x1234567890123456789012345678901234567890",
		"trigger_event":            "Transfer",
		"job_cost_prediction":      0.1,
		"target_chain_id":          "1",
		"target_contract_address":  "0x1234567890123456789012345678901234567890",
		"target_function":          "execute",
		"abi":                      "[{\"inputs\":[],\"name\":\"execute\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
		"arg_type":                 1,
		"arguments":                []string{},
		"job_type":                 "event",
	}

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/jobs", createJSONBody(t, validEventJob))
	router.ServeHTTP(w, req)

	// Print response body for debugging
	body, _ := io.ReadAll(w.Body)
	t.Logf("Event-based job response: %s", string(body))

	assert.Equal(t, http.StatusOK, w.Code)

	// Create a new response recorder for the next request
	w = httptest.NewRecorder()

	// Test invalid event-based job (missing trigger fields)
	invalidEventJob := validEventJob
	delete(invalidEventJob, "trigger_event")
	req, _ = http.NewRequest("POST", "/api/jobs", createJSONBody(t, invalidEventJob))
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
		"user_address":            "0x1234567890123456789012345678901234567890",
		"stake_amount":            1000000000000000000,
		"token_amount":            1000000000000000000,
		"task_definition_id":      5,
		"priority":                5,
		"security":                5,
		"custom":                  false,
		"job_title":               "Test Condition Job",
		"time_frame":              3600,
		"recurring":               true,
		"condition_type":          "price",
		"upper_limit":             100.0,
		"lower_limit":             50.0,
		"value_source_type":       "api",
		"value_source_url":        "https://api.example.com/price",
		"job_cost_prediction":     0.1,
		"target_chain_id":         "1",
		"target_contract_address": "0x1234567890123456789012345678901234567890",
		"target_function":         "execute",
		"abi":                     "[{\"inputs\":[],\"name\":\"execute\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
		"arg_type":                1,
		"arguments":               []string{},
		"job_type":                "condition",
	}

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/jobs", createJSONBody(t, validConditionJob))
	router.ServeHTTP(w, req)

	// Print response body for debugging
	body, _ := io.ReadAll(w.Body)
	t.Logf("Condition-based job response: %s", string(body))

	assert.Equal(t, http.StatusOK, w.Code)

	// Create a new response recorder for the next request
	w = httptest.NewRecorder()

	// Test invalid condition-based job (missing condition fields)
	invalidConditionJob := validConditionJob
	delete(invalidConditionJob, "condition_type")
	req, _ = http.NewRequest("POST", "/api/jobs", createJSONBody(t, invalidConditionJob))
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
		"user_address":            "0x1234567890123456789012345678901234567890",
		"stake_amount":            1000000000000000000,
		"token_amount":            1000000000000000000,
		"task_definition_id":      7, // Invalid ID
		"priority":                5,
		"security":                5,
		"custom":                  false,
		"job_title":               "Test Invalid Job",
		"time_frame":              3600,
		"recurring":               true,
		"job_cost_prediction":     0.1,
		"target_chain_id":         "1",
		"target_contract_address": "0x1234567890123456789012345678901234567890",
		"target_function":         "execute",
		"abi":                     "[{\"inputs\":[],\"name\":\"execute\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
		"arg_type":                1,
		"arguments":               []string{},
		"job_type":                "time", // Even though task_definition_id is invalid, we still need a valid job_type
	}

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/jobs", createJSONBody(t, invalidJob))
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




