package dataconv

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"

	tps "github.com/1set/starlet/dataconv/types"
	itn "github.com/1set/starlet/internal"
	starlarktime "go.starlark.net/lib/time"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
	"go.starlark.net/syntax"
)

// SharedDict wraps a Starlark dictionary with locks for indexing, dictionary
// methods, Len, CloneDict, LoadJSON, Freeze and comparison.
// Returned keys/values and shallow clones can retain mutable nested aliases;
// the mapping's lock does not synchronize access through those aliases.
// String and SnapshotString capture entries before formatting, and ToJSON
// serializes native Starlark data under the mapping lock. Output methods do not
// wait for a busy dictionary: formatting returns shared_dict(<busy>), while
// ToJSON returns ErrSharedDictBusy. Type names can be changed concurrently.
// A frozen SharedDict rejects mapping updates and freezes its contained values.
type SharedDict struct {
	_ itn.DoNotCompare
	sync.RWMutex
	outputMu sync.Mutex
	dict     *starlark.Dict
	frozen   bool
	name     atomic.Pointer[string]
}

const (
	defaultSharedDictSize = 8
	defaultSharedDictName = "shared_dict"
	sharedDictBusyString  = "shared_dict(<busy>)"
)

// ErrSharedDictBusy reports output that could not safely acquire its locks.
// Retry outside callbacks or critical sections; a busy result contains no JSON.
var ErrSharedDictBusy = errors.New("shared dictionary is busy")

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
	s := NewSharedDict()
	s.SetTypeName(name)
	return s
}

// NewSharedDictFromDict creates a new SharedDict instance from the given starlark.Dict.
// It copies top-level entries, retaining nested value aliases; if copying fails,
// it uses the original dictionary. The caller must synchronize access to d.
func NewSharedDictFromDict(d *starlark.Dict) *SharedDict {
	nd, err := CloneDict(d)
	if err != nil {
		nd = d
	}
	return &SharedDict{
		dict: nd,
	}
}

func (s *SharedDict) String() string {
	return s.SnapshotString()
}

// SnapshotString formats a shallow snapshot of the mapping in insertion order.
// It releases mapping locks before invoking formatting callbacks. Output from
// each visited SharedDict is serialized without waiting for a busy lock; a busy
// dictionary is rendered as shared_dict(<busy>), including reentrant calls.
// Mutable native aliases still require their own ownership or synchronization.
// Cycle detection spans
// native lists/dictionaries/sets, SharedDict and structs with unique field names,
// following references via tuples. Other values retain their own String behavior.
// Set elements are captured before formatting callbacks run; the values remain shared.
func (s *SharedDict) SnapshotString() string {
	var f sharedDictFormatter
	defer f.release()
	f.write(s, nil)
	return f.String()
}

// Preserve the interpreter's container-cycle notation while using the captured
// root entries. Other values keep their own String behavior, without rehashing
// keys or invoking host callbacks under the mapping lock.
type sharedDictFormatter struct {
	strings.Builder
	snapshots []sharedDictSnapshot
}

type sharedDictSnapshot struct {
	shared *SharedDict
	dict   *starlark.Dict
	name   string
	items  []starlark.Tuple
}

func (f *sharedDictFormatter) release() {
	for i := len(f.snapshots) - 1; i >= 0; i-- {
		f.snapshots[i].shared.outputMu.Unlock()
	}
}

func (f *sharedDictFormatter) capture(shared *SharedDict) (*sharedDictSnapshot, bool) {
	for i := range f.snapshots {
		if f.snapshots[i].shared == shared {
			return &f.snapshots[i], true
		}
	}
	if !shared.outputMu.TryLock() {
		return nil, false
	}
	if !shared.TryLock() {
		shared.outputMu.Unlock()
		return nil, false
	}
	captured := false
	defer func() {
		shared.Unlock()
		if !captured {
			shared.outputMu.Unlock()
		}
	}()
	snapshot := sharedDictSnapshot{shared: shared, dict: shared.dict, name: shared.getTypeName()}
	if snapshot.dict != nil {
		snapshot.items = snapshot.dict.Items()
	}
	f.snapshots = append(f.snapshots, snapshot)
	captured = true
	return &f.snapshots[len(f.snapshots)-1], true
}

