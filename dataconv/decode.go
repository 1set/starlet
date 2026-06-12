package dataconv

import (
	"errors"
	"fmt"
	"math/big"
	"reflect"
	"strings"
	"time"

	startime "go.starlark.net/lib/time"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

var (
	timeType     = reflect.TypeOf(time.Time{})
	bigIntType   = reflect.TypeOf(big.Int{})
	starValueTyp = reflect.TypeOf((*starlark.Value)(nil)).Elem()
)

// DecodeStarlark decodes a Starlark value directly into a typed Go
// destination, the way encoding/json.Unmarshal fills a struct: out must be
// a non-nil pointer. It walks the starlark.Value itself rather than going
// through Unmarshal's interface{} shapes, so type information (int width,
// bytes vs string, tuples) is checked against the destination instead of
// being collapsed, and every error names the path that failed (for
// example: messages[2].role: got int, want string).
//
// Supported destinations: structs (fields matched by the tagName struct
// tag, falling back to the field name; unknown source keys are ignored),
// pointers (None decodes to nil, so *T distinguishes absent from zero),
// slices (from list or tuple; []byte also accepts bytes), maps with string
// keys, booleans, integers (with overflow checks), floats, strings,
// time.Time, big.Int, starlark.Value (taken as-is), and interface{}
// (filled with Unmarshal's shapes). Struct sources may be dicts,
// starlarkstruct structs, or modules.
func DecodeStarlark(v starlark.Value, out interface{}, tagName string) error {
	rv := reflect.ValueOf(out)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return errors.New("decode destination must be a non-nil pointer")
	}
	if v == nil {
		return errors.New("nil source value")
	}
	return decodeValue(v, rv.Elem(), tagName, "")
}

// DecodeJSONStarlark is DecodeStarlark with the "json" tag name, matching
// the struct tags most API payload types already carry.
func DecodeJSONStarlark(v starlark.Value, out interface{}) error {
	return DecodeStarlark(v, out, "json")
}

