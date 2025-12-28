# Changelog

All notable changes to this project will be documented in this file.

The format is based on Keep a Changelog, and this project adheres to Semantic Versioning.

This repository contains a command-line interface for interacting with the Overland Trip Planning service.
Behavioral and API changes are defined in the spec repository; this changelog focuses on user-facing CLI behavior.

Each CLI release must declare which spec version it targets (see spec.lock).

Notes:

- Spec changelog = contract and behavior
- CLI changelog = commands, flags, UX, output, scripting impact

## [Unreleased]

### Added
- Added interactive `ebo auth login` (OIDC device flow) to obtain and store a bearer token.
- `ebo auth` token commands: `auth status`, `auth logout`, `auth token set`, and `auth token print`.
- `ebo profile` commands: manage profiles (list/show/create/set/use/delete) and switch current profile.
- Added `ebo config` commands (path/get/set/unset/list) with secret redaction and `--include-secrets` for JSON output.
- Added `--from-file <path>` JSON/YAML request mode for `ebo trip create`, `ebo trip update`, and `ebo member update`.
- Added `--edit` editor-based request mode for `ebo trip update` and `ebo member update`.
- Added `--prompt` interactive guided-entry mode for `ebo trip create`, `ebo trip update`, and `ebo member update`.
- Added trip read commands: `ebo trip list`, `ebo trip drafts`, and `ebo trip get <tripId>`.
- Added `ebo trip update` patch flags for common fields (including meeting location and artifacts), plus validation/mutual-exclusion rules for clear/replace semantics.
- Added trip lifecycle commands: `ebo trip visibility`, `ebo trip publish` (with `--print-announcement`), and `ebo trip cancel` (with `--force`).
- Added HTTP runtime helpers for per-request timeouts, verbose request logging, and token redaction.
- Added an architecture dependency guard test to enforce hex-layer import rules in CI.
- Added a `plannerapi` outbound port and HTTP adapter wrapping the generated OpenAPI client (auth + idempotency headers + normalized errors).
- Added deterministic OpenAPI client generation (`make gen`) driven by `spec.lock`.
- Added YAML-backed config storage with profile support and precedence resolution (preserves unknown fields on rewrite).
- Added a stable JSON output envelope and standardized exit-code mapping for errors.
- Initialized the Go module and added a minimal `ebo` root command with global flags and environment variable equivalents.

### Changed
- CI now runs `go test` with `-count=1` to disable test result caching.

### Deprecated

### Removed

### Fixed
- Fixed a panic in `ebo auth login` polling when the IdP returns `authorization_pending` during device flow.

### Security
