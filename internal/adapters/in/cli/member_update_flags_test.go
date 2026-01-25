package cli

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/Overland-East-Bay/trip-planner-cli/internal/platform/exitcode"
)

func TestMemberUpdate_FlagsAndFromFile_IsUsage_NoAPICall(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	store := &memStore{path: "/x", doc: baseDoc(t)}
	api := &fakePlannerAPI{}

	p := writeTempFile(t, "member.yaml", "displayName: x\n")
	cmd := NewRootCmd(RootDeps{ConfigStore: store, PlannerAPI: api, Stdout: stdout, Stderr: stderr})
	cmd.SetArgs([]string{"member", "update", "--display-name", "X", "--from-file", p})
	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected error")
	}
	if exitcode.Code(err) != exitcode.Usage {
		t.Fatalf("expected exit 2, got %d (%v)", exitcode.Code(err), err)
	}
	if api.memberCalls != 0 {
		t.Fatalf("expected no api calls, got %d", api.memberCalls)
	}
}

func TestMemberUpdate_AutoGeneratesIdempotency_JSONMetaAndCall(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	store := &memStore{path: "/x", doc: baseDoc(t)}
	api := &fakePlannerAPI{}

	cmd := NewRootCmd(RootDeps{ConfigStore: store, PlannerAPI: api, Stdout: stdout, Stderr: stderr})
	cmd.SetArgs([]string{"--output", "json", "member", "update", "--display-name", "X"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if api.memberCalls != 1 {
		t.Fatalf("expected 1 api call, got %d", api.memberCalls)
	}
	if api.lastMemberIdem == "" {
		t.Fatalf("expected idempotency key")
	}

	var env map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &env); err != nil {
		t.Fatalf("stdout not json: %v\n%s", err, stdout.String())
	}
	meta, _ := env["meta"].(map[string]any)
	if meta == nil || meta["idempotencyKey"] == "" {
		t.Fatalf("meta: %#v", meta)
	}
	if meta["idempotencyKey"] != api.lastMemberIdem {
		t.Fatalf("idempotency mismatch meta=%v call=%q", meta["idempotencyKey"], api.lastMemberIdem)
	}
}

func TestMemberUpdate_ClearVehicle_SetsAllFieldsEmpty(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	store := &memStore{path: "/x", doc: baseDoc(t)}
	api := &fakePlannerAPI{}

	cmd := NewRootCmd(RootDeps{ConfigStore: store, PlannerAPI: api, Stdout: stdout, Stderr: stderr})
	cmd.SetArgs([]string{"member", "update", "--clear-vehicle", "--idempotency-key", "k1"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if api.memberCalls != 1 {
		t.Fatalf("expected 1 api call, got %d", api.memberCalls)
	}
	if api.lastMemberReq.VehicleProfile == nil || api.lastMemberReq.VehicleProfile.Make == nil || api.lastMemberReq.VehicleProfile.Notes == nil {
		t.Fatalf("vehicleProfile: %#v", api.lastMemberReq.VehicleProfile)
	}
	if *api.lastMemberReq.VehicleProfile.Make != "" || *api.lastMemberReq.VehicleProfile.Notes != "" {
		t.Fatalf("expected cleared fields, got make=%q notes=%q", *api.lastMemberReq.VehicleProfile.Make, *api.lastMemberReq.VehicleProfile.Notes)
	}
}

func TestMemberUpdate_ClearVehicle_AndVehicleMake_IsUsage_NoAPICall(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	store := &memStore{path: "/x", doc: baseDoc(t)}
	api := &fakePlannerAPI{}

	cmd := NewRootCmd(RootDeps{ConfigStore: store, PlannerAPI: api, Stdout: stdout, Stderr: stderr})
	cmd.SetArgs([]string{"member", "update", "--clear-vehicle", "--vehicle-make", "Toyota", "--idempotency-key", "k1"})
	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected error")
	}
	if exitcode.Code(err) != exitcode.Usage {
		t.Fatalf("expected exit 2, got %d (%v)", exitcode.Code(err), err)
	}
	if api.memberCalls != 0 {
		t.Fatalf("expected no api calls, got %d", api.memberCalls)
	}
}

func TestMemberUpdate_DisplayNameAndClearDisplayName_IsUsage_NoAPICall(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	store := &memStore{path: "/x", doc: baseDoc(t)}
	api := &fakePlannerAPI{}

	cmd := NewRootCmd(RootDeps{ConfigStore: store, PlannerAPI: api, Stdout: stdout, Stderr: stderr})
	cmd.SetArgs([]string{"member", "update", "--display-name", "X", "--clear-display-name"})
	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected error")
	}
	if exitcode.Code(err) != exitcode.Usage {
		t.Fatalf("expected exit 2, got %d (%v)", exitcode.Code(err), err)
	}
	if api.memberCalls != 0 {
		t.Fatalf("expected no api calls, got %d", api.memberCalls)
	}
}

func TestMemberUpdate_GroupAliasAndClearGroupAlias_IsUsage_NoAPICall(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	store := &memStore{path: "/x", doc: baseDoc(t)}
	api := &fakePlannerAPI{}

	cmd := NewRootCmd(RootDeps{ConfigStore: store, PlannerAPI: api, Stdout: stdout, Stderr: stderr})
	cmd.SetArgs([]string{"member", "update", "--group-alias-email", "ga@example.com", "--clear-group-alias-email"})
	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected error")
	}
	if exitcode.Code(err) != exitcode.Usage {
		t.Fatalf("expected exit 2, got %d (%v)", exitcode.Code(err), err)
	}
	if api.memberCalls != 0 {
		t.Fatalf("expected no api calls, got %d", api.memberCalls)
	}
}

