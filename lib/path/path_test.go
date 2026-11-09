package path_test

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
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
		skipRoot    bool
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
				# Python semantics: an empty component contributes a
				# trailing separator instead of being dropped
				assert.eq(join("a", "", ""), "a/")
				assert.eq(join("a", "b", ""), "a/b/")
				assert.eq(join("a", "", "c"), "a/c")
				assert.eq(join("", "b", "c"), "b/c")
			`),
		},
		{
			name: `join: relative path`,
			script: itn.HereDoc(`
				load('path', 'join')
				# Python semantics: no lexical cleaning - ".." stays put
				# (use normpath to collapse it)
				assert.eq(join("a/b", "../../../xyz"), "a/b/../../../xyz")
				assert.eq(join("a/b", "../../xyz"), "a/b/../../xyz")
			`),
		},
		{
			name: `join: absolute path`,
			script: itn.HereDoc(`
				load('path', 'join')
				assert.eq(join("/a100"), "/a100")
				assert.eq(join("/a100", ""), "/a100/")
				assert.eq(join("/a100", "b"), "/a100/b")
				assert.eq(join("/a100", "b", "c"), "/a100/b/c")
				# an absolute component resets the result
				assert.eq(join("a", "/b"), "/b")
				assert.eq(join("a", "/b", "c"), "/b/c")
			`),
		},
		{
			name: `lexical: basename and dirname`,
			script: itn.HereDoc(`
				load('path', 'basename', 'dirname')
				assert.eq(basename('a/b/c.txt'), 'c.txt')
				assert.eq(basename('a/b/'), '')
				assert.eq(basename('plain'), 'plain')
				assert.eq(basename('/'), '')
				assert.eq(dirname('a/b/c.txt'), 'a/b')
				assert.eq(dirname('a/b/'), 'a/b')
				assert.eq(dirname('plain'), '')
				assert.eq(dirname('/a'), '/')
				assert.eq(dirname('//a'), '//')
			`),
		},
		{
			name: `lexical: normpath`,
			script: itn.HereDoc(`
				load('path', 'normpath')
				assert.eq(normpath('a//b'), 'a/b')
				assert.eq(normpath('a/./b'), 'a/b')
				assert.eq(normpath('a/c/../b'), 'a/b')
				assert.eq(normpath('../a'), '../a')
				assert.eq(normpath('a/../../b'), '../b')
				assert.eq(normpath('/a/../../b'), '/b')
				assert.eq(normpath('//a'), '//a')
				assert.eq(normpath('///a'), '/a')
				assert.eq(normpath(''), '.')
				assert.eq(normpath('a/b/'), 'a/b')
				assert.eq(normpath('../../a'), '../../a')
				assert.eq(normpath('/../a'), '/a')
				assert.eq(normpath('.'), '.')
			`),
		},
		{
			name: `lexical: split and splitext`,
			script: itn.HereDoc(`
				load('path', 'split', 'splitext')
				assert.eq(split('a/b/c.txt'), ('a/b', 'c.txt'))
				assert.eq(split('/a'), ('/', 'a'))
				assert.eq(split('a/b/'), ('a/b', ''))
				assert.eq(split('plain'), ('', 'plain'))
				assert.eq(splitext('a/b.tar.gz'), ('a/b.tar', '.gz'))
				assert.eq(splitext('.bashrc'), ('.bashrc', ''))
				assert.eq(splitext('a/.bashrc'), ('a/.bashrc', ''))
				assert.eq(splitext('noext'), ('noext', ''))
				assert.eq(splitext('a.b/c'), ('a.b/c', ''))
				assert.eq(splitext('..b'), ('..b', ''))
				assert.eq(splitext('a/..'), ('a/..', ''))
			`),
		},
		{
			name: `lexical: isabs`,
			script: itn.HereDoc(`
				load('path', 'isabs')
				assert.true(isabs('/a/b'))
				assert.true(not isabs('a/b'))
				assert.true(not isabs(''))
			`),
		},
		{
			name: `lexical: relpath`,
			script: itn.HereDoc(`
				load('path', 'relpath')
				assert.eq(relpath('/a/b/c', '/a'), 'b/c')
				assert.eq(relpath('/a/b', '/a/b'), '.')
				assert.eq(relpath('/a/b', '/a/c/d'), '../../b')
				assert.eq(relpath('a/b', 'a'), 'b')
				assert.eq(relpath('a/b'), 'a/b')
				assert.eq(relpath('a/b', ''), 'a/b')
				assert.eq(relpath('./a/b', '.'), 'a/b')
			`),
		},
		{
			name: `lexical: basename invalid`,
			script: `load('path', 'basename')
