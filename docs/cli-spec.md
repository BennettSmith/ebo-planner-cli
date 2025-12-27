# ebo-planner-cli — CLI Specification (Draft)

Status: **DRAFT (v1 conventions finalized; some per-command patch flags remain “(proposed)”)**

This document specifies the required behavior and command surface of the **East Bay Overland Trip Planning CLI**.

It is **spec-first**: API behavior and validation rules come from the spec repo and are consumed via `spec.lock`.
The CLI must not rely on undocumented service behavior.

---

## Scope

- **In scope**:
  - Commands that allow a member/organizer to perform **all v1 use cases** from the spec repo
  - CLI-only supporting workflows (auth/login/logout, config/profiles, output formatting)
  - Scriptability guarantees (stable JSON output, stable exit codes)
- **Out of scope**:
  - Service-side business rules not in the spec
  - Direct DB access
  - Non-API “admin” backdoors

---

## Terminology

- **Service / API**: the ebo-planner backend HTTP API defined by `ebo-planner-spec/openapi/openapi.yaml`.
- **Member**: authenticated caller. Some endpoints require the caller to also be a *provisioned* member profile.
- **Organizer**: member in a trip’s organizer set.
- **Draft visibility**:
  - `PRIVATE` drafts: visible only to creator
  - `PUBLIC` drafts: visible only to organizers

---

## Open design decisions (questions to answer)

Resolved decisions:

- **Binary name**: `ebo`
- **Auth UX**: interactive `ebo auth login`
- **Profiles**: required; `--profile <name>` supported
- **Default output**: human table by default; machine output via `--output json`
- **Idempotency keys**: auto-generate when omitted for operations requiring idempotency
- **Destructive confirmations**: require `--force` for destructive operations

Open questions:

None (v1 decisions below).

---

## Global behavior

### CLI identity

- **Binary name**: `ebo`

### Authentication requirement

- All commands that call the API require authentication (bearer JWT).
- Some commands (notably `/members/me`) are allowed for authenticated-but-not-provisioned callers; see UC-17/UC-18.

### Standard flags (all commands)

The CLI MUST support these flags on every command (either at root or inherited):

- `--api-url <url>`: override API base URL (default from config; see Profiles)
- `--profile <name>`: select a named profile (default: `default`)
- `--output <format>`: `table|json` (default: `table`)
- `--no-color`: disable ANSI coloring
- `--timeout <duration>`: request timeout (e.g., `10s`, `2m`)
- `--verbose`: verbose HTTP/debug logging to stderr (never to stdout)

Environment variable equivalents (MUST be supported):

- `EBO_API_URL` (equivalent to `--api-url`)
- `EBO_PROFILE` (equivalent to `--profile`)
- `EBO_OUTPUT` (equivalent to `--output`)
- `EBO_NO_COLOR=1` (equivalent to `--no-color`)
- `EBO_TIMEOUT` (equivalent to `--timeout`)
- `EBO_VERBOSE=1` (equivalent to `--verbose`)

### Exit codes

Minimum required exit code contract:

- `0`: success
- `1`: generic failure (unexpected)
- `2`: usage error (invalid flags/args)
- `3`: auth error (missing/invalid token; HTTP 401)
- `4`: not found (HTTP 404)
- `5`: conflict (HTTP 409)
- `6`: validation failed (HTTP 422)
- `7`: server error (HTTP 5xx or network error)

### Output contract

- **Human output**: intended for terminals. May include headings, tables, and short guidance.
- **Human errors**: MUST be multi-line and include actionable guidance when possible (e.g., “Try: `ebo member create ...`”).
- **JSON output**:
  - MUST be valid JSON written to stdout
  - SHOULD be a single JSON object per invocation (not NDJSON) unless explicitly designed otherwise
  - MUST NOT include ANSI color codes
  - SHOULD include a stable envelope:
    - `data`: the API response payload (or a CLI-defined payload for non-API commands)
    - `meta`: request metadata (e.g., `requestId`, `idempotencyKey`, `apiUrl`) when available
    - `error`: present only on failure; matches API error shape when available

