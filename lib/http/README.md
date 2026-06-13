# http

`http` is an HTTP **client** for Starlark scripts: a thin wrapper around Go's `net/http`, shaped after Python's `requests`. **Capability profile: Network** — every request function performs a real outbound HTTP request, so this module has network side effects.

Every request function (and `call`) has a `try_`-prefixed twin (`try_get`, `try_post`, …, `try_call`) that never aborts the script: it returns a `(response, error)` tuple where exactly one side is `None`, the same shape as the `json` module's `try_*` functions. All request functions also accept `raise_for_status=True` to turn any non-2xx response into an error.

> Note: `postForm` / `try_postForm` are **not** snake_case. The name is kept for historical compatibility (it is `post` with the form encoding forced to `application/x-www-form-urlencoded`).

The server-side helpers (`ExportedServerRequest`, `ServerResponse`) are Go types this package exposes for embedding scripts inside a Go HTTP handler; they are documented under [Types](#types) but are not part of the loadable `http` module surface.

## Functions

| function | description |
|----------|-------------|
| `call(method, url, *, params=None, headers=None, body=None, json_body=None, form_body=None, form_encoding="", auth=(), timeout=30, allow_redirects=True, verify=True, raise_for_status=False) -> response` / `try_call(...) -> (response, error)` | Perform a request with the HTTP method named by `method` (case-insensitive), dispatching to one of the verb functions below. |
| `get(url, ...) -> response` / `try_get(...) -> (response, error)` | Perform an HTTP GET request. |
| `put(url, ...) -> response` / `try_put(...) -> (response, error)` | Perform an HTTP PUT request. |
| `post(url, ...) -> response` / `try_post(...) -> (response, error)` | Perform an HTTP POST request. |
| `postForm(url, ...) -> response` / `try_postForm(...) -> (response, error)` | POST with `form_encoding` forced to `application/x-www-form-urlencoded`. Non-snake_case name. |
| `delete(url, ...) -> response` / `try_delete(...) -> (response, error)` | Perform an HTTP DELETE request. |
| `head(url, ...) -> response` / `try_head(...) -> (response, error)` | Perform an HTTP HEAD request (response body is empty). |
| `patch(url, ...) -> response` / `try_patch(...) -> (response, error)` | Perform an HTTP PATCH request. |
| `options(url, ...) -> response` / `try_options(...) -> (response, error)` | Perform an HTTP OPTIONS request. |
| `set_timeout(timeout)` | Set the default request timeout (seconds) for this module instance. |
| `get_timeout() -> float` | Return the current default request timeout (seconds) of this module instance. |

Every verb function (`get`, `put`, `post`, `postForm`, `delete`, `head`, `patch`, `options`) and `call` share the same keyword parameters; see [Request parameters](#request-parameters).

## Types

### `response`

The result of performing an HTTP request (a struct).

**Attributes**

| attribute | type | description |
|-----------|------|-------------|
| `url` | `string` | the URL that was ultimately requested (may differ from the input after redirects). |
| `status_code` | `int` | response status code (e.g. `200`). |
| `ok` | `bool` | `True` when `status_code` is in the 2xx range. |
| `headers` | `dict` | response headers; each value is the header's values joined by `,`. |
| `encoding` | `string` | transfer encoding(s) joined by `,` (empty when none). |

**Methods**

| method | description |
|--------|-------------|
| `body() -> string` | Read and return the whole response body as a string. Re-readable. |
| `json() -> object` | Parse the body as JSON; returns `None` when the body is empty or not valid JSON (a parse failure and a JSON `null` are indistinguishable — use `try_json` to tell them apart). |
| `try_body() -> (string, error)` | Like `body()` but returns a `(value, error)` pair instead of aborting (e.g. when the response-size limit is exceeded). |
| `try_json() -> (object, error)` | Like `json()` but a parse or read failure lands in the error slot instead of folding into `None`. |

### `ExportedServerRequest`

Go-side helper (constructed via `NewExportedServerRequest` / `ConvertServerRequest`) that exposes an incoming `http.Request` to a script as a **read-only** struct. Not part of the loadable `http` module; passed in by the host.

**Attributes**

| attribute | type | description |
|-----------|------|-------------|
| `method` | `string` | the HTTP method (e.g. `GET`, `POST`). |
| `url` | `string` | the request URL. |
| `proto` | `string` | the protocol (e.g. `HTTP/1.1`). |
| `host` | `string` | the request host. |
| `remote` | `string` | the client's remote address. |
| `headers` | `dict` | request headers (each value a list of strings). |
| `query` | `dict` | parsed query parameters (each value a list of strings). |
| `encoding` | `list` | transfer encodings specified in the request. |
| `body` | `string` | the raw request body. |
| `json` | `object` | the body parsed as JSON, or `None` if empty or invalid. |

### `ServerResponse`

Go-side helper (constructed via `NewServerResponse`) that lets a script build an HTTP response the host later writes to an `http.ResponseWriter`. Not part of the loadable `http` module; passed in by the host.

**Methods**

| method | description |
|--------|-------------|
| `set_status(code)` | Set the HTTP status code (must be 100–599). |
| `set_code(code)` | Alias for `set_status`. |
| `add_header(key, value)` | Append a header value under `key`. |
| `set_content_type(content_type)` | Set the `Content-Type` header, overriding any implicit one. |
| `set_data(data)` | Set the body as binary; implies `Content-Type: application/octet-stream`. |
| `set_json(data)` | Marshal a Starlark value to JSON and set it as the body; implies `Content-Type: application/json`. |
| `set_text(data)` | Set the body as plain text; implies `Content-Type: text/plain`. |
| `set_html(data)` | Set the body as HTML; implies `Content-Type: text/html`. |

## Details & examples

### Request parameters

All verb functions and `call` accept the same parameters (for `call`, `method` is an extra first positional argument). Only `url` is required.

| name | type | description |
|------|------|-------------|
| `url` | `string` | URL to request. |
| `params` | `dict` | optional. URL query parameters to append; values must be strings. |
| `headers` | `dict` | optional. headers to add; values must be strings. |
| `body` | `string`/`bytes` | optional. raw request body. |
| `json_body` | `any` | optional. JSON-serializable value sent with `Content-Type: application/json`. |
| `form_body` | `dict` | optional. values encoded as form data; a value is either a string (a field) or a two-element list/tuple `[filename, content]` (a file). |
| `form_encoding` | `string` | optional. `application/x-www-form-urlencoded` or `multipart/form-data`; inferred when omitted (multipart if any file is present, otherwise urlencoded). |
| `auth` | `tuple` | optional. `(username, password)` for HTTP Basic auth. |
| `timeout` | `float` | optional. seconds to wait before giving up; `0` means no timeout. Defaults to the instance timeout (30). |
| `allow_redirects` | `bool` | optional. whether to follow redirects (default `True`). |
| `verify` | `bool` | optional. whether to verify the server's TLS certificate (default `True`). |
| `raise_for_status` | `bool` | optional. if `True`, a non-2xx response is reported as an error (default `False`). |

**Errors on:** a non-string `url`; a non-string `params`/`headers` value; an `auth` tuple that is not length 2; a `form_body` value that is neither a string nor a `(filename, content)` pair (e.g. `got: "int"`); supplying more than one of `body`/`json_body`/`form_body` (`body, json_body and form_body are mutually exclusive`); a JSON-unserializable `json_body`; a transport failure (connection refused, DNS, TLS); `raise_for_status=True` with a non-2xx response; `verify=False` when the host forces TLS verification; or passing `timeout`/`allow_redirects`/`verify` when the host injected its own client.

```python
load('http', 'get')
res = get(test_server_url, params={"a": "b", "c": "d"})
print(res.url)
print(res.status_code)
print(res.body())
print(res.json())
# Output:
# http://127.0.0.1:PORT?a=b&c=d
# 200
# {"hello":"world"}
# {"hello": "world"}
```

(The server in the test returns `{"hello":"world"}`; `test_server_url` is the test server's base URL.)

#### POST with a JSON body

`json_body` is marshaled to JSON and sent with `Content-Type: application/json`.

```python
load('http', 'post')
res = post(test_server_url, json_body={"a": "b", "c": "d"})
b = res.body()           # the echo server returns the raw request it received
print(res.status_code)
print('application/json' in b)
print('{"a":"b","c":"d"}' in b)
# Output:
# 200
# True
# True
```

#### POST form data and files

A string value becomes a form field; a `[filename, content]` pair becomes a file. With files present (or `form_encoding="multipart/form-data"`) the request is multipart; otherwise it is `application/x-www-form-urlencoded`.

```python
load('http', 'post')
res = post(test_server_url, form_body={
    "a": ["better.txt", "123456"],
    "b": ["dance.md", '"abcdef(@!'],
})
rb = res.body()
print(res.status_code)
print('multipart/form-data; boundary=' in rb)
print('filename="better.txt"' in rb)
# Output:
# 200
# True
# True
```

### `call` / `try_call`

`call(method, url, ...)` dispatches to the verb function named by `method` (case-insensitive). The supported methods are `get`, `put`, `post`, `postForm`, `delete`, `head`, `patch`, `options`.

**Errors on:** a missing method name (`http.call: missing method name`); a non-string method name; or an unsupported method (`unsupported method: <name>`).

```python
load('http', 'call')
res = call('POST', test_server_url, params={"hello": "world"}, json_body={"a": "b", "c": "d"})
b = res.body()
print(res.status_code)
print('/?hello=world' in b)
print('{"a":"b","c":"d"}' in b)
# Output:
# 200
# True
# True
```

### `try_*` variants

A `try_` function returns `(response, error)` with the Go error always `nil`: on success the error slot is `None`; on failure the response slot is `None` and the error slot holds the message string. Argument-unpacking and dispatch errors are captured the same way.

```python
load('http', 'try_get', 'try_call')
# transport failure is captured, not raised
res, err = try_get('http://127.0.0.1:1/')
print(res == None)
print('connect' in err or 'refused' in err)
# an unsupported method is captured too
res2, err2 = try_call('TRACE', test_server_url)
print(res2 == None)
print('unsupported method' in err2)
# Output:
# True
# True
# True
# True
```

### `raise_for_status`

By default a non-2xx response is returned normally (`res.ok` is `False`); with `raise_for_status=True` it becomes an error.

```python
load('http', 'get')
res = get(nf_url)            # server replies 404
print(res.ok)
print(res.status_code)
# Output:
# False
# 404
```

```python
load('http', 'get')
get(nf_url, raise_for_status=True)
# Error: http.get: unexpected status: 404 Not Found
```

### `try_json` vs `json`

`json()` folds a read/parse failure into `None`; `try_json()` surfaces it in the error slot, so a parse failure is distinguishable from a JSON `null`.

```python
load('http', 'get')
res = get(ok_url)            # server replies {"a": 1}
v, err = res.try_json()
print(err == None)
print(v)
# Output:
# True
# {"a": 1}
```

### `set_timeout` / `get_timeout`

`set_timeout(timeout)` sets the default request timeout (seconds) for **this module instance** only — it does not leak into other machines in the process; the package-level `TimeoutSecond` seeds new instances. `get_timeout()` returns the current value. With a host-injected client the value is ignored (the client's own timeout applies).

**Errors on:** a non-numeric `timeout` (`got string, want float or int`); a negative `timeout` (`timeout must be non-negative`); or passing any argument to `get_timeout()` (`got 1 arguments, want 0`).

```python
load('http', 'get_timeout', 'set_timeout')
print(get_timeout())
set_timeout(10.5)
print(get_timeout())
# Output:
# 30.0
# 10.5
```

## Notes / boundaries

- **Engine.** A thin wrapper over Go `net/http`; request/response semantics follow that package. JSON is handled by starlet's `dataconv` (Starlark-aware), so structs, `module`, `time`, and starlight-wrapped Go values marshal correctly.
- **Instance vs package state.** `set_timeout` and the host-configurable knobs (`SetClient`, `SetGuard`, `SetMaxResponseBodyBytes`, `SetForceTLSVerify`) live on the module instance; the package-level `TimeoutSecond`, `UserAgent`, `SkipInsecureVerify`, `DisableRedirect`, `MaxResponseBodyBytes`, `ForceTLSVerify`, `Client`, and `Guard` only *seed* new instances at `LoadModule` time.
- **Security knobs.** A host may force TLS verification (`verify=False` is then rejected), cap the response body size (over-limit `body()`/`json()` error with `response body exceeds the N-byte limit`), and install a `RequestGuard` to allow/deny requests by URL. When the host injects its own `*http.Client`, the per-request `timeout`/`allow_redirects`/`verify` options are rejected rather than silently ignored.
- **Body kinds are mutually exclusive.** Pass at most one of `body`, `json_body`, `form_body`; supplying more than one is an error rather than a silent drop.
- **Determinism.** Response `headers` and `encoding` join multiple values with `,`. `body()`/`json()` reset the body reader so they may be called repeatedly.
- **Difference from `requests`.** `postForm` is a non-Pythonic convenience name; `params`/`headers` values must be strings; `json()` returns `None` (not raising) on parse failure — use `try_json()` for an explicit error.
</content>
</invoke>
