package goutils_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/skatiyar/goutils"
	"github.com/stretchr/testify/assert"
)

func TestConcatSlice(t *testing.T) {
	t.Run("should return correct values when error is nil", func(nt *testing.T) {
		collection := []string{"the brown", "fox", "jumps over the", "brown fence"}
		collectionResult := []string{"the", "brown", "fox", "jumps", "over", "the", "brown", "fence"}
		result, resultErr := goutils.ConcatSlice(collection, func(val string, idx int) ([]string, error) {
			return strings.Split(val, " "), nil
		})
		assert.NoError(nt, resultErr)
		assert.Equal(nt, result, collectionResult)
	})
	t.Run("should return correct values when error is not nil", func(nt *testing.T) {
		collection := []string{"the brown", "fox", "jumps over the", "brown fence"}
		result, resultErr := goutils.ConcatSlice(collection, func(val string, idx int) ([]string, error) {
			return strings.Split(val, " "), errors.New("an error")
		})
		assert.Error(nt, resultErr)
		assert.Equal(nt, result, []string(nil))
	})
}

func TestDetectSlice(t *testing.T) {
	t.Run("should return correct value when detected but error is nil", func(nt *testing.T) {
		collection := []string{"the brown", "fox", "jumps over the", "brown fence"}
		value, detected, err := goutils.DetectSlice(collection, func(val string, idx int) (bool, error) {
			return strings.Contains(val, "fox"), nil
		})
		assert.NoError(nt, err)
		assert.Truef(nt, detected, "Value found", value)
	})
	t.Run("should return correct value when not detected but error is nil", func(nt *testing.T) {
		collection := []string{"the brown", "fox", "jumps over the", "brown fence"}
		value, detected, err := goutils.DetectSlice(collection, func(val string, idx int) (bool, error) {
			return strings.Contains(val, "dog"), nil
		})
		assert.NoError(nt, err)
		assert.Falsef(nt, detected, "Value not found", value)
	})
	t.Run("should return correct value when error is not nil", func(nt *testing.T) {
		collection := []string{"the brown", "fox", "jumps over the", "brown fence"}
		value, detected, err := goutils.DetectSlice(collection, func(val string, idx int) (bool, error) {
			return strings.Contains(val, "fox"), errors.New("an error")
		})
		assert.Error(nt, err)
		assert.Falsef(nt, detected, "Value not found", value)
	})
}

func TestEachSlice(t *testing.T) {
	t.Run("should pass when iterator returns no error", func(nt *testing.T) {
		collection := []string{"the brown", "fox", "jumps over the", "brown fence"}
		collectionResult := []string{"brown", "fox", "jumps over", "brown fence"}
		result := make([]string, 0)
		err := goutils.EachSlice(collection, func(val string, idx int) error {
			result = append(result, strings.Trim(strings.ReplaceAll(val, "the", ""), " "))
			return nil
		})
		assert.NoError(nt, err)
		assert.Equal(nt, result, collectionResult)
	})
	t.Run("should pass when iterator returns error", func(nt *testing.T) {
		collection := []string{"the brown", "fox", "jumps over the", "brown fence"}
		result := make([]string, 0)
		err := goutils.EachSlice(collection, func(val string, idx int) error {
			result = append(result, strings.Trim(strings.ReplaceAll(val, "the", ""), " "))
			return errors.New("an error")
		})
		assert.Error(nt, err)
		assert.NotEmpty(nt, result)
	})
}

func TestSlice(t *testing.T) {
	t.Run("should return correct values when iterator returns no error", func(nt *testing.T) {
		collection := []string{"the brown", "fox", "jumps over the", "brown fence"}
		collectionResult := []string{"brown", "fox", "jumps over", "brown fence"}
		mapped, mappedErr := goutils.Slice(collection, func(val string, idx int) (string, error) {
			return strings.Trim(strings.ReplaceAll(val, "the", ""), " "), nil
		})
		assert.NoError(nt, mappedErr)
		assert.Equal(nt, mapped, collectionResult)
	})
	t.Run("should return correct values when iterator returns error", func(nt *testing.T) {
		collection := []string{"the brown", "fox", "jumps over the", "brown fence"}
		mapped, mappedErr := goutils.Slice(collection, func(val string, idx int) (string, error) {
			return strings.Trim(strings.ReplaceAll(val, "the", ""), " "), errors.New("an error")
		})
		assert.Error(nt, mappedErr)
		assert.Nil(nt, mapped)
	})
}

