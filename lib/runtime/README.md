# runtime

`runtime` is a Starlark module provides Go and app runtime information.

## Functions

### `uptime()`

Returns the uptime of the current process in `time.duration`.

#### Examples

**basic**

Returns the uptime of the current process immediately.

```python
load("runtime", "uptime")
print(uptime())
# Output: 883.583µs
```
