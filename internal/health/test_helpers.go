package health

import (
	"time"

	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// TestHelpers provides utility functions for testing health module components

// CreateTestKeeperInfo creates a test KeeperInfo with default values
func CreateTestKeeperInfo(address string, overrides ...func(*types.HealthKeeperInfo)) types.HealthKeeperInfo {
	keeper := types.HealthKeeperInfo{
		KeeperName:       "test-keeper-" + address,
		KeeperAddress:    address,
		ConsensusAddress: "0x" + address + "consensus",
		OperatorID:       "op-" + address,
		Version:          "1.0.0",
		PeerID:           "peer-" + address,
		IsActive:         false,
		LastCheckedIn:    time.Now().UTC().Add(-1 * time.Hour),
		IsImua:           false,
	}

	// Apply any overrides
	for _, override := range overrides {
		override(&keeper)
	}

	return keeper
}

// CreateTestKeeperHealthCheckIn creates a test KeeperHealthCheckIn with default values
func CreateTestKeeperHealthCheckIn(address string, overrides ...func(*types.KeeperHealthCheckIn)) types.KeeperHealthCheckIn {
	health := types.KeeperHealthCheckIn{
		KeeperAddress:    address,
		ConsensusPubKey:  "pubkey-" + address,
		ConsensusAddress: "0x" + address + "consensus",
		Version:          "1.0.0",
		Timestamp:        time.Now().UTC(),
		Signature:        "sig-" + address,
		PeerID:           "peer-" + address,
		IsImua:           false,
	}

	// Apply any overrides
	for _, override := range overrides {
		override(&health)
	}

	return health
}

// CreateTestLogger creates a no-op logger for testing
func CreateTestLogger() logging.Logger {
	return logging.NewNoOpLogger()
}

// Common test overrides

// WithActiveKeeper sets the keeper as active
func WithActiveKeeper() func(*types.HealthKeeperInfo) {
	return func(k *types.HealthKeeperInfo) {
		k.IsActive = true
		k.LastCheckedIn = time.Now().UTC()
	}
}

// WithInactiveKeeper sets the keeper as inactive
func WithInactiveKeeper() func(*types.HealthKeeperInfo) {
	return func(k *types.HealthKeeperInfo) {
		k.IsActive = false
		k.LastCheckedIn = time.Now().UTC().Add(-2 * time.Hour)
	}
}

// WithKeeperVersion sets a specific version
func WithKeeperVersion(version string) func(*types.HealthKeeperInfo) {
	return func(k *types.HealthKeeperInfo) {
		k.Version = version
	}
}

// WithImuaKeeper sets the keeper as Imua
func WithImuaKeeper() func(*types.HealthKeeperInfo) {
	return func(k *types.HealthKeeperInfo) {
		k.IsImua = true
	}
}

// WithHealthTimestamp sets a specific timestamp for health check-in
func WithHealthTimestamp(timestamp time.Time) func(*types.KeeperHealthCheckIn) {
	return func(h *types.KeeperHealthCheckIn) {
		h.Timestamp = timestamp
	}
}

// WithHealthVersion sets a specific version for health check-in
func WithHealthVersion(version string) func(*types.KeeperHealthCheckIn) {
	return func(h *types.KeeperHealthCheckIn) {
		h.Version = version
	}
}

// WithHealthImua sets the health check-in as Imua
func WithHealthImua() func(*types.KeeperHealthCheckIn) {
	return func(h *types.KeeperHealthCheckIn) {
		h.IsImua = true
	}
}
