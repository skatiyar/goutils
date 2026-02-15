package async_test

import (
	"errors"
	"math"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/skatiyar/goutils/async"
	"github.com/stretchr/testify/assert"
)

func TestEachSlice(t *testing.T) {
	t.Run("should return correct values for sync operations", func(nt *testing.T) {
		collection := []int{2, 7, 8, 9, 1, 3}
		collectionResult := []int{4, 49, 64, 81, 1, 9}
		rmu := sync.RWMutex{}
		results := make([]int, 0)
		err := async.EachSlice(collection, func(value, idx int) error {
			rmu.Lock()
			defer rmu.Unlock()
			results = append(results, int(math.Pow(float64(value), 2)))
			return nil
		})
		assert.NoError(nt, err)
		assert.ElementsMatch(nt, results, collectionResult)
	})
	t.Run("should return correct values for async operations", func(nt *testing.T) {
		collection := []int{2, 7, 8, 9, 1, 3}
		collectionResult := []int{4, 49, 64, 81, 1, 9}
		rmu := sync.RWMutex{}
		results := make([]int, 0)
		err := async.EachSlice(collection, func(value, idx int) error {
			time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
			rmu.Lock()
			defer rmu.Unlock()
			results = append(results, int(math.Pow(float64(value), 2)))
			return nil
		})
		assert.NoError(nt, err)
		assert.ElementsMatch(nt, results, collectionResult)
	})
}

func TestEachSliceLimit(t *testing.T) {
	t.Run("should return correct values for sync operations", func(nt *testing.T) {
		collection := []int{2, 7, 8, 9, 1, 3}
		collectionResult := []int{4, 49, 64, 81, 1, 9}
		rmu := sync.RWMutex{}
		results := make([]int, 0)
		maxLimit := 2
		currentLimit := 0
		limitExceeded := false
		err := async.EachSliceLimit(collection, func(value, idx int) error {
			rmu.Lock()
			currentLimit += 1
			defer func() {
				currentLimit -= 1
				rmu.Unlock()
			}()
			if currentLimit > maxLimit {
				limitExceeded = true
			}
			results = append(results, int(math.Pow(float64(value), 2)))
			return nil
		}, maxLimit)
		assert.NoError(nt, err)
		assert.ElementsMatch(nt, results, collectionResult)
		assert.False(nt, limitExceeded)
	})
	t.Run("should return correct values for async operations", func(nt *testing.T) {
		collection := []int{2, 7, 8, 9, 1, 3}
		collectionResult := []int{4, 49, 64, 81, 1, 9}
		rmu := sync.RWMutex{}
		results := make([]int, 0)
		maxLimit := 2
		currentLimit := 0
		limitExceeded := false
		err := async.EachSliceLimit(collection, func(value, idx int) error {
			time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
			rmu.Lock()
			currentLimit += 1
			defer func() {
				currentLimit -= 1
				rmu.Unlock()
			}()
			if currentLimit > maxLimit {
				limitExceeded = true
			}
			results = append(results, int(math.Pow(float64(value), 2)))
			return nil
		}, maxLimit)
		assert.NoError(nt, err)
		assert.ElementsMatch(nt, results, collectionResult)
		assert.False(nt, limitExceeded)
	})
}

func TestMapSlice(t *testing.T) {
	t.Run("should return correct values for square of integers", func(nt *testing.T) {
		collection := []int{2, 7, 8, 9, 1, 3}
		collectionResult := []int{4, 49, 64, 81, 1, 9}
		result, err := async.MapSlice(collection, func(val int, idx int) (int, error) {
			return int(math.Pow(float64(val), 2)), nil
		})
		assert.NoError(nt, err)
		assert.Equal(nt, result, collectionResult)
	})

	t.Run("should return correct values for async operations", func(nt *testing.T) {
		collection := []int{2, 7, 8, 9, 1, 3}
		collectionResult := []int{4, 49, 64, 81, 1, 9}
		result, err := async.MapSlice(collection, func(val int, idx int) (int, error) {
			time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
			return int(math.Pow(float64(val), 2)), nil
		})
		assert.NoError(nt, err)
		assert.Equal(nt, result, collectionResult)
	})
}

