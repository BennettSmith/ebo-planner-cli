# ADR 0001: Choose Go modules for v1 CLI framework + UX helpers

## Status

Accepted

## Context

Issue: `ebo-planner-cli#7`

The v1 CLI needs a small set of foundational dependencies to implement the CLI spec:

- A command framework (subcommands, flags, help, completion)
- Interactive prompting (`--prompt`)
- Editor-based input (`--edit`) and `--from-file`
- YAML config read/write (with forward compatibility)
- Human-friendly output (tables) and stable machine output (JSON)
- Color/no-color behavior
- Browser-open helper (for OAuth device login UX)
- OAuth 2.0 Device Authorization Grant (OIDC tenant support)
- OpenAPI client generation approach

We want boring, well-maintained libs that minimize future migration pain.

## Decision

### CLI framework

Use:

- `github.com/spf13/cobra` (commands, help, shell completion)
- `github.com/spf13/pflag` (flag parsing; already used by Cobra)

Rationale:

- Widely used in Go CLIs (including complex multi-command tools)
- First-class help/completion support
- Plenty of examples and ecosystem knowledge

### Interactive prompting (`--prompt`)

Use:

- `github.com/AlecAivazis/survey/v2`

Rationale:

- Mature, widely adopted, good UX for basic forms/selects/confirmations
- Easy to keep the CLI non-TUI (v1 is not a full-screen TUI)

### Editor invocation (`--edit`)

Use:

- `github.com/mattn/go-isatty` (detect TTY when needed)
- Standard library (`os/exec`, temp files, `$EDITOR`/`$VISUAL`)

Rationale:

- Prefer stdlib for editor execution to keep behavior explicit and testable
- Only use small helpers where it removes platform quirks (TTY detection)

### YAML config read/write (preserving unknown fields)

Use:

- `gopkg.in/yaml.v3`

Approach:

- Parse config into a `yaml.Node`.
- Read known fields for behavior.
- When writing, update only the nodes for known paths, leaving unknown keys intact.

Rationale:

- `yaml.v3` supports round-tripping via nodes
- Avoids surprising deletions when we add new fields in future versions

### Config location

Use:

- Standard library `os.UserConfigDir()` (baseline)
- Add `EBO_CONFIG_DIR` override per spec

Rationale:

- Spec already defines behavior; keep it simple and portable

### Human table output

Use:

- `github.com/olekukonko/tablewriter`

Rationale:

- Battle-tested for CLI tables
- Keeps default UX readable without committing to a TUI stack

### Color / no-color

Use:

- `github.com/muesli/termenv`

Rationale:

- Handles common terminal capability detection and `NO_COLOR` conventions
- Works well alongside table output and future styling needs

### Browser open (device login UX)

Use:

- `github.com/pkg/browser`

Rationale:

- Small, common helper for opening URLs on macOS/Linux/Windows

### OAuth 2.0 Device Authorization Grant

Use:

- Standard library HTTP + `golang.org/x/oauth2` for token representation where useful
- Implement Device Authorization Grant flow directly (a small internal client)

Rationale:

- Device flow varies slightly by provider; a small internal implementation keeps us in control
- Avoids adopting a larger identity SDK prematurely

### OpenAPI client generation

Use:

- `github.com/oapi-codegen/oapi-codegen/v2` (generate client + types)

Rationale:

- Backend already uses oapi-codegen; aligns generator semantics across repos
- Generates idiomatic Go client code with good compile-time typing

## Consequences

- **Pros**
  - Well-known libraries with lots of ecosystem familiarity
  - Minimizes bespoke plumbing while keeping us in control of critical flows (auth + config rewrite)
  - Compatible with spec-required behaviors (profiles, output contract, no-color, etc.)
- **Cons**
  - A few dependencies to learn (Cobra, Survey, termenv)
  - YAML node round-tripping adds some implementation complexity (but prevents future config breakage)

## Alternatives considered

- **Kong instead of Cobra**
  - Rejected (for v1): Cobra is more established for multi-command CLIs and completion.
- **Charm stack (Bubble Tea / Huh) for prompts**
  - Rejected (for v1): pushes us toward TUI patterns; v1 needs a simpler non-TUI CLI.
- **Viper for config**
  - Rejected: we need precise YAML shape + write-back that preserves unknown fields; Viper is read-focused.
- **OpenAPI Generator**
  - Rejected: heavier toolchain and less alignment with existing backend generator.

## Follow-up checklist (scaffolding work enabled by this ADR)

- Add `go.mod` + `go.sum` and basic `cmd/ebo` entrypoint (Cobra root)
- Add config package with YAML node round-trip strategy
- Implement global flags/env mapping (`--output`, `--no-color`, etc.)
- Implement `ebo auth login/logout/status/token print` with device flow
- Add oapi-codegen wiring + generated code location (exclude from coverage)
- Add `make ci` gates: unit tests, coverage >= 85% (non-generated), lint/staticcheck as desired


