# CLI Architecture (Hexagonal)

This document defines the **mandatory** layering and dependency rules for the `ebo` CLI implementation.

It exists to keep the CLI:

- **Spec-first** (behavior comes from `docs/cli-spec.md` + pinned `spec.lock`)
- **Testable** (clear seams; easy to fake ports)
- **Maintainable** (avoid “CLI glue” importing everything)

## Top-level structure (normative)

All production code lives under `internal/` and is organized into:

- `internal/adapters/in/cli/`: inbound adapter(s) that turn Cobra commands/flags into app calls.
- `internal/app/`: application layer (use-case orchestration; no UI, no HTTP).
- `internal/ports/out/`: outbound ports (interfaces) the app depends on.
- `internal/adapters/out/`: outbound adapters (implementations of ports; e.g., OpenAPI client wrapper, config store, token store).
- `internal/platform/`: shared cross-cutting helpers (output envelope, exit codes, validation helpers, time, IO abstractions).

## Dependency rules (normative)

Allowed import direction:

```
adapters/in/cli  ->  app  ->  ports/out
                         \->  platform

adapters/out/*   ->  ports/out
adapters/out/*   ->  platform

platform         ->  (no imports of app/adapters)
```

Hard rules:

- `internal/app/**` MUST NOT import `internal/adapters/**`.
- `internal/app/**` MUST depend only on:
  - `internal/ports/out/**` (interfaces)
  - `internal/platform/**`
  - standard library
- `internal/adapters/in/cli/**` MUST NOT call generated OpenAPI code directly.
  - CLI commands talk to the application layer, not the network.
- `internal/adapters/out/**` MUST be the only code that depends on:
  - generated OpenAPI client code (planned under `internal/gen/**`)
  - config file IO
  - OS keychain/secure storage (if introduced later)

## What goes where (examples)

- **Inbound adapter (`internal/adapters/in/cli/`)**
  - Cobra command definitions
  - Flag parsing and validation that is strictly “CLI UX” (mutual exclusions, required flags, etc.)
  - Mapping the CLI output mode (`table|json`) into platform renderers

- **Application layer (`internal/app/`)**
  - “Do the thing” orchestration: e.g. `TripService.ListVisibleTrips(ctx, ...)`
  - Idempotency key policy decisions that are CLI-defined (auto-generate vs prohibited), but *not* HTTP header logic
  - Selecting outbound ports to call (planner API, config store, etc.)

- **Outbound ports (`internal/ports/out/`)**
  - `PlannerAPI` (Trips/Members operations) interface
  - `ConfigStore` / `AuthStore` interfaces
  - `Clock`, `Editor`, `Prompter`, etc. if we need them as test seams

- **Outbound adapters (`internal/adapters/out/`)**
  - OpenAPI client wrapper that implements `PlannerAPI`
  - YAML config store that implements `ConfigStore`
  - Anything that touches the network, filesystem, environment, or OS

- **Platform (`internal/platform/`)**
  - Output envelope implementation (human vs JSON)
  - Exit code mapping utilities
  - Shared validation helpers (email/date parsing helpers, multi-line flag rejection)
  - IO abstractions for testing (stdout/stderr writers, etc.)

## Testing guidance (normative)

- App layer tests SHOULD use fake port implementations (in-memory) rather than mocking Cobra or HTTP.
- Inbound adapter tests SHOULD focus on:
  - flag parsing / mutual exclusivity
  - correct invocation of app layer with resolved inputs
- Outbound adapter tests SHOULD use:
  - `httptest.Server` for HTTP behavior
  - temporary directories for filesystem behavior

## Generated code (planned)

Generated OpenAPI client code SHOULD live under `internal/gen/**` and MUST NOT be imported from `internal/app/**`.
Only the outbound OpenAPI adapter should import `internal/gen/**`.
