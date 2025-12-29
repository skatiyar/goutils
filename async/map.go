package async

import (
	"fmt"
	"sync"
)

func stopChannelCloser(ch chan struct{}) {
	select {
	case <-ch:
		// already closed
	default:
		close(ch)
	}
}

// ConcatMap applies iteratee to each item in collection, concatenating the results and returns the concatenated list.
// The results array will be unorder as map iterations are unordered.
// If iterator returns an error, function returns immediately with an error and result as nil.
func ConcatMap[A comparable, B any, X any](collection map[A]B, fn func(key A, value B) ([]X, error)) ([]X, error) {
	return ConcatMapLimit(collection, fn, len(collection))
}

func ConcatMapLimit[A comparable, B any, X any](collection map[A]B, fn func(key A, value B) ([]X, error), limit int) ([]X, error) {
	wg := sync.WaitGroup{}
	resultChan := make(chan opresult[A, []X])
	gaurd := make(chan struct{}, limit)
	wg.Add(1)
	go func(icol map[A]B) {
		defer wg.Done()
		stop := make(chan struct{})
		for key, val := range icol {
			select {
			case <-stop:
				return
			default:
				gaurd <- struct{}{}
				wg.Add(1)
				go func(k A, v B) {
					defer func() {
						if r := recover(); r != nil {
							stopChannelCloser(stop)
							if err, ok := r.(error); ok {
								resultChan <- opresult[A, []X]{Error: err}
							} else {
								resultChan <- opresult[A, []X]{Error: fmt.Errorf("panic in function: %v", r)}
							}
						}
						wg.Done()
						<-gaurd
					}()
					rv, re := fn(k, v)
					if re != nil {
						stopChannelCloser(stop)
					}
					resultChan <- opresult[A, []X]{
						Key:   k,
						Value: rv,
						Error: re,
					}
				}(key, val)
			}
		}
	}(collection)
	go func() {
		wg.Wait()
		close(resultChan)
		close(gaurd)
	}()
	result := make([]X, 0)
	for resVal := range resultChan {
		if resVal.Error != nil {
			return nil, resVal.Error
		}
		result = append(result, resVal.Value...)
	}
	return result, nil
}

// DetectMap returns the first value in collection that passes truth test, with a boolean signifying if the value was detected.
// If iterator returns an error, function returns immediately with an error and detected as false.
func DetectMap[A comparable, B any](collection map[A]B, fn func(key A, value B) (bool, error)) (B, bool, error) {
	return DetectMapLimit(collection, fn, len(collection))
}

// DetectMap returns the first value in collection that passes truth test, with a boolean signifying if the value was detected.
// If iterator returns an error, function returns immediately with an error and detected as false.
func DetectMapLimit[A comparable, B any](collection map[A]B, fn func(key A, value B) (bool, error), limit int) (B, bool, error) {
	wg := sync.WaitGroup{}
	resultChan := make(chan opresult[B, bool])
	gaurd := make(chan struct{}, limit)
	wg.Add(1)
	go func(icol map[A]B) {
		defer wg.Done()
		stop := make(chan struct{})
		for key, val := range icol {
			select {
			case <-stop:
				return
			default:
				gaurd <- struct{}{}
				wg.Add(1)
				go func(k A, v B) {
					defer func() {
						if r := recover(); r != nil {
							stopChannelCloser(stop)
							if err, ok := r.(error); ok {
								resultChan <- opresult[B, bool]{Error: err}
							} else {
								resultChan <- opresult[B, bool]{Error: fmt.Errorf("panic in function: %v", r)}
							}
						}
						wg.Done()
						<-gaurd
					}()
					ro, re := fn(k, v)
					if re != nil || ro {
						stopChannelCloser(stop)
					}
					resultChan <- opresult[B, bool]{
						Key:   v,
						Value: ro,
						Error: re,
					}
				}(key, val)
			}
		}
	}(collection)
	go func() {
		wg.Wait()
		close(resultChan)
		close(gaurd)
	}()
	for resVal := range resultChan {
		if resVal.Error != nil || resVal.Value {
			return resVal.Key, resVal.Value, resVal.Error
		}
	}
	return *new(B), false, nil
}

func EachMap[A comparable, B any](collection map[A]B, fn func(key A, value B) error) error {
	return EachMapLimit(collection, fn, len(collection))
}

