package dataconv

// SharedDict tests cover constructors/names, script methods, frozen/nil values,
// concurrent access, nonblocking output, snapshots, and native JSON conversion.

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"os/exec"
	"reflect"
	"runtime/debug"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	itn "github.com/1set/starlet/internal"
	starlarktime "go.starlark.net/lib/time"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
	"go.starlark.net/syntax"
)

func TestSharedDict_SnapshotString(t *testing.T) {
	globals, err := starlark.ExecFile(&starlark.Thread{}, "values.star", "def identity(x):\n    return x\nr = range(3)\ns = 'abc'.elems()\nu = '雪'.codepoints()\nb = b'abc'.elems()\n", nil)
	if err != nil {
		t.Fatal(err)
	}
	set := starlark.NewSet(1)
	if err := set.Insert(starlark.Tuple{starlark.String("item"), starlark.MakeInt(1)}); err != nil {
		t.Fatal(err)
	}
	method, err := starlark.String("hello").Attr("upper")
	if err != nil {
		t.Fatal(err)
	}
	values := []starlark.Value{
		starlark.None, starlark.True, starlark.False, starlark.MakeInt(-3),
		starlark.MakeUint64(^uint64(0)), starlark.Float(1.25), starlark.Float(math.NaN()),
		starlark.Float(math.Inf(1)), starlark.String("quotes\"\n雪"), starlark.Bytes("\x00\xff"),
		starlark.Tuple{}, starlark.Tuple{starlark.MakeInt(1)},
		starlark.Tuple{starlark.NewList([]starlark.Value{starlark.True}), starlark.String("two")},
		starlark.NewList(nil), starlark.NewList([]starlark.Value{starlark.True, starlark.MakeInt(2)}),
		starlark.NewList([]starlark.Value{nil}), starlark.NewDict(0), set,
		globals["identity"], starlark.Universe["len"], method,
		globals["r"], globals["s"], globals["u"], globals["b"],
		starlarkstruct.FromStringDict(starlark.String("record"), starlark.StringDict{"value": starlark.True}),
		&starlarkstruct.Module{Name: "module", Members: starlark.StringDict{"value": starlark.True}},
		starlarktime.Time(time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC)), starlarktime.Duration(time.Second),
	}
	for _, value := range values {
		t.Run(value.Type()+"/"+value.String(), func(t *testing.T) {
			sd := NewNamedSharedDict("named")
			if err := sd.SetKey(starlark.String("value"), value); err != nil {
				t.Fatal(err)
			}
			want := "named(" + sd.dict.String() + ")"
			if got := sd.SnapshotString(); got != want {
				t.Fatalf("String() = %q; want upstream representation %q", got, want)
			}
		})
	}
	for _, key := range []starlark.Value{
		starlark.None, starlark.True, starlark.MakeUint64(^uint64(0)),
		starlark.Bytes("\x00\xff"), starlark.Tuple{}, starlark.Tuple{starlark.String("one")},
		starlark.Tuple{starlark.Tuple{starlark.MakeInt(1), starlark.True}, starlark.String("two")},
	} {
		t.Run("key/"+key.Type()+"/"+key.String(), func(t *testing.T) {
			sd := NewSharedDict()
			if err := sd.SetKey(key, starlark.String("value")); err != nil {
				t.Fatal(err)
			}
			want := "shared_dict(" + sd.dict.String() + ")"
			if got := sd.SnapshotString(); got != want {
				t.Fatalf("key String() = %q; want %q", got, want)
			}
		})
	}
	t.Run("root_cycles", func(t *testing.T) {
		sd := NewSharedDict()
		list := starlark.NewList(nil)
		other := starlark.NewDict(2)
		if err := other.SetKey(starlark.String("self"), other); err != nil {
			t.Fatal(err)
		}
		if err := other.SetKey(starlark.String("root"), sd.dict); err != nil {
			t.Fatal(err)
		}
		if err := list.Append(starlark.Tuple{sd.dict, list}); err != nil {
			t.Fatal(err)
		}
		for _, item := range []starlark.Tuple{
			{starlark.String("self"), sd.dict},
			{starlark.String("via_list_tuple"), list},
			{starlark.String("repeated_list"), list},
			{starlark.String("nested_dictionary"), other},
		} {
			if err := sd.SetKey(item[0], item[1]); err != nil {
				t.Fatal(err)
			}
		}
		want := "shared_dict(" + sd.dict.String() + ")"
		if got := sd.SnapshotString(); got != want {
			t.Fatalf("cyclic String() = %q; want %q", got, want)
		}
	})
	t.Run("insertion_order", func(t *testing.T) {
		sd := NewSharedDict()
		for _, key := range []string{"z", "a", "m"} {
			if err := sd.SetKey(starlark.String(key), starlark.MakeInt(1)); err != nil {
				t.Fatal(err)
			}
		}
		if got, want := sd.SnapshotString(), `shared_dict({"z": 1, "a": 1, "m": 1})`; got != want {
			t.Fatalf("String() = %q; want %q", got, want)
		}
	})
	t.Run("nil_underlying_dictionary", func(t *testing.T) {
		if got := (&SharedDict{}).String(); got != "shared_dict()" {
			t.Fatalf("String() = %q; want shared_dict()", got)
		}
		if got := (&SharedDict{}).SnapshotString(); got != "shared_dict()" {
			t.Fatalf("SnapshotString() = %q; want shared_dict()", got)
		}
	})
	t.Run("empty_root", func(t *testing.T) {
		zero := new(SharedDict)
		zero.SetTypeName("zero")
		for _, test := range []struct {
			dict *SharedDict
			want string
		}{{NewSharedDict(), "shared_dict({})"}, {NewNamedSharedDict("named"), "named({})"}, {zero, "zero()"}} {
			if got := test.dict.SnapshotString(); got != test.want {
				t.Fatalf("SnapshotString() = %q; want %q", got, test.want)
			}
		}
	})
}

func TestSharedDict_SnapshotStringLeavesNestedValuesMutable(t *testing.T) {
	sd := NewSharedDict()
	list, dict := starlark.NewList(nil), starlark.NewDict(0)
	for _, item := range []starlark.Tuple{{starlark.String("list"), list}, {starlark.String("dict"), dict}} {
		if err := sd.SetKey(item[0], item[1]); err != nil {
			t.Fatal(err)
		}
	}
	_ = sd.SnapshotString()
	if err := list.Append(starlark.True); err != nil {
		t.Fatalf("String froze the nested list: %v", err)
	}
	if err := dict.SetKey(starlark.String("new"), starlark.True); err != nil {
		t.Fatalf("String froze the nested dictionary: %v", err)
	}
}

func TestSharedDict_SnapshotStringSnapshot(t *testing.T) {
	sd := NewSharedDict()
	callback := &sharedDictFormatValue{stringFunc: func() string {
		// A formatter must be able to use the dictionary without its lock held.
		if !sd.TryLock() {
			t.Error("String callback ran while the mapping lock was held")
			return "callback"
		}
		sd.Unlock()
		if sd.Len() != 1 {
			t.Errorf("callback Len() = %d; want 1", sd.Len())
		}
		if _, found, err := sd.Get(starlark.String("first")); err != nil || !found {
			t.Errorf("callback Get() found=%v, err=%v", found, err)
		}
		if err := sd.SetKey(starlark.String("later"), starlark.MakeInt(2)); err != nil {
			t.Error(err)
		}
		return "callback"
	}}
	if err := sd.SetKey(starlark.String("first"), callback); err != nil {
		t.Fatal(err)
	}
	if got, want := sd.SnapshotString(), `shared_dict({"first": callback})`; got != want {
		t.Fatalf("String() = %q; want captured entries %q", got, want)
	}
	if sd.Len() != 2 {
		t.Fatal("String callback did not update the live dictionary")
	}
}

func TestSharedDict_SnapshotStringTypeCallback(t *testing.T) {
	sd := NewSharedDict()
	calls := 0
	receiver := &sharedDictFormatValue{stringFunc: func() string { return "receiver" }, typeFunc: func() string {
		calls++
		if !sd.TryLock() {
			t.Error("Type callback ran while the mapping lock was held")
			return "host"
		}
		sd.Unlock()
		if sd.Len() != 1 {
			t.Errorf("Type callback Len() = %d; want 1", sd.Len())
		}
		if _, found, err := sd.Get(starlark.String("first")); err != nil || !found {
			t.Errorf("Type callback Get found=%v err=%v", found, err)
		}
		if err := sd.SetKey(starlark.String("later"), starlark.MakeInt(2)); err != nil {
			t.Error(err)
		}
		return "host"
	}}
	method := starlark.NewBuiltin("method", func(*starlark.Thread, *starlark.Builtin, starlark.Tuple, []starlark.Tuple) (starlark.Value, error) {
		return starlark.None, nil
	}).BindReceiver(receiver)
	if err := sd.SetKey(starlark.String("first"), method); err != nil {
		t.Fatal(err)
	}
	if got, want := sd.SnapshotString(), `shared_dict({"first": <built-in method method of host value>})`; got != want {
		t.Fatalf("SnapshotString() = %q; want %q", got, want)
	}
	if calls != 1 || sd.Len() != 2 {
		t.Fatalf("Type callback count=%d, mapping length=%d; want 1 and 2", calls, sd.Len())
	}
}

func TestSharedDict_SnapshotStringDoesNotRehash(t *testing.T) {
	var hashes, comparisons int
	hash := func() (uint32, error) {
		hashes++
		return 1, nil
	}
	key := &sharedDictFormatValue{stringFunc: func() string { return "host_key" }, hashFunc: hash, comparisons: &comparisons}
	other := &sharedDictFormatValue{stringFunc: func() string { return "other_key" }, hashFunc: hash, comparisons: &comparisons}
	sd := NewSharedDict()
	if err := sd.SetKey(key, starlark.MakeInt(1)); err != nil {
		t.Fatal(err)
	}
	if err := sd.SetKey(other, starlark.MakeInt(2)); err != nil {
		t.Fatal(err)
	}
	beforeHashes, beforeComparisons := hashes, comparisons
	if got, want := sd.SnapshotString(), "shared_dict({host_key: 1, other_key: 2})"; got != want {
		t.Fatalf("String() = %q; want %q", got, want)
	}
	if hashes != beforeHashes || comparisons != beforeComparisons {
		t.Fatalf("String invoked key Hash/Equal: hashes %d -> %d, comparisons %d -> %d", beforeHashes, hashes, beforeComparisons, comparisons)
	}
}

func TestSharedDict_SnapshotStringCallbackPanic(t *testing.T) {
	sd := NewSharedDict()
	if err := sd.SetKey(starlark.String("value"), &sharedDictFormatValue{stringFunc: func() string { panic("format panic") }}); err != nil {
		t.Fatal(err)
	}
	func() {
		defer func() {
			if got := recover(); got != "format panic" {
				t.Errorf("String panic = %v; want format panic", got)
			}
		}()
		_ = sd.SnapshotString()
	}()
	if !sd.TryLock() {
		t.Fatal("mapping lock remained held after a formatting panic")
	}
	sd.Unlock()
	if err := sd.SetKey(starlark.String("value"), starlark.None); err != nil {
		t.Fatal(err)
	}
	if sd.String() == sharedDictBusyString || sd.SnapshotString() == sharedDictBusyString {
		t.Fatal("output lock leaked after formatting panic")
	}
	if _, err := sd.ToJSON(); err != nil {
		t.Fatalf("ToJSON after formatting panic: %v", err)
	}
}

func TestSharedDict_SnapshotStringConcurrent(t *testing.T) {
	sd := NewSharedDict()
	key := starlark.String("counter")
	if err := sd.SetKey(key, starlark.MakeInt(0)); err != nil {
		t.Fatal(err)
	}
	start := make(chan struct{})
	var workers sync.WaitGroup
	workers.Add(2)
	go func() {
		defer workers.Done()
		<-start
		for i := 0; i < 1000; i++ {
			if err := sd.SetKey(key, starlark.MakeInt(i)); err != nil {
				t.Error(err)
				return
			}
		}
	}()
	go func() {
		defer workers.Done()
		<-start
		for i := 0; i < 1000; i++ {
			got := sd.SnapshotString()
			if got != sharedDictBusyString && (!strings.HasPrefix(got, `shared_dict({"counter": `) || !strings.HasSuffix(got, "})")) {
				t.Errorf("malformed concurrent String: %q", got)
				return
			}
		}
	}()
	close(start)
	workers.Wait()
}

