package hash_test

import (
	"testing"

	"github.com/1set/starlet/lib/hash"
	itn "github.com/1set/starlet/lib/internal"
)

func TestLoadModule_Hash(t *testing.T) {
	code := itn.HereDoc(`
	load('assert.star', 'assert')
	load('hash', 'md5', 'sha1', 'sha256')

	assert.eq(md5(""), "d41d8cd98f00b204e9800998ecf8427e")
	assert.eq(md5("Aloha!"), "de424bf3e7dcba091c27d652ada485fb")
	assert.eq(sha1("Aloha!"), "c3dd37312ba987e1cc40ae021bc202c4a52d8afe")
	assert.eq(sha256("Aloha!"), "dea7e28aee505f2dd033de1427a517793e38b7605e8fc24da40151907e52cea3")
	`)

	_, err := itn.ExecModuleWithErrorTest(t, hash.ModuleName, hash.LoadModule, code, nil)
	if err != nil {
		return
	}
}
