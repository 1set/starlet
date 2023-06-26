package hash_test

import (
	"errors"
	"testing"

	"github.com/1set/starlet/lib/hash"
	itn "github.com/1set/starlet/lib/internal"
)

func TestLoadModule_Hash(t *testing.T) {
	header := `load('assert.star', 'assert')`
	tests := []struct {
		name    string
		script  string
		wantErr error
	}{
		{
			name: `MD5`,
			script: itn.HereDoc(`
				load('hash', 'md5')
				assert.eq(md5(""), "d41d8cd98f00b204e9800998ecf8427e")
				assert.eq(md5("Aloha!"), "de424bf3e7dcba091c27d652ada485fb")
			`),
		},
		{
			name: `SHA1`,
			script: itn.HereDoc(`
				load('hash', 'sha1')
				assert.eq(sha1("Aloha!"), "c3dd37312ba987e1cc40ae021bc202c4a52d8afe")
			`),
		},
		{
			name: `SHA256`,
			script: itn.HereDoc(`
				load('hash', 'sha256')
				assert.eq(sha256("Aloha!"), "dea7e28aee505f2dd033de1427a517793e38b7605e8fc24da40151907e52cea3")
			`),
		},
		{
			name: `Invalid Input Type`,
			script: itn.HereDoc(`
				load('hash', 'md5')
				md5(123)
				assert.fail("should not reach here")
			`),
			wantErr: errors.New("hash.md5: for parameter 1: got int, want string"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := itn.ExecModuleWithErrorTest(t, hash.ModuleName, hash.LoadModule, header+"\n"+tt.script, tt.wantErr)
			if (err != nil) != (tt.wantErr != nil) {
				t.Errorf("hash(%q) expects error = '%v', actual error = '%v', result = %v", tt.name, tt.wantErr, err, res)
				return
			}
		})
	}
}