func TestSharedDict_OutputLocks(t *testing.T) {
	sd := NewNamedSharedDict("named")
	if err := sd.SetKey(starlark.String("n"), starlark.MakeInt(1)); err != nil {
		t.Fatal(err)
	}
	encode, err := sd.Attr("to_json")
	if err != nil {
		t.Fatal(err)
	}
	for _, lock := range []struct {
		name   string
		lock   func()
		unlock func()
	}{{"write", sd.Lock, sd.Unlock}, {"read", sd.RLock, sd.RUnlock}, {"output", sd.outputMu.Lock, sd.outputMu.Unlock}} {
		t.Run(lock.name, func(t *testing.T) {
			lock.lock()
			defer lock.unlock()
			if got := sd.String(); got != sharedDictBusyString {
				t.Errorf("String while busy = %q", got)
			}
			if got := sd.SnapshotString(); got != sharedDictBusyString {
				t.Errorf("SnapshotString while busy = %q", got)
			}
			if text, err := sd.ToJSON(); text != "" || !errors.Is(err, ErrSharedDictBusy) {
				t.Errorf("ToJSON while busy = (%q, %v)", text, err)
			}
			if _, err := starlark.Call(&starlark.Thread{}, encode, nil, nil); !errors.Is(err, ErrSharedDictBusy) {
				t.Errorf("script to_json while busy = %v", err)
			}
			if lock.name == "write" {
				if method, err := sd.Attr("to_json"); method == nil || err != nil {
					t.Errorf("to_json lookup while locked = (%v,%v)", method, err)
				}
				if names := sd.AttrNames(); !reflect.DeepEqual(names, sharedDictAttrNames) {
					t.Errorf("attribute names while locked = %v", names)
				}
			}
		})
	}
	if got := sd.String(); got != `named({"n": 1})` {
		t.Errorf("String after busy calls = %q", got)
	}
	if text, err := sd.ToJSON(); text != `{"n":1}` || err != nil {
		t.Errorf("ToJSON after busy calls = (%q, %v)", text, err)
	}
	var nilDict *SharedDict
	if names := nilDict.AttrNames(); names != nil {
		t.Errorf("nil AttrNames = %v", names)
	}
	if got := nilDict.String(); got != "shared_dict()" {
		t.Errorf("nil String = %q", got)
	}
	if text, err := nilDict.ToJSON(); text != "" || err == nil {
		t.Errorf("nil ToJSON = (%q, %v)", text, err)
	}
	if text, err := new(SharedDict).ToJSON(); text != "{}" || err != nil {
		t.Errorf("zero ToJSON = (%q, %v)", text, err)
	}
}

func TestSharedDict_OutputConcurrent(t *testing.T) {
	for _, withWriter := range []bool{false, true} {
		t.Run(fmt.Sprintf("writer=%v", withWriter), func(t *testing.T) {
			sd := NewSharedDict()
			if err := sd.SetKey(starlark.String("counter"), starlark.MakeInt(0)); err != nil {
				t.Fatal(err)
			}
			set := starlark.NewSet(2)
			for _, n := range []int{1, 2} {
				if err := set.Insert(starlark.MakeInt(n)); err != nil {
					t.Fatal(err)
				}
			}
			if err := sd.SetKey(starlark.String("list"), starlark.NewList([]starlark.Value{starlark.MakeInt(1)})); err != nil {
				t.Fatal(err)
			}
			if err := sd.SetKey(starlark.String("set"), set); err != nil {
				t.Fatal(err)
			}
			start := make(chan struct{})
			var workers sync.WaitGroup
			for _, format := range []func() string{sd.String, sd.SnapshotString} {
				format := format
				workers.Add(1)
				go func() {
					defer workers.Done()
					<-start
					for i := 0; i < 128; i++ {
						got := format()
						if got != sharedDictBusyString && (!strings.HasPrefix(got, `shared_dict({"counter": `) || !strings.HasSuffix(got, `, "list": [1], "set": set([1, 2])})`)) {
							t.Errorf("concurrent format = %q", got)
							return
						}
					}
				}()
			}
			workers.Add(1)
			go func() {
				defer workers.Done()
				<-start
				for i := 0; i < 128; i++ {
					text, err := sd.ToJSON()
					if errors.Is(err, ErrSharedDictBusy) && text == "" {
						continue
					}
					var got struct {
						Counter int   `json:"counter"`
						List    []int `json:"list"`
						Set     []int `json:"set"`
					}
					if err != nil || json.Unmarshal([]byte(text), &got) != nil || got.Counter < 0 || got.Counter >= 128 || !reflect.DeepEqual(got.List, []int{1}) || !reflect.DeepEqual(got.Set, []int{1, 2}) {
						t.Errorf("concurrent ToJSON = (%q, %v)", text, err)
						return
					}
				}
			}()
			if withWriter {
				workers.Add(1)
				go func() {
					defer workers.Done()
					<-start
					for i := 0; i < 128; i++ {
						if err := sd.SetKey(starlark.String("counter"), starlark.MakeInt(i)); err != nil {
							t.Error(err)
							return
						}
					}
				}()
			}
			close(start)
			workers.Wait()
			counter := 0
			if withWriter {
				counter = 127
			}
			wantText := fmt.Sprintf(`shared_dict({"counter": %d, "list": [1], "set": set([1, 2])})`, counter)
			if got := sd.String(); got != wantText {
				t.Errorf("uncontended String = %q, want %q", got, wantText)
			}
			wantJSON := fmt.Sprintf(`{"counter":%d,"list":[1],"set":[1,2]}`, counter)
			if got, err := sd.ToJSON(); got != wantJSON || err != nil {
				t.Errorf("uncontended ToJSON = (%q,%v), want %q", got, err, wantJSON)
			}
		})
	}
}

func TestSharedDict_ConcurrentTypeName(t *testing.T) {
	sd := NewSharedDict()
	start := make(chan struct{})
	var workers sync.WaitGroup
	workers.Add(2)
	go func() {
		defer workers.Done()
		<-start
		for i := 0; i < 128; i++ {
			sd.SetTypeName([]string{"left", "right", ""}[i%3])
		}
	}()
	go func() {
		defer workers.Done()
		<-start
		for i := 0; i < 128; i++ {
			if name := sd.Type(); name != "left" && name != "right" && name != "shared_dict" {
				t.Errorf("concurrent Type = %q", name)
			}
			if text := sd.String(); text != "left({})" && text != "right({})" && text != "shared_dict({})" && text != sharedDictBusyString {
				t.Errorf("concurrent named String = %q", text)
			}
		}
	}()
	close(start)
	workers.Wait()
}

func TestSharedDict_ConcurrentNestedOutput(t *testing.T) {
	a, b := NewNamedSharedDict("a"), NewNamedSharedDict("b")
	for _, pair := range []struct{ owner, child *SharedDict }{{a, b}, {b, a}} {
		if err := pair.owner.SetKey(starlark.String("items"), starlark.NewList([]starlark.Value{pair.child})); err != nil {
			t.Fatal(err)
		}
	}
	start := make(chan struct{})
	var workers sync.WaitGroup
	for _, dict := range []*SharedDict{a, b} {
		dict := dict
		workers.Add(1)
		go func() {
			defer workers.Done()
			<-start
			for i := 0; i < 128; i++ {
				if text := dict.String(); text == "" {
					t.Error("empty opposite-root output")
					return
				}
			}
		}()
	}
	workers.Add(1)
	go func() {
		defer workers.Done()
		<-start
		for i := 0; i < 128; i++ {
			if err := b.SetKey(starlark.String("counter"), starlark.MakeInt(i)); err != nil {
				t.Error(err)
				return
			}
		}
	}()
	close(start)
	workers.Wait()
	if got, found, err := b.Get(starlark.String("counter")); err != nil || !found || got != starlark.MakeInt(127) {
		t.Fatalf("child writer result = (%v,%v,%v)", got, found, err)
	}
	if got := a.String(); got != `a({"items": [b({"items": [a({...})], "counter": 127})]})` {
		t.Errorf("cycle after concurrent output = %q", got)
	}
}

func TestSharedDict_NestedOutputLocks(t *testing.T) {
	root, child := NewNamedSharedDict("root"), NewNamedSharedDict("child")
	if err := child.SetKey(starlark.String("n"), starlark.MakeInt(1)); err != nil {
		t.Fatal(err)
	}
	if err := root.SetKey(starlark.String("items"), starlark.NewList([]starlark.Value{child, child})); err != nil {
		t.Fatal(err)
	}
	child.Lock()
	got := root.String()
	child.Unlock()
	if got != `root({"items": [shared_dict(<busy>), shared_dict(<busy>)]})` {
		t.Errorf("locked child = %q", got)
	}
	panicValue := &sharedDictFormatValue{stringFunc: func() string { panic("format panic") }}
	if err := child.SetKey(starlark.String("panic"), panicValue); err != nil {
		t.Fatal(err)
	}
	func() {
		defer func() {
			if recover() == nil {
				t.Error("expected nested formatting panic")
			}
		}()
		root.SnapshotString()
	}()
	if err := child.SetKey(starlark.String("panic"), starlark.None); err != nil {
		t.Fatal(err)
	}
	for _, dict := range []*SharedDict{root, child} {
		if got := dict.String(); got == sharedDictBusyString || strings.Contains(got, "<busy>") {
			t.Errorf("output lock leaked after panic: %q", got)
		}
	}
}

type sharedDictRejectedJSONValue struct{ calls *int }

func (v *sharedDictRejectedJSONValue) called()                      { (*v.calls)++; panic("JSON invoked host code") }
func (v *sharedDictRejectedJSONValue) String() string               { v.called(); return "host" }
func (v *sharedDictRejectedJSONValue) Type() string                 { v.called(); return "host" }
func (v *sharedDictRejectedJSONValue) Freeze()                      { v.called() }
func (*sharedDictRejectedJSONValue) Truth() starlark.Bool           { return true }
func (*sharedDictRejectedJSONValue) Hash() (uint32, error)          { return 1, nil }
func (v *sharedDictRejectedJSONValue) MarshalJSON() ([]byte, error) { v.called(); return nil, nil }
func (v *sharedDictRejectedJSONValue) AttrNames() []string          { v.called(); return nil }
func (v *sharedDictRejectedJSONValue) Attr(string) (starlark.Value, error) {
	v.called()
	return nil, nil
}

func TestSharedDict_JSONMethodThread(t *testing.T) {
	sd := NewSharedDict()
	method, err := sd.Attr("to_json")
	if err != nil {
		t.Fatal(err)
	}
	method.Freeze()
	if err := sd.SetKey(starlark.String("n"), starlark.MakeInt(1)); err != nil {
		t.Fatalf("freezing method froze receiver: %v", err)
	}
	thread := &starlark.Thread{Name: "script-caller"}
	marker := new(int)
	thread.SetLocal("json-marker", marker)
	old := starJSONEncode
	defer func() { starJSONEncode = old }()
	called := false
	starJSONEncode = starlark.NewBuiltin("encode", func(got *starlark.Thread, _ *starlark.Builtin, _ starlark.Tuple, _ []starlark.Tuple) (starlark.Value, error) {
		called = got == thread && got.Local("json-marker") == marker
		return starlark.String(`{"n":1}`), nil
	})
	if _, err := starlark.Call(thread, method, nil, nil); err != nil || !called {
		t.Fatalf("JSON lost caller thread: called=%v, err=%v", called, err)
	}
	starJSONEncode = old
	if err := sd.SetKey(starlark.String("bad"), starlark.Float(math.NaN())); err != nil {
		t.Fatal(err)
	}
	_, err = starlark.ExecFile(thread, "json-source.star", "def encode_here():\n    return sd.to_json()\nencode_here()\n", starlark.StringDict{"sd": sd})
	if eval, ok := err.(*starlark.EvalError); !ok || !strings.Contains(eval.Backtrace(), "json-source.star") || !strings.Contains(eval.Backtrace(), "encode_here") {
		t.Fatalf("JSON error source lost: %v", err)
	}
}

