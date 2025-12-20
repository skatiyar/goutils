package primitives_test

import (
	"errors"
	"testing"

	"github.com/skatiyar/goutils/internal/primitives"
)

func TestResult_Success(t *testing.T) {
	expectedValue := 42
	result := primitives.NewResult[int]()
	go func() {
		result.Resolve(expectedValue, nil)
	}()

	value, err := result.Await()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if value != expectedValue {
		t.Fatalf("expected %d, got %d", expectedValue, value)
	}
}

func TestResult_Error(t *testing.T) {
	expectedError := errors.New("test error")
	result := primitives.NewResult[int]()
	go func() {
		result.Resolve(0, expectedError)
	}()

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

func TestResult_MultipleResolvePicksValueOfFirstResolve(t *testing.T) {
	result := primitives.NewResult[int]()
	go func() {
		result.Resolve(1, nil)
		result.Resolve(2, errors.New("second resolve"))
	}()

	value, err := result.Await()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if value != 1 {
		t.Fatalf("expected value 1, got %d", value)
	}
}

func TestResult_ResolveDoesntBlock(t *testing.T) {
	result := primitives.NewResult[float64]()
	result.Resolve(1, nil)
	value, err := result.Await()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if value != 1 {
		t.Fatalf("expected 'completed', got '%f'", value)
	}
}
