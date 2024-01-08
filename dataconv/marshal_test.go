package dataconv

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/1set/starlight/convert"
	startime "go.starlark.net/lib/time"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkjson"
	"go.starlark.net/starlarkstruct"
	"go.starlark.net/syntax"
)

func TestMarshal(t *testing.T) {
	expectedStringDict := starlark.NewDict(1)
	if err := expectedStringDict.SetKey(starlark.String("foo"), starlark.MakeInt(42)); err != nil {
		t.Fatal(err)
	}

	expectedIntDict := starlark.NewDict(1)
	if err := expectedIntDict.SetKey(starlark.MakeInt(42*2), starlark.MakeInt(42)); err != nil {
		t.Fatal(err)
	}

	ct, _ := (&customType{42}).MarshalStarlark()
	expectedStrDictCustomType := starlark.NewDict(2)
	if err := expectedStrDictCustomType.SetKey(starlark.String("foo"), starlark.MakeInt(42)); err != nil {
		t.Fatal(err)
	}
	if err := expectedStrDictCustomType.SetKey(starlark.String("bar"), ct); err != nil {
		t.Fatal(err)
	}

	fnoop := func() {}
	fnow := time.Now

	now := time.Now()
	cases := []struct {
		in   interface{}
		want starlark.Value
		err  string
	}{
		{nil, starlark.None, ""},
		{true, starlark.True, ""},
		{"foo", starlark.String("foo"), ""},
		{42, starlark.MakeInt(42), ""},
		{int8(42), starlark.MakeInt(42), ""},
		{int16(42), starlark.MakeInt(42), ""},
		{int32(42), starlark.MakeInt(42), ""},
		{int64(42), starlark.MakeInt(42), ""},
		{int64(1 << 42), starlark.MakeInt(1 << 42), ""},
		{uint(42), starlark.MakeUint(42), ""},
		{uint8(42), starlark.MakeUint(42), ""},
		{uint16(42), starlark.MakeUint(42), ""},
		{uint32(42), starlark.MakeUint(42), ""},
		{uint64(42), starlark.MakeUint64(42), ""},
		{uint64(1 << 42), starlark.MakeUint64(1 << 42), ""},
		{float32(42), starlark.Float(42), ""},
		{42., starlark.Float(42), ""},
		{time.Unix(1588540633, 0), startime.Time(time.Unix(1588540633, 0)), ""},
		{now, startime.Time(now), ""},
		{[]interface{}{42}, starlark.NewList([]starlark.Value{starlark.MakeInt(42)}), ""},
		{map[string]interface{}{"foo": 42}, expectedStringDict, ""},
		{map[interface{}]interface{}{"foo": 42}, expectedStringDict, ""},
		{map[interface{}]interface{}{42 * 2: 42}, expectedIntDict, ""},
		{&customType{42}, ct, ""},
		{map[string]interface{}{"foo": 42, "bar": &customType{42}}, expectedStrDictCustomType, ""},
		{map[interface{}]interface{}{"foo": 42, "bar": &customType{42}}, expectedStrDictCustomType, ""},
		{[]interface{}{42, &customType{42}}, starlark.NewList([]starlark.Value{starlark.MakeInt(42), ct}), ""},
		{&invalidCustomType{42}, starlark.None, "unrecognized type: &dataconv.invalidCustomType{Foo:42}"},
		{complex(1, 2), starlark.None, "unrecognized type: (1+2i)"},
		{fnoop, starlark.None, "unrecognized type: (func())"},
		{fnow, starlark.None, "unrecognized type: (func() time.Time)"},
		{[]func(){fnoop}, starlark.None, "unrecognized type: []func(){(func())"},
		{[]interface{}{fnoop}, starlark.None, "unrecognized type: (func())"},
		{map[string]func(){"foo": fnoop}, starlark.None, "unrecognized type: map[string]func()"},
		{map[string]interface{}{"foo": fnow}, starlark.None, "unrecognized type: (func() time.Time)"},
		{map[string]complex64{"foo": 1 + 2i}, starlark.None, `unrecognized type: map[string]complex64{"foo":(1+2i)}`},
		{map[complex64]complex64{1 + 2i: 3 + 4i}, starlark.None, "unrecognized type: map[complex64]complex64{(1+2i):(3+4i)}"},
		{map[interface{}]interface{}{complex(1, 2): 34}, starlark.None, "unrecognized type: (1+2i)"},
		{map[interface{}]interface{}{12: complex(3, 4)}, starlark.None, "unrecognized type: (3+4i)"},
	}

	for i, c := range cases {
		got, err := Marshal(c.in)
		if !(err == nil && c.err == "" || err != nil && err.Error() == c.err || err != nil && strings.HasPrefix(err.Error(), c.err)) {
			t.Errorf("case %d. error mismatch. expected: %q, got: %q (%T -> %T)", i, c.err, err, c.in, c.want)
			continue
		}
		if err != nil {
			continue
		}

		compareResult, err := starlark.Equal(c.want, got)
		if err != nil {
			t.Errorf("case %d. error comparing results: %q", i, err)
			continue
		}
		if !compareResult {
			t.Errorf("case %d. expected: %#v, got: %#v (%T -> %T)", i, c.want, got, c.in, c.want)
		}
	}
}

