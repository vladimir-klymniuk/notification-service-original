package httpget

import (
	"context"
	"net/http"

	"github.com/Shopify/sarama"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/tkanos/konsumerou"
	"github.com/vladimir-klymniuk/notification-service-original/message"
	"github.com/vladimir-klymniuk/notification-service-original/runner"
)

// Worker processes messages.
type Worker interface {
	Process(context.Context, []byte) error
}

// MakeWMakeWorkerEndpoint creates handler.
func MakeWorkerEndpoint(s Worker) konsumerou.Handler {
	return func(ctx context.Context, msg *sarama.ConsumerMessage) error {
		return s.Process(ctx, msg.Value)
	}
}

// Sender sends http requests.
type Sender interface {
	Do(*http.Request) (*http.Response, error)
}

// Publisher publishes data.
type Publisher interface {
	Publish(context.Context, []byte) error
}

//
type Builder interface {
	CreateRunner() runner.Runner
}

// Decoder decodes message from bytes.
type Decoder interface {
	Decode(context.Context, []byte) (message.Message, error)
}

type worker struct {
	runners      chan runner.Runner
	decoder      Decoder
	sender       Sender
	rbuilder     Builder
	errPublisher Publisher
}

// NewWorker creates worker.
func NewWorker(sender Sender, errPublisher Publisher, decoder Decoder, number int, builder Builder) Worker {
	w := &worker{
		sender:       sender,
		decoder:      decoder,
		rbuilder:     builder,
		errPublisher: errPublisher,
	}

	w.runners = createRunners(w.rbuilder, number)

	return w
}

// Process processes message.
func (w *worker) Process(ctx context.Context, msg []byte) error {
	m, err := w.decoder.Decode(ctx, msg)
	if err != nil {
		log.Error().Err(err).Bytes("data", msg).Msg("unable to decode message")
		return err
	}

	log.Info().Msgf("process: %s", string(msg))

	// get free runner
	r := <-w.runners

	go func(ctx context.Context, r runner.Runner) {
		// put sender back to queue
		defer func() {
			w.runners <- r
		}()

		// create Task
		task := w.createTask(m.HTTPRequest)
		// execute task
		n, err := r.Execute(ctx, task)
		if err != nil {
			log.Error().Err(err).Int("attempt", n).Msg(m.HTTPRequest)

			w.logError(ctx, errors.Wrapf(err, "request: %s : attempt: %d ", m.HTTPRequest, n))
		}
	}(ctx, r)

	return nil
}

// logError logs error.
func (w *worker) logError(ctx context.Context, err error) {
	b := []byte(err.Error())

	err = w.errPublisher.Publish(ctx, b)
	if err != nil {
		log.Error().Err(err).Msg("unable to log error")
	}
}

// createTask creates task to execute.
func (w *worker) createTask(url string) func() error {
	return func() error {
		return send(w.sender, url)
	}
}

// createRunners creates chan of runners, which will execute tasks.
func createRunners(b Builder, number int) chan runner.Runner {
	runners := make(chan runner.Runner, number)
	// create runners
	for i := 0; i < number; i++ {
		runners <- b.CreateRunner()
	}

	return runners
}

// send sends GET request to url.
func send(sender Sender, url string) error {
	log.Info().Msgf("http GET: %s", url)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	r, err := sender.Do(req)
	if err != nil {
		return err
	}

	// close body
	r.Body.Close()

	// check status code for 200 OK
	if r.StatusCode != http.StatusOK {
		return errors.Wrapf(ErrStatusCode, "%d - %s", r.StatusCode, r.Status)
	}

	return nil
}
