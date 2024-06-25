# random

`random` defines functions that generate random values for various distributions, it's intended to be a drop-in subset of [Python's **random** module](https://docs.python.org/3/library/random.html) for Starlark.

## Functions

### `randbytes(n)`

Generate a random byte string containing n number of bytes.

#### Parameters

| name | type  | description                                                               |
|------|-------|---------------------------------------------------------------------------|
| `n`  | `int` | If n bytes is non-positive or not supplied, a reasonable default is used. |

#### Examples

**basic**

Generate a random byte string containing 10 bytes.

```python
load("random", "randbytes")
b = randbytes(10)
print(b)
# Output: b'K\xaa\xbb4\xbaEh0\x19\x9c'
```

### `randstr(chars, n)`

Generate a random string containing n number of unicode characters from the given unicode string.

#### Parameters

| name    | type     | description                                                                                   |
|---------|----------|-----------------------------------------------------------------------------------------------|
| `chars` | `string` | The characters to choose from.                                                                |
| `n`     | `int`    | The length of the string. If n is non-positive or not supplied, a reasonable default is used. |

#### Examples

**basic**

Generate a random string containing 10 characters from the given unicode string.

```python
load("random", "randstr")
s = randstr("abcdefghijklmnopqrstuvwxyz", 10)
print(s)
# Output: "enfknqfbra"
```

### `randb32(n, sep)`

Generate a random base32 string containing n number of bytes with optional separator dash for every sep characters.

#### Parameters

| name  | type  | description                                                                                                   |
|-------|-------|---------------------------------------------------------------------------------------------------------------|
| `n`   | `int` | The number of bytes to generate. If n is non-positive or not supplied, a reasonable default is used.          |
| `sep` | `int` | The number of characters to separate with a dash, if it's non-positive or not supplied, no separator is used. |

#### Examples

**basic**

Generate a random base32 string containing 10 bytes with no separator.

```python
load("random", "randb32")
s = randb32(10, 4)
print(s)
# Output: 2RXQ-H45H-WV
```

### `randint(a,b)`

Return a random integer N such that a <= N <= b.

#### Parameters

| name | type  | description                   |
|------|-------|-------------------------------|
| `a`  | `int` | The lower bound of the range. |
| `b`  | `int` | The upper bound of the range. |

#### Examples

**basic**

Return a random integer N such that 0 <= N <= 10.

```python
load("random", "randint")
n = randint(0, 10)
print(n)
# Output: 7
```

### `random()`

Return a random floating point number in the range 0.0 <= X < 1.0.

#### Examples

**basic**

Return a random floating point number in the range [0.0, 1.0).

```python
load("random", "random")
n = random()
print(n)
# Output: 0.7309677873766576
```

### `uniform(a, b)`

Return a random floating point number N such that a <= N <= b for a <= b and b <= N <= a for b < a.
The end-point value b may or may not be included in the range depending on floating-point rounding in the equation a + (b-a) * random().

#### Parameters

| name | type    | description                   |
|------|---------|-------------------------------|
| `a`  | `float` | The lower bound of the range. |
| `b`  | `float` | The upper bound of the range. |

#### Examples

**basic**

Return a random floating point number N such that 5.0 <= N <= 10.0.

```python
load("random", "uniform")
n = uniform(5, 10)
print(n)
# Output: 7.309677873766576
```

### `uuid()`

Generate a random UUID (RFC 4122 version 4).

#### Examples

**basic**

Generate a random UUID.

```python
load("random", "uuid")
u = uuid()
print(u)
# Output: 6e360b7a-f677-4f6c-9c57-8b09694d66b3
```

### `choice(seq)`

Return a random element from the non-empty sequence seq.

#### Parameters

| name  | type   | description           |
|-------|--------|-----------------------|
| `seq` | `list` | A non-empty sequence. |

#### Examples

**basic**

Return a random element from the non-empty sequence [1, 2, 3, 4, 5].

```python
load("random", "choice")
n = choice([1, 2, 3, 4, 5])
print(n)
# Output: 3
```

### `choices(population, weights=None, cum_weights=None, k=1)`

Return a k-sized list of elements chosen from the population with replacement. If the population is empty, raises a ValueError.

#### Parameters

| name          | type   | description                                                                                                                                                |
|---------------|--------|------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `population`  | `list` | A non-empty sequence to choose from.                                                                                                                       |
| `weights`     | `list` | A sequence of weights corresponding to the population, where weights[i] is the weight of population[i]. If not provided, all weights are considered equal. |
| `cum_weights` | `list` | A sequence of cumulative weights corresponding to the population. If provided, weights should not be provided.                                             |
| `k`           | `int`  | The size of the result list. If not provided, defaults to 1. If k is non-positive, an empty list is returned.                                              |

#### Examples

**basic**

Return a list of 2 random elements chosen from the population [1, 2, 3].

```python
load("random", "choices")
result = choices([1, 2, 3], k=2)
print(result)
# Output: [2, 3]
```

**with_weights**

Return a list of 3 random elements chosen from the population [1, 2, 3] with given weights.

```python
load("random", "choices")
result = choices([1, 2, 3], weights=[1, 2, 1], k=3)
print(result)
# Output: [2, 3, 1]
```

**with_cumulative_weights**

Return a list of 2 random elements chosen from the population [1, 2, 3] with given cumulative weights.

```python
load("random", "choices")
result = choices([1, 2, 3], cum_weights=[1, 3, 4], k=2)
print(result)
# Output: [3, 2]
```

### `shuffle(x)`

Shuffle the sequence x in place.

#### Parameters

| name | type   | description           |
|------|--------|-----------------------|
| `x`  | `list` | A non-empty sequence. |

#### Examples

**basic**

Shuffle the sequence [1, 2, 3, 4, 5] in place.

```python
load("random", "shuffle")
x = [1, 2, 3, 4, 5]
shuffle(x)
print(x)
# Output: [3, 1, 5, 4, 2]
```
