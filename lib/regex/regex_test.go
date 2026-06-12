package regex_test

import (
	"testing"

	itn "github.com/1set/starlet/internal"
	"github.com/1set/starlet/lib/regex"
)

func TestLoadModule_Regex(t *testing.T) {
	tests := []struct {
		name    string
		script  string
		wantErr string
	}{
		{
			name: `search: match and none`,
			script: itn.HereDoc(`
				load('regex', 'search')
				m = search(r'(\w+)@(\w+)', 'reach ann@host now')
				assert.eq(m.group(0), 'ann@host')
				assert.eq(m.group(1), 'ann')
				assert.eq(m.group(2), 'host')
				assert.eq(m.span(0), (6, 14))
				assert.eq(search('zzz', 'abc'), None)
			`),
		},
		{
			name: `match: anchored at start`,
			script: itn.HereDoc(`
				load('regex', 'match')
				assert.eq(match('hello', 'hello world').group(0), 'hello')
				assert.eq(match('world', 'hello world'), None)
				assert.eq(match('(a)(b)', 'abc').group(1), 'a')
			`),
		},
		{
			name: `fullmatch: whole string, with backtracking`,
			script: itn.HereDoc(`
				load('regex', 'fullmatch')
				assert.eq(fullmatch('a|ab', 'ab').group(0), 'ab')
				assert.eq(fullmatch('a+', 'aaa').group(0), 'aaa')
				assert.eq(fullmatch('a+', 'aaab'), None)
			`),
		},
		{
			name: `findall: python shaping by group count`,
			script: itn.HereDoc(`
				load('regex', 'findall')
				assert.eq(findall(r'\d+', 'a1 b22 c333'), ('1', '22', '333'))
				assert.eq(findall(r'(\w)(\d)', 'a1 b2'), (('a', '1'), ('b', '2')))
				assert.eq(findall(r'x(\d)', 'x1 x2'), ('1', '2'))
				assert.eq(findall('z', 'abc'), ())
			`),
		},
		{
			name: `finditer: list of matches`,
			script: itn.HereDoc(`
				load('regex', 'finditer')
				ms = finditer(r'\d+', 'a1 b22')
				assert.eq(len(ms), 2)
				assert.eq([m.group(0) for m in ms], ['1', '22'])
				assert.eq([m.span(0) for m in ms], [(1, 2), (4, 6)])
			`),
		},
		{
			name: `named groups: group by name, groupdict`,
			script: itn.HereDoc(`
				load('regex', 'search')
				m = search(r'(?P<user>\w+)@(?P<host>\w+)', 'ann@example')
				assert.eq(m.group('user'), 'ann')
				assert.eq(m.group('host'), 'example')
				assert.eq(m.groupdict(), {'user': 'ann', 'host': 'example'})
				assert.eq(m.group('user', 'host'), ('ann', 'example'))
			`),
		},
		{
			name: `groups: optional group is None or default`,
			script: itn.HereDoc(`
				load('regex', 'search')
				m = search(r'(a)(b)?', 'a')
				assert.eq(m.groups(), ('a', None))
				assert.eq(m.groups('X'), ('a', 'X'))
			`),
		},
		{
			name: `flags: ignorecase, multiline, dotall`,
			script: itn.HereDoc(`
				load('regex', 'search', 'findall', 'I', 'M', 'S')
				assert.eq(search('hello', 'HELLO', I).group(0), 'HELLO')
				assert.eq(findall('^x', 'x\nx\ny', M), ('x', 'x'))
				assert.eq(search('a.b', 'a\nb', S).group(0), 'a\nb')
				assert.eq(search('a.b', 'a\nb'), None)
			`),
		},
		{
			name: `sub: backreference translation and count`,
			script: itn.HereDoc(`
				load('regex', 'sub')
				assert.eq(sub(r'(\w+)@(\w+)', r'\2.\1', 'ann@host'), 'host.ann')
				assert.eq(sub(r'(?P<x>\d)', r'[\g<x>]', 'a1b2'), 'a[1]b[2]')
				assert.eq(sub('a', 'X', 'aaa', 2), 'XXa')
				assert.eq(sub('a', 'X', 'aaa', -1), 'aaa')
				assert.eq(sub('a', '$', 'a'), '$')
			`),
		},
		{
			name: `sub: function replacement`,
			script: itn.HereDoc(`
				load('regex', 'sub')
				def up(m):
					return m.group(0).upper()
				assert.eq(sub(r'[a-z]+', up, 'aa bb'), 'AA BB')
			`),
		},
		{
			name: `subn: returns count`,
			script: itn.HereDoc(`
				load('regex', 'subn')
				assert.eq(subn('a', 'X', 'aaa'), ('XXX', 3))
				assert.eq(subn('z', 'X', 'aaa'), ('aaa', 0))
			`),
		},
		{
			name: `split: includes capture groups, maxsplit`,
			script: itn.HereDoc(`
				load('regex', 'split')
				assert.eq(split(r'\s+', 'a b  c'), ('a', 'b', 'c'))
				assert.eq(split(r'(\s+)', 'a b'), ('a', ' ', 'b'))
				assert.eq(split(',', 'a,b,c', 1), ('a', 'b,c'))
			`),
		},
		{
			name: `escape: quote meta`,
			script: itn.HereDoc(`
				load('regex', 'escape', 'search')
				p = escape('a.b*c')
				assert.eq(search(p, 'a.b*c').group(0), 'a.b*c')
				assert.eq(search(p, 'axbyc'), None)
			`),
		},
		{
			name: `compile: pattern object and attrs`,
			script: itn.HereDoc(`
				load('regex', 'compile')
				p = compile(r'(?P<n>\d+)')
				assert.eq(p.pattern, r'(?P<n>\d+)')
				assert.eq(p.groups, 1)
				assert.eq(p.search('x42').group('n'), '42')
				assert.eq(p.findall('1 2 3'), ('1', '2', '3'))
				assert.eq(p.sub('#', 'a1b2'), 'a#b#')
				assert.true(p.match('5x') != None)
				assert.eq(p.match('x5'), None)
			`),
		},
		{
			name: `expand: template on a match`,
			script: itn.HereDoc(`
				load('regex', 'search')
				m = search(r'(\w+) (\w+)', 'hello world')
				assert.eq(m.expand(r'\2 \1'), 'world hello')
			`),
		},
		{
			name: `start end span`,
			script: itn.HereDoc(`
				load('regex', 'search')
				m = search(r'b(c)', 'abcd')
				assert.eq(m.start(0), 1)
				assert.eq(m.end(0), 3)
				assert.eq(m.start(1), 2)
				assert.eq(m.span(1), (2, 3))
			`),
		},
		{
			name: `match attrs: string and re`,
			script: itn.HereDoc(`
				load('regex', 'search', 'compile')
				m = search('b', 'abc')
				assert.eq(m.string, 'abc')
				assert.eq(m.re.pattern, 'b')
			`),
		},
		// error cases
		{
			name: `error: invalid pattern`,
			script: itn.HereDoc(`
				load('regex', 'compile')
				compile('(')
			`),
			wantErr: `cannot compile pattern`,
		},
		{
			name: `error: lookbehind unsupported by RE2`,
			script: itn.HereDoc(`
				load('regex', 'compile')
				compile(r'(?<=x)y')
			`),
			wantErr: `cannot compile pattern`,
		},
		{
			name: `error: unsupported flags value`,
			script: itn.HereDoc(`
				load('regex', 'search')
				search('a', 'a', 1024)
			`),
			wantErr: `unsupported flags value`,
		},
		{
			name: `error: no such group`,
			script: itn.HereDoc(`
				load('regex', 'search')
				search('(a)', 'a').group(5)
			`),
			wantErr: `no such group: 5`,
		},
		{
			name: `error: bad repl type`,
			script: itn.HereDoc(`
				load('regex', 'sub')
				sub('a', 123, 'a')
			`),
			wantErr: `repl must be a string or a function`,
		},
		{
			name: `try_compile: ok and error`,
			script: itn.HereDoc(`
				load('regex', 'try_compile')
				p, err = try_compile('a+')
				assert.eq(err, None)
				assert.true(p != None)
				bad, err2 = try_compile('(')
				assert.eq(bad, None)
				assert.true('cannot compile' in err2)
			`),
		},
		{
			name: `try_search: error captured`,
			script: itn.HereDoc(`
				load('regex', 'try_search')
				res, err = try_search('(', 'abc')
				assert.eq(res, None)
				assert.true('cannot compile' in err)
			`),
		},
		{
			name: `value protocol: str, type, bool, dir`,
			script: itn.HereDoc(`
				load('regex', 'compile', 'search')
				p = compile('a')
				assert.eq(type(p), 'regex.Pattern')
				assert.true(bool(p))
				assert.true('search' in dir(p))
				assert.true('pattern' in dir(p))
				assert.true('compile' in str(p))
				m = search('b', 'abc')
				assert.eq(type(m), 'regex.Match')
				assert.true(bool(m))
				assert.true('group' in dir(m))
				assert.true('string' in dir(m))
				assert.true('Match' in str(m))
			`),
		},
		{
			name: `match is unhashable`,
			script: itn.HereDoc(`
				load('regex', 'search')
				m = search('a', 'a')
				d = {}
				d[m] = 1
			`),
			wantErr: `unhashable type: regex.Match`,
		},
		{
			name: `sub: repl escape sequences`,
			script: itn.HereDoc(`
				load('regex', 'sub')
				assert.eq(sub('X', '\\n', 'aXb'), 'a\nb')
				assert.eq(sub('X', '\\t', 'aXb'), 'a\tb')
				assert.eq(sub('X', '\\\\', 'aXb'), 'a\\b')
				assert.eq(sub('X', 'lone\\q', 'X'), 'lone\\q')
			`),
		},
		{
			name: `findall via pattern object`,
			script: itn.HereDoc(`
				load('regex', 'compile')
				p = compile(r'(\w)(\d)')
				assert.eq(p.findall('a1 b2'), (('a', '1'), ('b', '2')))
				assert.eq(p.finditer('a1')[0].group(2), '1')
				assert.eq(p.split('a1xb2', 1)[0], '')
				assert.eq(p.subn('#', 'a1b2'), ('##', 2))
				assert.eq(p.fullmatch('a1').group(0), 'a1')
			`),
		},
		{
			name: `error: missing arguments`,
			script: itn.HereDoc(`
				load('regex', 'search')
				search()
			`),
			wantErr: `missing argument`,
		},
		{
			name: `error: sub function returns non-string`,
			script: itn.HereDoc(`
				load('regex', 'sub')
				def bad(m):
					return 42
				sub('a', bad, 'a')
			`),
			wantErr: `repl function must return a string`,
		},
		{
			name: `error: group by bad name`,
			script: itn.HereDoc(`
				load('regex', 'search')
				search('(a)', 'a').group('nope')
			`),
			wantErr: `no such group: "nope"`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := itn.ExecModuleWithErrorTest(t, regex.ModuleName, regex.LoadModule, tt.script, tt.wantErr, nil)
			if (err != nil) != (tt.wantErr != "") {
				t.Errorf("regex(%q) expects error = '%v', actual error = '%v', result = %v", tt.name, tt.wantErr, err, res)
			}
		})
	}
}
