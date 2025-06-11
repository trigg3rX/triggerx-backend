package worker

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/trigg3rX/triggerx-backend/internal/schedulers/condition/metrics"

	schedulerTypes "github.com/trigg3rX/triggerx-backend/internal/schedulers/condition/scheduler/types"
)

// checkCondition fetches the current value and checks if condition is satisfied
func (w *ConditionWorker) checkCondition() error {
	startTime := time.Now()

	// Track condition check
	metrics.ConditionsChecked.Inc()

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

		// Execute action
		executionSuccess := w.performActionExecution(currentValue)

		duration := time.Since(startTime)
		conditionContext["duration_ms"] = duration.Milliseconds()
		conditionContext["completed_at"] = time.Now().Unix()

		if executionSuccess {
			conditionContext["action_status"] = "completed"
		} else {
			conditionContext["action_status"] = "failed"
		}
	} else {
		w.ConditionMet = 0
		conditionContext["status"] = "not_satisfied"

		w.Logger.Debug("Condition not satisfied",
			"job_id", w.Job.JobID,
			"current_value", currentValue,
			"condition_type", w.Job.ConditionType,
		)
	}
	return nil
}

// fetchValueWithCache retrieves the current value with caching support
func (w *ConditionWorker) fetchValueWithCache() (float64, error) {
	// Fetch fresh value
	currentValue, err := w.fetchValue()
	if err != nil {
		return 0, err
	}

	return currentValue, nil
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
	defer func() {
		if err := resp.Body.Close(); err != nil {
			w.Logger.Errorf("Error closing response body: %v", err)
		}
	}()

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
