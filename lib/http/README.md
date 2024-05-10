# http

`http` defines an HTTP client implementation. It is a thin wrapper around the Go standard package `net/http` but in Python `requests` style.

## Functions

### `call(method, url, params=None, headers=None, auth=(), body=None, json_body=None, form_body=None, form_encoding="", timeout=30, allow_redirects=True, verify=True) response`

Perform an HTTP request of the specified method, returning a response.
The `call` method allows for flexibility in making HTTP requests by specifying the HTTP method as an argument.
It supports all common HTTP methods. This method dynamically dispatches the request based on the provided method name.
It is a convenience wrapper that enables users to use any supported HTTP method without needing separate method calls for each type of request.

#### Parameters

| name              | type     | description                                                                                                                                                   |
|-------------------|----------|---------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `method`          | `string` | The HTTP method to use for the request (e.g., GET, POST, PUT, DELETE).                                                                                        |
| `url`             | `string` | URL to request.                                                                                                                                               |
| `params`          | `dict`   | optional. dictionary of URL parameters to append to the request.                                                                                              |
| `headers`         | `dict`   | optional. dictionary of headers to add to request.                                                                                                            |
| `body`            | `string` | optional. raw string body to provide to the request.                                                                                                          |
| `form_body`       | `dict`   | optional. dict of values that will be encoded as form data. the value can be a string or a list of two strings (filename, file content) for file attachments. |
| `form_encoding`   | `string` | optional. `application/x-www-form-urlencoded` (default for form data) or `multipart/form-data`.                                                               |
| `json_body`       | `any`    | optional. JSON data to supply as a request. handy for working with JSON-API's.                                                                                |
| `auth`            | `tuple`  | optional. (username,password) tuple for HTTP Basic authorization.                                                                                             |
| `timeout`         | `float`  | optional. how many seconds to wait for the server to send all the data before giving up. 0 means no timeout.                                                  |
| `allow_redirects` | `bool`   | optional. whether to follow redirects.                                                                                                                        |
| `verify`          | `bool`   | optional. whether to verify the server's SSL certificate.                                                                                                     |

### `get(url, params=None, headers=None, auth=(), body=None, json_body=None, form_body=None, form_encoding="", timeout=30, allow_redirects=True, verify=True) response`

Perform an HTTP GET request, returning a response.

#### Parameters

| name              | type     | description                                                                                                                                                                                                        |
|-------------------|----------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `url`             | `string` | URL to request.                                                                                                                                                                                                    |
| `params`          | `dict`   | optional. dictionary of URL parameters to append to the request.                                                                                                                                                   |
| `headers`         | `dict`   | optional. dictionary of headers to add to request.                                                                                                                                                                 |
| `body`            | `string` | optional. raw string body to provide to the request.                                                                                                                                                               |
| `json_body`       | `any`    | optional. considered only if body is None. JSON data to supply as a request. handy for working with JSON-API's.                                                                                                    |
| `form_body`       | `dict`   | optional. considered only if both body and json_body are None. dict of values that will be encoded as form data. the value can be a string or a list of two strings (filename, file content) for file attachments. |
| `form_encoding`   | `string` | optional. `application/x-www-form-urlencoded` (default for form data) or `multipart/form-data`.                                                                                                                    |
| `auth`            | `tuple`  | optional. (username,password) tuple for HTTP Basic authorization.                                                                                                                                                  |
| `timeout`         | `float`  | optional. how many seconds to wait for the server to send all the data before giving up. 0 means no timeout.                                                                                                       |
| `allow_redirects` | `bool`   | optional. whether to follow redirects.                                                                                                                                                                             |
| `verify`          | `bool`   | optional. whether to verify the server's SSL certificate.                                                                                                                                                          |

### `put(url, params=None, headers=None, auth=(), body=None, json_body=None, form_body=None, form_encoding="", timeout=30, allow_redirects=True, verify=True) response`

Perform an HTTP PUT request, returning a response.

#### Parameters

