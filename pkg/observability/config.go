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
	if !env.IsValidURL(cfg.OTELExporterEndpoint) {
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
		LogLevel:             "debug",
		SuccessSamplingRate:  0.03, // 3% for success
		ErrorSamplingRate:    1.0,  // 100% for errors
	}
	return cfg
}