func (f *sharedDictFormatter) write(value starlark.Value, path []starlark.Value) {
	switch value := value.(type) {
	case nil:
		f.WriteString("<nil>")
	case *starlark.List:
		f.WriteByte('[')
		if sharedDictPathContains(path, value) {
			f.WriteString("...")
		} else {
			path = append(path, value)
			for i := 0; i < value.Len(); i++ {
				if i > 0 {
					f.WriteString(", ")
				}
				f.write(value.Index(i), path)
			}
		}
		f.WriteByte(']')
	case starlark.Tuple:
		f.WriteByte('(')
		for i, element := range value {
			if i > 0 {
				f.WriteString(", ")
			}
			f.write(element, path)
		}
		if len(value) == 1 {
			f.WriteByte(',')
		}
		f.WriteByte(')')
	case *SharedDict:
		if value == nil {
			f.WriteString(defaultSharedDictName + "()")
			return
		}
		snapshot, ok := f.capture(value)
		if !ok {
			f.WriteString(sharedDictBusyString)
			return
		}
		f.WriteString(snapshot.name)
		f.WriteByte('(')
		if snapshot.dict != nil {
			f.write(snapshot.dict, path)
		}
		f.WriteByte(')')
	case *starlark.Set:
		f.writeSet(value, path)
	case *starlarkstruct.Struct:
		f.writeStruct(value, path)
	case *starlark.Dict:
		f.WriteByte('{')
		if sharedDictPathContains(path, value) {
			f.WriteString("...")
		} else {
			var items []starlark.Tuple
			captured := false
			for _, snapshot := range f.snapshots {
				if snapshot.dict == value {
					items, captured = snapshot.items, true
					break
				}
			}
			if !captured {
				items = value.Items()
			}
			valuePath := append(path, value)
			keyPath := path[:len(path):len(path)]
			for i, item := range items {
				if i > 0 {
					f.WriteString(", ")
				}
				f.write(item[0], keyPath)
				f.WriteString(": ")
				f.write(item[1], valuePath)
			}
		}
		f.WriteByte('}')
	default:
		f.WriteString(value.String())
	}
}

func (f *sharedDictFormatter) writeSet(value *starlark.Set, path []starlark.Value) {
	f.WriteString("set([")
	if sharedDictPathContains(path, value) {
		f.WriteString("...")
	} else {
		var smallElements [8]starlark.Value
		elements := smallElements[:0]
		if size := value.Len(); size > len(smallElements) {
			elements = make([]starlark.Value, 0, size)
		}
		iter := value.Iterate()
		var element starlark.Value
		for iter.Next(&element) {
			elements = append(elements, element)
		}
		iter.Done()
		var smallPath [8]starlark.Value
		size := len(path) + 1
		nextPath := smallPath[:0]
		if size >= len(smallPath) {
			nextPath = make([]starlark.Value, size, size+1)
		} else {
			nextPath = smallPath[:size]
		}
		copy(nextPath, path)
		nextPath[size-1] = value
		for i, element := range elements {
			if i > 0 {
				f.WriteString(", ")
			}
			f.write(element, nextPath)
		}
	}
	f.WriteString("])")
}

func (f *sharedDictFormatter) writeStruct(value *starlarkstruct.Struct, path []starlark.Value) {
	names := value.AttrNames()
	for i := 1; i < len(names); i++ {
		if names[i] == names[i-1] {
			// Attr returns only the first entry of duplicate Go fields.
			f.WriteString(value.String())
			return
		}
	}
	if sharedDictPathContains(path, value) {
		f.WriteString("...")
		return
	}
	path = append(path, value)
	constructor := value.Constructor()
	if name, ok := constructor.(starlark.String); ok {
		f.WriteString(name.GoString())
	} else {
		if constructor == nil {
			f.WriteString(value.String())
			return
		}
		f.write(constructor, path)
	}
	f.WriteByte('(')
	for i, name := range names {
		if i > 0 {
			f.WriteString(", ")
		}
		f.WriteString(name)
		f.WriteString(" = ")
		field, _ := value.Attr(name)
		if field == nil {
			f.WriteString(value.String())
			return
		}
		f.write(field, path)
	}
	f.WriteByte(')')
}

func sharedDictPathContains(path []starlark.Value, value starlark.Value) bool {
	for _, ancestor := range path {
		if ancestor == value {
			return true
		}
	}
	return false
}

