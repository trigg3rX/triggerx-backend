package worker

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/trigg3rX/triggerx-backend/internal/schedulers/condition/metrics"
)

// checkCondition fetches the current value and checks if condition is satisfied
func (w *ConditionWorker) checkCondition() error {
	startTime := time.Now()

	// Track condition check by source type
	metrics.TrackConditionBySource(w.ConditionWorkerData.ValueSourceType)

	// Fetch current value from source (with caching)
	currentValue, err := w.fetchValueWithCache()
	if err != nil {
		metrics.TrackValueParsingError(w.ConditionWorkerData.ValueSourceType)
		return fmt.Errorf("failed to fetch value: %w", err)
	}

	w.LastValue = currentValue
	w.LastCheckTimestamp = time.Now()

	// Create condition check context for Redis streaming
	conditionContext := map[string]interface{}{
		"job_id":         w.ConditionWorkerData.JobID,
		"current_value":  currentValue,
		"condition_type": w.ConditionWorkerData.ConditionType,
		"upper_limit":    w.ConditionWorkerData.UpperLimit,
		"lower_limit":    w.ConditionWorkerData.LowerLimit,
		"checked_at":     startTime.Unix(),
	}

	// Check if condition is satisfied
	satisfied, err := w.evaluateCondition(currentValue)
	if err != nil {
		conditionContext["status"] = "evaluation_error"
		conditionContext["error"] = err.Error()
		metrics.TrackCriticalError("condition_evaluation")
		return fmt.Errorf("failed to evaluate condition: %w", err)
	}

	// Track condition evaluation
	evaluationDuration := time.Since(startTime)
	metrics.TrackConditionEvaluation(evaluationDuration)

	// Track condition check with success status
	chainID := fmt.Sprintf("%d", w.ConditionWorkerData.JobID) // Using job_id as chain identifier for consistency
	metrics.TrackConditionCheck(chainID, evaluationDuration, satisfied)

	if satisfied {
		w.ConditionMet++
		metrics.TrackConditionByType(w.ConditionWorkerData.ConditionType)

		conditionContext["status"] = "satisfied"
		conditionContext["consecutive_checks"] = w.ConditionMet

		w.Logger.Info("Condition satisfied",
			"job_id", w.ConditionWorkerData.JobID,
			"current_value", currentValue,
			"condition_type", w.ConditionWorkerData.ConditionType,
			"upper_limit", w.ConditionWorkerData.UpperLimit,
			"lower_limit", w.ConditionWorkerData.LowerLimit,
			"consecutive_checks", w.ConditionMet,
		)

		// Notify scheduler about the trigger
		if w.TriggerCallback != nil {
			notification := &TriggerNotification{
				JobID:           w.ConditionWorkerData.JobID,
				TriggerValue:    currentValue,
				TriggeredAt:     time.Now(),
			}

			if err := w.TriggerCallback(notification); err != nil {
				w.Logger.Error("Failed to notify scheduler about trigger",
					"job_id", w.ConditionWorkerData.JobID,
					"error", err,
				)
				metrics.TrackCriticalError("trigger_notification_failed")
			} else {
				w.Logger.Info("Successfully notified scheduler about trigger",
					"job_id", w.ConditionWorkerData.JobID,
					"trigger_value", currentValue,
				)
			}
		} else {
			w.Logger.Warn("No trigger callback configured for worker",
				"job_id", w.ConditionWorkerData.JobID,
			)
		}

		// For non-recurring jobs, stop the worker after triggering
		if !w.ConditionWorkerData.Recurring {
			w.Logger.Info("Non-recurring job triggered, stopping worker", "job_id", w.ConditionWorkerData.JobID)
			go w.Stop() // Stop in a goroutine to avoid deadlock
		}

		duration := time.Since(startTime)
		conditionContext["duration_ms"] = duration.Milliseconds()
		conditionContext["completed_at"] = time.Now().Unix()
		conditionContext["action_status"] = "triggered"
	} else {
		w.ConditionMet = 0
		conditionContext["status"] = "not_satisfied"

		w.Logger.Debug("Condition not satisfied",
			"job_id", w.ConditionWorkerData.JobID,
			"current_value", currentValue,
			"condition_type", w.ConditionWorkerData.ConditionType,
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
	switch w.ConditionWorkerData.ValueSourceType {
	case SourceTypeAPI:
		return w.fetchFromAPI()
	case SourceTypeOracle:
		return w.fetchFromOracle()
	case SourceTypeStatic:
		return w.fetchStaticValue()
	default:
		return 0, fmt.Errorf("unsupported value source type: %s", w.ConditionWorkerData.ValueSourceType)
	}
}

// isTimeoutError checks if an error is a timeout error
func isTimeoutError(err error) bool {
	if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
		return true
	}
	return strings.Contains(err.Error(), "timeout") ||
		strings.Contains(err.Error(), "deadline exceeded")
}

// fetchFromAPI fetches value from an HTTP API endpoint
func (w *ConditionWorker) fetchFromAPI() (float64, error) {
	req, err := http.NewRequest("GET", w.ConditionWorkerData.ValueSourceUrl, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := w.HttpClient.DoWithRetry(req)
	if err != nil {
		metrics.TrackHTTPRequest("GET", w.ConditionWorkerData.ValueSourceUrl, "error")
		metrics.TrackHTTPClientConnectionError()

		// Check if it's a timeout error
		if isTimeoutError(err) {
			metrics.TrackTimeout("http_api_request")
		}

		return 0, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			w.Logger.Errorf("Error closing response body: %v", err)
		}
	}()

	statusCode := strconv.Itoa(resp.StatusCode)
	metrics.TrackHTTPRequest("GET", w.ConditionWorkerData.ValueSourceUrl, statusCode)
	metrics.TrackAPIResponse(w.ConditionWorkerData.ValueSourceUrl, statusCode)

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("HTTP request failed with status: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		metrics.TrackHTTPRequest("GET", w.ConditionWorkerData.ValueSourceUrl, "read_error")
		return 0, fmt.Errorf("failed to read response body: %w", err)
	}

	// Try to parse response as ValueResponse struct
	var valueResp ValueResponse
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

	metrics.TrackInvalidValue(w.ConditionWorkerData.ValueSourceUrl)
	metrics.TrackValueParsingError(w.ConditionWorkerData.ValueSourceType)
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
	value, err := strconv.ParseFloat(w.ConditionWorkerData.ValueSourceUrl, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid static value: %s", w.ConditionWorkerData.ValueSourceUrl)
	}
	return value, nil
}

// evaluateCondition checks if the current value satisfies the condition
func (w *ConditionWorker) evaluateCondition(currentValue float64) (bool, error) {
	switch w.ConditionWorkerData.ConditionType {
	case ConditionGreaterThan:
		return currentValue > w.ConditionWorkerData.LowerLimit, nil
	case ConditionLessThan:
		return currentValue < w.ConditionWorkerData.UpperLimit, nil
	case ConditionBetween:
		return currentValue >= w.ConditionWorkerData.LowerLimit && currentValue <= w.ConditionWorkerData.UpperLimit, nil
	case ConditionEquals:
		return currentValue == w.ConditionWorkerData.LowerLimit, nil
	case ConditionNotEquals:
		return currentValue != w.ConditionWorkerData.LowerLimit, nil
	case ConditionGreaterEqual:
		return currentValue >= w.ConditionWorkerData.LowerLimit, nil
	case ConditionLessEqual:
		return currentValue <= w.ConditionWorkerData.UpperLimit, nil
	default:
		return false, fmt.Errorf("unsupported condition type: %s", w.ConditionWorkerData.ConditionType)
	}
}
