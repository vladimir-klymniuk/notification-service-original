package runner

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_newRunner(t *testing.T) {
	retries := 1
	wait := 1 * time.Millisecond

	r := newRunner(retries, wait)

	assert.NotNil(t, r)
	assert.Equal(t, retries, r.retries)
	assert.Equal(t, wait, r.wait)
}

func Test_runner_Execute_Should_Return_Nil_When_Success(t *testing.T) {
	ctx := context.Background()
	s := newRunner(1, 1*time.Microsecond)

	task := func() error { return nil }

	_, got := s.Execute(ctx, task)

	assert.Nil(t, got)
}

func Test_runner_Execute_Should_Return_Err_When_Task_Is_Emtpy(t *testing.T) {
	ctx := context.Background()
	s := newRunner(0, 1*time.Microsecond)

	_, got := s.Execute(ctx, nil)

	assert.Equal(t, ErrNilTask, got)
}

func Test_runner_Execute_Should_Return_Err_When_Retries_Is_0(t *testing.T) {
	ctx := context.Background()
	s := newRunner(0, 1*time.Microsecond)
	task := func() error { return nil }

	_, got := s.Execute(ctx, task)

	assert.Equal(t, ErrTaskFail, got)
}

func Test_runner_Execute_Should_Return_Err_When_Task_Returns_Err(t *testing.T) {
	ctx := context.Background()
	s := newRunner(1, 1*time.Microsecond)

	err := errors.New("task")

	task := func() error { return err }

	_, got := s.Execute(ctx, task)

	assert.Equal(t, err, got)
}

func Test_runner_Execute_Should_Return_Nil_When_Task_Returns_No_Error_On_Next_Attempt(t *testing.T) {
	ctx := context.Background()
	s := newRunner(2, 1*time.Microsecond)

	count := 0
	task := func() error {
		if count == 0 {
			count++
			return errors.New("firt attempt")
		}

		return nil
	}

	n, got := s.Execute(ctx, task)

	assert.Equal(t, 1, n)
	assert.Nil(t, got)
}