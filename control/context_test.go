package control_test

import (
	"context"
	"testing"

	"github.com/skatiyar/goutils/control"
	"github.com/stretchr/testify/assert"
)

func TestGetControlContextValue(t *testing.T) {
	t.Run("should return correct value", func(nt *testing.T) {
		ctx := context.WithValue(context.Background(), control.ContextKey("Hello"), "World")
		val, valErr := control.GetControlContextValue[string, string](ctx, "Hello")
		assert.NoError(nt, valErr)
		assert.Equal(nt, val, "World")
	})
	t.Run("should return error", func(nt *testing.T) {
		ctx := context.Background()
		val, valErr := control.GetControlContextValue[string, string](ctx, "Hello")
		assert.Error(nt, valErr)
		assert.Equal(nt, val, "")
	})
}

func TestSetControlContextValue(t *testing.T) {
	t.Run("should return correct value", func(nt *testing.T) {
		ctx := control.SetControlContextValue(context.Background(), "Hello", "World")
		val, ok := ctx.Value(control.ContextKey("Hello")).(string)
		assert.True(nt, ok)
		assert.Equal(nt, val, "World")
	})
}
