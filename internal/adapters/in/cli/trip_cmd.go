package cli

import (
	"fmt"
	"io"
	"os"
	"strings"

	gen "github.com/Overland-East-Bay/trip-planner-cli/internal/gen/plannerapi"
	"github.com/Overland-East-Bay/trip-planner-cli/internal/platform/cliopts"
	"github.com/Overland-East-Bay/trip-planner-cli/internal/platform/editmode"
	"github.com/Overland-East-Bay/trip-planner-cli/internal/platform/envelope"
	"github.com/Overland-East-Bay/trip-planner-cli/internal/platform/exitcode"
	"github.com/Overland-East-Bay/trip-planner-cli/internal/platform/idempotency"
	"github.com/Overland-East-Bay/trip-planner-cli/internal/platform/prompt"
	"github.com/Overland-East-Bay/trip-planner-cli/internal/platform/requestfile"
	"github.com/spf13/cobra"
)

func addTripCommands(root *cobra.Command, deps RootDeps) {
	tripCmd := &cobra.Command{
		Use:   "trip",
		Short: "Trip operations",
	}

	tripCmd.AddCommand(newTripListCmd(deps))
	tripCmd.AddCommand(newTripDraftsCmd(deps))
	tripCmd.AddCommand(newTripGetCmd(deps))
	tripCmd.AddCommand(newTripCreateCmd(deps))
	tripCmd.AddCommand(newTripUpdateCmd(deps))
	tripCmd.AddCommand(newTripVisibilityCmd(deps))
	tripCmd.AddCommand(newTripPublishCmd(deps))
	tripCmd.AddCommand(newTripCancelCmd(deps))
	tripCmd.AddCommand(newTripOrganizerCmd(deps))
	tripCmd.AddCommand(newTripRSVPCmd(deps))

	root.AddCommand(tripCmd)
}

func newTripRSVPCmd(deps RootDeps) *cobra.Command {
	rsvpCmd := &cobra.Command{
		Use:   "rsvp",
		Short: "Trip RSVP",
	}
	rsvpCmd.AddCommand(newTripRSVPSetCmd(deps))
	rsvpCmd.AddCommand(newTripRSVPGetCmd(deps))
	rsvpCmd.AddCommand(newTripRSVPSummaryCmd(deps))
	return rsvpCmd
}

func newTripRSVPSetCmd(deps RootDeps) *cobra.Command {
	var (
		yes            bool
		no             bool
		unset          bool
		idempotencyKey string
	)
	cmd := &cobra.Command{
		Use:   "set <tripId> --yes|--no|--unset",
		Short: "Set my RSVP for a trip",
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

			mode := 0
			if yes {
				mode++
			}
			if no {
				mode++
			}
			if unset {
				mode++
			}
			if mode != 1 {
				return exitcode.New(exitcode.KindUsage, "choose exactly one of --yes, --no, or --unset", nil)
			}
			if strings.TrimSpace(idempotencyKey) == "" {
				idempotencyKey = idempotency.NewKey()
			}

			resp := gen.YES
			if no {
				resp = gen.NO
			}
			if unset {
				resp = gen.UNSET
			}

			tripID := gen.TripId(args[0])
			out, err := deps.PlannerAPI.SetMyRSVP(ctx, apiCtx.APIURL, apiCtx.BearerToken, tripID, idempotencyKey, gen.SetMyRSVPRequest{Response: resp})
			if err != nil {
				return err
			}

			if resolved.Options.Output == cliopts.OutputJSON {
				return envelope.WriteJSON(deps.Stdout, envelope.Envelope{
					Data: out.JSON200,
					Meta: envelope.Meta{APIURL: apiCtx.APIURL, Profile: apiCtx.Profile, IdempotencyKey: idempotencyKey},
				})
			}

			_, _ = fmt.Fprintf(deps.Stderr, "Idempotency-Key: %s\n", idempotencyKey)
			_, _ = io.WriteString(deps.Stdout, "OK\n")
			return nil
		},
	}
	cmd.Flags().BoolVar(&yes, "yes", false, "RSVP yes")
	cmd.Flags().BoolVar(&no, "no", false, "RSVP no")
	cmd.Flags().BoolVar(&unset, "unset", false, "Unset RSVP")
	cmd.Flags().StringVar(&idempotencyKey, "idempotency-key", "", "Idempotency key (auto-generated if omitted)")
	return cmd
}

