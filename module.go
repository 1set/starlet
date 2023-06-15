package starlet

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"strings"

	"github.com/1set/starlight/convert"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

// ModuleLoader is a function that loads a Starlark module and returns the module's string dict.
type ModuleLoader func() (starlark.StringDict, error)

// NamedModuleLoader is a function that loads a Starlark module with the given name and returns the module's string dict.
// If the module is not found, it returns nil as the first and second return value.
type NamedModuleLoader func(string) (starlark.StringDict, error)

// ModuleLoaderList is a list of Starlark module loaders, usually used to load a list of modules in order.
type ModuleLoaderList []ModuleLoader

// Clone returns a copy of the list.
func (l ModuleLoaderList) Clone() []ModuleLoader {
	return append([]ModuleLoader{}, l...)
}

// LoadAll loads all modules in the list into the given StringDict.
// It returns an error as second return value if any module fails to load.
func (l ModuleLoaderList) LoadAll(d starlark.StringDict) error {
	if d == nil {
		return fmt.Errorf("starlet: cannot load modules into nil dict")
	}
	for _, ld := range l {
		if ld == nil {
			return fmt.Errorf("starlet: nil module loader")
		}
		m, err := ld()
		if err != nil {
			return fmt.Errorf("starlet: failed to load module: %w", err)
		}
		if m != nil {
			for k, v := range m {
				d[k] = v
			}
		}
	}
	return nil
}

// MakeBuiltinModuleLoaderList creates a list of module loaders from a list of module names.
// It returns an error as second return value if any module is not found.
func MakeBuiltinModuleLoaderList(names []string) (ModuleLoaderList, error) {
	ld := make(ModuleLoaderList, len(names))
	for i, name := range names {
		ld[i] = allBuiltinModules[name]
		if ld[i] == nil {
			return ld, fmt.Errorf("starlet: module %q: %w", name, ErrModuleNotFound)
		}
	}
	return ld, nil
}

// ModuleLoaderMap is a map of Starlark module loaders, usually used to load a map of modules by name.
type ModuleLoaderMap map[string]ModuleLoader

// Clone returns a copy of the map.
func (m ModuleLoaderMap) Clone() map[string]ModuleLoader {
	clone := make(map[string]ModuleLoader, len(m))
	for k, v := range m {
		clone[k] = v
	}
	return clone
}

// GetLazyLoader returns a lazy loader that loads the module with the given name.
// It returns an error as second return value if the module is found but fails to load.
// Otherwise, the first return value is nil if the module is not found.
func (m ModuleLoaderMap) GetLazyLoader() NamedModuleLoader {
	return func(s string) (starlark.StringDict, error) {
		// if the map or the name is empty, just return nil to indicate not found
		if m == nil || s == "" {
			return nil, nil
		}
		// attempt to find the module
		ld, ok := m[s]
		if !ok {
			// not found
			return nil, nil
		} else if ld == nil {
			// found but nil
			return nil, fmt.Errorf("nil module loader %q", s)
		}
		// try to load it
		d, err := ld()
		if err != nil {
			// failed to load
			return nil, err
		}
		// extract members of module from dict like `{name: module}`
		if len(d) == 1 {
			m, found := d[s]
			if found {
				if md, ok := m.(*starlarkstruct.Module); ok && md != nil {
					return md.Members, nil
				}
			}
		}
		// otherwise, just return the dict
		return d, nil
	}
}

// MakeBuiltinModuleLoaderMap creates a map of module loaders from a list of module names.
// It returns an error as second return value if any module is not found.
func MakeBuiltinModuleLoaderMap(names []string) (ModuleLoaderMap, error) {
	ld := make(ModuleLoaderMap, len(names))
	for _, name := range names {
		ld[name] = allBuiltinModules[name]
		if ld[name] == nil {
			return ld, fmt.Errorf("starlet: module %q: %w", name, ErrModuleNotFound)
		}
	}
	return ld, nil
}

// MakeModuleLoaderFromStringDict creates a module loader from the given string dict.
func MakeModuleLoaderFromStringDict(d starlark.StringDict) ModuleLoader {
	return func() (starlark.StringDict, error) {
		return d, nil
	}
}

// MakeModuleLoaderFromMap creates a module loader from the given map, it converts the map to a string dict when loading.
func MakeModuleLoaderFromMap(m StringAny) ModuleLoader {
	return func() (starlark.StringDict, error) {
		dict, err := convert.MakeStringDict(m)
		if err != nil {
			return nil, err
		}
		return dict, nil
	}
}

// MakeModuleLoaderFromString creates a module loader from the given source code.
func MakeModuleLoaderFromString(name, source string, predeclared starlark.StringDict) ModuleLoader {
	return func() (starlark.StringDict, error) {
		if name == "" {
			name = "load.star"
		}
		return starlark.ExecFile(&starlark.Thread{}, name, source, predeclared)
	}
}

// MakeModuleLoaderFromReader creates a module loader from the given IO reader.
func MakeModuleLoaderFromReader(name string, rd io.Reader, predeclared starlark.StringDict) ModuleLoader {
	return func() (starlark.StringDict, error) {
		if name == "" {
			name = "load.star"
		}
		return starlark.ExecFile(&starlark.Thread{}, name, rd, predeclared)
	}
}

// MakeModuleLoaderFromFile creates a module loader from the given file.
func MakeModuleLoaderFromFile(name string, fileSys fs.FS, predeclared starlark.StringDict) ModuleLoader {
	return func() (starlark.StringDict, error) {
		// read file content
		b, err := readScriptFile(name, fileSys)
		if err != nil {
			return nil, err
		}
		// execute file
		return starlark.ExecFile(&starlark.Thread{}, name, b, predeclared)
	}
}

// readScriptFile reads a script file from the given file system.
func readScriptFile(name string, fileSys fs.FS) ([]byte, error) {
	// precondition checks
	if name == "" {
		return nil, errors.New("no file name given")
	}
	if fileSys == nil {
		return nil, errors.New("no file system given")
	}

	// if file name does not end with ".star", append it
	if !strings.HasSuffix(name, ".star") {
		name += ".star"
	}

	// open file
	f, err := fileSys.Open(name)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = f.Close()
	}()

	// read
	return io.ReadAll(f)
}
