package middleware

import (
	"bytes"
	// "fmt"
	"io"
	"net/http"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

type Validator struct {
	validate *validator.Validate
	logger   logging.Logger
}

func NewValidator(logger logging.Logger) *Validator {
	v := validator.New()

	// Register custom validations
	err := v.RegisterValidation("ethereum_address", validateEthereumAddress)
	if err != nil {
		logger.Errorf("Error registering validation: %v", err)
	}
	err = v.RegisterValidation("ipfs_url", validateIPFSURL)
	if err != nil {
		logger.Errorf("Error registering validation: %v", err)
	}
	err = v.RegisterValidation("chain_id", validateChainID)
	if err != nil {
		logger.Errorf("Error registering validation: %v", err)
	}

	return &Validator{
		validate: v,
		logger:   logger,
	}
}

func (v *Validator) GinMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Read the request body first
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			v.logger.Errorf("Error reading request body: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid request body",
				"details": err.Error(),
			})
			c.Abort()
			return
		}

		// Log the raw request body
		// v.logger.Infof("Raw request body: %s", string(body))

		// Create a new reader with the body and restore it
		c.Request.Body = io.NopCloser(bytes.NewBuffer(body))

		// Validate based on the endpoint
		var validationError error
		switch c.Request.URL.Path {
		case "/api/jobs":
			var jobDataArray []types.CreateJobDataRequest
			if err := c.ShouldBindJSON(&jobDataArray); err != nil {
				validationError = err
			} else {
				for _, jobData := range jobDataArray {
					if err := v.validate.Struct(jobData); err != nil {
						validationError = err
						break
					}
				}
			}

		case "/api/admin/api-keys":
			var apiKeyData types.CreateApiKeyRequest
			if err := c.ShouldBindJSON(&apiKeyData); err != nil {
				validationError = err
			} else {
				validationError = v.validate.Struct(apiKeyData)
			}

		default:
			// For unknown endpoints, just pass through
			c.Next()
			return
		}

		if validationError != nil {
			v.logger.Errorf("Validation error: %v", validationError)
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Validation failed",
				"details": validationError.Error(),
			})
			c.Abort()
			return
		}

		// Validate based on the endpoint
		// var validationErrors []string
		// switch c.Request.URL.Path {
		// case "/api/jobs":
		// 	for i, jobData := range jobDataArray {
		// 		if err := v.validateCreateJob(c, jobData); err != nil {
		// 			validationErrors = append(validationErrors, fmt.Sprintf("Job %d: %v", i+1, err))
		// 		}
		// 	}
		// case "/api/tasks":
		// 	if err := v.validateCreateTask(c, jobDataArray); err != nil {
		// 		validationErrors = append(validationErrors, err.Error())
		// 	}
		// case "/api/keepers/form":
		// 	if err := v.validateCreateKeeperForm(c, jobDataArray); err != nil {
		// 		validationErrors = append(validationErrors, err.Error())
		// 	}
		// case "/api/admin/api-keys":
		// 	if err := v.validateCreateApiKey(c, jobDataArray); err != nil {
		// 		validationErrors = append(validationErrors, err.Error())
		// 	}
		// }

		// if len(validationErrors) > 0 {
		// 	c.JSON(http.StatusBadRequest, gin.H{
		// 		"error":   "Validation failed",
		// 		"details": validationErrors,
		// 	})
		// 	c.Abort()
		// 	return
		// }

		// Restore the body for subsequent handlers
		c.Request.Body = io.NopCloser(bytes.NewBuffer(body))

		c.Next()
	}
}

// func (v *Validator) validateCreateJob(c *gin.Context, jobData types.CreateJobData) error {
// 	// First validate the common fields
// 	if err := v.validate.Struct(jobData); err != nil {
// 		return err
// 	}

// Determine job type based on TaskDefinitionID
// switch {
// case jobData.TaskDefinitionID >= types.TaskDefTimeBasedStart && jobData.TaskDefinitionID <= types.TaskDefTimeBasedEnd:
// Time-based job validation
// if jobData.TimeInterval <= 0 {
// 	return fmt.Errorf("time_interval is required for time-based jobs (TaskDefinitionID: %d)", jobData.TaskDefinitionID)
// }
// Time jobs don't need trigger or condition fields
// if jobData.TriggerChainID != "" || jobData.TriggerContractAddress != "" || jobData.TriggerEvent != "" {
// 	return fmt.Errorf("trigger fields should not be set for time-based jobs (TaskDefinitionID: %d)", jobData.TaskDefinitionID)
// }
// if jobData.ConditionType != "" || jobData.UpperLimit != 0 || jobData.LowerLimit != 0 {
// 	return fmt.Errorf("condition fields should not be set for time-based jobs (TaskDefinitionID: %d)", jobData.TaskDefinitionID)
// }

