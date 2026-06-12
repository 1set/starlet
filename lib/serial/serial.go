// Package serial defines a Starlark module that serializes data values to and
// from a JSON envelope, round-tripping the types plain JSON cannot: bytes,
// set, tuple, big integers, time, and dicts with non-string keys.
//
// The contract is deliberately strict: a value either round-trips losslessly
// or dumps fails with an actionable error — there is no silently-lossy middle
// ground. Code (functions/builtins), host objects (structs, Go-backed
// wrappers), non-finite floats, and reference cycles are rejected rather than
// flattened, because flattening them would quietly drop type identity,
// methods, or live bindings. To persist behavior, store the .star script and
// load() it; to persist a struct/host object, convert it to a dict first.
package serial

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"reflect"
	"sort"
	"strings"
	"sync"
	"time"

	startime "go.starlark.net/lib/time"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

// ModuleName defines the expected name for this Module when used in starlark's
// load() function, eg: load('serial', 'dumps')
const ModuleName = "serial"

// Envelope keys and type tags.
const (
	tagKey = "$t"
	valKey = "v"

	tagBytes  = "bytes"
	tagSet    = "set"
	tagTuple  = "tuple"
	tagBigint = "bigint"
	tagTime   = "time"
	tagMapKV  = "mapkv"  // a dict with non-string keys, as [[k, v], ...]
	tagObject = "object" // escape for a real dict that itself contains a "$t" key
)

var (
	none   = starlark.None
	once   sync.Once
	module starlark.StringDict
)

// LoadModule loads the serial module. It is concurrency-safe and idempotent.
func LoadModule() (starlark.StringDict, error) {
	once.Do(func() {
		module = starlark.StringDict{
			ModuleName: &starlarkstruct.Module{
				Name: ModuleName,
				Members: starlark.StringDict{
					"dumps":     starlark.NewBuiltin(ModuleName+".dumps", generateDumps(false)),
					"try_dumps": starlark.NewBuiltin(ModuleName+".try_dumps", generateDumps(true)),
					"loads":     starlark.NewBuiltin(ModuleName+".loads", generateLoads(false)),
					"try_loads": starlark.NewBuiltin(ModuleName+".try_loads", generateLoads(true)),
				},
			},
		}
	})
	return module, nil
}

func env(tag string, v interface{}) map[string]interface{} {
	return map[string]interface{}{tagKey: tag, valKey: v}
}

// generateDumps builds serial.dumps / serial.try_dumps. dumps walks the value
// directly (never via Unmarshal, which would collapse type information) and
// returns deterministic JSON text suitable for use as a cache key.
func generateDumps(try bool) func(*starlark.Thread, *starlark.Builtin, starlark.Tuple, []starlark.Tuple) (starlark.Value, error) {
	return func(_ *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		var v starlark.Value
		if err := starlark.UnpackArgs(fn.Name(), args, kwargs, "value", &v); err != nil {
			return failDumps(try, err, fn, false)
		}
		enc, err := encode(v, map[uintptr]bool{})
		if err != nil {
			return failDumps(try, err, fn, true)
		}
		b, err := json.Marshal(enc)
		if err != nil {
			return failDumps(try, err, fn, true)
		}
		if try {
			return starlark.Tuple{starlark.String(b), none}, nil
		}
		return starlark.String(b), nil
	}
}

func failDumps(try bool, err error, fn *starlark.Builtin, wrap bool) (starlark.Value, error) {
	if try {
		return starlark.Tuple{none, starlark.String(err.Error())}, nil
	}
	if wrap {
		return none, fmt.Errorf("%s: %w", fn.Name(), err)
	}
	return none, err
}

// generateLoads builds serial.loads / serial.try_loads. The result is a fresh,
// unfrozen value (mirroring json.decode), so scripts can read or mutate it.
func generateLoads(try bool) func(*starlark.Thread, *starlark.Builtin, starlark.Tuple, []starlark.Tuple) (starlark.Value, error) {
	return func(_ *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		var s string
		if err := starlark.UnpackArgs(fn.Name(), args, kwargs, "s", &s); err != nil {
			if try {
				return starlark.Tuple{none, starlark.String(err.Error())}, nil
			}
			return none, err
		}
		dec := json.NewDecoder(strings.NewReader(s))
		dec.UseNumber()
		var raw interface{}
		if err := dec.Decode(&raw); err != nil {
			if try {
				return starlark.Tuple{none, starlark.String(err.Error())}, nil
			}
			return none, fmt.Errorf("%s: %w", fn.Name(), err)
		}
		val, err := decode(raw)
		if err != nil {
			if try {
				return starlark.Tuple{none, starlark.String(err.Error())}, nil
			}
			return none, fmt.Errorf("%s: %w", fn.Name(), err)
		}
		if try {
			return starlark.Tuple{val, none}, nil
		}
		return val, nil
	}
}

