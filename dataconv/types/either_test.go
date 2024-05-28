package types

import (
	"reflect"
	"testing"

	"go.starlark.net/starlark"
)

func TestEitherOrNone_Type(t *testing.T) {
	x := NewEitherOrNone[starlark.Int, starlark.String]()
	x.Unpack(starlark.None)
	t.Logf("Type: %s", x.Type())

	x.Unpack(starlark.MakeInt(42))
	t.Logf("Type: %s", x.Type())

	x.Unpack(starlark.String("hello"))
	t.Logf("Type: %s", x.Type())

	x.Unpack(starlark.NewDict(1))
	t.Logf("Type: %s", x.Type())

	y := NewEitherOrNone[*starlark.List, *starlark.Dict]()
	t.Logf("Type: %s", y.Type())

	y.Unpack(starlark.None)
	t.Logf("Type: %s", y.Type())

	y.Unpack(starlark.NewList([]starlark.Value{starlark.None}))
	t.Logf("Type: %s", y.Type())

	y.Unpack(starlark.NewDict(1))
	t.Logf("Type: %s", y.Type())
}

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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.target.Unpack(tt.inV)
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.target.Unpack(tt.inV)
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
