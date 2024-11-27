package config

import (
    "log"
    "os"

    "github.com/trigg3rX/go-backend/pkg/types"
)

type KeeperConfig struct {
    Name            string
    WorkerCount     int
    MaxConcurrentJobs int
    LogDirectory    string
    JobTypes        []string
}

func LoadConfig() *KeeperConfig {
    // Default configuration
    config := &KeeperConfig{
        Name:            "keeper-default",
        WorkerCount:     4,
        MaxConcurrentJobs: 10,
        LogDirectory:    "./logs",
        JobTypes:        []string{"default"},
    }

    // Create log directory if it doesn't exist
    if err := os.MkdirAll(config.LogDirectory, 0755); err != nil {
        log.Printf("Failed to create log directory: %v", err)
    }

    return config
}

// ValidateJob checks if the job is compatible with this keeper
func (c *KeeperConfig) ValidateJob(job *types.Job) bool {
    // Simple job type validation
    for _, supportedType := range c.JobTypes {
        if supportedType == "default" || supportedType == job.ArgType {
            return true
        }
    }
    return false
}