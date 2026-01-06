package env

// DatabaseConfig contains database connection configuration
type DatabaseConfig struct {
	HostAddress           string
	HostPort              string
	Username              string
	Password              string
	SSLEnabled            bool
	SSLCertPath           string
	SSLKeyPath            string
	SSLCAPath             string
	SSLInsecureSkipVerify bool
}

// GetDatabaseConfig returns database configuration from environment variables
func GetDatabaseConfig() DatabaseConfig {
	return DatabaseConfig{
		HostAddress:           GetEnvString("DATABASE_HOST_ADDRESS", "localhost"),
		HostPort:              GetEnvString("DATABASE_HOST_PORT", "9042"),
		Username:              GetEnvString("DATABASE_USERNAME", ""),
		Password:              GetEnvString("DATABASE_PASSWORD", ""),
		SSLEnabled:            GetEnvBool("DATABASE_SSL_ENABLED", true),
		SSLCertPath:           GetEnvString("DATABASE_SSL_CERT_PATH", ""),
		SSLKeyPath:            GetEnvString("DATABASE_SSL_KEY_PATH", ""),
		SSLCAPath:             GetEnvString("DATABASE_SSL_CA_PATH", ""),
		SSLInsecureSkipVerify: GetEnvBool("DATABASE_SSL_INSECURE_SKIP_VERIFY", true),
	}
}

// GetOTELExporterEndpoint returns the OpenTelemetry exporter endpoint
// Default: "localhost:4318"
func GetOTELExporterEndpoint() string {
	return GetEnvString("OTEL_EXPORTER_ENDPOINT", "localhost:4318")
}
