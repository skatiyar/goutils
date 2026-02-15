package control_test

import (
	"errors"
	"testing"
	"time"

	"github.com/skatiyar/goutils/control"
	"github.com/stretchr/testify/assert"
)

type TypeA struct{ Value string }
type TypeB struct{ Value string }
type TypeC struct{ Value string }
type TypeD struct{ Value string }
type TypeE struct{ Value string }
type TypeF struct{ Value string }
type TypeSeed struct{ Value string }

func TestNewAuto(t *testing.T) {
	t.Run("should return error when no functions provided", func(nt *testing.T) {
		_, err := control.NewAuto[TypeA]()
		assert.Error(nt, err)
	})
	t.Run("should return error for nil function", func(nt *testing.T) {
		_, err := control.NewAuto[TypeA](nil)
		assert.Error(nt, err)
	})
	t.Run("should return error for non-function argument", func(nt *testing.T) {
		_, err := control.NewAuto[TypeA]("not a function")
		assert.Error(nt, err)
	})
	t.Run("should return error when function returns less than 2 values", func(nt *testing.T) {
		_, err := control.NewAuto[TypeA](func() error { return nil })
		assert.Error(nt, err)
	})
	t.Run("should return error when last return is not error", func(nt *testing.T) {
		_, err := control.NewAuto[TypeA](func() (TypeA, string) { return TypeA{}, "" })
		assert.Error(nt, err)
	})
	t.Run("should return error when return value is not a struct", func(nt *testing.T) {
		_, err := control.NewAuto[TypeA](func() (int, error) { return 0, nil })
		assert.Error(nt, err)
	})
	t.Run("should return error when input is not a struct", func(nt *testing.T) {
		_, err := control.NewAuto[TypeA](func(s string) (TypeA, error) { return TypeA{}, nil })
		assert.Error(nt, err)
	})
	t.Run("should return error when duplicate output types exist", func(nt *testing.T) {
		_, err := control.NewAuto[TypeA](
			func() (TypeA, error) { return TypeA{}, nil },
			func() (TypeA, error) { return TypeA{}, nil },
		)
		assert.Error(nt, err)
	})
	t.Run("should return error when no function produces result type T", func(nt *testing.T) {
		_, err := control.NewAuto[TypeE](
			func() (TypeA, error) { return TypeA{}, nil },
		)
		assert.Error(nt, err)
	})
	t.Run("should return error for cyclic dependency", func(nt *testing.T) {
		_, err := control.NewAuto[TypeA](
			func(b TypeB) (TypeA, error) { return TypeA{}, nil },
			func(a TypeA) (TypeB, error) { return TypeB{}, nil },
		)
		assert.Error(nt, err)
	})
	t.Run("should succeed with valid single root function", func(nt *testing.T) {
		auto, err := control.NewAuto[TypeA](
			func() (TypeA, error) { return TypeA{Value: "hello"}, nil },
		)
		assert.NoError(nt, err)
		assert.NotEmpty(nt, auto.Describe())
	})
	t.Run("should succeed with multi-output function", func(nt *testing.T) {
		_, err := control.NewAuto[TypeC](
			func() (TypeA, TypeB, error) { return TypeA{}, TypeB{}, nil },
			func(a TypeA, b TypeB) (TypeC, error) { return TypeC{}, nil },
		)
		assert.NoError(nt, err)
	})
	t.Run("should return error for duplicate output across multi-output", func(nt *testing.T) {
		_, err := control.NewAuto[TypeB](
			func() (TypeA, TypeB, error) { return TypeA{}, TypeB{}, nil },
			func() (TypeB, error) { return TypeB{}, nil },
		)
		assert.Error(nt, err)
	})
}

