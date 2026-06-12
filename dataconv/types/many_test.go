package types

import (
	"reflect"
	"testing"

	"go.starlark.net/starlark"
)

func TestOneOrMany_Unpack(t *testing.T) {
	tests := []struct {
		name     string
		target   *OneOrMany[starlark.Int]
		inV      starlark.Value
		want     []starlark.Int
		wantNull bool
		wantErr  bool
	}{
		{
			name:    "nil receiver",
			target:  nil,
			inV:     starlark.MakeInt(10),
			wantErr: true,
		},
		{
			name:     "int value",
			target:   NewOneOrManyNoDefault[starlark.Int](),
			inV:      starlark.MakeInt(10),
			want:     []starlark.Int{starlark.MakeInt(10)},
			wantNull: false,
		},
		{
			name:     "none value",
			target:   NewOneOrManyNoDefault[starlark.Int](),
			inV:      none,
			want:     []starlark.Int{},
			wantNull: true,
		},
		{
			name:     "iterable value",
			target:   NewOneOrManyNoDefault[starlark.Int](),
			inV:      starlark.NewList([]starlark.Value{starlark.MakeInt(10), starlark.MakeInt(20)}),
			want:     []starlark.Int{starlark.MakeInt(10), starlark.MakeInt(20)},
			wantNull: false,
		},
		{
			name:    "wrong type value",
			target:  NewOneOrManyNoDefault[starlark.Int](),
			inV:     starlark.String("foo"),
			wantErr: true,
		},
		{
			name:    "iterable with wrong type",
			target:  NewOneOrManyNoDefault[starlark.Int](),
			inV:     starlark.NewList([]starlark.Value{starlark.MakeInt(10), starlark.String("foo")}),
			wantErr: true,
		},
		{
			name:    "iterable with mixed types",
			target:  NewOneOrManyNoDefault[starlark.Int](),
			inV:     starlark.NewList([]starlark.Value{starlark.MakeInt(10), starlark.MakeInt(20), starlark.String("foo")}),
			wantErr: true,
		},
		{
			name:     "iterable with empty list",
			target:   NewOneOrManyNoDefault[starlark.Int](),
			inV:      starlark.NewList([]starlark.Value{}),
			want:     []starlark.Int{},
			wantNull: false,
		},
		{
			name:   "iterable with default",
			target: NewOneOrMany(starlark.MakeInt(5)),
			inV:    starlark.NewList([]starlark.Value{starlark.MakeInt(10), starlark.MakeInt(20)}),
			want:   []starlark.Int{starlark.MakeInt(10), starlark.MakeInt(20)},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.target.Unpack(tt.inV)
			if (err != nil) != tt.wantErr {
				t.Errorf("OneOrMany[%s].Unpack() error = %v, wantErr %v", tt.name, err, tt.wantErr)
			} else if err != nil {
				t.Logf("OneOrMany[%s].Unpack() error = %v", tt.name, err)
			}
			if tt.wantErr {
				return
			}
			if tt.wantNull != tt.target.IsNull() {
				t.Errorf("OneOrMany[%s].IsNull() got = %v, want %v", tt.name, tt.target.IsNull(), tt.wantNull)
			}
			if !reflect.DeepEqual(tt.target.Slice(), tt.want) {
				t.Errorf("OneOrMany[%s].Unpack() got = %v, want %v", tt.name, tt.target.Slice(), tt.want)
			}
		})
	}
}

