package observability

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	otellog "go.opentelemetry.io/otel/log"
	sdklog "go.opentelemetry.io/otel/sdk/log"
)

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorCyan   = "\033[36m"
	colorWhite  = "\033[37m"
	colorGray   = "\033[90m"
)

// ConsoleExporter exports logs to the console with colors in dev mode
type ConsoleExporter struct {
	writer io.Writer
}

// NewConsoleExporter creates a new console exporter
func NewConsoleExporter() *ConsoleExporter {
	return &ConsoleExporter{
		writer: os.Stderr,
	}
}

// Export exports log records to the console
func (e *ConsoleExporter) Export(ctx context.Context, records []sdklog.Record) error {
	for _, record := range records {
		e.writeRecord(record)
	}
	return nil
}

// Shutdown shuts down the exporter
func (e *ConsoleExporter) Shutdown(ctx context.Context) error {
	return nil
}

// ForceFlush forces a flush of the exporter
func (e *ConsoleExporter) ForceFlush(ctx context.Context) error {
	return nil
}

// writeRecord writes a single log record to the console with colors
func (e *ConsoleExporter) writeRecord(record sdklog.Record) {
	severity := record.Severity()
	severityText := record.SeverityText()
	if severityText == "" {
		severityText = severity.String()
	}
	color := getColorForSeverity(severity)

	// Format timestamp
	timestamp := record.Timestamp().Format(time.RFC3339)

	// Format message
	body := record.Body()
	message := body.AsString()

	// Format attributes
	var attrs []otellog.KeyValue
	record.WalkAttributes(func(kv otellog.KeyValue) bool {
		attrs = append(attrs, kv)
		return true
	})
	attrStr := formatAttributes(attrs)

	// Format trace context if available
	traceStr := formatTraceContext(record)

	// Build the log line
	var logLine string
	if traceStr != "" {
		logLine = fmt.Sprintf("%s[%s]%s %s%s%s %s%s%s %s%s%s %s%s%s\n",
			colorGray, timestamp, colorReset,
			color, severityText, colorReset,
			colorCyan, traceStr, colorReset,
			colorWhite, message, colorReset,
			colorGray, attrStr, colorReset,
		)
	} else {
		logLine = fmt.Sprintf("%s[%s]%s %s%s%s %s%s%s %s%s%s\n",
			colorGray, timestamp, colorReset,
			color, severityText, colorReset,
			colorWhite, message, colorReset,
			colorGray, attrStr, colorReset,
		)
	}

	_, _ = e.writer.Write([]byte(logLine))
}

// getColorForSeverity returns the color code for a given severity level
func getColorForSeverity(severity otellog.Severity) string {
	switch severity {
	case otellog.SeverityTrace, otellog.SeverityDebug1, otellog.SeverityDebug2, otellog.SeverityDebug3, otellog.SeverityDebug4:
		return colorGray
	case otellog.SeverityInfo1, otellog.SeverityInfo2, otellog.SeverityInfo3, otellog.SeverityInfo4:
		return colorGreen
	case otellog.SeverityWarn1, otellog.SeverityWarn2, otellog.SeverityWarn3, otellog.SeverityWarn4:
		return colorYellow
	case otellog.SeverityError1, otellog.SeverityError2, otellog.SeverityError3, otellog.SeverityError4:
		return colorRed
	case otellog.SeverityFatal1, otellog.SeverityFatal2, otellog.SeverityFatal3, otellog.SeverityFatal4:
		return colorRed
	default:
		return colorWhite
	}
}

// formatAttributes formats attributes as a string
func formatAttributes(attrs []otellog.KeyValue) string {
	if len(attrs) == 0 {
		return ""
	}

	var parts []string
	for _, attr := range attrs {
		key := string(attr.Key)
		value := formatValue(attr.Value)
		parts = append(parts, fmt.Sprintf("%s=%s", key, value))
	}

	if len(parts) == 0 {
		return ""
	}
	return fmt.Sprintf("{%s}", fmt.Sprint(parts))
}

// formatValue converts an OpenTelemetry log value to a string representation
func formatValue(value otellog.Value) string {
	switch value.Kind() {
	case otellog.KindString:
		return value.AsString()
	case otellog.KindInt64:
		return fmt.Sprintf("%d", value.AsInt64())
	case otellog.KindFloat64:
		// Use %g to avoid unnecessary trailing zeros
		return fmt.Sprintf("%g", value.AsFloat64())
	case otellog.KindBool:
		return fmt.Sprintf("%t", value.AsBool())
	case otellog.KindBytes:
		return fmt.Sprintf("%x", value.AsBytes())
	case otellog.KindSlice:
		return formatSliceValue(value)
	default:
		return fmt.Sprintf("%v", value)
	}
}

// formatSliceValue formats a slice value
func formatSliceValue(value otellog.Value) string {
	switch value.Kind() {
	case otellog.KindSlice:
		slice := value.AsSlice()
		if len(slice) == 0 {
			return "[]"
		}
		var parts []string
		for _, v := range slice {
			parts = append(parts, formatValue(v))
		}
		return fmt.Sprintf("[%s]", fmt.Sprint(parts))
	default:
		return fmt.Sprintf("%v", value)
	}
}

// formatTraceContext formats trace context if available
func formatTraceContext(record sdklog.Record) string {
	traceID := record.TraceID()
	spanID := record.SpanID()

	if traceID.IsValid() {
		return fmt.Sprintf("trace_id=%s span_id=%s", traceID.String(), spanID.String())
	}

	return ""
}

// Ensure ConsoleExporter implements the sdklog.Exporter interface
var _ sdklog.Exporter = (*ConsoleExporter)(nil)
