package dataconv

import (
	"fmt"
	"sort"
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

func (s *SharedDict) Get(k starlark.Value) (v starlark.Value, found bool, err error) {
	s.RLock()
	defer s.RUnlock()

	if s.dict != nil {
		return s.dict.Get(k)
	}
	return nil, false, nil
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
	s.Lock()
	defer s.Unlock()

	// basic check
	if s.dict == nil {
		return nil, nil
	}

	var (
		attr starlark.Value
		err  error
	)
	// try to get the new custom builtin
	if b, ok := customSharedDictMethods[name]; ok {
		attr = b.BindReceiver(s.dict)
	} else {
		// get the builtin from the original dict
		attr, err = s.dict.Attr(name)
	}

	// convert to builtin
	if attr == nil || err != nil {
		return attr, err
	}
	btl, ok := attr.(*starlark.Builtin)
	if !ok {
		return nil, fmt.Errorf("unsupported attribute: %s", name)
	}

	// wrap the builtin
	return starlark.NewBuiltin(name, func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		// lock the shared dict
		s.Lock()
		defer s.Unlock()

		// call the original builtin
		return btl.CallInternal(thread, args, kwargs)
	}), nil
}

func (s *SharedDict) AttrNames() []string {
	if s.dict != nil {
		names := s.dict.AttrNames()
		for cn := range customSharedDictMethods {
			names = append(names, cn)
		}
		sort.Strings(names)
		return names
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

var (
	customSharedDictMethods = map[string]*starlark.Builtin{
		"len":     starlark.NewBuiltin("len", shardDictLen),
		"perform": starlark.NewBuiltin("perform", shardDictPerform),
	}
)

// shardDictLen returns the length of the underlying dictionary.
func shardDictLen(_ *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if err := starlark.UnpackPositionalArgs(b.Name(), args, kwargs, 0); err != nil {
		return nil, err
	}
	l := b.Receiver().(*starlark.Dict).Len()
	return starlark.MakeInt(l), nil
}

// shardDictPerform calls the given function with the underlying receiver dictionary, and returns the result.
// The function must be callable, like def perform(fn).
func shardDictPerform(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	// get the perform function
	var pr starlark.Value
	if err := starlark.UnpackArgs(b.Name(), args, kwargs, "fn", &pr); err != nil {
		return nil, err
	}

	// get the receiver
	d := b.Receiver().(*starlark.Dict)

	// call the function with the receiver
	switch pr := pr.(type) {
	case starlark.Callable:
		return pr.CallInternal(thread, starlark.Tuple{d}, nil)
	default:
		return nil, fmt.Errorf("%s: not callable type: %s", b.Name(), pr.Type())
	}
}
