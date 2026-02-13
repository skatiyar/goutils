package async_test

import (
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
