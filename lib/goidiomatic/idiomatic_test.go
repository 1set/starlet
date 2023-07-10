package goidiomatic_test

import (
	"testing"

	"github.com/1set/starlet/lib/goidiomatic"
	itn "github.com/1set/starlet/lib/internal"
	"github.com/1set/starlight/convert"
	"go.starlark.net/starlark"
)

type testStruct struct {
	Slice  []string
	Map    map[string]string
	Struct *struct {
		A string
		B string
	}
	NestedStruct *struct {
		Child *struct {
			C string
			D string
		}
	}
	Pointer interface{}
}

func TestLoadModule_GoIdiomatic(t *testing.T) {
	starlark.Universe["test_custom_struct"] = convert.NewStruct(testStruct{})
	starlark.Universe["test_custom_struct_pointer"] = convert.NewStruct(&testStruct{})

	// test cases
	tests := []struct {
		name    string
		script  string
		wantErr string
	}{
		{
			name: `boolean`,
			script: itn.HereDoc(`
				load('go_idiomatic', 'true', 'false')
				assert.eq(true, True)
				assert.eq(false, False)
			`),
		},
		{
			name: `nil`,
			script: itn.HereDoc(`
				load('go_idiomatic', 'nil')
				assert.eq(nil, None)
			`),
		},
		{
			name: `is_nil()`,
			script: itn.HereDoc(`
				load('go_idiomatic', 'is_nil')
				is_nil()
			`),
			wantErr: `is_nil: missing argument for x`,
		},
		{
			name: `is_nil(123)`,
			script: itn.HereDoc(`
				load('go_idiomatic', 'is_nil')
				is_nil(123)
			`),
			wantErr: `is_nil: unsupported type: starlark.Int`,
		},
		{
			name: `is_nil struct`,
			script: itn.HereDoc(`
				load('go_idiomatic', 'is_nil')
				cs = test_custom_struct
				assert.eq(is_nil(None), True)
				assert.eq(is_nil(cs), False)
				assert.eq(is_nil(cs.Slice), True)
				assert.eq(is_nil(cs.Map), True)
				assert.eq(is_nil(cs.Struct), True)
				assert.eq(is_nil(cs.NestedStruct), True)
				assert.eq(is_nil(cs.Pointer), True)
			`),
		},
		{
			name: `is_nil pointer`,
			script: itn.HereDoc(`
				load('go_idiomatic', 'is_nil')
				cs = test_custom_struct_pointer
				assert.eq(is_nil(cs), False)
				assert.eq(is_nil(cs.Slice), True)
				assert.eq(is_nil(cs.Map), True)
				assert.eq(is_nil(cs.Struct), True)
				assert.eq(is_nil(cs.NestedStruct), True)
				assert.eq(is_nil(cs.Pointer), True)
			`),
		},
		{
			name: `sleep`,
			script: itn.HereDoc(`
				load('go_idiomatic', 'sleep')
				sleep()
			`),
			wantErr: `sleep: missing argument for secs`,
		},
		{
			name: `sleep 0`,
			script: itn.HereDoc(`
				load('go_idiomatic', 'sleep')
				sleep(0)
			`),
		},
		{
			name: `sleep 10ms`,
			script: itn.HereDoc(`
				load('go_idiomatic', 'sleep')
				sleep(0.01)
			`),
		},
		{
			name: `sleep negative`,
			script: itn.HereDoc(`
				load('go_idiomatic', 'sleep')
				sleep(-1)
			`),
			wantErr: `secs must be non-negative`,
		},
		{
			name: `sleep hello`,
			script: itn.HereDoc(`
				load('go_idiomatic', 'sleep')
				sleep('hello')
			`),
			wantErr: `sleep: for parameter secs: got string, want float or int`,
		},
		{
			name: `sleep none`,
			script: itn.HereDoc(`
				load('go_idiomatic', 'sleep')
				sleep(None)
			`),
			wantErr: `sleep: for parameter secs: got NoneType, want float or int`,
		},
		{
			name: `exit`,
			script: itn.HereDoc(`
				load('go_idiomatic', 'exit')
				exit()
			`),
			wantErr: `starlet runtime system exit (Use Ctrl-D in REPL to exit)`,
		},
		{
			name: `exit 0`,
			script: itn.HereDoc(`
				load('go_idiomatic', 'exit')
				exit(0)
			`),
			wantErr: `starlet runtime system exit (Use Ctrl-D in REPL to exit)`,
		},
		{
			name: `exit 1`,
			script: itn.HereDoc(`
				load('go_idiomatic', 'exit')
				exit(1)
			`),
			wantErr: `starlet runtime system exit (Use Ctrl-D in REPL to exit)`,
		},
		{
			name: `exit -1`,
			script: itn.HereDoc(`
				load('go_idiomatic', 'exit')
				exit(-1)
			`),
			wantErr: `exit: for parameter code: -1 out of range (want value in unsigned 8-bit range)`,
		},
		{
			name: `exit hello`,
			script: itn.HereDoc(`
				load('go_idiomatic', 'exit')
				exit('hello')
			`),
			wantErr: `exit: for parameter code: got string, want int`,
		},
		{
			name: `quit`,
			script: itn.HereDoc(`
				load('go_idiomatic', 'quit')
				quit()
			`),
			wantErr: `starlet runtime system exit (Use Ctrl-D in REPL to exit)`,
		},
		{
			name: `length(string)`,
			script: itn.HereDoc(`
				load('go_idiomatic', 'length')
				assert.eq(0, length(''))
				assert.eq(1, length('a'))
				assert.eq(2, length('ab'))
				assert.eq(3, length('abc'))
				assert.eq(1, length('🙆'))
				assert.eq(1, length('✅'))
				assert.eq(3, length('水光肌'))
			`),
		},
		{
			name: `length(bytes)`,
			script: itn.HereDoc(`
				load('go_idiomatic', 'length')
				assert.eq(0, length(b''))
				assert.eq(9, length(b'水光肌'))
				assert.eq(7, length(b'🙆✅'))
			`),
		},
		{
			name: `length(list/tuple/dict/set)`,
			script: itn.HereDoc(`
				load('go_idiomatic', 'length')
				assert.eq(0, length([]))
				assert.eq(5, length([1, 2, "#", True, None]))
				assert.eq(1, length((3,)))
				assert.eq(2, length((1, 2)))
				assert.eq(0, length({}))
				assert.eq(2, length({'a': 1, 'b': 2}))
				assert.eq(0, length(set()))
				assert.eq(2, length(set(['a', 'b'])))
			`),
		},
		{
			name: `length(slice)`,
			script: itn.HereDoc(`
				load('go_idiomatic', 'length')
				assert.eq(3, length(slice))
			`),
		},
		{
			name: `length(map)`,
			script: itn.HereDoc(`
				load('go_idiomatic', 'length')
				assert.eq(2, length(map))
			`),
		},
		{
			name: `length()`,
			script: itn.HereDoc(`
				load('go_idiomatic', 'length')
				length()
			`),
			wantErr: `length() takes exactly one argument (0 given)`,
		},
		{
			name: `length(*2)`,
			script: itn.HereDoc(`
				load('go_idiomatic', 'length')
				length('a', 'b')
			`),
			wantErr: `length() takes exactly one argument (2 given)`,
		},
		{
			name: `length(bool)`,
			script: itn.HereDoc(`
				load('go_idiomatic', 'length')
				length(True)
			`),
			wantErr: `object of type 'bool' has no length()`,
		},
		{
			name: `sum()`,
			script: itn.HereDoc(`
				load('go_idiomatic', 'sum')
				sum()
			`),
			wantErr: `sum: missing argument for iterable`,
		},
		{
			name: `sum(>2)`,
			script: itn.HereDoc(`
				load('go_idiomatic', 'sum')
				sum(1, 2, 3)
			`),
			wantErr: `sum: got 3 arguments, want at most 2`,
		},
		{
			name: `sum(int)`,
			script: itn.HereDoc(`
				load('go_idiomatic', 'sum')
				sum(1)
			`),
			wantErr: `sum: for parameter iterable: got int, want iterable`,
		},
		{
			name: `sum(string)`,
			script: itn.HereDoc(`
				load('go_idiomatic', 'sum')
				sum('abc')
			`),
			wantErr: `sum: for parameter iterable: got string, want iterable`,
		},
		{
			name: `sum([int])`,
			script: itn.HereDoc(`
				load('go_idiomatic', 'sum')
				assert.eq(6, sum([1, 2, 3]))
			`),
		},
		{
			name: `sum([float])`,
			script: itn.HereDoc(`
				load('go_idiomatic', 'sum')
				assert.eq(6, sum([1.0, 2.0, 3.0]))
			`),
		},
		{
			name: `sum([string])`,
			script: itn.HereDoc(`
				load('go_idiomatic', 'sum')
				sum(['a', 'b', 'c'])
			`),
			wantErr: `unsupported type: string, expected float or int`,
		},
		{
			name: `sum([int,float])`,
			script: itn.HereDoc(`
				load('go_idiomatic', 'sum')
				assert.eq(7, sum([1, 2.0, 4]))
			`),
		},
		{
			name: `sum([int,string])`,
			script: itn.HereDoc(`
				load('go_idiomatic', 'sum')
				sum([1, 'a'])
			`),
			wantErr: `unsupported type: string, expected float or int`,
		},
		{
			name: `sum([int], float)`,
			script: itn.HereDoc(`
				load('go_idiomatic', 'sum')
				assert.eq(15, sum([1, 2, 4], 8.0))
			`),
		},
		{
			name: `sum([int], start=int)`,
			script: itn.HereDoc(`
				load('go_idiomatic', 'sum')
				assert.eq(15, sum([1, 2, 4], start=8))
			`),
		},
		{
			name: `hex()`,
			script: itn.HereDoc(`
				load('go_idiomatic', 'hex')
				hex()
			`),
			wantErr: `hex: missing argument for x`,
		},
		{
			name: `hex(>1)`,
			script: itn.HereDoc(`
				load('go_idiomatic', 'hex')
				hex(1, 2)
			`),
			wantErr: `hex: got 2 arguments, want at most 1`,
		},
		{
			name: `hex(1)`,
			script: itn.HereDoc(`
				load('go_idiomatic', 'hex')
				assert.eq('0x0', hex(0))
				assert.eq('0xf', hex(15))
				assert.eq('-0xf', hex(-15))
				assert.eq('-0x100', hex(-256))
				assert.eq('0x1a459b09a8bbc286c14756a86376e710', hex(34921340912409213842304823423424456464))
			`),
		},
		{
			name: `oct()`,
			script: itn.HereDoc(`
				load('go_idiomatic', 'oct')
				oct()
			`),
			wantErr: `oct: missing argument for x`,
		},
		{
			name: `oct(>1)`,
			script: itn.HereDoc(`
				load('go_idiomatic', 'oct')
				oct(1, 2)
			`),
			wantErr: `oct: got 2 arguments, want at most 1`,
		},
		{
			name: `oct(1)`,
			script: itn.HereDoc(`
				load('go_idiomatic', 'oct')
				assert.eq('0o0', oct(0))
				assert.eq('0o7', oct(7))
				assert.eq('0o10', oct(8))
				assert.eq('-0o70', oct(-56))
				assert.eq('0o25002121274216244622344125302630165706706267707144', oct(468409683456048976340589328520898324578930276))
			`),
		},
		{
			name: `bin()`,
			script: itn.HereDoc(`
				load('go_idiomatic', 'bin')
				bin()
			`),
			wantErr: `bin: missing argument for x`,
		},
		{
			name: `bin(>1)`,
			script: itn.HereDoc(`
				load('go_idiomatic', 'bin')
				bin(1, 2)
			`),
			wantErr: `bin: got 2 arguments, want at most 1`,
		},
		{
			name: `bin(1)`,
			script: itn.HereDoc(`
				load('go_idiomatic', 'bin')
				assert.eq('0b0', bin(0))
				assert.eq('0b111', bin(7))
				assert.eq('0b1000', bin(8))
				assert.eq('-0b11111111', bin(-255))
				assert.eq('0b11010010001011001101100001001101010001011101111000010100001101100000101000111010101101010100001100011011101101110011100010000', bin(34921340912409213842304823423424456464))
			`),
		},
		{
			name: `bytes_hex()`,
			script: itn.HereDoc(`
				load('go_idiomatic', 'bytes_hex')
				bytes_hex()
			`),
			wantErr: `bytes_hex: missing argument for bytes`,
		},
		{
			name: `bytes_hex(>3)`,
			script: itn.HereDoc(`
				load('go_idiomatic', 'bytes_hex')
				bytes_hex(b'123456', "_", 4, 5)
			`),
			wantErr: `bytes_hex: got 4 arguments, want at most 3`,
		},
		{
			name: `bytes_hex(b'123456')`,
			script: itn.HereDoc(`
				load('go_idiomatic', 'bytes_hex')
				assert.eq('313233343536', bytes_hex(b'123456'))
			`),
		},
		{
			name: `bytes_hex(b'123456', "_")`,
			script: itn.HereDoc(`
				load('go_idiomatic', 'bytes_hex')
				assert.eq('31_32_33_34_35_36', bytes_hex(b'123456', "_"))
			`),
		},
		{
			name: `bytes_hex(b'0', "-")`,
			script: itn.HereDoc(`
				load('go_idiomatic', 'bytes_hex')
				assert.eq('', bytes_hex(b'', "-"))
				assert.eq('30', bytes_hex(b'0', "-"))
			`),
		},
		{
			name: `bytes_hex(b'123456', "_", 4)`,
			script: itn.HereDoc(`
				load('go_idiomatic', 'bytes_hex')
				assert.eq('3132_33343536', bytes_hex(b'123456', "_", 4))
				assert.eq('31_3233343536', bytes_hex(b'123456', "_", 5))
				assert.eq('5555 44444c52 4c524142', bytes_hex(b'UUDDLRLRAB', " ", 4))
			`),
		},
		{
			name: `bytes_hex(b'123456', "_", -4)`,
			script: itn.HereDoc(`
				load('go_idiomatic', 'bytes_hex')
				assert.eq('31323334_3536', bytes_hex(b'123456', "_", -4))
				assert.eq('3132333435_36', bytes_hex(b'123456', "_", -5))
				assert.eq('55554444 4c524c52 4142', bytes_hex(b'UUDDLRLRAB', " ", -4))
			`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// prepare
			s, err := convert.ToValue([]string{"a", "b", "c"})
			if err != nil {
				t.Errorf("convert.ToValue Slice: %v", err)
				return
			}
			m, err := convert.ToValue(map[string]string{"a": "b", "c": "d"})
			if err != nil {
				t.Errorf("convert.ToValue Map: %v", err)
				return
			}
			starlark.Universe["slice"] = s
			starlark.Universe["map"] = m

			res, err := itn.ExecModuleWithErrorTest(t, goidiomatic.ModuleName, goidiomatic.LoadModule, tt.script, tt.wantErr)
			if (err != nil) != (tt.wantErr != "") {
				t.Errorf("go_idiomatic(%q) expects error = '%v', actual error = '%v', result = %v", tt.name, tt.wantErr, err, res)
				return
			}
		})
	}
}
