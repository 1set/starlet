// Package net provides network-related functions for Starlark, inspired by Go's net package and Python's socket module.
package net

import (
	"context"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/1set/starlet/dataconv"
	tps "github.com/1set/starlet/dataconv/types"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

// ModuleName defines the expected name for this Module when used in starlark's load() function, eg: load('net', 'tcping')
const ModuleName = "net"

var (
	none    = starlark.None
	once    sync.Once
	modFunc starlark.StringDict
)

// LoadModule loads the net module. It is concurrency-safe and idempotent.
func LoadModule() (starlark.StringDict, error) {
	once.Do(func() {
		modFunc = starlark.StringDict{
			ModuleName: &starlarkstruct.Module{
				Name: ModuleName,
				Members: starlark.StringDict{
					"nslookup": starlark.NewBuiltin(ModuleName+".nslookup", starLookup),
					"tcping":   starlark.NewBuiltin(ModuleName+".tcping", starTCPPing),
					"httping":  starlark.NewBuiltin(ModuleName+".httping", starHTTPing),
				},
			},
		}
	})
	return modFunc, nil
}

func goLookup(ctx context.Context, domain, dnsServer string, timeout time.Duration) ([]string, error) {
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

	// Create a new context with timeout
	ctxWithTimeout, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// perform the DNS lookup
	return r.LookupHost(ctxWithTimeout, domain)
}

func starLookup(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var (
		domain    tps.StringOrBytes
		dnsServer tps.NullableStringOrBytes
		timeout   tps.FloatOrInt = 10
	)
	if err := starlark.UnpackArgs(b.Name(), args, kwargs, "domain", &domain, "dns_server?", &dnsServer, "timeout?", &timeout); err != nil {
		return nil, err
	}

	// correct timeout value
	if timeout <= 0 {
		timeout = 10
	}

	// get the context
	ctx := dataconv.GetThreadContext(thread)

	// perform the DNS lookup
	ips, err := goLookup(ctx, domain.GoString(), dnsServer.GoString(), time.Duration(timeout)*time.Second)

	// return the result
	if err != nil {
		return none, fmt.Errorf("%s: %w", b.Name(), err)
	}
	var list []starlark.Value
	for _, ip := range ips {
		list = append(list, starlark.String(ip))
	}
	return starlark.NewList(list), nil
}