func TestMapSliceLimit(t *testing.T) {
	t.Run("should return correct values for square of integers", func(nt *testing.T) {
		collection := []int{2, 7, 8, 9, 1, 3}
		collectionResult := []int{4, 49, 64, 81, 1, 9}
		maxLimit := 2
		rmu := sync.RWMutex{}
		currentLimit := 0
		limitExceeded := false
		result, err := async.MapSliceLimit(collection, func(val int, idx int) (int, error) {
			rmu.Lock()
			currentLimit += 1
			defer func() {
				currentLimit -= 1
				rmu.Unlock()
			}()
			if currentLimit > maxLimit {
				limitExceeded = true
			}
			return int(math.Pow(float64(val), 2)), nil
		}, maxLimit)
		assert.NoError(nt, err)
		assert.Equal(nt, result, collectionResult)
		assert.False(nt, limitExceeded)
	})

	t.Run("should return correct values for async operations", func(nt *testing.T) {
		collection := []int{2, 7, 8, 9, 1, 3}
		collectionResult := []int{4, 49, 64, 81, 1, 9}
		maxLimit := 4
		rmu := sync.RWMutex{}
		currentLimit := 0
		limitExceeded := false
		result, err := async.MapSliceLimit(collection, func(val int, idx int) (int, error) {
			time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
			rmu.Lock()
			currentLimit += 1
			defer func() {
				currentLimit -= 1
				rmu.Unlock()
			}()
			if currentLimit > maxLimit {
				limitExceeded = true
			}
			return int(math.Pow(float64(val), 2)), nil
		}, maxLimit)
		assert.NoError(nt, err)
		assert.Equal(nt, result, collectionResult)
		assert.False(nt, limitExceeded)
	})
}

func TestFilterSlice(t *testing.T) {
	t.Run("should delegate to FilterSliceLimit with len(collection)", func(nt *testing.T) {
		collection := []int{1, 2, 3, 4, 5}
		expected := []int{2, 4}
		result, err := async.FilterSlice(collection, func(val, idx int) (bool, error) {
			return val%2 == 0, nil
		})
		assert.NoError(nt, err)
		assert.ElementsMatch(nt, result, expected)
	})
}

func TestFilterSliceLimit(t *testing.T) {
	t.Run("should filter correctly for sync operations", func(nt *testing.T) {
		collection := []int{1, 2, 3, 4, 5}
		expected := []int{2, 4}
		result, err := async.FilterSliceLimit(collection, func(val, idx int) (bool, error) {
			return val%2 == 0, nil
		}, 2)
		assert.NoError(nt, err)
		assert.ElementsMatch(nt, result, expected)
	})

	t.Run("should filter correctly for async operations", func(nt *testing.T) {
		collection := []int{1, 2, 3, 4, 5}
		expected := []int{2, 4}
		result, err := async.FilterSliceLimit(collection, func(val, idx int) (bool, error) {
			time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
			return val%2 == 0, nil
		}, 3)
		assert.NoError(nt, err)
		assert.ElementsMatch(nt, result, expected)
	})

	t.Run("should return all elements when all pass", func(nt *testing.T) {
		collection := []int{1, 2, 3, 4, 5}
		result, err := async.FilterSliceLimit(collection, func(val, idx int) (bool, error) {
			return true, nil
		}, 2)
		assert.NoError(nt, err)
		assert.ElementsMatch(nt, result, collection)
	})

	t.Run("should return empty slice when none pass", func(nt *testing.T) {
		collection := []int{1, 2, 3, 4, 5}
		result, err := async.FilterSliceLimit(collection, func(val, idx int) (bool, error) {
			return false, nil
		}, 2)
		assert.NoError(nt, err)
		assert.Empty(nt, result)
	})

	t.Run("should filter with unordered results", func(nt *testing.T) {
		collection := []int{5, 2, 4, 1, 3}
		expected := []int{5, 4, 3}
		result, err := async.FilterSliceLimit(collection, func(val, idx int) (bool, error) {
			return val > 2, nil
		}, 2)
		assert.NoError(nt, err)
		assert.ElementsMatch(nt, result, expected)
	})

	t.Run("should return error immediately", func(nt *testing.T) {
		collection := []int{1, 2, 3, 4, 5}
		result, err := async.FilterSliceLimit(collection, func(val, idx int) (bool, error) {
			return false, errors.New("some error")
		}, 2)
		assert.Error(nt, err)
		assert.Nil(nt, result)
	})

	t.Run("should return error immediately with early termination", func(nt *testing.T) {
		collection := []int{1, 2, 3, 4, 5, 6}
		startTime := time.Now()
		result, err := async.FilterSliceLimit(collection, func(val, idx int) (bool, error) {
			time.Sleep(200 * time.Millisecond)
			return false, errors.New("some error")
		}, 3)
		elapsedTime := time.Since(startTime)
		assert.Error(nt, err)
		assert.Nil(nt, result)
		assert.Less(nt, int(elapsedTime.Milliseconds()), 500)
	})

	t.Run("should handle panic", func(nt *testing.T) {
		collection := []int{1, 2, 3, 4, 5}
		result, err := async.FilterSliceLimit(collection, func(val, idx int) (bool, error) {
			panic("some panic")
		}, 2)
		assert.Error(nt, err)
		assert.Nil(nt, result)
	})

	t.Run("should not exceed concurrency limit", func(nt *testing.T) {
		collection := []int{1, 2, 3, 4, 5, 6, 7}
		maxLimit := 2
		rmu := sync.RWMutex{}
		currentLimit := 0
		limitExceeded := false

		_, _ = async.FilterSliceLimit(collection, func(val, idx int) (bool, error) {
			rmu.Lock()
			currentLimit += 1
			defer func() {
				currentLimit -= 1
				rmu.Unlock()
			}()
			if currentLimit > maxLimit {
				limitExceeded = true
			}
			time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
			return true, nil
		}, maxLimit)

		assert.False(nt, limitExceeded)
	})
}

