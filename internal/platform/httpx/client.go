package httpx

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Options struct {
	Timeout time.Duration
	Verbose bool
	LogSink io.Writer
}

// NewClient returns an http.Client configured for CLI runtime behavior.
//
// - Timeout is enforced via request context deadlines (not http.Client.Timeout).
// - Verbose logging writes method/url/status/timing to LogSink (stderr by convention).
// - Authorization headers are redacted in logs.
func NewClient(base *http.Client, opts Options) *http.Client {
	if base == nil {
		base = http.DefaultClient
	}

	transport := base.Transport
	if transport == nil {
		transport = http.DefaultTransport
	}

	rt := transport
	if opts.Verbose {
		if opts.LogSink == nil {
			opts.LogSink = io.Discard
		}
		rt = &loggingRoundTripper{base: rt, out: opts.LogSink}
	}
	if opts.Timeout > 0 {
		rt = &timeoutRoundTripper{base: rt, timeout: opts.Timeout}
	}

	c := *base
	c.Transport = rt
	return &c
}

type timeoutRoundTripper struct {
	base    http.RoundTripper
	timeout time.Duration
}

func (t *timeoutRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	ctx := req.Context()
	if _, ok := ctx.Deadline(); ok {
		return t.base.RoundTrip(req)
	}
	ctx, cancel := context.WithTimeout(ctx, t.timeout)
	defer cancel()
	return t.base.RoundTrip(req.WithContext(ctx))
}

type loggingRoundTripper struct {
	base http.RoundTripper
	out  io.Writer
}

func (l *loggingRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	start := time.Now()

	// Snapshot request for logging (redacted).
	method := req.Method
	urlStr := safeURL(req.URL)
	auth := req.Header.Get("Authorization")
	if auth != "" {
		auth = "REDACTED"
	}

	resp, err := l.base.RoundTrip(req)
	dur := time.Since(start)

	status := 0
	if resp != nil {
		status = resp.StatusCode
	}

	if err != nil {
		_, _ = fmt.Fprintf(l.out, "HTTP %s %s -> error (%s) auth=%s\n", method, urlStr, dur, auth)
		return resp, err
	}
	_, _ = fmt.Fprintf(l.out, "HTTP %s %s -> %d (%s) auth=%s\n", method, urlStr, status, dur, auth)
	return resp, nil
}

func safeURL(u *url.URL) string {
	if u == nil {
		return ""
	}
	// Ensure tokens never leak via query string.
	cu := *u
	if cu.RawQuery != "" {
		cu.RawQuery = "REDACTED"
	}
	return strings.TrimSpace(cu.String())
}
