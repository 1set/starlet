// Package stats provides a Starlark module for comprehensive statistics functions. It's a wrapper around the Go package: https://github.com/montanaflynn/stats
package stats

import (
	"sync"

	tps "github.com/1set/starlet/dataconv/types"
	gms "github.com/montanaflynn/stats"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

// ModuleName defines the expected name for this Module when used
// in starlark's load() function, eg: load('stats', 'md5')
const ModuleName = "stats"

var (
	once       sync.Once
	statModule starlark.StringDict
)

// LoadModule loads the hashlib module. It is concurrency-safe and idempotent.
func LoadModule() (starlark.StringDict, error) {
	once.Do(func() {
		statModule = starlark.StringDict{
			ModuleName: &starlarkstruct.Module{
				Name: ModuleName,
				Members: starlark.StringDict{
					"mean": newUnaryFloatBuiltin("mean", gms.Mean),
				},
			},
		}
	})
	return statModule, nil
}

// newUnaryFloatBuiltin wraps a unary function accepting []float64 and returning (float64, error) as a Starlark built-in.
func newUnaryFloatBuiltin(name string, fn func(data gms.Float64Data) (float64, error)) *starlark.Builtin {
	return starlark.NewBuiltin(name, func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		var data tps.FloatOrIntList
		if err := starlark.UnpackPositionalArgs(name, args, kwargs, 1, &data); err != nil {
			return nil, err
		}
		result, err := fn(data.GoSlice())
		if err != nil {
			return nil, err
		}
		return starlark.Float(result), nil
	})
}

// newUnaryFloatListBuiltin wraps a unary function accepting []float64 and returning ([]float64, error) as a Starlark built-in.
func newUnaryFloatListBuiltin(name string, fn func(gms.Float64Data) ([]float64, error)) *starlark.Builtin {
	return starlark.NewBuiltin(name, func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		var data tps.FloatOrIntList
		if err := starlark.UnpackPositionalArgs(name, args, kwargs, 1, &data); err != nil {
			return nil, err
		}
		result, err := fn(data.GoSlice())
		if err != nil {
			return nil, err
		}
		fls := make([]starlark.Value, len(result))
		for i, v := range result {
			fls[i] = starlark.Float(v)
		}
		return starlark.NewList(fls), nil
	})
}

// newBinaryFloatBuiltin wraps a binary function accepting two []float64 arguments and returning (float64, error) as a Starlark built-in.
func newBinaryFloatBuiltin(name string, fn func(gms.Float64Data, gms.Float64Data) (float64, error)) *starlark.Builtin {
	return starlark.NewBuiltin(name, func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		var data1, data2 tps.FloatOrIntList
		if err := starlark.UnpackPositionalArgs(name, args, kwargs, 2, &data1, &data2); err != nil {
			return nil, err
		}
		result, err := fn(data1.GoSlice(), data2.GoSlice())
		if err != nil {
			return nil, err
		}
		return starlark.Float(result), nil
	})
}

// newBinaryFloatToFloatBuiltin wraps a binary function accepting two float64 arguments and returning float64 as a Starlark built-in.
func newBinaryFloatToFloatBuiltin(name string, fn func(float64, float64) float64) *starlark.Builtin {
	return starlark.NewBuiltin(name, func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		var x, y tps.FloatOrInt
		if err := starlark.UnpackPositionalArgs(name, args, kwargs, 2, &x, &y); err != nil {
			return nil, err
		}
		result := fn(float64(x), float64(y))
		return starlark.Float(result), nil
	})
}

// newTernaryFloatToFloatBuiltin wraps a ternary function accepting three float64 arguments and returning float64 as a Starlark built-in.
func newTernaryFloatToFloatBuiltin(name string, fn func(float64, float64, float64) float64) *starlark.Builtin {
	return starlark.NewBuiltin(name, func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		var x, y, z tps.FloatOrInt
		if err := starlark.UnpackPositionalArgs(name, args, kwargs, 3, &x, &y, &z); err != nil {
			return nil, err
		}
		result := fn(float64(x), float64(y), float64(z))
		return starlark.Float(result), nil
	})
}
