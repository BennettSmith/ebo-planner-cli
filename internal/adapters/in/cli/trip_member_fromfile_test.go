package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	gen "github.com/Overland-East-Bay/trip-planner-cli/internal/gen/plannerapi"
	"github.com/Overland-East-Bay/trip-planner-cli/internal/platform/config"
	"github.com/Overland-East-Bay/trip-planner-cli/internal/platform/exitcode"
	outplannerapi "github.com/Overland-East-Bay/trip-planner-cli/internal/ports/out/plannerapi"
)

type fakePlannerAPI struct {
	createCalls int
	updateCalls int
	memberCalls int

	lastCreateIdem string
	lastUpdateIdem string
	lastMemberIdem string

	lastCreateReq gen.CreateTripDraftJSONRequestBody
	lastUpdateReq gen.UpdateTripJSONRequestBody
	lastMemberReq gen.UpdateMyMemberProfileJSONRequestBody
}

var _ outplannerapi.Client = (*fakePlannerAPI)(nil)

func (f *fakePlannerAPI) CreateTripDraft(ctx context.Context, baseURL string, bearerToken string, idempotencyKey string, req gen.CreateTripDraftJSONRequestBody) (*gen.CreateTripDraftClientResponse, error) {
	_ = ctx
	_ = baseURL
	_ = bearerToken
	f.createCalls++
	f.lastCreateIdem = idempotencyKey
	f.lastCreateReq = req
	return &gen.CreateTripDraftClientResponse{JSON201: &gen.CreateTripDraftResponse{Trip: gen.TripCreated{TripId: "t1"}}}, nil
}

func (f *fakePlannerAPI) UpdateTrip(ctx context.Context, baseURL string, bearerToken string, tripID gen.TripId, idempotencyKey string, req gen.UpdateTripJSONRequestBody) (*gen.UpdateTripClientResponse, error) {
	_ = ctx
	_ = baseURL
	_ = bearerToken
	_ = tripID
	f.updateCalls++
	f.lastUpdateIdem = idempotencyKey
	f.lastUpdateReq = req
	return &gen.UpdateTripClientResponse{JSON200: &gen.TripResponse{}}, nil
}

func (f *fakePlannerAPI) CancelTrip(ctx context.Context, baseURL string, bearerToken string, tripID gen.TripId, idempotencyKey *string) (*gen.CancelTripClientResponse, error) {
	_ = ctx
	_ = baseURL
	_ = bearerToken
	_ = tripID
	_ = idempotencyKey
	return &gen.CancelTripClientResponse{}, nil
}

func (f *fakePlannerAPI) ListMembers(ctx context.Context, baseURL string, bearerToken string, params *gen.ListMembersParams) (*gen.ListMembersClientResponse, error) {
	_ = ctx
	_ = baseURL
	_ = bearerToken
	_ = params
	return &gen.ListMembersClientResponse{}, nil
}

func (f *fakePlannerAPI) UpdateMyMemberProfile(ctx context.Context, baseURL string, bearerToken string, idempotencyKey string, req gen.UpdateMyMemberProfileJSONRequestBody) (*gen.UpdateMyMemberProfileClientResponse, error) {
	_ = ctx
	_ = baseURL
	_ = bearerToken
	f.memberCalls++
	f.lastMemberIdem = idempotencyKey
	f.lastMemberReq = req
	return &gen.UpdateMyMemberProfileClientResponse{JSON200: &gen.UpdateMyMemberProfileResponse{
		Member: gen.MemberProfile{MemberId: "m1", DisplayName: "n", Email: "a@example.com"},
	}}, nil
}

func writeTempFile(t *testing.T, name string, content string) string {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte(content), 0o600); err != nil {
		t.Fatalf("write file: %v", err)
	}
	return p
}

func writeEditorScript(t *testing.T, editedContent string) string {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, "fake-editor.sh")
	script := "#!/bin/sh\n" +
		"cat > \"$1\" <<'EOF'\n" +
		editedContent +
		"\nEOF\n"
	if err := os.WriteFile(p, []byte(script), 0o700); err != nil {
		t.Fatalf("write script: %v", err)
	}
	return p
}

func baseDoc(t *testing.T) config.Document {
	t.Helper()
	doc := config.NewEmptyDocument()
	var err error
	doc, err = config.WithCurrentProfile(doc, "default")
	if err != nil {
		t.Fatalf("current profile: %v", err)
	}
	doc, err = config.WithProfileAPIURL(doc, "default", "http://api")
	if err != nil {
		t.Fatalf("api url: %v", err)
	}
	doc, err = config.SetString(doc, "profiles.default.auth.accessToken", "tok")
	if err != nil {
		t.Fatalf("token: %v", err)
	}
	return doc
}

