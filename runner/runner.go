package runner

import (
	"context"

	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

// ErrTaskFail is raised when task is failed or was not executed.
var ErrTaskFail = errors.New("fail")

// ErrNilTask is raised when task is nil.
var ErrNilTask = errors.New("task is nil")

// Task is function to execute.
type Task func() error

// runner executes tasks.
type runner struct {
	retries int
	wait    time.Duration
}

// NewRunner creates runner for task execution
// retries is number of attempts with timeout between them.
func newRunner(retries int, wait time.Duration) *runner {
	return &runner{
		retries: retries,
		wait:    wait,
	}
}

// Execute executes given task.
func (s *runner) Execute(ctx context.Context, t Task) (int, error) {
	if t == nil {
		return 0, ErrNilTask
	}
	//
	err := ErrTaskFail

	for i := 0; i < s.retries; i++ {
		// execute task
		err = t()
		if err == nil {
			// task was executed successfully
			// no error to return
			return i, nil
		}
		// wait
		select {
		// wait before next attempt
		case <-time.After(s.wait):
		// cancel runner
		case <-ctx.Done():
			return 0, ctx.Err()
		}

		log.Info().Msgf("retries: %d", i)
	}

	// return fail
	return s.retries, err
}
