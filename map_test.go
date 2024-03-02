package goutils_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/skatiyar/goutils"
	"github.com/stretchr/testify/assert"
)

func TestConcatMap(t *testing.T) {
	t.Run("should return correct values when error is nil", func(nt *testing.T) {
		collection := map[string]string{"1": "the brown", "2": "fox", "3": "jumps over the", "4": "brown fence"}
		collectionResult := []string{"the", "brown", "fox", "jumps", "over", "the", "brown", "fence"}
		result, resultErr := goutils.ConcatMap(collection, func(key, val string) ([]string, error) {
			return strings.Split(val, " "), nil
		})
		assert.NoError(nt, resultErr)
		assert.ElementsMatch(nt, result, collectionResult)
	})
	t.Run("should return correct values when error is not nil", func(nt *testing.T) {
		collection := map[string]string{"1": "the brown", "2": "fox", "3": "jumps over the", "4": "brown fence"}
		result, resultErr := goutils.ConcatMap(collection, func(key, val string) ([]string, error) {
			return strings.Split(val, " "), errors.New("an error")
		})
		assert.Error(nt, resultErr)
		assert.Equal(nt, result, []string(nil))
	})
}

func TestDetectMap(t *testing.T) {
	t.Run("should return correct value when detected but error is nil", func(nt *testing.T) {
		collection := map[string]string{"1": "the brown", "2": "fox", "3": "jumps over the", "4": "brown fence"}
		value, detected, err := goutils.DetectMap(collection, func(key, val string) (bool, error) {
			return strings.Contains(val, "fox"), nil
		})
		assert.NoError(nt, err)
		assert.Truef(nt, detected, "Value found", value)
	})
	t.Run("should return correct value when not detected but error is nil", func(nt *testing.T) {
		collection := map[string]string{"1": "the brown", "2": "fox", "3": "jumps over the", "4": "brown fence"}
		value, detected, err := goutils.DetectMap(collection, func(key, val string) (bool, error) {
			return strings.Contains(val, "dog"), nil
		})
		assert.NoError(nt, err)
		assert.Falsef(nt, detected, "Value not found", value)
	})
	t.Run("should return correct value when error is not nil", func(nt *testing.T) {
		collection := map[string]string{"1": "the brown", "2": "fox", "3": "jumps over the", "4": "brown fence"}
		value, detected, err := goutils.DetectMap(collection, func(key, val string) (bool, error) {
			return strings.Contains(val, "fox"), errors.New("an error")
		})
		assert.Error(nt, err)
		assert.Falsef(nt, detected, "Value not found", value)
	})
}

func TestEachMap(t *testing.T) {
	t.Run("should pass when iterator returns no error", func(nt *testing.T) {
		collection := map[string]string{"1": "the brown", "2": "fox", "3": "jumps over the", "4": "brown fence"}
		collectionResult := map[string]string{"1": "brown", "2": "fox", "3": "jumps over", "4": "brown fence"}
		result := make(map[string]string)
		err := goutils.EachMap(collection, func(key, val string) error {
			result[key] = strings.Trim(strings.ReplaceAll(val, "the", ""), " ")
			return nil
		})
		assert.NoError(nt, err)
		assert.Equal(nt, result, collectionResult)
	})
	t.Run("should pass when iterator returns error", func(nt *testing.T) {
		collection := map[string]string{"1": "the brown", "2": "fox", "3": "jumps over the", "4": "brown fence"}
		result := make(map[string]string)
		err := goutils.EachMap(collection, func(key, val string) error {
			result[key] = strings.Trim(strings.ReplaceAll(val, "the", ""), " ")
			return errors.New("an error")
		})
		assert.Error(nt, err)
		assert.NotEmpty(nt, result)
	})
}

