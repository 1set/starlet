package serial_test

import (
	"testing"
	"time"

	"github.com/1set/starlet/dataconv"
	itn "github.com/1set/starlet/internal"
	"github.com/1set/starlet/lib/serial"
	startime "go.starlark.net/lib/time"
	"go.starlark.net/starlark"
)

func TestLoadModule_Serial(t *testing.T) {
	type Person struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}
	t0 := time.Date(2026, 6, 13, 12, 30, 0, 0, time.UTC)
	pred := starlark.StringDict{
		"t0": startime.Time(t0),
		"gs": dataconv.ConvertStruct(&Person{"Ann", 3}, "json"),
	}
	tests := []struct {
		name    string
		script  string
		wantErr string
	}{
		{
			name: `round-trip: scalars`,
			script: itn.HereDoc(`
				load('serial', 'dumps', 'loads')
				def rt(x): return loads(dumps(x))
				assert.eq(rt(None), None)
				assert.eq(rt(True), True)
				assert.eq(rt(False), False)
				assert.eq(rt(0), 0)
				assert.eq(rt(42), 42)
				assert.eq(rt(-7), -7)
				assert.eq(rt(1267650600228229401496703205376), 1267650600228229401496703205376)
				assert.eq(rt(3.14), 3.14)
				assert.eq(rt('héllo'), 'héllo')
				assert.eq(rt(''), '')
			`),
		},
		{
			name: `round-trip: containers and the five extra types`,
			script: itn.HereDoc(`
				load('serial', 'dumps', 'loads')
				def rt(x): return loads(dumps(x))
				assert.eq(rt(b'abc'), b'abc')
				assert.eq(rt([1, 'a', True, None]), [1, 'a', True, None])
				assert.eq(rt((1, 2)), (1, 2))
				assert.eq(rt({'a': 1, 'b': [2, 3]}), {'a': 1, 'b': [2, 3]})
				assert.eq(rt(set([1, 2, 3])), set([1, 2, 3]))
				assert.eq(rt({1: 'a', (2, 3): 'b'}), {1: 'a', (2, 3): 'b'})
				assert.eq(rt(t0), t0)
			`),
		},
		{
			name: `type preservation: tuple stays tuple, bytes stays bytes, set stays set`,
			script: itn.HereDoc(`
				load('serial', 'dumps', 'loads')
				def rt(x): return loads(dumps(x))
				assert.eq(type(rt((1, 2))), 'tuple')
				assert.eq(type(rt([1, 2])), 'list')
				assert.eq(type(rt(b'x')), 'bytes')
				assert.eq(type(rt(set([1]))), 'set')
				assert.eq(type(rt(1267650600228229401496703205376)), 'int')
			`),
		},
		{
			name: `determinism: key order independent, stable golden`,
			script: itn.HereDoc(`
				load('serial', 'dumps')
				assert.eq(dumps({'b': 2, 'a': 1}), dumps({'a': 1, 'b': 2}))
				assert.eq(dumps({'a': 1, 'b': 2}), '{"a":1,"b":2}')
				assert.eq(dumps(set([3, 1, 2])), dumps(set([2, 3, 1])))
			`),
		},
		{
			name: `$t escape: a real dict carrying a $t key round-trips`,
			script: itn.HereDoc(`
				load('serial', 'dumps', 'loads')
				def rt(x): return loads(dumps(x))
				assert.eq(rt({'$t': 'hello', 'y': 1}), {'$t': 'hello', 'y': 1})
				assert.true('object' in dumps({'$t': 'x'}))
			`),
		},
		{
			name: `round-trip: deep nesting mixing every type`,
			script: itn.HereDoc(`
				load('serial', 'dumps', 'loads')
				def rt(x): return loads(dumps(x))
				v = {'list': [1, (2, 3), set([4])], 'bytes': b'z', 'big': 1180591620717411303424, 'm': {1: 'x'}}
				assert.eq(rt(v), v)
			`),
		},
		{
			name: `try_dumps / try_loads: ok and error`,
			script: itn.HereDoc(`
				load('serial', 'try_dumps', 'try_loads')
				out, err = try_dumps(42)
				assert.eq(err, None)
				assert.eq(out, '42')
				bad, e2 = try_dumps(lambda x: x)
				assert.eq(bad, None)
				assert.true(e2 != None)
				val, e3 = try_loads('[1,2,3]')
				assert.eq(e3, None)
				assert.eq(val, [1, 2, 3])
				bad2, e4 = try_loads('not json')
				assert.eq(bad2, None)
				assert.true(e4 != None)
			`),
		},
		{
			name: `error: cannot serialize a function`,
			script: itn.HereDoc(`
				load('serial', 'dumps')
				dumps(lambda x: x)
			`),
			wantErr: `it is code`,
		},
		{
			name: `error: cannot serialize a builtin`,
			script: itn.HereDoc(`
				load('serial', 'dumps')
				dumps(dumps)
			`),
			wantErr: `it is code`,
		},
		{
			name: `error: reference cycle is reported, not a panic`,
			script: itn.HereDoc(`
				load('serial', 'dumps')
				x = [1]
				x.append(x)
				dumps(x)
			`),
			wantErr: `cycle`,
		},
		{
			name: `error: non-finite float`,
			script: itn.HereDoc(`
				load('serial', 'dumps')
				dumps(float('inf'))
			`),
			wantErr: `non-finite`,
		},
		{
			name: `error: struct must be converted to a dict first`,
			script: itn.HereDoc(`
				load('serial', 'dumps')
				load('struct.star', 'struct')
				dumps(struct(a=1))
			`),
			wantErr: `convert it to a dict`,
		},
		{
			name: `error: host Go object is not data`,
			script: itn.HereDoc(`
				load('serial', 'dumps')
				dumps(gs)
			`),
			wantErr: `host objects`,
		},
		{
			name: `error: loads rejects an unknown type tag`,
			script: itn.HereDoc(`
				load('serial', 'loads')
				loads('{"$t":"bogus","v":1}')
			`),
			wantErr: `unknown type tag`,
		},
		{
			name: `error: loads rejects invalid JSON`,
			script: itn.HereDoc(`
				load('serial', 'loads')
				loads('{not json')
			`),
			wantErr: `serial.loads:`,
		},
		{
			name: `error: dumps missing argument`,
			script: itn.HereDoc(`
				load('serial', 'dumps')
				dumps()
			`),
			wantErr: `serial.dumps: missing argument for value`,
		},
		{
			name: `error: loads wrong type`,
			script: itn.HereDoc(`
				load('serial', 'loads')
				loads(42)
			`),
			wantErr: `serial.loads: for parameter s: got int, want string`,
		},
		{
			name: `loads: bare big integer and floats (numberToValue branches)`,
			script: itn.HereDoc(`
				load('serial', 'loads')
				assert.eq(loads('123456789012345678901234567890'), 123456789012345678901234567890)
				assert.eq(loads('3.5'), 3.5)
				assert.eq(loads('-0.25'), -0.25)
				assert.eq(loads('1e3'), 1000.0)
				`),
		},
		{
			name: `error: unserializable inside a list propagates`,
			script: itn.HereDoc(`
				load('serial', 'dumps')
				dumps([1, lambda x: x])
				`),
			wantErr: `it is code`,
		},
		{
			name: `error: unserializable inside a tuple propagates`,
			script: itn.HereDoc(`
				load('serial', 'dumps')
				dumps((1, lambda x: x))
				`),
			wantErr: `it is code`,
		},
		{
			name: `error: unserializable as a dict value propagates`,
			script: itn.HereDoc(`
				load('serial', 'dumps')
				dumps({'a': lambda x: x})
				`),
			wantErr: `it is code`,
		},
		{
			name: `error: unserializable in a non-string-key dict propagates`,
			script: itn.HereDoc(`
				load('serial', 'dumps')
				dumps({1: lambda x: x})
				`),
			wantErr: `it is code`,
		},
		{
			name: `error: unserializable inside a set propagates`,
			script: itn.HereDoc(`
				load('serial', 'dumps')
				dumps(set([lambda x: x]))
				`),
			wantErr: `it is code`,
		},
		{
			name: `error: loads invalid bytes payload`,
			script: itn.HereDoc(`
				load('serial', 'loads')
				loads('{"$t":"bytes","v":"!!!!"}')
				`),
			wantErr: `bytes payload`,
		},
		{
			name: `error: loads invalid bigint payload`,
			script: itn.HereDoc(`
				load('serial', 'loads')
				loads('{"$t":"bigint","v":"x"}')
				`),
			wantErr: `invalid bigint`,
		},
		{
			name: `error: loads invalid time payload`,
			script: itn.HereDoc(`
				load('serial', 'loads')
				loads('{"$t":"time","v":"x"}')
				`),
			wantErr: `invalid time`,
		},
		{
			name: `error: loads malformed mapkv entry`,
			script: itn.HereDoc(`
				load('serial', 'loads')
				loads('{"$t":"mapkv","v":[[1]]}')
				`),
			wantErr: `invalid mapkv`,
		},
		{
			name: `error: loads invalid object payload`,
			script: itn.HereDoc(`
				load('serial', 'loads')
				loads('{"$t":"object","v":5}')
				`),
			wantErr: `invalid object payload`,
		},
		{
			name: `error: nested unknown tag propagates through a dict`,
			script: itn.HereDoc(`
				load('serial', 'loads')
				loads('{"k":{"$t":"nope"}}')
				`),
			wantErr: `unknown type tag`,
		},
		{
			name: `error: unserializable as a non-string dict key propagates`,
			script: itn.HereDoc(`
				load('serial', 'dumps')
				dumps({(1, lambda x: x): 1})
				`),
			wantErr: `it is code`,
		},
		{
			name: `error: loads set with an unhashable element`,
			script: itn.HereDoc(`
				load('serial', 'loads')
				loads('{"$t":"set","v":[[1,2]]}')
				`),
			wantErr: `unhashable`,
		},
		{
			name: `error: loads mapkv with an unhashable key`,
			script: itn.HereDoc(`
				load('serial', 'loads')
				loads('{"$t":"mapkv","v":[[[1],"x"]]}')
				`),
			wantErr: `unhashable`,
		},
		{
			name: `try_loads: decode error after valid JSON parses`,
			script: itn.HereDoc(`
				load('serial', 'try_loads')
				v, err = try_loads('{"$t":"bogus","v":1}')
				assert.eq(v, None)
				assert.true('unknown type tag' in err)
				`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := itn.ExecModuleWithErrorTest(t, serial.ModuleName, serial.LoadModule, tt.script, tt.wantErr, pred)
			if (err != nil) != (tt.wantErr != "") {
				t.Errorf("serial(%q) expects error = %q, got %v", tt.name, tt.wantErr, err)
			}
		})
	}
}
