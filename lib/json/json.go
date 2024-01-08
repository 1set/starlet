// Package json defines utilities for converting Starlark values to/from JSON strings based on go.starlark.net/lib/json.
package json

import (
	"sync"

	itn "github.com/1set/starlet/dataconv"
	stdjson "go.starlark.net/lib/json"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

// ModuleName defines the expected name for this Module when used
// in starlark's load() function, eg: load('json', 'encode')
const ModuleName = "json"

var (
	once       sync.Once
	jsonModule starlark.StringDict
)

// LoadModule loads the time module. It is concurrency-safe and idempotent.
func LoadModule() (starlark.StringDict, error) {
	once.Do(func() {
		mod := starlarkstruct.Module{
			Name: ModuleName,
			Members: starlark.StringDict{
				"dumps": starlark.NewBuiltin("json.dumps", dumps),
			},
		}
		for k, v := range stdjson.Module.Members {
			mod.Members[k] = v
		}
		jsonModule = starlark.StringDict{
			ModuleName: &mod,
		}
	})
	return jsonModule, nil
}

func dumps(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var (
		obj    starlark.Value
		indent = starlark.MakeInt(0)
	)
	if err := starlark.UnpackArgs(fn.Name(), args, kwargs, "obj", &obj, "indent?", &indent); err != nil {
		return starlark.None, err
	}

	// use 0 as default indent if failed to unpack indent
	it, ok := indent.Int64()
	if !ok || it < 0 {
		it = 0
	}

	// use internal marshaler to support starlark types
	data, err := itn.MarshalStarlarkJSON(obj, int(it))
	if err != nil {
		return starlark.None, err
	}
	return starlark.String(data), nil
}
