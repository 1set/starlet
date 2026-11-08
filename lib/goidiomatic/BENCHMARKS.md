# Expression evaluation benchmark notes

These local measurements compare direct compiled Starlark with the public
`eval` and `try_eval` builtins. See [usage and qualitative guidance](README.md#performance).
They describe the measured workloads and systems; run the benchmarks on your
deployment target before making latency or throughput assumptions.

Local measurements on an Apple M4 Pro, Go 1.27.3, darwin/arm64, `-cpu=1`:
medians of six 500 ms runs, in **microseconds per call** (1000 µs = 1 ms).
Both cases use `qty=3`; the scalar case uses `price=12`, and the nested case
evaluates `{"values": [x * qty for x in items]}` with `items=list(range(32))`.

| Path | `price * qty` | 32-item nested result |
| --- | ---: | ---: |
| Normal compiled Starlark function | 0.062 µs | 1.821 µs |
| `eval`, reused `vars` dict | 4.504 µs (72.4×) | 19.709 µs (10.8×) |
| `eval`, new `vars` dict per call | 4.657 µs (75.2×) | 20.022 µs (11.0×) |
| `try_eval`, reused `vars`, success | 4.630 µs (74.6×) | 19.896 µs (11.0×) |

The normal/reused-`eval` allocation costs were 80/6976 B and 1/111 allocations
for the scalar case, and 1760/25025 B and 11/682 allocations for the nested case.
Every path calls a compiled Starlark function with identical inputs. Outer
compilation, input preparation and warmup are excluded; per-call `eval` parsing,
compilation, copying, validation and freezing are included. Normal execution
does not provide this isolation. Ratios are medians of the six run-wise ratios;
these are local examples, not application-wide latency or throughput guarantees.

Use direct script code for fixed calculations and hot loops. `eval` is intended
for dynamic expressions: it does not cache compiled expressions, even when
`vars` is reused. Run the same workloads with:

```sh
go test ./lib/goidiomatic -run '^$' -bench '^BenchmarkEvalComparison$' \
  -benchmem -benchtime=500ms -cpu=1 -count=6
```

## Increasing the work per call

`eval` reuses the caller's Thread. Each call builds a private variable environment
and compiles an expression; it does not fork the whole runtime. To measure how
this preparation cost compares with longer computation, keep the source fixed:
`len([(i * 3 + 7) % 97 for i in range(n)])`. Only integer `n` is passed in and
returned, so larger input/output graphs do not add copying work.

On the same host, six alternating normal/eval pairs at 300 ms per workload gave:

| Iterations | Normal (µs) | eval (µs) | Paired extra time (µs) | Paired extra (%) |
| --- | ---: | ---: | ---: | ---: |
| 0 | 0.238 | 7.394 | 7.155 | 3013.1% |
| 10 | 1.336 | 8.603 | 7.255 | 544.4% |
| 100 | 9.514 | 17.097 | 7.619 | 80.9% |
| 1,000 | 91.241 | 99.590 | 8.878 | 9.7% |
| 10,000 | 984.503 | 1035.482 | 54.321 | 5.5% |
| 100,000 | 13076.609 | 12915.692 | -156.653* | -1.3%* |

Times are per-path medians; extra time/percentages are medians of paired
differences/ratios, so subtracting the displayed medians need not give the delta.
At zero iterations, about **7.2 µs** estimates the extra cost for this source and
small data graph. The extra allocation was roughly 9.4 KB per call. Relative
overhead becomes smaller as computation grows, but the time gap is not constant.
*At 100,000 iterations, individual paired differences ranged from -1.21 to
+1.38 ms: the negative median does not establish a speedup or absence of cost.*

These timings depend on CPU, OS, architecture, Go/compiler, GC and system load.
Larger source or input/output graphs also increase preparation work. Re-run
`BenchmarkEvalWorkload` with `-benchtime=300ms -cpu=1 -count=6` on your target
system; these results are estimates for the measured workloads, not a universal
per-call constant.

