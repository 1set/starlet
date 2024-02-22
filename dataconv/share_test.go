package dataconv

import (
	"testing"

	itn "github.com/1set/starlet/internal"
	"go.starlark.net/starlark"
)

func getSDLoader(name string, sd *SharedDict) func() (starlark.StringDict, error) {
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
			name: `attrs`,
			script: itn.HereDoc(`
				load('share', 'sd')
				l = dir(sd)
				print(l)
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
				sd.clear()
				assert.eq(sd.get("foo"), None)
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