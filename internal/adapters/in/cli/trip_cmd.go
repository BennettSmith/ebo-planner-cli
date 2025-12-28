package cli

import (
	"context"
	"fmt"
	"io"
	"strings"

	gen "github.com/BennettSmith/ebo-planner-cli/internal/gen/plannerapi"
	"github.com/BennettSmith/ebo-planner-cli/internal/platform/cliopts"
	"github.com/BennettSmith/ebo-planner-cli/internal/platform/envelope"
	"github.com/BennettSmith/ebo-planner-cli/internal/platform/exitcode"
	"github.com/BennettSmith/ebo-planner-cli/internal/platform/idempotency"
	"github.com/BennettSmith/ebo-planner-cli/internal/platform/requestfile"
	"github.com/spf13/cobra"
)

func addTripCommands(root *cobra.Command, deps RootDeps) {
	tripCmd := &cobra.Command{
		Use:   "trip",
		Short: "Trip operations",
	}

	tripCmd.AddCommand(newTripCreateCmd(deps))
	tripCmd.AddCommand(newTripUpdateCmd(deps))

	root.AddCommand(tripCmd)
}

func newTripCreateCmd(deps RootDeps) *cobra.Command {
	var (
		name           string
		fromFile       string
		idempotencyKey string
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a draft trip",
		RunE: func(cmd *cobra.Command, args []string) error {
			_ = args
			ctx := context.Background()

			if deps.PlannerAPI == nil {
				return exitcode.New(exitcode.KindUnexpected, "planner api", fmt.Errorf("nil planner api client"))
			}

			resolved, err := resolvedFromRoot(cmd, deps)
			if err != nil {
				return err
			}
			apiCtx, err := resolveAPIContext(ctx, deps, resolved)
			if err != nil {
				return err
			}

			if strings.TrimSpace(name) != "" && strings.TrimSpace(fromFile) != "" {
				return exitcode.New(exitcode.KindUsage, "choose exactly one of --name or --from-file", nil)
			}

			var req gen.CreateTripDraftRequest
			switch {
			case strings.TrimSpace(fromFile) != "":
				if err := requestfile.LoadStrict(fromFile, &req); err != nil {
					return exitcode.New(exitcode.KindUsage, "parse request file", err)
				}
			case strings.TrimSpace(name) != "":
				req = gen.CreateTripDraftRequest{Name: name}
			default:
				return exitcode.New(exitcode.KindUsage, "missing input (use --name or --from-file)", nil)
			}

			if strings.TrimSpace(idempotencyKey) == "" {
				idempotencyKey = idempotency.NewKey()
			}

			resp, err := deps.PlannerAPI.CreateTripDraft(ctx, apiCtx.APIURL, apiCtx.BearerToken, idempotencyKey, req)
			if err != nil {
				return err
			}

			if resolved.Options.Output == cliopts.OutputJSON {
				return envelope.WriteJSON(deps.Stdout, envelope.Envelope{
					Data: resp.JSON201,
					Meta: envelope.Meta{APIURL: apiCtx.APIURL, Profile: apiCtx.Profile, IdempotencyKey: idempotencyKey},
				})
			}

			// Keep stdout clean-ish but still human-friendly.
			_, _ = fmt.Fprintf(deps.Stderr, "Idempotency-Key: %s\n", idempotencyKey)
			if resp.JSON201 != nil {
				_, _ = fmt.Fprintf(deps.Stdout, "tripId=%s\n", resp.JSON201.Trip.TripId)
			} else {
				_, _ = io.WriteString(deps.Stdout, "OK\n")
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Trip name")
	cmd.Flags().StringVar(&fromFile, "from-file", "", "Read request body from file (JSON or YAML)")
	cmd.Flags().StringVar(&idempotencyKey, "idempotency-key", "", "Idempotency key (auto-generated if omitted)")

	return cmd
}

func newTripUpdateCmd(deps RootDeps) *cobra.Command {
	var (
		fromFile       string
		idempotencyKey string
	)

	cmd := &cobra.Command{
		Use:   "update <tripId>",
		Short: "Update a trip (patch)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			if deps.PlannerAPI == nil {
				return exitcode.New(exitcode.KindUnexpected, "planner api", fmt.Errorf("nil planner api client"))
			}

			resolved, err := resolvedFromRoot(cmd, deps)
			if err != nil {
				return err
			}
			apiCtx, err := resolveAPIContext(ctx, deps, resolved)
			if err != nil {
				return err
			}

			if strings.TrimSpace(fromFile) == "" {
				return exitcode.New(exitcode.KindUsage, "missing input (use --from-file)", nil)
			}

			var req gen.UpdateTripRequest
			if err := requestfile.LoadStrict(fromFile, &req); err != nil {
				return exitcode.New(exitcode.KindUsage, "parse request file", err)
			}

			if strings.TrimSpace(idempotencyKey) == "" {
				idempotencyKey = idempotency.NewKey()
			}

			tripID := gen.TripId(args[0])
			resp, err := deps.PlannerAPI.UpdateTrip(ctx, apiCtx.APIURL, apiCtx.BearerToken, tripID, idempotencyKey, req)
			if err != nil {
				return err
			}

			if resolved.Options.Output == cliopts.OutputJSON {
				return envelope.WriteJSON(deps.Stdout, envelope.Envelope{
					Data: resp.JSON200,
					Meta: envelope.Meta{APIURL: apiCtx.APIURL, Profile: apiCtx.Profile, IdempotencyKey: idempotencyKey},
				})
			}

			_, _ = fmt.Fprintf(deps.Stderr, "Idempotency-Key: %s\n", idempotencyKey)
			_, _ = io.WriteString(deps.Stdout, "OK\n")
			return nil
		},
	}

	cmd.Flags().StringVar(&fromFile, "from-file", "", "Read request body from file (JSON or YAML)")
	cmd.Flags().StringVar(&idempotencyKey, "idempotency-key", "", "Idempotency key (auto-generated if omitted)")

	return cmd
}
