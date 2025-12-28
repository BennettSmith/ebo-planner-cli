package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	gen "github.com/BennettSmith/ebo-planner-cli/internal/gen/plannerapi"
	"github.com/BennettSmith/ebo-planner-cli/internal/platform/config"
	"github.com/BennettSmith/ebo-planner-cli/internal/platform/exitcode"
	outplannerapi "github.com/BennettSmith/ebo-planner-cli/internal/ports/out/plannerapi"
)

type fakeTripReadAPI struct {
	listCalls   int
	draftsCalls int
	getCalls    int

	listTrips   []gen.TripSummary
	draftsTrips []gen.TripSummary
	getTrip     *gen.TripResponse
}

var _ outplannerapi.Client = (*fakeTripReadAPI)(nil)

func (f *fakeTripReadAPI) ListVisibleTripsForMember(ctx context.Context, baseURL string, bearerToken string) (*gen.ListVisibleTripsForMemberClientResponse, error) {
	_ = ctx
	_ = baseURL
	_ = bearerToken
	f.listCalls++
	return &gen.ListVisibleTripsForMemberClientResponse{
		JSON200: &struct {
			Trips []gen.TripSummary `json:"trips"`
		}{Trips: f.listTrips},
	}, nil
}

func (f *fakeTripReadAPI) ListMyDraftTrips(ctx context.Context, baseURL string, bearerToken string) (*gen.ListMyDraftTripsClientResponse, error) {
	_ = ctx
	_ = baseURL
	_ = bearerToken
	f.draftsCalls++
	return &gen.ListMyDraftTripsClientResponse{
		JSON200: &struct {
			Trips []gen.TripSummary `json:"trips"`
		}{Trips: f.draftsTrips},
	}, nil
}

func (f *fakeTripReadAPI) GetTripDetails(ctx context.Context, baseURL string, bearerToken string, tripID gen.TripId) (*gen.GetTripDetailsClientResponse, error) {
	_ = ctx
	_ = baseURL
	_ = bearerToken
	_ = tripID
	f.getCalls++
	return &gen.GetTripDetailsClientResponse{JSON200: f.getTrip}, nil
}

// Unused for these tests, but required by the interface.
func (f *fakeTripReadAPI) CreateTripDraft(ctx context.Context, baseURL string, bearerToken string, idempotencyKey string, req gen.CreateTripDraftJSONRequestBody) (*gen.CreateTripDraftClientResponse, error) {
	_ = ctx
	_ = baseURL
	_ = bearerToken
	_ = idempotencyKey
	_ = req
	return nil, exitcode.New(exitcode.KindUnexpected, "not implemented in test", nil)
}
func (f *fakeTripReadAPI) UpdateTrip(ctx context.Context, baseURL string, bearerToken string, tripID gen.TripId, idempotencyKey string, req gen.UpdateTripJSONRequestBody) (*gen.UpdateTripClientResponse, error) {
	_ = ctx
	_ = baseURL
	_ = bearerToken
	_ = tripID
	_ = idempotencyKey
	_ = req
	return nil, exitcode.New(exitcode.KindUnexpected, "not implemented in test", nil)
}
func (f *fakeTripReadAPI) CancelTrip(ctx context.Context, baseURL string, bearerToken string, tripID gen.TripId, idempotencyKey *string) (*gen.CancelTripClientResponse, error) {
	_ = ctx
	_ = baseURL
	_ = bearerToken
	_ = tripID
	_ = idempotencyKey
	return nil, exitcode.New(exitcode.KindUnexpected, "not implemented in test", nil)
}
func (f *fakeTripReadAPI) ListMembers(ctx context.Context, baseURL string, bearerToken string, params *gen.ListMembersParams) (*gen.ListMembersClientResponse, error) {
	_ = ctx
	_ = baseURL
	_ = bearerToken
	_ = params
	return nil, exitcode.New(exitcode.KindUnexpected, "not implemented in test", nil)
}
func (f *fakeTripReadAPI) UpdateMyMemberProfile(ctx context.Context, baseURL string, bearerToken string, idempotencyKey string, req gen.UpdateMyMemberProfileJSONRequestBody) (*gen.UpdateMyMemberProfileClientResponse, error) {
	_ = ctx
	_ = baseURL
	_ = bearerToken
	_ = idempotencyKey
	_ = req
	return nil, exitcode.New(exitcode.KindUnexpected, "not implemented in test", nil)
}