func decodeValue(v starlark.Value, dst reflect.Value, tagName, path string) error {
	at := func(format string, args ...interface{}) error {
		p := path
		if p == "" {
			p = "value"
		}
		return fmt.Errorf("%s: %s", p, fmt.Sprintf(format, args...))
	}

	// pointer destinations: None becomes nil, anything else allocates and
	// descends - this is how *T distinguishes absent from zero
	if dst.Kind() == reflect.Ptr {
		if v == starlark.None {
			dst.Set(reflect.Zero(dst.Type()))
			return nil
		}
		if dst.IsNil() {
			dst.Set(reflect.New(dst.Type().Elem()))
		}
		return decodeValue(v, dst.Elem(), tagName, path)
	}

	// interface destinations
	if dst.Kind() == reflect.Interface {
		if dst.Type() == starValueTyp {
			dst.Set(reflect.ValueOf(v))
			return nil
		}
		if dst.NumMethod() == 0 {
			gv, err := Unmarshal(v)
			if err != nil {
				return at("%v", err)
			}
			if gv == nil {
				dst.Set(reflect.Zero(dst.Type()))
			} else {
				dst.Set(reflect.ValueOf(gv))
			}
			return nil
		}
		return at("unsupported destination interface %s", dst.Type())
	}

	// concrete special types before the kind switch
	switch dst.Type() {
	case timeType:
		t, ok := v.(startime.Time)
		if !ok {
			return at("got %s, want time", v.Type())
		}
		dst.Set(reflect.ValueOf(time.Time(t)))
		return nil
	case bigIntType:
		i, ok := v.(starlark.Int)
		if !ok {
			return at("got %s, want int", v.Type())
		}
		dst.Set(reflect.ValueOf(*new(big.Int).Set(i.BigInt())))
		return nil
	}

	switch dst.Kind() {
	case reflect.Bool:
		b, ok := v.(starlark.Bool)
		if !ok {
			return at("got %s, want bool", v.Type())
		}
		dst.SetBool(bool(b))
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		i, ok := v.(starlark.Int)
		if !ok {
			return at("got %s, want int", v.Type())
		}
		i64, ok := i.Int64()
		if !ok || dst.OverflowInt(i64) {
			return at("value %s overflows %s", i.String(), dst.Type())
		}
		dst.SetInt(i64)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		i, ok := v.(starlark.Int)
		if !ok {
			return at("got %s, want int", v.Type())
		}
		u64, ok := i.Uint64()
		if !ok || dst.OverflowUint(u64) {
			return at("value %s overflows %s", i.String(), dst.Type())
		}
		dst.SetUint(u64)
	case reflect.Float32, reflect.Float64:
		f, ok := starlark.AsFloat(v)
		if !ok {
			return at("got %s, want float", v.Type())
		}
		if dst.OverflowFloat(f) {
			return at("value %v overflows %s", f, dst.Type())
		}
		dst.SetFloat(f)
	case reflect.String:
		switch sv := v.(type) {
		case starlark.String:
			dst.SetString(sv.GoString())
		case starlark.Bytes:
			dst.SetString(string(sv))
		default:
			return at("got %s, want string", v.Type())
		}
	case reflect.Slice:
		if dst.Type().Elem().Kind() == reflect.Uint8 {
			if bs, ok := v.(starlark.Bytes); ok {
				dst.SetBytes([]byte(bs))
				return nil
			}
		}
		seq, ok := v.(starlark.Sequence)
		if !ok {
			return at("got %s, want list or tuple", v.Type())
		}
		n := seq.Len()
		out := reflect.MakeSlice(dst.Type(), n, n)
		iter := seq.Iterate()
		defer iter.Done()
		var ev starlark.Value
		for i := 0; iter.Next(&ev); i++ {
			if err := decodeValue(ev, out.Index(i), tagName, fmt.Sprintf("%s[%d]", path, i)); err != nil {
				return err
			}
		}
		dst.Set(out)
	case reflect.Map:
		if dst.Type().Key().Kind() != reflect.String {
			return at("unsupported map key type %s", dst.Type().Key())
		}
		d, ok := v.(*starlark.Dict)
		if !ok {
			return at("got %s, want dict", v.Type())
		}
		out := reflect.MakeMapWithSize(dst.Type(), d.Len())
		for _, k := range d.Keys() {
			ks, ok := k.(starlark.String)
			if !ok {
				return at("dict key %s is not a string", k.String())
			}
			mv, _, err := d.Get(k)
			if err != nil {
				return at("%v", err)
			}
			elem := reflect.New(dst.Type().Elem()).Elem()
			if err := decodeValue(mv, elem, tagName, joinPath(path, ks.GoString())); err != nil {
				return err
			}
			out.SetMapIndex(reflect.ValueOf(ks.GoString()).Convert(dst.Type().Key()), elem)
		}
		dst.Set(out)
	case reflect.Struct:
		get, srcDesc := structGetter(v)
		if get == nil {
			return at("got %s, want dict or struct", v.Type())
		}
		t := dst.Type()
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			if f.PkgPath != "" {
				continue // unexported
			}
			name := f.Name
			if tagName != "" {
				if tag, ok := f.Tag.Lookup(tagName); ok {
					base := strings.Split(tag, ",")[0]
					if base == "-" {
						continue
					}
					if base != "" {
						name = base
					}
				}
			}
			fv, found, err := get(name)
			if err != nil {
				return at("reading %s of %s: %v", name, srcDesc, err)
			}
			if !found || fv == nil {
				continue // absent: leave the zero value (use *T to detect)
			}
			if err := decodeValue(fv, dst.Field(i), tagName, joinPath(path, name)); err != nil {
				return err
			}
		}
	default:
		return at("unsupported destination type %s", dst.Type())
	}
	return nil
}

// structGetter adapts the dict-like sources a struct can decode from.
func structGetter(v starlark.Value) (func(name string) (starlark.Value, bool, error), string) {
	switch src := v.(type) {
	case *starlark.Dict:
		return func(name string) (starlark.Value, bool, error) {
			val, found, err := src.Get(starlark.String(name))
			return val, found, err
		}, "dict"
	case *starlarkstruct.Struct:
		return func(name string) (starlark.Value, bool, error) {
			val, err := src.Attr(name)
			if err != nil || val == nil {
				// a missing attribute is "absent", not an error
				return nil, false, nil
			}
			return val, true, nil
		}, "struct"
	case *starlarkstruct.Module:
		return func(name string) (starlark.Value, bool, error) {
			val, err := src.Attr(name)
			if err != nil || val == nil {
				return nil, false, nil
			}
			return val, true, nil
		}, "module"
	}
	return nil, ""
}

// joinPath appends a field segment to an error path.
func joinPath(path, seg string) string {
	if path == "" {
		return seg
	}
	return path + "." + seg
}
