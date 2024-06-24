package types

import (
	"reflect"
	"testing"

	"go.starlark.net/starlark"
)

func TestEitherOrNone_Unpack(t *testing.T) {
	tests := []struct {
		name    string
		target  *EitherOrNone[starlark.String, starlark.Int]
		inV     starlark.Value
		want    starlark.Value
		wantErr bool
		isNone  bool
		isTypeA bool
		isTypeB bool
	}{
		{
			name:    "nil receiver",
			target:  nil,
			inV:     starlark.String("hello"),
			wantErr: true,
		},
		{
			name:    "empty struct",
			target:  &EitherOrNone[starlark.String, starlark.Int]{},
			inV:     nil,
			wantErr: true,
		},
		{
			name:    "string value",
			target:  NewEitherOrNone[starlark.String, starlark.Int](),
			inV:     starlark.String("hello"),
			want:    starlark.String("hello"),
			wantErr: false,
			isNone:  false,
			isTypeA: true,
			isTypeB: false,
		},
		{
			name:    "int value",
			target:  NewEitherOrNone[starlark.String, starlark.Int](),
			inV:     starlark.MakeInt(42),
			want:    starlark.MakeInt(42),
			wantErr: false,
			isNone:  false,
			isTypeA: false,
			isTypeB: true,
		},
		{
			name:    "none value",
			target:  NewEitherOrNone[starlark.String, starlark.Int](),
			inV:     starlark.None,
			want:    nil,
			wantErr: false,
			isNone:  true,
			isTypeA: false,
			isTypeB: false,
		},
		{
			name:    "wrong type value",
			target:  NewEitherOrNone[starlark.String, starlark.Int](),
			inV:     starlark.NewList(nil),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.target.Unpack(tt.inV)
			if (err != nil) != tt.wantErr {
				t.Errorf("EitherOrNone[%s].Unpack() error = %v, wantErr %v", tt.name, err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if got := tt.target.Value(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("EitherOrNone[%s].Value() = %v, want %v", tt.name, got, tt.want)
			}
			if got := tt.target.IsNone(); got != tt.isNone {
				t.Errorf("EitherOrNone[%s].IsNone() = %v, want %v", tt.name, got, tt.isNone)
			}
			if got := tt.target.IsTypeA(); got != tt.isTypeA {
				t.Errorf("EitherOrNone[%s].IsTypeA() = %v, want %v", tt.name, got, tt.isTypeA)
			}
			if got := tt.target.IsTypeB(); got != tt.isTypeB {
				t.Errorf("EitherOrNone[%s].IsTypeB() = %v, want %v", tt.name, got, tt.isTypeB)
			}
		})
	}
}

func TestEitherOrNone_Value(t *testing.T) {
	tests := []struct {
		name   string
		target *EitherOrNone[starlark.String, starlark.Int]
		inV    starlark.Value
		want   starlark.Value
	}{
		{
			name:   "string value",
			target: NewEitherOrNone[starlark.String, starlark.Int](),
			inV:    starlark.String("hello"),
			want:   starlark.String("hello"),
		},
		{
			name:   "int value",
			target: NewEitherOrNone[starlark.String, starlark.Int](),
			inV:    starlark.MakeInt(42),
			want:   starlark.MakeInt(42),
		},
		{
			name:   "none value",
			target: NewEitherOrNone[starlark.String, starlark.Int](),
			inV:    starlark.None,
			want:   nil,
		},
		{
			name:   "unknown type",
			target: new(EitherOrNone[starlark.String, starlark.Int]),
			want:   nil,
		},
		{
			name:   "nil receiver",
			target: nil,
			want:   nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.inV != nil && tt.target != nil {
				tt.target.Unpack(tt.inV)
			}
			got := tt.target.Value()
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("EitherOrNone[%s].Value() = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func TestEitherOrNone_ValueA(t *testing.T) {
	tests := []struct {
		name   string
		target *EitherOrNone[starlark.String, starlark.Int]
		inV    starlark.Value
		want   starlark.String
		wantOk bool
	}{
		{
			name:   "string value",
			target: NewEitherOrNone[starlark.String, starlark.Int](),
			inV:    starlark.String("hello"),
			want:   starlark.String("hello"),
			wantOk: true,
		},
		{
			name:   "int value",
			target: NewEitherOrNone[starlark.String, starlark.Int](),
			inV:    starlark.MakeInt(42),
			wantOk: false,
		},
		{
			name:   "none value",
			target: NewEitherOrNone[starlark.String, starlark.Int](),
			inV:    starlark.None,
			wantOk: false,
		},
		{
			name:   "nil receiver",
			target: nil,
			inV:    starlark.String("hello"),
			wantOk: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.target != nil {
				tt.target.Unpack(tt.inV)
			}
			got, ok := tt.target.ValueA()
			if ok != tt.wantOk {
				t.Errorf("EitherOrNone[%s].ValueA() ok = %v, want %v", tt.name, ok, tt.wantOk)
			}
			if ok && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("EitherOrNone[%s].ValueA() = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func TestEitherOrNone_ValueB(t *testing.T) {
	tests := []struct {
		name   string
		target *EitherOrNone[starlark.String, starlark.Int]
		inV    starlark.Value
		want   starlark.Int
		wantOk bool
	}{
		{
			name:   "string value",
			target: NewEitherOrNone[starlark.String, starlark.Int](),
			inV:    starlark.String("hello"),
			wantOk: false,
		},
		{
			name:   "int value",
			target: NewEitherOrNone[starlark.String, starlark.Int](),
			inV:    starlark.MakeInt(42),
			want:   starlark.MakeInt(42),
			wantOk: true,
		},
		{
			name:   "none value",
			target: NewEitherOrNone[starlark.String, starlark.Int](),
			inV:    starlark.None,
			wantOk: false,
		},
		{
			name:   "nil receiver",
			target: nil,
			inV:    starlark.String("hello"),
			wantOk: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.target != nil {
				tt.target.Unpack(tt.inV)
			}
			got, ok := tt.target.ValueB()
			if ok != tt.wantOk {
				t.Errorf("EitherOrNone[%s].ValueB() ok = %v, want %v", tt.name, ok, tt.wantOk)
			}
			if ok && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("EitherOrNone[%s].ValueB() = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func TestEitherOrNone_Type(t *testing.T) {
	type MyType interface {
		starlark.Unpacker
		Type() string
	}
	tests := []struct {
		name   string
		target MyType
		inV    starlark.Value
		want   string
	}{
		{
			name:   "string value",
			target: NewEitherOrNone[starlark.String, starlark.Int](),
			inV:    starlark.String("hello"),
			want:   "string",
		},
		{
			name:   "int value",
			target: NewEitherOrNone[starlark.String, starlark.Int](),
			inV:    starlark.MakeInt(42),
			want:   "int",
		},
		{
			name:   "none value",
			target: NewEitherOrNone[starlark.String, starlark.Int](),
			inV:    starlark.None,
			want:   "NoneType",
		},
		{
			name:   "dict type",
			target: NewEitherOrNone[*starlark.List, *starlark.Dict](),
			inV:    starlark.NewDict(1),
			want:   "dict",
		},
		{
			name:   "list type",
			target: NewEitherOrNone[*starlark.List, *starlark.Dict](),
			inV:    starlark.NewList([]starlark.Value{}),
			want:   "list",
		},
		{
			name:   "bool type",
			target: NewEitherOrNone[starlark.Bool, starlark.Float](),
			inV:    starlark.True,
			want:   "bool",
		},
		{
			name:   "float type",
			target: NewEitherOrNone[starlark.Bool, starlark.Float](),
			inV:    starlark.Float(3.14),
			want:   "float",
		},
		{
			name:   "builtin type",
			target: NewEitherOrNone[*starlark.Builtin, *starlark.Function](),
			inV: starlark.NewBuiltin("test", func(*starlark.Thread, *starlark.Builtin, starlark.Tuple, []starlark.Tuple) (starlark.Value, error) {
				return starlark.None, nil
			}),
			want: "builtin_function_or_method",
		},
		{
			name:   "function type",
			target: NewEitherOrNone[*starlark.Builtin, *starlark.Function](),
			inV:    &starlark.Function{},
			want:   "function",
		},
		{
			name:   "unknown type",
			target: new(EitherOrNone[starlark.String, starlark.Int]),
			want:   "Unknown",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.inV != nil && tt.target != nil {
				tt.target.Unpack(tt.inV)
			}
			if got := tt.target.Type(); got != tt.want {
				t.Errorf("EitherOrNone[%s].Type() = %v, want %v", tt.name, got, tt.want)
			}
		})
	}

	var target *EitherOrNone[starlark.String, starlark.Int]
	if got := target.Type(); got != "NilReceiver" {
		t.Errorf("EitherOrNone.Type() = %v, want %v", got, "NilReceiver")
	}
}