func TestUnmarshal(t *testing.T) {
	now := time.Now()

	strDict := starlark.NewDict(1)
	if err := strDict.SetKey(starlark.String("foo"), starlark.MakeInt(42)); err != nil {
		t.Fatal(err)
	}
	intDict := starlark.NewDict(1)
	if err := intDict.SetKey(starlark.MakeInt(42*2), starlark.MakeInt(42)); err != nil {
		t.Fatal(err)
	}
	nilDict := starlark.NewDict(1)
	if err := nilDict.SetKey(starlark.String("foo"), nil); err != nil {
		t.Fatal(err)
	}

	ct, _ := (&customType{42}).MarshalStarlark()
	strDictCT := starlark.NewDict(2)
	if err := strDictCT.SetKey(starlark.String("foo"), starlark.MakeInt(42)); err != nil {
		t.Fatal(err)
	}
	if err := strDictCT.SetKey(starlark.String("bar"), ct); err != nil {
		t.Fatal(err)
	}

	ss := starlark.NewSet(10)
	ss.Insert(starlark.String("Hello"))
	ss.Insert(starlark.String("World"))

	msb := mockStarlarkBuiltin("foo")
	sf := asStarlarkFunc("foo", `def foo(): return "foo"`)
	sse := starlark.NewSet(10)
	sse.Insert(msb)
	sle := starlark.NewList([]starlark.Value{msb})
	ste := starlark.Tuple{msb}
	sdke := starlark.NewDict(10)
	sdke.SetKey(msb, starlark.MakeInt(42))
	sdve := starlark.NewDict(10)
	sdve.SetKey(starlark.String("foo"), msb)

	srt := starlarkstruct.FromStringDict(starlarkstruct.Default, map[string]starlark.Value{
		"Message": starlark.String("Aloha"),
		"Times":   starlark.MakeInt(100),
		"Later":   startime.Time(now),
	})
	srtNil := starlarkstruct.FromStringDict(starlarkstruct.Default, map[string]starlark.Value{
		"Null": nil,
	})

	gs := struct {
		Message string
		Times   int
	}{"Aloha", 100}

	var (
		nilGs  *convert.GoSlice
		nilGm  *convert.GoMap
		nilGst *convert.GoStruct
		nilGif *convert.GoInterface
	)
	cases := []struct {
		in   starlark.Value
		want interface{}
		err  string
	}{
		{nil, nil, "unrecognized starlark type: <nil>"},
		{nilDict, nil, "unmarshaling starlark value: unrecognized starlark type: <nil>"},
		{srtNil, nil, "unrecognized starlark type: <nil>"},
		{starlark.NewList([]starlark.Value{starlark.MakeInt(42), nil}), nil, "unrecognized starlark type: <nil>"},
		{starlark.Tuple([]starlark.Value{starlark.MakeInt(42), nil}), nil, "unrecognized starlark type: <nil>"},
		{starlark.None, nil, ""},
		{starlark.True, true, ""},
		{starlark.String("foo"), "foo", ""},
		{starlark.MakeInt(0), 0, ""},
		{starlark.MakeInt(42), 42, ""},
		{starlark.MakeInt(42), int8(42), ""},
		{starlark.MakeInt(42), int16(42), ""},
		{starlark.MakeInt(42), int32(42), ""},
		{starlark.MakeInt(42), int64(42), ""},
		{starlark.MakeInt(1 << 42), int64(1 << 42), ""},
		{starlark.MakeUint(42), uint(42), ""},
		{starlark.MakeUint(42), uint8(42), ""},
		{starlark.MakeUint(42), uint16(42), ""},
		{starlark.MakeUint(42), uint32(42), ""},
		{starlark.MakeUint64(42), uint64(42), ""},
		{starlark.MakeUint64(1 << 42), uint64(1 << 42), ""},
		{starlark.Float(42), float32(42), ""},
		{starlark.Float(42), 42., ""},
		{starlark.Float(0), 0., ""},
		{startime.Time(time.Unix(1588540633, 0)), time.Unix(1588540633, 0), ""},
		{startime.Time(now), now, ""},
		{starlark.NewList([]starlark.Value{starlark.MakeInt(42)}), []interface{}{42}, ""},
		{strDict, map[string]interface{}{"foo": 42}, ""},
		{intDict, map[interface{}]interface{}{42 * 2: 42}, ""},
		{ct, &customType{42}, ""},
		{strDictCT, map[string]interface{}{"foo": 42, "bar": &customType{42}}, ""},
		{starlark.NewList([]starlark.Value{starlark.MakeInt(42), ct}), []interface{}{42, &customType{42}}, ""},
		{starlark.Tuple{starlark.String("foo"), starlark.MakeInt(42)}, []interface{}{"foo", 42}, ""},
		{ss, []interface{}{"Hello", "World"}, ""},
		{&starlarkstruct.Struct{}, map[string]interface{}{}, ""},
		{srt, map[string]interface{}{"Message": "Aloha", "Times": 100, "Later": now}, ""},
		{starlarkjson.Module, nil, "unrecognized starlark type: *starlarkstruct.Module"},
		{convert.NewGoSlice([]int{1, 2, 3}), []int{1, 2, 3}, ""},
		{convert.NewGoSlice([]string{"Hello", "World"}), []string{"Hello", "World"}, ""},
		{convert.NewGoMap(map[string]int{"foo": 42}), map[string]int{"foo": 42}, ""},
		{convert.NewStruct(gs), gs, ""},
		{convert.MakeGoInterface("Hello, World!"), "Hello, World!", ""},
		{(*convert.GoSlice)(nil), nil, "nil GoSlice"},
		{nilGs, nil, "nil GoSlice"},
		{(*convert.GoMap)(nil), nil, "nil GoMap"},
		{nilGm, nil, "nil GoMap"},
		{(*convert.GoStruct)(nil), nil, "nil GoStruct"},
		{nilGst, nil, "nil GoStruct"},
		{(*convert.GoInterface)(nil), nil, "nil GoInterface"},
		{nilGif, nil, "nil GoInterface"},
		{msb, nil, "unrecognized starlark type: *starlark.Builtin"},
		{sf, nil, "unrecognized starlark type: *starlark.Function"},
		{sse, nil, "unrecognized starlark type: *starlark.Builtin"},
		{sle, nil, "unrecognized starlark type: *starlark.Builtin"},
		{ste, nil, "unrecognized starlark type: *starlark.Builtin"},
		{sdke, nil, "unmarshaling starlark key: unrecognized starlark type: *starlark.Builtin"},
		{sdve, nil, "unmarshaling starlark value: unrecognized starlark type: *starlark.Builtin"},
	}
	for i, c := range cases {
		got, err := Unmarshal(c.in)
		if !(err == nil && c.err == "" || err != nil && err.Error() == c.err) {
			t.Errorf("case %d. error mismatch. expected: %q, got: %v, %T -> %T", i, c.err, err, c.in, c.want)
			continue
		}

		// convert to the same type as expected
		var act interface{}
		act = got
		switch c.want.(type) {
		case int8:
			act = int8(got.(int))
		case int16:
			act = int16(got.(int))
		case int32:
			act = int32(got.(int))
		case int64:
			act = int64(got.(int))
		case uint:
			act = uint(got.(int))
		case uint8:
			act = uint8(got.(int))
		case uint16:
			act = uint16(got.(int))
		case uint32:
			act = uint32(got.(int))
		case uint64:
			act = uint64(got.(int))
		case float32:
			act = float32(got.(float64))
		}

		// compare
		if !reflect.DeepEqual(c.want, act) {
			t.Errorf("case %d. expected: %#v (%T), got: %#v (%T), %T -> %T", i, c.want, c.want, got, got, c.in, c.want)
		}
	}
}

