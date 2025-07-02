package tracing

import (
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// GinMiddleware returns a Gin middleware that instruments HTTP requests with OpenTelemetry tracing
func GinMiddleware(serviceName string) gin.HandlerFunc {
	return otelgin.Middleware(serviceName)
}

// GetTracer returns an OpenTelemetry tracer for the given service
func GetTracer(serviceName string) trace.Tracer {
	return otel.Tracer(serviceName)
}

// AddBusinessAttributes adds TriggerX-specific attributes to a span
func AddBusinessAttributes(span trace.Span, attrs map[string]interface{}) {
	var otelAttrs []attribute.KeyValue

	for key, value := range attrs {
		switch v := value.(type) {
		case string:
			otelAttrs = append(otelAttrs, attribute.String(key, v))
		case int:
			otelAttrs = append(otelAttrs, attribute.Int(key, v))
		case int64:
			otelAttrs = append(otelAttrs, attribute.Int64(key, v))
		case float64:
			otelAttrs = append(otelAttrs, attribute.Float64(key, v))
		case bool:
			otelAttrs = append(otelAttrs, attribute.Bool(key, v))
		}
	}

	span.SetAttributes(otelAttrs...)
}

// TriggerXAttributes provides commonly used attribute keys for TriggerX services
var TriggerXAttributes = struct {
	// User attributes
	UserID      string
	UserAddress string

	// Job attributes
	JobID     string
	JobType   string
	TaskDefID string

	// Database attributes
	DBOperation string
	DBTable     string

	// Scheduler attributes
	SchedulerType     string
	BatchSize         string
	ProcessingTime    string
	NextExecutionTime string

	// Worker attributes
	WorkerType string
	WorkerID   string

	// Blockchain attributes
	BlockchainNetwork string
	EventType         string
	TransactionHash   string
	BlockNumber       string

	// Task attributes
	TaskID         string
	TaskStatus     string
	TaskRetryCount string

	// Keeper attributes
	KeeperID       string
	KeeperStatus   string
	AssignmentTime string
}{
	// User attributes
	UserID:      "user.id",
	UserAddress: "user.address",

	// Job attributes
	JobID:     "job.id",
	JobType:   "job.type",
	TaskDefID: "task.definition_id",

	// Database attributes
	DBOperation: "db.operation",
	DBTable:     "db.table",

	// Scheduler attributes
	SchedulerType:     "scheduler.type",
	BatchSize:         "batch.size",
	ProcessingTime:    "batch.processing_time",
	NextExecutionTime: "job.next_execution_time",

	// Worker attributes
	WorkerType: "worker.type",
	WorkerID:   "worker.id",

	// Blockchain attributes
	BlockchainNetwork: "blockchain.network",
	EventType:         "blockchain.event_type",
	TransactionHash:   "event.transaction_hash",
	BlockNumber:       "blockchain.block_number",

	// Task attributes
	TaskID:         "task.id",
	TaskStatus:     "task.status",
	TaskRetryCount: "task.retry_count",

	// Keeper attributes
	KeeperID:       "keeper.id",
	KeeperStatus:   "keeper.status",
	AssignmentTime: "keeper.assignment_time",
}
