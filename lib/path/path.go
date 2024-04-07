// Package path defines functions that manipulate directories, it's inspired by path helpers from Mojo.
package path

import (
	"fmt"
	"path/filepath"
	"sync"

	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

// ModuleName defines the expected name for this Module when used in starlark's load() function, eg: load('path', 'join')
const ModuleName = "path"

var (
	once       sync.Once
	pathModule starlark.StringDict
)

// LoadModule loads the path module. It is concurrency-safe and idempotent.
func LoadModule() (starlark.StringDict, error) {
	once.Do(func() {
		pathModule = starlark.StringDict{
			ModuleName: &starlarkstruct.Module{
				Name: ModuleName,
				Members: starlark.StringDict{
					"join":    starlark.NewBuiltin(ModuleName+".join", joinPaths),
					"abspath": starlark.NewBuiltin(ModuleName+".abspath", absPath),
				},
			},
		}
	})
	return pathModule, nil
}

// absPath returns the absolute representation of path.
func absPath(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var path string
	if err := starlark.UnpackArgs(b.Name(), args, kwargs, "path", &path); err != nil {
		return nil, err
	}
	// get absolute path
	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	return starlark.String(abs), nil
}

// joinPaths joins any number of path elements into a single path.
func joinPaths(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	// check arguments
	if len(args) < 1 {
		return nil, fmt.Errorf("%s: got %d arguments, want at least 1", b.Name(), len(args))
	}
	// unpack arguments
	paths := make([]string, len(args))
	for i, arg := range args {
		s, ok := starlark.AsString(arg)
		if !ok {
			return nil, fmt.Errorf("%s: got %s, want string", b.Name(), arg.Type())
		}
		paths[i] = s
	}
	// join paths
	joined := filepath.Join(paths...)
	return starlark.String(joined), nil
}