func TestMap(t *testing.T) {
	t.Run("should return correct values when iterator returns no error", func(nt *testing.T) {
		collection := map[string]string{"1": "the brown", "2": "fox", "3": "jumps over the", "4": "brown fence"}
		collectionResult := map[string]string{"1": "brown", "2": "fox", "3": "jumps over", "4": "brown fence"}
		mapped, mappedErr := goutils.Map(collection, func(key, val string) (string, string, error) {
			return key, strings.Trim(strings.ReplaceAll(val, "the", ""), " "), nil
		})
		assert.NoError(nt, mappedErr)
		assert.Equal(nt, mapped, collectionResult)
	})
	t.Run("should return correct values when iterator returns error", func(nt *testing.T) {
		collection := map[string]string{"1": "the brown", "2": "fox", "3": "jumps over the", "4": "brown fence"}
		mapped, mappedErr := goutils.Map(collection, func(key, val string) (string, string, error) {
			return key, strings.Trim(strings.ReplaceAll(val, "the", ""), " "), errors.New("an error")
		})
		assert.Error(nt, mappedErr)
		assert.Nil(nt, mapped)
	})
}

func TestReduceMap(t *testing.T) {
	t.Run("should return correct values when iterator returns no error", func(nt *testing.T) {
		collection := map[string]string{"1": "the brown", "2": "fox", "3": "jumps over the", "4": "brown fence"}
		collectionResult := []string{"brown", "fox", "jumps over", "brown fence"}
		mapped, mappedErr := goutils.ReduceMap(collection, func(acc []string, key, val string) ([]string, error) {
			return append(acc, strings.Trim(strings.ReplaceAll(val, "the", ""), " ")), nil
		}, []string{})
		assert.NoError(nt, mappedErr)
		assert.ElementsMatch(nt, mapped, collectionResult)
	})
	t.Run("should return correct values when iterator returns error", func(nt *testing.T) {
		collection := map[string]string{"1": "the brown", "2": "fox", "3": "jumps over the", "4": "brown fence"}
		mapped, mappedErr := goutils.ReduceMap(collection, func(acc string, key, val string) (string, error) {
			return acc + " " + val, errors.New("an error")
		}, "")
		assert.Error(nt, mappedErr)
		assert.Empty(nt, mapped)
	})
}

func TestEveryMap(t *testing.T) {
	t.Run("should return correct values when iterator returns no error and all values pass test", func(nt *testing.T) {
		collection := map[string]string{"1": "the brown", "2": "fox", "3": "jumps over the", "4": "brown fence"}
		mapped, mappedErr := goutils.EveryMap(collection, func(key, val string) (bool, error) {
			return strings.ContainsAny(val, "aeiou"), nil
		})
		assert.NoError(nt, mappedErr)
		assert.True(nt, mapped)
	})
	t.Run("should return correct values when iterator returns no error and one value returs false", func(nt *testing.T) {
		collection := map[string]string{"1": "the brown", "2": "fly", "3": "jumps over the", "4": "brown fence"}
		mapped, mappedErr := goutils.EveryMap(collection, func(key, val string) (bool, error) {
			return strings.ContainsAny(val, "aeiou"), nil
		})
		assert.NoError(nt, mappedErr)
		assert.False(nt, mapped)
	})
	t.Run("should return correct values when iterator returns error", func(nt *testing.T) {
		collection := map[string]string{"1": "the brown", "2": "fly", "3": "jumps over the", "4": "brown fence"}
		mapped, mappedErr := goutils.EveryMap(collection, func(key, val string) (bool, error) {
			return strings.ContainsAny(val, "aeiou"), errors.New("an error")
		})
		assert.Error(nt, mappedErr)
		assert.False(nt, mapped)
	})
}

func TestFilterMap(t *testing.T) {
	t.Run("should return correct values when iterator returns no error", func(nt *testing.T) {
		collection := map[string]string{"1": "the brown", "2": "fox", "3": "jumps over the", "4": "brown fence"}
		mapped, mappedErr := goutils.FilterMap(collection, func(key, val string) (bool, error) {
			return strings.ContainsAny(val, "aeiou"), nil
		})
		assert.NoError(nt, mappedErr)
		assert.Equal(nt, mapped, collection)
	})
	t.Run("should return correct values when iterator returns error", func(nt *testing.T) {
		collection := map[string]string{"1": "the brown", "2": "fly", "3": "jumps over the", "4": "brown fence"}
		mapped, mappedErr := goutils.FilterMap(collection, func(key, val string) (bool, error) {
			return strings.ContainsAny(val, "aeiou"), errors.New("an error")
		})
		assert.Error(nt, mappedErr)
		assert.Nil(nt, mapped)
	})
}

