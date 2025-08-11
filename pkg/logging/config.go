package logging

import (
	"os"
	"path/filepath"
)

var (
	BaseDataDir = getBaseDataDir()
	LogsDir     = "logs"
)

type ProcessName string

const (
	AggregatorProcess         ProcessName = "aggregator"
	DatabaseProcess           ProcessName = "dbserver"
	KeeperProcess             ProcessName = "keeper"
	RegistrarProcess          ProcessName = "registrar"
	HealthProcess             ProcessName = "health"
	TaskDispatcherProcess     ProcessName = "taskdispatcher"
	TaskMonitorProcess        ProcessName = "taskmonitor"
	TimeSchedulerProcess      ProcessName = "schedulers-time"
	ConditionSchedulerProcess ProcessName = "schedulers-condition"
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
	colorWhite   = "\x1b[37m"
)

func getBaseDataDir() string {
	if dataDir := os.Getenv("TRIGGERX_DATA_DIR"); dataDir != "" {
		return dataDir
	}

	currentDir, err := os.Getwd()
	if err != nil {
		return "data"
	}
	
	for {
		if _, err := os.Stat(filepath.Join(currentDir, "go.mod")); err == nil {
			return filepath.Join(currentDir, "data")
		}
		
		parent := filepath.Dir(currentDir)
		if parent == currentDir {
			break
		}
		currentDir = parent
	}
	return "data"
}
