package worker

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	redisx "github.com/trigg3rX/triggerx-backend/internal/redis"
	"github.com/trigg3rX/triggerx-backend/internal/schedulers/condition/metrics"

	schedulerTypes "github.com/trigg3rX/triggerx-backend/internal/schedulers/condition/scheduler/types"
)

// checkCondition fetches the current value and checks if condition is satisfied
func (w *ConditionWorker) checkCondition() error {
	startTime := time.Now()

	// Track condition check
	metrics.ConditionsChecked.Inc()

	// Check for duplicate condition check prevention
	checkKey := fmt.Sprintf("condition_check_%d_%d", w.Job.JobID, startTime.Unix())
	if w.Cache != nil {
		if _, err := w.Cache.Get(checkKey); err == nil {
			w.Logger.Debug("Condition check already performed recently, skipping", "job_id", w.Job.JobID)
			return nil
		}
		// Mark this check time to prevent duplicates
		if err := w.Cache.Set(checkKey, time.Now().Format(time.RFC3339), schedulerTypes.DuplicateConditionWindow); err != nil {
			w.Logger.Errorf("Failed to set condition check cache: %v", err)
		}
	}

	// Fetch current value from source (with caching)
	currentValue, err := w.fetchValueWithCache()
	if err != nil {
		return fmt.Errorf("failed to fetch value: %w", err)
	}

	w.LastValue = currentValue
	w.LastCheck = time.Now()

	// Create condition check context for Redis streaming
	conditionContext := map[string]interface{}{
		"job_id":         w.Job.JobID,
		"manager_id":     w.ManagerID,
		"current_value":  currentValue,
		"condition_type": w.Job.ConditionType,
		"upper_limit":    w.Job.UpperLimit,
		"lower_limit":    w.Job.LowerLimit,
		"checked_at":     startTime.Unix(),
	}

	// Check if condition is satisfied
	satisfied, err := w.evaluateCondition(currentValue)
	if err != nil {
		conditionContext["status"] = "evaluation_error"
		conditionContext["error"] = err.Error()
		if err := redisx.AddJobToStream(redisx.JobsRetryConditionStream, conditionContext); err != nil {
			w.Logger.Warnf("Failed to add condition evaluation error to Redis stream: %v", err)
		}
		return fmt.Errorf("failed to evaluate condition: %w", err)
	}

	if satisfied {
		w.ConditionMet++
		metrics.ConditionsSatisfied.Inc()

		conditionContext["status"] = "satisfied"
		conditionContext["consecutive_checks"] = w.ConditionMet

		w.Logger.Info("Condition satisfied",
			"job_id", w.Job.JobID,
			"current_value", currentValue,
			"condition_type", w.Job.ConditionType,
			"upper_limit", w.Job.UpperLimit,
			"lower_limit", w.Job.LowerLimit,
			"consecutive_checks", w.ConditionMet,
		)

		// Add satisfied condition to Redis stream
		if err := redisx.AddJobToStream(redisx.JobsReadyConditionStream, conditionContext); err != nil {
			w.Logger.Warnf("Failed to add condition satisfied event to Redis stream: %v", err)
		}

		// Execute action
		executionSuccess := w.performActionExecution(currentValue)

		duration := time.Since(startTime)
		conditionContext["duration_ms"] = duration.Milliseconds()
		conditionContext["completed_at"] = time.Now().Unix()

		if executionSuccess {
			conditionContext["action_status"] = "completed"
			if err := redisx.AddJobToStream(redisx.JobsReadyConditionStream, conditionContext); err != nil {
				w.Logger.Warnf("Failed to add condition action completion to Redis stream: %v", err)
			}
		} else {
			conditionContext["action_status"] = "failed"
			if err := redisx.AddJobToStream(redisx.JobsRetryConditionStream, conditionContext); err != nil {
				w.Logger.Warnf("Failed to add condition action failure to Redis stream: %v", err)
			}
		}
	} else {
		w.ConditionMet = 0
		conditionContext["status"] = "not_satisfied"

		w.Logger.Debug("Condition not satisfied",
			"job_id", w.Job.JobID,
			"current_value", currentValue,
			"condition_type", w.Job.ConditionType,
		)

		// Periodically log condition checks for monitoring
		if time.Now().Unix()%60 == 0 { // Every minute
			if err := redisx.AddJobToStream(redisx.JobsReadyConditionStream, conditionContext); err != nil {
				w.Logger.Warnf("Failed to add condition check status to Redis stream: %v", err)
			}
		}
	}

	// Cache the condition state
	if w.Cache != nil {
		w.cacheConditionState(currentValue, satisfied)
	}

	return nil
}

// fetchValueWithCache retrieves the current value with caching support
func (w *ConditionWorker) fetchValueWithCache() (float64, error) {
	// Try to get cached value first
	if w.Cache != nil {
		cacheKey := fmt.Sprintf("value_%d_%s", w.Job.JobID, w.Job.ValueSourceUrl)
		if cached, err := w.Cache.Get(cacheKey); err == nil {
			var cachedValue float64
			if _, err := fmt.Sscanf(cached, "%f", &cachedValue); err == nil {
				w.Logger.Debug("Using cached value", "job_id", w.Job.JobID, "value", cachedValue)
				return cachedValue, nil
			}
		}
	}

	// Fetch fresh value
	currentValue, err := w.fetchValue()
	if err != nil {
		return 0, err
	}

	// Cache the value
	if w.Cache != nil {
		cacheKey := fmt.Sprintf("value_%d_%s", w.Job.JobID, w.Job.ValueSourceUrl)
		if err := w.Cache.Set(cacheKey, fmt.Sprintf("%f", currentValue), schedulerTypes.ValueCacheTTL); err != nil {
			w.Logger.Errorf("Failed to set value cache: %v", err)
		}
	}

	return currentValue, nil
}

