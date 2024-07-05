package dataconv

import (
	"fmt"
	"sort"
	"sync"

	tps "github.com/1set/starlet/dataconv/types"
	itn "github.com/1set/starlet/internal"
	stdjson "go.starlark.net/lib/json"
	"go.starlark.net/starlark"
	"go.starlark.net/syntax"
)

// SharedDict represents a thread-safe dictionary that can be concurrently accessed and modified by multiple Starlark threads.
// This synchronization is crucial in concurrent environments where Starlark scripts are executed in parallel, ensuring data consistency and preventing race conditions.
//
// The internal state of a SharedDict includes a standard Starlark dictionary (`*starlark.Dict`), a mutex (`sync.RWMutex`) to manage concurrent access,
// and a boolean flag indicating whether the dictionary is frozen. A frozen SharedDict cannot be modified, aligning with Starlark's immutability rules for frozen values.
// Additionally, SharedDict supports custom naming through the 'name' field, allowing for more descriptive representations and debugging.
//
// Constructors:
// - NewSharedDict: Initializes a new SharedDict with default settings.
// - NewNamedSharedDict: Creates a new SharedDict with a specified name, providing clarity when multiple SharedDicts are used.
// - NewSharedDictFromDict: Generates a new SharedDict based on an existing Starlark dictionary. It attempts to clone the original dictionary to preserve immutability.
//
// Methods like Len, CloneDict, ToJSON, LoadJSON provide additional functionalities like determining the dictionary's length, cloning, JSON serialization, and deserialization, enhancing the utility of SharedDict in various use cases.
//
// SharedDict integrates tightly with Starlark's concurrency model, offering a robust solution for managing shared state across threads.
// By encapsulating thread safety mechanisms and providing a familiar dictionary interface, SharedDict facilitates the development of concurrent Starlark scripts with shared mutable state.
type SharedDict struct {
	_ itn.DoNotCompare
	sync.RWMutex
	dict   *starlark.Dict
	frozen bool
	name   string
}

const (
	defaultSharedDictSize = 8
	defaultSharedDictName = "shared_dict"
)

var (
	_ starlark.Value      = (*SharedDict)(nil)
	_ starlark.Comparable = (*SharedDict)(nil)
	_ starlark.Mapping    = (*SharedDict)(nil)
	_ starlark.HasAttrs   = (*SharedDict)(nil)
	_ starlark.HasSetKey  = (*SharedDict)(nil)
)

// NewSharedDict creates a new SharedDict instance.
func NewSharedDict() *SharedDict {
	return &SharedDict{
		dict: starlark.NewDict(defaultSharedDictSize),
	}
}

// NewNamedSharedDict creates a new SharedDict instance with the given name.
func NewNamedSharedDict(name string) *SharedDict {
	return &SharedDict{
		dict: starlark.NewDict(defaultSharedDictSize),
		name: name,
	}
}

// NewSharedDictFromDict creates a new SharedDict instance from the given starlark.Dict.
// It attempts to clone the dictionary, and returns the original dictionary if failed.
func NewSharedDictFromDict(d *starlark.Dict) *SharedDict {
	nd, err := cloneDict(d)
	if err != nil {
		nd = d
	}
	return &SharedDict{
		dict: nd,
	}
}

func (s *SharedDict) String() string {
	var v string
	if s != nil && s.dict != nil {
		v = s.dict.String()
	}
	return fmt.Sprintf("%s(%s)", s.getTypeName(), v)
}

// SetTypeName sets the type name of the SharedDict.
func (s *SharedDict) SetTypeName(name string) {
	s.name = name
}

// getTypeName returns the type name of the SharedDict.
func (s *SharedDict) getTypeName() string {
	if s.name == "" {
		return defaultSharedDictName
	}
	return s.name
}

// Type returns the type name of the SharedDict.
func (s *SharedDict) Type() string {
	return s.getTypeName()
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
	return 0, fmt.Errorf("unhashable type: %s", s.getTypeName())
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
		return fmt.Errorf("frozen %s", s.Type())
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
		return starlark.Call(thread, btl, args, kwargs)
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
func (s *SharedDict) CompareSameType(op syntax.Token, yv starlark.Value, depth int) (bool, error) {
	retEqualCheck := func(equal bool, op syntax.Token) (bool, error) {
		switch op {
		case syntax.EQL:
			return equal, nil
		case syntax.NEQ:
			return !equal, nil
		default:
			return false, fmt.Errorf("unsupported operator: %s", op)
		}
	}

	// if they are the same object, they are equal
	if s == yv {
		return retEqualCheck(true, op)
	}

	// scan the type
	y := yv.(*SharedDict)

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
		return retEqualCheck(true, op)
	}

	// one is nil, the other is not, they are not equal
	return retEqualCheck(false, op)
}

