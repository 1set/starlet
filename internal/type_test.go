package internal

import (
	"testing"

	"go.starlark.net/starlark"
)

func TestFloatOrInt_Unpack(t *testing.T) {
	tests := []struct {
		name    string
		v       starlark.Value
		wantNum FloatOrInt
		wantErr bool
	}{
		{
			name:    "int",
			v:       starlark.MakeInt(1),
			wantNum: 1,
		},
		{
			name:    "float",
			v:       starlark.Float(1.2),
			wantNum: 1.2,
		},
		{
			name:    "string",
			v:       starlark.String("1"),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var p FloatOrInt
			if err := p.Unpack(tt.v); (err != nil) != tt.wantErr {
				t.Errorf("FloatOrInt.Unpack() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && p != tt.wantNum {
				t.Errorf("FloatOrInt.Unpack() got = %v, want %v", p, tt.wantNum)
			}
		})
	}
}

func TestStringOrBytes_Unpack(t *testing.T) {
	tests := []struct {
		name    string
		v       starlark.Value
		wantStr StringOrBytes
		wantErr bool
	}{
		{
			name:    "string",
			v:       starlark.String("foo"),
			wantStr: "foo",
		},
		{
			name:    "bytes",
			v:       starlark.Bytes("foo"),
			wantStr: "foo",
		},
		{
			name:    "int",
			v:       starlark.MakeInt(1),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var p StringOrBytes
			if err := p.Unpack(tt.v); (err != nil) != tt.wantErr {
				t.Errorf("StringOrBytes.Unpack() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && p != tt.wantStr {
				t.Errorf("StringOrBytes.Unpack() got = %v, want %v", p, tt.wantStr)
			}
		})
	}
}

func TestStringOrBytes_Stringer(t *testing.T) {
	tests := []struct {
		name     string
		v        StringOrBytes
		wantGo   string
		wantStar starlark.String
	}{
		{
			name:     "empty",
			v:        "",
			wantGo:   "",
			wantStar: starlark.String(""),
		},
		{
			name:     "string",
			v:        "foo",
			wantGo:   "foo",
			wantStar: starlark.String("foo"),
		},
		{
			name:     "bytes",
			v:        "bar",
			wantGo:   "bar",
			wantStar: starlark.String("bar"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.v.GoString(); got != tt.wantGo {
				t.Errorf("StringOrBytes.GoString() = %v, want %v", got, tt.wantGo)
			}
			if got := tt.v.GoBytes(); string(got) != tt.wantGo {
				t.Errorf("StringOrBytes.GoBytes() = %v, want %v", got, []byte(tt.wantGo))
			}
			if got := tt.v.StarlarkString(); got != tt.wantStar {
				t.Errorf("StringOrBytes.StarlarkString() = %v, want %v", got, tt.wantStar)
			}
		})
	}
}

func TestNumericValue(t *testing.T) {
	integer := func(n int) starlark.Value { return starlark.MakeInt(n) }
	double := func(n float64) starlark.Value { return starlark.Float(n) }
	tests := []struct {
		name    string
		values  []starlark.Value
		wantVal starlark.Value
		wantErr bool
	}{
		{
			name:    "empty",
			values:  []starlark.Value{},
			wantVal: integer(0),
		},
		{
			name:    "single int",
			values:  []starlark.Value{integer(100)},
			wantVal: integer(100),
		},
		{
			name:    "single float",
			values:  []starlark.Value{double(2)},
			wantVal: double(2),
		},
		{
			name:    "int and float",
			values:  []starlark.Value{integer(100), double(2)},
			wantVal: double(102),
		},
		{
			name:    "float and int",
			values:  []starlark.Value{double(4), integer(100)},
			wantVal: double(104),
		},
		{
			name:    "string",
			values:  []starlark.Value{starlark.String("100")},
			wantErr: true,
		},
		{
			name:    "int and string",
			values:  []starlark.Value{integer(100), starlark.String("2")},
			wantVal: integer(100),
			wantErr: true,
		},
		{
			name:    "string and int",
			values:  []starlark.Value{starlark.String("2"), integer(100)},
			wantVal: integer(0),
			wantErr: true,
		},
		{
			name:    "float and string",
			values:  []starlark.Value{double(4), starlark.String("2")},
			wantVal: double(4),
			wantErr: true,
		},
		{
			name:    "string and float",
			values:  []starlark.Value{starlark.String("2"), double(4)},
			wantVal: integer(0),
			wantErr: true,
		},
		{
			name:    "more int",
			values:  []starlark.Value{integer(100), integer(2), integer(3)},
			wantVal: integer(105),
		},
		{
			name:    "more float",
			values:  []starlark.Value{double(4), double(2), double(3)},
			wantVal: double(9),
		},
		{
			name:    "int and nil",
			values:  []starlark.Value{integer(100), nil},
			wantVal: integer(100),
		},
		{
			name:    "float and nil",
			values:  []starlark.Value{double(4), nil},
			wantVal: double(4),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := NewNumericValue()
			var err error
			for _, v := range tt.values {
				if err = n.Add(v); err != nil {
					break
				}
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("Add() error = %v, wantErr %v", err, tt.wantErr)
			}
			gotVal := n.Value()
			switch expVal := tt.wantVal.(type) {
			case starlark.Int:
				if actVal, ok := gotVal.(starlark.Int); !ok || actVal != expVal {
					t.Errorf("Add() gotVal = %v, want int %v", gotVal, tt.wantVal)
				}
			case starlark.Float:
				if actVal, ok := gotVal.(starlark.Float); !ok || actVal != expVal {
					t.Errorf("Add() gotVal = %v, want float %v", gotVal, tt.wantVal)
				}
			}
		})
	}
}