func TestTripCreate_FromFile_ParseError_NoAPICall_Exit2(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	store := &memStore{path: "/x", doc: baseDoc(t)}
	api := &fakePlannerAPI{}

	cmd := NewRootCmd(RootDeps{ConfigStore: store, PlannerAPI: api, Stdout: stdout, Stderr: stderr})
	p := writeTempFile(t, "bad.json", "{")
	cmd.SetArgs([]string{"trip", "create", "--from-file", p})
	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected error")
	}
	if exitcode.Code(err) != exitcode.Usage {
		t.Fatalf("expected exit 2, got %d (%v)", exitcode.Code(err), err)
	}
	if api.createCalls != 0 {
		t.Fatalf("expected no API calls, got %d", api.createCalls)
	}
}

func TestTripUpdate_FromFile_ParseError_NoAPICall_Exit2(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	store := &memStore{path: "/x", doc: baseDoc(t)}
	api := &fakePlannerAPI{}

	cmd := NewRootCmd(RootDeps{ConfigStore: store, PlannerAPI: api, Stdout: stdout, Stderr: stderr})
	p := writeTempFile(t, "bad.yaml", ":\n- nope")
	cmd.SetArgs([]string{"trip", "update", "t1", "--from-file", p})
	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected error")
	}
	if exitcode.Code(err) != exitcode.Usage {
		t.Fatalf("expected exit 2, got %d (%v)", exitcode.Code(err), err)
	}
	if api.updateCalls != 0 {
		t.Fatalf("expected no API calls, got %d", api.updateCalls)
	}
}

func TestMemberUpdate_FromFile_ParseError_NoAPICall_Exit2(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	store := &memStore{path: "/x", doc: baseDoc(t)}
	api := &fakePlannerAPI{}

	cmd := NewRootCmd(RootDeps{ConfigStore: store, PlannerAPI: api, Stdout: stdout, Stderr: stderr})
	p := writeTempFile(t, "bad.json", "{")
	cmd.SetArgs([]string{"member", "update", "--from-file", p})
	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected error")
	}
	if exitcode.Code(err) != exitcode.Usage {
		t.Fatalf("expected exit 2, got %d (%v)", exitcode.Code(err), err)
	}
	if api.memberCalls != 0 {
		t.Fatalf("expected no API calls, got %d", api.memberCalls)
	}
}

