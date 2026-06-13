# CLAUDE.md

Guidance for Claude Code (claude.ai/code) working in this repository.

## What this is

`starlet` is the L2 layer of the Star* stack: stdlib-style [Starlark](https://github.com/google/starlark-go) modules (`lib/*`) plus the `Machine` runner that wires them into an embeddable scripting environment. It sits on `starlight` (L1, the Go⇄Starlark bridge — `dataconv` wraps it) and under `starbox`/`starpkg`/`starcli`. Pure library plus a nested `cmd/starlet` CLI module. Module floor is **Go 1.19**.

## Commands

```bash
make ci                                    # the CI bar: -race -cover -count 1 ./... + bench compile
go test -race -count=2 ./...               # the real local bar (race + repeat catches flakiness)
go test ./lib/json/ -run TestJSONRepair    # a single test
make test_integration                      # real-network smoke (//go:build integration); default suite is hermetic
go vet ./... && gofmt -l .                 # must be clean before commit

# Verify on the Go floor — local toolchains are newer than go 1.19 and differ
# in stdlib behavior and binary size; trust only the container:
docker run --rm -v "$PWD":/src -v "$HOME/go/pkg/mod":/go/pkg/mod -w /src golang:1.19 go test -race -count=1 ./...
```

CI (`.github/workflows/build.yml`): Go `1.19.x`/`1.25.x` × ubuntu-22.04 / macos-14 / windows-2022. The coverage gate is the `codecov/project` + `codecov/patch` commit statuses (not the upload step). `cmd/starlet` has its own go.mod and builds in a separate **non-gating** job. Merge only when `gh pr checks <n> --watch` exits 0, and read the bot comments (Codacy/Codecov) — they have caught real issues here.

## Architecture

- **Root package** — the runner: `machine.go` (the `Machine`), `run.go`/`exec.go`/`call.go` (script execution and Go⇄script calls), `cache.go`, `module.go`/`types.go`/`error.go`, and `config.go` — the **module registry** (`allBuiltinModules`) plus the **capability map** (`builtinModuleCapabilities`).
- **`dataconv`** — value conversion built on starlight's `convert`: `Marshal`/`Unmarshal` (Starlark⇄Go), `MarshalStarlarkJSON`/`UnmarshalStarlarkJSON`, `ConvertStruct`/`ConvertJSONStruct`.
- **`lib/<m>`** — one directory per module: `<m>.go` (a `LoadModule` returning a `starlarkstruct.Module`), `<m>_test.go`, `README.md`. Modules are self-contained; cross-module reuse goes through `dataconv` or `internal`.
- **`internal`** — test loader helpers: `ExecModuleWithErrorTest` (runs a script with `assert.star`, matches `wantErr` by substring) and `HereDoc`.

## Module selection standard — what belongs in starlet (ADR-016)

A new module must satisfy **all three**, otherwise it goes to `starpkg/*`:

1. **Pure or cleanly classifiable.** Its host effects fit one capability profile (see below); pure computation is the default expectation.
2. **Universally needed.** Broad, domain-neutral utility. Domain modules (sqlite, web, llm, mq, s3…) are starpkg's job no matter how clean they are.
3. **Zero third-party dependencies.** Stdlib-only (or an extension of an existing core module). Any `go.sum` entry is inherited by every downstream — one third-party requirement sends the module to starpkg.

**The vendoring exception** (the `lib/json/internal/jsonrepair` precedent): a frozen, **same-license (MIT)**, **stdlib-only** third-party runtime may be vendored under `lib/<m>/internal/<pkg>/` to keep `go.sum` clean, when the capability is judged worth it. Requirements: pin to a specific upstream release (record it), copy runtime `.go` files only (no `_test.go`, no upstream test deps), keep the upstream LICENSE in the directory, add a `doc.go` stating provenance + "do not edit by hand; re-vendor to update", golden-lock the observed behavior in our tests, and exclude the path in `codecov.yml` and `.codacy.yml`. Measure the binary delta in the go1.19 container before committing to it.

**License hygiene caps vendoring.** Never vendor differently-licensed source — even permissive (Apache-2.0) — into this MIT repository: the copied files keep their license and the repo becomes mixed-license. For a capability worth a differently-licensed library, use a **module dependency** instead, and only when it passes the evaluation bar: its go.mod must not exceed this repo's Go floor, it should bring zero (or near-zero) transitive dependencies into `go.sum`, the binary delta is measured in the go1.19 container, and its panic surface is audited + hostile-input tested (the `lib/json` jsonschema decision: Apache-2.0, go1.19 exactly, zero requires, +256 KiB measured → module dep, repo stays pure MIT).

**Python-parity rule.** If a module mirrors a Python stdlib API (`regex` ⇒ `re`), the shapes must match CPython exactly — signatures, return types (`findall`/`split` return **lists**, not tuples — a real bug class), group shaping, flag values. Where the Go engine genuinely can't (RE2: lookaround, backreferences), **fail to compile with a clear error**; never silently approximate. Same-name-different-shape is worse than absent.

**Honest-boundary rule.** A module either does the thing or errors loudly — no silently-lossy middle ground (`serial`: lossless round-trip or an actionable error; `json.repair`: already-valid input returned byte-for-byte). Error messages say what to do (`convert it to a dict first`), not just "unsupported".

## Adding a builtin module — the peripheral ring

A new `lib/<m>` is never enough on its own; wire the whole ring or the build/tests/docs go stale:

1. `config.go` — add the import, register in `allBuiltinModules`, AND classify in `builtinModuleCapabilities` (`CapPure` / `CapLog` / `CapProcess` / `CapFileSystem` / `CapNetwork`). `TestModuleCapabilities` fails the build if a registered module is unclassified — mandatory.
2. `module_test.go` — add the name to the `builtinModules` list (sorted; the test asserts the exact set and order).
3. Repo `README.md` — a row in the libraries table **and** the example `Modules: [...]` line.
4. `lib/<m>/README.md` — functions table, types, examples; state the module's boundaries (what errors and why).
5. `lib/<m>/<m>_test.go` — table-driven via `itn.ExecModuleWithErrorTest`; test-first; coverage must clear the codecov patch gate (~93%+; only genuinely unreachable defensive arms may stay uncovered).
6. Downstream (NOT this PR; flag for the release legs): repos asserting the module list (starbox's SafeModuleSet goldens) see the new name in their own pin PR. A *pure* module entering the Safe set is correct.

## Writing module tests — conventions and traps

- **No new file per fix/feature, no third-party test frameworks.** Tests live in the module's `<m>_test.go` (or the root's thematic files) as table sections; stdlib `testing` + `internal` helpers only.
- Scripts run through `itn.HereDoc` with **tab indentation matching the surrounding block** (4 tabs inside the standard table layout) — a mismatched level becomes a Starlark indent error.
- Starlark has no `**` operator (write big-int literals), and `load('assert.star', 'assert')` is prepended automatically.
- In-script error checks: `wantErr` matches by **substring**.
- Go values enter scripts via the `predecl` parameter (see `lib/serial`'s use of `startime.Time` / `dataconv.ConvertStruct`).
- The default suite must stay **hermetic** (no real network/DNS; `lib/net` uses local stubs) — real-network tests go behind `//go:build integration`.

## Release discipline

- **Never tag or publish autonomously.** Draft the release title + notes, show the user, and tag only after explicit approval. Patch bump by default. A published tag is immutable in the Go module proxy.
- Before tagging: run each key downstream (starbox first) on a **baseline leg** (its pins) and an **upgrade leg** (`go mod edit -replace` to this repo) in the go1.19 container — only failures new to the upgrade leg are regressions; measure the binary delta (go1.19 container, `-trimpath -ldflags="-s -w"`); scan tag/notes/PR text for internal names.
- Commit style: `[feat]`/`[fix]`/`[refactor]`/`[test]`/`[style]`/`[doc]`/`[ci]`/`[build]` prefixes, imperative subject; PRs reference the Backlog/requirement IDs.

## Reply marker

End every reply with the 🌟 emoji to confirm this file was read and is being followed.
