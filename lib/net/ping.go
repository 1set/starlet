package net

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httptrace"
	"strconv"
	"sync"
	"time"

	"github.com/1set/starlet/dataconv"
	tps "github.com/1set/starlet/dataconv/types"
	"github.com/montanaflynn/stats"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

func goPingWrap(ctx context.Context, address string, count int, timeout, interval time.Duration, pingFunc func(ctx context.Context, address string, timeout time.Duration) (time.Duration, error)) ([]time.Duration, error) {
	if count <= 0 {
		return nil, fmt.Errorf("count must be greater than 0")
	}

	rttDurations := make([]time.Duration, 0, count)
	for i := 1; i <= count; i++ {
		rtt, err := pingFunc(ctx, address, timeout)
		if err != nil {
			continue
		}
		rttDurations = append(rttDurations, rtt)
		if i < count {
			time.Sleep(interval)
		}
	}

	if len(rttDurations) == 0 {
		return nil, fmt.Errorf("no successful connections")
	}
	return rttDurations, nil
}

func tcpPingFunc(ctx context.Context, address string, timeout time.Duration) (time.Duration, error) {
	start := time.Now()
	conn, err := net.DialTimeout("tcp", address, timeout)
	if err != nil {
		return 0, err
	}
	rtt := time.Since(start)
	conn.Close()
	return rtt, nil
}

func httpPingFunc(ctx context.Context, url string, timeout time.Duration) (time.Duration, error) {
	// create a custom http client tracing
	var (
		onceStart, onceDone sync.Once
		connStart           time.Time
		connDur             time.Duration
	)
	trace := &httptrace.ClientTrace{
		ConnectStart: func(network, addr string) {
			onceStart.Do(func() {
				connStart = time.Now()
			})
		},
		ConnectDone: func(network, addr string, err error) {
			onceDone.Do(func() {
				connDur = time.Since(connStart)
			})
		},
	}

	// create a http client with timeout and tracing
	client := &http.Client{
		Timeout: timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse // do not follow redirects
		},
		Transport: &http.Transport{
			DisableKeepAlives: true,
		},
	}
	req, err := http.NewRequestWithContext(httptrace.WithClientTrace(ctx, trace), "GET", url, nil)
	if err != nil {
		return 0, err
	}

	// perform the HTTP request
	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		return 0, fmt.Errorf("unacceptable status: %d", resp.StatusCode)
	}
	return connDur, nil
}

func createPingStats(address string, count int, rtts []time.Duration) starlark.Value {
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
		"address": starlark.String(address),
		"total":   starlark.MakeInt(count),
		"success": starlark.MakeInt(succ),
		"loss":    starlark.Float(loss),
		"min":     starlark.Float(min),
		"avg":     starlark.Float(avg),
		"max":     starlark.Float(max),
		"stddev":  starlark.Float(stddev),
	}
	return starlarkstruct.FromStringDict(starlark.String(`statistics`), sd)
}

func starTCPPing(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var (
		hostname tps.StringOrBytes
		port                    = 80
		count                   = 4
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

	// get the context for the DNS lookup and TCP ping
	ctx := dataconv.GetThreadContext(thread)

	// resolve the hostname to an IP address
	ips, err := goLookup(ctx, hostname.GoString(), "", time.Duration(timeout)*time.Second)
	if err != nil {
		return none, fmt.Errorf("%s: %w", b.Name(), err)
	}
	if len(ips) == 0 {
		return none, fmt.Errorf("unable to resolve hostname")
	}
	address := net.JoinHostPort(ips[0], strconv.Itoa(port))

	// perform the TCP ping, and get the statistics
	rtts, err := goPingWrap(ctx, address, count, time.Duration(timeout)*time.Second, time.Duration(interval)*time.Second, tcpPingFunc)
	if err != nil {
		return none, fmt.Errorf("%s: %w", b.Name(), err)
	}
	return createPingStats(address, count, rtts), nil
}

func starHTTPing(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var (
		url      tps.StringOrBytes
		count                   = 4
		timeout  tps.FloatOrInt = 10
		interval tps.FloatOrInt = 1
	)
	if err := starlark.UnpackArgs(b.Name(), args, kwargs, "url", &url, "count?", &count, "timeout?", &timeout, "interval?", &interval); err != nil {
		return nil, err
	}

	// correct timeout value
	if timeout <= 0 {
		timeout = 10
	}
	if interval <= 0 {
		interval = 1
	}

	// perform the HTTP ping, and get the statistics
	address := url.GoString()
	ctx := dataconv.GetThreadContext(thread)
	rtts, err := goPingWrap(ctx, address, count, time.Duration(timeout)*time.Second, time.Duration(interval)*time.Second, httpPingFunc)
	if err != nil {
		return none, fmt.Errorf("%s: %w", b.Name(), err)
	}
	return createPingStats(address, count, rtts), nil
}
