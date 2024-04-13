# atom

atom provides atomic operations for integers, floats, and strings.

## Functions

### `new_int(value=0) -> AtomicInt`

create a new AtomicInt with an optional initial value

#### Parameters

| name    | type  | description                  |
|---------|-------|------------------------------|
| `value` | `int` | initial value, defaults to 0 |

#### Examples

**basic**

create a new AtomicInt with default value

```python
load("atom", "new_int")
ai = new_int()
ai.inc()
print(ai.get())
# Output: 1
```

**with value**

create a new AtomicInt with a specific value

```python
load("atom", "new_int")
ai = new_int(42)
ai.add(42)
print(ai.get())
# Output: 84
```

### `new_float(value=0.0) -> AtomicFloat`

create a new AtomicFloat with an optional initial value

#### Parameters

| name    | type    | description                    |
|---------|---------|--------------------------------|
| `value` | `float` | initial value, defaults to 0.0 |

#### Examples

**basic**

create a new AtomicFloat with default value

```python
load("atom", "new_float")
af = new_float()
print(af.get())
# Output: 0.0
```

**with value**

create a new AtomicFloat with a specific value

```python
load("atom", "new_float")
af = new_float(3.14)
print(af.get())
# Output: 3.14
```

### `new_string(value="") -> AtomicString`

create a new AtomicString with an optional initial value

#### Parameters

| name    | type     | description                                |
|---------|----------|--------------------------------------------|
| `value` | `string` | initial value, defaults to an empty string |

#### Examples

**basic**

create a new AtomicString with default value

```python
load("atom", "new_string")
as = new_string()
print(as.get())  # Output: ""
```

**with value**

create a new AtomicString with a specific value

```python
load("atom", "new_string")
as = new_string("hello")
print(as.get())
# Output: "hello"
```

## Types

### `AtomicInt`

an atomic integer type with various atomic operations

**Methods**

#### `get() -> int`

returns the current value

#### `set(value: int)`

sets the value

#### `cas(old: int, new: int) -> bool`

compares and swaps the value if it matches old

#### `add(delta: int) -> int`

adds delta to the value and returns the new value

#### `sub(delta: int) -> int`

subtracts delta from the value and returns the new value

#### `inc() -> int`

increments the value by 1 and returns the new value

#### `dec() -> int`

decrements the value by 1 and returns the new value

### `AtomicFloat`

an atomic float type with various atomic operations

**Methods**

#### `get() -> float`

returns the current value

#### `set(value: float)`

sets the value

#### `cas(old: float, new: float) -> bool`

compares and swaps the value if it matches old

#### `add(delta: float) -> float`

adds delta to the value and returns the new value

#### `sub(delta: float) -> float`

subtracts delta from the value and returns the new value

### `AtomicString`

an atomic string type with various atomic operations

**Methods**

#### `get() -> string`

returns the current value

#### `set(value: string)`

sets the value

#### `cas(old: string, new: string) -> bool`

compares and swaps the value if it matches old
