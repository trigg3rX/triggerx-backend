package yaml

import (
	"testing"
)

// TestStruct represents a struct for testing validation
type TestStruct struct {
	RequiredString string `validate:"required"`
	RequiredInt    int    `validate:"required"`
	RequiredBool   bool   `validate:"required"`
	OptionalString string `validate:"min=3"`

	// Validation types
	PortField       string `validate:"port"`
	IPField         string `validate:"ip"`
	EmailField      string `validate:"email"`
	EthAddressField string `validate:"eth_address"`
	DurationField   string `validate:"duration"`
	URLField        string `validate:"url"`
	StrategyField   string `validate:"oneof=round_robin|least_connections|weighted"`

	// Numeric validation
	MinIntField    int    `validate:"min=10"`
	MaxIntField    int    `validate:"max=100"`
	MinMaxIntField int    `validate:"min=5,max=50"`
	MinStringField string `validate:"min=5"`
	MaxStringField string `validate:"max=10"`
}

// Helper function to create a valid base test struct
func createValidTestStruct() TestStruct {
	return TestStruct{
		RequiredString:  "test",
		RequiredInt:     42,
		RequiredBool:    true,
		OptionalString:  "valid",
		PortField:       "8080",
		IPField:         "127.0.0.1",
		EmailField:      "test@example.com",
		EthAddressField: "0x1234567890123456789012345678901234567890",
		DurationField:   "30s",
		URLField:        "https://example.com",
		StrategyField:   "round_robin",
		MinIntField:     15,
		MaxIntField:     75,
		MinMaxIntField:  25,
		MinStringField:  "validstring",
		MaxStringField:  "short",
	}
}

func TestValidateConfig_ValidConfig(t *testing.T) {
	config := createValidTestStruct()

	validator := NewValidator()
	err := validator.ValidateConfig(&config)
	if err != nil {
		t.Errorf("Validation failed for valid config: %v", err)
	}
}

