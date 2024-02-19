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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := itn.ExecModuleWithErrorTest(t, ls.ModuleName, ls.LoadModule, tt.script, tt.wantErr)
			if (err != nil) != (tt.wantErr != "") {
				t.Errorf("hash(%q) expects error = '%v', actual error = '%v', result = %v", tt.name, tt.wantErr, err, res)
				return
			}
		})
	}
}
