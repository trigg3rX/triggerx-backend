package types

import "time"

const (
	ValueCacheTTL            = 30 * time.Second // Cache TTL for API values
	PerformerLockTTL         = 15 * time.Minute // Lock duration for condition monitoring
	DuplicateConditionWindow = 10 * time.Second // Window to prevent duplicate condition processing
	ConditionStateCacheTTL   = 5 * time.Minute  // Cache TTL for condition state
)

// Supported condition types
const (
	ConditionGreaterThan  = "greater_than"
	ConditionLessThan     = "less_than"
	ConditionBetween      = "between"
	ConditionEquals       = "equals"
	ConditionNotEquals    = "not_equals"
	ConditionGreaterEqual = "greater_equal"
	ConditionLessEqual    = "less_equal"
)

// Supported value source types
const (
	SourceTypeAPI    = "api"
	SourceTypeOracle = "oracle"
	SourceTypeStatic = "static"
)

// ValueResponse represents a generic response structure for fetching values
type ValueResponse struct {
	Value     float64 `json:"value"`
	Price     float64 `json:"price"`     // Common for price APIs
	USD       float64 `json:"usd"`       // Common for CoinGecko-style APIs
	Rate      float64 `json:"rate"`      // Common for exchange rate APIs
	Result    float64 `json:"result"`    // Generic result field
	Data      float64 `json:"data"`      // Generic data field
	Timestamp int64   `json:"timestamp"` // Optional timestamp
}
