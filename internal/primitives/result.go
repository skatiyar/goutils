package primitives

type resultValue[T any] struct {
	Value T
	Error error
}

// Result is a generic type that encapsulates the result of an asynchronous operation.
// It uses a channel to communicate the result value of type T, allowing for safe
// and concurrent access to the result.
type Result[T any] struct {
	result chan resultValue[T]
}

// Await waits for the asynchronous operation to complete and retrieves the result.
// It blocks until the result is available, then returns the value and any error
// that occurred during the operation. The result channel is closed after the
// value is retrieved.
func (r Result[T]) Await() (T, error) {
	data := <-r.result
	defer close(r.result)
	return data.Value, data.Error
}

// Resolve sets the result of the asynchronous operation by sending the provided
// value and error into the result channel. This method is typically called
// internally by the asynchronous operation to signal completion.
func (r Result[T]) Resolve(value T, err error) {
	r.result <- resultValue[T]{Value: value, Error: err}
}

// NewResult creates and returns a new instance of Result[T] with an initialized
// result channel. This function is used to create a Result for a specific type T.
func NewResult[T any]() Result[T] {
	return Result[T]{result: make(chan resultValue[T], 1)}
}
