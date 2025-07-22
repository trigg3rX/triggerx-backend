package performers

const (
	// Performer Management
	PerformerLockPrefix = "performer:lock:"      // Redis key prefix for performer locks
	PerformerListKey    = "performers:available" // List of available performers
)