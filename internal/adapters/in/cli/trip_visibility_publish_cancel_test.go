package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	gen "github.com/BennettSmith/ebo-planner-cli/internal/gen/plannerapi"
	"github.com/BennettSmith/ebo-planner-cli/internal/platform/exitcode"
	outplannerapi "github.com/BennettSmith/ebo-planner-cli/internal/ports/out/plannerapi"
)

type fakeTripMutationsAPI struct {
	visCalls     int
	publishCalls int
	cancelCalls  int

	lastVisIdem string
	lastVisReq  gen.SetDraftVisibilityRequest

	lastCancelIdem *string

	publishResp *gen.PublishTripResponse
}

var _ outplannerapi.Client = (*fakeTripMutationsAPI)(nil)

func (f *fakeTripMutationsAPI) ListVisibleTripsForMember(ctx context.Context, baseURL string, bearerToken string) (*gen.ListVisibleTripsForMemberClientResponse, error) {
	return nil, exitcode.New(exitcode.KindUnexpected, "not implemented in test", nil)
}
func (f *fakeTripMutationsAPI) ListMyDraftTrips(ctx context.Context, baseURL string, bearerToken string) (*gen.ListMyDraftTripsClientResponse, error) {
	return nil, exitcode.New(exitcode.KindUnexpected, "not implemented in test", nil)
}
func (f *fakeTripMutationsAPI) GetTripDetails(ctx context.Context, baseURL string, bearerToken string, tripID gen.TripId) (*gen.GetTripDetailsClientResponse, error) {
	return nil, exitcode.New(exitcode.KindUnexpected, "not implemented in test", nil)
}
func (f *fakeTripMutationsAPI) CreateTripDraft(ctx context.Context, baseURL string, bearerToken string, idempotencyKey string, req gen.CreateTripDraftJSONRequestBody) (*gen.CreateTripDraftClientResponse, error) {
	return nil, exitcode.New(exitcode.KindUnexpected, "not implemented in test", nil)
}
func (f *fakeTripMutationsAPI) UpdateTrip(ctx context.Context, baseURL string, bearerToken string, tripID gen.TripId, idempotencyKey string, req gen.UpdateTripJSONRequestBody) (*gen.UpdateTripClientResponse, error) {
	return nil, exitcode.New(exitcode.KindUnexpected, "not implemented in test", nil)
}
func (f *fakeTripMutationsAPI) ListMembers(ctx context.Context, baseURL string, bearerToken string, params *gen.ListMembersParams) (*gen.ListMembersClientResponse, error) {
	return nil, exitcode.New(exitcode.KindUnexpected, "not implemented in test", nil)
}
func (f *fakeTripMutationsAPI) UpdateMyMemberProfile(ctx context.Context, baseURL string, bearerToken string, idempotencyKey string, req gen.UpdateMyMemberProfileJSONRequestBody) (*gen.UpdateMyMemberProfileClientResponse, error) {
	return nil, exitcode.New(exitcode.KindUnexpected, "not implemented in test", nil)
}

func (f *fakeTripMutationsAPI) SetTripDraftVisibility(ctx context.Context, baseURL string, bearerToken string, tripID gen.TripId, idempotencyKey string, req gen.SetTripDraftVisibilityJSONRequestBody) (*gen.SetTripDraftVisibilityClientResponse, error) {
	_ = ctx
	_ = baseURL
	_ = bearerToken
	_ = tripID
	f.visCalls++
	f.lastVisIdem = idempotencyKey
	f.lastVisReq = req
	return &gen.SetTripDraftVisibilityClientResponse{JSON200: &gen.TripResponse{}}, nil
}

func (f *fakeTripMutationsAPI) PublishTrip(ctx context.Context, baseURL string, bearerToken string, tripID gen.TripId) (*gen.PublishTripClientResponse, error) {
	_ = ctx
	_ = baseURL
	_ = bearerToken
	_ = tripID
	f.publishCalls++
	resp := f.publishResp
	if resp == nil {
		resp = &gen.PublishTripResponse{AnnouncementCopy: "ANNOUNCE\n"}
	}
	return &gen.PublishTripClientResponse{JSON200: resp}, nil
}

func (f *fakeTripMutationsAPI) CancelTrip(ctx context.Context, baseURL string, bearerToken string, tripID gen.TripId, idempotencyKey *string) (*gen.CancelTripClientResponse, error) {
	_ = ctx
	_ = baseURL
	_ = bearerToken
	_ = tripID
	f.cancelCalls++
	f.lastCancelIdem = idempotencyKey
	return &gen.CancelTripClientResponse{JSON200: &gen.TripResponse{}}, nil
}

func (f *fakeTripMutationsAPI) AddTripOrganizer(ctx context.Context, baseURL string, bearerToken string, tripID gen.TripId, idempotencyKey string, req gen.AddTripOrganizerJSONRequestBody) (*gen.AddTripOrganizerClientResponse, error) {
	_ = ctx
	_ = baseURL
	_ = bearerToken
	_ = tripID
	_ = idempotencyKey
	_ = req
	return nil, exitcode.New(exitcode.KindUnexpected, "not implemented in test", nil)
}

func (f *fakeTripMutationsAPI) RemoveTripOrganizer(ctx context.Context, baseURL string, bearerToken string, tripID gen.TripId, memberID gen.MemberId, idempotencyKey string) (*gen.RemoveTripOrganizerClientResponse, error) {
	_ = ctx
	_ = baseURL
	_ = bearerToken
	_ = tripID
	_ = memberID
	_ = idempotencyKey
	return nil, exitcode.New(exitcode.KindUnexpected, "not implemented in test", nil)
}

