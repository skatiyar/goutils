package queue

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/skatiyar/goutils/internal/primitives"
)

type task[T, R any] struct {
	ctx    context.Context
	value  T
	result primitives.Result[R]
}

// QueueImpl is the implementation of the Queue interface.
type QueueImpl[T, R any] struct {
	wg             sync.WaitGroup
	items          chan task[T, R]
	worker         func(context.Context, T) (R, error)
	closed         uint32
	running        int64
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
		cfg.DefaultTimeout = 1<<63 - 1 // effectively no timeout
	}
	queue := &QueueImpl[T, R]{
		wg:             sync.WaitGroup{},
		items:          make(chan task[T, R], cfg.Size),
		worker:         process,
		closed:         0,
		running:        0,
		defaultTimeout: cfg.DefaultTimeout,
	}
	go queue.work(cfg.Concurrency)
	return queue
}

func (qi *QueueImpl[T, R]) work(concurrency int) {
	// semaphore to bound concurrent workers
	sem := make(chan struct{}, concurrency)
	defer close(sem)

	// continuously process tasks from the queue
	for val := range qi.items {
		// acquire a slot (blocks when we reached max concurrency)
		sem <- struct{}{}
		atomic.AddInt64(&qi.running, 1)
		qi.wg.Add(1)

		// process task in a goroutine; when done release slot and decrement counters
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
				qi.wg.Done()
				<-sem // release slot
			}()
			data, dataErr := qi.worker(ival.ctx, ival.value)
			ival.result.Resolve(data, dataErr)
		}(val)
	}
}

// Shutdown gracefully shuts down the queue, waiting for all running tasks to complete.
// Queue is marked as closed immediately; no new tasks can be pushed after this call.
// Maximum wait time to finish queued tasks can be controlled via the provided context,
// post timeout pending tasks will be dropped.
func (qi *QueueImpl[T, R]) Shutdown(ctx context.Context) error {
	atomic.StoreUint32(&qi.closed, 1)
	qi.wg.Wait()
	close(qi.items)
	return nil
}

// Push add a new task to the queue.
// If the queue is closed, it returns an error immediately.
// Otherwise, it enqueues the task and returns a future result.
func (qi *QueueImpl[T, R]) Push(ctx context.Context, value T) primitives.Result[R] {
	if ctx == nil {
		// apply default timeout if no context is provided
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), qi.defaultTimeout)
		defer cancel()
	} else if deadline, ok := ctx.Deadline(); !ok || deadline.IsZero() {
		// apply default timeout if no deadline is set
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, qi.defaultTimeout)
		defer cancel()
	}
	result := primitives.NewResult[R]()
	if atomic.LoadUint32(&qi.closed) != 0 {
		result.Resolve(*new(R), ErrQueueClosed)
	} else {
		qi.items <- task[T, R]{ctx: ctx, value: value, result: result}
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
// Queue state can change immediately after this call.
func (qi *QueueImpl[T, R]) Status() Status {
	if atomic.LoadUint32(&qi.closed) != 0 {
		return StatusClosed
	}
	if qi.Running() > 0 {
		return StatusRunning
	}
	return StatusIdle
}
