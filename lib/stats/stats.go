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
					// Distance Metrics
					"euclidean_distance": newBinaryFloatBuiltin("euclidean_distance", gms.EuclideanDistance), // Calculates the straight line distance between two points in Euclidean space.
					"manhattan_distance": newBinaryFloatBuiltin("manhattan_distance", gms.ManhattanDistance), // Computes the sum of the absolute differences of their coordinates.

					// Probability Distributions and Transformations
					"softmax": newUnaryFloatListBuiltin("softmax", gms.SoftMax), // Applies the SoftMax function, useful for converting scores to probabilities.
					"sigmoid": newUnaryFloatListBuiltin("sigmoid", gms.Sigmoid), // Applies the Sigmoid function, often used for binary classification.

					// Basic Statistical Measures
					"mode":     newUnaryFloatListBuiltin("mode", gms.Mode), // Determines the most frequently occurring data points in a dataset.
					"sum":      newUnaryFloatBuiltin("sum", gms.Sum),       // Computes the sum of a series of numbers.
					"max":      newUnaryFloatBuiltin("max", gms.Max),       // Finds the maximum value in a dataset.
					"min":      newUnaryFloatBuiltin("min", gms.Min),       // Finds the minimum value in a dataset.
					"midrange": starlark.NewBuiltin("midrange", midrange),  // Calculates the midrange, the average of the maximum and minimum values.

					// Measures of Central Tendency
					"average":        newUnaryFloatBuiltin("average", gms.Mean),                 // Alias for mean.
					"mean":           newUnaryFloatBuiltin("mean", gms.Mean),                    // Calculates the arithmetic mean of a dataset.
					"geometric_mean": newUnaryFloatBuiltin("geometric_mean", gms.GeometricMean), // Computes the geometric mean, useful for datasets with exponential growth.
					"harmonic_mean":  newUnaryFloatBuiltin("harmonic_mean", gms.HarmonicMean),   // Computes the harmonic mean, effective for rates and ratios.
					"trimean":        newUnaryFloatBuiltin("trimean", gms.Trimean),              // Calculates the trimean, a measure of a dataset's tendency.
					"median":         newUnaryFloatBuiltin("median", gms.Median),                // Finds the middle value of a dataset.

					// Measures of Variability
					"percentile":              newBinaryFloatSingleBuiltin("percentile", gms.Percentile),                         // Determines the value below which a given percentage of observations in a group of observations falls.
					"percentile_nearest_rank": newBinaryFloatSingleBuiltin("percentile_nearest_rank", gms.PercentileNearestRank), // Similar to percentile but uses the nearest rank method.
					"variance":                newUnaryFloatBuiltin("variance", gms.Variance),                                    // Calculates the variance of a dataset, a measure of dispersion.
					"covariance":              newBinaryFloatBuiltin("covariance", gms.Covariance),                               // Measures how changes in one variable are associated with changes in another variable.
					"covariance_population":   newBinaryFloatBuiltin("covariance_population", gms.CovariancePopulation),          // Calculates the covariance for an entire population.
					"population_variance":     newUnaryFloatBuiltin("population_variance", gms.PopulationVariance),               // Computes the variance of an entire population.
					"sample_variance":         newUnaryFloatBuiltin("sample_variance", gms.SampleVariance),                       // Computes the variance of a sample from the population.
					"sample":                  starlark.NewBuiltin("sample", sample),                                             // Randomly samples elements from the dataset.

					// Correlation and Regression
					"correlation": newBinaryFloatBuiltin("correlation", gms.Correlation), // Computes the correlation coefficient between two datasets.
					"pearson":     newBinaryFloatBuiltin("pearson", gms.Pearson),         // Computes Pearson's correlation coefficient specifically.

					// Standard Deviation
					"standard_deviation": newUnaryFloatBuiltin("standard_deviation", gms.StandardDeviation),  // Calculates the standard deviation of a dataset.
					"stddev":             newUnaryFloatBuiltin("stddev", gms.StandardDeviation),              // Alias for standard_deviation.
					"stddev_sample":      newUnaryFloatBuiltin("stddev_sample", gms.StandardDeviationSample), // Calculates the standard deviation of a sample.
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
			data tps.FloatOrIntList
			val  tps.FloatOrInt
		)
		if err := starlark.UnpackPositionalArgs(name, args, kwargs, 2, &data, &val); err != nil {
			return nil, err
		}
		result, err := fn(data.GoSlice(), val.GoFloat())
		if err != nil {
			return nil, err
		}
		return starlark.Float(result), nil
	})
}

// sample returns sample from input with replacement or without.
func sample(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	// check input
	var (
		data tps.FloatOrIntList
		take int
		repl bool
	)
	if err := starlark.UnpackArgs(b.Name(), args, kwargs, "data", &data, "take", &take, "replace?", &repl); err != nil {
		return nil, err
	}
	// call and convert result
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

// midrange calculates the midrange, the average of the maximum and minimum values.
func midrange(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	// check input
	var data tps.FloatOrIntList
	if err := starlark.UnpackPositionalArgs(b.Name(), args, kwargs, 1, &data); err != nil {
		return nil, err
	}
	ds := data.GoSlice()

	// check if data is empty
	if len(ds) == 0 {
		return nil, gms.EmptyInputErr
	}

	// calculate midrange
	var (
		min float64 = ds[0]
		max float64 = ds[0]
	)
	for _, v := range ds {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}
	mr := (min + max) / 2

	// return result
	return starlark.Float(mr), nil
}
