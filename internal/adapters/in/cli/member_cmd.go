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

func addMemberCommands(root *cobra.Command, deps RootDeps) {
	memberCmd := &cobra.Command{
		Use:   "member",
		Short: "Member operations",
	}
	memberCmd.AddCommand(newMemberUpdateCmd(deps))
	root.AddCommand(memberCmd)
}

func newMemberUpdateCmd(deps RootDeps) *cobra.Command {
	var (
		fromFile       string
		idempotencyKey string
	)

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update my member profile (patch)",
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

			if strings.TrimSpace(fromFile) == "" {
				return exitcode.New(exitcode.KindUsage, "missing input (use --from-file)", nil)
			}

			var req gen.UpdateMyMemberProfileRequest
			if err := requestfile.LoadStrict(fromFile, &req); err != nil {
				return exitcode.New(exitcode.KindUsage, "parse request file", err)
			}

			if strings.TrimSpace(idempotencyKey) == "" {
				idempotencyKey = idempotency.NewKey()
			}

			resp, err := deps.PlannerAPI.UpdateMyMemberProfile(ctx, apiCtx.APIURL, apiCtx.BearerToken, idempotencyKey, req)
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