func TestReduceSlice(t *testing.T) {
	t.Run("should return correct values when iterator returns no error", func(nt *testing.T) {
		collection := []string{"the brown", "fox", "jumps over the", "brown fence"}
		collectionResult := "the brown fox jumps over the brown fence"
		mapped, mappedErr := goutils.ReduceSlice(collection, func(acc string, val string, idx int) (string, error) {
			if len(acc) > 0 {
				return acc + " " + val, nil
			} else {
				return val, nil
			}
		}, "")
		assert.NoError(nt, mappedErr)
		assert.Equal(nt, mapped, collectionResult)
	})
	t.Run("should return correct values when iterator returns error", func(nt *testing.T) {
		collection := []string{"the brown", "fox", "jumps over the", "brown fence"}
		mapped, mappedErr := goutils.ReduceSlice(collection, func(acc string, val string, idx int) (string, error) {
			return acc + " " + val, errors.New("an error")
		}, "")
		assert.Error(nt, mappedErr)
		assert.Empty(nt, mapped)
	})
}

func TestReduceRightSlice(t *testing.T) {
	t.Run("should return correct values when iterator returns no error", func(nt *testing.T) {
		collection := []string{"the brown", "fox", "jumps over the", "brown fence"}
		collectionResult := "brown fence jumps over the fox the brown"
		mapped, mappedErr := goutils.ReduceRightSlice(collection, func(acc string, val string, idx int) (string, error) {
			if len(acc) > 0 {
				return acc + " " + val, nil
			} else {
				return val, nil
			}
		}, "")
		assert.NoError(nt, mappedErr)
		assert.Equal(nt, mapped, collectionResult)
	})
	t.Run("should return correct values when iterator returns error", func(nt *testing.T) {
		collection := []string{"the brown", "fox", "jumps over the", "brown fence"}
		mapped, mappedErr := goutils.ReduceRightSlice(collection, func(acc string, val string, idx int) (string, error) {
			return acc + " " + val, errors.New("an error")
		}, "")
		assert.Error(nt, mappedErr)
		assert.Empty(nt, mapped)
	})
}

func TestEverySlice(t *testing.T) {
	t.Run("should return correct values when iterator returns no error and all values pass test", func(nt *testing.T) {
		collection := []string{"the brown", "fox", "jumps over the", "brown fence"}
		mapped, mappedErr := goutils.EverySlice(collection, func(val string, idx int) (bool, error) {
			return strings.ContainsAny(val, "aeiou"), nil
		})
		assert.NoError(nt, mappedErr)
		assert.True(nt, mapped)
	})
	t.Run("should return correct values when iterator returns no error and one value returs false", func(nt *testing.T) {
		collection := []string{"the brown", "fly", "jumps over the", "brown fence"}
		mapped, mappedErr := goutils.EverySlice(collection, func(val string, idx int) (bool, error) {
			return strings.ContainsAny(val, "aeiou"), nil
		})
		assert.NoError(nt, mappedErr)
		assert.False(nt, mapped)
	})
	t.Run("should return correct values when iterator returns error", func(nt *testing.T) {
		collection := []string{"the brown", "fox", "jumps over the", "brown fence"}
		mapped, mappedErr := goutils.EverySlice(collection, func(val string, idx int) (bool, error) {
			return strings.ContainsAny(val, "aeiou"), errors.New("an error")
		})
		assert.Error(nt, mappedErr)
		assert.False(nt, mapped)
	})
}

