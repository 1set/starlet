# runtime

`runtime` is a Starlark module provides Go and app runtime information.

## Constants

- `hostname`: A string representing the hostname of the system where the script is being executed.
- `workdir`: A string representing the current working directory of the process.
- `os`: A string representing the operating system of the runtime. This value comes from Go's `runtime.GOOS`.
- `arch`: A string representing the architecture of the machine. This value is derived from Go's `runtime.GOARCH`.
- `gover`: A string representing the Go runtime version. This is obtained using `runtime.Version()` from the Go standard library.
- `pid`: An integer representing the process ID of the current process.
- `ppid`: An integer representing the parent process ID of the current process.
- `uid`: An integer representing the user ID of the process owner.
- `gid`: An integer representing the group ID of the process owner.
- `app_start`: A time value representing the moment when the application started. This is used to calculate uptime.

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
