package file_test

import (
	"os"
	"testing"

	itn "github.com/1set/starlet/internal"
	lf "github.com/1set/starlet/lib/file"
)

func TestLoadModule_File(t *testing.T) {
	//ioutil.TempFile("", "starlet-file-test-write")
	tf, err := os.CreateTemp("", "starlet-file-test-write")
	if err != nil {
		t.Errorf("os.CreateTemp() expects no error, actual error = '%v'", err)
		return
	}
	tp := tf.Name()
	t.Logf("Temp file to write: %s", tp)

	tests := []struct {
		name    string
		script  string
		wantErr string
	}{
		{
			name: `trim bom`,
			script: itn.HereDoc(`
				load('file', 'trim_bom')
				b1 = trim_bom('hello')
				assert.eq(b1, 'hello')
				b2 = trim_bom(b'\xef\xbb\xbfhello')
				assert.eq(b2, b'hello')
			`),
		},
		{
			name: `trim bom with no args`,
			script: itn.HereDoc(`
				load('file', 'trim_bom')
				trim_bom()
			`),
			wantErr: "file.trim_bom: takes exactly one argument (0 given)",
		},
		{
			name: `trim bom with invalid args`,
			script: itn.HereDoc(`
				load('file', 'trim_bom')
				trim_bom(123)
			`),
			wantErr: "file.trim_bom: expected string or bytes, got int",
		},
		{
			name: `read bytes`,
			script: itn.HereDoc(`
				load('file', 'read_bytes')
				b = read_bytes('testdata/aloha.txt')
				assert.eq(b, b'ALOHA\n')
			`),
		},
		{
			name: `read bom bytes`,
			script: itn.HereDoc(`
				load('file', 'read_bytes', 'trim_bom')
				b = read_bytes('testdata/bom.txt')
				assert.eq(b, b'\xef\xbb\xbfhas bom')
				t = trim_bom(b)
				assert.eq(t, b'has bom')
			`),
		},
		{
			name: `read string`,
			script: itn.HereDoc(`
				load('file', 'read_string')
				s = read_string('testdata/aloha.txt')
				assert.eq(s, 'ALOHA\n')
			`),
		},
		{
			name: `read lines`,
			script: itn.HereDoc(`
				load('file', 'read_lines')
				l1 = read_lines('testdata/line_mac.txt')
				assert.eq(len(l1), 3)
				assert.eq(l1, ['Line 1', 'Line 2', 'Line 3'])
				l2 = read_lines('testdata/line_win.txt')
				assert.eq(len(l2), 3)
				assert.eq(l2, ['Line 1', 'Line 2', 'Line 3'])
			`),
		},
		{
			name: `read one line`,
			script: itn.HereDoc(`
				load('file', 'read_lines')
				text = 'ABCDEFGHIJKLMNOPQRSTUVWXYZ甲乙丙丁戊己庚辛壬癸子丑寅卯辰巳午未申酉戌亥'
				l1 = read_lines('testdata/1line.txt')
				assert.eq(len(l1), 1)
				assert.eq(l1, [text])
				l2 = read_lines('testdata/1line_nl.txt')
				assert.eq(len(l2), 1)
				assert.eq(l2, [text])
			`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := itn.ExecModuleWithErrorTest(t, lf.ModuleName, lf.LoadModule, tt.script, tt.wantErr)
			if (err != nil) != (tt.wantErr != "") {
				t.Errorf("file(%q) expects error = '%v', actual error = '%v', result = %v", tt.name, tt.wantErr, err, res)
			}
		})
	}
}
