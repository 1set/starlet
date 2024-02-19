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
			`),
		},
		{
			name: `length with invalid args`,
			script: itn.HereDoc(`
				load('string', 'length')
				length(123)
			`),
			wantErr: `object of type 'int' has no length()`,
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
