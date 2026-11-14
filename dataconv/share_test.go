package dataconv

// SharedDict tests cover constructors/names, script methods, frozen/nil values,
// concurrent access, caller-synchronized output, snapshots, and JSON conversion.

import (
	"fmt"
	"math"
	"reflect"
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
			if !strings.HasPrefix(got, `shared_dict({"counter": `) || !strings.HasSuffix(got, "})") {
				t.Errorf("malformed concurrent String: %q", got)
				return
			}
		}
	}()
	close(start)
	workers.Wait()
}

// This verifies the caller's synchronization discipline; String and ToJSON do
// not acquire the application mutex or become safe for unsynchronized callers.
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

func TestSharedDict_StringLegacyReentry(t *testing.T) {
	for _, test := range []string{"external_lock", "Hash", "Equal", "Freeze"} {
		t.Run(test, func(t *testing.T) {
			sd := NewNamedSharedDict("named")
			var pending []chan string
			checks := 0
			checkString := func(want string) {
				checks++
				done := make(chan string, 1)
				go func() { done <- sd.String() }()
				select {
				case got := <-done:
					if got != want {
						t.Errorf("String while locked = %q; want %q", got, want)
					}
				case <-time.After(2 * time.Second):
					t.Error("legacy String unexpectedly waited for the mapping lock")
					// Return from the callback so the operation unlocks, then
					// drain the worker below even if a regression blocks it.
					pending = append(pending, done)
				}
			}
			key := &sharedDictFormatValue{stringFunc: func() string { return "key" }}
			switch test {
			case "external_lock":
				sd.Lock()
				checkString("named({})")
				sd.Unlock()
			case "Hash":
				key.hashFunc = func() (uint32, error) { checkString("named({})"); return 1, nil }
				if err := sd.SetKey(key, starlark.True); err != nil {
					t.Fatal(err)
				}
			case "Equal":
				if err := sd.SetKey(key, starlark.True); err != nil {
					t.Fatal(err)
				}
				other := &sharedDictFormatValue{stringFunc: func() string { return "other" }}
				other.compareFunc = func() { checkString("named({key: True})") }
				if err := sd.SetKey(other, starlark.False); err != nil {
					t.Fatal(err)
				}
			case "Freeze":
				key.freezeFunc = func() { checkString(`named({"value": key})`) }
				if err := sd.SetKey(starlark.String("value"), key); err != nil {
					t.Fatal(err)
				}
				sd.Freeze()
			}
			for _, done := range pending {
				select {
				case <-done:
				case <-time.After(2 * time.Second):
					t.Fatal("formatting worker remained blocked after the operation unlocked")
				}
			}
			if checks != 1 {
				t.Fatalf("String compatibility callback count = %d; want 1", checks)
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
	} {
		b.Run("Snapshot/"+test.name, func(b *testing.B) {
			format := test.dict.String
			if snapshot, ok := interface{}(test.dict).(interface{ SnapshotString() string }); ok {
				format = snapshot.SnapshotString
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
				sd.len()
			`),
			wantErr: `shared_dict has no .len field or method`,
		},
		{
			name: `to json`,
			script: itn.HereDoc(`
				load('share', 'sd')
				assert.eq(sd.to_json(), '{}')
			`),
			wantErr: `shared_dict has no .to_json field or method`,
		},
		{
			name: `attr names`,
			script: itn.HereDoc(`
				load('share', 'sd')
				assert.eq(dir(sd), [])
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
