package env

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Helper functions
func GetEnvString(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	fmt.Printf("Environment variable %s not found, using default value: %s\n", key, defaultValue)
	return defaultValue
}

func GetEnvBool(key string, defaultValue bool) bool {
	if value, exists := os.LookupEnv(key); exists {
		boolValue, err := strconv.ParseBool(value)
		if err != nil {
			fmt.Printf("Environment variable %s not found, using default value: %t\n", key, defaultValue)
			return defaultValue
		}
		return boolValue
	}
	fmt.Printf("Environment variable %s not found, using default value: %t\n", key, defaultValue)
	return defaultValue
}

func GetEnvInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		intValue, err := strconv.Atoi(value)
		if err != nil {
			return defaultValue
		}
		return intValue
	}
	fmt.Printf("Environment variable %s not found, using default value: %d\n", key, defaultValue)
	return defaultValue
}

func GetEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value, exists := os.LookupEnv(key); exists {
		duration, err := time.ParseDuration(value)
		if err != nil {
			return defaultValue
		}
		return duration
	}
	fmt.Printf("Environment variable %s not found, using default value: %v\n", key, defaultValue)
	return defaultValue
}