func TestDetectSlice(t *testing.T) {
	t.Run("should delegate to DetectSliceLimit with len(collection)", func(nt *testing.T) {
		collection := []int{1, 2, 3, 4, 5}
		result, detected, err := async.DetectSlice(collection, func(val, idx int) (bool, error) {
			return val == 3, nil
		})
		assert.NoError(nt, err)
		assert.True(nt, detected)
		assert.Equal(nt, result, 3)
	})
}

func TestDetectSliceLimit(t *testing.T) {
	t.Run("should detect value for sync operations", func(nt *testing.T) {
		collection := []int{1, 2, 3, 4, 5}
		result, detected, err := async.DetectSliceLimit(collection, func(val, idx int) (bool, error) {
			return val == 3, nil
		}, 2)
		assert.NoError(nt, err)
		assert.True(nt, detected)
		assert.Equal(nt, result, 3)
	})

	t.Run("should detect value for async operations", func(nt *testing.T) {
		collection := []int{1, 2, 3, 4, 5}
		result, detected, err := async.DetectSliceLimit(collection, func(val, idx int) (bool, error) {
			time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
			return val == 3, nil
		}, 3)
		assert.NoError(nt, err)
		assert.True(nt, detected)
		assert.Equal(nt, result, 3)
	})

	t.Run("should return zero value and false when no match", func(nt *testing.T) {
		collection := []int{1, 2, 3, 4, 5}
		result, detected, err := async.DetectSliceLimit(collection, func(val, idx int) (bool, error) {
			return val == 99, nil
		}, 2)
		assert.NoError(nt, err)
		assert.False(nt, detected)
		assert.Equal(nt, result, 0)
	})

	t.Run("should return immediately on first match", func(nt *testing.T) {
		collection := []int{1, 2, 3, 4, 5, 6}
		startTime := time.Now()
		result, detected, err := async.DetectSliceLimit(collection, func(val, idx int) (bool, error) {
			time.Sleep(200 * time.Millisecond)
			return val == 3, nil
		}, 3)
		elapsedTime := time.Since(startTime)
		assert.NoError(nt, err)
		assert.True(nt, detected)
		assert.Equal(nt, result, 3)
		assert.Less(nt, int(elapsedTime.Milliseconds()), 700) // 3*200ms + margin
	})

	t.Run("should return error immediately", func(nt *testing.T) {
		collection := []int{1, 2, 3, 4, 5}
		_, detected, err := async.DetectSliceLimit(collection, func(val, idx int) (bool, error) {
			return false, errors.New("some error")
		}, 2)
		assert.Error(nt, err)
		assert.False(nt, detected)
	})

	t.Run("should return error immediately with early termination", func(nt *testing.T) {
		collection := []int{1, 2, 3, 4, 5, 6}
		startTime := time.Now()
		_, detected, err := async.DetectSliceLimit(collection, func(val, idx int) (bool, error) {
			time.Sleep(200 * time.Millisecond)
			return false, errors.New("some error")
		}, 3)
		elapsedTime := time.Since(startTime)
		assert.Error(nt, err)
		assert.False(nt, detected)
		assert.Less(nt, int(elapsedTime.Milliseconds()), 700)
	})

	t.Run("should handle panic", func(nt *testing.T) {
		collection := []int{1, 2, 3, 4, 5}
		_, detected, err := async.DetectSliceLimit(collection, func(val, idx int) (bool, error) {
			panic("some panic")
		}, 2)
		assert.Error(nt, err)
		assert.False(nt, detected)
	})

	t.Run("should not exceed concurrency limit", func(nt *testing.T) {
		collection := []int{1, 2, 3, 4, 5, 6, 7}
		maxLimit := 2
		rmu := sync.RWMutex{}
		currentLimit := 0
		limitExceeded := false

		_, _, _ = async.DetectSliceLimit(collection, func(val, idx int) (bool, error) {
			rmu.Lock()
			currentLimit += 1
			defer func() {
				currentLimit -= 1
				rmu.Unlock()
			}()
			if currentLimit > maxLimit {
				limitExceeded = true
			}
			time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
			return false, nil
		}, maxLimit)

		assert.False(nt, limitExceeded)
	})
}

