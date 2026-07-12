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
			name: `int: compare`,
			script: itn.HereDoc(`
				load('atom', 'new_int')
				x = new_int(1)
				y = new_int(1)
				z = new_int(2)
				assert.eq(x, y)
				assert.ne(x, z)
				assert.true(x == y)
				assert.true(x != z)
				assert.true(x < z)
				assert.true(x <= z)
				assert.true(z > x)
				assert.true(z >= x)
				assert.true(x >= y)
				assert.true(x <= y)
			`),
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
		{
			name: `int: usage`,
			script: itn.HereDoc(`
				load('atom', 'new_int')
				x = new_int()
				def work():
					x.inc()
				[work() for _ in range(10)]
				assert.eq(x.get(), 10)
			`),
		},

		// for float
		{
			name: `float: missing attrs`,
			script: itn.HereDoc(`
				load('atom', 'new_float')
				x = new_float(1)
				x.guess()
			`),
			wantErr: "atom_float has no .guess field or method",
		},
		{
			name: `float: default`,
			script: itn.HereDoc(`
				load('atom', 'new_float')
				x = new_float()
				assert.eq(x.get(), 0)
				assert.eq(type(x), 'atom_float')
				assert.eq(str(x), '<atom_float:0>')
				assert.eq(dir(x), ["add", "cas", "get", "set", "sub"])
				assert.true(not bool(x))
			`),
		},
		{
			name: `float: new with args`,
			script: itn.HereDoc(`
				load('atom', 'new_float')
				x = new_float(42.1)
				assert.eq(str(x), '<atom_float:42.1>')
				assert.eq(x.get(), 42.1)
				assert.true(bool(x))
			`),
		},
		{
			name: `float: new invalid args`,
			script: itn.HereDoc(`
				load('atom', 'new_float')
				x = new_float('42.1')
			`),
			wantErr: "new_float: for parameter value: got string, want float",
		},
		{
			name: `float: set invalid args`,
			script: itn.HereDoc(`
				load('atom', 'new_float')
				x = new_float(1)
				x.set('2')
			`),
			wantErr: "set: for parameter value: got string, want float",
		},
		{
			name: `float: get with args`,
			script: itn.HereDoc(`
				load('atom', 'new_float')
				x = new_float(1)
				x.get(2)
			`),
			wantErr: "get: got 1 arguments, want 0",
		},
		{
			name: `float: add invalid args`,
			script: itn.HereDoc(`
				load('atom', 'new_float')
				x = new_float(1)	
				x.add('2')
			`),
			wantErr: "add: for parameter delta: got string, want float",
		},
		{
			name: `float: sub invalid args`,
			script: itn.HereDoc(`
				load('atom', 'new_float')
				x = new_float(1)
				x.sub('2')
			`),
			wantErr: "sub: for parameter delta: got string, want float",
		},
		{
			name: `float: cas invalid args`,
			script: itn.HereDoc(`
				load('atom', 'new_float')
				x = new_float(1)
				x.cas('1', 2)
			`),
			wantErr: "cas: for parameter old: got string, want float",
		},
		{
			name: `float: compare`,
			script: itn.HereDoc(`
				load('atom', 'new_float')
				x = new_float(1.1)
				y = new_float(1.1)
				z = new_float(2.2)
				assert.eq(x, y)
				assert.ne(x, z)
				assert.true(x == y)
				assert.true(x != z)
				assert.true(x < z)
				assert.true(x <= z)
				assert.true(z > x)
				assert.true(z >= x)
				assert.true(x >= y)
				assert.true(x <= y)
			`),
		},
		{
			name: `float: full`,
			script: itn.HereDoc(`
				load('atom', 'new_float')
				x = new_float(1)	
				assert.eq(x.get(), 1)	
				x.set(20.1)	
				assert.eq(x.get(), 20.1)
				assert.eq(x.add(5), 25.1)
				assert.eq(x.add(2.1), 27.2)
				assert.eq(x.sub(2.1), 25.1)
				assert.eq(x.sub(3), 22.1)
				assert.eq(x.cas(22.1, 100), True)
				assert.eq(x.get(), 100)
				assert.eq(x.cas(22.1, 200.5), False)
				assert.eq(x.get(), 100)
			`),
		},
		{
			name: `float: usage`,
			script: itn.HereDoc(`
				load('atom', 'new_float')
				x = new_float()
				def work():
					x.add(1)
				[work() for _ in range(10)]
				assert.eq(x.get(), 10.0)
			`),
		},

		// for string
		{
			name: `string: missing attrs`,
			script: itn.HereDoc(`
				load('atom', 'new_string')
				x = new_string("aloha")
				x.guess()
			`),
			wantErr: "atom_string has no .guess field or method",
		},
		{
			name: `string: default`,
			script: itn.HereDoc(`
				load('atom', 'new_string')
				x = new_string()
				assert.eq(x.get(), "")
				assert.eq(type(x), 'atom_string')
				assert.eq(str(x), '<atom_string:"">')
				assert.eq(dir(x), ["cas", "get", "set"])
				assert.true(not bool(x))
			`),
		},
		{
			name: `string: new with args`,
			script: itn.HereDoc(`
				load('atom', 'new_string')
				x = new_string("hello")
				assert.eq(str(x), '<atom_string:"hello">')
				assert.eq(x.get(), "hello")
				assert.true(bool(x))
			`),
		},
		{
			name: `string: new invalid args`,
			script: itn.HereDoc(`
				load('atom', 'new_string')
				x = new_string(1)
			`),
			wantErr: "new_string: for parameter value: got int, want string",
		},
		{
			name: `string: set invalid args`,
			script: itn.HereDoc(`
				load('atom', 'new_string')
				x = new_string("hello")
				x.set(1)
			`),
			wantErr: "set: for parameter value: got int, want string",
		},
		{
			name: `string: get with args`,
			script: itn.HereDoc(`
				load('atom', 'new_string')
				x = new_string("hello")
				x.get(2)
			`),
			wantErr: "get: got 1 arguments, want 0",
		},
		{
			name: `string: cas invalid args`,
			script: itn.HereDoc(`
				load('atom', 'new_string')
				x = new_string("hello")
				x.cas(1, "world")
			`),
			wantErr: "cas: for parameter old: got int, want string",
		},
		{
			name: `string: compare`,
			script: itn.HereDoc(`
				load('atom', 'new_string')
				x = new_string("hello")
				y = new_string("hello")
				z = new_string("world")
				assert.eq(x, y)
				assert.ne(x, z)
				assert.true(x == y)
				assert.true(x != z)
				assert.true(x < z)
				assert.true(x <= z)
				assert.true(z > x)
				assert.true(z >= x)
				assert.true(x >= y)
				assert.true(x <= y)
			`),
		},
		{
			name: `string: full`,
			script: itn.HereDoc(`
				load('atom', 'new_string')
				x = new_string("hello")	
				assert.eq(x.get(), "hello")	
				x.set("world")	
				assert.eq(x.get(), "world")
				assert.eq(x.cas("world", "new"), True)
				assert.eq(x.get(), "new")
				assert.eq(x.cas("world", "new2"), False)
				assert.eq(x.get(), "new")
			`),
		},
		{
			name: `string: usage`,
			script: itn.HereDoc(`
				load('atom', 'new_string')
				x = new_string()
				def work():
					s = x.get()
					x.set(s + "!")
				[work() for _ in range(10)]
				assert.eq(x.get(), "!!!!!!!!!!")
			`),
		},
		// unhashable: a mutable cell must not be usable as a dict/set key —
		// its live-value hash goes stale on mutation and the entry is lost
		{
			name: `int: unhashable as dict key`,
			script: itn.HereDoc(`
				load('atom', 'new_int')
				x = new_int(1)
				d = {x: 'a'}
			`),
			wantErr: "unhashable type: atom_int",
		},
		{
			name: `float: unhashable as dict key`,
			script: itn.HereDoc(`
				load('atom', 'new_float')
				x = new_float(1.5)
				d = {x: 'a'}
			`),
			wantErr: "unhashable type: atom_float",
		},
		{
			name: `string: unhashable as dict key`,
			script: itn.HereDoc(`
				load('atom', 'new_string')
				x = new_string("k")
				d = {x: 'a'}
			`),
			wantErr: "unhashable type: atom_string",
		},
		{
			name: `int: unhashable as set element`,
			script: itn.HereDoc(`
				load('atom', 'new_int')
				s = set([new_int(1)])
			`),
			wantErr: "unhashable type: atom_int",
		},
		// frozen: every mutating method must reject a frozen cell
		{
			name: `int: frozen set fails`,
			script: itn.HereDoc(`
				load('atom', 'new_int')
				load('freeze.star', 'freeze')
				x = new_int(1)
				freeze(x)
				x.set(2)
			`),
			wantErr: "cannot set frozen atom_int",
		},
		{
			name: `int: frozen cas fails`,
			script: itn.HereDoc(`
				load('atom', 'new_int')
				load('freeze.star', 'freeze')
				x = new_int(1)
				freeze(x)
				x.cas(1, 2)
			`),
			wantErr: "cannot cas frozen atom_int",
		},
		{
			name: `int: frozen add fails`,
			script: itn.HereDoc(`
				load('atom', 'new_int')
				load('freeze.star', 'freeze')
				x = new_int(1)
				freeze(x)
				x.add(1)
			`),
			wantErr: "cannot add frozen atom_int",
		},
		{
			name: `int: frozen sub fails`,
			script: itn.HereDoc(`
				load('atom', 'new_int')
				load('freeze.star', 'freeze')
				x = new_int(1)
				freeze(x)
				x.sub(1)
			`),
			wantErr: "cannot sub frozen atom_int",
		},
		{
			name: `int: frozen inc fails`,
			script: itn.HereDoc(`
				load('atom', 'new_int')
				load('freeze.star', 'freeze')
				x = new_int(1)
				freeze(x)
				x.inc()
			`),
			wantErr: "cannot inc frozen atom_int",
		},
		{
			name: `int: frozen dec fails`,
			script: itn.HereDoc(`
				load('atom', 'new_int')
				load('freeze.star', 'freeze')
				x = new_int(1)
				freeze(x)
				x.dec()
			`),
			wantErr: "cannot dec frozen atom_int",
		},
		{
			name: `float: frozen set fails`,
			script: itn.HereDoc(`
				load('atom', 'new_float')
				load('freeze.star', 'freeze')
				x = new_float(1.5)
				freeze(x)
				x.set(2.5)
			`),
			wantErr: "cannot set frozen atom_float",
		},
		{
			name: `float: frozen cas fails`,
			script: itn.HereDoc(`
				load('atom', 'new_float')
				load('freeze.star', 'freeze')
				x = new_float(1.5)
				freeze(x)
				x.cas(1.5, 2.5)
			`),
			wantErr: "cannot cas frozen atom_float",
		},
		{
			name: `float: frozen add fails`,
			script: itn.HereDoc(`
				load('atom', 'new_float')
				load('freeze.star', 'freeze')
				x = new_float(1.5)
				freeze(x)
				x.add(1.0)
			`),
			wantErr: "cannot add frozen atom_float",
		},
		{
			name: `float: frozen sub fails`,
			script: itn.HereDoc(`
				load('atom', 'new_float')
				load('freeze.star', 'freeze')
				x = new_float(1.5)
				freeze(x)
				x.sub(1.0)
			`),
			wantErr: "cannot sub frozen atom_float",
		},
		{
			name: `string: frozen set fails`,
			script: itn.HereDoc(`
				load('atom', 'new_string')
				load('freeze.star', 'freeze')
				x = new_string("a")
				freeze(x)
				x.set("b")
			`),
			wantErr: "cannot set frozen atom_string",
		},
		{
			name: `string: frozen cas fails`,
			script: itn.HereDoc(`
				load('atom', 'new_string')
				load('freeze.star', 'freeze')
				x = new_string("a")
				freeze(x)
				x.cas("a", "b")
			`),
			wantErr: "cannot cas frozen atom_string",
		},
		// frozen cells stay readable and comparable
		{
			name: `frozen cells still read and compare`,
			script: itn.HereDoc(`
				load('atom', 'new_int', 'new_float', 'new_string')
				load('freeze.star', 'freeze')
				i = new_int(7)
				f = new_float(2.5)
				s = new_string("hi")
				freeze(i)
				freeze(f)
				freeze(s)
				assert.eq(i.get(), 7)
				assert.eq(f.get(), 2.5)
				assert.eq(s.get(), "hi")
				assert.true(i == new_int(7))
				assert.true(f < new_float(3.0))
				assert.true(s != new_string("yo"))
				assert.true(bool(i))
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
