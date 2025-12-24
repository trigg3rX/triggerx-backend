package server

import (
	"context"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/trigg3rX/triggerx-backend/pkg/observability"
	metricspkg "github.com/trigg3rX/triggerx-backend/pkg/rpc/metrics"
)

// LoggingInterceptor provides request/response logging for gRPC
func LoggingInterceptor(logger observability.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()

		logger.Debug(ctx, "gRPC request started",
			observability.String("method", info.FullMethod),
			observability.Any("request", req))

		response, err := handler(ctx, req)

		duration := time.Since(start)

		if err != nil {
			logger.Error(ctx, "gRPC request failed",
				observability.String("method", info.FullMethod),
				observability.Duration("duration", duration),
				observability.Error(err))
		} else {
			logger.Debug(ctx, "gRPC request completed",
				observability.String("method", info.FullMethod),
				observability.Duration("duration", duration),
				observability.Any("response", response))
		}

		return response, err
	}
}

// MetricsInterceptor provides metrics collection for gRPC
func MetricsInterceptor(collector metricspkg.Collector, serviceName string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if collector == nil {
			// No-op if no collector configured
			return handler(ctx, req)
		}

		start := time.Now()
		collector.IncRequestsTotal(serviceName, info.FullMethod)

		response, err := handler(ctx, req)

		duration := time.Since(start)
		collector.ObserveRequestDuration(serviceName, info.FullMethod, duration)
		if err != nil {
			collector.IncErrorsTotal(serviceName, info.FullMethod)
		}

		return response, err
	}
}

// AuthInterceptor provides authentication for gRPC
func AuthInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// TODO: Add JWT authentication logic here
		// For now, just pass through
		return handler(ctx, req)
	}
}

// RecoveryInterceptor provides panic recovery for gRPC
func RecoveryInterceptor(logger observability.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		defer func() {
			if r := recover(); r != nil {
				logger.Error(ctx, "gRPC panic recovered",
					observability.String("method", info.FullMethod),
					observability.Any("panic", r))
				err = status.Errorf(codes.Internal, "internal server error")
			}
		}()
		return handler(ctx, req)
	}
}

// TimeoutInterceptor provides request timeout for gRPC
func TimeoutInterceptor(timeout time.Duration) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		ctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()
		return handler(ctx, req)
	}
}

// RateLimitInterceptor provides rate limiting for gRPC
func RateLimitInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// TODO: Add rate limiting logic here
		// For now, just pass through
		return handler(ctx, req)
	}
}