func TestSomeSlice(t *testing.T) {
	t.Run("should delegate to SomeSliceLimit with len(collection)", func(nt *testing.T) {
		collection := []int{1, 2, 3, 4, 5}
		result, err := async.SomeSlice(collection, func(val, idx int) (bool, error) {
			return val == 3, nil
		})
		assert.NoError(nt, err)
		assert.True(nt, result)
	})
}

func TestSomeSliceLimit(t *testing.T) {
	t.Run("should return true for sync operations", func(nt *testing.T) {
		collection := []int{1, 2, 3, 4, 5}
		result, err := async.SomeSliceLimit(collection, func(val, idx int) (bool, error) {
			return val == 3, nil
		}, 2)
		assert.NoError(nt, err)
		assert.True(nt, result)
	})

	t.Run("should return true for async operations", func(nt *testing.T) {
		collection := []int{1, 2, 3, 4, 5}
		result, err := async.SomeSliceLimit(collection, func(val, idx int) (bool, error) {
			time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
			return val == 3, nil
		}, 3)
		assert.NoError(nt, err)
		assert.True(nt, result)
	})

	t.Run("should return false when no match", func(nt *testing.T) {
		collection := []int{1, 2, 3, 4, 5}
		result, err := async.SomeSliceLimit(collection, func(val, idx int) (bool, error) {
			return val == 99, nil
		}, 2)
		assert.NoError(nt, err)
		assert.False(nt, result)
	})

	t.Run("should return immediately on first true", func(nt *testing.T) {
		collection := []int{1, 2, 3, 4, 5, 6}
		startTime := time.Now()
		result, err := async.SomeSliceLimit(collection, func(val, idx int) (bool, error) {
			time.Sleep(200 * time.Millisecond)
			return val == 3, nil
		}, 3)
		elapsedTime := time.Since(startTime)
		assert.NoError(nt, err)
		assert.True(nt, result)
		assert.Less(nt, int(elapsedTime.Milliseconds()), 700)
	})

	t.Run("should return error immediately", func(nt *testing.T) {
		collection := []int{1, 2, 3, 4, 5}
		result, err := async.SomeSliceLimit(collection, func(val, idx int) (bool, error) {
			return false, errors.New("some error")
		}, 2)
		assert.Error(nt, err)
		assert.False(nt, result)
	})

	t.Run("should return error immediately with early termination", func(nt *testing.T) {
		collection := []int{1, 2, 3, 4, 5, 6}
		startTime := time.Now()
		result, err := async.SomeSliceLimit(collection, func(val, idx int) (bool, error) {
			time.Sleep(200 * time.Millisecond)
			return false, errors.New("some error")
		}, 3)
		elapsedTime := time.Since(startTime)
		assert.Error(nt, err)
		assert.False(nt, result)
		assert.Less(nt, int(elapsedTime.Milliseconds()), 700)
	})

	t.Run("should handle panic", func(nt *testing.T) {
		collection := []int{1, 2, 3, 4, 5}
		result, err := async.SomeSliceLimit(collection, func(val, idx int) (bool, error) {
			panic("some panic")
		}, 2)
		assert.Error(nt, err)
		assert.False(nt, result)
	})

	t.Run("should not exceed concurrency limit", func(nt *testing.T) {
		collection := []int{1, 2, 3, 4, 5, 6, 7}
		maxLimit := 2
		rmu := sync.RWMutex{}
		currentLimit := 0
		limitExceeded := false

		_, _ = async.SomeSliceLimit(collection, func(val, idx int) (bool, error) {
			rmu.Lock()
			currentLimit += 1
			defer func() {
				currentLimit -= 1
				rmu.Unlock()
			}()
			if currentLimit > maxLimit {
				limitExceeded = true
			}
			time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
			return false, nil
		}, maxLimit)

		assert.False(nt, limitExceeded)
	})
}

