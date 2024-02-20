# string

`string` provides constants and functions to manipulate strings, it's intended to be a drop-in subset of Python's string module for Starlark.

## Functions

### `length(obj) int`

Returns the length of the object; for string, it returns the number of Unicode code points, instead of bytes like `len()`.

#### Parameters

| name  | type     | description                                 |
|-------|----------|---------------------------------------------|
| `obj` | `string` | The object whose length is to be calculated |

#### Examples

**String**

Calculate the length of a CJK string.

```python
load("string", "length")
s = "你好"
print(length(s), len(s))
# Output: 2 6
```

**Misc**

Calculate the length of a list, set and map.

```python
load("string", "length")
print(length([1, 2, 3]), length(set([1, 2])), length({1: 2}))
# Output: 3 2 1
```

### `reverse(str) string`

Returns the reversed string of the given value.

#### Parameters

| name  | type     | description                     |
|-------|----------|---------------------------------|
| `str` | `string` | A string that is to be reversed |

#### Examples

**String**

Reverse a string.

```python
load("string", "reverse")
s = "abcdef"
print(reverse(s))
# Output: "fedcba"
```

### `escape(str) string`

Converts the characters "&", "<", ">", '"' and "'" in string to their corresponding HTML entities.

#### Parameters

| name  | type     | description                          |
|-------|----------|--------------------------------------|
| `str` | `string` | A string which is to be HTML escaped |

#### Examples

**String**

Escape a string.

```python
load("string", "escape")
s = "Hello<World>"
print(escape(s))
# Output: "Hello&lt;World&gt;"
```

### `unescape(str) string`

Converts the HTML entities in a string back to their corresponding characters.

#### Parameters

| name  | type     | description           |
|-------|----------|-----------------------|
| `str` | `string` | A HTML escaped string |

#### Examples

**String**

Unescape a string.

```python
load("string", "unescape")
s = "Hello&amp;World"
print(unescape(s))
# Output: "Hello&World"
```

### `quote(str) string`

Returns a shell-escaped version of the string str. This returns a string that can safely be used as one token in a shell command line.

#### Parameters

| name  | type     | description                    |
|-------|----------|--------------------------------|
| `str` | `string` | A string which is to be quoted |

#### Examples

**String**

Quote a string.

```python
load("string", "quote")
s = "Hello World"
print(quote(s))
# Output: '"Hello World"'
```

### `unquote(str) string`

Returns a shell-unescaped version of the string str. This returns a string that was used as one token in a shell command line.

#### Parameters

| name  | type     | description                      |
|-------|----------|----------------------------------|
| `str` | `string` | A string which is to be unquoted |

#### Examples

**String**

Unquote a string.

```python
load("string", "unquote")
s = '"Hello World"'
print(unquote(s))
# Output: 'Hello World'
```
