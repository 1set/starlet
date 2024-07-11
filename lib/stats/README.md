# stats

`stats` provides a Starlark module for comprehensive statistics functions. It is a wrapper around the Go package: https://github.com/montanaflynn/stats

## Functions

### `euclidean_distance(data1, data2) float`

Calculates the straight line distance between two points in Euclidean space.

#### Parameters

| name    | type   | description                         |
|---------|--------|-------------------------------------|
| `data1` | `list` | First dataset of numerical values.  |
| `data2` | `list` | Second dataset of numerical values. |

#### Examples

**basic**

Calculate Euclidean distance between two points.

```python
load("stats", "euclidean_distance")
print(euclidean_distance([3, 4], [0, 0]))  # Output: 5.0
```

### `manhattan_distance(data1, data2) float`

Computes the sum of the absolute differences of their coordinates.

#### Parameters

| name    | type   | description                         |
|---------|--------|-------------------------------------|
| `data1` | `list` | First dataset of numerical values.  |
| `data2` | `list` | Second dataset of numerical values. |

#### Examples

**basic**

Calculate Manhattan distance between two points.

```python
load("stats", "manhattan_distance")
print(manhattan_distance([3, 4], [0, 0]))  # Output: 7.0
```

### `softmax(data) list`

Applies the Softmax function, useful for converting scores to probabilities.

#### Parameters

| name   | type   | description                  |
|--------|--------|------------------------------|
| `data` | `list` | Dataset of numerical values. |

#### Examples

**basic**

Apply Softmax function to a dataset.

```python
load("stats", "softmax")
print(softmax([1, 1, 1]))  # Output: [0.3333333333333333, 0.3333333333333333, 0.3333333333333333]
```

### `sigmoid(data) list`

Applies the Sigmoid function, often used for binary classification.

#### Parameters

| name   | type   | description                  |
|--------|--------|------------------------------|
| `data` | `list` | Dataset of numerical values. |

#### Examples

**basic**

Apply Sigmoid function to a dataset.

```python
load("stats", "sigmoid")
print(sigmoid([0, 2, 4]))  # Output: [0.5, 0.8807970779778823, 0.9820137900379085]
```

### `mode(data) list`

Determines the most frequently occurring data points in a dataset.

#### Parameters

| name   | type   | description                  |
|--------|--------|------------------------------|
| `data` | `list` | Dataset of numerical values. |

#### Examples

**basic**

Calculate mode of a dataset.

```python
load("stats", "mode")
print(mode([1, 2, 2, 3, 4]))  # Output: [2.0]
```

### `sum(data) float`

Computes the sum of a series of numbers.

#### Parameters

| name   | type   | description                  |
|--------|--------|------------------------------|
| `data` | `list` | Dataset of numerical values. |

#### Examples

**basic**

Calculate sum of a dataset.

```python
load("stats", "sum")
print(sum([1, 2, 3, 4]))  # Output: 10.0
```

### `max(data) float`

Finds the maximum value in a dataset.

#### Parameters

| name   | type   | description                  |
|--------|--------|------------------------------|
| `data` | `list` | Dataset of numerical values. |

#### Examples

**basic**

Find maximum value in a dataset.

```python
load("stats", "max")
print(max([1, 2, 3, 4]))  # Output: 4.0
```

### `min(data) float`

Finds the minimum value in a dataset.

#### Parameters

| name   | type   | description                  |
|--------|--------|------------------------------|
| `data` | `list` | Dataset of numerical values. |

#### Examples

**basic**

Find minimum value in a dataset.

```python
load("stats", "min")
print(min([1, 2, 3, 4]))  # Output: 1.0
```

### `midrange(data) float`

Calculates the midrange, the average of the maximum and minimum values.

#### Parameters

| name   | type   | description                  |
|--------|--------|------------------------------|
| `data` | `list` | Dataset of numerical values. |

#### Examples

**basic**

Calculate midrange of a dataset.

```python
load("stats", "midrange")
print(midrange([1, 2, 3, 4]))  # Output: 2.5
```

### `average(data) float`

Alias for mean. Calculates the arithmetic mean of a dataset.

#### Parameters

