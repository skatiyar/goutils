package async_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/skatiyar/goutils/async"
)

func TestAsync_Success(t *testing.T) {
	t.Parallel()

	expectedValue := 42
	result := async.Async(func() (int, error) {
		time.Sleep(100 * time.Millisecond)
		return expectedValue, nil
	})

	value, err := result.Await()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if value != expectedValue {
		t.Fatalf("expected %d, got %d", expectedValue, value)
	}
}

func TestAsync_Error(t *testing.T) {
	t.Parallel()

	expectedError := errors.New("test error")
	result := async.Async(func() (int, error) {
		time.Sleep(100 * time.Millisecond)
		return 0, expectedError
	})

	value, err := result.Await()
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
	if err.Error() != expectedError.Error() {
		t.Fatalf("expected error %v, got %v", expectedError, err)
	}
	if value != 0 {
		t.Fatalf("expected value 0, got %d", value)
	}
}

func TestAsync_Panic(t *testing.T) {
	t.Parallel()

	result := async.Async(func() (int, error) {
		panic("something went wrong")
	})

	value, err := result.Await()
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
	if err != async.ErrorPanicInGoroutine {
		t.Fatalf("expected panic error, got %v", err)
	}
	if value != 0 {
		t.Fatalf("expected value 0, got %d", value)
	}
}

func TestAsync_Concurrent(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		function      func() (int, error)
		expectedValue int
		expectedError error
	}{
		{
			name: "success",
			function: func() (int, error) {
				time.Sleep(50 * time.Millisecond)
				return 10, nil
			},
			expectedValue: 10,
			expectedError: nil,
		},
		{
			name: "error",
			function: func() (int, error) {
				time.Sleep(50 * time.Millisecond)
				return 0, errors.New("error occurred")
			},
			expectedValue: 0,
			expectedError: errors.New("error occurred"),
		},
		{
			name: "panic",
			function: func() (int, error) {
				panic("panic occurred")
			},
			expectedValue: 0,
			expectedError: async.ErrorPanicInGoroutine,
		},
	}

	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := async.Async(tt.function)
			value, err := result.Await()

			if err != nil && tt.expectedError == nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if err == nil && tt.expectedError != nil {
				t.Fatalf("expected error %v, got nil", tt.expectedError)
			}
			if err != nil && err.Error() != tt.expectedError.Error() {
				t.Fatalf("expected error %v, got %v", tt.expectedError, err)
			}
			if value != tt.expectedValue {
				t.Fatalf("expected value %d, got %d", tt.expectedValue, value)
			}
		})
	}
}
func TestAsyncWithContext_Success(t *testing.T) {
	t.Parallel()

	expectedValue := 42
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	time.Sleep(100 * time.Millisecond)
	result := async.AsyncWithContext(ctx, func() (int, error) {
		return expectedValue, nil
	})

	value, err := result.Await()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if value != expectedValue {
		t.Fatalf("expected %d, got %d", expectedValue, value)
	}
}

func TestAsyncWithContext_Timeout(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	time.Sleep(100 * time.Millisecond)
	result := async.AsyncWithContext(ctx, func() (int, error) {
		return 42, nil
	})

	value, err := result.Await()
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected context deadline exceeded error, got %v", err)
	}
	if value != 0 {
		t.Fatalf("expected value 0, got %d", value)
	}
}

func TestAsyncWithContext_Error(t *testing.T) {
	t.Parallel()

	expectedError := errors.New("test error")
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	time.Sleep(100 * time.Millisecond)
	result := async.AsyncWithContext(ctx, func() (int, error) {
		return 0, expectedError
	})

	value, err := result.Await()
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
	if err.Error() != expectedError.Error() {
		t.Fatalf("expected error %v, got %v", expectedError, err)
	}
	if value != 0 {
		t.Fatalf("expected value 0, got %d", value)
	}
}

func TestAsyncWithContext_Panic(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	result := async.AsyncWithContext(ctx, func() (int, error) {
		panic("something went wrong")
	})

	value, err := result.Await()
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
	if err != async.ErrorPanicInGoroutine {
		t.Fatalf("expected panic error, got %v", err)
	}
	if value != 0 {
		t.Fatalf("expected value 0, got %d", value)
	}
}

func TestAsyncWithContext_Cancel(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	result := async.AsyncWithContext(ctx, func() (int, error) {
		time.Sleep(100 * time.Millisecond)
		return 42, nil
	})

	value, err := result.Await()
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context canceled error, got %v", err)
	}
	if value != 0 {
		t.Fatalf("expected value 0, got %d", value)
	}
}
