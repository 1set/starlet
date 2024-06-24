package file

import (
	"fmt"
	"github.com/1set/starlet/dataconv"
	stdjson "go.starlark.net/lib/json"
	"go.starlark.net/starlark"
)

// readJSON reads the whole named file and decodes the contents as JSON for Starlark.
func readJSON(name string) (starlark.Value, error) {
	data, err := ReadFileBytes(name)
	if err != nil {
		return nil, err
	}
	return starlarkJSONDecode(data)
}

// starlarkJSONDecode decodes the JSON bytes into a Starlark value via standard JSON module from Starlark.
func starlarkJSONDecode(data []byte) (starlark.Value, error) {
	// get the JSON decoder
	jm, ok := stdjson.Module.Members["decode"]
	if !ok {
		return nil, fmt.Errorf("json.decode not found")
	}
	dec := jm.(*starlark.Builtin)

	// convert from JSON
	return starlark.Call(&starlark.Thread{Name: "file_module"}, dec, starlark.Tuple{starlark.String(data)}, nil)
}

// starlarkJSONEncode encodes the Starlark value into a JSON string via standard JSON module from Starlark.
func starlarkJSONEncode(v starlark.Value) (string, error) {
	// get the JSON encoder
	jm, ok := stdjson.Module.Members["encode"]
	if !ok {
		return emptyStr, fmt.Errorf("json.encode not found")
	}
	enc := jm.(*starlark.Builtin)

	// convert to JSON
	v, err := starlark.Call(&starlark.Thread{Name: "file_module"}, enc, starlark.Tuple{v}, nil)
	if err != nil {
		return emptyStr, err
	}

	// convert to string
	return dataconv.StarString(v), nil
}
