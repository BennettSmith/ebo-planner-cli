package plannerapi

import (
	"context"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/BennettSmith/ebo-planner-cli/internal/platform/httpx"

	gen "github.com/BennettSmith/ebo-planner-cli/internal/gen/plannerapi"
	"github.com/BennettSmith/ebo-planner-cli/internal/platform/exitcode"
)

type Adapter struct {
	HTTPClient *http.Client
	Timeout    time.Duration
	Verbose    bool
	LogSink    io.Writer
}

func (a Adapter) newClient(baseURL string, bearerToken string) (*gen.ClientWithResponses, error) {
	opts := []gen.ClientOption{}
	hc := a.HTTPClient
	hc = httpx.NewClient(hc, httpx.Options{Timeout: a.Timeout, Verbose: a.Verbose, LogSink: a.LogSink})
	opts = append(opts, gen.WithHTTPClient(hc))
	opts = append(opts, gen.WithRequestEditorFn(func(ctx context.Context, req *http.Request) error {
		_ = ctx
		if strings.TrimSpace(bearerToken) != "" {
			req.Header.Set("Authorization", "Bearer "+bearerToken)
		}
		return nil
	}))
	return gen.NewClientWithResponses(baseURL, opts...)
}

func (a Adapter) ListVisibleTripsForMember(ctx context.Context, baseURL string, bearerToken string) (*gen.ListVisibleTripsForMemberClientResponse, error) {
	client, err := a.newClient(baseURL, bearerToken)
	if err != nil {
		return nil, exitcode.New(exitcode.KindServer, "failed to init client", err)
	}
	resp, err := client.ListVisibleTripsForMemberWithResponse(ctx)
	if err != nil {
		return nil, exitcode.New(exitcode.KindServer, "request failed", err)
	}
	if resp.StatusCode() >= 400 {
		return nil, apiErrorFromAny(resp.StatusCode(), resp.JSON401, resp.JSON500)
	}
	return resp, nil
}

func (a Adapter) ListMyDraftTrips(ctx context.Context, baseURL string, bearerToken string) (*gen.ListMyDraftTripsClientResponse, error) {
	client, err := a.newClient(baseURL, bearerToken)
	if err != nil {
		return nil, exitcode.New(exitcode.KindServer, "failed to init client", err)
	}
	resp, err := client.ListMyDraftTripsWithResponse(ctx)
	if err != nil {
		return nil, exitcode.New(exitcode.KindServer, "request failed", err)
	}
	if resp.StatusCode() >= 400 {
		return nil, apiErrorFromAny(resp.StatusCode(), resp.JSON401, resp.JSON500)
	}
	return resp, nil
}

func (a Adapter) GetTripDetails(ctx context.Context, baseURL string, bearerToken string, tripID gen.TripId) (*gen.GetTripDetailsClientResponse, error) {
	client, err := a.newClient(baseURL, bearerToken)
	if err != nil {
		return nil, exitcode.New(exitcode.KindServer, "failed to init client", err)
	}
	resp, err := client.GetTripDetailsWithResponse(ctx, tripID)
	if err != nil {
		return nil, exitcode.New(exitcode.KindServer, "request failed", err)
	}
	if resp.StatusCode() >= 400 {
		return nil, apiErrorFromAny(resp.StatusCode(), resp.JSON401, resp.JSON404, resp.JSON500)
	}
	return resp, nil
}

func (a Adapter) CreateTripDraft(ctx context.Context, baseURL string, bearerToken string, idempotencyKey string, req gen.CreateTripDraftJSONRequestBody) (*gen.CreateTripDraftClientResponse, error) {
	client, err := a.newClient(baseURL, bearerToken)
	if err != nil {
		return nil, exitcode.New(exitcode.KindServer, "failed to init client", err)
	}
	params := &gen.CreateTripDraftParams{IdempotencyKey: idempotencyKey}
	resp, err := client.CreateTripDraftWithResponse(ctx, params, req)
	if err != nil {
		return nil, exitcode.New(exitcode.KindServer, "request failed", err)
	}
	if resp.StatusCode() >= 400 {
		return nil, apiErrorFromAny(resp.StatusCode(), resp.JSON401, resp.JSON500)
	}
	return resp, nil
}

