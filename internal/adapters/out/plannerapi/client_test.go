package plannerapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	gen "github.com/BennettSmith/ebo-planner-cli/internal/gen/plannerapi"
	"github.com/BennettSmith/ebo-planner-cli/internal/platform/exitcode"
)

func TestAdapter_SendsAuthorizationHeader(t *testing.T) {
	seenAuth := ""
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"members":[]}`))
	}))
	defer srv.Close()

	a := Adapter{}
	_, err := a.ListMembers(context.Background(), srv.URL, "tok", nil)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if seenAuth != "Bearer tok" {
		t.Fatalf("auth: got %q", seenAuth)
	}
}

func TestAdapter_IdempotencyHeader_RequiredAndOptional(t *testing.T) {
	var createKey, cancelKey string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/trips":
			createKey = r.Header.Get("Idempotency-Key")
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"trip":{"tripId":"t1"}}`))
		case r.Method == http.MethodPost && r.URL.Path == "/trips/t1/cancel":
			cancelKey = r.Header.Get("Idempotency-Key")
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{}`))
		default:
			w.WriteHeader(500)
		}
	}))
	defer srv.Close()

	a := Adapter{}
	_, err := a.CreateTripDraft(context.Background(), srv.URL, "tok", "k1", gen.CreateTripDraftJSONRequestBody{Name: "n"})
	if err != nil {
		t.Fatalf("create err: %v", err)
	}
	if createKey != "k1" {
		t.Fatalf("create idempotency: got %q", createKey)
	}

	_, err = a.CancelTrip(context.Background(), srv.URL, "tok", gen.TripId("t1"), nil)
	if err != nil {
		t.Fatalf("cancel err: %v", err)
	}
	if cancelKey != "" {
		t.Fatalf("cancel idempotency: expected empty, got %q", cancelKey)
	}

	k := "k2"
	_, err = a.CancelTrip(context.Background(), srv.URL, "tok", gen.TripId("t1"), &k)
	if err != nil {
		t.Fatalf("cancel2 err: %v", err)
	}
	if cancelKey != "k2" {
		t.Fatalf("cancel idempotency: expected k2, got %q", cancelKey)
	}
}

func TestAdapter_ErrorDecoding_MapsExitKind(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(401)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]any{
				"code":      "UNAUTHORIZED",
				"message":   "nope",
				"requestId": "req-1",
			},
		})
	}))
	defer srv.Close()

	a := Adapter{}
	_, err := a.ListMembers(context.Background(), srv.URL, "tok", nil)
	if err == nil {
		t.Fatalf("expected error")
	}
	if exitcode.Code(err) != exitcode.Auth {
		t.Fatalf("expected auth exit code, got %d", exitcode.Code(err))
	}
}

func TestAdapter_CreateTripDraft_RequestErrorIsServer(t *testing.T) {
	// invalid base URL forces client.NewClientWithResponses to error
	a := Adapter{}
	_, err := a.CreateTripDraft(context.Background(), "://bad", "tok", "k1", gen.CreateTripDraftJSONRequestBody{Name: "n"})
	if err == nil {
		t.Fatalf("expected error")
	}
	if exitcode.Code(err) != exitcode.Server {
		t.Fatalf("expected server, got %d", exitcode.Code(err))
	}
}

func TestAdapter_CancelTrip_Maps404ToNotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(404)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]any{"code": "NOT_FOUND", "message": "missing"},
		})
	}))
	defer srv.Close()

	a := Adapter{}
	_, err := a.CancelTrip(context.Background(), srv.URL, "tok", gen.TripId("t1"), nil)
	if err == nil {
		t.Fatalf("expected error")
	}
	if exitcode.Code(err) != exitcode.NotFound {
		t.Fatalf("expected notfound exit 4, got %d", exitcode.Code(err))
	}
}

func TestAdapter_UpdateTrip_SendsIdempotencyHeader(t *testing.T) {
	seenKey := ""
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			w.WriteHeader(500)
			return
		}
		if r.URL.Path != "/trips/t1" {
			w.WriteHeader(500)
			return
		}
		seenKey = r.Header.Get("Idempotency-Key")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"trip":{"tripId":"t1"}}`))
	}))
	defer srv.Close()

	a := Adapter{}
	_, err := a.UpdateTrip(context.Background(), srv.URL, "tok", gen.TripId("t1"), "k1", gen.UpdateTripJSONRequestBody{})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if seenKey != "k1" {
		t.Fatalf("idempotency: got %q", seenKey)
	}
}

func TestAdapter_UpdateMyMemberProfile_SendsIdempotencyHeader(t *testing.T) {
	seenKey := ""
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			w.WriteHeader(500)
			return
		}
		if r.URL.Path != "/members/me" {
			w.WriteHeader(500)
			return
		}
		seenKey = r.Header.Get("Idempotency-Key")
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"member": map[string]any{
				"memberId":    "m1",
				"displayName": "n",
				"email":       "a@example.com",
			},
		})
	}))
	defer srv.Close()

	a := Adapter{}
	_, err := a.UpdateMyMemberProfile(context.Background(), srv.URL, "tok", "k1", gen.UpdateMyMemberProfileJSONRequestBody{})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if seenKey != "k1" {
		t.Fatalf("idempotency: got %q", seenKey)
	}
}
