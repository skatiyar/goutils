package async

import (
	"context"
	"errors"

	"github.com/skatiyar/goutils/internal/primitives"
)

var ErrorPanicInGoroutine = errors.New("panic in go routine")

// Async executes a given function `f` asynchronously in a separate goroutine and
// returns a `Result[T]` that can be used to retrieve the result of the function
// execution. The function `f` is expected to return a value of type `T` and an error.
//
// If the function `f` executes successfully, the result will contain the value of type `T`.
// If the function `f` returns an error, the result will contain the error.
// If a panic occurs within the goroutine, the result will contain a predefined error
// or the recovered panic value if it is of type `error`.
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
func Async[T any](f func() (T, error)) primitives.Result[T] {
	result := primitives.NewResult[T]()

	go func() {
		defer func() {
			if r := recover(); r != nil {
				if rec, ok := r.(error); ok {
					result.Resolve(*new(T), rec)
				} else {
					result.Resolve(*new(T), ErrorPanicInGoroutine)
				}
			}
		}()
		val, err := f()
		result.Resolve(val, err)
	}()

	return result
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
func AsyncWithContext[T any](ctx context.Context, f func() (T, error)) primitives.Result[T] {
	result := primitives.NewResult[T]()

	go func() {
		defer func() {
			if r := recover(); r != nil {
				if rec, ok := r.(error); ok {
					result.Resolve(*new(T), rec)
				} else {
					result.Resolve(*new(T), ErrorPanicInGoroutine)
				}
			}
		}()
		select {
		case <-ctx.Done():
			result.Resolve(*new(T), ctx.Err())
			return
		default:
			val, err := f()
			result.Resolve(val, err)
		}
	}()

	return result
}
