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
					"trim_bom":     starlark.NewBuiltin(ModuleName+".trim_bom", trimBom),
					"read_bytes":   wrapReadFile("read_bytes", readBytes),
					"read_string":  wrapReadFile("read_string", readString),
					"read_lines":   wrapReadFile("read_lines", readLines),
					"write_bytes":  wrapWriteFile("write_bytes", writeBytes),
					"write_string": wrapWriteFile("write_string", writeString),
					"write_lines":  wrapWriteFile("write_lines", writeLines),
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

// wrapWriteFile wraps the file writing functions to be used in Starlark.
func wrapWriteFile(funcName string, workLoad func(name string, data starlark.Value) error) starlark.Callable {
	return starlark.NewBuiltin(ModuleName+"."+funcName, func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		var (
			fp   itn.StringOrBytes
			data starlark.Value
		)
		if err := starlark.UnpackArgs(b.Name(), args, kwargs, "name", &fp, "data", &data); err != nil {
			return starlark.None, err
		}
		return starlark.None, workLoad(fp.GoString(), data)
	})
}

// writeBytes writes the given data in bytes into a file.
func writeBytes(name string, data starlark.Value) error {
	switch v := data.(type) {
	case starlark.Bytes:
		return WriteFileBytes(name, []byte(v))
	case starlark.String:
		return WriteFileBytes(name, []byte(v))
	default:
		return fmt.Errorf(ModuleName+`.write_bytes: expected string or bytes, got %s`, data.Type())
	}
}

// writeString writes the given data in string into a file.
func writeString(name string, data starlark.Value) error {
	switch v := data.(type) {
	case starlark.Bytes:
		return WriteFileString(name, string(v))
	case starlark.String:
		return WriteFileString(name, string(v))
	default:
		return fmt.Errorf(ModuleName+`.write_string: expected string or bytes, got %s`, data.Type())
	}
}

// writeLines writes the lines into a file. The data should be a list, a tuple or a set of strings.
func writeLines(name string, data starlark.Value) error {
	switch v := data.(type) {
	case *starlark.List:
		if ls, err := convIterStrings(v); err != nil {
			return err
		} else {
			return WriteFileLines(name, ls)
		}
	case *starlark.Tuple:
		if ts, err := convIterStrings(v); err != nil {
			return err
		} else {
			return WriteFileLines(name, ts)
		}
	case *starlark.Set:
		if ss, err := convIterStrings(v); err != nil {
			return err
		} else {
			return WriteFileLines(name, ss)
		}
	default:
		return fmt.Errorf(ModuleName+`.write_lines: expected list/tuple/set, got %s`, data.Type())
	}
}

func convIterStrings(lst starlark.Iterable) ([]string, error) {
	var lines []string
	iter := lst.Iterate()
	defer iter.Done()

	var x starlark.Value
	for iter.Next(&x) {
		if s, ok := starlark.AsString(x); ok {
			lines = append(lines, s)
		} else {
			lines = append(lines, x.String())
		}
	}
	return lines, nil
}
