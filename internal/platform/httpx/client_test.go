package httpx

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestTimeoutRoundTripper_CancelsRequest(t *testing.T) {
	// Server never responds.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
		w.WriteHeader(200)

	}))
	defer srv.Close()

	c := NewClient(nil, Options{Timeout: 20 * time.Millisecond})
	ctx := context.Background()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, srv.URL, nil)
	if err != nil {
		t.Fatalf("new req: %v", err)
	}

	_, err = c.Do(req)
	if err == nil {
		t.Fatalf("expected error")
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		// Depending on transport, may wrap; still should contain deadline exceeded.
		if !strings.Contains(err.Error(), "deadline") {
			t.Fatalf("expected deadline error, got %v", err)
		}
	}
}

func TestLoggingRoundTripper_RedactsAuthorizationAndQuery(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))
	defer srv.Close()

	buf := &bytes.Buffer{}
	c := NewClient(nil, Options{Verbose: true, LogSink: buf})
	req, err := http.NewRequest(http.MethodGet, srv.URL+"/?access_token=secret", nil)
	if err != nil {
		t.Fatalf("new req: %v", err)
	}
	req.Header.Set("Authorization", "Bearer secret")

	_, err = c.Do(req)
	if err != nil {
		t.Fatalf("do: %v", err)
	}

	log := buf.String()
	if !strings.Contains(log, "HTTP GET") {
		t.Fatalf("expected method in log, got %q", log)
	}
	if strings.Contains(log, "Bearer") || strings.Contains(log, "secret") {
		t.Fatalf("expected token redacted, got %q", log)
	}
	if strings.Contains(log, "access_token") {
		t.Fatalf("expected query redacted, got %q", log)
	}
	if !strings.Contains(log, "REDACTED") {
		t.Fatalf("expected REDACTED marker, got %q", log)
	}
}
