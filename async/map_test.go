package async_test

import (
	"errors"
	"math/rand"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/skatiyar/goutils/async"
	"github.com/stretchr/testify/assert"
)

func TestConcatMapLimit(t *testing.T) {
	t.Run("should return correct values for sync operations", func(nt *testing.T) {
		collection := map[string]string{"1": "the brown", "2": "fox", "3": "jumps over the", "4": "brown fence"}
		expectedResult := []string{"brown", "fox", "jumps", "over", "brown", "fence"}
		r, rerr := async.ConcatMapLimit(collection, func(key, val string) ([]string, error) {
			return strings.Split(strings.Trim(strings.ReplaceAll(val, "the", ""), " "), " "), nil
		}, 2)
		assert.NoError(nt, rerr)
		assert.ElementsMatch(nt, r, expectedResult)
	})

	t.Run("should return correct values for async operations", func(nt *testing.T) {
		collection := map[string]string{"1": "the brown", "2": "fox", "3": "jumps over the", "4": "brown fence"}
		expectedResult := []string{"brown", "fox", "jumps", "over", "brown", "fence"}
		r, rerr := async.ConcatMapLimit(collection, func(key, val string) ([]string, error) {
			time.Sleep(time.Duration(rand.Intn(50)) * time.Millisecond)
			return strings.Split(strings.Trim(strings.ReplaceAll(val, "the", ""), " "), " "), nil
		}, 3)
		assert.NoError(nt, rerr)
		assert.ElementsMatch(nt, r, expectedResult)
	})

	t.Run("should return error if function returns error", func(nt *testing.T) {
		collection := map[string]string{"1": "the brown", "2": "fox", "3": "jumps over the", "4": "brown fence"}
		r, rerr := async.ConcatMapLimit(collection, func(key, val string) ([]string, error) {
			return nil, errors.New("some error")
		}, 2)
		assert.Error(nt, rerr)
		assert.Nil(nt, r)
	})

}

func TestConcatMap(t *testing.T) {
	t.Run("should spawn correct number of goroutines", func(nt *testing.T) {
		collection := map[string]string{"1": "the brown", "2": "fox", "3": "jumps over the", "4": "brown fence"}
		expectedGoroutines := len(collection)
		beforeGoroutines := runtime.NumGoroutine()
		_, _ = async.ConcatMap(collection, func(key, val string) ([]string, error) {
			time.Sleep(time.Duration(rand.Intn(50)) * time.Millisecond)
			return strings.Split(strings.Trim(strings.ReplaceAll(val, "the", ""), " "), " "), nil
		})
		afterGoroutines := runtime.NumGoroutine()
		actualGoroutines := afterGoroutines - beforeGoroutines
		assert.LessOrEqual(nt, actualGoroutines, expectedGoroutines)
	})
}

func TestEachMap(t *testing.T) {
	t.Run("should return correct values for sync operations", func(nt *testing.T) {
		collection := map[string]string{"1": "the brown", "2": "fox", "3": "jumps over the", "4": "brown fence"}
		expectedResult := []string{"brown", "fox", "jumps over", "brown fence"}
		rmu := sync.RWMutex{}
		results := make([]string, 0)
		async.EachMap(collection, func(key, val string) error {
			rmu.Lock()
			defer rmu.Unlock()
			results = append(results, strings.Trim(strings.ReplaceAll(val, "the", ""), " "))
			return nil
		})
		assert.ElementsMatch(nt, results, expectedResult)
	})
	t.Run("should return correct values for async operations", func(nt *testing.T) {
		collection := map[string]string{"1": "the brown", "2": "fox", "3": "jumps over the", "4": "brown fence"}
		expectedResult := []string{"brown", "fox", "jumps over", "brown fence"}
		rmu := sync.RWMutex{}
		results := make([]string, 0)
		async.EachMap(collection, func(key, val string) error {
			time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
			rmu.Lock()
			defer rmu.Unlock()
			results = append(results, strings.Trim(strings.ReplaceAll(val, "the", ""), " "))
			return nil
		})
		assert.ElementsMatch(nt, results, expectedResult)
	})
}

