package dataconv

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	itn "github.com/1set/starlet/internal"
	"github.com/1set/starlight/convert"
	startime "go.starlark.net/lib/time"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

func TestIsEmptyString(t *testing.T) {
	if !IsEmptyString(starlark.String("")) {
		t.Error("empty string should equal true")
	}

	if IsEmptyString(".") {
		t.Error("non-empty string shouldn't be empty")
	}
}

func TestIsInterfaceNil(t *testing.T) {
	var (
		ei  interface{}
		em  map[string]struct{}
		es  fmt.Stringer
		esp *customType
		est customType
	)
	tests := []struct {
		name string
		i    interface{}
		want bool
	}{
		{"nil", nil, true},
		{"nil interface", ei, true},
		{"nil map", em, true},
		{"nil stringer", es, true},
		{"nil pointer", esp, true},
		{"string", "1", false},
		{"custom type", est, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsInterfaceNil(tt.i); got != tt.want {
				t.Errorf("IsInterfaceNil(%v) = %v, want %v", tt.i, got, tt.want)
			}
		})
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

func TestMarshalStarlarkJSON(t *testing.T) {
	now := time.Now()
	sd := starlark.NewDict(1)
	sd.SetKey(starlark.String("foo"), starlark.MakeInt(42))
	sd2 := starlark.NewDict(1)
	sd2.SetKey(starlark.MakeUint(42), starlark.String("foo"))
	sd3 := starlark.NewDict(1)
	sd3.SetKey(starlark.Bool(true), starlark.MakeInt(42))

	ss := starlark.NewSet(1)
	ss.Insert(starlark.String("foo"))
	ss.Insert(starlark.String("bar"))

	gf := convert.MakeStarFn("sqr", func(x int) int { return x * 2 })
	sf := asStarlarkFunc("foobar", itn.HereDoc(`
		def foobar(x):
			return x*2
	`))
	sb := mockStarlarkBuiltin("foobar")
	sd4 := starlark.NewDict(1)
	sd4.SetKey(starlark.String("foo"), gf)
	sd5 := starlark.NewDict(1)
	sd5.SetKey(starlark.String("bar"), sf)

	stime := time.Unix(1689384600, 0)
	stime = stime.In(time.FixedZone("CST", 8*60*60))
	st := struct {
		Foo   string    `json:"foo"`
		Bar   int       `json:"bar"`
		Later time.Time `json:"later"`
	}{
		Foo:   "Hello, World!",
		Bar:   42,
		Later: stime,
	}
	ste := struct {
		Foo string   `json:"foo"`
		Car chan int `json:"car"`
	}{
		Foo: "Goodbye!",
		Car: make(chan int, 10),
	}

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
			name: "dict2",
			data: sd2,
			want: `{"42":"foo"}`,
		},
		{
			name: "dict3",
			data: sd3,
			want: `{"true":42}`,
		},
		{
			name:    "dict4",
			data:    sd4,
			wantErr: true,
		},
		{
			name:    "dict5",
			data:    sd5,
			wantErr: true,
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
			name: "starlark struct nil",
			data: &starlarkstruct.Struct{},
			want: `{}`,
		},
		{
			name: "starlark struct",
			data: starlarkstruct.FromStringDict(starlarkstruct.Default, starlark.StringDict{
				"foo": starlark.String("Hello, World!"),
				"bar": starlark.MakeInt(42),
			}),
			want: `{"bar":42,"foo":"Hello, World!"}`,
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
			name: "go struct more",
			data: convert.NewStruct(st),
			want: `{"foo":"Hello, World!","bar":42,"later":"2023-07-15T09:30:00+08:00"}`,
		},
		{
			name:    "go struct chan",
			data:    convert.NewStruct(ste),
			wantErr: true,
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
			want: itn.HereDoc(`
				{
				  "foo": 42
				}`),
		},
		{
			name:    "function",
			data:    sf,
			wantErr: true,
		},
		{
			name:    "builtin",
			data:    sb,
			wantErr: true,
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

func TestConvertStruct(t *testing.T) {
	type record1 struct {
		Name  string `sl:"name"`
		Index uint   `sl:"idx"`
	}
	v1 := ConvertStruct(&record1{Name: "foo", Index: 42}, "sl")
	j1, err := MarshalStarlarkJSON(v1, 0)
	if err != nil {
		t.Errorf("MarshalStarlarkJSON() error = %v", err)
		return
	}
	t.Logf(j1)
}

func TestConvertStructPanic(t *testing.T) {
	type record1 struct {
		Name  string `sl:"name"`
		Index uint   `sl:"idx"`
	}
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("ConvertStruct() should panic")
		}
	}()
	ConvertStruct(record1{Name: "foo", Index: 42}, "sl")
}