// case jobData.TaskDefinitionID >= types.TaskDefEventBasedStart && jobData.TaskDefinitionID <= types.TaskDefEventBasedEnd:
// 	// Event-based job validation
// 	if jobData.TriggerChainID == "" || jobData.TriggerContractAddress == "" || jobData.TriggerEvent == "" {
// 		return fmt.Errorf("trigger fields are required for event-based jobs (TaskDefinitionID: %d)", jobData.TaskDefinitionID)
// 	}
// Event jobs don't need time interval or condition fields
// if jobData.TimeInterval != 0 {
// 	return fmt.Errorf("time_interval should not be set for event-based jobs (TaskDefinitionID: %d)", jobData.TaskDefinitionID)
// }
// if jobData.ConditionType != "" || jobData.UpperLimit != 0 || jobData.LowerLimit != 0 {
// 	return fmt.Errorf("condition fields should not be set for event-based jobs (TaskDefinitionID: %d)", jobData.TaskDefinitionID)
// }

// case jobData.TaskDefinitionID >= types.TaskDefConditionBasedStart && jobData.TaskDefinitionID <= types.TaskDefConditionBasedEnd:
// 	// Condition-based job validation
// 	if jobData.ConditionType == "" || jobData.UpperLimit == 0 || jobData.LowerLimit == 0 {
// 		return fmt.Errorf("condition fields are required for condition-based jobs (TaskDefinitionID: %d)", jobData.TaskDefinitionID)
// 	}
// 	if jobData.ValueSourceType == "" || jobData.ValueSourceUrl == "" {
// 		return fmt.Errorf("value source fields are required for condition-based jobs (TaskDefinitionID: %d)", jobData.TaskDefinitionID)
// 	}
// Condition jobs don't need time interval or trigger fields
// if jobData.TimeInterval != 0 {
// 	return fmt.Errorf("time_interval should not be set for condition-based jobs (TaskDefinitionID: %d)", jobData.TaskDefinitionID)
// }
// if jobData.TriggerChainID != "" || jobData.TriggerContractAddress != "" || jobData.TriggerEvent != "" {
// 	return fmt.Errorf("trigger fields should not be set for condition-based jobs (TaskDefinitionID: %d)", jobData.TaskDefinitionID)
// }

// default:
// 	return fmt.Errorf("invalid TaskDefinitionID: %d", jobData.TaskDefinitionID)
// }

// 	return nil
// }

// func (v *Validator) validateCreateTask(c *gin.Context, body interface{}) error {
// 	var taskData types.CreateTaskDataRequest
// 	if err := c.ShouldBindJSON(&taskData); err != nil {
// 		return err
// 	}

// 	return v.validate.Struct(taskData)
// }

// func (v *Validator) validateCreateApiKey(c *gin.Context, body interface{}) error {
// 	var apiKeyData types.CreateApiKeyRequest
// 	if err := c.ShouldBindJSON(&apiKeyData); err != nil {
// 		return err
// 	}

// 	return v.validate.Struct(apiKeyData)
// }

// Custom validation functions
func validateEthereumAddress(fl validator.FieldLevel) bool {
	address := fl.Field().String()
	// Basic Ethereum address validation (0x followed by 40 hex characters)
	re := regexp.MustCompile("^0x[0-9a-fA-F]{40}$")
	return re.MatchString(address)
}

func validateIPFSURL(fl validator.FieldLevel) bool {
	url := fl.Field().String()

	// Check for native IPFS protocol
	if strings.HasPrefix(url, "ipfs://") {
		return true
	}

	// Check for HTTPS URLs that contain "/ipfs/" path (covers various gateways)
	if strings.HasPrefix(url, "https://") && strings.Contains(url, "/ipfs/") {
		return true
	}

	return false
}

func validateChainID(fl validator.FieldLevel) bool {
	chainID := fl.Field().String()
	// Add your chain ID validation logic here
	// For now, just checking if it's not empty
	return chainID != ""
}