func TestEverySlice(t *testing.T) {
	t.Run("should delegate to EverySliceLimit with len(collection)", func(nt *testing.T) {
		collection := []int{1, 2, 3, 4, 5}
		result, err := async.EverySlice(collection, func(val, idx int) (bool, error) {
			return val > 0, nil
		})
		assert.NoError(nt, err)
		assert.True(nt, result)
	})
}

func TestEverySliceLimit(t *testing.T) {
	t.Parallel()

	t.Run("should return true for sync operations when all pass", func(nt *testing.T) {
		collection := []int{1, 2, 3, 4, 5}
		result, err := async.EverySliceLimit(collection, func(val, idx int) (bool, error) {
			return val > 0, nil
		}, 2)
		assert.NoError(nt, err)
		assert.True(nt, result)
	})

	t.Run("should return true for async operations when all pass", func(nt *testing.T) {
		collection := []int{1, 2, 3, 4, 5}
		result, err := async.EverySliceLimit(collection, func(val, idx int) (bool, error) {
			time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
			return val > 0, nil
		}, 2)
		assert.NoError(nt, err)
		assert.True(nt, result)
	})

	t.Run("should return false when one element fails", func(nt *testing.T) {
		collection := []int{1, 2, 3, 4, 5}
		result, err := async.EverySliceLimit(collection, func(val, idx int) (bool, error) {
			return val != 3, nil
		}, 2)
		assert.NoError(nt, err)
		assert.False(nt, result)
	})

	t.Run("should return false immediately on first false with early termination", func(nt *testing.T) {
		collection := []int{1, 2, 3, 4, 5, 6}
		startTime := time.Now()
		result, err := async.EverySliceLimit(collection, func(val, idx int) (bool, error) {
			time.Sleep(200 * time.Millisecond)
			return val != 3, nil
		}, 3)
		elapsedTime := time.Since(startTime)
		assert.NoError(nt, err)
		assert.False(nt, result)
		assert.Less(nt, int(elapsedTime.Milliseconds()), 500)
	})

	t.Run("should return error immediately", func(nt *testing.T) {
		collection := []int{1, 2, 3, 4, 5}
		result, err := async.EverySliceLimit(collection, func(val, idx int) (bool, error) {
			if val == 3 {
				return false, errors.New("some error")
			}
			return true, nil
		}, 2)
		assert.Error(nt, err)
		assert.False(nt, result)
	})

	t.Run("should return error immediately with early termination", func(nt *testing.T) {
		collection := []int{1, 2, 3, 4, 5, 6}
		startTime := time.Now()
		result, err := async.EverySliceLimit(collection, func(val, idx int) (bool, error) {
			time.Sleep(200 * time.Millisecond)
			return false, errors.New("some error")
		}, 3)
		elapsedTime := time.Since(startTime)
		assert.Error(nt, err)
		assert.False(nt, result)
		assert.Less(nt, int(elapsedTime.Milliseconds()), 500)
	})

	t.Run("should handle panic", func(nt *testing.T) {
		collection := []int{1, 2, 3, 4, 5}
		result, err := async.EverySliceLimit(collection, func(val, idx int) (bool, error) {
			if val == 3 {
				panic("some panic")
			}
			return true, nil
		}, 2)
		assert.Error(nt, err)
		assert.False(nt, result)
	})

	t.Run("should not exceed concurrency limit", func(nt *testing.T) {
		collection := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
		maxLimit := 2
		rmu := sync.RWMutex{}
		currentLimit := 0
		limitExceeded := false

		result, err := async.EverySliceLimit(collection, func(val, idx int) (bool, error) {
			rmu.Lock()
			currentLimit += 1
			defer func() {
				currentLimit -= 1
				rmu.Unlock()
			}()
			if currentLimit > maxLimit {
				limitExceeded = true
			}
			time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
			return true, nil
		}, maxLimit)

		assert.NoError(nt, err)
		assert.True(nt, result)
		assert.False(nt, limitExceeded)
	})

	t.Run("should return true for empty collection", func(nt *testing.T) {
		collection := []int{}
		result, err := async.EverySliceLimit(collection, func(val, idx int) (bool, error) {
			return false, nil
		}, 2)
		assert.NoError(nt, err)
		assert.True(nt, result)
	})
}

