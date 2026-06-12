package net

import (
	"testing"
)

// White-box checks of internals that the script-level tests cannot reach.

// TestNewPingTransport pins that httping's transport inherits the default
// transport's settings — most importantly the system proxy hook — instead
// of a zero-value Transport that never consults a proxy (LET-14).
func TestNewPingTransport(t *testing.T) {
	tr := newPingTransport()
	if tr.Proxy == nil {
		t.Errorf("expected the ping transport to inherit the system proxy hook")
	}
	if !tr.DisableKeepAlives {
		t.Errorf("expected keep-alives to stay disabled for ping clients")
	}
}
