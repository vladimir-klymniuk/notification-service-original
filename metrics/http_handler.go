package metrics

import (
	"net/http"
	"time"
	"github.com/prometheus/client_golang/prometheus"
)

type httpService struct {
	latency     *prometheus.HistogramVec
	serviceName string
}

var ms *httpService

func init() {
	ms = metricsMiddleware("notification-service")
}

// NewHTTPMiddleware wraps an http.HandlerFunc to report metrics regarding HTTP requests.
// HTTP request duration and status code are reported.
func NewHTTPMiddleware(name string, handler http.HandlerFunc) http.HandlerFunc {
	return ms.chain(name, handler)
}

func metricsMiddleware(name string) *httpService {
	var m httpService
	fieldKeys := []string{"method", "endpoint"}

	m.latency = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "http",
			Subsystem: "request",
			Name:      "duration_milliseconds",
			Help:      "Total duration in milliseconds.",
			Buckets:   []float64{1, 1.5, 2, 2.5, 3, 3.5, 4, 5, 10, 50, 100},
		}, fieldKeys)
	prometheus.MustRegister(m.latency)

	m.serviceName = name

	return &m
}

func (m *httpService) chain(name string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		// add metrics to this method
		defer func(start time.Time) {
			v := float64(time.Since(start).Seconds() * 1e3)
			m.latency.WithLabelValues(m.serviceName, name).Observe(v)
		}(start)

		wh := &responseWriter{
			ResponseWriter: w,
		}

		next(wh, r)
	}
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader wraps the http.ResponseWriter in order to extract
// the status code into this package so it can be reported.
func (w *responseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}