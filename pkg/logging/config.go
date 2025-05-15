package logging

const (
	BaseDataDir   = "data"
	LogsDir       = "logs"
	LogFileFormat = "2006-01-02.log" // for daily files
	TimeFormat    = "2006-01-02 15:04:05"
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

type ProcessName string

const (
	ManagerProcess   ProcessName = "manager"
	DatabaseProcess  ProcessName = "database"
	KeeperProcess    ProcessName = "keeper"
	RegistrarProcess ProcessName = "registrar"
	HealthProcess    ProcessName = "health"
)

type LoggerConfig struct {
	LogDir      string
	ProcessName ProcessName
	Environment LogLevel
	UseColors   bool
}

func NewDefaultConfig(processName ProcessName) LoggerConfig {
	return LoggerConfig{
		LogDir:      BaseDataDir,
		ProcessName: processName,
		Environment: Development,
		UseColors:   true,
	}
}
