package stats_test

import (
	"testing"

	itn "github.com/1set/starlet/internal"
)

func TestLoadModule_Stats(t *testing.T) {
	tests := []struct {
		name    string
		script  string
		wantErr string
	}{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := itn.ExecModuleWithErrorTest(t, stats.ModuleName, stats.LoadModule, tt.script, tt.wantErr, nil)
			if (err != nil) != (tt.wantErr != "") {
				t.Errorf("stats(%q) expects error = '%v', actual error = '%v', result = %v", tt.name, tt.wantErr, err, res)
				return
			}
		})
	}
}
