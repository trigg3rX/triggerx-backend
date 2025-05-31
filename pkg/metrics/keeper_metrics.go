package metrics

import (
	"net/http"
	// "strconv"
	"strings"
	// "time"

	"github.com/prometheus/client_golang/prometheus"
	// "github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/trigg3rX/triggerx-backend/pkg/database"
	"github.com/trigg3rX/triggerx-backend/pkg/logging"
)

// var (
// 	keeperPoints = promauto.NewGaugeVec(
// 		prometheus.GaugeOpts{
// 			Name: "triggerx_keeper_points",
// 			Help: "The total points accumulated by each keeper",
// 		},
// 		[]string{"keeper_id", "keeper_address"},
// 	)

// 	keeperTaskCount = promauto.NewGaugeVec(
// 		prometheus.GaugeOpts{
// 			Name: "triggerx_keeper_task_count",
// 			Help: "The number of tasks executed by each keeper",
// 		},
// 		[]string{"keeper_id", "keeper_address"},
// 	)

// 	totalKeepers = promauto.NewGauge(
// 		prometheus.GaugeOpts{
// 			Name: "triggerx_total_keepers",
// 			Help: "The total number of keepers in the system",
// 		},
// 	)
// )

type MetricsServer struct {
	db     *database.Connection
	logger logging.Logger
	done   chan bool
}

func NewMetricsServer(db *database.Connection, logger logging.Logger) *MetricsServer {
	return &MetricsServer{
		db:     db,
		logger: logger,
		done:   make(chan bool),
	}
}

func (m *MetricsServer) Start() {
	http.Handle("/metrics", promhttp.Handler())

	http.HandleFunc("/metrics/keeper", m.filteredMetricsHandler)

	go func() {
		m.logger.Infof("Starting metrics server on :8081")
		if err := http.ListenAndServe(":8081", nil); err != nil {
			m.logger.Errorf("Failed to start metrics server: %v", err)
		}
	}()
}

func (m *MetricsServer) Stop() {
	m.done <- true
}

// func (m *MetricsServer) collectMetrics() {
// 	ticker := time.NewTicker(15 * time.Second)
// 	defer ticker.Stop()

// 	for {
// 		select {
// 		case <-ticker.C:
// 			m.updateKeeperMetrics()
// 		case <-m.done:
// 			return
// 		}
// 	}
// }

// func (m *MetricsServer) updateKeeperMetrics() {
// 	iter := m.db.Session().Query(`
// 		SELECT keeper_id, keeper_address, no_executed_tasks, keeper_points 
// 		FROM triggerx.keeper_data
// 	`).Iter()

// 	var (
// 		keeperID      int64
// 		keeperAddress string
// 		taskCount     int
// 		points        float64
// 	)

// 	keeperCount := 0

// 	keeperPoints.Reset()
// 	keeperTaskCount.Reset()

// 	for iter.Scan(&keeperID, &keeperAddress, &taskCount, &points) {
// 		keeperIDStr := strconv.FormatInt(keeperID, 10)

// 		keeperPoints.WithLabelValues(keeperIDStr, keeperAddress).Set(float64(points))

// 		keeperTaskCount.WithLabelValues(keeperIDStr, keeperAddress).Set(float64(taskCount))

// 		keeperCount++
// 	}

// 	totalKeepers.Set(float64(keeperCount))

// 	if err := iter.Close(); err != nil {
// 		m.logger.Errorf("Error fetching keeper metrics: %v", err)
// 	}
// }

func (m *MetricsServer) filteredMetricsHandler(w http.ResponseWriter, r *http.Request) {
	keeperAddress := r.URL.Query().Get("address")
	if keeperAddress == "" {
		http.Error(w, "keeper address parameter 'address' is required", http.StatusBadRequest)
		return
	}

	if !strings.HasPrefix(keeperAddress, "0x") {
		keeperAddress = "0x" + keeperAddress
	}

	keeperAddress = strings.ToLower(keeperAddress)

	registry := prometheus.NewRegistry()

	var keeperID int64
	var taskCount int
	var points float64

	err := m.db.Session().Query(`
		SELECT keeper_id, no_executed_tasks, keeper_points 
		FROM triggerx.keeper_data
		WHERE keeper_address = ? ALLOW FILTERING
	`, keeperAddress).Scan(&keeperID, &taskCount, &points)

	if err != nil {
		http.Error(w, "Error fetching keeper data", http.StatusInternalServerError)
		return
	}

	keeperPointsMetric := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "triggerx_keeper_points",
		Help: "The total points accumulated by this keeper",
	})
	keeperTaskCountMetric := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "triggerx_keeper_task_count",
		Help: "The number of tasks executed by this keeper",
	})

	registry.MustRegister(keeperPointsMetric)
	registry.MustRegister(keeperTaskCountMetric)

	keeperPointsMetric.Set(float64(points))
	keeperTaskCountMetric.Set(float64(taskCount))

	h := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	h.ServeHTTP(w, r)
}
