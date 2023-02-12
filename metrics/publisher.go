package metrics

import (
	"context"
	"github.com/prometheus/client_golang/prometheus"
)

// singleton of prometheus reporting service
var ps *publisherService

// publisherService reports metrics to prometheus
type publisherService struct {
	topic   *prometheus.CounterVec
	service *prometheus.CounterVec
}

func init() {
	ps = publisherMiddleware()
}

// publisherMiddleware initializes the publisherService singleton
func publisherMiddleware() *publisherService {
	var m publisherService

	m.topic = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "publisher",
			Subsystem: "topic",
			Name:      "count",
			Help:      "Topic count",
		},
		[]string{"topic"},
	)
	prometheus.MustRegister(m.topic)

	m.service = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "publisher",
			Subsystem: "service",
			Name:      "count",
			Help:      "Service count",
		},
		[]string{"service"},
	)
	prometheus.MustRegister(m.service)

	return &m
}

// Publisher interface to wrap
type Publisher interface {
	Publish(ctx context.Context, message []byte) error
}

// publisher middleware struct
type publisher struct {
	next    Publisher
	topic   string
	service string
}

// NewPublisher creates a new middleware for metrics reporting for Publisher
func NewPublisher(p Publisher, topic, service string) Publisher {
	return &publisher{
		next:    p,
		topic:   topic,
		service: service,
	}
}

// Send wraps the Publisher's Send fucntionality and reports metrics on topic and service
func (h *publisher) Publish(ctx context.Context, httpRequest []byte) error {
	err := h.next.Publish(ctx, httpRequest)
	if err == nil {
		ps.topic.WithLabelValues(h.topic).Inc()

		ps.service.WithLabelValues(h.service).Inc()
	}

	return err
}