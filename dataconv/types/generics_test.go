package types

import (
	"reflect"
	"testing"

	"go.starlark.net/starlark"
)

func TestNullableInt_Unpack(t *testing.T) {
	tests := []struct {
		name     string
		target   *NullableInt
		inV      starlark.Value
		want     starlark.Int
		wantNull bool
		wantErr  bool
	}{
		{
			name:    "nil int",
			target:  nil,
			inV:     starlark.MakeInt(10),
			wantErr: true,
		},
		{
			name:    "nil none",
			target:  nil,
			inV:     starlark.None,
			wantErr: true,
		},
		{
			name:   "no default val",
			target: &NullableInt{},
			inV:    starlark.MakeInt(10),
			want:   starlark.MakeInt(10),
		},
		{
			name:     "no default none",
			target:   &NullableInt{},
			inV:      starlark.None,
			want:     starlark.Int{},
			wantNull: true,
		},
		{
			name:   "int val",
			target: NewNullable(starlark.MakeInt(5)),
			inV:    starlark.MakeInt(10),
			want:   starlark.MakeInt(10),
		},
		{
			name:     "int none",
			target:   NewNullable(starlark.MakeInt(5)),
			inV:      starlark.None,
			want:     starlark.MakeInt(5),
			wantNull: true,
		},
		{
			name:    "int err",
			target:  NewNullable(starlark.MakeInt(5)),
			inV:     starlark.String("foo"),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// setup
			n := tt.name
			p := tt.target
			// run
			//err := p.Unpack(tt.inV)
			err := starlark.UnpackArgs("test", []starlark.Value{tt.inV}, nil, "v?", p)
			// check error
			if (err != nil) != tt.wantErr {
				t.Errorf("Nullable[%s].Unpack() error = %v, wantErr %v", n, err, tt.wantErr)
			} else if err != nil {
				t.Logf("Nullable[%s].Unpack() error = %v", n, err)
			}
			if tt.wantErr {
				return
			}
			// check methods
			if tt.wantNull != p.IsNull() {
				t.Errorf("Nullable[%s].IsNull() got = %v, want %v", n, p.IsNull(), tt.wantNull)
			}
			if !reflect.DeepEqual(p.Value(), tt.want) {
				t.Errorf("Nullable[%s].Unpack() got = %v, want %v", n, p.Value(), tt.want)
			}
		})
	}
}

func testNullableUnpack[T starlark.Value](t *testing.T, name string, target *Nullable[T], inV starlark.Value, want T, wantNull, wantErr bool) {
	t.Run(name, func(t *testing.T) {
		// run
		err := target.Unpack(inV)
		// check error
		if (err != nil) != wantErr {
			t.Errorf("Nullable[%s].Unpack() error = %v, wantErr %v", name, err, wantErr)
		} else if err != nil {
			t.Logf("Nullable[%s].Unpack() error = %v", name, err)
		}
		if wantErr {
			return
		}
		// check methods
		if wantNull != target.IsNull() {
			t.Errorf("Nullable[%s].IsNull() got = %v, want %v", name, target.IsNull(), wantNull)
		}
		if !reflect.DeepEqual(target.Value(), want) {
			t.Errorf("Nullable[%s].Unpack() got = %v, want %v", name, target.Value(), want)
		}
	})
}

func testNullableUnpackArgs[T starlark.Value](t *testing.T, name string, target *Nullable[T], inV starlark.Value, want T, wantNull, wantErr bool) {
	t.Run(name, func(t *testing.T) {
		// run
		err := starlark.UnpackArgs("test", []starlark.Value{inV}, nil, "v?", target)
		// check error
		if (err != nil) != wantErr {
			t.Errorf("Nullable[%s].UnpackArgs() error = %v, wantErr %v", name, err, wantErr)
		} else if err != nil {
			t.Logf("Nullable[%s].UnpackArgs() error = %v", name, err)
		}
		if wantErr {
			return
		}
		// check methods
		if wantNull != target.IsNull() {
			t.Errorf("Nullable[%s].IsNull() got = %v, want %v", name, target.IsNull(), wantNull)
		}
		if !reflect.DeepEqual(target.Value(), want) {
			t.Errorf("Nullable[%s].UnpackArgs() got = %v, want %v", name, target.Value(), want)
		}
	})
}

func TestNullableInt(t *testing.T) {
	testNullableUnpack(t, "int val", NewNullable(starlark.MakeInt(5)), starlark.MakeInt(10), starlark.MakeInt(10), false, false)
	testNullableUnpack(t, "int none", NewNullable(starlark.MakeInt(5)), starlark.None, starlark.MakeInt(5), true, false)
	testNullableUnpack(t, "int err", NewNullable(starlark.MakeInt(5)), starlark.String("foo"), starlark.MakeInt(5), true, true)
	testNullableUnpackArgs(t, "int val", NewNullable(starlark.MakeInt(5)), starlark.MakeInt(10), starlark.MakeInt(10), false, false)
	testNullableUnpackArgs(t, "int none", NewNullable(starlark.MakeInt(5)), starlark.None, starlark.MakeInt(5), true, false)
	testNullableUnpackArgs(t, "int err", NewNullable(starlark.MakeInt(5)), starlark.String("foo"), starlark.MakeInt(5), true, true)
}
