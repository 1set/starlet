package file_test

import (
	"io/ioutil"
	"os"
	"runtime"
	"testing"

	itn "github.com/1set/starlet/internal"
	lf "github.com/1set/starlet/lib/file"
	"go.starlark.net/starlark"
)

func TestLoadModule_FileCopy(t *testing.T) {
	isOnWindows := runtime.GOOS == "windows"
	tests := []struct {
		name        string
		script      string
		wantErr     string
		skipWindows bool
	}{
		{
			name: `copyfile: no args`,
			script: itn.HereDoc(`
				cf()
			`),
			wantErr: `file.copyfile: missing argument for src`,
		},
		{
			name: `copyfile: src only`,
			script: itn.HereDoc(`
				cf(src=temp_file)
			`),
			wantErr: `file.copyfile: missing argument for dst`,
		},
		{
			name: `copyfile: empty src`,
			script: itn.HereDoc(`
				cf(src="", dst=temp_file+"_another")
			`),
			wantErr: `source path is empty`,
		},
		{
			name: `copyfile: empty dst`,
			script: itn.HereDoc(`
				cf(src=temp_file, dst="")
			`),
			wantErr: `destination path is empty`,
		},
		{
			name: `copyfile: invalid args`,
			script: itn.HereDoc(`
				cf(src=temp_file, dst=temp_file+"_another", overwrite="abc")
			`),
			wantErr: `file.copyfile: for parameter "overwrite": got string, want bool`,
		},
		{
			name: `normal copy`,
			script: itn.HereDoc(`
				load('file', 'stat')
				s, d = temp_file, temp_file+"_another"
				d2 = cf(s, d)
				assert.eq(d2, d)
				ss, sd = stat(s), stat(d)
				assert.eq(ss.type, sd.type)
				assert.eq(ss.size, sd.size)
				assert.eq(ss.modified, sd.modified)
				assert.eq(ss.get_md5(), sd.get_md5())
			`),
		},
		{
			name: `overwrite copy enabled`,
			script: itn.HereDoc(`
				load('file', 'stat')
				s, d = temp_file, temp_file2
				d2 = cf(s, d, overwrite=True)
				assert.eq(d2, d)
				ss, sd = stat(s), stat(d)
				assert.eq(ss.type, sd.type)
				assert.eq(ss.size, sd.size)
				assert.eq(ss.modified, sd.modified)
				assert.eq(ss.get_md5(), sd.get_md5())
			`),
			wantErr: ``,
		},
		{
			name: `overwrite copy disabled`,
			script: itn.HereDoc(`
				cf(temp_file, temp_file2)
			`),
			wantErr: `file already exists`,
		},
		{
			name: `src dst same`,
			script: itn.HereDoc(`
				cf(temp_file, temp_file)
			`),
			wantErr: `source and destination are the same file`,
		},
		{
			name: `src not exists`,
			script: itn.HereDoc(`
				s, d = temp_file + "_not", temp_file+"_another"
				cf(s, d)
			`),
			wantErr:     `no such file or directory`,
			skipWindows: true,
		},
		{
			name: `src is dir`,
			script: itn.HereDoc(`
				s, d = temp_dir, temp_file+"_another"
				cf(s, d)
			`),
			wantErr: `source file is not a regular file`,
		},
		{
			name: `src is device`,
			script: itn.HereDoc(`
				s, d = "/dev/null", temp_file+"_another"
				cf(s, d)
			`),
			wantErr:     `source file is not a regular file`,
			skipWindows: true,
		},
		{
			name: `dst is dir`,
			script: itn.HereDoc(`
				load('file', 'stat')
				s, d = temp_file, temp_dir
				nd = cf(s, d)
				print("dst dir", d)
				print("dst file", nd)
				assert.ne(nd, d)
				ss, sd = stat(s), stat(nd)
				assert.eq(ss.type, sd.type)
				assert.eq(ss.size, sd.size)
				assert.eq(ss.modified, sd.modified)
				assert.eq(ss.get_md5(), sd.get_md5())
			`),
		},
		{
			name: `dst is device`,
			script: itn.HereDoc(`
				cf(temp_file, "/dev/null", overwrite=True)
			`),
			skipWindows: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// prepare temp file/dir if needed
			var (
				tp  string
				tp2 string
				tp3 string
				td  string
			)
			{
				// temp file
				if tf, err := os.CreateTemp("", "starlet-copy-test-write"); err != nil {
					t.Errorf("os.CreateTemp() expects no error, actual error = '%v'", err)
					return
				} else {
					tp = tf.Name()
					if err = ioutil.WriteFile(tp, []byte("Aloha"), 0644); err != nil {
						t.Errorf("ioutil.WriteFile() expects no error, actual error = '%v'", err)
						return
					}
				}
				// temp file 2
				if tf, err := os.CreateTemp("", "starlet-copy-test-write2"); err != nil {
					t.Errorf("os.CreateTemp() expects no error, actual error = '%v'", err)
					return
				} else {
					tp2 = tf.Name()
					if err = ioutil.WriteFile(tp2, []byte("A hui hou"), 0644); err != nil {
						t.Errorf("ioutil.WriteFile() expects no error, actual error = '%v'", err)
						return
					}
				}
				// temp file 3
				if tf, err := os.CreateTemp("", "starlet-copy-test-write3"); err != nil {
					t.Errorf("os.CreateTemp() expects no error, actual error = '%v'", err)
					return
				} else {
					tp3 = tf.Name()
				}
				// temp dir
				if tt, err := os.MkdirTemp("", "starlet-copy-test-dir"); err != nil {
					t.Errorf("os.MkdirTemp() expects no error, actual error = '%v'", err)
					return
				} else {
					td = tt
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
				"temp_file2": starlark.String(tp2),
				"temp_file3": starlark.String(tp3),
				"temp_dir":   starlark.String(td),
			}
			script := `load('file', cf='copyfile')` + "\n" + tt.script
			res, err := itn.ExecModuleWithErrorTest(t, lf.ModuleName, lf.LoadModule, script, tt.wantErr, globals)
			if (err != nil) != (tt.wantErr != "") {
				t.Errorf("path(%q) expects error = '%v', actual error = '%v', result = %v", tt.name, tt.wantErr, err, res)
			}
		})
	}
}
