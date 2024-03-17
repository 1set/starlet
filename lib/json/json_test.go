package json_test

import (
	"testing"

	itn "github.com/1set/starlet/internal"
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
			name: `dumps(def)`,
			script: itn.HereDoc(`
				load('json', 'dumps')
				def f(): pass
				d = { "a" : "b", "f" : f}
				dumps(d)
			`),
			wantErr: `unmarshaling starlark value: unrecognized starlark type: *starlark.Function`,
		},
		{
			name: `dumps(dict)`,
			script: itn.HereDoc(`
				load('json', 'dumps')
				d = { "a" : "b", "c" : "d"}
				s = '''{"a":"b","c":"d"}'''
				assert.eq(dumps(d), s)
			`),
		},
		{
			name: `dumps(dict, indent=0)`,
			script: itn.HereDoc(`
				load('json', 'dumps')
				d = { "a" : "b", "c" : "d"}
				s = '''{"a":"b","c":"d"}'''
				assert.eq(dumps(d, indent=0), s)
			`),
		},
		{
			name: `dumps(dict, indent=-2)`,
			script: itn.HereDoc(`
				load('json', 'dumps')
				d = { "a" : "b", "c" : "d"}
				s = '''{"a":"b","c":"d"}'''
				assert.eq(dumps(d, indent=-2), s)
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
		{
			name: `dumps struct`,
			script: itn.HereDoc(`
				load('json', 'dumps')
				load("struct.star", "struct")
				s = struct(a="Aloha", b=0x10, c=True, d=[1,2,3])
				d = '{"a":"Aloha","b":16,"c":true,"d":[1,2,3]}'
				assert.eq(dumps(s), d)
			`),
		},
		{
			name: `dumps module`,
			script: itn.HereDoc(`
				load('json', 'dumps')
				load("module.star", "module")
				m = module("hello", a="Bravo", b=0b10, c=False, d=[7,8,9])
				d = '{"a":"Bravo","b":2,"c":false,"d":[7,8,9]}'
				assert.eq(dumps(m), d)
			`),
		},
		{
			name: `encode struct`,
			script: itn.HereDoc(`
				load('json', 'encode')
				load("struct.star", "struct")
				s = struct(a="Aloha", b=0x10, c=True, d=[1,2,3])
				d = '{"a":"Aloha","b":16,"c":true,"d":[1,2,3]}'
				assert.eq(encode(s), d)
			`),
		},
		{
			name: `encode module`,
			script: itn.HereDoc(`
				load('json', 'encode')
				load("module.star", "module")
				m = module("hello", a="Bravo", b=0b10, c=False, d=[7,8,9])
				d = '{"a":"Bravo","b":2,"c":false,"d":[7,8,9]}'
				assert.eq(encode(m), d)
			`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := itn.ExecModuleWithErrorTest(t, json.ModuleName, json.LoadModule, tt.script, tt.wantErr, nil)
			if (err != nil) != (tt.wantErr != "") {
				t.Errorf("json(%q) expects error = '%v', actual error = '%v', result = %v", tt.name, tt.wantErr, err, res)
				return
			}
		})
	}
}
