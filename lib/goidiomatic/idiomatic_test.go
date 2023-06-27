package goidiomatic_test

import (
	"errors"
	"testing"

	"github.com/1set/starlet/lib/goidiomatic"
	itn "github.com/1set/starlet/lib/internal"
)

func TestLoadModule_GoIdiomatic(t *testing.T) {
	tests := []struct {
		name    string
		script  string
		wantErr error
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
			wantErr: errors.New(`sleep: missing argument for secs`),
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
			wantErr: errors.New(`secs must be non-negative`),
		},
		{
			name: `sleep hello`,
			script: itn.HereDoc(`
				load('go_idiomatic', 'sleep')
				sleep('hello')
			`),
			wantErr: errors.New(`sleep: for parameter secs: got string, want float or int`),
		},
		{
			name: `sleep none`,
			script: itn.HereDoc(`
				load('go_idiomatic', 'sleep')
				sleep(None)
			`),
			wantErr: errors.New(`sleep: for parameter secs: got NoneType, want float or int`),
		},
		{
			name: `exit`,
			script: itn.HereDoc(`
				load('go_idiomatic', 'exit')
				exit()
			`),
			wantErr: errors.New(`starlet runtime system exit (Use Ctrl-D in REPL to exit)`),
		},
		{
			name: `exit 0`,
			script: itn.HereDoc(`
				load('go_idiomatic', 'exit')
				exit(0)
			`),
			wantErr: errors.New(`starlet runtime system exit (Use Ctrl-D in REPL to exit)`),
		},
		{
			name: `exit 1`,
			script: itn.HereDoc(`
				load('go_idiomatic', 'exit')
				exit(1)
			`),
			wantErr: errors.New(`starlet runtime system exit (Use Ctrl-D in REPL to exit)`),
		},
		{
			name: `exit -1`,
			script: itn.HereDoc(`
				load('go_idiomatic', 'exit')
				exit(-1)
			`),
			wantErr: errors.New(`exit: for parameter code: -1 out of range (want value in unsigned 8-bit range)`),
		},
		{
			name: `exit hello`,
			script: itn.HereDoc(`
				load('go_idiomatic', 'exit')
				exit('hello')
			`),
			wantErr: errors.New(`exit: for parameter code: got string, want int`),
		},
		{
			name: `quit`,
			script: itn.HereDoc(`
				load('go_idiomatic', 'quit')
				quit()
			`),
			wantErr: errors.New(`starlet runtime system exit (Use Ctrl-D in REPL to exit)`),
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
			name: `length(list)`,
			script: itn.HereDoc(`
				load('go_idiomatic', 'length')
				assert.eq(0, length([]))
				assert.eq(5, length([1, 2, "#", True, None]))
			`),
		},
		{
			name: `length()`,
			script: itn.HereDoc(`
				load('go_idiomatic', 'length')
				length()
			`),
			wantErr: errors.New(`length() takes exactly one argument (0 given)`),
		},
		{
			name: `length(*2)`,
			script: itn.HereDoc(`
				load('go_idiomatic', 'length')
				length('a', 'b')
			`),
			wantErr: errors.New(`length() takes exactly one argument (2 given)`),
		},
		{
			name: `length(bool)`,
			script: itn.HereDoc(`
				load('go_idiomatic', 'length')
				length(True)
			`),
			wantErr: errors.New(`object of type 'bool' has no length()`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := itn.ExecModuleWithErrorTest(t, goidiomatic.ModuleName, goidiomatic.LoadModule, tt.script, tt.wantErr)
			if (err != nil) != (tt.wantErr != nil) {
				t.Errorf("go_idiomatic(%q) expects error = '%v', actual error = '%v', result = %v", tt.name, tt.wantErr, err, res)
				return
			}
		})
	}
}
