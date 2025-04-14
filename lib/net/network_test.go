package net_test

import (
	"runtime"
	"testing"

	itn "github.com/1set/starlet/internal"
	"github.com/1set/starlet/lib/net"
)

func TestLoadModule_NSLookUp(t *testing.T) {
	isOnWindows := runtime.GOOS == "windows"
	tests := []struct {
		name        string
		script      string
		wantErr     string
		skipWindows bool
	}{
		{
			name: `nslookup: normal`,
			script: itn.HereDoc(`
				load('net', 'nslookup')
				ips = nslookup('bing.com')
				print(ips)
				assert.true(len(ips) > 0)
			`),
		},
		{
			name: `nslookup: normal with timeout`,
			script: itn.HereDoc(`
				load('net', 'nslookup')
				ips = nslookup('bing.com', timeout=5)
				print(ips)
				assert.true(len(ips) > 0)
			`),
		},
		{
			name: `nslookup: normal with dns`,
			script: itn.HereDoc(`
				load('net', 'nslookup')
				ips = nslookup('bing.com', '8.8.8.8')
				print(ips)
				assert.true(len(ips) > 0)
			`),
		},
		{
			name: `nslookup: normal with dns:port`,
			script: itn.HereDoc(`
				load('net', 'nslookup')
				ips = nslookup('bing.com', '1.1.1.1:53')
				print(ips)
				assert.true(len(ips) > 0)
			`),
		},
		{
			name: `nslookup: ip`,
			script: itn.HereDoc(`
				load('net', 'nslookup')
				ips = nslookup('8.8.8.8', timeout=-1)
				print(ips)
				assert.true(len(ips) > 0)
			`),
		},
		{
			name: `nslookup: localhost`,
			script: itn.HereDoc(`
				load('net', 'nslookup')
				ips = nslookup('localhost')
				print(ips)
				assert.true(len(ips) > 0)
			`),
		},
		{
			name: `nslookup: not exists`,
			script: itn.HereDoc(`
				load('net', 'nslookup')
				ips = nslookup('missing.invalid')
			`),
			wantErr: `missing.invalid`, // mac/win: no such host, linux: server misbehaving
		},
		{
			name: `nslookup: wrong dns`,
			script: itn.HereDoc(`
				load('net', 'nslookup')
				ips = nslookup('bing.com', 'microsoft.com', timeout=1)
			`),
			wantErr:     `timeout`, // Accept any error containing "timeout"
			skipWindows: true,      // on Windows 2022 with Go 1.18.10, it returns results from the default DNS server
		},
		{
			name: `nslookup: no args`,
			script: itn.HereDoc(`
				load('net', 'nslookup')
				nslookup()
			`),
			wantErr: `net.nslookup: missing argument for domain`,
		},
		{
			name: `nslookup: invalid args`,
			script: itn.HereDoc(`
				load('net', 'nslookup')
				nslookup(1, 2, 3)
			`),
			wantErr: `net.nslookup: for parameter domain: got int, want string or bytes`,
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
