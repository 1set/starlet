package path_test

import (
	"os"
	"runtime"
	"testing"

	itn "github.com/1set/starlet/internal"
	lpath "github.com/1set/starlet/lib/path"
	"go.starlark.net/starlark"
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
			name: `join: empty path`,
			script: itn.HereDoc(`
				load('path', 'join')
				assert.eq(join(""), "")
				assert.eq(join("",""), "")
				assert.eq(join("","",""), "")
			`),
		},
		{
			name: `join: partial empty`,
			script: itn.HereDoc(`
				load('path', 'join')
				assert.eq(join("a", "", ""), "a")
				assert.eq(join("a", "b", ""), "a/b")
				assert.eq(join("a", "", "c"), "a/c")
				assert.eq(join("", "b", "c"), "b/c")
			`),
			skipWindows: true,
		},
		{
			name: `join: relative path`,
			script: itn.HereDoc(`
				load('path', 'join')
				assert.eq(join("a/b", "../../../xyz"), "../xyz")
				assert.eq(join("a/b", "../../xyz"), "xyz")
			`),
			skipWindows: true,
		},
		{
			name: `join: absolute path`,
			script: itn.HereDoc(`
				load('path', 'join')
				assert.eq(join("/a100"), "/a100")
				assert.eq(join("/a100", ""), "/a100")
				assert.eq(join("/a100", "b"), "/a100/b")
				assert.eq(join("/a100", "b", "c"), "/a100/b/c")
			`),
			skipWindows: true,
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
			name: `exists: no args`,
			script: itn.HereDoc(`
				load('path', 'exists')
				exists()
			`),
			wantErr: "path.exists: missing argument for path",
		},
		{
			name: `exists: invalid type`,
			script: itn.HereDoc(`
				load('path', 'exists')
				p = exists(1)
			`),
			wantErr: "path.exists: for parameter path: got int, want string",
		},
		{
			name: `exists: empty path`,
			script: itn.HereDoc(`
				load('path', 'exists')
				p = exists('')
				assert.true(not p)
			`),
		},
		{
			name: `exists: non-existent path`,
			script: itn.HereDoc(`
				load('path', 'exists')
				p = exists('non-existent-path')
				assert.true(not p)
			`),
		},
		{
			name: `exists: existing file`,
			script: itn.HereDoc(`
				load('path', 'exists')
				p = exists('path_test.go')
				assert.true(p)
			`),
		},
		{
			name: `exists: existing dot`,
			script: itn.HereDoc(`
				load('path', 'exists')
				p = exists('.')
				assert.true(p)
			`),
		},
		{
			name: `exists: existing dot-dot`,
			script: itn.HereDoc(`
				load('path', 'exists')
				p = exists('..')
				assert.true(p)
			`),
		},
		{
			name: `is_file: not exist`,
			script: itn.HereDoc(`
				load('path', 'is_file')
				p = is_file('non-existent-path')
				assert.true(not p)
			`),
		},
		{
			name: `is_file: check file`,
			script: itn.HereDoc(`
				load('path', 'is_file')
				p = is_file('path_test.go')
				assert.true(p)
			`),
		},
		{
			name: `is_file: check dir`,
			script: itn.HereDoc(`
				load('path', 'is_file')
				p = is_file('.')
				assert.true(not p)
			`),
		},
		{
			name: `is_dir: not exist`,
			script: itn.HereDoc(`
				load('path', 'is_dir')
				p = is_dir('non-existent-path')
				assert.true(not p)
			`),
		},
		{
			name: `is_dir: check file`,
			script: itn.HereDoc(`
				load('path', 'is_dir')
				p = is_dir('path_test.go')
				assert.true(not p)
			`),
		},
		{
			name: `is_dir: check dir`,
			script: itn.HereDoc(`
				load('path', 'is_dir')
				p = is_dir('.')
				assert.true(p)
			`),
		},
		{
			name: `is_link: not exist`,
			script: itn.HereDoc(`
				load('path', 'is_link')
				p = is_link('non-existent-path')
				assert.true(not p)
			`),
		},
		{
			name: `is_link: check file and dir`,
			script: itn.HereDoc(`
				load('path', 'is_link')
				assert.true(is_link('path_test.go') == False)
				assert.true(is_link('.') == False)
			`),
		},
		{
			name: `listdir: no args`,
			script: itn.HereDoc(`
				load('path', 'listdir')
				listdir()
			`),
			wantErr: `path.listdir: missing argument for path`,
		},
		{
			name: `listdir: invalid type`,
			script: itn.HereDoc(`
				load('path', 'listdir')
				p = listdir(1)
			`),
			wantErr: `path.listdir: for parameter path: got int, want string`,
		},
		{
			name: `listdir: non-existent path`,
			script: itn.HereDoc(`
				load('path', 'listdir')
				p = listdir('non-existent-path')
			`),
			wantErr:     `path.listdir: lstat non-existent-path`,
			skipWindows: true,
		},
		{
			name: `listdir: existing file`,
			script: itn.HereDoc(`
				load('path', 'listdir')
				p = listdir('path_test.go')
				assert.eq(p, [])
			`),
		},
		{
			name: `listdir: existing dir`,
			script: itn.HereDoc(`
				load('path', 'listdir')
				p = listdir('.')
				print("listdir:", p)
				assert.true('path_test.go' in p)
			`),
		},
		{
			name: `listdir: None filter`,
			script: itn.HereDoc(`
				load('path', 'listdir')
				p = listdir('.', filter=None)
				assert.true('path_test.go' in p)
			`),
		},
		{
			name: `listdir: invalid filter`,
			script: itn.HereDoc(`
				load('path', 'listdir')
				p = listdir('.', filter=123)
			`),
			wantErr: `expected <nil> or None, got int`,
		},
		{
			name: `listdir: mistyped filter`,
			script: itn.HereDoc(`
				load('path', 'listdir')
				p = listdir('.', filter=lambda x: len(x))
			`),
			wantErr: `got int, want bool`,
		},
		{
			name: `listdir: filter fail`,
			script: itn.HereDoc(`
				load('path', 'listdir')
				p = listdir('.', filter=lambda x: fail("inner"))
			`),
			wantErr: `fail: inner`,
		},
		{
			name: `listdir: filter file`,
			script: itn.HereDoc(`
				load('path', 'listdir')
				p = listdir('.', filter=lambda x: not x.endswith('.go'))
				print("listdir:", p)
				assert.true('path_test.go' not in p)
			`),
		},
		{
			name: `listdir: list dev`,
			script: itn.HereDoc(`
				load('path', 'listdir')
				p = listdir('/dev')
				print("listdir device:", p)
				assert.true(len(p) > 0)
			`),
			skipWindows: true,
		},
		{
			name: `listdir: existing dir parent`,
			script: itn.HereDoc(`
				load('path', 'listdir')
				p = listdir('..')
				print("listdir parent:", p)
				assert.true('../path' in p)
				assert.true('../path/path_test.go' not in p)
			`),
			skipWindows: true,
		},
		{
			name: `listdir: existing dir recursive`,
			script: itn.HereDoc(`
				load('path', 'listdir')
				p = listdir('..', True)
				print("listdir parent recursively:", p)
				assert.true('../path' in p)
				assert.true('../path/path_test.go' in p)
			`),
			skipWindows: true,
		},
		{
			name: `listdir: no permission`,
			script: itn.HereDoc(`
				load('path', 'listdir')
				def run():
					if runtime_os == "darwin":
						p = listdir('/var/root')
					elif runtime_os == "linux":
						p = listdir('/root')
					else:
						print("skip for", runtime_os)
				run()
			`),
			wantErr:     `path.listdir: open /`,
			skipWindows: true,
		},
		{
			name: `getcwd: no args`,
			script: itn.HereDoc(`
				load('path', 'getcwd')
				p = getcwd()
				print("cwd:", p)
				assert.true(p.endswith('path'))
			`),
		},
		{
			name: `getcwd: extra args`,
			script: itn.HereDoc(`
				load('path', 'getcwd')
				getcwd(123)
			`),
			wantErr: `path.getcwd: got 1 arguments, want 0`,
		},
		{
			name: `chdir: no args`,
			script: itn.HereDoc(`
				load('path', 'chdir')
				chdir()
			`),
			wantErr: `path.chdir: missing argument for path`,
		},
		{
			name: `chdir: invalid type`,
			script: itn.HereDoc(`
				load('path', 'chdir')
				chdir(123)
			`),
			wantErr: `path.chdir: for parameter path: got int, want string`,
		},
		{
			name: `chdir: non-existent path`,
			script: itn.HereDoc(`
				load('path', 'chdir')
				chdir('non-existent-path')
			`),
			wantErr: `path.chdir: chdir non-existent-path`,
		},
		{
			name: `chdir: current path`,
			script: itn.HereDoc(`
				load('path', 'chdir', 'abs')
				a = abs('.')
				chdir('.')
				b = abs('.')
				chdir('../path')
				c = abs('.')
				assert.eq(a, b)
				assert.eq(a, c)
			`),
		},
		{
			name: `chdir: parent path`,
			script: itn.HereDoc(`
				load('path', 'chdir', 'abs')
				a = abs('.')
				chdir('..')
				b = abs('.')
				assert.ne(a, b)
				assert.true(a.startswith(b))
			`),
		},
		{
			name: `chdir: file path`,
			script: itn.HereDoc(`
				load('path', 'chdir')
				chdir('path_test.go')
			`),
			wantErr: `path.chdir: chdir path_test.go`,
		},
		{
			name: `make dir: existing`,
			script: itn.HereDoc(`
				load('path', 'mkdir')
				mkdir(temp_dir)
			`),
		},
		{
			name: `make dir: conflict`,
			script: itn.HereDoc(`
				load('path', 'mkdir')
				mkdir(temp_file)
			`),
			wantErr:     `not a directory`,
			skipWindows: true,
		},
		{
			name: `make dir: new`,
			script: itn.HereDoc(`
				load('path', 'mkdir')
				mkdir(temp_dir + "/newdir")
			`),
		},
		{
			name: `make dir: invalid`,
			script: itn.HereDoc(`
				load('path', 'mkdir')
				mkdir()
			`),
			wantErr: `path.mkdir: missing argument for path`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// prepare temp file if needed
			var (
				tp string
				td string
			)
			{
				// temp file
				tf, err := os.CreateTemp("", "starlet-path-test-write")
				if err != nil {
					t.Errorf("os.CreateTemp() expects no error, actual error = '%v'", err)
					return
				}
				tp = tf.Name()

				// temp dir
				td, err = os.MkdirTemp("", "starlet-path-test-dir")
				if err != nil {
					t.Errorf("os.MkdirTemp() expects no error, actual error = '%v'", err)
					return
				}
			}

			// execute test
			if isOnWindows && tt.skipWindows {
				t.Skipf("Skip test on Windows")
				return
			}
			globals := starlark.StringDict{
				"runtime_os": starlark.String(runtime.GOOS),
				"temp_file":  starlark.String(tp),
				"temp_dir":   starlark.String(td),
			}
			res, err := itn.ExecModuleWithErrorTest(t, lpath.ModuleName, lpath.LoadModule, tt.script, tt.wantErr, globals)
			if (err != nil) != (tt.wantErr != "") {
				t.Errorf("path(%q) expects error = '%v', actual error = '%v', result = %v", tt.name, tt.wantErr, err, res)
			}
		})
	}
}
