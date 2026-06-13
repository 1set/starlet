# string

`string` provides constants and functions for manipulating strings — length, reversal, searching, slicing, HTML escape/unescape, Go-syntax quote/unquote, and rune-aware head/tail/truncate helpers. It is intended to be a drop-in subset of [Python's `string` module](https://docs.python.org/3/library/string.html) for Starlark, extended with a few utilities of its own. Capability profile: **Pure** — no filesystem, network, process, or log side effects.

## Functions

| function | description |
|----------|-------------|
| `length(obj) -> int` | Number of Unicode code points in a string (bytes for `bytes`, element count for any sequence) |
| `reverse(s) -> string` | The value reversed (by rune for strings, by byte for `bytes`) |
| `index(s, sub) -> int` | Rune index of the first occurrence of `sub`; errors if not found |
| `rindex(s, sub) -> int` | Rune index of the last occurrence of `sub`; errors if not found |
| `find(s, sub) -> int` | Rune index of the first occurrence of `sub`, or `-1` if not found |
| `rfind(s, sub) -> int` | Rune index of the last occurrence of `sub`, or `-1` if not found |
| `substring(s, start, end=None) -> string` | Rune slice `[start:end)`; `end` defaults to the end of `s`; supports negative indices |
| `codepoint(s, index) -> string` | The single character (code point) at rune `index`; supports a negative index |
| `head(s, n) -> string` | First `n` runes of `s`, clamped to its length |
| `tail(s, n) -> string` | Last `n` runes of `s`, clamped to its length |
| `head_lines(s, n) -> string` | First `n` lines of `s` (split on `\n`), clamped to the line count |
| `tail_lines(s, n) -> string` | Last `n` lines of `s` (split on `\n`), clamped to the line count |
| `truncate(s, length, suffix="...") -> string` | Shorten `s` to at most `length` runes, appending `suffix` when cut |
| `escape(s) -> string` | HTML-escape `&`, `<`, `>`, `"`, `'` |
| `unescape(s) -> string` | Reverse of `escape`: HTML entities back to characters |
| `quote(s) -> string` | Go-syntax double-quoted string literal (like `strconv.Quote`) |
| `unquote(s) -> string` | Reverse of `quote`; returns the input unchanged if it is not a valid quoted literal |

## Constants

Each constant is a `string` (matching Python's `string` module names).

| constant | meaning |
|----------|---------|
| `ascii_lowercase` | `abcdefghijklmnopqrstuvwxyz` |
| `ascii_uppercase` | `ABCDEFGHIJKLMNOPQRSTUVWXYZ` |
| `ascii_letters` | `ascii_lowercase` + `ascii_uppercase` |
| `digits` | decimal digits `0123456789` |
| `hexdigits` | hexadecimal digits `0123456789abcdefABCDEF` |
| `octdigits` | octal digits `01234567` |
| `punctuation` | ASCII punctuation: `` !"#$%&'()*+,-./:;<=>?@[\]^_{|}~` `` |
| `whitespace` | whitespace characters: space, `\t`, `\n`, `\r`, `\v`, `\f` |
| `printable` | `digits` + `ascii_letters` + `punctuation` + `whitespace` |

This module exposes no custom Starlark types — only the functions and constants above.

## Details & examples

### `length(obj) -> int`

Returns the number of Unicode code points in a `string` (not bytes, unlike the built-in `len()`), the number of bytes in a `bytes` value, or the element count of any `starlark.Sequence` (list, tuple, set). Errors with `length() takes exactly one argument` when not given exactly one argument, and `length() function isn't supported for '<type>' type object` for an unsupported type (e.g. `int`).

```python
load("string", "length")
print(length("我爱你"), length(b"☕"), length([1, 2, "#", True, None]))
# Output: 3 3 5
```

### `reverse(s) -> string`

Reverses a `string` by rune (so multi-byte characters stay intact) or a `bytes` value by byte. Same single-argument and type errors as `length`.

```python
load("string", "reverse")
print(reverse("123我爱你"))
# Output: 你爱我321
```

### `index(s, sub) -> int` / `rindex(s, sub) -> int`

Return the rune index of the first (`index`) or last (`rindex`) occurrence of `sub` in `s`. They error with `<name>: substring not found` when `sub` is absent. Indices are counted in runes, so they are correct for multi-byte text.

```python
load("string", "index", "rindex")
print(index("你好世界", "好"), rindex("你好世界你好", "好"))
# Output: 1 5
```

### `find(s, sub) -> int` / `rfind(s, sub) -> int`

Like `index` / `rindex`, but return `-1` instead of erroring when `sub` is not found.

```python
load("string", "find", "rfind")
print(find("hello", "o"), find("hello", "x"), rfind("hello hello", "o"))
# Output: 4 -1 10
```

### `substring(s, start, end=None) -> string`

Returns the rune slice `s[start:end)`. `end` defaults to the end of `s` when omitted or `None`; an explicit `end` of `0` (or `-len(s)`) means an empty range, as in Python slicing. Negative `start`/`end` count from the end. Errors with `substring: indices are out of range` if, after normalization, an index falls outside `[0, len(s)]` or `start > end`.

```python
load("string", "substring")
print(substring("hello", 1, 4), substring("你好世界", 2, -1), substring("hello", 1))
# Output: ell 世 ello
```

### `codepoint(s, index) -> string`

Returns the single character at rune `index` (negative indices count from the end). Errors with `codepoint: index out of range` when `index` is outside the string.

```python
load("string", "codepoint")
print(codepoint("a☕c", 1), codepoint("a☕c", -1))
# Output: ☕ c
```

### `head(s, n) -> string` / `tail(s, n) -> string`

Return the first (`head`) or last (`tail`) `n` runes of `s`. `n` must be non-negative (else `<name>: n must be non-negative`) and is clamped to the string length, so a short input never errors. The cut is rune-aware and never splits a multi-byte character.

```python
load("string", "head", "tail")
print(head("你好世界", 2), tail("a☕c", 2), head("hello", 99))
# Output: 你好 ☕c hello
```

### `head_lines(s, n) -> string` / `tail_lines(s, n) -> string`

Return the first (`head_lines`) or last (`tail_lines`) `n` lines of `s`, splitting on `\n` and clamping `n` to the line count. `n` must be non-negative.

```python
load("string", "head_lines", "tail_lines")
s = "a\nb\nc"
print(head_lines(s, 2))
print(tail_lines(s, 2))
# Output:
# a
# b
# b
# c
```

### `truncate(s, length, suffix="...") -> string`

Shortens `s` to at most `length` runes. A string already within the limit is returned unchanged; otherwise `suffix` is appended and the result — including the suffix — never exceeds `length` runes (if `length` is shorter than `suffix`, the suffix itself is truncated). `length` must be non-negative (else `truncate: length must be non-negative`).

```python
load("string", "truncate")
print(truncate("hello world", 8))
print(truncate("hello world", 8, suffix="~"))
print(truncate("hello", 2))
# Output:
# hello...
# hello w~
# ..
```

### `escape(s) -> string` / `unescape(s) -> string`

`escape` converts `&`, `<`, `>`, `"`, and `'` to their HTML entities; `unescape` is the inverse. Both accept a `string` or `bytes` (exactly one argument) and preserve the input type. Same single-argument and type errors as `length`.

```python
load("string", "escape", "unescape")
print(escape("<&>"))
print(unescape("我&amp;你"))
# Output:
# &lt;&amp;&gt;
# 我&你
```

### `quote(s) -> string` / `unquote(s) -> string`

`quote` returns a Go-syntax double-quoted string literal for `s`, escaping control characters and non-printable runes (like Go's `strconv.Quote`). NOTE: this is **not** shell escaping — do not use it to build shell command lines. `unquote` reverses it: it unquotes a Go double-quoted literal, and robustly returns the input unchanged if it is not a valid quoted literal (e.g. shorter than 2 characters, or only one side quoted). Both accept a `string` or `bytes`.

```python
load("string", "quote", "unquote")
print(quote("\n1"))
print(unquote('"我爱你"'), unquote('"我爱你'))
# Output:
# "\n1"
# 我爱你 "我爱你
```

## Notes / boundaries

- **Rune-based indexing.** `length`, `index`/`rindex`/`find`/`rfind`, `substring`, `codepoint`, `head`/`tail`, and `truncate` all operate on Unicode code points, not bytes — counts and slices never split a multi-byte character. This differs from the built-in `len()` and the `s[a:b]` slice operator, which work on bytes.
- **`bytes` support.** Only `length`, `reverse`, `escape`, `unescape`, `quote`, and `unquote` accept `bytes` (and preserve the type). The index/slice/line helpers take `string` only.
- **Pure module.** No I/O, no global state, deterministic output for a given input.
- **Python parity.** The constants mirror Python's `string` module; `index`/`rindex`/`find`/`rfind` mirror the `str` methods of the same names. `quote`/`unquote`/`escape`/`unescape` and the `head`/`tail`/`*_lines`/`truncate`/`substring`/`codepoint` helpers are starlet extensions, not part of Python's `string` module.