| name              | type     | description                                                                                                                                                                                                        |
|-------------------|----------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `url`             | `string` | URL to request.                                                                                                                                                                                                    |
| `params`          | `dict`   | optional. dictionary of URL parameters to append to the request.                                                                                                                                                   |
| `headers`         | `dict`   | optional. dictionary of headers to add to request.                                                                                                                                                                 |
| `body`            | `string` | optional. raw string body to provide to the request.                                                                                                                                                               |
| `json_body`       | `any`    | optional. considered only if body is None. JSON data to supply as a request. handy for working with JSON-API's.                                                                                                    |
| `form_body`       | `dict`   | optional. considered only if both body and json_body are None. dict of values that will be encoded as form data. the value can be a string or a list of two strings (filename, file content) for file attachments. |
| `form_encoding`   | `string` | optional. `application/x-www-form-urlencoded` (default for form data) or `multipart/form-data`.                                                                                                                    |
| `auth`            | `tuple`  | optional. (username,password) tuple for HTTP Basic authorization.                                                                                                                                                  |
| `timeout`         | `float`  | optional. how many seconds to wait for the server to send all the data before giving up. 0 means no timeout.                                                                                                       |
| `allow_redirects` | `bool`   | optional. whether to follow redirects.                                                                                                                                                                             |
| `verify`          | `bool`   | optional. whether to verify the server's SSL certificate.                                                                                                                                                          |

### `post(url, params=None, headers=None, auth=(), body=None, json_body=None, form_body=None, form_encoding="", timeout=30, allow_redirects=True, verify=True) response`

Perform an HTTP POST request, returning a response.

#### Parameters

| name              | type     | description                                                                                                                                                                                                        |
|-------------------|----------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `url`             | `string` | URL to request.                                                                                                                                                                                                    |
| `params`          | `dict`   | optional. dictionary of URL parameters to append to the request.                                                                                                                                                   |
| `headers`         | `dict`   | optional. dictionary of headers to add to request.                                                                                                                                                                 |
| `body`            | `string` | optional. raw string body to provide to the request.                                                                                                                                                               |
| `json_body`       | `any`    | optional. considered only if body is None. JSON data to supply as a request. handy for working with JSON-API's.                                                                                                    |
| `form_body`       | `dict`   | optional. considered only if both body and json_body are None. dict of values that will be encoded as form data. the value can be a string or a list of two strings (filename, file content) for file attachments. |
| `form_encoding`   | `string` | optional. `application/x-www-form-urlencoded` (default for form data) or `multipart/form-data`.                                                                                                                    |
| `auth`            | `tuple`  | optional. (username,password) tuple for HTTP Basic authorization.                                                                                                                                                  |
| `timeout`         | `float`  | optional. how many seconds to wait for the server to send all the data before giving up. 0 means no timeout.                                                                                                       |
| `allow_redirects` | `bool`   | optional. whether to follow redirects.                                                                                                                                                                             |
| `verify`          | `bool`   | optional. whether to verify the server's SSL certificate.                                                                                                                                                          |

### `postForm(url, params=None, headers=None, auth=(), body=None, json_body=None, form_body=None, form_encoding="", timeout=30, allow_redirects=True, verify=True) response`

Perform an HTTP POST request with form data, returning a response.

#### Parameters

| name              | type     | description                                                                                                                                                                                                        |
|-------------------|----------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `url`             | `string` | URL to request.                                                                                                                                                                                                    |
| `params`          | `dict`   | optional. dictionary of URL parameters to append to the request.                                                                                                                                                   |
| `headers`         | `dict`   | optional. dictionary of headers to add to request.                                                                                                                                                                 |
| `body`            | `string` | optional. raw string body to provide to the request.                                                                                                                                                               |
| `json_body`       | `any`    | optional. considered only if body is None. JSON data to supply as a request. handy for working with JSON-API's.                                                                                                    |
| `form_body`       | `dict`   | optional. considered only if both body and json_body are None. dict of values that will be encoded as form data. the value can be a string or a list of two strings (filename, file content) for file attachments. |
| `form_encoding`   | `string` | optional. `application/x-www-form-urlencoded` (default for form data) or `multipart/form-data`.                                                                                                                    |
| `auth`            | `tuple`  | optional. (username,password) tuple for HTTP Basic authorization.                                                                                                                                                  |
| `timeout`         | `float`  | optional. how many seconds to wait for the server to send all the data before giving up. 0 means no timeout.                                                                                                       |
| `allow_redirects` | `bool`   | optional. whether to follow redirects.                                                                                                                                                                             |
| `verify`          | `bool`   | optional. whether to verify the server's SSL certificate.                                                                                                                                                          |