// encode converts a Starlark value to a JSON-marshalable Go value. seen tracks
// list/dict pointers so a reference cycle errors instead of recursing forever.
func encode(v starlark.Value, seen map[uintptr]bool) (interface{}, error) {
	switch t := v.(type) {
	case starlark.NoneType:
		return nil, nil
	case starlark.Bool:
		return bool(t), nil
	case starlark.Int:
		if i, ok := t.Int64(); ok {
			return i, nil
		}
		return env(tagBigint, t.String()), nil
	case starlark.Float:
		f := float64(t)
		if math.IsNaN(f) || math.IsInf(f, 0) {
			return nil, fmt.Errorf("cannot serialize non-finite float %v", f)
		}
		return f, nil
	case starlark.String:
		return string(t), nil
	case starlark.Bytes:
		return env(tagBytes, base64.StdEncoding.EncodeToString([]byte(t))), nil
	case startime.Time:
		return env(tagTime, time.Time(t).Format(time.RFC3339Nano)), nil
	case *starlark.List:
		p := reflect.ValueOf(t).Pointer()
		if seen[p] {
			return nil, fmt.Errorf("cannot serialize a value that refers to itself (cycle in list)")
		}
		seen[p] = true
		defer delete(seen, p)
		arr := make([]interface{}, 0, t.Len())
		for i := 0; i < t.Len(); i++ {
			e, err := encode(t.Index(i), seen)
			if err != nil {
				return nil, err
			}
			arr = append(arr, e)
		}
		return arr, nil
	case starlark.Tuple:
		arr, err := encodeElems(t, seen)
		if err != nil {
			return nil, err
		}
		return env(tagTuple, arr), nil
	case *starlark.Set:
		elems, err := encodeSet(t, seen)
		if err != nil {
			return nil, err
		}
		return env(tagSet, elems), nil
	case *starlark.Dict:
		return encodeDict(t, seen)
	case *starlark.Function, *starlark.Builtin:
		return nil, fmt.Errorf("cannot serialize %s: it is code — store the .star script and load() it instead", v.Type())
	case *starlarkstruct.Struct:
		return nil, fmt.Errorf("cannot serialize struct: convert it to a dict first (struct identity cannot be persisted)")
	default:
		return nil, fmt.Errorf("cannot serialize value of type %s: serial round-trips data, not host objects", v.Type())
	}
}

func encodeElems(elems []starlark.Value, seen map[uintptr]bool) ([]interface{}, error) {
	out := make([]interface{}, 0, len(elems))
	for _, e := range elems {
		ee, err := encode(e, seen)
		if err != nil {
			return nil, err
		}
		out = append(out, ee)
	}
	return out, nil
}

// encodeSet encodes a set and sorts the elements by their JSON form so the
// output is deterministic (a set has no inherent order).
func encodeSet(s *starlark.Set, seen map[uintptr]bool) ([]interface{}, error) {
	var elems []interface{}
	iter := s.Iterate()
	defer iter.Done()
	var x starlark.Value
	for iter.Next(&x) {
		e, err := encode(x, seen)
		if err != nil {
			return nil, err
		}
		elems = append(elems, e)
	}
	return sortByJSON(elems), nil
}

// encodeDict encodes a dict. With all-string keys it becomes a JSON object
// (json.Marshal sorts the keys); with any non-string key it becomes the mapkv
// tag. A real object carrying a "$t" key is wrapped in the object tag so it is
// never mistaken for an envelope on the way back.
func encodeDict(d *starlark.Dict, seen map[uintptr]bool) (interface{}, error) {
	p := reflect.ValueOf(d).Pointer()
	if seen[p] {
		return nil, fmt.Errorf("cannot serialize a value that refers to itself (cycle in dict)")
	}
	seen[p] = true
	defer delete(seen, p)

	keys := d.Keys()
	allString := true
	for _, k := range keys {
		if _, ok := k.(starlark.String); !ok {
			allString = false
			break
		}
	}

	if allString {
		m := make(map[string]interface{}, len(keys))
		hasTag := false
		for _, k := range keys {
			ks := string(k.(starlark.String))
			if ks == tagKey {
				hasTag = true
			}
			val, _, _ := d.Get(k)
			e, err := encode(val, seen)
			if err != nil {
				return nil, err
			}
			m[ks] = e
		}
		if hasTag {
			return env(tagObject, m), nil
		}
		return m, nil
	}

	type pair struct {
		entry []interface{}
		order string
	}
	pairs := make([]pair, 0, len(keys))
	for _, k := range keys {
		ke, err := encode(k, seen)
		if err != nil {
			return nil, err
		}
		val, _, _ := d.Get(k)
		ve, err := encode(val, seen)
		if err != nil {
			return nil, err
		}
		kb, _ := json.Marshal(ke)
		pairs = append(pairs, pair{[]interface{}{ke, ve}, string(kb)})
	}
	sort.SliceStable(pairs, func(i, j int) bool { return pairs[i].order < pairs[j].order })
	out := make([]interface{}, len(pairs))
	for i, pr := range pairs {
		out[i] = pr.entry
	}
	return env(tagMapKV, out), nil
}