func (a Adapter) UpdateTrip(ctx context.Context, baseURL string, bearerToken string, tripID gen.TripId, idempotencyKey string, req gen.UpdateTripJSONRequestBody) (*gen.UpdateTripClientResponse, error) {
	client, err := a.newClient(baseURL, bearerToken)
	if err != nil {
		return nil, exitcode.New(exitcode.KindServer, "failed to init client", err)
	}
	params := &gen.UpdateTripParams{IdempotencyKey: idempotencyKey}
	resp, err := client.UpdateTripWithResponse(ctx, tripID, params, req)
	if err != nil {
		return nil, exitcode.New(exitcode.KindServer, "request failed", err)
	}
	if resp.StatusCode() >= 400 {
		return nil, apiErrorFromAny(resp.StatusCode(), resp.JSON401, resp.JSON404, resp.JSON409, resp.JSON422, resp.JSON500)
	}
	return resp, nil
}

func (a Adapter) SetTripDraftVisibility(ctx context.Context, baseURL string, bearerToken string, tripID gen.TripId, idempotencyKey string, req gen.SetTripDraftVisibilityJSONRequestBody) (*gen.SetTripDraftVisibilityClientResponse, error) {
	client, err := a.newClient(baseURL, bearerToken)
	if err != nil {
		return nil, exitcode.New(exitcode.KindServer, "failed to init client", err)
	}
	params := &gen.SetTripDraftVisibilityParams{IdempotencyKey: idempotencyKey}
	resp, err := client.SetTripDraftVisibilityWithResponse(ctx, tripID, params, req)
	if err != nil {
		return nil, exitcode.New(exitcode.KindServer, "request failed", err)
	}
	if resp.StatusCode() >= 400 {
		return nil, apiErrorFromAny(resp.StatusCode(), resp.JSON401, resp.JSON404, resp.JSON409, resp.JSON422, resp.JSON500)
	}
	return resp, nil
}

func (a Adapter) PublishTrip(ctx context.Context, baseURL string, bearerToken string, tripID gen.TripId) (*gen.PublishTripClientResponse, error) {
	client, err := a.newClient(baseURL, bearerToken)
	if err != nil {
		return nil, exitcode.New(exitcode.KindServer, "failed to init client", err)
	}
	resp, err := client.PublishTripWithResponse(ctx, tripID)
	if err != nil {
		return nil, exitcode.New(exitcode.KindServer, "request failed", err)
	}
	if resp.StatusCode() >= 400 {
		return nil, apiErrorFromAny(resp.StatusCode(), resp.JSON401, resp.JSON404, resp.JSON409, resp.JSON422, resp.JSON500)
	}
	return resp, nil
}

func (a Adapter) AddTripOrganizer(ctx context.Context, baseURL string, bearerToken string, tripID gen.TripId, idempotencyKey string, req gen.AddTripOrganizerJSONRequestBody) (*gen.AddTripOrganizerClientResponse, error) {
	client, err := a.newClient(baseURL, bearerToken)
	if err != nil {
		return nil, exitcode.New(exitcode.KindServer, "failed to init client", err)
	}
	params := &gen.AddTripOrganizerParams{IdempotencyKey: idempotencyKey}
	resp, err := client.AddTripOrganizerWithResponse(ctx, tripID, params, req)
	if err != nil {
		return nil, exitcode.New(exitcode.KindServer, "request failed", err)
	}
	if resp.StatusCode() >= 400 {
		return nil, apiErrorFromAny(resp.StatusCode(), resp.JSON401, resp.JSON404, resp.JSON409, resp.JSON422, resp.JSON500)
	}
	return resp, nil
}

func (a Adapter) RemoveTripOrganizer(ctx context.Context, baseURL string, bearerToken string, tripID gen.TripId, memberID gen.MemberId, idempotencyKey string) (*gen.RemoveTripOrganizerClientResponse, error) {
	client, err := a.newClient(baseURL, bearerToken)
	if err != nil {
		return nil, exitcode.New(exitcode.KindServer, "failed to init client", err)
	}
	params := &gen.RemoveTripOrganizerParams{IdempotencyKey: idempotencyKey}
	resp, err := client.RemoveTripOrganizerWithResponse(ctx, tripID, memberID, params)
	if err != nil {
		return nil, exitcode.New(exitcode.KindServer, "request failed", err)
	}
	if resp.StatusCode() >= 400 {
		return nil, apiErrorFromAny(resp.StatusCode(), resp.JSON401, resp.JSON404, resp.JSON409, resp.JSON422, resp.JSON500)
	}
	return resp, nil
}

