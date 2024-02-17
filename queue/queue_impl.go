package queue

import "sync"

type QueueImpl struct {
	wg sync.WaitGroup
}

func NewQueue(fn func() error, workers int) *QueueImpl {
	return &QueueImpl{}
}

func (qi *QueueImpl) Drain() {
	qi.wg.Wait()
}
