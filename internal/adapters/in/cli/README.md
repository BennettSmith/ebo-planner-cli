# `internal/adapters/in/cli`

Inbound adapter for the CLI UX (Cobra commands, flags, help).

- Owns command structure and argument parsing.
- Calls into `internal/app`.
- MUST NOT call HTTP/OpenAPI generated clients directly.
