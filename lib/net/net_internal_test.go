package net

import (
	"net/http"
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

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

// TestNewPingTransportFallback pins the fallback used when the host has
// replaced http.DefaultTransport with a custom RoundTripper.
func TestNewPingTransportFallback(t *testing.T) {
	orig := http.DefaultTransport
	defer func() { http.DefaultTransport = orig }()
	http.DefaultTransport = roundTripperFunc(func(r *http.Request) (*http.Response, error) { return nil, nil })

	tr := newPingTransport()
	if tr.Proxy == nil {
		t.Errorf("expected the fallback transport to keep the system proxy hook")
	}
	if !tr.DisableKeepAlives {
		t.Errorf("expected keep-alives to stay disabled in the fallback transport")
	}
}
