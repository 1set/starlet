package csv_test

import (
	"testing"

	itn "github.com/1set/starlet/internal"
	libcsv "github.com/1set/starlet/lib/csv"
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := itn.ExecModuleWithErrorTest(t, libcsv.ModuleName, libcsv.LoadModule, tt.script, tt.wantErr, nil)
			if (err != nil) != (tt.wantErr != "") {
				t.Errorf("csv(%q) expects error = '%v', actual error = '%v', result = %v", tt.name, tt.wantErr, err, res)
				return
			}
		})
	}
}