### Idempotency contract

For mutating API operations where the OpenAPI spec requires `Idempotency-Key`:

- The CLI MUST provide a way to set the idempotency key per invocation (e.g., `--idempotency-key`).
- If `--idempotency-key` is omitted, the CLI MUST auto-generate one.
- The generated idempotency key MUST be:
  - included in JSON output as `meta.idempotencyKey`
  - printed to stderr when `--output table` (so stdout stays clean for piping)

### Destructive operation confirmation

For destructive operations, the CLI MUST require an explicit `--force` flag.

- If `--force` is omitted, the CLI MUST:
  - exit with code `2` (usage error)
  - print an actionable error message to stderr (e.g., “Refusing to cancel without --force”)

### Text fields (plain text only)

For user-supplied free-text fields (e.g., trip `name`, `description`, `difficultyText`, comms text, recommended text, and location label/address):

- The CLI MUST treat inputs as **plain text**.
- Multi-line text MUST be accepted where applicable (e.g., `--description` and editor/file modes).
- The CLI MUST NOT support Markdown or HTML rendering/preview features in v1.
- The CLI MAY warn (but MUST NOT transform) if input appears to contain HTML/Markdown.

---

## Config and profiles

Profiles are **required**.

### Storage location (normative)

The CLI MUST store its configuration in the OS user config directory:

- Determine `CONFIG_DIR` using the platform’s standard user config directory resolution (e.g., Go `os.UserConfigDir()`).
- Store config under:
  - `CONFIG_DIR/ebo/config.yaml`

If `EBO_CONFIG_DIR` is set, the CLI MUST use it as `CONFIG_DIR` (overriding OS resolution).

The CLI MUST NOT write secrets to project directories or the working directory by default.

### File permissions (normative)

- Config files that contain credentials MUST be created with restrictive permissions (best effort):
  - POSIX: `0600`
- The CLI MUST NOT print tokens in normal output (only `ebo auth token print`).

### Config schema (normative)

The config file MUST be YAML with this shape (keys are normative):

- `currentProfile: string` (default: `default`)
- `profiles: map[string]Profile`

Where `Profile` has:

- `apiUrl: string` (required)
- `auth: object` (optional)
  - `accessToken: string` (optional; bearer token used for API calls)
  - `tokenType: string` (optional; default `Bearer`)
  - `expiresAt: string` (optional; RFC3339 timestamp)

Notes:

- The CLI MAY store additional fields, but MUST preserve unknown fields when rewriting config (forward compatibility).

### Required profile behavior (normative)

- The CLI MUST support `--profile <name>` (default: `default`).
- Each profile MUST have its own:
  - `apiUrl`
  - auth session/token fields
- `--api-url` / `EBO_API_URL` MUST override the selected profile’s `apiUrl` for that invocation only (no persistence).

### Precedence rules (normative)

When resolving effective settings:

1. CLI flags (highest precedence)
2. Environment variables
3. Config file (`currentProfile` + selected profile values)
4. Built-in defaults (lowest precedence)

If no `apiUrl` can be resolved for a command that requires the API, the CLI MUST fail with exit code `2` and guidance to set it (e.g., `ebo profile set ... --api-url ...`).

### Minimum required config items

- API base URL
- stored access token (or token retrieval method)

- API base URL (per profile: `profiles.<name>.apiUrl`)
- stored access token/session (per profile: `profiles.<name>.auth.*`)

---

## Command inventory (must cover all v1 use cases)

This table is normative: the CLI MUST provide commands that map to these API operations.

