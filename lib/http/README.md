# HTTP

`http` defines an HTTP client implementation. It is a thin wrapper around the Go standard package `net/http` but in Python `requests` style.

## Functions

### `get(url,params={},headers={},auth=(),timeout=30,allow_redirects=True,verify=True) response`

Perform an HTTP GET request, returning a response.

#### Parameters

| name              | type     | description                                                                                                  |
|-------------------|----------|--------------------------------------------------------------------------------------------------------------|
| `url`             | `string` | URL to request.                                                                                              |
| `headers`         | `dict`   | optional. dictionary of headers to add to request.                                                           |
| `auth`            | `tuple`  | optional. (username,password) tuple for HTTP Basic authorization.                                            |
| `timeout`         | `float`  | optional. how many seconds to wait for the server to send all the data before giving up. 0 means no timeout. |
| `allow_redirects` | `bool`   | optional. whether to follow redirects.                                                                       |
| `verify`          | `bool`   | optional. whether to verify the server's SSL certificate.                                                    |

### `put(url,params={},headers={},body="",form_body={},form_encoding="",json_body={},auth=(),timeout=30,allow_redirects=True,verify=True) response`

Perform an HTTP PUT request, returning a response.

#### Parameters

| name              | type     | description                                                                                                  |
|-------------------|----------|--------------------------------------------------------------------------------------------------------------|
| `url`             | `string` | URL to request.                                                                                              |
| `headers`         | `dict`   | optional. dictionary of headers to add to request.                                                           |
| `body`            | `string` | optional. raw string body to provide to the request.                                                         |
| `form_body`       | `dict`   | optional. dict of values that will be encoded as form data.                                                  |
| `form_encoding`   | `string` | optional. `application/x-www-form-url-encoded` (default) or `multipart/form-data`.                           |
| `json_body`       | `any`    | optional. JSON data to supply as a request. handy for working with JSON-API's.                               |
| `auth`            | `tuple`  | optional. (username,password) tuple for HTTP Basic authorization.                                            |
| `timeout`         | `float`  | optional. how many seconds to wait for the server to send all the data before giving up. 0 means no timeout. |
| `allow_redirects` | `bool`   | optional. whether to follow redirects.                                                                       |
| `verify`          | `bool`   | optional. whether to verify the server's SSL certificate.                                                    |

### `post(url,params={},headers={},body="",form_body={},form_encoding="",json_body={},auth=(),timeout=30,allow_redirects=True,verify=True) response`

Perform an HTTP POST request, returning a response.

#### Parameters

| name              | type     | description                                                                                                  |
|-------------------|----------|--------------------------------------------------------------------------------------------------------------|
| `url`             | `string` | URL to request.                                                                                              |
| `headers`         | `dict`   | optional. dictionary of headers to add to request.                                                           |
| `body`            | `string` | optional. raw string body to provide to the request.                                                         |
| `form_body`       | `dict`   | optional. dict of values that will be encoded as form data.                                                  |
| `form_encoding`   | `string` | optional. `application/x-www-form-url-encoded` (default) or `multipart/form-data`.                           |
| `json_body`       | `any`    | optional. JSON data to supply as a request. handy for working with JSON-API's.                               |
| `auth`            | `tuple`  | optional. (username,password) tuple for HTTP Basic authorization.                                            |
| `timeout`         | `float`  | optional. how many seconds to wait for the server to send all the data before giving up. 0 means no timeout. |
| `allow_redirects` | `bool`   | optional. whether to follow redirects.                                                                       |
| `verify`          | `bool`   | optional. whether to verify the server's SSL certificate.                                                    |

### `postForm(url,params={},headers={},form_body={},form_encoding="",auth=(),timeout=30,allow_redirects=True,verify=True) response`

Perform an HTTP POST request with form data, returning a response.

#### Parameters

| name              | type     | description                                                                                                  |
|-------------------|----------|--------------------------------------------------------------------------------------------------------------|
| `url`             | `string` | URL to request.                                                                                              |
| `headers`         | `dict`   | optional. dictionary of headers to add to request.                                                           |
| `form_body`       | `dict`   | optional. dict of values that will be encoded as form data.                                                  |
| `form_encoding`   | `string` | optional. `application/x-www-form-url-encoded` (default) or `multipart/form-data`.                           |
| `auth`            | `tuple`  | optional. (username,password) tuple for HTTP Basic authorization.                                            |
| `timeout`         | `float`  | optional. how many seconds to wait for the server to send all the data before giving up. 0 means no timeout. |
| `allow_redirects` | `bool`   | optional. whether to follow redirects.                                                                       |
| `verify`          | `bool`   | optional. whether to verify the server's SSL certificate.                                                    |

