# ebo-planner-cli

TBD: Create the CLI

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
- Opens a PR that closes the issue and enables **auto-merge (squash)**
- Moves the Project item to **Done**

## Changelog & releases

- **For PRs**: update `CHANGELOG.md` under **`## [Unreleased]`** with a short note for user-visible CLI changes (commands/flags/output/behavior).
- **Spec pinning**: keep `spec.lock` updated to the spec tag the CLI targets.
- **To cut a release**: run `make changelog-release VERSION=x.y.z`, commit `CHANGELOG.md` (and `spec.lock` if changed), tag `vX.Y.Z`, and push the tag.

More details: `docs/releasing.md`
