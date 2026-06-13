# stats

`stats` provides a comprehensive set of statistics functions for Starlark, a thin wrapper around the Go package [`github.com/montanaflynn/stats`](https://github.com/montanaflynn/stats). It is **pure** (no filesystem, network, process, or log side effects) — every function is a deterministic computation over its arguments, with the sole exception of `sample`, which draws on a random number generator.

Every data argument accepts any Starlark **iterable of `int` or `float`** (a `list`, `tuple`, …); ints are promoted to floats. Scalar results are returned as `float`, vector results as a `list` of `float`.

## Functions

Each function takes one or two data arguments (an iterable of numbers) unless noted. `data`, `data1`, `data2` are iterables of `int`/`float`; `p` is a number; `take` is an `int`; `replace` is a `bool`.

| function | description |
|----------|-------------|
| `euclidean_distance(data1, data2) -> float` | Straight-line (L2) distance between two points. |
| `manhattan_distance(data1, data2) -> float` | Sum of absolute coordinate differences (L1 distance). |
| `softmax(data) -> list` | Softmax transform; converts scores to probabilities summing to 1. |
| `sigmoid(data) -> list` | Element-wise sigmoid (logistic) transform. |
| `mode(data) -> list` | Most frequently occurring value(s); may return several. |
| `sum(data) -> float` | Sum of the values. |
| `max(data) -> float` | Maximum value. |
| `min(data) -> float` | Minimum value. |
| `midrange(data) -> float` | Average of the maximum and minimum values. |
| `average(data) -> float` | Arithmetic mean (alias of `mean`). |
| `mean(data) -> float` | Arithmetic mean. |
| `geometric_mean(data) -> float` | Geometric mean. |
| `harmonic_mean(data) -> float` | Harmonic mean. |
| `trimean(data) -> float` | Trimean, a robust measure of central tendency. |
| `median(data) -> float` | Middle value. |
| `percentile(data, p) -> float` | Value below which `p`% of observations fall (interpolated). |
| `percentile_nearest_rank(data, p) -> float` | Percentile via the nearest-rank method. |
| `variance(data) -> float` | Variance (population variance; alias of `population_variance`). |
| `population_variance(data) -> float` | Variance of an entire population. |
| `sample_variance(data) -> float` | Variance of a sample (Bessel-corrected, `n-1`). |
| `covariance(data1, data2) -> float` | Sample covariance between two datasets. |
| `covariance_population(data1, data2) -> float` | Population covariance between two datasets. |
| `correlation(data1, data2) -> float` | Correlation coefficient between two datasets. |
| `pearson(data1, data2) -> float` | Pearson product-moment correlation coefficient. |
| `standard_deviation(data) -> float` | Population standard deviation. |
| `stddev(data) -> float` | Alias of `standard_deviation`. |
| `stddev_sample(data) -> float` | Sample standard deviation (Bessel-corrected, `n-1`). |
| `sample(data, take, replace=False) -> list` | Randomly draw `take` elements, with or without replacement. |

There are no exported constants and no custom value types; inputs are plain Starlark iterables and outputs are plain `float`/`list` values.

## Details & examples

All functions delegate validation to the underlying engine. Common errors propagate verbatim:

- Empty input where a value is required → `Input must not be empty.`
- Mismatched lengths for a two-dataset function → `Must be the same length.`
- A percentile / sample size outside the valid range → `Input is outside of range.`
- A non-iterable data argument → `<name>: for parameter 1: got <type>, want iterable`
- Wrong argument count → `<name>: got N arguments, want M`

### Distance metrics — `euclidean_distance`, `manhattan_distance`

`euclidean_distance(data1, data2)` returns the L2 distance; `manhattan_distance(data1, data2)` returns the L1 distance. Both treat the two iterables as coordinate vectors.

```python
load("stats", "euclidean_distance", "manhattan_distance")
print(euclidean_distance([3, 4], [0, 0]))
print(manhattan_distance([3, 4], [0, 0]))
# Output:
# 5.0
# 7.0
```

### Transforms — `softmax`, `sigmoid`

`softmax(data)` returns a list whose entries are non-negative and sum to 1. `sigmoid(data)` applies the logistic function element-wise. Both return a `list` of `float` the same length as the input, and error on empty input.

```python
load("stats", "softmax", "sigmoid")
print(softmax([1, 2, 3]))
print(sigmoid([0, 2, 4]))
# Output:
# [0.09003057317038046, 0.24472847105479764, 0.6652409557748218]
# [0.5, 0.8807970779778823, 0.9820137900379085]
```

### Basic measures — `sum`, `max`, `min`, `midrange`, `mode`

`sum`, `max`, `min`, and `midrange` return a single `float`. `mode(data)` returns a `list` because a dataset can have more than one most-frequent value. `midrange` is `(min + max) / 2`. All error on empty input.

```python
load("stats", "sum", "max", "min", "midrange", "mode")
print(sum([1, 2, 3, 4]))
print(max([1, 2, 3, 4]))
print(min([1, 2, 3, 4]))
print(midrange([1, 2, 3, 4]))
print(mode([1, 1, 2, 3, 3]))
# Output:
# 10.0
# 4.0
# 1.0
# 2.5
# [1.0, 3.0]
```

### Central tendency — `average`, `mean`, `geometric_mean`, `harmonic_mean`, `trimean`, `median`

`average` is an alias of `mean`. All return a single `float`.

```python
load("stats", "mean", "average", "geometric_mean", "harmonic_mean", "trimean", "median")
print(mean([1, 2, 3, 4]))
print(average([1, 2, 2, 3]))
print(geometric_mean([1, 2, 3, 4]))
print(harmonic_mean([1, 2, 3, 6]))
print(trimean([1, 2, 3, 4, 5]))
print(median([1, 2, 3, 4, 5]))
# Output:
# 2.5
# 2.0
# 2.2133638394006434
# 2.0
# 3.0
# 3.0
```

### Variability — `percentile`, `percentile_nearest_rank`, `variance`, `population_variance`, `sample_variance`

`percentile(data, p)` and `percentile_nearest_rank(data, p)` take a second positional argument `p` (the percentile, 0–100); both require exactly two arguments and error with `Input is outside of range.` if `p` is out of bounds. `variance` is population variance (same as `population_variance`); `sample_variance` uses the Bessel `n-1` correction.

```python
load("stats", "percentile", "percentile_nearest_rank", "variance", "sample_variance")
print(percentile([1, 2, 3, 4, 5], 50))
print(percentile_nearest_rank([1, 2, 3, 4, 5], 50))
print(variance([1, 2, 3, 4, 5]))
print(sample_variance([1, 2, 3, 4, 5]))
# Output:
# 2.5
# 3.0
# 2.0
# 2.5
```

### Covariance & correlation — `covariance`, `covariance_population`, `correlation`, `pearson`

Each takes two equal-length datasets and returns a single `float`; unequal lengths error with `Must be the same length.`. `pearson` is the Pearson product-moment coefficient; `correlation` uses the engine's correlation routine.

```python
load("stats", "covariance", "covariance_population", "correlation", "pearson")
print(covariance([1, 2, 3], [4, 5, 6]))
print(covariance_population([1, 2, 3], [4, 5, 6]))
print(correlation([1, 2, 3], [6, 5, 4]))
print(pearson([1, 2, 3], [1, 2, 3]))
# Output:
# 1.0
# 0.6666666666666666
# -1.0
# 1.0
```

### Standard deviation — `standard_deviation`, `stddev`, `stddev_sample`

`stddev` is an alias of `standard_deviation` (population standard deviation). `stddev_sample` uses the sample (`n-1`) correction.

```python
load("stats", "standard_deviation", "stddev", "stddev_sample")
print(standard_deviation([1, 2, 3, 4, 5]))
print(stddev([1, 2, 3, 4, 5]))
print(stddev_sample([1, 2, 3, 4, 5]))
# Output:
# 1.4142135623730951
# 1.4142135623730951
# 1.5811388300841898
```

### `sample(data, take, replace=False) -> list`

Randomly draws `take` elements from `data` and returns them as a `list` of `float`. Arguments are keyword-capable: `data`, `take`, and the optional `replace`. With `replace=False` the result is a subset (so `take` must not exceed `len(data)`, else `Input is outside of range.`); with `replace=True` the same element may be drawn more than once and `take` may exceed the input length. `take` is required — omitting it errors with `sample: missing argument for take`.

Because it draws on a random number generator, `sample` is the one **non-deterministic** function here; its element order and selection vary per call, so only the length is asserted in tests:

```python
load("stats", "sample")
r1 = sample(data=[1, 2, 3, 4], take=3, replace=False)
print(len(r1))
r2 = sample(data=[1, 2, 3, 4], take=5, replace=True)
print(len(r2))
# Output:
# 3
# 5
```

## Notes & boundaries

- **Engine.** Computation is delegated to `github.com/montanaflynn/stats`; numeric results match that library exactly (full IEEE-754 `float` precision, as shown in the examples).
- **Input domain.** Any iterable of `int`/`float` is accepted; ints are promoted to `float`. A non-iterable argument errors with `want iterable`. Validation messages (empty input, length mismatch, out-of-range) come straight from the engine.
- **Aliases.** `average` ≡ `mean`; `stddev` ≡ `standard_deviation`; `variance` ≡ `population_variance`. The sample-corrected (`n-1`) variants are `sample_variance` and `stddev_sample`.
- **Determinism.** All functions are pure and deterministic except `sample`, which is random.
- **Naming.** All exported members use `snake_case`.