### `delete(url, params=None, headers=None, auth=(), body=None, json_body=None, form_body=None, form_encoding="", timeout=30, allow_redirects=True, verify=True) response`

Perform an HTTP DELETE request, returning a response.

#### Parameters

| name              | type     | description                                                                                                                                                                                                        |
|-------------------|----------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `url`             | `string` | URL to request.                                                                                                                                                                                                    |
| `params`          | `dict`   | optional. dictionary of URL parameters to append to the request.                                                                                                                                                   |
| `headers`         | `dict`   | optional. dictionary of headers to add to request.                                                                                                                                                                 |
| `body`            | `string` | optional. raw string body to provide to the request.                                                                                                                                                               |
| `json_body`       | `any`    | optional. considered only if body is None. JSON data to supply as a request. handy for working with JSON-API's.                                                                                                    |
| `form_body`       | `dict`   | optional. considered only if both body and json_body are None. dict of values that will be encoded as form data. the value can be a string or a list of two strings (filename, file content) for file attachments. |
| `form_encoding`   | `string` | optional. `application/x-www-form-urlencoded` (default for form data) or `multipart/form-data`.                                                                                                                    |
| `auth`            | `tuple`  | optional. (username,password) tuple for HTTP Basic authorization.                                                                                                                                                  |
| `timeout`         | `float`  | optional. how many seconds to wait for the server to send all the data before giving up. 0 means no timeout.                                                                                                       |
| `allow_redirects` | `bool`   | optional. whether to follow redirects.                                                                                                                                                                             |
| `verify`          | `bool`   | optional. whether to verify the server's SSL certificate.                                                                                                                                                          |

### `patch(url, params=None, headers=None, auth=(), body=None, json_body=None, form_body=None, form_encoding="", timeout=30, allow_redirects=True, verify=True) response`

Perform an HTTP PATCH request, returning a response.

#### Parameters

| name              | type     | description                                                                                                                                                                                                        |
|-------------------|----------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `url`             | `string` | URL to request.                                                                                                                                                                                                    |
| `params`          | `dict`   | optional. dictionary of URL parameters to append to the request.                                                                                                                                                   |
| `headers`         | `dict`   | optional. dictionary of headers to add to request.                                                                                                                                                                 |
| `body`            | `string` | optional. raw string body to provide to the request.                                                                                                                                                               |
| `json_body`       | `any`    | optional. considered only if body is None. JSON data to supply as a request. handy for working with JSON-API's.                                                                                                    |
| `form_body`       | `dict`   | optional. considered only if both body and json_body are None. dict of values that will be encoded as form data. the value can be a string or a list of two strings (filename, file content) for file attachments. |
| `form_encoding`   | `string` | optional. `application/x-www-form-urlencoded` (default for form data) or `multipart/form-data`.                                                                                                                    |
| `auth`            | `tuple`  | optional. (username,password) tuple for HTTP Basic authorization.                                                                                                                                                  |
| `timeout`         | `float`  | optional. how many seconds to wait for the server to send all the data before giving up. 0 means no timeout.                                                                                                       |
| `allow_redirects` | `bool`   | optional. whether to follow redirects.                                                                                                                                                                             |
| `verify`          | `bool`   | optional. whether to verify the server's SSL certificate.                                                                                                                                                          |

### `options(url,params={},headers={},body="",form_body={},form_encoding="",json_body={},auth=(),timeout=30,allow_redirects=True,verify=True) response`

Perform an HTTP OPTIONS request, returning a response.

#### Parameters

