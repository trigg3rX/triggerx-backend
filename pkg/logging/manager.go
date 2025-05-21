package logging

import (
	"fmt"
	"strings"
	"sync"
)

type LoggerManager struct {
	serviceLogger Logger
	once          sync.Once
}

var loggerManager = &LoggerManager{}

func InitServiceLogger(config LoggerConfig) error {
	var err error
	loggerManager.once.Do(func() {
		loggerManager.serviceLogger, err = NewZapLogger(config)
	})
	return err
}

func GetServiceLogger() Logger {
	if loggerManager.serviceLogger == nil {
		panic("logger not initialized")
	}
	return loggerManager.serviceLogger
}

// Shutdown safely cleans up the logger
func Shutdown() error {
	if loggerManager.serviceLogger == nil {
		return nil
	}

	if zl, ok := loggerManager.serviceLogger.(*ZapLogger); ok {
		if err := zl.logger.Sync(); err != nil {
			// Ignore sync errors for stdout
			if !strings.Contains(err.Error(), "sync /dev/stdout") {
				return fmt.Errorf("failed to sync logger during shutdown: %w", err)
			}
		}
	}
	return nil
}