func TestConcatSlice(t *testing.T) {
	t.Run("should delegate to ConcatSliceLimit with len(collection)", func(nt *testing.T) {
		collection := []int{1, 2, 3}
		result, err := async.ConcatSlice(collection, func(val, idx int) ([]int, error) {
			return []int{val, val * 2}, nil
		})
		assert.NoError(nt, err)
		assert.Len(nt, result, 6)
		assert.ElementsMatch(nt, result, []int{1, 2, 2, 4, 3, 6})
	})
}

func TestConcatSliceLimit(t *testing.T) {
	t.Run("should concat correctly for sync operations", func(nt *testing.T) {
		collection := []int{1, 2, 3}
		result, err := async.ConcatSliceLimit(collection, func(val, idx int) ([]int, error) {
			return []int{val, val * 2}, nil
		}, 2)
		assert.NoError(nt, err)
		assert.Len(nt, result, 6)
		assert.ElementsMatch(nt, result, []int{1, 2, 2, 4, 3, 6})
	})

	t.Run("should concat correctly for async operations", func(nt *testing.T) {
		collection := []int{1, 2, 3}
		result, err := async.ConcatSliceLimit(collection, func(val, idx int) ([]int, error) {
			time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
			return []int{val, val * 2}, nil
		}, 3)
		assert.NoError(nt, err)
		assert.Len(nt, result, 6)
		assert.ElementsMatch(nt, result, []int{1, 2, 2, 4, 3, 6})
	})

	t.Run("should handle empty results", func(nt *testing.T) {
		collection := []int{1, 2, 3}
		result, err := async.ConcatSliceLimit(collection, func(val, idx int) ([]int, error) {
			if val%2 == 0 {
				return []int{}, nil
			}
			return []int{val}, nil
		}, 2)
		assert.NoError(nt, err)
		assert.ElementsMatch(nt, result, []int{1, 3})
	})

	t.Run("should return error immediately", func(nt *testing.T) {
		collection := []int{1, 2, 3}
		result, err := async.ConcatSliceLimit(collection, func(val, idx int) ([]int, error) {
			return nil, errors.New("some error")
		}, 2)
		assert.Error(nt, err)
		assert.Nil(nt, result)
	})

	t.Run("should return error immediately with early termination", func(nt *testing.T) {
		collection := []int{1, 2, 3, 4, 5, 6}
		startTime := time.Now()
		result, err := async.ConcatSliceLimit(collection, func(val, idx int) ([]int, error) {
			time.Sleep(200 * time.Millisecond)
			return nil, errors.New("some error")
		}, 3)
		elapsedTime := time.Since(startTime)
		assert.Error(nt, err)
		assert.Nil(nt, result)
		assert.Less(nt, int(elapsedTime.Milliseconds()), 700)
	})

	t.Run("should handle panic", func(nt *testing.T) {
		collection := []int{1, 2, 3}
		result, err := async.ConcatSliceLimit(collection, func(val, idx int) ([]int, error) {
			panic("some panic")
		}, 2)
		assert.Error(nt, err)
		assert.Nil(nt, result)
	})

	t.Run("should not exceed concurrency limit", func(nt *testing.T) {
		collection := []int{1, 2, 3, 4, 5, 6, 7}
		maxLimit := 2
		rmu := sync.RWMutex{}
		currentLimit := 0
		limitExceeded := false

		_, _ = async.ConcatSliceLimit(collection, func(val, idx int) ([]int, error) {
			rmu.Lock()
			currentLimit += 1
			defer func() {
				currentLimit -= 1
				rmu.Unlock()
			}()
			if currentLimit > maxLimit {
				limitExceeded = true
			}
			time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
			return []int{val}, nil
		}, maxLimit)

		assert.False(nt, limitExceeded)
	})
}

func TestRejectSlice(t *testing.T) {
	t.Run("should delegate to RejectSliceLimit with len(collection)", func(nt *testing.T) {
		collection := []int{1, 2, 3, 4, 5}
		expected := []int{1, 3, 5}
		result, err := async.RejectSlice(collection, func(val, idx int) (bool, error) {
			return val%2 == 0, nil
		})
		assert.NoError(nt, err)
		assert.ElementsMatch(nt, result, expected)
	})
}

