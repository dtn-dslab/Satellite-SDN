package metrics

import "github.com/prometheus/client_golang/prometheus"

type LatencyHistograms struct {
	Histograms *prometheus.HistogramVec
}

func NewLatencyHistograms() *LatencyHistograms {
	return &LatencyHistograms{
		Histograms: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "kubedtnd_request_duration_milliseconds",
				Help:    "Latency of requests in milliseconds",
				Buckets: []float64{0, 1, 5, 10, 20, 50, 100, 200, 500, 1000, 2000, 5000},
			},
			[]string{"method"},
		),
	}
}

func (l *LatencyHistograms) Observe(method string, latency int64) {
	l.Histograms.WithLabelValues(method).Observe(float64(latency))
}

func (l *LatencyHistograms) Register(registry *prometheus.Registry) {
	registry.MustRegister(
		l.Histograms,
	)
}
