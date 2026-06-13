package dataconv

// Based on https://github.com/qri-io/starlib/tree/master/util with some modifications and additions

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/1set/starlight/convert"
	startime "go.starlark.net/lib/time"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

// Marshal converts Go values into Starlark types, like ToValue() of package starlight does.
// It only supports common Go types, won't wrap any custom types like Starlight does.
func Marshal(data interface{}) (v starlark.Value, err error) {
	switch x := data.(type) {
	case nil:
		v = starlark.None
	case bool:
		v = starlark.Bool(x)
	case string:
		v = starlark.String(x)
	case json.Number:
		// a JSON number stays a number: Int for an integer literal (exact,
		// arbitrary precision) and Float otherwise — matching json.decode and
		// serial. A caller that decodes raw JSON into map[string]interface{}
		// with dec.UseNumber() reaches this; it is the int-vs-float-preserving
		// path that a plain json.Unmarshal (which collapses every number to
		// float64, losing the int/float distinction) cannot offer.
		v, err = marshalJSONNumber(x)
	case *big.Int:
		// the inverse of Unmarshal, which returns *big.Int for integers
		// beyond uint64, so a marshal/unmarshal round-trip stays exact.
		v = starlark.MakeBigInt(x)
	case big.Int:
		v = starlark.MakeBigInt(&x)
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
		v = starlark.Float(x)
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
	case []string:
		var elems = make([]starlark.Value, len(x))
		for i, val := range x {
			elems[i] = starlark.String(val)
		}
		v = starlark.NewList(elems)
	case []byte:
		v = starlark.Bytes(x)
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
	case map[string]string:
		dict := &starlark.Dict{}
		for key, val := range x {
			if err = dict.SetKey(starlark.String(key), starlark.String(val)); err != nil {
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

// marshalJSONNumber maps a json.Number to a Starlark Int (an integer literal,
// at arbitrary precision) or Float (a literal with a decimal point or
// exponent). It is the same int-vs-float rule json.decode and serial use, so a
// number written without a fractional part round-trips as an int and large
// integers keep their exact value instead of degrading through float64.
func marshalJSONNumber(n json.Number) (starlark.Value, error) {
	s := n.String()
	if !strings.ContainsAny(s, ".eE") {
		if i, err := n.Int64(); err == nil {
			return starlark.MakeInt64(i), nil
		}
		if bi, ok := new(big.Int).SetString(s, 10); ok {
			return starlark.MakeBigInt(bi), nil
		}
	}
	f, err := n.Float64()
	if err != nil {
		return nil, fmt.Errorf("invalid number %q: %w", s, err)
	}
	return starlark.Float(f), nil
}

// Unmarshal converts a starlark.Value into its Golang counterpart, like FromValue() of package starlight does.
//
// The contract:
//   - Int values within the platform int range come back as int; larger
//     values degrade losslessly to uint64 and then *big.Int (they used to
//     be an error).
//   - Dicts always come back as map[string]interface{}: string keys keep
//     their value, any other hashable key is stringified via its Starlark
//     representation (bool keys use json-style "true"/"false"), and a
//     post-stringification collision is an error. This deliberately
//     prefers JSON-friendly shapes over key fidelity; use starlight's
//     FromValue for typed dict keys.
//   - Cyclic values, direct or indirect, are reported as errors.
//
// It's the opposite of Marshal().
func Unmarshal(x starlark.Value) (val interface{}, err error) {
	return unmarshalValue(x, nil)
}

// unmarshalValue is Unmarshal's engine; visited tracks the mutable
// containers (dict/list/set) on the current descent path so that cyclic
// values - direct or indirect - are reported instead of overflowing the
// stack. The map is allocated lazily on the first container.
func unmarshalValue(x starlark.Value, visited map[starlark.Value]bool) (val interface{}, err error) {
	enter := func(c starlark.Value) (map[starlark.Value]bool, error) {
		if visited[c] {
			return nil, fmt.Errorf("cyclic reference found")
		}
		if visited == nil {
			visited = make(map[starlark.Value]bool)
		}
		visited[c] = true
		return visited, nil
	}
	iterAttrs := func(v starlark.HasAttrs) (map[string]interface{}, error) {
		jo := make(map[string]interface{})
		for _, name := range v.AttrNames() {
			sv, err := v.Attr(name)
			if err != nil {
				return nil, err
			}
			jo[name], err = unmarshalValue(sv, visited)
			if err != nil {
				return nil, err
			}
		}
		return jo, nil
	}

	// for typed nil or nil
	if IsInterfaceNil(x) {
		if x == nil {
			return nil, errors.New("nil value")
		}
		return nil, fmt.Errorf("typed nil value: %T", x)
	}

	// switch on the type of the value (common types)
	switch v := x.(type) {
	case starlark.NoneType:
		val = nil
	case starlark.Bool:
		val = v.Truth() == starlark.True
	case starlark.Int:
		// in-range values keep returning the platform int (the historical
		// dynamic type, which downstream type assertions depend on); values
		// beyond it used to be an error and now degrade losslessly through
		// uint64 to *big.Int instead
		var tmp int
		if errConv := starlark.AsInt(x, &tmp); errConv == nil {
			val = tmp
		} else if u, ok := v.Uint64(); ok {
			val = u
		} else {
			val = new(big.Int).Set(v.BigInt())
		}
	case starlark.Float:
		if f, ok := starlark.AsFloat(x); !ok {
			err = fmt.Errorf("couldn't parse float")
		} else {
			val = f
		}
	case starlark.String:
		val = v.GoString()
	case starlark.Bytes:
		val = string(v)
	case startime.Time:
		val = time.Time(v)
	case *starlark.Dict:
		if visited, err = enter(v); err != nil {
			return nil, err
		}
		defer delete(visited, v)
		// the result is uniformly map[string]interface{} (JSON-marshalable):
		// string keys keep their value, every other hashable key uses its
		// Starlark representation. The old behavior flipped the whole map to
		// map[interface{}]interface{} on the first non-string key - which
		// json.Marshal rejects - and crashed the host outright on a tuple
		// key. A post-stringification collision is reported instead of
		// silently dropping an entry.
		rs := make(map[string]interface{}, v.Len())
		for _, k := range v.Keys() {
			dictVal, _, errGet := v.Get(k)
			if errGet != nil {
				return nil, errGet
			}
			pval, errVal := unmarshalValue(dictVal, visited)
			if errVal != nil {
				return nil, fmt.Errorf("unmarshaling starlark value: %w", errVal)
			}
			var ks string
			switch kk := k.(type) {
			case starlark.String:
				ks = kk.GoString()
			case starlark.Bool:
				// json-style lowercase, matching the historical dumps output
				if bool(kk) {
					ks = "true"
				} else {
					ks = "false"
				}
			default:
				ks = kk.String()
			}
			if _, dup := rs[ks]; dup {
				return nil, fmt.Errorf("dict key collision after stringification: %q", ks)
			}
			rs[ks] = pval
		}
		val = rs
	case *starlark.List:
		if visited, err = enter(v); err != nil {
			return nil, err
		}
		defer delete(visited, v)
		var (
			i       int
			listVal starlark.Value
			iter    = v.Iterate()
			value   = make([]interface{}, v.Len())
		)

		defer iter.Done()
		for iter.Next(&listVal) {
			value[i], err = unmarshalValue(listVal, visited)
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
			// tuples are immutable and cannot contain themselves, so they
			// are not tracked in visited - their elements still are
			value[i], err = unmarshalValue(tupleVal, visited)
			if err != nil {
				return
			}
			i++
		}
		val = value
	case *starlark.Set:
		// sets cannot participate in cycles: their elements must be
		// hashable, and every container that could lead back to the set
		// (dict/list/set, or a tuple holding one) is unhashable - so no
		// visited tracking is needed here
		var (
			i      int
			setVal starlark.Value
			iter   = v.Iterate()
			value  = make([]interface{}, v.Len())
		)

		defer iter.Done()
		for iter.Next(&setVal) {
			value[i], err = unmarshalValue(setVal, visited)
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
				err = fmt.Errorf("failed to marshal %T to Starlark object: %w", v.Constructor(), err)
				return
			}
			val = _var
		} else {
			am, err := iterAttrs(v)
			if err != nil {
				return nil, err
			}
			val = am
		}
	case *starlarkstruct.Module:
		am, err := iterAttrs(v)
		if err != nil {
			return nil, err
		}
		val = am
	case *convert.GoSlice:
		val = v.Value().Interface()
	case *convert.GoMap:
		val = v.Value().Interface()
	case *convert.GoStruct:
		val = v.Value().Interface()
	case *convert.GoInterface:
		val = v.Value().Interface()
	default:
		err = fmt.Errorf("unrecognized starlark type: %T", x)
	}
	return
}