func newTripRSVPGetCmd(deps RootDeps) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <tripId>",
		Short: "Get my RSVP for a trip",
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

			tripID := gen.TripId(args[0])
			resp, err := deps.PlannerAPI.GetMyRSVPForTrip(ctx, apiCtx.APIURL, apiCtx.BearerToken, tripID)
			if err != nil {
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
			_, _ = io.WriteString(deps.Stdout, "TRIP_ID\tRESPONSE\n")
			_, _ = fmt.Fprintf(deps.Stdout, "%s\t%s\n", tripID, resp.JSON200.MyRsvp.Response)
			return nil
		},
	}
	return cmd
}

func newTripRSVPSummaryCmd(deps RootDeps) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "summary <tripId>",
		Short: "Get RSVP summary for a trip",
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

			tripID := gen.TripId(args[0])
			resp, err := deps.PlannerAPI.GetTripRSVPSummary(ctx, apiCtx.APIURL, apiCtx.BearerToken, tripID)
			if err != nil {
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
			s := resp.JSON200.RsvpSummary
			_, _ = io.WriteString(deps.Stdout, "TRIP_ID\tATTENDING_RIGS\tATTENDING_MEMBERS\tNOT_ATTENDING_MEMBERS\n")
			_, _ = fmt.Fprintf(deps.Stdout, "%s\t%d\t%d\t%d\n", tripID, s.AttendingRigs, len(s.AttendingMembers), len(s.NotAttendingMembers))
			return nil
		},
	}
	return cmd
}

func newTripOrganizerCmd(deps RootDeps) *cobra.Command {
	orgCmd := &cobra.Command{
		Use:   "organizer",
		Short: "Trip organizer management",
	}
	orgCmd.AddCommand(newTripOrganizerAddCmd(deps))
	orgCmd.AddCommand(newTripOrganizerRemoveCmd(deps))
	return orgCmd
}

func newTripOrganizerAddCmd(deps RootDeps) *cobra.Command {
	var (
		memberID       string
		idempotencyKey string
	)
	cmd := &cobra.Command{
		Use:   "add <tripId> --member <memberId>",
		Short: "Add an organizer to a trip",
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

			if strings.TrimSpace(memberID) == "" {
				return exitcode.New(exitcode.KindUsage, "missing --member", nil)
			}
			if strings.TrimSpace(idempotencyKey) == "" {
				idempotencyKey = idempotency.NewKey()
			}

			tripID := gen.TripId(args[0])
			resp, err := deps.PlannerAPI.AddTripOrganizer(ctx, apiCtx.APIURL, apiCtx.BearerToken, tripID, idempotencyKey, gen.AddOrganizerRequest{MemberId: memberID})
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
	cmd.Flags().StringVar(&memberID, "member", "", "Member ID to add as organizer")
	cmd.Flags().StringVar(&idempotencyKey, "idempotency-key", "", "Idempotency key (auto-generated if omitted)")
	return cmd
}

func newTripOrganizerRemoveCmd(deps RootDeps) *cobra.Command {
	var (
		memberID       string
		force          bool
		idempotencyKey string
	)
	cmd := &cobra.Command{
		Use:   "remove <tripId> --member <memberId>",
		Short: "Remove an organizer from a trip",
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

			if strings.TrimSpace(memberID) == "" {
				return exitcode.New(exitcode.KindUsage, "missing --member", nil)
			}
			if !force {
				return exitcode.New(exitcode.KindUsage, "refusing to remove organizer without --force", nil)
			}
			if strings.TrimSpace(idempotencyKey) == "" {
				idempotencyKey = idempotency.NewKey()
			}

			tripID := gen.TripId(args[0])
			resp, err := deps.PlannerAPI.RemoveTripOrganizer(ctx, apiCtx.APIURL, apiCtx.BearerToken, tripID, gen.MemberId(memberID), idempotencyKey)
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
	cmd.Flags().StringVar(&memberID, "member", "", "Member ID to remove as organizer")
	cmd.Flags().BoolVar(&force, "force", false, "Required: confirm organizer removal")
	cmd.Flags().StringVar(&idempotencyKey, "idempotency-key", "", "Idempotency key (auto-generated if omitted)")
	return cmd
}

func newTripListCmd(deps RootDeps) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List visible trips",
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

			resp, err := deps.PlannerAPI.ListVisibleTripsForMember(ctx, apiCtx.APIURL, apiCtx.BearerToken)
			if err != nil {
				return err
			}

			trips := []gen.TripSummary{}
			if resp.JSON200 != nil {
				trips = resp.JSON200.Trips
			}

			if resolved.Options.Output == cliopts.OutputJSON {
				return envelope.WriteJSON(deps.Stdout, envelope.Envelope{
					Data: resp.JSON200,
					Meta: envelope.Meta{APIURL: apiCtx.APIURL, Profile: apiCtx.Profile},
				})
			}

			_, _ = io.WriteString(deps.Stdout, "TRIP_ID\tSTATUS\tNAME\n")
			for _, t := range trips {
				name := ""
				if t.Name != nil {
					name = *t.Name
				}
				_, _ = fmt.Fprintf(deps.Stdout, "%s\t%s\t%s\n", t.TripId, t.Status, name)
			}
			return nil
		},
	}
}

