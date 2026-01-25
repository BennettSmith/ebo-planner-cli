package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"

	gen "github.com/Overland-East-Bay/trip-planner-cli/internal/gen/plannerapi"
	"github.com/Overland-East-Bay/trip-planner-cli/internal/platform/exitcode"
	outplannerapi "github.com/Overland-East-Bay/trip-planner-cli/internal/ports/out/plannerapi"
)

type fakeMemberDeleteAPI struct {
	deleteCalls int

	lastIdem string
	lastReq  gen.DeleteMyMemberAccountJSONRequestBody

	resp *gen.DeleteMyMemberAccountClientResponse
}

var _ outplannerapi.Client = (*fakeMemberDeleteAPI)(nil)

// Trips (unused)
func (f *fakeMemberDeleteAPI) ListVisibleTripsForMember(ctx context.Context, baseURL string, bearerToken string) (*gen.ListVisibleTripsForMemberClientResponse, error) {
	return nil, exitcode.New(exitcode.KindUnexpected, "not implemented in test", nil)
}
func (f *fakeMemberDeleteAPI) ListMyDraftTrips(ctx context.Context, baseURL string, bearerToken string) (*gen.ListMyDraftTripsClientResponse, error) {
	return nil, exitcode.New(exitcode.KindUnexpected, "not implemented in test", nil)
}
func (f *fakeMemberDeleteAPI) GetTripDetails(ctx context.Context, baseURL string, bearerToken string, tripID gen.TripId) (*gen.GetTripDetailsClientResponse, error) {
	return nil, exitcode.New(exitcode.KindUnexpected, "not implemented in test", nil)
}
func (f *fakeMemberDeleteAPI) CreateTripDraft(ctx context.Context, baseURL string, bearerToken string, idempotencyKey string, req gen.CreateTripDraftJSONRequestBody) (*gen.CreateTripDraftClientResponse, error) {
	return nil, exitcode.New(exitcode.KindUnexpected, "not implemented in test", nil)
}
func (f *fakeMemberDeleteAPI) UpdateTrip(ctx context.Context, baseURL string, bearerToken string, tripID gen.TripId, idempotencyKey string, req gen.UpdateTripJSONRequestBody) (*gen.UpdateTripClientResponse, error) {
	return nil, exitcode.New(exitcode.KindUnexpected, "not implemented in test", nil)
}
func (f *fakeMemberDeleteAPI) SetTripDraftVisibility(ctx context.Context, baseURL string, bearerToken string, tripID gen.TripId, idempotencyKey string, req gen.SetTripDraftVisibilityJSONRequestBody) (*gen.SetTripDraftVisibilityClientResponse, error) {
	return nil, exitcode.New(exitcode.KindUnexpected, "not implemented in test", nil)
}
func (f *fakeMemberDeleteAPI) PublishTrip(ctx context.Context, baseURL string, bearerToken string, tripID gen.TripId) (*gen.PublishTripClientResponse, error) {
	return nil, exitcode.New(exitcode.KindUnexpected, "not implemented in test", nil)
}
func (f *fakeMemberDeleteAPI) AddTripOrganizer(ctx context.Context, baseURL string, bearerToken string, tripID gen.TripId, idempotencyKey string, req gen.AddTripOrganizerJSONRequestBody) (*gen.AddTripOrganizerClientResponse, error) {
	return nil, exitcode.New(exitcode.KindUnexpected, "not implemented in test", nil)
}
func (f *fakeMemberDeleteAPI) RemoveTripOrganizer(ctx context.Context, baseURL string, bearerToken string, tripID gen.TripId, memberID gen.MemberId, idempotencyKey string) (*gen.RemoveTripOrganizerClientResponse, error) {
	return nil, exitcode.New(exitcode.KindUnexpected, "not implemented in test", nil)
}
func (f *fakeMemberDeleteAPI) CancelTrip(ctx context.Context, baseURL string, bearerToken string, tripID gen.TripId, idempotencyKey *string) (*gen.CancelTripClientResponse, error) {
	return nil, exitcode.New(exitcode.KindUnexpected, "not implemented in test", nil)
}
func (f *fakeMemberDeleteAPI) SetMyRSVP(ctx context.Context, baseURL string, bearerToken string, tripID gen.TripId, idempotencyKey string, req gen.SetMyRSVPJSONRequestBody) (*gen.SetMyRSVPClientResponse, error) {
	return nil, exitcode.New(exitcode.KindUnexpected, "not implemented in test", nil)
}
func (f *fakeMemberDeleteAPI) GetMyRSVPForTrip(ctx context.Context, baseURL string, bearerToken string, tripID gen.TripId) (*gen.GetMyRSVPForTripClientResponse, error) {
	return nil, exitcode.New(exitcode.KindUnexpected, "not implemented in test", nil)
}
func (f *fakeMemberDeleteAPI) GetTripRSVPSummary(ctx context.Context, baseURL string, bearerToken string, tripID gen.TripId) (*gen.GetTripRSVPSummaryClientResponse, error) {
	return nil, exitcode.New(exitcode.KindUnexpected, "not implemented in test", nil)
}