basename(1)`,
			wantErr: `path.basename: for parameter path: got int, want string`,
		},
		{
			name: `lexical: normpath invalid`,
			script: `load('path', 'normpath')
normpath(1)`,
			wantErr: `path.normpath: for parameter path: got int, want string`,
		},
		{
			name: `lexical: split invalid`,
			script: `load('path', 'split')
split(1)`,
			wantErr: `path.split: for parameter path: got int, want string`,
		},
		{
			name: `lexical: splitext invalid`,
			script: `load('path', 'splitext')
splitext(1)`,
			wantErr: `path.splitext: for parameter path: got int, want string`,
		},
		{
			name: `lexical: isabs invalid`,
			script: `load('path', 'isabs')
isabs(1)`,
			wantErr: `path.isabs: for parameter path: got int, want string`,
		},
		{
			name: `lexical: expanduser invalid`,
			script: `load('path', 'expanduser')
expanduser(1)`,
			wantErr: `path.expanduser: for parameter path: got int, want string`,
		},
		{
			name: `lexical: relpath invalid`,
			script: `load('path', 'relpath')
relpath(1)`,
			wantErr: `path.relpath: for parameter path: got int, want string`,
		},
		{
			name: `lexical: relpath empty path`,
			script: itn.HereDoc(`
				load('path', 'relpath')
				relpath('')
			`),
			wantErr: `path.relpath: no path specified`,
		},
		{
			name: `lexical: relpath mixed kinds`,
			script: itn.HereDoc(`
				load('path', 'relpath')
				relpath('/a/b', 'c')
			`),
			wantErr: `path.relpath: cannot mix an absolute path with a relative start`,
		},
		{
			name: `lexical: try_relpath`,
			script: itn.HereDoc(`
				load('path', 'try_relpath')
				v, err = try_relpath('/a/b/c', '/a')
				assert.eq(err, None)
				assert.eq(v, 'b/c')
				v2, err2 = try_relpath('/a', 'c')
				assert.eq(v2, None)
				assert.true('cannot mix' in err2)
			`),
		},
		{
			name: `try_abs`,
			script: itn.HereDoc(`
				load('path', 'try_abs')
				v, err = try_abs('.')
				assert.eq(err, None)
				assert.true(len(v) > 0)
			`),
		},
		{
			name: `expanduser`,
			script: itn.HereDoc(`
				load('path', 'expanduser')
				h = expanduser('~')
				assert.true(h != '~')
				hs = expanduser('~/x')
				assert.true(hs.endswith('/x'))
				assert.eq(expanduser('~user/x'), '~user/x')
				assert.eq(expanduser('plain'), 'plain')
			`),
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
						p = listdir('/var/db/sudo')
					elif runtime_os == "linux":
						p = listdir('/root')
					else:
						print("skip for", runtime_os)
				run()
			`),
			wantErr:     `path.listdir: open /`,
			skipWindows: true,
			skipRoot:    true, // root can read these directories, so the error never happens
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
			if tt.skipRoot && os.Geteuid() == 0 {
				t.Skipf("Skip test as root")
				return
			}
			// path.chdir mutates the process-global CWD; restore it after each
			// case so cases stay independent and the suite survives -count=2.
			origWD, err := os.Getwd()
			if err != nil {
				t.Fatalf("os.Getwd() expects no error, actual error = '%v'", err)
			}
			defer func() {
				if err := os.Chdir(origWD); err != nil {
					t.Fatalf("failed to restore working directory: %v", err)
				}
			}()

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

