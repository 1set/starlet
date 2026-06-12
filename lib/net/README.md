# net

`net` provides network diagnostics for Starlark: DNS lookup, TCP ping, and HTTP ping. It is inspired by Go's `net` package and Python's `socket` module.

All three functions honor the machine's context: a `RunWithTimeout`/`RunWithContext` deadline aborts a lookup or an in-flight ping loop promptly.

## Functions

### `nslookup(domain, dns_server=None, timeout=10) []string`

Looks up the IP addresses of a domain name, returning a list of strings.

| name         | type     | description                                                                                                                                                                      |
|--------------|----------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `domain`     | `string` | the domain name to resolve (an IP literal is returned as-is)                                                                                                                    |
| `dns_server` | `string` | optional DNS server as `host` or `host:port` (port defaults to 53). Uses the system resolver when omitted. NOTE: a custom server requires the pure-Go resolver, which is a no-op on Windows before Go 1.19 — the system resolver is used there instead. |
| `timeout`    | `float/int` | lookup timeout in seconds; non-positive values fall back to 10                                                                                                                |

### `tcping(hostname, port=80, count=4, timeout=10, interval=1) statistics`

Measures TCP connect round-trip times to `hostname:port` and returns a `statistics` struct.

| name       | type        | description                                                              |
|------------|-------------|--------------------------------------------------------------------------|
| `hostname` | `string`    | the host to ping (resolved first; an IP literal skips DNS)               |
| `port`     | `int`       | TCP port, defaults to 80                                                 |
| `count`    | `int`       | number of rounds, `1..1024`                                              |
| `timeout`  | `float/int` | per-connect timeout in seconds (sub-second values work); at most 3600    |
| `interval` | `float/int` | pause between rounds in seconds (sub-second values work); at most 3600   |

### `httping(url, count=4, timeout=10, interval=1) statistics`

Measures HTTP connect round-trip times by issuing GET requests to `url` (redirects are not followed; a status outside `200..399` counts as a failed round) and returns a `statistics` struct.

Takes the same `count`/`timeout`/`interval` parameters and bounds as `tcping`. The client honors the system proxy settings (`HTTP_PROXY`, `HTTPS_PROXY`, `NO_PROXY`), just like `lib/http`; behind a proxy, the measured connect duration is the TCP connection to the proxy (the first hop), not to the origin server.

## Types

### `statistics`

| member    | type     | description                                  |
|-----------|----------|----------------------------------------------|
| `address` | `string` | the resolved address or URL that was pinged  |
| `total`   | `int`    | rounds attempted                             |
| `success` | `int`    | rounds that succeeded                        |
| `loss`    | `float`  | failed percentage (0–100)                    |
| `min`/`avg`/`max`/`stddev` | `float` | round-trip statistics in milliseconds |

If no round succeeds, the function reports a `no successful connections` error instead of returning a struct.
