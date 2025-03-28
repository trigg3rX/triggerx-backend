package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"
)

// KeeperData represents the structure of keeper information
type KeeperData struct {
	KeeperID          int64  `json:"keeper_id"`
	KeeperAddress     string `json:"keeper_address"`
	ConnectionAddress string `json:"connection_address"`
	Status            bool   `json:"status"`
	Verified          bool   `json:"verified"`
}

// KeeperHealth represents the health status of a keeper
type KeeperHealth struct {
	KeeperID          int64     `json:"keeper_id"`
	ConnectionAddress string    `json:"connection_address"`
	IsHealthy         bool      `json:"is_healthy"`
	LastChecked       time.Time `json:"last_checked"`
	Error             string    `json:"error,omitempty"`
}

// KeeperHealthMonitor manages health checks for keepers
type KeeperHealthMonitor struct {
	keepers      []KeeperData
	healthStatus map[int64]KeeperHealth
	mu           sync.RWMutex
}

// NewKeeperHealthMonitor creates a new health monitor
func NewKeeperHealthMonitor() *KeeperHealthMonitor {
	return &KeeperHealthMonitor{
		healthStatus: make(map[int64]KeeperHealth),
	}
}

// FetchKeepers retrieves keeper information from the API
func (m *KeeperHealthMonitor) FetchKeepers() error {
	resp, err := http.Get("https://data.triggerx.network/api/keepers/all")
	if err != nil {
		return fmt.Errorf("failed to fetch keepers: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %v", err)
	}

	var keepers []KeeperData
	if err := json.Unmarshal(body, &keepers); err != nil {
		return fmt.Errorf("failed to unmarshal keepers: %v", err)
	}

	m.mu.Lock()
	m.keepers = keepers
	m.mu.Unlock()

	return nil
}

// CheckKeeperHealth checks the health of a single keeper
func (m *KeeperHealthMonitor) CheckKeeperHealth(keeper KeeperData) KeeperHealth {
	health := KeeperHealth{
		KeeperID:          keeper.KeeperID,
		ConnectionAddress: keeper.ConnectionAddress,
		LastChecked:       time.Now(),
	}

	// Skip health check if no connection address or not verified/active
	if keeper.ConnectionAddress == "" || !keeper.Verified || !keeper.Status {
		health.IsHealthy = false
		health.Error = "Invalid or inactive keeper"
		return health
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/health", keeper.ConnectionAddress), nil)
	if err != nil {
		health.IsHealthy = false
		health.Error = fmt.Sprintf("Request creation error: %v", err)
		return health
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		health.IsHealthy = false
		health.Error = fmt.Sprintf("Health check failed: %v", err)
		return health
	}
	defer resp.Body.Close()

	health.IsHealthy = resp.StatusCode == http.StatusOK
	if !health.IsHealthy {
		health.Error = fmt.Sprintf("Unexpected status code: %d", resp.StatusCode)
	}

	return health
}

// MonitorKeeperHealth periodically checks health of all keepers
func (m *KeeperHealthMonitor) MonitorKeeperHealth() {
	for {
		// Fetch latest keeper information
		if err := m.FetchKeepers(); err != nil {
			log.Printf("Error fetching keepers: %v", err)
			time.Sleep(30 * time.Second)
			continue
		}

		// Check health for each keeper
		var wg sync.WaitGroup
		healthResults := make(chan KeeperHealth, len(m.keepers))

		m.mu.RLock()
		for _, keeper := range m.keepers {
			wg.Add(1)
			go func(k KeeperData) {
				defer wg.Done()
				health := m.CheckKeeperHealth(k)
				healthResults <- health
			}(keeper)
		}
		m.mu.RUnlock()

		// Wait for all health checks to complete
		go func() {
			wg.Wait()
			close(healthResults)
		}()

		// Update health status
		m.mu.Lock()
		for health := range healthResults {
			m.healthStatus[health.KeeperID] = health

			// Log health status
			if !health.IsHealthy {
				log.Printf("Keeper %d at %s is NOT HEALTHY. Error: %s",
					health.KeeperID, health.ConnectionAddress, health.Error)
			} else {
				log.Printf("Keeper %d at %s is healthy",
					health.KeeperID, health.ConnectionAddress)
			}
		}
		m.mu.Unlock()

		// Wait before next health check
		time.Sleep(30 * time.Second)
	}
}

// PrintHealthStatus prints the current health status of keepers
func (m *KeeperHealthMonitor) PrintHealthStatus() {
	m.mu.RLock()
	defer m.mu.RUnlock()

	fmt.Println("\n--- Keeper Health Status ---")
	for id, health := range m.healthStatus {
		status := "Healthy"
		if !health.IsHealthy {
			status = "Unhealthy"
		}
		fmt.Printf("Keeper %d: %s | Address: %s | Last Checked: %v\n",
			id, status, health.ConnectionAddress, health.LastChecked)
		if !health.IsHealthy {
			fmt.Printf("  Error: %s\n", health.Error)
		}
	}
}

func main() {
	monitor := NewKeeperHealthMonitor()

	// Start monitoring in a goroutine
	go monitor.MonitorKeeperHealth()

	// Optional: Periodic status printing
	go func() {
		for {
			time.Sleep(60 * time.Second)
			monitor.PrintHealthStatus()
		}
	}()

	// Keep the main goroutine running
	select {}
}
