# string

`string` provides constants and functions to manipulate strings, it's intended to be a drop-in subset of Python's string module for Starlark.

## Constants

- `ascii_lowercase`: A string containing all the characters that are considered lowercase letters.
- `ascii_uppercase`: A string containing all the characters that are considered uppercase letters.
- `ascii_letters`: A string containing all the characters that are considered letters.
- `digits`: A string containing all characters considered decimal digits: `0123456789`.
- `hexdigits`: A string containing all characters considered hexadecimal digits: `0123456789abcdefABCDEF`.
- `octdigits`: A string containing all characters considered octal digits: `01234567`.
- `punctuation`: A string containing all characters which are considered punctuation characters.
- `whitespace`: A string containing all characters that are considered whitespace.
- `printable`: A string containing all characters that are considered printable. This is a combination of digits, ascii_letters, punctuation, and whitespace

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

**Basic**

Reverse a string.

```python
load("string", "reverse")
s = "123我爱你"
print(reverse(s))
# Output: 你爱我321
```

### `index(s, sub) int`

Returns the index of the first occurrence of the substring `sub` in `s`. If the substring is not found, an error is raised.

#### Parameters

| name  | type     | description                 |
|-------|----------|-----------------------------|
| `s`   | `string` | The string to be searched   |
| `sub` | `string` | The substring to search for |

#### Examples

**Basic**

Find the first occurrence of a substring in a string.

```python
load("string", "index")
s = "hello world"
print(index(s, "o"))
# Output: 4
```

### `rindex(s, sub) int`

Returns the index of the last occurrence of the substring `sub` in `s`. If the substring is not found, an error is raised.

#### Parameters

| name  | type     | description                 |
|-------|----------|-----------------------------|
| `s`   | `string` | The string to be searched   |
| `sub` | `string` | The substring to search for |

#### Examples

**Basic**

Find the last occurrence of a substring in a string.

```python
load("string", "rindex")
s = "hello world"
print(rindex(s, "o"))
# Output: 7
```

### `find(s, sub) int`

Returns the index of the first occurrence of the substring `sub` in `s`. If the substring is not found, returns -1.

#### Parameters

| name  | type     | description                 |
|-------|----------|-----------------------------|
| `s`   | `string` | The string to be searched   |
| `sub` | `string` | The substring to search for |

#### Examples

**Basic**

Find the first occurrence of a substring in a string, returning -1 if not found.

```python
load("string", "find")
s = "hello world"
print(find(s, "o"))
print(find(s, "x"))
# Output: 4
# Output: -1
```

### `rfind(s, sub) int`

Returns the index of the last occurrence of the substring `sub` in `s`. If the substring is not found, returns -1.

#### Parameters

| name  | type     | description                 |
|-------|----------|-----------------------------|
| `s`   | `string` | The string to be searched   |
| `sub` | `string` | The substring to search for |

#### Examples

**Basic**

Find the last occurrence of a substring in a string, returning -1 if not found.

```python
load("string", "rfind")
s = "hello world"
print(rfind(s, "o"))
print(rfind(s, "x"))
# Output: 7
# Output: -1
```

### `substring(s, start, end=None) string`

Returns a substring of `s` from index `start` to `end` (exclusive). If `end` is omitted or `None`, the substring extends to the end of the string. An explicit `end` of `0` (or `-len(s)`) means an empty range, as in Python slicing; out-of-range indices are reported as errors.

#### Parameters

| name    | type     | description                          |
|---------|----------|--------------------------------------|
| `s`     | `string` | The string to be sliced              |
| `start` | `int`    | The starting index for the substring |
| `end`   | `int` or `None` | Optional ending index (exclusive); defaults to the end of the string |

#### Examples

**Basic**

Get a substring of a string.

```python
load("string", "substring")
s = "hello world"
print(substring(s, 1, 5))
# Output: "ello"
```

**Negative Indices**

Get a substring of a string using negative indices.

```python
load("string", "substring")
s = "hello world"
print(substring(s, -5, -1))
# Output: "worl"
```

### `codepoint(s, index) string`

Returns the Unicode codepoint of the character at the given `index` in `s`.

#### Parameters

| name    | type     | description                                |
|---------|----------|--------------------------------------------|
| `s`     | `string` | The string from which to get the codepoint |
| `index` | `int`    | The index of the character                 |

#### Examples

**Basic**

Get the Unicode codepoint of a character at a specific index.

```python
load("string", "codepoint")
s = "hello world"
print(codepoint(s, 4))
# Output: "o"
```

### `escape(str) string`

Converts the characters "&", "<", ">", '"' and "'" in string to their corresponding HTML entities.

#### Parameters

| name  | type     | description                          |
|-------|----------|--------------------------------------|
| `str` | `string` | A string which is to be HTML escaped |

#### Examples

**Basic**

Escape a string.

```python
load("string", "escape")
s = "Hello<World>"
print(escape(s))
# Output: Hello&lt;World&gt;
```

### `unescape(str) string`

Converts the HTML entities in a string back to their corresponding characters.

#### Parameters

| name  | type     | description           |
|-------|----------|-----------------------|
| `str` | `string` | A HTML escaped string |

#### Examples

**Basic**

Unescape a string.

```python
load("string", "unescape")
s = "You&amp;Me"
print(unescape(s))
# Output: "You&Me"
```

### `quote(str) string`

Returns a double-quoted string literal in Go syntax representing str, with control characters and non-printable runes escaped (like Go's `strconv.Quote`). NOTE: this is **not** shell escaping — do not use it to build shell command lines.

#### Parameters

| name  | type     | description                    |
|-------|----------|--------------------------------|
| `str` | `string` | A string which is to be quoted |

#### Examples

**Basic**

Quote a string.

```python
load("string", "quote")
s = "Hello World"
print(quote(s))
# Output: "Hello World"
```

### `unquote(str) string`

Returns a shell-unescaped version of the string str. This returns a string that was used as one token in a shell command line.

#### Parameters

| name  | type     | description                      |
|-------|----------|----------------------------------|
| `str` | `string` | A string which is to be unquoted |

#### Examples

**Basic**

Unquote a string.

```python
load("string", "unquote")
s = '"Hello\tWorld"'
print(unquote(s))
World
```

### `head(s, n) string`

Returns the first `n` characters (runes) of `s`. `n` must be non-negative and is clamped at the length of the string, so a short input never errors. Unlike the language-level `s[:n]`, the cut is rune-aware and never splits a multi-byte character.

### `tail(s, n) string`

Returns the last `n` characters (runes) of `s`, with the same clamping and rune-awareness as `head`.

### `head_lines(s, n) string`

Returns the first `n` lines of `s` (split on `\n`), clamped at the number of lines.

### `tail_lines(s, n) string`

Returns the last `n` lines of `s` (split on `\n`), clamped at the number of lines.

### `truncate(s, length, suffix="...") string`

Shortens `s` to at most `length` characters (runes). If a cut happens, `suffix` is appended and the result — including the suffix — never exceeds `length` runes; a string already within the limit is returned unchanged.
