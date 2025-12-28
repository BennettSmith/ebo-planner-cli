package plannerapi

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
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

func TestAdapter_ListVisibleTripsForMember_HitsTripsEndpoint(t *testing.T) {
	seen := ""
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = r.Method + " " + r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"trips":[]}`))
	}))
	defer srv.Close()

	a := Adapter{}
	_, err := a.ListVisibleTripsForMember(context.Background(), srv.URL, "tok")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if seen != "GET /trips" {
		t.Fatalf("got %q", seen)
	}
}

func TestAdapter_ListMyDraftTrips_HitsDraftsEndpoint(t *testing.T) {
	seen := ""
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = r.Method + " " + r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"trips":[]}`))
	}))
	defer srv.Close()

	a := Adapter{}
	_, err := a.ListMyDraftTrips(context.Background(), srv.URL, "tok")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if seen != "GET /trips/drafts" {
		t.Fatalf("got %q", seen)
	}
}

func TestAdapter_GetTripDetails_HitsTripEndpoint(t *testing.T) {
	seen := ""
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = r.Method + " " + r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"trip":{"tripId":"t1","status":"PUBLISHED","rsvpActionsEnabled":false,"organizers":[],"artifacts":[]}}`))
	}))
	defer srv.Close()

	a := Adapter{}
	_, err := a.GetTripDetails(context.Background(), srv.URL, "tok", gen.TripId("t1"))
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if seen != "GET /trips/t1" {
		t.Fatalf("got %q", seen)
	}
}

func TestAdapter_GetTripDetails_Maps404ToNotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(404)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]any{"code": "NOT_FOUND", "message": "missing"},
		})
	}))
	defer srv.Close()

	a := Adapter{}
	_, err := a.GetTripDetails(context.Background(), srv.URL, "tok", gen.TripId("t1"))
	if err == nil {
		t.Fatalf("expected error")
	}
	if exitcode.Code(err) != exitcode.NotFound {
		t.Fatalf("expected notfound exit 4, got %d", exitcode.Code(err))
	}
}

func TestAdapter_SetTripDraftVisibility_SendsIdempotencyHeader(t *testing.T) {
	seenKey := ""
	seen := ""
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = r.Method + " " + r.URL.Path
		seenKey = r.Header.Get("Idempotency-Key")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"trip":{"tripId":"t1","status":"DRAFT","rsvpActionsEnabled":false,"organizers":[],"artifacts":[]}}`))
	}))
	defer srv.Close()

	a := Adapter{}
	_, err := a.SetTripDraftVisibility(context.Background(), srv.URL, "tok", gen.TripId("t1"), "k1", gen.SetDraftVisibilityRequest{DraftVisibility: "PUBLIC"})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if seen != "PUT /trips/t1/draft-visibility" {
		t.Fatalf("got %q", seen)
	}
	if seenKey != "k1" {
		t.Fatalf("idempotency: got %q", seenKey)
	}
}

