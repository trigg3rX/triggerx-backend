package observability

import (
	"context"
	"fmt"
	"os"
	"time"

	"go.opentelemetry.io/otel/attribute"
	otlploghttp "go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	otellog "go.opentelemetry.io/otel/log"
	"go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/trace"
)

// otelLogger wraps the OpenTelemetry logger implementation
type otelLogger struct {
	logger         otellog.Logger
	loggerProvider *log.LoggerProvider
	baseFields     []Field
	cfg            Config
	devMode        bool
}

// NewLogger creates a new logger instance with the provided configuration and resource
func NewLogger(cfg Config, res *resource.Resource) (Logger, func(context.Context) error, error) {
	var exporters []log.Exporter

	if err := validateConfig(cfg); err != nil {
		return nil, nil, fmt.Errorf("invalid configuration: %w", err)
	}

	// Always add console exporter for INFO level and above
	// DEBUG level will be filtered based on DevMode flag
	consoleExporter := NewConsoleExporter(cfg.DevMode)
	exporters = append(exporters, consoleExporter)

	otlpExporter, err := otlploghttp.New(
		context.Background(),
		otlploghttp.WithEndpoint(cfg.OTELExporterEndpoint),
		otlploghttp.WithInsecure(), // TODO: make this configurable
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create OTLP log exporter: %w", err)
	}
	exporters = append(exporters, otlpExporter)

	// Create processor
	var processor log.Processor
	if len(exporters) == 1 {
		processor = log.NewSimpleProcessor(exporters[0])
	} else if len(exporters) > 1 {
		// Create a multi-exporter that forwards to all exporters
		multiExporter := NewMultiExporter(exporters...)
		// Use batch processor with the multi-exporter
		processor = log.NewBatchProcessor(
			multiExporter,
			log.WithExportInterval(cfg.BatchTimeout),
			log.WithExportTimeout(cfg.ExportTimeout),
			log.WithExportMaxBatchSize(cfg.MaxExportBatch),
		)
	} else {
		// No exporters - use console exporter as fallback
		processor = log.NewSimpleProcessor(NewConsoleExporter(cfg.DevMode))
	}

	// Create logger provider
	loggerProvider := log.NewLoggerProvider(
		log.WithResource(res),
		log.WithProcessor(processor),
	)

	// Create logger from provider
	apiLogger := loggerProvider.Logger(
		string(cfg.ServiceName),
		otellog.WithSchemaURL("https://opentelemetry.io/schemas/1.27.0"),
	)

	shutdown := func(ctx context.Context) error {
		return loggerProvider.Shutdown(ctx)
	}

	return &otelLogger{
		logger:         apiLogger,
		loggerProvider: loggerProvider,
		cfg:            cfg,
		devMode:        cfg.DevMode,
	}, shutdown, nil
}

// Debug logs a debug message
func (l *otelLogger) Debug(ctx context.Context, msg string, fields ...Field) {
	// Only log DEBUG level if DevMode is enabled
	if !l.devMode {
		return
	}
	l.log(ctx, otellog.SeverityDebug1, msg, fields...)
}

// Info logs an info message
func (l *otelLogger) Info(ctx context.Context, msg string, fields ...Field) {
	l.log(ctx, otellog.SeverityInfo1, msg, fields...)
}

// Warn logs a warning message
func (l *otelLogger) Warn(ctx context.Context, msg string, fields ...Field) {
	l.log(ctx, otellog.SeverityWarn1, msg, fields...)
}

// Error logs an error message
func (l *otelLogger) Error(ctx context.Context, msg string, fields ...Field) {
	l.log(ctx, otellog.SeverityError1, msg, fields...)
}

// Fatal logs a fatal message and exits
func (l *otelLogger) Fatal(ctx context.Context, msg string, fields ...Field) {
	l.log(ctx, otellog.SeverityFatal1, msg, fields...)
	os.Exit(1)
}

// With returns a new logger with the given fields
func (l *otelLogger) With(fields ...Field) Logger {
	return &otelLogger{
		logger:         l.logger,
		loggerProvider: l.loggerProvider,
		baseFields:     append(l.baseFields, fields...),
		cfg:            l.cfg,
		devMode:        l.devMode,
	}
}

// log is the internal logging method
func (l *otelLogger) log(ctx context.Context, severity otellog.Severity, msg string, fields ...Field) {
	// Combine base fields with provided fields
	allFields := append(l.baseFields, fields...)

	// Convert fields to attributes
	attrs := FieldsToAttributes(allFields)

	// Extract trace context from context
	spanContext := trace.SpanContextFromContext(ctx)

	// Create log record
	record := otellog.Record{}
	record.SetSeverity(severity)
	record.SetSeverityText(severity.String())
	record.SetBody(otellog.StringValue(msg))
	record.SetTimestamp(time.Now())

	// Add trace context as attributes if available
	if spanContext.IsValid() {
		attrs = append(attrs,
			attribute.String("trace_id", spanContext.TraceID().String()),
			attribute.String("span_id", spanContext.SpanID().String()),
		)
	}

	// Add attributes - convert attribute.KeyValue to otellog.KeyValue
	for _, attr := range attrs {
		kv := otellog.KeyValueFromAttribute(attr)
		record.AddAttributes(kv)
	}

	// Emit the log record
	l.logger.Emit(ctx, record)
}
