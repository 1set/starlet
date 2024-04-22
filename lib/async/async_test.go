package async_test

import (
	"testing"

	itn "github.com/1set/starlet/internal"
	la "github.com/1set/starlet/lib/async"
	lg "github.com/1set/starlet/lib/goidiomatic"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

func TestLoadModule_Async(t *testing.T) {
	tests := []struct {
		name    string
		script  string
		wantErr string
	}{
		{
			name: `basic`,
			script: itn.HereDoc(`
				load('async', 'run')
				print(type(run))
			`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := itn.ExecModuleWithErrorTest(t, la.ModuleName, LoadModules, tt.script, tt.wantErr, nil)
			if (err != nil) != (tt.wantErr != "") {
				t.Errorf("async(%q) expects error = '%v', actual error = '%v', result = %v", tt.name, tt.wantErr, err, res)
				return
			}
		})
	}
}

func LoadModules() (starlark.StringDict, error) {
	sd := starlark.StringDict{}
	// load go idiomatic module
	if md, err := lg.LoadModule(); err == nil {
		for k, v := range md {
			sd[k] = v
		}
	}
	// load async module
	if md, err := la.LoadModule(); err == nil {
		for _, mv := range md {
			if mm, ok := mv.(*starlarkstruct.Module); ok && mm != nil {
				for k, v := range mm.Members {
					sd[k] = v
				}
			}
		}
	}
	return sd, nil
}
