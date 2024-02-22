package dataconv

import (
	itn "github.com/1set/starlet/internal"
	"go.starlark.net/starlark"
	"testing"
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
