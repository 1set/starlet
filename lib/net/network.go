// Package net provides network-related functions for Starlark, inspired by Go's net package and Python's socket module.
package net

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/montanaflynn/stats"

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

// goTCPPing performs a TCP ping to the given address and port. It returns a slice of round-trip times (RTTs) for each successful connection attempt and an error if any.
func goTCPPing(ctx context.Context, hostname string, port int, count int, timeout, interval time.Duration) ([]time.Duration, string, error) {
	if count <= 0 {
		return nil, "", fmt.Errorf("count must be greater than 0")
	}

	// resolve the hostname to an IP address
	ips, err := goLookup(ctx, hostname, "", timeout)
	if err != nil {
		return nil, "", err
	}
	if len(ips) == 0 {
		return nil, "", fmt.Errorf("unable to resolve hostname")
	}
	addr := net.JoinHostPort(ips[0], strconv.Itoa(port))

	// slice to hold the RTTs of successful pings
	var rttDurations []time.Duration
	for i := 1; i <= count; i++ {
		start := time.Now()
		conn, err := net.DialTimeout("tcp", addr, timeout)
		if err != nil {
			// if the connection fails, continue to the next attempt without adding RTT
			continue
		}
		rtt := time.Since(start)
		_ = conn.Close()

		// store the successful RTT, and wait for the interval before the next attempt
		rttDurations = append(rttDurations, rtt)
		if i < count {
			time.Sleep(interval)
		}
	}

	// if there were no successful connections, return an error.
	if len(rttDurations) == 0 {
		return nil, addr, fmt.Errorf("no successful connections")
	}

	// return the slice of RTTs.
	return rttDurations, addr, nil
}

// starTCPPing performs a TCP ping to the given address and port. It returns the average round-trip time (RTT) and the packet loss percentage.
// for statistics result like: 4 packets transmitted, 4 packets received, 0.0% packet loss, round-trip min/avg/max/stddev = 0.409/0.666/0.773/0.149 ms
func starTCPPing(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var (
		hostname tps.StringOrBytes
		port     int            = 80
		count    int            = 4
		timeout  tps.FloatOrInt = 10
		interval tps.FloatOrInt = 1
	)
	if err := starlark.UnpackArgs(b.Name(), args, kwargs, "hostname", &hostname, "port?", &port, "count?", &count, "timeout?", &timeout, "interval?", &interval); err != nil {
		return nil, err
	}

	// correct timeout value
	if timeout <= 0 {
		timeout = 10
	}
	if interval <= 0 {
		interval = 1
	}

	// get the context
	ctx := dataconv.GetThreadContext(thread)

	// perform the TCP ping
	rtts, addr, err := goTCPPing(ctx, hostname.GoString(), port, count, time.Duration(timeout)*time.Second, time.Duration(interval)*time.Second)

	// return the result
	if err != nil {
		return none, fmt.Errorf("%s: %w", b.Name(), err)
	}

	// statistics
	vals := make([]float64, len(rtts))
	for i, rtt := range rtts {
		vals[i] = float64(rtt) / float64(time.Millisecond)
	}
	succ := len(rtts)
	loss := float64(count-succ) / float64(count) * 100
	avg, _ := stats.Mean(vals)
	min, _ := stats.Min(vals)
	max, _ := stats.Max(vals)
	stddev, _ := stats.StandardDeviation(vals)
	sd := starlark.StringDict{
		"hostname": starlark.String(addr),
		"total":    starlark.MakeInt(count),
		"success":  starlark.MakeInt(succ),
		"loss":     starlark.Float(loss),
		"min":      starlark.Float(min),
		"avg":      starlark.Float(avg),
		"max":      starlark.Float(max),
		"stddev":   starlark.Float(stddev),
	}
	return starlarkstruct.FromStringDict(starlark.String(`statistics`), sd), nil
}
