package hash_test

import (
	"errors"
	"testing"

	"github.com/1set/starlet/lib/hash"
	itn "github.com/1set/starlet/lib/internal"
)

func TestLoadModule_Hash(t *testing.T) {
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
				assert.eq(sha1(""), "da39a3ee5e6b4b0d3255bfef95601890afd80709")
				assert.eq(sha1("Aloha!"), "c3dd37312ba987e1cc40ae021bc202c4a52d8afe")
			`),
		},
		{
			name: `SHA256`,
			script: itn.HereDoc(`
				load('hash', 'sha256')
				assert.eq(sha256(""), "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855")
				assert.eq(sha256("Aloha!"), "dea7e28aee505f2dd033de1427a517793e38b7605e8fc24da40151907e52cea3")
			`),
		},
		{
			name: `SHA512`,
			script: itn.HereDoc(`
				load('hash', 'sha512')
				assert.eq(sha512(""), "cf83e1357eefb8bdf1542850d66d8007d620e4050b5715dc83f4a921d36ce9ce47d0d13c5d85f2b0ff8318d2877eec2f63b931bd47417a81a538327af927da3e")
				assert.eq(sha512("Aloha!"), "d9cb95ad9d916a0781b3339424d5eb11c476405dfba7af7fabf4981fdd3291c27e8006e4cca617beae70dd00ab86a0213c44ed461229b16b45db45f64691049e")
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
			res, err := itn.ExecModuleWithErrorTest(t, hash.ModuleName, hash.LoadModule, tt.script, tt.wantErr)
			if (err != nil) != (tt.wantErr != nil) {
				t.Errorf("hash(%q) expects error = '%v', actual error = '%v', result = %v", tt.name, tt.wantErr, err, res)
				return
			}
		})
	}
}
