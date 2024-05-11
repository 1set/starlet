package types

import (
	"reflect"
	"testing"

	"go.starlark.net/starlark"
)

var (
	none = starlark.None
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
			inV:     none,
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
			inV:      none,
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
			inV:      none,
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
	name = name + " args"
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
	testNullableUnpack(t, "empty val", &NullableInt{}, starlark.MakeInt(5), starlark.MakeInt(5), false, false)
	testNullableUnpack(t, "empty none", &NullableInt{}, none, starlark.Int{}, true, false)
	testNullableUnpack(t, "int val", NewNullable(starlark.MakeInt(5)), starlark.MakeInt(10), starlark.MakeInt(10), false, false)
	testNullableUnpack(t, "int none", NewNullable(starlark.MakeInt(5)), none, starlark.MakeInt(5), true, false)
	testNullableUnpack(t, "int err", NewNullable(starlark.MakeInt(5)), starlark.String("foo"), starlark.MakeInt(5), true, true)

	testNullableUnpackArgs(t, "empty val", &NullableInt{}, starlark.MakeInt(5), starlark.MakeInt(5), false, false)
	testNullableUnpackArgs(t, "empty none", &NullableInt{}, none, starlark.Int{}, true, false)
	testNullableUnpackArgs(t, "int val", NewNullable(starlark.MakeInt(5)), starlark.MakeInt(10), starlark.MakeInt(10), false, false)
	testNullableUnpackArgs(t, "int none", NewNullable(starlark.MakeInt(5)), none, starlark.MakeInt(5), true, false)
	testNullableUnpackArgs(t, "int err", NewNullable(starlark.MakeInt(5)), starlark.String("foo"), starlark.MakeInt(5), true, true)
}

func TestNullableFloat(t *testing.T) {
	defaultVal := starlark.Float(1.5)
	newVal := starlark.Float(2.5)

	testNullableUnpack(t, "empty val", &NullableFloat{}, newVal, newVal, false, false)
	testNullableUnpack(t, "empty none", &NullableFloat{}, none, starlark.Float(0), true, false)
	testNullableUnpack(t, "float val", NewNullable(defaultVal), newVal, newVal, false, false)
	testNullableUnpack(t, "float none", NewNullable(defaultVal), none, defaultVal, true, false)
	testNullableUnpack(t, "float err", NewNullable(defaultVal), starlark.String("not a float"), defaultVal, true, true)

	testNullableUnpackArgs(t, "empty val", &NullableFloat{}, newVal, newVal, false, false)
	testNullableUnpackArgs(t, "empty none", &NullableFloat{}, none, starlark.Float(0), true, false)
	testNullableUnpackArgs(t, "float val", NewNullable(defaultVal), newVal, newVal, false, false)
	testNullableUnpackArgs(t, "float none", NewNullable(defaultVal), none, defaultVal, true, false)
	testNullableUnpackArgs(t, "float err", NewNullable(defaultVal), starlark.String("not a float"), defaultVal, true, true)
}

func TestNullableBool(t *testing.T) {
	trueVal := starlark.Bool(true)
	falseVal := starlark.Bool(false)

	testNullableUnpack(t, "empty val", &NullableBool{}, trueVal, trueVal, false, false)
	testNullableUnpack(t, "empty none", &NullableBool{}, none, false, true, false)
	testNullableUnpack(t, "bool true", NewNullable(trueVal), trueVal, trueVal, false, false)
	testNullableUnpack(t, "bool false", NewNullable(trueVal), falseVal, falseVal, false, false)
	testNullableUnpack(t, "bool none", NewNullable(trueVal), none, trueVal, true, false)
	testNullableUnpack(t, "bool err", NewNullable(trueVal), starlark.String("not a bool"), trueVal, true, true)

	testNullableUnpackArgs(t, "empty val", &NullableBool{}, trueVal, trueVal, false, false)
	testNullableUnpackArgs(t, "empty none", &NullableBool{}, none, false, true, false)
	testNullableUnpackArgs(t, "bool true", NewNullable(trueVal), trueVal, trueVal, false, false)
	testNullableUnpackArgs(t, "bool false", NewNullable(trueVal), falseVal, falseVal, false, false)
	testNullableUnpackArgs(t, "bool none", NewNullable(trueVal), none, trueVal, true, false)
	testNullableUnpackArgs(t, "bool err", NewNullable(trueVal), starlark.String("not a bool"), trueVal, true, true)
}

