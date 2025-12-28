package cli

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	plannerapiout "github.com/BennettSmith/ebo-planner-cli/internal/adapters/out/plannerapi"
	gen "github.com/BennettSmith/ebo-planner-cli/internal/gen/plannerapi"
	"github.com/BennettSmith/ebo-planner-cli/internal/platform/cliopts"
	"github.com/BennettSmith/ebo-planner-cli/internal/platform/editmode"
	"github.com/BennettSmith/ebo-planner-cli/internal/platform/envelope"
	"github.com/BennettSmith/ebo-planner-cli/internal/platform/exitcode"
	"github.com/BennettSmith/ebo-planner-cli/internal/platform/idempotency"
	"github.com/BennettSmith/ebo-planner-cli/internal/platform/prompt"
	"github.com/BennettSmith/ebo-planner-cli/internal/platform/requestfile"
	"github.com/spf13/cobra"
)

func addMemberCommands(root *cobra.Command, deps RootDeps) {
	memberCmd := &cobra.Command{
		Use:   "member",
		Short: "Member operations",
	}
	memberCmd.AddCommand(newMemberListCmd(deps))
	memberCmd.AddCommand(newMemberSearchCmd(deps))
	memberCmd.AddCommand(newMemberMeCmd(deps))
	memberCmd.AddCommand(newMemberUpdateCmd(deps))
	root.AddCommand(memberCmd)
}

func newMemberListCmd(deps RootDeps) *cobra.Command {
	var includeInactive bool
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List members",
		RunE: func(cmd *cobra.Command, args []string) error {
			_ = args
			ctx := cmd.Context()
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

			var params *gen.ListMembersParams
			if includeInactive {
				v := gen.IncludeInactive(true)
				params = &gen.ListMembersParams{IncludeInactive: &v}
			}

			resp, err := deps.PlannerAPI.ListMembers(ctx, apiCtx.APIURL, apiCtx.BearerToken, params)
			if err != nil {
				return err
			}

			entries := []gen.MemberDirectoryEntry{}
			if resp.JSON200 != nil {
				entries = resp.JSON200.Members
			}

			if resolved.Options.Output == cliopts.OutputJSON {
				return envelope.WriteJSON(deps.Stdout, envelope.Envelope{
					Data: resp.JSON200,
					Meta: envelope.Meta{APIURL: apiCtx.APIURL, Profile: apiCtx.Profile},
				})
			}

			_, _ = io.WriteString(deps.Stdout, "MEMBER_ID\tDISPLAY_NAME\n")
			for _, m := range entries {
				_, _ = fmt.Fprintf(deps.Stdout, "%s\t%s\n", m.MemberId, m.DisplayName)
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&includeInactive, "include-inactive", false, "Include inactive members")
	return cmd
}

func newMemberSearchCmd(deps RootDeps) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "search <query>",
		Short: "Search members by display name",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
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

			q := strings.TrimSpace(args[0])
			if len(q) < 3 {
				return exitcode.New(exitcode.KindUsage, "query must be at least 3 characters", nil)
			}
			params := &gen.SearchMembersParams{Q: gen.SearchQuery(q)}
			resp, err := deps.PlannerAPI.SearchMembers(ctx, apiCtx.APIURL, apiCtx.BearerToken, params)
			if err != nil {
				return err
			}

			entries := []gen.MemberDirectoryEntry{}
			if resp.JSON200 != nil {
				entries = resp.JSON200.Members
			}

			if resolved.Options.Output == cliopts.OutputJSON {
				return envelope.WriteJSON(deps.Stdout, envelope.Envelope{
					Data: resp.JSON200,
					Meta: envelope.Meta{APIURL: apiCtx.APIURL, Profile: apiCtx.Profile},
				})
			}

			_, _ = io.WriteString(deps.Stdout, "MEMBER_ID\tDISPLAY_NAME\n")
			for _, m := range entries {
				_, _ = fmt.Fprintf(deps.Stdout, "%s\t%s\n", m.MemberId, m.DisplayName)
			}
			return nil
		},
	}
	return cmd
}

func newMemberMeCmd(deps RootDeps) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "me",
		Short: "Get my member profile",
		RunE: func(cmd *cobra.Command, args []string) error {
			_ = args
			ctx := cmd.Context()
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

			resp, err := deps.PlannerAPI.GetMyMemberProfile(ctx, apiCtx.APIURL, apiCtx.BearerToken)
			if err != nil {
				var ae *plannerapiout.APIError
				if errors.As(err, &ae) && ae != nil && ae.ErrorCode == "MEMBER_NOT_PROVISIONED" {
					return exitcode.New(exitcode.KindNotFound, "member not provisioned; run `ebo member create ...`", err)
				}
				return err
			}

			if resolved.Options.Output == cliopts.OutputJSON {
				return envelope.WriteJSON(deps.Stdout, envelope.Envelope{
					Data: resp.JSON200,
					Meta: envelope.Meta{APIURL: apiCtx.APIURL, Profile: apiCtx.Profile},
				})
			}

			if resp.JSON200 == nil {
				_, _ = io.WriteString(deps.Stdout, "OK\n")
				return nil
			}
			m := resp.JSON200.Member
			groupAlias := ""
			if m.GroupAliasEmail != nil {
				groupAlias = string(*m.GroupAliasEmail)
			}
			_, _ = fmt.Fprintf(deps.Stdout, "MemberId: %s\nDisplayName: %s\nEmail: %s\nGroupAliasEmail: %s\n", m.MemberId, m.DisplayName, string(m.Email), groupAlias)
			return nil
		},
	}
	return cmd
}

