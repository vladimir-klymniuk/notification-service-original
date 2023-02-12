package producer

import (
	"context"
	"time"

	"github.com/Shopify/sarama"
	"github.com/rs/zerolog/log"
)

// Publisher publishes messages to a broker
type Publisher interface {
	Publish(ctx context.Context, message []byte) error
}

// publisher holds references to the producer (publisher),
// default headers, default key, and a timestamp function.
type publisher struct {
	key       string
	topic     string
	headers   []sarama.RecordHeader
	producer  sarama.AsyncProducer
	timestamp func() time.Time
}

// Options that modify publisher. Used in NewPublisher.
type Option func(*publisher)

// WithTime sets the publisher's timestamp function. Useful for unit tests.
func WithTime(f func() time.Time) Option {
	return func(p *publisher) {
		p.timestamp = f
	}
}

// NewPublisher returns a publisher that writes to the given topic
// at the addresses specified by the given brokers.
// A kafka producer is created and configured for use.
func NewPublisher(key, topic string, brokers []string, config *sarama.Config, options ...Option) (Publisher, error) {
	producer, err := sarama.NewAsyncProducer(brokers, config)
	if err != nil {
		return nil, err
	}

	// TODO: Handle this error with more elegance and thought. Maybe omit this.
	go func() {
		for err := range producer.Errors() {
			log.Warn().Msgf("failed to write message: %v", err)
		}
	}()

	// TODO: fill these from config; Maybe not required for our purposes,
	//  and we can remove the field from the publisher struct.
	// headers := []sarama.RecordHeader{
	// 	{
	// 		Key:   []byte("X-NS-TENANTID"),
	// 		Value: nil,
	// 	},
	// 	{
	// 		Key:   []byte("X-NS-SERVICE"),
	// 		Value: nil,
	// 	},
	// }

	p := &publisher{
		producer:  producer,
		key:       key,
		topic:     topic,
		timestamp: func() time.Time { return time.Now().UTC() },
	}

	for _, option := range options {
		option(p)
	}

	return p, nil
}

// Publish creates a ProducerMessage from the provided message and writes it
// to the producer's input channel.
func (p *publisher) Publish(_ context.Context, message []byte) error {
	m := &sarama.ProducerMessage{
		Topic: p.topic,
		// Key:       sarama.StringEncoder(p.key),
		Value:     sarama.ByteEncoder(message),
		Headers:   p.headers,
		Timestamp: p.timestamp(),
	}

	log.Info().Msgf("new message: %s", message)

	p.producer.Input() <- m

	return nil
}

// Close wraps the producer's Close method.
func (p *publisher) Close() error {
	return p.producer.Close()
}