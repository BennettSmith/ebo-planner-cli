package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"

	gen "github.com/BennettSmith/ebo-planner-cli/internal/gen/plannerapi"
	"github.com/BennettSmith/ebo-planner-cli/internal/platform/exitcode"
	outplannerapi "github.com/BennettSmith/ebo-planner-cli/internal/ports/out/plannerapi"
)

type fakeRSVPAPI struct {
	setCalls     int
	getCalls     int
	summaryCalls int

	lastSetIdem string
	lastSetReq  gen.SetMyRSVPRequest

	getResp     *gen.GetMyRSVPForTripClientResponse
	summaryResp *gen.GetTripRSVPSummaryClientResponse
}

var _ outplannerapi.Client = (*fakeRSVPAPI)(nil)

func (f *fakeRSVPAPI) ListVisibleTripsForMember(ctx context.Context, baseURL string, bearerToken string) (*gen.ListVisibleTripsForMemberClientResponse, error) {
	return nil, exitcode.New(exitcode.KindUnexpected, "not implemented in test", nil)
}
func (f *fakeRSVPAPI) ListMyDraftTrips(ctx context.Context, baseURL string, bearerToken string) (*gen.ListMyDraftTripsClientResponse, error) {
	return nil, exitcode.New(exitcode.KindUnexpected, "not implemented in test", nil)
}
func (f *fakeRSVPAPI) GetTripDetails(ctx context.Context, baseURL string, bearerToken string, tripID gen.TripId) (*gen.GetTripDetailsClientResponse, error) {
	return nil, exitcode.New(exitcode.KindUnexpected, "not implemented in test", nil)
}
func (f *fakeRSVPAPI) CreateTripDraft(ctx context.Context, baseURL string, bearerToken string, idempotencyKey string, req gen.CreateTripDraftJSONRequestBody) (*gen.CreateTripDraftClientResponse, error) {
	return nil, exitcode.New(exitcode.KindUnexpected, "not implemented in test", nil)
}
func (f *fakeRSVPAPI) UpdateTrip(ctx context.Context, baseURL string, bearerToken string, tripID gen.TripId, idempotencyKey string, req gen.UpdateTripJSONRequestBody) (*gen.UpdateTripClientResponse, error) {
	return nil, exitcode.New(exitcode.KindUnexpected, "not implemented in test", nil)
}
func (f *fakeRSVPAPI) SetTripDraftVisibility(ctx context.Context, baseURL string, bearerToken string, tripID gen.TripId, idempotencyKey string, req gen.SetTripDraftVisibilityJSONRequestBody) (*gen.SetTripDraftVisibilityClientResponse, error) {
	return nil, exitcode.New(exitcode.KindUnexpected, "not implemented in test", nil)
}
func (f *fakeRSVPAPI) PublishTrip(ctx context.Context, baseURL string, bearerToken string, tripID gen.TripId) (*gen.PublishTripClientResponse, error) {
	return nil, exitcode.New(exitcode.KindUnexpected, "not implemented in test", nil)
}
func (f *fakeRSVPAPI) AddTripOrganizer(ctx context.Context, baseURL string, bearerToken string, tripID gen.TripId, idempotencyKey string, req gen.AddTripOrganizerJSONRequestBody) (*gen.AddTripOrganizerClientResponse, error) {
	return nil, exitcode.New(exitcode.KindUnexpected, "not implemented in test", nil)
}
func (f *fakeRSVPAPI) RemoveTripOrganizer(ctx context.Context, baseURL string, bearerToken string, tripID gen.TripId, memberID gen.MemberId, idempotencyKey string) (*gen.RemoveTripOrganizerClientResponse, error) {
	return nil, exitcode.New(exitcode.KindUnexpected, "not implemented in test", nil)
}
func (f *fakeRSVPAPI) CancelTrip(ctx context.Context, baseURL string, bearerToken string, tripID gen.TripId, idempotencyKey *string) (*gen.CancelTripClientResponse, error) {
	return nil, exitcode.New(exitcode.KindUnexpected, "not implemented in test", nil)
}
func (f *fakeRSVPAPI) ListMembers(ctx context.Context, baseURL string, bearerToken string, params *gen.ListMembersParams) (*gen.ListMembersClientResponse, error) {
	return nil, exitcode.New(exitcode.KindUnexpected, "not implemented in test", nil)
}
func (f *fakeRSVPAPI) UpdateMyMemberProfile(ctx context.Context, baseURL string, bearerToken string, idempotencyKey string, req gen.UpdateMyMemberProfileJSONRequestBody) (*gen.UpdateMyMemberProfileClientResponse, error) {
	return nil, exitcode.New(exitcode.KindUnexpected, "not implemented in test", nil)
}

func (f *fakeRSVPAPI) SearchMembers(ctx context.Context, baseURL string, bearerToken string, params *gen.SearchMembersParams) (*gen.SearchMembersClientResponse, error) {
	return nil, exitcode.New(exitcode.KindUnexpected, "not implemented in test", nil)
}