func TestAdapter_PublishTrip_HitsPublishEndpoint(t *testing.T) {
	seen := ""
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = r.Method + " " + r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"announcementCopy":"A","trip":{"tripId":"t1","status":"PUBLISHED","rsvpActionsEnabled":false,"organizers":[],"artifacts":[]}}`))
	}))
	defer srv.Close()

	a := Adapter{}
	_, err := a.PublishTrip(context.Background(), srv.URL, "tok", gen.TripId("t1"))
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if seen != "POST /trips/t1/publish" {
		t.Fatalf("got %q", seen)
	}
}

func TestAdapter_PublishTrip_Maps404ToNotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(404)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]any{"code": "NOT_FOUND", "message": "missing"},
		})
	}))
	defer srv.Close()

	a := Adapter{}
	_, err := a.PublishTrip(context.Background(), srv.URL, "tok", gen.TripId("t1"))
	if err == nil {
		t.Fatalf("expected error")
	}
	if exitcode.Code(err) != exitcode.NotFound {
		t.Fatalf("expected notfound exit 4, got %d", exitcode.Code(err))
	}
}

func TestAdapter_PublishTrip_Maps401ToAuth(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(401)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]any{"code": "UNAUTHORIZED", "message": "nope"},
		})
	}))
	defer srv.Close()

	a := Adapter{}
	_, err := a.PublishTrip(context.Background(), srv.URL, "tok", gen.TripId("t1"))
	if err == nil {
		t.Fatalf("expected error")
	}
	if exitcode.Code(err) != exitcode.Auth {
		t.Fatalf("expected auth exit 3, got %d", exitcode.Code(err))
	}
}

func TestAdapter_SetTripDraftVisibility_Maps404ToNotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(404)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]any{"code": "NOT_FOUND", "message": "missing"},
		})
	}))
	defer srv.Close()

	a := Adapter{}
	_, err := a.SetTripDraftVisibility(context.Background(), srv.URL, "tok", gen.TripId("t1"), "k1", gen.SetDraftVisibilityRequest{DraftVisibility: "PUBLIC"})
	if err == nil {
		t.Fatalf("expected error")
	}
	if exitcode.Code(err) != exitcode.NotFound {
		t.Fatalf("expected notfound exit 4, got %d", exitcode.Code(err))
	}
}

func TestAdapter_AddTripOrganizer_HitsEndpointAndSendsIdempotencyHeader(t *testing.T) {
	seen := ""
	seenKey := ""
	seenBody := ""
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = r.Method + " " + r.URL.Path
		seenKey = r.Header.Get("Idempotency-Key")
		b, _ := io.ReadAll(r.Body)
		seenBody = string(b)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"trip":{"tripId":"t1","status":"DRAFT","rsvpActionsEnabled":false,"organizers":[],"artifacts":[]}}`))
	}))
	defer srv.Close()

	a := Adapter{}
	_, err := a.AddTripOrganizer(context.Background(), srv.URL, "tok", gen.TripId("t1"), "k1", gen.AddOrganizerRequest{MemberId: "m1"})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if seen != "POST /trips/t1/organizers" {
		t.Fatalf("got %q", seen)
	}
	if seenKey != "k1" {
		t.Fatalf("idempotency: got %q", seenKey)
	}
	if !strings.Contains(seenBody, `"memberId":"m1"`) {
		t.Fatalf("body: %q", seenBody)
	}
}