| Use case | OpenAPI operationId | Endpoint | Required CLI command (proposed) |
| --- | --- | --- | --- |
| UC-01 list visible trips | `listVisibleTripsForMember` | `GET /trips` | `trip list` |
| UC-01 list my draft trips | `listMyDraftTrips` | `GET /trips/drafts` | `trip drafts` |
| UC-02 get trip details | `getTripDetails` | `GET /trips/{tripId}` | `trip get TRIP_ID` |
| UC-03 create draft | `createTripDraft` | `POST /trips` | `trip create --name NAME` |
| UC-04 update draft | `updateTrip` | `PATCH /trips/{tripId}` | `trip update TRIP_ID [patch flags]` |
| UC-05 set draft visibility | `setTripDraftVisibility` | `PUT /trips/{tripId}/draft-visibility` | `trip visibility TRIP_ID --public\|--private` |
| UC-06 publish trip | `publishTrip` | `POST /trips/{tripId}/publish` | `trip publish TRIP_ID` |
| UC-07 update published | `updateTrip` | `PATCH /trips/{tripId}` | `trip update TRIP_ID [patch flags]` |
| UC-08 cancel trip | `cancelTrip` | `POST /trips/{tripId}/cancel` | `trip cancel TRIP_ID` |
| UC-09 add organizer | `addTripOrganizer` | `POST /trips/{tripId}/organizers` | `trip organizer add TRIP_ID --member MEMBER_ID` |
| UC-10 remove organizer | `removeTripOrganizer` | `DELETE /trips/{tripId}/organizers/{memberId}` | `trip organizer remove TRIP_ID --member MEMBER_ID` |
| UC-11 set my RSVP | `setMyRSVP` | `PUT /trips/{tripId}/rsvp` | `trip rsvp set TRIP_ID --yes\|--no\|--unset` |
| UC-12 RSVP summary | `getTripRSVPSummary` | `GET /trips/{tripId}/rsvps` | `trip rsvp summary TRIP_ID` |
| UC-13 get my RSVP | `getMyRSVPForTrip` | `GET /trips/{tripId}/rsvp/me` | `trip rsvp get TRIP_ID` |
| UC-14 list members | `listMembers` | `GET /members` | `member list [--include-inactive]` |
| UC-15 search members | `searchMembers` | `GET /members/search?q=...` | `member search QUERY` |
| UC-16 update my profile | `updateMyMemberProfile` | `PATCH /members/me` | `member update [patch flags]` |
| UC-17 get my profile | `getMyMemberProfile` | `GET /members/me` | `member me` |
| UC-18 create my member | `createMyMember` | `POST /members` | `member create --display-name DISPLAY_NAME --email EMAIL` |

---

## CLI-only command requirements

These commands are required to make the CLI usable, even though they do not correspond to a v1 use case.

### `auth` commands (required)

The CLI MUST provide:

- `auth login` (interactive)
- `auth logout`
- `auth status` (prints whether a token is configured and which profile is active)
- `auth token set --token <jwt>` (non-interactive path to configure credentials)
- `auth token print` (prints the current token to stdout only when explicitly requested; never print tokens by default)

`auth login` requirements:

- MUST be interactive.
- MUST initiate a browser-based or device-code OIDC login flow (exact mechanism is an implementation detail).
- MUST store the resulting bearer access token in the active profile.

### `profile` and `config` commands (required)

The CLI MUST provide a stable, scriptable interface for managing profiles and config.

#### `ebo profile` (required)

- `ebo profile list`
  - Lists known profiles.
- `ebo profile show [PROFILE]`
  - Shows resolved profile config (default: current profile).
- `ebo profile create <PROFILE> --api-url <url>`
  - Creates a new profile. MUST fail if the profile already exists (exit `5`).
- `ebo profile set <PROFILE> --api-url <url>`
  - Sets the base URL for a profile. MUST create the profile if it does not exist.
- `ebo profile use <PROFILE>`
  - Sets `currentProfile`. MUST fail if profile does not exist (exit `2`).
- `ebo profile delete <PROFILE>`
  - Deletes a profile. MUST fail (exit `5`) if deleting would leave zero profiles.
  - If deleting the current profile, the CLI MUST set `currentProfile` to `default` if it exists, otherwise fail with guidance.

#### `ebo config` (required)

These are low-level escape hatches for automation and debugging. Keys are dot-paths into the YAML schema.

- `ebo config path`
  - Prints the config file path to stdout.
- `ebo config get <key>`
  - Prints the value for `<key>` to stdout. If key not found, exit `4`.
