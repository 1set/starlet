package starlet_test

import (
	"context"
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
	_, err := m.Run(context.Background())
	expectErr(t, err, `starlet: run: no code to run`)
	if err == nil {
		t.Errorf("expected error, got nil")
	}
}
