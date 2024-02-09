package dataconv

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/1set/starlight/convert"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

var (
	emptyStr string
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
// It first converts the starlark.Value into a Golang value, then marshals it into JSON.
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

// TypingConvert converts all values to their appropriate types.
func TypingConvert(data interface{}) interface{} {
	switch v := data.(type) {
	case string:
		// If the string is a valid time, return a time.Time.
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			return t
		}
		return v
	case float64:
		// If the float is actually an int, return an int.
		if v == float64(int(v)) {
			return int(v)
		}
		return v
	case map[string]interface{}:
		// If the value is a map, recursively call this function on all map values.
		newMap := make(map[string]interface{})
		for key, value := range v {
			newMap[key] = TypingConvert(value)
		}
		return newMap
	case []interface{}:
		// If the value is a slice, recursively call this function on all slice values.
		newSlice := make([]interface{}, len(v))
		for i, value := range v {
			newSlice[i] = TypingConvert(value)
		}
		return newSlice
	default:
		// Otherwise, just return the value.
		return v
	}
}

/*
func TypeConvert(data interface{}) interface{} {
	switch v := data.(type) {
	case string:
		// Attempt parsing in different formats
		for _, format := range []string{time.RFC3339, time.RFC3339Nano, time.RFC822, time.RFC1123} {
			if t, err := time.Parse(format, v); err == nil {
				return t
			}
		}
		// Attempt to parse as JSON number (int or float)
		var num json.Number
		if err := json.Unmarshal([]byte(v), &num); err == nil {
			nf, ef := num.Float64()
			ni, ei := num.Int64()
			if ef == nil && ei == nil {
				if int64(nf) == ni {
					return ni
				}
				return nf
			} else if ef == nil {
				return nf
			} else if ei == nil {
				return ni
			}
		}
		// If not a time or number, return the original string
		return v

	case float64:
		// Check for exact int match
		// if float64(int64(v)) == v {
		if math.Floor(v) == v {
			return int64(v)
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
*/