func TestTripPublish_PrintAnnouncement_StdoutOnly(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	store := &memStore{path: "/x", doc: baseDoc(t)}
	api := &fakeTripMutationsAPI{publishResp: &gen.PublishTripResponse{AnnouncementCopy: "HELLO\n"}}

	cmd := NewRootCmd(RootDeps{ConfigStore: store, PlannerAPI: api, Stdout: stdout, Stderr: stderr})
	cmd.SetArgs([]string{"trip", "publish", "t1", "--print-announcement"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if stdout.String() != "HELLO\n" {
		t.Fatalf("stdout: %q", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr should be empty, got %q", stderr.String())
	}
}

func TestTripCancel_RequiresForce(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	store := &memStore{path: "/x", doc: baseDoc(t)}
	api := &fakeTripMutationsAPI{}

	cmd := NewRootCmd(RootDeps{ConfigStore: store, PlannerAPI: api, Stdout: stdout, Stderr: stderr})
	cmd.SetArgs([]string{"trip", "cancel", "t1"})
	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected error")
	}
	if exitcode.Code(err) != exitcode.Usage {
		t.Fatalf("expected exit 2, got %d (%v)", exitcode.Code(err), err)
	}
	if api.cancelCalls != 0 {
		t.Fatalf("expected no api calls, got %d", api.cancelCalls)
	}
}

func TestTripCancel_IdempotencyOptional_NoAutoGenerate(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	store := &memStore{path: "/x", doc: baseDoc(t)}
	api := &fakeTripMutationsAPI{}

	cmd := NewRootCmd(RootDeps{ConfigStore: store, PlannerAPI: api, Stdout: stdout, Stderr: stderr})
	cmd.SetArgs([]string{"trip", "cancel", "t1", "--force"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if api.cancelCalls != 1 {
		t.Fatalf("expected 1 call, got %d", api.cancelCalls)
	}
	if api.lastCancelIdem != nil {
		t.Fatalf("expected nil idempotency, got %v", *api.lastCancelIdem)
	}
}

func TestTripVisibility_JSON_IncludesMetaIdempotencyKey(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	store := &memStore{path: "/x", doc: baseDoc(t)}
	api := &fakeTripMutationsAPI{}

	cmd := NewRootCmd(RootDeps{ConfigStore: store, PlannerAPI: api, Stdout: stdout, Stderr: stderr})
	cmd.SetArgs([]string{"--output", "json", "trip", "visibility", "t1", "--public"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if api.visCalls != 1 {
		t.Fatalf("expected 1 call, got %d", api.visCalls)
	}
	if api.lastVisIdem == "" {
		t.Fatalf("expected generated idempotency key")
	}

	var env map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &env); err != nil {
		t.Fatalf("stdout not json: %v\n%s", err, stdout.String())
	}
	meta, _ := env["meta"].(map[string]any)
	if meta == nil || meta["idempotencyKey"] == "" {
		t.Fatalf("expected meta.idempotencyKey, got %#v", meta)
	}
}

func TestTripVisibility_PublicPrivateMutualExclusion(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	store := &memStore{path: "/x", doc: baseDoc(t)}
	api := &fakeTripMutationsAPI{}

	cmd := NewRootCmd(RootDeps{ConfigStore: store, PlannerAPI: api, Stdout: stdout, Stderr: stderr})
	cmd.SetArgs([]string{"trip", "visibility", "t1", "--public", "--private"})
	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected error")
	}
	if exitcode.Code(err) != exitcode.Usage {
		t.Fatalf("expected exit 2, got %d (%v)", exitcode.Code(err), err)
	}
	if api.visCalls != 0 {
		t.Fatalf("expected no api calls, got %d", api.visCalls)
	}
}

func TestTripPublish_JSONOutput_Envelope(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	store := &memStore{path: "/x", doc: baseDoc(t)}
	api := &fakeTripMutationsAPI{publishResp: &gen.PublishTripResponse{AnnouncementCopy: "HELLO\n"}}

	cmd := NewRootCmd(RootDeps{ConfigStore: store, PlannerAPI: api, Stdout: stdout, Stderr: stderr})
	cmd.SetArgs([]string{"--output", "json", "trip", "publish", "t1"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if api.publishCalls != 1 {
		t.Fatalf("expected 1 api call, got %d", api.publishCalls)
	}

	var env map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &env); err != nil {
		t.Fatalf("stdout not json: %v\n%s", err, stdout.String())
	}
	if env["data"] == nil {
		t.Fatalf("expected data")
	}
}

func TestTripCancel_WithIdempotency_JSONMetaAndPointer(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	store := &memStore{path: "/x", doc: baseDoc(t)}
	api := &fakeTripMutationsAPI{}

	cmd := NewRootCmd(RootDeps{ConfigStore: store, PlannerAPI: api, Stdout: stdout, Stderr: stderr})
	cmd.SetArgs([]string{"--output", "json", "trip", "cancel", "t1", "--force", "--idempotency-key", "k1"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if api.cancelCalls != 1 {
		t.Fatalf("expected 1 call, got %d", api.cancelCalls)
	}
	if api.lastCancelIdem == nil || *api.lastCancelIdem != "k1" {
		if api.lastCancelIdem == nil {
			t.Fatalf("expected idempotency pointer")
		}
		t.Fatalf("idempotency: %q", *api.lastCancelIdem)
	}

	var env map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &env); err != nil {
		t.Fatalf("stdout not json: %v\n%s", err, stdout.String())
	}
	meta, _ := env["meta"].(map[string]any)
	if meta == nil || meta["idempotencyKey"] != "k1" {
		t.Fatalf("meta: %#v", meta)
	}
}
