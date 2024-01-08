package dataconv

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"go.starlark.net/starlark"
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
