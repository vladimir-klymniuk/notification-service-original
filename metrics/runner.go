package metrics

import (
	"context"
	"strconv"
	"time"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/vladimir-klymniuk/notification-service-original/runner"
)

var rs *runnerService

type runnerService struct {
	request        *prometheus.CounterVec
	requestSuccess *prometheus.CounterVec
	requestFailed  *prometheus.CounterVec
	latency        *prometheus.HistogramVec

	serviceName string
}

func init() {
	rs = runnerMiddleware()
}

func runnerMiddleware() *runnerService {
	var m runnerService

	fieldKeys := []string{"service", "attempt"}

	m.request = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "runner",
			Subsystem: "service",
			Name:      "count",
			Help:      "Number of tasks processed",
		}, fieldKeys)
	prometheus.MustRegister(m.request)

	m.requestFailed = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "runner",
			Subsystem: "service",
			Name:      "failed_count",
			Help:      "Number of tasks failed",
		}, fieldKeys)
	prometheus.MustRegister(m.requestFailed)

	m.requestSuccess = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "runner",
			Subsystem: "service",
			Name:      "success_count",
			Help:      "Number of tasks succseed",
		}, fieldKeys)
	prometheus.MustRegister(m.requestSuccess)

	m.latency = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "runner",
			Subsystem: "service",
			Name:      "duration_milliseconds",
			Help:      "Total duration in milliseconds.",
			Buckets:   []float64{1, 5, 10, 50, 100, 150, 250, 500, 1000},
		}, fieldKeys)
	prometheus.MustRegister(m.latency)

	return &m
}

// RunnerBuilder create Runners and attaches serviceName to them
type RunnerBuilder struct {
	b           *runner.Builder
	serviceName string
}

// NewRunnerBuilder creates RunnerBuilder using provided builder
func NewRunnerBuilder(rb *runner.Builder, serviceName string) *RunnerBuilder {
	return &RunnerBuilder{
		b:           rb,
		serviceName: serviceName,
	}
}

// CreateRunner creates Runner
func (rb *RunnerBuilder) CreateRunner() runner.Runner {
	r := rb.b.CreateRunner()

	return &mrunner{
		next:        r,
		serviceName: rb.serviceName,
	}
}

type mrunner struct {
	next        runner.Runner
	serviceName string
}

// Execute executes task using internal runner
func (r *mrunner) Execute(ctx context.Context, task runner.Task) (n int, err error) {
	start := time.Now()
	// add metrics to this method
	defer func() {
		v := time.Since(start).Seconds() * 1e3
		rs.latency.WithLabelValues(r.serviceName, strconv.Itoa(n)).Observe(v)
	}()
	defer rs.request.WithLabelValues(r.serviceName, strconv.Itoa(n)).Inc()

	n, err = r.next.Execute(ctx, task)
	if err != nil {
		rs.requestFailed.WithLabelValues(r.serviceName, strconv.Itoa(n)).Inc()
	} else {
		rs.requestSuccess.WithLabelValues(r.serviceName, strconv.Itoa(n)).Inc()
	}

	return n, err
}