| name   | type   | description                  |
|--------|--------|------------------------------|
| `data` | `list` | Dataset of numerical values. |

#### Examples

**basic**

Calculate average of a dataset.

```python
load("stats", "average")
print(average([1, 2, 2, 3]))  # Output: 2.0
```

### `mean(data) float`

Calculates the arithmetic mean of a dataset.

#### Parameters

| name   | type   | description                  |
|--------|--------|------------------------------|
| `data` | `list` | Dataset of numerical values. |

#### Examples

**basic**

Calculate mean of a dataset.

```python
load("stats", "mean")
print(mean([1, 2, 3, 4]))  # Output: 2.5
```

### `geometric_mean(data) float`

Computes the geometric mean, useful for datasets with exponential growth.

#### Parameters

| name   | type   | description                  |
|--------|--------|------------------------------|
| `data` | `list` | Dataset of numerical values. |

#### Examples

**basic**

Calculate geometric mean of a dataset.

```python
load("stats", "geometric_mean")
print(geometric_mean([1, 2, 3, 4]))  # Output: 2.213363839400643
```

### `harmonic_mean(data) float`

Computes the harmonic mean, effective for rates and ratios.

#### Parameters

| name   | type   | description                  |
|--------|--------|------------------------------|
| `data` | `list` | Dataset of numerical values. |

#### Examples

**basic**

Calculate harmonic mean of a dataset.

```python
load("stats", "harmonic_mean")
print(harmonic_mean([1, 2, 3, 6]))  # Output: 2.0
```

### `trimean(data) float`

Calculates the trimean, a measure of a dataset's tendency.

#### Parameters

| name   | type   | description                  |
|--------|--------|------------------------------|
| `data` | `list` | Dataset of numerical values. |

#### Examples

**basic**

Calculate trimean of a dataset.

```python
load("stats", "trimean")
print(trimean([1, 2, 3, 4, 5]))  # Output: 3.0
```

### `median(data) float`

Finds the middle value of a dataset.

#### Parameters

| name   | type   | description                  |
|--------|--------|------------------------------|
| `data` | `list` | Dataset of numerical values. |

#### Examples

**basic**

Calculate median of a dataset.

```python
load("stats", "median")
print(median([1, 2, 3, 4, 5]))  # Output: 3.0
```

### `percentile(data, p) float`

Determines the value below which a given percentage of observations in a group of observations falls.

#### Parameters

| name   | type    | description                  |
|--------|---------|------------------------------|
| `data` | `list`  | Dataset of numerical values. |
| `p`    | `float` | The percentile to compute.   |

#### Examples

**basic**

Calculate 50th percentile of a dataset.

```python
load("stats", "percentile")
print(percentile([1, 2, 3, 4, 5], 50))  # Output: 2.5
```

### `percentile_nearest_rank(data, p) float`

Determines the value below which a given percentage of observations in a group of observations falls using the nearest rank method.

#### Parameters

| name   | type    | description                  |
|--------|---------|------------------------------|
| `data` | `list`  | Dataset of numerical values. |
| `p`    | `float` | The percentile to compute.   |

#### Examples

**basic**

Calculate 50th percentile using nearest rank method.

```python
load("stats", "percentile_nearest_rank")
print(percentile_nearest_rank([1, 2, 3, 4, 5], 50))  # Output: 3.0
```

### `variance(data) float`

Calculates the variance of a dataset, a measure of dispersion.

#### Parameters

| name   | type   | description                  |
|--------|--------|------------------------------|
| `data` | `list` | Dataset of numerical values. |

#### Examples

**basic**

Calculate variance of a dataset.

```python
load("stats", "variance")
print(variance([1, 2, 3, 4, 5]))  # Output: 2.0
```

### `covariance(data1, data2) float`

Measures how changes in one variable are associated with changes in another variable.

#### Parameters

| name    | type   | description                         |
|---------|--------|-------------------------------------|
| `data1` | `list` | First dataset of numerical values.  |
| `data2` | `list` | Second dataset of numerical values. |

#### Examples

**basic**

Calculate covariance between two datasets.

