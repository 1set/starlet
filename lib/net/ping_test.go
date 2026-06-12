package net_test

import (
	stdnet "net"
	"net/http"
	"net/http/httptest"
	"runtime"
	"testing"

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
				assert.true(s.min > 0)
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
				assert.true(s.min > 0)
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
