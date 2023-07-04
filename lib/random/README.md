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
