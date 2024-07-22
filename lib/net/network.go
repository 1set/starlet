// Package net provides network-related functions for Starlark, inspired by Go's net package and Python's socket module.
package net

import (
	"context"
	"net"
	"strings"
	"sync"
	"time"

	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

// ModuleName defines the expected name for this Module when used in starlark's load() function, eg: load('net', 'tcping')
const ModuleName = "net"

var (
	once    sync.Once
	modFunc starlark.StringDict
)

// LoadModule loads the net module. It is concurrency-safe and idempotent.
func LoadModule() (starlark.StringDict, error) {
	once.Do(func() {
		modFunc = starlark.StringDict{
			ModuleName: &starlarkstruct.Module{
				Name:    ModuleName,
				Members: starlark.StringDict{
					//"tcping": starlark.NewBuiltin("net.tcping", tcping),
					//"nslookup": starlark.NewBuiltin("net.nslookup", nslookup),
				},
			},
		}
	})
	return modFunc, nil
}

func nsLookup(ctx context.Context, domain, dnsServer string, timeout time.Duration) ([]string, error) {
	// create a custom resolver if a DNS server is specified
	var r *net.Resolver
	if dnsServer != "" {
		if !strings.Contains(dnsServer, ":") {
			// append default DNS port if not specified
			dnsServer = net.JoinHostPort(dnsServer, "53")
		}
		r = &net.Resolver{
			PreferGo: true,
			Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
				d := net.Dialer{
					Timeout: timeout,
				}
				return d.DialContext(ctx, "udp", dnsServer)
			},
		}
	} else {
		r = net.DefaultResolver
	}

	// perform the DNS lookup
	return r.LookupHost(ctx, domain)
}
