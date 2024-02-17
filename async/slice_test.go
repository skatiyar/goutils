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
		async.EachSlice(collection, func(idx, value int) {
			rmu.Lock()
			defer rmu.Unlock()
			results = append(results, int(math.Pow(float64(value), 2)))
		})
		assert.ElementsMatch(nt, results, collectionResult)
	})
	t.Run("should return correct values for async operations", func(nt *testing.T) {
		collection := []int{2, 7, 8, 9, 1, 3}
		collectionResult := []int{4, 49, 64, 81, 1, 9}
		rmu := sync.RWMutex{}
		results := make([]int, 0)
		async.EachSlice(collection, func(idx, value int) {
			time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
			rmu.Lock()
			defer rmu.Unlock()
			results = append(results, int(math.Pow(float64(value), 2)))
		})
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
		async.EachSliceLimit(collection, func(idx, value int) {
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
		}, maxLimit)
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
		async.EachSlice(collection, func(idx, value int) {
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
		})
		assert.ElementsMatch(nt, results, collectionResult)
		assert.False(nt, limitExceeded)
	})
}

func TestSlice(t *testing.T) {
	t.Run("should return correct values for square of integers", func(nt *testing.T) {
		collection := []int{2, 7, 8, 9, 1, 3}
		collectionResult := []int{4, 49, 64, 81, 1, 9}
		assert.Equal(nt, async.Slice(collection, func(val int) int {
			return int(math.Pow(float64(val), 2))
		}), collectionResult)
	})

	t.Run("should return correct values for async operations", func(nt *testing.T) {
		collection := []int{2, 7, 8, 9, 1, 3}
		collectionResult := []int{4, 49, 64, 81, 1, 9}
		assert.Equal(nt, async.Slice(collection, func(val int) int {
			time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
			return int(math.Pow(float64(val), 2))
		}), collectionResult)
	})
}

func TestSliceLimit(t *testing.T) {
	t.Run("should return correct values for square of integers", func(nt *testing.T) {
		collection := []int{2, 7, 8, 9, 1, 3}
		collectionResult := []int{4, 49, 64, 81, 1, 9}
		maxLimit := 2
		rmu := sync.RWMutex{}
		currentLimit := 0
		limitExceeded := false
		assert.Equal(nt, async.SliceLimit(collection, func(val int) int {
			rmu.Lock()
			currentLimit += 1
			defer func() {
				currentLimit -= 1
				rmu.Unlock()
			}()
			if currentLimit > maxLimit {
				limitExceeded = true
			}
			return int(math.Pow(float64(val), 2))
		}, maxLimit), collectionResult)
		assert.False(nt, limitExceeded)
	})

	t.Run("should return correct values for async operations", func(nt *testing.T) {
		collection := []int{2, 7, 8, 9, 1, 3}
		collectionResult := []int{4, 49, 64, 81, 1, 9}
		maxLimit := 4
		rmu := sync.RWMutex{}
		currentLimit := 0
		limitExceeded := false
		assert.Equal(nt, async.SliceLimit(collection, func(val int) int {
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
			return int(math.Pow(float64(val), 2))
		}, maxLimit), collectionResult)
		assert.False(nt, limitExceeded)
	})
}
