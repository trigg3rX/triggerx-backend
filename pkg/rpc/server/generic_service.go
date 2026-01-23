package server

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/trigg3rX/triggerx-backend/pkg/observability"
	rpcpkg "github.com/trigg3rX/triggerx-backend/pkg/rpc"
	rpcproto "github.com/trigg3rX/triggerx-backend/pkg/rpc/proto"
)

// GenericService implements the proto.GenericServiceServer interface
type GenericService struct {
	rpcproto.UnimplementedGenericServiceServer
	serviceName string
	handler     rpcpkg.RPCHandler
	logger      observability.Logger
}

// NewGenericService creates a new GenericService
func NewGenericService(serviceName string, handler rpcpkg.RPCHandler, logger observability.Logger) *GenericService {
	return &GenericService{
		serviceName: serviceName,
		handler:     handler,
		logger:      logger,
	}
}

// Call handles generic RPC calls
func (s *GenericService) Call(ctx context.Context, req *rpcproto.RPCRequest) (*rpcproto.RPCResponse, error) {
	s.logger.Debug(ctx, "gRPC call received", observability.String("service", s.serviceName), observability.String("method", req.Method))

	// Extract request payload
	var request interface{}
	if req.Payload != nil {
		if req.Payload.TypeUrl == "application/json" {
			// Get the expected request type from the handler's method definition
			methods := s.handler.GetMethods()
			var requestType interface{}
			for _, method := range methods {
				if method.Name == req.Method {
					requestType = method.RequestType
					break
				}
			}

			if requestType != nil {
				// Create a new instance of the request type using reflection
				requestTypeValue := reflect.ValueOf(requestType)
				if requestTypeValue.Kind() == reflect.Ptr {
					// If it's a pointer, get the element type and create a new instance
					elemType := requestTypeValue.Elem().Type()
					newValue := reflect.New(elemType)

					// Unmarshal JSON into the new instance
					if err := json.Unmarshal(req.Payload.Value, newValue.Interface()); err != nil {
						s.logger.Error(ctx, "Failed to deserialize JSON payload into expected type", observability.String("service", s.serviceName), observability.String("method", req.Method), observability.Error(err))
						return &rpcproto.RPCResponse{
							Error: fmt.Sprintf("failed to deserialize JSON payload: %v", err),
						}, nil
					}
					request = newValue.Interface()
				} else {
					// If it's not a pointer, create a new instance directly
					newValue := reflect.New(requestTypeValue.Type())
					if err := json.Unmarshal(req.Payload.Value, newValue.Interface()); err != nil {
						s.logger.Error(ctx, "Failed to deserialize JSON payload into expected type", observability.String("service", s.serviceName), observability.String("method", req.Method), observability.Error(err))
						return &rpcproto.RPCResponse{
							Error: fmt.Sprintf("failed to deserialize JSON payload: %v", err),
						}, nil
					}
					request = newValue.Elem().Interface()
				}
			} else {
				// No method definition found, deserialize as map
				var jsonData map[string]interface{}
				if err := json.Unmarshal(req.Payload.Value, &jsonData); err != nil {
					s.logger.Error(ctx, "Failed to deserialize JSON payload", observability.String("service", s.serviceName), observability.String("method", req.Method), observability.Error(err))
					return &rpcproto.RPCResponse{
						Error: fmt.Sprintf("failed to deserialize JSON payload: %v", err),
					}, nil
				}
				request = jsonData
			}
		} else {
			// For protobuf messages, pass as-is
			request = req.Payload
		}
	}

	// Call the handler
	result, err := s.handler.Handle(ctx, req.Method, request)
	if err != nil {
		s.logger.Error(ctx, "gRPC call failed", observability.String("service", s.serviceName), observability.String("method", req.Method), observability.Error(err))

		return &rpcproto.RPCResponse{
			Error: err.Error(),
		}, nil
	}

	// Convert result to Any
	var resultAny *anypb.Any
	if result != nil {
		if protoMsg, ok := result.(proto.Message); ok {
			var err error
			resultAny, err = anypb.New(protoMsg)
			if err != nil {
				s.logger.Error(ctx, "Failed to convert result to Any", observability.String("service", s.serviceName), observability.String("method", req.Method), observability.Error(err))
				return &rpcproto.RPCResponse{
					Error: fmt.Sprintf("failed to serialize result: %v", err),
				}, nil
			}
		} else {
			// For non-proto messages, serialize as JSON
			jsonData, err := json.Marshal(result)
			if err != nil {
				s.logger.Error(ctx, "Failed to serialize result as JSON", observability.String("service", s.serviceName), observability.String("method", req.Method), observability.Error(err))
				return &rpcproto.RPCResponse{
					Error: fmt.Sprintf("failed to serialize result as JSON: %v", err),
				}, nil
			}
			resultAny = &anypb.Any{
				TypeUrl: "application/json",
				Value:   jsonData,
			}
		}
	}

	s.logger.Debug(ctx, "gRPC call completed", observability.String("service", s.serviceName), observability.String("method", req.Method))

	return &rpcproto.RPCResponse{
		Result: resultAny,
	}, nil
}

// HealthCheck performs a health check
func (s *GenericService) HealthCheck(ctx context.Context, req *rpcproto.HealthCheckRequest) (*rpcproto.HealthCheckResponse, error) {
	s.logger.Debug(ctx, "Health check requested", observability.String("service", s.serviceName))

	// Create a simple health status
	healthStatus := &rpcproto.HealthStatus{
		Status:  "healthy",
		Message: fmt.Sprintf("Service %s is healthy", s.serviceName),
	}

	return &rpcproto.HealthCheckResponse{
		Status: healthStatus,
	}, nil
}

// GetMethods returns available methods
func (s *GenericService) GetMethods(ctx context.Context, _ *emptypb.Empty) (*rpcproto.GetMethodsResponse, error) {
	s.logger.Debug(ctx, "Get methods requested", observability.String("service", s.serviceName))

	methods := s.handler.GetMethods()
	protoMethods := make([]*rpcproto.RPCMethod, len(methods))

	for i, method := range methods {
		protoMethods[i] = &rpcproto.RPCMethod{
			Name:         method.Name,
			Description:  method.Description,
			RequestType:  fmt.Sprintf("%T", method.RequestType),
			ResponseType: fmt.Sprintf("%T", method.ResponseType),
			TimeoutNanos: int64(method.Timeout),
		}
	}

	return &rpcproto.GetMethodsResponse{
		Methods: protoMethods,
	}, nil
}
