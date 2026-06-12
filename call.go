package starlet

import (
	"context"
	"time"

	"github.com/1set/starlight/convert"
	"go.starlark.net/starlark"
)

// Call executes a Starlark function or builtin saved in the thread and returns the result.
// The function runs without a context: it cannot be cancelled or time-bounded;
// use CallWithContext or CallWithTimeout for that.
func (m *Machine) Call(name string, args ...interface{}) (out interface{}, err error) {
	return m.callInternal(nil, name, args)
}

// CallWithTimeout executes like Call, but the function call (and any
// context-aware library builtin it invokes) is aborted once the timeout
// elapses, matching RunWithTimeout semantics.
func (m *Machine) CallWithTimeout(timeout time.Duration, name string, args ...interface{}) (out interface{}, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return m.callInternal(ctx, name, args)
}

// CallWithContext executes like Call, but the function call (and any
// context-aware library builtin it invokes) is aborted when ctx is
// cancelled, matching RunWithContext semantics. A nil context behaves like
// plain Call; an already-cancelled context fails immediately.
func (m *Machine) CallWithContext(ctx context.Context, name string, args ...interface{}) (out interface{}, err error) {
	return m.callInternal(ctx, name, args)
}

func (m *Machine) callInternal(ctx context.Context, name string, args []interface{}) (out interface{}, err error) {
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

	// reset the thread and wire the context
	m.thread.Uncancel()
	if ctx == nil {
		// plain Call: no cancellation channel, as before
		m.thread.SetLocal("context", context.TODO())
	} else {
		if err := ctx.Err(); err != nil {
			// fail fast instead of silently calling with a context that
			// can never fire again
			return nil, errorStarletError("call", err)
		}
		m.thread.SetLocal("context", ctx)
		stop := m.watchContextCancel(ctx)
		defer stop()
	}

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