| name              | type     | description                                                                                                                                                                                                        |
|-------------------|----------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `url`             | `string` | URL to request.                                                                                                                                                                                                    |
| `params`          | `dict`   | optional. dictionary of URL parameters to append to the request.                                                                                                                                                   |
| `headers`         | `dict`   | optional. dictionary of headers to add to request.                                                                                                                                                                 |
| `body`            | `string` | optional. raw string body to provide to the request.                                                                                                                                                               |
| `json_body`       | `any`    | optional. considered only if body is None. JSON data to supply as a request. handy for working with JSON-API's.                                                                                                    |
| `form_body`       | `dict`   | optional. considered only if both body and json_body are None. dict of values that will be encoded as form data. the value can be a string or a list of two strings (filename, file content) for file attachments. |
| `form_encoding`   | `string` | optional. `application/x-www-form-urlencoded` (default for form data) or `multipart/form-data`.                                                                                                                    |
| `auth`            | `tuple`  | optional. (username,password) tuple for HTTP Basic authorization.                                                                                                                                                  |
| `timeout`         | `float`  | optional. how many seconds to wait for the server to send all the data before giving up. 0 means no timeout.                                                                                                       |
| `allow_redirects` | `bool`   | optional. whether to follow redirects.                                                                                                                                                                             |
| `verify`          | `bool`   | optional. whether to verify the server's SSL certificate.                                                                                                                                                          |

### `set_timeout(timeout)`

Set the global timeout for all HTTP requests.

#### Parameters

| name      | type    | description                                                                                                                  |
|-----------|---------|------------------------------------------------------------------------------------------------------------------------------|
| `timeout` | `float` | The timeout in seconds. Must be non-negative. This timeout will be used for all subsequent HTTP requests made by the module. |

### `get_timeout() float`

Get the current global timeout setting for HTTP requests.
returns:
The current timeout in seconds used for HTTP requests.

## Types

### `response`

The result of performing a HTTP request.

**Fields**

| name          | type     | description                                                         |
|---------------|----------|---------------------------------------------------------------------|
| `url`         | `string` | the URL that was ultimately requested (may change after redirects). |
| `status_code` | `int`    | response status code (for example: `200 == OK`).                    |
| `headers`     | `dict`   | dictionary of response headers.                                     |
| `encoding`    | `string` | transfer encoding. example: "octet-stream" or "application/json".   |

**Methods**

#### `body() string`

output response body as a string.

#### `json() object`

attempt to parse response body as json, returning a JSON-decoded result, or None if the response body is empty or not valid JSON.

### `ExportedServerRequest`

Encapsulates HTTP request data in a format accessible to both Go code and Starlark scripts.

**Fields**

| name       | type       | description                                                                            |
|------------|------------|----------------------------------------------------------------------------------------|
| `method`   | `string`   | The HTTP method (e.g., GET, POST, PUT, DELETE)                                         |
| `url`      | `string`   | The request URL.                                                                       |
| `proto`    | `string`   | The protocol used for the request (e.g., HTTP/1.1).                                    |
| `host`     | `string`   | The host specified in the request.                                                     |
| `remote`   | `string`   | The remote address of the client.                                                      |
| `headers`  | `dict`     | The HTTP headers included in the request.                                              |
| `query`    | `dict`     | The query parameters included in the request.                                          |
| `encoding` | `[]string` | The transfer encodings specified in the request.                                       |
| `body`     | `string`   | The request body data                                                                  |
| `json`     | `any`      | The request body data as JSON, or None if the request body is empty or not valid JSON. |

### `ServerResponse`

Enables HTTP response manipulation within Starlark scripts, facilitating dynamic preparation of HTTP responses in Go-based web servers.

**Methods**

#### `set_status(code int)`

Sets the HTTP status code for the response.

#### `set_code(code int)`

Alias for set_status.

#### `add_header(key string, value string)`

Adds a header with the given key and value to the response.

#### `set_content_type(content_type string)`

Sets the Content-Type header for the response, it will overwrite any existing or implicit Content-Type header.

#### `set_data(data string|bytes)`

Sets the response data as binary data, and the Content-Type header to `application/octet-stream`.

#### `set_json(data any)`

Sets the response data as JSON, marshaling the given Starlark value to JSON, and the Content-Type header to `application/json`.

#### `set_text(data string|bytes)`

Sets the response data as plain text, and the Content-Type header to `text/plain`.

#### `set_html(data string|bytes)`

Sets the response data as HTML, and the Content-Type header to `text/html`.
