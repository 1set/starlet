package file

import (
	"fmt"
	"strings"

	"github.com/1set/starlet/dataconv"
	"go.starlark.net/starlark"
)

// readJSON reads the whole named file and decodes the contents as JSON for Starlark.
func readJSON(name string) (starlark.Value, error) {
	data, err := ReadFileBytes(name)
	if err != nil {
		return nil, err
	}
	return dataconv.DecodeStarlarkJSON(data)
}

// readJSONL reads the whole named file and decodes the contents as JSON lines for Starlark.
func readJSONL(name string) (starlark.Value, error) {
	var (
		cnt    int
		values []starlark.Value
	)
	if err := readFileByLine(name, func(line string) error {
		cnt++
		// skip empty lines
		if strings.TrimSpace(line) == emptyStr {
			return nil
		}
		// convert to Starlark value
		v, err := dataconv.DecodeStarlarkJSON([]byte(line))
		if err != nil {
			return fmt.Errorf("line %d: %w", cnt, err)
		}
		values = append(values, v)
		return nil
	}); err != nil {
		return nil, err
	}
	return starlark.NewList(values), nil
}

// writeJSON writes the given JSON as string into a file.
func writeJSON(name, funcName string, override bool, data starlark.Value) error {
	wf := AppendFileString
	if override {
		wf = WriteFileString
	}
	// treat starlark.Bytes and starlark.String as the same type, just convert to string, for other types, encode to JSON
	switch v := data.(type) {
	case starlark.Bytes:
		return wf(name, string(v))
	case starlark.String:
		return wf(name, string(v))
	default:
		// convert to JSON
		s, err := dataconv.EncodeStarlarkJSON(v)
		if err != nil {
			return err
		}
		return wf(name, s)
	}
}

// writeJSONL writes the given JSON lines into a file.
func writeJSONL(name, funcName string, override bool, data starlark.Value) error {
	wf := AppendFileLines
	if override {
		wf = WriteFileLines
	}

	// handle all types of iterable, and allow string or bytes, for other types, encode to lines of JSON
	var (
		ls  []string
		err error
	)
	switch v := data.(type) {
	case starlark.String:
		return wf(name, []string{v.GoString()})
	case starlark.Bytes:
		return wf(name, []string{string(v)})
	case *starlark.List:
		ls, err = convIterJSONL(v)
	case starlark.Tuple:
		ls, err = convIterJSONL(v)
	case *starlark.Set:
		ls, err = convIterJSONL(v)
	default:
		// convert to JSON
		s, err := dataconv.EncodeStarlarkJSON(v)
		if err != nil {
			return err
		}
		return wf(name, []string{s})
	}
	if err != nil {
		return err
	}

	// write lines
	return wf(name, ls)
}

func convIterJSONL(lst starlark.Iterable) (lines []string, err error) {
	iter := lst.Iterate()
	defer iter.Done()

	var (
		s string
		x starlark.Value
	)
	for iter.Next(&x) {
		s, err = dataconv.EncodeStarlarkJSON(x)
		if err != nil {
			return
		}
		lines = append(lines, s)
	}
	return
}
