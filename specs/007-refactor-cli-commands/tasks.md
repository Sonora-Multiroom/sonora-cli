---

description: "Task list for 007-refactor-cli-commands"
---

# Tasks: Adopt Verb-First Command Landscape

**Input**: Design documents from `/specs/007-refactor-cli-commands/`

**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/, quickstart.md

**Tests**: Included and REQUIRED — Constitution Principle VI (Test-First, NON-NEGOTIABLE)
mandates a failing test before implementation for every command and client-method change in
this repo; this is not the template's optional default.

**Organization**: Tasks are grouped by user story (spec.md) to enable independent
implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependency on an incomplete task)
- **[Story]**: Which user story this task belongs to (US1/US2/US3)

## Path Conventions

Single Go project. Source under `cmd/` and `internal/`; tests under `tests/unit`,
`tests/contract`, `tests/integration` (existing repo convention — unit tests for
internal packages live in `tests/unit`, not alongside the package).

---

## Phase 1: Setup

**Purpose**: Scaffold the one new package this feature adds.

- [ ] T001 Create the `internal/cli/respath` package skeleton (package declaration only, no
      logic yet) in `internal/cli/respath/respath.go`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: The shared resource-path parser both US1 (get/list dispatch) and US3 (play's
target) depend on. Aliases are deliberately NOT included yet — canonical resource names only
— so US1 can be built and validated independently of US2 (research.md §1).

**⚠️ CRITICAL**: No user story work can begin until this phase is complete.

- [ ] T002 [P] Write failing unit tests in `tests/unit/respath_test.go` for
      `internal/cli/respath`'s `ResourceKind`/`Path` parsing: canonical names only
      (`inputs`/`outputs`/`groups`/`routes`), splitting a resource-path argument on the first
      `/`, the id pattern `^[a-zA-Z0-9_-]{1,255}$` (FR-004a), and rejection of an unrecognized
      resource name or a malformed id (extra `/`, disallowed characters)
- [ ] T003 Implement `ResourceKind`, `Path`, and a `Parse(arg string) (Path, error)` function
      in `internal/cli/respath/respath.go` to make T002 pass (depends on T002)

**Checkpoint**: `internal/cli/respath` parses and validates resource paths by canonical name.
User story implementation can now begin.

---

## Phase 3: User Story 1 - Read commands follow the verb-first, path-style shape (Priority: P1) 🎯 MVP

**Goal**: `sonora get <resource>[/<id>]` and `sonora list <resource>` replace
`sonora <resource> list`/`get <id>` for inputs, outputs, groups, and routes; the old grammar
is removed, not deprecated.

**Independent Test**: Run `sonora get inputs`, `sonora get outputs`, `sonora get groups`,
`sonora get routes` (each with and without a trailing `/<id>`) and confirm identical data to
today's `sonora <resource> list`/`get <id>`; confirm `sonora <resource> list`/`get <id>` now
fail with a usage error.

### Tests for User Story 1 (write first; confirm they fail before implementing)

- [ ] T004 [P] [US1] Rewrite `cmd/sonora/main_test.go`: replace `TestRunUnknownNoun`,
      `TestRunMissingVerb`, `TestRunUnknownVerb` with tests covering `get <resource>`,
      `get <resource>/<id>`, `list <resource>`, `list <resource>/<id>` (usage error, FR-003),
      an unrecognized resource name (usage error, FR-006), and old-style
      `sonora outputs list` / `sonora outputs get <id>` now failing as an "unknown command"
      usage error, exit `2` (FR-005); update `TestRunHelp`'s expected help-text substrings
- [ ] T005 [P] [US1] Update `tests/integration/outputs_list_test.go` and
      `tests/integration/outputs_get_test.go` to invoke `sonora get outputs` /
      `sonora get outputs/<id>` instead of `sonora outputs list` / `sonora outputs get <id>`
- [ ] T006 [P] [US1] Update `tests/integration/inputs_list_test.go` and
      `tests/integration/inputs_get_test.go` to the new `sonora get inputs[/<id>]` invocation
- [ ] T007 [P] [US1] Update `tests/integration/groups_list_test.go` and
      `tests/integration/groups_get_test.go` to the new `sonora get groups[/<id>]` invocation
- [ ] T008 [P] [US1] Update `tests/integration/routes_list_test.go` and
      `tests/integration/routes_get_test.go` to the new `sonora get routes[/<id>]`
      invocation, preserving `--status`/`--input-id`/`--target-id` filter coverage

