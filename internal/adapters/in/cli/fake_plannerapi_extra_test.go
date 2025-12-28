package cli

import (
	"context"

	gen "github.com/BennettSmith/ebo-planner-cli/internal/gen/plannerapi"
	"github.com/BennettSmith/ebo-planner-cli/internal/platform/exitcode"
)

// Satisfy newly-added plannerapi.Client methods for existing tests that use fakePlannerAPI.
func (f *fakePlannerAPI) ListVisibleTripsForMember(ctx context.Context, baseURL string, bearerToken string) (*gen.ListVisibleTripsForMemberClientResponse, error) {
	_ = ctx
	_ = baseURL
	_ = bearerToken
	return nil, exitcode.New(exitcode.KindUnexpected, "not implemented in this test fake", nil)
}

func (f *fakePlannerAPI) ListMyDraftTrips(ctx context.Context, baseURL string, bearerToken string) (*gen.ListMyDraftTripsClientResponse, error) {
	_ = ctx
	_ = baseURL
	_ = bearerToken
	return nil, exitcode.New(exitcode.KindUnexpected, "not implemented in this test fake", nil)
}

func (f *fakePlannerAPI) GetTripDetails(ctx context.Context, baseURL string, bearerToken string, tripID gen.TripId) (*gen.GetTripDetailsClientResponse, error) {
	_ = ctx
	_ = baseURL
	_ = bearerToken
	_ = tripID
	return nil, exitcode.New(exitcode.KindUnexpected, "not implemented in this test fake", nil)
}

func (f *fakePlannerAPI) SetTripDraftVisibility(ctx context.Context, baseURL string, bearerToken string, tripID gen.TripId, idempotencyKey string, req gen.SetTripDraftVisibilityJSONRequestBody) (*gen.SetTripDraftVisibilityClientResponse, error) {
	_ = ctx
	_ = baseURL
	_ = bearerToken
	_ = tripID
	_ = idempotencyKey
	_ = req
	return nil, exitcode.New(exitcode.KindUnexpected, "not implemented in this test fake", nil)
}

func (f *fakePlannerAPI) PublishTrip(ctx context.Context, baseURL string, bearerToken string, tripID gen.TripId) (*gen.PublishTripClientResponse, error) {
	_ = ctx
	_ = baseURL
	_ = bearerToken
	_ = tripID
	return nil, exitcode.New(exitcode.KindUnexpected, "not implemented in this test fake", nil)
}

func (f *fakePlannerAPI) AddTripOrganizer(ctx context.Context, baseURL string, bearerToken string, tripID gen.TripId, idempotencyKey string, req gen.AddTripOrganizerJSONRequestBody) (*gen.AddTripOrganizerClientResponse, error) {
	_ = ctx
	_ = baseURL
	_ = bearerToken
	_ = tripID
	_ = idempotencyKey
	_ = req
	return nil, exitcode.New(exitcode.KindUnexpected, "not implemented in this test fake", nil)
}

func (f *fakePlannerAPI) RemoveTripOrganizer(ctx context.Context, baseURL string, bearerToken string, tripID gen.TripId, memberID gen.MemberId, idempotencyKey string) (*gen.RemoveTripOrganizerClientResponse, error) {
	_ = ctx
	_ = baseURL
	_ = bearerToken
	_ = tripID
	_ = memberID
	_ = idempotencyKey
	return nil, exitcode.New(exitcode.KindUnexpected, "not implemented in this test fake", nil)
}

func (f *fakePlannerAPI) SetMyRSVP(ctx context.Context, baseURL string, bearerToken string, tripID gen.TripId, idempotencyKey string, req gen.SetMyRSVPJSONRequestBody) (*gen.SetMyRSVPClientResponse, error) {
	_ = ctx
	_ = baseURL
	_ = bearerToken
	_ = tripID
	_ = idempotencyKey
	_ = req
	return nil, exitcode.New(exitcode.KindUnexpected, "not implemented in this test fake", nil)
}

func (f *fakePlannerAPI) GetMyRSVPForTrip(ctx context.Context, baseURL string, bearerToken string, tripID gen.TripId) (*gen.GetMyRSVPForTripClientResponse, error) {
	_ = ctx
	_ = baseURL
	_ = bearerToken
	_ = tripID
	return nil, exitcode.New(exitcode.KindUnexpected, "not implemented in this test fake", nil)
}

func (f *fakePlannerAPI) GetTripRSVPSummary(ctx context.Context, baseURL string, bearerToken string, tripID gen.TripId) (*gen.GetTripRSVPSummaryClientResponse, error) {
	_ = ctx
	_ = baseURL
	_ = bearerToken
	_ = tripID
	return nil, exitcode.New(exitcode.KindUnexpected, "not implemented in this test fake", nil)
}

func (f *fakePlannerAPI) SearchMembers(ctx context.Context, baseURL string, bearerToken string, params *gen.SearchMembersParams) (*gen.SearchMembersClientResponse, error) {
	_ = ctx
	_ = baseURL
	_ = bearerToken
	_ = params
	return nil, exitcode.New(exitcode.KindUnexpected, "not implemented in this test fake", nil)
}

func (f *fakePlannerAPI) GetMyMemberProfile(ctx context.Context, baseURL string, bearerToken string) (*gen.GetMyMemberProfileClientResponse, error) {
	_ = ctx
	_ = baseURL
	_ = bearerToken
	return nil, exitcode.New(exitcode.KindUnexpected, "not implemented in this test fake", nil)
}

func (f *fakePlannerAPI) CreateMyMember(ctx context.Context, baseURL string, bearerToken string, req gen.CreateMyMemberJSONRequestBody) (*gen.CreateMyMemberClientResponse, error) {
	_ = ctx
	_ = baseURL
	_ = bearerToken
	_ = req
	return nil, exitcode.New(exitcode.KindUnexpected, "not implemented in this test fake", nil)
}
