package queue

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/skatiyar/goutils/internal/primitives"
)

const DefaultTimeout = 1<<63 - 1 // effectively no timeout

type task[T, R any] struct {
	ctx       context.Context
	ctxCancel context.CancelFunc
	value     T
	result    primitives.Result[R]
}

// QueueImpl is the implementation of the Queue interface.
type QueueImpl[T, R any] struct {
	items          chan task[T, R]
	signalClose    chan struct{}
	exitChan       chan struct{}
	worker         func(context.Context, T) (R, error)
	closed         uint32
	running        int64
	concurrency    int
	defaultTimeout time.Duration
}

// New creates a new Queue with the given configuration and processing function.
func New[T, R any](
	cfg Config,
	process func(context.Context, T) (R, error),
) Queue[T, R] {
	if cfg.Size <= 0 {
		cfg.Size = 100
	}
	if cfg.Concurrency <= 0 {
		cfg.Concurrency = 10
	}
	if cfg.DefaultTimeout <= 0 {
		cfg.DefaultTimeout = DefaultTimeout // effectively no timeout
	}
	queue := &QueueImpl[T, R]{
		items:          make(chan task[T, R], cfg.Size),
		signalClose:    make(chan struct{}),
		exitChan:       make(chan struct{}),
		worker:         process,
		closed:         0,
		running:        0,
		concurrency:    cfg.Concurrency,
		defaultTimeout: cfg.DefaultTimeout,
	}
	go queue.work(cfg.Concurrency)
	return queue
}

func (qi *QueueImpl[T, R]) work(concurrency int) {
	// semaphore to bound concurrent workers
	sem := make(chan struct{}, concurrency)
	defer func() {
		defer close(qi.items)
		defer close(sem)
		qi.exitChan <- struct{}{}
	}()

	// continuously process tasks from the queue
	for {
		if qi.isClosed() && len(qi.items) == 0 {
			return // exit if closed and no more items
		}
		select {
		case <-qi.signalClose:
			atomic.StoreUint32(&qi.closed, 1)
		default:
			sem <- struct{}{}
			val, ok := <-qi.items
			if !ok {
				return // queue closed
			}
			atomic.AddInt64(&qi.running, 1)
			go func(ival task[T, R]) {
				defer func() {
					if r := recover(); r != nil {
						if err, ok := r.(error); ok {
							ival.result.Resolve(*new(R), err)
						} else {
							ival.result.Resolve(*new(R), fmt.Errorf("panic in worker %v", r))
						}
					}
					atomic.AddInt64(&qi.running, -1)
					ival.ctxCancel()
					<-sem // release slot
				}()
				select {
				case <-ival.ctx.Done():
					ival.result.Resolve(*new(R), ival.ctx.Err())
					return
				default:
					data, dataErr := qi.worker(ival.ctx, ival.value)
					ival.result.Resolve(data, dataErr)
				}
			}(val)
		}
	}
}

// Shutdown gracefully shuts down the queue, waiting for all running tasks to complete.
// Queue is marked as closed immediately; no new tasks can be pushed after this call.
// Maximum wait time to finish queued tasks can be controlled via the provided context,
// post timeout pending tasks will be dropped.
func (qi *QueueImpl[T, R]) Shutdown(ctx context.Context) error {
	newCtx, ctxCancel := qi.context(ctx)
	defer ctxCancel()
	qi.signalClose <- struct{}{}
	defer close(qi.signalClose)
	select {
	case <-newCtx.Done():
		return newCtx.Err()
	case <-qi.exitChan:
		close(qi.exitChan)
	}
	return nil
}

// context prepares a context with default timeout if needed.
func (qi *QueueImpl[T, R]) context(ctx context.Context) (context.Context, context.CancelFunc) {
	if ctx == nil {
		// apply default timeout if no context is provided
		return context.WithTimeout(context.Background(), qi.defaultTimeout)
	} else if _, ok := ctx.Deadline(); !ok {
		// apply default timeout if no deadline is set
		return context.WithTimeout(ctx, qi.defaultTimeout)
	} else {
		// use provided context as is and let caller handle timeout/cancellation
		return ctx, func() {}
	}
}

// Push add a new task to the queue.
// If the queue is closed, it returns an error immediately.
// Otherwise, it enqueues the task and returns a future result.
func (qi *QueueImpl[T, R]) Push(ctx context.Context, value T) primitives.Result[R] {
	newCtx, ctxCancel := qi.context(ctx)
	result := primitives.NewResult[R]()

	if qi.isClosed() {
		defer ctxCancel()
		result.Resolve(*new(R), ErrQueueClosed)
		return result
	}

	select {
	case qi.items <- task[T, R]{ctx: newCtx, ctxCancel: ctxCancel, value: value, result: result}:
		// successfully enqueued
	case <-newCtx.Done():
		defer ctxCancel()
		result.Resolve(*new(R), ErrPushTimeout)
	}
	return result
}

// Queued returns the number of tasks currently queued in the queue.
func (qi *QueueImpl[T, R]) Queued() int {
	return len(qi.items)
}

// Running returns the number of tasks currently being processed by the queue.
func (qi *QueueImpl[T, R]) Running() int {
	return int(atomic.LoadInt64(&qi.running))
}

// Status returns the current status of the queue.
// If there are no tasks in the queue and no tasks are being processed, the status is StatusIdle.
// If there are tasks being processed, the status is StatusRunning.
// If the queue has been closed, the status is StatusClosed.
// This method provides a quick way to check the state of the queue for debugging and monitoring purposes.
// Queue state can change between StatusRunning & StatusIdle right after this call, but StatusClosed is final.
func (qi *QueueImpl[T, R]) Status() Status {
	if qi.isClosed() {
		return StatusClosed
	}
	if qi.Running() > 0 || qi.Queued() > 0 {
		return StatusRunning
	}
	return StatusIdle
}

// Config returns the actual configuration of the queue.
func (qi *QueueImpl[T, R]) Config() Config {
	return Config{
		Size:           cap(qi.items),
		Concurrency:    qi.concurrency,
		DefaultTimeout: qi.defaultTimeout,
	}
}

func (qi *QueueImpl[T, R]) isClosed() bool {
	return atomic.LoadUint32(&qi.closed) != 0
}
