// Package json defines utilities for converting Starlark values to/from JSON strings based on go.starlark.net/lib/json.
package json

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"sync"

	itn "github.com/1set/starlet/dataconv"
	"github.com/spyzhov/ajson"
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
				"path":       starlark.NewBuiltin(ModuleName+".path", generateJsonPath(false)),
				"try_path":   starlark.NewBuiltin(ModuleName+".try_path", generateJsonPath(true)),
				"eval":       starlark.NewBuiltin(ModuleName+".eval", generateJsonEval(false)),
				"try_eval":   starlark.NewBuiltin(ModuleName+".try_eval", generateJsonEval(true)),
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

// generateJsonPath generates a Starlark function that performs a JSONPath query on the given JSON data and returns the matching elements.
func generateJsonPath(try bool) itn.StarlarkFunc {
	return func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (res starlark.Value, err error) {
		var (
			data     starlark.Value
			pathExpr string
		)
		if err := starlark.UnpackArgs(fn.Name(), args, kwargs, "data", &data, "path", &pathExpr); err != nil {
			if try {
				return starlark.Tuple{starlark.None, starlark.String(err.Error())}, nil
			}
			return none, err
		}

		jb, err := getJsonBytes(data)
		if err != nil {
			if try {
				return starlark.Tuple{starlark.None, starlark.String(err.Error())}, nil
			}
			return none, fmt.Errorf("json.path: %w", err)
		}

		nodes, err := ajson.JSONPath(jb, pathExpr)
		if err != nil {
			if try {
				return starlark.Tuple{starlark.None, starlark.String(err.Error())}, nil
			}
			return none, fmt.Errorf("json.path: %w", err)
		}

		results, err := ajsonNodesToStarlarkList(nodes)
		if err != nil {
			if try {
				return starlark.Tuple{starlark.None, starlark.String(err.Error())}, nil
			}
			return none, fmt.Errorf("json.path: %w", err)
		}

		if try {
			return starlark.Tuple{results, starlark.None}, nil
		}
		return results, nil
	}
}

// generateJsonEval generates a Starlark function that evaluates a JSONPath query with an expression on the given JSON data and returns the evaluation result.
func generateJsonEval(try bool) itn.StarlarkFunc {
	return func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (res starlark.Value, err error) {
		var (
			data starlark.Value
			expr string
		)
		if err := starlark.UnpackArgs(fn.Name(), args, kwargs, "data", &data, "expr", &expr); err != nil {
			if try {
				return starlark.Tuple{starlark.None, starlark.String(err.Error())}, nil
			}
			return none, err
		}

		jb, err := getJsonBytes(data)
		if err != nil {
			if try {
				return starlark.Tuple{starlark.None, starlark.String(err.Error())}, nil
			}
			return none, fmt.Errorf("json.eval: %w", err)
		}

		root, err := ajson.Unmarshal(jb)
		if err != nil {
			if try {
				return starlark.Tuple{starlark.None, starlark.String(err.Error())}, nil
			}
			return none, fmt.Errorf("json.eval: %w", err)
		}

		result, err := ajson.Eval(root, expr)
		if err != nil {
			if try {
				return starlark.Tuple{starlark.None, starlark.String(err.Error())}, nil
			}
			return none, fmt.Errorf("json.eval: %w", err)
		}

		val, err := ajsonNodeToStarlarkValue(result)
		if err != nil {
			if try {
				return starlark.Tuple{starlark.None, starlark.String(err.Error())}, nil
			}
			return none, fmt.Errorf("json.eval: %w", err)
		}

		if try {
			return starlark.Tuple{val, starlark.None}, nil
		}
		return val, nil
	}
}

// getJsonBytes converts a Starlark value to a JSON byte slice.
func getJsonBytes(data starlark.Value) ([]byte, error) {
	switch v := data.(type) {
	case starlark.String:
		return []byte(v.GoString()), nil
	case starlark.Bytes:
		return []byte(v), nil
	default:
		js, err := itn.MarshalStarlarkJSON(data, 0)
		if err != nil {
			return nil, err
		}
		return []byte(js), nil
	}
}

// ajsonNodesToStarlarkList converts a slice of ajson.Node to a Starlark list.
func ajsonNodesToStarlarkList(nodes []*ajson.Node) (starlark.Value, error) {
	results := make([]starlark.Value, 0, len(nodes))
	for _, node := range nodes {
		val, err := ajsonNodeToStarlarkValue(node)
		if err != nil {
			return nil, err
		}
		results = append(results, val)
	}
	return starlark.NewList(results), nil
}

// ajsonNodeToStarlarkValue converts an ajson.Node to a Starlark value.
// It recursively traverses the node tree and constructs the corresponding Starlark values.
func ajsonNodeToStarlarkValue(node *ajson.Node) (starlark.Value, error) {
	switch node.Type() {
	case ajson.Object:
		dict := &starlark.Dict{}
		for _, key := range node.Keys() {
			valNode, err := node.GetKey(key)
			if err != nil {
				return nil, err
			}
			val, err := ajsonNodeToStarlarkValue(valNode)
			if err != nil {
				return nil, err
			}
			err = dict.SetKey(starlark.String(key), val)
			if err != nil {
				return nil, err
			}
		}
		return dict, nil
	case ajson.Array:
		elements, err := node.GetArray()
		if err != nil {
			return nil, err
		}
		vals := make([]starlark.Value, len(elements))
		for i, elem := range elements {
			val, err := ajsonNodeToStarlarkValue(elem)
			if err != nil {
				return nil, err
			}
			vals[i] = val
		}
		return starlark.NewList(vals), nil
	case ajson.String:
		return starlark.String(node.MustString()), nil
	case ajson.Numeric:
		num := node.MustNumeric()
		if math.Mod(num, 1.0) == 0 {
			return starlark.MakeInt64(int64(num)), nil
		} else {
			return starlark.Float(num), nil
		}
	case ajson.Bool:
		return starlark.Bool(node.MustBool()), nil
	case ajson.Null:
		return starlark.None, nil
	default:
		return nil, fmt.Errorf("unsupported JSON node type: %v", node.Type())
	}
}