### Implementation for User Story 1

- [ ] T009 [US1] Rewrite `run()` in `cmd/sonora/main.go`: recognize `get` and `list` as
      top-level verbs, resolve the resource-path argument via
      `internal/cli/respath.Parse`, and translate the result into the existing
      `inputs`/`outputs`/`groups`/`routes` `RunList`/`RunGet` calls (same `rest []string`
      shape those functions already accept, per research.md §2); delete the old
      noun-first switch entirely so an unrecognized first token (including the old
      `inputs`/`outputs`/`groups`/`routes`) falls through to the existing generic
      "unknown command" error (depends on T003; must be written after T004–T008 fail)
- [ ] T010 [US1] Update the `helpText` constant in `cmd/sonora/main.go` to document
      `get`/`list <resource>[/<id>]` instead of `<noun> <verb>`
- [ ] T011 [US1] Run `go test ./...` and fix any remaining regressions from the T004–T008
      test rewrites (depends on T009, T010)

**Checkpoint**: `get`/`list` work identically to today's output for all four resources; the
old noun-verb grammar is gone. This is a shippable MVP on its own.

---

## Phase 4: User Story 2 - Resource aliases work everywhere a resource path appears (Priority: P2)

**Goal**: `in`/`out`/`gr`/`rt` are accepted anywhere `inputs`/`outputs`/`groups`/`routes` is,
for both the collection and single-item forms.

**Independent Test**: Substitute each alias for its full resource name in a `get`/`list`
command and confirm identical output to using the full name.

### Tests for User Story 2 (write first; confirm they fail before implementing)

- [ ] T012 [P] [US2] Add failing unit tests to `tests/unit/respath_test.go` asserting
      `in`/`out`/`gr`/`rt` resolve to the same `ResourceKind` as
      `inputs`/`outputs`/`groups`/`routes`, both bare and with a `/<id>` suffix
- [ ] T013 [P] [US2] Extend `tests/integration/outputs_get_test.go` with a case asserting
      `sonora get out/<id>` returns output identical to `sonora get outputs/<id>`

### Implementation for User Story 2

- [ ] T014 [US2] Extend `internal/cli/respath`'s resource table in
      `internal/cli/respath/respath.go` with the `in`/`out`/`gr`/`rt` aliases (depends on
      T012, T013)
- [ ] T015 [US2] Run `go test ./...` to confirm T012/T013 pass with no regression in US1
      (depends on T014)

**Checkpoint**: aliases work everywhere a resource path is accepted for `get`/`list`.

---

## Phase 5: User Story 3 - `play`'s target and flags match the new landscape (Priority: P3)

**Goal**: `sonora play <uri> <outputs|groups>/<id> [--display-name NAME]` replaces
`sonora play <uri> <target-id> [--group|--output] [--name NAME]`.

**Independent Test**: Run `sonora play <uri> outputs/<id>` and
`sonora play <uri> groups/<id>` (with `--display-name`) and confirm the same playback result
as today's `--output`/`--group`/`--name` forms; confirm the old flags now fail.

### Tests for User Story 3 (write first; confirm they fail before implementing)

- [ ] T016 [P] [US3] Rewrite `tests/unit/cli_play_test.go` for the new invocation: a
      `<target-path>` positional replacing `<target-id>` + `--group`/`--output`, and
      `--display-name` replacing `--name`; assert the old `--group`, `--output`, and
      `--name` flags now fail as unrecognized flags (exit `2`)
- [ ] T017 [P] [US3] Update `tests/integration/play_test.go` to invoke
      `sonora play <uri> outputs/<id>` / `groups/<id>` instead of
      `<target-id> --output`/`--group`, and `--display-name` instead of `--name`
- [ ] T018 [P] [US3] Update `tests/contract/play_resolve_test.go`: remove the "target
      matches both an output and a group" ambiguous-case test (now structurally
      unreachable per data-model.md's exit-code-`7` removal), keeping/adjusting the
      per-type "not found" cases

### Implementation for User Story 3

- [ ] T019 [US3] Rewrite `internal/cli/play/play.go`: drop the `--group`, `--output`, and
      `--name` flag definitions; parse the second positional argument via
      `internal/cli/respath.Parse` restricted to `outputs`/`groups` (a resolved `inputs` or
      `routes` kind here is a usage error); add `--display-name`; pass the now-known target
      type straight into the playback request instead of calling the old auto-detect path
      (depends on T016, T017, T018)
