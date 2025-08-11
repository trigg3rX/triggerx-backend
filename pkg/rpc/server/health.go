package server

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	rpcproto "github.com/trigg3rX/triggerx-backend/pkg/rpc/proto"
)

// HealthService implements a simple health check service
type HealthService struct {
	rpcproto.UnimplementedGenericServiceServer
	logger      logging.Logger
	serviceName string
	startTime   time.Time
}

// NewHealthService creates a new health service
func NewHealthService(serviceName string, logger logging.Logger) *HealthService {
	return &HealthService{
		logger:      logger,
		serviceName: serviceName,
		startTime:   time.Now(),
	}
}

// HealthCheck performs a health check
func (s *HealthService) HealthCheck(ctx context.Context, req *rpcproto.HealthCheckRequest) (*rpcproto.HealthCheckResponse, error) {
	s.logger.Debug("Health check requested",
		"service", s.serviceName,
		"requested_service", req.Service)

	// Create health status
	healthStatus := &rpcproto.HealthStatus{
		Status:    "healthy",
		Message:   fmt.Sprintf("Service %s is healthy", s.serviceName),
		Timestamp: timestamppb.Now(),
		Details: map[string]string{
			"uptime":     time.Since(s.startTime).String(),
			"start_time": s.startTime.Format(time.RFC3339),
		},
	}

	return &rpcproto.HealthCheckResponse{
		Status: healthStatus,
	}, nil
}

// GetMethods returns available methods
func (s *HealthService) GetMethods(ctx context.Context, _ *emptypb.Empty) (*rpcproto.GetMethodsResponse, error) {
	s.logger.Debug("Get methods requested",
		"service", s.serviceName)

	methods := []*rpcproto.RPCMethod{
		{
			Name:        "HealthCheck",
			Description: "Perform a health check on the service",
		},
	}

	return &rpcproto.GetMethodsResponse{
		Methods: methods,
	}, nil
}
