package async

import (
	"sync"
)

type opresult[A, B any] struct {
	Key   A
	Value B
	Error error
}

func EachSlice[T any](collection []T, fn func(idx int, value T)) {
	wg := sync.WaitGroup{}
	for idx := range collection {
		wg.Add(1)
		go func(i int, val T) {
			defer wg.Done()
			fn(i, val)
		}(idx, collection[idx])
	}
	wg.Wait()
}

func EachSliceLimit[T any](collection []T, fn func(idx int, value T), limit int) {
	wg := sync.WaitGroup{}
	guard := make(chan struct{}, limit)
	defer close(guard)
	for idx := range collection {
		wg.Add(1)
		guard <- struct{}{}
		go func(i int, val T) {
			defer wg.Done()
			fn(i, val)
			<-guard
		}(idx, collection[idx])
	}
	wg.Wait()
}

func Slice[T any, S any](collection []T, fn func(val T) S) []S {
	result := make([]S, len(collection))
	resultChan := make(chan opresult[int, S])
	wg := sync.WaitGroup{}
	for idx := range collection {
		wg.Add(1)
		go func(i int, val T) {
			defer wg.Done()
			resultChan <- opresult[int, S]{
				Key:   i,
				Value: fn(val),
			}
		}(idx, collection[idx])
	}
	go func() {
		wg.Wait()
		close(resultChan)
	}()
	for resVal := range resultChan {
		result[resVal.Key] = resVal.Value
	}
	return result
}

func SliceLimit[T any, S any](collection []T, fn func(val T) S, limit int) []S {
	result := make([]S, len(collection))
	resultChan := make(chan opresult[int, S])
	wg := sync.WaitGroup{}
	guard := make(chan struct{}, limit)
	defer close(guard)
	for idx := range collection {
		wg.Add(1)
		guard <- struct{}{}
		go func(i int, val T) {
			defer wg.Done()
			// Guard needs to be received before sending result to prevent deadlock.
			// As results channel is not buffered and guard will block for loop
			// till existing go routines are able to send on result channel
			<-guard
			resultChan <- opresult[int, S]{
				Key:   i,
				Value: fn(val),
			}
		}(idx, collection[idx])
	}
	go func() {
		wg.Wait()
		close(resultChan)
	}()
	for resVal := range resultChan {
		result[resVal.Key] = resVal.Value
	}
	return result
}