- `ebo config set <key> <value>`
  - Sets a config value (stringly-typed). MUST create missing objects/maps along the path.
- `ebo config unset <key>`
  - Removes a key (no-op if missing).
- `ebo config list`
  - Prints the entire config file.

Secrets redaction rules (normative):

- By default, `ebo config list` MUST redact secret values in all output formats.
  - Redaction string: `REDACTED`
- A new flag `--include-secrets` MUST be supported on `ebo config list` only:
  - If provided, secrets MAY be included in output.
  - If `--output table`, secrets MUST still be redacted (to reduce accidental screen leaks).
  - If `--output json`, secrets MAY be included when `--include-secrets` is set.

Required keys that MUST be supported by `ebo config get/set`:

- `currentProfile`
- `profiles.<name>.apiUrl`
- `profiles.<name>.auth.accessToken`
- `profiles.<name>.auth.tokenType`
- `profiles.<name>.auth.expiresAt`

---

## Worked examples

These examples are normative demonstrations of intended UX. Exact formatting of tables may vary, but behavior and flag semantics MUST match.

### Create and use profiles with different base URLs

Create profiles:

```bash
ebo profile create dev --api-url http://localhost:8081
ebo profile create staging --api-url https://staging-api.eastbayoverland.com
```

Select a profile for subsequent commands:

```bash
ebo profile use dev
ebo auth login
ebo trip list
```

Run a single command against a different profile without changing `currentProfile`:

```bash
ebo --profile staging trip list
```

### Override base URL for a single invocation

```bash
ebo --profile staging --api-url http://localhost:8081 trip list
```

Same via env vars:

```bash
EBO_PROFILE=staging EBO_API_URL=http://localhost:8081 ebo trip list
```

### Configure base URL via profile vs config

Preferred (profile command):

```bash
ebo profile set staging --api-url https://staging-api.eastbayoverland.com
```

Low-level escape hatch (config command):

```bash
ebo config set profiles.staging.apiUrl https://staging-api.eastbayoverland.com
```

### Create a trip from a file (YAML)

`trip.yaml`:

```yaml
name: "Snow Run Planning Day"
```

Create the draft:

```bash
ebo trip create --from-file trip.yaml
```

### Update a trip from a file (YAML patch)

`trip-patch.yaml`:

```yaml
description: |-
  Meet at the usual spot.
  Bring full fuel and recovery gear.
startDate: "2026-01-10"
endDate: "2026-01-10"
```

Apply patch:

```bash
ebo trip update TRIP_ID --from-file trip-patch.yaml
```

### Update a trip using the editor

```bash
ebo trip update TRIP_ID --edit
```

### Update a trip using interactive prompt mode

```bash
ebo trip update TRIP_ID --prompt
```

### Publish a trip and print announcement copy

To print announcement copy, `--print-announcement` is required and stdout MUST contain only the announcement:

```bash
ebo trip publish TRIP_ID --print-announcement
```

### Scripts: JSON output and idempotency metadata

```bash
ebo --output json trip create --name "Test Trip" | jq .
```

- For operations requiring idempotency, if you omit `--idempotency-key`, the CLI auto-generates one and includes it in JSON at `meta.idempotencyKey`.

## Command specifications (proposed)

This section defines commands and flags. Items labeled **(proposed)** remain open; everything else is normative.

### Trips

#### `trip list`

- **Maps to**: `GET /trips` (`listVisibleTripsForMember`)
- **Description**: Lists trips in `PUBLISHED` or `CANCELED` visible to the authenticated member.
- **Options**:
  - `--output table|json`

#### `trip drafts`

- **Maps to**: `GET /trips/drafts` (`listMyDraftTrips`)
- **Description**: Lists `DRAFT` trips where the caller is authorized (creator for private drafts; organizer for public drafts).

#### `trip get <tripId>`

- **Maps to**: `GET /trips/{tripId}` (`getTripDetails`)
- **Description**: Prints trip details if visible; otherwise surfaces a 404.

#### `trip create --name <name>`

