package goidiomatic_test

import (
	"errors"
	"testing"

	"github.com/1set/starlet/lib/goidiomatic"
	itn "github.com/1set/starlet/lib/internal"
)

func TestLoadModule_GoIdiomatic(t *testing.T) {
	header := `load('assert.star', 'assert')`
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
			wantErr: errors.New(`starlet runtime system exit`),
		},
		{
			name: `exit 0`,
			script: itn.HereDoc(`
				load('go_idiomatic', 'exit')
				exit(0)
			`),
			wantErr: errors.New(`starlet runtime system exit`),
		},
		{
			name: `exit 1`,
			script: itn.HereDoc(`
				load('go_idiomatic', 'exit')
				exit(1)
			`),
			wantErr: errors.New(`starlet runtime system exit`),
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
			wantErr: errors.New(`starlet runtime system exit`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := itn.ExecModuleWithErrorTest(t, goidiomatic.ModuleName, goidiomatic.LoadModule, header+"\n"+tt.script, tt.wantErr)
			if (err != nil) != (tt.wantErr != nil) {
				t.Errorf("go_idiomatic(%q) expects error = '%v', actual error = '%v', result = %v", tt.name, tt.wantErr, err, res)
				return
			}
		})
	}
}
