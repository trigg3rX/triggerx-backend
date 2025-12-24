package observability

import (
	"context"
	"fmt"

	sdklog "go.opentelemetry.io/otel/sdk/log"
)

// MultiExporter forwards log records to multiple exporters
type MultiExporter struct {
	exporters []sdklog.Exporter
}

// NewMultiExporter creates a new multi-exporter that forwards to all provided exporters
func NewMultiExporter(exporters ...sdklog.Exporter) *MultiExporter {
	return &MultiExporter{
		exporters: exporters,
	}
}

// Export exports log records to all exporters
// If any exporter fails, the error is collected and returned
func (e *MultiExporter) Export(ctx context.Context, records []sdklog.Record) error {
	var errs []error
	for _, exporter := range e.exporters {
		if err := exporter.Export(ctx, records); err != nil {
			errs = append(errs, fmt.Errorf("exporter error: %w", err))
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("multi-exporter errors: %v", errs)
	}
	return nil
}

// Shutdown shuts down all exporters
func (e *MultiExporter) Shutdown(ctx context.Context) error {
	var errs []error
	for _, exporter := range e.exporters {
		if err := exporter.Shutdown(ctx); err != nil {
			errs = append(errs, fmt.Errorf("exporter shutdown error: %w", err))
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("multi-exporter shutdown errors: %v", errs)
	}
	return nil
}

// ForceFlush forces a flush of all exporters
func (e *MultiExporter) ForceFlush(ctx context.Context) error {
	var errs []error
	for _, exporter := range e.exporters {
		if err := exporter.ForceFlush(ctx); err != nil {
			errs = append(errs, fmt.Errorf("exporter force flush error: %w", err))
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("multi-exporter force flush errors: %v", errs)
	}
	return nil
}

// Ensure MultiExporter implements the sdklog.Exporter interface
var _ sdklog.Exporter = (*MultiExporter)(nil)