func TestLoadModule_PathListdirFilterDepth(t *testing.T) {
	// Keep the fixture on the working directory's volume so both relative and
	// absolute paths work on Windows without changing the process-wide CWD.
	base, err := os.MkdirTemp(".", ".listdir-filter-")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll(base); err != nil {
			t.Error(err)
		}
	})
	root := filepath.Join(base, "root")
	for _, dir := range []string{filepath.Join(root, "child", "deeper"), filepath.Join(base, "empty")} {
		if err := os.MkdirAll(dir, 0700); err != nil {
			t.Fatal(err)
		}
	}
	for _, name := range []string{"top.txt", "skip.bin", "child/nested.txt", "child/deeper/deep.txt"} {
		if err := os.WriteFile(filepath.Join(root, filepath.FromSlash(name)), []byte("fixture"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	abs, err := filepath.Abs(root)
	if err != nil {
		t.Fatal(err)
	}
	shallow := []string{"child", "skip.bin", "top.txt"}
	deep := []string{"child", "child/deeper", "child/deeper/deep.txt", "child/nested.txt", "skip.bin", "top.txt"}
	for _, prefix := range []struct{ name, path string }{
		{"relative", root},
		{"absolute", abs},
		{"trailing_separator", root + string(filepath.Separator)},
	} {
		t.Run(prefix.name, func(t *testing.T) {
			for _, tc := range []struct {
				name, call string
				accept     func(string) bool
				want       []string
				visits     []string
				filterErr  bool
				nonBool    bool
				wantErr    string
			}{
				{name: "default", call: "list_root", want: shallow},
				{name: "none_filter", call: "list_root, filter=None", want: shallow},
				{name: "accept_all", call: "list_root, filter=filter_func", accept: func(string) bool { return true }, want: shallow, visits: shallow},
				{name: "reject_directories_default", call: "list_root, filter=filter_func", accept: func(p string) bool { return strings.HasSuffix(p, ".txt") }, want: []string{"top.txt"}, visits: shallow},
				{name: "reject_directories_nonrecursive", call: "list_root, recursive=False, filter=filter_func", accept: func(p string) bool { return strings.HasSuffix(p, ".txt") }, want: []string{"top.txt"}, visits: shallow},
				{name: "reject_all", call: "list_root, filter=filter_func", accept: func(string) bool { return false }, visits: shallow},
				{name: "reject_file", call: "list_root, filter=filter_func", accept: func(p string) bool { return !strings.HasSuffix(p, ".bin") }, want: []string{"child", "top.txt"}, visits: shallow},
				{name: "recursive_none", call: "list_root, recursive=True, filter=None", want: deep},
				{name: "recursive_rejected_directories", call: "list_root, True, filter_func", accept: func(p string) bool { return strings.HasSuffix(p, ".txt") }, want: []string{"child/deeper/deep.txt", "child/nested.txt", "top.txt"}, visits: deep},
				{name: "filter_error", call: "list_root, filter=filter_func", filterErr: true, visits: []string{"child"}, wantErr: "filter stopped"},
				{name: "filter_non_bool", call: "list_root, filter=filter_func", nonBool: true, visits: []string{"child"}, wantErr: "got int, want bool"},
			} {
				t.Run(tc.name, func(t *testing.T) {
					var visits []string
					filter := starlark.NewBuiltin("listdir_filter", func(_ *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
						if len(args) != 1 || len(kwargs) != 0 {
							return nil, errors.New("filter must receive one positional path")
						}
						p, ok := starlark.AsString(args[0])
						if !ok {
							return nil, errors.New("filter path must be a string")
						}
						visits = append(visits, p)
						if tc.filterErr {
							return nil, errors.New("filter stopped")
						}
						if tc.nonBool {
							return starlark.MakeInt(1), nil
						}
						return starlark.Bool(tc.accept(p)), nil
					})
					script := "load('path', 'listdir')\nresult = listdir(" + tc.call + ")\n"
					out, err := itn.ExecModuleWithErrorTest(t, lpath.ModuleName, lpath.LoadModule, script, tc.wantErr, starlark.StringDict{
						"list_root":   starlark.String(prefix.path),
						"filter_func": filter,
					})
					if (err != nil) != (tc.wantErr != "") {
						t.Fatalf("error = %v, want %q", err, tc.wantErr)
					}
					paths := func(relative []string) []string {
						out := make([]string, len(relative))
						for i, p := range relative {
							out[i] = filepath.Join(prefix.path, filepath.FromSlash(p))
						}
						return out
					}
					if !reflect.DeepEqual(append([]string{}, visits...), paths(tc.visits)) {
						t.Errorf("filter visited %q, want %q", visits, paths(tc.visits))
					}
					if err != nil {
						return
					}
					list, ok := out["result"].(*starlark.List)
					if !ok {
						t.Fatalf("result type = %T, want *starlark.List", out["result"])
					}
					got := make([]string, list.Len())
					for i := range got {
						got[i], ok = starlark.AsString(list.Index(i))
						if !ok {
							t.Fatalf("result[%d] type = %T, want string", i, list.Index(i))
						}
					}
					if !reflect.DeepEqual(got, paths(tc.want)) {
						t.Errorf("result = %q, want %q", got, paths(tc.want))
					}
				})
			}
		})
	}
	for _, path := range []string{filepath.Join(base, "empty"), filepath.Join(root, "top.txt")} {
		globals := starlark.StringDict{"list_root": starlark.String(path)}
		itn.ExecModuleWithErrorTest(t, lpath.ModuleName, lpath.LoadModule, "load('path', 'listdir')\nassert.eq(listdir(list_root, filter=lambda p: fail('must not run')), [])", "", globals)
	}
}