func TestTripList_TableOutput(t *testing.T) {
	name := "Trip"
	api := &fakeTripReadAPI{
		listTrips: []gen.TripSummary{{TripId: "t1", Status: "PUBLISHED", Name: &name}},
	}
	store := &memStore{path: "/x", doc: baseDoc(t)}
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	cmd := NewRootCmd(RootDeps{ConfigStore: store, PlannerAPI: api, Stdout: stdout, Stderr: stderr})
	cmd.SetArgs([]string{"trip", "list"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if api.listCalls != 1 {
		t.Fatalf("expected 1 call, got %d", api.listCalls)
	}
	if !bytes.Contains(stdout.Bytes(), []byte("TRIP_ID\tSTATUS\tNAME\n")) {
		t.Fatalf("stdout: %q", stdout.String())
	}
	if !bytes.Contains(stdout.Bytes(), []byte("t1\tPUBLISHED\tTrip\n")) {
		t.Fatalf("stdout: %q", stdout.String())
	}
}

func TestTripList_JSONOutput(t *testing.T) {
	name := "Trip"
	api := &fakeTripReadAPI{
		listTrips: []gen.TripSummary{{TripId: "t1", Status: "PUBLISHED", Name: &name}},
	}
	store := &memStore{path: "/x", doc: baseDoc(t)}
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	cmd := NewRootCmd(RootDeps{ConfigStore: store, PlannerAPI: api, Stdout: stdout, Stderr: stderr})
	cmd.SetArgs([]string{"--output", "json", "trip", "list"})
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

func TestTripDrafts_TableOutput(t *testing.T) {
	name := "Draft"
	api := &fakeTripReadAPI{
		draftsTrips: []gen.TripSummary{{TripId: "t1", Status: "DRAFT", Name: &name}},
	}
	store := &memStore{path: "/x", doc: baseDoc(t)}
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	cmd := NewRootCmd(RootDeps{ConfigStore: store, PlannerAPI: api, Stdout: stdout, Stderr: stderr})
	cmd.SetArgs([]string{"trip", "drafts"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if api.draftsCalls != 1 {
		t.Fatalf("expected 1 call, got %d", api.draftsCalls)
	}
	if !bytes.Contains(stdout.Bytes(), []byte("t1\tDRAFT\tDraft\n")) {
		t.Fatalf("stdout: %q", stdout.String())
	}
}

func TestTripDrafts_JSONOutput(t *testing.T) {
	name := "Draft"
	api := &fakeTripReadAPI{
		draftsTrips: []gen.TripSummary{{TripId: "t1", Status: "DRAFT", Name: &name}},
	}
	store := &memStore{path: "/x", doc: baseDoc(t)}
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	cmd := NewRootCmd(RootDeps{ConfigStore: store, PlannerAPI: api, Stdout: stdout, Stderr: stderr})
	cmd.SetArgs([]string{"--output", "json", "trip", "drafts"})
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

func TestTripGet_TableOutput(t *testing.T) {
	name := "Trip"
	api := &fakeTripReadAPI{
		getTrip: &gen.TripResponse{Trip: gen.TripDetails{TripId: "t1", Status: "PUBLISHED", Name: &name}},
	}
	store := &memStore{path: "/x", doc: baseDoc(t)}
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	cmd := NewRootCmd(RootDeps{ConfigStore: store, PlannerAPI: api, Stdout: stdout, Stderr: stderr})
	cmd.SetArgs([]string{"trip", "get", "t1"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if api.getCalls != 1 {
		t.Fatalf("expected 1 call, got %d", api.getCalls)
	}
	if !bytes.Contains(stdout.Bytes(), []byte("TripId: t1")) {
		t.Fatalf("stdout: %q", stdout.String())
	}
}

func TestTripGet_JSONOutput(t *testing.T) {
	name := "Trip"
	api := &fakeTripReadAPI{
		getTrip: &gen.TripResponse{Trip: gen.TripDetails{TripId: "t1", Status: "PUBLISHED", Name: &name}},
	}
	store := &memStore{path: "/x", doc: baseDoc(t)}
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	cmd := NewRootCmd(RootDeps{ConfigStore: store, PlannerAPI: api, Stdout: stdout, Stderr: stderr})
	cmd.SetArgs([]string{"--output", "json", "trip", "get", "t1"})
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

func TestTripRead_MissingToken_IsAuth(t *testing.T) {
	doc := config.NewEmptyDocument()
	doc, _ = config.WithCurrentProfile(doc, "default")
	doc, _ = config.WithProfileAPIURL(doc, "default", "http://api")
	store := &memStore{path: "/x", doc: doc}

	api := &fakeTripReadAPI{}
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	cmd := NewRootCmd(RootDeps{ConfigStore: store, PlannerAPI: api, Stdout: stdout, Stderr: stderr})
	cmd.SetArgs([]string{"trip", "list"})
	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected error")
	}
	if exitcode.Code(err) != exitcode.Auth {
		t.Fatalf("expected exit 3, got %d (%v)", exitcode.Code(err), err)
	}
	if api.listCalls != 0 {
		t.Fatalf("expected no API calls, got %d", api.listCalls)
	}
}

func TestTripRead_MissingAPIURL_IsUsage(t *testing.T) {
	doc := config.NewEmptyDocument()
	doc, _ = config.WithCurrentProfile(doc, "default")
	// Provide token but no apiUrl.
	doc, _ = config.SetString(doc, "profiles.default.auth.accessToken", "tok")
	store := &memStore{path: "/x", doc: doc}

	api := &fakeTripReadAPI{}
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	cmd := NewRootCmd(RootDeps{ConfigStore: store, PlannerAPI: api, Stdout: stdout, Stderr: stderr})
	cmd.SetArgs([]string{"trip", "list"})
	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected error")
	}
	if exitcode.Code(err) != exitcode.Usage {
		t.Fatalf("expected exit 2, got %d (%v)", exitcode.Code(err), err)
	}
	if api.listCalls != 0 {
		t.Fatalf("expected no API calls, got %d", api.listCalls)
	}
}

func TestTripList_NilPlannerAPI_IsUnexpected(t *testing.T) {
	store := &memStore{path: "/x", doc: baseDoc(t)}
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	cmd := NewRootCmd(RootDeps{ConfigStore: store, PlannerAPI: nil, Stdout: stdout, Stderr: stderr})
	cmd.SetArgs([]string{"trip", "list"})
	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected error")
	}
	if exitcode.Code(err) != exitcode.Unexpected {
		t.Fatalf("expected exit 1, got %d (%v)", exitcode.Code(err), err)
	}
}
