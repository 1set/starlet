package dataconv

// Tests for DecodeStarlark/DecodeJSONStarlark: typed decoding from
// Starlark values into Go destinations with path-carrying errors.
//
// Sections:
//   1. happy path: full struct with tags, nesting, slices, maps, pointers
//   2. type and overflow errors carry the failing path
//   3. destination edge cases (bad out, interface, starlark.Value, bytes)

import (
	"math/big"
	"strings"
	"testing"
	"time"

	startime "go.starlark.net/lib/time"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

func mkDict(t *testing.T, pairs map[string]starlark.Value) *starlark.Dict {
	t.Helper()
	d := starlark.NewDict(len(pairs))
	for k, v := range pairs {
		if err := d.SetKey(starlark.String(k), v); err != nil {
			t.Fatal(err)
		}
	}
	return d
}

func TestDecodeStarlarkHappyPath(t *testing.T) {
	type message struct {
		Role    string  `json:"role"`
		Tokens  int     `json:"tokens"`
		Score   float64 `json:"score"`
		Hidden  string  `json:"-"`
		NoTag   bool
		Comment *string `json:"comment"`
	}
	type payload struct {
		Name     string         `json:"name"`
		Messages []message      `json:"messages"`
		Extra    map[string]int `json:"extra"`
		Origin   time.Time      `json:"origin"`
		Big      big.Int        `json:"big"`
		Raw      starlark.Value `json:"raw"`
		Any      interface{}    `json:"any"`
		Blob     []byte         `json:"blob"`
		Pos      [0]int         `json:"-"`
	}

	cmt := "first"
	origin := time.Date(2023, 1, 15, 12, 30, 45, 0, time.UTC)
	msg1 := mkDict(t, map[string]starlark.Value{
		"role":    starlark.String("user"),
		"tokens":  starlark.MakeInt(42),
		"score":   starlark.Float(0.5),
		"NoTag":   starlark.Bool(true),
		"comment": starlark.String(cmt),
		"ignored": starlark.MakeInt(1), // unknown source keys are skipped
	})
	msg2 := mkDict(t, map[string]starlark.Value{
		"role":    starlark.String("assistant"),
		"tokens":  starlark.MakeInt(7),
		"comment": starlark.None, // explicit None -> nil pointer
	})
	src := mkDict(t, map[string]starlark.Value{
		"name":     starlark.String("conv"),
		"messages": starlark.NewList([]starlark.Value{msg1, msg2}),
		"extra":    mkDict(t, map[string]starlark.Value{"a": starlark.MakeInt(1)}),
		"origin":   startime.Time(origin),
		"big":      starlark.MakeInt64(1).Lsh(80),
		"raw":      starlark.Tuple{starlark.MakeInt(1)},
		"any":      starlark.MakeInt(5),
		"blob":     starlark.Bytes("\x01\x02"),
	})

	var out payload
	if err := DecodeJSONStarlark(src, &out); err != nil {
		t.Fatalf("decode expects no error, got: %v", err)
	}
	if out.Name != "conv" || len(out.Messages) != 2 {
		t.Fatalf("unexpected top level: %+v", out)
	}
	m1, m2 := out.Messages[0], out.Messages[1]
	if m1.Role != "user" || m1.Tokens != 42 || m1.Score != 0.5 || !m1.NoTag || m1.Comment == nil || *m1.Comment != "first" {
		t.Errorf("unexpected message 1: %+v", m1)
	}
	if m1.Hidden != "" {
		t.Errorf("the '-' tagged field must not be filled, got %q", m1.Hidden)
	}
	if m2.Comment != nil {
		t.Errorf("None must decode to a nil pointer, got %v", *m2.Comment)
	}
	if out.Extra["a"] != 1 {
		t.Errorf("unexpected map: %v", out.Extra)
	}
	if !out.Origin.Equal(origin) {
		t.Errorf("unexpected time: %v", out.Origin)
	}
	if out.Big.BitLen() != 81 {
		t.Errorf("unexpected big int: %v", &out.Big)
	}
	if _, ok := out.Raw.(starlark.Tuple); !ok {
		t.Errorf("starlark.Value field must take the value as-is, got %T", out.Raw)
	}
	if out.Any != 5 {
		t.Errorf("interface{} field must reuse Unmarshal shapes, got %T %v", out.Any, out.Any)
	}
	if string(out.Blob) != "\x01\x02" {
		t.Errorf("unexpected bytes: %v", out.Blob)
	}
}

func TestDecodeStarlarkSources(t *testing.T) {
	type opts struct {
		Level int `json:"level"`
	}
	// starlarkstruct.Struct and Module sources work like dicts
	st := starlarkstruct.FromStringDict(starlarkstruct.Default, starlark.StringDict{"level": starlark.MakeInt(3)})
	var o1 opts
	if err := DecodeJSONStarlark(st, &o1); err != nil || o1.Level != 3 {
		t.Errorf("struct source: got %+v, err %v", o1, err)
	}
	mod := &starlarkstruct.Module{Name: "m", Members: starlark.StringDict{"level": starlark.MakeInt(4)}}
	var o2 opts
	if err := DecodeJSONStarlark(mod, &o2); err != nil || o2.Level != 4 {
		t.Errorf("module source: got %+v, err %v", o2, err)
	}
	// a tuple decodes into a slice
	var ints []int
	if err := DecodeJSONStarlark(starlark.Tuple{starlark.MakeInt(1), starlark.MakeInt(2)}, &ints); err != nil || len(ints) != 2 || ints[1] != 2 {
		t.Errorf("tuple source: got %v, err %v", ints, err)
	}
	// bytes decode into a string destination as raw bytes
	var s string
	if err := DecodeJSONStarlark(starlark.Bytes("raw"), &s); err != nil || s != "raw" {
		t.Errorf("bytes->string: got %q, err %v", s, err)
	}
}

func TestDecodeStarlarkErrorsCarryPath(t *testing.T) {
	type message struct {
		Role string `json:"role"`
		N    int8   `json:"n"`
	}
	type payload struct {
		Messages []message      `json:"messages"`
		Extra    map[string]int `json:"extra"`
	}
	cases := []struct {
		name string
		src  *starlark.Dict
		want string
	}{
		{
			name: "wrong type deep in a list",
			src: mkDict(t, map[string]starlark.Value{
				"messages": starlark.NewList([]starlark.Value{
					mkDict(t, map[string]starlark.Value{"role": starlark.String("ok")}),
					mkDict(t, map[string]starlark.Value{"role": starlark.MakeInt(3)}),
				}),
			}),
			want: "messages[1].role: got int, want string",
		},
		{
			name: "overflow names the field",
			src: mkDict(t, map[string]starlark.Value{
				"messages": starlark.NewList([]starlark.Value{
					mkDict(t, map[string]starlark.Value{"n": starlark.MakeInt(1000)}),
				}),
			}),
			want: "messages[0].n: value 1000 overflows int8",
		},
		{
			name: "map value error names the key",
			src: mkDict(t, map[string]starlark.Value{
				"extra": mkDict(t, map[string]starlark.Value{"k": starlark.String("x")}),
			}),
			want: "extra.k: got string, want int",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var out payload
			err := DecodeJSONStarlark(c.src, &out)
			if err == nil || !strings.Contains(err.Error(), c.want) {
				t.Errorf("expected error containing %q, got: %v", c.want, err)
			}
		})
	}
}

