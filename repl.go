package starlet

import (
	"go.starlark.net/repl"
)

// REPL is a Read-Eval-Print-Loop for Starlark.
func (m *Machine) REPL() {
	repl.REPL(m.thread, m.predeclared)
}
