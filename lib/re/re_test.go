package re_test

import (
	"testing"

	itn "github.com/1set/starlet/internal"
	"github.com/1set/starlet/lib/re"
)

func TestLoadModule_Re(t *testing.T) {
	tests := []struct {
		name    string
		script  string
		wantErr string
	}{
		{
			name: `match`,
			script: itn.HereDoc(`
			load('re', 'match')
			match_pattern = r"(\w*)\s*(ADD|REM|DEL|EXT|TRF)\s*(.*)\s*(NAT|INT)\s*(.*)\s*(\(\w{2}\))\s*(.*)"
			match_test = "EDM ADD FROM INJURED NAT Jordan BEAULIEU (DB) Western University"
			
			assert.eq(match(match_pattern,match_test), [(match_test, "EDM", "ADD", "FROM INJURED ", "NAT", "Jordan BEAULIEU ", "(DB)", "Western University")])
			assert.eq(match(match_pattern,"what"), [])
			`),
		},
		{
			name: `match error pattern`,
			script: itn.HereDoc(`
			load('re', 'match')
			match(123, "foo")
			`),
			wantErr: `re.match: for parameter pattern: got int, want string`,
		},
		{
			name: `match error string`,
			script: itn.HereDoc(`
			load('re', 'match')
			match("foobar", 2)
			`),
			wantErr: `re.match: for parameter string: got int, want string`,
		},
		{
			name: `search`,
			script: itn.HereDoc(`
			load('re', 'search')
			match_pattern = r"(\w*)\s*(ADD|REM|DEL|EXT|TRF)\s*(.*)\s*(NAT|INT)\s*(.*)\s*(\(\w{2}\))\s*(.*)"
			match_test = "EDM ADD FROM INJURED NAT Jordan BEAULIEU (DB) Western University"
			assert.eq(search(match_pattern, match_test), [0, 64])
			assert.eq(search(match_pattern, "what"), None)
			`),
		},
		{
			name: `search error`,
			script: itn.HereDoc(`
			load('re', 'search')
			search(123, "foo")
			`),
			wantErr: `re.search: for parameter pattern: got int, want string`,
		},
		{
			name: `compile`,
			script: itn.HereDoc(`
			load('re', 'compile')
			match_pattern = r"(\w*)\s*(ADD|REM|DEL|EXT|TRF)\s*(.*)\s*(NAT|INT)\s*(.*)\s*(\(\w{2}\))\s*(.*)"
			match_test = "EDM ADD FROM INJURED NAT Jordan BEAULIEU (DB) Western University"
			
			match_r = compile(match_pattern)
			assert.eq(match_r.match(match_test), [(match_test, "EDM", "ADD", "FROM INJURED ", "NAT", "Jordan BEAULIEU ", "(DB)", "Western University")])
			assert.eq(match_r.match("ab acdef"), [])
			assert.eq(match_r.sub("", match_test), "")
			`),
		},
		{
			name: `compile error`,
			script: itn.HereDoc(`
			load('re', 'compile')
			compile(123)
			`),
			wantErr: `re.compile: for parameter pattern: got int, want string`,
		},
		{
			name: `compile fail`,
			script: itn.HereDoc(`
			load('re', 'compile')
			compile("\q")
			`),
			wantErr: `re_test.star:3:9: invalid escape sequence \q`,
		},
		{
			name: `sub`,
			script: itn.HereDoc(`
			load('re', 'sub')
			match_pattern = r"(\w*)\s*(ADD|REM|DEL|EXT|TRF)\s*(.*)\s*(NAT|INT)\s*(.*)\s*(\(\w{2}\))\s*(.*)"
			match_test = "EDM ADD FROM INJURED NAT Jordan BEAULIEU (DB) Western University"
			
			assert.eq(sub(match_pattern, "", match_test), "")
			assert.eq(sub(match_pattern, "", "ab acdef"), "ab acdef")
			`),
		},
		{
			name: `sub error`,
			script: itn.HereDoc(`
			load('re', 'sub')
			sub(123, "", "foo")
			`),
			wantErr: `re.sub: for parameter pattern: got int, want string`,
		},
		{
			name: `split`,
			script: itn.HereDoc(`
			load('re', 'split', 'compile')
			space_r = compile(" ")
			assert.eq(split(" ", "foo bar baz bat"), ("foo", "bar", "baz", "bat"))
			assert.eq(space_r.split("foo bar baz bat"), ("foo", "bar", "baz", "bat"))
			assert.eq(split(" ", "foobar"), ("foobar",))
			`),
		},
		{
			name: `split error`,
			script: itn.HereDoc(`
			load('re', 'split')
			split(123, "foo")
			`),
			wantErr: `re.split: for parameter pattern: got int, want string`,
		},
		{
			name: `findall`,
			script: itn.HereDoc(`
			load('re', 'compile', 'findall')
			foo_r = compile("foo")
			assert.eq(findall("foo", "foo bar baz"), ("foo",))
			assert.eq(foo_r.findall("foo bar baz"), ("foo",))
			assert.eq(findall("foo", "bar baz"), ())
			`),
		},
		{
			name: `findall error`,
			script: itn.HereDoc(`
			load('re', 'findall')
			findall(123, "foo")
			`),
			wantErr: `re.findall: for parameter pattern: got int, want string`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := itn.ExecModuleWithErrorTest(t, re.ModuleName, re.LoadModule, tt.script, tt.wantErr, nil)
			if (err != nil) != (tt.wantErr != "") {
				t.Errorf("re(%q) expects error = '%v', actual error = '%v', result = %v", tt.name, tt.wantErr, err, res)
				return
			}
		})
	}
}
