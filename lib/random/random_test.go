package random_test

import (
	"testing"

	itn "github.com/1set/starlet/lib/internal"
	"github.com/1set/starlet/lib/random"
	"go.starlark.net/starlark"
)

func TestLoadModule_Hash(t *testing.T) {
	var (
		repeatTimes = 10
		one         = starlark.MakeInt(1)
		two         = starlark.MakeInt(2)
		three       = starlark.MakeInt(3)
	)
	tests := []struct {
		name        string
		script      string
		wantErr     string
		checkResult func(res starlark.Value) bool
	}{
		{
			name: `nil choice`,
			script: itn.HereDoc(`
				load('random', 'choice')
				choice()
			`),
			wantErr: `choice: missing argument for seq`,
		},
		{
			name: `no choice`,
			script: itn.HereDoc(`
				load('random', 'choice')
				choice([])
			`),
			wantErr: `cannot choose from an empty sequence`,
		},
		{
			name: `one choice`,
			script: itn.HereDoc(`
				load('random', 'choice')
				val = choice([1])
			`),
			checkResult: func(res starlark.Value) bool {
				return res.(starlark.Int) == one
			},
		},
		{
			name: `two choices`,
			script: itn.HereDoc(`
				load('random', 'choice')
				val = choice([1, 2])
			`),
			checkResult: func(res starlark.Value) bool {
				val := res.(starlark.Int)
				return val == one || val == two
			},
		},
		{
			name: `duplicate choices`,
			script: itn.HereDoc(`
				load('random', 'choice')
				val = choice([1, 1, 2, 3, 3, 2, 1])
			`),
			checkResult: func(res starlark.Value) bool {
				val := res.(starlark.Int)
				return val == one || val == two || val == three
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for i := 1; i <= repeatTimes; i++ {
				res, err := itn.ExecModuleWithErrorTest(t, random.ModuleName, random.LoadModule, tt.script, tt.wantErr)
				if (err != nil) != (tt.wantErr != "") {
					t.Errorf("random(%q) expects error = '%v', actual error = '%v', result = %v", tt.name, tt.wantErr, err, res)
					return
				}
				if tt.wantErr != "" {
					return
				}
				if tt.checkResult != nil && !tt.checkResult(res["val"]) {
					t.Errorf("random(%q) got unexpected result, actual result = %v", tt.name, res)
				}
			}
		})
	}
}
