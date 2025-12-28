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
		promptMode     bool
		idempotencyKey string
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a draft trip",
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

			if strings.TrimSpace(name) != "" && strings.TrimSpace(fromFile) != "" {
				return exitcode.New(exitcode.KindUsage, "choose exactly one of --name or --from-file", nil)
			}
			if (strings.TrimSpace(name) != "" || strings.TrimSpace(fromFile) != "") && promptMode {
				return exitcode.New(exitcode.KindUsage, "choose exactly one of --name, --from-file, or --prompt", nil)
			}

			var req gen.CreateTripDraftRequest
			switch {
			case strings.TrimSpace(fromFile) != "":
				if err := requestfile.LoadStrict(fromFile, &req); err != nil {
					return exitcode.New(exitcode.KindUsage, "parse request file", err)
				}
			case promptMode:
				p := prompt.New(cmd.InOrStdin(), deps.Stderr, nil)
				n, err := p.PromptRequiredString(ctx, "Name")
				if err != nil {
					if err == prompt.ErrAborted {
						return exitcode.New(exitcode.KindInterrupted, "interrupted", err)
					}
					return exitcode.New(exitcode.KindServer, "prompt", err)
				}
				req = gen.CreateTripDraftRequest{Name: n}
			case strings.TrimSpace(name) != "":
				req = gen.CreateTripDraftRequest{Name: name}
			default:
				return exitcode.New(exitcode.KindUsage, "missing input (use --name, --from-file, or --prompt)", nil)
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
	cmd.Flags().BoolVar(&promptMode, "prompt", false, "Interactive guided entry")
	cmd.Flags().StringVar(&idempotencyKey, "idempotency-key", "", "Idempotency key (auto-generated if omitted)")

	return cmd
}

func newTripUpdateCmd(deps RootDeps) *cobra.Command {
	var (
		fromFile       string
		edit           bool
		promptMode     bool
		idempotencyKey string
	)

	cmd := &cobra.Command{
		Use:   "update <tripId>",
		Short: "Update a trip (patch)",
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

			var req gen.UpdateTripRequest
			switch {
			case edit:
				edited, err := editmode.EditTemp(updateTripTemplateYAML)
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
				desc, usedEditor, err := p.PromptMultilineOrInline(ctx, "description (optional)", editTextTemplateYAML("text", "Description"))
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
					if err := os.WriteFile(tmpName, []byte(desc), 0o600); err != nil {
						return exitcode.New(exitcode.KindServer, "write edited buffer", err)
					}
					var td textDoc
					if err := requestfile.LoadStrict(tmpName, &td); err != nil {
						return exitcode.New(exitcode.KindUsage, "parse edited buffer", err)
					}
					desc = td.Text
				}
				if strings.TrimSpace(desc) != "" {
					req.Description = &desc
				}

				ids, err := p.PromptStringList(ctx, "artifactIds")
				if err != nil {
					if err == prompt.ErrAborted {
						return exitcode.New(exitcode.KindInterrupted, "interrupted", err)
					}
					return exitcode.New(exitcode.KindServer, "prompt", err)
				}
				if len(ids) > 0 {
					req.ArtifactIds = &ids
				}
			default:
				if err := requestfile.LoadStrict(fromFile, &req); err != nil {
					return exitcode.New(exitcode.KindUsage, "parse request file", err)
				}
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
	cmd.Flags().BoolVar(&edit, "edit", false, "Edit request body in $EBO_EDITOR/$EDITOR (YAML template)")
	cmd.Flags().BoolVar(&promptMode, "prompt", false, "Interactive guided entry")
	cmd.Flags().StringVar(&idempotencyKey, "idempotency-key", "", "Idempotency key (auto-generated if omitted)")

	return cmd
}

const updateTripTemplateYAML = `# UpdateTripRequest (patch)
#
# - Omitted fields are unchanged.
# - Set fields to update.
#
# Example:
# description: |-
#   line1
#   line2
#
# name:
# description:
# startDate:
# endDate:
# capacityRigs:
# difficultyText:
# commsRequirementsText:
# recommendedRequirementsText:
# artifactIds:
# meetingLocation:
#   label:
#   address:
#   latitudeLongitude:
#     latitude:
#     longitude:
{}
`

func editTextTemplateYAML(key string, title string) string {
	return fmt.Sprintf(`# %s (plain text)
#
# Put your text under "%s".
%s: |-
  `+"\n", title, key, key)
}
