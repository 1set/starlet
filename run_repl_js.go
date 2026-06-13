//go:build js

package starlet

// REPL is unavailable on js/wasm builds: the upstream go.starlark.net/repl
// pulls in a terminal library (chzyer/readline) that does not compile for
// GOOS=js, and an interactive terminal REPL is meaningless in a browser.
// Drive the Machine with Run/RunScript instead. This stub keeps the method
// present so consumer code referencing REPL still compiles for wasm.
func (m *Machine) REPL() {
	// no-op in js/wasm; use Run/RunScript
}
