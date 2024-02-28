package log_test

import (
	"testing"

	itn "github.com/1set/starlet/internal"
	lg "github.com/1set/starlet/lib/log"
)

func TestLoadModule_Log(t *testing.T) {
	tests := []struct {
		name    string
		script  string
		wantErr string
	}{
		{
			name: `debug message`,
			script: itn.HereDoc(`
				load('log', 'debug')
				debug('this is a debug message only')
			`),
		},
		{
			name: `debug with no args`,
			script: itn.HereDoc(`
				load('log', 'debug')
				debug()
			`),
			wantErr: "log.debug: expected at least 1 argument, got 0",
		},
		{
			name: `debug with invalid arg type`,
			script: itn.HereDoc(`
				load('log', 'debug')
				debug(520)
			`),
			wantErr: "log.debug: expected string as first argument, got int",
		},
		{
			name: `debug with key values`,
			script: itn.HereDoc(`
				load('log', 'debug')
				m = {"mm": "this is more"}
				l = [2, "LIST", 3.14, True]
				debug('this is a debug message', "map", m, "list", l)
			`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := itn.ExecModuleWithErrorTest(t, lg.ModuleName, lg.LoadModule, tt.script, tt.wantErr)
			if (err != nil) != (tt.wantErr != "") {
				t.Errorf("log(%q) expects error = '%v', actual error = '%v', result = %v", tt.name, tt.wantErr, err, res)
				return
			}
		})
	}
}
