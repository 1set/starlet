# random

`random` generates random values for various distributions — a drop-in subset of [Python's `random` module](https://docs.python.org/3/library/random.html) for Starlark. All randomness is drawn from the OS cryptographic source (`crypto/rand`), so it is suitable for security-sensitive use. **Capability profile: Pure** — no filesystem, network, process, or log side effects.

## Functions

| function | description |
|----------|-------------|
| `randbytes(n=10) -> bytes` | Random byte string of `n` bytes. |
| `randstr(chars, n=10) -> str` | Random string of `n` characters drawn from `chars`. |
| `randb32(n=10, sep=0) -> str` | Random base32 string of `n` characters, optionally dash-separated every `sep` characters. |
| `randint(a, b) -> int` | Random integer `N` with `a <= N <= b`. |
| `random() -> float` | Random float in `[0.0, 1.0)`. |
| `uniform(a, b) -> float` | Random float between `a` and `b`. |
| `choice(seq) -> value` | Random element from the non-empty sequence `seq`. |
| `choices(population, weights=None, cum_weights=None, k=1) -> list` | `k`-sized list chosen from `population` with replacement, optionally weighted. |
| `shuffle(seq) -> None` | Shuffle the mutable sequence `seq` in place. |
| `uuid() -> str` | Random RFC 4122 version 4 UUID string. |

This module exposes no constants and no custom types — every member is a function on the `random` module.

## Details & examples

### `randbytes(n=10) -> bytes`

Returns a random byte string of length `n`. If `n` is non-positive or omitted, the default length `10` is used. `n` is capped at **1 MiB (1048576)** — a larger value errors (`n is too large: … (max 1048576)`) rather than letting a script drive an unbounded allocation. Also errors if the underlying RNG read fails.

```python
load("random", "randbytes")
b = randbytes(10)
print(len(b))
# Output: 10
```

### `randstr(chars, n=10) -> str`

Returns a random string of `n` characters, each drawn uniformly from the characters in `chars`. `chars` is split into Unicode runes, so multi-byte characters are selected as whole code points (length is measured in bytes, e.g. each Chinese character contributes 3). If `n` is non-positive or omitted, the default length `10` is used; `n` is capped at **1 MiB (1048576)** and a larger value errors (`n is too large`). Errors when `chars` is empty (`chars must not be empty`).

```python
load("random", "randstr")
x = randstr("AAA", 10)
print(x)
# Output: AAAAAAAAAA
```

### `randb32(n=10, sep=0) -> str`

Returns a random base32 string of `n` characters using the standard RFC 4648 alphabet (`A-Z`, `2-7`). If `sep` is positive and smaller than the string length, a dash is inserted every `sep` characters (this adds separator characters to the total length). If `n` is non-positive or omitted, the default length `10` is used; `n` is capped at **1 MiB (1048576)** and a larger value errors (`n is too large`). If `sep` is non-positive or omitted, no separator is inserted.

```python
load("random", "randb32")
x = randb32(20, 5)
print(len(x), x[5], len(x.split("-")))
# Output: 23 - 4
```

### `randint(a, b) -> int`

Returns a random integer `N` such that `a <= N <= b` (inclusive on both ends). Both `a` and `b` must be integers; backed by arbitrary-precision big integers. Errors when `a > b` (`a must be less than or equal to b`).

```python
load("random", "randint")
val = randint(1, 1)
print(val)
# Output: 1
```

### `random() -> float`

Returns a random float in the range `[0.0, 1.0)`. Takes no arguments; passing any argument errors (`random.random: got 1 arguments, want 0`).

```python
load("random", "random")
val = random()
print((0 <= val) and (val < 1))
# Output: True
```

### `uniform(a, b) -> float`

Returns a random float `N` between `a` and `b`, computed as `a + (b - a) * random()`. For `a <= b` the range is `a <= N <= b`; for `b < a` it is `b <= N <= a`. The end-point `b` may or may not be included depending on floating-point rounding. Both `a` and `b` accept `int` or `float`.

```python
load("random", "uniform")
val = uniform(1, 1)
print(val)
# Output: 1.0
```

### `choice(seq) -> value`

Returns a single random element from the non-empty indexable sequence `seq` (e.g. list, tuple, range). Errors when `seq` is empty (`cannot choose from an empty sequence`) or not indexable.

```python
load("random", "choice")
val = choice((3, 3, 3, 3, 3))
print(val)
# Output: 3
```

### `choices(population, weights=None, cum_weights=None, k=1) -> list`

Returns a `k`-sized list of elements chosen from `population` **with replacement**.

- `population` — a non-empty indexable sequence.
- `weights` — relative weights per element; if omitted, selection is uniform.
- `cum_weights` — cumulative weights per element; cannot be combined with `weights`.
- `k` — result size (default `1`); if `k <= 0`, an empty list is returned.

Errors on: empty `population` (`population is empty`); both `weights` and `cum_weights` given (`cannot specify both weights and cumulative weights`); a weight list whose length differs from `population` (`the number of weights does not match the population`); non-numeric weights (`weights must be numeric`); decreasing cumulative weights (`cumulative weights must be non-decreasing`); a non-positive weight total (`total of weights must be greater than zero`); or a non-finite total (`total of weights must be finite`).

```python
load("random", "choices")
a = choices([1, 2, 3], weights=[0, 1, 0])
print(a)
# Output: [2]
```

A zero weight (or a flat cumulative segment) makes an element unreachable, so deterministic weight vectors yield deterministic results:

```python
load("random", "choices")
a = choices([1, 2, 3, 4, 5], cum_weights=[0, 0, 1, 1, 1])
print(a)
# Output: [3]
```

### `shuffle(seq) -> None`

Shuffles the mutable sequence `seq` in place using the Fisher-Yates algorithm and returns `None`. `seq` must support index assignment (a list); tuples and other immutable sequences error (`want starlark.HasSetIndex`), and a frozen list errors (`cannot assign to element of frozen list`). Sequences of length 0 or 1 are left unchanged.

```python
load("random", "shuffle")
val = [1]
shuffle(val)
print(val)
# Output: [1]
```

### `uuid() -> str`

Returns a random UUID (RFC 4122 version 4) as a 36-character string (32 hex digits plus 4 dashes). Takes no arguments; passing any argument errors (`random.uuid: got 1 arguments, want 0`).

```python
load("random", "uuid")
val = uuid()
print(len(val), len(val.replace("-", "")))
# Output: 36 32
```

## Notes / boundaries

- **Engine.** All values come from `crypto/rand` (the OS CSPRNG); integers use `math/big`, so `randint` is exact for arbitrarily large bounds. There is no seeding API and no `random.seed`/`getrandbits` equivalent — output is non-deterministic by design and cannot be made reproducible.
- **Float precision.** `random()` and `uniform()` quantize to `1/2^53` (53 bits of mantissa), matching CPython's effective precision.
- **Python parity.** This is a subset: `randbytes`, `randstr`, and `randb32` are extensions not present in CPython; `randint`, `random`, `uniform`, `choice`, `choices`, and `shuffle` mirror their CPython signatures. `choices` returns a `list` (never a tuple). Not provided: `randrange`, `sample`, `seed`, `getstate`/`setstate`, and the distribution helpers (`gauss`, `betavariate`, …).
