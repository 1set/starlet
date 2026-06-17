//go:build !js && !wasip1

package starlet

import "go.starlark.net/repl"

// REPL is a Read-Eval-Print-Loop for Starlark. It loads the predeclared
// symbols and modules into the global environment and reads from the terminal.
//
// This lives in a terminal-target file because go.starlark.net/repl pulls in
// chzyer/readline, a terminal library that does not compile for browser
// js/wasm or WASI; see run_repl_stub.go for the no-terminal stub.
func (m *Machine) REPL() {
	if err := m.prepareThread(nil); err != nil {
		repl.PrintError(err)
		return
	}
	repl.REPLOptions(m.getFileOptions(), m.thread, m.predeclared)
}
