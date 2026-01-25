package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	gen "github.com/Overland-East-Bay/trip-planner-cli/internal/gen/plannerapi"
	"github.com/Overland-East-Bay/trip-planner-cli/internal/platform/exitcode"
	outplannerapi "github.com/Overland-East-Bay/trip-planner-cli/internal/ports/out/plannerapi"
)

type fakeMemberCreateAPI struct {
	createCalls int
	lastReq     gen.CreateMyMemberJSONRequestBody

	resp *gen.CreateMyMemberClientResponse
}

var _ outplannerapi.Client = (*fakeMemberCreateAPI)(nil)

// Trips (unused)
func (f *fakeMemberCreateAPI) ListVisibleTripsForMember(ctx context.Context, baseURL string, bearerToken string) (*gen.ListVisibleTripsForMemberClientResponse, error) {
	return nil, exitcode.New(exitcode.KindUnexpected, "not implemented in test", nil)
}
func (f *fakeMemberCreateAPI) ListMyDraftTrips(ctx context.Context, baseURL string, bearerToken string) (*gen.ListMyDraftTripsClientResponse, error) {
	return nil, exitcode.New(exitcode.KindUnexpected, "not implemented in test", nil)
}
func (f *fakeMemberCreateAPI) GetTripDetails(ctx context.Context, baseURL string, bearerToken string, tripID gen.TripId) (*gen.GetTripDetailsClientResponse, error) {
	return nil, exitcode.New(exitcode.KindUnexpected, "not implemented in test", nil)
}
func (f *fakeMemberCreateAPI) CreateTripDraft(ctx context.Context, baseURL string, bearerToken string, idempotencyKey string, req gen.CreateTripDraftJSONRequestBody) (*gen.CreateTripDraftClientResponse, error) {
	return nil, exitcode.New(exitcode.KindUnexpected, "not implemented in test", nil)
}
func (f *fakeMemberCreateAPI) UpdateTrip(ctx context.Context, baseURL string, bearerToken string, tripID gen.TripId, idempotencyKey string, req gen.UpdateTripJSONRequestBody) (*gen.UpdateTripClientResponse, error) {
	return nil, exitcode.New(exitcode.KindUnexpected, "not implemented in test", nil)
}
func (f *fakeMemberCreateAPI) SetTripDraftVisibility(ctx context.Context, baseURL string, bearerToken string, tripID gen.TripId, idempotencyKey string, req gen.SetTripDraftVisibilityJSONRequestBody) (*gen.SetTripDraftVisibilityClientResponse, error) {
	return nil, exitcode.New(exitcode.KindUnexpected, "not implemented in test", nil)
}
func (f *fakeMemberCreateAPI) PublishTrip(ctx context.Context, baseURL string, bearerToken string, tripID gen.TripId) (*gen.PublishTripClientResponse, error) {
	return nil, exitcode.New(exitcode.KindUnexpected, "not implemented in test", nil)
}
func (f *fakeMemberCreateAPI) AddTripOrganizer(ctx context.Context, baseURL string, bearerToken string, tripID gen.TripId, idempotencyKey string, req gen.AddTripOrganizerJSONRequestBody) (*gen.AddTripOrganizerClientResponse, error) {
	return nil, exitcode.New(exitcode.KindUnexpected, "not implemented in test", nil)
}
func (f *fakeMemberCreateAPI) RemoveTripOrganizer(ctx context.Context, baseURL string, bearerToken string, tripID gen.TripId, memberID gen.MemberId, idempotencyKey string) (*gen.RemoveTripOrganizerClientResponse, error) {
	return nil, exitcode.New(exitcode.KindUnexpected, "not implemented in test", nil)
}
func (f *fakeMemberCreateAPI) CancelTrip(ctx context.Context, baseURL string, bearerToken string, tripID gen.TripId, idempotencyKey *string) (*gen.CancelTripClientResponse, error) {
	return nil, exitcode.New(exitcode.KindUnexpected, "not implemented in test", nil)
}
func (f *fakeMemberCreateAPI) SetMyRSVP(ctx context.Context, baseURL string, bearerToken string, tripID gen.TripId, idempotencyKey string, req gen.SetMyRSVPJSONRequestBody) (*gen.SetMyRSVPClientResponse, error) {
	return nil, exitcode.New(exitcode.KindUnexpected, "not implemented in test", nil)
}
func (f *fakeMemberCreateAPI) GetMyRSVPForTrip(ctx context.Context, baseURL string, bearerToken string, tripID gen.TripId) (*gen.GetMyRSVPForTripClientResponse, error) {
	return nil, exitcode.New(exitcode.KindUnexpected, "not implemented in test", nil)
}
func (f *fakeMemberCreateAPI) GetTripRSVPSummary(ctx context.Context, baseURL string, bearerToken string, tripID gen.TripId) (*gen.GetTripRSVPSummaryClientResponse, error) {
	return nil, exitcode.New(exitcode.KindUnexpected, "not implemented in test", nil)
}

