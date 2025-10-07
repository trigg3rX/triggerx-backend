package interfaces

import (
	"context"
	"time"

	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// DatabaseManagerInterface defines the interface for database operations
type DatabaseManagerInterface interface {
	UpdateKeeperStatus(
		ctx context.Context,
		keeperAddress string,
		consensusAddress string,
		version string,
		uptime int64,
		timestamp time.Time,
		publicIP string,
		isActive bool,
	) error
	GetVerifiedKeepers(ctx context.Context) ([]types.HealthKeeperInfo, error)
	UpdateAllKeepersStatus(ctx context.Context, onlineKeepers []types.HealthKeeperInfo) error
	UpdateKeeperChatID(ctx context.Context, keeperAddress string, chatID int64) error
	GetKeeperChatInfo(ctx context.Context, keeperAddress string) (int64, string, error)
}