func TestDecodeStarlarkDestinationEdges(t *testing.T) {
	// out must be a non-nil pointer
	var i int
	if err := DecodeJSONStarlark(starlark.MakeInt(1), i); err == nil {
		t.Errorf("expected an error for a non-pointer destination")
	}
	var pi *int
	if err := DecodeJSONStarlark(starlark.MakeInt(1), pi); err == nil {
		t.Errorf("expected an error for a nil pointer destination")
	}
	// nil source
	if err := DecodeJSONStarlark(nil, &i); err == nil {
		t.Errorf("expected an error for a nil source")
	}
	// top-level scalar works
	if err := DecodeJSONStarlark(starlark.MakeInt(9), &i); err != nil || i != 9 {
		t.Errorf("expected 9, got %d / %v", i, err)
	}
	// unsupported destination kind
	var ch chan int
	if err := DecodeJSONStarlark(starlark.MakeInt(1), &ch); err == nil || !strings.Contains(err.Error(), "unsupported destination type") {
		t.Errorf("expected an unsupported-type error, got: %v", err)
	}
	// wrong shapes for containers
	var sl []int
	if err := DecodeJSONStarlark(starlark.MakeInt(1), &sl); err == nil || !strings.Contains(err.Error(), "want list or tuple") {
		t.Errorf("expected a list-shape error, got: %v", err)
	}
	var mp map[string]int
	if err := DecodeJSONStarlark(starlark.MakeInt(1), &mp); err == nil || !strings.Contains(err.Error(), "want dict") {
		t.Errorf("expected a dict-shape error, got: %v", err)
	}
	var mk map[int]int
	if err := DecodeJSONStarlark(starlark.NewDict(0), &mk); err == nil || !strings.Contains(err.Error(), "unsupported map key type") {
		t.Errorf("expected a map-key error, got: %v", err)
	}
	// non-string dict key into a string-keyed map
	bad := starlark.NewDict(1)
	_ = bad.SetKey(starlark.MakeInt(1), starlark.MakeInt(2))
	var mo map[string]int
	if err := DecodeJSONStarlark(bad, &mo); err == nil || !strings.Contains(err.Error(), "is not a string") {
		t.Errorf("expected a key-type error, got: %v", err)
	}
	// uint overflow and negative
	var u8 uint8
	if err := DecodeJSONStarlark(starlark.MakeInt(300), &u8); err == nil || !strings.Contains(err.Error(), "overflows uint8") {
		t.Errorf("expected a uint overflow error, got: %v", err)
	}
	if err := DecodeJSONStarlark(starlark.MakeInt(-1), &u8); err == nil {
		t.Errorf("expected an error for a negative value into uint8")
	}
	// float from int, and float32 overflow
	var f32 float32
	if err := DecodeJSONStarlark(starlark.MakeInt(2), &f32); err != nil || f32 != 2 {
		t.Errorf("expected 2.0, got %v / %v", f32, err)
	}
	if err := DecodeJSONStarlark(starlark.Float(1e300), &f32); err == nil || !strings.Contains(err.Error(), "overflows float32") {
		t.Errorf("expected a float overflow error, got: %v", err)
	}
	// time and big from wrong types
	var tt time.Time
	if err := DecodeJSONStarlark(starlark.String("now"), &tt); err == nil || !strings.Contains(err.Error(), "want time") {
		t.Errorf("expected a time-type error, got: %v", err)
	}
	var bi big.Int
	if err := DecodeJSONStarlark(starlark.String("1"), &bi); err == nil || !strings.Contains(err.Error(), "want int") {
		t.Errorf("expected a big-int-type error, got: %v", err)
	}
	// bool from wrong type
	var b bool
	if err := DecodeJSONStarlark(starlark.MakeInt(1), &b); err == nil || !strings.Contains(err.Error(), "want bool") {
		t.Errorf("expected a bool-type error, got: %v", err)
	}
	// struct from a non-dict
	type s struct{}
	var so s
	if err := DecodeJSONStarlark(starlark.MakeInt(1), &so); err == nil || !strings.Contains(err.Error(), "want dict or struct") {
		t.Errorf("expected a struct-shape error, got: %v", err)
	}
}
