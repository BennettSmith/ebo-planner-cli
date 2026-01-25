package authloginapp

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Overland-East-Bay/trip-planner-cli/internal/platform/config"
	"github.com/Overland-East-Bay/trip-planner-cli/internal/platform/exitcode"
	"github.com/Overland-East-Bay/trip-planner-cli/internal/platform/oidcdevice"
)

type memStore struct{ doc config.Document }

func (m memStore) Path(ctx context.Context) (string, error)             { return "/x", nil }
func (m memStore) Load(ctx context.Context) (config.Document, error)    { return m.doc, nil }
func (m *memStore) Save(ctx context.Context, doc config.Document) error { m.doc = doc; return nil }

type fakeOpen struct{ last string }

func (f *fakeOpen) Open(url string) error { f.last = url; return nil }

type fixedClock struct{ t time.Time }

func (f fixedClock) Now() time.Time { return f.t }

func TestLogin_PersistsTokenFields(t *testing.T) {
	var baseURL string
	var tokenCalls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/json")
		switch r.URL.Path {
		case "/.well-known/openid-configuration":
			_, _ = w.Write([]byte(`{"device_authorization_endpoint":"` + baseURL + `/device","token_endpoint":"` + baseURL + `/token"}`))
		case "/device":
			_, _ = w.Write([]byte(`{"device_code":"dc","user_code":"UC","verification_uri":"` + baseURL + `/verify","expires_in":600,"interval":1}`))
		case "/token":
			tokenCalls++
			_, _ = w.Write([]byte(`{"access_token":"a.b.c","token_type":"Bearer","expires_in":60}`))
		default:
			w.WriteHeader(404)
		}
	}))
	defer srv.Close()
	baseURL = srv.URL

	doc := config.NewEmptyDocument()
	doc, _ = config.WithProfileAPIURL(doc, "default", "http://x")
	doc, _ = config.WithProfileOIDC(doc, "default", baseURL, "cid", []string{"openid"})

	m := &memStore{doc: doc}
	op := &fakeOpen{}
	clock := fixedClock{t: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)}

	svc := Service{Store: m, OIDC: oidcdevice.Client{HTTP: srv.Client()}, Open: op, Clock: clock}
	res, err := svc.Login(context.Background(), config.Effective{Profile: "default", APIURL: ""})
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	if res.Profile != "default" {
		t.Fatalf("profile: %q", res.Profile)
	}
	if op.last == "" {
		t.Fatalf("expected browser open")
	}

	// Verify persisted values.
	gotTok, _ := config.Get(m.doc, "profiles.default.auth.accessToken")
	if gotTok != "a.b.c" {
		t.Fatalf("token: %q", gotTok)
	}
	gotType, _ := config.Get(m.doc, "profiles.default.auth.tokenType")
	if gotType != "Bearer" {
		t.Fatalf("type: %q", gotType)
	}
	gotExp, _ := config.Get(m.doc, "profiles.default.auth.expiresAt")
	if gotExp == "" {
		t.Fatalf("expected expiresAt")
	}
	if tokenCalls == 0 {
		t.Fatalf("expected token call")
	}
}

func TestLogin_UsesRealClockWhenNil(t *testing.T) {
	var base string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/json")
		switch r.URL.Path {
		case "/.well-known/openid-configuration":
			_, _ = w.Write([]byte(`{"device_authorization_endpoint":"` + base + `/device","token_endpoint":"` + base + `/token"}`))
		case "/device":
			_, _ = w.Write([]byte(`{"device_code":"dc","user_code":"UC","verification_uri":"` + base + `/verify","expires_in":600,"interval":0}`))
		case "/token":
			_, _ = w.Write([]byte(`{"access_token":"a.b.c","token_type":"Bearer","expires_in":1}`))
		default:
			w.WriteHeader(404)
		}
	}))
	defer srv.Close()
	base = srv.URL

	doc := config.NewEmptyDocument()
	doc, _ = config.WithProfileAPIURL(doc, "default", "http://x")
	doc, _ = config.WithProfileOIDC(doc, "default", base, "cid", []string{"openid"})
	m := &memStore{doc: doc}
	op := &fakeOpen{}

	svc := Service{Store: m, OIDC: oidcdevice.Client{HTTP: srv.Client()}, Open: op, Clock: nil}
	_, err := svc.Login(context.Background(), config.Effective{Profile: "default"})
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	exp, _ := config.Get(m.doc, "profiles.default.auth.expiresAt")
	if exp == "" {
		t.Fatalf("expected expiresAt")
	}
}

func TestLogin_NilOpenerIsUnexpected(t *testing.T) {
	svc := Service{Store: &memStore{doc: config.NewEmptyDocument()}, OIDC: oidcdevice.Client{HTTP: &http.Client{}}, Open: nil}
	_, err := svc.Login(context.Background(), config.Effective{Profile: "default"})
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestLogin_NilStoreIsUnexpected(t *testing.T) {
	svc := Service{Store: nil, OIDC: oidcdevice.Client{HTTP: &http.Client{}}, Open: &fakeOpen{}}
	_, err := svc.Login(context.Background(), config.Effective{Profile: "default"})
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestLogin_MissingOIDCConfig_IsUsage(t *testing.T) {
	doc := config.NewEmptyDocument()
	doc, _ = config.WithProfileAPIURL(doc, "default", "http://x")

	m := &memStore{doc: doc}
	svc := Service{Store: m, OIDC: oidcdevice.Client{HTTP: &http.Client{}}, Open: &fakeOpen{}}
	_, err := svc.Login(context.Background(), config.Effective{Profile: "default"})
	if err == nil {
		t.Fatalf("expected error")
	}
	if exitcode.Code(err) != exitcode.Usage {
		t.Fatalf("expected usage exit 2, got %d", exitcode.Code(err))
	}
}
