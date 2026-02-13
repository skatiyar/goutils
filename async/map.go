package async

import (
	"fmt"
	"sync"
)

// ConcatMap applies iteratee to each item in collection, concatenating the results and returns the concatenated list.
// The results array will be unorder as map iterations are unordered.
// If iterator returns an error, function returns immediately with an error. But some iteratee functions may still be running.
func ConcatMap[A comparable, B any, X any](collection map[A]B, fn func(key A, value B) ([]X, error)) ([]X, error) {
	return ConcatMapLimit(collection, fn, len(collection))
}

// ConcatMap applies iteratee to each item in collection, concatenating the results and returns the concatenated list.
// The results array will be unorder as map iterations are unordered.
// If iterator returns an error, function returns immediately with an error. But some iteratee functions may still be running.
func ConcatMapLimit[A comparable, B any, X any](collection map[A]B, fn func(key A, value B) ([]X, error), limit int) ([]X, error) {
	resultChan := executeMapLimit(
		collection,
		func(k A, v B) (opresult[A, []X], error) {
			rv, re := fn(k, v)
			return opresult[A, []X]{Key: k, Value: rv, Error: re}, re
		},
		limit,
		func(_ opresult[A, []X], err error) bool { return err != nil },
		func(err error) opresult[A, []X] { return opresult[A, []X]{Error: err} },
	)

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
	resultChan := executeMapLimit(
		collection,
		func(k A, v B) (opresult[B, bool], error) {
			ro, re := fn(k, v)
			return opresult[B, bool]{Key: v, Value: ro, Error: re}, re
		},
		limit,
		func(res opresult[B, bool], err error) bool { return err != nil || res.Value },
		func(err error) opresult[B, bool] { return opresult[B, bool]{Error: err} },
	)

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
	resultChan := executeMapLimit(
		collection,
		func(k A, v B) (error, error) {
			err := fn(k, v)
			return err, err
		},
		limit,
		func(_ error, err error) bool { return err != nil },
		func(err error) error { return err },
	)

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
	resultChan := executeMapLimit(
		collection,
		func(k A, v B) (opresult[X, Z], error) {
			rk, rv, re := fn(k, v)
			return opresult[X, Z]{Key: rk, Value: rv, Error: re}, re
		},
		limit,
		func(_ opresult[X, Z], err error) bool { return err != nil },
		func(err error) opresult[X, Z] { return opresult[X, Z]{Error: err} },
	)

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
	resultChan := executeMapLimit(
		collection,
		func(k A, v B) (opresult[A, bool], error) {
			rk, re := fn(k, v)
			return opresult[A, bool]{Key: k, Value: rk, Error: re}, re
		},
		limit,
		func(res opresult[A, bool], err error) bool { return err != nil || res.Value },
		func(err error) opresult[A, bool] { return opresult[A, bool]{Error: err} },
	)

	for resVal := range resultChan {
		if resVal.Error != nil || resVal.Value {
			return resVal.Value, resVal.Error
		}
	}
	return false, nil
}

// executeMapLimit executes fn on each map entry with concurrency limiting and panic recovery.
// shouldStop is called after each fn execution to determine if processing should halt early.
// makeErrorResult converts an error into a result of type R (needed for panic recovery).
// Returns a channel that emits results (including errors from panics or fn).
func executeMapLimit[K comparable, V any, R any](
	collection map[K]V,
	fn func(K, V) (R, error),
	limit int,
	shouldStop func(R, error) bool,
	makeErrorResult func(error) R,
) chan R {
	wg := sync.WaitGroup{}
	resultChan := make(chan R)
	guard := make(chan struct{}, limit)
	var stopOnce sync.Once
	stop := make(chan struct{})

	closeStop := func() {
		stopOnce.Do(func() {
			close(stop)
		})
	}

	wg.Add(1)
	go func(icol map[K]V) {
		defer wg.Done()
		for key, val := range icol {
			select {
			case <-stop:
				return
			default:
				guard <- struct{}{}
				wg.Add(1)
				go func(k K, v V) {
					defer func() {
						if r := recover(); r != nil {
							closeStop()
							var err error
							if e, ok := r.(error); ok {
								err = e
							} else {
								err = fmt.Errorf("panic in function: %v", r)
							}
							resultChan <- makeErrorResult(err)
						}
						wg.Done()
						<-guard
					}()
					result, err := fn(k, v)
					if shouldStop != nil && shouldStop(result, err) {
						closeStop()
					}
					resultChan <- result
				}(key, val)
			}
		}
	}(collection)
	go func() {
		wg.Wait()
		close(resultChan)
		close(guard)
	}()
	return resultChan
}

