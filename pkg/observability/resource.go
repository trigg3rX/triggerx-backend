package observability

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.27.0"
)

// NewResource creates a new OpenTelemetry resource with service metadata from the provided configuration.
func NewResource(cfg Config) (*resource.Resource, error) {
	attrs := []attribute.KeyValue{
		// Required service attributes
		semconv.ServiceName(string(cfg.ServiceName)),
		semconv.ServiceVersion(cfg.ServiceVersion),
		semconv.ServiceInstanceID(cfg.InstanceID),
	}

	// Create resource with default detection (host, OS, etc.)
	res, err := resource.New(
		context.Background(),
		resource.WithAttributes(attrs...),
	)

	if err != nil {
		return nil, err
	}

	return res, nil
}
