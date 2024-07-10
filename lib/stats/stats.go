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
					"euclidean_distance": newBinaryFloatBuiltin("euclidean_distance", gms.EuclideanDistance),
					"manhattan_distance": newBinaryFloatBuiltin("manhattan_distance", gms.ManhattanDistance),
					"softmax":            newUnaryFloatListBuiltin("softmax", gms.SoftMax),
					"sigmoid":            newUnaryFloatListBuiltin("sigmoid", gms.Sigmoid),

					"mode": newUnaryFloatListBuiltin("mode", gms.Mode),
					"sum":  newUnaryFloatBuiltin("sum", gms.Sum),
					"max":  newUnaryFloatBuiltin("max", gms.Max),
					"min":  newUnaryFloatBuiltin("min", gms.Min),

					"mean":           newUnaryFloatBuiltin("mean", gms.Mean),
					"norm_mean":      newBinaryFloatToFloatBuiltin("norm_mean", gms.NormMean),
					"geometric_mean": newUnaryFloatBuiltin("geometric_mean", gms.GeometricMean),
					"harmonic_mean":  newUnaryFloatBuiltin("harmonic_mean", gms.HarmonicMean),
					"trimean":        newUnaryFloatBuiltin("trimean", gms.Trimean),
					"median":         newUnaryFloatBuiltin("median", gms.Median),
					"norm_median":    newBinaryFloatToFloatBuiltin("norm_median", gms.NormMedian),

					"percentile":              newBinaryFloatSingleBuiltin("percentile", gms.Percentile),
					"percentile_nearest_rank": newBinaryFloatSingleBuiltin("percentile_nearest_rank", gms.PercentileNearestRank),

					"variance":              newUnaryFloatBuiltin("variance", gms.Variance),
					"covariance":            newBinaryFloatBuiltin("covariance", gms.Covariance),
					"covariance_population": newBinaryFloatBuiltin("covariance_population", gms.CovariancePopulation),
					"population_variance":   newUnaryFloatBuiltin("population_variance", gms.PopulationVariance),
					"sample_variance":       newUnaryFloatBuiltin("sample_variance", gms.SampleVariance),
					"sample":                starlark.NewBuiltin("sample", sample),

					"correlation": newBinaryFloatBuiltin("correlation", gms.Correlation),
					"pearson":     newBinaryFloatBuiltin("pearson", gms.Pearson),

					"standard_deviation": newUnaryFloatBuiltin("standard_deviation", gms.StandardDeviation),
					"stddev":             newUnaryFloatBuiltin("stddev", gms.StandardDeviation),
					"norm_stddev":        newBinaryFloatToFloatBuiltin("norm_stddev", gms.NormStd),
					"stddev_sample":      newUnaryFloatBuiltin("stddev_sample", gms.StandardDeviationSample),
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
		return float64List(result), nil
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

// newBinaryFloatSingleBuiltin wraps a binary function accepting a []float64 and a float64 arguments and returning (float64, error) as a Starlark built-in.
func newBinaryFloatSingleBuiltin(name string, fn func(gms.Float64Data, float64) (float64, error)) *starlark.Builtin {
	return starlark.NewBuiltin(name, func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		var (
			data1 tps.FloatOrIntList
			data2 float64
		)
		if err := starlark.UnpackPositionalArgs(name, args, kwargs, 2, &data1, &data2); err != nil {
			return nil, err
		}
		result, err := fn(data1.GoSlice(), data2)
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

// sample returns sample from input with replacement or without.
func sample(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var (
		data tps.FloatOrIntList
		take int
		repl bool
	)
	if err := starlark.UnpackArgs(b.Name(), args, kwargs, "data", &data, "take", &take, "replace", &repl); err != nil {
		return nil, err
	}
	result, err := gms.Sample(data.GoSlice(), take, repl)
	if err != nil {
		return nil, err
	}
	return float64List(result), nil
}

func float64List(data []float64) starlark.Value {
	fls := make([]starlark.Value, len(data))
	for i, v := range data {
		fls[i] = starlark.Float(v)
	}
	return starlark.NewList(fls)
}