### `delete(url,params={},headers={},body="",form_body={},form_encoding="",json_body={},auth=(),timeout=30,allow_redirects=True,verify=True) response`

Perform an HTTP DELETE request, returning a response.

#### Parameters

| name              | type     | description                                                                                                  |
|-------------------|----------|--------------------------------------------------------------------------------------------------------------|
| `url`             | `string` | URL to request.                                                                                              |
| `headers`         | `dict`   | optional. dictionary of headers to add to request.                                                           |
| `body`            | `string` | optional. raw string body to provide to the request.                                                         |
| `form_body`       | `dict`   | optional. dict of values that will be encoded as form data.                                                  |
| `form_encoding`   | `string` | optional. `application/x-www-form-url-encoded` (default) or `multipart/form-data`                            |
| `json_body`       | `any`    | optional. JSON data to supply as a request. handy for working with JSON-API's.                               |
| `auth`            | `tuple`  | optional. (username,password) tuple for HTTP Basic authorization.                                            |
| `timeout`         | `float`  | optional. how many seconds to wait for the server to send all the data before giving up. 0 means no timeout. |
| `allow_redirects` | `bool`   | optional. whether to follow redirects.                                                                       |
| `verify`          | `bool`   | optional. whether to verify the server's SSL certificate.                                                    |

### `patch(url,params={},headers={},body="",form_body={},form_encoding="",json_body={},auth=(),timeout=30,allow_redirects=True,verify=True) response`

Perform an HTTP PATCH request, returning a response.

#### Parameters

| name              | type     | description                                                                                                  |
|-------------------|----------|--------------------------------------------------------------------------------------------------------------|
| `url`             | `string` | URL to request.                                                                                              |
| `headers`         | `dict`   | optional. dictionary of headers to add to request.                                                           |
| `body`            | `string` | optional. raw string body to provide to the request.                                                         |
| `form_body`       | `dict`   | optional. dict of values that will be encoded as form data.                                                  |
| `form_encoding`   | `string` | optional. `application/x-www-form-url-encoded` (default) or `multipart/form-data`.                           |
| `json_body`       | `any`    | optional. JSON data to supply as a request. handy for working with JSON-API's.                               |
| `auth`            | `tuple`  | optional. (username,password) tuple for HTTP Basic authorization.                                            |
| `timeout`         | `float`  | optional. how many seconds to wait for the server to send all the data before giving up. 0 means no timeout. |
| `allow_redirects` | `bool`   | optional. whether to follow redirects.                                                                       |
| `verify`          | `bool`   | optional. whether to verify the server's SSL certificate.                                                    |

### `options(url,params={},headers={},body="",form_body={},form_encoding="",json_body={},auth=(),timeout=30,allow_redirects=True,verify=True) response`

Perform an HTTP OPTIONS request, returning a response.

#### Parameters

| name              | type     | description                                                                                                  |
|-------------------|----------|--------------------------------------------------------------------------------------------------------------|
| `url`             | `string` | URL to request.                                                                                              |
| `headers`         | `dict`   | optional. dictionary of headers to add to request.                                                           |
| `body`            | `string` | optional. raw string body to provide to the request.                                                         |
| `form_body`       | `dict`   | optional. dict of values that will be encoded as form data.                                                  |
| `form_encoding`   | `string` | optional. `application/x-www-form-url-encoded` (default) or `multipart/form-data`.                           |
| `json_body`       | `any`    | optional. JSON data to supply as a request. handy for working with JSON-API's.                               |
| `auth`            | `tuple`  | optional. (username,password) tuple for HTTP Basic authorization.                                            |
| `timeout`         | `float`  | optional. how many seconds to wait for the server to send all the data before giving up. 0 means no timeout. |
| `allow_redirects` | `bool`   | optional. whether to follow redirects.                                                                       |
| `verify`          | `bool`   | optional. whether to verify the server's SSL certificate.                                                    |

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

output response body as a string

#### `json() object`

attempt to parse response body as json, returning a JSON-decoded result.
