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
	// Hold the write lock for the whole REPL session, matching runInternal:
	// prepareThread and repl.REPLOptions read and mutate m.thread/m.predeclared,
	// and every other execution path already serializes on m.mu (String() uses
	// TryRLock and returns the running snapshot rather than blocking). A REPL
	// owns the Machine exclusively for its duration, so holding the lock until it
	// exits is the right scope — without it, a REPL concurrent with Run/Reset
	// races on m.thread.
	m.mu.Lock()
	defer m.mu.Unlock()

	if err := m.prepareThread(nil); err != nil {
		repl.PrintError(err)
		return
	}
	repl.REPLOptions(m.getFileOptions(), m.thread, m.predeclared)
}
