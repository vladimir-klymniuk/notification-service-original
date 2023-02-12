package notify

import (
	"context"
	"github.com/pkg/errors"
	"github.com/vladimir-klymniuk/notification-service-original/message"
)

// errPublishing is raised when an error occurs with the publisher.
var errPublishing = errors.New("error publishing message")

// errPublishing is raised when an error occurs with the encoder.
var errEncoding = errors.New("error encoding message")

// Service ...
type Service interface {
	Send(ctx context.Context, httpRequest string) error
}

// Publisher publishes messages
type Publisher interface {
	Publish(context.Context, []byte) error
}

// Encoder encodes messages
type Encoder interface {
	Encode(context.Context, message.Message) ([]byte, error)
}

type service struct {
	publisher Publisher
	encoder   Encoder
}

// NewService returns an instance of a new notifier service.
// Requires an injected publisher and encoder services.
func NewService(publisher Publisher, encoder Encoder) Service {
	return &service{
		publisher: publisher,
		encoder:   encoder,
	}
}

// Send wraps the provided message inside a JSON object, and publishes
// it with the injected publisher service.
func (s *service) Send(ctx context.Context, httpRequest string) error {
	m := message.Message{
		Type:        message.TypeHTTPGet,
		HTTPRequest: httpRequest,
	}

	b, err := s.encoder.Encode(ctx, m)
	if err != nil {
		return errors.Wrap(errEncoding, err.Error())
	}

	if err = s.publisher.Publish(ctx, b); err != nil {
		return errors.Wrap(errPublishing, err.Error())
	}

	return nil
}