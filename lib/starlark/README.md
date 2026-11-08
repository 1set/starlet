# starlark

`starlark` evaluates a single **Starlark expression** against an explicit,
read-only data snapshot. It provides normal and tuple-returning error handling,
without inheriting caller functions, modules or host objects.

The name identifies the language. It does not evaluate
[Expr-language](https://github.com/expr-lang/expr) syntax, and does not guess or
retry another language. `expr` is left available for a separate future adapter.

## Functions

| Function | Result |
| --- | --- |
| `eval(source, vars={})` | A frozen native data value; raises an error on failure. |
| `try_eval(source, vars={})` | `(value, None)` on success, or `(None, error_info)` on an ordinary evaluation failure. |

`source` must be a string. `vars` must be a native dict whose keys are valid
Starlark identifiers. Omitting `vars` supplies an empty environment. The module
does not inspect caller locals or copy a Machine's implicit globals.

```python
load("starlark", "eval", "try_eval")

items = [2, 3]
result = eval(
    '{"title": title.upper(), "values": [x * 2 for x in items]}',
    vars={"title": "example", "items": items},
)
# {"title": "EXAMPLE", "values": [4, 6]}

value, error = try_eval("items.append(9)", vars={"items": items})
# value == None; error["phase"] == "runtime"
items.append(4)  # The original input was neither changed nor frozen.

value, error = try_eval("None")
# (None, None): a valid None result is distinct from failure.
```

A preloaded module is also available as `starlark.eval(...)` and
`starlark.try_eval(...)`. When importing both this module and `json.eval`, use
explicit load aliases to distinguish the engines.

## Data and method boundaries

Inputs and results preserve native `None`, bool, int (including big integers),
float, string, bytes, list, dict, tuple and set values. Nested data is supported;
aliases are preserved within a copied graph. Reference cycles, nil values,
functions, bound methods, modules, iterators and custom or Go-backed values are
rejected with a data path. Convert a host object into an explicit native data
projection first; wrapping it with Starlight does not create a snapshot.

Only private input copies are frozen. Native read methods such as `upper`,
`lower`, `get` and `index`, comprehensions and approved computational builtins
remain usable. Mutating a frozen input fails. Expressions may create and modify
their own temporary containers; results are checked, copied and frozen before
publication. Even a frozen lambda or bound method remains callable, so callable
results are rejected recursively.

Caller functions and external modules are not inherited. Unapproved Universe
builtins, including `print`, are blocked without changing the global Universe or
calling the parent's Print handler. The standard computational builtin instances
are captured during package initialization; the embedding application's startup
must retain the standard definitions then. As upstream requires, do not modify
Universe concurrently with execution. This policy is independent of Starlet's
CapPure classification: random and clock sources are not automatically exposed.

The source graph must be exclusively owned/synchronized during evaluation, or
already deeply frozen native data. A lack of writers alone is insufficient:
iterating an unfrozen native set can change its iterator bookkeeping. The helper
does not acquire a Machine or application data lock for the caller.

## Errors and execution state

`error_info` is a frozen dict with these fields:

- `phase` and `kind`: the failing stage or policy category;
- `message`: a diagnostic string;
- `position`: `None` or `{file, line, column}`;
- `frames`: a tuple of `{file, line, column, name}` dicts;
- `truncated`: whether diagnostic text or frames were shortened.

Stages distinguish arguments, input, source admission, parsing, name resolution,
runtime, policy and output checks. Messages are limited to 4096 bytes, the most
recent 32 frames are retained, and frame file/name text is bounded.

Both functions execute on the caller's existing Thread. Its cancellation,
OnMaxSteps behavior and cumulative VM step count remain in force; no new step
allowance is installed or reset. Execution checkpoints also consume VM steps.
`try_eval` does **not** turn caller cancellation, exhausted total steps or a caller
panic into a successful tuple. It only handles errors occurring inside the call;
errors while evaluating the call's arguments still belong to the outer script.

There is no rollback of host callback effects. Such callbacks are excluded from
the data boundary. Errors discard the private evaluation data.

## Admission limits

| Limit | Value |
| --- | --- |
| Source bytes | 4096 |
| Non-whitespace source bytes | 512 |
| AST depth / nodes | 64 / 512 |
| Each copied data graph's depth / visited edges | 64 / 4096 |
| Each copied graph's scalar/name payload | 1 MiB |
| Integer magnitude | At most 4096 bits |

The pre-parser complexity count includes non-whitespace bytes in strings and
comments; only ASCII space, tab, CR and LF are excluded. Put large data in `vars`.
Input and output copying have separate graph quotas. Every visited edge counts,
including memo hits; payload accounting follows the copied graph rather than an
expanded JSON serialization.

These are admission and result limits, **not hard RSS or CPU limits**. The pinned
parser's general recursion issue is not repaired here. VM steps do not account for
every operation inside a native builtin; an intermediate allocation can occur
before output admission. Configure the caller's execution limits, and use a
separately validated process boundary when stronger isolation is required.

## Go usage

```go
import (
    evalmod "github.com/1set/starlet/lib/starlark"
    "go.starlark.net/starlark"
)

thread := &starlark.Thread{Name: "expression"}
thread.SetMaxExecutionSteps(10000)
result, err := evalmod.Eval(thread, "qty * 2", starlark.StringDict{
    "qty": starlark.MakeInt(3),
})
```

The caller owns the Thread and source graph. Direct Go callers must wire
cancellation to `Thread.Cancel` when cancellation during VM execution is needed;
a context stored in Thread.Local is checked at helper checkpoints, not a new
cancellation goroutine. Machine calls retain their existing cancellation bridge.
The Go helper returns `starlark.Value` and an error preserving parser, resolver and
EvalError causes through `errors.As`/`errors.Is`. It does not implicitly convert
the result to Go or JSON.

The module is registered as CapPure because it grants no filesystem, network,
process or logging access. Downstream module sets still decide whether to expose
it; CapPure does not itself establish deterministic execution or a sandbox.