- **Maps to**: `POST /trips` (`createTripDraft`)
- **Idempotency**: requires `Idempotency-Key` (see Idempotency contract)
- **Description**: Creates a draft trip with `draftVisibility=PRIVATE`. Creator becomes initial organizer.
- **Options**:
  - `--name <string>` (required)
  - `--idempotency-key <string>` (optional; auto-generated if omitted)
  - `--from-file <path>` (optional; JSON or YAML; see File-based requests)
  - `--prompt` (optional; interactive guided entry; see Interactive prompt mode)

Notes:

- Exactly one of `--name`, `--from-file`, or `--prompt` MUST be used.

#### `trip update <tripId>`

- **Maps to**: `PATCH /trips/{tripId}` (`updateTrip`)
- **Idempotency**: requires `Idempotency-Key` (see Idempotency contract)
- **Description**: Applies a partial update to a trip (draft or published). For drafts, auth depends on draft visibility; for published, caller must be organizer.
- **Patch options (proposed)**:
  - `--name <string>`
  - `--description <string>` / `--clear-description`
  - `--start-date <YYYY-MM-DD>` / `--clear-start-date`
  - `--end-date <YYYY-MM-DD>` / `--clear-end-date`
  - `--capacity-rigs <int>` / `--clear-capacity-rigs`
  - `--difficulty <string>` / `--clear-difficulty`
  - `--meeting-label <string>`
  - `--meeting-address <string>` / `--clear-meeting-address`
  - `--meeting-lat <float>` / `--meeting-lng <float>` / `--clear-meeting-latlng`
  - `--comms <string>` / `--clear-comms`
  - `--recommended <string>` / `--clear-recommended`
  - `--artifact-id <id>` (repeatable; replaces ordered list) / `--clear-artifacts`
  - `--idempotency-key <string>` (optional; auto-generated if omitted)
  - `--from-file <path>` (optional; JSON or YAML; see File-based requests)
  - `--edit` (optional; open $EDITOR; see Editor mode)
  - `--prompt` (optional; interactive guided entry; see Interactive prompt mode)

Notes:

- The CLI MUST support three patch input modes:
  - flags (individual fields)
  - file-based patch (`--from-file`)
  - editor mode (`--edit`)
- The CLI SHOULD support `--prompt` as a convenience for quick entry.
- If multiple patch modes are specified together, the CLI MUST error (exit code `2`).

#### `trip visibility <tripId> --public|--private`

- **Maps to**: `PUT /trips/{tripId}/draft-visibility` (`setTripDraftVisibility`)
- **Idempotency**: requires `Idempotency-Key` (see Idempotency contract)
- **Description**: Sets draft visibility. **Creator-only**.

#### `trip publish <tripId>`

- **Maps to**: `POST /trips/{tripId}/publish` (`publishTrip`)
- **Description**: Publishes a PUBLIC draft after validating required fields; returns `announcementCopy`.
- **Options**:
  - `--print-announcement` (required to output announcement copy; see Output behavior below)

Output behavior:

- If `--print-announcement` is provided:
  - stdout MUST contain **only** the announcement copy (plain text).
  - the CLI MUST still exit `0` on success.
- If `--print-announcement` is omitted:
  - the CLI MUST NOT print announcement copy (even if returned by the API).
  - normal `--output table|json` output applies.

#### `trip cancel <tripId>`

- **Maps to**: `POST /trips/{tripId}/cancel` (`cancelTrip`)
- **Idempotency**: API allows optional idempotency key.
- **CLI behavior**:
  - If `--idempotency-key` is provided, the CLI MUST send it.
  - If omitted, the CLI MUST NOT auto-generate one for this operation.
- **Description**: Cancels a trip (draft or published). Organizer-only. Cancellation is idempotent.
- **Options**:
  - `--force` (required)
  - `--idempotency-key <string>` (optional)

#### `trip organizer add <tripId> --member <memberId>`

- **Maps to**: `POST /trips/{tripId}/organizers` (`addTripOrganizer`)
- **Idempotency**: requires `Idempotency-Key` (see Idempotency contract)

