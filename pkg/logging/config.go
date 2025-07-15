package logging

// import "path/filepath"

const (
	BaseDataDir = "data"
	LogsDir     = "logs"
)

type ProcessName string

const (
	AggregatorProcess         ProcessName = "aggregator"
	DatabaseProcess           ProcessName = "dbserver"
	KeeperProcess             ProcessName = "keeper"
	RegistrarProcess          ProcessName = "registrar"
	HealthProcess             ProcessName = "health"
	RedisProcess              ProcessName = "redis"
	TimeSchedulerProcess      ProcessName = "scheduler-time"
	ConditionSchedulerProcess ProcessName = "scheduler-condition"
	TestProcess               ProcessName = "test"
)

type LoggerConfig struct {
	ProcessName   ProcessName
	IsDevelopment bool
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
