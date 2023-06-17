package starlet

import (
	"go.starlark.net/repl"
)

// REPL is a Read-Eval-Print-Loop for Starlark.
// It loads the predeclared symbols and modules into the global environment,
func (m *Machine) REPL() {
	if err := m.prepareThread(nil); err != nil {
		repl.PrintError(err)
		return
	}
	repl.REPL(m.thread, m.predeclared)
}