func TestGroupByMap(t *testing.T) {
	t.Run("should return correct values when iterator returns no error", func(nt *testing.T) {
		collection := map[string]string{"1": "the brown", "2": "fox", "3": "jumps over the", "4": "brown fence", "5": "fly"}
		collectionResult := map[int][]string{9: {"the brown"}, 3: {"fox", "fly"}, 14: {"jumps over the"}, 11: {"brown fence"}}
		grouped, groupedErr := goutils.GroupByMap(collection, func(key, val string) (int, string, error) {
			return len(val), val, nil
		})
		assert.NoError(nt, groupedErr)
		assert.Len(nt, grouped, len(collectionResult))
		for key, val := range grouped {
			assert.ElementsMatch(nt, val, collectionResult[key])
		}
	})
	t.Run("should return correct values when iterator returns error", func(nt *testing.T) {
		collection := map[string]string{"1": "the brown", "2": "fox", "3": "jumps over the", "4": "brown fence"}
		grouped, groupedErr := goutils.GroupByMap(collection, func(key, val string) (int, string, error) {
			return len(val), val, errors.New("an error")
		})
		assert.Error(nt, groupedErr)
		assert.Nil(nt, grouped)
	})
}

func TestRejectMap(t *testing.T) {
	t.Run("should return correct values when iterator returns no error", func(nt *testing.T) {
		collection := map[string]string{"1": "the brown", "2": "fox", "3": "jumps over the", "4": "brown fence"}
		mapped, mappedErr := goutils.RejectMap(collection, func(key, val string) (bool, error) {
			return !strings.ContainsAny(val, "aeiou"), nil
		})
		assert.NoError(nt, mappedErr)
		assert.Equal(nt, mapped, collection)
	})
	t.Run("should return correct values when iterator returns error", func(nt *testing.T) {
		collection := map[string]string{"1": "the brown", "2": "fly", "3": "jumps over the", "4": "brown fence"}
		mapped, mappedErr := goutils.RejectMap(collection, func(key, val string) (bool, error) {
			return !strings.ContainsAny(val, "aeiou"), errors.New("an error")
		})
		assert.Error(nt, mappedErr)
		assert.Nil(nt, mapped)
	})
}

func TestSomeMap(t *testing.T) {
	t.Run("should return correct values when iterator returns no error and atleast one value tests true", func(nt *testing.T) {
		collection := map[string]string{"1": "the brown", "2": "fly", "3": "jumps over the", "4": "brown fence"}
		mapped, mappedErr := goutils.SomeMap(collection, func(key, val string) (bool, error) {
			return strings.ContainsAny(val, "aeiou"), nil
		})
		assert.NoError(nt, mappedErr)
		assert.True(nt, mapped)
	})
	t.Run("should return correct values when iterator returns no error and no value tests true", func(nt *testing.T) {
		collection := map[string]string{"1": "", "2": "fly", "3": "2024"}
		mapped, mappedErr := goutils.SomeMap(collection, func(key, val string) (bool, error) {
			return strings.ContainsAny(val, "aeiou"), nil
		})
		assert.NoError(nt, mappedErr)
		assert.False(nt, mapped)
	})
	t.Run("should return correct values when iterator returns error", func(nt *testing.T) {
		collection := map[string]string{"1": "the brown", "2": "fly", "3": "jumps over the", "4": "brown fence"}
		mapped, mappedErr := goutils.SomeMap(collection, func(key, val string) (bool, error) {
			return !strings.ContainsAny(val, "aeiou"), errors.New("an error")
		})
		assert.Error(nt, mappedErr)
		assert.False(nt, mapped)
	})
}
