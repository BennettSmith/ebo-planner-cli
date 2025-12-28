package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"

	plannerapiout "github.com/BennettSmith/ebo-planner-cli/internal/adapters/out/plannerapi"
	gen "github.com/BennettSmith/ebo-planner-cli/internal/gen/plannerapi"
	"github.com/BennettSmith/ebo-planner-cli/internal/platform/exitcode"
	outplannerapi "github.com/BennettSmith/ebo-planner-cli/internal/ports/out/plannerapi"
)

type fakeMemberReadAPI struct {
	listCalls   int
	searchCalls int
	meCalls     int

	lastListParams   *gen.ListMembersParams
	lastSearchParams *gen.SearchMembersParams

	meResp *gen.GetMyMemberProfileClientResponse
	meErr  error
}

var _ outplannerapi.Client = (*fakeMemberReadAPI)(nil)

// Trips (unused)
func (f *fakeMemberReadAPI) ListVisibleTripsForMember(ctx context.Context, baseURL string, bearerToken string) (*gen.ListVisibleTripsForMemberClientResponse, error) {
	return nil, exitcode.New(exitcode.KindUnexpected, "not implemented in test", nil)
}
func (f *fakeMemberReadAPI) ListMyDraftTrips(ctx context.Context, baseURL string, bearerToken string) (*gen.ListMyDraftTripsClientResponse, error) {
	return nil, exitcode.New(exitcode.KindUnexpected, "not implemented in test", nil)
}
func (f *fakeMemberReadAPI) GetTripDetails(ctx context.Context, baseURL string, bearerToken string, tripID gen.TripId) (*gen.GetTripDetailsClientResponse, error) {
	return nil, exitcode.New(exitcode.KindUnexpected, "not implemented in test", nil)
}
func (f *fakeMemberReadAPI) CreateTripDraft(ctx context.Context, baseURL string, bearerToken string, idempotencyKey string, req gen.CreateTripDraftJSONRequestBody) (*gen.CreateTripDraftClientResponse, error) {
	return nil, exitcode.New(exitcode.KindUnexpected, "not implemented in test", nil)
}
func (f *fakeMemberReadAPI) UpdateTrip(ctx context.Context, baseURL string, bearerToken string, tripID gen.TripId, idempotencyKey string, req gen.UpdateTripJSONRequestBody) (*gen.UpdateTripClientResponse, error) {
	return nil, exitcode.New(exitcode.KindUnexpected, "not implemented in test", nil)
}
func (f *fakeMemberReadAPI) SetTripDraftVisibility(ctx context.Context, baseURL string, bearerToken string, tripID gen.TripId, idempotencyKey string, req gen.SetTripDraftVisibilityJSONRequestBody) (*gen.SetTripDraftVisibilityClientResponse, error) {
	return nil, exitcode.New(exitcode.KindUnexpected, "not implemented in test", nil)
}
func (f *fakeMemberReadAPI) PublishTrip(ctx context.Context, baseURL string, bearerToken string, tripID gen.TripId) (*gen.PublishTripClientResponse, error) {
	return nil, exitcode.New(exitcode.KindUnexpected, "not implemented in test", nil)
}
func (f *fakeMemberReadAPI) AddTripOrganizer(ctx context.Context, baseURL string, bearerToken string, tripID gen.TripId, idempotencyKey string, req gen.AddTripOrganizerJSONRequestBody) (*gen.AddTripOrganizerClientResponse, error) {
	return nil, exitcode.New(exitcode.KindUnexpected, "not implemented in test", nil)
}
func (f *fakeMemberReadAPI) RemoveTripOrganizer(ctx context.Context, baseURL string, bearerToken string, tripID gen.TripId, memberID gen.MemberId, idempotencyKey string) (*gen.RemoveTripOrganizerClientResponse, error) {
	return nil, exitcode.New(exitcode.KindUnexpected, "not implemented in test", nil)
}
func (f *fakeMemberReadAPI) CancelTrip(ctx context.Context, baseURL string, bearerToken string, tripID gen.TripId, idempotencyKey *string) (*gen.CancelTripClientResponse, error) {
	return nil, exitcode.New(exitcode.KindUnexpected, "not implemented in test", nil)
}
func (f *fakeMemberReadAPI) SetMyRSVP(ctx context.Context, baseURL string, bearerToken string, tripID gen.TripId, idempotencyKey string, req gen.SetMyRSVPJSONRequestBody) (*gen.SetMyRSVPClientResponse, error) {
	return nil, exitcode.New(exitcode.KindUnexpected, "not implemented in test", nil)
}
func (f *fakeMemberReadAPI) GetMyRSVPForTrip(ctx context.Context, baseURL string, bearerToken string, tripID gen.TripId) (*gen.GetMyRSVPForTripClientResponse, error) {
	return nil, exitcode.New(exitcode.KindUnexpected, "not implemented in test", nil)
}
func (f *fakeMemberReadAPI) GetTripRSVPSummary(ctx context.Context, baseURL string, bearerToken string, tripID gen.TripId) (*gen.GetTripRSVPSummaryClientResponse, error) {
	return nil, exitcode.New(exitcode.KindUnexpected, "not implemented in test", nil)
}

