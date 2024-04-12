package atom_test

import (
	"testing"

	itn "github.com/1set/starlet/internal"
	libatom "github.com/1set/starlet/lib/atom"
)

func TestLoadModule_Atom(t *testing.T) {
	tests := []struct {
		name    string
		script  string
		wantErr string
	}{
		// for integer
		{
			name: `int: missing attrs`,
			script: itn.HereDoc(`
				load('atom', 'new_int')
				x = new_int(1)
				x.guess()
			`),
			wantErr: "atom_int has no .guess field or method",
		},
		{
			name: `int: default`,
			script: itn.HereDoc(`
				load('atom', 'new_int')
				x = new_int()
				assert.eq(x.get(), 0)
				assert.eq(type(x), 'atom_int')
				assert.eq(str(x), '<atom_int:0>')
				assert.eq(dir(x), ["add", "cas", "dec", "get", "inc", "set", "sub"])
				assert.true(not bool(x))
				
				m = {}
				m[x] = 1
				# assert.eq(m[x], 1)
			`),
		},
		{
			name: `int: new with args`,
			script: itn.HereDoc(`
				load('atom', 'new_int')
				x = new_int(42)
				assert.eq(str(x), '<atom_int:42>')
				assert.eq(x.get(), 42)
				assert.true(bool(x))
			`),
		},
		{
			name: `int: new invalid args`,
			script: itn.HereDoc(`
				load('atom', 'new_int')
				x = new_int('42')
			`),
			wantErr: "new_int: for parameter value: got string, want int",
		},
		{
			name: `int: set invalid args`,
			script: itn.HereDoc(`
				load('atom', 'new_int')
				x = new_int(1)
				x.set('2')
			`),
			wantErr: "set: for parameter value: got string, want int",
		},
		{
			name: `int: get with args`,
			script: itn.HereDoc(`
				load('atom', 'new_int')
				x = new_int(1)
				x.get(2)
			`),
			wantErr: "get: got 1 arguments, want 0",
		},
		{
			name: `int: inc with args`,
			script: itn.HereDoc(`
				load('atom', 'new_int')
				x = new_int(1)
				x.inc(2)
			`),
			wantErr: "inc: got 1 arguments, want 0",
		},
		{
			name: `int: dec with args`,
			script: itn.HereDoc(`
				load('atom', 'new_int')
				x = new_int(1)	
				x.dec(2)	
			`),
			wantErr: "dec: got 1 arguments, want 0",
		},
		{
			name: `int: add invalid args`,
			script: itn.HereDoc(`
				load('atom', 'new_int')
				x = new_int(1)	
				x.add('2')
			`),
			wantErr: "add: for parameter delta: got string, want int",
		},
		{
			name: `int: sub invalid args`,
			script: itn.HereDoc(`
				load('atom', 'new_int')
				x = new_int(1)
				x.sub('2')
			`),
			wantErr: "sub: for parameter delta: got string, want int",
		},
		{
			name: `int: cas invalid args`,
			script: itn.HereDoc(`
				load('atom', 'new_int')
				x = new_int(1)
				x.cas('1', 2)
			`),
			wantErr: "cas: for parameter old: got string, want int",
		},
		{
			name: `int: full`,
			script: itn.HereDoc(`
				load('atom', 'new_int')
				x = new_int(1)	
				assert.eq(x.get(), 1)	
				x.set(20)	
				assert.eq(x.get(), 20)
				assert.eq(x.add(5), 25)
				assert.eq(x.sub(3), 22)
				assert.eq(x.inc(), 23)
				assert.eq(x.dec(), 22)
				assert.eq(x.cas(22, 100), True)
				assert.eq(x.get(), 100)
				assert.eq(x.cas(22, 200), False)
				assert.eq(x.get(), 100)
			`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := itn.ExecModuleWithErrorTest(t, libatom.ModuleName, libatom.LoadModule, tt.script, tt.wantErr, nil)
			if (err != nil) != (tt.wantErr != "") {
				t.Errorf("atom(%q) expects error = '%v', actual error = '%v', result = %v", tt.name, tt.wantErr, err, res)
				return
			}
		})
	}
}
