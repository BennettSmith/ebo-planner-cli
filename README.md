# ebo-planner-cli

`ebo` is a command-line interface for interacting with the **East Bay Overland Trip Planning** service.

- Normative behavior: `docs/cli-spec.md`
- Releasing + spec pinning: `docs/releasing.md`

## Quickstart

### Build

```bash
go build ./cmd/ebo
./ebo --help
```

### Configure an API URL (profiles are required)

Profiles store settings like the API base URL and auth token.

```bash
./ebo profile set default --api-url https://api.example.com
./ebo profile use default
```

Environment-variable equivalents are supported for all commands:

- `EBO_API_URL` (equivalent to `--api-url`)
- `EBO_PROFILE` (equivalent to `--profile`)
- `EBO_OUTPUT` (equivalent to `--output`)
- `EBO_NO_COLOR=1` (equivalent to `--no-color`)
- `EBO_TIMEOUT` (equivalent to `--timeout`)
- `EBO_VERBOSE=1` (equivalent to `--verbose`)
- `EBO_CONFIG_DIR` (override config directory)

### Authenticate

Interactive login (OIDC device flow):

```bash
./ebo auth login
./ebo auth status
```

Or set a token directly:

```bash
./ebo auth token set --token "$JWT"
```

## Output modes

- Default is human-friendly output (`--output table`).
- For scripting, use `--output json` (stable envelope; no ANSI; stdout-only JSON).

Example:

```bash
./ebo --output json trip list | jq .
```

## Common workflows (examples)

### Member profile

Some endpoints require you to be a **provisioned member profile**.

```bash
./ebo member me
./ebo member create --display-name "My Name" --email me@example.com
./ebo member update --display-name "New Name"
```

Notes:

- `member create` intentionally does **not** accept `--idempotency-key` (natural retry is handled by the API).
- `member update` auto-generates an idempotency key when omitted; in table mode it prints `Idempotency-Key: ...` to **stderr** (stdout remains script-safe).

### Trips (read)

```bash
./ebo trip list
./ebo trip drafts
./ebo trip get t1
```

### Trips (create/update)

Create a draft (name required):

```bash
./ebo trip create --name "Snow Run"
```

Update a draft using flags:

```bash
./ebo trip update t1 --name "Snow Run (updated)"
```

For multi-line fields, use an input mode instead of flags:

- `--from-file <path>` (strict JSON/YAML)
- `--edit` (edit a YAML template in `$EBO_EDITOR` / `$EDITOR` / `vi`)
- `--prompt` (interactive guided entry)

### Trip lifecycle + safety gates

Some operations are destructive and require `--force`:

```bash
./ebo trip cancel t1 --force
./ebo trip organizer remove t1 --member m1 --force
```

Some operations are naturally idempotent and do not accept idempotency flags:

```bash
./ebo trip publish t1
./ebo trip publish t1 --print-announcement
```

## Spec pinning + client generation

This repo pins the targeted spec version in `spec.lock`.

- `make gen` uses the pinned spec via `./tools/specpin` and regenerates the OpenAPI client.
- `make ci` verifies `spec.lock` and enforces formatting, tests, and **>= 85% internal (non-generated) coverage**.

## Working with Cursor agents (solo-dev workflow)

This repo includes a Cursor rule that encodes the expected agent workflow (issue picking, TDD, coverage gate, PR hygiene).

### Where the rule lives

- `/.cursor/rules/working-cadence.md`

### How to use it (what to type in Cursor)

To start the next GitHub issue in order:

```text
Follow the working-cadence rule and start the next GitHub issue in order.
```

To be explicit about the file:

```text
Follow .cursor/rules/working-cadence.md and begin the next GitHub issue in order.
```

### What it will do

- Picks the **lowest-Order** issue in the **EBO Planner CLI v1** GitHub Project with **Status = Todo**
- Moves it to **In Progress**
- Implements it using **TDD** while maintaining **>= 85% coverage for non-generated code**
- Opens a PR that closes the issue (**do not enable auto-merge unless explicitly requested**)
- Moves the Project item to **Done**

## Changelog & releases

- **For PRs**: update `CHANGELOG.md` under **`## [Unreleased]`** with a short note for user-visible CLI changes (commands/flags/output/behavior).
- **Spec pinning**: keep `spec.lock` updated to the spec tag the CLI targets.
- **To cut a release**: run `make changelog-release VERSION=x.y.z`, commit `CHANGELOG.md` (and `spec.lock` if changed), tag `vX.Y.Z`, and push the tag.

More details: `docs/releasing.md`
