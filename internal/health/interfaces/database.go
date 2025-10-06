package interfaces

import (
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// DatabaseManagerInterface defines the interface for database operations
type DatabaseManagerInterface interface {
	UpdateKeeperHealth(keeperHealth types.KeeperHealthCheckIn, isActive bool) error
	GetVerifiedKeepers() ([]types.HealthKeeperInfo, error)
}
