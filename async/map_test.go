package async_test

import (
	"math/rand"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/skatiyar/goutils/async"
	"github.com/stretchr/testify/assert"
)

func TestEachMap(t *testing.T) {
	t.Run("should return correct values for sync operations", func(nt *testing.T) {
		collection := map[string]string{"1": "the brown", "2": "fox", "3": "jumps over the", "4": "brown fence"}
		expectedResult := []string{"brown", "fox", "jumps over", "brown fence"}
		rmu := sync.RWMutex{}
		results := make([]string, 0)
		async.EachMap(collection, func(key, val string) {
			rmu.Lock()
			defer rmu.Unlock()
			results = append(results, strings.Trim(strings.ReplaceAll(val, "the", ""), " "))
		})
		assert.ElementsMatch(nt, results, expectedResult)
	})
	t.Run("should return correct values for async operations", func(nt *testing.T) {
		collection := map[string]string{"1": "the brown", "2": "fox", "3": "jumps over the", "4": "brown fence"}
		expectedResult := []string{"brown", "fox", "jumps over", "brown fence"}
		rmu := sync.RWMutex{}
		results := make([]string, 0)
		async.EachMap(collection, func(key, val string) {
			time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
			rmu.Lock()
			defer rmu.Unlock()
			results = append(results, strings.Trim(strings.ReplaceAll(val, "the", ""), " "))
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
		async.EachMapLimit(collection, func(key, val string) {
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
		async.EachMapLimit(collection, func(key, val string) {
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
		}, maxLimit)
		assert.ElementsMatch(nt, results, expectedResult)
		assert.False(nt, limitExceeded)
	})
}

func TestMap(t *testing.T) {
	t.Run("should return correct values for sync operations", func(nt *testing.T) {
		collection := map[string]string{"1": "the brown", "2": "fox", "3": "jumps over the", "4": "brown fence"}
		collectionResult := map[string]string{"1": "brown", "2": "fox", "3": "jumps over", "4": "brown fence"}
		assert.Equal(nt, async.Map(collection, func(key, val string) (string, string) {
			return key, strings.Trim(strings.ReplaceAll(val, "the", ""), " ")
		}), collectionResult)
	})

	t.Run("should return correct values for async operations", func(nt *testing.T) {
		collection := map[string]string{"1": "the brown", "2": "fox", "3": "jumps over the", "4": "brown fence"}
		collectionResult := map[string]string{"1": "brown", "2": "fox", "3": "jumps over", "4": "brown fence"}
		assert.Equal(nt, async.Map(collection, func(key, val string) (string, string) {
			time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
			return key, strings.Trim(strings.ReplaceAll(val, "the", ""), " ")
		}), collectionResult)
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
		result := async.MapLimit(collection, func(key, val string) (string, string) {
			rmu.Lock()
			currentLimit += 1
			defer func() {
				currentLimit -= 1
				rmu.Unlock()
			}()
			if currentLimit > maxLimit {
				limitExceeded = true
			}
			return key, strings.Trim(strings.ReplaceAll(val, "the", ""), " ")
		}, maxLimit)
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
		result := async.MapLimit(collection, func(key, val string) (string, string) {
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
			return key, strings.Trim(strings.ReplaceAll(val, "the", ""), " ")
		}, maxLimit)
		assert.Equal(nt, result, collectionResult)
		assert.False(nt, limitExceeded)
	})
}
