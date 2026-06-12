package net_test

import (
	"fmt"
	stdnet "net"
	"net/http"
	"net/http/httptest"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/1set/starlet"
	itn "github.com/1set/starlet/internal"
	"github.com/1set/starlet/lib/net"
	"go.starlark.net/starlark"
)

// A helper function to create a mock server that returns the specified status code
func createMockServer(statusCode int) *httptest.Server {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(statusCode)
	})
	return httptest.NewServer(handler)
}

// Create a mock server that returns a 301 status code with a Location header
func createRedirectMockServer(location string) *httptest.Server {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Location", location)
		w.WriteHeader(http.StatusMovedPermanently)
	})
	return httptest.NewServer(handler)
}

// startLocalTCPServer listens on the loopback interface and accepts-then-closes
// connections, giving tcping a hermetic target.
func startLocalTCPServer(t *testing.T) (port int) {
	t.Helper()
	ln, err := stdnet.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to start local TCP server: %v", err)
	}
	t.Cleanup(func() { _ = ln.Close() })
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return // listener closed
			}
			_ = conn.Close()
		}
	}()
	return ln.Addr().(*stdnet.TCPAddr).Port
}

// closedTCPPort returns a loopback port that was just released, so dialing it
// fails immediately with a connection error.
func closedTCPPort(t *testing.T) (port int) {
	t.Helper()
	ln, err := stdnet.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to probe for a closed port: %v", err)
	}
	port = ln.Addr().(*stdnet.TCPAddr).Port
	_ = ln.Close()
	return port
}

func TestLoadModule_Ping(t *testing.T) {
	// create mock servers for testing
	serverOK := createMockServer(http.StatusOK)
	defer serverOK.Close()
	server301 := createRedirectMockServer("https://notgoingthere.invalid")
	defer server301.Close()
	server404 := createMockServer(http.StatusNotFound)
	defer server404.Close()
	server500 := createMockServer(http.StatusInternalServerError)
	defer server500.Close()

	// hermetic TCP targets: one accepting, one just released (closed)
	tcpPort := startLocalTCPServer(t)
	tcpPortClosed := closedTCPPort(t)

	isOnWindows := runtime.GOOS == "windows"
	tests := []struct {
		name        string
		script      string
		wantErr     string
		skipWindows bool
	}{
		// TCPing tests (IP-literal targets resolve without touching DNS)
		{
			name: `tcping: normal`,
			script: itn.HereDoc(`
				load('net', 'tcping')
				s = tcping('127.0.0.1', port=tcp_port, interval=0.1)
				print(s)
				assert.eq(s.total, 4)
				assert.true(s.success > 0)
			`),
		},
		{
			name: `tcping: abnormal`,
			script: itn.HereDoc(`
				load('net', 'tcping')
				s = tcping('127.0.0.1', port=tcp_port, count=1, timeout=-5, interval=-2)
				print(s)
				assert.eq(s.total, 1)
				assert.true(s.success > 0)
				assert.eq(s.stddev, 0)
			`),
		},
		{
			name: `tcping: faster`,
			script: itn.HereDoc(`
				load('net', 'tcping')
				s = tcping('127.0.0.1', port=tcp_port, count=10, timeout=5, interval=0.01)
				print(s)
				assert.eq(s.total, 10)
				assert.true(s.success > 0)
			`),
		},
		{
			name: `tcping: closed port`,
			script: itn.HereDoc(`
				load('net', 'tcping')
				s = tcping('127.0.0.1', port=tcp_port_closed, count=2, timeout=2, interval=0.01)
			`),
			wantErr: `net.tcping: no successful connections`,
		},
		{
			name: `tcping: not exists`,
			script: itn.HereDoc(`
				load('net', 'tcping')
				s = tcping('missing.invalid')
			`),
			wantErr: `missing.invalid`, // the error always names the host, online (NXDOMAIN) or offline
		},
		{
			name: `tcping: wrong count`,
			script: itn.HereDoc(`
				load('net', 'tcping')
				s = tcping('127.0.0.1', count=0)
			`),
			wantErr: `net.tcping: count must be greater than 0`,
		},
		{
			name: `tcping: no args`,
			script: itn.HereDoc(`
				load('net', 'tcping')
				tcping()
			`),
			wantErr: `net.tcping: missing argument for hostname`,
		},
		{
			name: `tcping: invalid args`,
			script: itn.HereDoc(`
				load('net', 'tcping')
				tcping(123)
			`),
			wantErr: `net.tcping: for parameter hostname: got int, want string or bytes`,
		},

		// HTTPing tests (loopback httptest targets)
		{
			name: `httping: normal`,
			script: itn.HereDoc(`
				load('net', 'httping')
				s = httping(server_ok, interval=0.1)
				print(s)
				assert.eq(s.total, 4)
				assert.true(s.success > 0)
				# loopback connects can finish within one Windows clock tick (~0.5ms)
				assert.true(s.min >= 0)
			`),
		},
		{
			name: `httping: abnormal`,
			script: itn.HereDoc(`
				load('net', 'httping')
				s = httping(server_ok, count=1, timeout=-5, interval=-2)
				print(s)
				assert.eq(s.total, 1)
				assert.true(s.success > 0)
				assert.eq(s.stddev, 0)
			`),
		},
		{
			name: `httping: faster`,
			script: itn.HereDoc(`
				load('net', 'httping')
				s = httping(server_ok, count=10, timeout=5, interval=0.01)
				print(s)
				assert.eq(s.total, 10)
				assert.true(s.success > 0)
				# loopback connects can finish within one Windows clock tick (~0.5ms)
				assert.true(s.min >= 0)
			`),
		},
		{
			name: `httping: ignore redirect`,
			script: itn.HereDoc(`
				load('net', 'httping')
				s = httping(server_301, interval=0.1)
				assert.eq(s.total, 4)
				assert.eq(s.success, 4)
			`),
		},
		{
			name: `httping: status 404`,
			script: itn.HereDoc(`
				load('net', 'httping')
				s = httping(server_404, interval=0.1)
			`),
			wantErr: `net.httping: no successful connections`,
		},
		{
			name: `httping: status 500`,
			script: itn.HereDoc(`
				load('net', 'httping')
				s = httping(server_500, interval=0.1)
			`),
			wantErr: `net.httping: no successful connections`,
		},
		{
			name: `httping: not exists`,
			script: itn.HereDoc(`
				load('net', 'httping')
				s = httping('http://missing.invalid')
			`),
			wantErr: `net.httping: no successful connections`,
		},
		{
			name: `httping: wrong count`,
			script: itn.HereDoc(`
				load('net', 'httping')
				s = httping(server_ok, count=0)
			`),
			wantErr: `net.httping: count must be greater than 0`,
		},
		{
			name: `httping: no args`,
			script: itn.HereDoc(`
				load('net', 'httping')
				httping()
			`),
			wantErr: `net.httping: missing argument for url`,
		},
		{
			name: `httping: invalid args`,
			script: itn.HereDoc(`
				load('net', 'httping')
				httping(123)
			`),
			wantErr: `net.httping: for parameter url: got int, want string or bytes`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if isOnWindows && tt.skipWindows {
				t.Skipf("Skip test on Windows")
				return
			}
			extra := starlark.StringDict{
				"server_ok":       starlark.String(serverOK.URL),
				"server_301":      starlark.String(server301.URL),
				"server_404":      starlark.String(server404.URL),
				"server_500":      starlark.String(server500.URL),
				"tcp_port":        starlark.MakeInt(tcpPort),
				"tcp_port_closed": starlark.MakeInt(tcpPortClosed),
			}
			res, err := itn.ExecModuleWithErrorTest(t, net.ModuleName, net.LoadModule, tt.script, tt.wantErr, extra)
			if (err != nil) != (tt.wantErr != "") {
				t.Errorf("net(%q) expects error = '%v', actual error = '%v', result = %v", tt.name, tt.wantErr, err, res)
				return
			}
		})
	}
}

