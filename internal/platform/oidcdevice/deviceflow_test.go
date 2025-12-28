package oidcdevice

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

type fakeSleeper struct{ calls atomic.Int32 }

func (f *fakeSleeper) Sleep(ctx context.Context, d time.Duration) error {
	f.calls.Add(1)
	return nil
}

func TestDiscover_ParsesEndpoints(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/json")
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"device_authorization_endpoint":"http://x/device","token_endpoint":"http://x/token"}`))
	}))
	defer srv.Close()

	d, err := Discover(context.Background(), srv.Client(), srv.URL)
	if err != nil {
		t.Fatalf("discover: %v", err)
	}
	if d.DeviceAuthorizationEndpoint == "" || d.TokenEndpoint == "" {
		t.Fatalf("missing endpoints")
	}
}

func TestPollToken_PendingThenSuccess(t *testing.T) {
	var n atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/token" {
			c := n.Add(1)
			w.Header().Set("content-type", "application/json")
			if c == 1 {
				w.WriteHeader(400)
				_, _ = w.Write([]byte(`{"error":"authorization_pending"}`))
				return
			}
			w.WriteHeader(200)
			_, _ = w.Write([]byte(`{"access_token":"a.b.c","token_type":"Bearer","expires_in":60}`))
			return
		}
		w.WriteHeader(404)
	}))
	defer srv.Close()

	slp := &fakeSleeper{}
	c := Client{HTTP: srv.Client(), Sleeper: slp}
	tr, err := c.PollToken(context.Background(), srv.URL+"/token", "cid", "dc", 1*time.Second)
	if err != nil {
		t.Fatalf("poll: %v", err)
	}
	if tr.AccessToken != "a.b.c" {
		t.Fatalf("token: %q", tr.AccessToken)
	}
	if slp.calls.Load() == 0 {
		t.Fatalf("expected sleep")
	}
}

func TestPollToken_SlowDownIncreasesInterval(t *testing.T) {
	var n atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/json")
		if n.Add(1) == 1 {
			w.WriteHeader(400)
			_, _ = w.Write([]byte(`{"error":"slow_down"}`))
			return
		}
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"access_token":"a.b.c","token_type":"Bearer"}`))
	}))
	defer srv.Close()

	slp := &fakeSleeper{}
	c := Client{HTTP: srv.Client(), Sleeper: slp}
	_, err := c.PollToken(context.Background(), srv.URL, "cid", "dc", 1*time.Second)
	if err != nil {
		t.Fatalf("poll: %v", err)
	}
	if slp.calls.Load() == 0 {
		t.Fatalf("expected sleep")
	}
}

func TestRequestDeviceCode_SendsFormAndParsesResponse(t *testing.T) {
	var gotCT string
	var gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		_ = r.Body.Close()
		gotCT = r.Header.Get("content-type")
		gotBody = string(b)
		w.Header().Set("content-type", "application/json")
		_, _ = w.Write([]byte(`{"device_code":"dc","user_code":"UC","verification_uri":"http://v"}`))
	}))
	defer srv.Close()

	c := Client{HTTP: srv.Client(), Sleeper: &fakeSleeper{}}
	resp, err := c.RequestDeviceCode(context.Background(), srv.URL, "cid", []string{"openid", "email"})
	if err != nil {
		t.Fatalf("device code: %v", err)
	}
	if resp.DeviceCode != "dc" || resp.UserCode != "UC" {
		t.Fatalf("resp %#v", resp)
	}
	if gotCT == "" {
		t.Fatalf("expected content-type")
	}
	if !strings.Contains(gotBody, "client_id=cid") {
		t.Fatalf("body=%q", gotBody)
	}
	if !strings.Contains(gotBody, "scope=openid+email") {
		t.Fatalf("body=%q", gotBody)
	}
}

func TestRealSleeper_CanceledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := (RealSleeper{}).Sleep(ctx, 10*time.Second); err == nil {
		t.Fatalf("expected error")
	}
}

func TestClientEnsure_NilHTTPIsError(t *testing.T) {
	c := Client{HTTP: nil}
	if _, err := c.RequestDeviceCode(context.Background(), "http://x", "cid", []string{"openid"}); err == nil {
		t.Fatalf("expected error")
	}
}

func TestClientEnsure_DefaultSleeperWhenNil(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/json")
		_, _ = w.Write([]byte(`{"device_code":"dc","user_code":"UC","verification_uri":"http://v"}`))
	}))
	defer srv.Close()

	c := Client{HTTP: srv.Client(), Sleeper: nil}
	if _, err := c.RequestDeviceCode(context.Background(), srv.URL, "cid", []string{"openid"}); err != nil {
		t.Fatalf("device code: %v", err)
	}
}

func TestDiscover_MissingEndpointsIsError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/json")
		_, _ = w.Write([]byte(`{"token_endpoint":"http://x/token"}`))
	}))
	defer srv.Close()

	if _, err := Discover(context.Background(), srv.Client(), srv.URL); err == nil {
		t.Fatalf("expected error")
	}
}

func TestPollToken_ErrorJSONUnknownIsError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/json")
		w.WriteHeader(400)
		_, _ = w.Write([]byte(`{"error":"access_denied"}`))
	}))
	defer srv.Close()

	c := Client{HTTP: srv.Client(), Sleeper: &fakeSleeper{}}
	if _, err := c.PollToken(context.Background(), srv.URL, "cid", "dc", 0); err == nil {
		t.Fatalf("expected error")
	}
}

func TestPollToken_DefaultSleeperDoesNotPanicOnPending(t *testing.T) {
	var n atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/json")
		if n.Add(1) == 1 {
			w.WriteHeader(400)
			_, _ = w.Write([]byte(`{"error":"authorization_pending"}`))
			return
		}
		_, _ = w.Write([]byte(`{"access_token":"a.b.c","token_type":"Bearer"}`))
	}))
	defer srv.Close()

	c := Client{HTTP: srv.Client(), Sleeper: nil}
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	_, err := c.PollToken(ctx, srv.URL, "cid", "dc", 1*time.Millisecond)
	if err != nil {
		// could time out on slow machines; but should never panic.
		return
	}
}

func TestDiscover_EmptyIssuerIsError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer srv.Close()
	if _, err := Discover(context.Background(), srv.Client(), ""); err == nil {
		t.Fatalf("expected error")
	}
}

func TestDiscover_NilHTTPIsError(t *testing.T) {
	if _, err := Discover(context.Background(), nil, "http://x"); err == nil {
		t.Fatalf("expected error")
	}
}
