package types

import (
	"reflect"
	"testing"

	"go.starlark.net/starlark"
)

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
