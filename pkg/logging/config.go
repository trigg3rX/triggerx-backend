package logging

// import "path/filepath"

const (
	BaseDataDir   = "data"
	LogsDir       = "logs"
)

type Environment string

const (
	Development Environment = "development"
	Production  Environment = "production"
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
	Environment      Environment
	UseColors        bool
}

const (
	colorReset   = "\x1b[0m"
	colorRed     = "\x1b[31m"
	colorGreen   = "\x1b[32m"
	colorYellow  = "\x1b[33m"
	colorBlue    = "\x1b[34m"
	colorMagenta = "\x1b[35m"
	colorCyan    = "\x1b[36m"
	colorWhite   = "\x1b[37m"
)