func TestRejectSliceLimit(t *testing.T) {
	t.Run("should reject correctly for sync operations", func(nt *testing.T) {
		collection := []int{1, 2, 3, 4, 5}
		expected := []int{1, 3, 5}
		result, err := async.RejectSliceLimit(collection, func(val, idx int) (bool, error) {
			return val%2 == 0, nil
		}, 2)
		assert.NoError(nt, err)
		assert.ElementsMatch(nt, result, expected)
	})

	t.Run("should reject correctly for async operations", func(nt *testing.T) {
		collection := []int{1, 2, 3, 4, 5}
		expected := []int{1, 3, 5}
		result, err := async.RejectSliceLimit(collection, func(val, idx int) (bool, error) {
			time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
			return val%2 == 0, nil
		}, 3)
		assert.NoError(nt, err)
		assert.ElementsMatch(nt, result, expected)
	})

	t.Run("should return empty slice when all rejected", func(nt *testing.T) {
		collection := []int{1, 2, 3, 4, 5}
		result, err := async.RejectSliceLimit(collection, func(val, idx int) (bool, error) {
			return true, nil
		}, 2)
		assert.NoError(nt, err)
		assert.Empty(nt, result)
	})

	t.Run("should return all elements when none rejected", func(nt *testing.T) {
		collection := []int{1, 2, 3, 4, 5}
		result, err := async.RejectSliceLimit(collection, func(val, idx int) (bool, error) {
			return false, nil
		}, 2)
		assert.NoError(nt, err)
		assert.ElementsMatch(nt, result, collection)
	})

	t.Run("should reject with unordered results", func(nt *testing.T) {
		collection := []int{5, 2, 4, 1, 3}
		expected := []int{5, 1, 3}
		result, err := async.RejectSliceLimit(collection, func(val, idx int) (bool, error) {
			return val%2 == 0, nil
		}, 2)
		assert.NoError(nt, err)
		assert.ElementsMatch(nt, result, expected)
	})

	t.Run("should return error immediately", func(nt *testing.T) {
		collection := []int{1, 2, 3, 4, 5}
		result, err := async.RejectSliceLimit(collection, func(val, idx int) (bool, error) {
			return false, errors.New("some error")
		}, 2)
		assert.Error(nt, err)
		assert.Nil(nt, result)
	})

	t.Run("should return error immediately with early termination", func(nt *testing.T) {
		collection := []int{1, 2, 3, 4, 5, 6}
		startTime := time.Now()
		result, err := async.RejectSliceLimit(collection, func(val, idx int) (bool, error) {
			time.Sleep(200 * time.Millisecond)
			return false, errors.New("some error")
		}, 3)
		elapsedTime := time.Since(startTime)
		assert.Error(nt, err)
		assert.Nil(nt, result)
		assert.Less(nt, int(elapsedTime.Milliseconds()), 700)
	})

	t.Run("should handle panic", func(nt *testing.T) {
		collection := []int{1, 2, 3, 4, 5}
		result, err := async.RejectSliceLimit(collection, func(val, idx int) (bool, error) {
			panic("some panic")
		}, 2)
		assert.Error(nt, err)
		assert.Nil(nt, result)
	})

	t.Run("should not exceed concurrency limit", func(nt *testing.T) {
		collection := []int{1, 2, 3, 4, 5, 6, 7}
		maxLimit := 2
		rmu := sync.RWMutex{}
		currentLimit := 0
		limitExceeded := false

		_, _ = async.RejectSliceLimit(collection, func(val, idx int) (bool, error) {
			rmu.Lock()
			currentLimit += 1
			defer func() {
				currentLimit -= 1
				rmu.Unlock()
			}()
			if currentLimit > maxLimit {
				limitExceeded = true
			}
			time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
			return false, nil
		}, maxLimit)

		assert.False(nt, limitExceeded)
	})
}

func TestGroupBySlice(t *testing.T) {
	t.Run("should delegate to GroupBySliceLimit with len(collection)", func(nt *testing.T) {
		collection := []int{1, 2, 3, 4, 5, 6}
		result, err := async.GroupBySlice(collection, func(val, idx int) (string, int, error) {
			if val%2 == 0 {
				return "even", val, nil
			}
			return "odd", val, nil
		})
		assert.NoError(nt, err)
		assert.Len(nt, result, 2)
		assert.ElementsMatch(nt, result["odd"], []int{1, 3, 5})
		assert.ElementsMatch(nt, result["even"], []int{2, 4, 6})
	})
}