// Members (unused except create)
func (f *fakeMemberCreateAPI) ListMembers(ctx context.Context, baseURL string, bearerToken string, params *gen.ListMembersParams) (*gen.ListMembersClientResponse, error) {
	return nil, exitcode.New(exitcode.KindUnexpected, "not implemented in test", nil)
}
func (f *fakeMemberCreateAPI) SearchMembers(ctx context.Context, baseURL string, bearerToken string, params *gen.SearchMembersParams) (*gen.SearchMembersClientResponse, error) {
	return nil, exitcode.New(exitcode.KindUnexpected, "not implemented in test", nil)
}
func (f *fakeMemberCreateAPI) GetMyMemberProfile(ctx context.Context, baseURL string, bearerToken string) (*gen.GetMyMemberProfileClientResponse, error) {
	return nil, exitcode.New(exitcode.KindUnexpected, "not implemented in test", nil)
}

func (f *fakeMemberCreateAPI) DeleteMyMemberAccount(ctx context.Context, baseURL string, bearerToken string, idempotencyKey string, req gen.DeleteMyMemberAccountJSONRequestBody) (*gen.DeleteMyMemberAccountClientResponse, error) {
	return nil, exitcode.New(exitcode.KindUnexpected, "not implemented in test", nil)
}

func (f *fakeMemberCreateAPI) CreateMyMember(ctx context.Context, baseURL string, bearerToken string, req gen.CreateMyMemberJSONRequestBody) (*gen.CreateMyMemberClientResponse, error) {
	_ = ctx
	_ = baseURL
	_ = bearerToken
	f.createCalls++
	f.lastReq = req
	if f.resp != nil {
		return f.resp, nil
	}
	return &gen.CreateMyMemberClientResponse{JSON201: &gen.CreateMemberResponse{Member: gen.MemberProfile{MemberId: "m1", DisplayName: req.DisplayName, Email: req.Email}}}, nil
}

func (f *fakeMemberCreateAPI) UpdateMyMemberProfile(ctx context.Context, baseURL string, bearerToken string, idempotencyKey string, req gen.UpdateMyMemberProfileJSONRequestBody) (*gen.UpdateMyMemberProfileClientResponse, error) {
	return nil, exitcode.New(exitcode.KindUnexpected, "not implemented in test", nil)
}

func TestMemberCreate_RequiredFlagsMissing_IsUsage_NoAPICall(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	store := &memStore{path: "/x", doc: baseDoc(t)}
	api := &fakeMemberCreateAPI{}

	cmd := NewRootCmd(RootDeps{ConfigStore: store, PlannerAPI: api, Stdout: stdout, Stderr: stderr})
	cmd.SetArgs([]string{"member", "create", "--display-name", "A"})
	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected error")
	}
	if exitcode.Code(err) != exitcode.Usage {
		t.Fatalf("expected exit 2, got %d (%v)", exitcode.Code(err), err)
	}
	if api.createCalls != 0 {
		t.Fatalf("expected no api calls, got %d", api.createCalls)
	}
}