func TestAdapter_RemoveTripOrganizer_HitsEndpointAndSendsIdempotencyHeader(t *testing.T) {
	seen := ""
	seenKey := ""
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = r.Method + " " + r.URL.Path
		seenKey = r.Header.Get("Idempotency-Key")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"trip":{"tripId":"t1","status":"DRAFT","rsvpActionsEnabled":false,"organizers":[],"artifacts":[]}}`))
	}))
	defer srv.Close()

	a := Adapter{}
	_, err := a.RemoveTripOrganizer(context.Background(), srv.URL, "tok", gen.TripId("t1"), gen.MemberId("m1"), "k1")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if seen != "DELETE /trips/t1/organizers/m1" {
		t.Fatalf("got %q", seen)
	}
	if seenKey != "k1" {
		t.Fatalf("idempotency: got %q", seenKey)
	}
}

func TestAdapter_AddTripOrganizer_RequestErrorIsServer(t *testing.T) {
	a := Adapter{}
	_, err := a.AddTripOrganizer(context.Background(), "://bad", "tok", gen.TripId("t1"), "k1", gen.AddOrganizerRequest{MemberId: "m1"})
	if err == nil {
		t.Fatalf("expected error")
	}
	if exitcode.Code(err) != exitcode.Server {
		t.Fatalf("expected server, got %d", exitcode.Code(err))
	}
}

func TestAdapter_RemoveTripOrganizer_RequestErrorIsServer(t *testing.T) {
	a := Adapter{}
	_, err := a.RemoveTripOrganizer(context.Background(), "://bad", "tok", gen.TripId("t1"), gen.MemberId("m1"), "k1")
	if err == nil {
		t.Fatalf("expected error")
	}
	if exitcode.Code(err) != exitcode.Server {
		t.Fatalf("expected server, got %d", exitcode.Code(err))
	}
}

func TestAdapter_AddTripOrganizer_Maps401ToAuth(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(401)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]any{"code": "UNAUTHORIZED", "message": "nope"},
		})
	}))
	defer srv.Close()

	a := Adapter{}
	_, err := a.AddTripOrganizer(context.Background(), srv.URL, "tok", gen.TripId("t1"), "k1", gen.AddOrganizerRequest{MemberId: "m1"})
	if err == nil {
		t.Fatalf("expected error")
	}
	if exitcode.Code(err) != exitcode.Auth {
		t.Fatalf("expected auth exit 3, got %d", exitcode.Code(err))
	}
}

func TestAdapter_RemoveTripOrganizer_Maps404ToNotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(404)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]any{"code": "NOT_FOUND", "message": "missing"},
		})
	}))
	defer srv.Close()

	a := Adapter{}
	_, err := a.RemoveTripOrganizer(context.Background(), srv.URL, "tok", gen.TripId("t1"), gen.MemberId("m1"), "k1")
	if err == nil {
		t.Fatalf("expected error")
	}
	if exitcode.Code(err) != exitcode.NotFound {
		t.Fatalf("expected notfound exit 4, got %d", exitcode.Code(err))
	}
}

func TestAdapter_AddTripOrganizer_Maps404ToNotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(404)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]any{"code": "NOT_FOUND", "message": "missing"},
		})
	}))
	defer srv.Close()

	a := Adapter{}
	_, err := a.AddTripOrganizer(context.Background(), srv.URL, "tok", gen.TripId("t1"), "k1", gen.AddOrganizerRequest{MemberId: "m1"})
	if err == nil {
		t.Fatalf("expected error")
	}
	if exitcode.Code(err) != exitcode.NotFound {
		t.Fatalf("expected notfound exit 4, got %d", exitcode.Code(err))
	}
}

func TestAdapter_ListVisibleTripsForMember_Maps401ToAuth(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(401)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]any{"code": "UNAUTHORIZED", "message": "nope"},
		})
	}))
	defer srv.Close()

	a := Adapter{}
	_, err := a.ListVisibleTripsForMember(context.Background(), srv.URL, "tok")
	if err == nil {
		t.Fatalf("expected error")
	}
	if exitcode.Code(err) != exitcode.Auth {
		t.Fatalf("expected auth exit 3, got %d", exitcode.Code(err))
	}
}

func TestAdapter_ListMyDraftTrips_RequestErrorIsServer(t *testing.T) {
	a := Adapter{}
	_, err := a.ListMyDraftTrips(context.Background(), "://bad", "tok")
	if err == nil {
		t.Fatalf("expected error")
	}
	if exitcode.Code(err) != exitcode.Server {
		t.Fatalf("expected server, got %d", exitcode.Code(err))
	}
}

func TestAdapter_PublishTrip_RequestErrorIsServer(t *testing.T) {
	a := Adapter{}
	_, err := a.PublishTrip(context.Background(), "://bad", "tok", gen.TripId("t1"))
	if err == nil {
		t.Fatalf("expected error")
	}
	if exitcode.Code(err) != exitcode.Server {
		t.Fatalf("expected server, got %d", exitcode.Code(err))
	}
}

func TestAdapter_SetTripDraftVisibility_RequestErrorIsServer(t *testing.T) {
	a := Adapter{}
	_, err := a.SetTripDraftVisibility(context.Background(), "://bad", "tok", gen.TripId("t1"), "k1", gen.SetDraftVisibilityRequest{DraftVisibility: "PUBLIC"})
	if err == nil {
		t.Fatalf("expected error")
	}
	if exitcode.Code(err) != exitcode.Server {
		t.Fatalf("expected server, got %d", exitcode.Code(err))
	}
}

func TestAdapter_SetTripDraftVisibility_Maps401ToAuth(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(401)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]any{"code": "UNAUTHORIZED", "message": "nope"},
		})
	}))
	defer srv.Close()

	a := Adapter{}
	_, err := a.SetTripDraftVisibility(context.Background(), srv.URL, "tok", gen.TripId("t1"), "k1", gen.SetDraftVisibilityRequest{DraftVisibility: "PUBLIC"})
	if err == nil {
		t.Fatalf("expected error")
	}
	if exitcode.Code(err) != exitcode.Auth {
		t.Fatalf("expected auth exit 3, got %d", exitcode.Code(err))
	}
}