func TestGroupBySliceLimit(t *testing.T) {
	t.Run("should group correctly for sync operations", func(nt *testing.T) {
		collection := []int{1, 2, 3, 4, 5, 6}
		result, err := async.GroupBySliceLimit(collection, func(val, idx int) (string, int, error) {
			if val%2 == 0 {
				return "even", val, nil
			}
			return "odd", val, nil
		}, 2)
		assert.NoError(nt, err)
		assert.Len(nt, result, 2)
		assert.ElementsMatch(nt, result["odd"], []int{1, 3, 5})
		assert.ElementsMatch(nt, result["even"], []int{2, 4, 6})
	})

	t.Run("should group correctly for async operations", func(nt *testing.T) {
		collection := []int{1, 2, 3, 4, 5, 6}
		result, err := async.GroupBySliceLimit(collection, func(val, idx int) (string, int, error) {
			time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
			if val%2 == 0 {
				return "even", val, nil
			}
			return "odd", val, nil
		}, 3)
		assert.NoError(nt, err)
		assert.Len(nt, result, 2)
		assert.ElementsMatch(nt, result["odd"], []int{1, 3, 5})
		assert.ElementsMatch(nt, result["even"], []int{2, 4, 6})
	})

	t.Run("should handle multiple values per group", func(nt *testing.T) {
		collection := []int{1, 2, 3, 4, 5, 6}
		result, err := async.GroupBySliceLimit(collection, func(val, idx int) (string, int, error) {
			if val%2 == 0 {
				return "even", val, nil
			}
			return "odd", val, nil
		}, 2)
		assert.NoError(nt, err)
		assert.Len(nt, result, 2)
		assert.Len(nt, result["odd"], 3)
		assert.Len(nt, result["even"], 3)
	})

	t.Run("should return error immediately", func(nt *testing.T) {
		collection := []int{1, 2, 3, 4, 5}
		result, err := async.GroupBySliceLimit(collection, func(val, idx int) (string, int, error) {
			return "", 0, errors.New("some error")
		}, 2)
		assert.Error(nt, err)
		assert.Nil(nt, result)
	})

	t.Run("should return error immediately with early termination", func(nt *testing.T) {
		collection := []int{1, 2, 3, 4, 5, 6}
		startTime := time.Now()
		result, err := async.GroupBySliceLimit(collection, func(val, idx int) (string, int, error) {
			time.Sleep(200 * time.Millisecond)
			return "", 0, errors.New("some error")
		}, 3)
		elapsedTime := time.Since(startTime)
		assert.Error(nt, err)
		assert.Nil(nt, result)
		assert.Less(nt, int(elapsedTime.Milliseconds()), 700)
	})

	t.Run("should handle panic", func(nt *testing.T) {
		collection := []int{1, 2, 3, 4, 5}
		result, err := async.GroupBySliceLimit(collection, func(val, idx int) (string, int, error) {
			panic("some panic")
		}, 2)
		assert.Error(nt, err)
		assert.Nil(nt, result)
	})

	t.Run("should not exceed concurrency limit", func(nt *testing.T) {
		collection := []int{1, 2, 3, 4, 5, 6, 7}
		maxLimit := 2
		rmu := sync.RWMutex{}
		currentLimit := 0
		limitExceeded := false

		_, _ = async.GroupBySliceLimit(collection, func(val, idx int) (string, int, error) {
			rmu.Lock()
			currentLimit += 1
			defer func() {
				currentLimit -= 1
				rmu.Unlock()
			}()
			if currentLimit > maxLimit {
				limitExceeded = true
			}
			time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
			return "group", val, nil
		}, maxLimit)

		assert.False(nt, limitExceeded)
	})
}

func TestEachSliceLimit_ErrorHandling(t *testing.T) {
	t.Run("should return error immediately", func(nt *testing.T) {
		collection := []int{1, 2, 3, 4, 5}
		err := async.EachSliceLimit(collection, func(val, idx int) error {
			return errors.New("some error")
		}, 2)
		assert.Error(nt, err)
	})

	t.Run("should handle panic", func(nt *testing.T) {
		collection := []int{1, 2, 3, 4, 5}
		err := async.EachSliceLimit(collection, func(val, idx int) error {
			panic("some panic")
		}, 2)
		assert.Error(nt, err)
	})
}

func TestMapSliceLimit_ErrorHandling(t *testing.T) {
	t.Run("should return error immediately", func(nt *testing.T) {
		collection := []int{1, 2, 3, 4, 5}
		result, err := async.MapSliceLimit(collection, func(val, idx int) (int, error) {
			return 0, errors.New("some error")
		}, 2)
		assert.Error(nt, err)
		assert.Nil(nt, result)
	})

	t.Run("should handle panic", func(nt *testing.T) {
		collection := []int{1, 2, 3, 4, 5}
		result, err := async.MapSliceLimit(collection, func(val, idx int) (int, error) {
			panic("some panic")
		}, 2)
		assert.Error(nt, err)
		assert.Nil(nt, result)
	})
}
