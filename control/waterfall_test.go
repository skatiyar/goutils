package control_test

import (
	"context"
	"errors"
	"testing"

	"github.com/skatiyar/goutils/control"
	"github.com/stretchr/testify/assert"
)

func TestWaterfall(t *testing.T) {
	t.Run("should return correct values when no error is returned", func(nt *testing.T) {
		fctx, fctxErr := control.Waterfall(
			control.WaterfallBaseValue("First", "Hello"),
			func(ctx context.Context) (context.Context, error) {
				data := make(map[string]string)
				val, valErr := control.GetControlContextValue[string, string](ctx, "First")
				if valErr != nil {
					return ctx, valErr
				} else {
					data["First"] = val
					data["Second"] = "World"
				}
				return control.SetControlContextValue(ctx, "Data", data), nil
			},
		)
		value, valueErr := control.GetControlContextValue[string, map[string]string](fctx, "Data")
		assert.NoError(nt, fctxErr)
		assert.NoError(nt, valueErr)
		assert.Equal(nt, value, map[string]string{"First": "Hello", "Second": "World"})
	})
	t.Run("should return correct values when error is returned", func(nt *testing.T) {
		fctx, fctxErr := control.Waterfall(
			control.WaterfallBaseValue("First", "Hello"),
			func(ctx context.Context) (context.Context, error) {
				return ctx, errors.New("some error")
			},
		)
		assert.Error(nt, fctxErr)
		value, valueErr := control.GetControlContextValue[string, string](fctx, "First")
		assert.NoError(nt, valueErr)
		assert.Equal(nt, value, "Hello")
	})
}