// Members (unused except delete)
func (f *fakeMemberDeleteAPI) ListMembers(ctx context.Context, baseURL string, bearerToken string, params *gen.ListMembersParams) (*gen.ListMembersClientResponse, error) {
	return nil, exitcode.New(exitcode.KindUnexpected, "not implemented in test", nil)
}
func (f *fakeMemberDeleteAPI) SearchMembers(ctx context.Context, baseURL string, bearerToken string, params *gen.SearchMembersParams) (*gen.SearchMembersClientResponse, error) {
	return nil, exitcode.New(exitcode.KindUnexpected, "not implemented in test", nil)
}
func (f *fakeMemberDeleteAPI) GetMyMemberProfile(ctx context.Context, baseURL string, bearerToken string) (*gen.GetMyMemberProfileClientResponse, error) {
	return nil, exitcode.New(exitcode.KindUnexpected, "not implemented in test", nil)
}
func (f *fakeMemberDeleteAPI) CreateMyMember(ctx context.Context, baseURL string, bearerToken string, req gen.CreateMyMemberJSONRequestBody) (*gen.CreateMyMemberClientResponse, error) {
	return nil, exitcode.New(exitcode.KindUnexpected, "not implemented in test", nil)
}
func (f *fakeMemberDeleteAPI) UpdateMyMemberProfile(ctx context.Context, baseURL string, bearerToken string, idempotencyKey string, req gen.UpdateMyMemberProfileJSONRequestBody) (*gen.UpdateMyMemberProfileClientResponse, error) {
	return nil, exitcode.New(exitcode.KindUnexpected, "not implemented in test", nil)
}

func (f *fakeMemberDeleteAPI) DeleteMyMemberAccount(ctx context.Context, baseURL string, bearerToken string, idempotencyKey string, req gen.DeleteMyMemberAccountJSONRequestBody) (*gen.DeleteMyMemberAccountClientResponse, error) {
	_ = ctx
	_ = baseURL
	_ = bearerToken
	f.deleteCalls++
	f.lastIdem = idempotencyKey
	f.lastReq = req
	if f.resp != nil {
		return f.resp, nil
	}
	return &gen.DeleteMyMemberAccountClientResponse{JSON200: &gen.DeleteMyMemberResponse{Deleted: true}}, nil
}

func TestMemberDelete_RequiresForce_Exit2_NoAPICall(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	store := &memStore{path: "/x", doc: baseDoc(t)}
	api := &fakeMemberDeleteAPI{}

	cmd := NewRootCmd(RootDeps{ConfigStore: store, PlannerAPI: api, Stdout: stdout, Stderr: stderr})
	cmd.SetArgs([]string{"member", "delete"})
	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected error")
	}
	if exitcode.Code(err) != exitcode.Usage {
		t.Fatalf("expected exit 2, got %d (%v)", exitcode.Code(err), err)
	}
	if api.deleteCalls != 0 {
		t.Fatalf("expected no api calls, got %d", api.deleteCalls)
	}
}

func TestMemberDelete_AutoGeneratesIdempotency_AndSendsConfirmTrue(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	store := &memStore{path: "/x", doc: baseDoc(t)}
	api := &fakeMemberDeleteAPI{}

	cmd := NewRootCmd(RootDeps{ConfigStore: store, PlannerAPI: api, Stdout: stdout, Stderr: stderr})
	cmd.SetArgs([]string{"member", "delete", "--force"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if api.deleteCalls != 1 {
		t.Fatalf("expected 1 call, got %d", api.deleteCalls)
	}
	if api.lastIdem == "" {
		t.Fatalf("expected generated idempotency key")
	}
	if !strings.Contains(stderr.String(), "Idempotency-Key: ") {
		t.Fatalf("expected idempotency printed to stderr, got %q", stderr.String())
	}
	if api.lastReq.Confirm != true {
		t.Fatalf("expected confirm=true, got %#v", api.lastReq.Confirm)
	}
}

func TestMemberDelete_ReasonIsTrimmedAndIncluded(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	store := &memStore{path: "/x", doc: baseDoc(t)}
	api := &fakeMemberDeleteAPI{}

	cmd := NewRootCmd(RootDeps{ConfigStore: store, PlannerAPI: api, Stdout: stdout, Stderr: stderr})
	cmd.SetArgs([]string{"member", "delete", "--force", "--reason", "  bye  "})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if api.lastReq.Reason == nil || *api.lastReq.Reason != "bye" {
		t.Fatalf("expected trimmed reason, got %#v", api.lastReq.Reason)
	}
}

func TestMemberDelete_JSONEnvelopeHasMetaIdempotencyKey(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	store := &memStore{path: "/x", doc: baseDoc(t)}
	api := &fakeMemberDeleteAPI{}

	cmd := NewRootCmd(RootDeps{ConfigStore: store, PlannerAPI: api, Stdout: stdout, Stderr: stderr})
	cmd.SetArgs([]string{"--output", "json", "member", "delete", "--force", "--idempotency-key", "k1"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if api.lastIdem != "k1" {
		t.Fatalf("expected idempotency k1, got %q", api.lastIdem)
	}

	var env map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &env); err != nil {
		t.Fatalf("stdout not json: %v\n%s", err, stdout.String())
	}
	meta := env["meta"].(map[string]any)
	if meta["idempotencyKey"] != "k1" {
		t.Fatalf("meta.idempotencyKey: %#v", meta["idempotencyKey"])
	}
}

func TestMemberDelete_RejectsMultilineReason(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	store := &memStore{path: "/x", doc: baseDoc(t)}
	api := &fakeMemberDeleteAPI{}

	cmd := NewRootCmd(RootDeps{ConfigStore: store, PlannerAPI: api, Stdout: stdout, Stderr: stderr})
	cmd.SetArgs([]string{"member", "delete", "--force", "--reason", "line1\nline2"})
	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected error")
	}
	if exitcode.Code(err) != exitcode.Usage {
		t.Fatalf("expected exit 2, got %d (%v)", exitcode.Code(err), err)
	}
}