func (f *fakeRSVPAPI) GetMyMemberProfile(ctx context.Context, baseURL string, bearerToken string) (*gen.GetMyMemberProfileClientResponse, error) {
	return nil, exitcode.New(exitcode.KindUnexpected, "not implemented in test", nil)
}

func (f *fakeRSVPAPI) CreateMyMember(ctx context.Context, baseURL string, bearerToken string, req gen.CreateMyMemberJSONRequestBody) (*gen.CreateMyMemberClientResponse, error) {
	return nil, exitcode.New(exitcode.KindUnexpected, "not implemented in test", nil)
}

func (f *fakeRSVPAPI) SetMyRSVP(ctx context.Context, baseURL string, bearerToken string, tripID gen.TripId, idempotencyKey string, req gen.SetMyRSVPJSONRequestBody) (*gen.SetMyRSVPClientResponse, error) {
	_ = ctx
	_ = baseURL
	_ = bearerToken
	_ = tripID
	f.setCalls++
	f.lastSetIdem = idempotencyKey
	f.lastSetReq = req
	return &gen.SetMyRSVPClientResponse{JSON200: &gen.SetMyRSVPResponse{MyRsvp: gen.MyRSVP{Response: req.Response}}}, nil
}

func (f *fakeRSVPAPI) GetMyRSVPForTrip(ctx context.Context, baseURL string, bearerToken string, tripID gen.TripId) (*gen.GetMyRSVPForTripClientResponse, error) {
	_ = ctx
	_ = baseURL
	_ = bearerToken
	_ = tripID
	f.getCalls++
	if f.getResp != nil {
		return f.getResp, nil
	}
	return &gen.GetMyRSVPForTripClientResponse{JSON200: &struct {
		MyRsvp gen.MyRSVP `json:"myRsvp"`
	}{MyRsvp: gen.MyRSVP{Response: gen.YES}}}, nil
}

func (f *fakeRSVPAPI) GetTripRSVPSummary(ctx context.Context, baseURL string, bearerToken string, tripID gen.TripId) (*gen.GetTripRSVPSummaryClientResponse, error) {
	_ = ctx
	_ = baseURL
	_ = bearerToken
	_ = tripID
	f.summaryCalls++
	if f.summaryResp != nil {
		return f.summaryResp, nil
	}
	return &gen.GetTripRSVPSummaryClientResponse{JSON200: &struct {
		RsvpSummary gen.TripRSVPSummary `json:"rsvpSummary"`
	}{RsvpSummary: gen.TripRSVPSummary{AttendingRigs: 1}}}, nil
}

func TestTripRSVPSet_MutuallyExclusiveFlags(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	store := &memStore{path: "/x", doc: baseDoc(t)}
	api := &fakeRSVPAPI{}

	cmd := NewRootCmd(RootDeps{ConfigStore: store, PlannerAPI: api, Stdout: stdout, Stderr: stderr})
	cmd.SetArgs([]string{"trip", "rsvp", "set", "t1", "--yes", "--no"})
	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected error")
	}
	if exitcode.Code(err) != exitcode.Usage {
		t.Fatalf("expected exit 2, got %d (%v)", exitcode.Code(err), err)
	}
	if api.setCalls != 0 {
		t.Fatalf("expected no api calls, got %d", api.setCalls)
	}
}

func TestTripRSVPSet_MissingChoice_IsUsage_NoAPICall(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	store := &memStore{path: "/x", doc: baseDoc(t)}
	api := &fakeRSVPAPI{}

	cmd := NewRootCmd(RootDeps{ConfigStore: store, PlannerAPI: api, Stdout: stdout, Stderr: stderr})
	cmd.SetArgs([]string{"trip", "rsvp", "set", "t1"})
	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected error")
	}
	if exitcode.Code(err) != exitcode.Usage {
		t.Fatalf("expected exit 2, got %d (%v)", exitcode.Code(err), err)
	}
	if api.setCalls != 0 {
		t.Fatalf("expected no api calls, got %d", api.setCalls)
	}
}

