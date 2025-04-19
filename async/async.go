package async

import (
	"context"
	"errors"
)

var ErrorPanicInGoroutine = errors.New("panic in go routine")

type resultValue[T any] struct {
	Value T
	Error error
}

// Result is a generic type that encapsulates the result of an asynchronous operation.
// It uses a channel to communicate the result value of type T, allowing for safe
// and concurrent access to the result of a computation or process.
type Result[T any] struct {
	result chan resultValue[T]
}

// Await waits for the asynchronous operation to complete and retrieves the result.
// It blocks until the result is available, then returns the value and any error
// that occurred during the operation. The result channel is closed after the
// value is retrieved.
func (f *Result[T]) Await() (T, error) {
	data := <-f.result
	defer close(f.result)
	return data.Value, data.Error
}

// Async executes a given function `f` asynchronously in a separate goroutine and
// returns a `Result[T]` that can be used to retrieve the result of the function
// execution. The function `f` is expected to return a value of type `T` and an error.
//
// If the function `f` executes successfully, the result will contain the value of type `T`.
// If the function `f` returns an error, the result will contain the error.
// If a panic occurs within the goroutine, the result will contain a predefined error
// or the recovered panic value if it is of type `error`.
//
// Type Parameters:
//   - T: The type of the value returned by the function `f`.
//
// Parameters:
//   - f: A function that returns a value of type `T` and an error.
//
// Returns:
//   - Result[T]: A wrapper that allows retrieving the result of the asynchronous
//     execution.
//
// Example usage:
//
//	func fetchData() (string, error) {
//	    // Simulate some work
//	    time.Sleep(1 * time.Second)
//	    return "data", nil
//	}
//
//	func main() {
//	    asyncResult := Async(fetchData)
//
//	    // Do other work while fetchData executes
//
//	    result := asyncResult.Await() // Blocks until the result is available
//	    if result.Error != nil {
//	        fmt.Println("Error:", result.Error)
//	    } else {
//	        fmt.Println("Value:", result.Value)
//	    }
//	}
func Async[T any](f func() (T, error)) Result[T] {
	result := make(chan resultValue[T], 1)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				if rec, ok := r.(error); ok {
					result <- resultValue[T]{Error: rec}
				} else {
					result <- resultValue[T]{Error: ErrorPanicInGoroutine}
				}
			}
		}()
		val, err := f()
		if err != nil {
			result <- resultValue[T]{Error: err}
		} else {
			result <- resultValue[T]{Value: val}
		}
	}()

	return Result[T]{result: result}
}

// AsyncWithContext executes a function asynchronously in a separate goroutine,
// while associating it with a given context. It returns a Result[T] that can
// be used to retrieve the result of the function execution.
//
// If the context is canceled or its deadline is exceeded before the function
// executes, the returned Result will contain the context's error.
//
// If a panic occurs within the goroutine, the result will contain a predefined error
// or the recovered panic value if it is of type `error`.
//
// Type Parameters:
//   - T: The type of the value returned by the function.
//
// Parameters:
//   - ctx: The context to associate with the asynchronous operation.
//   - f: A function that returns a value of type T and an error.
//
// Returns:
//
//	A Result[T] containing the result of the function execution or an error.
//
// Example usage:
//
//	func fetchData() (string, error) {
//	    // Simulate some work
//	    time.Sleep(1 * time.Second)
//	    return "data", nil
//	}
//
//	func main() {
//	    ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
//	    defer cancel()
//
//	    asyncResult := AsyncWithContext(ctx, fetchData)
//
//	    // Do other work while fetchData executes
//
//	    result, err := asyncResult.Await() // Blocks until the result is available
//	    if err != nil {
//	        fmt.Println("Error:", err)
//	    } else {
//	        fmt.Println("Value:", result)
//	    }
//	}
func AsyncWithContext[T any](ctx context.Context, f func() (T, error)) Result[T] {
	result := make(chan resultValue[T], 1)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				if rec, ok := r.(error); ok {
					result <- resultValue[T]{Error: rec}
				} else {
					result <- resultValue[T]{Error: ErrorPanicInGoroutine}
				}
			}
		}()
		select {
		case <-ctx.Done():
			result <- resultValue[T]{Error: ctx.Err()}
			return
		default:
			val, err := f()
			result <- resultValue[T]{Error: err, Value: val}
		}
	}()

	return Result[T]{result: result}
}
