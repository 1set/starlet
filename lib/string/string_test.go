package string_test

import (
	"testing"

	itn "github.com/1set/starlet/internal"
	ls "github.com/1set/starlet/lib/string"
)

func TestLoadModule_String(t *testing.T) {
	tests := []struct {
		name    string
		script  string
		wantErr string
	}{
		{
			name: `ascii_lowercase`,
			script: itn.HereDoc(`
				load('string', s='ascii_lowercase')
				assert.eq(s, "abcdefghijklmnopqrstuvwxyz")
			`),
		},
		{
			name: `ascii_uppercase`,
			script: itn.HereDoc(`
				load('string', s='ascii_uppercase')
				assert.eq(s, "ABCDEFGHIJKLMNOPQRSTUVWXYZ")
			`),
		},
		{
			name: `ascii_letters`,
			script: itn.HereDoc(`
				load('string', s='ascii_letters')
				assert.eq(s, "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
			`),
		},
		{
			name: `digits`,
			script: itn.HereDoc(`
				load('string', s='digits')
				assert.eq(s, "0123456789")
			`),
		},
		{
			name: `hexdigits`,
			script: itn.HereDoc(`
				load('string', s='hexdigits')
				assert.eq(s, "0123456789abcdefABCDEF")
			`),
		},
		{
			name: `octdigits`,
			script: itn.HereDoc(`
				load('string', s='octdigits')
				assert.eq(s, "01234567")
			`),
		},
		{
			name: `punctuation`,
			script: itn.HereDoc(`
				load('string', s='punctuation')
				assert.eq(s, r"""!"#$%&'()*+,-./:;<=>?@[\]^_{|}~` + "`" + `""")
				print('punctuation', s)
			`),
		},
		{
			name: `whitespace`,
			script: itn.HereDoc(`
				load('string', s='whitespace')
				assert.eq(s, ' \t\n\r\v\f')
			`),
		},
		{
			name: `printable`,
			script: itn.HereDoc(`
				load('string', s='printable')
				assert.eq(s, r"""0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ!"#$%&'()*+,-./:;<=>?@[\]^_{|}~` + "`" + `""" + ' \t\n\r\v\f')
				print('printable', s)
			`),
		},
		{
			name: `length`,
			script: itn.HereDoc(`
				load('string', 'length')
				assert.eq(length(''), 0)
				assert.eq(length('a'), 1)
				assert.eq(length('abc'), 3)
				assert.eq(length('我爱你'), 3)
				assert.eq(length("☕"), 1)
				assert.eq(length(b"☕"), 3)
				assert.eq(length([1, 2, "#", True, None]), 5)
			`),
		},
		{
			name: `length with missing args`,
			script: itn.HereDoc(`
				load('string', 'length')
				length()
			`),
			wantErr: `length() takes exactly one argument (0 given)`,
		},
		{
			name: `length with invalid args`,
			script: itn.HereDoc(`
				load('string', 'length')
				length(123)
			`),
			wantErr: `length() function isn't supported for 'int' type object`,
		},
		{
			name: `reverse`,
			script: itn.HereDoc(`
				load('string', 'reverse')
				assert.eq(reverse(''), '')
				assert.eq(reverse('a'), 'a')
				assert.eq(reverse('abc'), 'cba')
				assert.eq(reverse('我爱你'), '你爱我')
				assert.eq(reverse("☕"), "☕")
				assert.eq(reverse(b"☕"), b"\x95\x98\xe2")
			`),
		},
		{
			name: `reverse with missing args`,
			script: itn.HereDoc(`
				load('string', 'reverse')
				reverse()
			`),
			wantErr: `reverse() takes exactly one argument (0 given)`,
		},
		{
			name: `reverse with invalid args`,
			script: itn.HereDoc(`
				load('string', 'reverse')
				reverse(123)
			`),
			wantErr: `reverse() function isn't supported for 'int' type object`,
		},
		{
			name: `escape with missing args`,
			script: itn.HereDoc(`
				load('string', 'escape')
				escape()
			`),
			wantErr: `escape() takes exactly one argument (0 given)`,
		},
		{
			name: `escape with invalid args`,
			script: itn.HereDoc(`
				load('string', 'escape')
				escape(123)
			`),
			wantErr: `escape() function isn't supported for 'int' type object`,
		},
		{
			name: `escape`,
			script: itn.HereDoc(`
				load('string', 'escape')
				assert.eq(escape(''), '')
				assert.eq(escape('abc'), 'abc')
				assert.eq(escape("☕"), "☕")
				assert.eq(escape(b"☕"), b"☕")
				assert.eq(escape('我爱你'), '我爱你')
				assert.eq(escape('我&你'), '我&amp;你')
				assert.eq(escape('<&>'), '&lt;&amp;&gt;')
				assert.eq(escape(b'<&>'), b'&lt;&amp;&gt;')
			`),
		},
		{
			name: `unescape`,
			script: itn.HereDoc(`
				load('string', 'unescape')
				assert.eq(unescape(''), '')
				assert.eq(unescape('abc'), 'abc')
				assert.eq(unescape("☕"), "☕")
				assert.eq(unescape(b"☕"), b"☕")
				assert.eq(unescape('我爱你'), '我爱你')
				assert.eq(unescape('我&amp;你'), '我&你')
				assert.eq(unescape('&lt;&amp;&gt;'), '<&>')
				assert.eq(unescape(b'&lt;&amp;&gt;'), b'<&>')
			`),
		},
		{
			name: `quote`,
			script: itn.HereDoc(`
				load('string', 'quote')
				assert.eq(quote(''), '""')
				assert.eq(quote('abc'), '"abc"')
				assert.eq(quote("☕"), '"☕"')
				assert.eq(quote('我爱你'), '"我爱你"')
				assert.eq(quote('我&你'), '"我&你"')
				assert.eq(quote('<&>'), '"<&>"')
				assert.eq(quote(b'<&>'), b'"<&>"')
				assert.eq(quote('\n1'), '"\\n1"')
			`),
		},
		{
			name: `unquote`,
			script: itn.HereDoc(`
				load('string', 'unquote')
				assert.eq(unquote(''), '')
				assert.eq(unquote('""'), '')
				assert.eq(unquote('"abc"'), 'abc')
				assert.eq(unquote('"☕"'), '☕')
				assert.eq(unquote('我爱你'), '我爱你')
				assert.eq(unquote('"我爱你'), '"我爱你')
				assert.eq(unquote('我爱你"'), '我爱你"')
				assert.eq(unquote('"我爱你"'), '我爱你')
				assert.eq(unquote('"我&你"'), '我&你')
				assert.eq(unquote('"<&>"'), '<&>')
				assert.eq(unquote(b'"<&>"'), b'<&>')
				assert.eq(unquote('\\n1'), '\n1')
				print("{"+unquote('\\n1')+"}")
			`),
		},
		{
			name: `index`,
			script: itn.HereDoc(`
				load('string', 'index')
				assert.eq(index('hello', 'e'), 1)
				assert.eq(index('hello', 'o'), 4)
				assert.eq(index('你好世界', '好'), 1)
				assert.eq(index('你好世界', '界'), 3)
				assert.eq(index('a☕c', '☕'), 1)
			`),
		},
		{
			name: `index not found`,
			script: itn.HereDoc(`
				load('string', 'index')
				index('hello', 'x')
			`),
			wantErr: `substring not found`,
		},
		{
			name: `index with missing args`,
			script: itn.HereDoc(`
				load('string', 'index')
				index('hello')
			`),
			wantErr: `index: missing argument for sub`,
		},
		{
			name: `index with invalid args`,
			script: itn.HereDoc(`
				load('string', 'index')
				index(123, 'hello')
			`),
			wantErr: `index: for parameter s: got int, want string`,
		},
		{
			name: `rindex`,
			script: itn.HereDoc(`
				load('string', 'rindex')
				assert.eq(rindex('hello hello', 'e'), 7)
				assert.eq(rindex('hello hello', 'o'), 10)
				assert.eq(rindex('你好世界你好', '好'), 5)
				assert.eq(rindex('你好世界你好', '界'), 3)
				assert.eq(rindex('a☕c☕a', '☕'), 3)
			`),
		},
		{
			name: `rindex not found`,
			script: itn.HereDoc(`
				load('string', 'rindex')
				rindex('hello', 'x')
			`),
			wantErr: `substring not found`,
		},
		{
			name: `rindex with missing args`,
			script: itn.HereDoc(`
				load('string', 'rindex')
				rindex('hello')
			`),
			wantErr: `rindex: missing argument for sub`,
		},
		{
			name: `rindex with invalid args`,
			script: itn.HereDoc(`
				load('string', 'rindex')
				rindex(123, 'hello')
			`),
			wantErr: `rindex: for parameter s: got int, want string`,
		},
		{
			name: `find`,
			script: itn.HereDoc(`
				load('string', 'find')
				assert.eq(find('hello', 'e'), 1)
				assert.eq(find('hello', 'o'), 4)
				assert.eq(find('hello', 'x'), -1)
				assert.eq(find('你好世界', '好'), 1)
				assert.eq(find('你好世界', '界'), 3)
				assert.eq(find('你好世界', 'b'), -1)
				assert.eq(find('a☕c', '☕'), 1)
			`),
		},
		{
			name: `find with missing args`,
			script: itn.HereDoc(`
				load('string', 'find')
				find('hello')
			`),
			wantErr: `find: missing argument for sub`,
		},
		{
			name: `find with invalid args`,
			script: itn.HereDoc(`
				load('string', 'find')
				find(123, 'hello')
			`),
			wantErr: `find: for parameter s: got int, want string`,
		},
		{
			name: `rfind`,
			script: itn.HereDoc(`
				load('string', 'rfind')
				assert.eq(rfind('hello hello', 'e'), 7)
				assert.eq(rfind('hello hello', 'o'), 10)
				assert.eq(rfind('hello', 'x'), -1)
				assert.eq(rfind('你好世界你好', '好'), 5)
				assert.eq(rfind('你好世界你好', '界'), 3)
				assert.eq(rfind('你好世界', 'b'), -1)
				assert.eq(rfind('a☕c☕a', '☕'), 3)
			`),
		},
		{
			name: `rfind with missing args`,
			script: itn.HereDoc(`
				load('string', 'rfind')
				rfind('hello')
			`),
			wantErr: `rfind: missing argument for sub`,
		},
		{
			name: `rfind with invalid args`,
			script: itn.HereDoc(`
				load('string', 'rfind')
				rfind(123, 'hello')
			`),
			wantErr: `rfind: for parameter s: got int, want string`,
		},
		{
			name: `substring`,
			script: itn.HereDoc(`
				load('string', 'substring')
				assert.eq(substring('hello', 1, 4), 'ell')
				assert.eq(substring('hello', 1, -1), 'ell')
				assert.eq(substring('你好世界', 1, 3), '好世')
				assert.eq(substring('你好世界', 2, -1), '世')
				assert.eq(substring('a☕c', 1, 2), '☕')
				assert.eq(substring('a☕c', -2, -1), '☕')
			`),
		},
		{
			name: `substring with missing end`,
			script: itn.HereDoc(`
				load('string', 'substring')
				x = substring('hello', 1)
				assert.eq(x, 'ello')
			`),
		},
		{
			name: `substring with invalid args`,
			script: itn.HereDoc(`
				load('string', 'substring')
				substring(123, 1, 4)
			`),
			wantErr: `substring: for parameter s: got int, want string`,
		},
		{
			name: `substring with out of range args`,
			script: itn.HereDoc(`
				load('string', 'substring')
				substring('hello', 1, 6)
			`),
			wantErr: `substring: indices are out of range`,
		},
		{
			name: `substring with start greater than end`,
			script: itn.HereDoc(`
				load('string', 'substring')
				substring('hello', 3, 1)
			`),
			wantErr: `substring: indices are out of range`,
		},
		{
			name: `codepoint`,
			script: itn.HereDoc(`
				load('string', 'codepoint')
				assert.eq(codepoint('hello', 1), "e")
				assert.eq(codepoint('你好世界', 0), "你")
				assert.eq(codepoint('你好世界', 1), "好")
				assert.eq(codepoint('a☕c', 0), "a")
				assert.eq(codepoint('a☕c', 1), "☕")
				assert.eq(codepoint('a☕c', -1), "c")
			`),
		},
		{
			name: `codepoint with missing args`,
			script: itn.HereDoc(`
				load('string', 'codepoint')
				codepoint('hello')
			`),
			wantErr: `codepoint: missing argument for index`,
		},
		{
			name: `codepoint with invalid args`,
			script: itn.HereDoc(`
				load('string', 'codepoint')
				codepoint(123, 1)
			`),
			wantErr: `codepoint: for parameter s: got int, want string`,
		},
		{
			name: `codepoint with out of range index`,
			script: itn.HereDoc(`
				load('string', 'codepoint')
				codepoint('hello', 5)
			`),
			wantErr: `codepoint: index out of range`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := itn.ExecModuleWithErrorTest(t, ls.ModuleName, ls.LoadModule, tt.script, tt.wantErr, nil)
			if (err != nil) != (tt.wantErr != "") {
				t.Errorf("string(%q) expects error = '%v', actual error = '%v', result = %v", tt.name, tt.wantErr, err, res)
				return
			}
		})
	}
}