func TestFilterSlice(t *testing.T) {
	t.Run("should return correct values when iterator returns no error", func(nt *testing.T) {
		collection := []string{"the brown", "fox", "jumps over the", "brown fence"}
		mapped, mappedErr := goutils.FilterSlice(collection, func(val string, idx int) (bool, error) {
			return strings.ContainsAny(val, "aeiou"), nil
		})
		assert.NoError(nt, mappedErr)
		assert.Equal(nt, mapped, collection)
	})
	t.Run("should return correct values when iterator returns error", func(nt *testing.T) {
		collection := []string{"the brown", "fox", "jumps over the", "brown fence"}
		mapped, mappedErr := goutils.FilterSlice(collection, func(val string, idx int) (bool, error) {
			return strings.ContainsAny(val, "aeiou"), errors.New("an error")
		})
		assert.Error(nt, mappedErr)
		assert.Nil(nt, mapped)
	})
}

func TestGroupBySlice(t *testing.T) {
	t.Run("should return correct values when iterator returns no error", func(nt *testing.T) {
		collection := []string{"the brown", "fox", "jumps over the", "brown fence"}
		collectionResult := map[int][]string{9: {"the brown"}, 3: {"fox"}, 14: {"jumps over the"}, 11: {"brown fence"}}
		grouped, groupedErr := goutils.GroupBySlice(collection, func(val string, idx int) (int, string, error) {
			return len(val), val, nil
		})
		assert.NoError(nt, groupedErr)
		assert.Equal(nt, grouped, collectionResult)
	})
	t.Run("should return correct values when iterator returns error", func(nt *testing.T) {
		collection := []string{"the brown", "fox", "jumps over the", "brown fence"}
		grouped, groupedErr := goutils.GroupBySlice(collection, func(val string, idx int) (int, string, error) {
			return len(val), val, errors.New("an error")
		})
		assert.Error(nt, groupedErr)
		assert.Nil(nt, grouped)
	})
}

func TestRejectSlice(t *testing.T) {
	t.Run("should return correct values when iterator returns no error", func(nt *testing.T) {
		collection := []string{"the brown", "fox", "jumps over the", "brown fence"}
		mapped, mappedErr := goutils.RejectSlice(collection, func(val string, idx int) (bool, error) {
			return !strings.ContainsAny(val, "aeiou"), nil
		})
		assert.NoError(nt, mappedErr)
		assert.Equal(nt, mapped, collection)
	})
	t.Run("should return correct values when iterator returns error", func(nt *testing.T) {
		collection := []string{"the brown", "fox", "jumps over the", "brown fence"}
		mapped, mappedErr := goutils.RejectSlice(collection, func(val string, idx int) (bool, error) {
			return !strings.ContainsAny(val, "aeiou"), errors.New("an error")
		})
		assert.Error(nt, mappedErr)
		assert.Nil(nt, mapped)
	})
}

func TestSomeSlice(t *testing.T) {
	t.Run("should return correct values when iterator returns no error and atleast one value tests true", func(nt *testing.T) {
		collection := []string{"the brown", "fly", "jumps over the", "brown fence"}
		mapped, mappedErr := goutils.SomeSlice(collection, func(val string, idx int) (bool, error) {
			return strings.ContainsAny(val, "aeiou"), nil
		})
		assert.NoError(nt, mappedErr)
		assert.True(nt, mapped)
	})
	t.Run("should return correct values when iterator returns no error and no value tests true", func(nt *testing.T) {
		collection := []string{"", "fly", "2024"}
		mapped, mappedErr := goutils.SomeSlice(collection, func(val string, idx int) (bool, error) {
			return strings.ContainsAny(val, "aeiou"), nil
		})
		assert.NoError(nt, mappedErr)
		assert.False(nt, mapped)
	})
	t.Run("should return correct values when iterator returns error", func(nt *testing.T) {
		collection := []string{"the brown", "fly", "jumps over the", "brown fence"}
		mapped, mappedErr := goutils.SomeSlice(collection, func(val string, idx int) (bool, error) {
			return !strings.ContainsAny(val, "aeiou"), errors.New("an error")
		})
		assert.Error(nt, mappedErr)
		assert.False(nt, mapped)
	})
}
