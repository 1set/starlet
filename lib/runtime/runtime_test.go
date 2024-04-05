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
				load('runtime', 'hostname', 'workdir', 'os', 'arch', 'gover')
				ss = [hostname, workdir, os, arch, gover]
				print(ss)
				_ = [assert.eq(type(s), "string") for s in ss]
			`),
		},
		{
			name: `process`,
			script: itn.HereDoc(`
				load('runtime', 'pid', 'ppid', 'uid', 'gid')
				si = [pid, ppid, uid, gid]
				print(si)
				_ = [assert.eq(type(s), "int") for s in si]
			`),
		},
		{
			name: `app`,
			script: itn.HereDoc(`
				load('runtime', s='app_start', ut='uptime')
				assert.eq(type(s), "time.time")
				u = ut()
				assert.eq(type(u), "time.duration")
				print(s, u)
			`),
		},
		{
			name: `uptime invalid`,
			script: itn.HereDoc(`
				load('runtime', 'uptime')
				uptime(123)
			`),
			wantErr: `runtime.uptime: got 1 arguments, want 0`,
		},
		{
			name: `getenv: no args`,
			script: itn.HereDoc(`
				load('runtime', 'getenv')
				getenv()
			`),
			wantErr: `runtime.getenv: missing argument for key`,
		},
		{
			name: `getenv: invalid`,
			script: itn.HereDoc(`
				load('runtime', 'getenv')
				getenv(123)
			`),
			wantErr: `runtime.getenv: for parameter key: got int, want string`,
		},
		{
			name: `getenv: no result`,
			script: itn.HereDoc(`
				load('runtime', 'getenv')
				x = getenv("very-long-long-non-existent")
				assert.eq(x, None)
				y = getenv("very-long-long-non-existent", 1000)
				assert.eq(y, 1000)
			`),
		},
		{
			name: `getenv: with result`,
			script: itn.HereDoc(`
				load('runtime', 'getenv')
				x = getenv("PATH")
				print("PATH:", x)
				assert.eq(type(x), "string")
			`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := itn.ExecModuleWithErrorTest(t, rt.ModuleName, rt.LoadModule, tt.script, tt.wantErr, nil)
			if (err != nil) != (tt.wantErr != "") {
				t.Errorf("runtime(%q) expects error = '%v', actual error = '%v', result = %v", tt.name, tt.wantErr, err, res)
				return
			}
		})
	}
}
