package file_test

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"testing"

	itn "github.com/1set/starlet/internal"
	lf "github.com/1set/starlet/lib/file"
)

func TestLoadModule_File(t *testing.T) {
	isOnWindows := runtime.GOOS == "windows"
	tests := []struct {
		name        string
		script      string
		wantErr     string
		fileContent string
		skipWindows bool
	}{
		{
			name: `trim bom`,
			script: itn.HereDoc(`
				load('file', 'trim_bom')
				b0 = trim_bom('A')
				assert.eq(b0, 'A')
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
			name: `read empty file`,
			script: itn.HereDoc(`
				load('file', 'read_bytes', 'read_string', 'read_lines', 'count_lines', 'head_lines', 'tail_lines')
				fp = 'testdata/empty.txt'
				b = read_bytes(fp)
				assert.eq(b, b'')
				s = read_string(fp)
				assert.eq(s, '')
				l = read_lines(fp)
				assert.eq(l, [])
				c = count_lines(fp)
				assert.eq(c, 0)
				h = head_lines(fp, 10)
				assert.eq(h, [])
				t = tail_lines(fp, 10)
				assert.eq(t, [])
			`),
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
			name: `read no args`,
			script: itn.HereDoc(`
				load('file', 'read_string')
				read_string()
			`),
			wantErr: "file.read_string: missing argument for name",
		},
		{
			name: `read invalid args`,
			script: itn.HereDoc(`
				load('file', 'read_string')
				read_string(123)
			`),
			wantErr: "file.read_string: for parameter name: got int, want string or bytes",
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
		{
			name: `write no args`,
			script: itn.HereDoc(`
				load('file', 'write_bytes')
				write_bytes()
			`),
			wantErr: `file.write_bytes: missing argument for name`,
		},
		{
			name: `write no data`,
			script: itn.HereDoc(`
				load('file', 'write_bytes')
				write_bytes("try-one-test.txt")
			`),
			wantErr: `file.write_bytes: missing argument for data`,
		},
		{
			name: `write_bytes invalid data`,
			script: itn.HereDoc(`
				load('file', 'write_bytes')
				fp = %q
				write_bytes(fp, 123)
			`),
			wantErr: "file.write_bytes: expected string or bytes, got int",
		},
		{
			name: `write_string invalid data`,
			script: itn.HereDoc(`
				load('file', 'write_string')
				fp = %q
				write_string(fp, 123)
			`),
			wantErr: "file.write_string: expected string or bytes, got int",
		},
		{
			name: `write_lines invalid data`,
			script: itn.HereDoc(`
				load('file', 'write_lines')
				fp = %q
				write_lines(fp, 123)
			`),
			wantErr: "file.write_lines: expected list/tuple/set, got int",
		},
		{
			name: `write bytes`,
			script: itn.HereDoc(`
				load('file', 'write_bytes')
				fp = %q
				write_bytes(fp, b'hello')
				write_bytes(fp, b'world')
				write_bytes(fp, 'Word')
			`),
			fileContent: "Word",
		},
		{
			name: `write string`,
			script: itn.HereDoc(`
				load('file', 'write_string')
				fp = %q
				write_string(fp, "Hello World")
				write_string(fp, "Aloha!\nA hui hou!\n")
			`),
			fileContent: "Aloha!\nA hui hou!\n",
		},
		{
			name: `write lines: string`,
			script: itn.HereDoc(`
				load('file', 'write_lines')
				fp = %q
				write_lines(fp, "Hello World")
				write_lines(fp, "Goodbye~")
			`),
			fileContent: "Goodbye~\n",
		},
		{
			name: `write lines: bytes`,
			script: itn.HereDoc(`
				load('file', 'write_lines')
				fp = %q
				write_lines(fp, b"Hello World")
				write_lines(fp, b"Goodbye!")
			`),
			fileContent: "Goodbye!\n",
		},
		{
			name: `write lines: list`,
			script: itn.HereDoc(`
				load('file', 'write_lines')
				fp = %q
				write_lines(fp, ["Hello", "World"])
				write_lines(fp, ["Great", "Job"])
			`),
			fileContent: "Great\nJob\n",
		},
		{
			name: `write lines: tuple`,
			script: itn.HereDoc(`
				load('file', 'write_lines')
				fp = %q
				write_lines(fp, ("Hello", "World"))
				write_lines(fp, ("Nice", "Job"))
			`),
			fileContent: "Nice\nJob\n",
		},
		{
			name: `write lines: set`,
			script: itn.HereDoc(`
				load('file', 'write_lines')
				fp = %q
				write_lines(fp, set(["Hello", "World"]))
				write_lines(fp, set(["Good", "Job"]))
			`),
			fileContent: "Good\nJob\n",
		},
		{
			name: `write lines: various type`,
			script: itn.HereDoc(`
				load('file', 'write_lines')
				fp = %q
				write_lines(fp, ["Hello", b'world', 123, [True, False]])
			`),
			fileContent: "Hello\nworld\n123\n[True, False]\n",
		},
		{
			name: `append bytes`,
			script: itn.HereDoc(`
				load('file', 'append_bytes')
				fp = %q
				append_bytes(fp, b'hello')
				append_bytes(fp, b'world')
			`),
			fileContent: "helloworld",
		},
		{
			name: `append string`,
			script: itn.HereDoc(`
				load('file', 'append_string', 'write_string')
				fp = %q
				write_string(fp, "Hello World\n")
				append_string(fp, "Aloha!\nA hui hou!\n")
			`),
			fileContent: "Hello World\nAloha!\nA hui hou!\n",
		},
		{
			name: `append lines`,
			script: itn.HereDoc(`
				load('file', 'write_lines', 'append_lines')
				fp = %q
				write_lines(fp, ["Hello", "World"])
				append_lines(fp, ["Great", "Job"])
				append_lines(fp, ["Bye"])
			`),
			fileContent: "Hello\nWorld\nGreat\nJob\nBye\n",
		},
		{
			name: `count lines: no args`,
			script: itn.HereDoc(`
				load('file', 'count_lines')
				count_lines()
			`),
			wantErr: `file.count_lines: missing argument for name`,
		},
		{
			name: `count lines`,
			script: itn.HereDoc(`
				load('file', 'count_lines')
				assert.eq(3, count_lines('testdata/line_mac.txt'))
				assert.eq(3, count_lines('testdata/line_win.txt'))
			`),
		},
		{
			name: `read head lines`,
			script: itn.HereDoc(`
				load('file', 'head_lines')
				l1 = head_lines('testdata/line_mac.txt', 10)
				assert.eq(len(l1), 3)
				assert.eq(l1, ['Line 1', 'Line 2', 'Line 3'])
				l2 = head_lines('testdata/line_win.txt', 2)
				assert.eq(len(l2), 2)
				assert.eq(l2, ['Line 1', 'Line 2'])
			`),
		},
		{
			name: `read tail lines`,
			script: itn.HereDoc(`
				load('file', 'tail_lines')
				l1 = tail_lines('testdata/line_mac.txt', 10)
				assert.eq(len(l1), 3)
				assert.eq(l1, ['Line 1', 'Line 2', 'Line 3'])
				l2 = tail_lines('testdata/line_win.txt', 2)
				assert.eq(len(l2), 2)
				assert.eq(l2, ['Line 2', 'Line 3'])
			`),
		},
		{
			name: `read head lines: invalid args`,
			script: itn.HereDoc(`
				load('file', 'head_lines')
				l1 = head_lines(123, 10)
			`),
			wantErr: `file.head_lines: for parameter name: got int, want string or bytes`,
		},
		{
			name: `read tail lines: invalid n`,
			script: itn.HereDoc(`
				load('file', 'tail_lines')
				l1 = tail_lines('testdata/line_mac.txt', -7)
			`),
			wantErr: `file.tail_lines: expected positive integer, got -7`,
		},
		{
			name: `write_bytes: conflict`,
			script: itn.HereDoc(`
				load('file', 'write_bytes')
				write_bytes('testdata/', b'hello')
			`),
			wantErr: `open testdata/:`,
		},
		{
			name: `write_string: conflict`,
			script: itn.HereDoc(`
				load('file', 'write_string')
				write_string('testdata/', b'hello')
			`),
			wantErr: `open testdata/:`,
		},
		{
			name: `write_lines: conflict`,
			script: itn.HereDoc(`
				load('file', 'write_lines')
				write_lines('testdata/', b'hello')
			`),
			wantErr: `open testdata/:`,
		},
		{
			name: `append_bytes: conflict`,
			script: itn.HereDoc(`
				load('file', 'append_bytes')
				append_bytes('testdata/', b'hello')
			`),
			wantErr: `open testdata/:`,
		},
		{
			name: `append_string: conflict`,
			script: itn.HereDoc(`
				load('file', 'append_string')
				append_string('testdata/', b'hello')
			`),
			wantErr: `open testdata/:`,
		},
		{
			name: `append_lines: conflict`,
			script: itn.HereDoc(`
				load('file', 'append_lines')
				append_lines('testdata/', b'hello')
			`),
			wantErr: `open testdata/:`,
		},
		{
			name: `read_bytes: not exist`,
			script: itn.HereDoc(`
				load('file', 'read_bytes')
				s = read_bytes('not-such-file1.txt')
			`),
			wantErr: `open not-such-file1.txt:`,
		},
		{
			name: `read_string: not exist`,
			script: itn.HereDoc(`
				load('file', 'read_string')
				s = read_string('not-such-file2.txt')
			`),
			wantErr: `open not-such-file2.txt:`,
		},
		{
			name: `read_lines: not exist`,
			script: itn.HereDoc(`
				load('file', 'read_lines')
				s = read_lines('not-such-file3.txt')
			`),
			wantErr: `open not-such-file3.txt:`,
		},
		{
			name: `count not exist`,
			script: itn.HereDoc(`
				load('file', 'count_lines')
				s = count_lines('not-such-file4.txt')
			`),
			wantErr: `open not-such-file4.txt:`,
		},
		{
			name: `head not exist`,
			script: itn.HereDoc(`
				load('file', 'head_lines')
				s = head_lines('not-such-file5.txt', 10)
			`),
			wantErr: `open not-such-file5.txt:`,
		},
		{
			name: `tail not exist`,
			script: itn.HereDoc(`
				load('file', 'tail_lines')
				s = tail_lines('not-such-file6.txt', 10)
			`),
			wantErr: `open not-such-file6.txt:`,
		},
		{
			name: `stat: no arg`,
			script: itn.HereDoc(`
				load('file', 'stat')
				s = stat()
			`),
			wantErr: `file.stat: missing argument for name`,
		},
		{
			name: `stat: not exist`,
			script: itn.HereDoc(`
				load('file', 'stat')
				s = stat('not-such-file7.txt')
			`),
			wantErr:     `file.stat: lstat not-such-file7.txt`,
			skipWindows: true,
		},
		{
			name: `stat: empty path`,
			script: itn.HereDoc(`
				load('file', 'stat')
				s = stat('')
			`),
			wantErr:     `file.stat: lstat :`,
			skipWindows: true,
		},
		{
			name: `stat: file`,
			script: itn.HereDoc(`
				load('file', 'stat')
				fp = 'testdata/aloha.txt'
				s = stat(fp)
				assert.eq(s.name, 'aloha.txt')
				assert.eq(s.size, 6)
				assert.eq(s.type, 'file')
				assert.eq(s.ext, '.txt')
				assert.true("testdata" in s.path)
				assert.true("aloha.txt" in s.path)
				assert.true(s.modified.unix > 0)
				assert.eq(s.get_md5(), '6a12867bd5e0810f2dae51da4a51f001')
				assert.eq(s.get_sha1(), '71a45eadccd2f29bbf60f46b13e019ae62c8b0bd')
				assert.eq(s.get_sha256(), 'eb69c86a84164b23808dcda13fdfbe664760701cf605df28272d4efd2ed18ab4')
				assert.eq(s.get_sha512(), '9946cf3ba83eff33d9798fe933785b8a0f20aa179ca1d18418fd401d955a1270edf9542eb7b10833bbfe1de84b31dde6924e4e01f0e335174c45fede9f9e80ef')
			`),
		},
		{
			name: `stat: follow file`,
			script: itn.HereDoc(`
				load('file', 'stat')
				fp = 'testdata/aloha.txt'
				s = stat(fp, True)
				assert.eq(s.name, 'aloha.txt')
				assert.eq(s.size, 6)
				assert.eq(s.type, 'file')
				assert.eq(s.ext, '.txt')
				assert.true("testdata" in s.path)
				assert.true("aloha.txt" in s.path)
				assert.true(s.modified.unix > 0)
				assert.eq(s.get_md5(), '6a12867bd5e0810f2dae51da4a51f001')
				assert.eq(s.get_sha1(), '71a45eadccd2f29bbf60f46b13e019ae62c8b0bd')
				assert.eq(s.get_sha256(), 'eb69c86a84164b23808dcda13fdfbe664760701cf605df28272d4efd2ed18ab4')
				assert.eq(s.get_sha512(), '9946cf3ba83eff33d9798fe933785b8a0f20aa179ca1d18418fd401d955a1270edf9542eb7b10833bbfe1de84b31dde6924e4e01f0e335174c45fede9f9e80ef')
			`),
		},
		{
			name: `stat: empty file`,
			script: itn.HereDoc(`
				load('file', 'stat')
				fp = 'testdata/empty.txt'
				s = stat(fp)
				assert.eq(s.name, 'empty.txt')
				assert.eq(s.size, 0)
				assert.eq(s.type, 'file')
				assert.eq(s.ext, '.txt')
			`),
		},
		{
			name: `stat: file no ext`,
			script: itn.HereDoc(`
				load('file', 'stat')
				fp = 'testdata/noext'
				s = stat(fp)
				assert.eq(s.type, 'file')
				assert.eq(s.name, 'noext')
				assert.eq(s.ext, '')
			`),
		},
		//{
		//	name: `stat: file dot ext`,
		//	script: itn.HereDoc(`
		//		load('file', 'stat')
		//		fp = 'testdata/dotext.'
		//		s = stat(fp)
		//		assert.eq(s.type, 'file')
		//		assert.eq(s.name, 'dotext.')
		//		assert.eq(s.ext, '.')
		//	`),
		//},
		{
			name: `stat: dir`,
			script: itn.HereDoc(`
				load('file', 'stat')
				fp = 'testdata'
				s = stat(fp)
				assert.eq(s.name, 'testdata')
				assert.eq(s.type, 'dir')
				assert.eq(s.ext, '')
				assert.true(s.path.endswith(fp))
				assert.true(s.modified.unix > 0)
			`),
		},
		{
			name: `stat: follow dir`,
			script: itn.HereDoc(`
				load('file', 'stat')
				fp = 'testdata'
				s = stat(fp, follow=True)
				assert.eq(s.name, 'testdata')
				assert.eq(s.type, 'dir')
				assert.eq(s.ext, '')
				assert.true(s.path.endswith(fp))
				assert.true(s.modified.unix > 0)
			`),
		},
		{
			name: `stat: device`,
			script: itn.HereDoc(`
				load('file', 'stat')
				fp = '/dev/null'
				s = stat(fp)
				assert.eq(s.name, 'null')
				assert.eq(s.path, fp)
				assert.eq(s.ext, '')
				assert.eq(s.type, 'char_device')
			`),
			skipWindows: true,
		},
		{
			name: `stat: dir get hash`,
			script: itn.HereDoc(`
				load('file', 'stat')
				fp = 'testdata'
				s = stat(fp)
				s.get_md5()
			`),
			wantErr:     `testdata: is a directory`,
			skipWindows: true,
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
			res, err := itn.ExecModuleWithErrorTest(t, lf.ModuleName, lf.LoadModule, tt.script, tt.wantErr, nil)
			if (err != nil) != (tt.wantErr != "") {
				t.Errorf("file(%q) expects error = '%v', actual error = '%v', result = %v", tt.name, tt.wantErr, err, res)
			}
			// check file content if needed
			if cont := tt.fileContent; cont != "" {
				b, err := os.ReadFile(tp)
				if err != nil {
					t.Errorf("os.ReadFile() expects no error, actual error = '%v'", err)
					return
				}
				if string(b) != cont {
					t.Errorf("file(%q) expects (%s) file content = %q, actual content = %q", tt.name, tp, cont, string(b))
				}
			}
		})
	}
}
