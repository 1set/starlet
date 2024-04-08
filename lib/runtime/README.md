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

### `getenv(key, default=None)`

Returns the value of the environment variable key as a string if it exists, or default if it doesn't.

#### Examples

**basic**

Returns the value of the environment variable PATH if it exists, or None if it doesn't.

```python
load("runtime", "getenv")
print(getenv("PATH"))
# Output: /usr/local/bin:/usr/bin:/bin:/usr/sbin:/sbin
```

### `putenv(key, value)`

Sets the value of the environment variable named by the key, returning an error if any.

#### Examples

**basic**

Sets the environment variable `STARLET_TEST` to the value `123456`.

```python
load("runtime", "putenv")
putenv("STARLET_TEST", 123456)
```

### `setenv(key, value)`

Sets the value of the environment variable named by the key, returning an error if any.
Alias of `putenv`.

#### Examples

**basic**

Sets the environment variable `STARLET_TEST` to the value `ABC`.

```python
load("runtime", "setenv")
setenv("STARLET_TEST", "ABC")
```

### `unsetenv(key)`

Unsets a single environment variable.

#### Examples

**basic**

Unsets the environment variable STARLET_TEST.

```python
load("runtime", "unsetenv")
unsetenv("STARLET_TEST")
```