func TestOneOrMany_First(t *testing.T) {
	tests := []struct {
		name     string
		target   *OneOrMany[starlark.Int]
		want     starlark.Int
		wantNull bool
	}{
		{
			name:     "nil receiver",
			target:   nil,
			want:     starlark.Int{},
			wantNull: true,
		},
		{
			name:     "empty no default",
			target:   NewOneOrManyNoDefault[starlark.Int](),
			want:     starlark.Int{},
			wantNull: true,
		},
		{
			name:     "empty with default",
			target:   NewOneOrMany(starlark.MakeInt(5)),
			want:     starlark.MakeInt(5),
			wantNull: true,
		},
		{
			name:     "single value",
			target:   &OneOrMany[starlark.Int]{values: []starlark.Int{starlark.MakeInt(10)}, hasDefault: true, defaultValue: starlark.MakeInt(5)},
			want:     starlark.MakeInt(10),
			wantNull: false,
		},
		{
			name:     "multiple values",
			target:   &OneOrMany[starlark.Int]{values: []starlark.Int{starlark.MakeInt(10), starlark.MakeInt(20)}, hasDefault: true, defaultValue: starlark.MakeInt(5)},
			want:     starlark.MakeInt(10),
			wantNull: false,
		},
		{
			name:     "iterable with empty list",
			target:   &OneOrMany[starlark.Int]{values: []starlark.Int{}},
			want:     starlark.Int{},
			wantNull: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.target.First(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("OneOrMany[%s].First() = %v, want %v", tt.name, got, tt.want)
			}
			if tt.wantNull != tt.target.IsNull() {
				t.Errorf("OneOrMany[%s].IsNull() got = %v, want %v", tt.name, tt.target.IsNull(), tt.wantNull)
			}
		})
	}
}

