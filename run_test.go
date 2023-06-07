package starlet_test

import (
	"context"
	"os"
	"starlet"
	"testing"
)

func expectErr(t *testing.T, err error, expected string) {
	if err == nil {
		t.Errorf("unexpected nil error")
	}
	if err.Error() != expected {
		t.Errorf(`expected error %q, got: %v`, expected, err)
	}
}

func Test_EmptyMachine_RunNoCode(t *testing.T) {
	m := starlet.NewEmptyMachine()
	// run with empty script
	_, err := m.Run(context.Background())
	expectErr(t, err, `starlet: run: no code to run`)
}

func Test_EmptyMachine_RunNoSpecificCode(t *testing.T) {
	m := starlet.NewEmptyMachine()
	m.SetScript("", nil, os.DirFS("example"))
	// run with no specific file name
	_, err := m.Run(context.Background())
	expectErr(t, err, `starlet: run: no code to run`)
}
