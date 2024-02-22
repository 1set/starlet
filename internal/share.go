package internal

import (
	"fmt"
	"sync"

	"go.starlark.net/starlark"
	"go.starlark.net/syntax"
)

// SharedDict is a dictionary that can be shared among multiple Starlark threads.
type SharedDict struct {
	sync.RWMutex
	dict   *starlark.Dict
	frozen bool
}

// NewSharedDict creates a new SharedDict instance.
func NewSharedDict() *SharedDict {
	return &SharedDict{
		dict: starlark.NewDict(1),
	}
}

//// Assert at compile time that Duration implements Unpacker.
//var _ starlark.Unpacker = (*SharedDict)(nil)
//
//// Unpack implements Unpacker for custom type SharedDict.
//func (s *SharedDict) Unpack(v starlark.Value) error {
//	dict, ok := v.(*starlark.Dict)
//	if !ok {
//		return fmt.Errorf("got %s, want dict", v.Type())
//	}
//
//	s.Lock()
//	defer s.Unlock()
//	s.dict = dict
//	return nil
//}

var (
	_ starlark.Value      = (*SharedDict)(nil)
	_ starlark.Comparable = (*SharedDict)(nil)
	_ starlark.Mapping    = (*SharedDict)(nil)
	_ starlark.HasAttrs   = (*SharedDict)(nil)
	_ starlark.HasSetKey  = (*SharedDict)(nil)
)

func (s *SharedDict) String() string {
	var v string
	if s != nil && s.dict != nil {
		v = s.dict.String()
	}
	return fmt.Sprintf("shared_dict(%s)", v)
}

func (s *SharedDict) Type() string {
	return "shared_dict"
}

// Freeze prevents the SharedDict from being modified.
func (s *SharedDict) Freeze() {
	s.Lock()
	defer s.Unlock()

	s.frozen = true
	if s.dict != nil {
		s.dict.Freeze()
	}
}

// Truth returns the truth value of the SharedDict.
func (s *SharedDict) Truth() starlark.Bool {
	s.RLock()
	defer s.RUnlock()

	return s != nil && s.dict != nil && s.dict.Truth()
}

// Hash returns the hash value of the SharedDict, actually it's not hashable.
func (s *SharedDict) Hash() (uint32, error) {
	return 0, fmt.Errorf("unhashable type: shared_dict")
}

// The following methods are wrapped around the underlying starlark.Dict.

func (s *SharedDict) Clear() error {
	s.Lock()
	defer s.Unlock()

	if s.frozen {
		return fmt.Errorf("cannot clear a frozen shared_dict")
	}
	if s.dict != nil {
		return s.dict.Clear()
	}
	return nil
}

func (s *SharedDict) Delete(k starlark.Value) (v starlark.Value, found bool, err error) {
	s.Lock()
	defer s.Unlock()

	if s.frozen {
		return nil, false, fmt.Errorf("cannot delete from a frozen shared_dict")
	}
	if s.dict != nil {
		return s.dict.Delete(k)
	}
	return nil, false, nil
}

func (s *SharedDict) Get(k starlark.Value) (v starlark.Value, found bool, err error) {
	s.RLock()
	defer s.RUnlock()

	if s.dict != nil {
		return s.dict.Get(k)
	}
	return nil, false, nil
}

func (s *SharedDict) Items() []starlark.Tuple {
	s.RLock()
	defer s.RUnlock()

	if s.dict != nil {
		return s.dict.Items()
	}
	return nil
}

func (s *SharedDict) Keys() []starlark.Value {
	s.RLock()
	defer s.RUnlock()

	if s.dict != nil {
		return s.dict.Keys()
	}
	return nil
}

func (s *SharedDict) Len() int {
	s.RLock()
	defer s.RUnlock()

	if s.dict != nil {
		return s.dict.Len()
	}
	return 0
}

func (s *SharedDict) SetKey(k, v starlark.Value) error {
	s.Lock()
	defer s.Unlock()

	if s.frozen {
		return fmt.Errorf("frozen dict")
	}
	if s.dict == nil {
		s.dict = &starlark.Dict{}
	}
	return s.dict.SetKey(k, v)
}

func (s *SharedDict) Attr(name string) (starlark.Value, error) {
	s.RLock()
	defer s.RUnlock()

	if s.dict != nil {
		return s.dict.Attr(name)
	}
	return nil, nil
}

func (s *SharedDict) AttrNames() []string {
	if s.dict != nil {
		return s.dict.AttrNames()
	}
	return nil
}

func (s *SharedDict) CompareSameType(op syntax.Token, y_ starlark.Value, depth int) (bool, error) {
	// if they are the same object, they are equal
	if s == y_ {
		return true, nil
	}

	// scan the type
	y := y_.(*SharedDict)

	// lock both objects
	s.RLock()
	defer s.RUnlock()
	y.RLock()
	defer y.RUnlock()

	// compare the underlying dictionaries
	if s.dict != nil && y.dict != nil {
		return s.dict.CompareSameType(op, y.dict, depth)
	}
	return false, nil
}
