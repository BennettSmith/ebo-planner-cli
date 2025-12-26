# Releasing (CLI)

## Updating the changelog in PRs

- Add an entry to `CHANGELOG.md` under **`## [Unreleased]`**.
- Put it under the most appropriate section:
  - `### Added`, `### Changed`, `### Deprecated`, `### Removed`, `### Fixed`, `### Security`
- Changelog entries are required when PRs change:
  - Commands/subcommands, flags, arguments, defaults
  - Output format (human or machine-readable): **treat as potentially breaking**
  - Exit codes, error messages, auth/token handling
  - Configuration/env var behavior
  - Scripting ergonomics (JSON fields, ordering guarantees, headers, etc.)

## How this differs from spec/service changelogs

- **Spec repo changelog**: defines contract and behavior of the overall system (API + requirements).
- **Service repo changelog**: deployment/runtime/migrations and which spec version is implemented.
- **CLI changelog (this repo)**: end-user UX: commands, flags, output, scripting impact. Output format changes are breaking unless explicitly backwards compatible.

## Spec pinning policy (`spec.lock`)

This repo must pin the spec version the CLI targets in `spec.lock` (a spec git tag like `v1.2.3`).

- Update `spec.lock` when adopting a new spec version.
- Each CLI release must include a changelog line: `- Targets spec \`vX.Y.Z\`` (the release script will ensure this).

## Cutting a CLI release

1. Update `spec.lock` to the spec tag targeted (for example: `v1.2.3`).
2. Ensure `CHANGELOG.md` has entries under `## [Unreleased]`.
3. Cut the release section (moves Unreleased entries into a dated version section and ensures it includes the pinned spec version):

```bash
make changelog-release VERSION=x.y.z
```

4. Commit the changelog update (and `spec.lock` if it changed):

```bash
git add CHANGELOG.md spec.lock
git commit -m "chore(release): vX.Y.Z"
```

5. Create and push the git tag:

```bash
git tag vX.Y.Z
git push --tags
```

## SemVer (very short)

- **MAJOR**: breaking changes (including output format changes unless proven compatible)
- **MINOR**: backwards-compatible features
- **PATCH**: backwards-compatible fixes


