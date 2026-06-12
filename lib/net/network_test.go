package net_test

import (
	stdnet "net"
	"runtime"
	"testing"

	itn "github.com/1set/starlet/internal"
	"github.com/1set/starlet/lib/net"
	"go.starlark.net/starlark"
)

// startDNSStub runs a minimal UDP DNS server on 127.0.0.1 and returns its
// host:port. It answers every A question with 127.0.0.1 and any other
// question with an empty NOERROR response. With blackhole=true it reads
// queries and never replies, so clients hit a deterministic timeout.
//
// NOTE: a custom dns_server only takes effect where the pure-Go resolver
// honours a custom Dial — on Windows PreferGo is a no-op before Go 1.19
// (the system resolver is used instead), so stub-based cases must skip
// Windows while the module floor is go 1.18.
func startDNSStub(t *testing.T, blackhole bool) string {
	t.Helper()
	pc, err := stdnet.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to start DNS stub: %v", err)
	}
	t.Cleanup(func() { _ = pc.Close() })
	go func() {
		buf := make([]byte, 512)
		for {
			n, addr, err := pc.ReadFrom(buf)
			if err != nil {
				return // listener closed
			}
			if blackhole || n < 12 {
				continue
			}
			q := buf[:n]
			// skip the QNAME labels of the first question
			i := 12
			for i < len(q) && q[i] != 0 {
				i += int(q[i]) + 1
			}
			qend := i + 5 // NUL + QTYPE(2) + QCLASS(2)
			if qend > len(q) {
				continue
			}
			isA := q[i+1] == 0 && q[i+2] == 1
			ancount := byte(0)
			if isA {
				ancount = 1
			}
			resp := make([]byte, 0, qend+16)
			resp = append(resp, q[0], q[1], 0x81, 0x80)       // ID, QR|RD|RA, NOERROR
			resp = append(resp, 0, 1, 0, ancount, 0, 0, 0, 0) // QD=1, AN, NS=0, AR=0
			resp = append(resp, q[12:qend]...)                // echo the question
			if isA {
				resp = append(resp,
					0xC0, 0x0C, // NAME: pointer to the question name
					0, 1, 0, 1, // TYPE A, CLASS IN
					0, 0, 0, 60, // TTL 60s
					0, 4, 127, 0, 0, 1, // RDLENGTH 4, RDATA 127.0.0.1
				)
			}
			_, _ = pc.WriteTo(resp, addr)
		}
	}()
	return pc.LocalAddr().String()
}

func TestLoadModule_NSLookUp(t *testing.T) {
	// hermetic local DNS servers: one answering, one swallowing queries
	dnsStub := startDNSStub(t, false)
	dnsBlackhole := startDNSStub(t, true)

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
				ips = nslookup('hermetic.test', dns_stub)
				print(ips)
				assert.true(len(ips) > 0)
				assert.true('127.0.0.1' in ips)
			`),
			skipWindows: true, // PreferGo resolver is a no-op on Windows before Go 1.19
		},
		{
			name: `nslookup: normal with timeout`,
			script: itn.HereDoc(`
				load('net', 'nslookup')
				ips = nslookup('hermetic.test', dns_stub, timeout=5)
				print(ips)
				assert.true(len(ips) > 0)
			`),
			skipWindows: true, // PreferGo resolver is a no-op on Windows before Go 1.19
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
			wantErr: `missing.invalid`, // the error always names the domain, online (NXDOMAIN) or offline
		},
		{
			name: `nslookup: wrong dns`,
			script: itn.HereDoc(`
				load('net', 'nslookup')
				ips = nslookup('hermetic.test', dns_blackhole, timeout=1)
			`),
			wantErr:     `timeout`,
			skipWindows: true, // PreferGo resolver is a no-op on Windows before Go 1.19
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
			extra := starlark.StringDict{
				"dns_stub":      starlark.String(dnsStub),
				"dns_blackhole": starlark.String(dnsBlackhole),
			}
			res, err := itn.ExecModuleWithErrorTest(t, net.ModuleName, net.LoadModule, tt.script, tt.wantErr, extra)
			if (err != nil) != (tt.wantErr != "") {
				t.Errorf("net(%q) expects error = '%v', actual error = '%v', result = %v", tt.name, tt.wantErr, err, res)
				return
			}
		})
	}
}
