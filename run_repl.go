//go:build !js

package starlet

import "go.starlark.net/repl"

// REPL is a Read-Eval-Print-Loop for Starlark. It loads the predeclared
// symbols and modules into the global environment and reads from the terminal.
//
// This lives in a non-js file because go.starlark.net/repl pulls in
// chzyer/readline (a terminal library) which does not compile for GOOS=js;
// see run_repl_js.go for the wasm stub.
func (m *Machine) REPL() {
	if err := m.prepareThread(nil); err != nil {
		repl.PrintError(err)
		return
	}
	repl.REPLOptions(m.getFileOptions(), m.thread, m.predeclared)
}
