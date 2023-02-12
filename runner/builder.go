package runner

import (
	"context"
	"time"
)

type Runner interface {
	Execute(context.Context, Task) (int, error)
}

type Builder struct {
	retries int
	wait    time.Duration
}

func NewBuilder(retries int, wait time.Duration) *Builder {
	return &Builder{
		retries: retries,
		wait:    wait,
	}
}

func (b *Builder) CreateRunner() Runner {
	return newRunner(b.retries, b.wait)
}