func EachMapLimit[A comparable, B any](collection map[A]B, fn func(key A, value B) error, limit int) error {
	wg := sync.WaitGroup{}
	resultChan := make(chan error)
	gaurd := make(chan struct{}, limit)
	wg.Add(1)
	go func(icol map[A]B) {
		defer wg.Done()
		stop := make(chan struct{})
		for key, val := range icol {
			select {
			case <-stop:
				return
			default:
				gaurd <- struct{}{}
				wg.Add(1)
				go func(k A, v B) {
					defer func() {
						if r := recover(); r != nil {
							stopChannelCloser(stop)
							if err, ok := r.(error); ok {
								resultChan <- err
							} else {
								resultChan <- fmt.Errorf("panic in function: %v", r)
							}
						}
						wg.Done()
						<-gaurd
					}()
					re := fn(k, v)
					if re != nil {
						stopChannelCloser(stop)
					}
					resultChan <- re
				}(key, val)
			}
		}
	}(collection)
	go func() {
		wg.Wait()
		close(resultChan)
		close(gaurd)
	}()
	for resVal := range resultChan {
		if resVal != nil {
			return resVal
		}
	}
	return nil
}

// Map produces a new collection by mapping each key and value in collection through the iteratee function in parallel.
// The iteratee is called with key and value from collection, returns new key and value.
// If the iterator returns an error, function returns immediately with an error. But some iteratee functions may still be running.
func Map[A comparable, B any, X comparable, Z any](collection map[A]B, fn func(key A, value B) (X, Z, error)) (map[X]Z, error) {
	return MapLimit(collection, fn, len(collection))
}

func MapLimit[A comparable, B any, X comparable, Z any](collection map[A]B, fn func(key A, value B) (X, Z, error), limit int) (map[X]Z, error) {
	wg := sync.WaitGroup{}
	resultChan := make(chan opresult[X, Z])
	gaurd := make(chan struct{}, limit)
	wg.Add(1)
	go func(icol map[A]B) {
		defer wg.Done()
		stop := make(chan struct{})
		for key, val := range icol {
			select {
			case <-stop:
				return
			default:
				gaurd <- struct{}{}
				wg.Add(1)
				go func(k A, v B) {
					defer func() {
						if r := recover(); r != nil {
							stopChannelCloser(stop)
							if err, ok := r.(error); ok {
								resultChan <- opresult[X, Z]{Error: err}
							} else {
								resultChan <- opresult[X, Z]{Error: fmt.Errorf("panic in function: %v", r)}
							}
						}
						wg.Done()
						<-gaurd
					}()
					rk, rv, re := fn(k, v)
					if re != nil {
						stopChannelCloser(stop)
					}
					resultChan <- opresult[X, Z]{
						Key:   rk,
						Value: rv,
						Error: re,
					}
				}(key, val)
			}
		}
	}(collection)
	go func() {
		wg.Wait()
		close(resultChan)
		close(gaurd)
	}()
	result := make(map[X]Z)
	for resVal := range resultChan {
		if resVal.Error != nil {
			return nil, resVal.Error
		}
		result[resVal.Key] = resVal.Value
	}
	return result, nil
}

// SomeMap returns true if at least one element in the collection satisfies test.
// Test are applied in parallel with max concurrency equal to number of keys in collection.
// If any test call returns true or error, the function is returned immediately. But some test functions may still be running.
func SomeMap[A comparable, B any](collection map[A]B, fn func(key A, value B) (bool, error)) (bool, error) {
	return SomeMapLimit(collection, fn, len(collection))
}

// SomeMapLimit is similar to SomeMap, returns true if at least one element in the collection satisfies test.
// Test are applied in parallel with max concurrency restricted to limit provided.
// If any test call returns true or error, the function is returned immediately. But some test functions may still be running.
func SomeMapLimit[A comparable, B any](collection map[A]B, fn func(key A, value B) (bool, error), limit int) (bool, error) {
	wg := sync.WaitGroup{}
	resultChan := make(chan opresult[A, bool])
	gaurd := make(chan struct{}, limit)
	wg.Add(1)
	go func(icol map[A]B) {
		defer wg.Done()
		stop := make(chan struct{})
		for key, val := range icol {
			select {
			case <-stop:
				return
			default:
				gaurd <- struct{}{}
				wg.Add(1)
				go func(k A, v B) {
					defer func() {
						if r := recover(); r != nil {
							stopChannelCloser(stop)
							if err, ok := r.(error); ok {
								resultChan <- opresult[A, bool]{Error: err}
							} else {
								resultChan <- opresult[A, bool]{Error: fmt.Errorf("panic in function: %v", r)}
							}
						}
						wg.Done()
						<-gaurd
					}()
					rk, re := fn(k, v)
					if re != nil || rk {
						stopChannelCloser(stop)
					}
					resultChan <- opresult[A, bool]{
						Key:   k,
						Value: rk,
						Error: re,
					}
				}(key, val)
			}
		}
	}(collection)
	go func() {
		wg.Wait()
		close(resultChan)
		close(gaurd)
	}()
	for resVal := range resultChan {
		if resVal.Error != nil || resVal.Value {
			return resVal.Value, resVal.Error
		}
	}
	return false, nil
}