func TestEachMapLimit(t *testing.T) {
	t.Run("should return correct values for sync operations", func(nt *testing.T) {
		collection := map[string]string{"1": "the brown", "2": "fox", "3": "jumps over the", "4": "brown fence", "5": "and over", "6": "the lazy", "7": "dog"}
		expectedResult := []string{"brown", "fox", "jumps over", "brown fence", "and over", "lazy", "dog"}
		rmu := sync.RWMutex{}
		results := make([]string, 0)
		maxLimit := 2
		currentLimit := 0
		limitExceeded := false
		async.EachMapLimit(collection, func(key, val string) error {
			rmu.Lock()
			currentLimit += 1
			defer func() {
				currentLimit -= 1
				rmu.Unlock()
			}()
			if currentLimit > maxLimit {
				limitExceeded = true
			}
			results = append(results, strings.Trim(strings.ReplaceAll(val, "the", ""), " "))
			return nil
		}, maxLimit)
		assert.ElementsMatch(nt, results, expectedResult)
		assert.False(nt, limitExceeded)
	})
	t.Run("should return correct values for async operations", func(nt *testing.T) {
		collection := map[string]string{"1": "the brown", "2": "fox", "3": "jumps over the", "4": "brown fence", "5": "and over", "6": "the lazy", "7": "dog"}
		expectedResult := []string{"brown", "fox", "jumps over", "brown fence", "and over", "lazy", "dog"}
		rmu := sync.RWMutex{}
		results := make([]string, 0)
		maxLimit := 4
		currentLimit := 0
		limitExceeded := false
		async.EachMapLimit(collection, func(key, val string) error {
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
			results = append(results, strings.Trim(strings.ReplaceAll(val, "the", ""), " "))
			return nil
		}, maxLimit)
		assert.ElementsMatch(nt, results, expectedResult)
		assert.False(nt, limitExceeded)
	})
}

func TestMap(t *testing.T) {
	t.Run("should return correct values for sync operations", func(nt *testing.T) {
		collection := map[string]string{"1": "the brown", "2": "fox", "3": "jumps over the", "4": "brown fence"}
		collectionResult := map[string]string{"1": "brown", "2": "fox", "3": "jumps over", "4": "brown fence"}
		r, rerr := async.Map(collection, func(key, val string) (string, string, error) {
			return key, strings.Trim(strings.ReplaceAll(val, "the", ""), " "), nil
		})
		assert.NoError(nt, rerr)
		assert.Equal(nt, r, collectionResult)
	})

	t.Run("should return correct values for async operations", func(nt *testing.T) {
		collection := map[string]string{"1": "the brown", "2": "fox", "3": "jumps over the", "4": "brown fence"}
		collectionResult := map[string]string{"1": "brown", "2": "fox", "3": "jumps over", "4": "brown fence"}
		r, rerr := async.Map(collection, func(key, val string) (string, string, error) {
			time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
			return key, strings.Trim(strings.ReplaceAll(val, "the", ""), " "), nil
		})
		assert.NoError(nt, rerr)
		assert.Equal(nt, collectionResult, r)
	})

	t.Run("should return error if function returns error", func(nt *testing.T) {
		collection := map[string]string{"1": "the brown", "2": "fox", "3": "jumps over the", "4": "brown fence"}
		r, rerr := async.Map(collection, func(key, val string) (string, string, error) {
			return key, "", errors.New("some error")
		})
		assert.Error(nt, rerr)
		assert.Nil(nt, r)
	})

	t.Run("should return immediately post error in function", func(nt *testing.T) {
		collection := map[string]string{"1": "the brown", "2": "fox", "3": "jumps over the", "4": "brown fence"}
		startTime := time.Now()
		r, rerr := async.Map(collection, func(key, val string) (string, string, error) {
			time.Sleep(200 * time.Millisecond)
			return key, "", errors.New("some error")
		})
		elapsedTime := time.Since(startTime)
		assert.Error(nt, rerr)
		assert.Nil(nt, r)
		assert.Less(nt, int(elapsedTime.Milliseconds()), 300) // ensuring it returned immediately after first error
	})

	t.Run("should return error if function panics", func(nt *testing.T) {
		collection := map[string]string{"1": "the brown", "2": "fox", "3": "jumps over the", "4": "brown fence"}
		r, rerr := async.Map(collection, func(key, val string) (string, string, error) {
			panic("some panic")
		})
		assert.Error(nt, rerr)
		assert.Nil(nt, r)
	})
}

