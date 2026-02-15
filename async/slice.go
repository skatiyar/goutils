package async

import (
	"fmt"
	"sync"
)

type opresult[A, B any] struct {
	Key   A
	Value B
	Error error
}

// EachSlice applies the function iteratee to each item in slice in parallel.
// The iteratee is called with value and index from the slice.
// If any iteratee returns an error, function returns immediately with an error. But some iteratee functions may still be running.
func EachSlice[T any](collection []T, fn func(value T, idx int) error) error {
	return EachSliceLimit(collection, fn, len(collection))
}

// EachSliceLimit is similar to EachSlice but limits concurrent executions.
// Iteratees are executed in parallel with max concurrency restricted to limit.
// If any iteratee returns an error, function returns immediately with an error. But some iteratee functions may still be running.
func EachSliceLimit[T any](collection []T, fn func(value T, idx int) error, limit int) error {
	resultChan := executeSliceLimit(
		collection,
		func(idx int, val T) (error, error) {
			err := fn(val, idx)
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

// MapSlice produces a new slice by mapping each value in slice through the iteratee function in parallel.
// The iteratee is called with value and index from slice, returns new value.
// If any iterator returns an error, function returns immediately with an error. But some iteratee functions may still be running.
// Order is preserved in the result slice.
func MapSlice[T any, S any](collection []T, fn func(value T, idx int) (S, error)) ([]S, error) {
	return MapSliceLimit(collection, fn, len(collection))
}

// MapSliceLimit is similar to MapSlice but limits concurrent executions.
// Transformations are executed in parallel with max concurrency restricted to limit.
// If any iterator returns an error, function returns immediately with an error. But some iteratee functions may still be running.
// Order is preserved in the result slice.
func MapSliceLimit[T any, S any](collection []T, fn func(value T, idx int) (S, error), limit int) ([]S, error) {
	result := make([]S, len(collection))
	resultChan := executeSliceLimit(
		collection,
		func(idx int, val T) (opresult[int, S], error) {
			mapped, err := fn(val, idx)
			return opresult[int, S]{Key: idx, Value: mapped, Error: err}, err
		},
		limit,
		func(_ opresult[int, S], err error) bool { return err != nil },
		func(err error) opresult[int, S] { return opresult[int, S]{Error: err} },
	)

	for resVal := range resultChan {
		if resVal.Error != nil {
			return nil, resVal.Error
		}
		result[resVal.Key] = resVal.Value
	}
	return result, nil
}

// executeSliceLimit executes fn on each slice element with concurrency limiting and panic recovery.
// shouldStop is called after each fn execution to determine if processing should halt early.
// makeErrorResult converts an error into a result of type R (needed for panic recovery).
// Returns a channel that emits results (including errors from panics or fn).
func executeSliceLimit[T any, R any](
	collection []T,
	fn func(int, T) (R, error),
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
	go func(icol []T) {
		defer wg.Done()
		for idx := range icol {
			select {
			case <-stop:
				return
			default:
				guard <- struct{}{}
				wg.Add(1)
				go func(i int, val T) {
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
					result, err := fn(i, val)
					if shouldStop != nil && shouldStop(result, err) {
						closeStop()
					}
					resultChan <- result
				}(idx, icol[idx])
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

// FilterSlice returns a new slice of all the values in slice which pass truth test in parallel.
// Tests are executed in parallel with max concurrency equal to slice length.
// If any test returns an error, function returns immediately with error.
func FilterSlice[T any](collection []T, fn func(value T, idx int) (bool, error)) ([]T, error) {
	return FilterSliceLimit(collection, fn, len(collection))
}

// FilterSliceLimit is similar to FilterSlice but limits concurrent executions.
// Tests are executed in parallel with max concurrency restricted to limit.
// If any test returns an error, function returns immediately with error.
func FilterSliceLimit[T any](collection []T, fn func(value T, idx int) (bool, error), limit int) ([]T, error) {
	resultChan := executeSliceLimit(
		collection,
		func(idx int, val T) (opresult[int, bool], error) {
			passes, err := fn(val, idx)
			return opresult[int, bool]{Key: idx, Value: passes, Error: err}, err
		},
		limit,
		func(_ opresult[int, bool], err error) bool { return err != nil },
		func(err error) opresult[int, bool] { return opresult[int, bool]{Error: err} },
	)

	result := make([]T, 0)
	for resVal := range resultChan {
		if resVal.Error != nil {
			return nil, resVal.Error
		}
		if resVal.Value {
			result = append(result, collection[resVal.Key])
		}
	}
	return result, nil
}

// DetectSlice returns the first value in slice that passes truth test in parallel, with a boolean signifying if the value was detected.
// Tests are executed in parallel with max concurrency equal to slice length.
// If any test returns an error, function returns immediately with error and detected as false.
func DetectSlice[T any](collection []T, fn func(value T, idx int) (bool, error)) (T, bool, error) {
	return DetectSliceLimit(collection, fn, len(collection))
}

// DetectSliceLimit is similar to DetectSlice but limits concurrent executions.
// Tests are executed in parallel with max concurrency restricted to limit.
// If any test returns an error, function returns immediately with error and detected as false.
func DetectSliceLimit[T any](collection []T, fn func(value T, idx int) (bool, error), limit int) (T, bool, error) {
	resultChan := executeSliceLimit(
		collection,
		func(idx int, val T) (opresult[T, bool], error) {
			found, err := fn(val, idx)
			return opresult[T, bool]{Key: val, Value: found, Error: err}, err
		},
		limit,
		func(res opresult[T, bool], err error) bool { return err != nil || res.Value },
		func(err error) opresult[T, bool] { return opresult[T, bool]{Error: err} },
	)

	for resVal := range resultChan {
		if resVal.Error != nil || resVal.Value {
			return resVal.Key, resVal.Value, resVal.Error
		}
	}
	return *new(T), false, nil
}

// SomeSlice returns true if at least one element in the slice satisfies test in parallel.
// Tests are executed in parallel with max concurrency equal to slice length.
// If any test returns true or error, the function returns immediately.
func SomeSlice[T any](collection []T, fn func(value T, idx int) (bool, error)) (bool, error) {
	return SomeSliceLimit(collection, fn, len(collection))
}

// SomeSliceLimit is similar to SomeSlice but limits concurrent executions.
// Tests are executed in parallel with max concurrency restricted to limit.
// If any test returns true or error, the function returns immediately.
func SomeSliceLimit[T any](collection []T, fn func(value T, idx int) (bool, error), limit int) (bool, error) {
	resultChan := executeSliceLimit(
		collection,
		func(idx int, val T) (opresult[int, bool], error) {
			passes, err := fn(val, idx)
			return opresult[int, bool]{Key: idx, Value: passes, Error: err}, err
		},
		limit,
		func(res opresult[int, bool], err error) bool { return err != nil || res.Value },
		func(err error) opresult[int, bool] { return opresult[int, bool]{Error: err} },
	)

	for resVal := range resultChan {
		if resVal.Error != nil || resVal.Value {
			return resVal.Value, resVal.Error
		}
	}
	return false, nil
}

// EverySlice returns true if every element in the slice satisfies test in parallel.
// Tests are executed in parallel with max concurrency equal to slice length.
// If any test returns false or error, the function returns immediately.
func EverySlice[T any](collection []T, fn func(value T, idx int) (bool, error)) (bool, error) {
	return EverySliceLimit(collection, fn, len(collection))
}

// EverySliceLimit is similar to EverySlice but limits concurrent executions.
// Tests are executed in parallel with max concurrency restricted to limit.
// If any test returns false or error, the function returns immediately.
func EverySliceLimit[T any](collection []T, fn func(value T, idx int) (bool, error), limit int) (bool, error) {
	resultChan := executeSliceLimit(
		collection,
		func(idx int, val T) (opresult[int, bool], error) {
			passes, err := fn(val, idx)
			return opresult[int, bool]{Key: idx, Value: passes, Error: err}, err
		},
		limit,
		func(res opresult[int, bool], err error) bool { return err != nil || !res.Value },
		func(err error) opresult[int, bool] { return opresult[int, bool]{Error: err} },
	)

	for resVal := range resultChan {
		if resVal.Error != nil || !resVal.Value {
			return resVal.Value, resVal.Error
		}
	}
	return true, nil
}

// ConcatSlice applies iteratee to each item in slice, concatenating the results and returns the concatenated list in parallel.
// Results are concatenated as they complete. The order may not match the input slice order as execution happens asynchronously.
// If any iterator returns an error, function returns immediately with an error.
func ConcatSlice[T any, R any](collection []T, fn func(value T, idx int) ([]R, error)) ([]R, error) {
	return ConcatSliceLimit(collection, fn, len(collection))
}

// ConcatSliceLimit is similar to ConcatSlice but limits concurrent executions.
// Iteratees are executed in parallel with max concurrency restricted to limit.
// If any iterator returns an error, function returns immediately with an error.
func ConcatSliceLimit[T any, R any](collection []T, fn func(value T, idx int) ([]R, error), limit int) ([]R, error) {
	resultChan := executeSliceLimit(
		collection,
		func(idx int, val T) (opresult[int, []R], error) {
			arr, err := fn(val, idx)
			return opresult[int, []R]{Key: idx, Value: arr, Error: err}, err
		},
		limit,
		func(_ opresult[int, []R], err error) bool { return err != nil },
		func(err error) opresult[int, []R] { return opresult[int, []R]{Error: err} },
	)

	result := make([]R, 0)
	for resVal := range resultChan {
		if resVal.Error != nil {
			return nil, resVal.Error
		}
		result = append(result, resVal.Value...)
	}
	return result, nil
}

// RejectSlice is the opposite of FilterSlice. Removes values that pass truth test in parallel.
// Tests are executed in parallel with max concurrency equal to slice length.
// If any test returns an error, function returns immediately with error.
func RejectSlice[T any](collection []T, fn func(value T, idx int) (bool, error)) ([]T, error) {
	return RejectSliceLimit(collection, fn, len(collection))
}

// RejectSliceLimit is similar to RejectSlice but limits concurrent executions.
// Tests are executed in parallel with max concurrency restricted to limit.
// If any test returns an error, function returns immediately with error.
func RejectSliceLimit[T any](collection []T, fn func(value T, idx int) (bool, error), limit int) ([]T, error) {
	resultChan := executeSliceLimit(
		collection,
		func(idx int, val T) (opresult[int, bool], error) {
			passes, err := fn(val, idx)
			return opresult[int, bool]{Key: idx, Value: passes, Error: err}, err
		},
		limit,
		func(_ opresult[int, bool], err error) bool { return err != nil },
		func(err error) opresult[int, bool] { return opresult[int, bool]{Error: err} },
	)

	result := make([]T, 0)
	for resVal := range resultChan {
		if resVal.Error != nil {
			return nil, resVal.Error
		}
		if !resVal.Value {
			result = append(result, collection[resVal.Key])
		}
	}
	return result, nil
}

// GroupBySlice returns a new map, where each value corresponds to a slice of items, from slice, that returned the corresponding key in parallel.
// That is, the keys of the map correspond to the values passed to the iteratee callback.
// If any iterator returns an error, function returns immediately with an error.
func GroupBySlice[T any, K comparable, V any](collection []T, fn func(value T, idx int) (K, V, error)) (map[K][]V, error) {
	return GroupBySliceLimit(collection, fn, len(collection))
}

// GroupBySliceLimit is similar to GroupBySlice but limits concurrent executions.
// Iteratees are executed in parallel with max concurrency restricted to limit.
// If any iterator returns an error, function returns immediately with an error.
func GroupBySliceLimit[T any, K comparable, V any](collection []T, fn func(value T, idx int) (K, V, error), limit int) (map[K][]V, error) {
	resultChan := executeSliceLimit(
		collection,
		func(idx int, val T) (opresult[K, V], error) {
			key, value, err := fn(val, idx)
			return opresult[K, V]{Key: key, Value: value, Error: err}, err
		},
		limit,
		func(_ opresult[K, V], err error) bool { return err != nil },
		func(err error) opresult[K, V] { return opresult[K, V]{Error: err} },
	)

	result := make(map[K][]V)
	for resVal := range resultChan {
		if resVal.Error != nil {
			return nil, resVal.Error
		}
		result[resVal.Key] = append(result[resVal.Key], resVal.Value)
	}
	return result, nil
}
