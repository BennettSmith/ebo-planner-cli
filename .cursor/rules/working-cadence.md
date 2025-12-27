# EBO Planner CLI — Working Cadence (Agents)

This rule is **mandatory** for autonomous Cursor agents working in this repo.

## How to pick work

- Work is tracked in the GitHub Project **EBO Planner CLI v1**.
- Issues have a numeric **Order** field that defines the intended dependency order.
- Unless explicitly instructed otherwise, pick the next issue as the **lowest-Order** issue with **Status = Todo**.
- Work on **at most one** issue at a time.

## Before you start

- Move the Project item to **In Progress**.
- Assign the issue to yourself (or leave a comment stating an agent is working it).
- Create a branch named `{type}/{slug}` where `{type}` is one of: `chore`, `bug`, `refactor`, `feature`.

## Quality gates (TDD + coverage)

- **TDD is mandatory**:
  - write/update tests first (or in the same PR) to specify the behavior
  - do not land behavior changes without tests
- **Coverage gate**: maintain **>= 85% test coverage for non-generated code**.
  - Generated code (e.g., OpenAPI client code under `internal/gen/` or similar) must be excluded from coverage calculations.
  - Do not merge if non-generated coverage would fall below 85%.

## While working

- You MAY make incremental commits on your branch to checkpoint progress.
- Keep commits small and clearly named.

## Definition of Done

An issue is Done only when:

- `make ci` passes
- the issue acceptance criteria are satisfied
- a PR is opened that **closes** the issue (e.g., “Closes #123”)
- auto-merge is enabled using **squash** (unless explicitly told otherwise)
- the Project item is moved to **Done**

## When in doubt

- Follow `CONSTITUTION.md` (especially §6.2).


