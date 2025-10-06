package yaml

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v2"
)

// LoadYAML loads a YAML file into the provided struct
func LoadYAML(path string, target interface{}) error {
	if path == "" {
		return fmt.Errorf("yaml path cannot be empty")
	}

	if target == nil {
		return fmt.Errorf("target cannot be nil")
	}

	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("yaml file does not exist: %s", path)
	}

	// Read file content
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read yaml file %s: %w", path, err)
	}

	// Unmarshal YAML into target struct
	if err := yaml.Unmarshal(data, target); err != nil {
		return fmt.Errorf("failed to unmarshal yaml file %s: %w", path, err)
	}

	return nil
}

// LoadEnvironmentSpecificYAML loads environment-specific YAML files with fallback
// Looks for files in order: base.yaml, base.{env}.yaml
func LoadEnvironmentSpecificYAML(basePath string, target interface{}, environment string) error {
	// Remove extension if present
	basePath = strings.TrimSuffix(basePath, filepath.Ext(basePath))

	// Try to load base configuration first
	baseFile := basePath + ".yaml"
	if err := LoadYAML(baseFile, target); err != nil {
		// Base file is optional, continue if it doesn't exist
		if !os.IsNotExist(err) {
			return fmt.Errorf("failed to load base config: %w", err)
		}
	}

	// Load environment-specific overrides if environment is specified
	if environment != "" {
		envFile := fmt.Sprintf("%s.%s.yaml", basePath, environment)
		if err := LoadYAML(envFile, target); err != nil {
			// Environment file is optional, continue if it doesn't exist
			if !os.IsNotExist(err) {
				return fmt.Errorf("failed to load environment config: %w", err)
			}
		}
	}

	return nil
}

// SaveYAML saves a struct to a YAML file
func SaveYAML(path string, data interface{}) error {
	// Marshal struct to YAML
	yamlData, err := yaml.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal to yaml: %w", err)
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Write YAML data to file
	if err := os.WriteFile(path, yamlData, 0644); err != nil {
		return fmt.Errorf("failed to write yaml file %s: %w", path, err)
	}

	return nil
}

// ValidateYAMLStructure validates that a YAML file has the expected structure
func ValidateYAMLStructure(path string, validator func(interface{}) error) error {
	var data interface{}
	if err := LoadYAML(path, &data); err != nil {
		return fmt.Errorf("failed to load yaml for validation: %w", err)
	}

	if err := validator(data); err != nil {
		return fmt.Errorf("yaml validation failed: %w", err)
	}

	return nil
}

// GetYAMLField retrieves a specific field from a YAML file using dot notation
func GetYAMLField(path, fieldPath string) (interface{}, error) {
	var data map[interface{}]interface{}
	if err := LoadYAML(path, &data); err != nil {
		return nil, fmt.Errorf("failed to load yaml: %w", err)
	}

	value, err := getNestedValue(data, strings.Split(fieldPath, "."))
	if err != nil {
		return nil, fmt.Errorf("failed to get field %s: %w", fieldPath, err)
	}

	return value, nil
}

// Helper function to get nested values from map
func getNestedValue(data map[interface{}]interface{}, keys []string) (interface{}, error) {
	if len(keys) == 0 {
		return data, nil
	}

	value, exists := data[keys[0]]
	if !exists {
		return nil, fmt.Errorf("key %s not found", keys[0])
	}

	if len(keys) == 1 {
		return value, nil
	}

	// Recursively get nested value
	if nestedMap, ok := value.(map[interface{}]interface{}); ok {
		return getNestedValue(nestedMap, keys[1:])
	}

	return nil, fmt.Errorf("key %s is not a map", keys[0])
}
