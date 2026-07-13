# atom

`atom` provides atomic operations for integers, floats, and strings — lock-free counters and compare-and-swap cells safe to share across concurrent script work (backed by Go's `sync/atomic` via `go.uber.org/atomic`). Capability profile: **pure** (no filesystem, network, process, or log side effects).

## Functions

| function | description |
|----------|-------------|
| `new_int(value=0) -> atom_int` | create an atomic integer cell, optionally seeded with `value` |
| `new_float(value=0.0) -> atom_float` | create an atomic float cell, optionally seeded with `value` |
| `new_string(value="") -> atom_string` | create an atomic string cell, optionally seeded with `value` |

The constructors are the module's only top-level members; all mutation happens through methods on the returned cells (see Types).

## Types

The cells created above are custom Starlark values. Each is truthy when its value is non-zero / non-empty, and ordered (`==`, `!=`, `<`, `<=`, `>`, `>=` compare by current value against another cell of the same type). Cells are **not hashable** — a mutable cell cannot be a dict/set key (a live-value hash would go stale on mutation and strand the entry); key by `x.get()` when a snapshot key is what you mean. `str()` renders as `<atom_int:N>`, `<atom_float:N>`, `<atom_string:"S">`.

### `atom_int`

An atomic `int64` cell. Methods:

| method | description |
|--------|-------------|
| `get() -> int` | return the current value |
| `set(value)` | store `value` (int); returns `None` |
| `cas(old, new) -> bool` | atomically set to `new` only if the current value equals `old`; returns whether the swap happened |
| `add(delta) -> int` | add `delta` (int) and return the new value |
| `sub(delta) -> int` | subtract `delta` (int) and return the new value |
| `inc() -> int` | add 1 and return the new value |
| `dec() -> int` | subtract 1 and return the new value |

### `atom_float`

An atomic `float64` cell. `set`, `cas`, `add`, and `sub` accept either a float or an int (ints are widened to float). Methods:

| method | description |
|--------|-------------|
| `get() -> float` | return the current value |
| `set(value)` | store `value` (float or int); returns `None` |
| `cas(old, new) -> bool` | atomically set to `new` only if the current value equals `old`; returns whether the swap happened |
| `add(delta) -> float` | add `delta` (float or int) and return the new value |
| `sub(delta) -> float` | subtract `delta` (float or int) and return the new value |

### `atom_string`

An atomic string cell. Methods:

| method | description |
|--------|-------------|
| `get() -> string` | return the current value |
| `set(value)` | store `value` (string); returns `None` |
| `cas(old, new) -> bool` | atomically set to `new` only if the current value equals `old`; returns whether the swap happened |

## Details & examples

### `new_int`

`new_int(value=0) -> atom_int` — `value` is an optional initial `int` (defaults to `0`). Errors when `value` is not an int (e.g. `new_int('42')` → `new_int: for parameter value: got string, want int`).

```python
load("atom", "new_int")
x = new_int()
x.inc()
x.set(20)
print(x.add(5), x.sub(3), x.cas(22, 100), x.get())
# Output: 25 22 True 100
```

### `new_float`

`new_float(value=0.0) -> atom_float` — `value` is an optional initial number; an int is accepted and widened to float. Errors when `value` is a non-numeric type (e.g. `new_float('42.1')` → `new_float: for parameter value: got string, want float`).

```python
load("atom", "new_float")
x = new_float(1)
x.set(20.1)
print(x.add(5), x.cas(22.1, 200.5), x.get())
# Output: 25.1 False 25.1
```

### `new_string`

`new_string(value="") -> atom_string` — `value` is an optional initial `string` (defaults to `""`). Errors when `value` is not a string (e.g. `new_string(1)` → `new_string: for parameter value: got int, want string`).

```python
load("atom", "new_string")
x = new_string("hello")
x.set("world")
print(x.cas("world", "new"), x.get(), x.cas("world", "new2"), x.get())
# Output: True new False new
```

### Methods, errors, and concurrency

- `get`, `inc`, and `dec` take no arguments — passing any errors (e.g. `x.get(2)` → `get: got 1 arguments, want 0`).
- `set`, `add`, `sub`, and `cas` validate their argument types and error on a mismatch (e.g. `x.add('2')` on an `atom_int` → `add: for parameter delta: got string, want int`).
- An unknown attribute errors via the standard Starlark message (e.g. `x.guess()` → `atom_int has no .guess field or method`).
- Operations are atomic, so a cell can be safely mutated from comprehensions or callbacks sharing it:

```python
load("atom", "new_int")
x = new_int()
def work():
    x.inc()
[work() for _ in range(10)]
print(x.get())
# Output: 10
```

## Notes / boundaries

- **Type names.** Script-visible type names are `atom_int`, `atom_float`, `atom_string` (as returned by `type()`); the underlying Go types are `AtomicInt`, `AtomicFloat`, `AtomicString`.
- **Truthiness.** A cell is falsy when its value is `0` / `0.0` / `""` and truthy otherwise (`bool(new_int(0))` is `False`).
- **Comparison and hashing.** Cells are comparable only against the same cell type. They are **not hashable**: using a cell as a dict/set key errors with `unhashable type: atom_int` (or `atom_float` / `atom_string`) — a live-value hash goes stale when the cell mutates and the keyed entry becomes unreachable. Key by `x.get()` instead.
- **Freezing.** Once a cell is frozen (Starlark freezes values that cross a module boundary), every mutating method — `set`, `cas`, `add`, `sub`, `inc`, `dec` — errors with `cannot <method> frozen atom_int` (or `atom_float` / `atom_string`); `get`, truthiness, and comparisons still work.
- **Range.** `atom_int` is a 64-bit signed integer; `atom_float` is IEEE-754 `float64`. `add`/`sub` wrap or lose precision exactly as the underlying 64-bit types do.
- **Float widening.** `atom_float` accepts ints for `set`/`cas`/`add`/`sub` and stores them as floats; `cas` compares with float equality, so seed `old` with the exact stored float.
