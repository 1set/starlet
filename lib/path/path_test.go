package path_test

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"testing"

	itn "github.com/1set/starlet/internal"
	lpath "github.com/1set/starlet/lib/path"
)

func TestLoadModule_Path(t *testing.T) {
	isOnWindows := runtime.GOOS == "windows"
	tests := []struct {
		name        string
		script      string
		wantErr     string
		skipWindows bool
	}{
		{
			name: `join: no args`,
			script: itn.HereDoc(`
				load('path', 'join')
				join()
			`),
			wantErr: "path.join: got 0 arguments, want at least 1",
		},
		{
			name: `join: 1 arg`,
			script: itn.HereDoc(`
				load('path', 'join')
				p = join('a')
				assert.eq(p, 'a')
			`),
		},
		{
			name: `join: 2 args`,
			script: itn.HereDoc(`
				load('path', 'join')
				p = join('a', 'b')
				assert.eq(p, 'a/b')
			`),
			skipWindows: true,
		},
		{
			name: `join: invalid type`,
			script: itn.HereDoc(`
				load('path', 'join')
				p = join('a', 1)
			`),
			wantErr: "path.join: for parameter path: got int, want string",
		},
		{
			name: `abs: no args`,
			script: itn.HereDoc(`
				load('path', 'abs')
				abs()
			`),
			wantErr: "path.abs: missing argument for path",
		},
		{
			name: `abs: invalid type`,
			script: itn.HereDoc(`
				load('path', 'abs')
				p = abs(1)
			`),
			wantErr: "path.abs: for parameter path: got int, want string",
		},
		{
			name: `abs: empty path`,
			script: itn.HereDoc(`
				load('path', 'abs')
				p = abs('')
				assert.true(p.endswith('lib/path'))
			`),
			skipWindows: true,
		},
		{
			name: `abs: non-existent path`,
			script: itn.HereDoc(`
				load('path', 'abs')
				p = abs('non-existent-path')
				assert.true(p.endswith('lib/path/non-existent-path'))
			`),
			skipWindows: true,
		},
		{
			name: `abs: existing path`,
			script: itn.HereDoc(`
				load('path', 'abs')
				p = abs('path_test.go')
				assert.true(p.endswith('lib/path/path_test.go'))
			`),
			skipWindows: true,
		},
		{
			name: `exist: no args`,
			script: itn.HereDoc(`
				load('path', 'exist')
				exist()
			`),
			wantErr: "path.exist: missing argument for path",
		},
		{
			name: `exist: invalid type`,
			script: itn.HereDoc(`
				load('path', 'exist')
				p = exist(1)
			`),
			wantErr: "path.exist: for parameter path: got int, want string",
		},
		{
			name: `exist: empty path`,
			script: itn.HereDoc(`
				load('path', 'exist')
				p = exist('')
				assert.true(not p)
			`),
		},
		{
			name: `exist: non-existent path`,
			script: itn.HereDoc(`
				load('path', 'exist')
				p = exist('non-existent-path')
				assert.true(not p)
			`),
		},
		{
			name: `exist: existing file`,
			script: itn.HereDoc(`
				load('path', 'exist')
				p = exist('path_test.go')
				assert.true(p)
			`),
		},
		{
			name: `exist: existing dot`,
			script: itn.HereDoc(`
				load('path', 'exist')
				p = exist('.')
				assert.true(p)
			`),
		},
		{
			name: `exist: existing dot-dot`,
			script: itn.HereDoc(`
				load('path', 'exist')
				p = exist('..')
				assert.true(p)
			`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// prepare temp file if needed
			var tp string
			if strings.Contains(tt.script, "%q") {
				tf, err := os.CreateTemp("", "starlet-file-test-write")
				if err != nil {
					t.Errorf("os.CreateTemp() expects no error, actual error = '%v'", err)
					return
				}
				tp = tf.Name()
				//t.Logf("Temp file to write: %s", tp)
				tt.script = fmt.Sprintf(tt.script, tp)
			}
			// execute test
			if isOnWindows && tt.skipWindows {
				t.Skipf("Skip test on Windows")
				return
			}
			res, err := itn.ExecModuleWithErrorTest(t, lpath.ModuleName, lpath.LoadModule, tt.script, tt.wantErr, nil)
			if (err != nil) != (tt.wantErr != "") {
				t.Errorf("path(%q) expects error = '%v', actual error = '%v', result = %v", tt.name, tt.wantErr, err, res)
			}
		})
	}
}
