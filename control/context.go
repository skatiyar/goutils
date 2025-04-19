package control

import (
	"context"
	"errors"
)

type ContextKey any

var (
	ErrValueTypeNotFound = errors.New("context value of type not found")
)

// GetControlContextValue retrieves a value of a specified type from the given context
// using the provided key. It returns the value if it exists and matches the expected type,
// or an error if the value is not found or does not match the expected type.
func GetControlContextValue[K, V any](ctx context.Context, key K) (value V, err error) {
	if val, ok := ctx.Value(ContextKey(key)).(V); ok {
		value = val
		return
	} else {
		err = ErrValueTypeNotFound
		return
	}
}

// SetControlContextValue sets a value in the provided context using a specified key.
// This function is generic and can accept any types for the key and value.
//
// Note:
//
//	The key is converted to a ContextKey type before being used to store the value.
//	Ensure that the key type is unique to avoid collisions in the context.
func SetControlContextValue[K, V any](ctx context.Context, key K, value V) context.Context {
	return context.WithValue(ctx, ContextKey(key), value)
}