// Len returns the length of the underlying dictionary.
// Notice that this method is not a must for the starlark.Value interface, but it's useful for Go code.
func (s *SharedDict) Len() int {
	s.RLock()
	defer s.RUnlock()

	if s.dict != nil {
		return s.dict.Len()
	}
	return 0
}

// CloneDict creates a shallow copy of the underlying Starlark dictionary contained within the SharedDict instance.
// This method is particularly valuable when a snapshot of the current state of the dictionary is needed without affecting the original dictionary.
// It ensures that modifications to the returned dictionary do not impact the source SharedDict, providing a mechanism for safe, concurrent read operations.
func (s *SharedDict) CloneDict() (*starlark.Dict, error) {
	s.RLock()
	defer s.RUnlock()

	return cloneDict(s.dict)
}

// ToJSON serializes the SharedDict instance into a JSON string representation.
// This method facilitates the conversion of complex, nested data structures stored within a SharedDict into a universally recognizable format (JSON),
// making it easier to export or log the data contained within the SharedDict.
//
// It is important to note that the serialization process adheres to JSON's limitations, such as not supporting circular references. If the SharedDict contains
// circular references or types not supported by JSON (e.g., functions), `ToJSON` will return an error.
func (s *SharedDict) ToJSON() (string, error) {
	return EncodeStarlarkJSON(s.dict)
}

// LoadJSON updates the SharedDict instance with key-value pairs decoded from a given JSON string.
// This method provides a convenient way to populate or update the contents of a SharedDict with data received in JSON format,
// such as from a configuration file, a network request, or any external data source.
//
// The method attempts to merge the contents of the JSON string into the existing SharedDict. In cases where keys overlap,
// the values specified in the JSON string will overwrite those in the SharedDict.
//
// It's important to ensure that the JSON string represents a dictionary/object structure; otherwise, `LoadJSON` will return an error.
// Also, the SharedDict must not be frozen; attempting to modify a frozen SharedDict will result in an error.
func (s *SharedDict) LoadJSON(jsonStr string) error {
	// check the dict itself
	if s == nil {
		return fmt.Errorf("nil shared dict")
	}

	// json decode
	val, err := DecodeStarlarkJSON([]byte(jsonStr))
	if err != nil {
		return err
	}

	// convert to dict
	nd, ok := val.(*starlark.Dict)
	if !ok {
		return fmt.Errorf("got %s result, want dict", val.Type())
	}

	// lock the shared dict
	s.Lock()
	defer s.Unlock()

	// merge the new dict into the shared dict
	for _, r := range nd.Items() {
		if len(r) < 2 {
			continue
		}
		if e := s.dict.SetKey(r[0], r[1]); e != nil {
			return e
		}
	}
	return nil
}

var (
	customSharedDictMethods = map[string]*starlark.Builtin{
		"len":       starlark.NewBuiltin("len", sharedDictLen),
		"perform":   starlark.NewBuiltin("perform", sharedDictPerform),
		"to_dict":   starlark.NewBuiltin("to_dict", sharedDictToDict),
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
		return starlark.Call(thread, pr, starlark.Tuple{d}, nil)
	default:
		return nil, fmt.Errorf("%s: not callable type: %s", b.Name(), pr.Type())
	}
}

// sharedDictToDict returns the shadow-clone of underlying dictionary.
func sharedDictToDict(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if err := starlark.UnpackPositionalArgs(b.Name(), args, kwargs, 0); err != nil {
		return nil, err
	}
	// get the receiver
	od := b.Receiver().(*starlark.Dict)

	// clone the dictionary
	return cloneDict(od)
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
	return starlark.Call(thread, enc, starlark.Tuple{d}, nil)
}

// sharedDictFromJSON converts a starlark.Value to a starlark.Dict, and wraps it with a SharedDict.
func sharedDictFromJSON(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	// check the arguments
	var s tps.StringOrBytes
	if err := starlark.UnpackArgs(b.Name(), args, kwargs, "x", &s); err != nil {
		return nil, err
	}

	// json decode
	v, err := DecodeStarlarkJSON(s.GoBytes())
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

// cloneDict returns a shadow-clone of the given dictionary. It's safe to call it with a nil dictionary, it will return a new empty dictionary.
func cloneDict(od *starlark.Dict) (*starlark.Dict, error) {
	if od == nil {
		return starlark.NewDict(defaultSharedDictSize), nil
	}
	nd := starlark.NewDict(od.Len())
	for _, r := range od.Items() {
		if len(r) < 2 {
			continue
		}
		if e := nd.SetKey(r[0], r[1]); e != nil {
			return nil, e
		}
	}
	return nd, nil
}