func TestSharedDict_AttributeNames(t *testing.T) {
	sd := new(SharedDict)
	want := starlark.NewDict(0).AttrNames()
	for name := range customSharedDictMethods {
		want = append(want, name)
	}
	want = append(want, "to_json")
	sort.Strings(want)
	if got := sd.AttrNames(); !reflect.DeepEqual(got, want) {
		t.Fatalf("zero AttrNames = %v, want %v", got, want)
	}
	names := sd.AttrNames()
	names[0] = "changed"
	if reflect.DeepEqual(names, sd.AttrNames()) {
		t.Fatal("AttrNames returned shared mutable list")
	}
	for _, frozen := range []bool{false, true} {
		sd := new(SharedDict)
		if frozen {
			sd.Freeze()
		}
		method, err := sd.Attr("to_json")
		if err != nil {
			t.Fatal(err)
		}
		if got, err := starlark.Call(&starlark.Thread{}, method, nil, nil); err != nil || got != starlark.String("{}") {
			t.Fatalf("zero JSON = (%v,%v)", got, err)
		}
		update, err := sd.Attr("update")
		if err != nil {
			t.Fatal(err)
		}
		arg := starlark.NewDict(1)
		if err := arg.SetKey(starlark.String("n"), starlark.MakeInt(1)); err != nil {
			t.Fatal(err)
		}
		_, err = starlark.Call(&starlark.Thread{}, update, starlark.Tuple{arg}, nil)
		if (err != nil) != frozen {
			t.Fatalf("zero initialization lost frozen=%v: %v", frozen, err)
		}
	}
}

func TestSharedDict_ToJSONAdmission(t *testing.T) {
	calls := 0
	host := &sharedDictRejectedJSONValue{calls: &calls}
	var list *starlark.List
	var dict *starlark.Dict
	var set *starlark.Set
	var record *starlarkstruct.Struct
	var module *starlarkstruct.Module
	longTuple := starlark.Tuple{starlark.MakeInt(1), host}
	badKey := starlark.NewDict(1)
	if err := badKey.SetKey(host, starlark.MakeInt(1)); err != nil {
		t.Fatal(err)
	}
	for _, value := range []starlark.Value{host, list, dict, set, record, module, nil, badKey, starlarkstruct.FromStringDict(starlarkstruct.Default, starlark.StringDict{"nil": nil}), &starlarkstruct.Module{Name: "nil_member", Members: starlark.StringDict{"nil": nil}}, starlarkstruct.FromStringDict(starlarkstruct.Default, starlark.StringDict{"bad": host}), &starlarkstruct.Module{Name: "host_member", Members: starlark.StringDict{"bad": host}}, starlark.Bytes("x"), starlark.Universe["len"], starlark.NewList([]starlark.Value{longTuple, longTuple[:1]})} {
		sd := NewSharedDict()
		if err := sd.SetKey(starlark.String("value"), value); err != nil {
			t.Fatal(err)
		}
		if text, err := sd.ToJSON(); text != "" || err == nil || errors.Is(err, ErrSharedDictBusy) {
			t.Errorf("unsupported %T produced (%q, %v)", value, text, err)
		}
		if calls != 0 {
			t.Fatal("JSON invoked a host callback")
		}
		if err := sd.SetKey(starlark.String("value"), starlark.MakeInt(1)); err != nil {
			t.Fatal(err)
		}
		if text, err := sd.ToJSON(); text != `{"value":1}` || err != nil {
			t.Errorf("JSON locks leaked after rejection: (%q, %v)", text, err)
		}
	}
	globals, err := starlark.ExecFile(&starlark.Thread{}, "json-values.star", "values = [range(3), 'abc'.elems(), 'abc'.elem_ords(), '雪'.codepoints(), '雪'.codepoint_ords(), b'abc'.elems()]", nil)
	if err != nil {
		t.Fatal(err)
	}
	aliasSet := starlark.NewSet(1)
	if err := aliasSet.Insert(starlark.MakeInt(1)); err != nil {
		t.Fatal(err)
	}
	aliasList := starlark.NewList([]starlark.Value{starlark.MakeInt(2)})
	aliasRecord := starlarkstruct.FromStringDict(starlarkstruct.Default, starlark.StringDict{"n": starlark.MakeInt(1)})
	aliasModule := &starlarkstruct.Module{Name: "alias", Members: starlark.StringDict{"n": starlark.MakeInt(1)}}
	aliases := starlark.Tuple{aliasList, aliasList, aliasSet, aliasSet, aliasRecord, aliasRecord, aliasModule, aliasModule}
	values := []starlark.Value{aliases, globals["values"], starlarktime.Duration(time.Second), starlarktime.Time(time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC)), &starlarkstruct.Struct{}, starlarkstruct.FromStringDict(host, starlark.StringDict{"n": starlark.MakeInt(1)}), &starlarkstruct.Module{Name: "module", Members: starlark.StringDict{"n": starlark.MakeInt(1)}}}
	for _, value := range values {
		sd := NewSharedDict()
		if err := sd.SetKey(starlark.String("value"), value); err != nil {
			t.Fatal(err)
		}
		want, wantErr := EncodeStarlarkJSON(sd.dict)
		got, gotErr := sd.ToJSON()
		if got != want || fmt.Sprint(gotErr) != fmt.Sprint(wantErr) {
			t.Errorf("native %T: got (%q,%v), want (%q,%v)", value, got, gotErr, want, wantErr)
		}
	}
	if calls != 0 {
		t.Fatal("native struct JSON invoked its constructor")
	}
	cycle := make(starlark.Tuple, 1)
	cycle[0] = cycle
	sd := NewSharedDict()
	if err := sd.SetKey(starlark.String("cycle"), cycle); err != nil {
		t.Fatal(err)
	}
	if text, err := sd.ToJSON(); text != "" || err == nil || !strings.Contains(err.Error(), "cycle in JSON structure") {
		t.Errorf("tuple cycle = (%q, %v)", text, err)
	}
}

// An application mutex is still needed for compound updates to native aliases;
// output locks do not make the two assignments one atomic application operation.
func TestSharedDict_ExternalSynchronization(t *testing.T) {
	sd := NewSharedDict()
	key := starlark.String("generation")
	items := starlark.NewList([]starlark.Value{starlark.MakeInt(0)})
	if err := sd.SetKey(key, starlark.MakeInt(0)); err != nil {
		t.Fatal(err)
	}
	if err := sd.SetKey(starlark.String("items"), items); err != nil {
		t.Fatal(err)
	}

	var gate sync.Mutex
	generation := 0
	update := func() error {
		gate.Lock()
		defer gate.Unlock()
		generation++
		value := starlark.MakeInt(generation)
		if err := sd.SetKey(key, value); err != nil {
			return err
		}
		// The alias mutation belongs to the same compound update as SetKey.
		return items.SetIndex(0, value)
	}
	check := func() error {
		gate.Lock()
		defer gate.Unlock()
		wantText := fmt.Sprintf("shared_dict({\"generation\": %d, \"items\": [%d]})", generation, generation)
		if got := sd.String(); got != wantText {
			return fmt.Errorf("String() = %q, want %q", got, wantText)
		}
		got, err := sd.ToJSON()
		if err != nil {
			return err
		}
		wantJSON := fmt.Sprintf("{\"generation\":%d,\"items\":[%d]}", generation, generation)
		if got != wantJSON {
			return fmt.Errorf("ToJSON() = %q, want %q", got, wantJSON)
		}
		return nil
	}

	const iterations = 128
	start := make(chan struct{})
	var workers sync.WaitGroup
	workers.Add(4)
	run := func(name string, operation func() error) {
		defer workers.Done()
		<-start
		for i := 0; i < iterations; i++ {
			if err := operation(); err != nil {
				t.Errorf("%s iteration %d: %v", name, i, err)
				return
			}
		}
	}
	for i := 0; i < 2; i++ {
		go run("writer", update)
		go run("reader", check)
	}
	close(start)
	workers.Wait()
	if generation != 2*iterations {
		t.Errorf("generation = %d, want %d", generation, 2*iterations)
	}
	if err := check(); err != nil {
		t.Error(err)
	}
}

