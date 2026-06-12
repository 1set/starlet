//go:build integration

// Real-network smoke tests for lib/net, excluded from the default (hermetic)
// suite. They exercise the paths a local stub cannot reach: the system
// resolver, the default-DNS-port append branch, and real TCP/HTTPS targets.
//
// Run with: make test_integration
//
//	or: go test -tags integration ./lib/net/...
package net_test

import (
	"testing"

	itn "github.com/1set/starlet/internal"
	"github.com/1set/starlet/lib/net"
)

func TestIntegration_RealNetwork(t *testing.T) {
	tests := []struct {
		name    string
		script  string
		wantErr string
	}{
		{
			name: `nslookup: system resolver`,
			script: itn.HereDoc(`
				load('net', 'nslookup')
				ips = nslookup('bing.com')
				print(ips)
				assert.true(len(ips) > 0)
			`),
		},
		{
			name: `nslookup: custom dns without port`, // covers the default-:53 append branch
			script: itn.HereDoc(`
				load('net', 'nslookup')
				ips = nslookup('bing.com', '8.8.8.8')
				print(ips)
				assert.true(len(ips) > 0)
			`),
		},
		{
			name: `nslookup: custom dns with port`,
			script: itn.HereDoc(`
				load('net', 'nslookup')
				ips = nslookup('bing.com', '1.1.1.1:53')
				print(ips)
				assert.true(len(ips) > 0)
			`),
		},
		{
			name: `tcping: real host`,
			script: itn.HereDoc(`
				load('net', 'tcping')
				s = tcping('bing.com', count=2, interval=0.1)
				print(s)
				assert.eq(s.total, 2)
				assert.true(s.success > 0)
			`),
		},
		{
			name: `httping: real host`,
			script: itn.HereDoc(`
				load('net', 'httping')
				s = httping('https://www.bing.com', count=2, interval=0.1)
				print(s)
				assert.eq(s.total, 2)
				assert.true(s.success > 0)
				assert.true(s.min > 0)
			`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := itn.ExecModuleWithErrorTest(t, net.ModuleName, net.LoadModule, tt.script, tt.wantErr, nil)
			if (err != nil) != (tt.wantErr != "") {
				t.Errorf("net(%q) expects error = '%v', actual error = '%v', result = %v", tt.name, tt.wantErr, err, res)
				return
			}
		})
	}
}
