package yaml

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// TestYAMLStruct represents a simple struct for testing
type TestYAMLStruct struct {
	Name       string `yaml:"name" validate:"required,min=1"`
	Age        int    `yaml:"age" validate:"min=1,max=120"`
	Email      string `yaml:"email" validate:"email"`
	Port       string `yaml:"port" validate:"port"`
	IP         string `yaml:"ip" validate:"ip"`
	EthAddress string `yaml:"eth_address" validate:"eth_address"`
	Duration   string `yaml:"duration" validate:"duration"`
	URL        string `yaml:"url" validate:"url"`
	Strategy   string `yaml:"strategy" validate:"oneof=round_robin|least_connections|weighted"`
	Optional   string `yaml:"optional"`
}

func TestLoadYAML(t *testing.T) {
	// Create a temporary YAML file
	tempFile := createTempYAML(t, `
name: John Doe
age: 30
email: john@example.com
port: "8080"
ip: "127.0.0.1"
eth_address: "0x1234567890123456789012345678901234567890"
duration: "30s"
url: "https://example.com"
strategy: "round_robin"
optional: "test"
`)
	defer os.Remove(tempFile)

	var config TestYAMLStruct
	err := LoadYAML(tempFile, &config)
	if err != nil {
		t.Fatalf("LoadYAML failed: %v", err)
	}

	// Verify loaded values
	expected := TestYAMLStruct{
		Name:       "John Doe",
		Age:        30,
		Email:      "john@example.com",
		Port:       "8080",
		IP:         "127.0.0.1",
		EthAddress: "0x1234567890123456789012345678901234567890",
		Duration:   "30s",
		URL:        "https://example.com",
		Strategy:   "round_robin",
		Optional:   "test",
	}

	if config != expected {
		t.Errorf("Expected %+v, got %+v", expected, config)
	}
}

func TestLoadYAML_FileNotFound(t *testing.T) {
	var config TestYAMLStruct
	err := LoadYAML("nonexistent.yaml", &config)
	if err == nil {
		t.Error("Expected error for nonexistent file")
	}
}

func TestLoadYAML_InvalidYAML(t *testing.T) {
	// Create a temporary file with invalid YAML
	tempFile := createTempYAML(t, `
invalid: yaml: content
  - this is broken
`)
	defer os.Remove(tempFile)

	var config TestYAMLStruct
	err := LoadYAML(tempFile, &config)
	if err == nil {
		t.Error("Expected error for invalid YAML")
	}
}

func TestLoadEnvironmentSpecificYAML(t *testing.T) {
	// Create base config file
	baseFile := createTempYAML(t, `
name: Base Config
age: 25
email: base@example.com
port: "8080"
strategy: "round_robin"
`)
	defer os.Remove(baseFile)

	// Extract base name without extension
	baseName := filepath.Base(baseFile)
	baseName = baseName[:len(baseName)-len(filepath.Ext(baseName))]

	// Create environment-specific config file
	envFile := filepath.Join(filepath.Dir(baseFile), baseName+".env.yaml")
	err := os.WriteFile(envFile, []byte(`
name: Environment Config
age: 35
port: "9090"
strategy: "weighted"
`), 0644)
	if err != nil {
		t.Fatalf("Failed to create env file: %v", err)
	}
	defer os.Remove(envFile)

	// Remove extension from base path
	basePath := filepath.Join(filepath.Dir(baseFile), baseName)

	var config TestYAMLStruct
	err = LoadEnvironmentSpecificYAML(basePath, &config, "env")
	if err != nil {
		t.Fatalf("LoadEnvironmentSpecificYAML failed: %v", err)
	}

	// Verify that environment values override base values
	expected := TestYAMLStruct{
		Name:     "Environment Config", // Overridden
		Age:      35,                   // Overridden
		Email:    "base@example.com",   // From base
		Port:     "9090",               // Overridden
		Strategy: "weighted",           // Overridden
	}

	if config.Name != expected.Name {
		t.Errorf("Expected name %s, got %s", expected.Name, config.Name)
	}
	if config.Age != expected.Age {
		t.Errorf("Expected age %d, got %d", expected.Age, config.Age)
	}
	if config.Email != expected.Email {
		t.Errorf("Expected email %s, got %s", expected.Email, config.Email)
	}
	if config.Port != expected.Port {
		t.Errorf("Expected port %s, got %s", expected.Port, config.Port)
	}
	if config.Strategy != expected.Strategy {
		t.Errorf("Expected strategy %s, got %s", expected.Strategy, config.Strategy)
	}
}

