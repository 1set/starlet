package starlet

import (
	"context"
	"errors"
	"fmt"

	"github.com/1set/starlight/convert"
	"go.starlark.net/starlark"
)

func (m *Machine) Call(name string, args ...interface{}) (out interface{}, err error) {
	// preconditions
	if name == "" {
		return nil, errors.New("no function name")
	}
	if m.predeclared == nil || m.thread == nil {
		return nil, errors.New("no function loaded")
	}
	var starFunc *starlark.Function
	if rf, ok := m.predeclared[name]; !ok {
		return nil, errors.New("function not found")
	} else if starFunc, ok = rf.(*starlark.Function); !ok {
		return nil, errors.New("mistyped function")
	}

	// convert arguments
	sl := starlark.Tuple{}
	for _, arg := range args {
		sv, err := convert.ToValue(arg)
		if err != nil {
			return nil, fmt.Errorf("starlet: convert arg: %w", err)
		}
		sl = append(sl, sv)
	}

	// reset thread
	m.thread.Uncancel()
	m.thread.SetLocal("context", context.TODO())

	// call and convert result
	res, err := starlark.Call(m.thread, starFunc, sl, nil)
	out = convert.FromValue(res)
	if err != nil {
		return out, fmt.Errorf("starlet: call: %w", err)
	}
	return out, nil
}