func TestMemberUpdate_InvalidEmail_IsUsage_NoAPICall(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	store := &memStore{path: "/x", doc: baseDoc(t)}
	api := &fakePlannerAPI{}

	cmd := NewRootCmd(RootDeps{ConfigStore: store, PlannerAPI: api, Stdout: stdout, Stderr: stderr})
	cmd.SetArgs([]string{"member", "update", "--email", "not-an-email", "--idempotency-key", "k1"})
	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected error")
	}
	if exitcode.Code(err) != exitcode.Usage {
		t.Fatalf("expected exit 2, got %d (%v)", exitcode.Code(err), err)
	}
	if api.memberCalls != 0 {
		t.Fatalf("expected no api calls, got %d", api.memberCalls)
	}
}

func TestMemberUpdate_SetEmail_SetsReqField(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	store := &memStore{path: "/x", doc: baseDoc(t)}
	api := &fakePlannerAPI{}

	cmd := NewRootCmd(RootDeps{ConfigStore: store, PlannerAPI: api, Stdout: stdout, Stderr: stderr})
	cmd.SetArgs([]string{"member", "update", "--email", "a@example.com", "--idempotency-key", "k1"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if api.memberCalls != 1 {
		t.Fatalf("expected 1 api call, got %d", api.memberCalls)
	}
	if api.lastMemberReq.Email == nil || string(*api.lastMemberReq.Email) != "a@example.com" {
		t.Fatalf("email: %#v", api.lastMemberReq.Email)
	}
}

func TestMemberUpdate_ClearGroupAlias_SetsEmpty(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	store := &memStore{path: "/x", doc: baseDoc(t)}
	api := &fakePlannerAPI{}

	cmd := NewRootCmd(RootDeps{ConfigStore: store, PlannerAPI: api, Stdout: stdout, Stderr: stderr})
	cmd.SetArgs([]string{"member", "update", "--clear-group-alias-email", "--idempotency-key", "k1"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if api.memberCalls != 1 {
		t.Fatalf("expected 1 api call, got %d", api.memberCalls)
	}
	if api.lastMemberReq.GroupAliasEmail == nil || string(*api.lastMemberReq.GroupAliasEmail) != "" {
		t.Fatalf("groupAliasEmail: %#v", api.lastMemberReq.GroupAliasEmail)
	}
}

func TestMemberUpdate_ClearVehicleMake_SetsEmpty(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	store := &memStore{path: "/x", doc: baseDoc(t)}
	api := &fakePlannerAPI{}

	cmd := NewRootCmd(RootDeps{ConfigStore: store, PlannerAPI: api, Stdout: stdout, Stderr: stderr})
	cmd.SetArgs([]string{"member", "update", "--clear-vehicle-make", "--idempotency-key", "k1"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if api.lastMemberReq.VehicleProfile == nil || api.lastMemberReq.VehicleProfile.Make == nil || *api.lastMemberReq.VehicleProfile.Make != "" {
		t.Fatalf("vehicleProfile.make: %#v", api.lastMemberReq.VehicleProfile)
	}
}

func TestMemberUpdate_EditAndFlags_IsUsage_NoAPICall(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	store := &memStore{path: "/x", doc: baseDoc(t)}
	api := &fakePlannerAPI{}

	cmd := NewRootCmd(RootDeps{ConfigStore: store, PlannerAPI: api, Stdout: stdout, Stderr: stderr})
	cmd.SetArgs([]string{"member", "update", "--edit", "--display-name", "X"})
	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected error")
	}
	if exitcode.Code(err) != exitcode.Usage {
		t.Fatalf("expected exit 2, got %d (%v)", exitcode.Code(err), err)
	}
	if api.memberCalls != 0 {
		t.Fatalf("expected no api calls, got %d", api.memberCalls)
	}
}

func TestMemberUpdate_PromptAndFlags_IsUsage_NoAPICall(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	store := &memStore{path: "/x", doc: baseDoc(t)}
	api := &fakePlannerAPI{}

	cmd := NewRootCmd(RootDeps{ConfigStore: store, PlannerAPI: api, Stdout: stdout, Stderr: stderr})
	cmd.SetArgs([]string{"member", "update", "--prompt", "--display-name", "X"})
	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected error")
	}
	if exitcode.Code(err) != exitcode.Usage {
		t.Fatalf("expected exit 2, got %d (%v)", exitcode.Code(err), err)
	}
	if api.memberCalls != 0 {
		t.Fatalf("expected no api calls, got %d", api.memberCalls)
	}
}

func TestMemberUpdate_ClearDisplayName_SetsEmpty(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	store := &memStore{path: "/x", doc: baseDoc(t)}
	api := &fakePlannerAPI{}

	cmd := NewRootCmd(RootDeps{ConfigStore: store, PlannerAPI: api, Stdout: stdout, Stderr: stderr})
	cmd.SetArgs([]string{"member", "update", "--clear-display-name", "--idempotency-key", "k1"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if api.lastMemberReq.DisplayName == nil || *api.lastMemberReq.DisplayName != "" {
		t.Fatalf("displayName: %#v", api.lastMemberReq.DisplayName)
	}
}

func TestMemberUpdate_SetGroupAliasEmail_SetsField(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	store := &memStore{path: "/x", doc: baseDoc(t)}
	api := &fakePlannerAPI{}

	cmd := NewRootCmd(RootDeps{ConfigStore: store, PlannerAPI: api, Stdout: stdout, Stderr: stderr})
	cmd.SetArgs([]string{"member", "update", "--group-alias-email", "ga@example.com", "--idempotency-key", "k1"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if api.lastMemberReq.GroupAliasEmail == nil || string(*api.lastMemberReq.GroupAliasEmail) != "ga@example.com" {
		t.Fatalf("groupAliasEmail: %#v", api.lastMemberReq.GroupAliasEmail)
	}
}

func TestMemberUpdate_SetVehicleMake_SetsField(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	store := &memStore{path: "/x", doc: baseDoc(t)}
	api := &fakePlannerAPI{}

	cmd := NewRootCmd(RootDeps{ConfigStore: store, PlannerAPI: api, Stdout: stdout, Stderr: stderr})
	cmd.SetArgs([]string{"member", "update", "--vehicle-make", "Toyota", "--idempotency-key", "k1"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if api.lastMemberReq.VehicleProfile == nil || api.lastMemberReq.VehicleProfile.Make == nil || *api.lastMemberReq.VehicleProfile.Make != "Toyota" {
		t.Fatalf("vehicleProfile: %#v", api.lastMemberReq.VehicleProfile)
	}
}

func TestMemberUpdate_VehicleMakeEmpty_IsUsage_NoAPICall(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	store := &memStore{path: "/x", doc: baseDoc(t)}
	api := &fakePlannerAPI{}

	cmd := NewRootCmd(RootDeps{ConfigStore: store, PlannerAPI: api, Stdout: stdout, Stderr: stderr})
	// Use = to force the flag to be treated as "changed" with an empty value.
	cmd.SetArgs([]string{"member", "update", "--vehicle-make=", "--idempotency-key", "k1"})
	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected error")
	}
	if exitcode.Code(err) != exitcode.Usage {
		t.Fatalf("expected exit 2, got %d (%v)", exitcode.Code(err), err)
	}
	if api.memberCalls != 0 {
		t.Fatalf("expected no api calls, got %d", api.memberCalls)
	}
}

func TestMemberUpdate_MultilineVehicleNotes_IsUsage_NoAPICall(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	store := &memStore{path: "/x", doc: baseDoc(t)}
	api := &fakePlannerAPI{}

	cmd := NewRootCmd(RootDeps{ConfigStore: store, PlannerAPI: api, Stdout: stdout, Stderr: stderr})
	cmd.SetArgs([]string{"member", "update", "--vehicle-notes", "a\nb", "--idempotency-key", "k1"})
	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected error")
	}
	if exitcode.Code(err) != exitcode.Usage {
		t.Fatalf("expected exit 2, got %d (%v)", exitcode.Code(err), err)
	}
	if api.memberCalls != 0 {
		t.Fatalf("expected no api calls, got %d", api.memberCalls)
	}
}
