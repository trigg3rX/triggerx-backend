package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/trigg3rX/triggerx-backend/internal/dbserver/config"
	"github.com/trigg3rX/triggerx-backend/pkg/database"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

type NotificationConfig struct {
	EmailFrom     string
	EmailPassword string
	BotToken      string
}

type Handler struct {
	db     *database.Connection
	logger logging.Logger
	config NotificationConfig
}

func NewHandler(db *database.Connection, logger logging.Logger, config NotificationConfig) *Handler {
	return &Handler{
		db:     db,
		logger: logger,
		config: config,
	}
}

func (h *Handler) SendDataToManager(route string, data interface{}) (bool, error) {
	apiURL := fmt.Sprintf("%s%s", config.GetManagerRPCAddress(), route)

	jsonData, err := json.Marshal(data)
	if err != nil {
		return false, fmt.Errorf("error marshaling data: %v", err)
	}

	// Create a client with aggressive timeouts and connection pooling
	client := &http.Client{
		Timeout: 3 * time.Second, // Reduced timeout
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 100,
			IdleConnTimeout:     30 * time.Second,
			DisableKeepAlives:   false,
			DialContext: (&net.Dialer{
				Timeout:   2 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			TLSHandshakeTimeout:   2 * time.Second,
			ResponseHeaderTimeout: 2 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}

	// Retry logic with shorter backoff
	maxRetries := 3
	backoff := 200 * time.Millisecond

	for i := 0; i < maxRetries; i++ {
		req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
		if err != nil {
			return false, fmt.Errorf("error creating request: %v", err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Close = true // Ensure connection is closed after request

		resp, err := client.Do(req)
		if err != nil {
			if i == maxRetries-1 {
				return false, fmt.Errorf("error sending data to manager after %d retries: %v", maxRetries, err)
			}
			h.logger.Warnf("Attempt %d failed, retrying in %v: %v", i+1, backoff, err)
			time.Sleep(backoff)
			backoff *= 2 // Exponential backoff
			continue
		}

		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			if i == maxRetries-1 {
				return false, fmt.Errorf("manager service error (status=%d): %s", resp.StatusCode, string(body))
			}
			h.logger.Warnf("Attempt %d failed with status %d, retrying in %v", i+1, resp.StatusCode, backoff)
			time.Sleep(backoff)
			backoff *= 2
			continue
		}

		return true, nil
	}

	return false, fmt.Errorf("failed to send data to manager after %d retries", maxRetries)
}

// SendDataToEventScheduler sends data to the event scheduler
func (h *Handler) SendDataToEventScheduler(route string, data interface{}) (bool, error) {
	apiURL := fmt.Sprintf("%s%s", config.GetEventSchedulerRPCAddress(), route)
	return h.sendDataToScheduler(apiURL, data, "event scheduler")
}

// SendDataToConditionScheduler sends data to the condition scheduler
func (h *Handler) SendDataToConditionScheduler(route string, data interface{}) (bool, error) {
	apiURL := fmt.Sprintf("%s%s", config.GetConditionSchedulerRPCAddress(), route)
	return h.sendDataToScheduler(apiURL, data, "condition scheduler")
}

// sendDataToScheduler is a generic function to send data to any scheduler
func (h *Handler) sendDataToScheduler(apiURL string, data interface{}, schedulerName string) (bool, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return false, fmt.Errorf("error marshaling data: %v", err)
	}

	// Create a client with aggressive timeouts and connection pooling
	client := &http.Client{
		Timeout: 3 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 100,
			IdleConnTimeout:     30 * time.Second,
			DisableKeepAlives:   false,
			DialContext: (&net.Dialer{
				Timeout:   2 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			TLSHandshakeTimeout:   2 * time.Second,
			ResponseHeaderTimeout: 2 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}

	// Retry logic with shorter backoff
	maxRetries := 3
	backoff := 200 * time.Millisecond

	for i := 0; i < maxRetries; i++ {
		req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
		if err != nil {
			return false, fmt.Errorf("error creating request: %v", err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Close = true

		resp, err := client.Do(req)
		if err != nil {
			if i == maxRetries-1 {
				return false, fmt.Errorf("error sending data to %s after %d retries: %v", schedulerName, maxRetries, err)
			}
			h.logger.Warnf("%s attempt %d failed, retrying in %v: %v", schedulerName, i+1, backoff, err)
			time.Sleep(backoff)
			backoff *= 2
			continue
		}

		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			if i == maxRetries-1 {
				return false, fmt.Errorf("%s service error (status=%d): %s", schedulerName, resp.StatusCode, string(body))
			}
			h.logger.Warnf("%s attempt %d failed with status %d, retrying in %v", schedulerName, i+1, resp.StatusCode, backoff)
			time.Sleep(backoff)
			backoff *= 2
			continue
		}

		h.logger.Infof("Successfully sent data to %s", schedulerName)
		return true, nil
	}

	return false, fmt.Errorf("failed to send data to %s after %d retries", schedulerName, maxRetries)
}

// SendDeleteToEventScheduler sends a DELETE request to the event scheduler
func (h *Handler) SendDeleteToEventScheduler(route string) (bool, error) {
	apiURL := fmt.Sprintf("%s%s", config.GetEventSchedulerRPCAddress(), route)
	return h.sendDeleteToScheduler(apiURL, "event scheduler")
}

// SendDeleteToConditionScheduler sends a DELETE request to the condition scheduler
func (h *Handler) SendDeleteToConditionScheduler(route string) (bool, error) {
	apiURL := fmt.Sprintf("%s%s", config.GetConditionSchedulerRPCAddress(), route)
	return h.sendDeleteToScheduler(apiURL, "condition scheduler")
}

// sendDeleteToScheduler sends a DELETE request to any scheduler
func (h *Handler) sendDeleteToScheduler(apiURL string, schedulerName string) (bool, error) {
	// Create a client with aggressive timeouts and connection pooling
	client := &http.Client{
		Timeout: 3 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 100,
			IdleConnTimeout:     30 * time.Second,
			DisableKeepAlives:   false,
			DialContext: (&net.Dialer{
				Timeout:   2 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			TLSHandshakeTimeout:   2 * time.Second,
			ResponseHeaderTimeout: 2 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}

	// Retry logic with shorter backoff
	maxRetries := 3
	backoff := 200 * time.Millisecond

	for i := 0; i < maxRetries; i++ {
		req, err := http.NewRequest("DELETE", apiURL, nil)
		if err != nil {
			return false, fmt.Errorf("error creating DELETE request: %v", err)
		}
		req.Close = true

		resp, err := client.Do(req)
		if err != nil {
			if i == maxRetries-1 {
				return false, fmt.Errorf("error sending DELETE to %s after %d retries: %v", schedulerName, maxRetries, err)
			}
			h.logger.Warnf("%s DELETE attempt %d failed, retrying in %v: %v", schedulerName, i+1, backoff, err)
			time.Sleep(backoff)
			backoff *= 2
			continue
		}

		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			if i == maxRetries-1 {
				return false, fmt.Errorf("%s DELETE service error (status=%d): %s", schedulerName, resp.StatusCode, string(body))
			}
			h.logger.Warnf("%s DELETE attempt %d failed with status %d, retrying in %v", schedulerName, i+1, resp.StatusCode, backoff)
			time.Sleep(backoff)
			backoff *= 2
			continue
		}

		h.logger.Infof("Successfully sent DELETE to %s", schedulerName)
		return true, nil
	}

	return false, fmt.Errorf("failed to send DELETE to %s after %d retries", schedulerName, maxRetries)
}
