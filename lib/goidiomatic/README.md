# Go Idiomatic

`go_idiomatic` provides a Starlark module that defines Go idiomatic functions and values.

## Functions

### `length(obj) int`

Returns the length of the object, for string it returns the number of Unicode code points, instead of bytes like `len()`.

#### Examples

**String**

Calculate the length of a CJK string.

```python
load("go_idiomatic", "length")
s = "你好"
print(length(s), len(s))
# Output: 2 6
```

**Misc**

Calculate the length of a list, set and map.

```python
load("go_idiomatic", "length")
print(length([1, 2, 3]), length(set([1, 2])), length({1: 2}))
# Output: 3 2 1
```

### `sum(iterable, start=0)`

Returns the sum of `start` and the items of an iterable from left to right. The iterable's items and the `start` value are normally numbers.

#### Examples

**Basic**

Calculate the sum of a list.

```python
load("go_idiomatic", "sum")
print(sum([1, 2, 3]))
# Output: 6
```

**Start**

Calculate the sum of a list with a start value.

```python
load("go_idiomatic", "sum")
print(sum([1, 2, 3], 10))
# Output: 16
```

### `sleep(secs)`

Sleeps for the given number of seconds.

#### Examples

**Basic**

Sleep for 1 second.

```python
load("go_idiomatic", "sleep")
sleep(1)
```

### `exit(code=0)`

Exits the program with the given exit code.

#### Examples

**Default**

Exit with default code (0).

```python
load("go_idiomatic", "exit")
exit()
```

**Non-zero**

Exit with code 1.

```python
load("go_idiomatic", "exit")
exit(1)
```

### `quit(code=0)`

Alias for `exit()`.

#### Examples

**Default**

Exit with default code (0).

```python
load("go_idiomatic", "quit")
quit()
```

**Non-zero**

Exit with code 1.

```python
load("go_idiomatic", "quit")
quit(1)
```

## Types

### `nil`

Value as an alias for `None`.

### `true`

Value as an alias for `True`.

### `false`

Value as an alias for `False`.
