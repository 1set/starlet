package dataconv

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"strings"
	"time"

	"github.com/1set/starlight/convert"
	stdjson "go.starlark.net/lib/json"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

var (
	emptyStr       string
	noopPrintFunc  = func(thread *starlark.Thread, msg string) {}
	starJSONEncode = stdjson.Module.Members["encode"].(*starlark.Builtin)
	starJSONDecode = stdjson.Module.Members["decode"].(*starlark.Builtin)
)

// IsEmptyString checks is a starlark string is empty ("" for a go string)
// starlark.String.String performs repr-style quotation, which is necessary
// for the starlark.Value contract but a frequent source of errors in API
// clients. This helper method makes sure it'll work properly
func IsEmptyString(s starlark.String) bool {
	return s.String() == `""`
}

// IsInterfaceNil returns true if the given interface is nil.
func IsInterfaceNil(i interface{}) bool {
	if i == nil {
		return true
	}
	defer func() { recover() }()
	switch reflect.TypeOf(i).Kind() {
	case reflect.Ptr, reflect.UnsafePointer, reflect.Interface, reflect.Struct, reflect.Slice, reflect.Map, reflect.Chan, reflect.Func:
		return reflect.ValueOf(i).IsNil()
	}
	return false
}

// MarshalStarlarkJSON marshals a starlark.Value into a JSON string.
// It first converts the starlark.Value into a Go value using the Unmarshal function,
// then marshals the Go value into a JSON string.
// It handles Go-specific types (supports more types like Go slices, maps, structs, interfaces, time) better than EncodeStarlarkJSON but may be slower due to intermediate conversion.
func MarshalStarlarkJSON(data starlark.Value, indent int) (string, error) {
	// convert starlark value to a go value
	v, err := Unmarshal(data)
	if err != nil {
		return emptyStr, err
	}

	// fix map[interface {}]interface {}
	if m, ok := v.(map[interface{}]interface{}); ok {
		mm := make(map[string]interface{})
		for k, v := range m {
			mm[fmt.Sprintf("%v", k)] = v
		}
		v = mm
	}

	// prepare json encoder
	var bf bytes.Buffer
	enc := json.NewEncoder(&bf)
	enc.SetEscapeHTML(false)
	if indent > 0 {
		enc.SetIndent("", strings.Repeat(" ", indent))
	}

	// convert go to string
	if err = enc.Encode(v); err != nil {
		return emptyStr, err
	}
	return strings.TrimSpace(bf.String()), nil
}

// UnmarshalStarlarkJSON unmarshals JSON bytes into a starlark.Value.
// It first unmarshals the JSON bytes into a Go value using the standard JSON Unmarshal function,
// then converts the Go value into a starlark.Value using the Marshal function.
// Time strings are parsed as starlark Time objects.
// In comparison with DecodeStarlarkJSON, it gives you more control over type conversion but may be less efficient due to intermediate steps.
func UnmarshalStarlarkJSON(data []byte) (starlark.Value, error) {
	var m interface{}
	err := json.Unmarshal(data, &m)
	if err != nil {
		return starlark.None, err
	}

	// fix all values to their appropriate types
	f := TypeConvert(m)

	// convert go value to a starlark value
	return Marshal(f)
}

// EncodeStarlarkJSON encodes a starlark.Value into a JSON string using the official JSON module of Starlark Go.
// It retrieves the "json.encode" function from the Starlark standard library and calls it to perform the encoding.
// It's less flexible but more straightforward for pure Starlark values.
// Compared to the Marshal/Unmarshal* pair, which is more feature-rich and flexible, the Encode/Decode* pair is more focused on core Starlark types.
func EncodeStarlarkJSON(v starlark.Value) (string, error) {
	// convert to JSON
	v, err := starlark.Call(&starlark.Thread{Name: "dataconv", Print: noopPrintFunc}, starJSONEncode, starlark.Tuple{v}, nil)
	if err != nil {
		return emptyStr, err
	}

	// convert to string
	return StarString(v), nil
}

// DecodeStarlarkJSON decodes JSON bytes into a starlark.Value using the official JSON module of Starlark Go.
// It retrieves the "json.decode" function from the Starlark standard library and calls it to perform the decoding.
// It handles certain Starlark-specific structures better but less control over Go-specific types.
// Compared to the Marshal/Unmarshal* pair, which is more feature-rich and flexible, the Encode/Decode* pair is more focused on core Starlark types.
func DecodeStarlarkJSON(data []byte) (starlark.Value, error) {
	return starlark.Call(&starlark.Thread{Name: "dataconv", Print: noopPrintFunc}, starJSONDecode, starlark.Tuple{starlark.String(data)}, nil)
}