func TestSharedDict_SnapshotStringFormattingGraph(t *testing.T) {
	const childEnv = "STARLET_TEST_SNAPSHOT_GRAPH_CHILD"
	if os.Getenv(childEnv) != "1" {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestSharedDict_SnapshotStringFormattingGraph$", "-test.v")
		cmd.Env = append(os.Environ(), childEnv+"=1")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("bounded formatting child: %v (timeout: %v)\n%s", err, ctx.Err(), output)
		}
		// Exercise the same guarded cases here so Go 1.19 records coverage.
	} else {
		debug.SetMaxStack(128 << 10)
	}
	t.Run("json_metadata_held_lock", func(t *testing.T) {
		sd := NewSharedDict()
		sd.Lock()
		text, err := EncodeStarlarkJSON(sd)
		sd.Unlock()
		if text != "" || err == nil {
			t.Fatalf("generic JSON while locked = (%q,%v)", text, err)
		}
	})
	t.Run("json_method_freeze", func(t *testing.T) {
		sd := NewSharedDict()
		method, err := sd.Attr("to_json")
		if err != nil {
			t.Fatal(err)
		}
		if err := sd.SetKey(starlark.String("method"), method); err != nil {
			t.Fatal(err)
		}
		sd.Freeze()
		if err := sd.SetKey(starlark.String("n"), starlark.MakeInt(1)); err == nil {
			t.Fatal("Freeze did not freeze mapping")
		}
	})
	put := func(t *testing.T, dict starlark.HasSetKey, key string, value starlark.Value) {
		t.Helper()
		if err := dict.SetKey(starlark.String(key), value); err != nil {
			t.Fatal(err)
		}
	}
	check := func(t *testing.T, dict *SharedDict, want string) {
		t.Helper()
		if got := dict.SnapshotString(); got != want {
			t.Fatalf("SnapshotString() = %q; want %q", got, want)
		}
	}
	t.Run("shared_dict", func(t *testing.T) {
		for _, test := range []struct{ name, want string }{
			{"list", `root({"items": [root({...})]})`},
			{"dict", `root({"items": {"back": root({...})}})`},
			{"tuple", `root({"items": (root({...}),)})`},
		} {
			t.Run(test.name, func(t *testing.T) {
				root := NewNamedSharedDict("root")
				var value starlark.Value
				switch test.name {
				case "list":
					value = starlark.NewList([]starlark.Value{root})
				case "dict":
					dict := starlark.NewDict(1)
					put(t, dict, "back", root)
					value = dict
				case "tuple":
					value = starlark.Tuple{root}
				}
				put(t, root, "items", value)
				check(t, root, test.want)
			})
		}
		t.Run("mutual", func(t *testing.T) {
			a, b, c := NewNamedSharedDict("a"), NewNamedSharedDict("b"), NewNamedSharedDict("c")
			back := starlark.NewDict(1)
			put(t, back, "back", a)
			put(t, a, "items", starlark.NewList([]starlark.Value{b}))
			put(t, b, "items", starlark.Tuple{c})
			put(t, c, "items", back)
			check(t, a, `a({"items": [b({"items": (c({"items": {"back": a({...})}}),)})]})`)
		})
		t.Run("aliases_and_empty", func(t *testing.T) {
			zero := new(SharedDict)
			zero.SetTypeName("zero")
			root, child := NewNamedSharedDict("root"), NewNamedSharedDict("child")
			put(t, child, "n", starlark.MakeInt(1))
			put(t, root, "items", starlark.NewList([]starlark.Value{child, child, NewNamedSharedDict("empty"), zero}))
			check(t, root, `root({"items": [child({"n": 1}), child({"n": 1}), empty({}), zero()]})`)
		})
	})
	t.Run("root_snapshot", func(t *testing.T) {
		root := NewNamedSharedDict("root")
		locked, calls := false, 0
		defer func() {
			if locked {
				root.Unlock()
			}
		}()
		callback := &sharedDictFormatValue{stringFunc: func() string {
			calls++
			put(t, root, "later", starlark.True)
			root.Lock()
			locked = true
			return "changed"
		}}
		put(t, root, "change", callback)
		put(t, root, "back", starlark.NewList([]starlark.Value{root}))
		put(t, root, "native", root.dict)
		check(t, root, `root({"change": changed, "back": [root({...})], "native": {...}})`)
		if calls != 1 {
			t.Fatalf("format callback calls = %d; want 1", calls)
		}
	})
	t.Run("external_nested_lock", func(t *testing.T) {
		root, child := NewNamedSharedDict("root"), NewNamedSharedDict("child")
		put(t, root, "next", starlark.NewList([]starlark.Value{child}))
		put(t, child, "next", starlark.NewList([]starlark.Value{root}))
		child.Lock()
		got := root.SnapshotString()
		child.Unlock()
		if want := `root({"next": [shared_dict(<busy>)]})`; got != want {
			t.Fatalf("locked child snapshot = %q; want %q", got, want)
		}
		check(t, root, `root({"next": [child({"next": [root({...})]})]})`)
	})
	t.Run("callbacks", func(t *testing.T) {
		for _, mode := range []string{"String", "Type"} {
			t.Run(mode, func(t *testing.T) {
				root, child := NewNamedSharedDict("root"), NewNamedSharedDict("child")
				calls := 0
				callback := func() string {
					calls++
					for _, dict := range []*SharedDict{root, child} {
						if !dict.TryLock() {
							t.Fatal("format callback ran with a mapping lock held")
						}
						dict.Unlock()
						if dict.Len() != 1 {
							t.Fatal("unexpected mapping length")
						}
					}
					put(t, root, "later", starlark.True)
					return "host"
				}
				host := &sharedDictFormatValue{stringFunc: func() string { return "receiver" }}
				var value starlark.Value = host
				wantValue := "host"
				if mode == "String" {
					host.stringFunc = callback
				} else {
					host.typeFunc = callback
					value = starlark.NewBuiltin("method", func(*starlark.Thread, *starlark.Builtin, starlark.Tuple, []starlark.Tuple) (starlark.Value, error) {
						return starlark.None, nil
					}).BindReceiver(host)
					wantValue = "<built-in method method of host value>"
				}
				put(t, child, "value", value)
				put(t, root, "items", starlark.NewList([]starlark.Value{child}))
				check(t, root, `root({"items": [child({"value": `+wantValue+`})]})`)
				if calls != 1 || root.Len() != 2 {
					t.Fatalf("callback calls=%d, root length=%d; want 1 and 2", calls, root.Len())
				}
			})
		}
	})
	t.Run("concurrent_root_writer", func(t *testing.T) {
		root, child := NewNamedSharedDict("root"), NewNamedSharedDict("child")
		put(t, child, "n", starlark.MakeInt(1))
		put(t, root, "items", starlark.NewList([]starlark.Value{child}))
		put(t, root, "counter", starlark.MakeInt(0))
		var workers sync.WaitGroup
		workers.Add(2)
		go func() {
			defer workers.Done()
			for i := 0; i < 128; i++ {
				if err := root.SetKey(starlark.String("counter"), starlark.MakeInt(i)); err != nil {
					t.Error(err)
					return
				}
			}
		}()
		go func() {
			defer workers.Done()
			for i := 0; i < 128; i++ {
				if got := root.SnapshotString(); got != sharedDictBusyString && (!strings.HasPrefix(got, `root({"items": [child({"n": 1})], "counter": `) || !strings.HasSuffix(got, "})")) {
					t.Errorf("malformed snapshot: %q", got)
					return
				}
			}
		}()
		workers.Wait()
	})
	t.Run("opposite_snapshots", func(t *testing.T) {
		a, b := NewNamedSharedDict("a"), NewNamedSharedDict("b")
		put(t, a, "next", starlark.NewList([]starlark.Value{b}))
		put(t, b, "next", starlark.NewList([]starlark.Value{a}))
		var workers sync.WaitGroup
		for _, test := range []struct {
			dict *SharedDict
			want string
			busy string
		}{{a, `a({"next": [b({"next": [a({...})]})]})`, `a({"next": [shared_dict(<busy>)]})`}, {b, `b({"next": [a({"next": [b({...})]})]})`, `b({"next": [shared_dict(<busy>)]})`}} {
			workers.Add(1)
			go func(dict *SharedDict, want, busy string) {
				defer workers.Done()
				for i := 0; i < 128; i++ {
					if got := dict.SnapshotString(); got != want && got != busy && got != sharedDictBusyString {
						t.Errorf("SnapshotString() = %q; want %q", got, want)
						return
					}
				}
			}(test.dict, test.want, test.busy)
		}
		workers.Wait()
		check(t, a, `a({"next": [b({"next": [a({...})]})]})`)
		check(t, b, `b({"next": [a({"next": [b({...})]})]})`)
	})
	t.Run("struct", func(t *testing.T) {
		t.Run("list_cycle", func(t *testing.T) {
			root := NewNamedSharedDict("root")
			list := starlark.NewList(nil)
			record := starlarkstruct.FromStringDict(starlark.String("record"), starlark.StringDict{"items": list})
			if err := list.Append(record); err != nil {
				t.Fatal(err)
			}
			put(t, root, "items", list)
			check(t, root, `root({"items": [record(items = [...])]})`)
		})
		t.Run("root_wrapper", func(t *testing.T) {
			root := NewNamedSharedDict("root")
			put(t, root, "items", starlarkstruct.FromStringDict(starlark.String("record"), starlark.StringDict{"root": root}))
			check(t, root, `root({"items": record(root = root({...}))})`)
		})
		t.Run("constructor_cycle", func(t *testing.T) {
			root := NewNamedSharedDict("root")
			constructor := starlark.NewList(nil)
			record := starlarkstruct.FromStringDict(constructor, starlark.StringDict{"n": starlark.MakeInt(1)})
			if err := constructor.Append(record); err != nil {
				t.Fatal(err)
			}
			put(t, root, "items", record)
			check(t, root, `root({"items": [...](n = 1)})`)
		})
		t.Run("oracle", func(t *testing.T) {
			for _, constructor := range []starlark.Value{starlarkstruct.Default, starlark.String("record\"\n"), starlark.MakeInt(7), starlark.Tuple{starlark.String("kind")}} {
				root := NewSharedDict()
				record := starlarkstruct.FromStringDict(constructor, starlark.StringDict{"z": starlark.None, "a": starlark.Tuple{starlark.True}})
				put(t, root, "items", starlark.NewList([]starlark.Value{record, record}))
				check(t, root, "shared_dict("+root.dict.String()+")")
			}
		})
		t.Run("duplicate_fields_oracle", func(t *testing.T) {
			root := NewSharedDict()
			record := starlarkstruct.FromKeywords(starlark.String("record"), []starlark.Tuple{{starlark.String("a"), starlark.MakeInt(1)}, {starlark.String("a"), starlark.MakeInt(2)}})
			put(t, root, "items", record)
			check(t, root, "shared_dict("+root.dict.String()+")")
		})
		for _, tupleKey := range []bool{false, true} {
			t.Run(fmt.Sprintf("key_path/tuple=%v", tupleKey), func(t *testing.T) {
				root := NewNamedSharedDict("root")
				record := starlarkstruct.FromStringDict(starlark.String("key"), starlark.StringDict{"n": starlark.MakeInt(1)})
				var key starlark.Value = record
				wantKey := "key(n = 1)"
				if tupleKey {
					key = starlark.Tuple{record}
					wantKey = "(key(n = 1),)"
				}
				bottom := starlark.NewDict(2)
				if err := bottom.SetKey(key, record); err != nil {
					t.Fatal(err)
				}
				put(t, bottom, "self", bottom)
				put(t, root, "items", starlark.NewList([]starlark.Value{starlark.NewList([]starlark.Value{bottom})}))
				check(t, root, `root({"items": [[{`+wantKey+`: key(n = 1), "self": {...}}]]})`)
			})
		}
		for _, test := range []struct {
			name  string
			value starlark.Value
		}{{"nil_field", starlarkstruct.FromStringDict(starlarkstruct.Default, starlark.StringDict{"value": nil})}, {"zero_value", new(starlarkstruct.Struct)}} {
			t.Run(test.name, func(t *testing.T) {
				root := NewSharedDict()
				put(t, root, "items", test.value)
				defer func() {
					if recover() == nil {
						t.Fatal("invalid Go struct no longer preserves upstream String panic")
					}
				}()
				root.SnapshotString()
			})
		}
	})
	t.Run("set", func(t *testing.T) {
		t.Run("constructor_cycle", func(t *testing.T) {
			root := NewNamedSharedDict("root")
			set, constructor := starlark.NewSet(1), starlark.NewList(nil)
			record := starlarkstruct.FromStringDict(constructor, nil)
			if err := set.Insert(record); err != nil {
				t.Fatal(err)
			}
			if err := constructor.Append(set); err != nil {
				t.Fatal(err)
			}
			put(t, root, "items", set)
			check(t, root, `root({"items": set([[set([...])]()])})`)
		})
		t.Run("oracle", func(t *testing.T) {
			for _, values := range [][]starlark.Value{nil, {starlark.None, starlark.True, starlark.Tuple{starlark.MakeInt(1)}}} {
				root, set := NewSharedDict(), starlark.NewSet(len(values))
				for _, value := range values {
					if err := set.Insert(value); err != nil {
						t.Fatal(err)
					}
				}
				put(t, root, "items", set)
				check(t, root, "shared_dict("+root.dict.String()+")")
			}
			for _, size := range []int{8, 9, 32} {
				t.Run(fmt.Sprintf("size=%d", size), func(t *testing.T) {
					root, set := NewSharedDict(), starlark.NewSet(size)
					for i := 0; i < size; i++ {
						if err := set.Insert(starlark.MakeInt(i)); err != nil {
							t.Fatal(err)
						}
					}
					put(t, root, "items", set)
					check(t, root, "shared_dict("+root.dict.String()+")")
				})
			}
		})
		t.Run("deep_path", func(t *testing.T) {
			root, set := NewSharedDict(), starlark.NewSet(1)
			if err := set.Insert(starlark.MakeInt(1)); err != nil {
				t.Fatal(err)
			}
			var value starlark.Value = set
			for i := 0; i < 9; i++ {
				value = starlark.NewList([]starlark.Value{value})
			}
			put(t, root, "items", value)
			check(t, root, "shared_dict("+root.dict.String()+")")
		})
		t.Run("deep_ancestor_cycle", func(t *testing.T) {
			root, set := NewSharedDict(), starlark.NewSet(1)
			var value starlark.Value = set
			for i := 0; i < 9; i++ {
				value = starlark.NewList([]starlark.Value{value})
			}
			record := starlarkstruct.FromStringDict(value, nil)
			if err := set.Insert(record); err != nil {
				t.Fatal(err)
			}
			put(t, root, "items", value)
			check(t, root, `shared_dict({"items": `+strings.Repeat("[", 9)+`set([[...]()])`+strings.Repeat("]", 9)+`})`)
		})
		t.Run("callback_insert", func(t *testing.T) {
			root, set := NewSharedDict(), starlark.NewSet(1)
			calls := 0
			value := &sharedDictFormatValue{stringFunc: func() string {
				calls++
				if err := set.Insert(starlark.String("later")); err != nil {
					t.Fatal(err)
				}
				return "first"
			}}
			if err := set.Insert(value); err != nil {
				t.Fatal(err)
			}
			put(t, root, "items", set)
			check(t, root, `shared_dict({"items": set([first])})`)
			if calls != 1 || set.Len() != 2 {
				t.Fatalf("format calls=%d, set length=%d; want 1 and 2", calls, set.Len())
			}
		})
		t.Run("callback_panic", func(t *testing.T) {
			root, set := NewSharedDict(), starlark.NewSet(1)
			value := &sharedDictFormatValue{stringFunc: func() string { panic("set format panic") }}
			if err := set.Insert(value); err != nil {
				t.Fatal(err)
			}
			put(t, root, "items", set)
			func() {
				defer func() {
					if got := recover(); got != "set format panic" {
						t.Fatalf("format panic = %v", got)
					}
				}()
				root.SnapshotString()
			}()
			if err := set.Insert(starlark.True); err != nil {
				t.Fatalf("set iterator remained active after formatting panic: %v", err)
			}
		})
	})
}

