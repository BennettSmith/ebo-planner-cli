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
