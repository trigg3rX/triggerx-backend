package server

import (
	"context"
	"time"

	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	rpcpkg "github.com/trigg3rX/triggerx-backend/pkg/rpc"
	metricspkg "github.com/trigg3rX/triggerx-backend/pkg/rpc/metrics"
)

// LoggingMiddleware provides request/response logging
type LoggingMiddleware struct {
	logger logging.Logger
}

// NewLoggingMiddleware creates a new logging middleware
func NewLoggingMiddleware(logger logging.Logger) *LoggingMiddleware {
	return &LoggingMiddleware{logger: logger}
}

// Process implements the middleware interface
func (m *LoggingMiddleware) Process(ctx context.Context, method string, request interface{}, next rpcpkg.RPCHandler) (interface{}, error) {
	start := time.Now()

	m.logger.Debug("RPC request started",
		"method", method,
		"request", request)

	response, err := next.Handle(ctx, method, request)

	duration := time.Since(start)

	if err != nil {
		m.logger.Error("RPC request failed",
			"method", method,
			"duration", duration,
			"error", err)
	} else {
		m.logger.Debug("RPC request completed",
			"method", method,
			"duration", duration,
			"response", response)
	}

	return response, err
}

// MetricsMiddleware provides metrics collection
type MetricsMiddleware struct {
	collector metricspkg.Collector
	service   string
}

// NewMetricsMiddleware creates a new metrics middleware
func NewMetricsMiddleware(collector metricspkg.Collector, serviceName string) *MetricsMiddleware {
	return &MetricsMiddleware{collector: collector, service: serviceName}
}

// Process implements the middleware interface
func (m *MetricsMiddleware) Process(ctx context.Context, method string, request interface{}, next rpcpkg.RPCHandler) (interface{}, error) {
	if m.collector == nil {
		// No-op if no collector configured
		return next.Handle(ctx, method, request)
	}

	start := time.Now()
	m.collector.IncRequestsTotal(m.service, method)

	response, err := next.Handle(ctx, method, request)

	duration := time.Since(start)
	m.collector.ObserveRequestDuration(m.service, method, duration)
	if err != nil {
		m.collector.IncErrorsTotal(m.service, method)
	}

	return response, err
}

// AuthMiddleware provides authentication
type AuthMiddleware struct {
	// Add auth logic here
}

// NewAuthMiddleware creates a new auth middleware
func NewAuthMiddleware() *AuthMiddleware {
	return &AuthMiddleware{}
}

// Process implements the middleware interface
func (m *AuthMiddleware) Process(ctx context.Context, method string, request interface{}, next rpcpkg.RPCHandler) (interface{}, error) {
	// Add authentication logic here

	return next.Handle(ctx, method, request)
}