func TestNullableList(t *testing.T) {
	var nilList *starlark.List

	list12 := starlark.NewList([]starlark.Value{starlark.MakeInt(1), starlark.MakeInt(2)})
	list34 := starlark.NewList([]starlark.Value{starlark.MakeInt(3), starlark.MakeInt(4)})

	testNullableUnpack(t, "empty val", &NullableList{}, list12, list12, false, false)
	testNullableUnpack(t, "empty none", &NullableList{}, none, nilList, true, false)
	testNullableUnpack(t, "list val", NewNullable(list12), list34, list34, false, false)
	testNullableUnpack(t, "list none", NewNullable(list12), none, list12, true, false)
	testNullableUnpack(t, "list err", NewNullable(list12), starlark.String("foo"), list12, true, true)

	testNullableUnpackArgs(t, "empty val", &NullableList{}, list12, list12, false, false)
	testNullableUnpackArgs(t, "empty none", &NullableList{}, none, nilList, true, false)
	testNullableUnpackArgs(t, "list val", NewNullable(list12), list34, list34, false, false)
	testNullableUnpackArgs(t, "list none", NewNullable(list12), none, list12, true, false)
	testNullableUnpackArgs(t, "list err", NewNullable(list12), starlark.String("foo"), list12, true, true)
}

func TestNullableTuple(t *testing.T) {
	var nilTuple starlark.Tuple

	tuple12 := starlark.Tuple{starlark.MakeInt(1), starlark.MakeInt(2)}
	tuple34 := starlark.Tuple{starlark.MakeInt(3), starlark.MakeInt(4)}

	testNullableUnpack(t, "empty val", &NullableTuple{}, tuple12, tuple12, false, false)
	testNullableUnpack(t, "empty none", &NullableTuple{}, none, nilTuple, true, false)
	testNullableUnpack(t, "tuple val", NewNullable(tuple12), tuple34, tuple34, false, false)
	testNullableUnpack(t, "tuple none", NewNullable(tuple12), none, tuple12, true, false)
	testNullableUnpack(t, "tuple err", NewNullable(tuple12), starlark.String("foo"), tuple12, true, true)

	testNullableUnpackArgs(t, "empty val", &NullableTuple{}, tuple12, tuple12, false, false)
	testNullableUnpackArgs(t, "empty none", &NullableTuple{}, none, nilTuple, true, false)
	testNullableUnpackArgs(t, "tuple val", NewNullable(tuple12), tuple34, tuple34, false, false)
	testNullableUnpackArgs(t, "tuple none", NewNullable(tuple12), none, tuple12, true, false)
	testNullableUnpackArgs(t, "tuple err", NewNullable(tuple12), starlark.String("foo"), tuple12, true, true)
}

func TestNullableSet(t *testing.T) {
	var nilSet *starlark.Set
	defaultSet := starlark.NewSet(2)
	defaultSet.Insert(starlark.MakeInt(1))
	defaultSet.Insert(starlark.MakeInt(2))
	newSet := starlark.NewSet(2)
	newSet.Insert(starlark.MakeInt(3))
	newSet.Insert(starlark.MakeInt(4))

	testNullableUnpack(t, "empty val", &NullableSet{}, newSet, newSet, false, false)
	testNullableUnpack(t, "empty none", &NullableSet{}, none, nilSet, true, false)
	testNullableUnpack(t, "set val", NewNullable(defaultSet), newSet, newSet, false, false)
	testNullableUnpack(t, "set none", NewNullable(defaultSet), none, defaultSet, true, false)
	testNullableUnpack(t, "set err", NewNullable(defaultSet), starlark.String("foo"), defaultSet, true, true)

	testNullableUnpackArgs(t, "empty val", &NullableSet{}, newSet, newSet, false, false)
	testNullableUnpackArgs(t, "empty none", &NullableSet{}, none, nilSet, true, false)
	testNullableUnpackArgs(t, "set val", NewNullable(defaultSet), newSet, newSet, false, false)
	testNullableUnpackArgs(t, "set none", NewNullable(defaultSet), none, defaultSet, true, false)
	testNullableUnpackArgs(t, "set err", NewNullable(defaultSet), starlark.String("foo"), defaultSet, true, true)
}

func TestNullableDict(t *testing.T) {
	var nilDict *starlark.Dict
	newDict := starlark.NewDict(1)
	newDict.SetKey(starlark.String("aloha"), starlark.MakeInt(100))
	defaultDict := starlark.NewDict(1)
	defaultDict.SetKey(starlark.String("got"), starlark.MakeInt(1))

	type NullableDict = Nullable[*starlark.Dict]

	testNullableUnpack(t, "empty val", &NullableDict{}, newDict, newDict, false, false)
	testNullableUnpack(t, "empty none", &NullableDict{}, none, nilDict, true, false)
	testNullableUnpack(t, "dict val", NewNullable(defaultDict), newDict, newDict, false, false)
	testNullableUnpack(t, "dict none", NewNullable(defaultDict), none, defaultDict, true, false)
	testNullableUnpack(t, "dict err", NewNullable(defaultDict), starlark.String("foo"), defaultDict, true, true)

	testNullableUnpackArgs(t, "empty val", &NullableDict{}, newDict, newDict, false, false)
	testNullableUnpackArgs(t, "empty none", &NullableDict{}, none, nilDict, true, false)
	testNullableUnpackArgs(t, "dict val", NewNullable(defaultDict), newDict, newDict, false, false)
	testNullableUnpackArgs(t, "dict none", NewNullable(defaultDict), none, defaultDict, true, false)
	testNullableUnpackArgs(t, "dict err", NewNullable(defaultDict), starlark.String("foo"), defaultDict, true, true)
}

