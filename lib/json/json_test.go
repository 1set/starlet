package json_test

import (
	"testing"

	itn "github.com/1set/starlet/lib/internal"
	"github.com/1set/starlet/lib/json"
)

func TestLoadModule_JSON(t *testing.T) {
	tests := []struct {
		name    string
		script  string
		wantErr string
	}{
		{
			name: `stdlib can be loaded`,
			script: itn.HereDoc(`
				load('json', 'dumps', 'encode', 'decode', 'indent')
			`),
		},
		{
			name: `dumps()`,
			script: itn.HereDoc(`
				load('json', 'dumps')
				dumps()
			`),
			wantErr: `json.dumps: missing argument for obj`,
		},
		{
			name: `dumps(>2)`,
			script: itn.HereDoc(`
				load('json', 'dumps')
				dumps(1, 2, 3)
			`),
			wantErr: `json.dumps: got 3 arguments, want at most 2`,
		},
		{
			name: `dumps(1, "a")`,
			script: itn.HereDoc(`
				load('json', 'dumps')
				dumps(1, "a")
			`),
			wantErr: `json.dumps: for parameter indent: got string, want int`,
		},
		{
			name: `dumps(1)`,
			script: itn.HereDoc(`
				load('json', 'dumps')
				assert.eq(dumps(1), '1')
			`),
		},
		{
			name: `dumps(1, indent=2)`,
			script: itn.HereDoc(`
				load('json', 'dumps')
				assert.eq(dumps(1, indent=2), '1')
			`),
		},
		{
			name: `dumps(dict, indent=2)`,
			script: itn.HereDoc(`
				load('json', 'dumps')
				d = { "a" : "b", "c" : "d"}
				s = '''
				{
				  "a": "b",
				  "c": "d"
				}
				'''.strip()
				assert.eq(dumps(d, indent=2), s)
			`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := itn.ExecModuleWithErrorTest(t, json.ModuleName, json.LoadModule, tt.script, tt.wantErr)
			if (err != nil) != (tt.wantErr != "") {
				t.Errorf("json(%q) expects error = '%v', actual error = '%v', result = %v", tt.name, tt.wantErr, err, res)
				return
			}
		})
	}
}