// Members
func (f *fakeMemberReadAPI) ListMembers(ctx context.Context, baseURL string, bearerToken string, params *gen.ListMembersParams) (*gen.ListMembersClientResponse, error) {
	_ = ctx
	_ = baseURL
	_ = bearerToken
	f.listCalls++
	f.lastListParams = params
	return &gen.ListMembersClientResponse{JSON200: &struct {
		Members []gen.MemberDirectoryEntry `json:"members"`
	}{Members: []gen.MemberDirectoryEntry{{MemberId: "m1", DisplayName: "Alice"}}}}, nil
}

func (f *fakeMemberReadAPI) SearchMembers(ctx context.Context, baseURL string, bearerToken string, params *gen.SearchMembersParams) (*gen.SearchMembersClientResponse, error) {
	_ = ctx
	_ = baseURL
	_ = bearerToken
	f.searchCalls++
	f.lastSearchParams = params
	return &gen.SearchMembersClientResponse{JSON200: &struct {
		Members []gen.MemberDirectoryEntry `json:"members"`
	}{Members: []gen.MemberDirectoryEntry{{MemberId: "m2", DisplayName: "Bob"}}}}, nil
}

func (f *fakeMemberReadAPI) GetMyMemberProfile(ctx context.Context, baseURL string, bearerToken string) (*gen.GetMyMemberProfileClientResponse, error) {
	_ = ctx
	_ = baseURL
	_ = bearerToken
	f.meCalls++
	if f.meErr != nil {
		return nil, f.meErr
	}
	if f.meResp != nil {
		return f.meResp, nil
	}
	return &gen.GetMyMemberProfileClientResponse{JSON200: &struct {
		Member gen.MemberProfile `json:"member"`
	}{Member: gen.MemberProfile{MemberId: "m1", DisplayName: "Me", Email: "me@example.com"}}}, nil
}

func (f *fakeMemberReadAPI) UpdateMyMemberProfile(ctx context.Context, baseURL string, bearerToken string, idempotencyKey string, req gen.UpdateMyMemberProfileJSONRequestBody) (*gen.UpdateMyMemberProfileClientResponse, error) {
	return nil, exitcode.New(exitcode.KindUnexpected, "not implemented in test", nil)
}

func TestMemberSearch_MinLen3_IsUsage_NoAPICall(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	store := &memStore{path: "/x", doc: baseDoc(t)}
	api := &fakeMemberReadAPI{}

	cmd := NewRootCmd(RootDeps{ConfigStore: store, PlannerAPI: api, Stdout: stdout, Stderr: stderr})
	cmd.SetArgs([]string{"member", "search", "ab"})
	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected error")
	}
	if exitcode.Code(err) != exitcode.Usage {
		t.Fatalf("expected exit 2, got %d (%v)", exitcode.Code(err), err)
	}
	if api.searchCalls != 0 {
		t.Fatalf("expected no api calls, got %d", api.searchCalls)
	}
}

func TestMemberMe_MemberNotProvisioned_IsNotFound_WithGuidance(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	store := &memStore{path: "/x", doc: baseDoc(t)}
	api := &fakeMemberReadAPI{
		meErr: exitcode.New(exitcode.KindConflict, "conflict", &plannerapiout.APIError{StatusCode: 409, ErrorCode: "MEMBER_NOT_PROVISIONED", Message: "missing"}),
	}

	cmd := NewRootCmd(RootDeps{ConfigStore: store, PlannerAPI: api, Stdout: stdout, Stderr: stderr})
	cmd.SetArgs([]string{"member", "me"})
	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected error")
	}
	if exitcode.Code(err) != exitcode.NotFound {
		t.Fatalf("expected exit 4, got %d (%v)", exitcode.Code(err), err)
	}
	if !strings.Contains(err.Error(), "ebo member create") {
		t.Fatalf("expected guidance, got %q", err.Error())
	}
}

