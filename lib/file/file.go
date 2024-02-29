// Package file defines functions that manipulate files, it's inspired by file helpers from Amoy.
package file

import (
	"fmt"
	"sync"

	dc "github.com/1set/starlet/dataconv"
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
					"trim_bom":      starlark.NewBuiltin(ModuleName+".trim_bom", trimBom),
					"count_lines":   starlark.NewBuiltin(ModuleName+".count_lines", countLinesInFile),
					"head_lines":    readTopOrBottomLines("head_lines", ReadFirstLines),
					"tail_lines":    readTopOrBottomLines("tail_lines", ReadLastLines),
					"read_bytes":    wrapReadFile("read_bytes", readBytes),
					"read_string":   wrapReadFile("read_string", readString),
					"read_lines":    wrapReadFile("read_lines", readLines),
					"write_bytes":   wrapWriteFile("write_bytes", true, writeBytes),
					"write_string":  wrapWriteFile("write_string", true, writeString),
					"write_lines":   wrapWriteFile("write_lines", true, writeLines),
					"append_bytes":  wrapWriteFile("append_bytes", false, writeBytes),
					"append_string": wrapWriteFile("append_string", false, writeString),
					"append_lines":  wrapWriteFile("append_lines", false, writeLines),
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

// countLinesInFile counts the number of lines in a file.
func countLinesInFile(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var fp itn.StringOrBytes
	if err := starlark.UnpackArgs(b.Name(), args, kwargs, "name", &fp); err != nil {
		return starlark.None, err
	}
	// do the work
	cnt, err := CountFileLines(fp.GoString())
	if err != nil {
		return nil, err
	}
	return starlark.MakeInt(cnt), nil
}

// readTopOrBottomLines wraps the file reading functions for top or bottom lines to be used in Starlark.
func readTopOrBottomLines(funcName string, workLoad func(name string, n int) ([]string, error)) starlark.Callable {
	return starlark.NewBuiltin(ModuleName+"."+funcName, func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		// unpack arguments
		var (
			fp itn.StringOrBytes
			n  starlark.Int
		)
		if err := starlark.UnpackArgs(b.Name(), args, kwargs, "name", &fp, "n", &n); err != nil {
			return starlark.None, err
		}
		nInt, _ := n.Int64()
		if nInt <= 0 {
			return starlark.None, fmt.Errorf(`%s: expected positive integer, got %d`, b.Name(), n)
		}

		// read lines
		ls, err := workLoad(fp.GoString(), int(nInt))
		if err != nil {
			return nil, err
		}

		// return as list
		sl := make([]starlark.Value, len(ls))
		for i, l := range ls {
			sl[i] = starlark.String(l)
		}
		return starlark.NewList(sl), nil
	})
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
func wrapWriteFile(funcName string, override bool, workLoad func(name, funcName string, override bool, data starlark.Value) error) starlark.Callable {
	fullName := ModuleName + "." + funcName
	return starlark.NewBuiltin(fullName, func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		var (
			fp   itn.StringOrBytes
			data starlark.Value
		)
		if err := starlark.UnpackArgs(b.Name(), args, kwargs, "name", &fp, "data", &data); err != nil {
			return starlark.None, err
		}
		return starlark.None, workLoad(fp.GoString(), fullName, override, data)
	})
}

// writeBytes writes the given data in bytes into a file.
func writeBytes(name, funcName string, override bool, data starlark.Value) error {
	wf := AppendFileBytes
	if override {
		wf = WriteFileBytes
	}
	// treat starlark.Bytes and starlark.String as the same type
	switch v := data.(type) {
	case starlark.Bytes:
		return wf(name, []byte(v))
	case starlark.String:
		return wf(name, []byte(v))
	default:
		return fmt.Errorf(`%s: expected string or bytes, got %s`, funcName, data.Type())
	}
}

// writeString writes the given data in string into a file.
func writeString(name, funcName string, override bool, data starlark.Value) error {
	wf := AppendFileString
	if override {
		wf = WriteFileString
	}
	// treat starlark.Bytes and starlark.String as the same type
	switch v := data.(type) {
	case starlark.Bytes:
		return wf(name, string(v))
	case starlark.String:
		return wf(name, string(v))
	default:
		return fmt.Errorf(`%s: expected string or bytes, got %s`, funcName, data.Type())
	}
}

// writeLines writes the lines into a file. The data should be a list, a tuple or a set of strings.
func writeLines(name, funcName string, override bool, data starlark.Value) error {
	wf := AppendFileLines
	if override {
		wf = WriteFileLines
	}
	// handle all types of iterable, and allow string or bytes
	switch v := data.(type) {
	case starlark.String:
		return wf(name, []string{v.GoString()})
	case starlark.Bytes:
		return wf(name, []string{string(v)})
	case *starlark.List:
		return wf(name, convIterStrings(v))
	case starlark.Tuple:
		return wf(name, convIterStrings(v))
	case *starlark.Set:
		return wf(name, convIterStrings(v))
	default:
		return fmt.Errorf(`%s: expected list/tuple/set, got %s`, funcName, data.Type())
	}
}

func convIterStrings(lst starlark.Iterable) (lines []string) {
	iter := lst.Iterate()
	defer iter.Done()

	var x starlark.Value
	for iter.Next(&x) {
		lines = append(lines, dc.StarString(x))
	}
	return
}