- [ ] T020 [US3] Simplify `internal/hub/play.go`'s `ResolveTarget` (or replace it with a
      narrower per-type existence check) to drop the both-unset auto-detect/ambiguous
      branch, and remove `hub.ClassAmbiguous` from `internal/hub/errors.go` — exit code `7`
      is unreachable now that the CLI always supplies a known target type (depends on T019)
- [ ] T021 [US3] Update the `usage` string in `internal/cli/play/play.go` and the `play`
      line in `cmd/sonora/main.go`'s `helpText` for the new target-path syntax
- [ ] T022 [US3] Run `go test ./...` to confirm T016–T018 pass with no regression in
      US1/US2 (depends on T019, T020, T021)

**Checkpoint**: all three user stories complete; the CLI matches
`docs/cli-command-landscape.md` for every already-implemented command.

---

## Phase 6: Polish & Cross-Cutting Concerns

- [ ] T023 [P] Update README.md's example commands (`sonora outputs list`,
      `sonora routes list --status active`, `sonora groups get <id> --json`,
      `sonora play "..." office-speaker --volume 40`) to the new grammar
- [ ] T024 [P] Update the `sonora outputs list` example in AGENTS.md to the new grammar
- [ ] T025 [P] Fix the stale `` `sonora <noun> <verb>` `` doc comment at the top of
      `internal/cli/clihelp/usage.go`
- [ ] T026 Run through `quickstart.md` end-to-end against a local mock hub and fix any
      discrepancy found
- [ ] T027 Run `gofmt`, `go vet ./...`, and the project's configured linter per the
      Constitution's Development Workflow, fixing any findings

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies.
- **Foundational (Phase 2)**: Depends on Setup — BLOCKS all user stories.
- **US1 (Phase 3)**: Depends on Foundational. No dependency on US2/US3.
- **US2 (Phase 4)**: Depends on Foundational and on US1 existing (it extends the resource
  table US1's dispatcher already calls) — not independently deployable before US1, but
  independently testable as its own increment on top of US1.
- **US3 (Phase 5)**: Depends on Foundational only for `respath`'s canonical parsing; benefits
  from US2's aliases if present but does not require them (FR-004a/FR-007 hold with canonical
  names alone).
- **Polish (Phase 6)**: Depends on whichever of US1/US2/US3 are in scope for this release.

### Within Each Phase

- Tests MUST be written and confirmed failing before their corresponding implementation task
  (Constitution Principle VI).
- Implementation tasks within a phase follow the dependency notes inline above.

### Parallel Opportunities

- T002 has no other Foundational task to run alongside.
- T004–T008 (US1 tests) touch five disjoint files — all `[P]`.
- T012–T013 (US2 tests) touch two disjoint files — both `[P]`.
- T016–T018 (US3 tests) touch three disjoint files — all `[P]`.
- T023–T025 (Polish) touch three disjoint files — all `[P]`.

---

## Parallel Example: User Story 1

```bash
Task: "Rewrite cmd/sonora/main_test.go dispatch tests for verb-first grammar"
Task: "Update tests/integration/outputs_list_test.go and outputs_get_test.go"
Task: "Update tests/integration/inputs_list_test.go and inputs_get_test.go"
Task: "Update tests/integration/groups_list_test.go and groups_get_test.go"
Task: "Update tests/integration/routes_list_test.go and routes_get_test.go"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1 (Setup) and Phase 2 (Foundational).
2. Complete Phase 3 (US1) — this alone replaces every already-shipped `list`/`get` command
   with the new grammar and removes the old one. **STOP and VALIDATE** against
   `quickstart.md` steps 1–4 before continuing.

### Incremental Delivery

1. Setup + Foundational → foundation ready.
2. US1 → validate independently → the CLI's read commands are fully migrated (MVP).
3. US2 → validate independently → aliases work everywhere.
4. US3 → validate independently → `play` matches the landscape; run the full
   `quickstart.md`.
5. Polish → docs and lint pass.

## Notes

- No `[Story]` label on Setup/Foundational/Polish tasks, per format rules.
- Every `[P]` pair above was checked for disjoint files — none share a file with another
  task marked `[P]` in the same phase.
- Commit after each task or logical group; stop at any checkpoint to validate that story
  independently before continuing.