```python
load("stats", "covariance")
print(covariance([1, 2, 3], [4, 5, 6]))  # Output: 1.0
```

### `covariance_population(data1, data2) float`

Calculates the covariance for an entire population.

#### Parameters

| name    | type   | description                         |
|---------|--------|-------------------------------------|
| `data1` | `list` | First dataset of numerical values.  |
| `data2` | `list` | Second dataset of numerical values. |

#### Examples

**basic**

Calculate population covariance between two datasets.

```python
load("stats", "covariance_population")
print(covariance_population([1, 2, 3], [4, 5, 6]))  # Output: 1.0
```

### `population_variance(data) float`

Computes the variance of an entire population.

#### Parameters

| name   | type   | description                  |
|--------|--------|------------------------------|
| `data` | `list` | Dataset of numerical values. |

#### Examples

**basic**

Calculate population variance of a dataset.

```python
load("stats", "population_variance")
print(population_variance([1, 2, 3, 4, 5]))  # Output: 2.0
```

### `sample_variance(data) float`

Computes the variance of a sample from the population.

#### Parameters

| name   | type   | description                  |
|--------|--------|------------------------------|
| `data` | `list` | Dataset of numerical values. |

#### Examples

**basic**

Calculate sample variance of a dataset.

```python
load("stats", "sample_variance")
print(sample_variance([1, 2, 3, 4, 5]))  # Output: 2.5
```

### `correlation(data1, data2) float`

Computes the correlation coefficient between two datasets.

#### Parameters

| name    | type   | description                         |
|---------|--------|-------------------------------------|
| `data1` | `list` | First dataset of numerical values.  |
| `data2` | `list` | Second dataset of numerical values. |

#### Examples

**basic**

Calculate correlation between two datasets.

```python
load("stats", "correlation")
print(correlation([1, 2, 3], [1, 2, 3]))  # Output: 1.0
```

### `pearson(data1, data2) float`

Computes Pearson's correlation coefficient.

#### Parameters

| name    | type   | description                         |
|---------|--------|-------------------------------------|
| `data1` | `list` | First dataset of numerical values.  |
| `data2` | `list` | Second dataset of numerical values. |

#### Examples

**basic**

Calculate Pearson's correlation coefficient.

```python
load("stats", "pearson")
print(pearson([1, 2, 3], [1, 2, 3]))  # Output: 1.0
```

### `standard_deviation(data) float`

Calculates the standard deviation of a dataset.

#### Parameters

| name   | type   | description                  |
|--------|--------|------------------------------|
| `data` | `list` | Dataset of numerical values. |

#### Examples

**basic**

Calculate standard deviation of a dataset.

```python
load("stats", "standard_deviation")
print(standard_deviation([1, 2, 3, 4, 5]))  # Output: 1.4142135623730951
```

### `stddev(data) float`

Alias for standard_deviation. Calculates the standard deviation of a dataset.

#### Parameters

| name   | type   | description                  |
|--------|--------|------------------------------|
| `data` | `list` | Dataset of numerical values. |

#### Examples

**basic**

Calculate stddev of a dataset.

```python
load("stats", "stddev")
print(stddev([1, 2, 3, 4, 5]))  # Output: 1.4142135623730951
```

### `stddev_sample(data) float`

Calculates the standard deviation of a sample.

#### Parameters

| name   | type   | description                  |
|--------|--------|------------------------------|
| `data` | `list` | Dataset of numerical values. |

#### Examples

**basic**

Calculate standard deviation of a sample.

```python
load("stats", "stddev_sample")
print(stddev_sample([1, 2, 3, 4, 5]))  # Output: 1.5811388300841898
```

### `sample(data, take, replace=False) list`

Randomly samples elements from the dataset.

#### Parameters

| name      | type   | description                                           |
|-----------|--------|-------------------------------------------------------|
| `data`    | `list` | Dataset of numerical values.                          |
| `take`    | `int`  | Number of elements to sample.                         |
| `replace` | `bool` | Optional. If True, sampling is done with replacement. |

#### Examples

**basic**

Sample elements from a dataset.

```python
load("stats", "sample")
print(sample(data=[1, 2, 3, 4], take=3, replace=False))
```