func TestAutoRun(t *testing.T) {
	t.Run("should execute single root function", func(nt *testing.T) {
		auto, err := control.NewAuto[TypeA](
			func() (TypeA, error) { return TypeA{Value: "hello"}, nil },
		)
		assert.NoError(nt, err)
		result, runErr := auto.Run()
		assert.NoError(nt, runErr)
		assert.Equal(nt, "hello", result.Value)
	})
	t.Run("should execute linear chain", func(nt *testing.T) {
		auto, err := control.NewAuto[TypeC](
			func() (TypeA, error) { return TypeA{Value: "a"}, nil },
			func(a TypeA) (TypeB, error) { return TypeB{Value: a.Value + "b"}, nil },
			func(b TypeB) (TypeC, error) { return TypeC{Value: b.Value + "c"}, nil },
		)
		assert.NoError(nt, err)
		result, runErr := auto.Run()
		assert.NoError(nt, runErr)
		assert.Equal(nt, "abc", result.Value)
	})
	t.Run("should execute diamond dependency", func(nt *testing.T) {
		auto, err := control.NewAuto[TypeD](
			func() (TypeA, error) { return TypeA{Value: "root"}, nil },
			func(a TypeA) (TypeB, error) { return TypeB{Value: a.Value + "-B"}, nil },
			func(a TypeA) (TypeC, error) { return TypeC{Value: a.Value + "-C"}, nil },
			func(b TypeB, c TypeC) (TypeD, error) {
				return TypeD{Value: b.Value + "+" + c.Value}, nil
			},
		)
		assert.NoError(nt, err)
		result, runErr := auto.Run()
		assert.NoError(nt, runErr)
		assert.Equal(nt, "root-B+root-C", result.Value)
	})
	t.Run("should execute parallel roots", func(nt *testing.T) {
		auto, err := control.NewAuto[TypeC](
			func() (TypeA, error) { return TypeA{Value: "a"}, nil },
			func() (TypeB, error) { return TypeB{Value: "b"}, nil },
			func(a TypeA, b TypeB) (TypeC, error) {
				return TypeC{Value: a.Value + b.Value}, nil
			},
		)
		assert.NoError(nt, err)
		result, runErr := auto.Run()
		assert.NoError(nt, runErr)
		assert.Equal(nt, "ab", result.Value)
	})
	t.Run("should execute with multi-output function", func(nt *testing.T) {
		auto, err := control.NewAuto[TypeC](
			func() (TypeA, TypeB, error) {
				return TypeA{Value: "a"}, TypeB{Value: "b"}, nil
			},
			func(a TypeA, b TypeB) (TypeC, error) {
				return TypeC{Value: a.Value + b.Value}, nil
			},
		)
		assert.NoError(nt, err)
		result, runErr := auto.Run()
		assert.NoError(nt, runErr)
		assert.Equal(nt, "ab", result.Value)
	})
	t.Run("should execute with seed values", func(nt *testing.T) {
		auto, err := control.NewAuto[TypeB](
			func(s TypeSeed) (TypeA, error) {
				return TypeA{Value: s.Value + "-processed"}, nil
			},
			func(a TypeA) (TypeB, error) {
				return TypeB{Value: a.Value + "-done"}, nil
			},
		)
		assert.NoError(nt, err)
		result, runErr := auto.Run(TypeSeed{Value: "seed"})
		assert.NoError(nt, runErr)
		assert.Equal(nt, "seed-processed-done", result.Value)
	})
	t.Run("should fail fast on error", func(nt *testing.T) {
		auto, err := control.NewAuto[TypeB](
			func() (TypeA, error) { return TypeA{}, errors.New("intentional error") },
			func(a TypeA) (TypeB, error) { return TypeB{}, nil },
		)
		assert.NoError(nt, err)
		_, runErr := auto.Run()
		assert.Error(nt, runErr)
		assert.Equal(nt, "intentional error", runErr.Error())
	})
	t.Run("should recover from panic in goroutine", func(nt *testing.T) {
		auto, err := control.NewAuto[TypeB](
			func() (TypeA, error) { panic("test panic") },
			func(a TypeA) (TypeB, error) { return TypeB{}, nil },
		)
		assert.NoError(nt, err)
		_, runErr := auto.Run()
		assert.Error(nt, runErr)
	})
	t.Run("should return error for missing seed", func(nt *testing.T) {
		auto, err := control.NewAuto[TypeB](
			func(s TypeSeed) (TypeA, error) { return TypeA{}, nil },
			func(a TypeA) (TypeB, error) { return TypeB{}, nil },
		)
		assert.NoError(nt, err)
		_, runErr := auto.Run()
		assert.Error(nt, runErr)
	})
	t.Run("should return error for non-struct seed", func(nt *testing.T) {
		auto, err := control.NewAuto[TypeA](
			func() (TypeA, error) { return TypeA{}, nil },
		)
		assert.NoError(nt, err)
		_, runErr := auto.Run("not a struct")
		assert.Error(nt, runErr)
	})
	t.Run("should return error for duplicate seed type", func(nt *testing.T) {
		auto, err := control.NewAuto[TypeB](
			func(s TypeSeed) (TypeA, error) { return TypeA{}, nil },
			func(a TypeA) (TypeB, error) { return TypeB{}, nil },
		)
		assert.NoError(nt, err)
		_, runErr := auto.Run(TypeSeed{Value: "a"}, TypeSeed{Value: "b"})
		assert.Error(nt, runErr)
	})
	t.Run("should return error when seed conflicts with function output", func(nt *testing.T) {
		auto, err := control.NewAuto[TypeB](
			func() (TypeA, error) { return TypeA{}, nil },
			func(a TypeA) (TypeB, error) { return TypeB{}, nil },
		)
		assert.NoError(nt, err)
		_, runErr := auto.Run(TypeA{Value: "conflict"})
		assert.Error(nt, runErr)
	})
	t.Run("should return error for nil seed arg", func(nt *testing.T) {
		auto, err := control.NewAuto[TypeA](
			func() (TypeA, error) { return TypeA{}, nil },
		)
		assert.NoError(nt, err)
		_, runErr := auto.Run(nil)
		assert.Error(nt, runErr)
	})
	t.Run("should return zero value when running empty auto", func(nt *testing.T) {
		auto := control.Auto[TypeA]{}
		result, runErr := auto.Run()
		assert.NoError(nt, runErr)
		assert.Equal(nt, TypeA{}, result)
	})
	t.Run("should recover from panic with error value", func(nt *testing.T) {
		auto, err := control.NewAuto[TypeB](
			func() (TypeA, error) { panic(errors.New("error panic")) },
			func(a TypeA) (TypeB, error) { return TypeB{}, nil },
		)
		assert.NoError(nt, err)
		_, runErr := auto.Run()
		assert.Error(nt, runErr)
		assert.Equal(nt, "error panic", runErr.Error())
	})
	t.Run("should fail fast and drain parallel functions on error", func(nt *testing.T) {
		auto, err := control.NewAuto[TypeC](
			func() (TypeA, error) {
				return TypeA{}, errors.New("early failure")
			},
			func() (TypeB, error) {
				time.Sleep(20 * time.Millisecond)
				return TypeB{}, nil
			},
			func(a TypeA, b TypeB) (TypeC, error) {
				return TypeC{}, nil
			},
		)
		assert.NoError(nt, err)
		_, runErr := auto.Run()
		assert.Error(nt, runErr)
		assert.Equal(nt, "early failure", runErr.Error())
	})
	t.Run("should run functions in parallel when possible", func(nt *testing.T) {
		auto, err := control.NewAuto[TypeC](
			func() (TypeA, error) {
				time.Sleep(50 * time.Millisecond)
				return TypeA{Value: "a"}, nil
			},
			func() (TypeB, error) {
				time.Sleep(50 * time.Millisecond)
				return TypeB{Value: "b"}, nil
			},
			func(a TypeA, b TypeB) (TypeC, error) {
				return TypeC{Value: a.Value + b.Value}, nil
			},
		)
		assert.NoError(nt, err)
		start := time.Now()
		result, runErr := auto.Run()
		elapsed := time.Since(start)
		assert.NoError(nt, runErr)
		assert.Equal(nt, "ab", result.Value)
		// If parallel, should take ~50ms, not ~100ms
		assert.Less(nt, elapsed, 90*time.Millisecond)
	})
}

