# net

`net` provides network diagnostics for Starlark: DNS lookup (`nslookup`), TCP connect ping (`tcping`), and HTTP connect ping (`httping`). It is inspired by Go's `net` package and Python's `socket` module.

Capability profile: **Network** — every function performs real DNS resolution and/or outbound connections. All three honor the machine's context: a `RunWithTimeout`/`RunWithContext` deadline aborts a lookup or an in-flight ping loop promptly, between rounds.

## Functions

| function | description |
|----------|-------------|
| `nslookup(domain, dns_server=None, timeout=10) -> list[string]` | resolve a domain to its IP addresses |
| `tcping(hostname, port=80, count=4, timeout=10, interval=1) -> statistics` | measure TCP connect round-trip times to `hostname:port` |
| `httping(url, count=4, timeout=10, interval=1) -> statistics` | measure HTTP connect round-trip times by issuing GET requests to `url` |

## Types

### `statistics`

A `starlarkstruct` returned by `tcping` and `httping`. Round-trip values are in milliseconds. Read-only; access by attribute (`s.avg`).

| attribute | type | description |
|-----------|------|-------------|
| `address` | `string` | the resolved `host:port` (`tcping`) or the URL (`httping`) that was pinged |
| `total` | `int` | rounds attempted (equals `count`) |
| `success` | `int` | rounds that succeeded |
| `loss` | `float` | failed percentage, `0`–`100` |
| `min` | `float` | minimum round-trip time, ms |
| `avg` | `float` | mean round-trip time, ms |
| `max` | `float` | maximum round-trip time, ms |
| `stddev` | `float` | standard deviation of round-trip times, ms (`0` for a single successful round) |

If no round succeeds, the function errors with `no successful connections` instead of returning a `statistics` struct.

## Details & examples

### `nslookup`

`nslookup(domain, dns_server=None, timeout=10) -> list[string]`

Looks up the IP addresses of `domain`, returning a list of address strings.

- `domain` (`string`/`bytes`) — the name to resolve. An IP literal is returned as-is.
- `dns_server` (`string`/`bytes`, optional) — a DNS server as `host` or `host:port`; the port defaults to `53` when omitted. Uses the system resolver when not given. A custom server requires the pure-Go resolver, which is a no-op on Windows before Go 1.19 — the system resolver is used there instead.
- `timeout` (`float`/`int`, optional) — lookup timeout in seconds; non-positive values fall back to `10`.

Errors when the lookup fails — the error always names the domain, whether the failure is an offline timeout, a server timeout (a `timeout` error against an unreachable `dns_server`), or an online `NXDOMAIN` (e.g. `missing.invalid`).

```python
load('net', 'nslookup')
# an IP literal resolves to itself without touching DNS
print(nslookup('8.8.8.8'))
# Output:
# ["8.8.8.8"]
```

### `tcping`

`tcping(hostname, port=80, count=4, timeout=10, interval=1) -> statistics`

Resolves `hostname` (an IP literal skips DNS), then opens and immediately closes a TCP connection to `host:port` `count` times, returning a `statistics` struct of the connect times.

- `port` (`int`) — TCP port, default `80`.
- `count` (`int`) — number of rounds, `1`–`1024`.
- `timeout` (`float`/`int`) — per-connect timeout in seconds; sub-second values work; at most `3600`. Non-positive falls back to `10`.
- `interval` (`float`/`int`) — pause between rounds in seconds; sub-second values work; at most `3600`. Non-positive falls back to `1`.

Errors when: `count <= 0` (`count must be greater than 0`); `count > 1024` (`count must be at most 1024`); `timeout`/`interval > 3600` (`... must be at most 3600 seconds`); the hostname cannot be resolved (the error names the host); or no round connects (`no successful connections`). A round that fails to connect is counted as a loss, not an error, as long as at least one round succeeds.

```python
load('net', 'tcping')
s = tcping('127.0.0.1', port=local_port, count=4, interval=0.1)
print(s.total, s.success > 0)
# Output:
# 4 True
```

### `httping`

`httping(url, count=4, timeout=10, interval=1) -> statistics`

Issues a GET request to `url` `count` times and returns a `statistics` struct of the connect (TCP-establish) times. Redirects are not followed; a status outside `200`–`399` makes the round a failure.

Takes the same `count`/`timeout`/`interval` parameters and bounds as `tcping`. The client honors the system proxy settings (`HTTP_PROXY`, `HTTPS_PROXY`, `NO_PROXY`), just like `lib/http`. Behind a proxy, the measured connect duration is the TCP connection to the proxy (the first hop), not to the origin server.

Errors when: `count <= 0`; `count > 1024`; `timeout`/`interval > 3600`; or no round succeeds (`no successful connections`) — this is also what an unresolvable URL or an all-`4xx`/`5xx` target produces. A `3xx` redirect counts as a *success* (the response is used directly rather than followed).

```python
load('net', 'httping')
s = httping(server_url, count=4, interval=0.1)
print(s.total, s.success > 0)
# Output:
# 4 True
```

## Notes / boundaries

- **Engine.** DNS uses Go's `net.Resolver` (a custom `dns_server` forces `PreferGo`, sending UDP queries to that server); `tcping` uses `net.Dialer.DialContext` over `tcp`; `httping` uses `net/http` with redirect-following disabled and keep-alives off, tracing the first `ConnectStart`/`ConnectDone` pair.
- **Cancellation.** A builtin cannot be stopped mid-flight, so the ping loop checks the context between rounds and uses an interruptible timer for the inter-round pause; a `RunWithTimeout` deadline aborts with a `context` error rather than running all rounds out.
- **Bounds rationale.** `count` is capped at `1024` and `timeout`/`interval` at `3600` seconds so a script (e.g. `count=10**9`) cannot park the host goroutine indefinitely.
- **Determinism.** Round-trip values depend on the live network and are not reproducible; only `address`, `total`, and (given a stable target) `success` are deterministic. Examples above print only those fields and are grounded in the hermetic test stubs (`local_port`/`server_url` are loopback targets).
- **`statistics` shape.** Statistics are computed only over the successful rounds; `min`/`avg`/`max`/`stddev` ignore losses. `loss` is `(total - success) / total * 100`.
