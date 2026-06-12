package csv_test

import (
	"testing"
	"time"

	itn "github.com/1set/starlet/internal"
	libcsv "github.com/1set/starlet/lib/csv"
	startime "go.starlark.net/lib/time"
	"go.starlark.net/starlark"
)

func TestLoadModule_CSV(t *testing.T) {
	tests := []struct {
		name    string
		script  string
		wantErr string
	}{
		{
			name: `read_all: no args`,
			script: itn.HereDoc(`
load('csv', 'read_all')
read_all()
			`),
			wantErr: "csv.read_all: missing argument for source",
		},
		{
			name: `read_all: invalid type`,
			script: itn.HereDoc(`
load('csv', 'read_all')
read_all(1)
			`),
			wantErr: "csv.read_all: for parameter source: got int, want string",
		},
		{
			name: `read_all: invalid comma`,
			script: itn.HereDoc(`
load('csv', 'read_all')
read_all("1,2,3", comma=", ")
			`),
			wantErr: "csv.read_all: expected comma param to be a single-character string",
		},
		{
			name: `read_all: invalid comment`,
			script: itn.HereDoc(`
load('csv', 'read_all')
read_all("1,2,3", comment="##")
			`),
			wantErr: "csv.read_all: expected comment param to be a single-character string",
		},
		{
			name: `read_all: empty`,
			script: itn.HereDoc(`
load('csv', 'read_all')
assert.eq(read_all(''), [])
			`),
		},
		{
			name: `read_all: one`,
			script: itn.HereDoc(`
load('csv', 'read_all')
assert.eq(read_all('1,2,3\n\n'), [["1","2","3"]])
			`),
		},
		{
			name: `read_all: normal`,
			script: itn.HereDoc(`
load('csv', 'read_all')
csv_string_1 = """a,b,c
1,2,3
4,5,6
7,8,9
"""
assert.eq(read_all(csv_string_1), [["a","b","c"],["1","2","3"],["4","5","6"],["7","8","9"]])
			`),
		},
		{
			name: `read_all: abnormal`,
			script: itn.HereDoc(`
load('csv', 'read_all')
csv_string_1 = """a,b,c,d
1,2,3
4,5,6
7,8,9
"""
assert.eq(read_all(csv_string_1, fields_per_record=-1), [["a","b","c", "d"],["1","2","3"],["4","5","6"],["7","8","9"]])
			`),
		},
		{
			name: `read_all: check`,
			script: itn.HereDoc(`
load('csv', 'read_all')
csv_string_1 = """a,b,c,d
1,2,3
4,5,6
7,8,9
"""
assert.eq(read_all(csv_string_1, fields_per_record=3), [["a","b","c", "d"],["1","2","3"],["4","5","6"],["7","8","9"]])
			`),
			wantErr: `csv.read_all: record on line 1: wrong number of fields`,
		},
		{
			name: `read_all: skip check`,
			script: itn.HereDoc(`
load('csv', 'read_all')
csv_string_1 = """a,b,c,d
1,2,3
4,5,6
7,8,9
"""
assert.eq(read_all(csv_string_1, skip=1, fields_per_record=3), [["a","b","c", "d"],["1","2","3"],["4","5","6"],["7","8","9"]])
			`),
			wantErr: `csv.read_all: record on line 1: wrong number of fields`,
		},
		{
			name: `read_all: default check`,
			script: itn.HereDoc(`
load('csv', 'read_all')
csv_string_1 = """a,b,c
1,2,3,4
4,5,6
7,8,9
"""
assert.eq(read_all(csv_string_1), [["a","b","c", "d"],["1","2","3"],["4","5","6"],["7","8","9"]])
			`),
			wantErr: `csv.read_all: record on line 2: wrong number of fields`,
		},
		{
			name: `read_all: skip`,
			script: itn.HereDoc(`
load('csv', 'read_all')
csv_string_1 = """a,b,c
1,2,3
4,5,6
7,8,9
"""
assert.eq(read_all(csv_string_1, skip=1, limit=2, fields_per_record=3), [["1","2","3"],["4","5","6"]])
			`),
		},
		{
			name: `read_all: comment and comma`,
			script: itn.HereDoc(`
load('csv', 'read_all')
csv_string_1 = """a|b|c
#1,2,3
4|5|6
7|8|9
"""
assert.eq(read_all(csv_string_1, comma="|", comment="#"), [["a","b","c"],["4","5","6"],["7","8","9"]])
			`),
		},
		{
			name: `read_all: malformed row after limit`,
			script: itn.HereDoc(`
load('csv', 'read_all')
csv_string = 'a,b\nc,d\n"bad\n'
assert.eq(read_all(csv_string, limit=2), [["a","b"],["c","d"]])
			`),
		},
		{
			name: `read_all: malformed row within limit`,
			script: itn.HereDoc(`
load('csv', 'read_all')
read_all('a,b\n"bad\n', limit=2)
			`),
			wantErr: `csv.read_all: parse error on line 2`,
		},
		{
			name: `try_read_all: ok and error`,
			script: itn.HereDoc(`
load('csv', 'try_read_all')
rows, err = try_read_all('a,b\nc,d\n')
assert.eq(err, None)
assert.eq(rows, [["a","b"],["c","d"]])
bad, err2 = try_read_all('"bad\n')
assert.eq(bad, None)
assert.true('parse error' in err2)
			`),
		},
		{
			name: `read_dict: normal`,
			script: itn.HereDoc(`
load('csv', 'read_dict')
rows = read_dict('a,b\n1,2\n3,4\n')
assert.eq(rows, [{"a": "1", "b": "2"}, {"a": "3", "b": "4"}])
			`),
		},
		{
			name: `read_dict: skip and limit`,
			script: itn.HereDoc(`
load('csv', 'read_dict')
csv_string = '# note\na,b\n1,2\n3,4\n'
rows = read_dict(csv_string, skip=1, limit=1)
assert.eq(rows, [{"a": "1", "b": "2"}])
			`),
		},
		{
			name: `read_dict: empty source`,
			script: itn.HereDoc(`
load('csv', 'read_dict')
assert.eq(read_dict(''), [])
			`),
		},
		{
			name: `read_dict: duplicate header`,
			script: itn.HereDoc(`
load('csv', 'read_dict')
read_dict('a,a\n1,2\n')
			`),
			wantErr: `csv.read_dict: duplicate header field "a"`,
		},
		{
			name: `read_dict: mismatched row`,
			script: itn.HereDoc(`
load('csv', 'read_dict')
read_dict('a,b\n1\n')
			`),
			wantErr: `wrong number of fields`,
		},
		{
			name: `try_read_dict: error`,
			script: itn.HereDoc(`
load('csv', 'try_read_dict')
rows, err = try_read_dict('a,a\n1,2\n')
assert.eq(rows, None)
assert.true('duplicate header' in err)
			`),
		},
		{
			name: `write_all: no args`,
			script: itn.HereDoc(`
load('csv', 'write_all')
write_all()
			`),
			wantErr: "csv.write_all: missing argument for data",
		},
		{
			name: `write_all: invalid comma`,
			script: itn.HereDoc(`
load('csv', 'write_all')
write_all([[1,2,3]], comma=", ")
			`),
			wantErr: "csv.write_all: expected comma param to be a single-character string",
		},
		{
			name: `write_all: invalid data`,
			script: itn.HereDoc(`
load('csv', 'write_all')
def hello():
	print("Hello, World!")
write_all(hello)
			`),
			wantErr: "csv.write_all: unrecognized starlark type: *starlark.Function",
		},
		{
			name: `write_all: invalid type`,
			script: itn.HereDoc(`
load('csv', 'write_all')
write_all(1)
			`),
			wantErr: "csv.write_all: expected value to be an array type",
		},
		{
			name: `write_all: invalid list`,
			script: itn.HereDoc(`
load('csv', 'write_all')
write_all([1,2,3])
			`),
			wantErr: "csv.write_all: row 0 is not an array type",
		},
		{
			name: `write_all: empty`,
			script: itn.HereDoc(`
load('csv', 'write_all')
assert.eq(write_all([]), "")
assert.eq(write_all([[]]), "\n")
			`),
		},
		{
			name: `write_all: one`,
			script: itn.HereDoc(`
load('csv', 'write_all')
assert.eq(write_all([["1","2","3"]]), "1,2,3\n")
			`),
		},
		{
			name: `write_all: normal`,
			script: itn.HereDoc(`
load('csv', 'write_all')
csv_data = [[1,2,3],[4,5,6],['a','b','c']]
csv_data_string = """1,2,3
4,5,6
a,b,c
"""
assert.eq(write_all(csv_data), csv_data_string)
			`),
		},
		{
			name: `write_all: cell types`,
			script: itn.HereDoc(`
load('csv', 'write_all')
assert.eq(write_all([[1000000.0, 0.00001, -2.5]]), "1000000,0.00001,-2.5\n")
assert.eq(write_all([[None, True, False]]), ",true,false\n")
assert.eq(write_all([[test_time]]), "2023-01-15T12:30:45Z\n")
			`),
		},
		{
			name: `write_all: non-finite float cell`,
			script: itn.HereDoc(`
load('csv', 'write_all')
write_all([[float("nan")]])
			`),
			wantErr: `not representable in CSV`,
		},
		{
			name: `write_all: nested cell`,
			script: itn.HereDoc(`
load('csv', 'write_all')
write_all([[[1, 2]]])
			`),
			wantErr: `unsupported cell type`,
		},
		{
			name: `write_dict: no args`,
			script: itn.HereDoc(`
load('csv', 'write_dict')
write_dict()
			`),
			wantErr: "csv.write_dict: missing argument for data",
		},
		{
			name: `write_dict: invalid header`,
			script: itn.HereDoc(`
load('csv', 'write_dict')
write_dict([], header="a")
			`),
			wantErr: "csv.write_dict: for parameter \"header\": got string, want iterable",
		},
		{
			name: `write_dict: invalid header type`,
			script: itn.HereDoc(`
load('csv', 'write_dict')
write_dict([], header=[1,2,3])
			`),
			wantErr: "csv.write_dict: for parameter header: got int, want string",
		},
		{
			name: `write_dict: empty header`,
			script: itn.HereDoc(`
load('csv', 'write_dict')
write_dict([], header=[])
			`),
			wantErr: "csv.write_dict: header cannot be empty",
		},
		{
			name: `write_dict: invalid comma`,
			script: itn.HereDoc(`
load('csv', 'write_dict')
write_dict([], header=["a"], comma=", ")
			`),
			wantErr: "csv.write_dict: expected comma param to be a single-character string",
		},
		{
			name: `write_dict: invalid data`,
			script: itn.HereDoc(`
load('csv', 'write_dict')
x = write_dict("123", header=["a"])
			`),
			wantErr: `csv.write_dict: expected value to be an array type`,
		},
		{
			name: `write_all: invalid data type`,
			script: itn.HereDoc(`
load('csv', 'write_dict')
def hello():
	print("Hello, World!")
write_dict(hello, header=["a"])
			`),
			wantErr: "csv.write_dict: unrecognized starlark type: *starlark.Function",
		},
		{
			name: `write_dict: invalid list`,
			script: itn.HereDoc(`
load('csv', 'write_dict')
x = write_dict([[1,2]], header=["a"])
			`),
			wantErr: `csv.write_dict: expected value to be a map type`,
		},
		{
			name: `write_dict: empty`,
			script: itn.HereDoc(`
load('csv', 'write_dict')
x = write_dict([], header=["a"])
assert.eq(x, "a\n")
			`),
		},
		{
			name: `write_dict: one`,
			script: itn.HereDoc(`
load('csv', 'write_dict')
x = write_dict([{"a": 100}], header=["a"])
assert.eq(x, "a\n100\n")
			`),
		},
		{
			name: `write_dict: cell types`,
			script: itn.HereDoc(`
load('csv', 'write_dict')
x = write_dict([{"a": 1000000.0, "b": None, "c": True}], header=["a", "b", "c", "d"])
assert.eq(x, "a,b,c,d\n1000000,,true,\n")
			`),
		},
		{
			name: `write_dict: nested cell`,
			script: itn.HereDoc(`
load('csv', 'write_dict')
write_dict([{"a": {"x": 1}}], header=["a"])
			`),
			wantErr: `unsupported cell type`,
		},
		{
			name: `try_write_all: ok and error`,
			script: itn.HereDoc(`
load('csv', 'try_write_all')
text, err = try_write_all([[1, 2]])
assert.eq(err, None)
assert.eq(text, "1,2\n")
bad, err2 = try_write_all([[[1]]])
assert.eq(bad, None)
assert.true('unsupported cell type' in err2)
			`),
		},
		{
			name: `try_write_dict: ok and error`,
			script: itn.HereDoc(`
load('csv', 'try_write_dict')
text, err = try_write_dict([{"a": 1}], header=["a"])
assert.eq(err, None)
assert.eq(text, "a\n1\n")
bad, err2 = try_write_dict([{"a": {}}], header=["a"])
assert.eq(bad, None)
assert.true('unsupported cell type' in err2)
			`),
		},
		{
			name: `write_dict: normal`,
			script: itn.HereDoc(`
load('csv', 'write_dict')
x = write_dict([{"a": 200, "b": 100, "c": 500},{"b": 1024, "C": 2048}], header=["c","b"])
assert.eq(x, "c,b\n500,100\n,1024\n")
			`),
		},
		{
			name: `read_all: with UTF-8 BOM`,
			script: itn.HereDoc(`
load('csv', 'read_all')
# UTF-8 BOM is represented by bytes EF BB BF at the beginning of the file
# In this test we use the hex representation as a string
csv_with_bom = b"\xef\xbb\xbfa,b,c\n1,2,3\n4,5,6"
assert.eq(read_all(csv_with_bom), [["a","b","c"],["1","2","3"],["4","5","6"]])
			`),
		},
		{
			name: `read_all: with UTF-8 BOM and different comma`,
			script: itn.HereDoc(`
load('csv', 'read_all')
# UTF-8 BOM with semicolon as delimiter
csv_with_bom = b"\xef\xbb\xbfa;b;c\n1;2;3\n4;5;6"
assert.eq(read_all(csv_with_bom, comma=";"), [["a","b","c"],["1","2","3"],["4","5","6"]])
			`),
		},
		{
			name: `read_all: with UTF-8 BOM and comments`,
			script: itn.HereDoc(`
load('csv', 'read_all')
# UTF-8 BOM with comments
csv_with_bom = b"\xef\xbb\xbfa,b,c\n#comment line\n1,2,3\n4,5,6"
assert.eq(read_all(csv_with_bom, comment="#"), [["a","b","c"],["1","2","3"],["4","5","6"]])
			`),
		},
		{
			name: `read_all: ensure BOM is properly removed`,
			script: itn.HereDoc(`
load('csv', 'read_all')
# UTF-8 BOM should be removed properly and not affect the first field
# If not handled properly, the first character "a" would include the BOM bytes
csv_with_bom = b"\xef\xbb\xbfa,b,c\n1,2,3\n4,5,6"
result = read_all(csv_with_bom)
first_field = result[0][0]
assert.eq(first_field, "a")
assert.eq(len(first_field), 1)  # Length should be 1, not 4 (3 BOM bytes + "a")
			`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			predecl := starlark.StringDict{
				"test_time": startime.Time(time.Date(2023, 1, 15, 12, 30, 45, 0, time.UTC)),
			}
			res, err := itn.ExecModuleWithErrorTest(t, libcsv.ModuleName, libcsv.LoadModule, tt.script, tt.wantErr, predecl)
			if (err != nil) != (tt.wantErr != "") {
				t.Errorf("csv(%q) expects error = '%v', actual error = '%v', result = %v", tt.name, tt.wantErr, err, res)
				return
			}
		})
	}
}
