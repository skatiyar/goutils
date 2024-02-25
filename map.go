package goutils

// ConcatMap applies iteratee to each item in collection, concatenating the results and returns the concatenated list.
// The results array will be unorder as map iterations are unordered.
// If iterator returns an error, function returns immediately with an error and result as nil.
func ConcatMap[A comparable, B any, X any](collection map[A]B, fn func(key A, value B) ([]X, error)) ([]X, error) {
	result := make([]X, 0)
	for key, val := range collection {
		ival, ierr := fn(key, val)
		if ierr != nil {
			return nil, ierr
		}
		result = append(result, ival...)
	}
	return result, nil
}

// DetectMap returns the first value in collection that passes truth test, with a boolean signifying if the value was detected.
// If iterator returns an error, function returns immediately with an error and detected as false.
func DetectMap[A comparable, B any](collection map[A]B, fn func(key A, value B) (bool, error)) (result B, detected bool, err error) {
	for key, val := range collection {
		id, ierr := fn(key, val)
		if ierr != nil {
			detected, err = false, ierr
			return
		}
		if id {
			result, detected = val, true
			return
		}
	}
	return
}

// EachMap applies the function iteratee to each item in collection.
// The iteratee is called with an item from the collection.
// If the iterator returns an error, function returns immediately with an error.
func EachMap[A comparable, B any](collection map[A]B, fn func(key A, value B) error) error {
	for key, val := range collection {
		if err := fn(key, val); err != nil {
			return err
		}
	}
	return nil
}

// Map produces a new collection by mapping each key and value in collection through the iteratee function.
// The iteratee is called with key and value from collection, returns new key and value.
// If the iterator returns an error, function returns immediately with an error.
func Map[A comparable, X comparable, B any, Z any](collection map[A]B, fn func(key A, value B) (X, Z, error)) (map[X]Z, error) {
	result := make(map[X]Z)
	for key, val := range collection {
		if rk, rv, re := fn(key, val); re != nil {
			return nil, re
		} else {
			result[rk] = rv
		}
	}
	return result, nil
}

// ReduceMap reduces collection into a single value using an iteratee to return each successive step.
// If the iterator returns an error, function returns immediately with an error.
func ReduceMap[A comparable, B any, X any](collection map[A]B, fn func(accumulator X, key A, value B) (X, error), initial X) (X, error) {
	for key, value := range collection {
		if acc, accErr := fn(initial, key, value); accErr != nil {
			return initial, accErr
		} else {
			initial = acc
		}
	}
	return initial, nil
}

// EveryMap returns true if every element in collection satisfies a test.
// If any iteratee call returns false or an error, function returns immediately.
func EveryMap[A comparable, B any](collection map[A]B, fn func(key A, value B) (bool, error)) (bool, error) {
	for key, value := range collection {
		if test, testErr := fn(key, value); testErr != nil {
			return false, testErr
		} else if !test {
			return false, nil
		}
	}
	return true, nil
}

// FilterMap return a new map of all the values in collection which pass truth test.
// If the iterator returns an error, function returns immediately with an error.
func FilterMap[A comparable, B any](collection map[A]B, fn func(key A, value B) (bool, error)) (map[A]B, error) {
	result := make(map[A]B)
	for key, value := range collection {
		if test, testErr := fn(key, value); testErr != nil {
			return nil, testErr
		} else if test {
			result[key] = value
		}
	}
	return result, nil
}

// GroupByMap returns a new map, where each value corresponds to an array of items, from collection, that returned the corresponding key.
// That is, the keys of the object correspond to the values passed to the iteratee callback.
// If the iterator returns an error, function returns immediately with an error.
func GroupByMap[A comparable, B any, X comparable, Y any](collection map[A]B, fn func(key A, value B) (X, Y, error)) (map[X][]Y, error) {
	result := make(map[X][]Y)
	for key, value := range collection {
		if group, groupValue, groupErr := fn(key, value); groupErr != nil {
			return nil, groupErr
		} else if val, ok := result[group]; ok {
			result[group] = append(val, groupValue)
		} else {
			result[group] = []Y{groupValue}
		}
	}
	return result, nil
}

// RejectMap is the opposite of FilterMap. Removes values that pass truth test.
// If the iterator returns an error, function returns immediately with an error.
func RejectMap[A comparable, B any](collection map[A]B, fn func(key A, value B) (bool, error)) (map[A]B, error) {
	result := make(map[A]B)
	for key, value := range collection {
		if test, testErr := fn(key, value); testErr != nil {
			return nil, testErr
		} else if !test {
			result[key] = value
		}
	}
	return result, nil
}

// SomeMap returns true if at least one element in the collection satisfies test.
// If any iteratee call returns true, the function is returned immediately.
func SomeMap[A comparable, B any](collection map[A]B, fn func(key A, value B) (bool, error)) (bool, error) {
	for key, value := range collection {
		if test, testErr := fn(key, value); testErr != nil {
			return false, testErr
		} else if test {
			return true, nil
		}
	}
	return false, nil
}