func TestNullableIterable(t *testing.T) {
	defaultList := starlark.NewList([]starlark.Value{starlark.MakeInt(1), starlark.MakeInt(2), starlark.MakeInt(3)})
	defaultSet := starlark.NewSet(2)
	defaultSet.Insert(starlark.MakeInt(10))
	defaultSet.Insert(starlark.MakeInt(20))
	defaultDict := starlark.NewDict(1)
	defaultDict.SetKey(starlark.String("aloha"), starlark.MakeInt(100))

	list := starlark.NewList([]starlark.Value{starlark.MakeInt(4), starlark.MakeInt(5), starlark.MakeInt(6)})
	set := starlark.NewSet(2)
	set.Insert(starlark.MakeInt(30))
	set.Insert(starlark.MakeInt(40))
	dict := starlark.NewDict(1)
	dict.SetKey(starlark.String("flower"), starlark.MakeInt(200))

	var nilItr starlark.Iterable
	testNullableUnpack(t, "empty none", &NullableIterable{}, none, nilItr, true, false)
	testNullableUnpackArgs(t, "empty none", &NullableIterable{}, none, nilItr, true, false)

	testNullableUnpack(t, "list val", NewNullable(defaultList), list, list, false, false)
	testNullableUnpack(t, "list none", NewNullable(defaultList), none, defaultList, true, false)
	testNullableUnpack(t, "list err", NewNullable(defaultList), starlark.String("foo"), defaultList, true, true)
	testNullableUnpackArgs(t, "list val", NewNullable(defaultList), list, list, false, false)
	testNullableUnpackArgs(t, "list none", NewNullable(defaultList), none, defaultList, true, false)
	testNullableUnpackArgs(t, "list err", NewNullable(defaultList), starlark.String("foo"), defaultList, true, true)

	testNullableUnpack(t, "set val", NewNullable(defaultSet), set, set, false, false)
	testNullableUnpack(t, "set none", NewNullable(defaultSet), none, defaultSet, true, false)
	testNullableUnpack(t, "set err", NewNullable(defaultSet), starlark.String("foo"), defaultSet, true, true)
	testNullableUnpackArgs(t, "set val", NewNullable(defaultSet), set, set, false, false)
	testNullableUnpackArgs(t, "set none", NewNullable(defaultSet), none, defaultSet, true, false)
	testNullableUnpackArgs(t, "set err", NewNullable(defaultSet), starlark.String("foo"), defaultSet, true, true)

	testNullableUnpack(t, "dict val", NewNullable(defaultDict), dict, dict, false, false)
	testNullableUnpack(t, "dict none", NewNullable(defaultDict), none, defaultDict, true, false)
	testNullableUnpack(t, "dict err", NewNullable(defaultDict), starlark.String("foo"), defaultDict, true, true)
	testNullableUnpackArgs(t, "dict val", NewNullable(defaultDict), dict, dict, false, false)
	testNullableUnpackArgs(t, "dict none", NewNullable(defaultDict), none, defaultDict, true, false)
	testNullableUnpackArgs(t, "dict err", NewNullable(defaultDict), starlark.String("foo"), defaultDict, true, true)
}

func TestNullableCallable(t *testing.T) {
	defaultCallable := starlark.NewBuiltin("default", func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		return starlark.MakeInt(1), nil
	})
	callable := starlark.NewBuiltin("foo", func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		return starlark.MakeInt(42), nil
	})

	testNullableUnpack(t, "empty none", &NullableCallable{}, none, nil, true, false)
	testNullableUnpack(t, "callable val", NewNullable(defaultCallable), callable, callable, false, false)
	testNullableUnpack(t, "callable none", NewNullable(defaultCallable), none, defaultCallable, true, false)
	testNullableUnpack(t, "callable err", NewNullable(defaultCallable), starlark.String("foo"), defaultCallable, true, true)

	testNullableUnpackArgs(t, "empty none", &NullableCallable{}, none, nil, true, false)
	testNullableUnpackArgs(t, "callable val", NewNullable(defaultCallable), callable, callable, false, false)
	testNullableUnpackArgs(t, "callable none", NewNullable(defaultCallable), none, defaultCallable, true, false)
	testNullableUnpackArgs(t, "callable err", NewNullable(defaultCallable), starlark.String("foo"), defaultCallable, true, true)
}
