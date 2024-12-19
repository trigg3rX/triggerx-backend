package metrics

import (
	"github.com/Layr-Labs/eigensdk-go/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type Metrics interface {
	metrics.Metrics
	IncNumTasksReceived()
}

type AvsAndEigenMetrics struct {
	metrics.Metrics
	numTasksReceived prometheus.Counter
}

func NewAvsAndEigenMetrics(avsName string, eigenMetrics *metrics.EigenMetrics, reg *prometheus.Registry) *AvsAndEigenMetrics {
	return &AvsAndEigenMetrics{
		Metrics: eigenMetrics,
		numTasksReceived: promauto.With(reg).NewCounter(
			prometheus.CounterOpts{
				Namespace: "triggerx",
				Name:      "num_tasks_received",
				Help:      "The number of tasks received by reading from the avs service manager contract",
			}),
	}
}

func (m *AvsAndEigenMetrics) IncNumTasksReceived() {
	m.numTasksReceived.Inc()
}
