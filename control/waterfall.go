package control

import (
	"context"
)

// Waterfall runs the executors in series, each passing their results to the next through context.
// However, if any of the tasks returns an error, the next task is not executed,
// and the function immediately returns with the error.
func Waterfall(executors ...func(context.Context) (context.Context, error)) (context.Context, error) {
	ctx := context.Background()
	for idx := range executors {
		execCtx, execErr := executors[idx](ctx)
		if execErr != nil {
			return ctx, execErr
		} else {
			ctx = execCtx
		}
	}
	return ctx, nil
}

// WaterfallBaseValue returns a function that when called, returns context with the values provided.
// Useful as the first function in a waterfall.
func WaterfallBaseValue(key, value interface{}) func(context.Context) (context.Context, error) {
	return func(ctx context.Context) (context.Context, error) {
		return context.WithValue(context.Background(), ContextKey(key), value), nil
	}
}