type sharedDictFormatValue struct {
	stringFunc  func() string
	hashFunc    func() (uint32, error)
	comparisons *int
	compareFunc func()
	freezeFunc  func()
	typeFunc    func() string
}

func (v *sharedDictFormatValue) String() string { return v.stringFunc() }
func (v *sharedDictFormatValue) Type() string {
	if v.typeFunc != nil {
		return v.typeFunc()
	}
	return "format_value"
}
func (v *sharedDictFormatValue) Freeze() {
	if v.freezeFunc != nil {
		v.freezeFunc()
	}
}
func (*sharedDictFormatValue) Truth() starlark.Bool { return true }
func (v *sharedDictFormatValue) Hash() (uint32, error) {
	if v.hashFunc != nil {
		return v.hashFunc()
	}
	return 1, nil
}
func (v *sharedDictFormatValue) CompareSameType(op syntax.Token, other starlark.Value, _ int) (bool, error) {
	if v.compareFunc != nil {
		v.compareFunc()
	}
	if v.comparisons != nil {
		*v.comparisons++
	}
	equal := v == other.(*sharedDictFormatValue)
	return equal == (op == syntax.EQL), nil
}

func TestSharedDict_StringBusyReentry(t *testing.T) {
	for _, test := range []string{"external_lock", "external_read_lock", "Hash", "Equal", "Freeze"} {
		t.Run(test, func(t *testing.T) {
			sd := NewNamedSharedDict("named")
			type result struct {
				text string
				err  error
			}
			var pending []chan result
			checks := 0
			checkOutput := func() {
				checks++
				for _, output := range []struct {
					name string
					call func() (string, error)
					json bool
				}{{"String", func() (string, error) { return sd.String(), nil }, false}, {"SnapshotString", func() (string, error) { return sd.SnapshotString(), nil }, false}, {"ToJSON", sd.ToJSON, true}} {
					done := make(chan result, 1)
					call := output.call
					go func() { text, err := call(); done <- result{text, err} }()
					select {
					case got := <-done:
						if output.json {
							if got.text != "" || !errors.Is(got.err, ErrSharedDictBusy) {
								t.Errorf("locked %s = (%q,%v)", output.name, got.text, got.err)
							}
						} else if got.text != sharedDictBusyString || got.err != nil {
							t.Errorf("locked %s = (%q,%v)", output.name, got.text, got.err)
						}
					case <-time.After(2 * time.Second):
						t.Errorf("%s waited for the mapping lock", output.name)
						pending = append(pending, done)
					}
				}
			}
			key := &sharedDictFormatValue{stringFunc: func() string { return "key" }}
			switch test {
			case "external_lock":
				sd.Lock()
				checkOutput()
				sd.Unlock()
			case "external_read_lock":
				key.stringFunc = func() string {
					if err := sd.SetKey(starlark.String("new"), starlark.None); err != nil {
						t.Error(err)
					}
					return "key"
				}
				if err := sd.SetKey(starlark.String("value"), key); err != nil {
					t.Fatal(err)
				}
				sd.RLock()
				checkOutput()
				sd.RUnlock()
			case "Hash":
				key.hashFunc = func() (uint32, error) { checkOutput(); return 1, nil }
				if err := sd.SetKey(key, starlark.True); err != nil {
					t.Fatal(err)
				}
			case "Equal":
				if err := sd.SetKey(key, starlark.True); err != nil {
					t.Fatal(err)
				}
				other := &sharedDictFormatValue{stringFunc: func() string { return "other" }, compareFunc: checkOutput}
				if err := sd.SetKey(other, starlark.False); err != nil {
					t.Fatal(err)
				}
			case "Freeze":
				key.freezeFunc = checkOutput
				if err := sd.SetKey(starlark.String("value"), key); err != nil {
					t.Fatal(err)
				}
				sd.Freeze()
			}
			for _, done := range pending {
				select {
				case <-done:
				case <-time.After(2 * time.Second):
					t.Fatal("output worker remained blocked after unlocking")
				}
			}
			if checks != 1 {
				t.Fatalf("callback count = %d; want 1", checks)
			}
		})
	}
}

func TestSharedDict_SnapshotStringGoOnly(t *testing.T) {
	sd := NewSharedDict()
	for _, name := range sd.AttrNames() {
		if name == "SnapshotString" || name == "snapshot_string" {
			t.Fatalf("Go snapshot method was added to the Starlark surface: %s", name)
		}
	}
	if value, _ := sd.Attr("SnapshotString"); value != nil {
		t.Fatal("Go SnapshotString unexpectedly exposed as a Starlark attribute")
	}
}

func ExampleSharedDict_SnapshotString() {
	shared := NewSharedDict()
	if err := shared.SetKey(starlark.String("count"), starlark.MakeInt(1)); err != nil {
		panic(err)
	}
	fmt.Println(shared.SnapshotString())
	// Output: shared_dict({"count": 1})
}

var sharedDictStringBenchmarkResult string

func BenchmarkSharedDictReadPaths(b *testing.B) {
	makeFlat := func(n int) *SharedDict {
		dict := NewSharedDict()
		for i := 0; i < n; i++ {
			if err := dict.SetKey(starlark.String(fmt.Sprintf("key_%03d", i)), starlark.MakeInt(i)); err != nil {
				b.Fatal(err)
			}
		}
		return dict
	}
	flat8 := makeFlat(8)
	nested := NewSharedDict()
	if err := nested.SetKey(starlark.String("nested"), flat8.dict); err != nil {
		b.Fatal(err)
	}
	nestedShared := NewSharedDict()
	if err := nestedShared.SetKey(starlark.String("nested"), starlark.NewList([]starlark.Value{flat8})); err != nil {
		b.Fatal(err)
	}
	repeatedShared := NewSharedDict()
	if err := repeatedShared.SetKey(starlark.String("nested"), starlark.NewList([]starlark.Value{flat8, flat8})); err != nil {
		b.Fatal(err)
	}
	record := NewSharedDict()
	if err := record.SetKey(starlark.String("record"), starlarkstruct.FromStringDict(starlark.String("record"), starlark.StringDict{"entries": flat8.dict})); err != nil {
		b.Fatal(err)
	}
	makeSet := func(n int, records bool) *SharedDict {
		dict, set := NewSharedDict(), starlark.NewSet(n)
		for i := 0; i < n; i++ {
			var value starlark.Value = starlark.MakeInt(i)
			if records {
				value = starlarkstruct.FromStringDict(starlark.String("key"), starlark.StringDict{"id": value})
			}
			if err := set.Insert(value); err != nil {
				b.Fatal(err)
			}
		}
		if err := dict.SetKey(starlark.String("set"), set); err != nil {
			b.Fatal(err)
		}
		return dict
	}
	cycle := NewSharedDict()
	if err := cycle.SetKey(starlark.String("self"), cycle.dict); err != nil {
		b.Fatal(err)
	}
	for _, test := range []struct {
		name string
		dict *SharedDict
	}{
		{"Flat0", makeFlat(0)}, {"Flat1", makeFlat(1)}, {"Flat8", flat8},
		{"Flat128", makeFlat(128)}, {"Nested8", nested}, {"Cycle", cycle},
		{"NestedShared8", nestedShared}, {"RepeatedShared8", repeatedShared}, {"Struct8", record},
		{"Set8", makeSet(8, false)}, {"SetStruct8", makeSet(8, true)},
		{"Set32", makeSet(32, false)}, {"SetStruct32", makeSet(32, true)},
	} {
		b.Run("Snapshot/"+test.name, func(b *testing.B) {
			format := test.dict.SnapshotString
			if got, want := format(), test.dict.String(); got != want {
				b.Fatalf("SnapshotString() = %q; want %q", got, want)
			}
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				sharedDictStringBenchmarkResult = format()
			}
		})
	}
	b.Run("ControlLegacyString8", func(b *testing.B) {
		format := flat8.String
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			sharedDictStringBenchmarkResult = format()
		}
	})
	_, program, err := starlark.SourceProgram("print.star", "print(shared)", func(name string) bool { return name == "shared" })
	if err != nil {
		b.Fatal(err)
	}
	thread := &starlark.Thread{Print: func(_ *starlark.Thread, text string) { sharedDictStringBenchmarkResult = text }}
	globals := starlark.StringDict{"shared": flat8}
	b.Run("ControlLegacyPrintFlat8", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if _, err := program.Init(thread, globals); err != nil {
				b.Fatal(err)
			}
		}
	})
	b.Run("ControlPlainDictString8", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			sharedDictStringBenchmarkResult = flat8.dict.String()
		}
	})
	b.Run("ControlToJSONFlat8", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			got, err := flat8.ToJSON()
			if err != nil {
				b.Fatal(err)
			}
			sharedDictStringBenchmarkResult = got
		}
	})
}

func TestSharedDict_Clone(t *testing.T) {
	sd1 := NewSharedDictFromDict(nil)
	if sd1 == nil {
		t.Errorf("NewSharedDictFromDict: 1 nil")
		return
	}
	if sd1.dict == nil {
		t.Errorf("NewSharedDictFromDict: 1 nil dict")
		return
	}
	if sd1.Len() != 0 {
		t.Errorf("NewSharedDictFromDict: 1 non-empty dict")
		return
	}

	d2 := starlark.NewDict(1)
	d2.SetKey(starlark.String("foo"), starlark.String("bar"))
	sd2 := NewSharedDictFromDict(d2)
	if sd2 == nil {
		t.Errorf("NewSharedDictFromDict: 2 nil")
		return
	}
	if sd2.dict == nil {
		t.Errorf("NewSharedDictFromDict: 2 nil dict")
		return
	}
	if sd2.Len() != 1 {
		t.Errorf("NewSharedDictFromDict: 2 empty dict")
		return
	}
	if d2 == sd2.dict {
		t.Errorf("NewSharedDictFromDict: 2 not cloned")
		return
	}

	dd, err := sd2.CloneDict()
	if dd == nil || err != nil {
		t.Errorf("Clone: nil, Error: %v", err)
		return
	}
	if dd == sd2.dict {
		t.Errorf("Clone: not cloned")
		return
	}
	if !reflect.DeepEqual(dd, sd2.dict) {
		t.Errorf("Clone: not deep cloned")
		return
	}
}

func TestSharedDict_TypeName(t *testing.T) {
	sd := NewSharedDict()
	if sd.Type() != "shared_dict" {
		t.Errorf("TypeName1: %s", sd.Type())
		return
	}

	sd.SetTypeName("Kioku")
	if sd.Type() != "Kioku" {
		t.Errorf("TypeName2: %s", sd.Type())
		return
	}

	sd.SetTypeName("")
	if sd.Type() != "shared_dict" {
		t.Errorf("TypeName3: %s", sd.Type())
		return
	}
}

