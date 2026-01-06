package yaml

// DatabaseOperationsConfig contains database operations configuration
type DatabaseOperationsConfig struct {
	Timeout     Duration `yaml:"timeout"`
	ConnectWait Duration `yaml:"connect_wait"`
	Retries     int      `yaml:"retries"`
}

// MetricsConfig contains metrics configuration
type MetricsConfig struct {
	UpdateInterval Duration `yaml:"update_interval"`
}

// ShutdownConfig contains shutdown configuration
type ShutdownConfig struct {
	Timeout Duration `yaml:"timeout"`
}

// VersionConfig contains version configuration
type VersionConfig struct {
	Version string `yaml:"version"`
}
