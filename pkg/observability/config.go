package observability

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"time"

	"github.com/trigg3rX/triggerx-backend/pkg/env"
)

var (
	BaseDataDir = getBaseDataDir()
	InstanceID  = generateInstanceID()
)

// ServiceName represents a service identifier
type ServiceName string

const (
	// AggregatorService         ServiceName = "aggregator"
	ServerService             ServiceName = "api-server"
	HealthService             ServiceName = "health"
	TimeSchedulerService      ServiceName = "time-scheduler"
	ConditionSchedulerService ServiceName = "condition-scheduler"
	EventMonitorService       ServiceName = "event-monitor"
	TaskDispatcherService     ServiceName = "task-dispatcher"
	TaskMonitorService        ServiceName = "task-monitor"
	KeeperService             ServiceName = "keeper"
	TestService               ServiceName = "test123"
)

// Config represents the unified observability configuration
type Config struct {
	// Service metadata (resource attributes)
	ServiceName    ServiceName
	ServiceVersion string
	InstanceID     string

	// OTel collector endpoint
	OTELExporterEndpoint string

	DevMode  bool
	LogLevel string

	// Export settings
	BatchTimeout   time.Duration
	ExportTimeout  time.Duration
	MaxExportBatch int

	// Prometheus export settings
	// EnablePrometheusExport: if true, exposes metrics endpoint for Prometheus scraping
	EnablePrometheusExport bool

	// Trace sampling settings
	// SuccessSamplingRate: sampling rate for successful traces (0.0 to 1.0)
	// ErrorSamplingRate: sampling rate for traces with errors (0.0 to 1.0)
	// Defaults: 0.03 (3%) for success, 1.0 (100%) for errors
	SuccessSamplingRate float64
	ErrorSamplingRate   float64
}

func validateConfig(cfg Config) error {
	if env.IsEmpty(string(cfg.ServiceName)) {
		return fmt.Errorf("ServiceName is required")
	}
	if env.IsEmpty(cfg.ServiceVersion) {
		return fmt.Errorf("ServiceVersion is required")
	}
	if env.IsEmpty(cfg.InstanceID) {
		return fmt.Errorf("InstanceID is required")
	}
	if !env.IsValidHostPort(cfg.OTELExporterEndpoint) {
		return fmt.Errorf("invalid OTELExporterEndpoint: %s", cfg.OTELExporterEndpoint)
	}
	return nil
}

// generateInstanceID generates a unique instance ID
func generateInstanceID() string {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}
	pid := os.Getpid()
	return fmt.Sprintf("%s-%d", hostname, pid)
}

func getBaseDataDir() string {
	if dataDir := os.Getenv("TRIGGERX_DATA_DIR"); dataDir != "" {
		return dataDir
	}

	currentUser, err := user.Current()
	if err != nil {
		return "data"
	}
	baseDataDir := filepath.Join(currentUser.HomeDir, ".triggerx", "data")
	if _, err := os.Stat(baseDataDir); os.IsNotExist(err) {
		err = os.MkdirAll(baseDataDir, 0755)
		if err != nil {
			return "data"
		}
	}
	return baseDataDir
}

// NewConfig creates a new Config with the provided service information and sensible defaults.
// This is the recommended way to create observability configuration for services.
//
// Parameters:
//   - serviceName: The name of the service (use ServiceName constants)
//   - serviceVersion: The version of the service (e.g., "1.0.0")
//   - otelEndpoint: The OTLP exporter endpoint (e.g., "http://localhost:4318")
//   - devMode: Whether the service is running in development mode
//
// Returns a Config with default values for:
//   - BatchTimeout: 5 seconds
//   - ExportTimeout: 30 seconds
//   - MaxExportBatch: 512
//   - LogLevel: "info" (or "debug" if devMode is true)
//   - SuccessSamplingRate: 0.03 (3%)
//   - ErrorSamplingRate: 1.0 (100%)
func NewConfig(serviceName ServiceName, serviceVersion, otelEndpoint string, devMode bool) Config {
	logLevel := "info"
	if devMode {
		logLevel = "debug"
	}

	return Config{
		ServiceName:            serviceName,
		ServiceVersion:         serviceVersion,
		InstanceID:             InstanceID,
		OTELExporterEndpoint:   otelEndpoint,
		DevMode:                devMode,
		LogLevel:               logLevel,
		BatchTimeout:           5 * time.Second,
		ExportTimeout:          30 * time.Second,
		MaxExportBatch:         512,
		EnablePrometheusExport: false, // Default: disabled
		SuccessSamplingRate:    0.03,  // 3% for success
		ErrorSamplingRate:      1.0,   // 100% for errors
	}
}

// NewConfigWithOptions creates a new Config with custom options.
// Use this when you need to override default values.
func NewConfigWithOptions(serviceName ServiceName, serviceVersion, otelEndpoint string, devMode bool, opts ...ConfigOption) Config {
	cfg := NewConfig(serviceName, serviceVersion, otelEndpoint, devMode)
	for _, opt := range opts {
		opt(&cfg)
	}
	return cfg
}

// ConfigOption is a function that modifies a Config
type ConfigOption func(*Config)

// WithBatchTimeout sets the batch timeout for exports
func WithBatchTimeout(timeout time.Duration) ConfigOption {
	return func(cfg *Config) {
		cfg.BatchTimeout = timeout
	}
}

// WithExportTimeout sets the export timeout
func WithExportTimeout(timeout time.Duration) ConfigOption {
	return func(cfg *Config) {
		cfg.ExportTimeout = timeout
	}
}

// WithMaxExportBatch sets the maximum export batch size
func WithMaxExportBatch(size int) ConfigOption {
	return func(cfg *Config) {
		cfg.MaxExportBatch = size
	}
}

// WithLogLevel sets the log level
func WithLogLevel(level string) ConfigOption {
	return func(cfg *Config) {
		cfg.LogLevel = level
	}
}

// WithSamplingRates sets the trace sampling rates
func WithSamplingRates(successRate, errorRate float64) ConfigOption {
	return func(cfg *Config) {
		cfg.SuccessSamplingRate = successRate
		cfg.ErrorSamplingRate = errorRate
	}
}

// WithPrometheusExport enables Prometheus metrics export and sets the metrics path
func WithPrometheusExport(enabled bool) ConfigOption {
	return func(cfg *Config) {
		cfg.EnablePrometheusExport = enabled
	}
}

// SetTestConfig returns a Config with sensible defaults from environment variables
func SetTestConfig() Config {
	cfg := Config{
		ServiceName:          TestService,
		ServiceVersion:       "0.1.0",
		InstanceID:           InstanceID,
		OTELExporterEndpoint: "http://localhost:4318",
		BatchTimeout:         5 * time.Second,
		ExportTimeout:        30 * time.Second,
		MaxExportBatch:       512,
		DevMode:              true,
		EnablePrometheusExport: false,
		LogLevel:             "debug",
		SuccessSamplingRate:  0.03, // 3% for success
		ErrorSamplingRate:    1.0,  // 100% for errors
	}
	return cfg
}