func TestMapLimit(t *testing.T) {
	t.Run("should return correct values for sync operations", func(nt *testing.T) {
		collection := map[string]string{"1": "the brown", "2": "fox", "3": "jumps over the", "4": "brown fence", "5": "and over", "6": "the lazy", "7": "dog"}
		collectionResult := map[string]string{"1": "brown", "2": "fox", "3": "jumps over", "4": "brown fence", "5": "and over", "6": "lazy", "7": "dog"}
		maxLimit := 2
		rmu := sync.RWMutex{}
		currentLimit := 0
		limitExceeded := false
		result, resultErr := async.MapLimit(collection, func(key, val string) (string, string, error) {
			rmu.Lock()
			currentLimit += 1
			defer func() {
				currentLimit -= 1
				rmu.Unlock()
			}()
			if currentLimit > maxLimit {
				limitExceeded = true
			}
			return key, strings.Trim(strings.ReplaceAll(val, "the", ""), " "), nil
		}, maxLimit)
		assert.NoError(nt, resultErr)
		assert.Equal(nt, result, collectionResult)
		assert.False(nt, limitExceeded)
	})

	t.Run("should return correct values for async operations", func(nt *testing.T) {
		collection := map[string]string{"1": "the brown", "2": "fox", "3": "jumps over the", "4": "brown fence", "5": "and over", "6": "the lazy", "7": "dog"}
		collectionResult := map[string]string{"1": "brown", "2": "fox", "3": "jumps over", "4": "brown fence", "5": "and over", "6": "lazy", "7": "dog"}
		maxLimit := 4
		rmu := sync.RWMutex{}
		currentLimit := 0
		limitExceeded := false
		result, resultErr := async.MapLimit(collection, func(key, val string) (string, string, error) {
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
			return key, strings.Trim(strings.ReplaceAll(val, "the", ""), " "), nil
		}, maxLimit)
		assert.NoError(nt, resultErr)
		assert.Equal(nt, result, collectionResult)
		assert.False(nt, limitExceeded)
	})
}

func TestSomeMapLimit(t *testing.T) {
	t.Run("should return true if any function call returns true", func(nt *testing.T) {
		collection := map[string]string{"1": "the brown", "2": "fox", "3": "jumps over the", "4": "brown fence"}
		result, resultErr := async.SomeMapLimit(collection, func(key, val string) (bool, error) {
			time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
			return strings.Contains(val, "jumps"), nil
		}, 2)
		assert.NoError(nt, resultErr)
		assert.True(nt, result)
	})

	t.Run("should return false if all function calls return false", func(nt *testing.T) {
		collection := map[string]string{"1": "the brown", "2": "fox", "3": "jumps over the", "4": "brown fence"}
		result, resultErr := async.SomeMapLimit(collection, func(key, val string) (bool, error) {
			time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
			return strings.Contains(val, "cat"), nil
		}, 3)
		assert.NoError(nt, resultErr)
		assert.False(nt, result)
	})

	t.Run("should return error if any function call returns error", func(nt *testing.T) {
		collection := map[string]string{"1": "the brown", "2": "fox", "3": "jumps over the", "4": "brown fence"}
		result, resultErr := async.SomeMapLimit(collection, func(key, val string) (bool, error) {
			time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
			if strings.Contains(val, "jumps") {
				return false, errors.New("some error")
			}
			return false, nil
		}, 2)
		assert.Error(nt, resultErr)
		assert.False(nt, result)
	})

	t.Run("should return immediately post true or error in function", func(nt *testing.T) {
		collection := map[string]string{"1": "the brown", "2": "fox", "3": "jumps over the", "4": "brown fence"}
		startTime := time.Now()
		result, resultErr := async.SomeMapLimit(collection, func(key, val string) (bool, error) {
			time.Sleep(200 * time.Millisecond)
			if strings.Contains(val, "jumps") {
				return true, nil
			}
			return false, nil
		}, 3)
		elapsedTime := time.Since(startTime)
		assert.NoError(nt, resultErr)
		assert.True(nt, result)
		assert.Less(nt, int(elapsedTime.Milliseconds()), 500) // ensuring it returned immediately after first true (2 batches worst case + margin)
	})

	t.Run("should return error if function panics", func(nt *testing.T) {
		collection := map[string]string{"1": "the brown", "2": "fox", "3": "jumps over the", "4": "brown fence"}
		result, resultErr := async.SomeMapLimit(collection, func(key, val string) (bool, error) {
			time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
			panic("some panic")
		}, 2)
		assert.Error(nt, resultErr)
		assert.False(nt, result)
	})

	t.Run("should not exceed limit", func(nt *testing.T) {
		collection := map[string]string{"1": "the brown", "2": "fox", "3": "jumps over the", "4": "brown fence", "5": "and over", "6": "the lazy", "7": "dog"}
		maxLimit := 3
		rmu := sync.RWMutex{}
		currentLimit := 0
		limitExceeded := false
		_, _ = async.SomeMapLimit(collection, func(key, val string) (bool, error) {
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
			return false, nil
		}, maxLimit)
		assert.False(nt, limitExceeded)
	})
}

