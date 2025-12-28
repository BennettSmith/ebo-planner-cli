package cli

import (
	"fmt"
	"io"
	"os"
	"strings"

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
	memberCmd.AddCommand(newMemberUpdateCmd(deps))
	root.AddCommand(memberCmd)
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