// ConvertStruct converts a struct to a Starlark wrapper.
func ConvertStruct(v interface{}, tagName string) starlark.Value {
	// if not a pointer, convert to pointer
	if reflect.TypeOf(v).Kind() != reflect.Ptr {
		panic("v must be a pointer")
	}
	return convert.NewStructWithTag(v, tagName)
}

// ConvertJSONStruct converts a struct to a Starlark wrapper with JSON tag.
func ConvertJSONStruct(v interface{}) starlark.Value {
	return ConvertStruct(v, "json")
}

// WrapModuleData wraps data from the given starlark.StringDict into a Starlark module loader, which can be used to load the module into a Starlark interpreter and accessed via `load("name", "data_key")`.
func WrapModuleData(name string, data starlark.StringDict) func() (starlark.StringDict, error) {
	return func() (starlark.StringDict, error) {
		return starlark.StringDict{
			name: &starlarkstruct.Module{
				Name:    name,
				Members: data,
			},
		}, nil
	}
}

// WrapStructData wraps data from the given starlark.StringDict into a Starlark struct loader, which can be used to load the struct into a Starlark interpreter and accessed via `load("name", "data_key")`.
func WrapStructData(name string, data starlark.StringDict) func() (starlark.StringDict, error) {
	return func() (starlark.StringDict, error) {
		ss := starlarkstruct.FromStringDict(starlark.String(name), data)
		return starlark.StringDict{name: ss}, nil
	}
}

// MakeModule creates a Starlark module from the given name and data.
func MakeModule(name string, data starlark.StringDict) *starlarkstruct.Module {
	return &starlarkstruct.Module{
		Name:    name,
		Members: data,
	}
}

// MakeStruct creates a Starlark struct from the given name and data.
func MakeStruct(name string, data starlark.StringDict) *starlarkstruct.Struct {
	return starlarkstruct.FromStringDict(starlark.String(name), data)
}

// TypeConvert converts JSON decoded values to their appropriate types.
// Usually it's used after JSON Unmarshal, Starlark Unmarshal, or similar.
func TypeConvert(data interface{}) interface{} {
	switch v := data.(type) {
	case string:
		// Attempt parsing in different formats
		for _, format := range []string{time.RFC3339, time.RFC3339Nano, time.RFC822, time.RFC1123} {
			if t, err := time.Parse(format, v); err == nil {
				return t
			}
		}
		// Attempt to parse as JSON number (integer only)
		var num json.Number
		if err := json.Unmarshal([]byte(v), &num); err == nil {
			if ni, ei := num.Int64(); ei == nil {
				return ni
			}
		}
		// If not a time or number, return the original string
		return v

	case float64:
		// Check for exact int match
		if math.Floor(v) == v {
			return int(v)
		}
		return v

	case map[string]interface{}:
		// If the value is a map, recursively call this function on all map values.
		newMap := make(map[string]interface{}, len(v))
		for key, value := range v {
			newMap[key] = TypeConvert(value)
		}
		return newMap

	case []interface{}:
		// If the value is a slice, recursively call this function on all slice values.
		newSlice := make([]interface{}, len(v))
		for i, value := range v {
			newSlice[i] = TypeConvert(value)
		}
		return newSlice

	default:
		// Return original value for other types
		return v
	}
}

// StarString returns the string representation of a starlark.Value, i.e. converts Starlark values to Go strings.
func StarString(x starlark.Value) string {
	if IsInterfaceNil(x) {
		return emptyStr
	}
	switch v := x.(type) {
	case starlark.String:
		return v.GoString()
	case starlark.Bytes:
		return string(v)
	default:
		return v.String()
	}
}

// GoToStarlarkViaJSON converts any Go value that can be marshaled into JSON into a corresponding Starlark value.
// It first marshals the Go value into a JSON string with the standard Go JSON package,
// then decodes this JSON string into a Starlark value with the official Starlark JSON module.
func GoToStarlarkViaJSON(v interface{}) (starlark.Value, error) {
	if v == nil {
		return starlark.None, nil
	}
	bs, err := json.Marshal(v)
	if err != nil {
		return starlark.None, err
	}
	return DecodeStarlarkJSON(bs)
}

// GetThreadContext returns the context of the given thread, or new context if not found.
func GetThreadContext(thread *starlark.Thread) context.Context {
	if thread != nil {
		if c := thread.Local("context"); c != nil {
			if co, ok := c.(context.Context); ok {
				return co
			}
		}
	}
	return context.Background()
}
