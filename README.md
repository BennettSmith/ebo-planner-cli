# ebo-planner-cli

## Changelog & releases

- **For PRs**: update `CHANGELOG.md` under **`## [Unreleased]`** with a short note for user-visible CLI changes (commands/flags/output/behavior).
- **Spec pinning**: keep `spec.lock` updated to the spec tag the CLI targets.
- **To cut a release**: run `make changelog-release VERSION=x.y.z`, commit `CHANGELOG.md` (and `spec.lock` if changed), tag `vX.Y.Z`, and push the tag.

More details: `docs/releasing.md`
