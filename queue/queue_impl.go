package queue

import (
	"errors"
	"sync"
)

var (
	ErrorQueueClosed = "queue has been closed"
)

type task[T any] struct {
	value         T
	errorCallback func(error)
}

type QueueImpl[T any] struct {
	wg          sync.WaitGroup
	items       chan task[T]
	worker      func(T) error
	concurrency int
	closed      bool
}

func NewQueue[T any](fn func(T) error, concurrency int) *QueueImpl[T] {
	queue := &QueueImpl[T]{
		wg:          sync.WaitGroup{},
		items:       make(chan task[T]),
		worker:      fn,
		concurrency: concurrency,
	}
	go queue.workers()
	return queue
}

func (qi *QueueImpl[T]) workers() {
	for {
		select {
		case val, ok := <-qi.items:
			if !ok {
				return
			}
			if err := qi.worker(val.value); err != nil {
				val.errorCallback(err)
			}
		default:

		}
	}
}

func (qi *QueueImpl[T]) Drain() {
	qi.wg.Wait()
	close(qi.items)
}

// Push add a new task to the queue. Calls callback once the worker has finished processing the task.
func (qi *QueueImpl[T]) Push(value T, callback func(err error)) {
	if !qi.closed {
		qi.items <- task[T]{value: value, errorCallback: callback}
	} else {
		callback(errors.New(ErrorQueueClosed))
	}
}
