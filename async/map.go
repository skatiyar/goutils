package async

import (
	"sync"
)

func EachMap[A comparable, B any](collection map[A]B, fn func(key A, value B)) {
	wg := sync.WaitGroup{}
	for key, value := range collection {
		wg.Add(1)
		go func(k A, v B) {
			defer wg.Done()
			fn(k, v)
		}(key, value)
	}
	wg.Wait()
}

func EachMapLimit[A comparable, B any](collection map[A]B, fn func(key A, value B), limit int) {
	wg := sync.WaitGroup{}
	gaurd := make(chan struct{}, limit)
	defer close(gaurd)
	for key, value := range collection {
		wg.Add(1)
		gaurd <- struct{}{}
		go func(k A, v B) {
			defer wg.Done()
			fn(k, v)
			<-gaurd
		}(key, value)
	}
	wg.Wait()
}

func Map[A comparable, X comparable, B any, Z any](collection map[A]B, fn func(key A, value B) (X, Z)) map[X]Z {
	result := make(map[X]Z)
	wg := sync.WaitGroup{}
	resultChan := make(chan mapResult[X, Z])
	for key, val := range collection {
		wg.Add(1)
		go func(k A, v B) {
			defer wg.Done()
			rk, rv := fn(k, v)
			resultChan <- mapResult[X, Z]{
				Key:   rk,
				Value: rv,
			}
		}(key, val)
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

func MapLimit[A comparable, B any, X comparable, Z any](collection map[A]B, fn func(key A, value B) (X, Z), limit int) map[X]Z {
	result := make(map[X]Z)
	wg := sync.WaitGroup{}
	resultChan := make(chan mapResult[X, Z], len(collection))
	gaurd := make(chan struct{}, limit)
	defer close(gaurd)
	for key, val := range collection {
		wg.Add(1)
		gaurd <- struct{}{}
		go func(k A, v B) {
			defer wg.Done()
			rk, rv := fn(k, v)
			// Gaurd needs to be received before sending result to prevent deadlock.
			// As results channel is not buffered and gaurd will block for loop
			// till existing go routines are able to send on result channel
			<-gaurd
			resultChan <- mapResult[X, Z]{
				Key:   rk,
				Value: rv,
			}
		}(key, val)
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
