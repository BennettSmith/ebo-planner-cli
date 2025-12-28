package plannerapi

import (
	"context"

	gen "github.com/BennettSmith/ebo-planner-cli/internal/gen/plannerapi"
)

// Client is the outbound port used by the application layer.
//
// It wraps the generated OpenAPI client behind a stable interface.
//
// Note: this port currently exposes generated types. As the CLI matures,
// we can introduce domain-level DTOs.
type Client interface {
	// Trips
	ListVisibleTripsForMember(ctx context.Context, baseURL string, bearerToken string) (*gen.ListVisibleTripsForMemberClientResponse, error)
	ListMyDraftTrips(ctx context.Context, baseURL string, bearerToken string) (*gen.ListMyDraftTripsClientResponse, error)
	GetTripDetails(ctx context.Context, baseURL string, bearerToken string, tripID gen.TripId) (*gen.GetTripDetailsClientResponse, error)
	CreateTripDraft(ctx context.Context, baseURL string, bearerToken string, idempotencyKey string, req gen.CreateTripDraftJSONRequestBody) (*gen.CreateTripDraftClientResponse, error)
	UpdateTrip(ctx context.Context, baseURL string, bearerToken string, tripID gen.TripId, idempotencyKey string, req gen.UpdateTripJSONRequestBody) (*gen.UpdateTripClientResponse, error)
	CancelTrip(ctx context.Context, baseURL string, bearerToken string, tripID gen.TripId, idempotencyKey *string) (*gen.CancelTripClientResponse, error)

	// Members
	ListMembers(ctx context.Context, baseURL string, bearerToken string, params *gen.ListMembersParams) (*gen.ListMembersClientResponse, error)
	UpdateMyMemberProfile(ctx context.Context, baseURL string, bearerToken string, idempotencyKey string, req gen.UpdateMyMemberProfileJSONRequestBody) (*gen.UpdateMyMemberProfileClientResponse, error)
}
