package net_test

import (
	"testing"

	itn "github.com/1set/starlet/internal"
	"github.com/1set/starlet/lib/net"
)

func TestLoadModule_Network(t *testing.T) {
	tests := []struct {
		name    string
		script  string
		wantErr string
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
			name: `nslookup: ip`,
			script: itn.HereDoc(`
				load('net', 'nslookup')
				ips = nslookup('8.8.8.8')
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
			wantErr: `no such host`,
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