func newTripDraftsCmd(deps RootDeps) *cobra.Command {
	return &cobra.Command{
		Use:   "drafts",
		Short: "List my draft trips",
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

			resp, err := deps.PlannerAPI.ListMyDraftTrips(ctx, apiCtx.APIURL, apiCtx.BearerToken)
			if err != nil {
				return err
			}

			trips := []gen.TripSummary{}
			if resp.JSON200 != nil {
				trips = resp.JSON200.Trips
			}

			if resolved.Options.Output == cliopts.OutputJSON {
				return envelope.WriteJSON(deps.Stdout, envelope.Envelope{
					Data: resp.JSON200,
					Meta: envelope.Meta{APIURL: apiCtx.APIURL, Profile: apiCtx.Profile},
				})
			}

			_, _ = io.WriteString(deps.Stdout, "TRIP_ID\tSTATUS\tNAME\n")
			for _, t := range trips {
				name := ""
				if t.Name != nil {
					name = *t.Name
				}
				_, _ = fmt.Fprintf(deps.Stdout, "%s\t%s\t%s\n", t.TripId, t.Status, name)
			}
			return nil
		},
	}
}

func newTripGetCmd(deps RootDeps) *cobra.Command {
	return &cobra.Command{
		Use:   "get <tripId>",
		Short: "Get trip details",
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

			tripID := gen.TripId(args[0])
			resp, err := deps.PlannerAPI.GetTripDetails(ctx, apiCtx.APIURL, apiCtx.BearerToken, tripID)
			if err != nil {
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

			t := resp.JSON200.Trip
			name := ""
			if t.Name != nil {
				name = *t.Name
			}
			_, _ = fmt.Fprintf(deps.Stdout, "TripId: %s\nStatus: %s\nName: %s\n", t.TripId, t.Status, name)
			return nil
		},
	}
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
				if err := validateSingleLineFlag(name, "--name"); err != nil {
					return exitcode.New(exitcode.KindUsage, "invalid flag", err)
				}
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

		patchName                    string
		patchDescription             string
		patchDifficultyText          string
		patchCommsRequirementsText   string
		patchRecommendedRequirements string
		patchCapacityRigs            int
		clearMeetingLocation         bool
		meetingLabel                 string
		meetingAddress               string
		meetingLat                   float64
		meetingLng                   float64
		artifactIDs                  []string
		clearArtifacts               bool
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
			patchMode := cmd.Flags().Changed("name") ||
				cmd.Flags().Changed("description") ||
				cmd.Flags().Changed("difficulty-text") ||
				cmd.Flags().Changed("comms-requirements-text") ||
				cmd.Flags().Changed("recommended-requirements-text") ||
				cmd.Flags().Changed("capacity-rigs") ||
				cmd.Flags().Changed("meeting-label") ||
				cmd.Flags().Changed("meeting-address") ||
				cmd.Flags().Changed("meeting-lat") ||
				cmd.Flags().Changed("meeting-lng") ||
				cmd.Flags().Changed("artifact-id") ||
				clearMeetingLocation ||
				clearArtifacts
			if patchMode {
				modeCount++
			}
			if modeCount == 0 {
				return exitcode.New(exitcode.KindUsage, "missing input (use patch flags, --from-file, --edit, or --prompt)", nil)
			}
			if modeCount > 1 {
				return exitcode.New(exitcode.KindUsage, "choose exactly one patch mode: flags, --from-file, --edit, or --prompt", nil)
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
			case patchMode:
				// Validation: multiline text is not allowed via flags.
				for _, v := range []struct {
					changed bool
					val     string
					name    string
				}{
					{cmd.Flags().Changed("name"), patchName, "--name"},
					{cmd.Flags().Changed("description"), patchDescription, "--description"},
					{cmd.Flags().Changed("difficulty-text"), patchDifficultyText, "--difficulty-text"},
					{cmd.Flags().Changed("comms-requirements-text"), patchCommsRequirementsText, "--comms-requirements-text"},
					{cmd.Flags().Changed("recommended-requirements-text"), patchRecommendedRequirements, "--recommended-requirements-text"},
					{cmd.Flags().Changed("meeting-label"), meetingLabel, "--meeting-label"},
					{cmd.Flags().Changed("meeting-address"), meetingAddress, "--meeting-address"},
				} {
					if v.changed {
						if err := validateSingleLineFlag(v.val, v.name); err != nil {
							return exitcode.New(exitcode.KindUsage, "invalid flag", err)
						}
					}
				}

				if clearArtifacts && cmd.Flags().Changed("artifact-id") {
					return exitcode.New(exitcode.KindUsage, "choose exactly one of --artifact-id or --clear-artifacts", nil)
				}

				if clearMeetingLocation && (cmd.Flags().Changed("meeting-label") || cmd.Flags().Changed("meeting-address") || cmd.Flags().Changed("meeting-lat") || cmd.Flags().Changed("meeting-lng")) {
					return exitcode.New(exitcode.KindUsage, "--clear-meeting-location is mutually exclusive with meeting location flags", nil)
				}

				latSet := cmd.Flags().Changed("meeting-lat")
				lngSet := cmd.Flags().Changed("meeting-lng")
				if latSet != lngSet {
					return exitcode.New(exitcode.KindUsage, "meeting lat/lng must be provided as a pair", nil)
				}

				if cmd.Flags().Changed("name") {
					req.Name = &patchName
				}
				if cmd.Flags().Changed("description") {
					req.Description = &patchDescription
				}
				if cmd.Flags().Changed("difficulty-text") {
					req.DifficultyText = &patchDifficultyText
				}
				if cmd.Flags().Changed("comms-requirements-text") {
					req.CommsRequirementsText = &patchCommsRequirementsText
				}
				if cmd.Flags().Changed("recommended-requirements-text") {
					req.RecommendedRequirementsText = &patchRecommendedRequirements
				}
				if cmd.Flags().Changed("capacity-rigs") {
					req.CapacityRigs = &patchCapacityRigs
				}

				if clearArtifacts {
					empty := []string{}
					req.ArtifactIds = &empty
				}
				if cmd.Flags().Changed("artifact-id") {
					dedup := dedupStringsPreserveFirst(artifactIDs)
					req.ArtifactIds = &dedup
				}

				if clearMeetingLocation {
					req.MeetingLocation = &gen.LocationPatch{}
				} else if cmd.Flags().Changed("meeting-label") || cmd.Flags().Changed("meeting-address") || latSet || lngSet {
					loc := &gen.LocationPatch{}
					if cmd.Flags().Changed("meeting-label") {
						loc.Label = &meetingLabel
					}
					if cmd.Flags().Changed("meeting-address") {
						loc.Address = &meetingAddress
					}
					if latSet && lngSet {
						lat := meetingLat
						lng := meetingLng
						loc.LatitudeLongitude = &struct {
							Latitude  *float64 `json:"latitude"`
							Longitude *float64 `json:"longitude"`
						}{Latitude: &lat, Longitude: &lng}
					}
					req.MeetingLocation = loc
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
	cmd.Flags().StringVar(&patchName, "name", "", "Trip name (single line; disallows newlines)")
	cmd.Flags().StringVar(&patchDescription, "description", "", "Trip description (single line; disallows newlines)")
	cmd.Flags().StringVar(&patchDifficultyText, "difficulty-text", "", "Difficulty text (single line; disallows newlines)")
	cmd.Flags().StringVar(&patchCommsRequirementsText, "comms-requirements-text", "", "Comms requirements text (single line; disallows newlines)")
	cmd.Flags().StringVar(&patchRecommendedRequirements, "recommended-requirements-text", "", "Recommended requirements text (single line; disallows newlines)")
	cmd.Flags().IntVar(&patchCapacityRigs, "capacity-rigs", 0, "Capacity rigs")
	cmd.Flags().BoolVar(&clearMeetingLocation, "clear-meeting-location", false, "Clear meeting location")
	cmd.Flags().StringVar(&meetingLabel, "meeting-label", "", "Meeting location label (single line)")
	cmd.Flags().StringVar(&meetingAddress, "meeting-address", "", "Meeting location address (single line)")
	cmd.Flags().Float64Var(&meetingLat, "meeting-lat", 0, "Meeting latitude (requires --meeting-lng)")
	cmd.Flags().Float64Var(&meetingLng, "meeting-lng", 0, "Meeting longitude (requires --meeting-lat)")
	cmd.Flags().StringArrayVar(&artifactIDs, "artifact-id", nil, "Replace artifacts with this ordered list of artifact IDs (repeatable; deduped)")
	cmd.Flags().BoolVar(&clearArtifacts, "clear-artifacts", false, "Clear artifacts list")

	return cmd
}

func newTripVisibilityCmd(deps RootDeps) *cobra.Command {
	var (
		public         bool
		private        bool
		idempotencyKey string
	)
	cmd := &cobra.Command{
		Use:   "visibility <tripId> --public|--private",
		Short: "Set draft trip visibility",
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

			if public == private {
				return exitcode.New(exitcode.KindUsage, "choose exactly one of --public or --private", nil)
			}
			if strings.TrimSpace(idempotencyKey) == "" {
				idempotencyKey = idempotency.NewKey()
			}

			vis := gen.DraftVisibility("PRIVATE")
			if public {
				vis = gen.DraftVisibility("PUBLIC")
			}

			tripID := gen.TripId(args[0])
			resp, err := deps.PlannerAPI.SetTripDraftVisibility(ctx, apiCtx.APIURL, apiCtx.BearerToken, tripID, idempotencyKey, gen.SetDraftVisibilityRequest{DraftVisibility: vis})
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

	cmd.Flags().BoolVar(&public, "public", false, "Set draft visibility to PUBLIC")
	cmd.Flags().BoolVar(&private, "private", false, "Set draft visibility to PRIVATE")
	cmd.Flags().StringVar(&idempotencyKey, "idempotency-key", "", "Idempotency key (auto-generated if omitted)")
	return cmd
}

func newTripPublishCmd(deps RootDeps) *cobra.Command {
	var printAnnouncement bool
	cmd := &cobra.Command{
		Use:   "publish <tripId>",
		Short: "Publish a trip",
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

			tripID := gen.TripId(args[0])
			resp, err := deps.PlannerAPI.PublishTrip(ctx, apiCtx.APIURL, apiCtx.BearerToken, tripID)
			if err != nil {
				return err
			}

			if printAnnouncement {
				if resp.JSON200 != nil {
					_, _ = io.WriteString(deps.Stdout, resp.JSON200.AnnouncementCopy)
				}
				return nil
			}

			if resolved.Options.Output == cliopts.OutputJSON {
				return envelope.WriteJSON(deps.Stdout, envelope.Envelope{
					Data: resp.JSON200,
					Meta: envelope.Meta{APIURL: apiCtx.APIURL, Profile: apiCtx.Profile},
				})
			}

			_, _ = io.WriteString(deps.Stdout, "OK\n")
			return nil
		},
	}
	cmd.Flags().BoolVar(&printAnnouncement, "print-announcement", false, "Print announcement copy to stdout (plain text only)")
	return cmd
}

func newTripCancelCmd(deps RootDeps) *cobra.Command {
	var (
		force          bool
		idempotencyKey string
	)
	cmd := &cobra.Command{
		Use:   "cancel <tripId>",
		Short: "Cancel a trip",
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

			if !force {
				return exitcode.New(exitcode.KindUsage, "refusing to cancel without --force", nil)
			}

			var idemPtr *string
			if strings.TrimSpace(idempotencyKey) != "" {
				idemPtr = &idempotencyKey
			}

			tripID := gen.TripId(args[0])
			resp, err := deps.PlannerAPI.CancelTrip(ctx, apiCtx.APIURL, apiCtx.BearerToken, tripID, idemPtr)
			if err != nil {
				return err
			}

			if resolved.Options.Output == cliopts.OutputJSON {
				meta := envelope.Meta{APIURL: apiCtx.APIURL, Profile: apiCtx.Profile}
				if idemPtr != nil {
					meta.IdempotencyKey = idempotencyKey
				}
				return envelope.WriteJSON(deps.Stdout, envelope.Envelope{
					Data: resp.JSON200,
					Meta: meta,
				})
			}

			if idemPtr != nil {
				_, _ = fmt.Fprintf(deps.Stderr, "Idempotency-Key: %s\n", idempotencyKey)
			}
			_, _ = io.WriteString(deps.Stdout, "OK\n")
			return nil
		},
	}
	cmd.Flags().BoolVar(&force, "force", false, "Required: confirm trip cancellation")
	cmd.Flags().StringVar(&idempotencyKey, "idempotency-key", "", "Optional idempotency key (not auto-generated)")
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

func validateSingleLineFlag(v string, flagName string) error {
	if strings.Contains(v, "\n") || strings.Contains(v, "\r") {
		return fmt.Errorf("%s must be single line (no newlines)", flagName)
	}
	return nil
}

func dedupStringsPreserveFirst(in []string) []string {
	seen := make(map[string]struct{}, len(in))
	out := make([]string, 0, len(in))
	for _, v := range in {
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}
	return out
}