func newMemberUpdateCmd(deps RootDeps) *cobra.Command {
	var (
		fromFile       string
		edit           bool
		promptMode     bool
		idempotencyKey string
	)

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update my member profile (patch)",
		RunE: func(cmd *cobra.Command, args []string) error {
			_ = args
			ctx := cmd.Context()

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

			modeCount := 0
			if strings.TrimSpace(fromFile) != "" {
				modeCount++
			}
			if edit {
				modeCount++
			}
			if promptMode {
				modeCount++
			}
			if modeCount == 0 {
				return exitcode.New(exitcode.KindUsage, "missing input (use --from-file, --edit, or --prompt)", nil)
			}
			if modeCount > 1 {
				return exitcode.New(exitcode.KindUsage, "choose exactly one of --from-file, --edit, or --prompt", nil)
			}

			var req gen.UpdateMyMemberProfileRequest
			switch {
			case edit:
				edited, err := editmode.EditTemp(updateMemberTemplateYAML)
				if err != nil {
					return exitcode.New(exitcode.KindServer, "open editor", err)
				}
				tmp, err := os.CreateTemp("", "ebo-editbuf-*.req")
				if err != nil {
					return exitcode.New(exitcode.KindServer, "temp file", err)
				}
				tmpName := tmp.Name()
				_ = tmp.Close()
				defer func() { _ = os.Remove(tmpName) }()
				if err := os.WriteFile(tmpName, edited, 0o600); err != nil {
					return exitcode.New(exitcode.KindServer, "write edited buffer", err)
				}
				if err := requestfile.LoadStrict(tmpName, &req); err != nil {
					return exitcode.New(exitcode.KindUsage, "parse edited buffer", err)
				}
			case promptMode:
				p := prompt.New(cmd.InOrStdin(), deps.Stderr, func(template string) ([]byte, error) { return editmode.EditTemp(template) })

				name, err := p.PromptOptionalString(ctx, "Display name (optional)")
				if err != nil {
					if err == prompt.ErrAborted {
						return exitcode.New(exitcode.KindInterrupted, "interrupted", err)
					}
					return exitcode.New(exitcode.KindServer, "prompt", err)
				}
				if strings.TrimSpace(name) != "" {
					req.DisplayName = &name
				}

				wantVehicle, err := p.PromptYesNo(ctx, "Configure vehicle profile?", true)
				if err != nil {
					if err == prompt.ErrAborted {
						return exitcode.New(exitcode.KindInterrupted, "interrupted", err)
					}
					return exitcode.New(exitcode.KindServer, "prompt", err)
				}
				if wantVehicle {
					notes, usedEditor, err := p.PromptMultilineOrInline(ctx, "vehicle notes (optional)", editTextTemplateYAML("text", "Vehicle notes"))
					if err != nil {
						if err == prompt.ErrAborted {
							return exitcode.New(exitcode.KindInterrupted, "interrupted", err)
						}
						return exitcode.New(exitcode.KindServer, "prompt", err)
					}
					if usedEditor {
						type textDoc struct {
							Text string `json:"text"`
						}
						tmp, err := os.CreateTemp("", "ebo-edittext-*.req")
						if err != nil {
							return exitcode.New(exitcode.KindServer, "temp file", err)
						}
						tmpName := tmp.Name()
						_ = tmp.Close()
						defer func() { _ = os.Remove(tmpName) }()
						if err := os.WriteFile(tmpName, []byte(notes), 0o600); err != nil {
							return exitcode.New(exitcode.KindServer, "write edited buffer", err)
						}
						var td textDoc
						if err := requestfile.LoadStrict(tmpName, &td); err != nil {
							return exitcode.New(exitcode.KindUsage, "parse edited buffer", err)
						}
						notes = td.Text
					}
					if strings.TrimSpace(notes) != "" {
						req.VehicleProfile = &gen.VehicleProfile{Notes: &notes}
					}
				}
			default:
				if err := requestfile.LoadStrict(fromFile, &req); err != nil {
					return exitcode.New(exitcode.KindUsage, "parse request file", err)
				}
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
	cmd.Flags().BoolVar(&edit, "edit", false, "Edit request body in $EBO_EDITOR/$EDITOR (YAML template)")
	cmd.Flags().BoolVar(&promptMode, "prompt", false, "Interactive guided entry")
	cmd.Flags().StringVar(&idempotencyKey, "idempotency-key", "", "Idempotency key (auto-generated if omitted)")

	return cmd
}

const updateMemberTemplateYAML = `# UpdateMyMemberProfileRequest (patch)
#
# - Omitted fields are unchanged.
# - Set fields to update.
#
# displayName:
# email:
# groupAliasEmail:
# vehicleProfile:
#   make:
#   model:
#   tireSize:
#   liftLockers:
#   fuelRange:
#   recoveryGear:
#   hamRadioCallSign:
#   notes:
{}
`
