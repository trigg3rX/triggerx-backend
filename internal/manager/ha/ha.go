package ha

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/trigg3rX/triggerx-backend/internal/manager/config"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

const (
	// Lock constants
	leaderLockKey    = "triggerx:manager:leader"
	leaderLockExpiry = 10 * time.Second
	leaderLockTTL    = 5 * time.Second

	// HA roles
	RoleLeader   = "leader"
	RoleFollower = "follower"
)

var (
	logger          logging.Logger
	currentRole     string
	instanceID      string
	roleChangeMutex sync.RWMutex
	redisClient     *redis.Client
	leaderCtx       context.Context
	leaderCancel    context.CancelFunc
)

// Manager high availability configuration
type HAConfig struct {
	// Redis connection for distributed locking
	RedisAddress  string
	RedisPassword string

	// Other managers in cluster for health checks
	OtherManagerAddresses []string

	// Role change callback
	OnRoleChange func(newRole string)
}

// Initialize high availability mode
func Init(haConfig *HAConfig) {
	logger = logging.GetLogger(logging.Development, logging.ManagerProcess)

	// Generate unique instance ID using hostname as fallback
	hostname, err := os.Hostname()
	if err != nil {
		hostname = fmt.Sprintf("manager-%d", time.Now().UnixNano())
	}
	instanceID = fmt.Sprintf("%s:%s", hostname, config.ManagerRPCPort)

	logger.Infof("Starting manager in HA mode with instance ID: %s", instanceID)
	logger.Infof("Redis address: %s", haConfig.RedisAddress)

	// Initialize Redis client for distributed locking
	redisClient = redis.NewClient(&redis.Options{
		Addr:     haConfig.RedisAddress,
		Password: haConfig.RedisPassword,
		DB:       0,
		Network:  "tcp",
	})

	// Test Redis connection
	ctx := context.Background()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		logger.Errorf("Failed to connect to Redis at %s: %v", haConfig.RedisAddress, err)
		panic(fmt.Sprintf("Failed to connect to Redis: %v", err))
	}

	logger.Infof("Redis client initialized with address: %s", haConfig.RedisAddress)
	// Start role election
	go startElection(haConfig)
}

// Start leader election process
func startElection(haConfig *HAConfig) {
	ctx := context.Background()

	for {
		// Try to acquire the leader lock
		acquired, err := tryAcquireLeaderLock(ctx)

		if err != nil {
			logger.Errorf("Error in leader election: %v", err)
			logger.Errorf("Redis address being used: %s", haConfig.RedisAddress)
			setRole(RoleFollower, haConfig.OnRoleChange)
			time.Sleep(1 * time.Second)
			continue
		}

		if acquired {
			// We're the leader now
			if getCurrentRole() != RoleLeader {
				logger.Info("This manager instance is now the LEADER")
				setRole(RoleLeader, haConfig.OnRoleChange)

				// Create new context for leader operations
				leaderCtx, leaderCancel = context.WithCancel(ctx)

				// Start leader heartbeat to maintain lock
				go runLeaderHeartbeat(leaderCtx, haConfig)
			}
		} else {
			// We're a follower now
			if getCurrentRole() != RoleFollower {
				logger.Info("This manager instance is now a FOLLOWER")
				setRole(RoleFollower, haConfig.OnRoleChange)

				// Cancel leader context if we were previously leader
				if leaderCancel != nil {
					leaderCancel()
				}
			}
		}

		// Check again after some time
		time.Sleep(1 * time.Second)
	}
}

// Run leader heartbeat to maintain the lock
func runLeaderHeartbeat(ctx context.Context, haConfig *HAConfig) {
	ticker := time.NewTicker(leaderLockTTL / 2) // Use half the TTL for more frequent updates
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Check if we still hold the lock with our instance ID
			val, err := redisClient.Get(ctx, leaderLockKey).Result()
			if err != nil && err != redis.Nil {
				logger.Errorf("Error checking leader lock: %v", err)
				continue
			}

			// If the lock doesn't exist or has a different value, try to reacquire
			if err == redis.Nil || val != instanceID {
				logger.Warn("Leader lock lost or acquired by another instance")
				setRole(RoleFollower, haConfig.OnRoleChange)
				return
			}

			// We still own the lock, extend it
			success, err := redisClient.Expire(ctx, leaderLockKey, leaderLockExpiry).Result()
			if err != nil || !success {
				logger.Error("Failed to refresh leader lock, stepping down")
				setRole(RoleFollower, haConfig.OnRoleChange)
				return
			}

		case <-ctx.Done():
			logger.Info("Leader heartbeat stopped")
			return
		}
	}
}

// Try to acquire the leader lock
func tryAcquireLeaderLock(ctx context.Context) (bool, error) {
	// First check if lock exists
	val, err := redisClient.Get(ctx, leaderLockKey).Result()

	// If there's no error and the value matches our instance ID, we already have the lock
	if err == nil && val == instanceID {
		// Refresh the lock
		_, err = redisClient.Expire(ctx, leaderLockKey, leaderLockExpiry).Result()
		return true, err
	}

	// If the key doesn't exist, try to acquire the lock
	if err == redis.Nil {
		return redisClient.SetNX(ctx, leaderLockKey, instanceID, leaderLockExpiry).Result()
	}

	// Otherwise there was an error or someone else has the lock
	return false, err
}

// Get the current role (leader or follower)
func getCurrentRole() string {
	roleChangeMutex.RLock()
	defer roleChangeMutex.RUnlock()
	return currentRole
}

// Set the current role
func setRole(role string, onRoleChange func(string)) {
	roleChangeMutex.Lock()
	currentRole = role
	roleChangeMutex.Unlock()

	if onRoleChange != nil {
		onRoleChange(role)
	}
}

// IsLeader returns whether this instance is currently the leader
func IsLeader() bool {
	return getCurrentRole() == RoleLeader
}

// GetInstanceID returns this manager's unique instance ID
func GetInstanceID() string {
	return instanceID
}
