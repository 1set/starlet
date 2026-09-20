# Security and compatibility

## Supported use

starlet supports Go 1.19 and keeps the ecosystem interpreter baseline at
`go.starlark.net v0.0.0-20260324133313-ffb3f39dd27a`. Build production
applications with a supported Go toolchain containing current security fixes.
Keeping a library compatible with Go 1.19 does not make that old toolchain a
recommended production compiler.

## Known parser limitation

The pinned interpreter predates the upstream parser recursion fix described in
[google/starlark-go commit 5395d018f003](https://github.com/google/starlark-go/commit/5395d018f003e2a08bfbca6dcb2562acee700f62)
(GHSA-wcqm-92f8-mh2v). Excessively nested source can exhaust the Go stack and
terminate the host process. This release retains that known limitation to
preserve the established interpreter and Go compatibility baseline; it does
not claim to fix the vulnerability.

The boundary covers direct source, readers, files, compilation or checking,
and every module loaded through `load()`. Source must be selected and reviewed
by the host, including generated code and loaded modules. Data accepted by a
trusted script is not permission to evaluate that data as new source.

Execution-step limits, cancellation, timeouts, capability gates and `recover`
do not prevent a fatal stack overflow during parsing. An input-size limit can
reduce exposure but is not a complete substitute for a parser recursion limit.
Documentation states this boundary; it does not remove the vulnerable code.

For attacker-controlled source, the host needs a separately reviewed deployment
with process isolation and external CPU, memory, time and output limits, or an
explicit interpreter upgrade/backport with compatibility testing. Those are
host changes, not protections supplied by this package. A newer Go compiler
alone does not patch the old interpreter. The library is not an in-process
security boundary for untrusted or multi-tenant script execution.

Go module requirements are minimum versions: another dependency can select a
newer interpreter through minimal version selection. Hosts should verify the
actual selection with `go list -m go.starlark.net` and review any change. An
application's `replace` directive is local to that application and does not
provide a backport to consumers of this library.

## Reporting

Report vulnerabilities privately through
[GitHub security advisories](https://github.com/1set/starlet/security/advisories/new).
Use synthetic examples without credentials or business data. No response-time
SLA is promised.