func TestValidateConfig_RequiredFields(t *testing.T) {
	tests := []struct {
		name    string
		config  TestStruct
		wantErr bool
	}{
		{
			name: "missing required string",
			config: TestStruct{
				RequiredInt:  42,
				RequiredBool: true,
			},
			wantErr: true,
		},
		{
			name: "missing required int",
			config: TestStruct{
				RequiredString: "test",
				RequiredBool:   true,
				OptionalString: "valid",
			},
			wantErr: true,
		},
		{
			name: "required bool false (should be valid)",
			config: TestStruct{
				RequiredString:  "test",
				RequiredInt:     42,
				RequiredBool:    false,
				OptionalString:  "valid",
				PortField:       "8080",
				IPField:         "127.0.0.1",
				EmailField:      "test@example.com",
				EthAddressField: "0x1234567890123456789012345678901234567890",
				DurationField:   "30s",
				URLField:        "https://example.com",
				StrategyField:   "round_robin",
				MinIntField:     15,
				MaxIntField:     75,
				MinMaxIntField:  25,
				MinStringField:  "validstring",
				MaxStringField:  "short",
			},
			wantErr: false,
		},
		{
			name: "empty required string",
			config: TestStruct{
				RequiredString: "",
				RequiredInt:    42,
				RequiredBool:   true,
				OptionalString: "valid",
			},
			wantErr: true,
		},
		{
			name: "zero required int",
			config: TestStruct{
				RequiredString: "test",
				RequiredInt:    0,
				RequiredBool:   true,
				OptionalString: "valid",
			},
			wantErr: true,
		},
	}

	validator := NewValidator()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateConfig(&tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateConfig_PortValidation(t *testing.T) {
	tests := []struct {
		name    string
		port    string
		wantErr bool
	}{
		{"valid port", "8080", false},
		{"valid port range", "3000", false},
		{"valid high port", "65535", false},
		{"invalid port", "99999", true},
		{"invalid port", "80", true}, // Too low
		{"invalid port", "abc", true},
		{"empty port", "", true},
	}

	validator := NewValidator()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := createValidTestStruct()
			config.PortField = tt.port
			err := validator.ValidateConfig(&config)
			if (err != nil) != tt.wantErr {
				t.Errorf("Port validation error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateConfig_IPValidation(t *testing.T) {
	tests := []struct {
		name    string
		ip      string
		wantErr bool
	}{
		{"valid localhost", "localhost", false},
		{"valid IP", "127.0.0.1", false},
		{"valid IP", "192.168.1.1", false},
		{"invalid IP", "999.999.999.999", true},
		{"invalid IP", "not.an.ip", true},
		{"empty IP", "", true},
	}

	validator := NewValidator()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := createValidTestStruct()
			config.IPField = tt.ip
			err := validator.ValidateConfig(&config)
			if (err != nil) != tt.wantErr {
				t.Errorf("IP validation error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateConfig_EmailValidation(t *testing.T) {
	tests := []struct {
		name    string
		email   string
		wantErr bool
	}{
		{"valid email", "test@example.com", false},
		{"valid email with subdomain", "user@mail.example.com", false},
		{"valid email with hyphen", "user-name@example-domain.com", false},
		{"invalid email", "notanemail", true},
		{"invalid email", "@example.com", true},
		{"invalid email", "test@", true},
		{"empty email", "", true},
	}

	validator := NewValidator()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := createValidTestStruct()
			config.EmailField = tt.email
			err := validator.ValidateConfig(&config)
			if (err != nil) != tt.wantErr {
				t.Errorf("Email validation error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateConfig_EthAddressValidation(t *testing.T) {
	tests := []struct {
		name    string
		address string
		wantErr bool
	}{
		{"valid address", "0x1234567890123456789012345678901234567890", false},
		{"valid address lowercase", "0xabcdef123456789012345678901234567890abcd", false},
		{"valid address mixed case", "0x1234567890123456789012345678901234567890", false},
		{"invalid address too short", "0x123456789012345678901234567890123456789", true},
		{"invalid address too long", "0x12345678901234567890123456789012345678901", true},
		{"invalid address no 0x", "1234567890123456789012345678901234567890", true},
		{"empty address", "", true},
	}

	validator := NewValidator()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := createValidTestStruct()
			config.EthAddressField = tt.address
			err := validator.ValidateConfig(&config)
			if (err != nil) != tt.wantErr {
				t.Errorf("EthAddress validation error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateConfig_DurationValidation(t *testing.T) {
	tests := []struct {
		name     string
		duration string
		wantErr  bool
	}{
		{"valid duration", "30s", false},
		{"valid duration", "5m", false},
		{"valid duration", "1h", false},
		{"valid duration", "2h30m", false},
		{"invalid duration", "invalid", true},
		{"invalid duration", "30", true},
		{"empty duration", "", true},
	}

	validator := NewValidator()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := createValidTestStruct()
			config.DurationField = tt.duration
			err := validator.ValidateConfig(&config)
			if (err != nil) != tt.wantErr {
				t.Errorf("Duration validation error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateConfig_URLValidation(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{"valid http URL", "http://example.com", false},
		{"valid https URL", "https://example.com", false},
		{"valid URL with port", "http://example.com:8080", false},
		{"valid URL with path", "https://example.com/path", false},
		{"invalid URL", "not-a-url", true},
		{"invalid URL", "ftp://example.com", true},
		{"empty URL", "", true},
	}

	validator := NewValidator()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := createValidTestStruct()
			config.URLField = tt.url
			err := validator.ValidateConfig(&config)
			if (err != nil) != tt.wantErr {
				t.Errorf("URL validation error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateConfig_OneOfValidation(t *testing.T) {
	tests := []struct {
		name     string
		strategy string
		wantErr  bool
	}{
		{"valid round_robin", "round_robin", false},
		{"valid least_connections", "least_connections", false},
		{"valid weighted", "weighted", false},
		{"invalid strategy", "invalid", true},
		{"empty strategy", "", true},
	}

	validator := NewValidator()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := createValidTestStruct()
			config.StrategyField = tt.strategy
			err := validator.ValidateConfig(&config)
			if (err != nil) != tt.wantErr {
				t.Errorf("OneOf validation error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateConfig_MinMaxValidation(t *testing.T) {
	tests := []struct {
		name    string
		config  TestStruct
		wantErr bool
	}{
		{
			name: "valid min int",
			config: TestStruct{
				RequiredString:  "test",
				RequiredInt:     42,
				RequiredBool:    true,
				OptionalString:  "valid",
				PortField:       "8080",
				IPField:         "127.0.0.1",
				EmailField:      "test@example.com",
				EthAddressField: "0x1234567890123456789012345678901234567890",
				DurationField:   "30s",
				URLField:        "https://example.com",
				StrategyField:   "round_robin",
				MinIntField:     15,
				MaxIntField:     75,
				MinMaxIntField:  25,
				MinStringField:  "validstring",
				MaxStringField:  "short",
			},
			wantErr: false,
		},
		{
			name: "invalid min int",
			config: TestStruct{
				RequiredString:  "test",
				RequiredInt:     42,
				RequiredBool:    true,
				OptionalString:  "valid",
				PortField:       "8080",
				IPField:         "127.0.0.1",
				EmailField:      "test@example.com",
				EthAddressField: "0x1234567890123456789012345678901234567890",
				DurationField:   "30s",
				URLField:        "https://example.com",
				StrategyField:   "round_robin",
				MinIntField:     5,
				MaxIntField:     75,
				MinMaxIntField:  25,
				MinStringField:  "validstring",
				MaxStringField:  "short",
			},
			wantErr: true,
		},
		{
			name: "valid max int",
			config: TestStruct{
				RequiredString:  "test",
				RequiredInt:     42,
				RequiredBool:    true,
				OptionalString:  "valid",
				PortField:       "8080",
				IPField:         "127.0.0.1",
				EmailField:      "test@example.com",
				EthAddressField: "0x1234567890123456789012345678901234567890",
				DurationField:   "30s",
				URLField:        "https://example.com",
				StrategyField:   "round_robin",
				MinIntField:     15,
				MaxIntField:     75,
				MinMaxIntField:  25,
				MinStringField:  "validstring",
				MaxStringField:  "short",
			},
			wantErr: false,
		},
		{
			name: "invalid max int",
			config: TestStruct{
				RequiredString:  "test",
				RequiredInt:     42,
				RequiredBool:    true,
				OptionalString:  "valid",
				PortField:       "8080",
				IPField:         "127.0.0.1",
				EmailField:      "test@example.com",
				EthAddressField: "0x1234567890123456789012345678901234567890",
				DurationField:   "30s",
				URLField:        "https://example.com",
				StrategyField:   "round_robin",
				MinIntField:     15,
				MaxIntField:     150,
				MinMaxIntField:  25,
				MinStringField:  "validstring",
				MaxStringField:  "short",
			},
			wantErr: true,
		},
		{
			name: "valid min max int",
			config: TestStruct{
				RequiredString:  "test",
				RequiredInt:     42,
				RequiredBool:    true,
				OptionalString:  "valid",
				PortField:       "8080",
				IPField:         "127.0.0.1",
				EmailField:      "test@example.com",
				EthAddressField: "0x1234567890123456789012345678901234567890",
				DurationField:   "30s",
				URLField:        "https://example.com",
				StrategyField:   "round_robin",
				MinIntField:     15,
				MaxIntField:     75,
				MinMaxIntField:  25,
				MinStringField:  "validstring",
				MaxStringField:  "short",
			},
			wantErr: false,
		},
		{
			name: "invalid min max int - too low",
			config: TestStruct{
				RequiredString:  "test",
				RequiredInt:     42,
				RequiredBool:    true,
				OptionalString:  "valid",
				PortField:       "8080",
				IPField:         "127.0.0.1",
				EmailField:      "test@example.com",
				EthAddressField: "0x1234567890123456789012345678901234567890",
				DurationField:   "30s",
				URLField:        "https://example.com",
				StrategyField:   "round_robin",
				MinIntField:     15,
				MaxIntField:     75,
				MinMaxIntField:  2,
				MinStringField:  "validstring",
				MaxStringField:  "short",
			},
			wantErr: true,
		},
		{
			name: "invalid min max int - too high",
			config: TestStruct{
				RequiredString:  "test",
				RequiredInt:     42,
				RequiredBool:    true,
				OptionalString:  "valid",
				PortField:       "8080",
				IPField:         "127.0.0.1",
				EmailField:      "test@example.com",
				EthAddressField: "0x1234567890123456789012345678901234567890",
				DurationField:   "30s",
				URLField:        "https://example.com",
				StrategyField:   "round_robin",
				MinIntField:     15,
				MaxIntField:     75,
				MinMaxIntField:  75,
				MinStringField:  "validstring",
				MaxStringField:  "short",
			},
			wantErr: true,
		},
		{
			name: "valid min string",
			config: TestStruct{
				RequiredString:  "test",
				RequiredInt:     42,
				RequiredBool:    true,
				OptionalString:  "valid",
				PortField:       "8080",
				IPField:         "127.0.0.1",
				EmailField:      "test@example.com",
				EthAddressField: "0x1234567890123456789012345678901234567890",
				DurationField:   "30s",
				URLField:        "https://example.com",
				StrategyField:   "round_robin",
				MinIntField:     15,
				MaxIntField:     75,
				MinMaxIntField:  25,
				MinStringField:  "validstring",
				MaxStringField:  "short",
			},
			wantErr: false,
		},
		{
			name: "invalid min string",
			config: TestStruct{
				RequiredString:  "test",
				RequiredInt:     42,
				RequiredBool:    true,
				OptionalString:  "valid",
				PortField:       "8080",
				IPField:         "127.0.0.1",
				EmailField:      "test@example.com",
				EthAddressField: "0x1234567890123456789012345678901234567890",
				DurationField:   "30s",
				URLField:        "https://example.com",
				StrategyField:   "round_robin",
				MinIntField:     15,
				MaxIntField:     75,
				MinMaxIntField:  25,
				MinStringField:  "hi",
				MaxStringField:  "short",
			},
			wantErr: true,
		},
		{
			name: "valid max string",
			config: TestStruct{
				RequiredString:  "test",
				RequiredInt:     42,
				RequiredBool:    true,
				OptionalString:  "valid",
				PortField:       "8080",
				IPField:         "127.0.0.1",
				EmailField:      "test@example.com",
				EthAddressField: "0x1234567890123456789012345678901234567890",
				DurationField:   "30s",
				URLField:        "https://example.com",
				StrategyField:   "round_robin",
				MinIntField:     15,
				MaxIntField:     75,
				MinMaxIntField:  25,
				MinStringField:  "validstring",
				MaxStringField:  "short",
			},
			wantErr: false,
		},
		{
			name: "invalid max string",
			config: TestStruct{
				RequiredString:  "test",
				RequiredInt:     42,
				RequiredBool:    true,
				OptionalString:  "valid",
				PortField:       "8080",
				IPField:         "127.0.0.1",
				EmailField:      "test@example.com",
				EthAddressField: "0x1234567890123456789012345678901234567890",
				DurationField:   "30s",
				URLField:        "https://example.com",
				StrategyField:   "round_robin",
				MinIntField:     15,
				MaxIntField:     75,
				MinMaxIntField:  25,
				MinStringField:  "validstring",
				MaxStringField:  "thisiswaytoolong",
			},
			wantErr: true,
		},
	}

	validator := NewValidator()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateConfig(&tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("MinMax validation error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateConfig_NonStruct(t *testing.T) {
	validator := NewValidator()

	// Test with non-struct value
	err := validator.ValidateConfig("not a struct")
	if err == nil {
		t.Error("Expected error for non-struct input")
	}
}

func TestValidateConfig_PointerToStruct(t *testing.T) {
	config := createValidTestStruct()

	validator := NewValidator()
	err := validator.ValidateConfig(&config)
	if err != nil {
		t.Errorf("Validation failed for pointer to struct: %v", err)
	}
}

func TestValidateConfig_UnknownValidationRule(t *testing.T) {
	// Create a struct with unknown validation rule
	type UnknownRuleStruct struct {
		Field string `validate:"unknown_rule"`
	}

	config := UnknownRuleStruct{
		Field: "test",
	}

	validator := NewValidator()
	err := validator.ValidateConfig(&config)
	if err == nil {
		t.Error("Expected error for unknown validation rule")
	}
}

func TestValidateHealthConfig(t *testing.T) {
	config := createValidTestStruct()

	err := ValidateConfig(&config)
	if err != nil {
		t.Errorf("ValidateHealthConfig failed: %v", err)
	}
}

func TestValidateConfig_NestedStruct(t *testing.T) {
	type NestedStruct struct {
		Value string `validate:"required"`
	}

	type ParentStruct struct {
		Nested NestedStruct `validate:"required"`
	}

	config := ParentStruct{
		Nested: NestedStruct{
			Value: "test",
		},
	}

	validator := NewValidator()
	err := validator.ValidateConfig(&config)
	if err != nil {
		t.Errorf("Validation failed for nested struct: %v", err)
	}

	// Test with invalid nested struct
	invalidConfig := ParentStruct{
		Nested: NestedStruct{
			Value: "",
		},
	}

	err = validator.ValidateConfig(&invalidConfig)
	if err == nil {
		t.Error("Expected error for invalid nested struct")
	}
}
