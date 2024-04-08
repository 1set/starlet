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
			name: `write_all: no args`,
			script: itn.HereDoc(`
load('csv', 'write_all')
write_all()
			`),
			wantErr: "csv.write_all: missing argument for source",
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
			name: `read_all`,
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
			name: `write_all`,
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
