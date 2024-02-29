# log

`log` provides functionality for logging messages at various severity levels.

## Functions

### `debug(msg, *misc, **kv)`

Logs a message at the debug log level.

#### Parameters

| name   | type       | description                                                                                   |
|--------|------------|-----------------------------------------------------------------------------------------------|
| `msg`  | `string`   | The message to log.                                                                           |
| `misc` | `*args`    | Additional message arguments will be concatenated to the message string separated by a space. |
| `kv`   | `**kwargs` | Key-value pairs to provide additional debug information.                                      |

#### Examples

**basic**

Log a debug message with additional information.

```python
load("log", "debug")
debug("Fetching data", retry_attempt=1)
```

### `info(msg, *misc, **kv)`

Logs a message at the info log level.

#### Parameters

| name   | type       | description                                                                                   |
|--------|------------|-----------------------------------------------------------------------------------------------|
| `msg`  | `string`   | The message to log.                                                                           |
| `misc` | `*args`    | Additional message arguments will be concatenated to the message string separated by a space. |
| `kv`   | `**kwargs` | Key-value pairs to provide additional information.                                            |

#### Examples

**basic**

Log an info message with additional information.

```python
load("log", "info")
info("Data fetched", response_time=42)
```

### `warn(msg, *misc, **kv)`

Logs a message at the warn log level.

#### Parameters

| name   | type       | description                                                                                   |
|--------|------------|-----------------------------------------------------------------------------------------------|
| `msg`  | `string`   | The message to log.                                                                           |
| `misc` | `*args`    | Additional message arguments will be concatenated to the message string separated by a space. |
| `kv`   | `**kwargs` | Key-value pairs to provide additional warning information.                                    |

#### Examples

**basic**

Log a warning message with additional information.

```python
load("log", "warn")
warn("Fetching data took longer than expected", response_time=123)
```

### `error(msg, *misc, **kv)`

Logs a message at the error log level and returns an error.

#### Parameters

| name   | type       | description                                                                                   |
|--------|------------|-----------------------------------------------------------------------------------------------|
| `msg`  | `string`   | The message to log.                                                                           |
| `misc` | `*args`    | Additional message arguments will be concatenated to the message string separated by a space. |
| `kv`   | `**kwargs` | Key-value pairs to provide additional error information.                                      |

#### Examples

**basic**

Log an error message with additional information.

```python
load("log", "error")
error("Failed to fetch data", response_time=240)
```

### `fatal(msg, *misc, **kv)`

Logs a message at the error log level, returns a `fail(msg)` to halt program execution.

#### Parameters

| name   | type       | description                                                                                   |
|--------|------------|-----------------------------------------------------------------------------------------------|
| `msg`  | `string`   | The message to log.                                                                           |
| `misc` | `*args`    | Additional message arguments will be concatenated to the message string separated by a space. |
| `kv`   | `**kwargs` | Key-value pairs to provide additional fatal error information.                                |

#### Examples

**basic**

Log a fatal error message with additional information.

```python
load("log", "fatal")
fatal("Failed to fetch data and cannot recover", retry_attempts=3, response_time=360)
```
