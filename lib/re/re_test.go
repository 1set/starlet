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
		{
			name: `sub invalid pattern`,
			script: itn.HereDoc(`
			load('re', 'sub')
			sub('(', 'x', 'abc')
			`),
			wantErr: `missing closing`,
		},
		{
			name: `sub with count`,
			script: itn.HereDoc(`
			load('re', 'sub')
			assert.eq(sub('a', 'X', 'aaa', 1), 'Xaa')
			assert.eq(sub('a', 'X', 'aaa', count=2), 'XXa')
			assert.eq(sub('a', 'X', 'aaa', 9), 'XXX')
			assert.eq(sub('a', 'X', 'aaa', 0), 'XXX')
			assert.eq(sub('a', 'X', 'aaa', -1), 'aaa')
			`),
		},
		{
			name: `sub group template`,
			script: itn.HereDoc(`
			load('re', 'sub')
			assert.eq(sub('(a)(b)', '${2}${1}', 'ab ab'), 'ba ba')
			assert.eq(sub('a', '$$5', 'a'), '$5')
			`),
		},
		{
			name: `sub flags rejected`,
			script: itn.HereDoc(`
			load('re', 'sub')
			sub('a', 'b', 'c', 0, 2)
			`),
			wantErr: `re.sub: flags are not supported`,
		},
		{
			name: `split maxsplit`,
			script: itn.HereDoc(`
			load('re', 'split')
			assert.eq(split(',', 'a,b,c', 1), ('a', 'b,c'))
			assert.eq(split(',', 'a,b,c', maxsplit=2), ('a', 'b', 'c'))
			assert.eq(split(',', 'a,b,c', 99), ('a', 'b', 'c'))
			assert.eq(split(',', 'a,b,c', -1), ('a,b,c',))
			`),
		},
		{
			name: `split flags rejected`,
			script: itn.HereDoc(`
			load('re', 'split')
			split(',', 'a,b', 0, 1)
			`),
			wantErr: `re.split: flags are not supported`,
		},
		{
			name: `match anchored`,
			script: itn.HereDoc(`
			load('re', 'match')
			assert.eq(match('world', 'hello world'), [])
			assert.eq(match('hello', 'hello world'), [('hello',)])
			assert.eq(match('(h)(e)', 'hello'), [('he', 'h', 'e')])
			`),
		},
		{
			name: `findall groups`,
			script: itn.HereDoc(`
			load('re', 'findall')
			assert.eq(findall(r'(\w)(\d)', 'a1 b2'), (('a', '1'), ('b', '2')))
			assert.eq(findall(r'\w(\d)', 'a1 b2'), ('1', '2'))
			`),
		},
		{
			name: `findall flags rejected`,
			script: itn.HereDoc(`
			load('re', 'findall')
			findall('a', 'a', flags=1)
			`),
			wantErr: `re.findall: flags are not supported`,
		},
		{
			name: `compile flags rejected`,
			script: itn.HereDoc(`
			load('re', 'compile')
			compile('a', 1)
			`),
			wantErr: `re.compile: flags are not supported`,
		},
		{
			name: `compiled methods semantics`,
			script: itn.HereDoc(`
			load('re', 'compile')
			c = compile(',')
			assert.eq(c.split('a,b,c', 1), ('a', 'b,c'))
			a = compile('a')
			assert.eq(a.sub('X', 'aaa', 2), 'XXa')
			assert.eq(a.match('abc'), [('a',)])
			assert.eq(a.match('bca'), [])
			`),
		},
		{
			name: `compiled flags rejected`,
			script: itn.HereDoc(`
			load('re', 'compile')
			compile('a').findall('a', flags=1)
			`),
			wantErr: `findall: flags are not supported`,
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
