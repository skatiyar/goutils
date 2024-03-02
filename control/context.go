package control

import (
	"context"
	"errors"
)

type ContextKey any

var (
	ErrKeyNotFound       = errors.New("context key not found")
	ErrValueTypeNotFound = errors.New("context value of type not found")
)

func GetControlContextValue[K, V any](ctx context.Context, key K) (value V, err error) {
	if val, ok := ctx.Value(key).(V); ok {
		value = val
		return
	} else {
		err = ErrValueTypeNotFound
		return
	}
}

func SetControlContextValue[K, V any](ctx context.Context, key K, value V) context.Context {
	return context.WithValue(ctx, ContextKey(key), value)
}
