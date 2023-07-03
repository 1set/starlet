package base64_test

import (
	"testing"

	"github.com/1set/starlet/lib/base64"
	itn "github.com/1set/starlet/lib/internal"
)

func TestLoadModule_Re(t *testing.T) {
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
			name: `decode`,
			script: itn.HereDoc(`
				load('base64', 'decode')
				assert.eq(decode("aGVsbG8="),"hello")
				assert.eq(decode("aGVsbG8", encoding="standard_raw"),"hello")
				assert.eq(decode("aGVsbG8gZnJpZW5kIQ==", encoding="url"),"hello friend!")
				assert.eq(decode("aGVsbG8gZnJpZW5kIQ", encoding="url_raw"),"hello friend!")
			`),
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
