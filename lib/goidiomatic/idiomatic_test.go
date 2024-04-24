package goidiomatic_test

import (
	"fmt"
	"testing"

	itn "github.com/1set/starlet/internal"
	"github.com/1set/starlet/lib/goidiomatic"
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

// customIntRange represents a range of integers [start, end).
type customIntRange struct {
	starlark.Value
	start, end int
}

// String returns the string representation of the customIntRange.
func (r *customIntRange) String() string {
	return fmt.Sprintf("customIntRange(%d, %d)", r.start, r.end)
}

// Type returns the type name of the customIntRange.
func (r *customIntRange) Type() string {
	return "customIntRange"
}

// Freeze makes the customIntRange immutable. Required by starlark.Value interface.
func (r *customIntRange) Freeze() {}

// Truth returns the truth value of the customIntRange.
func (r *customIntRange) Truth() starlark.Bool {
	return r.start < r.end // true if the range is not empty
}

// Hash returns the hash value of the customIntRange.
func (r *customIntRange) Hash() (uint32, error) {
	return uint32(r.start*31 + r.end), nil
}

// Iterate returns an iterator for the customIntRange.
func (r *customIntRange) Iterate() starlark.Iterator {
	return &customIntRangeIterator{ranger: r, next: r.start}
}

// customIntRangeIterator implements the Iterator interface for customIntRange.
type customIntRangeIterator struct {
	ranger *customIntRange
	next   int
}

// Next moves the iterator to the next value and returns true if there was a next value.
func (it *customIntRangeIterator) Next(p *starlark.Value) bool {
	if it.next >= it.ranger.end {
		return false
	}
	*p = starlark.MakeInt(it.next)
	it.next++
	return true
}

// Done does nothing but necessary to implement the Iterator interface.
func (it *customIntRangeIterator) Done() {}

// newCustomIntRange creates a new customIntRange value.
func newCustomIntRange(start, end int) *customIntRange {
	return &customIntRange{start: start, end: end}
}

func TestLoadModule_GoIdiomatic(t *testing.T) {
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
			wantErr: `length() function isn't supported for 'bool' type object`,
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
		{
			name: `module: missing name`,
			script: itn.HereDoc(`
				load('go_idiomatic', 'module')
				module()
			`),
			wantErr: `module: got 0 arguments, want 1`,
		},
		{
			name: `module: only name`,
			script: itn.HereDoc(`
				load('go_idiomatic', 'module')
				r = module("rose")
				assert.eq(str(r), '<module "rose">')
			`),
		},
		{
			name: `module: invalid value`,
			script: itn.HereDoc(`
				load('go_idiomatic', 'module')
				r = module("rose", 200)
			`),
			wantErr: `module: got 2 arguments, want 1`,
		},
		{
			name: `module: values`,
			script: itn.HereDoc(`
				load('go_idiomatic', 'module')
				r = module("rose", a=100)
				assert.eq(str(r), '<module "rose">')
				assert.eq(r.a, 100)
				s = module("rose", a=200, b="hello")
				assert.eq(str(s), '<module "rose">')
				assert.eq(s.a, 200)
				assert.eq(s.b, "hello")
			`),
		},
		{
			name: `module: compare`,
			script: itn.HereDoc(`
				load('go_idiomatic', 'module')
				a = module("rose", a=100)
				b = module("rose", a=100)
				c = module("rose", a=200)
				d = module("lily", a=100)
				assert.true(a != b)
				assert.true(a != c)
				assert.true(a != d)
			`),
		},
		{
			name: `struct: no name`,
			script: itn.HereDoc(`
				load('go_idiomatic', 'struct')
				s = struct()
				assert.eq(str(s), 'struct()')
			`),
		},
		{
			name: `struct: invalid value`,
			script: itn.HereDoc(`
				load('go_idiomatic', 'struct')
				r = struct("red")
			`),
			wantErr: `struct: unexpected positional arguments`,
		},
		{
			name: `struct: one value`,
			script: itn.HereDoc(`
				load('go_idiomatic', 'struct')
				r = struct(rose="red")
				assert.eq(str(r), 'struct(rose = "red")')
			`),
		},
		{
			name: `struct: values`,
			script: itn.HereDoc(`
				load('go_idiomatic', 'struct')
				r = struct(rose="red", lily="white")
				assert.eq(str(r), 'struct(lily = "white", rose = "red")')
			`),
		},
		{
			name: `struct: compare`,
			script: itn.HereDoc(`
				load('go_idiomatic', 'struct')
				a = struct(a=100)
				b = struct(a=100)
				c = struct(a=200)
				d = struct(a=100, b=200)
				assert.true(a == b)
				assert.true(a != c)
				assert.true(a != d)
				assert.true(c != d)
			`),
		},
		{
			name: `make_struct: no name`,
			script: itn.HereDoc(`
				load('go_idiomatic', 'make_struct')
				s = make_struct()
				assert.eq(str(s), 'make_struct()')
			`),
			wantErr: `make_struct: got 0 arguments, want 1`,
		},
		{
			name: `make_struct: only name`,
			script: itn.HereDoc(`
				load('go_idiomatic', 'make_struct')
				r = make_struct("rose")
				assert.eq(str(r), 'rose()')
			`),
		},
		{
			name: `make_struct: invalid value`,
			script: itn.HereDoc(`
				load('go_idiomatic', 'make_struct')
				r = make_struct("rose", 100)
			`),
			wantErr: `make_struct: got 2 arguments, want 1`,
		},
		{
			name: `make_struct: one value`,
			script: itn.HereDoc(`
				load('go_idiomatic', 'make_struct')
				r = make_struct("rose", price=100)
				assert.eq(str(r), 'rose(price = 100)')
			`),
		},
		{
			name: `make_struct: values`,
			script: itn.HereDoc(`
				load('go_idiomatic', 'make_struct')
				r = make_struct("rose", price=100, color="red")
				assert.eq(str(r), 'rose(color = "red", price = 100)')
			`),
		},
		{
			name: `make_struct: compare`,
			script: itn.HereDoc(`
				load('go_idiomatic', 'make_struct')
				a = make_struct("rose", n=100)
				b = make_struct("rose", n=100)
				c = make_struct("rose", n=200)
				d = make_struct("rose", n=100, m=200)
				e = make_struct("lily", n=100)
				assert.eq(a, b)
				assert.true(a == b)
				assert.true(a != c)
				assert.true(a != d)
				assert.true(a != e)
			`),
		},
		{
			name: `shared_dict: args`,
			script: itn.HereDoc(`
				load('go_idiomatic', 'shared_dict')
				shared_dict(123)
			`),
			wantErr: `shared_dict: got 1 arguments, want 0`,
		},
		{
			name: `shared_dict: no args`,
			script: itn.HereDoc(`
				load('go_idiomatic', 'shared_dict')
				d = shared_dict()
				assert.eq(type(d), 'shared_dict')
			`),
		},
		{
			name: `make_shared_dict: no args`,
			script: itn.HereDoc(`
				load('go_idiomatic', msd='make_shared_dict')
				d = msd()
				assert.eq(type(d), 'shared_dict')
			`),
		},
		{
			name: `make_shared_dict: invalid`,
			script: itn.HereDoc(`
				load('go_idiomatic', msd='make_shared_dict')
				d = msd({"abc": 100})
			`),
			wantErr: `make_shared_dict: for parameter name: got dict, want string`,
		},
		{
			name: `make_shared_dict: name`,
			script: itn.HereDoc(`
				load('go_idiomatic', msd='make_shared_dict')
				d = msd("manaʻo")
				assert.eq(type(d), 'manaʻo')
				assert.eq(str(d), 'manaʻo({})')
			`),
		},
		{
			name: `make_shared_dict: values`,
			script: itn.HereDoc(`
				load('go_idiomatic', msd='make_shared_dict')
				d = msd(data={"abc": 100})
				assert.eq(type(d), 'shared_dict')
				assert.eq(str(d), 'shared_dict({"abc": 100})')
			`),
		},
		{
			name: `make_shared_dict: name and values`,
			script: itn.HereDoc(`
				load('go_idiomatic', msd='make_shared_dict')
				d = msd("manaʻo", {"abc": 123})
				assert.eq(type(d), 'manaʻo')
				assert.eq(str(d), 'manaʻo({"abc": 123})')
			`),
		},
		{
			name: `distinct with list`,
			script: itn.HereDoc(`
        load('go_idiomatic', 'distinct')
        result = distinct([1, 2, 2, 3, 3, 3])
        assert.eq([1, 2, 3], result)
    `),
		},
		{
			name: `distinct with tuple`,
			script: itn.HereDoc(`
        load('go_idiomatic', 'distinct')
        result = distinct((1, 2, 2, 3, 3, 3))
        assert.eq((1, 2, 3), result)
    `),
		},
		{
			name: `distinct with dict`,
			script: itn.HereDoc(`
        load('go_idiomatic', 'distinct')
        result = distinct({'a': 1, 'b': 2, 'c': 3})
        # Note: Dict keys order is not guaranteed
        assert.eq(['a', 'b', 'c'], sorted(result))
    `),
		},
		{
			name: `distinct with set`,
			script: itn.HereDoc(`
        load('go_idiomatic', 'distinct')
        result = distinct(set([1, 2, 3, 3, 2, 1]))
        assert.eq(set([1, 2, 3]), result)
    `),
		},
		{
			name: `distinct with empty list`,
			script: itn.HereDoc(`
        load('go_idiomatic', 'distinct')
        result = distinct([])
        assert.eq([], result)
    `),
		},
		{
			name: "distinct with list of single element",
			script: itn.HereDoc(`
        load('go_idiomatic', 'distinct')
        assert.eq([1], distinct([1]))
    `),
		},
		{
			name: `distinct with list of non-hashable elements`,
			script: itn.HereDoc(`
		load('go_idiomatic', 'distinct')
		l = []
		distinct([12, l])
	`),
			wantErr: `distinct: unhashable type: list`,
		},
		{
			name: `distinct with incorrect type`,
			script: itn.HereDoc(`
        load('go_idiomatic', 'distinct')
        distinct(123)
    `),
			wantErr: `distinct: for parameter iterable: got int, want iterable`,
		},
		{
			name: `distinct with none`,
			script: itn.HereDoc(`
        load('go_idiomatic', 'distinct')
        distinct(None)
    `),
			wantErr: `distinct: for parameter iterable: got NoneType, want iterable`,
		},
		{
			name: "distinct with non-iterable arg",
			script: itn.HereDoc(`
        load('go_idiomatic', 'distinct')
        distinct(True)
    `),
			wantErr: `distinct: for parameter iterable: got bool, want iterable`,
		},
		{
			name: "distinct with no args",
			script: itn.HereDoc(`
        load('go_idiomatic', 'distinct')
        distinct()
    `),
			wantErr: `distinct: missing argument for iterable`,
		},
		{
			name: `distinct with multiple arguments`,
			script: itn.HereDoc(`
        load('go_idiomatic', 'distinct')
        distinct([1, 2, 3], [4, 5, 6])
    `),
			wantErr: `distinct: got 2 arguments, want at most 1`,
		},
		{
			name: `distinct with custom type`,
			script: itn.HereDoc(`
				load('go_idiomatic', 'distinct')
				r1 = make_range(1, 6)
				r2 = distinct(r1)
				assert.eq([1, 2, 3, 4, 5], r2)
			`),
		},
		{
			name: `distinct with invalid custom type`,
			script: itn.HereDoc(`
				load('go_idiomatic', 'distinct')
				distinct(test_custom_struct_pointer)
			`),
			wantErr: `distinct: for parameter iterable: got starlight_struct<*goidiomatic_test.testStruct>, want iterable`,
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
			globals := starlark.StringDict{
				"slice":                      s,
				"map":                        m,
				"make_range":                 convert.MakeStarFn("make_range", newCustomIntRange),
				"test_custom_struct":         convert.NewStruct(testStruct{}),
				"test_custom_struct_pointer": convert.NewStruct(&testStruct{}),
			}

			res, err := itn.ExecModuleWithErrorTest(t, goidiomatic.ModuleName, goidiomatic.LoadModule, tt.script, tt.wantErr, globals)
			if (err != nil) != (tt.wantErr != "") {
				t.Errorf("go_idiomatic(%q) expects error = '%v', actual error = '%v', result = %v", tt.name, tt.wantErr, err, res)
				return
			}
		})
	}
}
