package dataconv

import (
	"sync"
	"testing"

	itn "github.com/1set/starlet/internal"
	"go.starlark.net/starlark"
)

func getSDLoader(name string, sd *SharedDict) func() (starlark.StringDict, error) {
	md := NewSharedDict()
	if err := md.SetKey(starlark.String("your"), starlark.String("name")); err != nil {
		panic(err)
	}
	return func() (starlark.StringDict, error) {
		return starlark.StringDict{
			name:      sd,
			"another": md,
		}, nil
	}
}

func getDictLoader(name string, sd *starlark.Dict) func() (starlark.StringDict, error) {
	return func() (starlark.StringDict, error) {
		return starlark.StringDict{
			name: sd,
		}, nil
	}
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
			name: `equal check`,
			script: itn.HereDoc(`
				load('share', 'sd', sd3='another')
				sd1 = sd
				sd2 = sd
				assert.true(sd1 == sd2)
				assert.true(sd1 != sd3)

				sd1.clear()
				sd1["your"] = "name"
				assert.true(sd1 == sd2)
				assert.true(sd1 == sd3)
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
				assert.eq(l, ["clear", "get", "items", "keys", "len", "perform", "pop", "popitem", "setdefault", "update", "values"])
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := itn.ExecModuleWithErrorTest(t, "share", getSDLoader("sd", NewSharedDict()), tt.script, tt.wantErr)
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
			name: `equal`,
			script: itn.HereDoc(`
				load('share', 'sd', sd3='another')
				assert.true(sd != sd3)
				sd3.clear()
				sd3["foo"] = "bar"
				assert.true(sd == sd3)
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
			wantErr: `frozen dict`,
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sd := NewSharedDict()
			if err := sd.SetKey(starlark.String("foo"), starlark.String("bar")); err != nil {
				t.Errorf("set key error: %v", err)
				return
			}
			sd.Freeze()
			res, err := itn.ExecModuleWithErrorTest(t, "share", getSDLoader("sd", sd), tt.script, tt.wantErr)
			if (err != nil) != (tt.wantErr != "") {
				t.Errorf("frozen sd(%q) expects error = '%v', actual error = '%v', result = %v", tt.name, tt.wantErr, err, res)
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

			res, err := itn.ExecModuleWithErrorTest(t, "share", getSDLoader("sd", sd), s1, "")
			if err != nil {
				t.Errorf("sd concurrent error: %v, res: %v", err, res)
			}
		}()
	}
	wg.Wait()

	// check the result
	res, err := itn.ExecModuleWithErrorTest(t, "share", getSDLoader("sd", sd), s2, "")
	if err != nil {
		t.Errorf("sd check error: %v, res: %v", err, res)
	}
}
