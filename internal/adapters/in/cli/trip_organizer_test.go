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

type fakeTripOrganizerAPI struct {
	addCalls    int
	removeCalls int

	lastAddIdem    string
	lastRemoveIdem string

	lastAddTripID    gen.TripId
	lastRemoveTripID gen.TripId
	lastAddReq       gen.AddTripOrganizerJSONRequestBody
	lastRemoveMember gen.MemberId
}

var _ outplannerapi.Client = (*fakeTripOrganizerAPI)(nil)

func (f *fakeTripOrganizerAPI) ListVisibleTripsForMember(ctx context.Context, baseURL string, bearerToken string) (*gen.ListVisibleTripsForMemberClientResponse, error) {
	return nil, exitcode.New(exitcode.KindUnexpected, "not implemented in test", nil)
}
func (f *fakeTripOrganizerAPI) ListMyDraftTrips(ctx context.Context, baseURL string, bearerToken string) (*gen.ListMyDraftTripsClientResponse, error) {
	return nil, exitcode.New(exitcode.KindUnexpected, "not implemented in test", nil)
}
func (f *fakeTripOrganizerAPI) GetTripDetails(ctx context.Context, baseURL string, bearerToken string, tripID gen.TripId) (*gen.GetTripDetailsClientResponse, error) {
	return nil, exitcode.New(exitcode.KindUnexpected, "not implemented in test", nil)
}
func (f *fakeTripOrganizerAPI) CreateTripDraft(ctx context.Context, baseURL string, bearerToken string, idempotencyKey string, req gen.CreateTripDraftJSONRequestBody) (*gen.CreateTripDraftClientResponse, error) {
	return nil, exitcode.New(exitcode.KindUnexpected, "not implemented in test", nil)
}
func (f *fakeTripOrganizerAPI) UpdateTrip(ctx context.Context, baseURL string, bearerToken string, tripID gen.TripId, idempotencyKey string, req gen.UpdateTripJSONRequestBody) (*gen.UpdateTripClientResponse, error) {
	return nil, exitcode.New(exitcode.KindUnexpected, "not implemented in test", nil)
}
func (f *fakeTripOrganizerAPI) SetTripDraftVisibility(ctx context.Context, baseURL string, bearerToken string, tripID gen.TripId, idempotencyKey string, req gen.SetTripDraftVisibilityJSONRequestBody) (*gen.SetTripDraftVisibilityClientResponse, error) {
	return nil, exitcode.New(exitcode.KindUnexpected, "not implemented in test", nil)
}
func (f *fakeTripOrganizerAPI) PublishTrip(ctx context.Context, baseURL string, bearerToken string, tripID gen.TripId) (*gen.PublishTripClientResponse, error) {
	return nil, exitcode.New(exitcode.KindUnexpected, "not implemented in test", nil)
}
func (f *fakeTripOrganizerAPI) CancelTrip(ctx context.Context, baseURL string, bearerToken string, tripID gen.TripId, idempotencyKey *string) (*gen.CancelTripClientResponse, error) {
	return nil, exitcode.New(exitcode.KindUnexpected, "not implemented in test", nil)
}
func (f *fakeTripOrganizerAPI) ListMembers(ctx context.Context, baseURL string, bearerToken string, params *gen.ListMembersParams) (*gen.ListMembersClientResponse, error) {
	return nil, exitcode.New(exitcode.KindUnexpected, "not implemented in test", nil)
}
func (f *fakeTripOrganizerAPI) UpdateMyMemberProfile(ctx context.Context, baseURL string, bearerToken string, idempotencyKey string, req gen.UpdateMyMemberProfileJSONRequestBody) (*gen.UpdateMyMemberProfileClientResponse, error) {
	return nil, exitcode.New(exitcode.KindUnexpected, "not implemented in test", nil)
}

func (f *fakeTripOrganizerAPI) AddTripOrganizer(ctx context.Context, baseURL string, bearerToken string, tripID gen.TripId, idempotencyKey string, req gen.AddTripOrganizerJSONRequestBody) (*gen.AddTripOrganizerClientResponse, error) {
	_ = ctx
	_ = baseURL
	_ = bearerToken
	f.addCalls++
	f.lastAddTripID = tripID
	f.lastAddIdem = idempotencyKey
	f.lastAddReq = req
	return &gen.AddTripOrganizerClientResponse{JSON200: &gen.TripResponse{}}, nil
}

func (f *fakeTripOrganizerAPI) RemoveTripOrganizer(ctx context.Context, baseURL string, bearerToken string, tripID gen.TripId, memberID gen.MemberId, idempotencyKey string) (*gen.RemoveTripOrganizerClientResponse, error) {
	_ = ctx
	_ = baseURL
	_ = bearerToken
	f.removeCalls++
	f.lastRemoveTripID = tripID
	f.lastRemoveMember = memberID
	f.lastRemoveIdem = idempotencyKey
	return &gen.RemoveTripOrganizerClientResponse{JSON200: &gen.TripResponse{}}, nil
}

