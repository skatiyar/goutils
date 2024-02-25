package goutils

import "context"

// Waterfall runs the executors in series, each passing their results to the next through context.
// However, if any of the tasks returns an error, the next function is not executed,
// and the function immediately returns with the error.
func Waterfall(ctx context.Context, executors ...func(context.Context) (context.Context, error)) (context.Context, error) {
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