func TestPingBounds(t *testing.T) {
	port := startLocalTCPServer(t)
	tests := []struct {
		name    string
		script  string
		wantErr string
	}{
		{
			name:    `tcping: count too large`,
			script:  "load('net', 'tcping')\ns = tcping('127.0.0.1', port=tcp_port, count=2000)",
			wantErr: `net.tcping: count must be at most 1024`,
		},
		{
			name:    `tcping: interval too large`,
			script:  "load('net', 'tcping')\ns = tcping('127.0.0.1', port=tcp_port, interval=7200)",
			wantErr: `net.tcping: interval must be at most 3600 seconds`,
		},
		{
			name:    `tcping: timeout too large`,
			script:  "load('net', 'tcping')\ns = tcping('127.0.0.1', port=tcp_port, timeout=1000000)",
			wantErr: `net.tcping: timeout must be at most 3600 seconds`,
		},
		{
			name:    `httping: count too large`,
			script:  "load('net', 'httping')\ns = httping('http://127.0.0.1/', count=2000)",
			wantErr: `net.httping: count must be at most 1024`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			extra := starlark.StringDict{"tcp_port": starlark.MakeInt(port)}
			res, err := itn.ExecModuleWithErrorTest(t, net.ModuleName, net.LoadModule, tt.script, tt.wantErr, extra)
			if (err != nil) != (tt.wantErr != "") {
				t.Errorf("net(%q) expects error = '%v', actual error = '%v', result = %v", tt.name, tt.wantErr, err, res)
			}
		})
	}
}

func TestTCPingSubSecondInterval(t *testing.T) {
	// the old time.Duration(f)*time.Second conversion truncated sub-second
	// values to 0, so this two-round ping paused for no time at all
	port := startLocalTCPServer(t)
	script := fmt.Sprintf("load('net', 'tcping')\ns = tcping('127.0.0.1', port=%d, count=2, interval=0.3)", port)
	ts := time.Now()
	if _, err := itn.ExecModuleWithErrorTest(t, net.ModuleName, net.LoadModule, script, "", nil); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if elapsed := time.Since(ts); elapsed < 250*time.Millisecond {
		t.Errorf("sub-second interval was truncated away, two rounds took only %v", elapsed)
	}
}

func TestTCPingCancellation(t *testing.T) {
	// the ping loop must honor the machine context: with an interruptible
	// pause the run aborts within the timeout instead of finishing all
	// rounds first
	port := startLocalTCPServer(t)
	m := starlet.NewWithNames(nil, nil, []string{"net"})
	m.SetScript("ping.star", []byte(fmt.Sprintf("load('net', 'tcping')\ns = tcping('127.0.0.1', port=%d, count=5, interval=1)", port)), nil)
	ts := time.Now()
	_, err := m.RunWithTimeout(time.Second, nil)
	if err == nil || !strings.Contains(err.Error(), "context") {
		t.Errorf("expected a context cancellation error, got: %v", err)
	}
	if elapsed := time.Since(ts); elapsed > 3*time.Second {
		t.Errorf("ping was not cancelled in time, took %v", elapsed)
	}
}
