package goutils

func Map[A comparable, X comparable, B any, Z any](collection map[A]B, fn func(key A, value B) (X, Z)) map[X]Z {
	result := make(map[X]Z)
	for key, val := range collection {
		rk, rv := fn(key, val)
		result[rk] = rv
	}
	return result
}
