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
				load('runtime', 'hostname', 'workdir', 'homedir', 'os', 'arch', 'gover')
				ss = [hostname, workdir, homedir, os, arch, gover]
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
		{
			name: `putenv: no args`,
			script: itn.HereDoc(`
				load('runtime', 'putenv')
				putenv()	
			`),
			wantErr: `runtime.putenv: missing argument for key`,
		},
		{
			name: `putenv: invalid`,
			script: itn.HereDoc(`
				load('runtime', 'putenv')
				putenv(123, "value")
			`),
			wantErr: `runtime.putenv: for parameter key: got int, want string`,
		},
		{
			name: `putenv: no value`,
			script: itn.HereDoc(`
				load('runtime', 'putenv')
				putenv("key")
			`),
			wantErr: `runtime.putenv: missing argument for value`,
		},
		{
			name: `putenv: new value`,
			script: itn.HereDoc(`
				load('runtime', 'putenv', 'getenv')
				putenv("STARLET_TEST", 123456)
				x = getenv("STARLET_TEST")
				print("STARLET_TEST:", x)
				assert.eq(x, "123456")
			`),
		},
		{
			name: `putenv: existing value`,
			script: itn.HereDoc(`
				load('runtime', 'putenv', 'getenv')
				putenv("STARLET_TEST", 123456)
				putenv("STARLET_TEST", 654321)
				x = getenv("STARLET_TEST")
				print("STARLET_TEST:", x)
				assert.eq(x, "654321")
			`),
		},
		{
			name: `unsetenv: no args`,
			script: itn.HereDoc(`
				load('runtime', 'unsetenv')
				unsetenv()
			`),
			wantErr: `runtime.unsetenv: missing argument for key`,
		},
		{
			name: `unsetenv: invalid`,
			script: itn.HereDoc(`
				load('runtime', 'unsetenv')
				unsetenv(123)
			`),
			wantErr: `runtime.unsetenv: for parameter key: got int, want string`,
		},
		{
			name: `unsetenv: non-existent`,
			script: itn.HereDoc(`
				load('runtime', 'unsetenv')
				unsetenv("very-long-long-non-existent")
			`),
		},
		{
			name: `unsetenv: existing`,
			script: itn.HereDoc(`
				load('runtime', 'putenv', 'unsetenv', 'getenv')
				putenv("STARLET_TEST", 123456)
				x = getenv("STARLET_TEST")
				print("STARLET_TEST:", x)
				assert.eq(x, "123456")
				unsetenv("STARLET_TEST")
				y = getenv("STARLET_TEST")
				print("STARLET_TEST:", y)
				assert.eq(y, None)
			`),
		},
		{
			name: `setenv like putenv`,
			script: itn.HereDoc(`
				load('runtime', 'setenv', 'getenv')
				setenv("STARLET_TEST", 123456)
				x = getenv("STARLET_TEST")
				print("STARLET_TEST:", x)
				assert.eq(x, "123456")
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
