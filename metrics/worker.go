package metrics

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Shopify/sarama"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/tkanos/konsumerou"
)

const (
	requestName       = "request_total"
	requestFailedName = "request_failed"
	latencyName       = "request_latency_milliseconds"
)

type metricsWorker struct {
	request       *prometheus.CounterVec
	requestFailed *prometheus.CounterVec
	latency       *prometheus.SummaryVec
	serviceName   string
}

// NewMetricsService creates a layer of service that add metrics capability
func NewMetricssWorker(serviceName string, next konsumerou.Handler) konsumerou.Handler {
	m := metricsMiddlewareWorker(serviceName)
	return m.instrumentation(next)
}

func metricsMiddlewareWorker(name string) *metricsWorker {
	var m metricsWorker

	fieldKeys := []string{"service"}

	m.request = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "consumer",
			Subsystem: "service",
			Name:      fmt.Sprintf("%v_%v", strings.Replace(name, "-", "_", -1), requestName),
			Help:      "Number of requests processed",
		}, fieldKeys)
	prometheus.MustRegister(m.request)

	m.requestFailed = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "consumer",
			Subsystem: "service",
			Name:      fmt.Sprintf("%v_%v", strings.Replace(name, "-", "_", -1), requestFailedName),
			Help:      "Number of requests failed",
		}, fieldKeys)
	prometheus.MustRegister(m.requestFailed)

	m.latency = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: "consumer",
			Subsystem: "service",
			Name:      fmt.Sprintf("%v_%v", strings.Replace(name, "-", "_", -1), latencyName),
			Help:      "Total duration in miliseconds.",
		}, fieldKeys)
	prometheus.MustRegister(m.latency)

	m.serviceName = name

	return &m
}

func (m *metricsWorker) instrumentation(next konsumerou.Handler) konsumerou.Handler {
	return func(ctx context.Context, msg *sarama.ConsumerMessage) (err error) {
		start := time.Now()
		// add metrics to this method
		defer m.latency.WithLabelValues(m.serviceName).Observe(time.Since(start).Seconds() * 1e3)
		defer m.request.WithLabelValues(m.serviceName).Inc()

		// If error is not empty, we add to metrics that it failed
		err = next(ctx, msg)
		if err != nil {
			m.requestFailed.WithLabelValues(m.serviceName).Inc()
		}

		return
	}
}