func TestSharedDict_Unpacker(t *testing.T) {
	// test unpacker with starlark.Dict
	d := starlark.NewDict(1)
	d.SetKey(starlark.String("foo"), starlark.String("bar"))

	var ad *starlark.Dict
	err := starlark.UnpackArgs("f", starlark.Tuple{d}, []starlark.Tuple{}, "a", &ad)
	if err != nil {
		t.Errorf("unpack standard args error: %v", err)
		return
	}
	t.Logf("dict: %v", ad)

	// test unpacker with SharedDict
	sd := NewSharedDict()
	sd.SetKey(starlark.String("four"), starlark.String("five"))

	var asd *SharedDict
	err = starlark.UnpackArgs("f2", starlark.Tuple{sd}, []starlark.Tuple{}, "a", &asd)
	if err != nil {
		t.Errorf("unpack shared dict args error: %v", err)
		return
	}
	t.Logf("sdict: %v", asd)

	// check if it's the same instance inside
	if &sd.RWMutex != &asd.RWMutex {
		t.Errorf("unpack shared dict: not the same RWMutex")
		return
	}
	t.Logf("shared sdict lock: %p --> %p", &sd.RWMutex, &asd.RWMutex)

	if sd.dict != asd.dict {
		t.Errorf("unpack shared dict: not the same dict")
		return
	}
	t.Logf("shared sdict dict: %p --> %p", sd.dict, asd.dict)
}

func TestSharedDict_Functions(t *testing.T) {
	tests := []struct {
		name    string
		script  string
		wantErr string
	}{
		{
			name: `get: not found`,
			script: itn.HereDoc(`
				load('share', 'sd')
				v = sd["foo"]
			`),
			wantErr: `key "foo" not in shared_dict`,
		},
		{
			name: `set then get`,
			script: itn.HereDoc(`
				load('share', 'sd')
				e = "bar"
				sd["foo"] = e
				v = sd["foo"]
				assert.eq(v, e)
			`),
		},
		{
			name: `set twice`,
			script: itn.HereDoc(`
				load('share', 'sd')
				sd["foo"] = "bar"
				sd["foo"] = "cat"
				v = sd["foo"]
				assert.eq(v, "cat")
			`),
		},
		{
			name: `type`,
			script: itn.HereDoc(`
				load('share', 'sd')
				t = type(sd)
				assert.eq(t, "shared_dict")
			`),
		},
		{
			name: `no len`,
			script: itn.HereDoc(`
				load('share', 'sd')
				assert.eq(len(sd), 0)
			`),
			wantErr: `len: value of type shared_dict has no len`,
		},
		{
			name: `truth`,
			script: itn.HereDoc(`
				load('share', 'sd')

				def truth(v):
					t = False
					if v:
						t = True
					return t

				assert.true(bool(sd) == False)
				assert.true(truth(sd) == False)

				sd["foo"] = "bar"
				assert.true(bool(sd) == True)
				assert.true(truth(sd) == True)
			`),
		},
		{
			name: `compare`,
			script: itn.HereDoc(`
				load('share', 'sd', sd3='another')
				sd1 = sd
				sd2 = sd
				assert.true(sd1 == sd2)
				assert.true(sd1 != sd3)
				assert.true(not (sd1 != sd2))

				sd1.clear()
				sd1["your"] = "name"
				assert.true(sd1 == sd2)
				assert.true(sd1 == sd3)

				assert.true(None != sd1)
				assert.true(None != sd2)
				assert.true(None != sd3)
			`),
		},
		{
			name: `compare: invalid operator`,
			script: itn.HereDoc(`
				load('share', 'sd')
				sd1 = sd
				sd2 = sd
				sd1 < sd2
			`),
			wantErr: `unsupported operator: <`,
		},
		{
			name: `compare: diff type`,
			script: itn.HereDoc(`
				load('share', 'sd')
				sd["foo"] = "bar"
				dd = {"foo": "bar"}
				print(sd, dd)
				assert.true(sd != dd)
				assert.true(sd != 123)
				assert.true(dd != 456)
			`),
		},
		{
			name: `string`,
			script: itn.HereDoc(`
				load('share', 'sd')
				s = str(sd)
				assert.eq(s, "shared_dict({})")
				sd["foo"] = "bar"
				s2 = str(sd)
				assert.eq(s2, 'shared_dict({"foo": "bar"})')
			`),
		},
		{
			name: `self contain key`,
			script: itn.HereDoc(`
				load('share', 'sd')
				sd[sd] = "self"
			`),
			wantErr: `unhashable type: shared_dict`,
		},
		{
			name: `self contain value`,
			script: itn.HereDoc(`
				load('share', 'sd')
				sd["self"] = sd
			`),
			wantErr: `unsupported value: shared_dict`,
		},
		{
			name: `no shared dict as value`,
			script: itn.HereDoc(`
				load('share', 'sd', sd3='another')
				sd["another"] = sd3
			`),
			wantErr: `unsupported value: shared_dict`,
		},
		{
			name: `attrs`,
			script: itn.HereDoc(`
				load('share', 'sd')
				l = dir(sd)
				print(l)
				assert.eq(l, ["clear", "from_json", "get", "items", "keys", "len", "perform", "pop", "popitem", "setdefault", "to_dict", "to_json", "update", "values"])
			`),
		},
		{
			name: `attr: not found`,
			script: itn.HereDoc(`
				load('share', 'sd')
				v = sd.foo
			`),
			wantErr: `shared_dict has no .foo field or method`,
		},
		{
			name: `attr: setdefault`,
			script: itn.HereDoc(`
				load('share', 'sd')
				v = sd.get("foo")
				assert.eq(v, None)

				sd["foo"] = "bar"			
				assert.eq(sd.setdefault("foo"), "bar")

				assert.eq(sd.setdefault("bar"), None)
				assert.eq(sd.setdefault("cat", 123), 123)
				assert.eq(sd.setdefault("bar"), None)
				assert.eq(sd.setdefault("cat"), 123)
				print(sd)
			`),
		},
		{
			name: `attr: clear`,
			script: itn.HereDoc(`
				load('share', 'sd')
				sd["foo"] = "bar"
				assert.eq(sd.get("foo"), "bar")
				assert.eq(sd.len(), 1)
				sd.clear()
				assert.eq(sd.get("foo"), None)
				assert.eq(sd.len(), 0)
			`),
		},
		{
			name: `attr: items`,
			script: itn.HereDoc(`
				load('share', 'sd')	
				sd["foo"] = "cat"
				sd["bar"] = "dog"
				assert.eq(sd.items(), [("foo", "cat"), ("bar", "dog")])
				assert.eq(sd.keys(), ["foo", "bar"])
				assert.eq(sd.values(), ["cat", "dog"])
				
				sd.update([("foo", "dog")], bar="cat")
				assert.eq(sd.items(), [("foo", "dog"), ("bar", "cat")])

				sd.pop("foo")
				assert.eq(sd.items(), [("bar", "cat")])

				sd.popitem()
				assert.eq(sd.items(), [])
			`),
		},
		{
			name: `attr: pop missing`,
			script: itn.HereDoc(`
				load('share', 'sd')
				v = sd.pop("foo")
			`),
			wantErr: `pop: missing key`,
		},
		{
			name: `attr: popitem empty`,
			script: itn.HereDoc(`
				load('share', 'sd')
				v = sd.popitem()
			`),
			wantErr: `popitem: empty dict`,
		},
		{
			name: `attr: upsert`,
			script: itn.HereDoc(`
				load('share', 'sd')
				def inc(i = 1):
					sd["cnt"] = sd.get("cnt", 0) + i

				inc()
				inc()
				inc(3)

				assert.eq(sd["cnt"], 5)
			`),
		},
		{
			name: `attr: custom len`,
			script: itn.HereDoc(`
				load('share', 'sd')
				assert.eq(sd.len(), 0)
				sd["foo"] = "bar"
				assert.eq(sd.len(), 1)
			`),
		},
		{
			name: `attr: custom len -- invalid args`,
			script: itn.HereDoc(`
				load('share', 'sd')
				sd.len(123)
			`),
			wantErr: `len: got 1 arguments, want 0`,
		},
		{
			name: `attr: custom perform`,
			script: itn.HereDoc(`
				load('share', 'sd')
				def act(d):
					d["cnt"] = d.get("cnt", 100) + 1
				assert.eq(sd.get("cnt"), None)
				sd.perform(act)
				assert.eq(sd["cnt"], 101)
			`),
		},
		{
			name: `attr: custom perform -- error`,
			script: itn.HereDoc(`
				load('share', 'sd')
				def ungood(d):
					fail("not good~{}".format(d))
				sd["foo"] = "bar"
				sd.perform(ungood)
			`),
			wantErr: `fail: not good~{"foo": "bar"}`,
		},
		{
			name: `attr: custom perform -- error 2`,
			script: itn.HereDoc(`
				load('share', 'sd')
				def ungood(d, a, b, c):
					print("two")
				sd["foo"] = "bar"
				sd.perform(ungood)
			`),
			wantErr: `function ungood missing 3 arguments (a, b, c)`,
		},
		{
			name: `attr: custom perform -- error 3`,
			script: itn.HereDoc(`
				load('share', 'sd')
				def ungood():
					print("three")
				sd["foo"] = "bar"
				sd.perform(ungood)
			`),
			wantErr: `function ungood accepts no arguments (1 given)`,
		},
		{
			name: `attr: custom perform`,
			script: itn.HereDoc(`
				load('share', 'sd')
				def act(d):
					d["cnt"] = d.get("cnt", 0) + 1
				print(sd)
				sd.perform(act)
				print(sd)
			`),
		},
		{
			name: `attr: custom perform return`,
			script: itn.HereDoc(`
				load('share', 'sd')
				sd["cnt"] = 100
				def act(d):
					return d["cnt"] + 1
				x = sd.perform(act)
				assert.eq(x, 101)
			`),
		},
		{
			name: `attr: custom perform return`,
			script: itn.HereDoc(`
				load('share', 'sd')
				sd["a"] = 100
				sd["b"] = 200
				x = sd.perform(lambda d: d.get("a", 0) + d.get("b", 0) + 50)
				assert.eq(x, 350)
			`),
		},
		{
			name: `attr: custom perform -- invalid args`,
			script: itn.HereDoc(`
				load('share', 'sd')
				sd.perform()
			`),
			wantErr: `perform: missing argument for fn`,
		},
		{
			name: `attr: custom perform -- invalid type`,
			script: itn.HereDoc(`
				load('share', 'sd')
				sd.perform(123)
			`),
			wantErr: `perform: not callable type: int`,
		},
		{
			name: `attr: custom to_dict -- invalid args`,
			script: itn.HereDoc(`
				load('share', 'sd')
				sd.to_dict(123)
			`),
			wantErr: `to_dict: got 1 arguments, want 0`,
		},
		{
			name: `attr: custom to_dict -- empty`,
			script: itn.HereDoc(`
				load('share', 'sd')
				d = sd.to_dict()
				assert.eq(d, {})
			`),
		},
		{
			name: `attr: custom to_dict -- one`,
			script: itn.HereDoc(`
				load('share', 'sd')
				sd["foo"] = "bar"
				d = sd.to_dict()
				assert.eq(d, {"foo":"bar"})
			`),
		},
		{
			name: `attr: custom to_dict -- shallow nested aliases`,
			script: itn.HereDoc(`
				load('share', 'sd')
				sd["items"] = [1]
				d = sd.to_dict()
				d["new"] = 3
				d["items"].append(2)
				assert.eq("new" in sd, False)
				assert.eq(sd["items"], [1, 2])
			`),
		},
		{
			name: `attr: custom to_dict -- two`,
			script: itn.HereDoc(`
				load('share', 'sd')
				sd["foo"] = "bar"
				sd["cnt"] = 100
				d = sd.to_dict()
				assert.eq(d, {"foo":"bar", "cnt":100})
			`),
		},
		{
			name: `attr: custom to_json -- invalid args`,
			script: itn.HereDoc(`
				load('share', 'sd')
				sd.to_json(123)
			`),
			wantErr: `to_json: got 1 arguments, want 0`,
		},
		{
			name: `attr: custom to_json -- empty`,
			script: itn.HereDoc(`
				load('share', 'sd')
				j = sd.to_json()
				assert.eq(j, "{}")
			`),
		},
		{
			name: `attr: custom to_json -- one`,
			script: itn.HereDoc(`
				load('share', 'sd')
				sd["foo"] = "bar"
				j = sd.to_json()
				assert.eq(j, '{"foo":"bar"}')
			`),
		},
		{
			name: `attr: custom from_json -- missing args`,
			script: itn.HereDoc(`
				load('share', 'sd')
				sd.from_json()
			`),
			wantErr: `from_json: missing argument for x`,
		},
		{
			name: `attr: custom from_json -- invalid type`,
			script: itn.HereDoc(`
				load('share', 'sd')
				sd.from_json(123)
			`),
			wantErr: `from_json: for parameter x: got int, want string or bytes`,
		},
		{
			name: `attr: custom from_json -- not dict`,
			script: itn.HereDoc(`
				load('share', 'sd')
				sd.from_json("123")
			`),
			wantErr: `got int, want dict`,
		},
		{
			name: `attr: custom from_json`,
			script: itn.HereDoc(`
				load('share', 'sd')
				sd["a"] = 100
				sd["b"] = 200

				x = sd.from_json('{}')
				assert.eq(sd.items(), {"a": 100, "b": 200}.items())
				assert.eq(x, {})

				y = sd.from_json('{"b":300, "c":400}')
				print(sd, y)
				assert.eq(sd.items(), {"a": 100, "b": 300, "c": 400}.items())
				assert.eq(y, {"b": 300, "c": 400})
			`),
		},
		{
			name: `args`,
			script: itn.HereDoc(`
				load('share', 'sd')
				def work(d):
					d["cnt"] = d.get("cnt", 100) + 1
				work(sd)
				assert.eq(sd["cnt"], 101)
			`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := itn.ExecModuleWithErrorTest(t, "share", getSDLoader("sd", NewSharedDict()), tt.script, tt.wantErr, nil)
			if (err != nil) != (tt.wantErr != "") {
				t.Errorf("sd(%q) expects error = '%v', actual error = '%v', result = %v", tt.name, tt.wantErr, err, res)
				return
			}
		})
	}
}

