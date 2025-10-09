package rpc

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/trigg3rX/triggerx-backend/internal/health/keeper"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	pb "github.com/trigg3rX/triggerx-backend/pkg/rpc/proto"
	"github.com/trigg3rX/triggerx-backend/pkg/types"
)

// Server implements the Health gRPC service
type Server struct {
	pb.UnimplementedHealthServiceServer
	stateManager *keeper.StateManager
	logger       logging.Logger
	grpcServer   *grpc.Server
	listener     net.Listener
	address      string
	port         string
}

// NewServer creates a new gRPC server for health service
func NewServer(stateManager *keeper.StateManager, logger logging.Logger, address, port string) *Server {
	return &Server{
		stateManager: stateManager,
		logger:       logger.With("component", "grpc_server"),
		address:      address,
		port:         port,
	}
}

// Start starts the gRPC server
func (s *Server) Start() error {
	// Create listener
	addr := fmt.Sprintf("%s:%s", s.address, s.port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to create listener: %w", err)
	}
	s.listener = listener

	// Create gRPC server with interceptors
	s.grpcServer = grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			s.loggingInterceptor(),
		),
	)

	// Register health service
	pb.RegisterHealthServiceServer(s.grpcServer, s)

	// Enable reflection for debugging
	reflection.Register(s.grpcServer)

	s.logger.Info("Starting gRPC server", "address", addr)

	// Start serving
	go func() {
		if err := s.grpcServer.Serve(listener); err != nil {
			s.logger.Error("gRPC server error", "error", err)
		}
	}()

	return nil
}

// Stop gracefully stops the gRPC server
func (s *Server) Stop(ctx context.Context) error {
	s.logger.Info("Stopping gRPC server")

	if s.grpcServer != nil {
		// Graceful shutdown
		stopped := make(chan struct{})
		go func() {
			s.grpcServer.GracefulStop()
			close(stopped)
		}()

		select {
		case <-stopped:
			s.logger.Info("gRPC server stopped gracefully")
		case <-ctx.Done():
			s.grpcServer.Stop() // Force stop
			s.logger.Warn("gRPC server force stopped")
		}
	}

	return nil
}

// GetActivePerformers returns active keepers available for task assignment
func (s *Server) GetActivePerformers(ctx context.Context, req *pb.GetActivePerformersRequest) (*pb.GetActivePerformersResponse, error) {
	s.logger.Debug("GetActivePerformers called",
		"include_imua", req.IncludeImua,
		"limit", req.Limit,
	)

	// Get detailed keeper info
	keepers := s.stateManager.GetDetailedKeeperInfo()

	// Filter for active keepers and convert to performers
	performers := make([]*pb.Performer, 0)
	for _, keeper := range keepers {
		if keeper.IsActive {
			// Apply IMUA filter if requested
			if req.IncludeImua && !keeper.IsImua {
				continue
			}

			operatorID, _ := strconv.ParseInt(keeper.OperatorID, 10, 64)
			performer := &pb.Performer{
				OperatorId:    operatorID,
				KeeperAddress: keeper.KeeperAddress,
				IsImua:        keeper.IsImua,
				Version:       keeper.Version,
				UptimeSeconds: keeper.Uptime,
				LastSeen:      timestamppb.New(keeper.LastCheckedIn),
			}
			performers = append(performers, performer)

			// Apply limit if specified
			if req.Limit > 0 && len(performers) >= int(req.Limit) {
				break
			}
		}
	}

	return &pb.GetActivePerformersResponse{
		Performers: performers,
		Count:      int32(len(performers)),
		Timestamp:  timestamppb.Now(),
	}, nil
}

// GetNextPerformer returns the next keeper using load balancing
func (s *Server) GetNextPerformer(ctx context.Context, req *pb.GetNextPerformerRequest) (*pb.GetNextPerformerResponse, error) {
	s.logger.Debug("GetNextPerformer called",
		"is_imua_task", req.IsImuaTask,
		"task_type", req.TaskType,
	)

	// Get detailed keeper info
	keepers := s.stateManager.GetDetailedKeeperInfo()

	// Filter for active keepers matching criteria
	var activeKeepers []types.HealthKeeperInfo
	for _, keeper := range keepers {
		if keeper.IsActive {
			// Filter by IMUA requirement
			if req.IsImuaTask && !keeper.IsImua {
				continue
			}
			activeKeepers = append(activeKeepers, keeper)
		}
	}

	if len(activeKeepers) == 0 {
		return nil, fmt.Errorf("no active keepers available")
	}

	// Simple round-robin selection (can be enhanced with load balancing strategies)
	// For now, return the first active keeper
	selected := activeKeepers[0]
	operatorID, _ := strconv.ParseInt(selected.OperatorID, 10, 64)

	return &pb.GetNextPerformerResponse{
		Performer: &pb.Performer{
			OperatorId:    operatorID,
			KeeperAddress: selected.KeeperAddress,
			IsImua:        selected.IsImua,
			Version:       selected.Version,
			UptimeSeconds: selected.Uptime,
			LastSeen:      timestamppb.New(selected.LastCheckedIn),
		},
		Strategy: "round_robin",
	}, nil
}

// GetKeeperStatus returns the status of keepers
func (s *Server) GetKeeperStatus(ctx context.Context, req *pb.GetKeeperStatusRequest) (*pb.GetKeeperStatusResponse, error) {
	s.logger.Debug("GetKeeperStatus called",
		"keeper_addresses", req.KeeperAddresses,
	)

	total, active := s.stateManager.GetKeeperCount()
	keepers := s.stateManager.GetDetailedKeeperInfo()

	// Filter if specific keeper addresses requested
	var filteredKeepers []types.HealthKeeperInfo
	if len(req.KeeperAddresses) > 0 {
		addressMap := make(map[string]bool)
		for _, addr := range req.KeeperAddresses {
			addressMap[addr] = true
		}
		for _, keeper := range keepers {
			if addressMap[keeper.KeeperAddress] {
				filteredKeepers = append(filteredKeepers, keeper)
			}
		}
	} else {
		filteredKeepers = keepers
	}

	// Convert to proto message
	protoKeepers := make([]*pb.KeeperInfo, len(filteredKeepers))
	for i, keeper := range filteredKeepers {
		protoKeepers[i] = &pb.KeeperInfo{
			KeeperName:       keeper.KeeperName,
			KeeperAddress:    keeper.KeeperAddress,
			ConsensusAddress: keeper.ConsensusAddress,
			OperatorId:       keeper.OperatorID,
			Version:          keeper.Version,
			IsActive:         keeper.IsActive,
			IsImua:           keeper.IsImua,
			UptimeSeconds:    keeper.Uptime,
			LastCheckedIn:    timestamppb.New(keeper.LastCheckedIn),
		}
	}

	return &pb.GetKeeperStatusResponse{
		Keepers:       protoKeepers,
		TotalKeepers:  int32(total),
		ActiveKeepers: int32(active),
		Timestamp:     timestamppb.Now(),
	}, nil
}

// loggingInterceptor logs all gRPC requests
func (s *Server) loggingInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()

		resp, err := handler(ctx, req)

		duration := time.Since(start)
		s.logger.Debug("gRPC request",
			"method", info.FullMethod,
			"duration_ms", duration.Milliseconds(),
			"error", err,
		)

		return resp, err
	}
}
