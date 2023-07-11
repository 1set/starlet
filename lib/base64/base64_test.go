package base64_test

import (
	"testing"

	"github.com/1set/starlet/lib/base64"
	itn "github.com/1set/starlet/lib/internal"
)

func TestLoadModule_Base64(t *testing.T) {
	tests := []struct {
		name    string
		script  string
		wantErr string
	}{
		{
			name: `encode`,
			script: itn.HereDoc(`
				load('base64', 'encode')
				assert.eq(encode("hello"), "aGVsbG8=")
				assert.eq(encode("hello", encoding="standard_raw"), "aGVsbG8")
				assert.eq(encode("hello friend!", encoding="url"), "aGVsbG8gZnJpZW5kIQ==")
				assert.eq(encode("hello friend!", encoding="url_raw"), "aGVsbG8gZnJpZW5kIQ")
			`),
		},
		{
			name: `encode with invalid encoding`,
			script: itn.HereDoc(`
				load('base64', 'encode')
				encode("hello", encoding="invalid")
			`),
			wantErr: `unsupported encoding format: "invalid"`,
		},
		{
			name: `encode with invalid input`,
			script: itn.HereDoc(`
				load('base64', 'encode')
				encode(123)
			`),
			wantErr: `base64.encode: for parameter data: got int, want string or bytes`,
		},
		{
			name: `decode`,
			script: itn.HereDoc(`
				load('base64', 'decode')
				assert.eq(decode("aGVsbG8="),"hello")
				assert.eq(decode("aGVsbG8", encoding="standard_raw"),"hello")
				assert.eq(decode("aGVsbG8gZnJpZW5kIQ==", encoding="url"),"hello friend!")
				assert.eq(decode("aGVsbG8gZnJpZW5kIQ", encoding="url_raw"),"hello friend!")
			`),
		},
		{
			name: `decode with invalid encoding`,
			script: itn.HereDoc(`
				load('base64', 'decode')
				decode("aGVsbG8=", encoding="invalid")
			`),
			wantErr: `unsupported encoding format: "invalid"`,
		},
		{
			name: `decode with invalid input`,
			script: itn.HereDoc(`
				load('base64', 'decode')
				decode(123)
			`),
			wantErr: `base64.decode: for parameter data: got int, want string or bytes`,
		},
		{
			name: `decode fail`,
			script: itn.HereDoc(`
				load('base64', 'decode')
				decode("aGVsbG8")
			`),
			wantErr: `illegal base64 data at input byte 4`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := itn.ExecModuleWithErrorTest(t, base64.ModuleName, base64.LoadModule, tt.script, tt.wantErr)
			if (err != nil) != (tt.wantErr != "") {
				t.Errorf("base64(%q) expects error = '%v', actual error = '%v', result = %v", tt.name, tt.wantErr, err, res)
				return
			}
		})
	}
}
