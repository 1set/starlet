package json_test

import (
	"testing"

	"github.com/1set/starlet/dataconv"
	itn "github.com/1set/starlet/internal"
	"github.com/1set/starlet/lib/json"
	"go.starlark.net/starlark"
)

func TestLoadModule_JSON(t *testing.T) {
	rs := struct {
		Message string `json:"msg,omitempty"`
		Times   int    `json:"t,omitempty"`
	}{"Bravo", 200}
	pred := starlark.StringDict{
		"vj": dataconv.ConvertJSONStruct(&rs),
		"vs": dataconv.ConvertStruct(&rs, "star"),
	}

	tests := []struct {
		name    string
		script  string
		wantErr string
	}{
		{
			name: `func can be loaded`,
			script: itn.HereDoc(`
				load('json', 'dumps', 'encode', 'decode', 'indent', 'try_dumps', 'try_encode', 'try_decode', 'try_indent')
			`),
		},
		{
			name: `dumps()`,
			script: itn.HereDoc(`
				load('json', 'dumps')
				dumps()
			`),
			wantErr: `json.dumps: missing argument for obj`,
		},
		{
			name: `dumps(>2)`,
			script: itn.HereDoc(`
				load('json', 'dumps')
				dumps(1, 2, 3)
			`),
			wantErr: `json.dumps: got 3 arguments, want at most 2`,
		},
		{
			name: `dumps(1, "a")`,
			script: itn.HereDoc(`
				load('json', 'dumps')
				dumps(1, "a")
			`),
			wantErr: `json.dumps: for parameter indent: got string, want int`,
		},
		{
			name: `dumps(1)`,
			script: itn.HereDoc(`
				load('json', 'dumps')
				assert.eq(dumps(1), '1')
			`),
		},
		{
			name: `dumps(1, indent=2)`,
			script: itn.HereDoc(`
				load('json', 'dumps')
				assert.eq(dumps(1, indent=2), '1')
			`),
		},
		{
			name: `dumps(def)`,
			script: itn.HereDoc(`
				load('json', 'dumps')
				def f(): pass
				d = { "a" : "b", "f" : f}
				dumps(d)
			`),
			wantErr: `unmarshaling starlark value: unrecognized starlark type: *starlark.Function`,
		},
		{
			name: `dumps(dict)`,
			script: itn.HereDoc(`
				load('json', 'dumps')
				d = { "a" : "b", "c" : "d"}
				s = '''{"a":"b","c":"d"}'''
				assert.eq(dumps(d), s)
			`),
		},
		{
			name: `dumps(dict, indent=0)`,
			script: itn.HereDoc(`
				load('json', 'dumps')
				d = { "a" : "b", "c" : "d"}
				s = '''{"a":"b","c":"d"}'''
				assert.eq(dumps(d, indent=0), s)
			`),
		},
		{
			name: `dumps(dict, indent=-2)`,
			script: itn.HereDoc(`
				load('json', 'dumps')
				d = { "a" : "b", "c" : "d"}
				s = '''{"a":"b","c":"d"}'''
				assert.eq(dumps(d, indent=-2), s)
			`),
		},
		{
			name: `dumps(dict, indent=2)`,
			script: itn.HereDoc(`
				load('json', 'dumps')
				d = { "a" : "b", "c" : "d"}
				s = '''
				{
				  "a": "b",
				  "c": "d"
				}
				'''.strip()
				assert.eq(dumps(d, indent=2), s)
			`),
		},
		{
			name: `dumps struct`,
			script: itn.HereDoc(`
				load('json', 'dumps')
				load("struct.star", "struct")
				s = struct(a="Aloha", b=0x10, c=True, d=[1,2,3])
				d = '{"a":"Aloha","b":16,"c":true,"d":[1,2,3]}'
				assert.eq(dumps(s), d)
			`),
		},
		{
			name: `dumps module`,
			script: itn.HereDoc(`
				load('json', 'dumps')
				load("module.star", "module")
				m = module("hello", a="Bravo", b=0b10, c=False, d=[7,8,9])
				d = '{"a":"Bravo","b":2,"c":false,"d":[7,8,9]}'
				assert.eq(dumps(m), d)
			`),
		},
		{
			name: `dumps struct with json tag`,
			script: itn.HereDoc(`
				load('json', 'dumps')
				d = '{"msg":"Bravo","t":200}'
				assert.eq(dumps(vj), d)
			`),
		},
		{
			name: `dumps struct with star tag`,
			script: itn.HereDoc(`
				load('json', 'dumps')
				d = '{"msg":"Bravo","t":200}'
				assert.eq(dumps(vs), d)
			`),
		},
		{
			name: `encode struct`,
			script: itn.HereDoc(`
				load('json', 'encode')
				load("struct.star", "struct")
				s = struct(a="Aloha", b=0x10, c=True, d=[1,2,3])
				d = '{"a":"Aloha","b":16,"c":true,"d":[1,2,3]}'
				assert.eq(encode(s), d)
			`),
		},
		{
			name: `encode module`,
			script: itn.HereDoc(`
				load('json', 'encode')
				load("module.star", "module")
				m = module("hello", a="Bravo", b=0b10, c=False, d=[7,8,9])
				d = '{"a":"Bravo","b":2,"c":false,"d":[7,8,9]}'
				assert.eq(encode(m), d)
			`),
		},
		{
			name: `encode struct with json tag`,
			script: itn.HereDoc(`
				load('json', 'encode')
				d = '{"msg":"Bravo","t":200}'
				assert.eq(encode(vj), d)
			`),
		},
		{
			// notice that the result is different from the previous test for dumps,
			// because dumps() uses Go's json.Marshal() and encode() uses Starlark's lib json,
			// and the later one falls into GoStruct wraps.
			name: `encode struct with star tag`,
			script: itn.HereDoc(`
				load('json', 'encode')
				d = '{"Message":"Bravo","Times":200}'
				assert.eq(encode(vs), d)
			`),
		},
		{
			name: `try_dumps(1)`,
			script: itn.HereDoc(`
				load('json', 'try_dumps')
				assert.eq(try_dumps(1), ('1', None))
			`),
		},
		{
			name: `try_dumps(1, indent=-7)`,
			script: itn.HereDoc(`
				load('json', 'try_dumps')
				assert.eq(try_dumps(1, indent=-7), ('1', None))
			`),
		},
		{
			name: `try_dumps(1, indent=2)`,
			script: itn.HereDoc(`
				load('json', 'try_dumps')
				assert.eq(try_dumps(1, indent=2), ('1', None))
			`),
		},
		{
			name: `try_dumps(def)`,
			script: itn.HereDoc(`
				load('json', 'try_dumps')
				def f(): pass
				d = { "a" : "b", "f" : f}
				assert.eq(try_dumps(d), (None, 'unmarshaling starlark value: unrecognized starlark type: *starlark.Function'))
			`),
		},
		{
			name: `try_dumps(1, indent="abc")`,
			script: itn.HereDoc(`
				load('json', 'try_dumps')
				r, e = try_dumps(1, indent="abc")
				assert.eq(r, None)
				assert.true("got string, want int" in e)
			`),
		},
		{
			name: `try_encode no args`,
			script: itn.HereDoc(`
				load('json', 'try_encode')
				r, e = try_encode()
				assert.eq(r, None)
				assert.true("got 0 arguments, want 1" in e)
			`),
		},
		{
			name: `try_encode invalid`,
			script: itn.HereDoc(`
				load('json', 'try_encode')
				r, e = try_encode(lambda x: x+1)
				assert.eq(r, None)
				assert.true("cannot encode function as JSON" in e)
			`),
		},
		{
			name: `try_encode struct`,
			script: itn.HereDoc(`
				load('json', 'try_encode')
				load("struct.star", "struct")
				s = struct(a="Aloha", b=0x10, c=True, d=[1,2,3])
				d = '{"a":"Aloha","b":16,"c":true,"d":[1,2,3]}'
				assert.eq(try_encode(s), (d, None))
			`),
		},
		{
			name: `try_encode module`,
			script: itn.HereDoc(`
				load('json', 'try_encode')
				load("module.star", "module")
				m = module("hello", a="Bravo", b=0b10, c=False, d=[7,8,9])
				d = '{"a":"Bravo","b":2,"c":false,"d":[7,8,9]}'
				assert.eq(try_encode(m), (d, None))
			`),
		},
		{
			name: `try_encode struct with json tag`,
			script: itn.HereDoc(`
				load('json', 'try_encode')
				d = '{"msg":"Bravo","t":200}'
				assert.eq(try_encode(vj), (d, None))
			`),
		},
		{
			name: `try_encode struct with star tag`,
			script: itn.HereDoc(`
				load('json', 'try_encode')
				d = '{"Message":"Bravo","Times":200}'
				assert.eq(try_encode(vs), (d, None))
			`),
		},
		{
			name: `try_decode valid`,
			script: itn.HereDoc(`
				load('json', 'try_decode')
				d = '{"a": "b"}'
				assert.eq(try_decode(d), ({'a': 'b'}, None))
			`),
		},
		{
			name: `try_decode no args`,
			script: itn.HereDoc(`
				load('json', 'try_decode')
				r, e = try_decode()
				assert.eq(r, None)
				assert.true("missing argument for x" in e)
			`),
		},
		{
			name: `try_decode invalid`,
			script: itn.HereDoc(`
				load('json', 'try_decode')
				d = '{"a": b"}'
				r, e = try_decode(d)
				assert.eq(r, None)
				assert.true("unexpected character" in e)
			`),
		},
		{
			name: `try_indent no args`,
			script: itn.HereDoc(`
				load('json', 'try_indent')
				r, e = try_indent()
				assert.eq(r, None)
				assert.true("missing argument for str" in e)
			`),
		},
		{
			name: `try_indent valid`,
			script: itn.HereDoc(`
				load('json', 'try_indent')
				d = '{"a":"b","c":"d"}'
				expected = '''
				{
				  "a": "b",
				  "c": "d"
				}
				'''.strip()
				assert.eq(try_indent(d, indent='  '), (expected, None))
			`),
		},
		{
			name: `try_indent invalid`,
			script: itn.HereDoc(`
				load('json', 'try_indent')
				d = '{"a": b"}'
				r, e = try_indent(d, indent='  ')
				assert.eq(r, None)
				assert.true("invalid character" in e)
			`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := itn.ExecModuleWithErrorTest(t, json.ModuleName, json.LoadModule, tt.script, tt.wantErr, pred)
			if (err != nil) != (tt.wantErr != "") {
				t.Errorf("json(%q) expects error = '%v', actual error = '%v', result = %v", tt.name, tt.wantErr, err, res)
				return
			}
		})
	}
}

