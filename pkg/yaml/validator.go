package yaml

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/trigg3rX/triggerx-backend/pkg/env"
)

// Validator provides validation functions for YAML configuration
type Validator struct{}

// NewValidator creates a new YAML validator
func NewValidator() *Validator {
	return &Validator{}
}

// ValidateConfig validates a configuration struct
func (v *Validator) ValidateConfig(config interface{}) error {
	configValue := reflect.ValueOf(config)
	if configValue.Kind() == reflect.Ptr {
		configValue = configValue.Elem()
	}

	if configValue.Kind() != reflect.Struct {
		return fmt.Errorf("config must be a struct")
	}

	return v.validateStruct(configValue)
}

// validateStruct recursively validates struct fields
func (v *Validator) validateStruct(structValue reflect.Value) error {
	structType := structValue.Type()

	for i := 0; i < structValue.NumField(); i++ {
		field := structValue.Field(i)
		fieldType := structType.Field(i)

		// Skip unexported fields
		if !field.CanInterface() {
			continue
		}

		// Get validation tag
		tag := fieldType.Tag.Get("validate")
		if tag == "" {
			continue
		}

		// Validate field based on tags
		if err := v.validateField(field, fieldType, tag); err != nil {
			return fmt.Errorf("field %s: %w", fieldType.Name, err)
		}

		// Recursively validate nested structs
		if field.Kind() == reflect.Struct {
			if err := v.validateStruct(field); err != nil {
				return fmt.Errorf("nested field %s: %w", fieldType.Name, err)
			}
		}
	}

	return nil
}

// validateField validates a single field based on validation tags
func (v *Validator) validateField(field reflect.Value, fieldType reflect.StructField, tag string) error {
	rules := strings.Split(tag, ",")

	for _, rule := range rules {
		rule = strings.TrimSpace(rule)
		if rule == "" {
			continue
		}

		if err := v.applyValidationRule(field, fieldType, rule); err != nil {
			return err
		}
	}

	return nil
}

// applyValidationRule applies a single validation rule
func (v *Validator) applyValidationRule(field reflect.Value, fieldType reflect.StructField, rule string) error {
	parts := strings.Split(rule, "=")
	ruleName := parts[0]
	ruleValue := ""
	if len(parts) > 1 {
		ruleValue = parts[1]
	}

	switch ruleName {
	case "required":
		return v.validateRequired(field, fieldType)
	case "port":
		return v.validatePort(field, fieldType)
	case "ip":
		return v.validateIP(field, fieldType)
	case "email":
		return v.validateEmail(field, fieldType)
	case "eth_address":
		return v.validateEthAddress(field, fieldType)
	case "duration":
		return v.validateDuration(field, fieldType)
	case "url":
		return v.validateURL(field, fieldType)
	case "oneof":
		return v.validateOneOf(field, fieldType, ruleValue)
	case "min":
		return v.validateMin(field, fieldType, ruleValue)
	case "max":
		return v.validateMax(field, fieldType, ruleValue)
	default:
		return fmt.Errorf("unknown validation rule: %s", ruleName)
	}
}

// validateRequired checks if a required field is not empty
func (v *Validator) validateRequired(field reflect.Value, fieldType reflect.StructField) error {
	if field.Kind() == reflect.String {
		if env.IsEmpty(field.String()) {
			return fmt.Errorf("required field %s cannot be empty", fieldType.Name)
		}
	} else if field.Kind() == reflect.Int || field.Kind() == reflect.Int64 {
		if field.Int() == 0 {
			return fmt.Errorf("required field %s cannot be zero", fieldType.Name)
		}
	} else if field.Kind() == reflect.Bool {
		// Bool fields are always valid
		return nil
	}

	return nil
}

// validatePort validates port numbers
func (v *Validator) validatePort(field reflect.Value, fieldType reflect.StructField) error {
	if field.Kind() != reflect.String {
		return fmt.Errorf("port field %s must be a string", fieldType.Name)
	}

	if !env.IsValidPort(field.String()) {
		return fmt.Errorf("invalid port number for field %s: %s", fieldType.Name, field.String())
	}

	return nil
}

// validateIP validates IP addresses
func (v *Validator) validateIP(field reflect.Value, fieldType reflect.StructField) error {
	if field.Kind() != reflect.String {
		return fmt.Errorf("ip field %s must be a string", fieldType.Name)
	}

	if !env.IsValidIPAddress(field.String()) {
		return fmt.Errorf("invalid IP address for field %s: %s", fieldType.Name, field.String())
	}

	return nil
}

// validateEmail validates email addresses
func (v *Validator) validateEmail(field reflect.Value, fieldType reflect.StructField) error {
	if field.Kind() != reflect.String {
		return fmt.Errorf("email field %s must be a string", fieldType.Name)
	}

	if !env.IsValidEmail(field.String()) {
		return fmt.Errorf("invalid email address for field %s: %s", fieldType.Name, field.String())
	}

	return nil
}