func TestMemberCreate_EmailValidation_IsUsage_NoAPICall(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	store := &memStore{path: "/x", doc: baseDoc(t)}
	api := &fakeMemberCreateAPI{}

	cmd := NewRootCmd(RootDeps{ConfigStore: store, PlannerAPI: api, Stdout: stdout, Stderr: stderr})
	cmd.SetArgs([]string{"member", "create", "--display-name", "A", "--email", "not-an-email"})
	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected error")
	}
	if exitcode.Code(err) != exitcode.Usage {
		t.Fatalf("expected exit 2, got %d (%v)", exitcode.Code(err), err)
	}
	if api.createCalls != 0 {
		t.Fatalf("expected no api calls, got %d", api.createCalls)
	}
}

func TestMemberCreate_NoIdempotencyFlagAccepted_IsUsage(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	store := &memStore{path: "/x", doc: baseDoc(t)}
	api := &fakeMemberCreateAPI{}

	cmd := NewRootCmd(RootDeps{ConfigStore: store, PlannerAPI: api, Stdout: stdout, Stderr: stderr})
	cmd.SetArgs([]string{"member", "create", "--idempotency-key", "k1", "--display-name", "A", "--email", "a@example.com"})
	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected error")
	}
	if exitcode.Code(err) != exitcode.Usage {
		t.Fatalf("expected exit 2, got %d (%v)", exitcode.Code(err), err)
	}
}

func TestMemberCreate_MultilineNotesViaFlag_IsUsage_NoAPICall(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	store := &memStore{path: "/x", doc: baseDoc(t)}
	api := &fakeMemberCreateAPI{}

	cmd := NewRootCmd(RootDeps{ConfigStore: store, PlannerAPI: api, Stdout: stdout, Stderr: stderr})
	cmd.SetArgs([]string{"member", "create", "--display-name", "A", "--email", "a@example.com", "--vehicle-notes", "line1\nline2"})
	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected error")
	}
	if exitcode.Code(err) != exitcode.Usage {
		t.Fatalf("expected exit 2, got %d (%v)", exitcode.Code(err), err)
	}
	if api.createCalls != 0 {
		t.Fatalf("expected no api calls, got %d", api.createCalls)
	}
}

func TestMemberCreate_RejectsMultilineEmail(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	store := &memStore{path: "/x", doc: baseDoc(t)}
	api := &fakeMemberCreateAPI{}

	cmd := NewRootCmd(RootDeps{ConfigStore: store, PlannerAPI: api, Stdout: stdout, Stderr: stderr})
	cmd.SetArgs([]string{"member", "create", "--display-name", "A", "--email", "a@example.com\nb"})
	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected error")
	}
	if exitcode.Code(err) != exitcode.Usage {
		t.Fatalf("expected exit 2, got %d (%v)", exitcode.Code(err), err)
	}
	if api.createCalls != 0 {
		t.Fatalf("expected no api calls, got %d", api.createCalls)
	}
}

func TestMemberCreate_MissingDisplayName_IsUsage_NoAPICall(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	store := &memStore{path: "/x", doc: baseDoc(t)}
	api := &fakeMemberCreateAPI{}

	cmd := NewRootCmd(RootDeps{ConfigStore: store, PlannerAPI: api, Stdout: stdout, Stderr: stderr})
	cmd.SetArgs([]string{"member", "create", "--email", "a@example.com"})
	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected error")
	}
	if exitcode.Code(err) != exitcode.Usage {
		t.Fatalf("expected exit 2, got %d (%v)", exitcode.Code(err), err)
	}
	if api.createCalls != 0 {
		t.Fatalf("expected no api calls, got %d", api.createCalls)
	}
}

