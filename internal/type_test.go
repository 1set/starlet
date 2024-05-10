package internal

import (
	"reflect"
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

func TestFloatOrInt_Value(t *testing.T) {
	tests := []struct {
		name    string
		v       FloatOrInt
		wantInt int
		wantFlt float64
	}{
		{
			name:    "zero",
			v:       0,
			wantInt: 0,
			wantFlt: 0,
		},
		{
			name:    "int",
			v:       1,
			wantInt: 1,
			wantFlt: 1,
		},
		{
			name:    "float",
			v:       1.2,
			wantInt: 1,
			wantFlt: 1.2,
		},
		{
			name:    "large",
			v:       1e12 + 1,
			wantInt: 1000000000001,
			wantFlt: 1e12 + 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.v.GoInt(); got != tt.wantInt {
				t.Errorf("FloatOrInt.GoInt() = %v, want %v", got, tt.wantInt)
			}
			if got := tt.v.GoFloat(); got != tt.wantFlt {
				t.Errorf("FloatOrInt.GoFloat() = %v, want %v", got, tt.wantFlt)
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
			name:    "float and None",
			values:  []starlark.Value{double(6), starlark.None},
			wantVal: double(6),
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

func TestNullableString_Unpack(t *testing.T) {
	tests := []struct {
		name    string
		v       starlark.Value
		wantStr string
		wantErr bool
	}{
		{
			name:    "string",
			v:       starlark.String("foo"),
			wantStr: "foo",
		},
		{
			name:    "bytes",
			v:       starlark.Bytes("bar"),
			wantStr: "bar",
		},
		{
			name:    "none",
			v:       starlark.None,
			wantStr: "",
		},
		{
			name:    "int",
			v:       starlark.MakeInt(1),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var p NullableString
			if err := p.Unpack(tt.v); (err != nil) != tt.wantErr {
				t.Errorf("NullableString.Unpack() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && p.GoString() != tt.wantStr {
				t.Errorf("NullableString.Unpack() got = %v, want %v", p.GoString(), tt.wantStr)
			}
		})
	}
}

func TestNullableString_Methods(t *testing.T) {
	tests := []struct {
		name        string
		str         *NullableString
		wantStr     string
		wantIsNull  bool
		wantIsEmpty bool
	}{
		{
			name:        "nil",
			str:         nil,
			wantStr:     "",
			wantIsNull:  true,
			wantIsEmpty: true,
		},
		{
			name:        "nil value",
			str:         &NullableString{},
			wantStr:     "",
			wantIsNull:  true,
			wantIsEmpty: true,
		},
		{
			name:        "empty string",
			str:         NewNullableString(""),
			wantStr:     "",
			wantIsNull:  false,
			wantIsEmpty: true,
		},
		{
			name:        "non-empty string",
			str:         NewNullableString("hello"),
			wantStr:     "hello",
			wantIsNull:  false,
			wantIsEmpty: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotStr := tt.str.GoString(); gotStr != tt.wantStr {
				t.Errorf("NullableString.GoString() = %v, want %v", gotStr, tt.wantStr)
			}
			if gotIsNull := tt.str.IsNull(); gotIsNull != tt.wantIsNull {
				t.Errorf("NullableString.IsNull() = %v, want %v", gotIsNull, tt.wantIsNull)
			}
			if gotIsEmpty := tt.str.IsNullOrEmpty(); gotIsEmpty != tt.wantIsEmpty {
				t.Errorf("NullableString.IsNullOrEmpty() = %v, want %v", gotIsEmpty, tt.wantIsEmpty)
			}
		})
	}
}

func TestNullableDict_Unpack(t *testing.T) {
	tests := []struct {
		name    string
		v       starlark.Value
		wantErr bool
	}{
		{
			name: "dict",
			v:    starlark.NewDict(0),
		},
		{
			name: "none",
			v:    starlark.None,
		},
		{
			name:    "string",
			v:       starlark.String("foo"),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var p NullableDict
			if err := p.Unpack(tt.v); (err != nil) != tt.wantErr {
				t.Errorf("NullableDict.Unpack() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNullableDict_AsDict(t *testing.T) {
	tests := []struct {
		name string
		dict *NullableDict
		want *starlark.Dict
	}{
		{
			name: "nil",
			dict: &NullableDict{},
			want: starlark.NewDict(0),
		},
		{
			name: "non-nil",
			dict: &NullableDict{dict: starlark.NewDict(1)},
			want: starlark.NewDict(1),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.dict.AsDict(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NullableDict.AsDict() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNumericValue_Unpack(t *testing.T) {
	tests := []struct {
		name    string
		v       starlark.Value
		wantInt starlark.Int
		wantFlt float64
		hasFlt  bool
		wantErr bool
	}{
		{
			name:    "int",
			v:       starlark.MakeInt(42),
			wantInt: starlark.MakeInt(42),
			hasFlt:  false,
		},
		{
			name:    "float",
			v:       starlark.Float(3.14),
			wantFlt: 3.14,
			hasFlt:  true,
		},
		{
			name:    "string error",
			v:       starlark.String("not a number"),
			wantErr: true,
		},
		{
			name:    "none error",
			v:       starlark.None,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := NewNumericValue()
			err := n.Unpack(tt.v)
			if (err != nil) != tt.wantErr {
				t.Errorf("NumericValue.Unpack() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				if tt.hasFlt {
					if float64(n.floatValue) != tt.wantFlt {
						t.Errorf("NumericValue.Unpack() got = %v, want %v", n.floatValue, tt.wantFlt)
					}
					if n.hasFloat != tt.hasFlt {
						t.Errorf("NumericValue.Unpack() got hasFloat = %v, want %v", n.hasFloat, tt.hasFlt)
					}
				} else {
					if n.intValue != tt.wantInt {
						t.Errorf("NumericValue.Unpack() got = %v, want %v", n.intValue, tt.wantInt)
					}
				}
			}
		})
	}
}

func TestNumericValue_Value(t *testing.T) {
	tests := []struct {
		name    string
		n       *NumericValue
		wantVal starlark.Value
	}{
		{
			name:    "int value",
			n:       &NumericValue{intValue: starlark.MakeInt(42)},
			wantVal: starlark.MakeInt(42),
		},
		{
			name:    "float value",
			n:       &NumericValue{floatValue: starlark.Float(3.14), hasFloat: true},
			wantVal: starlark.Float(3.14),
		},
		{
			name:    "int and float value",
			n:       &NumericValue{intValue: starlark.MakeInt(100), floatValue: starlark.Float(3.14), hasFloat: true},
			wantVal: starlark.Float(103.14),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotVal := tt.n.Value(); gotVal != tt.wantVal {
				t.Errorf("NumericValue.Value() = %v, want %v", gotVal, tt.wantVal)
			}
		})
	}
}
