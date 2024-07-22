package net_test

import (
	"runtime"
	"testing"

	itn "github.com/1set/starlet/internal"
	"github.com/1set/starlet/lib/net"
)

func TestLoadModule_Ping(t *testing.T) {
	isOnWindows := runtime.GOOS == "windows"
	tests := []struct {
		name        string
		script      string
		wantErr     string
		skipWindows bool
	}{
		// TCPing tests
		{
			name: `tcping: normal`,
			script: itn.HereDoc(`
				load('net', 'tcping')
				s = tcping('bing.com')
				print(s)
				assert.eq(s.total, 4)
				assert.true(s.success > 0)
			`),
		},
		{
			name: `tcping: abnormal`,
			script: itn.HereDoc(`
				load('net', 'tcping')
				s = tcping('bing.com', count=1, timeout=-5, interval=-2)
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
				s = tcping('bing.com', count=10, timeout=5, interval=0.01)
				print(s)
				assert.eq(s.total, 10)
				assert.true(s.success > 0)
			`),
		},
		{
			name: `tcping: not exists`,
			script: itn.HereDoc(`
				load('net', 'tcping')
				s = tcping('missing.invalid')
			`),
			wantErr: `missing.invalid`, // mac/win: no such host, linux: server misbehaving
		},
		{
			name: `tcping: wrong count`,
			script: itn.HereDoc(`
				load('net', 'tcping')
				s = tcping('bing.com', count=0)
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

		// HTTPing tests
		{
			name: `httping: normal`,
			script: itn.HereDoc(`
				load('net', 'httping')
				s = httping('https://www.bing.com')
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
				s = httping('https://www.bing.com', count=1, timeout=-5, interval=-2)
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
				s = httping('https://www.bing.com', count=10, timeout=5, interval=0.01)
				print(s)
				assert.eq(s.total, 10)
				assert.true(s.success > 0)
				assert.true(s.min > 0)
			`),
		},
		{
			name: `httping: not exists`,
			script: itn.HereDoc(`
				load('net', 'httping')
				s = httping('http://missing.invalid')
			`),
			wantErr: `net.httping: no successful connections`, // mac/win: no such host, linux: server misbehaving
		},
		{
			name: `httping: wrong count`,
			script: itn.HereDoc(`
				load('net', 'httping')
				s = httping('https://www.bing.com', count=0)
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
			res, err := itn.ExecModuleWithErrorTest(t, net.ModuleName, net.LoadModule, tt.script, tt.wantErr, nil)
			if (err != nil) != (tt.wantErr != "") {
				t.Errorf("net(%q) expects error = '%v', actual error = '%v', result = %v", tt.name, tt.wantErr, err, res)
				return
			}
		})
	}
}