func TestSharedDict_Frozen(t *testing.T) {
	tests := []struct {
		name    string
		script  string
		wantErr string
	}{
		// frozen dict can be read
		{
			name: `len`,
			script: itn.HereDoc(`
				load('share', 'sd')
				assert.eq(sd.len(), 1)
			`),
		},
		{
			name: `get`,
			script: itn.HereDoc(`
				load('share', 'sd')
				assert.eq(sd["foo"], "bar")
				assert.eq(sd.get("foo"), "bar")
			`),
		},
		{
			name: `items`,
			script: itn.HereDoc(`
				load('share', 'sd')	
				assert.eq(sd.items(), [("foo", "bar")])
			`),
		},
		{
			name: `keys`,
			script: itn.HereDoc(`
				load('share', 'sd')	
				assert.eq(sd.keys(), ["foo"])
			`),
		},
		{
			name: `values`,
			script: itn.HereDoc(`
				load('share', 'sd')	
				assert.eq(sd.values(), ["bar"])
			`),
		},
		{
			name: `compare`,
			script: itn.HereDoc(`
				load('share', 'sd', sd3='another')
				assert.true(sd != sd3)
				sd3.clear()
				sd3["foo"] = "bar"
				assert.true(sd == sd3)
				assert.true(None != sd)
				assert.true(None != sd3)
			`),
		},
		{
			name: `perform: get`,
			script: itn.HereDoc(`
				load('share', 'sd')
				v = []
				def act(d):
					v.append(d["foo"])
				sd.perform(act)
				assert.eq(v, ["bar"])
			`),
		},
		{
			name: `to_json`,
			script: itn.HereDoc(`
				load('share', 'sd')
				j = sd.to_json()
				assert.eq(j, '{"foo":"bar"}')
			`),
		},
		{
			name: `to_dict`,
			script: itn.HereDoc(`
				load('share', 'sd')
				d = sd.to_dict()
				assert.eq(d, {"foo":"bar"})
			`),
		},

		// frozen dict cannot be modified
		{
			name: `clear`,
			script: itn.HereDoc(`
				load('share', 'sd')
				sd.clear()
			`),
			wantErr: `cannot clear frozen hash table`,
		},
		{
			name: `delete`,
			script: itn.HereDoc(`
				load('share', 'sd')
				sd.popitem()
			`),
			wantErr: `popitem: cannot delete from frozen hash table`,
		},
		{
			name: `delete 2`,
			script: itn.HereDoc(`
				load('share', 'sd')
				sd.pop("foo")
			`),
			wantErr: `pop: cannot delete from frozen hash table`,
		},
		{
			name: `set key`,
			script: itn.HereDoc(`
				load('share', 'sd')
				sd["fly"] = "away"
			`),
			wantErr: `frozen shared_dict`,
		},
		{
			name: `setdefault`,
			script: itn.HereDoc(`
				load('share', 'sd')
				assert.eq(sd.setdefault("foo", "too"), "bar")
				sd.setdefault("zoo", "bar")
			`),
			wantErr: `setdefault: cannot insert into frozen hash table`,
		},
		{
			name: `attr: update`,
			script: itn.HereDoc(`
				load('share', 'sd')	
				sd.update([("foo", "dog")], bar="cat")
			`),
			wantErr: `update: cannot insert into frozen hash table`,
		},
		{
			name: `perform: set`,
			script: itn.HereDoc(`
				load('share', 'sd')
				v = "bar"
				def act(d):
					d[v] = d["foo"]
				sd.perform(act)
			`),
			wantErr: `cannot insert into frozen hash table`,
		},
		{
			name: `attr: custom from_json`,
			script: itn.HereDoc(`
				load('share', 'sd')
				print(sd)
				x = sd.from_json('{"a": 120}')
			`),
			wantErr: `cannot insert into frozen hash table`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sd := NewSharedDict()
			if err := sd.SetKey(starlark.String("foo"), starlark.String("bar")); err != nil {
				t.Errorf("set key error: %v", err)
				return
			}
			sd.Freeze()
			res, err := itn.ExecModuleWithErrorTest(t, "share", getSDLoader("sd", sd), tt.script, tt.wantErr, nil)
			if (err != nil) != (tt.wantErr != "") {
				t.Errorf("frozen sd(%q) expects error = '%v', actual error = '%v', result = %v", tt.name, tt.wantErr, err, res)
				return
			}
		})
	}
}

func TestSharedDict_NilDict(t *testing.T) {
	tests := []struct {
		name    string
		script  string
		wantErr string
	}{
		{
			name: `get`,
			script: itn.HereDoc(`
				load('share', 'sd')
				v = sd["foo"]
			`),
			wantErr: `key "foo" not in shared_dict`,
		},
		{
			name: `set`,
			script: itn.HereDoc(`
				load('share', 'sd')
				sd["foo"] = "bar"
				assert.eq(sd["foo"], "bar")
			`),
		},
		{
			name: `attr: len`,
			script: itn.HereDoc(`
				load('share', 'sd')
				assert.eq(sd.len(), 0)
			`),
		},
		{
			name: `to json`,
			script: itn.HereDoc(`
				load('share', 'sd')
				assert.eq(sd.to_json(), '{}')
			`),
		},
		{
			name: `attr names`,
			script: itn.HereDoc(`
				load('share', 'sd')
				assert.eq(dir(sd), ['clear', 'from_json', 'get', 'items', 'keys', 'len', 'perform', 'pop', 'popitem', 'setdefault', 'to_dict', 'to_json', 'update', 'values'])
			`),
		},
		{
			name: `compare`,
			script: itn.HereDoc(`
				load('share', 'sd', sd2='broken', sd3='another')
				assert.true(sd == sd)
				assert.true(sd == sd2)
				assert.true(not(sd != sd2))
				assert.true(sd != sd3)
				assert.true(not(sd == sd3))
			`),
		},
		{
			name: `compare unsupported`,
			script: itn.HereDoc(`
				load('share', 'sd', sd3='another')
				sd >= sd3
			`),
			wantErr: `unsupported operator: >=`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sd := NewSharedDict()
			sd.dict = nil
			res, err := itn.ExecModuleWithErrorTest(t, "share", getSDLoader("sd", sd), tt.script, tt.wantErr, nil)
			if (err != nil) != (tt.wantErr != "") {
				t.Errorf("nil inside sd(%q) expects error = '%v', actual error = '%v', result = %v", tt.name, tt.wantErr, err, res)
				return
			}
		})
	}
}

func TestSharedDict_Named(t *testing.T) {
	tests := []struct {
		name    string
		script  string
		wantErr string
	}{
		{
			name: `get error`,
			script: itn.HereDoc(`
				load('share', 'sd')
				v = sd["foo"]
			`),
			wantErr: `key "foo" not in Kioku`,
		},
		{
			name: `string`,
			script: itn.HereDoc(`
				load('share', 'sd')
				sd[1] = 2
				assert.eq(str(sd), "Kioku({1: 2})")
			`),
		},
		{
			name: `type`,
			script: itn.HereDoc(`
				load('share', 'sd')
				assert.eq(type(sd), "Kioku")
			`),
		},
		{
			name: `hash error`,
			script: itn.HereDoc(`
				load('share', 'sd')
				d = { sd: 123 }
			`),
			wantErr: `unhashable type: Kioku`,
		},
		{
			name: `set frozen`,
			script: itn.HereDoc(`
				load('share', sf='frozen')
				sf["foo"] = "bar"
			`),
			wantErr: `frozen shared_dict`,
		},
		{
			name: `set same type`,
			script: itn.HereDoc(`
				load('share', 'sd', sd3='another')
				sd3["foo"] = sd
			`),
			wantErr: `unsupported value: Kioku`,
		},
		{
			name: `compare`,
			script: itn.HereDoc(`
				load('share', 'sd', sd3='another')
				assert.true(sd == sd)
				assert.true(sd != sd3)
				assert.true(not(sd == sd3))
			`),
		},
		{
			name: `compare unsupported`,
			script: itn.HereDoc(`
				load('share', 'sd', sd3='another')
				sd >= sd3
			`),
			wantErr: `unsupported operator: >=`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sd := NewNamedSharedDict("Kioku")
			sd.dict = nil
			res, err := itn.ExecModuleWithErrorTest(t, "share", getSDLoader("sd", sd), tt.script, tt.wantErr, nil)
			if (err != nil) != (tt.wantErr != "") {
				t.Errorf("nil inside sd(%q) expects error = '%v', actual error = '%v', result = %v", tt.name, tt.wantErr, err, res)
				return
			}
		})
	}
}

func TestSharedDict_Concurrent(t *testing.T) {
	s1 := itn.HereDoc(`
		load('share', 'sd')
		def act(d):
			d["cnt"] = d.get("cnt", 100) + 1
		sd.perform(act)
	`)
	s2 := itn.HereDoc(`
		load('share', 'sd')
		assert.eq(sd["cnt"], 200)
	`)

	// concurrent access to shared dict
	var (
		sd = NewSharedDict()
		wg sync.WaitGroup
	)
	for i := 1; i <= 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			res, err := itn.ExecModuleWithErrorTest(t, "share", getSDLoader("sd", sd), s1, "", nil)
			if err != nil {
				t.Errorf("sd concurrent error: %v, res: %v", err, res)
			}
		}()
	}
	wg.Wait()

	// check the result
	res, err := itn.ExecModuleWithErrorTest(t, "share", getSDLoader("sd", sd), s2, "", nil)
	if err != nil {
		t.Errorf("sd check error: %v, res: %v", err, res)
	}
}

