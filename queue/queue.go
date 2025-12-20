package queue

import (
	"context"
	"errors"
	"time"

	"github.com/skatiyar/goutils/internal/primitives"
)

type Status int

const (
	StatusIdle Status = iota
	StatusRunning
	StatusClosed
)

type Config struct {
	Size           int           // size of the queue buffer, less than equal to 0: defaults to 100
	Concurrency    int           // number of concurrent workers, less than equal to 0: defaults to 10
	DefaultTimeout time.Duration // default timeout for push operations, less than equal to 0: defaults to no timeout
}

type Queue[T, R any] interface {
	Push(ctx context.Context, value T) primitives.Result[R]
	Shutdown(ctx context.Context) error
	Queued() int
	Running() int
	Status() Status
}

var (
	ErrQueueClosed = errors.New("queue is closed")
	ErrPushTimeout = errors.New("push timeout exceeded")
)