// asString unquotes a starlark string value
func asString(x starlark.Value) (string, error) {
	return strconv.Unquote(x.String())
}

// asStarlarkFunc returns a starlark function from a string for testing.
func asStarlarkFunc(fname, code string) *starlark.Function {
	thread := &starlark.Thread{Name: "test"}
	globals, err := starlark.ExecFile(thread, fname+".star", code, nil)
	if err != nil {
		panic(err)
	}
	return globals[fname].(*starlark.Function)
}

// mockStarlarkBuiltin returns a starlark builtin function for testing.
func mockStarlarkBuiltin(fname string) *starlark.Builtin {
	return starlark.NewBuiltin(fname, func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		return starlark.String("aloha " + fname), nil
	})
}

type invalidCustomType struct {
	Foo int64
}

type customType invalidCustomType

var (
	_ Unmarshaler    = (*customType)(nil)
	_ Marshaler      = (*customType)(nil)
	_ starlark.Value = (*customType)(nil)
)

func (c *customType) CompareSameType(op syntax.Token, v starlark.Value, depth int) (bool, error) {
	if op != syntax.EQL {
		return false, fmt.Errorf("not expected operator %q", op)
	}
	other := v.(*customType)
	return c.Foo == other.Foo, nil
}

func (c *customType) UnmarshalStarlark(v starlark.Value) error {
	// asserts
	if v.Type() != "struct" {
		return fmt.Errorf("not expected top level type, want struct, got %q", v.Type())
	}
	if _, ok := v.(*starlarkstruct.Struct).Constructor().(*customType); !ok {
		return fmt.Errorf("not expected construct type got %T, want %T", v.(*starlarkstruct.Struct).Constructor(), c)
	}

	// TODO: refactoring transform data
	mustInt64 := func(sv starlark.Value) int64 {
		i, _ := sv.(starlark.Int).Int64()
		return i
	}

	data := starlark.StringDict{}
	v.(*starlarkstruct.Struct).ToStringDict(data)

	*c = customType{
		Foo: mustInt64(data["foo"]),
	}
	return nil
}

func (c *customType) MarshalStarlark() (starlark.Value, error) {
	v := starlarkstruct.FromStringDict(&customType{}, starlark.StringDict{
		"foo": starlark.MakeInt64(c.Foo),
	})
	return v, nil
}

func (c customType) String() string {
	return "customType"
}

func (c customType) Type() string { return "test.customType" }

func (customType) Freeze() {}

func (c customType) Truth() starlark.Bool {
	return starlark.True
}

func (c customType) Hash() (uint32, error) {
	return 0, fmt.Errorf("unhashable: %s", c.Type())
}