// FilterMap returns a new map of all entries in map which pass truth test in parallel.
// Tests are executed in parallel with max concurrency equal to map size.
// If any test returns an error, function returns immediately with error.
func FilterMap[K comparable, V any](collection map[K]V, fn func(key K, value V) (bool, error)) (map[K]V, error) {
	return FilterMapLimit(collection, fn, len(collection))
}

// FilterMapLimit is similar to FilterMap but limits concurrent executions.
// Tests are executed in parallel with max concurrency restricted to limit.
// If any test returns an error, function returns immediately with error.
func FilterMapLimit[K comparable, V any](collection map[K]V, fn func(key K, value V) (bool, error), limit int) (map[K]V, error) {
	resultChan := executeMapLimit(
		collection,
		func(k K, v V) (opresult[K, bool], error) {
			passes, err := fn(k, v)
			return opresult[K, bool]{Key: k, Value: passes, Error: err}, err
		},
		limit,
		func(_ opresult[K, bool], err error) bool { return err != nil },
		func(err error) opresult[K, bool] { return opresult[K, bool]{Error: err} },
	)

	result := make(map[K]V)
	for resVal := range resultChan {
		if resVal.Error != nil {
			return nil, resVal.Error
		}
		if resVal.Value {
			result[resVal.Key] = collection[resVal.Key]
		}
	}
	return result, nil
}

// RejectMap is the opposite of FilterMap. Removes entries that pass truth test in parallel.
// Tests are executed in parallel with max concurrency equal to map size.
// If any test returns an error, function returns immediately with error.
func RejectMap[K comparable, V any](collection map[K]V, fn func(key K, value V) (bool, error)) (map[K]V, error) {
	return RejectMapLimit(collection, fn, len(collection))
}

// RejectMapLimit is similar to RejectMap but limits concurrent executions.
// Tests are executed in parallel with max concurrency restricted to limit.
// If any test returns an error, function returns immediately with error.
func RejectMapLimit[K comparable, V any](collection map[K]V, fn func(key K, value V) (bool, error), limit int) (map[K]V, error) {
	resultChan := executeMapLimit(
		collection,
		func(k K, v V) (opresult[K, bool], error) {
			passes, err := fn(k, v)
			return opresult[K, bool]{Key: k, Value: passes, Error: err}, err
		},
		limit,
		func(_ opresult[K, bool], err error) bool { return err != nil },
		func(err error) opresult[K, bool] { return opresult[K, bool]{Error: err} },
	)

	result := make(map[K]V)
	for resVal := range resultChan {
		if resVal.Error != nil {
			return nil, resVal.Error
		}
		if !resVal.Value {
			result[resVal.Key] = collection[resVal.Key]
		}
	}
	return result, nil
}

// GroupByMap returns a new map, where each value corresponds to a slice of items, from map, that returned the corresponding key in parallel.
// That is, the keys of the result map correspond to the values returned by the iteratee callback.
// If any iterator returns an error, function returns immediately with an error.
func GroupByMap[K comparable, V any, GK comparable, GV any](collection map[K]V, fn func(key K, value V) (GK, GV, error)) (map[GK][]GV, error) {
	return GroupByMapLimit(collection, fn, len(collection))
}

// GroupByMapLimit is similar to GroupByMap but limits concurrent executions.
// Iteratees are executed in parallel with max concurrency restricted to limit.
// If any iterator returns an error, function returns immediately with an error.
func GroupByMapLimit[K comparable, V any, GK comparable, GV any](collection map[K]V, fn func(key K, value V) (GK, GV, error), limit int) (map[GK][]GV, error) {
	resultChan := executeMapLimit(
		collection,
		func(k K, v V) (opresult[GK, GV], error) {
			gk, gv, err := fn(k, v)
			return opresult[GK, GV]{Key: gk, Value: gv, Error: err}, err
		},
		limit,
		func(_ opresult[GK, GV], err error) bool { return err != nil },
		func(err error) opresult[GK, GV] { return opresult[GK, GV]{Error: err} },
	)

	result := make(map[GK][]GV)
	for resVal := range resultChan {
		if resVal.Error != nil {
			return nil, resVal.Error
		}
		result[resVal.Key] = append(result[resVal.Key], resVal.Value)
	}
	return result, nil
}
