package env

import (
	"fmt"
	"os"
	"strconv"
)

// Helper functions
func GetEnv(key, defaultValue string) string {
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
