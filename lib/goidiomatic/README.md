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

### `hex(x)`

Convert an integer number to a lowercase hexadecimal string prefixed with `0x`.

#### Examples

**Basic**

Convert an integer to a hexadecimal string.

```python
load("go_idiomatic", "hex")
print(hex(255))
# Output: 0xff
```

**Negative**

Convert a negative integer to a hexadecimal string.

```python
load("go_idiomatic", "hex")
print(hex(-42))
# Output: -0x2a
```

### `oct(x)`

Convert an integer number to an octal string prefixed with `0o`.

#### Examples

**Basic**

Convert an integer to an octal string.

```python
load("go_idiomatic", "oct")
print(oct(255))
# Output: 0o377
```

**Negative**

Convert a negative integer to an octal string.

```python
load("go_idiomatic", "oct")
print(oct(-56))
# Output: -0o70
```

### `bin(x)`

Convert an integer number to a binary string prefixed with `0b`.

#### Examples

**Basic**

Convert an integer to a binary string.

```python
load("go_idiomatic", "bin")
print(bin(255))
# Output: 0b11111111
```

**Negative**

Convert a negative integer to a binary string.

```python
load("go_idiomatic", "bin")
print(bin(-10))
# Output: -0b1010
```

### `bytes_hex(bytes,sep="",bytes_per_sep=1)`

Return a string containing two hexadecimal digits for each byte in the instance.
If you want to make the hex string easier to read, you can specify a single character separator sep parameter to include in the output.
By default, this separator will be included between each byte.
A second optional bytes_per_sep parameter controls the spacing. Positive values calculate the separator position from the right, negative values from the left.

#### Parameters

| name            | type     | description                        |
|-----------------|----------|------------------------------------|
| `bytes`         | `bytes`  | The bytes to convert.              |
| `sep`           | `string` | The separator to use.              |
| `bytes_per_sep` | `int`    | The number of bytes per separator. |

#### Examples

**Basic**

Convert bytes to a hexadecimal string.

```python
load("go_idiomatic", "bytes_hex")
print(bytes_hex(b"hello"))
# Output: 68656c6c6f
```

**Separator**

Convert bytes to a hexadecimal string with a separator.

```python
load("go_idiomatic", "bytes_hex")
print(bytes_hex(b"hello", sep=":"))
# Output: 68:65:6c:6c:6f
```

**Bytes per separator**

Convert bytes to a hexadecimal string with a separator and bytes per separator.

```python
load("go_idiomatic", "bytes_hex")
print(bytes_hex(b"hello", sep=":", bytes_per_sep=2))
# Output: 68:656c:6c6f
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

### `module(name, **kv)`

Returns the module with the given name and keyword arguments.
The main difference between the `module` and the `struct` is that the string representation of the `module` does not enumerate its fields.
The module can't be compared with `==` and `!=`, but the `struct` can.

#### Parameters

| name   | type       | description                            |
|--------|------------|----------------------------------------|
| `name` | `string`   | The name of the module to return.      |
| `kv`   | `**kwargs` | Key-value pairs to provide attributes. |

#### Examples

**Basic**

Get the `os` module with pid attribute.

```python
load("go_idiomatic", "module")
os = module("os", pid=1)
print(os)
# Output: <module "os">
```

### `struct(**kv)`

Returns a new struct with the given keyword arguments.

#### Parameters

| name | type       | description                            |
|------|------------|----------------------------------------|
| `kv` | `**kwargs` | Key-value pairs to provide attributes. |

#### Examples

**Basic**

Create a struct with name and age attributes.

```python
load("go_idiomatic", "struct")
person = struct(name="Alice", age=30)
print(person)
# Output: struct(age = 30, name = "Alice")
```

### `make_struct(name, **kv)`

Returns a new struct with the given name as constructor and keyword arguments.
Comparing two structs with `==` and `!=` will compare their constructors first and then their fields.

#### Parameters

| name   | type       | description                            |
|--------|------------|----------------------------------------|
| `name` | `string`   | The name to use as constructor.        |
| `kv`   | `**kwargs` | Key-value pairs to provide attributes. |

#### Examples

**Basic**

Create a struct with name and age attributes.

```python
load("go_idiomatic", "make_struct")
person = make_struct("Person", name="Alice", age=30)
print(person)
# Output: Person(age = 30, name = "Alice")
```

## Types

### `nil`

Value as an alias for `None`.

### `true`

Value as an alias for `True`.

### `false`

Value as an alias for `False`.