func TestTripRSVPSet_AutoGeneratesIdempotency_JSONMeta(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	store := &memStore{path: "/x", doc: baseDoc(t)}
	api := &fakeRSVPAPI{}

	cmd := NewRootCmd(RootDeps{ConfigStore: store, PlannerAPI: api, Stdout: stdout, Stderr: stderr})
	cmd.SetArgs([]string{"--output", "json", "trip", "rsvp", "set", "t1", "--yes"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if api.setCalls != 1 {
		t.Fatalf("expected 1 call, got %d", api.setCalls)
	}
	if api.lastSetIdem == "" {
		t.Fatalf("expected generated idempotency key")
	}
	if api.lastSetReq.Response != gen.YES {
		t.Fatalf("response: %q", api.lastSetReq.Response)
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

func TestTripRSVPGet_JSONEnvelope(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	store := &memStore{path: "/x", doc: baseDoc(t)}
	api := &fakeRSVPAPI{}

	cmd := NewRootCmd(RootDeps{ConfigStore: store, PlannerAPI: api, Stdout: stdout, Stderr: stderr})
	cmd.SetArgs([]string{"--output", "json", "trip", "rsvp", "get", "t1"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if api.getCalls != 1 {
		t.Fatalf("expected 1 call, got %d", api.getCalls)
	}
	var env map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &env); err != nil {
		t.Fatalf("stdout not json: %v\n%s", err, stdout.String())
	}
	if env["data"] == nil {
		t.Fatalf("expected data")
	}
}

func TestTripRSVPSet_Unset_SendsUNSET(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	store := &memStore{path: "/x", doc: baseDoc(t)}
	api := &fakeRSVPAPI{}

	cmd := NewRootCmd(RootDeps{ConfigStore: store, PlannerAPI: api, Stdout: stdout, Stderr: stderr})
	cmd.SetArgs([]string{"trip", "rsvp", "set", "t1", "--unset", "--idempotency-key", "k1"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if api.setCalls != 1 {
		t.Fatalf("expected 1 call, got %d", api.setCalls)
	}
	if api.lastSetIdem != "k1" {
		t.Fatalf("idempotency: %q", api.lastSetIdem)
	}
	if api.lastSetReq.Response != gen.UNSET {
		t.Fatalf("response: %q", api.lastSetReq.Response)
	}
}

func TestTripRSVPSummary_JSONEnvelope(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	store := &memStore{path: "/x", doc: baseDoc(t)}
	api := &fakeRSVPAPI{}

	cmd := NewRootCmd(RootDeps{ConfigStore: store, PlannerAPI: api, Stdout: stdout, Stderr: stderr})
	cmd.SetArgs([]string{"--output", "json", "trip", "rsvp", "summary", "t1"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if api.summaryCalls != 1 {
		t.Fatalf("expected 1 call, got %d", api.summaryCalls)
	}
	var env map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &env); err != nil {
		t.Fatalf("stdout not json: %v\n%s", err, stdout.String())
	}
	if env["data"] == nil {
		t.Fatalf("expected data")
	}
}

func TestTripRSVPSet_DefaultOutput_PrintsIdempotency(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	store := &memStore{path: "/x", doc: baseDoc(t)}
	api := &fakeRSVPAPI{}

	cmd := NewRootCmd(RootDeps{ConfigStore: store, PlannerAPI: api, Stdout: stdout, Stderr: stderr})
	cmd.SetArgs([]string{"trip", "rsvp", "set", "t1", "--yes"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if !strings.Contains(stderr.String(), "Idempotency-Key: ") {
		t.Fatalf("stderr: %q", stderr.String())
	}
	if !strings.Contains(stdout.String(), "OK") {
		t.Fatalf("stdout: %q", stdout.String())
	}
}

func TestTripRSVPGet_TableOutput(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	store := &memStore{path: "/x", doc: baseDoc(t)}
	api := &fakeRSVPAPI{}

	cmd := NewRootCmd(RootDeps{ConfigStore: store, PlannerAPI: api, Stdout: stdout, Stderr: stderr})
	cmd.SetArgs([]string{"trip", "rsvp", "get", "t1"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if want := "TRIP_ID\tRESPONSE\nt1\tYES\n"; stdout.String() != want {
		t.Fatalf("stdout: %q", stdout.String())
	}
}

func TestTripRSVPSummary_TableOutput(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	store := &memStore{path: "/x", doc: baseDoc(t)}
	api := &fakeRSVPAPI{}

	cmd := NewRootCmd(RootDeps{ConfigStore: store, PlannerAPI: api, Stdout: stdout, Stderr: stderr})
	cmd.SetArgs([]string{"trip", "rsvp", "summary", "t1"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if want := "TRIP_ID\tATTENDING_RIGS\tATTENDING_MEMBERS\tNOT_ATTENDING_MEMBERS\nt1\t1\t0\t0\n"; stdout.String() != want {
		t.Fatalf("stdout: %q", stdout.String())
	}
}

func TestTripRSVPGet_TableOutput_JSON200Nil_PrintsOK(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	store := &memStore{path: "/x", doc: baseDoc(t)}
	api := &fakeRSVPAPI{getResp: &gen.GetMyRSVPForTripClientResponse{}}

	cmd := NewRootCmd(RootDeps{ConfigStore: store, PlannerAPI: api, Stdout: stdout, Stderr: stderr})
	cmd.SetArgs([]string{"trip", "rsvp", "get", "t1"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if stdout.String() != "OK\n" {
		t.Fatalf("stdout: %q", stdout.String())
	}
}

func TestTripRSVPSummary_TableOutput_JSON200Nil_PrintsOK(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	store := &memStore{path: "/x", doc: baseDoc(t)}
	api := &fakeRSVPAPI{summaryResp: &gen.GetTripRSVPSummaryClientResponse{}}

	cmd := NewRootCmd(RootDeps{ConfigStore: store, PlannerAPI: api, Stdout: stdout, Stderr: stderr})
	cmd.SetArgs([]string{"trip", "rsvp", "summary", "t1"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if stdout.String() != "OK\n" {
		t.Fatalf("stdout: %q", stdout.String())
	}
}
