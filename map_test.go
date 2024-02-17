package goutils_test

import (
	"strings"
	"testing"

	"github.com/skatiyar/goutils"
	"github.com/stretchr/testify/assert"
)

func TestMap(t *testing.T) {
	t.Run("should return correct values", func(nt *testing.T) {
		collection := map[string]string{"1": "the brown", "2": "fox", "3": "jumps over the", "4": "brown fence"}
		collectionResult := map[string]string{"1": "brown", "2": "fox", "3": "jumps over", "4": "brown fence"}
		assert.Equal(nt, goutils.Map(collection, func(key, val string) (string, string) {
			return key, strings.Trim(strings.ReplaceAll(val, "the", ""), " ")
		}), collectionResult)
	})
}
