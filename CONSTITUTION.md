# Constitution — Overland Trip Planning CLI

## 1. Purpose

This repository provides a **command-line interface** for interacting with
the Overland Trip Planning service.

It exists to:

- Enable organizers to create, manage, and publish trips
- Provide power-user workflows unavailable in the web UI
- Act as a first-class API consumer and contract validator

## 2. Source of Truth

- API contract and behavior are defined in the **spec repo**.
- This repo consumes a pinned spec version via `spec.lock`.

The CLI must not rely on undocumented or unspecified service behavior.

**We practice spec-first development.** All requirements changes — new features, changes to behavior defined by a use case, and API contract changes — **must originate in the spec repo**. This repo may only implement/consume those changes after updating `spec.lock`.

## 3. Scope

### 3.1 Allowed content

- CLI command definitions
- Client-side validation and UX enhancements
- API client code generated from OpenAPI
- Auth/token handling for end users
- Terminal UI, formatting, and ergonomics

### 3.2 Disallowed content

- Service implementation logic
- Business rules not defined in the spec
- Direct database or internal API access
- Spec or domain model definitions

## 4. CLI principles

- **Thin client**: the service enforces truth; the CLI assists users.
- **Use-case oriented commands**:
  - CLI commands map clearly to spec use cases.
- **Safe by default**:
  - Idempotent operations
  - Explicit confirmations for destructive actions
- **Scriptable**:
  - Machine-readable output formats (`--json`)
  - Stable exit codes

## 5. Contract compliance

### 5.1 Spec pinning

- `spec.lock` defines the exact spec version the CLI targets.
- Generated client code must be derived from that version only.

### 5.2 Validation strategy

- The CLI may perform preflight validation for UX
- The service remains authoritative
- The CLI must handle and surface API errors faithfully

## 6. Change workflow

1) Spec repo change (if behavior/contract changes)
2) Tag spec
3) Update `spec.lock`
4) Regenerate client
5) Update CLI commands/UX

CLI-only UX improvements may skip step (1).

If a CLI feature would require a new endpoint, a changed request/response shape, new validation rules, or any new/changed use-case behavior, step (1) is mandatory.

## 6.1 Development workflow (mandatory)

### 6.1.1 Branches only

- All work MUST happen on a branch (no direct commits to `main`).
- Branch names MUST be: `{type}/{slug}`
  - `{type}` MUST be one of: `chore`, `bug`, `refactor`, `feature`
  - `{slug}` MUST be a short, lowercase, hyphenated description
  - Examples:
    - `feature/trip-list-command`
    - `bug/json-output-stability`
    - `refactor/client-layering`
    - `chore/ci-target-wiring`

### 6.1.2 Pre-flight before PR

- Before creating or updating a PR, you MUST run `make ci` locally and it MUST pass.

### 6.1.3 Pull requests required

- Every change MUST be delivered via a pull request.
- CI must be green before merge (required checks).

### 6.1.4 Automation via `gh`

- Cursor agents SHOULD use the GitHub CLI (`gh`) to create PRs and set titles/descriptions.
- Cursor agents MUST enable auto-merge using **squash** for routine changes (so PRs merge automatically once required checks/reviews are satisfied).

Example:

```bash
gh pr create --fill
gh pr merge --auto --squash
```

## 6.2 Cursor agent working cadence (mandatory)

This repo is designed for autonomous Cursor agents working issue-by-issue with tight feedback loops.

### 6.2.1 Source of work: GitHub Project ordering

- Work is tracked in the GitHub Project **EBO Planner CLI v1**.
- Issues have a numeric **Order** field that defines the intended dependency order.
- Unless explicitly instructed otherwise, an agent MUST pick the next issue as the **lowest-Order** issue in **Status: Todo**.

### 6.2.2 One issue at a time

- An agent MUST work on **at most one** issue at a time.
- Before starting implementation, the agent MUST:
  - set the Project item to **In Progress**
  - assign the issue to themselves (or clearly comment that an agent is working it)

### 6.2.3 TDD + coverage gate

- The CLI MUST be developed using **TDD**:
  - tests are written/updated first (or in the same PR) to specify the behavior being added
  - behavior changes MUST NOT land without tests
- The repo MUST maintain **>= 85% test coverage for non-generated code**.
  - Generated code (e.g., OpenAPI client code under `internal/gen/` or similar) MUST be excluded from coverage calculations.
  - PRs MUST NOT reduce non-generated coverage below 85%.

### 6.2.4 Definition of Done for an issue

An issue is Done only when:

- `make ci` passes
- acceptance criteria in the issue are satisfied
- the PR links the issue (e.g., “Closes #123”)
- the Project item is moved to **Done**

### 6.2.5 PR hygiene

- PRs should stay small and focused on one issue.
- Prefer squash-merge with auto-merge enabled (see 6.1.4).
- If an issue needs to be split, the agent MUST create follow-up issues and keep the Project Order consistent.

### 6.2.6 Incremental commits (recommended)

- While working on a branch, an agent MAY make incremental commits as they go to checkpoint progress.
- Checkpoint commits SHOULD be:
  - small and scoped (one logical change)
  - named clearly (imperative mood, e.g., “Add config loader tests”)
- It is OK for intermediate commits to be “work in progress”, but the branch MUST be green (`make ci`) before opening/marking ready a PR.
- The default merge strategy remains **squash**, so incremental commits do not leak into `main` unless explicitly desired.

## 7. Versioning & distribution

- CLI versioning is independent of spec and service.
- Each CLI release must declare the spec version it targets.
- Binary releases must be reproducible from source.

## 8. Testing philosophy

- **TDD is mandatory** (see 6.2.3).
- **Coverage is a release-quality gate**: non-generated code coverage MUST remain **>= 85%**.
- **Command tests** (argument parsing, flags)
- **Client integration tests** (mocked HTTP)
- **End-to-end tests** against a running service (optional)
- Acceptance scenarios from the spec repo are encouraged as test inputs

## 9. UX guarantees

- Non-interactive mode must always be available
- Interactive features must degrade gracefully
- Errors must be human-readable and actionable

## 10. Non-goals

- This repo is not a general SDK.
- This repo is not a replacement for the web UI.
- This repo does not define product behavior.