func TestTripCreate_FromFile_MapsToRequestStruct(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	store := &memStore{path: "/x", doc: baseDoc(t)}
	api := &fakePlannerAPI{}

	cmd := NewRootCmd(RootDeps{ConfigStore: store, PlannerAPI: api, Stdout: stdout, Stderr: stderr})
	p := writeTempFile(t, "trip.yaml", "name: Trip From File\n")
	cmd.SetArgs([]string{"trip", "create", "--from-file", p, "--idempotency-key", "k1"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if api.createCalls != 1 {
		t.Fatalf("expected 1 API call, got %d", api.createCalls)
	}
	if api.lastCreateReq.Name != "Trip From File" {
		t.Fatalf("name: %q", api.lastCreateReq.Name)
	}
}

func TestTripCreate_Prompt_MapsToRequestStruct(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	store := &memStore{path: "/x", doc: baseDoc(t)}
	api := &fakePlannerAPI{}

	cmd := NewRootCmd(RootDeps{ConfigStore: store, PlannerAPI: api, Stdout: stdout, Stderr: stderr})
	cmd.SetIn(bytes.NewBufferString("My Trip\n"))
	cmd.SetArgs([]string{"trip", "create", "--prompt", "--idempotency-key", "k1"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if api.createCalls != 1 {
		t.Fatalf("expected 1 API call, got %d", api.createCalls)
	}
	if api.lastCreateReq.Name != "My Trip" {
		t.Fatalf("name: %q", api.lastCreateReq.Name)
	}
}

func TestTripCreate_Prompt_Aborted_NoAPICall_Exit130(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	store := &memStore{path: "/x", doc: baseDoc(t)}
	api := &fakePlannerAPI{}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	cmd := NewRootCmd(RootDeps{ConfigStore: store, PlannerAPI: api, Stdout: stdout, Stderr: stderr})
	cmd.SetContext(ctx)
	cmd.SetIn(bytes.NewBufferString("anything\n"))
	cmd.SetArgs([]string{"trip", "create", "--prompt"})
	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected error")
	}
	if exitcode.Code(err) != 130 {
		t.Fatalf("expected exit 130, got %d (%v)", exitcode.Code(err), err)
	}
	if api.createCalls != 0 {
		t.Fatalf("expected no API calls, got %d", api.createCalls)
	}
}

func TestTripUpdate_Prompt_MapsToRequestStruct(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	store := &memStore{path: "/x", doc: baseDoc(t)}
	api := &fakePlannerAPI{}

	// description inline, then artifactIds list, then blank to finish list
	cmd := NewRootCmd(RootDeps{ConfigStore: store, PlannerAPI: api, Stdout: stdout, Stderr: stderr})
	cmd.SetIn(bytes.NewBufferString("desc line\nid1\nid2\n\n"))
	cmd.SetArgs([]string{"trip", "update", "t1", "--prompt", "--idempotency-key", "k1"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if api.updateCalls != 1 {
		t.Fatalf("expected 1 API call, got %d", api.updateCalls)
	}
	if api.lastUpdateReq.Description == nil || *api.lastUpdateReq.Description != "desc line" {
		if api.lastUpdateReq.Description == nil {
			t.Fatalf("expected description set")
		}
		t.Fatalf("description: %#v", *api.lastUpdateReq.Description)
	}
	if api.lastUpdateReq.ArtifactIds == nil || len(*api.lastUpdateReq.ArtifactIds) != 2 {
		t.Fatalf("artifactIds: %#v", api.lastUpdateReq.ArtifactIds)
	}
	if (*api.lastUpdateReq.ArtifactIds)[0] != "id1" || (*api.lastUpdateReq.ArtifactIds)[1] != "id2" {
		t.Fatalf("artifactIds: %#v", *api.lastUpdateReq.ArtifactIds)
	}
}

func TestTripUpdate_Prompt_EditorPath_ParsesEditedContent(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	store := &memStore{path: "/x", doc: baseDoc(t)}
	api := &fakePlannerAPI{}

	editor := writeEditorScript(t, "text: |-\n  hi\n  there")
	if err := os.Setenv("EBO_EDITOR", editor); err != nil {
		t.Fatalf("setenv: %v", err)
	}
	t.Cleanup(func() { _ = os.Unsetenv("EBO_EDITOR") })

	// blank line triggers editor for description, then blank ends artifactIds list
	cmd := NewRootCmd(RootDeps{ConfigStore: store, PlannerAPI: api, Stdout: stdout, Stderr: stderr})
	cmd.SetIn(bytes.NewBufferString("\n\n"))
	cmd.SetArgs([]string{"trip", "update", "t1", "--prompt", "--idempotency-key", "k1"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if api.updateCalls != 1 {
		t.Fatalf("expected 1 API call, got %d", api.updateCalls)
	}
	if api.lastUpdateReq.Description == nil || *api.lastUpdateReq.Description != "hi\nthere" {
		if api.lastUpdateReq.Description == nil {
			t.Fatalf("expected description set")
		}
		t.Fatalf("description: %#v", *api.lastUpdateReq.Description)
	}
}

func TestTripUpdate_Prompt_AndFromFile_IsUsage_NoAPICall(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	store := &memStore{path: "/x", doc: baseDoc(t)}
	api := &fakePlannerAPI{}

	p := writeTempFile(t, "patch.yaml", "description: x\n")
	cmd := NewRootCmd(RootDeps{ConfigStore: store, PlannerAPI: api, Stdout: stdout, Stderr: stderr})
	cmd.SetArgs([]string{"trip", "update", "t1", "--prompt", "--from-file", p})
	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected error")
	}
	if exitcode.Code(err) != exitcode.Usage {
		t.Fatalf("expected exit 2, got %d (%v)", exitcode.Code(err), err)
	}
	if api.updateCalls != 0 {
		t.Fatalf("expected no API calls, got %d", api.updateCalls)
	}
}

func TestTripUpdate_NoMode_IsUsage_NoAPICall(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	store := &memStore{path: "/x", doc: baseDoc(t)}
	api := &fakePlannerAPI{}

	cmd := NewRootCmd(RootDeps{ConfigStore: store, PlannerAPI: api, Stdout: stdout, Stderr: stderr})
	cmd.SetArgs([]string{"trip", "update", "t1", "--idempotency-key", "k1"})
	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected error")
	}
	if exitcode.Code(err) != exitcode.Usage {
		t.Fatalf("expected exit 2, got %d (%v)", exitcode.Code(err), err)
	}
	if api.updateCalls != 0 {
		t.Fatalf("expected no API calls, got %d", api.updateCalls)
	}
}

func TestTripUpdate_EditAndPrompt_IsUsage_NoAPICall(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	store := &memStore{path: "/x", doc: baseDoc(t)}
	api := &fakePlannerAPI{}

	cmd := NewRootCmd(RootDeps{ConfigStore: store, PlannerAPI: api, Stdout: stdout, Stderr: stderr})
	cmd.SetArgs([]string{"trip", "update", "t1", "--edit", "--prompt"})
	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected error")
	}
	if exitcode.Code(err) != exitcode.Usage {
		t.Fatalf("expected exit 2, got %d (%v)", exitcode.Code(err), err)
	}
	if api.updateCalls != 0 {
		t.Fatalf("expected no API calls, got %d", api.updateCalls)
	}
}

func TestTripUpdate_Prompt_Aborted_NoAPICall_Exit130(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	store := &memStore{path: "/x", doc: baseDoc(t)}
	api := &fakePlannerAPI{}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	cmd := NewRootCmd(RootDeps{ConfigStore: store, PlannerAPI: api, Stdout: stdout, Stderr: stderr})
	cmd.SetContext(ctx)
	cmd.SetIn(bytes.NewBufferString("anything\n"))
	cmd.SetArgs([]string{"trip", "update", "t1", "--prompt"})
	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected error")
	}
	if exitcode.Code(err) != 130 {
		t.Fatalf("expected exit 130, got %d (%v)", exitcode.Code(err), err)
	}
	if api.updateCalls != 0 {
		t.Fatalf("expected no API calls, got %d", api.updateCalls)
	}
}

func TestMemberUpdate_Prompt_MapsToRequestStruct(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	store := &memStore{path: "/x", doc: baseDoc(t)}
	api := &fakePlannerAPI{}

	// displayName, then configure vehicle? (blank -> default no)
	cmd := NewRootCmd(RootDeps{ConfigStore: store, PlannerAPI: api, Stdout: stdout, Stderr: stderr})
	cmd.SetIn(bytes.NewBufferString("New Name\n\n"))
	cmd.SetArgs([]string{"member", "update", "--prompt", "--idempotency-key", "k1"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if api.memberCalls != 1 {
		t.Fatalf("expected 1 API call, got %d", api.memberCalls)
	}
	if api.lastMemberReq.DisplayName == nil || *api.lastMemberReq.DisplayName != "New Name" {
		if api.lastMemberReq.DisplayName == nil {
			t.Fatalf("expected displayName set")
		}
		t.Fatalf("displayName: %#v", *api.lastMemberReq.DisplayName)
	}
	if api.lastMemberReq.VehicleProfile != nil {
		t.Fatalf("expected vehicleProfile unset, got %#v", api.lastMemberReq.VehicleProfile)
	}
}

func TestMemberUpdate_Prompt_VehicleNotesInline_MapsToRequestStruct(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	store := &memStore{path: "/x", doc: baseDoc(t)}
	api := &fakePlannerAPI{}

	// displayName, configure vehicle yes, notes inline
	cmd := NewRootCmd(RootDeps{ConfigStore: store, PlannerAPI: api, Stdout: stdout, Stderr: stderr})
	cmd.SetIn(bytes.NewBufferString("New Name\ny\nnote\n"))
	cmd.SetArgs([]string{"member", "update", "--prompt", "--idempotency-key", "k1"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if api.memberCalls != 1 {
		t.Fatalf("expected 1 API call, got %d", api.memberCalls)
	}
	if api.lastMemberReq.VehicleProfile == nil || api.lastMemberReq.VehicleProfile.Notes == nil {
		t.Fatalf("expected vehicleProfile.notes set")
	}
	if *api.lastMemberReq.VehicleProfile.Notes != "note" {
		t.Fatalf("notes: %#v", *api.lastMemberReq.VehicleProfile.Notes)
	}
}

func TestMemberUpdate_Prompt_VehicleNotesEditor_ParsesEditedContent(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	store := &memStore{path: "/x", doc: baseDoc(t)}
	api := &fakePlannerAPI{}

	editor := writeEditorScript(t, "text: |-\n  n1\n  n2")
	if err := os.Setenv("EBO_EDITOR", editor); err != nil {
		t.Fatalf("setenv: %v", err)
	}
	t.Cleanup(func() { _ = os.Unsetenv("EBO_EDITOR") })

	// displayName blank, configure vehicle yes, notes blank -> editor
	cmd := NewRootCmd(RootDeps{ConfigStore: store, PlannerAPI: api, Stdout: stdout, Stderr: stderr})
	cmd.SetIn(bytes.NewBufferString("\ny\n\n"))
	cmd.SetArgs([]string{"member", "update", "--prompt", "--idempotency-key", "k1"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if api.memberCalls != 1 {
		t.Fatalf("expected 1 API call, got %d", api.memberCalls)
	}
	if api.lastMemberReq.VehicleProfile == nil || api.lastMemberReq.VehicleProfile.Notes == nil {
		t.Fatalf("expected vehicleProfile.notes set")
	}
	if *api.lastMemberReq.VehicleProfile.Notes != "n1\nn2" {
		t.Fatalf("notes: %#v", *api.lastMemberReq.VehicleProfile.Notes)
	}
}

func TestMemberUpdate_Prompt_AndFromFile_IsUsage_NoAPICall(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	store := &memStore{path: "/x", doc: baseDoc(t)}
	api := &fakePlannerAPI{}

	p := writeTempFile(t, "member.yaml", "displayName: x\n")
	cmd := NewRootCmd(RootDeps{ConfigStore: store, PlannerAPI: api, Stdout: stdout, Stderr: stderr})
	cmd.SetArgs([]string{"member", "update", "--prompt", "--from-file", p})
	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected error")
	}
	if exitcode.Code(err) != exitcode.Usage {
		t.Fatalf("expected exit 2, got %d (%v)", exitcode.Code(err), err)
	}
	if api.memberCalls != 0 {
		t.Fatalf("expected no API calls, got %d", api.memberCalls)
	}
}

func TestMemberUpdate_NoMode_IsUsage_NoAPICall(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	store := &memStore{path: "/x", doc: baseDoc(t)}
	api := &fakePlannerAPI{}

	cmd := NewRootCmd(RootDeps{ConfigStore: store, PlannerAPI: api, Stdout: stdout, Stderr: stderr})
	cmd.SetArgs([]string{"member", "update", "--idempotency-key", "k1"})
	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected error")
	}
	if exitcode.Code(err) != exitcode.Usage {
		t.Fatalf("expected exit 2, got %d (%v)", exitcode.Code(err), err)
	}
	if api.memberCalls != 0 {
		t.Fatalf("expected no API calls, got %d", api.memberCalls)
	}
}

func TestMemberUpdate_EditAndPrompt_IsUsage_NoAPICall(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	store := &memStore{path: "/x", doc: baseDoc(t)}
	api := &fakePlannerAPI{}

	cmd := NewRootCmd(RootDeps{ConfigStore: store, PlannerAPI: api, Stdout: stdout, Stderr: stderr})
	cmd.SetArgs([]string{"member", "update", "--edit", "--prompt"})
	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected error")
	}
	if exitcode.Code(err) != exitcode.Usage {
		t.Fatalf("expected exit 2, got %d (%v)", exitcode.Code(err), err)
	}
	if api.memberCalls != 0 {
		t.Fatalf("expected no API calls, got %d", api.memberCalls)
	}
}

func TestMemberUpdate_Prompt_Aborted_NoAPICall_Exit130(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	store := &memStore{path: "/x", doc: baseDoc(t)}
	api := &fakePlannerAPI{}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	cmd := NewRootCmd(RootDeps{ConfigStore: store, PlannerAPI: api, Stdout: stdout, Stderr: stderr})
	cmd.SetContext(ctx)
	cmd.SetIn(bytes.NewBufferString("anything\n"))
	cmd.SetArgs([]string{"member", "update", "--prompt"})
	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected error")
	}
	if exitcode.Code(err) != 130 {
		t.Fatalf("expected exit 130, got %d (%v)", exitcode.Code(err), err)
	}
	if api.memberCalls != 0 {
		t.Fatalf("expected no API calls, got %d", api.memberCalls)
	}
}

func TestTripCreate_NameFlag_MapsToRequestStruct(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	store := &memStore{path: "/x", doc: baseDoc(t)}
	api := &fakePlannerAPI{}

	cmd := NewRootCmd(RootDeps{ConfigStore: store, PlannerAPI: api, Stdout: stdout, Stderr: stderr})
	cmd.SetArgs([]string{"trip", "create", "--name", "Trip From Flag", "--idempotency-key", "k1"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if api.createCalls != 1 {
		t.Fatalf("expected 1 API call, got %d", api.createCalls)
	}
	if api.lastCreateReq.Name != "Trip From Flag" {
		t.Fatalf("name: %q", api.lastCreateReq.Name)
	}
	if api.lastCreateIdem != "k1" {
		t.Fatalf("idempotency: %q", api.lastCreateIdem)
	}
}

func TestTripCreate_BothNameAndFromFile_IsUsage_NoAPICall(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	store := &memStore{path: "/x", doc: baseDoc(t)}
	api := &fakePlannerAPI{}

	cmd := NewRootCmd(RootDeps{ConfigStore: store, PlannerAPI: api, Stdout: stdout, Stderr: stderr})
	p := writeTempFile(t, "trip.yaml", "name: Trip From File\n")
	cmd.SetArgs([]string{"trip", "create", "--name", "x", "--from-file", p})
	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected error")
	}
	if exitcode.Code(err) != exitcode.Usage {
		t.Fatalf("expected exit 2, got %d", exitcode.Code(err))
	}
	if api.createCalls != 0 {
		t.Fatalf("expected no API calls, got %d", api.createCalls)
	}
}

func TestTripCreate_AutoGeneratesIdempotencyKey_WhenOmitted(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	store := &memStore{path: "/x", doc: baseDoc(t)}
	api := &fakePlannerAPI{}

	cmd := NewRootCmd(RootDeps{ConfigStore: store, PlannerAPI: api, Stdout: stdout, Stderr: stderr})
	cmd.SetArgs([]string{"trip", "create", "--name", "Trip", "--output", "table"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if api.createCalls != 1 {
		t.Fatalf("expected 1 API call, got %d", api.createCalls)
	}
	if api.lastCreateIdem == "" {
		t.Fatalf("expected idempotency key generated")
	}
	if !bytes.Contains(stderr.Bytes(), []byte("Idempotency-Key: ")) {
		t.Fatalf("expected idempotency printed to stderr, got: %q", stderr.String())
	}
}

func TestTripUpdate_FromFile_Success_MapsToRequestStruct(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	store := &memStore{path: "/x", doc: baseDoc(t)}
	api := &fakePlannerAPI{}

	cmd := NewRootCmd(RootDeps{ConfigStore: store, PlannerAPI: api, Stdout: stdout, Stderr: stderr})
	p := writeTempFile(t, "patch.yaml", "description: |\n  a\n  b\n")
	cmd.SetArgs([]string{"trip", "update", "t1", "--from-file", p, "--idempotency-key", "k1"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if api.updateCalls != 1 {
		t.Fatalf("expected 1 API call, got %d", api.updateCalls)
	}
	if api.lastUpdateIdem != "k1" {
		t.Fatalf("idempotency: %q", api.lastUpdateIdem)
	}
	if api.lastUpdateReq.Description == nil || *api.lastUpdateReq.Description != "a\nb\n" {
		if api.lastUpdateReq.Description == nil {
			t.Fatalf("expected description set")
		}
		t.Fatalf("description: %#v", *api.lastUpdateReq.Description)
	}
}

func TestMemberUpdate_FromFile_Success_MapsToRequestStruct(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	store := &memStore{path: "/x", doc: baseDoc(t)}
	api := &fakePlannerAPI{}

	cmd := NewRootCmd(RootDeps{ConfigStore: store, PlannerAPI: api, Stdout: stdout, Stderr: stderr})
	p := writeTempFile(t, "member.yaml", "displayName: New Name\n")
	cmd.SetArgs([]string{"member", "update", "--from-file", p, "--idempotency-key", "k1"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if api.memberCalls != 1 {
		t.Fatalf("expected 1 API call, got %d", api.memberCalls)
	}
	if api.lastMemberIdem != "k1" {
		t.Fatalf("idempotency: %q", api.lastMemberIdem)
	}
	if api.lastMemberReq.DisplayName == nil || *api.lastMemberReq.DisplayName != "New Name" {
		if api.lastMemberReq.DisplayName == nil {
			t.Fatalf("expected displayName set")
		}
		t.Fatalf("displayName: %#v", *api.lastMemberReq.DisplayName)
	}
}

func TestTripCreate_MissingAPIURL_IsUsage(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	doc := config.NewEmptyDocument()
	doc, _ = config.WithCurrentProfile(doc, "default")
	// no apiUrl configured
	doc, _ = config.SetString(doc, "profiles.default.auth.accessToken", "tok")
	store := &memStore{path: "/x", doc: doc}
	api := &fakePlannerAPI{}

	cmd := NewRootCmd(RootDeps{ConfigStore: store, PlannerAPI: api, Stdout: stdout, Stderr: stderr})
	cmd.SetArgs([]string{"trip", "create", "--name", "Trip", "--idempotency-key", "k1"})
	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected error")
	}
	if exitcode.Code(err) != exitcode.Usage {
		t.Fatalf("expected exit 2, got %d (%v)", exitcode.Code(err), err)
	}
	if api.createCalls != 0 {
		t.Fatalf("expected no API calls, got %d", api.createCalls)
	}
}

func TestTripCreate_MissingToken_IsAuth(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	doc := config.NewEmptyDocument()
	doc, _ = config.WithCurrentProfile(doc, "default")
	doc, _ = config.WithProfileAPIURL(doc, "default", "http://api")
	// no token configured
	store := &memStore{path: "/x", doc: doc}
	api := &fakePlannerAPI{}

	cmd := NewRootCmd(RootDeps{ConfigStore: store, PlannerAPI: api, Stdout: stdout, Stderr: stderr})
	cmd.SetArgs([]string{"trip", "create", "--name", "Trip", "--idempotency-key", "k1"})
	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected error")
	}
	if exitcode.Code(err) != exitcode.Auth {
		t.Fatalf("expected exit 3, got %d (%v)", exitcode.Code(err), err)
	}
	if api.createCalls != 0 {
		t.Fatalf("expected no API calls, got %d", api.createCalls)
	}
}

func TestTripUpdate_Edit_InvalidBuffer_IsUsage_NoAPICall(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	store := &memStore{path: "/x", doc: baseDoc(t)}
	api := &fakePlannerAPI{}

	editor := writeEditorScript(t, "{")
	if err := os.Setenv("EBO_EDITOR", editor); err != nil {
		t.Fatalf("setenv: %v", err)
	}
	t.Cleanup(func() { _ = os.Unsetenv("EBO_EDITOR") })

	cmd := NewRootCmd(RootDeps{ConfigStore: store, PlannerAPI: api, Stdout: stdout, Stderr: stderr})
	cmd.SetArgs([]string{"trip", "update", "t1", "--edit"})
	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected error")
	}
	if exitcode.Code(err) != exitcode.Usage {
		t.Fatalf("expected exit 2, got %d (%v)", exitcode.Code(err), err)
	}
	if api.updateCalls != 0 {
		t.Fatalf("expected no API calls, got %d", api.updateCalls)
	}
}

func TestTripUpdate_Edit_ParsesEditedContent(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	store := &memStore{path: "/x", doc: baseDoc(t)}
	api := &fakePlannerAPI{}

	editor := writeEditorScript(t, "description: hi\n")
	if err := os.Setenv("EBO_EDITOR", editor); err != nil {
		t.Fatalf("setenv: %v", err)
	}
	t.Cleanup(func() { _ = os.Unsetenv("EBO_EDITOR") })

	cmd := NewRootCmd(RootDeps{ConfigStore: store, PlannerAPI: api, Stdout: stdout, Stderr: stderr})
	cmd.SetArgs([]string{"trip", "update", "t1", "--edit", "--idempotency-key", "k1"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if api.updateCalls != 1 {
		t.Fatalf("expected 1 API call, got %d", api.updateCalls)
	}
	if api.lastUpdateReq.Description == nil || *api.lastUpdateReq.Description != "hi" {
		if api.lastUpdateReq.Description == nil {
			t.Fatalf("expected description set")
		}
		t.Fatalf("description: %#v", *api.lastUpdateReq.Description)
	}
}

func TestTripUpdate_Edit_AndFromFile_IsUsage_NoAPICall(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	store := &memStore{path: "/x", doc: baseDoc(t)}
	api := &fakePlannerAPI{}

	editor := writeEditorScript(t, "description: hi\n")
	if err := os.Setenv("EBO_EDITOR", editor); err != nil {
		t.Fatalf("setenv: %v", err)
	}
	t.Cleanup(func() { _ = os.Unsetenv("EBO_EDITOR") })

	p := writeTempFile(t, "patch.yaml", "description: x\n")
	cmd := NewRootCmd(RootDeps{ConfigStore: store, PlannerAPI: api, Stdout: stdout, Stderr: stderr})
	cmd.SetArgs([]string{"trip", "update", "t1", "--edit", "--from-file", p})
	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected error")
	}
	if exitcode.Code(err) != exitcode.Usage {
		t.Fatalf("expected exit 2, got %d (%v)", exitcode.Code(err), err)
	}
	if api.updateCalls != 0 {
		t.Fatalf("expected no API calls, got %d", api.updateCalls)
	}
}

func TestMemberUpdate_Edit_InvalidBuffer_IsUsage_NoAPICall(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	store := &memStore{path: "/x", doc: baseDoc(t)}
	api := &fakePlannerAPI{}

	editor := writeEditorScript(t, "{")
	if err := os.Setenv("EBO_EDITOR", editor); err != nil {
		t.Fatalf("setenv: %v", err)
	}
	t.Cleanup(func() { _ = os.Unsetenv("EBO_EDITOR") })

	cmd := NewRootCmd(RootDeps{ConfigStore: store, PlannerAPI: api, Stdout: stdout, Stderr: stderr})
	cmd.SetArgs([]string{"member", "update", "--edit"})
	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected error")
	}
	if exitcode.Code(err) != exitcode.Usage {
		t.Fatalf("expected exit 2, got %d (%v)", exitcode.Code(err), err)
	}
	if api.memberCalls != 0 {
		t.Fatalf("expected no API calls, got %d", api.memberCalls)
	}
}

func TestMemberUpdate_Edit_ParsesEditedContent(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	store := &memStore{path: "/x", doc: baseDoc(t)}
	api := &fakePlannerAPI{}

	editor := writeEditorScript(t, "displayName: New Name\n")
	if err := os.Setenv("EBO_EDITOR", editor); err != nil {
		t.Fatalf("setenv: %v", err)
	}
	t.Cleanup(func() { _ = os.Unsetenv("EBO_EDITOR") })

	cmd := NewRootCmd(RootDeps{ConfigStore: store, PlannerAPI: api, Stdout: stdout, Stderr: stderr})
	cmd.SetArgs([]string{"member", "update", "--edit", "--idempotency-key", "k1"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if api.memberCalls != 1 {
		t.Fatalf("expected 1 API call, got %d", api.memberCalls)
	}
	if api.lastMemberReq.DisplayName == nil || *api.lastMemberReq.DisplayName != "New Name" {
		if api.lastMemberReq.DisplayName == nil {
			t.Fatalf("expected displayName set")
		}
		t.Fatalf("displayName: %#v", *api.lastMemberReq.DisplayName)
	}
}

func TestMemberUpdate_Edit_AndFromFile_IsUsage_NoAPICall(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	store := &memStore{path: "/x", doc: baseDoc(t)}
	api := &fakePlannerAPI{}

	editor := writeEditorScript(t, "displayName: New Name\n")
	if err := os.Setenv("EBO_EDITOR", editor); err != nil {
		t.Fatalf("setenv: %v", err)
	}
	t.Cleanup(func() { _ = os.Unsetenv("EBO_EDITOR") })

	p := writeTempFile(t, "member.yaml", "displayName: x\n")
	cmd := NewRootCmd(RootDeps{ConfigStore: store, PlannerAPI: api, Stdout: stdout, Stderr: stderr})
	cmd.SetArgs([]string{"member", "update", "--edit", "--from-file", p})
	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected error")
	}
	if exitcode.Code(err) != exitcode.Usage {
		t.Fatalf("expected exit 2, got %d (%v)", exitcode.Code(err), err)
	}
	if api.memberCalls != 0 {
		t.Fatalf("expected no API calls, got %d", api.memberCalls)
	}
}

func TestTripUpdate_MissingFromFile_IsUsage_NoAPICall(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	store := &memStore{path: "/x", doc: baseDoc(t)}
	api := &fakePlannerAPI{}

	cmd := NewRootCmd(RootDeps{ConfigStore: store, PlannerAPI: api, Stdout: stdout, Stderr: stderr})
	cmd.SetArgs([]string{"trip", "update", "t1"})
	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected error")
	}
	if exitcode.Code(err) != exitcode.Usage {
		t.Fatalf("expected exit 2, got %d", exitcode.Code(err))
	}
	if api.updateCalls != 0 {
		t.Fatalf("expected no API calls, got %d", api.updateCalls)
	}
}

func TestMemberUpdate_MissingFromFile_IsUsage_NoAPICall(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	store := &memStore{path: "/x", doc: baseDoc(t)}
	api := &fakePlannerAPI{}

	cmd := NewRootCmd(RootDeps{ConfigStore: store, PlannerAPI: api, Stdout: stdout, Stderr: stderr})
	cmd.SetArgs([]string{"member", "update"})
	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected error")
	}
	if exitcode.Code(err) != exitcode.Usage {
		t.Fatalf("expected exit 2, got %d", exitcode.Code(err))
	}
	if api.memberCalls != 0 {
		t.Fatalf("expected no API calls, got %d", api.memberCalls)
	}
}

func TestTripCreate_JSONOutput_IncludesMetaIdempotencyKey(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	store := &memStore{path: "/x", doc: baseDoc(t)}
	api := &fakePlannerAPI{}

	cmd := NewRootCmd(RootDeps{ConfigStore: store, PlannerAPI: api, Stdout: stdout, Stderr: stderr})
	cmd.SetArgs([]string{"--output", "json", "trip", "create", "--name", "Trip", "--idempotency-key", "k1"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}

	var env map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &env); err != nil {
		t.Fatalf("stdout not json: %v\n%s", err, stdout.String())
	}
	meta := env["meta"].(map[string]any)
	if meta["idempotencyKey"] != "k1" {
		t.Fatalf("idempotencyKey: %#v", meta["idempotencyKey"])
	}
}
