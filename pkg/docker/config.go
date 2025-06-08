package docker

import "github.com/docker/go-units"

type DockerConfig struct {
    Image          string
    TimeoutSeconds int
    AutoCleanup    bool
    MemoryLimit    string
    CPULimit       float64
}

type FeeConfig struct {
    PricePerTG           float64
    FixedCost            float64
    TransactionSimulation float64
    OverheadCost         float64
}

type ExecutorConfig struct {
    Docker DockerConfig
    Fees   FeeConfig
}

func DefaultConfig() ExecutorConfig {
    return ExecutorConfig{
        Docker: DockerConfig{
            Image:          "golang:latest",
            TimeoutSeconds: 600,
            AutoCleanup:    true,
            MemoryLimit:    "1024m",
            CPULimit:       1.0,
        },
        Fees: FeeConfig{
            PricePerTG:           0.0001,
            FixedCost:            1.0,
            TransactionSimulation: 1.0,
            OverheadCost:         0.1,
        },
    }
}

func (c *DockerConfig) MemoryLimitBytes() uint64 {
	memoryLimit, err := units.RAMInBytes(c.MemoryLimit)
	if err != nil {
		return 0
	}
	return uint64(memoryLimit)
}