func (f *fakeTripOrganizerAPI) SetMyRSVP(ctx context.Context, baseURL string, bearerToken string, tripID gen.TripId, idempotencyKey string, req gen.SetMyRSVPJSONRequestBody) (*gen.SetMyRSVPClientResponse, error) {
	_ = ctx
	_ = baseURL
	_ = bearerToken
	_ = tripID
	_ = idempotencyKey
	_ = req
	return nil, exitcode.New(exitcode.KindUnexpected, "not implemented in test", nil)
}

func (f *fakeTripOrganizerAPI) GetMyRSVPForTrip(ctx context.Context, baseURL string, bearerToken string, tripID gen.TripId) (*gen.GetMyRSVPForTripClientResponse, error) {
	_ = ctx
	_ = baseURL
	_ = bearerToken
	_ = tripID
	return nil, exitcode.New(exitcode.KindUnexpected, "not implemented in test", nil)
}

func (f *fakeTripOrganizerAPI) GetTripRSVPSummary(ctx context.Context, baseURL string, bearerToken string, tripID gen.TripId) (*gen.GetTripRSVPSummaryClientResponse, error) {
	_ = ctx
	_ = baseURL
	_ = bearerToken
	_ = tripID
	return nil, exitcode.New(exitcode.KindUnexpected, "not implemented in test", nil)
}

func TestTripOrganizerAdd_AutoGeneratesIdempotency_JSONMetaAndCall(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	store := &memStore{path: "/x", doc: baseDoc(t)}
	api := &fakeTripOrganizerAPI{}

	cmd := NewRootCmd(RootDeps{ConfigStore: store, PlannerAPI: api, Stdout: stdout, Stderr: stderr})
	cmd.SetArgs([]string{"--output", "json", "trip", "organizer", "add", "t1", "--member", "m1"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if api.addCalls != 1 {
		t.Fatalf("expected 1 call, got %d", api.addCalls)
	}
	if api.lastAddTripID != "t1" || api.lastAddReq.MemberId != "m1" {
		t.Fatalf("call args: trip=%q member=%q", api.lastAddTripID, api.lastAddReq.MemberId)
	}
	if api.lastAddIdem == "" {
		t.Fatalf("expected generated idempotency key")
	}

	var env map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &env); err != nil {
		t.Fatalf("stdout not json: %v\n%s", err, stdout.String())
	}
	meta, _ := env["meta"].(map[string]any)
	if meta == nil || meta["idempotencyKey"] == "" {
		t.Fatalf("meta: %#v", meta)
	}
}

func TestTripOrganizerAdd_MissingMember_IsUsage_NoAPICall(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	store := &memStore{path: "/x", doc: baseDoc(t)}
	api := &fakeTripOrganizerAPI{}

	cmd := NewRootCmd(RootDeps{ConfigStore: store, PlannerAPI: api, Stdout: stdout, Stderr: stderr})
	cmd.SetArgs([]string{"trip", "organizer", "add", "t1"})
	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected error")
	}
	if exitcode.Code(err) != exitcode.Usage {
		t.Fatalf("expected exit 2, got %d (%v)", exitcode.Code(err), err)
	}
	if api.addCalls != 0 {
		t.Fatalf("expected no api calls, got %d", api.addCalls)
	}
}

func TestTripOrganizerRemove_RequiresForce(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	store := &memStore{path: "/x", doc: baseDoc(t)}
	api := &fakeTripOrganizerAPI{}

	cmd := NewRootCmd(RootDeps{ConfigStore: store, PlannerAPI: api, Stdout: stdout, Stderr: stderr})
	cmd.SetArgs([]string{"trip", "organizer", "remove", "t1", "--member", "m1"})
	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected error")
	}
	if exitcode.Code(err) != exitcode.Usage {
		t.Fatalf("expected exit 2, got %d (%v)", exitcode.Code(err), err)
	}
	if api.removeCalls != 0 {
		t.Fatalf("expected no api calls, got %d", api.removeCalls)
	}
}

func TestTripOrganizerRemove_AutoGeneratesIdempotency(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	store := &memStore{path: "/x", doc: baseDoc(t)}
	api := &fakeTripOrganizerAPI{}

	cmd := NewRootCmd(RootDeps{ConfigStore: store, PlannerAPI: api, Stdout: stdout, Stderr: stderr})
	cmd.SetArgs([]string{"trip", "organizer", "remove", "t1", "--member", "m1", "--force"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if api.removeCalls != 1 {
		t.Fatalf("expected 1 call, got %d", api.removeCalls)
	}
	if api.lastRemoveIdem == "" {
		t.Fatalf("expected generated idempotency key")
	}
	if api.lastRemoveTripID != "t1" || api.lastRemoveMember != "m1" {
		t.Fatalf("call args: trip=%q member=%q", api.lastRemoveTripID, api.lastRemoveMember)
	}
}
