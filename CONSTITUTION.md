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

## 7. Versioning & distribution

- CLI versioning is independent of spec and service.
- Each CLI release must declare the spec version it targets.
- Binary releases must be reproducible from source.

## 8. Testing philosophy

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