// cacheConditionState caches the current condition state
func (w *ConditionWorker) cacheConditionState(value float64, satisfied bool) {
	if w.Cache == nil {
		return
	}

	stateData := map[string]interface{}{
		"job_id":        w.Job.JobID,
		"current_value": value,
		"satisfied":     satisfied,
		"condition_met": w.ConditionMet,
		"last_check":    w.LastCheck.Unix(),
		"cached_at":     time.Now().Unix(),
	}

	if jsonData, err := json.Marshal(stateData); err == nil {
		cacheKey := fmt.Sprintf("condition_state_%d", w.Job.JobID)
		if err := w.Cache.Set(cacheKey, string(jsonData), schedulerTypes.ConditionStateCacheTTL); err != nil {
			w.Logger.Errorf("Failed to set condition state cache: %v", err)
		}
	}
}

// fetchValue retrieves the current value from the configured source
func (w *ConditionWorker) fetchValue() (float64, error) {
	switch w.Job.ValueSourceType {
	case schedulerTypes.SourceTypeAPI:
		return w.fetchFromAPI()
	case schedulerTypes.SourceTypeOracle:
		return w.fetchFromOracle()
	case schedulerTypes.SourceTypeStatic:
		return w.fetchStaticValue()
	default:
		return 0, fmt.Errorf("unsupported value source type: %s", w.Job.ValueSourceType)
	}
}

// fetchFromAPI fetches value from an HTTP API endpoint
func (w *ConditionWorker) fetchFromAPI() (float64, error) {
	metrics.ValueSourceRequests.Inc()

	resp, err := w.HttpClient.Get(w.Job.ValueSourceUrl)
	if err != nil {
		metrics.ValueSourceErrors.Inc()
		return 0, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		metrics.ValueSourceErrors.Inc()
		return 0, fmt.Errorf("HTTP request failed with status: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		metrics.ValueSourceErrors.Inc()
		return 0, fmt.Errorf("failed to read response body: %w", err)
	}

	// Try to parse response as ValueResponse struct
	var valueResp schedulerTypes.ValueResponse
	if err := json.Unmarshal(body, &valueResp); err == nil {
		// Check which field has a non-zero value
		if valueResp.Value != 0 {
			return valueResp.Value, nil
		}
		if valueResp.Price != 0 {
			return valueResp.Price, nil
		}
		if valueResp.USD != 0 {
			return valueResp.USD, nil
		}
		if valueResp.Rate != 0 {
			return valueResp.Rate, nil
		}
		if valueResp.Result != 0 {
			return valueResp.Result, nil
		}
		if valueResp.Data != 0 {
			return valueResp.Data, nil
		}
	}

	// Try to parse as a simple float
	var floatValue float64
	if err := json.Unmarshal(body, &floatValue); err == nil {
		return floatValue, nil
	}

	// Try to parse as a simple string and convert to float
	var stringValue string
	if err := json.Unmarshal(body, &stringValue); err == nil {
		if floatVal, parseErr := strconv.ParseFloat(stringValue, 64); parseErr == nil {
			return floatVal, nil
		}
	}

	metrics.ValueSourceErrors.Inc()
	return 0, fmt.Errorf("could not extract numeric value from response: %s", string(body))
}

// fetchFromOracle fetches value from an oracle (placeholder implementation)
func (w *ConditionWorker) fetchFromOracle() (float64, error) {
	// TODO: Implement oracle-specific logic
	// For now, treat as API endpoint
	return w.fetchFromAPI()
}

// fetchStaticValue returns a static value (for testing purposes)
func (w *ConditionWorker) fetchStaticValue() (float64, error) {
	// Parse URL as the static value
	value, err := strconv.ParseFloat(w.Job.ValueSourceUrl, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid static value: %s", w.Job.ValueSourceUrl)
	}
	return value, nil
}

// evaluateCondition checks if the current value satisfies the condition
func (w *ConditionWorker) evaluateCondition(currentValue float64) (bool, error) {
	switch w.Job.ConditionType {
	case schedulerTypes.ConditionGreaterThan:
		return currentValue > w.Job.LowerLimit, nil
	case schedulerTypes.ConditionLessThan:
		return currentValue < w.Job.UpperLimit, nil
	case schedulerTypes.ConditionBetween:
		return currentValue >= w.Job.LowerLimit && currentValue <= w.Job.UpperLimit, nil
	case schedulerTypes.ConditionEquals:
		return currentValue == w.Job.LowerLimit, nil
	case schedulerTypes.ConditionNotEquals:
		return currentValue != w.Job.LowerLimit, nil
	case schedulerTypes.ConditionGreaterEqual:
		return currentValue >= w.Job.LowerLimit, nil
	case schedulerTypes.ConditionLessEqual:
		return currentValue <= w.Job.UpperLimit, nil
	default:
		return false, fmt.Errorf("unsupported condition type: %s", w.Job.ConditionType)
	}
}
