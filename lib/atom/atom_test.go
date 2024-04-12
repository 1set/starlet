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
		{
			name: `int: new, no args`,
			script: itn.HereDoc(`
				load('atom', 'new_int')
				x = new_int()
				assert.eq(str(x), '<atom_int:0>')
				assert.eq(x.get(), 0)
			`),
		},
		{
			name: `int: new, with args`,
			script: itn.HereDoc(`
				load('atom', 'new_int')
				x = new_int(42)
				assert.eq(str(x), '<atom_int:42>')
				assert.eq(x.get(), 42)
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
