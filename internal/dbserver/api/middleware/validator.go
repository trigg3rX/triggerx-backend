package middleware

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/trigg3rX/triggerx-backend/pkg/observability"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

type Validator struct {
	validate *validator.Validate
	logger   observability.Logger
}

func NewValidator(ctx context.Context, logger observability.Logger) *Validator {
	v := validator.New()

	// Register custom validations
	err := v.RegisterValidation("ethereum_address", validateEthereumAddress)
	if err != nil {
		logger.Error(ctx, "Error registering validation", observability.Error(err))
	}
	err = v.RegisterValidation("ipfs_url", validateIPFSURL)
	if err != nil {
		logger.Error(ctx, "Error registering validation", observability.Error(err))
	}
	err = v.RegisterValidation("cron", validateCronExpression)
	if err != nil {
		logger.Error(ctx, "Error registering cron validation", observability.Error(err))
	}
	err = v.RegisterValidation("timezone", validateTimezone)
	if err != nil {
		logger.Error(ctx, "Error registering timezone validation", observability.Error(err))
	}

	// Register struct-level validation for CreateJobData
	v.RegisterStructValidation(validateCreateJobData, types.CreateJobData{})

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
			v.logger.Error(c.Request.Context(), "Error reading request body", observability.Error(err))
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid request body",
				"details": err.Error(),
			})
			c.Abort()
			return
		}

		// Create a new reader with the body and restore it
		c.Request.Body = io.NopCloser(bytes.NewBuffer(body))

		// Validate based on the endpoint
		var validationError error
		switch c.Request.URL.Path {
		case "/api/jobs":
			var jobDataArray []types.CreateJobData
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
			v.logger.Error(c.Request.Context(), "Validation error", observability.Error(validationError))
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Validation failed",
				"details": validationError.Error(),
			})
			c.Abort()
			return
		}

		// Restore the body for subsequent handlers
		c.Request.Body = io.NopCloser(bytes.NewBuffer(body))

		c.Next()
	}
}

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

func validateCronExpression(fl validator.FieldLevel) bool {
	cronExpr := fl.Field().String()
	if cronExpr == "" {
		return true // omitempty handles empty values
	}

	// Basic cron expression validation (supports 5 or 6 field formats)
	// Format: [second] minute hour day_of_month month day_of_week
	// Standard 5-field: minute hour day_of_month month day_of_week
	// Extended 6-field: second minute hour day_of_month month day_of_week

	// Split by whitespace
	fields := strings.Fields(cronExpr)
	if len(fields) != 5 && len(fields) != 6 {
		return false
	}

	// Validate each field with cron pattern
	// Each field can contain: numbers, ranges (1-5), lists (1,2,3), wildcards (*), steps (*/5), and special values
	cronFieldPattern := regexp.MustCompile(`^(\*|(\d+(-\d+)?)(,\d+(-\d+)?)*|(\*/\d+)|(\d+(-\d+)?/\d+))$`)

	for _, field := range fields {
		// Handle special cases like @yearly, @monthly, etc.
		if strings.HasPrefix(field, "@") {
			special := []string{"@yearly", "@annually", "@monthly", "@weekly", "@daily", "@midnight", "@hourly"}
			isSpecial := false
			for _, s := range special {
				if field == s {
					isSpecial = true
					break
				}
			}
			if isSpecial {
				continue
			}
			return false
		}

		// Validate standard cron field format
		if !cronFieldPattern.MatchString(field) {
			return false
		}
	}

	return true
}

func validateTimezone(fl validator.FieldLevel) bool {
	timezone := fl.Field().String()
	if timezone == "" {
		return false
	}

	// Try to load the timezone
	_, err := time.LoadLocation(timezone)
	return err == nil
}

