package starlet

import (
	"context"
	"errors"
)

var (
	ErrUnknownScriptSource = errors.New("unknown script source")
)

// Run runs the preset script with given globals and returns the result.
func (m *Machine) run(ctx context.Context) (DataStore, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// run the script from various sources
	switch m.scriptSource {
	case scriptSourceContent:
		return m.runContent(ctx)
	case scriptSourceFileSystem:
		return m.runFileSystem(ctx)
	case scriptSourceUnknown:
		fallthrough
	default:
		return nil, ErrUnknownScriptSource
	}
}

func (m *Machine) runContent(ctx context.Context) (DataStore, error) {
	// TODO: implement
	return nil, nil
}

func (m *Machine) runFileSystem(ctx context.Context) (DataStore, error) {
	// TODO: implement
	return nil, nil
}

// TODO: Multiple FS for script and modules
// TODO: Reset machine
