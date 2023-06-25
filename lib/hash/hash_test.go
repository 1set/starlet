package hash_test

import (
	"testing"

	"github.com/1set/starlet/lib/hash"
	itn "github.com/1set/starlet/lib/internal"
)

func TestLoadModule_Hash(t *testing.T) {
	code := itn.HereDoc(`
	load('hash', 'hash')
	load('assert.star', 'assert')

	assert.eq(hash.md5("helloworld"), "fc5e038d38a57032085441e7fe7010b0")
	assert.eq(hash.sha1("helloworld"), "6adfb183a4a2c94a2f92dab5ade762a47889a5a1")
	assert.eq(hash.sha256("helloworld"), "936a185caaa266bb9cbe981e9e05cb78cd732b0b3280eb944412bb6f8f8f07af")
	`)

	_, err := itn.ExecModuleWithErrorTest(t, hash.ModuleName, hash.LoadModule, code, nil)
	if err != nil {
		return
	}
}
