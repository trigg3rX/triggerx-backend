package rpc

import (
	"context"
	"time"
)

// ServiceInfo represents information about an RPC service
type ServiceInfo struct {
	Name     string            `json:"name"`
	Version  string            `json:"version"`
	Address  string            `json:"address"`
	Port     int               `json:"port"`
	Metadata map[string]string `json:"metadata,omitempty"`
	LastSeen time.Time         `json:"last_seen"`
	Health   HealthStatus      `json:"health"`
}

// HealthStatus represents the health status of a service
type HealthStatus struct {
	Status    string                 `json:"status"` // "healthy", "unhealthy", "degraded"
	Message   string                 `json:"message,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
	Details   map[string]interface{} `json:"details,omitempty"`
}

// RPCMethod represents an RPC method definition
type RPCMethod struct {
	Name         string        `json:"name"`
	Description  string        `json:"description"`
	RequestType  interface{}   `json:"request_type"`
	ResponseType interface{}   `json:"response_type"`
	Timeout      time.Duration `json:"timeout"`
}

// ServiceRegistry defines the interface for service discovery
type ServiceRegistry interface {
	Register(ctx context.Context, info ServiceInfo) error
	Deregister(ctx context.Context, name string) error
	GetService(ctx context.Context, name string) (*ServiceInfo, error)
	ListServices(ctx context.Context) ([]ServiceInfo, error)
	Watch(ctx context.Context, name string) (<-chan ServiceInfo, error)
}

// RPCHandler defines the interface for RPC method handlers
type RPCHandler interface {
	Handle(ctx context.Context, method string, request interface{}) (interface{}, error)
	GetMethods() []RPCMethod
}

// RPCArgs represents the arguments for an RPC call
type RPCArgs struct {
	Context context.Context `json:"context"`
	Method  string          `json:"method"`
	Request interface{}     `json:"request"`
}

// RPCReply represents the reply from an RPC call
type RPCReply struct {
	Result interface{} `json:"result"`
}

// Middleware defines RPC middleware interface
type Middleware interface {
	Process(ctx context.Context, method string, request interface{}, next RPCHandler) (interface{}, error)
}
