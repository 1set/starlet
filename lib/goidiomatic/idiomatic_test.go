package goidiomatic_test

import (
	"testing"

	"github.com/1set/starlet/lib/goidiomatic"
	itn "github.com/1set/starlet/lib/internal"
	"github.com/1set/starlight/convert"
	"go.starlark.net/starlark"
)

func TestLoadModule_GoIdiomatic(t *testing.T) {
	// test cases
	tests := []struct {
		name    string
		script  string
		wantErr string
	}{
		{
			name: `boolean`,
			script: itn.HereDoc(`
				load('go_idiomatic', 'true', 'false')
				assert.eq(true, True)
				assert.eq(false, False)
			`),
		},
		{
			name: `nil`,
			script: itn.HereDoc(`
				load('go_idiomatic', 'nil')
				assert.eq(nil, None)
			`),
		},
		{
			name: `sleep`,
			script: itn.HereDoc(`
				load('go_idiomatic', 'sleep')
				sleep()
			`),
			wantErr: `sleep: missing argument for secs`,
		},
		{
			name: `sleep 0`,
			script: itn.HereDoc(`
				load('go_idiomatic', 'sleep')
				sleep(0)
			`),
		},
		{
			name: `sleep 10ms`,
			script: itn.HereDoc(`
				load('go_idiomatic', 'sleep')
				sleep(0.01)
			`),
		},
		{
			name: `sleep negative`,
			script: itn.HereDoc(`
				load('go_idiomatic', 'sleep')
				sleep(-1)
			`),
			wantErr: `secs must be non-negative`,
		},
		{
			name: `sleep hello`,
			script: itn.HereDoc(`
				load('go_idiomatic', 'sleep')
				sleep('hello')
			`),
			wantErr: `sleep: for parameter secs: got string, want float or int`,
		},
		{
			name: `sleep none`,
			script: itn.HereDoc(`
				load('go_idiomatic', 'sleep')
				sleep(None)
			`),
			wantErr: `sleep: for parameter secs: got NoneType, want float or int`,
		},
		{
			name: `exit`,
			script: itn.HereDoc(`
				load('go_idiomatic', 'exit')
				exit()
			`),
			wantErr: `starlet runtime system exit (Use Ctrl-D in REPL to exit)`,
		},
		{
			name: `exit 0`,
			script: itn.HereDoc(`
				load('go_idiomatic', 'exit')
				exit(0)
			`),
			wantErr: `starlet runtime system exit (Use Ctrl-D in REPL to exit)`,
		},
		{
			name: `exit 1`,
			script: itn.HereDoc(`
				load('go_idiomatic', 'exit')
				exit(1)
			`),
			wantErr: `starlet runtime system exit (Use Ctrl-D in REPL to exit)`,
		},
		{
			name: `exit -1`,
			script: itn.HereDoc(`
				load('go_idiomatic', 'exit')
				exit(-1)
			`),
			wantErr: `exit: for parameter code: -1 out of range (want value in unsigned 8-bit range)`,
		},
		{
			name: `exit hello`,
			script: itn.HereDoc(`
				load('go_idiomatic', 'exit')
				exit('hello')
			`),
			wantErr: `exit: for parameter code: got string, want int`,
		},
		{
			name: `quit`,
			script: itn.HereDoc(`
				load('go_idiomatic', 'quit')
				quit()
			`),
			wantErr: `starlet runtime system exit (Use Ctrl-D in REPL to exit)`,
		},
		{
			name: `length(string)`,
			script: itn.HereDoc(`
				load('go_idiomatic', 'length')
				assert.eq(0, length(''))
				assert.eq(1, length('a'))
				assert.eq(2, length('ab'))
				assert.eq(3, length('abc'))
				assert.eq(1, length('🙆'))
				assert.eq(1, length('✅'))
				assert.eq(3, length('水光肌'))
			`),
		},
		{
			name: `length(bytes)`,
			script: itn.HereDoc(`
				load('go_idiomatic', 'length')
				assert.eq(0, length(b''))
				assert.eq(9, length(b'水光肌'))
				assert.eq(7, length(b'🙆✅'))
			`),
		},
		{
			name: `length(list/tuple/dict/set)`,
			script: itn.HereDoc(`
				load('go_idiomatic', 'length')
				assert.eq(0, length([]))
				assert.eq(5, length([1, 2, "#", True, None]))
				assert.eq(1, length((3,)))
				assert.eq(2, length((1, 2)))
				assert.eq(0, length({}))
				assert.eq(2, length({'a': 1, 'b': 2}))
				assert.eq(0, length(set()))
				assert.eq(2, length(set(['a', 'b'])))
			`),
		},
		{
			name: `length(slice)`,
			script: itn.HereDoc(`
				load('go_idiomatic', 'length')
				assert.eq(3, length(slice))
			`),
		},
		{
			name: `length(map)`,
			script: itn.HereDoc(`
				load('go_idiomatic', 'length')
				assert.eq(2, length(map))
			`),
		},
		{
			name: `length()`,
			script: itn.HereDoc(`
				load('go_idiomatic', 'length')
				length()
			`),
			wantErr: `length() takes exactly one argument (0 given)`,
		},
		{
			name: `length(*2)`,
			script: itn.HereDoc(`
				load('go_idiomatic', 'length')
				length('a', 'b')
			`),
			wantErr: `length() takes exactly one argument (2 given)`,
		},
		{
			name: `length(bool)`,
			script: itn.HereDoc(`
				load('go_idiomatic', 'length')
				length(True)
			`),
			wantErr: `object of type 'bool' has no length()`,
		},
		{
			name: `sum()`,
			script: itn.HereDoc(`
				load('go_idiomatic', 'sum')
				sum()
			`),
			wantErr: `sum() takes at least 1 positional argument (0 given)`,
		},
		{
			name: `sum(>2)`,
			script: itn.HereDoc(`
				load('go_idiomatic', 'sum')
				sum(1, 2, 3)
			`),
			wantErr: `sum() takes at most 2 arguments (3 given)`,
		},
		{
			name: `sum(int)`,
			script: itn.HereDoc(`
				load('go_idiomatic', 'sum')
				sum(1)
			`),
			wantErr: `object of type 'int' is not iterable`,
		},
		{
			name: `sum([int])`,
			script: itn.HereDoc(`
				load('go_idiomatic', 'sum')
				assert.eq(6, sum([1, 2, 3]))
			`),
		},
		{
			name: `sum([float])`,
			script: itn.HereDoc(`
				load('go_idiomatic', 'sum')
				assert.eq(6, sum([1.0, 2.0, 3.0]))
			`),
		},
		{
			name: `sum([int,float])`,
			script: itn.HereDoc(`
				load('go_idiomatic', 'sum')
				assert.eq(7, sum([1, 2.0, 4]))
			`),
		},
		{
			name: `sum([int,string])`,
			script: itn.HereDoc(`
				load('go_idiomatic', 'sum')
				sum([1, 'a'])
			`),
			wantErr: `got string, want float or int`,
		},
		{
			name: `sum([int], float)`,
			script: itn.HereDoc(`
				load('go_idiomatic', 'sum')
				assert.eq(15, sum([1, 2, 4], 8.0))
			`),
		},
		{
			name: `sum([int], start=int)`,
			script: itn.HereDoc(`
				load('go_idiomatic', 'sum')
				assert.eq(15, sum([1, 2, 4], start=8))
			`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// prepare
			s, err := convert.ToValue([]string{"a", "b", "c"})
			if err != nil {
				t.Errorf("convert.ToValue Slice: %v", err)
				return
			}
			m, err := convert.ToValue(map[string]string{"a": "b", "c": "d"})
			if err != nil {
				t.Errorf("convert.ToValue Map: %v", err)
				return
			}
			starlark.Universe["slice"] = s
			starlark.Universe["map"] = m

			res, err := itn.ExecModuleWithErrorTest(t, goidiomatic.ModuleName, goidiomatic.LoadModule, tt.script, tt.wantErr)
			if (err != nil) != (tt.wantErr != "") {
				t.Errorf("go_idiomatic(%q) expects error = '%v', actual error = '%v', result = %v", tt.name, tt.wantErr, err, res)
				return
			}
		})
	}
}
