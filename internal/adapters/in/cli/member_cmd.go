package cli

import (
	"errors"
	"fmt"
	"io"
	"net/mail"
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
	openapi_types "github.com/oapi-codegen/runtime/types"
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
	memberCmd.AddCommand(newMemberCreateCmd(deps))
	memberCmd.AddCommand(newMemberUpdateCmd(deps))
	root.AddCommand(memberCmd)
}

func newMemberCreateCmd(deps RootDeps) *cobra.Command {
	var (
		displayName     string
		email           string
		groupAliasEmail string

		vehicleMake             string
		vehicleModel            string
		vehicleTireSize         string
		vehicleLiftLockers      string
		vehicleFuelRange        string
		vehicleRecoveryGear     string
		vehicleHamRadioCallSign string
		vehicleNotes            string
	)

	cmd := &cobra.Command{
		Use:   "create --display-name <name> --email <email>",
		Short: "Create (provision) my member profile",
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

			rejectMultiline := func(label, v string) error {
				if strings.ContainsAny(v, "\r\n") {
					return exitcode.New(exitcode.KindUsage, fmt.Sprintf("%s must not be multi-line; use --from-file/--edit/--prompt instead", label), nil)
				}
				return nil
			}
			if err := rejectMultiline("--display-name", displayName); err != nil {
				return err
			}
			if err := rejectMultiline("--email", email); err != nil {
				return err
			}
			if err := rejectMultiline("--group-alias-email", groupAliasEmail); err != nil {
				return err
			}
			if err := rejectMultiline("--vehicle-make", vehicleMake); err != nil {
				return err
			}
			if err := rejectMultiline("--vehicle-model", vehicleModel); err != nil {
				return err
			}
			if err := rejectMultiline("--vehicle-tire-size", vehicleTireSize); err != nil {
				return err
			}
			if err := rejectMultiline("--vehicle-lift-lockers", vehicleLiftLockers); err != nil {
				return err
			}
			if err := rejectMultiline("--vehicle-fuel-range", vehicleFuelRange); err != nil {
				return err
			}
			if err := rejectMultiline("--vehicle-recovery-gear", vehicleRecoveryGear); err != nil {
				return err
			}
			if err := rejectMultiline("--vehicle-ham-radio-call-sign", vehicleHamRadioCallSign); err != nil {
				return err
			}
			if err := rejectMultiline("--vehicle-notes", vehicleNotes); err != nil {
				return err
			}

			if strings.TrimSpace(displayName) == "" {
				return exitcode.New(exitcode.KindUsage, "missing --display-name", nil)
			}
			if strings.TrimSpace(email) == "" {
				return exitcode.New(exitcode.KindUsage, "missing --email", nil)
			}
			if _, err := mail.ParseAddress(email); err != nil {
				return exitcode.New(exitcode.KindUsage, "invalid --email", err)
			}
			if strings.TrimSpace(groupAliasEmail) != "" {
				if _, err := mail.ParseAddress(groupAliasEmail); err != nil {
					return exitcode.New(exitcode.KindUsage, "invalid --group-alias-email", err)
				}
			}

			req := gen.CreateMemberRequest{
				DisplayName: strings.TrimSpace(displayName),
				Email:       openapi_types.Email(strings.TrimSpace(email)),
			}
			if strings.TrimSpace(groupAliasEmail) != "" {
				v := openapi_types.Email(strings.TrimSpace(groupAliasEmail))
				req.GroupAliasEmail = &v
			}

			// Vehicle profile is optional; include if any fields are set.
			var vp gen.VehicleProfile
			setAny := false
			if strings.TrimSpace(vehicleMake) != "" {
				v := strings.TrimSpace(vehicleMake)
				vp.Make = &v
				setAny = true
			}
			if strings.TrimSpace(vehicleModel) != "" {
				v := strings.TrimSpace(vehicleModel)
				vp.Model = &v
				setAny = true
			}
			if strings.TrimSpace(vehicleTireSize) != "" {
				v := strings.TrimSpace(vehicleTireSize)
				vp.TireSize = &v
				setAny = true
			}
			if strings.TrimSpace(vehicleLiftLockers) != "" {
				v := strings.TrimSpace(vehicleLiftLockers)
				vp.LiftLockers = &v
				setAny = true
			}
			if strings.TrimSpace(vehicleFuelRange) != "" {
				v := strings.TrimSpace(vehicleFuelRange)
				vp.FuelRange = &v
				setAny = true
			}
			if strings.TrimSpace(vehicleRecoveryGear) != "" {
				v := strings.TrimSpace(vehicleRecoveryGear)
				vp.RecoveryGear = &v
				setAny = true
			}
			if strings.TrimSpace(vehicleHamRadioCallSign) != "" {
				v := strings.TrimSpace(vehicleHamRadioCallSign)
				vp.HamRadioCallSign = &v
				setAny = true
			}
			if strings.TrimSpace(vehicleNotes) != "" {
				v := strings.TrimSpace(vehicleNotes)
				vp.Notes = &v
				setAny = true
			}
			if setAny {
				req.VehicleProfile = &vp
			}

			resp, err := deps.PlannerAPI.CreateMyMember(ctx, apiCtx.APIURL, apiCtx.BearerToken, req)
			if err != nil {
				return err
			}

			if resolved.Options.Output == cliopts.OutputJSON {
				return envelope.WriteJSON(deps.Stdout, envelope.Envelope{
					Data: resp.JSON201,
					Meta: envelope.Meta{APIURL: apiCtx.APIURL, Profile: apiCtx.Profile},
				})
			}

			if resp.JSON201 != nil {
				_, _ = fmt.Fprintf(deps.Stdout, "memberId=%s\n", resp.JSON201.Member.MemberId)
				return nil
			}
			_, _ = io.WriteString(deps.Stdout, "OK\n")
			return nil
		},
	}

	cmd.Flags().StringVar(&displayName, "display-name", "", "Display name")
	cmd.Flags().StringVar(&email, "email", "", "Email")
	cmd.Flags().StringVar(&groupAliasEmail, "group-alias-email", "", "Group alias email (optional)")

	cmd.Flags().StringVar(&vehicleMake, "vehicle-make", "", "Vehicle make")
	cmd.Flags().StringVar(&vehicleModel, "vehicle-model", "", "Vehicle model")
	cmd.Flags().StringVar(&vehicleTireSize, "vehicle-tire-size", "", "Vehicle tire size")
	cmd.Flags().StringVar(&vehicleLiftLockers, "vehicle-lift-lockers", "", "Vehicle lift/lockers")
	cmd.Flags().StringVar(&vehicleFuelRange, "vehicle-fuel-range", "", "Vehicle fuel range")
	cmd.Flags().StringVar(&vehicleRecoveryGear, "vehicle-recovery-gear", "", "Vehicle recovery gear")
	cmd.Flags().StringVar(&vehicleHamRadioCallSign, "vehicle-ham-radio-call-sign", "", "Vehicle HAM radio call sign")
	cmd.Flags().StringVar(&vehicleNotes, "vehicle-notes", "", "Vehicle notes (single-line only; use file/edit/prompt for multi-line)")

	return cmd
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

		displayName      string
		clearDisplayName bool
		email            string
		groupAliasEmail  string
		clearGroupAlias  bool

		vehicleMake             string
		vehicleModel            string
		vehicleTireSize         string
		vehicleLiftLockers      string
		vehicleFuelRange        string
		vehicleRecoveryGear     string
		vehicleHamRadioCallSign string
		vehicleNotes            string

		clearVehicle             bool
		clearVehicleMake         bool
		clearVehicleModel        bool
		clearVehicleTireSize     bool
		clearVehicleLiftLockers  bool
		clearVehicleFuelRange    bool
		clearVehicleRecoveryGear bool
		clearVehicleHamRadio     bool
		clearVehicleNotes        bool
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

			flagsMode := false
			for _, name := range []string{
				"display-name",
				"email",
				"group-alias-email",
				"vehicle-make",
				"vehicle-model",
				"vehicle-tire-size",
				"vehicle-lift-lockers",
				"vehicle-fuel-range",
				"vehicle-recovery-gear",
				"vehicle-ham-radio-call-sign",
				"vehicle-notes",
			} {
				if cmd.Flags().Changed(name) {
					flagsMode = true
				}
			}
			if clearDisplayName || clearGroupAlias || clearVehicle ||
				clearVehicleMake || clearVehicleModel || clearVehicleTireSize ||
				clearVehicleLiftLockers || clearVehicleFuelRange || clearVehicleRecoveryGear ||
				clearVehicleHamRadio || clearVehicleNotes {
				flagsMode = true
			}

			modeCount := 0
			if flagsMode {
				modeCount++
			}
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
				return exitcode.New(exitcode.KindUsage, "missing input (provide patch flags or use --from-file/--edit/--prompt)", nil)
			}
			if modeCount > 1 {
				return exitcode.New(exitcode.KindUsage, "choose exactly one patch mode: flags, --from-file, --edit, or --prompt", nil)
			}

			var req gen.UpdateMyMemberProfileRequest
			switch {
			case flagsMode:
				rejectMultiline := func(label, v string) error {
					if strings.ContainsAny(v, "\r\n") {
						return exitcode.New(exitcode.KindUsage, fmt.Sprintf("%s must not be multi-line; use --from-file/--edit/--prompt instead", label), nil)
					}
					return nil
				}

				if cmd.Flags().Changed("display-name") {
					if err := rejectMultiline("--display-name", displayName); err != nil {
						return err
					}
					v := strings.TrimSpace(displayName)
					if v == "" {
						return exitcode.New(exitcode.KindUsage, "--display-name must be non-empty (use --clear-display-name to clear)", nil)
					}
					req.DisplayName = &v
				}
				if clearDisplayName {
					if cmd.Flags().Changed("display-name") {
						return exitcode.New(exitcode.KindUsage, "cannot use --display-name with --clear-display-name", nil)
					}
					empty := ""
					req.DisplayName = &empty
				}

				if cmd.Flags().Changed("email") {
					if err := rejectMultiline("--email", email); err != nil {
						return err
					}
					v := strings.TrimSpace(email)
					if v == "" {
						return exitcode.New(exitcode.KindUsage, "--email must be non-empty", nil)
					}
					if _, err := mail.ParseAddress(v); err != nil {
						return exitcode.New(exitcode.KindUsage, "invalid --email", err)
					}
					ev := openapi_types.Email(v)
					req.Email = &ev
				}

				if cmd.Flags().Changed("group-alias-email") {
					if err := rejectMultiline("--group-alias-email", groupAliasEmail); err != nil {
						return err
					}
					v := strings.TrimSpace(groupAliasEmail)
					if v == "" {
						return exitcode.New(exitcode.KindUsage, "--group-alias-email must be non-empty (use --clear-group-alias-email to clear)", nil)
					}
					if _, err := mail.ParseAddress(v); err != nil {
						return exitcode.New(exitcode.KindUsage, "invalid --group-alias-email", err)
					}
					ev := openapi_types.Email(v)
					req.GroupAliasEmail = &ev
				}
				if clearGroupAlias {
					if cmd.Flags().Changed("group-alias-email") {
						return exitcode.New(exitcode.KindUsage, "cannot use --group-alias-email with --clear-group-alias-email", nil)
					}
					empty := openapi_types.Email("")
					req.GroupAliasEmail = &empty
				}

				vehicleAnySetOrClear := false
				if cmd.Flags().Changed("vehicle-make") || cmd.Flags().Changed("vehicle-model") || cmd.Flags().Changed("vehicle-tire-size") ||
					cmd.Flags().Changed("vehicle-lift-lockers") || cmd.Flags().Changed("vehicle-fuel-range") || cmd.Flags().Changed("vehicle-recovery-gear") ||
					cmd.Flags().Changed("vehicle-ham-radio-call-sign") || cmd.Flags().Changed("vehicle-notes") ||
					clearVehicleMake || clearVehicleModel || clearVehicleTireSize || clearVehicleLiftLockers || clearVehicleFuelRange ||
					clearVehicleRecoveryGear || clearVehicleHamRadio || clearVehicleNotes || clearVehicle {
					vehicleAnySetOrClear = true
				}

				if clearVehicle && (cmd.Flags().Changed("vehicle-make") || cmd.Flags().Changed("vehicle-model") || cmd.Flags().Changed("vehicle-tire-size") ||
					cmd.Flags().Changed("vehicle-lift-lockers") || cmd.Flags().Changed("vehicle-fuel-range") || cmd.Flags().Changed("vehicle-recovery-gear") ||
					cmd.Flags().Changed("vehicle-ham-radio-call-sign") || cmd.Flags().Changed("vehicle-notes") ||
					clearVehicleMake || clearVehicleModel || clearVehicleTireSize || clearVehicleLiftLockers || clearVehicleFuelRange ||
					clearVehicleRecoveryGear || clearVehicleHamRadio || clearVehicleNotes) {
					return exitcode.New(exitcode.KindUsage, "--clear-vehicle cannot be combined with other vehicle flags", nil)
				}

				if vehicleAnySetOrClear {
					var vp gen.VehicleProfile
					empty := ""
					if clearVehicle {
						vp = gen.VehicleProfile{
							Make:             &empty,
							Model:            &empty,
							TireSize:         &empty,
							LiftLockers:      &empty,
							FuelRange:        &empty,
							RecoveryGear:     &empty,
							HamRadioCallSign: &empty,
							Notes:            &empty,
						}
						req.VehicleProfile = &vp
					} else {
						setField := func(flag string, label string, raw string, dst **string) error {
							if !cmd.Flags().Changed(flag) {
								return nil
							}
							if err := rejectMultiline(label, raw); err != nil {
								return err
							}
							v := strings.TrimSpace(raw)
							if v == "" {
								return exitcode.New(exitcode.KindUsage, fmt.Sprintf("%s must be non-empty (use the corresponding --clear-* flag to clear)", label), nil)
							}
							*dst = &v
							return nil
						}
						if err := setField("vehicle-make", "--vehicle-make", vehicleMake, &vp.Make); err != nil {
							return err
						}
						if err := setField("vehicle-model", "--vehicle-model", vehicleModel, &vp.Model); err != nil {
							return err
						}
						if err := setField("vehicle-tire-size", "--vehicle-tire-size", vehicleTireSize, &vp.TireSize); err != nil {
							return err
						}
						if err := setField("vehicle-lift-lockers", "--vehicle-lift-lockers", vehicleLiftLockers, &vp.LiftLockers); err != nil {
							return err
						}
						if err := setField("vehicle-fuel-range", "--vehicle-fuel-range", vehicleFuelRange, &vp.FuelRange); err != nil {
							return err
						}
						if err := setField("vehicle-recovery-gear", "--vehicle-recovery-gear", vehicleRecoveryGear, &vp.RecoveryGear); err != nil {
							return err
						}
						if err := setField("vehicle-ham-radio-call-sign", "--vehicle-ham-radio-call-sign", vehicleHamRadioCallSign, &vp.HamRadioCallSign); err != nil {
							return err
						}
						if err := setField("vehicle-notes", "--vehicle-notes", vehicleNotes, &vp.Notes); err != nil {
							return err
						}

						if clearVehicleMake {
							vp.Make = &empty
						}
						if clearVehicleModel {
							vp.Model = &empty
						}
						if clearVehicleTireSize {
							vp.TireSize = &empty
						}
						if clearVehicleLiftLockers {
							vp.LiftLockers = &empty
						}
						if clearVehicleFuelRange {
							vp.FuelRange = &empty
						}
						if clearVehicleRecoveryGear {
							vp.RecoveryGear = &empty
						}
						if clearVehicleHamRadio {
							vp.HamRadioCallSign = &empty
						}
						if clearVehicleNotes {
							vp.Notes = &empty
						}

						req.VehicleProfile = &vp
					}
				}

				// Must change at least one field.
				if req.DisplayName == nil && req.Email == nil && req.GroupAliasEmail == nil && req.VehicleProfile == nil {
					return exitcode.New(exitcode.KindUsage, "no changes specified", nil)
				}
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

	cmd.Flags().StringVar(&displayName, "display-name", "", "Display name")
	cmd.Flags().BoolVar(&clearDisplayName, "clear-display-name", false, "Clear display name")
	cmd.Flags().StringVar(&email, "email", "", "Email")
	cmd.Flags().StringVar(&groupAliasEmail, "group-alias-email", "", "Group alias email")
	cmd.Flags().BoolVar(&clearGroupAlias, "clear-group-alias-email", false, "Clear group alias email")

	cmd.Flags().StringVar(&vehicleMake, "vehicle-make", "", "Vehicle make")
	cmd.Flags().StringVar(&vehicleModel, "vehicle-model", "", "Vehicle model")
	cmd.Flags().StringVar(&vehicleTireSize, "vehicle-tire-size", "", "Vehicle tire size")
	cmd.Flags().StringVar(&vehicleLiftLockers, "vehicle-lift-lockers", "", "Vehicle lift/lockers")
	cmd.Flags().StringVar(&vehicleFuelRange, "vehicle-fuel-range", "", "Vehicle fuel range")
	cmd.Flags().StringVar(&vehicleRecoveryGear, "vehicle-recovery-gear", "", "Vehicle recovery gear")
	cmd.Flags().StringVar(&vehicleHamRadioCallSign, "vehicle-ham-radio-call-sign", "", "Vehicle HAM radio call sign")
	cmd.Flags().StringVar(&vehicleNotes, "vehicle-notes", "", "Vehicle notes (single-line only; use file/edit/prompt for multi-line)")

	cmd.Flags().BoolVar(&clearVehicle, "clear-vehicle", false, "Clear the entire vehicle profile")
	cmd.Flags().BoolVar(&clearVehicleMake, "clear-vehicle-make", false, "Clear vehicle make")
	cmd.Flags().BoolVar(&clearVehicleModel, "clear-vehicle-model", false, "Clear vehicle model")
	cmd.Flags().BoolVar(&clearVehicleTireSize, "clear-vehicle-tire-size", false, "Clear vehicle tire size")
	cmd.Flags().BoolVar(&clearVehicleLiftLockers, "clear-vehicle-lift-lockers", false, "Clear vehicle lift/lockers")
	cmd.Flags().BoolVar(&clearVehicleFuelRange, "clear-vehicle-fuel-range", false, "Clear vehicle fuel range")
	cmd.Flags().BoolVar(&clearVehicleRecoveryGear, "clear-vehicle-recovery-gear", false, "Clear vehicle recovery gear")
	cmd.Flags().BoolVar(&clearVehicleHamRadio, "clear-vehicle-ham-radio-call-sign", false, "Clear vehicle HAM radio call sign")
	cmd.Flags().BoolVar(&clearVehicleNotes, "clear-vehicle-notes", false, "Clear vehicle notes")

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
