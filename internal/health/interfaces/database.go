package interfaces

import (
	"github.com/trigg3rX/triggerx-backend/internal/health/types"
	commonTypes "github.com/trigg3rX/triggerx-backend/pkg/types"
)

// DatabaseManagerInterface defines the interface for database operations
type DatabaseManagerInterface interface {
	UpdateKeeperHealth(keeperHealth commonTypes.KeeperHealthCheckIn, isActive bool) error
	GetVerifiedKeepers() ([]types.KeeperInfo, error)
}
