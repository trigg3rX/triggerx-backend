package config

import (
	"fmt"

	"github.com/joho/godotenv"

	"github.com/trigg3rX/triggerx-backend/pkg/env"
	"github.com/trigg3rX/triggerx-backend/pkg/yaml"
)

type Config struct {
	devMode bool

	// OTel exporter endpoint
	otelExporterEndpoint string

	// Service ID for OpenTelemetry
	serviceID string

	// Aggregator Log Directory
	aggregatorLogDir string

	// event monitor RPC URL
	eventMonitorRPCUrl string
}

var cfg Config

func Init() error {
	// Load secrets from .env file
	if err := godotenv.Load(); err != nil {
		return fmt.Errorf("error loading .env file: %w", err)
	}

	cfg = Config{
		devMode:                      env.GetEnvBool("DEV_MODE", false),
		otelExporterEndpoint:         env.GetOTELExporterEndpoint(),
		serviceID:                    env.GetEnvString("AGGREGATOR_LOG_WATCHER_SERVICE_ID", "1"),
		aggregatorLogDir:             env.GetEnvString("AGGREGATOR_LOG_DIR", "data/logs/aggregator"),
		eventMonitorRPCUrl:           env.GetEnvString("EVENT_MONITOR_RPC_URL", "localhost:9018"),
	}
	if err := validateConfig(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}
	if err := yaml.ValidateConfig(cfg); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}
	return nil
}

func validateConfig() error {
	if !env.IsValidHostPort(cfg.otelExporterEndpoint) {
		return fmt.Errorf("invalid OTEL exporter endpoint: %s (must be a valid host:port, e.g., localhost:4318)", cfg.otelExporterEndpoint)
	}
	return nil
}

func IsDevMode() bool {
	return cfg.devMode
}

func GetOTELExporterEndpoint() string {
	return cfg.otelExporterEndpoint
}

func GetServiceID() string {
	return cfg.serviceID
}

func GetAggregatorLogDir() string {
	return cfg.aggregatorLogDir
}

func GetEventMonitorRPCUrl() string {
	return cfg.eventMonitorRPCUrl
}