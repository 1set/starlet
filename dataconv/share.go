package dataconv

import (
	"fmt"
	"sort"
	"sync"

	itn "github.com/1set/starlet/internal"
	stdjson "go.starlark.net/lib/json"
	"go.starlark.net/starlark"
	"go.starlark.net/syntax"
)

// SharedDict is a dictionary that can be shared among multiple Starlark threads.
type SharedDict struct {
	sync.RWMutex
	dict   *starlark.Dict
	frozen bool
}

const (
	defaultSharedDictSize = 8
)

// NewSharedDict creates a new SharedDict instance.
func NewSharedDict() *SharedDict {
	return &SharedDict{
		dict: starlark.NewDict(defaultSharedDictSize),
	}
}

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

// Type returns the type name of the SharedDict.
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

// Get returns the value corresponding to the specified key, or not found if the mapping does not contain the key.
// It implements the starlark.Mapping interface.
func (s *SharedDict) Get(k starlark.Value) (v starlark.Value, found bool, err error) {
	s.RLock()
	defer s.RUnlock()

	if s.dict != nil {
		return s.dict.Get(k)
	}
	return nil, false, nil
}

// SetKey sets the value for the specified key, supports update using x[k]=v syntax, like a dictionary.
// It implements the starlark.HasSetKey interface.
func (s *SharedDict) SetKey(k, v starlark.Value) error {
	s.Lock()
	defer s.Unlock()

	// basic check
	if s.frozen {
		return fmt.Errorf("frozen dict")
	}

	// maybe create the dictionary (perhaps this line is unreachable)
	if s.dict == nil {
		s.dict = starlark.NewDict(defaultSharedDictSize)
	}

	// check if the value is a shared dict -- reject it
	if sd, ok := v.(*SharedDict); ok {
		return fmt.Errorf("unsupported value: %s", sd.Type())
	}
	return s.dict.SetKey(k, v)
}

// Attr returns the value of the specified attribute, or (nil, nil) if the attribute is not found.
// It implements the starlark.HasAttrs interface.
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

// AttrNames returns a new slice containing the names of all the attributes of the SharedDict.
// It implements the starlark.HasAttrs interface.
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

// CompareSameType compares the SharedDict with another value of the same type.
// It implements the starlark.Comparable interface.
func (s *SharedDict) CompareSameType(op syntax.Token, y_ starlark.Value, depth int) (bool, error) {
	// if they are the same object, they are equal
	if s == y_ {
		switch op {
		case syntax.EQL:
			return true, nil
		case syntax.NEQ:
			return false, nil
		default:
			return false, fmt.Errorf("unsupported operator: %s", op)
		}
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
	} else if s.dict == nil && y.dict == nil {
		// both are nil, they are equal, aha! (nil == nil)
		return true, nil
	}

	// one is nil, the other is not, they are not equal
	switch op {
	case syntax.EQL:
		return false, nil
	case syntax.NEQ:
		return true, nil
	default:
		return false, fmt.Errorf("unsupported operator: %s", op)
	}
}

var (
	customSharedDictMethods = map[string]*starlark.Builtin{
		"len":       starlark.NewBuiltin("len", sharedDictLen),
		"perform":   starlark.NewBuiltin("perform", sharedDictPerform),
		"to_json":   starlark.NewBuiltin("to_json", sharedDictToJSON),
		"from_json": starlark.NewBuiltin("from_json", sharedDictFromJSON),
	}
)

// sharedDictLen returns the length of the underlying dictionary.
func sharedDictLen(_ *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if err := starlark.UnpackPositionalArgs(b.Name(), args, kwargs, 0); err != nil {
		return nil, err
	}
	l := b.Receiver().(*starlark.Dict).Len()
	return starlark.MakeInt(l), nil
}

// sharedDictPerform calls the given function with the underlying receiver dictionary, and returns the result.
// The function must be callable, like def perform(fn).
func sharedDictPerform(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
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

// sharedDictToJSON converts the underlying dictionary to a JSON string.
func sharedDictToJSON(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	// check the arguments: no arguments
	if err := starlark.UnpackPositionalArgs(b.Name(), args, kwargs, 0); err != nil {
		return nil, err
	}

	// get the receiver
	d := b.Receiver().(*starlark.Dict)

	// get the JSON encoder
	jm, ok := stdjson.Module.Members["encode"]
	if !ok {
		return nil, fmt.Errorf("json.encode not found")
	}
	enc := jm.(*starlark.Builtin)

	// convert to JSON
	return enc.CallInternal(thread, starlark.Tuple{d}, nil)
}

// sharedDictFromJSON converts a starlark.Value to a starlark.Dict, and wraps it with a SharedDict.
func sharedDictFromJSON(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	// check the arguments
	var s itn.StringOrBytes
	if err := starlark.UnpackArgs(b.Name(), args, kwargs, "x", &s); err != nil {
		return nil, err
	}

	// get the JSON decoder
	jm, ok := stdjson.Module.Members["decode"]
	if !ok {
		return nil, fmt.Errorf("json.decode not found")
	}
	dec := jm.(*starlark.Builtin)

	// convert from JSON
	v, err := dec.CallInternal(thread, starlark.Tuple{s.StarlarkString()}, nil)
	if err != nil {
		return nil, err
	}

	// convert to dict
	nd, ok := v.(*starlark.Dict)
	if !ok {
		return nil, fmt.Errorf("got %s, want dict", v.Type())
	}

	// merge the new dict into a shared dict
	od := b.Receiver().(*starlark.Dict)
	for _, r := range nd.Items() {
		if len(r) < 2 {
			continue
		}
		if e := od.SetKey(r[0], r[1]); e != nil {
			return nil, e
		}
	}

	// return new json dict
	return nd, nil
}
