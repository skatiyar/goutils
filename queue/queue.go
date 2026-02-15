/*
Package queue provides a concurrent work queue with configurable concurrency and buffering.

The queue processes tasks asynchronously with a fixed number of workers, providing
backpressure through a buffered channel. Each pushed task returns a Result that
can be awaited for completion.
*/
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
	Config() Config
}

var (
	ErrQueueClosed = errors.New("queue is closed")
	ErrPushTimeout = errors.New("push timeout exceeded")
)
