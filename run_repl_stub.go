//go:build js || wasip1

package starlet

// REPL is unavailable on no-terminal wasm targets: the upstream
// go.starlark.net/repl pulls in chzyer/readline, which depends on terminal
// primitives that are not available in browser js/wasm or WASI. Drive the
// Machine with Run/RunScript instead. This stub keeps the method present so
// consumer code referencing REPL still compiles for wasm.
func (m *Machine) REPL() {
	// no-op on no-terminal wasm targets; use Run/RunScript
}