func TestLoadEnvironmentSpecificYAML_NoEnvFile(t *testing.T) {
	// Create base config file
	baseFile := createTempYAML(t, `
name: Base Config
age: 25
`)
	defer os.Remove(baseFile)

	// Extract base name without extension
	baseName := filepath.Base(baseFile)
	baseName = baseName[:len(baseName)-len(filepath.Ext(baseName))]

	// Remove extension from base path
	basePath := filepath.Join(filepath.Dir(baseFile), baseName)

	var config TestYAMLStruct
	err := LoadEnvironmentSpecificYAML(basePath, &config, "nonexistent")
	if err != nil {
		t.Fatalf("LoadEnvironmentSpecificYAML failed: %v", err)
	}

	// Verify base config is loaded
	if config.Name != "Base Config" {
		t.Errorf("Expected name 'Base Config', got %s", config.Name)
	}
	if config.Age != 25 {
		t.Errorf("Expected age 25, got %d", config.Age)
	}
}

func TestSaveYAML(t *testing.T) {
	config := TestYAMLStruct{
		Name:     "Test Config",
		Age:      42,
		Email:    "test@example.com",
		Port:     "8080",
		Strategy: "round_robin",
	}

	// Create temporary file
	tempFile := filepath.Join(t.TempDir(), "test_config.yaml")

	err := SaveYAML(tempFile, config)
	if err != nil {
		t.Fatalf("SaveYAML failed: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(tempFile); os.IsNotExist(err) {
		t.Error("YAML file was not created")
	}

	// Load and verify content
	var loadedConfig TestYAMLStruct
	err = LoadYAML(tempFile, &loadedConfig)
	if err != nil {
		t.Fatalf("Failed to load saved YAML: %v", err)
	}

	if loadedConfig != config {
		t.Errorf("Expected %+v, got %+v", config, loadedConfig)
	}
}

func TestValidateYAMLStructure(t *testing.T) {
	// Create valid YAML file
	validFile := createTempYAML(t, `
name: Valid Config
age: 30
`)
	defer os.Remove(validFile)

	// Test with valid validator
	err := ValidateYAMLStructure(validFile, func(data interface{}) error {
		// Simple validation - check if data is a map (YAML creates map[interface{}]interface{})
		if _, ok := data.(map[interface{}]interface{}); !ok {
			return fmt.Errorf("expected map, got %T", data)
		}
		return nil
	})
	if err != nil {
		t.Errorf("ValidateYAMLStructure failed with valid YAML: %v", err)
	}

	// Test with invalid validator
	err = ValidateYAMLStructure(validFile, func(data interface{}) error {
		return fmt.Errorf("validation failed")
	})
	if err == nil {
		t.Error("Expected validation error")
	}
}

func TestGetYAMLField(t *testing.T) {
	// Create YAML file with nested structure
	yamlContent := `
server:
  http:
    port: "8080"
    host: "localhost"
  grpc:
    port: "9090"
database:
  host: "db.example.com"
  port: "5432"
`
	tempFile := createTempYAML(t, yamlContent)
	defer os.Remove(tempFile)

	// Test getting nested field
	value, err := GetYAMLField(tempFile, "server.http.port")
	if err != nil {
		t.Fatalf("GetYAMLField failed: %v", err)
	}
	if value != "8080" {
		t.Errorf("Expected '8080', got %v", value)
	}

	// Test getting top-level field
	value, err = GetYAMLField(tempFile, "database.host")
	if err != nil {
		t.Fatalf("GetYAMLField failed: %v", err)
	}
	if value != "db.example.com" {
		t.Errorf("Expected 'db.example.com', got %v", value)
	}

	// Test getting nonexistent field
	_, err = GetYAMLField(tempFile, "nonexistent.field")
	if err == nil {
		t.Error("Expected error for nonexistent field")
	}
}

func TestGetYAMLField_InvalidPath(t *testing.T) {
	tempFile := createTempYAML(t, `
server:
  port: "8080"
`)
	defer os.Remove(tempFile)

	// Test accessing non-map value as map
	_, err := GetYAMLField(tempFile, "server.port.subfield")
	if err == nil {
		t.Error("Expected error for invalid path")
	}
}

// Helper function to create temporary YAML file
func createTempYAML(t *testing.T, content string) string {
	file, err := os.CreateTemp("", "test_*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	_, err = file.WriteString(content)
	if err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}

	err = file.Close()
	if err != nil {
		t.Fatalf("Failed to close temp file: %v", err)
	}

	return file.Name()
}

// Test error cases
func TestLoadYAML_EmptyPath(t *testing.T) {
	var config TestYAMLStruct
	err := LoadYAML("", &config)
	if err == nil {
		t.Error("Expected error for empty path")
	}
}

func TestSaveYAML_CreateDirectory(t *testing.T) {
	config := TestYAMLStruct{Name: "Test"}

	// Test saving to nested directory that doesn't exist
	tempFile := filepath.Join(t.TempDir(), "nested", "config.yaml")

	err := SaveYAML(tempFile, config)
	if err != nil {
		t.Fatalf("SaveYAML failed to create directory: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(tempFile); os.IsNotExist(err) {
		t.Error("YAML file was not created in nested directory")
	}
}

func TestLoadYAML_NilTarget(t *testing.T) {
	tempFile := createTempYAML(t, `name: test`)
	defer os.Remove(tempFile)

	err := LoadYAML(tempFile, nil)
	if err == nil {
		t.Error("Expected error for nil target")
	}
}
