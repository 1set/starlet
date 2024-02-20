package runtime_test

import (
	"testing"

	itn "github.com/1set/starlet/internal"
	rt "github.com/1set/starlet/lib/runtime"
)

func TestLoadModule_Runtime(t *testing.T) {
	tests := []struct {
		name    string
		script  string
		wantErr string
	}{
		{
			name: `host`,
			script: itn.HereDoc(`
				load('runtime', 'hostname', 'workdir', 'os', 'arch')
				ss = [hostname, workdir, os, arch]
				print(ss)
				_ = [assert.eq(type(s), "string") for s in ss]
			`),
		},
		{
			name: `pid`,
			script: itn.HereDoc(`
				load('runtime', 'pid')
				assert.eq(type(pid), "int")
				assert.ne(pid, 0)
				print(pid)
			`),
		},
		{
			name: `app`,
			script: itn.HereDoc(`
				load('runtime', s='app_start', u='app_uptime')
				assert.eq(type(s), "time.time")
				assert.eq(type(u), "time.duration")
				print(s, u)
			`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := itn.ExecModuleWithErrorTest(t, rt.ModuleName, rt.LoadModule, tt.script, tt.wantErr)
			if (err != nil) != (tt.wantErr != "") {
				t.Errorf("runtime(%q) expects error = '%v', actual error = '%v', result = %v", tt.name, tt.wantErr, err, res)
				return
			}
		})
	}
}
