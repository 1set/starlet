package net

import (
	"context"
	"errors"
	"net/http"
	"reflect"
	"testing"
	"time"
)

// White-box checks of internals that the script-level tests cannot reach.
// Sections: transport selection; cancellation before and between ping rounds.

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

// TestGoPingWrapCancellation controls cancellation directly instead of racing
// a one-second timeout against a one-second interval in the public API test.
// Probe callbacks perform no network I/O, and the long pause cannot win over
// the already-cancelled context.
func TestGoPingWrapCancellation(t *testing.T) {
	probeErr := errors.New("controlled probe failure")
	for _, tc := range []struct {
		name         string
		preCancelled bool
		cancelProbe  bool
		probeErr     error
		interval     time.Duration
		wantCalls    int
		wantRTTs     []time.Duration
	}{
		{name: "cancelled before first probe", preCancelled: true, interval: time.Hour},
		{name: "cancelled after successful probe", cancelProbe: true, interval: time.Hour, wantCalls: 1},
		{name: "cancelled after failed probe", cancelProbe: true, probeErr: probeErr, interval: time.Hour, wantCalls: 1},
		{name: "successful rounds", wantCalls: 3, wantRTTs: []time.Duration{time.Millisecond, 2 * time.Millisecond, 3 * time.Millisecond}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			if tc.preCancelled {
				cancel()
			}
			const address = "fixture.invalid:1234"
			const timeout = 7 * time.Millisecond
			calls := 0
			var addresses []string
			var timeouts []time.Duration
			probe := func(_ context.Context, gotAddress string, gotTimeout time.Duration) (time.Duration, error) {
				calls++
				addresses = append(addresses, gotAddress)
				timeouts = append(timeouts, gotTimeout)
				if tc.cancelProbe {
					cancel()
				}
				return time.Duration(calls) * time.Millisecond, tc.probeErr
			}
			rtts, err := runPingWrapBounded(t, ctx, address, 3, timeout, tc.interval, probe)
			if calls != tc.wantCalls {
				t.Fatalf("probe calls = %d, want %d", calls, tc.wantCalls)
			}
			if tc.preCancelled || tc.cancelProbe {
				if !errors.Is(err, context.Canceled) || rtts != nil {
					t.Fatalf("cancelled result = (%v, %v), want (nil, context.Canceled)", rtts, err)
				}
			} else if err != nil || !reflect.DeepEqual(rtts, tc.wantRTTs) {
				t.Fatalf("successful result = (%v, %v), want (%v, nil)", rtts, err, tc.wantRTTs)
			}
			for i := range addresses {
				if addresses[i] != address || timeouts[i] != timeout {
					t.Errorf("probe %d received (%q, %v), want (%q, %v)", i, addresses[i], timeouts[i], address, timeout)
				}
			}
		})
	}
}

func runPingWrapBounded(t *testing.T, ctx context.Context, address string, count int, timeout, interval time.Duration, probe func(context.Context, string, time.Duration) (time.Duration, error)) ([]time.Duration, error) {
	t.Helper()
	type result struct {
		rtts []time.Duration
		err  error
	}
	done := make(chan result, 1)
	go func() {
		rtts, err := goPingWrap(ctx, address, count, timeout, interval, probe)
		done <- result{rtts: rtts, err: err}
	}()
	watchdog := time.NewTimer(5 * time.Second)
	defer watchdog.Stop()
	select {
	case got := <-done:
		return got.rtts, got.err
	case <-watchdog.C:
		t.Fatal("goPingWrap did not return after the controlled cancellation")
		return nil, nil
	}
}
