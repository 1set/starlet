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