// SetTypeName sets the type name of the SharedDict.
func (s *SharedDict) SetTypeName(name string) {
	if name == "" {
		s.name.Store(nil)
		return
	}
	s.name.Store(&name)
}

// getTypeName returns the type name of the SharedDict.
func (s *SharedDict) getTypeName() string {
	name := s.name.Load()
	if name == nil {
		return defaultSharedDictName
	}
	return *name
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
	if name == "to_json" {
		return starlark.NewBuiltin(name, s.sharedDictToJSON), nil
	}
	btl, ok := customSharedDictMethods[name]
	if !ok {
		attr, err := sharedDictMethodPrototype.Attr(name)
		if attr == nil || err != nil {
			return attr, err
		}
		btl, ok = attr.(*starlark.Builtin)
		if !ok {
			return nil, fmt.Errorf("unsupported attribute: %s", name)
		}
	}
	// Metadata lookup does not touch the live mapping. Acquire its lock only
	// when calling the method, preserving the existing method-call contract.
	return starlark.NewBuiltin(name, func(thread *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		s.Lock()
		defer s.Unlock()
		s.ensureDict()
		return starlark.Call(thread, btl.BindReceiver(s.dict), args, kwargs)
	}), nil
}

// AttrNames returns a new slice containing the names of all the attributes of the SharedDict.
// It implements the starlark.HasAttrs interface.
func (s *SharedDict) AttrNames() []string {
	if s == nil {
		return nil
	}
	return append([]string(nil), sharedDictAttrNames...)
}

var sharedDictMethodPrototype = starlark.NewDict(0)

var sharedDictAttrNames = []string{
	"clear", "from_json", "get", "items", "keys", "len", "perform", "pop",
	"popitem", "setdefault", "to_dict", "to_json", "update", "values",
}

