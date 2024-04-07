// Package path defines functions that manipulate directories, it's inspired by path helpers from Mojo.
package path

import (
	"fmt"
	"os"
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
					"abs":     starlark.NewBuiltin(ModuleName+".abs", absPath),
					"join":    starlark.NewBuiltin(ModuleName+".join", joinPaths),
					"exists":  wrapExistPath("exists", checkExistPath),
					"is_file": wrapExistPath("is_file", checkFileExist),
					"is_dir":  wrapExistPath("is_dir", checkDirExist),
					"is_link": wrapExistPath("is_link", checkSymlinkExist),
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
			return nil, fmt.Errorf("%s: for parameter path: got %s, want string", b.Name(), arg.Type())
		}
		paths[i] = s
	}
	// join paths
	joined := filepath.Join(paths...)
	return starlark.String(joined), nil
}

// wrapExistPath wraps the existPath function to be used in Starlark with a given function to check if the path exists.
func wrapExistPath(funcName string, workLoad func(path string) bool) starlark.Callable {
	return starlark.NewBuiltin(ModuleName+"."+funcName, func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		var path string
		if err := starlark.UnpackArgs(b.Name(), args, kwargs, "path", &path); err != nil {
			return starlark.None, err
		}
		return starlark.Bool(workLoad(path)), nil
	})
}

// checkExistPath returns true if the path exists, if it's a symbolic link, the symbolic link is followed.
func checkExistPath(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// checkFileExist returns true if the file exists, if it's a symbolic link, the symbolic link is followed.
func checkFileExist(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info != nil && info.Mode().IsRegular()
}

// checkDirExist returns true if the directory exists, if it's a symbolic link, the symbolic link is followed.
func checkDirExist(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info != nil && info.IsDir()
}

// checkSymlinkExist returns true if the symbolic link exists.
func checkSymlinkExist(path string) bool {
	info, err := os.Lstat(path)
	return err == nil && info != nil && info.Mode()&os.ModeSymlink == os.ModeSymlink
}
