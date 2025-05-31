package logging

const (
	BaseDataDir   = "data"
	LogsDir       = "logs"
	LogFileFormat = "2006-01-02.log" // for daily files
	TimeFormat    = "2006-01-02 15:04:05"
	MaxFileSize   = 30 * 1024 * 1024 // 30MB in bytes
)

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorPurple = "\033[35m"
)

type LogLevel string

const (
	Development LogLevel = "development"
	Production  LogLevel = "production"
)

type Level int

const (
	DebugLevel Level = iota
	InfoLevel
	WarnLevel
	ErrorLevel
	FatalLevel
)

type ProcessName string

const (
	DatabaseProcess           ProcessName = "database"
	KeeperProcess             ProcessName = "keeper"
	RegistrarProcess          ProcessName = "registrar"
	HealthProcess             ProcessName = "health"
	RedisProcess              ProcessName = "redis"
	TimeSchedulerProcess      ProcessName = "time-scheduler"
	EventSchedulerProcess     ProcessName = "event-scheduler"
	ConditionSchedulerProcess ProcessName = "condition-scheduler"
)

type LoggerConfig struct {
	LogDir           string
	ProcessName      ProcessName
	Environment      LogLevel
	UseColors        bool
	MinStdoutLevel   Level // Minimum level for stdout logging
	MinFileLogLevel  Level // Minimum level for file logging
	CompressOldFiles bool  // Whether to compress old log files
}

func NewDefaultConfig(processName ProcessName) LoggerConfig {
	return LoggerConfig{
		LogDir:           BaseDataDir,
		ProcessName:      processName,
		Environment:      Development,
		UseColors:        true,
		MinStdoutLevel:   DebugLevel, // Development defaults to all levels
		MinFileLogLevel:  DebugLevel, // Always log all levels to file
		CompressOldFiles: true,       // Default to compressing old files
	}
}

// AdjustForProduction modifies the config for production environment
func (c *LoggerConfig) AdjustForProduction() {
	c.Environment = Production
	c.MinStdoutLevel = InfoLevel // Only Info and above in production stdout
}
