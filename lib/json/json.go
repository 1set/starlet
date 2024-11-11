// Package json defines utilities for converting Starlark values to/from JSON strings based on go.starlark.net/lib/json.
package json

import (
	"bytes"
	"encoding/json"
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
	none       = starlark.None
	once       sync.Once
	jsonModule starlark.StringDict
)

// LoadModule loads the json module. It is concurrency-safe and idempotent.
func LoadModule() (starlark.StringDict, error) {
	once.Do(func() {
		mod := starlarkstruct.Module{
			Name: ModuleName,
			Members: starlark.StringDict{
				"dumps":      starlark.NewBuiltin(ModuleName+".dumps", dumps),
				"try_dumps":  starlark.NewBuiltin(ModuleName+".try_dumps", tryDumps),
				"try_encode": starlark.NewBuiltin(ModuleName+".try_encode", tryEncode),
				"try_decode": starlark.NewBuiltin(ModuleName+".try_decode", tryDecode),
				"try_indent": starlark.NewBuiltin(ModuleName+".try_indent", tryIndent),
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
		return none, err
	}

	// use 0 as default indent if failed to unpack indent
	it, ok := indent.Int64()
	if !ok || it < 0 {
		it = 0
	}

	// use internal marshaler to support starlark types
	data, err := itn.MarshalStarlarkJSON(obj, int(it))
	if err != nil {
		return none, err
	}
	return starlark.String(data), nil
}

func tryDumps(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var (
		obj    starlark.Value
		indent = starlark.MakeInt(0)
	)
	if err := starlark.UnpackArgs(fn.Name(), args, kwargs, "obj", &obj, "indent?", &indent); err != nil {
		return starlark.Tuple{none, starlark.String(err.Error())}, nil
	}

	it, ok := indent.Int64()
	if !ok || it < 0 {
		it = 0
	}

	data, err := itn.MarshalStarlarkJSON(obj, int(it))
	if err != nil {
		return starlark.Tuple{none, starlark.String(err.Error())}, nil
	}
	return starlark.Tuple{starlark.String(data), none}, nil
}

func tryEncode(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var x starlark.Value
	if err := starlark.UnpackPositionalArgs(fn.Name(), args, kwargs, 1, &x); err != nil {
		return starlark.Tuple{none, starlark.String(err.Error())}, nil
	}

	encoded, err := itn.EncodeStarlarkJSON(x)
	if err != nil {
		return starlark.Tuple{none, starlark.String(err.Error())}, nil
	}
	return starlark.Tuple{starlark.String(encoded), none}, nil
}

func tryDecode(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var s string
	if err := starlark.UnpackArgs(fn.Name(), args, kwargs, "x", &s); err != nil {
		return starlark.Tuple{none, starlark.String(err.Error())}, nil
	}

	decoded, err := itn.DecodeStarlarkJSON([]byte(s))
	if err != nil {
		return starlark.Tuple{none, starlark.String(err.Error())}, nil
	}
	return starlark.Tuple{decoded, none}, nil
}

func tryIndent(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var (
		str    string
		prefix string
		indent string
	)
	if err := starlark.UnpackArgs(fn.Name(), args, kwargs, "str", &str, "prefix?", &prefix, "indent?", &indent); err != nil {
		return starlark.Tuple{none, starlark.String(err.Error())}, nil
	}

	buf := new(bytes.Buffer)
	if err := json.Indent(buf, []byte(str), prefix, indent); err != nil {
		return starlark.Tuple{none, starlark.String(err.Error())}, nil
	}
	return starlark.Tuple{starlark.String(buf.String()), none}, nil
}