func TestEveryMap(t *testing.T) {
	t.Run("should delegate to EveryMapLimit with len(collection)", func(nt *testing.T) {
		collection := map[string]string{"1": "the brown", "2": "fox", "3": "jumps over the", "4": "brown fence"}
		result, err := async.EveryMap(collection, func(key, val string) (bool, error) {
			return len(val) > 0, nil
		})
		assert.NoError(nt, err)
		assert.True(nt, result)
	})
}

func TestEveryMapLimit(t *testing.T) {
	t.Parallel()

	t.Run("should return true if all function calls return true", func(nt *testing.T) {
		collection := map[string]string{"1": "the brown", "2": "fox", "3": "jumps over the", "4": "brown fence"}
		result, resultErr := async.EveryMapLimit(collection, func(key, val string) (bool, error) {
			return len(val) > 0, nil
		}, 2)
		assert.NoError(nt, resultErr)
		assert.True(nt, result)
	})

	t.Run("should return false if any function call returns false", func(nt *testing.T) {
		collection := map[string]string{"1": "the brown", "2": "fox", "3": "jumps over the", "4": "brown fence"}
		result, resultErr := async.EveryMapLimit(collection, func(key, val string) (bool, error) {
			return !strings.Contains(val, "jumps"), nil
		}, 2)
		assert.NoError(nt, resultErr)
		assert.False(nt, result)
	})

	t.Run("should return error if any function call returns error", func(nt *testing.T) {
		collection := map[string]string{"1": "the brown", "2": "fox", "3": "jumps over the", "4": "brown fence"}
		result, resultErr := async.EveryMapLimit(collection, func(key, val string) (bool, error) {
			if strings.Contains(val, "jumps") {
				return false, errors.New("some error")
			}
			return true, nil
		}, 2)
		assert.Error(nt, resultErr)
		assert.False(nt, result)
	})

	t.Run("should return immediately post false or error in function", func(nt *testing.T) {
		collection := map[string]string{"1": "the brown", "2": "fox", "3": "jumps over the", "4": "brown fence", "5": "quick", "6": "dog"}
		startTime := time.Now()
		result, resultErr := async.EveryMapLimit(collection, func(key, val string) (bool, error) {
			time.Sleep(200 * time.Millisecond)
			return !strings.Contains(val, "jumps"), nil
		}, 3)
		elapsedTime := time.Since(startTime)
		assert.NoError(nt, resultErr)
		assert.False(nt, result)
		assert.Less(nt, int(elapsedTime.Milliseconds()), 500)
	})

	t.Run("should handle panic", func(nt *testing.T) {
		collection := map[string]string{"1": "the brown", "2": "fox", "3": "jumps over the", "4": "brown fence"}
		result, resultErr := async.EveryMapLimit(collection, func(key, val string) (bool, error) {
			if strings.Contains(val, "jumps") {
				panic("some panic")
			}
			return true, nil
		}, 2)
		assert.Error(nt, resultErr)
		assert.False(nt, result)
	})

	t.Run("should not exceed limit", func(nt *testing.T) {
		collection := map[string]string{"1": "the", "2": "brown", "3": "fox", "4": "jumps", "5": "over", "6": "the", "7": "lazy", "8": "dog", "9": "quick", "10": "fence"}
		maxLimit := 2
		rmu := sync.RWMutex{}
		currentLimit := 0
		limitExceeded := false

		result, resultErr := async.EveryMapLimit(collection, func(key, val string) (bool, error) {
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

		assert.NoError(nt, resultErr)
		assert.True(nt, result)
		assert.False(nt, limitExceeded)
	})

	t.Run("should return true for empty collection", func(nt *testing.T) {
		collection := map[string]string{}
		result, resultErr := async.EveryMapLimit(collection, func(key, val string) (bool, error) {
			return false, nil
		}, 2)
		assert.NoError(nt, resultErr)
		assert.True(nt, result)
	})
}

func TestDetectMap(t *testing.T) {
	t.Run("should delegate to DetectMapLimit with len(collection)", func(nt *testing.T) {
		collection := map[string]string{"1": "the brown", "2": "fox", "3": "jumps over the", "4": "brown fence"}
		result, detected, err := async.DetectMap(collection, func(key, val string) (bool, error) {
			return strings.Contains(val, "jumps"), nil
		})
		assert.NoError(nt, err)
		assert.True(nt, detected)
		assert.Equal(nt, result, "jumps over the")
	})
}

func TestDetectMapLimit(t *testing.T) {
	t.Run("should detect matching value for sync operations", func(nt *testing.T) {
		collection := map[string]string{"1": "the brown", "2": "fox", "3": "jumps over the", "4": "brown fence"}
		result, detected, err := async.DetectMapLimit(collection, func(key, val string) (bool, error) {
			return strings.Contains(val, "jumps"), nil
		}, 2)
		assert.NoError(nt, err)
		assert.True(nt, detected)
		assert.Equal(nt, result, "jumps over the")
	})

	t.Run("should detect matching value for async operations", func(nt *testing.T) {
		collection := map[string]string{"1": "the brown", "2": "fox", "3": "jumps over the", "4": "brown fence"}
		result, detected, err := async.DetectMapLimit(collection, func(key, val string) (bool, error) {
			time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
			return strings.Contains(val, "jumps"), nil
		}, 3)
		assert.NoError(nt, err)
		assert.True(nt, detected)
		assert.Equal(nt, result, "jumps over the")
	})

	t.Run("should return zero value and false when no match", func(nt *testing.T) {
		collection := map[string]string{"1": "the brown", "2": "fox", "3": "jumps over the", "4": "brown fence"}
		result, detected, err := async.DetectMapLimit(collection, func(key, val string) (bool, error) {
			return strings.Contains(val, "cat"), nil
		}, 2)
		assert.NoError(nt, err)
		assert.False(nt, detected)
		assert.Equal(nt, result, "")
	})

	t.Run("should return immediately on first match", func(nt *testing.T) {
		collection := map[string]string{"1": "the brown", "2": "fox", "3": "jumps over the", "4": "brown fence"}
		startTime := time.Now()
		result, detected, err := async.DetectMapLimit(collection, func(key, val string) (bool, error) {
			time.Sleep(200 * time.Millisecond)
			return strings.Contains(val, "jumps"), nil
		}, 3)
		elapsedTime := time.Since(startTime)
		assert.NoError(nt, err)
		assert.True(nt, detected)
		assert.Equal(nt, result, "jumps over the")
		assert.Less(nt, int(elapsedTime.Milliseconds()), 500) // 2*200ms + margin for early termination
	})

	t.Run("should return error immediately", func(nt *testing.T) {
		collection := map[string]string{"1": "the brown", "2": "fox", "3": "jumps over the", "4": "brown fence"}
		_, detected, err := async.DetectMapLimit(collection, func(key, val string) (bool, error) {
			return false, errors.New("some error")
		}, 2)
		assert.Error(nt, err)
		assert.False(nt, detected)
	})

	t.Run("should return error immediately with early termination", func(nt *testing.T) {
		collection := map[string]string{"1": "the brown", "2": "fox", "3": "jumps over the", "4": "brown fence"}
		startTime := time.Now()
		_, detected, err := async.DetectMapLimit(collection, func(key, val string) (bool, error) {
			time.Sleep(200 * time.Millisecond)
			return false, errors.New("some error")
		}, 3)
		elapsedTime := time.Since(startTime)
		assert.Error(nt, err)
		assert.False(nt, detected)
		assert.Less(nt, int(elapsedTime.Milliseconds()), 500)
	})

	t.Run("should handle panic with string", func(nt *testing.T) {
		collection := map[string]string{"1": "the brown", "2": "fox", "3": "jumps over the", "4": "brown fence"}
		_, detected, err := async.DetectMapLimit(collection, func(key, val string) (bool, error) {
			panic("some panic")
		}, 2)
		assert.Error(nt, err)
		assert.False(nt, detected)
		assert.Contains(nt, err.Error(), "panic")
	})

	t.Run("should handle panic with error", func(nt *testing.T) {
		collection := map[string]string{"1": "the brown", "2": "fox", "3": "jumps over the", "4": "brown fence"}
		_, detected, err := async.DetectMapLimit(collection, func(key, val string) (bool, error) {
			panic(errors.New("panic error"))
		}, 2)
		assert.Error(nt, err)
		assert.False(nt, detected)
	})

	t.Run("should not exceed concurrency limit", func(nt *testing.T) {
		collection := map[string]string{"1": "a", "2": "b", "3": "c", "4": "d", "5": "e", "6": "f", "7": "g"}
		maxLimit := 2
		rmu := sync.RWMutex{}
		currentLimit := 0
		limitExceeded := false

		_, _, _ = async.DetectMapLimit(collection, func(key, val string) (bool, error) {
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

func TestSomeMap(t *testing.T) {
	t.Run("should delegate to SomeMapLimit with len(collection)", func(nt *testing.T) {
		collection := map[string]string{"1": "the brown", "2": "fox", "3": "jumps over the", "4": "brown fence"}
		result, err := async.SomeMap(collection, func(key, val string) (bool, error) {
			return strings.Contains(val, "jumps"), nil
		})
		assert.NoError(nt, err)
		assert.True(nt, result)
	})
}

func TestFilterMap(t *testing.T) {
	t.Run("should delegate to FilterMapLimit with len(collection)", func(nt *testing.T) {
		collection := map[string]string{"1": "the brown", "2": "fox", "3": "jumps over the", "4": "brown fence"}
		expected := map[string]string{"2": "fox", "4": "brown fence"}
		result, err := async.FilterMap(collection, func(key, val string) (bool, error) {
			return !strings.Contains(val, "the"), nil
		})
		assert.NoError(nt, err)
		assert.Equal(nt, result, expected)
	})
}

func TestFilterMapLimit(t *testing.T) {
	t.Run("should filter correctly for sync operations", func(nt *testing.T) {
		collection := map[string]string{"1": "the brown", "2": "fox", "3": "jumps over the", "4": "brown fence"}
		expected := map[string]string{"2": "fox", "4": "brown fence"}
		result, err := async.FilterMapLimit(collection, func(key, val string) (bool, error) {
			return !strings.Contains(val, "the"), nil
		}, 2)
		assert.NoError(nt, err)
		assert.Equal(nt, result, expected)
	})

	t.Run("should filter correctly for async operations", func(nt *testing.T) {
		collection := map[string]string{"1": "the brown", "2": "fox", "3": "jumps over the", "4": "brown fence"}
		expected := map[string]string{"2": "fox", "4": "brown fence"}
		result, err := async.FilterMapLimit(collection, func(key, val string) (bool, error) {
			time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
			return !strings.Contains(val, "the"), nil
		}, 3)
		assert.NoError(nt, err)
		assert.Equal(nt, result, expected)
	})

	t.Run("should return all entries when all pass", func(nt *testing.T) {
		collection := map[string]string{"1": "the brown", "2": "fox", "3": "jumps over the", "4": "brown fence"}
		result, err := async.FilterMapLimit(collection, func(key, val string) (bool, error) {
			return true, nil
		}, 2)
		assert.NoError(nt, err)
		assert.Equal(nt, result, collection)
	})

	t.Run("should return empty map when none pass", func(nt *testing.T) {
		collection := map[string]string{"1": "the brown", "2": "fox", "3": "jumps over the", "4": "brown fence"}
		result, err := async.FilterMapLimit(collection, func(key, val string) (bool, error) {
			return false, nil
		}, 2)
		assert.NoError(nt, err)
		assert.Empty(nt, result)
	})

	t.Run("should return error immediately", func(nt *testing.T) {
		collection := map[string]string{"1": "the brown", "2": "fox", "3": "jumps over the", "4": "brown fence"}
		result, err := async.FilterMapLimit(collection, func(key, val string) (bool, error) {
			return false, errors.New("some error")
		}, 2)
		assert.Error(nt, err)
		assert.Nil(nt, result)
	})

	t.Run("should return error immediately with early termination", func(nt *testing.T) {
		collection := map[string]string{"1": "the brown", "2": "fox", "3": "jumps over the", "4": "brown fence"}
		startTime := time.Now()
		result, err := async.FilterMapLimit(collection, func(key, val string) (bool, error) {
			time.Sleep(200 * time.Millisecond)
			return false, errors.New("some error")
		}, 3)
		elapsedTime := time.Since(startTime)
		assert.Error(nt, err)
		assert.Nil(nt, result)
		assert.Less(nt, int(elapsedTime.Milliseconds()), 500)
	})

	t.Run("should handle panic", func(nt *testing.T) {
		collection := map[string]string{"1": "the brown", "2": "fox", "3": "jumps over the", "4": "brown fence"}
		result, err := async.FilterMapLimit(collection, func(key, val string) (bool, error) {
			panic("some panic")
		}, 2)
		assert.Error(nt, err)
		assert.Nil(nt, result)
	})

	t.Run("should not exceed concurrency limit", func(nt *testing.T) {
		collection := map[string]string{"1": "a", "2": "b", "3": "c", "4": "d", "5": "e", "6": "f", "7": "g"}
		maxLimit := 3
		rmu := sync.RWMutex{}
		currentLimit := 0
		limitExceeded := false

		_, _ = async.FilterMapLimit(collection, func(key, val string) (bool, error) {
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

func TestRejectMap(t *testing.T) {
	t.Run("should delegate to RejectMapLimit with len(collection)", func(nt *testing.T) {
		collection := map[string]string{"1": "the brown", "2": "fox", "3": "jumps over the", "4": "brown fence"}
		expected := map[string]string{"1": "the brown", "3": "jumps over the"}
		result, err := async.RejectMap(collection, func(key, val string) (bool, error) {
			return !strings.Contains(val, "the"), nil
		})
		assert.NoError(nt, err)
		assert.Equal(nt, result, expected)
	})
}

func TestRejectMapLimit(t *testing.T) {
	t.Run("should reject correctly for sync operations", func(nt *testing.T) {
		collection := map[string]string{"1": "the brown", "2": "fox", "3": "jumps over the", "4": "brown fence"}
		expected := map[string]string{"1": "the brown", "3": "jumps over the"}
		result, err := async.RejectMapLimit(collection, func(key, val string) (bool, error) {
			return !strings.Contains(val, "the"), nil
		}, 2)
		assert.NoError(nt, err)
		assert.Equal(nt, result, expected)
	})

	t.Run("should reject correctly for async operations", func(nt *testing.T) {
		collection := map[string]string{"1": "the brown", "2": "fox", "3": "jumps over the", "4": "brown fence"}
		expected := map[string]string{"1": "the brown", "3": "jumps over the"}
		result, err := async.RejectMapLimit(collection, func(key, val string) (bool, error) {
			time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
			return !strings.Contains(val, "the"), nil
		}, 3)
		assert.NoError(nt, err)
		assert.Equal(nt, result, expected)
	})

	t.Run("should return empty map when all rejected", func(nt *testing.T) {
		collection := map[string]string{"1": "the brown", "2": "fox", "3": "jumps over the", "4": "brown fence"}
		result, err := async.RejectMapLimit(collection, func(key, val string) (bool, error) {
			return true, nil
		}, 2)
		assert.NoError(nt, err)
		assert.Empty(nt, result)
	})

	t.Run("should return all entries when none rejected", func(nt *testing.T) {
		collection := map[string]string{"1": "the brown", "2": "fox", "3": "jumps over the", "4": "brown fence"}
		result, err := async.RejectMapLimit(collection, func(key, val string) (bool, error) {
			return false, nil
		}, 2)
		assert.NoError(nt, err)
		assert.Equal(nt, result, collection)
	})

	t.Run("should return error immediately", func(nt *testing.T) {
		collection := map[string]string{"1": "the brown", "2": "fox", "3": "jumps over the", "4": "brown fence"}
		result, err := async.RejectMapLimit(collection, func(key, val string) (bool, error) {
			return false, errors.New("some error")
		}, 2)
		assert.Error(nt, err)
		assert.Nil(nt, result)
	})

	t.Run("should return error immediately with early termination", func(nt *testing.T) {
		collection := map[string]string{"1": "the brown", "2": "fox", "3": "jumps over the", "4": "brown fence"}
		startTime := time.Now()
		result, err := async.RejectMapLimit(collection, func(key, val string) (bool, error) {
			time.Sleep(200 * time.Millisecond)
			return false, errors.New("some error")
		}, 3)
		elapsedTime := time.Since(startTime)
		assert.Error(nt, err)
		assert.Nil(nt, result)
		assert.Less(nt, int(elapsedTime.Milliseconds()), 500)
	})

	t.Run("should handle panic", func(nt *testing.T) {
		collection := map[string]string{"1": "the brown", "2": "fox", "3": "jumps over the", "4": "brown fence"}
		result, err := async.RejectMapLimit(collection, func(key, val string) (bool, error) {
			panic("some panic")
		}, 2)
		assert.Error(nt, err)
		assert.Nil(nt, result)
	})

	t.Run("should not exceed concurrency limit", func(nt *testing.T) {
		collection := map[string]string{"1": "a", "2": "b", "3": "c", "4": "d", "5": "e", "6": "f", "7": "g"}
		maxLimit := 3
		rmu := sync.RWMutex{}
		currentLimit := 0
		limitExceeded := false

		_, _ = async.RejectMapLimit(collection, func(key, val string) (bool, error) {
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

func TestGroupByMap(t *testing.T) {
	t.Run("should delegate to GroupByMapLimit with len(collection)", func(nt *testing.T) {
		collection := map[string]int{"a": 1, "b": 2, "c": 3, "d": 4}
		result, err := async.GroupByMap(collection, func(key string, val int) (string, int, error) {
			if val%2 == 0 {
				return "even", val, nil
			}
			return "odd", val, nil
		})
		assert.NoError(nt, err)
		assert.Len(nt, result, 2)
		assert.ElementsMatch(nt, result["odd"], []int{1, 3})
		assert.ElementsMatch(nt, result["even"], []int{2, 4})
	})
}

func TestGroupByMapLimit(t *testing.T) {
	t.Run("should group correctly for sync operations", func(nt *testing.T) {
		collection := map[string]int{"a": 1, "b": 2, "c": 3, "d": 4}
		result, err := async.GroupByMapLimit(collection, func(key string, val int) (string, int, error) {
			if val%2 == 0 {
				return "even", val, nil
			}
			return "odd", val, nil
		}, 2)
		assert.NoError(nt, err)
		assert.Len(nt, result, 2)
		assert.ElementsMatch(nt, result["odd"], []int{1, 3})
		assert.ElementsMatch(nt, result["even"], []int{2, 4})
	})

	t.Run("should group correctly for async operations", func(nt *testing.T) {
		collection := map[string]int{"a": 1, "b": 2, "c": 3, "d": 4}
		result, err := async.GroupByMapLimit(collection, func(key string, val int) (string, int, error) {
			time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
			if val%2 == 0 {
				return "even", val, nil
			}
			return "odd", val, nil
		}, 3)
		assert.NoError(nt, err)
		assert.Len(nt, result, 2)
		assert.ElementsMatch(nt, result["odd"], []int{1, 3})
		assert.ElementsMatch(nt, result["even"], []int{2, 4})
	})

	t.Run("should handle multiple values per group", func(nt *testing.T) {
		collection := map[string]int{"a": 1, "b": 2, "c": 3, "d": 4, "e": 5, "f": 6}
		result, err := async.GroupByMapLimit(collection, func(key string, val int) (string, int, error) {
			if val%2 == 0 {
				return "even", val, nil
			}
			return "odd", val, nil
		}, 2)
		assert.NoError(nt, err)
		assert.Len(nt, result, 2)
		assert.Len(nt, result["odd"], 3)
		assert.Len(nt, result["even"], 3)
		assert.ElementsMatch(nt, result["odd"], []int{1, 3, 5})
		assert.ElementsMatch(nt, result["even"], []int{2, 4, 6})
	})

	t.Run("should return error immediately", func(nt *testing.T) {
		collection := map[string]int{"a": 1, "b": 2, "c": 3, "d": 4}
		result, err := async.GroupByMapLimit(collection, func(key string, val int) (string, int, error) {
			return "", 0, errors.New("some error")
		}, 2)
		assert.Error(nt, err)
		assert.Nil(nt, result)
	})

	t.Run("should return error immediately with early termination", func(nt *testing.T) {
		collection := map[string]int{"a": 1, "b": 2, "c": 3, "d": 4}
		startTime := time.Now()
		result, err := async.GroupByMapLimit(collection, func(key string, val int) (string, int, error) {
			time.Sleep(200 * time.Millisecond)
			return "", 0, errors.New("some error")
		}, 3)
		elapsedTime := time.Since(startTime)
		assert.Error(nt, err)
		assert.Nil(nt, result)
		assert.Less(nt, int(elapsedTime.Milliseconds()), 500)
	})

	t.Run("should handle panic", func(nt *testing.T) {
		collection := map[string]int{"a": 1, "b": 2, "c": 3, "d": 4}
		result, err := async.GroupByMapLimit(collection, func(key string, val int) (string, int, error) {
			panic("some panic")
		}, 2)
		assert.Error(nt, err)
		assert.Nil(nt, result)
	})

	t.Run("should not exceed concurrency limit", func(nt *testing.T) {
		collection := map[string]int{"a": 1, "b": 2, "c": 3, "d": 4, "e": 5, "f": 6, "g": 7}
		maxLimit := 2
		rmu := sync.RWMutex{}
		currentLimit := 0
		limitExceeded := false

		_, _ = async.GroupByMapLimit(collection, func(key string, val int) (string, int, error) {
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

func TestEachMapLimit_ErrorHandling(t *testing.T) {
	t.Run("should return error immediately", func(nt *testing.T) {
		collection := map[string]string{"1": "a", "2": "b", "3": "c"}
		err := async.EachMapLimit(collection, func(key, val string) error {
			return errors.New("some error")
		}, 2)
		assert.Error(nt, err)
	})

	t.Run("should handle panic", func(nt *testing.T) {
		collection := map[string]string{"1": "a", "2": "b", "3": "c"}
		err := async.EachMapLimit(collection, func(key, val string) error {
			panic("some panic")
		}, 2)
		assert.Error(nt, err)
	})
}

func TestConcatMapLimit_ErrorHandling(t *testing.T) {
	t.Run("should handle panic", func(nt *testing.T) {
		collection := map[string]string{"1": "a", "2": "b", "3": "c"}
		result, err := async.ConcatMapLimit(collection, func(key, val string) ([]string, error) {
			panic("some panic")
		}, 2)
		assert.Error(nt, err)
		assert.Nil(nt, result)
	})
}
