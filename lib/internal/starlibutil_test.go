package internal

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/1set/starlight/convert"
	startime "go.starlark.net/lib/time"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkjson"
	"go.starlark.net/starlarkstruct"
	"go.starlark.net/syntax"
)

func TestIsEmptyString(t *testing.T) {
	if !IsEmptyString(starlark.String("")) {
		t.Error("empty string should equal true")
	}

	if IsEmptyString(".") {
		t.Error("non-empty string shouldn't be empty")
	}
}

func TestAsString(t *testing.T) {
	cases := []struct {
		in       starlark.Value
		got, err string
	}{
		{starlark.String("foo"), "foo", ""},
		{starlark.String("\"foo'"), "\"foo'", ""},
		{starlark.Bool(true), "", "invalid syntax"},
	}

	for i, c := range cases {
		got, err := asString(c.in)
		if !(err == nil && c.err == "" || err != nil && err.Error() == c.err) {
			t.Errorf("case %d error mismatch. expected: '%s', got: '%s'", i, c.err, err)
			continue
		}

		if c.got != got {
			t.Errorf("case %d. expected: '%s', got: '%s'", i, c.got, got)
		}
	}
}

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
		{[]interface{}{42}, starlark.NewList([]starlark.Value{starlark.MakeInt(42)}), ""},
		{map[string]interface{}{"foo": 42}, expectedStringDict, ""},
		{map[interface{}]interface{}{"foo": 42}, expectedStringDict, ""},
		{map[interface{}]interface{}{42 * 2: 42}, expectedIntDict, ""},
		{&customType{42}, ct, ""},
		{map[string]interface{}{"foo": 42, "bar": &customType{42}}, expectedStrDictCustomType, ""},
		{map[interface{}]interface{}{"foo": 42, "bar": &customType{42}}, expectedStrDictCustomType, ""},
		{[]interface{}{42, &customType{42}}, starlark.NewList([]starlark.Value{starlark.MakeInt(42), ct}), ""},
		{&invalidCustomType{42}, starlark.None, "unrecognized type: &internal.invalidCustomType{Foo:42}"},
	}

	for i, c := range cases {
		got, err := Marshal(c.in)
		if !(err == nil && c.err == "" || err != nil && err.Error() == c.err) {
			t.Errorf("case %d. error mismatch. expected: %q, got: %q (%T -> %T)", i, c.err, err, c.in, c.want)
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
	strDict := starlark.NewDict(1)
	if err := strDict.SetKey(starlark.String("foo"), starlark.MakeInt(42)); err != nil {
		t.Fatal(err)
	}

	intDict := starlark.NewDict(1)
	if err := intDict.SetKey(starlark.MakeInt(42*2), starlark.MakeInt(42)); err != nil {
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
		{starlark.None, nil, ""},
		{starlark.True, true, ""},
		{starlark.String("foo"), "foo", ""},
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
		{startime.Time(time.Unix(1588540633, 0)), time.Unix(1588540633, 0), ""},
		{starlark.NewList([]starlark.Value{starlark.MakeInt(42)}), []interface{}{42}, ""},
		{strDict, map[string]interface{}{"foo": 42}, ""},
		{intDict, map[interface{}]interface{}{42 * 2: 42}, ""},
		{ct, &customType{42}, ""},
		{strDictCT, map[string]interface{}{"foo": 42, "bar": &customType{42}}, ""},
		{starlark.NewList([]starlark.Value{starlark.MakeInt(42), ct}), []interface{}{42, &customType{42}}, ""},
		{starlark.Tuple{starlark.String("foo"), starlark.MakeInt(42)}, []interface{}{"foo", 42}, ""},
		{ss, []interface{}{"Hello", "World"}, ""},
		{&starlarkstruct.Struct{}, nil, "constructor object from *starlarkstruct.Struct not supported Marshaler to starlark object: <nil>"},
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
	}
	for i, c := range cases {
		got, err := Unmarshal(c.in)
		if !(err == nil && c.err == "" || err != nil && err.Error() == c.err) {
			t.Errorf("case %d. error mismatch. expected: %q, got: %q, %T -> %T", i, c.err, err, c.in, c.want)
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

func TestMarshalStarlarkJSON(t *testing.T) {
	now := time.Now()
	sd := starlark.NewDict(1)
	sd.SetKey(starlark.String("foo"), starlark.MakeInt(42))
	ss := starlark.NewSet(1)
	ss.Insert(starlark.String("foo"))
	ss.Insert(starlark.String("bar"))

	tests := []struct {
		name    string
		data    starlark.Value
		indent  int
		want    string
		wantErr bool
	}{
		{
			name: "none",
			data: starlark.None,
			want: "null",
		},
		{
			name: "true",
			data: starlark.Bool(true),
			want: "true",
		},
		{
			name: "false",
			data: starlark.Bool(false),
			want: "false",
		},
		{
			name: "int",
			data: starlark.MakeInt(42),
			want: "42",
		},
		{
			name: "float",
			data: starlark.Float(1.23),
			want: "1.23",
		},
		{
			name: "string",
			data: starlark.String("Aloha!"),
			want: `"Aloha!"`,
		},
		{
			name: "time",
			data: startime.Time(now),
			want: fmt.Sprintf("%q", now.Format(time.RFC3339Nano)),
		},
		{
			name: "dict",
			data: sd,
			want: `{"foo":42}`,
		},
		{
			name: "list",
			data: starlark.NewList([]starlark.Value{starlark.MakeInt(43), starlark.String("foo")}),
			want: `[43,"foo"]`,
		},
		{
			name: "tuple",
			data: starlark.Tuple{starlark.MakeInt(60), starlark.String("bar")},
			want: `[60,"bar"]`,
		},
		{
			name: "set",
			data: ss,
			want: `["foo","bar"]`,
		},
		{
			name:    "struct",
			data:    &starlarkstruct.Struct{},
			wantErr: true,
		},
		{
			name: "go slice",
			data: convert.NewGoSlice([]int{1, 2, 3}),
			want: `[1,2,3]`,
		},
		{
			name: "go map",
			data: convert.NewGoMap(map[string]int{"foo": 42}),
			want: `{"foo":42}`,
		},
		{
			name: "go struct",
			data: convert.NewStruct(struct {
				Ace int `json:"a"`
			}{42}),
			want: `{"a":42}`,
		},
		{
			name: "go interface",
			data: convert.MakeGoInterface(42),
			want: `42`,
		},
		{
			name:   "plain indent",
			data:   starlark.String("Aloha!"),
			indent: 2,
			want:   `"Aloha!"`,
		},
		{
			name:   "dict indent",
			data:   sd,
			indent: 2,
			want: HereDoc(`
				{
				  "foo": 42
				}`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := MarshalStarlarkJSON(tt.data, tt.indent)
			if (err != nil) != tt.wantErr {
				t.Errorf("MarshalStarlarkJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("MarshalStarlarkJSON() got = %q, want %q", got, tt.want)
			}
		})
	}
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