func TestMemberCreate_JSONEnvelope(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	store := &memStore{path: "/x", doc: baseDoc(t)}
	api := &fakeMemberCreateAPI{}

	cmd := NewRootCmd(RootDeps{ConfigStore: store, PlannerAPI: api, Stdout: stdout, Stderr: stderr})
	cmd.SetArgs([]string{"--output", "json", "member", "create", "--display-name", "A", "--email", "a@example.com"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if api.createCalls != 1 {
		t.Fatalf("expected 1 call, got %d", api.createCalls)
	}
	var env map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &env); err != nil {
		t.Fatalf("stdout not json: %v\n%s", err, stdout.String())
	}
	if env["data"] == nil {
		t.Fatalf("expected data")
	}
}

func TestMemberCreate_TableOutput_PrintsMemberID(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	store := &memStore{path: "/x", doc: baseDoc(t)}
	api := &fakeMemberCreateAPI{}

	cmd := NewRootCmd(RootDeps{ConfigStore: store, PlannerAPI: api, Stdout: stdout, Stderr: stderr})
	cmd.SetArgs([]string{"member", "create", "--display-name", "A", "--email", "a@example.com"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if api.createCalls != 1 {
		t.Fatalf("expected 1 call, got %d", api.createCalls)
	}
	if stdout.String() != "memberId=m1\n" {
		t.Fatalf("stdout: %q", stdout.String())
	}
}

func TestMemberCreate_SendsVehicleProfileFields(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	store := &memStore{path: "/x", doc: baseDoc(t)}
	api := &fakeMemberCreateAPI{}

	cmd := NewRootCmd(RootDeps{ConfigStore: store, PlannerAPI: api, Stdout: stdout, Stderr: stderr})
	cmd.SetArgs([]string{
		"member", "create",
		"--display-name", "A",
		"--email", "a@example.com",
		"--vehicle-make", "Toyota",
		"--vehicle-ham-radio-call-sign", "KX6ABC",
	})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if api.createCalls != 1 {
		t.Fatalf("expected 1 call, got %d", api.createCalls)
	}
	if api.lastReq.VehicleProfile == nil || api.lastReq.VehicleProfile.Make == nil || *api.lastReq.VehicleProfile.Make != "Toyota" {
		t.Fatalf("vehicle profile: %#v", api.lastReq.VehicleProfile)
	}
	if api.lastReq.VehicleProfile.HamRadioCallSign == nil || *api.lastReq.VehicleProfile.HamRadioCallSign != "KX6ABC" {
		t.Fatalf("vehicle profile: %#v", api.lastReq.VehicleProfile)
	}
}

func TestMemberCreate_TrimsInputs(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	store := &memStore{path: "/x", doc: baseDoc(t)}
	api := &fakeMemberCreateAPI{}

	cmd := NewRootCmd(RootDeps{ConfigStore: store, PlannerAPI: api, Stdout: stdout, Stderr: stderr})
	cmd.SetArgs([]string{"member", "create", "--display-name", "  A  ", "--email", "  a@example.com  "})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if api.lastReq.DisplayName != "A" {
		t.Fatalf("displayName: %q", api.lastReq.DisplayName)
	}
	if string(api.lastReq.Email) != "a@example.com" {
		t.Fatalf("email: %q", string(api.lastReq.Email))
	}
}

func TestMemberCreate_GroupAliasEmail_Validates(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	store := &memStore{path: "/x", doc: baseDoc(t)}
	api := &fakeMemberCreateAPI{}

	cmd := NewRootCmd(RootDeps{ConfigStore: store, PlannerAPI: api, Stdout: stdout, Stderr: stderr})
	cmd.SetArgs([]string{"member", "create", "--display-name", "A", "--email", "a@example.com", "--group-alias-email", "bad"})
	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected error")
	}
	if exitcode.Code(err) != exitcode.Usage {
		t.Fatalf("expected exit 2, got %d (%v)", exitcode.Code(err), err)
	}
	if api.createCalls != 0 {
		t.Fatalf("expected no api calls, got %d", api.createCalls)
	}
}

func TestMemberCreate_GroupAliasEmail_SetsField(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	store := &memStore{path: "/x", doc: baseDoc(t)}
	api := &fakeMemberCreateAPI{}

	cmd := NewRootCmd(RootDeps{ConfigStore: store, PlannerAPI: api, Stdout: stdout, Stderr: stderr})
	cmd.SetArgs([]string{"member", "create", "--display-name", "A", "--email", "a@example.com", "--group-alias-email", "ga@example.com"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if api.lastReq.GroupAliasEmail == nil || string(*api.lastReq.GroupAliasEmail) != "ga@example.com" {
		t.Fatalf("groupAliasEmail: %#v", api.lastReq.GroupAliasEmail)
	}
}

func TestMemberCreate_RejectsMultilineDisplayName(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	store := &memStore{path: "/x", doc: baseDoc(t)}
	api := &fakeMemberCreateAPI{}

	cmd := NewRootCmd(RootDeps{ConfigStore: store, PlannerAPI: api, Stdout: stdout, Stderr: stderr})
	cmd.SetArgs([]string{"member", "create", "--display-name", "a\nb", "--email", "a@example.com"})
	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected error")
	}
	if exitcode.Code(err) != exitcode.Usage {
		t.Fatalf("expected exit 2, got %d (%v)", exitcode.Code(err), err)
	}
	if api.createCalls != 0 {
		t.Fatalf("expected no api calls, got %d", api.createCalls)
	}
}

func TestMemberCreate_TableOutput_JSON201Nil_PrintsOK(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	store := &memStore{path: "/x", doc: baseDoc(t)}
	api := &fakeMemberCreateAPI{resp: &gen.CreateMyMemberClientResponse{}}

	cmd := NewRootCmd(RootDeps{ConfigStore: store, PlannerAPI: api, Stdout: stdout, Stderr: stderr})
	cmd.SetArgs([]string{"member", "create", "--display-name", "A", "--email", "a@example.com"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if stdout.String() != "OK\n" {
		t.Fatalf("stdout: %q", stdout.String())
	}
}

func TestMemberCreate_RejectsMultilineForAllStringFlags(t *testing.T) {
	cases := []struct {
		name     string
		extraArg []string
	}{
		{name: "group-alias-email", extraArg: []string{"--group-alias-email", "x\ny"}},
		{name: "vehicle-make", extraArg: []string{"--vehicle-make", "x\ny"}},
		{name: "vehicle-model", extraArg: []string{"--vehicle-model", "x\ny"}},
		{name: "vehicle-tire-size", extraArg: []string{"--vehicle-tire-size", "x\ny"}},
		{name: "vehicle-lift-lockers", extraArg: []string{"--vehicle-lift-lockers", "x\ny"}},
		{name: "vehicle-fuel-range", extraArg: []string{"--vehicle-fuel-range", "x\ny"}},
		{name: "vehicle-recovery-gear", extraArg: []string{"--vehicle-recovery-gear", "x\ny"}},
		{name: "vehicle-ham-radio-call-sign", extraArg: []string{"--vehicle-ham-radio-call-sign", "x\ny"}},
		{name: "vehicle-notes", extraArg: []string{"--vehicle-notes", "x\ny"}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			stdout := &bytes.Buffer{}
			stderr := &bytes.Buffer{}
			store := &memStore{path: "/x", doc: baseDoc(t)}
			api := &fakeMemberCreateAPI{}

			cmd := NewRootCmd(RootDeps{ConfigStore: store, PlannerAPI: api, Stdout: stdout, Stderr: stderr})
			args := []string{"member", "create", "--display-name", "A", "--email", "a@example.com"}
			args = append(args, tc.extraArg...)
			cmd.SetArgs(args)
			err := cmd.Execute()
			if err == nil {
				t.Fatalf("expected error")
			}
			if exitcode.Code(err) != exitcode.Usage {
				t.Fatalf("expected exit 2, got %d (%v)", exitcode.Code(err), err)
			}
			if api.createCalls != 0 {
				t.Fatalf("expected no api calls, got %d", api.createCalls)
			}
		})
	}
}
