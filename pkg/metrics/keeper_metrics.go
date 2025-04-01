package metrics

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/trigg3rX/triggerx-backend/pkg/database"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

var (
	// Define Prometheus metrics
	keeperPoints = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "triggerx_keeper_points",
			Help: "The total points accumulated by each keeper",
		},
		[]string{"keeper_id", "keeper_address"},
	)

	keeperTaskCount = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "triggerx_keeper_task_count",
			Help: "The number of tasks executed by each keeper",
		},
		[]string{"keeper_id", "keeper_address"},
	)

	totalKeepers = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "triggerx_total_keepers",
			Help: "The total number of keepers in the system",
		},
	)
)

// MetricsServer handles the metrics collection and HTTP server
type MetricsServer struct {
	db     *database.Connection
	logger logging.Logger
	done   chan bool
}

// NewMetricsServer creates a new metrics server
func NewMetricsServer(db *database.Connection, logger logging.Logger) *MetricsServer {
	return &MetricsServer{
		db:     db,
		logger: logger,
		done:   make(chan bool),
	}
}

// Start begins the metrics collection and HTTP server
func (m *MetricsServer) Start() {
	// Register the default metrics handler
	http.Handle("/metrics", promhttp.Handler())

	// Register a filtered metrics handler
	http.HandleFunc("/metrics/keeper", m.filteredMetricsHandler)

	// Start the HTTP server on port 8081
	go func() {
		m.logger.Infof("Starting metrics server on :8081")
		if err := http.ListenAndServe(":8081", nil); err != nil {
			m.logger.Errorf("Failed to start metrics server: %v", err)
		}
	}()
}

// Stop signals the metrics collection to stop
func (m *MetricsServer) Stop() {
	m.done <- true
}

// collectMetrics periodically collects metrics from the database
func (m *MetricsServer) collectMetrics() {
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.updateKeeperMetrics()
		case <-m.done:
			return
		}
	}
}

// updateKeeperMetrics fetches keeper data from the database and updates Prometheus metrics
func (m *MetricsServer) updateKeeperMetrics() {
	// Query all keepers
	iter := m.db.Session().Query(`
		SELECT keeper_id, keeper_address, no_exctask, keeper_points 
		FROM triggerx.keeper_data
	`).Iter()

	var (
		keeperID      int64
		keeperAddress string
		taskCount     int
		points        int64
	)

	keeperCount := 0

	// Reset metrics before updating to handle removed keepers
	keeperPoints.Reset()
	keeperTaskCount.Reset()

	// Process each keeper
	for iter.Scan(&keeperID, &keeperAddress, &taskCount, &points) {
		keeperIDStr := strconv.FormatInt(keeperID, 10)

		// Update keeper points metric
		keeperPoints.WithLabelValues(keeperIDStr, keeperAddress).Set(float64(points))

		// Update keeper task count metric
		keeperTaskCount.WithLabelValues(keeperIDStr, keeperAddress).Set(float64(taskCount))

		keeperCount++
	}

	// Update total keepers metric
	totalKeepers.Set(float64(keeperCount))

	if err := iter.Close(); err != nil {
		m.logger.Errorf("Error fetching keeper metrics: %v", err)
	}
}

// filteredMetricsHandler provides metrics filtered by keeper_address
func (m *MetricsServer) filteredMetricsHandler(w http.ResponseWriter, r *http.Request) {
	keeperAddress := r.URL.Query().Get("address")
	if keeperAddress == "" {
		http.Error(w, "keeper address parameter 'address' is required", http.StatusBadRequest)
		return
	}

	// Create a registry for this specific keeper
	registry := prometheus.NewRegistry()

	// Query just this keeper's data
	var keeperID int64
	var taskCount int
	var points int64

	err := m.db.Session().Query(`
		SELECT keeper_id, no_exctask, keeper_points 
		FROM triggerx.keeper_data
		WHERE keeper_address = ? ALLOW FILTERING
	`, keeperAddress).Scan(&keeperID, &taskCount, &points)

	if err != nil {
		m.logger.Errorf("Error fetching keeper metrics for %s: %v", keeperAddress, err)
		http.Error(w, "Error fetching keeper data", http.StatusInternalServerError)
		return
	}

	// Create keeper-specific metrics
	keeperPointsMetric := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "triggerx_keeper_points",
		Help: "The total points accumulated by this keeper",
	})
	keeperTaskCountMetric := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "triggerx_keeper_task_count",
		Help: "The number of tasks executed by this keeper",
	})

	// Register metrics with this registry
	registry.MustRegister(keeperPointsMetric)
	registry.MustRegister(keeperTaskCountMetric)

	// Set values
	keeperPointsMetric.Set(float64(points))
	keeperTaskCountMetric.Set(float64(taskCount))

	// Generate response
	h := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	h.ServeHTTP(w, r)
}