func TestJSONPathAndEvalFunctions(t *testing.T) {
	const jsonData = `
				{
					"store": {
						"book": [
							{ "category": "reference", "author": "Nigel Rees", "title": "Sayings of the Century", "price": 8.95 },
							{ "category": "fiction", "author": "Evelyn Waugh", "title": "Sword of Honour", "price": 12.99 },
							{ "category": "fiction", "author": "Herman Melville", "title": "Moby Dick", "isbn": "0-553-21311-3", "price": 8.99 },
							{ "category": "fiction", "author": "J. R. R. Tolkien", "title": "The Lord of the Rings", "isbn": "0-395-19395-8", "price": 22.99 }
						],
						"bicycle": { "color": "red", "price": 19.95 }
					}
				}`

	tests := []struct {
		name    string
		script  string
		want    string
		wantErr string
	}{
		{
			name: `stdlib can be loaded`,
			script: itn.HereDoc(`
				load('json', 'path', 'try_path', 'eval', 'try_eval')
			`),
		},

		// path
		{
			name: "json.path - missing path",
			script: itn.HereDoc(`
				load('json', 'path')
				data = '''` + jsonData + `'''
				result = path(data)
			`),
			wantErr: `json.path: missing argument for path`,
		},
		{
			name: "json.path - retrieve all prices",
			script: itn.HereDoc(`
				load('json', 'path')
				data = '''` + jsonData + `'''
				result = path(data, '$..price')
				assert.eq(result, [19.95, 8.95, 12.99, 8.99, 22.99])
			`),
		},
		{
			name: "json.path - retrieve all book titles",
			script: itn.HereDoc(`
				load('json', 'path')
				data = '''` + jsonData + `'''
				result = path(bytes(data), '$.store.book[*].title')
				assert.eq(result, ['Sayings of the Century', 'Sword of Honour', 'Moby Dick', 'The Lord of the Rings'])
			`),
		},
		{
			name: "json.path - retrieve non-existent path",
			script: itn.HereDoc(`
				load('json', 'path')
				data = '''` + jsonData + `'''
				result = path(data, '$.store.nonexistent')
				assert.eq(result, [])
			`),
		},
		{
			name: "json.path - invalid JSON data",
			script: itn.HereDoc(`
				load('json', 'path')
				data = 'invalid json data'
				path(data, '$..price')
			`),
			wantErr: "json.path: wrong symbol 'i' at 0",
		},
		{
			name: "json.path - data as starlark dict",
			script: itn.HereDoc(`
				load('json', 'path')
				data = {
					"a": [1, 2, {"b": 3}],
					"c": {"d": 4}
				}
				result = path(data, '$..b')
				assert.eq(result, [3])
			`),
		},
		{
			name: "json.path - data as starlark list",
			script: itn.HereDoc(`
				load('json', 'path')
				data = [1, 2, {"a": 3, "b": [4, 5]}]
				result = path(data, '$..b')
				assert.eq(result, [[4, 5]])
			`),
		},
		{
			name: "json.path - invalid JSONPath expression",
			script: itn.HereDoc(`
				load('json', 'path')
				data = '''` + jsonData + `'''
				path(data, '$..[?(@.price > 10)]X')
			`),
			wantErr: "json.path: wrong symbol 'X' at 20",
		},
		{
			name: "json.path - JSON data with multiple types",
			script: itn.HereDoc(`
				load('json', 'path')
				data = {
					"string": "text",
					"number": 42,
					"bool": True,
					"null": None,
					"array": [1, 2, {'key': 'value'}],
					"object": {"nested": {"inner": "value"}}
				}
				result = path(data, '$..*')
				assert.eq(result, [[1, 2, {"key": "value"}],True,None,42,{"nested": {"inner": "value"}},"text",1,2,{"key": "value"},{"inner": "value"},"value","value"])
			`),
		},
		{
			name: "json.path - wildcard and recursive descent",
			script: itn.HereDoc(`
				load('json', 'path')
				data = {'a': {'b': {'c': 1}}, 'd': {'b': {'c': 2}}}
				result = path(data, '$..b.c')
				assert.eq(result, [1, 2])
			`),
		},
		{
			name: "json.path - array slicing",
			script: itn.HereDoc(`
				load('json', 'path')
				data = [0, 1, 2, 3, 4, 5, 6, 7, 8, 9]
				result = path(data, '$[2:5]')
				assert.eq(result, [2,3,4])
			`),
		},
		{
			name: "json.path - filter expression",
			script: itn.HereDoc(`
				load('json', 'path')
				data = {'items': [{'value': 5}, {'value': 10}, {'value': 15}]}
				result = path(data, '$.items[?(@.value > 7)].value')
				assert.eq(result, [10, 15])
			`),
		},
		{
			name: "json.path - parent operator",
			script: itn.HereDoc(`
				load('json', 'path')
				data = {'store': {'book': [{'title': 'A', 'price': 5}, {'title': 'B', 'price': 15}]}}
				result = path(data, '$.store.book[?(@.price > 10)]..title')
				assert.eq(result, ['B'])
			`),
		},
		{
			name: "json.path - use of '@' current object",
			script: itn.HereDoc(`
				load('json', 'path')
				data = {'items': [{'name': 'A'}, {'name': 'B'}, {'name': 'C'}]}
				result = path(data, '$.items[?(@.name == "B")].name')
				assert.eq(result, ['B'])
			`),
		},
		{
			name: "json.path - access keys with special characters",
			script: itn.HereDoc(`
				load('json', 'path')
				data = {'weird:key': {'another:weird': 42}}
				result = path(data, '$["weird:key"]["another:weird"]')
				assert.eq(result, [42])
			`),
		},
		{
			name: "json.path - data as invalid type",
			script: itn.HereDoc(`
				load('json', 'path')
				data = lambda x: x + 5
				path(data, '$ + 5')
			`),
			wantErr: `json.path: unrecognized starlark type`,
		},
		{
			name: "json.path - int as key",
			script: itn.HereDoc(`
				load('json', 'path')
				data = {
					42: {"a": 100}
				}
				result = path(data, '$.42.a')
				assert.eq(result, [100])
			`),
		},
		{
			name: "json.path - broken array expression",
			script: itn.HereDoc(`
				load('json', 'path')
				data = [{"abc": 123}]
				result = path(data, '$..length')
				assert.eq(result, [1])
				result2 = path(data, '$..$#')
				assert.eq(result2, [])
			`),
		},

		// eval
		{
			name: "json.eval - missing expr",
			script: itn.HereDoc(`
				load('json', 'eval')
				data = '''` + jsonData + `'''
				result = eval(data)
			`),
			wantErr: `json.eval: missing argument for expr`,
		},
		{
			name: "json.eval - average price of all items",
			script: itn.HereDoc(`
				load('json', 'eval')
				data = '''` + jsonData + `'''
				result = eval(data, 'avg($..price)')
				assert.eq(result, 14.774)
			`),
		},
		{
			name: "json.eval - sum of all book prices",
			script: itn.HereDoc(`
				load('json', 'eval')
				data = '''` + jsonData + `'''
				result = eval(data, 'sum($.store.book[*].price)')
				assert.eq(result, 53.92)
			`),
		},
		{
			name: "json.eval - invalid expression",
			script: itn.HereDoc(`
				load('json', 'eval')
				data = '''` + jsonData + `'''
				eval(data, 'invalid($..price)')
			`),
			wantErr: "json.eval: wrong request: wrong formula, 'invalid' is not a function",
		},
		{
			name: "json.eval - division by zero",
			script: itn.HereDoc(`
				load('json', 'eval')
				data = '''` + jsonData + `'''
				eval(data, '10 / 0')
			`),
			wantErr: "json.eval: wrong request: division by zero",
		},
		{
			name: "json.eval - data as starlark dict",
			script: itn.HereDoc(`
				load('json', 'eval')
				data = {
					"numbers": [1, 2, 3, 4, 5]
				}
				result = eval(data, 'sum($.numbers)')
				assert.eq(result, 15)
			`),
		},
		{
			name: "json.eval - division with floating point result",
			script: itn.HereDoc(`
				load('json', 'eval')
				data = '''` + jsonData + `'''
				result = eval(data, 'sum($..price) / size($..price)')
				assert.eq(result, 14.774)
			`),
		},
		{
			name: "json.eval - expression returning boolean",
			script: itn.HereDoc(`
				load('json', 'eval')
				data = {'value': 10}
				result = eval(data, '$.value > 5')
				assert.true(result)
			`),
		},
		{
			name: "json.eval - expression returning string",
			script: itn.HereDoc(`
				load('json', 'eval')
				data = {'greeting': 'Hello', 'name': 'World'}
				result = eval(data, '$.greeting + ", " + $.name + "!"')
				assert.eq(result, 'Hello, World!')
			`),
		},
		{
			name: "json.eval - nested JSONPath in expression",
			script: itn.HereDoc(`
				load('json', 'eval')
				data = {'nums': [1, 2, 3, 4, 5]}
				result = eval(data, 'sum($.nums[0:3])')
				assert.eq(result, 6)
			`),
		},
		{
			name: "json.eval - accessing non-existent key",
			script: itn.HereDoc(`
				load('json', 'eval')
				data = {'a': 1}
				result = eval(data, '$.b')
				assert.eq(result, None)
			`),
		},
		{
			name: "json.eval - use of built-in constants",
			script: itn.HereDoc(`
				load('json', 'eval')
				data = {}
				result = eval(data, 'pi * 2')
				assert.eq(result, 6.283185307179586)
			`),
		},
		{
			name: "json.eval - calling undefined function",
			script: itn.HereDoc(`
				load('json', 'eval')
				data = {}
				eval(data, 'undefined_function()')
			`),
			wantErr: "json.eval: wrong request: wrong formula, 'undefined_function' is not a function",
		},
		{
			name: "json.eval - invalid syntax in expression",
			script: itn.HereDoc(`
				load('json', 'eval')
				data = {}
				eval(data, 'sum(')
			`),
			wantErr: "json.eval: wrong request: wrong formula, '(' is not an operation or function",
		},
		{
			name: "json.eval - invalid data",
			script: itn.HereDoc(`
				load('json', 'eval')
				eval('{"a": 123', '$ + 5')
			`),
			wantErr: "json.eval: unexpected end of file",
		},
		{
			name: "json.eval - accessing array elements by index",
			script: itn.HereDoc(`
				load('json', 'eval')
				data = {'array': [10, 20, 30]}
				result = eval(data, '$.array[1]')
				assert.eq(result, 20)
			`),
		},
		{
			name: "json.eval - data as invalid type",
			script: itn.HereDoc(`
				load('json', 'eval')
				data = lambda x: x + 5
				eval(data, '$ + 5')
			`),
			wantErr: `json.eval: unrecognized starlark type`,
		},
		{
			name: "json.eval - int as key",
			script: itn.HereDoc(`
				load('json', 'eval')
				data = {
					42: {"a": 100}
				}
				result = eval(data, '$.42.a')
				assert.eq(result, 100)
			`),
		},
		{
			name: "json.eval - broken array expression",
			script: itn.HereDoc(`
				load('json', 'eval')
				data = [{"abc": 123}]
				r0 = eval(data, '$..length')
				assert.eq(r0, 1)
				r1 = eval(data, '$.')
				assert.eq(r1, [{"abc": 123}])
				r2 = eval(data, '$..')
				assert.eq(r2, [[{"abc": 123}],{"abc": 123}])
				r3 = eval(data, '$...')
				assert.eq(r3, [[{"abc": 123}],{"abc": 123},{"abc": 123}])
			`),
		},

		// try path
		{
			name: "json.try_path - missing path",
			script: itn.HereDoc(`
				load('json', 'try_path')
				data = '''` + jsonData + `'''
				result, err = try_path(data)
				assert.eq(result, None)
				assert.true('json.try_path: missing argument for path' in err)
			`),
		},
		{
			name: "json.try_path - retrieve all prices",
			script: itn.HereDoc(`
				load('json', 'try_path')
				data = '''` + jsonData + `'''
				result, err = try_path(data, '$..price')
				assert.eq(result, [19.95, 8.95, 12.99, 8.99, 22.99])
				assert.eq(err, None)
			`),
		},
		{
			name: "json.try_path - invalid JSONPath",
			script: itn.HereDoc(`
				load('json', 'try_path')
				data = '''` + jsonData + `'''
				result, err = try_path(data, '$..[invalid]')
				assert.eq(result, [])
				assert.eq(err, None)
			`),
		},
		{
			name: "json.try_path - wrong JSONPath",
			script: itn.HereDoc(`
				load('json', 'try_path')
				data = '''` + jsonData + `'''
				result, err = try_path(data, '$..[invalid]X')
				assert.eq(result, None)
				assert.true("wrong symbol 'X' at 12" in err)
			`),
		},
		{
			name: "json.try_path - data as invalid JSON",
			script: itn.HereDoc(`
				load('json', 'try_path')
				data = '{"invalid": json'
				result, err = try_path(data, '$..*')
				assert.eq(result, None)
				assert.true("wrong symbol 'j' at 12" in err)
			`),
		},
		{
			name: "json.try_path - data as invalid type",
			script: itn.HereDoc(`
				load('json', 'try_path')
				data = lambda x: x + 5
				result, err = try_path(data, '$ + 5')
				assert.eq(result, None)
				assert.true("unrecognized starlark" in err)
			`),
		},

		// try eval
		{
			name: "json.try_eval - average price of all items",
			script: itn.HereDoc(`
				load('json', 'try_eval')
				data = '''` + jsonData + `'''
				result, err = try_eval(data, 'avg($..price)')
				assert.eq(result, 14.774)
				assert.eq(err, None)
			`),
		},
		{
			name: "json.try_eval - invalid expression",
			script: itn.HereDoc(`
				load('json', 'try_eval')
				data = '''` + jsonData + `'''
				result, err = try_eval(data, 'invalid($..price)')
				assert.eq(result, None)
				assert.true("wrong request: wrong formula, 'invalid' is not a function" in err)
			`),
		},
		{
			name: "json.try_eval - missing expr",
			script: itn.HereDoc(`
				load('json', 'try_eval')
				data = '''` + jsonData + `'''
				result, err = try_eval(data)
				assert.eq(result, None)
				assert.true('json.try_eval: missing argument for expr' in err)
			`),
		},
		{
			name: "json.try_eval - division by zero",
			script: itn.HereDoc(`
				load('json', 'try_eval')
				data = '''` + jsonData + `'''
				result, err = try_eval(data, '10 / 0')
				assert.eq(result, None)
				assert.true("division by zero" in err)
			`),
		},
		{
			name: "json.try_eval - data as invalid type",
			script: itn.HereDoc(`
				load('json', 'try_eval')
				data = lambda x: x + 5
				result, err = try_eval(data, '$ + 5')
				assert.eq(result, None)
				assert.true("unrecognized starlark" in err)
			`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := itn.ExecModuleWithErrorTest(t, json.ModuleName, json.LoadModule, tt.script, tt.wantErr, nil)
			if (err != nil) != (tt.wantErr != "") {
				t.Errorf("json(%q) expects error = '%v', actual error = '%v', result = %v", tt.name, tt.wantErr, err, res)
				return
			}
		})
	}
}
