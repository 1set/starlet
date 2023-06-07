package starlet

import (
	"context"
	"errors"
)

var (
	ErrUnknownScriptSource = errors.New("starlet: unknown script source")
)

// Run runs the preset script with given globals and returns the result.
func (m *Machine) run(ctx context.Context) (DataStore, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// either script content or name and FS must be set
	if !((m.scriptContent != nil) || (m.scriptName != "" && m.scriptFS != nil)) {
		return nil, ErrUnknownScriptSource
	}

	// TODO: implement
	return nil, nil
}

// TODO: Multiple FS for script and modules
// TODO: Reset machine
