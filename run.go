package starlet

import (
	"context"
	"errors"
)

var (
	ErrUnknownScriptSource = errors.New("unknown script source")
)

func (m *Machine) run(ctx context.Context) (map[string]interface{}, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

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
	return nil, nil
}

func (m *Machine) runContent(ctx context.Context) (map[string]interface{}, error) {
	// TODO: implement
	return nil, nil
}

func (m *Machine) runFileSystem(ctx context.Context) (map[string]interface{}, error) {
	// TODO: implement
	return nil, nil
}
