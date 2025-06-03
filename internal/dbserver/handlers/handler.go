package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/config"
	"github.com/trigg3rX/triggerx-backend/internal/dbserver/repository"
	"github.com/trigg3rX/triggerx-backend/pkg/database"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
	"github.com/trigg3rX/triggerx-backend/pkg/retry"
)

type NotificationConfig struct {
	EmailFrom     string
	EmailPassword string
	BotToken      string
}

type Handler struct {
	db                     *database.Connection
	logger                 logging.Logger
	config                 NotificationConfig
	jobRepository          repository.JobRepository
	timeJobRepository      repository.TimeJobRepository
	eventJobRepository     repository.EventJobRepository
	conditionJobRepository repository.ConditionJobRepository
	taskRepository         repository.TaskRepository
	userRepository         repository.UserRepository
	keeperRepository       repository.KeeperRepository
	apiKeysRepository      repository.ApiKeysRepository

	scanNowQuery func(*time.Time) error // for testability
}

func NewHandler(db *database.Connection, logger logging.Logger, config NotificationConfig) *Handler {
	h := &Handler{
		db:                     db,
		logger:                 logger,
		config:                 config,
		jobRepository:          repository.NewJobRepository(db),
		timeJobRepository:      repository.NewTimeJobRepository(db),
		eventJobRepository:     repository.NewEventJobRepository(db),
		conditionJobRepository: repository.NewConditionJobRepository(db),
		taskRepository:         repository.NewTaskRepository(db),
		userRepository:         repository.NewUserRepository(db),
		keeperRepository:       repository.NewKeeperRepository(db),
		apiKeysRepository:      repository.NewApiKeysRepository(db),
	}
	h.scanNowQuery = h.defaultScanNowQuery
	return h
}

func (h *Handler) defaultScanNowQuery(timestamp *time.Time) error {
	return h.db.Session().Query("SELECT now() FROM system.local").Scan(timestamp)
}

// SendDataToEventScheduler sends data to the event scheduler
func (h *Handler) SendDataToEventScheduler(route string, data interface{}) (bool, error) {
	apiURL := fmt.Sprintf("%s%s", config.GetEventSchedulerRPCUrl(), route)
	return h.sendDataToScheduler(apiURL, data, "event scheduler")
}

// SendDataToConditionScheduler sends data to the condition scheduler
func (h *Handler) SendDataToConditionScheduler(route string, data interface{}) (bool, error) {
	apiURL := fmt.Sprintf("%s%s", config.GetConditionSchedulerRPCUrl(), route)
	return h.sendDataToScheduler(apiURL, data, "condition scheduler")
}

// sendDataToScheduler is a generic function to send data to any scheduler
func (h *Handler) sendDataToScheduler(apiURL string, data interface{}, schedulerName string) (bool, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return false, fmt.Errorf("error marshaling data: %v", err)
	}

	// Create a client with aggressive timeouts and connection pooling
	retryConfig := &retry.HTTPRetryConfig{
		MaxRetries:      3,
		InitialDelay:    200 * time.Millisecond,
		MaxDelay:        2 * time.Second,
		BackoffFactor:   2.0,
		LogRetryAttempt: true,
		RetryStatusCodes: []int{
			http.StatusInternalServerError,
			http.StatusBadGateway,
			http.StatusServiceUnavailable,
			http.StatusGatewayTimeout,
		},
		Timeout:             3 * time.Second,
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 100,
		IdleConnTimeout:     30 * time.Second,
	}

	client := retry.NewHTTPClient(retryConfig, h.logger)

	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return false, fmt.Errorf("error creating request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Close = true

	resp, err := client.DoWithRetry(req)
	if err != nil {
		return false, fmt.Errorf("error sending data to %s: %v", schedulerName, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return false, fmt.Errorf("%s service error (status=%d): %s", schedulerName, resp.StatusCode, string(body))
	}

	h.logger.Infof("Successfully sent data to %s", schedulerName)
	return true, nil
}

// SendPauseToEventScheduler sends a DELETE request to the event scheduler
func (h *Handler) SendPauseToEventScheduler(route string) (bool, error) {
	apiURL := fmt.Sprintf("%s%s", config.GetEventSchedulerRPCUrl(), route)
	return h.sendPauseToScheduler(apiURL, "event scheduler")
}

// SendPauseToConditionScheduler sends a DELETE request to the condition scheduler
func (h *Handler) SendPauseToConditionScheduler(route string) (bool, error) {
	apiURL := fmt.Sprintf("%s%s", config.GetConditionSchedulerRPCUrl(), route)
	return h.sendPauseToScheduler(apiURL, "condition scheduler")
}

// sendPauseToScheduler sends a DELETE request to any scheduler
func (h *Handler) sendPauseToScheduler(apiURL string, schedulerName string) (bool, error) {
	// Create a client with aggressive timeouts and connection pooling
	retryConfig := &retry.HTTPRetryConfig{
		MaxRetries:      3,
		InitialDelay:    200 * time.Millisecond,
		MaxDelay:        2 * time.Second,
		BackoffFactor:   2.0,
		LogRetryAttempt: true,
		RetryStatusCodes: []int{
			http.StatusInternalServerError,
			http.StatusBadGateway,
			http.StatusServiceUnavailable,
			http.StatusGatewayTimeout,
		},
		Timeout:             3 * time.Second,
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 100,
		IdleConnTimeout:     30 * time.Second,
	}

	client := retry.NewHTTPClient(retryConfig, h.logger)

	req, err := http.NewRequest("POST", apiURL, nil)
	if err != nil {
		return false, fmt.Errorf("error creating request: %v", err)
	}
	req.Close = true

	resp, err := client.DoWithRetry(req)
	if err != nil {
		return false, fmt.Errorf("error sending DELETE to %s: %v", schedulerName, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return false, fmt.Errorf("%s service error (status=%d): %s", schedulerName, resp.StatusCode, string(body))
	}

	h.logger.Infof("Successfully sent DELETE to %s", schedulerName)
	return true, nil
}

func (h *Handler) HealthCheck(c *gin.Context) {
	startTime := time.Now()

	// Check database connection by executing a simple query
	dbStatus := "healthy"
	dbError := ""

	// Use a simple system query to test the connection
	var timestamp time.Time
	if err := h.scanNowQuery(&timestamp); err != nil {
		dbStatus = "unhealthy"
		dbError = err.Error()
		h.logger.Errorf("Database health check failed: %v", err)
	}

	// Prepare response
	response := gin.H{
		"status":    "ok",
		"timestamp": startTime.Unix(),
		"service":   "dbserver",
		"version":   "1.0.0",
		"uptime":    time.Since(startTime).String(),
		"database": gin.H{
			"status": dbStatus,
			"error":  dbError,
		},
		"checks": gin.H{
			"database_connection": dbStatus == "healthy",
		},
	}

	// Set appropriate HTTP status
	httpStatus := http.StatusOK
	if dbStatus != "healthy" {
		httpStatus = http.StatusServiceUnavailable
		response["status"] = "degraded"
	}

	// Log health check
	duration := time.Since(startTime)
	h.logger.Infof("Health check completed: status=%s, db_status=%s, duration=%v",
		response["status"], dbStatus, duration)

	c.JSON(httpStatus, response)
}