func TestMemberList_JSONEnvelope(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	store := &memStore{path: "/x", doc: baseDoc(t)}
	api := &fakeMemberReadAPI{}

	cmd := NewRootCmd(RootDeps{ConfigStore: store, PlannerAPI: api, Stdout: stdout, Stderr: stderr})
	cmd.SetArgs([]string{"--output", "json", "member", "list"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if api.listCalls != 1 {
		t.Fatalf("expected 1 call, got %d", api.listCalls)
	}
	var env map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &env); err != nil {
		t.Fatalf("stdout not json: %v\n%s", err, stdout.String())
	}
	if env["data"] == nil {
		t.Fatalf("expected data")
	}
}

func TestMemberList_IncludeInactive_SetsParam(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	store := &memStore{path: "/x", doc: baseDoc(t)}
	api := &fakeMemberReadAPI{}

	cmd := NewRootCmd(RootDeps{ConfigStore: store, PlannerAPI: api, Stdout: stdout, Stderr: stderr})
	cmd.SetArgs([]string{"member", "list", "--include-inactive"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if api.lastListParams == nil || api.lastListParams.IncludeInactive == nil || bool(*api.lastListParams.IncludeInactive) != true {
		t.Fatalf("params: %#v", api.lastListParams)
	}
}

func TestMemberSearch_JSONEnvelope(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	store := &memStore{path: "/x", doc: baseDoc(t)}
	api := &fakeMemberReadAPI{}

	cmd := NewRootCmd(RootDeps{ConfigStore: store, PlannerAPI: api, Stdout: stdout, Stderr: stderr})
	cmd.SetArgs([]string{"--output", "json", "member", "search", "bob"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if api.searchCalls != 1 {
		t.Fatalf("expected 1 call, got %d", api.searchCalls)
	}
	if api.lastSearchParams == nil || api.lastSearchParams.Q != "bob" {
		t.Fatalf("params: %#v", api.lastSearchParams)
	}
	var env map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &env); err != nil {
		t.Fatalf("stdout not json: %v\n%s", err, stdout.String())
	}
	if env["data"] == nil {
		t.Fatalf("expected data")
	}
}

func TestMemberSearch_TableOutput(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	store := &memStore{path: "/x", doc: baseDoc(t)}
	api := &fakeMemberReadAPI{}

	cmd := NewRootCmd(RootDeps{ConfigStore: store, PlannerAPI: api, Stdout: stdout, Stderr: stderr})
	cmd.SetArgs([]string{"member", "search", "bob"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if !strings.Contains(stdout.String(), "MEMBER_ID") || !strings.Contains(stdout.String(), "m2") {
		t.Fatalf("stdout: %q", stdout.String())
	}
}

func TestMemberMe_JSONEnvelope(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	store := &memStore{path: "/x", doc: baseDoc(t)}
	api := &fakeMemberReadAPI{}

	cmd := NewRootCmd(RootDeps{ConfigStore: store, PlannerAPI: api, Stdout: stdout, Stderr: stderr})
	cmd.SetArgs([]string{"--output", "json", "member", "me"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	var env map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &env); err != nil {
		t.Fatalf("stdout not json: %v\n%s", err, stdout.String())
	}
	if env["data"] == nil {
		t.Fatalf("expected data")
	}
}

func TestMemberMe_TableOutput(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	store := &memStore{path: "/x", doc: baseDoc(t)}
	api := &fakeMemberReadAPI{}

	cmd := NewRootCmd(RootDeps{ConfigStore: store, PlannerAPI: api, Stdout: stdout, Stderr: stderr})
	cmd.SetArgs([]string{"member", "me"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if !strings.Contains(stdout.String(), "MemberId:") {
		t.Fatalf("stdout: %q", stdout.String())
	}
}

func TestMemberMe_TableOutput_JSON200Nil_PrintsOK(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	store := &memStore{path: "/x", doc: baseDoc(t)}
	api := &fakeMemberReadAPI{meResp: &gen.GetMyMemberProfileClientResponse{}}

	cmd := NewRootCmd(RootDeps{ConfigStore: store, PlannerAPI: api, Stdout: stdout, Stderr: stderr})
	cmd.SetArgs([]string{"member", "me"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if stdout.String() != "OK\n" {
		t.Fatalf("stdout: %q", stdout.String())
	}
}
