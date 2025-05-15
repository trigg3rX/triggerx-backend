package logging

import "sync"

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
func Shutdown() {
	if zl, ok := loggerManager.serviceLogger.(*ZapLogger); ok && zl != nil {
		// Ignore sync errors on shutdown as they're expected for stdout
		_ = zl.logger.Sync()
	}
}