func TestSharedDict_ToJSON(t *testing.T) {
	sd0 := NewSharedDict()
	sd1 := NewSharedDict()
	if err := sd1.SetKey(starlark.String("foo"), starlark.String("bar")); err != nil {
		t.Errorf("set key error: %v", err)
		return
	}
	sd2 := NewSharedDict()
	if err := sd2.SetKey(starlark.String("foo"), starlark.String("bar")); err != nil {
		t.Errorf("set key error: %v", err)
		return
	}
	if err := sd2.SetKey(starlark.String("cat"), starlark.MakeInt(100)); err != nil {
		t.Errorf("set key error: %v", err)
		return
	}
	sd3 := NewSharedDict()
	if err := sd3.SetKey(starlark.String("self"), sd3.dict); err != nil {
		t.Errorf("set key error: %v", err)
		return
	}

	tests := []struct {
		name    string
		sd      *SharedDict
		wantStr string
		wantErr string
	}{
		{
			name:    "empty",
			sd:      sd0,
			wantStr: `{}`,
		},
		{
			name:    "one",
			sd:      sd1,
			wantStr: `{"foo":"bar"}`,
		},
		{
			name:    "two",
			sd:      sd2,
			wantStr: `{"cat":100,"foo":"bar"}`,
		},
		{
			name:    "cycle",
			sd:      sd3,
			wantErr: `json.encode: in dict key "self": cycle in JSON structure`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := tt.sd.ToJSON()
			if (tt.wantErr == "") && (tt.wantStr != res) {
				t.Errorf("sd(%q).ToJSON() expects '%v', actual '%v'", tt.name, tt.wantStr, res)
				return
			}
			if (err != nil) != (tt.wantErr != "") {
				t.Errorf("sd(%q).ToJSON() expects error = '%v', actual error = '%v', result = %v", tt.name, tt.wantErr, err, res)
				return
			}
		})
	}
}

func getSDLoader(name string, sd *SharedDict) func() (starlark.StringDict, error) {
	md := NewSharedDict()
	if err := md.SetKey(starlark.String("your"), starlark.String("name")); err != nil {
		panic(err)
	}
	md2 := NewSharedDict()
	if err := md2.SetKey(starlark.String("my"), starlark.String("heart")); err != nil {
		panic(err)
	}
	md2.Freeze()
	md3 := NewSharedDict()
	md3.dict = nil
	return func() (starlark.StringDict, error) {
		return starlark.StringDict{
			name:      sd,
			"another": md,
			"frozen":  md2,
			"broken":  md3,
		}, nil
	}
}

func TestSharedDict_LoadJSON(t *testing.T) {
	sd0 := NewSharedDict()
	sd1 := NewSharedDict()
	if err := sd1.SetKey(starlark.String("foo"), starlark.String("bar")); err != nil {
		t.Errorf("set key error: %v", err)
		return
	}
	sd2 := NewSharedDict()
	if err := sd2.SetKey(starlark.String("foo"), starlark.String("bar")); err != nil {
		t.Errorf("set key error: %v", err)
		return
	}
	if err := sd2.SetKey(starlark.String("cat"), starlark.MakeInt(100)); err != nil {
		t.Errorf("set key error: %v", err)
		return
	}
	sd3 := NewSharedDict()
	if err := sd3.SetKey(starlark.String("self"), sd3.dict); err != nil {
		t.Errorf("set key error: %v", err)
		return
	}
	if err := sd3.SetKey(starlark.String("foo"), starlark.String("bar")); err != nil {
		t.Errorf("set key error: %v", err)
		return
	}
	sd4 := NewSharedDict()
	if err := sd4.SetKey(starlark.String("foo"), starlark.String("bar")); err != nil {
		t.Errorf("set key error: %v", err)
		return
	}
	if err := sd4.SetKey(starlark.String("123"), starlark.String("456")); err != nil {
		t.Errorf("set key error: %v", err)
		return
	}
	if err := sd4.SetKey(starlark.String("cat"), starlark.Bool(true)); err != nil {
		t.Errorf("set key error: %v", err)
		return
	}

	sdA := NewSharedDict()
	if err := sdA.SetKey(starlark.String("foo"), starlark.String("ball")); err != nil {
		t.Errorf("set key error: %v", err)
		return
	}
	sdB := NewSharedDict()
	if err := sdB.SetKey(starlark.String("self"), sd3.dict); err != nil {
		t.Errorf("set key error: %v", err)
		return
	}
	sdC := NewSharedDict()
	if err := sdC.SetKey(starlark.String("foo"), starlark.String("okay")); err != nil {
		t.Errorf("set key error: %v", err)
		return
	}
	if err := sdC.SetKey(starlark.String("123"), starlark.String("456")); err != nil {
		t.Errorf("set key error: %v", err)
		return
	}
	sdD := NewSharedDict()
	if err := sdD.SetKey(starlark.String("cat"), starlark.Bool(true)); err != nil {
		t.Errorf("set key error: %v", err)
		return
	}
	sdD.Freeze()

	tests := []struct {
		name    string
		startSD *SharedDict
		jsonStr string
		wantSD  *SharedDict
		wantErr string
	}{
		{
			name:    "empty->empty",
			jsonStr: `{}`,
			startSD: NewSharedDict(),
			wantSD:  sd0,
		},
		{
			name:    "empty->one",
			jsonStr: `{"foo":"bar"}`,
			startSD: NewSharedDict(),
			wantSD:  sd1,
		},
		{
			name:    "nil->one",
			jsonStr: `{"foo":"bar"}`,
			startSD: nil,
			wantSD:  sd1,
			wantErr: `nil shared dict`,
		},
		{
			name:    "one override",
			jsonStr: `{"foo":"bar"}`,
			startSD: sdA,
			wantSD:  sd1,
		},
		{
			name:    "cycle",
			jsonStr: `{"foo":"bar"}`,
			startSD: sdB,
			wantSD:  sd3,
		},
		{
			name:    "two->three",
			jsonStr: `{"cat":true,"foo":"bar"}`,
			startSD: sdC,
			wantSD:  sd4,
		},
		{
			name:    "frozen",
			jsonStr: `{"foo":"bar"}`,
			startSD: sdD,
			wantErr: `cannot insert into frozen hash table`,
		},
		{
			name:    "invalid json",
			jsonStr: `{"foo":"bar"`,
			startSD: NewSharedDict(),
			wantErr: `json.decode: at offset 12, unexpected end of file`,
		},
		{
			name:    "invalid type",
			jsonStr: `123`,
			startSD: NewSharedDict(),
			wantErr: `got int result, want dict`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.startSD.LoadJSON(tt.jsonStr)
			if (err != nil) && (tt.wantErr == "") {
				t.Errorf("sd(%q).LoadJSON() expects no error, actual error = '%v'", tt.name, err)
				return
			}
			if err == nil {
				if !reflect.DeepEqual(tt.startSD, tt.wantSD) {
					t.Errorf("sd(%q).LoadJSON() expects '%v', actual '%v'", tt.name, tt.wantSD, tt.startSD)
					return
				}
			}
		})
	}
}

// Shared by the baseline validation copy and candidate. Native children remain
// mutable and owner-exclusive during these single-threaded success-path runs.
var (
	sharedDictOutputSafetyStringSink string
	sharedDictOutputSafetyDictSink   *SharedDict
	sharedDictOutputSafetyValueSink  starlark.Value
)

func BenchmarkSharedDictOutputSafety(b *testing.B) {
	put := func(dict *SharedDict, key string, value starlark.Value) {
		if err := dict.SetKey(starlark.String(key), value); err != nil {
			b.Fatal(err)
		}
	}
	flat := func(size int) (*SharedDict, string, string) {
		dict := NewSharedDict()
		text, encoded := make([]string, size), make([]string, size)
		for i := 0; i < size; i++ {
			key := fmt.Sprintf("k%d", i)
			put(dict, key, starlark.MakeInt(i))
			text[i] = fmt.Sprintf("%q: %d", key, i)
			encoded[i] = fmt.Sprintf("%q:%d", key, i)
		}
		return dict, "shared_dict({" + strings.Join(text, ", ") + "})", "{" + strings.Join(encoded, ",") + "}"
	}
	flat0, text0, _ := flat(0)
	flat8, text8, json8 := flat(8)
	flat64, text64, _ := flat(64)
	values, numbers := make([]starlark.Value, 32), make([]string, 32)
	set := starlark.NewSet(32)
	for i := range values {
		values[i] = starlark.MakeInt(i)
		numbers[i] = fmt.Sprint(i)
		if err := set.Insert(values[i]); err != nil {
			b.Fatal(err)
		}
	}
	list := starlark.NewList(values)
	setDict, listDict := NewSharedDict(), NewSharedDict()
	put(setDict, "items", set)
	put(listDict, "items", list)
	setText := "shared_dict({\"items\": set([" + strings.Join(numbers, ", ") + "])})"
	sequenceJSON := "{\"items\":[" + strings.Join(numbers, ",") + "]}"
	for _, test := range []struct {
		name     string
		dict     *SharedDict
		want     string
		snapshot bool
		json     bool
	}{
		{name: "StringFlat0", dict: flat0, want: text0},
		{name: "StringFlat8", dict: flat8, want: text8},
		{name: "StringFlat64", dict: flat64, want: text64},
		{name: "StringSet32", dict: setDict, want: setText},
		{name: "SnapshotSet32", dict: setDict, want: setText, snapshot: true},
		{name: "JSONFlat8", dict: flat8, want: json8, json: true},
		{name: "JSONNestedList32", dict: listDict, want: sequenceJSON, json: true},
		{name: "JSONSet32", dict: setDict, want: sequenceJSON, json: true},
	} {
		test := test
		b.Run(test.name, func(b *testing.B) {
			b.ReportAllocs()
			switch {
			case test.json:
				if got, err := test.dict.ToJSON(); err != nil || got != test.want {
					b.Fatalf("ToJSON() = %q, %v; want %q, nil", got, err, test.want)
				}
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					got, err := test.dict.ToJSON()
					if err != nil {
						b.Fatal(err)
					}
					sharedDictOutputSafetyStringSink = got
				}
			case test.snapshot:
				if got := test.dict.SnapshotString(); got != test.want {
					b.Fatalf("SnapshotString() = %q; want %q", got, test.want)
				}
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					sharedDictOutputSafetyStringSink = test.dict.SnapshotString()
				}
			default:
				if got := test.dict.String(); got != test.want {
					b.Fatalf("String() = %q; want %q", got, test.want)
				}
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					sharedDictOutputSafetyStringSink = test.dict.String()
				}
			}
		})
	}
	b.Run("NewShared", func(b *testing.B) {
		if dict := NewSharedDict(); dict.Len() != 0 || dict.Type() != "shared_dict" {
			b.Fatal("unexpected default constructor result")
		}
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			sharedDictOutputSafetyDictSink = NewSharedDict()
		}
	})
	b.Run("NewNamed", func(b *testing.B) {
		if dict := NewNamedSharedDict("bench"); dict.Len() != 0 || dict.Type() != "bench" {
			b.Fatal("unexpected named constructor result")
		}
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			sharedDictOutputSafetyDictSink = NewNamedSharedDict("bench")
		}
	})
	b.Run("SetTypeName", func(b *testing.B) {
		dict := NewSharedDict()
		names := [2]string{"bench_a", "bench_b"}
		sharedDictOutputSafetyDictSink = dict
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			dict.SetTypeName(names[i&1])
		}
		b.StopTimer()
		if b.N > 0 && dict.Type() != names[(b.N-1)&1] {
			b.Fatal("SetTypeName did not store the final name")
		}
	})
	b.Run("SetKeyExisting", func(b *testing.B) {
		dict := NewSharedDict()
		key := starlark.String("existing")
		values := [2]starlark.Value{starlark.MakeInt(0), starlark.MakeInt(1)}
		if err := dict.SetKey(key, values[0]); err != nil {
			b.Fatal(err)
		}
		sharedDictOutputSafetyDictSink = dict
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if err := dict.SetKey(key, values[i&1]); err != nil {
				b.Fatal(err)
			}
		}
	})
	b.Run("GetExisting", func(b *testing.B) {
		dict := NewSharedDict()
		key, want := starlark.String("existing"), starlark.MakeInt(1)
		if err := dict.SetKey(key, want); err != nil {
			b.Fatal(err)
		}
		if got, found, err := dict.Get(key); err != nil || !found || got != want {
			b.Fatalf("Get() = %v, %v, %v; want %v, true, nil", got, found, err, want)
		}
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			got, found, err := dict.Get(key)
			if err != nil || !found {
				b.Fatalf("Get() found=%v, err=%v", found, err)
			}
			sharedDictOutputSafetyValueSink = got
		}
	})
}
