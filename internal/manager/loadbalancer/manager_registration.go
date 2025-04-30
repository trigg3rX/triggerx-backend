package loadbalancer

import (
	"context"
	"time"

	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

type ManagerRegistration struct {
	managerID    string
	address      string
	loadBalancer *LoadBalancer
	logger       logging.Logger
}

func NewManagerRegistration(managerID, address string, lb *LoadBalancer) *ManagerRegistration {
	return &ManagerRegistration{
		managerID:    managerID,
		address:      address,
		loadBalancer: lb,
		logger:       logging.GetLogger(logging.Development, logging.ManagerProcess),
	}
}

func (mr *ManagerRegistration) Register() error {
	// Register with load balancer
	if err := mr.loadBalancer.AddManager(mr.managerID, mr.address); err != nil {
		mr.logger.Errorf("Failed to register manager: %v", err)
		return err
	}

	mr.logger.Infof("Manager %s registered successfully", mr.managerID)
	return nil
}

func (mr *ManagerRegistration) StartHeartbeat(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			mr.sendHeartbeat()
		}
	}
}

func (mr *ManagerRegistration) sendHeartbeat() {
	// Update heartbeat timestamp
	if err := mr.loadBalancer.UpdateManagerHeartbeat(mr.managerID); err != nil {
		mr.logger.Errorf("Failed to send heartbeat: %v", err)
		return
	}

	mr.logger.Debugf("Heartbeat sent for manager %s", mr.managerID)
}

func (mr *ManagerRegistration) Unregister() error {
	// Remove from load balancer
	if err := mr.loadBalancer.RemoveManager(mr.managerID); err != nil {
		mr.logger.Errorf("Failed to unregister manager: %v", err)
		return err
	}

	mr.logger.Infof("Manager %s unregistered successfully", mr.managerID)
	return nil
}