// validateEthAddress validates Ethereum addresses
func (v *Validator) validateEthAddress(field reflect.Value, fieldType reflect.StructField) error {
	if field.Kind() != reflect.String {
		return fmt.Errorf("ethereum address field %s must be a string", fieldType.Name)
	}

	if !env.IsValidEthAddress(field.String()) {
		return fmt.Errorf("invalid ethereum address for field %s: %s", fieldType.Name, field.String())
	}

	return nil
}

// validateDuration validates duration strings
func (v *Validator) validateDuration(field reflect.Value, fieldType reflect.StructField) error {
	if field.Kind() != reflect.String {
		return fmt.Errorf("duration field %s must be a string", fieldType.Name)
	}

	_, err := time.ParseDuration(field.String())
	if err != nil {
		return fmt.Errorf("invalid duration for field %s: %s", fieldType.Name, field.String())
	}

	return nil
}

// validateURL validates URLs
func (v *Validator) validateURL(field reflect.Value, fieldType reflect.StructField) error {
	if field.Kind() != reflect.String {
		return fmt.Errorf("url field %s must be a string", fieldType.Name)
	}

	if !env.IsValidURL(field.String()) {
		return fmt.Errorf("invalid URL for field %s: %s", fieldType.Name, field.String())
	}

	return nil
}

// validateOneOf validates that a field value is one of the allowed values
func (v *Validator) validateOneOf(field reflect.Value, fieldType reflect.StructField, allowedValues string) error {
	if field.Kind() != reflect.String {
		return fmt.Errorf("oneof field %s must be a string", fieldType.Name)
	}

	allowed := strings.Split(allowedValues, "|")
	value := field.String()

	for _, allowedValue := range allowed {
		if value == allowedValue {
			return nil
		}
	}

	return fmt.Errorf("value %s for field %s not in allowed values: %s", value, fieldType.Name, allowedValues)
}

// validateMin validates minimum values for numeric fields
func (v *Validator) validateMin(field reflect.Value, fieldType reflect.StructField, minValue string) error {
	switch field.Kind() {
	case reflect.Int, reflect.Int64:
		min, err := parseInt(minValue)
		if err != nil {
			return fmt.Errorf("invalid min value for field %s: %s", fieldType.Name, minValue)
		}
		if field.Int() < min {
			return fmt.Errorf("value %d for field %s is less than minimum %d", field.Int(), fieldType.Name, min)
		}
	case reflect.Float32, reflect.Float64:
		min, err := parseFloat(minValue)
		if err != nil {
			return fmt.Errorf("invalid min value for field %s: %s", fieldType.Name, minValue)
		}
		if field.Float() < min {
			return fmt.Errorf("value %f for field %s is less than minimum %f", field.Float(), fieldType.Name, min)
		}
	case reflect.String:
		min, err := parseInt(minValue)
		if err != nil {
			return fmt.Errorf("invalid min value for field %s: %s", fieldType.Name, minValue)
		}
		if len(field.String()) < int(min) {
			return fmt.Errorf("string length %d for field %s is less than minimum %d", len(field.String()), fieldType.Name, min)
		}
	}

	return nil
}

// validateMax validates maximum values for numeric fields
func (v *Validator) validateMax(field reflect.Value, fieldType reflect.StructField, maxValue string) error {
	switch field.Kind() {
	case reflect.Int, reflect.Int64:
		max, err := parseInt(maxValue)
		if err != nil {
			return fmt.Errorf("invalid max value for field %s: %s", fieldType.Name, maxValue)
		}
		if field.Int() > max {
			return fmt.Errorf("value %d for field %s is greater than maximum %d", field.Int(), fieldType.Name, max)
		}
	case reflect.Float32, reflect.Float64:
		max, err := parseFloat(maxValue)
		if err != nil {
			return fmt.Errorf("invalid max value for field %s: %s", fieldType.Name, maxValue)
		}
		if field.Float() > max {
			return fmt.Errorf("value %f for field %s is greater than maximum %f", field.Float(), fieldType.Name, max)
		}
	case reflect.String:
		max, err := parseInt(maxValue)
		if err != nil {
			return fmt.Errorf("invalid max value for field %s: %s", fieldType.Name, maxValue)
		}
		if len(field.String()) > int(max) {
			return fmt.Errorf("string length %d for field %s is greater than maximum %d", len(field.String()), fieldType.Name, max)
		}
	}

	return nil
}

// Helper functions for parsing validation values
func parseInt(value string) (int64, error) {
	return strconv.ParseInt(value, 10, 64)
}

func parseFloat(value string) (float64, error) {
	return strconv.ParseFloat(value, 64)
}

// ValidateHealthConfig validates the health service configuration
func ValidateHealthConfig(config interface{}) error {
	validator := NewValidator()
	return validator.ValidateConfig(config)
}
