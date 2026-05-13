package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	HTTPRequests = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "moistello_http_requests_total",
		Help: "Total HTTP requests",
	}, []string{"method", "path", "status"})

	HTTPDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "moistello_http_duration_seconds",
		Help:    "HTTP request duration in seconds",
		Buckets: prometheus.DefBuckets,
	}, []string{"method", "path"})

	CirclesCreated = promauto.NewCounter(prometheus.CounterOpts{
		Name: "moistello_circles_created_total",
		Help: "Total circles created",
	})

	ContributionsRecorded = promauto.NewCounter(prometheus.CounterOpts{
		Name: "moistello_contributions_total",
		Help: "Total contributions recorded",
	})

	ActiveUsers = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "moistello_active_users",
		Help: "Number of active users",
	})
)
