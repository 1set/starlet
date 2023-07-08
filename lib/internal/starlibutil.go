package internal

// Based on https://github.com/qri-io/starlib/tree/master/util with some modifications and additions

import (
	"fmt"
	"reflect"
	"strconv"
	"time"

	"github.com/1set/starlight/convert"
	startime "go.starlark.net/lib/time"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

// asString unquotes a starlark string value
func asString(x starlark.Value) (string, error) {
	return strconv.Unquote(x.String())
}

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

// Unmarshal decodes a starlark.Value into it's Golang counterpart.
func Unmarshal(x starlark.Value) (val interface{}, err error) {
	switch v := x.(type) {
	case starlark.NoneType:
		val = nil
	case starlark.Bool:
		val = v.Truth() == starlark.True
	case starlark.Int:
		var tmp int
		err = starlark.AsInt(x, &tmp)
		val = tmp
	case starlark.Float:
		if f, ok := starlark.AsFloat(x); !ok {
			err = fmt.Errorf("couldn't parse float")
		} else {
			val = f
		}
	case starlark.String:
		val = v.GoString()
	case startime.Time:
		val = time.Time(v)
	case *starlark.Dict:
		var (
			dictVal starlark.Value
			pval    interface{}
			kval    interface{}
			keys    []interface{}
			vals    []interface{}
			// key as interface if found one key is not a string
			ki bool
		)

		for _, k := range v.Keys() {
			dictVal, _, err = v.Get(k)
			if err != nil {
				return
			}

			pval, err = Unmarshal(dictVal)
			if err != nil {
				err = fmt.Errorf("unmarshaling starlark value: %w", err)
				return
			}

			kval, err = Unmarshal(k)
			if err != nil {
				err = fmt.Errorf("unmarshaling starlark key: %w", err)
				return
			}

			if _, ok := kval.(string); !ok {
				// found key as not a string
				ki = true
			}

			keys = append(keys, kval)
			vals = append(vals, pval)
		}

		// prepare result

		rs := map[string]interface{}{}
		ri := map[interface{}]interface{}{}

		for i, key := range keys {
			// key as interface
			if ki {
				ri[key] = vals[i]
			} else {
				rs[key.(string)] = vals[i]
			}
		}

		if ki {
			val = ri // map[interface{}]interface{}
		} else {
			val = rs // map[string]interface{}
		}
	case *starlark.List:
		var (
			i       int
			listVal starlark.Value
			iter    = v.Iterate()
			value   = make([]interface{}, v.Len())
		)

		defer iter.Done()
		for iter.Next(&listVal) {
			value[i], err = Unmarshal(listVal)
			if err != nil {
				return
			}
			i++
		}
		val = value
	case starlark.Tuple:
		var (
			i        int
			tupleVal starlark.Value
			iter     = v.Iterate()
			value    = make([]interface{}, v.Len())
		)

		defer iter.Done()
		for iter.Next(&tupleVal) {
			value[i], err = Unmarshal(tupleVal)
			if err != nil {
				return
			}
			i++
		}
		val = value
	case *starlark.Set:
		var (
			i      int
			setVal starlark.Value
			iter   = v.Iterate()
			value  = make([]interface{}, v.Len())
		)

		defer iter.Done()
		for iter.Next(&setVal) {
			value[i], err = Unmarshal(setVal)
			if err != nil {
				return
			}
			i++
		}
		val = value
	case *starlarkstruct.Struct:
		if _var, ok := v.Constructor().(Unmarshaler); ok {
			err = _var.UnmarshalStarlark(x)
			if err != nil {
				err = fmt.Errorf("failed marshal %q to Starlark object: %w", v.Constructor().Type(), err)
				return
			}
			val = _var
		} else {
			err = fmt.Errorf("constructor object from *starlarkstruct.Struct not supported Marshaler to starlark object: %s", v.Constructor().Type())
		}
	case *convert.GoSlice:
		if IsInterfaceNil(v) {
			err = fmt.Errorf("nil GoSlice")
			return
		}
		val = v.Value().Interface()
	case *convert.GoMap:
		if IsInterfaceNil(v) {
			err = fmt.Errorf("nil GoMap")
			return
		}
		val = v.Value().Interface()
	case *convert.GoStruct:
		if IsInterfaceNil(v) {
			err = fmt.Errorf("nil GoStruct")
			return
		}
		val = v.Value().Interface()
	case *convert.GoInterface:
		if IsInterfaceNil(v) {
			err = fmt.Errorf("nil GoInterface")
			return
		}
		val = v.Value().Interface()
	default:
		//fmt.Println("errbadtype:", x.Type())
		//err = fmt.Errorf("unrecognized starlark type: %s", x.Type())
		err = fmt.Errorf("unrecognized starlark type: %T", x)
	}
	return
}

// Marshal turns go values into Starlark types.
func Marshal(data interface{}) (v starlark.Value, err error) {
	switch x := data.(type) {
	case nil:
		v = starlark.None
	case bool:
		v = starlark.Bool(x)
	case string:
		v = starlark.String(x)
	case int:
		v = starlark.MakeInt(x)
	case int8:
		v = starlark.MakeInt(int(x))
	case int16:
		v = starlark.MakeInt(int(x))
	case int32:
		v = starlark.MakeInt(int(x))
	case int64:
		v = starlark.MakeInt64(x)
	case uint:
		v = starlark.MakeUint(x)
	case uint8:
		v = starlark.MakeUint(uint(x))
	case uint16:
		v = starlark.MakeUint(uint(x))
	case uint32:
		v = starlark.MakeUint(uint(x))
	case uint64:
		v = starlark.MakeUint64(x)
	case float32:
		v = starlark.Float(float64(x))
	case float64:
		v = starlark.Float(x)
	case time.Time:
		v = startime.Time(x)
	case []interface{}:
		var elems = make([]starlark.Value, len(x))
		for i, val := range x {
			elems[i], err = Marshal(val)
			if err != nil {
				return
			}
		}
		v = starlark.NewList(elems)
	case map[interface{}]interface{}:
		dict := &starlark.Dict{}
		var elem starlark.Value
		for ki, val := range x {
			var key starlark.Value
			key, err = Marshal(ki)
			if err != nil {
				return
			}

			elem, err = Marshal(val)
			if err != nil {
				return
			}
			if err = dict.SetKey(key, elem); err != nil {
				return
			}
		}
		v = dict
	case map[string]interface{}:
		dict := &starlark.Dict{}
		var elem starlark.Value
		for key, val := range x {
			elem, err = Marshal(val)
			if err != nil {
				return
			}
			if err = dict.SetKey(starlark.String(key), elem); err != nil {
				return
			}
		}
		v = dict
	case Marshaler:
		v, err = x.MarshalStarlark()
	default:
		return starlark.None, fmt.Errorf("unrecognized type: %#v", x)
	}
	return
}

// Unmarshaler is the interface use to unmarshal Starlark custom types.
type Unmarshaler interface {
	// UnmarshalStarlark unmarshal a starlark object to custom type.
	UnmarshalStarlark(starlark.Value) error
}

// Marshaler is the interface use to marshal Starlark custom types.
type Marshaler interface {
	// MarshalStarlark marshal a custom type to starlark object.
	MarshalStarlark() (starlark.Value, error)
}