#### `trip organizer remove <tripId> --member <memberId>`

- **Maps to**: `DELETE /trips/{tripId}/organizers/{memberId}` (`removeTripOrganizer`)
- **Idempotency**: requires `Idempotency-Key` (see Idempotency contract)
- **Options**:
  - `--force` (required)

#### `trip rsvp set <tripId> --yes|--no|--unset`

- **Maps to**: `PUT /trips/{tripId}/rsvp` (`setMyRSVP`)
- **Idempotency**: requires `Idempotency-Key` (see Idempotency contract)
- **Description**: Sets caller’s RSVP. Only allowed for `PUBLISHED` trips; YES enforces capacity.

#### `trip rsvp get <tripId>`

- **Maps to**: `GET /trips/{tripId}/rsvp/me` (`getMyRSVPForTrip`)

#### `trip rsvp summary <tripId>`

- **Maps to**: `GET /trips/{tripId}/rsvps` (`getTripRSVPSummary`)

---

### Members

#### `member list`

- **Maps to**: `GET /members` (`listMembers`)
- **Options**:
  - `--include-inactive`

#### `member search <query>`

- **Maps to**: `GET /members/search?q=...` (`searchMembers`)

#### `member me`

- **Maps to**: `GET /members/me` (`getMyMemberProfile`)
- **Behavior**:
  - If `MEMBER_NOT_PROVISIONED`, exit code MUST reflect 404 and JSON/human output should guide users to `member create`.

#### `member create --display-name <name> --email <email> [--group-alias-email <email>] [vehicle flags...]`

- **Maps to**: `POST /members` (`createMyMember`)
- **Description**: Provisions the caller’s member record.

#### `member update`

- **Maps to**: `PATCH /members/me` (`updateMyMemberProfile`)
- **Idempotency**: requires `Idempotency-Key` (see Idempotency contract)
- **Patch options (proposed)**:
  - `--display-name <string>` / `--clear-display-name` (note: server requires non-empty if provided)
  - `--email <email>` (cannot be cleared)
  - `--group-alias-email <email>` / `--clear-group-alias-email`
  - `--vehicle-make <string>` / `--vehicle-model <string>` / `--vehicle-tire-size <string>` / `--vehicle-notes <string>` (etc.)
  - `--clear-vehicle`
  - `--idempotency-key <string>` (optional; auto-generated if omitted)
  - `--from-file <path>` (optional; JSON or YAML; see File-based requests)
  - `--edit` (optional; open $EDITOR; see Editor mode)
  - `--prompt` (optional; interactive guided entry; see Interactive prompt mode)

Notes:

- If multiple patch modes are specified together, the CLI MUST error (exit code `2`).

---

## File-based requests

When `--from-file <path>` is provided:

- The CLI MUST support **JSON** and **YAML**.
- Format detection:
  - `.json` => JSON
  - `.yaml`/`.yml` => YAML
  - otherwise: attempt JSON first, then YAML; if both fail, exit `2`.
- The file content MUST map to the corresponding request shape:
  - `trip create`: `CreateTripDraftRequest`
  - `trip update`: `UpdateTripRequest`
  - `member update`: `UpdateMyMemberProfileRequest`
- Multi-line plain text MUST be preserved exactly (no markdown/html processing).

---

## Editor mode

When `--edit` is provided:

- The CLI MUST open the user’s editor:
  - `$EBO_EDITOR` if set, else `$EDITOR`, else fall back to a sensible default (implementation detail).
- The CLI MUST present a JSON or YAML template of the request (format selection is an implementation detail, but MUST be deterministic).
- The CLI MUST parse the edited content as JSON/YAML and submit it as the request.
- The CLI MUST NOT perform markdown/html rendering; the buffer is plain text.

---

## Interactive prompt mode (quick entry)

When `--prompt` is provided:

- The CLI MUST guide the user through entering fields **one at a time**.
- It MUST support multi-line entry for text fields (implementation detail: repeated prompts or editor handoff are acceptable).
- It MUST build the same JSON/YAML request shape as the file/editor modes and submit it.