// sortByJSON orders encoded elements by their marshaled bytes, for determinism.
func sortByJSON(elems []interface{}) []interface{} {
	type keyed struct {
		v     interface{}
		order string
	}
	ks := make([]keyed, len(elems))
	for i, e := range elems {
		b, _ := json.Marshal(e)
		ks[i] = keyed{e, string(b)}
	}
	sort.SliceStable(ks, func(i, j int) bool { return ks[i].order < ks[j].order })
	out := make([]interface{}, len(elems))
	for i, k := range ks {
		out[i] = k.v
	}
	return out
}

// decode converts the parsed JSON (json.Number-preserving) back to a Starlark
// value, interpreting type tags.
func decode(v interface{}) (starlark.Value, error) {
	switch t := v.(type) {
	case nil:
		return starlark.None, nil
	case bool:
		return starlark.Bool(t), nil
	case json.Number:
		return numberToValue(t)
	case string:
		return starlark.String(t), nil
	case []interface{}:
		elems := make([]starlark.Value, len(t))
		for i, e := range t {
			ev, err := decode(e)
			if err != nil {
				return nil, err
			}
			elems[i] = ev
		}
		return starlark.NewList(elems), nil
	case map[string]interface{}:
		return decodeObject(t)
	default:
		return nil, fmt.Errorf("invalid serialized value of type %T", v)
	}
}

func decodeObject(m map[string]interface{}) (starlark.Value, error) {
	tag, ok := m[tagKey].(string)
	if !ok {
		return decodeDict(m)
	}
	raw := m[valKey]
	switch tag {
	case tagBytes:
		s, _ := raw.(string)
		b, err := base64.StdEncoding.DecodeString(s)
		if err != nil {
			return nil, fmt.Errorf("invalid bytes payload: %w", err)
		}
		return starlark.Bytes(b), nil
	case tagBigint:
		s, _ := raw.(string)
		bi, ok := new(big.Int).SetString(s, 10)
		if !ok {
			return nil, fmt.Errorf("invalid bigint payload %q", s)
		}
		return starlark.MakeBigInt(bi), nil
	case tagTuple:
		arr, _ := raw.([]interface{})
		elems := make([]starlark.Value, len(arr))
		for i, e := range arr {
			ev, err := decode(e)
			if err != nil {
				return nil, err
			}
			elems[i] = ev
		}
		return starlark.Tuple(elems), nil
	case tagSet:
		arr, _ := raw.([]interface{})
		set := starlark.NewSet(len(arr))
		for _, e := range arr {
			ev, err := decode(e)
			if err != nil {
				return nil, err
			}
			if err := set.Insert(ev); err != nil {
				return nil, err
			}
		}
		return set, nil
	case tagTime:
		s, _ := raw.(string)
		tm, err := time.Parse(time.RFC3339Nano, s)
		if err != nil {
			return nil, fmt.Errorf("invalid time payload %q: %w", s, err)
		}
		return startime.Time(tm), nil
	case tagMapKV:
		arr, _ := raw.([]interface{})
		d := starlark.NewDict(len(arr))
		for _, pr := range arr {
			kvp, ok := pr.([]interface{})
			if !ok || len(kvp) != 2 {
				return nil, fmt.Errorf("invalid mapkv entry")
			}
			kv, err := decode(kvp[0])
			if err != nil {
				return nil, err
			}
			vv, err := decode(kvp[1])
			if err != nil {
				return nil, err
			}
			if err := d.SetKey(kv, vv); err != nil {
				return nil, err
			}
		}
		return d, nil
	case tagObject:
		mm, ok := raw.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid object payload")
		}
		return decodeDict(mm)
	default:
		return nil, fmt.Errorf("unknown type tag %q", tag)
	}
}

func decodeDict(m map[string]interface{}) (starlark.Value, error) {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	d := starlark.NewDict(len(m))
	for _, k := range keys {
		vv, err := decode(m[k])
		if err != nil {
			return nil, err
		}
		if err := d.SetKey(starlark.String(k), vv); err != nil {
			return nil, err
		}
	}
	return d, nil
}

func numberToValue(n json.Number) (starlark.Value, error) {
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