func TestConvertJSONStruct(t *testing.T) {
	type record2 struct {
		Name  string `json:"name"`
		Index uint   `json:"idx"`
	}
	v2 := ConvertJSONStruct(&record2{Name: "bar", Index: 64})
	j2, err := MarshalStarlarkJSON(v2, 0)
	if err != nil {
		t.Errorf("MarshalStarlarkJSON() error = %v", err)
		return
	}
	t.Logf(j2)
}

func TestConvertJSONStructPanic(t *testing.T) {
	type record2 struct {
		Name  string `json:"name"`
		Index uint   `json:"idx"`
	}
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("ConvertJSONStruct() should panic")
		}
	}()
	ConvertJSONStruct(record2{Name: "bar", Index: 64})
}

func TestWrapModuleData(t *testing.T) {
	name := "test_module"
	data := starlark.StringDict{
		"foo": starlark.String("bar"),
		"baz": starlark.MakeInt(42),
	}

	wrapFunc := WrapModuleData(name, data)
	result, err := wrapFunc()
	if err != nil {
		t.Errorf("WrapModuleData() returned an error: %v", err)
	}

	module, ok := result[name].(*starlarkstruct.Module)
	if !ok {
		t.Errorf("WrapModuleData() did not return a module")
	}
	if module.Name != name {
		t.Errorf("WrapModuleData() returned a module with incorrect name. Expected: %s, Got: %s", name, module.Name)
	}
	if len(module.Members) != len(data) {
		t.Errorf("WrapModuleData() returned a module with incorrect number of members. Expected: %d, Got: %d", len(data), len(module.Members))
	}
	for key, value := range data {
		member, found := module.Members[key]
		if !found {
			t.Errorf("WrapModuleData() returned a module without the expected member: %s", key)
		}
		if member != value {
			t.Errorf("WrapModuleData() returned a module with incorrect member value. Key: %s, Expected: %v, Got: %v", key, value, member)
		}
	}
}

func TestTypingConvert(t *testing.T) {
	timestr := "2021-09-07T21:30:43Z"
	timestamp, _ := time.Parse(time.RFC3339, timestr)
	tests := []struct {
		name  string
		input interface{}
		want  interface{}
	}{
		{
			name:  "test float to int",
			input: float64(10),
			want:  10,
		},
		{
			name:  "test float remains same",
			input: 10.5,
			want:  10.5,
		},
		{
			name:  "valid time string to time.Time",
			input: timestr,
			want:  timestamp,
		},
		{
			name:  "normal string",
			input: "test string",
			want:  "test string",
		},
		{
			name:  "array of different values",
			input: []interface{}{float64(20), timestr, "test string"},
			want:  []interface{}{20, timestamp, "test string"},
		},
		{
			name:  "map of different values",
			input: map[string]interface{}{"age": float64(30), "dob": timestr, "name": "John Doe"},
			want:  map[string]interface{}{"age": 30, "dob": timestamp, "name": "John Doe"},
		},
		{
			name:  "boolean value",
			input: true,
			want:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TypingConvert(tt.input)

			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("TypingConvert() = %v, want %v", got, tt.want)
			}
		})
	}
}