func (a Adapter) SetMyRSVP(ctx context.Context, baseURL string, bearerToken string, tripID gen.TripId, idempotencyKey string, req gen.SetMyRSVPJSONRequestBody) (*gen.SetMyRSVPClientResponse, error) {
	client, err := a.newClient(baseURL, bearerToken)
	if err != nil {
		return nil, exitcode.New(exitcode.KindServer, "failed to init client", err)
	}
	params := &gen.SetMyRSVPParams{IdempotencyKey: idempotencyKey}
	resp, err := client.SetMyRSVPWithResponse(ctx, tripID, params, req)
	if err != nil {
		return nil, exitcode.New(exitcode.KindServer, "request failed", err)
	}
	if resp.StatusCode() >= 400 {
		return nil, apiErrorFromAny(resp.StatusCode(), resp.JSON401, resp.JSON404, resp.JSON409, resp.JSON422, resp.JSON500)
	}
	return resp, nil
}

func (a Adapter) GetMyRSVPForTrip(ctx context.Context, baseURL string, bearerToken string, tripID gen.TripId) (*gen.GetMyRSVPForTripClientResponse, error) {
	client, err := a.newClient(baseURL, bearerToken)
	if err != nil {
		return nil, exitcode.New(exitcode.KindServer, "failed to init client", err)
	}
	resp, err := client.GetMyRSVPForTripWithResponse(ctx, tripID)
	if err != nil {
		return nil, exitcode.New(exitcode.KindServer, "request failed", err)
	}
	if resp.StatusCode() >= 400 {
		return nil, apiErrorFromAny(resp.StatusCode(), resp.JSON401, resp.JSON404, resp.JSON409, resp.JSON500)
	}
	return resp, nil
}

func (a Adapter) GetTripRSVPSummary(ctx context.Context, baseURL string, bearerToken string, tripID gen.TripId) (*gen.GetTripRSVPSummaryClientResponse, error) {
	client, err := a.newClient(baseURL, bearerToken)
	if err != nil {
		return nil, exitcode.New(exitcode.KindServer, "failed to init client", err)
	}
	resp, err := client.GetTripRSVPSummaryWithResponse(ctx, tripID)
	if err != nil {
		return nil, exitcode.New(exitcode.KindServer, "request failed", err)
	}
	if resp.StatusCode() >= 400 {
		return nil, apiErrorFromAny(resp.StatusCode(), resp.JSON401, resp.JSON404, resp.JSON409, resp.JSON500)
	}
	return resp, nil
}

func (a Adapter) CancelTrip(ctx context.Context, baseURL string, bearerToken string, tripID gen.TripId, idempotencyKey *string) (*gen.CancelTripClientResponse, error) {
	client, err := a.newClient(baseURL, bearerToken)
	if err != nil {
		return nil, exitcode.New(exitcode.KindServer, "failed to init client", err)
	}
	params := &gen.CancelTripParams{IdempotencyKey: idempotencyKey}
	resp, err := client.CancelTripWithResponse(ctx, tripID, params)
	if err != nil {
		return nil, exitcode.New(exitcode.KindServer, "request failed", err)
	}
	if resp.StatusCode() >= 400 {
		return nil, apiErrorFromAny(resp.StatusCode(), resp.JSON401, resp.JSON404, resp.JSON409, resp.JSON422, resp.JSON500)
	}
	return resp, nil
}

func (a Adapter) ListMembers(ctx context.Context, baseURL string, bearerToken string, params *gen.ListMembersParams) (*gen.ListMembersClientResponse, error) {
	client, err := a.newClient(baseURL, bearerToken)
	if err != nil {
		return nil, exitcode.New(exitcode.KindServer, "failed to init client", err)
	}
	resp, err := client.ListMembersWithResponse(ctx, params)
	if err != nil {
		return nil, exitcode.New(exitcode.KindServer, "request failed", err)
	}
	if resp.StatusCode() >= 400 {
		return nil, apiErrorFromAny(resp.StatusCode(), resp.JSON401, resp.JSON500)
	}
	return resp, nil
}

func (a Adapter) UpdateMyMemberProfile(ctx context.Context, baseURL string, bearerToken string, idempotencyKey string, req gen.UpdateMyMemberProfileJSONRequestBody) (*gen.UpdateMyMemberProfileClientResponse, error) {
	client, err := a.newClient(baseURL, bearerToken)
	if err != nil {
		return nil, exitcode.New(exitcode.KindServer, "failed to init client", err)
	}
	params := &gen.UpdateMyMemberProfileParams{IdempotencyKey: idempotencyKey}
	resp, err := client.UpdateMyMemberProfileWithResponse(ctx, params, req)
	if err != nil {
		return nil, exitcode.New(exitcode.KindServer, "request failed", err)
	}
	if resp.StatusCode() >= 400 {
		return nil, apiErrorFromAny(resp.StatusCode(), resp.JSON401, resp.JSON404, resp.JSON409, resp.JSON422, resp.JSON500)
	}
	return resp, nil
}