func TestAutoRunWithMetrics(t *testing.T) {
	t.Run("should return metrics with duration and func metrics", func(nt *testing.T) {
		auto, err := control.NewAuto[TypeB](
			func() (TypeA, error) {
				time.Sleep(20 * time.Millisecond)
				return TypeA{Value: "a"}, nil
			},
			func(a TypeA) (TypeB, error) {
				time.Sleep(20 * time.Millisecond)
				return TypeB{Value: a.Value + "b"}, nil
			},
		)
		assert.NoError(nt, err)
		result, metrics, runErr := auto.RunWithMetrics()
		assert.NoError(nt, runErr)
		assert.Equal(nt, "ab", result.Value)
		assert.GreaterOrEqual(nt, metrics.Duration, 40*time.Millisecond)
		assert.Len(nt, metrics.FuncMetrics, 2)
		for _, dur := range metrics.FuncMetrics {
			assert.GreaterOrEqual(nt, dur, 20*time.Millisecond)
		}
	})
}

func TestAutoDescribe(t *testing.T) {
	t.Run("should return empty for no nodes", func(nt *testing.T) {
		auto := control.Auto[TypeA]{}
		assert.Empty(nt, auto.Describe())
	})
	t.Run("should describe dependency graph", func(nt *testing.T) {
		auto, err := control.NewAuto[TypeC](
			func() (TypeA, error) { return TypeA{}, nil },
			func(a TypeA) (TypeB, error) { return TypeB{}, nil },
			func(b TypeB) (TypeC, error) { return TypeC{}, nil },
		)
		assert.NoError(nt, err)
		desc := auto.Describe()
		assert.Len(nt, desc, 3)
	})
	t.Run("should describe external seed dependencies", func(nt *testing.T) {
		auto, err := control.NewAuto[TypeB](
			func(s TypeSeed) (TypeA, error) { return TypeA{}, nil },
			func(a TypeA) (TypeB, error) { return TypeB{}, nil },
		)
		assert.NoError(nt, err)
		desc := auto.Describe()
		assert.Len(nt, desc, 2)
		assert.Contains(nt, desc[0], "seed(")
	})
}
