package goutils

// ConcatSlice applies iteratee to each item in slice, concatenating the results and returns the concatenated list.
// The results array will be ordered with respect to slice provided.
// If iterator returns an error, function returns immediately with an error and result as nil.
func ConcatSlice[A any, X any](collection []A, fn func(value A, idx int) ([]X, error)) ([]X, error) {
	result := make([]X, 0)
	for idx, value := range collection {
		ival, ierr := fn(value, idx)
		if ierr != nil {
			return nil, ierr
		}
		result = append(result, ival...)
	}
	return result, nil
}

// DetectSlice returns the first value in slice that passes truth test, with a boolean signifying if the value was detected.
// If iterator returns an error, function returns immediately with an error and detected as false.
func DetectSlice[A any](collection []A, fn func(value A, idx int) (bool, error)) (result A, detected bool, err error) {
	for idx, value := range collection {
		id, ierr := fn(value, idx)
		if ierr != nil {
			detected, err = false, ierr
			return
		}
		if id {
			result, detected = collection[idx], true
			return
		}
	}
	return
}

// EachSlice applies the function iteratee to each item in slice.
// The iteratee is called with an item from the slice.
// If the iterator returns an error, function returns immediately with an error.
func EachSlice[A any](collection []A, fn func(value A, idx int) error) error {
	for idx, value := range collection {
		if err := fn(value, idx); err != nil {
			return err
		}
	}
	return nil
}

// Slice produces a new slice by mapping each value in slice through the iteratee function.
// The iteratee is called with value from slice, returns new value.
// If the iterator returns an error, function returns immediately with an error.
func Slice[A any, X any](collection []A, fn func(value A, idx int) (X, error)) ([]X, error) {
	result := make([]X, 0)
	for idx, value := range collection {
		if rv, re := fn(value, idx); re != nil {
			return nil, re
		} else {
			result = append(result, rv)
		}
	}
	return result, nil
}

// ReduceSlice reduces slice into a single value using an iteratee to return each successive step.
// If the iterator returns an error, function returns immediately with an error.
func ReduceSlice[A any, X any](collection []A, fn func(accumulator X, value A, idx int) (X, error), initial X) (X, error) {
	for idx, value := range collection {
		if acc, accErr := fn(initial, value, idx); accErr != nil {
			return initial, accErr
		} else {
			initial = acc
		}
	}
	return initial, nil
}

// ReduceRightSlice reduces slice from right into a single value using an iteratee to return each successive step.
// If the iterator returns an error, function returns immediately with an error.
func ReduceRightSlice[A any, X any](collection []A, fn func(accumulator X, value A, idx int) (X, error), initial X) (X, error) {
	for idx := len(collection) - 1; idx >= 0; idx -= 1 {
		if acc, accErr := fn(initial, collection[idx], idx); accErr != nil {
			return initial, accErr
		} else {
			initial = acc
		}
	}
	return initial, nil
}

// EverySlice returns true if every element in slice satisfies a test.
// If any iteratee call returns false or an error, function returns immediately.
func EverySlice[A any](collection []A, fn func(value A, idx int) (bool, error)) (bool, error) {
	for idx, value := range collection {
		if test, testErr := fn(value, idx); testErr != nil {
			return false, testErr
		} else if !test {
			return false, nil
		}
	}
	return true, nil
}

// FilterSlice returns a new slice of all the values in slice which pass truth test.
// If the iterator returns an error, function returns immediately with an error.
func FilterSlice[A any](collection []A, fn func(value A, idx int) (bool, error)) ([]A, error) {
	result := make([]A, 0)
	for idx, value := range collection {
		if test, testErr := fn(value, idx); testErr != nil {
			return nil, testErr
		} else if test {
			result = append(result, value)
		}
	}
	return result, nil
}

// GroupBySlice returns a new map, where each value corresponds to an slice of items, from slice, that returned the corresponding key.
// That is, the keys of the object correspond to the values passed to the iteratee callback.
// If the iterator returns an error, function returns immediately with an error.
func GroupBySlice[A any, X comparable, Y any](collection []A, fn func(value A, idx int) (X, Y, error)) (map[X][]Y, error) {
	result := make(map[X][]Y)
	for idx, value := range collection {
		if group, groupValue, groupErr := fn(value, idx); groupErr != nil {
			return nil, groupErr
		} else if val, ok := result[group]; ok {
			result[group] = append(val, groupValue)
		} else {
			result[group] = []Y{groupValue}
		}
	}
	return result, nil
}

// RejectSlice is the opposite of FilterSlice. Removes values that pass truth test.
// If the iterator returns an error, function returns immediately with an error.
func RejectSlice[A any](collection []A, fn func(value A, idx int) (bool, error)) ([]A, error) {
	result := make([]A, 0)
	for idx, value := range collection {
		if test, testErr := fn(value, idx); testErr != nil {
			return nil, testErr
		} else if !test {
			result = append(result, value)
		}
	}
	return result, nil
}

// SomeSlice returns true if at least one element in the slice satisfies test.
// If any iteratee call returns true or error, the function is returned immediately.
func SomeSlice[A any](collection []A, fn func(value A, idx int) (bool, error)) (bool, error) {
	for idx, value := range collection {
		if test, testErr := fn(value, idx); testErr != nil {
			return false, testErr
		} else if test {
			return true, nil
		}
	}
	return false, nil
}
