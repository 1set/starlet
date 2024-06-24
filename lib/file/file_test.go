package file_test

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"testing"

	itn "github.com/1set/starlet/internal"
	lf "github.com/1set/starlet/lib/file"
	"go.starlark.net/starlark"
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
			name: `read json no args`,
			script: itn.HereDoc(`
				load('file', 'read_json')
				j1 = read_json()
			`),
			wantErr: `file.read_json: missing argument for name`,
		},
		{
			name: `read json not found`,
			script: itn.HereDoc(`
				load('file', 'read_json')
				j1 = read_json("no-such-file")
			`),
			wantErr: `open no-such-file`,
		},
		{
			name: `read broken json`,
			script: itn.HereDoc(`
				load('file', 'read_json')
				j1 = read_json('testdata/1line.txt')
			`),
			wantErr: `json.decode: at offset`,
		},
		{
			name: `read json`,
			script: itn.HereDoc(`
				load('file', 'read_json')
				j1 = read_json('testdata/json1.json')
				assert.true(type(j1) == "dict")
				assert.eq(j1["num"], 42)
				assert.eq(j1["undef"], None)
				assert.eq(j1["bool"], True)
				assert.eq(j1["arr"], [1,2,3])
				assert.eq(j1["obj"], {"foo": "bar", "baz": "qux"})
			`),
		},
		{
			name: `read jsonl no args`,
			script: itn.HereDoc(`
				load('file', 'read_jsonl')
				j1 = read_jsonl()
			`),
			wantErr: `file.read_jsonl: missing argument for name`,
		},
		{
			name: `read jsonl not found`,
			script: itn.HereDoc(`
				load('file', 'read_jsonl')
				j1 = read_jsonl("no-such-file")
			`),
			wantErr: `open no-such-file`,
		},
		{
			name: `read broken jsonl`,
			script: itn.HereDoc(`
				load('file', 'read_jsonl')
				j1 = read_jsonl('testdata/1line.txt')
			`),
			wantErr: `line 1: json.decode: at offset`,
		},
		{
			name: `read jsonl`,
			script: itn.HereDoc(`
				load('file', 'read_jsonl')
				js = read_jsonl('testdata/json2.json')
				assert.eq(len(js), 3)
				assert.eq(js[-1], {"name": "Mike", "age": 32, "city": "Chicago", "opt": True})
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
			name: `write json no args`,
			script: itn.HereDoc(`
				load('file', 'write_json')
				write_json()
			`),
			wantErr: `file.write_json: missing argument for name`,
		},
		{
			name: `write json no data`,
			script: itn.HereDoc(`
				load('file', 'write_json')
				fp = %q
				write_json(fp)
			`),
			wantErr: `file.write_json: missing argument for data`,
		},
		{
			name: `write json invalid data`,
			script: itn.HereDoc(`
				load('file', 'write_json')
				fp = %q
				write_json(fp, lambda x: x*2)
			`),
			wantErr: `json.encode: cannot encode function as JSON`,
		},
		{
			name: `write json string`,
			script: itn.HereDoc(`
				load('file', 'write_json')
				fp = %q
				write_json(fp, "abc")
			`),
			fileContent: `abc`,
		},
		{
			name: `write json bytes`,
			script: itn.HereDoc(`
				load('file', 'write_json')
				fp = %q
				write_json(fp, b"123")
			`),
			fileContent: `123`,
		},
		{
			name: `write json dict`,
			script: itn.HereDoc(`
				load('file', 'write_json')
				fp = %q
				write_json(fp, {"b": True})
				write_json(fp, {"a": 520})
			`),
			fileContent: `{"a":520}`,
		},
		{
			name: `append json no args`,
			script: itn.HereDoc(`
				load('file', 'append_json')
				append_json()
			`),
			wantErr: `file.append_json: missing argument for name`,
		},
		{
			name: `append json no data`,
			script: itn.HereDoc(`
				load('file', 'append_json')
				fp = %q
				append_json(fp)
			`),
			wantErr: `file.append_json: missing argument for data`,
		},
		{
			name: `append json invalid data`,
			script: itn.HereDoc(`
				load('file', 'append_json')
				fp = %q
				append_json(fp, lambda x: x*2)
			`),
			wantErr: `json.encode: cannot encode function as JSON`,
		},
		{
			name: `append json string`,
			script: itn.HereDoc(`
				load('file', 'append_json')
				fp = %q
				append_json(fp, "abc")
				append_json(fp, "ABC")
			`),
			fileContent: `abcABC`,
		},
		{
			name: `append json bytes`,
			script: itn.HereDoc(`
				load('file', 'append_json')
				fp = %q
				append_json(fp, b"123")
				append_json(fp, b"456")
			`),
			fileContent: `123456`,
		},
		{
			name: `append json dict`,
			script: itn.HereDoc(`
				load('file', 'append_json')
				fp = %q
				append_json(fp, {"a": 520})
				append_json(fp, {"b": 1==1})
			`),
			fileContent: `{"a":520}{"b":true}`,
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
			name: `write jsonl no args`,
			script: itn.HereDoc(`
				load('file', 'write_jsonl')
				write_jsonl()
			`),
			wantErr: `file.write_jsonl: missing argument for name`,
		},
		{
			name: `write jsonl no data`,
			script: itn.HereDoc(`
				load('file', 'write_jsonl')
				fp = %q
				write_jsonl(fp)
			`),
			wantErr: `file.write_jsonl: missing argument for data`,
		},
		{
			name: `write jsonl invalid data`,
			script: itn.HereDoc(`
				load('file', 'write_jsonl')
				fp = %q
				write_jsonl(fp, lambda x: x*2)
			`),
			wantErr: `json.encode: cannot encode function as JSON`,
		},
		{
			name: `write jsonl string`,
			script: itn.HereDoc(`
				load('file', 'write_jsonl')
				fp = %q
				write_jsonl(fp, "abc")
			`),
			fileContent: "abc\n",
		},
		{
			name: `write jsonl bytes`,
			script: itn.HereDoc(`
				load('file', 'write_jsonl')
				fp = %q
				write_jsonl(fp, b"123")
			`),
			fileContent: "123\n",
		},
		{
			name: `write jsonl dict`,
			script: itn.HereDoc(`
				load('file', 'write_jsonl')
				fp = %q
				write_jsonl(fp, {"b": True})
				write_jsonl(fp, {"a": 520})
			`),
			fileContent: "{\"a\":520}\n",
		},
		{
			name: `write jsonl list`,
			script: itn.HereDoc(`
				load('file', 'write_jsonl')
				fp = %q
				l = [{"a": 520}, {"b": True}]
				write_jsonl(fp, l)
			`),
			fileContent: "{\"a\":520}\n{\"b\":true}\n",
		},
		{
			name: `write jsonl tuple`,
			script: itn.HereDoc(`
				load('file', 'write_jsonl')
				fp = %q
				l = ({"a": 520}, {"b": True}, {"c": "hello"})
				write_jsonl(fp, l)
			`),
			fileContent: "{\"a\":520}\n{\"b\":true}\n{\"c\":\"hello\"}\n",
		},
		{
			name: `write jsonl set`,
			script: itn.HereDoc(`
				load('file', 'write_jsonl')
				fp = %q
				write_jsonl(fp, set({"a": 520}))
			`),
			fileContent: "\"a\"\n",
		},
		{
			name: `append jsonl no args`,
			script: itn.HereDoc(`
				load('file', 'append_jsonl')
				append_jsonl()
			`),
			wantErr: `file.append_jsonl: missing argument for name`,
		},
		{
			name: `append jsonl no data`,
			script: itn.HereDoc(`
				load('file', 'append_jsonl')
				fp = %q
				append_jsonl(fp)
			`),
			wantErr: `file.append_jsonl: missing argument for data`,
		},
		{
			name: `append jsonl invalid data`,
			script: itn.HereDoc(`
				load('file', 'append_jsonl')
				fp = %q
				append_jsonl(fp, [lambda x: x*2])
			`),
			wantErr: `json.encode: cannot encode function as JSON`,
		},
		{
			name: `append jsonl string`,
			script: itn.HereDoc(`
				load('file', 'append_jsonl')
				fp = %q
				append_jsonl(fp, "abc")
			`),
			fileContent: "abc\n",
		},
		{
			name: `append jsonl bytes`,
			script: itn.HereDoc(`
				load('file', 'append_jsonl')
				fp = %q
				append_jsonl(fp, b"123")
			`),
			fileContent: "123\n",
		},
		{
			name: `append jsonl dict`,
			script: itn.HereDoc(`
				load('file', 'append_jsonl')
				fp = %q
				append_jsonl(fp, {"b": True})
				append_jsonl(fp, {"a": 520})
			`),
			fileContent: "{\"b\":true}\n{\"a\":520}\n",
		},
		{
			name: `append jsonl list`,
			script: itn.HereDoc(`
				load('file', 'append_jsonl')
				fp = %q
				l = [{"a": 520}, {"b": True}]
				append_jsonl(fp, l)
			`),
			fileContent: "{\"a\":520}\n{\"b\":true}\n",
		},
		{
			name: `append jsonl tuple`,
			script: itn.HereDoc(`
				load('file', 'append_jsonl')
				fp = %q
				l = ({"a": 520}, {"b": True}, {"c": "hello"})
				append_jsonl(fp, l)
			`),
			fileContent: "{\"a\":520}\n{\"b\":true}\n{\"c\":\"hello\"}\n",
		},
		{
			name: `append jsonl set`,
			script: itn.HereDoc(`
				load('file', 'append_jsonl')
				fp = %q
				append_jsonl(fp, set({"a": 520}))
				append_jsonl(fp, set({"c": 800}))
			`),
			fileContent: "\"a\"\n\"c\"\n",
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
			name: `stat: symlink`,
			script: itn.HereDoc(`
				load('file', 'stat')
				fp = '/dev/stdin'
				s = stat(fp)
				assert.eq(s.name, 'stdin')
				assert.eq(s.type, 'symlink')
				assert.eq(s.ext, '')
			`),
			skipWindows: true,
		},
		{
			name: `stat: char device`,
			script: itn.HereDoc(`
				load('file', 'stat')
				fp = '/dev/null'
				s = stat(fp)
				assert.eq(s.name, 'null')
				assert.eq(s.path, fp)
				assert.eq(s.ext, '')
				assert.eq(s.type, 'char')
			`),
			skipWindows: true,
		},
		//{
		//	name: `stat: block device`,
		//	script: itn.HereDoc(`
		//		load('file', 'stat')
		//		fp = '/dev/disk0'
		//		s = stat(fp)
		//		assert.eq(s.name, 'disk0')
		//		assert.eq(s.path, fp)
		//		assert.eq(s.ext, '')
		//		assert.eq(s.type, 'block')
		//	`),
		//	skipWindows: true,
		//},
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
		{
			name: `stat: no perm`,
			script: itn.HereDoc(`
				load('file', 'stat')
				fp = '/etc/sudoers'
				s = stat(fp)
				s.get_md5()
			`),
			wantErr:     `get_md5: open /etc/sudoers`,
			skipWindows: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// skip windows
			if isOnWindows && tt.skipWindows {
				t.Skipf("Skip test on Windows")
				return
			}

			// prepare temp file if needed
			var (
				tp      string
				td      string
				predecl starlark.StringDict
			)
			if strings.Contains(tt.script, "%q") {
				tf, err := os.CreateTemp("", "starlet-file-test-write")
				if err != nil {
					t.Errorf("os.CreateTemp() expects no error, actual error = '%v'", err)
					return
				}
				tp = tf.Name()
				//t.Logf("Temp file to write: %s", tp)
				tt.script = fmt.Sprintf(tt.script, tp)

				td, err = os.MkdirTemp("", "starlet-file-test-dir")
				if err != nil {
					t.Errorf("os.MkdirTemp() expects no error, actual error = '%v'", err)
					return
				}

				predecl = starlark.StringDict{
					"temp_file": starlark.String(tp),
					"temp_dir":  starlark.String(td),
				}
			}

			// execute test
			res, err := itn.ExecModuleWithErrorTest(t, lf.ModuleName, lf.LoadModule, tt.script, tt.wantErr, predecl)
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