func TestOneOrMany_Slice(t *testing.T) {
	tests := []struct {
		name   string
		target *OneOrMany[starlark.Int]
		want   []starlark.Int
	}{
		{
			name:   "nil receiver",
			target: nil,
			want:   []starlark.Int{},
		},
		{
			name:   "empty no default",
			target: NewOneOrManyNoDefault[starlark.Int](),
			want:   []starlark.Int{},
		},
		{
			name:   "empty with default",
			target: NewOneOrMany(starlark.MakeInt(5)),
			want:   []starlark.Int{starlark.MakeInt(5)},
		},
		{
			name:   "single value",
			target: &OneOrMany[starlark.Int]{values: []starlark.Int{starlark.MakeInt(10)}, hasDefault: true, defaultValue: starlark.MakeInt(5)},
			want:   []starlark.Int{starlark.MakeInt(10)},
		},
		{
			name:   "multiple values",
			target: &OneOrMany[starlark.Int]{values: []starlark.Int{starlark.MakeInt(10), starlark.MakeInt(20)}, hasDefault: true, defaultValue: starlark.MakeInt(5)},
			want:   []starlark.Int{starlark.MakeInt(10), starlark.MakeInt(20)},
		},
		{
			name:   "iterable with empty list",
			target: &OneOrMany[starlark.Int]{values: []starlark.Int{}},
			want:   []starlark.Int{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.target.Slice(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("OneOrMany[%s].Slice() = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func TestOneOrMany_IsNull(t *testing.T) {
	noDefault := NewOneOrManyNoDefault[starlark.Int]()
	noDefault.Unpack(none)
	withDefault := NewOneOrMany(starlark.MakeInt(5))
	withDefault.Unpack(none)
	tests := []struct {
		name   string
		target *OneOrMany[starlark.Int]
		want   bool
	}{
		{
			name:   "nil receiver",
			target: nil,
			want:   true,
		},
		{
			name:   "used no default",
			target: noDefault,
			want:   true,
		},
		{
			name:   "used with default",
			target: withDefault,
			want:   true,
		},
		{
			name:   "empty struct",
			target: &OneOrMany[starlark.Int]{},
			want:   true,
		},
		{
			name:   "empty no default",
			target: NewOneOrManyNoDefault[starlark.Int](),
			want:   true,
		},
		{
			name:   "empty with default",
			target: NewOneOrMany(starlark.MakeInt(5)),
			want:   true,
		},
		{
			name:   "single value",
			target: &OneOrMany[starlark.Int]{values: []starlark.Int{starlark.MakeInt(10)}, hasDefault: true, defaultValue: starlark.MakeInt(5)},
			want:   false,
		},
		{
			name:   "multiple values",
			target: &OneOrMany[starlark.Int]{values: []starlark.Int{starlark.MakeInt(10), starlark.MakeInt(20)}, hasDefault: true, defaultValue: starlark.MakeInt(5)},
			want:   false,
		},
		{
			name:   "iterable with empty list",
			target: &OneOrMany[starlark.Int]{values: []starlark.Int{}},
			want:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.target.IsNull(); got != tt.want {
				t.Errorf("OneOrMany[%s].IsNull() = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func TestOneOrMany_Len(t *testing.T) {
	noDefault := NewOneOrManyNoDefault[starlark.Int]()
	noDefault.Unpack(none)
	withDefault := NewOneOrMany(starlark.MakeInt(5))
	withDefault.Unpack(none)
	tests := []struct {
		name   string
		target *OneOrMany[starlark.Int]
		want   int
	}{
		{
			name:   "nil receiver",
			target: nil,
			want:   0,
		},
		{
			name:   "used no default",
			target: noDefault,
			want:   0,
		},
		{
			name:   "used with default",
			target: withDefault,
			want:   1,
		},
		{
			name:   "empty no default",
			target: NewOneOrManyNoDefault[starlark.Int](),
			want:   0,
		},
		{
			name:   "empty with default",
			target: NewOneOrMany(starlark.MakeInt(5)),
			want:   1,
		},
		{
			name:   "single value",
			target: &OneOrMany[starlark.Int]{values: []starlark.Int{starlark.MakeInt(10)}},
			want:   1,
		},
		{
			name:   "multiple values",
			target: &OneOrMany[starlark.Int]{values: []starlark.Int{starlark.MakeInt(10), starlark.MakeInt(20)}},
			want:   2,
		},
		{
			name:   "multiple values with default",
			target: &OneOrMany[starlark.Int]{values: []starlark.Int{starlark.MakeInt(10), starlark.MakeInt(20)}, hasDefault: true, defaultValue: starlark.MakeInt(5)},
			want:   2,
		},
		{
			name:   "empty without default, no values",
			target: &OneOrMany[starlark.Int]{},
			want:   0,
		},
		{
			name:   "iterable with empty list",
			target: &OneOrMany[starlark.Int]{values: []starlark.Int{}},
			want:   0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.target.Len(); got != tt.want {
				t.Errorf("OneOrMany[%s].Len() = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func TestOneOrMany_UnpackArgs(t *testing.T) {
	tests := []struct {
		name     string
		target   *OneOrMany[starlark.Int]
		inV      starlark.Value
		want     []starlark.Int
		wantNull bool
		wantErr  bool
	}{
		{
			name:    "nil receiver",
			target:  nil,
			inV:     starlark.MakeInt(10),
			wantErr: true,
		},
		{
			name:    "empty struct with list of mistype",
			target:  &OneOrMany[starlark.Int]{},
			inV:     starlark.NewList([]starlark.Value{starlark.String("s")}),
			wantErr: true,
		},
		{
			name:    "empty struct with list of nil",
			target:  &OneOrMany[starlark.Int]{},
			inV:     starlark.NewList([]starlark.Value{nil}),
			wantErr: true,
		},
		{
			name:     "int value",
			target:   NewOneOrManyNoDefault[starlark.Int](),
			inV:      starlark.MakeInt(10),
			want:     []starlark.Int{starlark.MakeInt(10)},
			wantNull: false,
		},
		{
			name:     "none value",
			target:   NewOneOrManyNoDefault[starlark.Int](),
			inV:      none,
			want:     []starlark.Int{},
			wantNull: true,
		},
		{
			name:     "iterable empty without default",
			target:   NewOneOrManyNoDefault[starlark.Int](),
			inV:      starlark.NewList([]starlark.Value{}),
			want:     []starlark.Int{},
			wantNull: false,
		},
		{
			name:     "iterable empty with default",
			target:   NewOneOrMany(starlark.MakeInt(5)),
			inV:      starlark.NewList([]starlark.Value{}),
			want:     []starlark.Int{},
			wantNull: false,
		},
		{
			name:     "iterable value",
			target:   NewOneOrManyNoDefault[starlark.Int](),
			inV:      starlark.NewList([]starlark.Value{starlark.MakeInt(10), starlark.MakeInt(20)}),
			want:     []starlark.Int{starlark.MakeInt(10), starlark.MakeInt(20)},
			wantNull: false,
		},
		{
			name:    "wrong type value",
			target:  NewOneOrManyNoDefault[starlark.Int](),
			inV:     starlark.String("foo"),
			wantErr: true,
		},
		{
			name:    "iterable with wrong type",
			target:  NewOneOrManyNoDefault[starlark.Int](),
			inV:     starlark.NewList([]starlark.Value{starlark.MakeInt(10), starlark.String("foo")}),
			wantErr: true,
		},
		{
			name:    "iterable with mixed types",
			target:  NewOneOrManyNoDefault[starlark.Int](),
			inV:     starlark.NewList([]starlark.Value{starlark.MakeInt(10), starlark.MakeInt(20), starlark.String("foo")}),
			wantErr: true,
		},
		{
			name:     "iterable with empty list",
			target:   NewOneOrManyNoDefault[starlark.Int](),
			inV:      starlark.NewList([]starlark.Value{}),
			want:     []starlark.Int{},
			wantNull: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := starlark.UnpackArgs("test", []starlark.Value{tt.inV}, nil, "v?", tt.target)
			if (err != nil) != tt.wantErr {
				t.Errorf("OneOrMany[%s].UnpackArgs() error = %v, wantErr %v", tt.name, err, tt.wantErr)
			} else if err != nil {
				t.Logf("OneOrMany[%s].UnpackArgs() error = %v", tt.name, err)
			}
			if tt.wantErr {
				return
			}
			if tt.wantNull != tt.target.IsNull() {
				t.Errorf("OneOrMany[%s].IsNull() got = %v, want %v", tt.name, tt.target.IsNull(), tt.wantNull)
			}
			if !reflect.DeepEqual(tt.target.Slice(), tt.want) {
				t.Errorf("OneOrMany[%s].UnpackArgs() got = %v, want %v", tt.name, tt.target.Slice(), tt.want)
			}
		})
	}
}

func TestOneOrMany_StringVsValue(t *testing.T) {
	// Create a list of strings
	strList := starlark.NewList([]starlark.Value{
		starlark.String("abc"),
		starlark.String("def"),
	})

	// Test OneOrMany[starlark.String]
	t.Run("OneOrMany[starlark.String]", func(t *testing.T) {
		target := NewOneOrManyNoDefault[starlark.String]()
		err := target.Unpack(strList)
		if err != nil {
			t.Errorf("OneOrMany[starlark.String].Unpack() error = %v", err)
			return
		}

		expected := []starlark.String{starlark.String("abc"), starlark.String("def")}
		if !reflect.DeepEqual(target.Slice(), expected) {
			t.Errorf("OneOrMany[starlark.String].Unpack() got = %v, want %v", target.Slice(), expected)
		}

		// Verify first element
		if target.First() != starlark.String("abc") {
			t.Errorf("OneOrMany[starlark.String].First() got = %v, want %v", target.First(), starlark.String("abc"))
		}

		// Verify length
		if target.Len() != 2 {
			t.Errorf("OneOrMany[starlark.String].Len() got = %v, want %v", target.Len(), 2)
		}
	})

	// Test OneOrMany[starlark.Value]
	t.Run("OneOrMany[starlark.Value]", func(t *testing.T) {
		target := NewOneOrManyNoDefault[starlark.Value]()
		err := target.Unpack(strList)
		if err != nil {
			t.Errorf("OneOrMany[starlark.Value].Unpack() error = %v", err)
			return
		}

		// For starlark.Value type, the whole list is treated as a single value
		values := target.Slice()
		if len(values) != 1 {
			t.Errorf("OneOrMany[starlark.Value].Unpack() got len = %v, want %v", len(values), 1)
			return
		}

		// Verify the value is a list type
		list, ok := values[0].(*starlark.List)
		if !ok {
			t.Errorf("OneOrMany[starlark.Value].Unpack() value is not starlark.List type, but %T", values[0])
			return
		}

		// Verify list contents
		if list.Len() != 2 {
			t.Errorf("List length got = %v, want %v", list.Len(), 2)
			return
		}

		v1, ok := starlark.AsString(list.Index(0))
		if !ok || v1 != "abc" {
			t.Errorf("First list element got = %v, want %v", v1, "abc")
		}

		v2, ok := starlark.AsString(list.Index(1))
		if !ok || v2 != "def" {
			t.Errorf("Second list element got = %v, want %v", v2, "def")
		}

		// Verify first element is the list itself
		first, ok := target.First().(*starlark.List)
		if !ok || first != list {
			t.Errorf("OneOrMany[starlark.Value].First() result is not the expected list")
		}
	})

	// Test unpacking a single string
	t.Run("Single_String", func(t *testing.T) {
		singleStr := starlark.String("xyz")

		// For starlark.String type
		stringTarget := NewOneOrManyNoDefault[starlark.String]()
		if err := stringTarget.Unpack(singleStr); err != nil {
			t.Errorf("OneOrMany[starlark.String].Unpack(single) error = %v", err)
			return
		}

		if stringTarget.First() != starlark.String("xyz") {
			t.Errorf("OneOrMany[starlark.String].First() got = %v, want %v",
				stringTarget.First(), starlark.String("xyz"))
		}

		// For starlark.Value type
		valueTarget := NewOneOrManyNoDefault[starlark.Value]()
		if err := valueTarget.Unpack(singleStr); err != nil {
			t.Errorf("OneOrMany[starlark.Value].Unpack(single) error = %v", err)
			return
		}

		first, ok := valueTarget.First().(starlark.String)
		if !ok || first != starlark.String("xyz") {
			t.Errorf("OneOrMany[starlark.Value].First() got = %v, want %v",
				valueTarget.First(), starlark.String("xyz"))
		}
	})

	// Test UnpackArgs method differences
	t.Run("UnpackArgs", func(t *testing.T) {
		// For starlark.String type
		stringTarget := NewOneOrManyNoDefault[starlark.String]()
		if err := starlark.UnpackArgs("test", []starlark.Value{strList}, nil, "v?", stringTarget); err != nil {
			t.Errorf("OneOrMany[starlark.String].UnpackArgs() error = %v", err)
			return
		}

		expected := []starlark.String{starlark.String("abc"), starlark.String("def")}
		if !reflect.DeepEqual(stringTarget.Slice(), expected) {
			t.Errorf("OneOrMany[starlark.String].UnpackArgs() got = %v, want %v",
				stringTarget.Slice(), expected)
		}

		// For starlark.Value type
		valueTarget := NewOneOrManyNoDefault[starlark.Value]()
		if err := starlark.UnpackArgs("test", []starlark.Value{strList}, nil, "v?", valueTarget); err != nil {
			t.Errorf("OneOrMany[starlark.Value].UnpackArgs() error = %v", err)
			return
		}

		values := valueTarget.Slice()
		if len(values) != 1 {
			t.Errorf("OneOrMany[starlark.Value].UnpackArgs() got len = %v, want %v",
				len(values), 1)
			return
		}

		// Verify first element is a list
		list, ok := values[0].(*starlark.List)
		if !ok {
			t.Errorf("OneOrMany[starlark.Value].UnpackArgs() value is not starlark.List type, but %T", values[0])
			return
		}

		// Verify list contents
		if list.Len() != 2 {
			t.Errorf("List length got = %v, want %v", list.Len(), 2)
			return
		}

		v1, ok := starlark.AsString(list.Index(0))
		if !ok || v1 != "abc" {
			t.Errorf("First list element got = %v, want %v", v1, "abc")
		}

		v2, ok := starlark.AsString(list.Index(1))
		if !ok || v2 != "def" {
			t.Errorf("Second list element got = %v, want %v", v2, "def")
		}
	})
}
