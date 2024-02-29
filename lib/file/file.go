// Package file defines functions that manipulate files, it's inspired by file helpers from Amoy.
package file

import (
	"fmt"
	"sync"

	itn "github.com/1set/starlet/internal"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

// ModuleName defines the expected name for this Module when used
// in starlark's load() function, eg: load('file', 'trim_bom')
const ModuleName = "file"

var (
	once       sync.Once
	fileModule starlark.StringDict
	none       = starlark.None
)

// LoadModule loads the file module. It is concurrency-safe and idempotent.
func LoadModule() (starlark.StringDict, error) {
	once.Do(func() {
		fileModule = starlark.StringDict{
			ModuleName: &starlarkstruct.Module{
				Name: ModuleName,
				Members: starlark.StringDict{
					"trim_bom":    starlark.NewBuiltin(ModuleName+".trim_bom", trimBom),
					"read_bytes":  wrapReadFile("read_bytes", readBytes),
					"read_string": wrapReadFile("read_string", readString),
					"read_lines":  wrapReadFile("read_lines", readLines),
				},
			},
		}
	})
	return fileModule, nil
}

// trimBom removes the UTF-8 BOM (Byte Order Mark) from the beginning of a string.
func trimBom(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if l := len(args); l != 1 {
		return nil, fmt.Errorf(`%s: takes exactly one argument (%d given)`, b.Name(), l)
	}

	switch r := args[0]; v := r.(type) {
	case starlark.String:
		return starlark.String(TrimUTF8BOM([]byte(v))), nil
	case starlark.Bytes:
		return starlark.Bytes(TrimUTF8BOM([]byte(v))), nil
	default:
		return none, fmt.Errorf(`%s: expected string or bytes, got %s`, b.Name(), r.Type())
	}
}

// wrapReadFile wraps the file reading functions to be used in Starlark.
func wrapReadFile(funcName string, workLoad func(name string) (starlark.Value, error)) starlark.Callable {
	return starlark.NewBuiltin(ModuleName+"."+funcName, func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		var fp itn.StringOrBytes
		if err := starlark.UnpackArgs(b.Name(), args, kwargs, "name", &fp); err != nil {
			return starlark.None, err
		}
		return workLoad(fp.GoString())
	})
}

// readBytes reads the whole named file and returns the contents as bytes.
func readBytes(name string) (starlark.Value, error) {
	data, err := ReadFileBytes(name)
	if err != nil {
		return nil, err
	}
	return starlark.Bytes(data), nil
}

// readString reads the whole named file and returns the contents as string.
func readString(name string) (starlark.Value, error) {
	data, err := ReadFileString(name)
	if err != nil {
		return nil, err
	}
	return starlark.String(data), nil
}

// readLines reads the whole named file and returns the contents as a list of lines.
func readLines(name string) (starlark.Value, error) {
	ls, err := ReadFileLines(name)
	if err != nil {
		return nil, err
	}
	sl := make([]starlark.Value, len(ls))
	for i, l := range ls {
		sl[i] = starlark.String(l)
	}
	return starlark.NewList(sl), nil
}
