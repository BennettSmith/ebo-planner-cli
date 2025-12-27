# `internal/platform`

Shared cross-cutting helpers used across layers.

Examples (planned):

- output envelope (human vs JSON)
- exit code mapping
- validation helpers
- IO abstractions for tests (stdout/stderr writers)

`platform` MUST NOT import `internal/app` or `internal/adapters`.
