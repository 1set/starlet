package dataconv

import (
	"fmt"
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

	ss := starlark.NewSet(1)
	ss.Insert(starlark.String("foo"))
	ss.Insert(starlark.String("bar"))

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