// The caller holds the mapping's write lock. A zero-value dictionary is empty;
// initialization must preserve a Freeze that happened before its first use.
func (s *SharedDict) ensureDict() {
	if s.dict == nil {
		s.dict = starlark.NewDict(defaultSharedDictSize)
		if s.frozen {
			s.dict.Freeze()
		}
	}
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

// CloneDict copies the mapping's entries under its read lock.
// Top-level insertions and deletions on the returned dictionary are independent;
// nested values remain shared and require their own ownership or synchronization.
func (s *SharedDict) CloneDict() (*starlark.Dict, error) {
	s.RLock()
	defer s.RUnlock()

	return CloneDict(s.dict)
}

// ToJSON serializes native Starlark data with the official JSON encoder.
// It returns ErrSharedDictBusy rather than waiting for another output operation
// or an already-held mapping lock. Supported JSON data includes native scalars,
// containers, structs/modules, duration and the interpreter's immutable range
// and string/bytes iterator sources. Dictionary keys must be strings.
// Other Go-backed values and custom JSON/attribute implementations are rejected
// before encoding, so no host code runs while the mapping lock is held.
// Mutable native aliases still require ownership or external synchronization.
// The Starlark to_json method uses the same rules. Errors return an empty string.
func (s *SharedDict) ToJSON() (string, error) {
	return s.toJSON(nil)
}

func (s *SharedDict) toJSON(thread *starlark.Thread) (string, error) {
	if s == nil {
		return "", fmt.Errorf("shared_dict.to_json: nil shared dictionary")
	}
	if !s.outputMu.TryLock() {
		return "", ErrSharedDictBusy
	}
	defer s.outputMu.Unlock()
	if !s.TryLock() {
		return "", ErrSharedDictBusy
	}
	defer s.Unlock()
	s.ensureDict()
	if err := sharedDictCheckJSON(s.dict); err != nil {
		return "", err
	}
	if thread == nil {
		return EncodeStarlarkJSON(s.dict)
	}
	value, err := starlark.Call(thread, starJSONEncode, starlark.Tuple{s.dict}, nil)
	if err != nil {
		return "", err
	}
	return string(value.(starlark.String)), nil
}

// Validate only the exact native types whose JSON traversal invokes no host
// callbacks. A work list and visited identities also terminate on Go-created
// tuple cycles; the encoder itself reports supported-container cycle errors.
func sharedDictCheckJSON(root starlark.Value) error {
	type tupleIdentity struct {
		first *starlark.Value
		len   int
	}
	seen := make(map[interface{}]bool)
	visit := func(key interface{}) bool {
		if seen[key] {
			return false
		}
		seen[key] = true
		return true
	}
	invalid := func(value starlark.Value) error {
		return fmt.Errorf("shared_dict.to_json: unsupported value %T; want native Starlark JSON data", value)
	}
	pending := []starlark.Value{root}
	for len(pending) > 0 {
		value := pending[len(pending)-1]
		pending = pending[:len(pending)-1]
		switch value := value.(type) {
		case starlark.NoneType, starlark.Bool, starlark.Int, starlark.Float, starlark.String,
			starlarktime.Duration, starlarktime.Time:
			continue
		case *starlark.Dict:
			if value == nil {
				return invalid(value)
			}
			if !visit(value) {
				continue
			}
			for _, item := range value.Items() {
				if _, ok := item[0].(starlark.String); !ok {
					return fmt.Errorf("shared_dict.to_json: dictionary key must be string (got %T)", item[0])
				}
				pending = append(pending, item[1])
			}
		case *starlark.List:
			if value == nil {
				return invalid(value)
			}
			if !visit(value) {
				continue
			}
			for i := 0; i < value.Len(); i++ {
				pending = append(pending, value.Index(i))
			}
		case starlark.Tuple:
			if len(value) == 0 || !visit(tupleIdentity{&value[0], len(value)}) {
				continue
			}
			pending = append(pending, value...)
		case *starlark.Set:
			if value == nil {
				return invalid(value)
			}
			if !visit(value) {
				continue
			}
			func() {
				iter := value.Iterate()
				defer iter.Done()
				var element starlark.Value
				for iter.Next(&element) {
					pending = append(pending, element)
				}
			}()
		case *starlarkstruct.Struct:
			if value == nil {
				return invalid(value)
			}
			if !visit(value) {
				continue
			}
			for _, name := range value.AttrNames() {
				field, err := value.Attr(name)
				if err != nil || field == nil {
					return fmt.Errorf("shared_dict.to_json: invalid native struct field")
				}
				pending = append(pending, field)
			}
		case *starlarkstruct.Module:
			if value == nil {
				return invalid(value)
			}
			if !visit(value) {
				continue
			}
			for _, name := range value.AttrNames() {
				field, err := value.Attr(name)
				if err != nil || field == nil {
					return fmt.Errorf("shared_dict.to_json: invalid native module member")
				}
				pending = append(pending, field)
			}
		default:
			if sharedDictSafeJSONIterable(value) {
				continue
			}
			return invalid(value)
		}
	}
	return nil
}

// These four immutable, private iterator sources are closed over scalar data in
// the pinned interpreter. Do not trust arbitrary Iterable implementations or
// pre-traverse a potentially large range merely to validate its element types.
func sharedDictSafeJSONIterable(value starlark.Value) bool {
	typ := reflect.TypeOf(value)
	if typ == nil || typ.Kind() != reflect.Struct || typ.PkgPath() != "go.starlark.net/starlark" {
		return false
	}
	switch typ.Name() {
	case "rangeValue", "stringElems", "stringCodepoints", "bytesIterable":
		return true
	}
	return false
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

// sharedDictToDict returns a shallow copy of the underlying dictionary.
func sharedDictToDict(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if err := starlark.UnpackPositionalArgs(b.Name(), args, kwargs, 0); err != nil {
		return nil, err
	}
	// get the receiver
	od := b.Receiver().(*starlark.Dict)

	// clone the dictionary
	return CloneDict(od)
}

// sharedDictToJSON converts the underlying dictionary to a JSON string.
func (s *SharedDict) sharedDictToJSON(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	// check the arguments: no arguments
	if err := starlark.UnpackPositionalArgs(b.Name(), args, kwargs, 0); err != nil {
		return nil, err
	}

	text, err := s.toJSON(thread)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", b.Name(), err)
	}
	return starlark.String(text), nil
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

// CloneDict copies the dictionary's top-level entries, retaining nested aliases.
// A nil dictionary produces an empty dictionary. It does not synchronize access
// to od; callers must own or lock it while copying.
func CloneDict(od *starlark.Dict) (*starlark.Dict, error) {
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
