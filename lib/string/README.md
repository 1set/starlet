# string

`string` provides constants and functions manipulate strings, it's intended to be a drop-in subset of Python's string module for Starlark.

## Functions

### `length(obj) int`

Returns the length of the object, for string it returns the number of Unicode code points, instead of bytes like `len()`.

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