// validateCreateJobData performs struct-level validation for CreateJobData
func validateCreateJobData(sl validator.StructLevel) {
	jobData := sl.Current().Interface().(types.CreateJobData)

	// Validate email_id: required for TDI 3,4,5,6,8,9 when recurring is true
	if jobData.Recurring {
		requiresEmail := jobData.TaskDefinitionID == 3 || jobData.TaskDefinitionID == 4 ||
			jobData.TaskDefinitionID == 5 || jobData.TaskDefinitionID == 6 ||
			jobData.TaskDefinitionID == 8 || jobData.TaskDefinitionID == 9
		if requiresEmail && jobData.EmailID == "" {
			sl.ReportError(jobData.EmailID, "email_id", "EmailID", "required_for_recurring", "")
		}
	}

	// Validate safe_address and safe_name: required when is_safe is true
	if jobData.IsSafe {
		if jobData.SafeAddress == "" {
			sl.ReportError(jobData.SafeAddress, "safe_address", "SafeAddress", "required_when_is_safe", "")
		}
		if jobData.SafeName == "" {
			sl.ReportError(jobData.SafeName, "safe_name", "SafeName", "required_when_is_safe", "")
		}
	}

	// Validate job-type-specific fields based on TaskDefinitionID
	switch jobData.TaskDefinitionID {
	case 1, 2, 7: // Time-based jobs
		// ScheduleType is required
		if jobData.ScheduleType == "" {
			sl.ReportError(jobData.ScheduleType, "schedule_type", "ScheduleType", "required_for_time_jobs", "")
		} else {
			// Validate schedule-specific fields
			switch jobData.ScheduleType {
			case "interval":
				if jobData.TimeInterval <= 0 {
					sl.ReportError(jobData.TimeInterval, "time_interval", "TimeInterval", "required_for_interval", "")
				}
			case "cron":
				if jobData.CronExpression == "" {
					sl.ReportError(jobData.CronExpression, "cron_expression", "CronExpression", "required_for_cron", "")
				}
			case "specific":
				if jobData.SpecificSchedule == "" {
					sl.ReportError(jobData.SpecificSchedule, "specific_schedule", "SpecificSchedule", "required_for_specific", "")
				}
			}
		}

		// For dynamic jobs (TDI 2) and agent jobs (TDI 7), execution script fields are required
		if jobData.TaskDefinitionID == 2 || jobData.TaskDefinitionID == 7 {
			if jobData.ExecutionScriptURL == "" {
				sl.ReportError(jobData.ExecutionScriptURL, "execution_script_url", "ExecutionScriptURL", "required_for_agent_jobs", "")
			}
			if jobData.ExecutionScriptLanguage == "" {
				sl.ReportError(jobData.ExecutionScriptLanguage, "execution_script_language", "ExecutionScriptLanguage", "required_for_agent_jobs", "")
			}
		}

	case 3, 4, 8: // Event-based jobs
		// Trigger fields are required
		if jobData.TriggerChainID == "" {
			sl.ReportError(jobData.TriggerChainID, "trigger_chain_id", "TriggerChainID", "required_for_event_jobs", "")
		}
		if jobData.TriggerContractAddress == "" {
			sl.ReportError(jobData.TriggerContractAddress, "trigger_contract_address", "TriggerContractAddress", "required_for_event_jobs", "")
		}
		if jobData.TriggerEvent == "" {
			sl.ReportError(jobData.TriggerEvent, "trigger_event", "TriggerEvent", "required_for_event_jobs", "")
		}

		// For dynamic jobs (TDI 4) and agent jobs (TDI 8), execution script fields are required
		if jobData.TaskDefinitionID == 4 || jobData.TaskDefinitionID == 8 {
			if jobData.ExecutionScriptURL == "" {
				sl.ReportError(jobData.ExecutionScriptURL, "execution_script_url", "ExecutionScriptURL", "required_for_agent_jobs", "")
			}
			if jobData.ExecutionScriptLanguage == "" {
				sl.ReportError(jobData.ExecutionScriptLanguage, "execution_script_language", "ExecutionScriptLanguage", "required_for_agent_jobs", "")
			}
		}

	case 5, 6, 9: // Condition-based jobs
		// Condition fields are required
		if jobData.ConditionType == "" {
			sl.ReportError(jobData.ConditionType, "condition_type", "ConditionType", "required_for_condition_jobs", "")
		}
		if jobData.ValueSourceType == "" {
			sl.ReportError(jobData.ValueSourceType, "value_source_type", "ValueSourceType", "required_for_condition_jobs", "")
		}
		if jobData.ValueSourceUrl == "" {
			sl.ReportError(jobData.ValueSourceUrl, "value_source_url", "ValueSourceUrl", "required_for_condition_jobs", "")
		}

		// Validate condition limits based on condition type
		if jobData.ConditionType == "between" {
			if jobData.UpperLimit == 0 && jobData.LowerLimit == 0 {
				sl.ReportError(jobData.UpperLimit, "upper_limit", "UpperLimit", "required_for_between", "")
			}
			if jobData.LowerLimit >= jobData.UpperLimit && jobData.UpperLimit != 0 && jobData.LowerLimit != 0 {
				sl.ReportError(jobData.UpperLimit, "upper_limit", "UpperLimit", "must_be_greater_than_lower", "")
			}
		} else if jobData.ConditionType != "" {
			// For other condition types, at least one limit should be set
			if jobData.UpperLimit == 0 && jobData.LowerLimit == 0 {
				sl.ReportError(jobData.UpperLimit, "upper_limit", "UpperLimit", "required_for_condition", "")
			}
		}

		// For dynamic jobs (TDI 6) and agent jobs (TDI 9), execution script fields are required
		if jobData.TaskDefinitionID == 6 || jobData.TaskDefinitionID == 9 {
			if jobData.ExecutionScriptURL == "" {
				sl.ReportError(jobData.ExecutionScriptURL, "execution_script_url", "ExecutionScriptURL", "required_for_agent_jobs", "")
			}
			if jobData.ExecutionScriptLanguage == "" {
				sl.ReportError(jobData.ExecutionScriptLanguage, "execution_script_language", "ExecutionScriptLanguage", "required_for_agent_jobs", "")
			}
		}

	default:
		sl.ReportError(jobData.TaskDefinitionID, "task_definition_id", "TaskDefinitionID", "invalid_task_definition_id", fmt.Sprintf("invalid TaskDefinitionID: %d", jobData.TaskDefinitionID))
	}

	// Validate target fields for non-agent jobs (TDI 1-6)
	if jobData.TaskDefinitionID >= 1 && jobData.TaskDefinitionID <= 6 {
		if jobData.TargetChainID == "" {
			sl.ReportError(jobData.TargetChainID, "target_chain_id", "TargetChainID", "required_for_non_agent_jobs", "")
		}
		if jobData.TargetContractAddress == "" {
			sl.ReportError(jobData.TargetContractAddress, "target_contract_address", "TargetContractAddress", "required_for_non_agent_jobs", "")
		}
		if jobData.TargetFunction == "" {
			sl.ReportError(jobData.TargetFunction, "target_function", "TargetFunction", "required_for_non_agent_jobs", "")
		}
	}
}
