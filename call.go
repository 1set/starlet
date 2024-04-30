package starlet

import (
	"context"

	"github.com/1set/starlight/convert"
	"go.starlark.net/starlark"
)

// Call executes a Starlark function or builtin saved in the thread and returns the result.
func (m *Machine) Call(name string, args ...interface{}) (out interface{}, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	defer func() {
		if r := recover(); r != nil {
			err = errorStarlarkPanic("call", r)
		}
	}()

	// preconditions
	if name == "" {
		return nil, errorStarletErrorf("call", "no function name")
	}
	if m.predeclared == nil || m.thread == nil {
		return nil, errorStarletErrorf("call", "no function loaded")
	}
	var callFunc starlark.Callable
	if rf, ok := m.predeclared[name]; !ok {
		return nil, errorStarletErrorf("call", "no such function: %s", name)
	} else if sf, ok := rf.(*starlark.Function); ok {
		callFunc = sf
	} else if sb, ok := rf.(*starlark.Builtin); ok {
		callFunc = sb
	} else {
		return nil, errorStarletErrorf("call", "mistyped function: %s", name)
	}

	// convert arguments
	sl := starlark.Tuple{}
	for _, arg := range args {
		sv, err := convert.ToValueWithTag(arg, m.customTag)
		if err != nil {
			return nil, errorStarlightConvert("args", err)
		}
		sl = append(sl, sv)
	}

	// reset thread
	m.thread.Uncancel()
	m.thread.SetLocal("context", context.TODO())

	// call and convert result
	res, err := starlark.Call(m.thread, callFunc, sl, nil)
	if m.enableOutConv { // convert to interface{} if enabled
		out = convert.FromValue(res)
	} else {
		out = res
	}
	// handle error
	if err != nil {
		return out, errorStarlarkError("call", err)
	}
	return out, nil
}
