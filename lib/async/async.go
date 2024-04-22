// Package async provides asynchronous operations for Starlark, inspired by Python's asyncio module and Go's goroutines.
package async

import (
	"fmt"
	"sync"

	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

// ModuleName defines the expected name for this Module when used in starlark's load() function, eg: load('async', 'go')
const ModuleName = "async"

var (
	once        sync.Once
	asyncModule starlark.StringDict
)

// LoadModule loads the async module. It is concurrency-safe and idempotent.
func LoadModule() (starlark.StringDict, error) {
	once.Do(func() {
		asyncModule = starlark.StringDict{
			ModuleName: &starlarkstruct.Module{
				Name: ModuleName,
				Members: starlark.StringDict{
					"run": starlark.NewBuiltin(ModuleName+".run", asyncCallFunc),
				},
			},
		}
	})
	return asyncModule, nil
}

func asyncCallFunc(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if l := len(args); l < 1 {
		return nil, fmt.Errorf("%s: takes at least one argument (%d given)", b.Name(), l)
	}

	// first argument must be a callable
	fnc, ok := args[0].(starlark.Callable)
	if !ok {
		return nil, fmt.Errorf("%s: first argument must be callable", b.Name())
	}

	// call the function asynchronously
	go func() {
		// call the function
		val, err := fnc.CallInternal(thread, args[1:], kwargs)
		fmt.Println("[done callable]", val, err)
	}()
	return nil, nil
}
