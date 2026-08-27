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

- [X] T001 Create the `internal/cli/respath` package skeleton (package declaration only, no
      logic yet) in `internal/cli/respath/respath.go`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: The shared resource-path parser both US1 (get/list dispatch) and US3 (play's
target) depend on. Aliases are deliberately NOT included yet — canonical resource names only
— so US1 can be built and validated independently of US2 (research.md §1).

**⚠️ CRITICAL**: No user story work can begin until this phase is complete.

- [X] T002 [P] Write failing unit tests in `tests/unit/respath_test.go` for
      `internal/cli/respath`'s `ResourceKind`/`Path` parsing: canonical names only
      (`inputs`/`outputs`/`groups`/`routes`), splitting a resource-path argument on the first
      `/`, the id pattern `^[a-zA-Z0-9_-]{1,255}$` (FR-004a), and rejection of an unrecognized
      resource name or a malformed id (extra `/`, disallowed characters)
- [X] T003 Implement `ResourceKind`, `Path`, and a `Parse(arg string) (Path, error)` function
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

- [X] T004 [P] [US1] Rewrite `cmd/sonora/main_test.go`: replace `TestRunUnknownNoun`,
      `TestRunMissingVerb`, `TestRunUnknownVerb` with tests covering `get <resource>`,
      `get <resource>/<id>`, `list <resource>`, `list <resource>/<id>` (usage error, FR-003),
      an unrecognized resource name (usage error, FR-006), a bare `sonora get` and bare
      `sonora list` with no resource argument — usage error, exit `2`, with the message
      asserted to enumerate the valid resource names (FR-006a) — and old-style
      `sonora outputs list` / `sonora outputs get <id>` now failing as an "unknown command"
      usage error, exit `2` (FR-005); update `TestRunHelp`'s expected help-text substrings
- [X] T005 [P] [US1] Update `tests/integration/outputs_list_test.go` and
      `tests/integration/outputs_get_test.go` to invoke `sonora get outputs` /
      `sonora get outputs/<id>` instead of `sonora outputs list` / `sonora outputs get <id>`
- [X] T006 [P] [US1] Update `tests/integration/inputs_list_test.go` and
      `tests/integration/inputs_get_test.go` to the new `sonora get inputs[/<id>]` invocation
- [X] T007 [P] [US1] Update `tests/integration/groups_list_test.go` and
      `tests/integration/groups_get_test.go` to the new `sonora get groups[/<id>]` invocation
- [X] T008 [P] [US1] Update `tests/integration/routes_list_test.go` and
      `tests/integration/routes_get_test.go` to the new `sonora get routes[/<id>]`
      invocation, preserving `--status`/`--input-id`/`--target-id` filter coverage
- [X] T008a [P] [US1] Add failing assertions to `tests/unit/cli_inputs_test.go`,
      `cli_inputs_get_test.go`, `cli_outputs_test.go`, `cli_outputs_get_test.go`,
      `cli_groups_test.go`, `cli_groups_get_test.go`, `cli_routes_test.go`, and
      `cli_routes_get_test.go` that each package's `--help` usage line names the NEW
      grammar (`usage: sonora get <resource>[/<id>] …` / `usage: sonora list <resource> …`)
      and does NOT contain the removed `sonora <resource> list`/`get` form. The existing
      `--help` tests only assert exit `2` and the presence of a `Flags:` section, so the
      eight stale usage constants would otherwise ship unnoticed (FR-005, FR-006, SC-001)

### Implementation for User Story 1

- [X] T009 [US1] Rewrite `run()` in `cmd/sonora/main.go`: recognize `get` and `list` as
      top-level verbs, resolve the resource-path argument via
      `internal/cli/respath.Parse`, and translate the result into the existing
      `inputs`/`outputs`/`groups`/`routes` `RunList`/`RunGet` calls (same `rest []string`
      shape those functions already accept, per research.md §2); delete the old
      noun-first switch entirely so an unrecognized first token (including the old
      `inputs`/`outputs`/`groups`/`routes`) falls through to the existing generic
      "unknown command" error. A `get`/`list` with no resource argument MUST NOT reuse that
      generic error: emit a distinct usage message enumerating the valid resource names
      (FR-006a) — note today's `len(args) < 2` guard prints the old
      `usage: sonora <noun> <verb> [flags]` line and must be replaced, not just reworded
      (depends on T003; must be written after T004–T008a fail)
- [X] T010 [US1] Update the `helpText` constant in `cmd/sonora/main.go` to document
      `get`/`list <resource>[/<id>]` instead of `<noun> <verb>`
- [X] T010a [US1] Rewrite the eight stale usage constants to the new grammar so no
      user-facing string names a removed command form: `listUsage` and `getUsage` in
      `internal/cli/inputs/list.go`/`get.go`, `internal/cli/outputs/list.go`/`get.go`,
      `internal/cli/groups/list.go`/`get.go`, and `internal/cli/routes/list.go`/`get.go`
      (e.g. `usage: sonora outputs list [--include-disabled] …` →
      `usage: sonora get|list outputs [--include-disabled] …`, and
      `usage: sonora outputs get <output-id> …` → `usage: sonora get outputs/<output-id> …`),
      preserving each command's existing flag list verbatim (depends on T008a)
- [X] T011 [US1] Run `go test ./...` and fix any remaining regressions from the T004–T008a
      test rewrites (depends on T009, T010, T010a)

**Checkpoint**: `get`/`list` work identically to today's output for all four resources; the
old noun-verb grammar is gone. This is a shippable MVP on its own.

---

## Phase 4: User Story 2 - Resource aliases work everywhere a resource path appears (Priority: P2)

**Goal**: `in`/`out`/`gr`/`rt` are accepted anywhere `inputs`/`outputs`/`groups`/`routes` is,
for both the collection and single-item forms.

**Independent Test**: Substitute each alias for its full resource name in a `get`/`list`
command and confirm identical output to using the full name.

### Tests for User Story 2 (write first; confirm they fail before implementing)

- [X] T012 [P] [US2] Add failing unit tests to `tests/unit/respath_test.go` asserting
      `in`/`out`/`gr`/`rt` resolve to the same `ResourceKind` as
      `inputs`/`outputs`/`groups`/`routes`, both bare and with a `/<id>` suffix
- [X] T013 [P] [US2] Extend `tests/integration/outputs_get_test.go` with a case asserting
      `sonora get out/<id>` returns output identical to `sonora get outputs/<id>`

### Implementation for User Story 2

- [X] T014 [US2] Extend `internal/cli/respath`'s resource table in
      `internal/cli/respath/respath.go` with the `in`/`out`/`gr`/`rt` aliases (depends on
      T012, T013)
- [X] T015 [US2] Run `go test ./...` to confirm T012/T013 pass with no regression in US1
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

- [X] T016 [P] [US3] Rewrite `tests/unit/cli_play_test.go` for the new invocation: a
      `<target-path>` positional replacing `<target-id>` + `--group`/`--output`, and
      `--display-name` replacing `--name`; assert the old `--group`, `--output`, and
      `--name` flags now fail as unrecognized flags (exit `2`)
- [X] T017 [P] [US3] Rework `tests/integration/play_test.go` for the new invocation. Every
      test in this file passes a bare `<target-id>`, and four of them exercise behavior that
      no longer exists. Specifically:
      - **Delete** `TestPlay_AmbiguousTarget_NoFlag_ExitsAmbiguous` — it asserts exit `7`,
        which is removed (data-model.md); path-style addressing makes the case unreachable.
      - **Delete** `TestPlay_GroupAndOutputTogether_UsageError` — both flags are gone, so the
        "supplied together" case cannot occur; the unregistered-flag error is covered by T016.
      - **Rewrite** `TestPlay_GroupFlag_ForcesGroupEvenWhenOutputAlsoExists` and
        `TestPlay_OutputFlag_ForcesOutputEvenWhenGroupAlsoExists` to use
        `groups/shared-id` / `outputs/shared-id` — these keep their value as proof the path
        prefix (not collision detection) picks the target when an id names both.
      - **Rewrite** `TestPlay_GroupFlag_NotFoundWhenOnlyOutputExists` and
        `TestPlay_OutputFlag_NotFoundWhenOnlyGroupExists` to the path form, still asserting
        exit `5`.
      - **Update** the remaining tests (`TestPlay_Success_*`, `TestPlay_GroupTarget_Success`,
        `TestPlay_TargetNotFound`, `TestPlay_HubErrorStatuses_*`,
        `TestPlay_MalformedSuccessBody_*`, `TestPlay_Volume_*`) to pass
        `outputs/<id>`/`groups/<id>`, and `TestPlay_Name_SetsDisplayNameField` /
        `TestPlay_NoName_OmitsDisplayNameField` to use `--display-name`
- [X] T018 [P] [US3] Update `tests/contract/play_resolve_test.go`: remove the "target
      matches both an output and a group" ambiguous-case test (now structurally
      unreachable per data-model.md's exit-code-`7` removal), keeping/adjusting the
      per-type "not found" cases
- [X] T018a [P] [US3] Update `tests/unit/hub_client_test.go` for the removal of
      `hub.ClassAmbiguous` (T020) — without this the `unit` package will not compile:
      delete `TestClassifyError_AmbiguousTarget` (it calls
      `hub.ClassifyError(&hub.AmbiguousTargetError{…})` and asserts `hub.ClassAmbiguous`),
      and drop the `hub.ClassAmbiguous: 7` entry from `TestErrorClass_NewExitCodes`'s map
      while leaving the `6`/`8`/`9`/`10` entries asserting their current values

### Implementation for User Story 3

- [X] T019 [US3] Rewrite `internal/cli/play/play.go`: drop the `--group`, `--output`, and
      `--name` flag definitions; parse the second positional argument via
      `internal/cli/respath.Parse` restricted to `outputs`/`groups` (a resolved `inputs` or
      `routes` kind here is a usage error); add `--display-name`; pass the now-known target
      type straight into the playback request instead of calling the old auto-detect path
      (depends on T016, T017, T018, T018a)
- [X] T020 [US3] Replace `internal/hub/play.go`'s `ResolveTarget` with a narrower per-type
      existence check (the `forceGroup`/`forceOutput` parameters and the both-unset
      auto-detect/ambiguous branch all become dead once the CLI always supplies a known
      target type), and remove the now-unreachable ambiguity machinery from
      `internal/hub/errors.go`: the `ClassAmbiguous` enum member, its `ExitCode()` case, the
      `AmbiguousTargetError` type with its `Error()` method, and the `errors.As` branch in
      `ClassifyError` that maps it. **Note**: `ErrorClass` is `iota`-based, so deleting the
      `ClassAmbiguous` member shifts the underlying integer values of `ClassRouteFailed`,
      `ClassSourceUnreachable`, and `ClassServiceUnavailable` — this is safe only because
      `ExitCode()` switches on the constant *names*; exit codes `8`/`9`/`10` must stay
      unchanged (data-model.md), so verify them via `TestErrorClass_NewExitCodes` rather
      than assuming (depends on T019)
- [X] T021 [US3] Update the `usage` string in `internal/cli/play/play.go` and the `play`
      line in `cmd/sonora/main.go`'s `helpText` for the new target-path syntax
- [X] T022 [US3] Run `go test ./...` to confirm T016–T018a pass with no regression in
      US1/US2 (depends on T019, T020, T021)
- [X] T022a [US2+US3] Close FR-004's "everywhere a resource path is accepted" claim for
      `play`: add a case to `tests/integration/play_test.go` asserting
      `sonora play <uri> out/<id>` and `sonora play <uri> gr/<id>` behave identically to
      their `outputs/`/`groups/` spellings — this is the invocation in
      contracts/cli-play.md example #3, which nothing else covers. No implementation
      accompanies this task: `play` calls the same `respath.Parse` the aliases were added
      to in T014, so the test passes on arrival and exists to keep that shared-table
      guarantee from silently regressing (depends on T014 and T019)

**Checkpoint**: all three user stories complete; the CLI matches
`docs/cli-command-landscape.md` for every already-implemented command.

---

## Phase 6: Polish & Cross-Cutting Concerns

- [X] T023 [P] Update every old-grammar occurrence in README.md — this is a rewrite of the
      "Commands" section, not just the example block:
      - line 7: the banner example `sonora outputs list`
      - line 32: `Usage: `sonora <noun> <verb> [flags]`` → the verb-first usage line
      - lines 36–42: the `| Noun | Verbs |` table, whose whole axis is now wrong — replace
        with a verb-oriented table (`get`/`list` × the four resources, plus their
        `in`/`out`/`gr`/`rt` aliases and the `[/<id>]` single-item form)
      - lines 44–49: the `play` paragraph, which documents auto-detection, `--group`/
        `--output`, and `--name` — all removed; restate it as
        `sonora play <uri> <outputs|groups>/<id>` with `--display-name`
      - line 54: `Run `sonora <noun> <verb> --help`` and the `routes list` filter prose
      - lines 58–61: the four example commands
- [X] T024 [P] Update the `sonora outputs list` example in AGENTS.md line 9 to the new
      grammar (the only old-grammar occurrence in that file)
- [X] T025 [P] Fix the stale `` `sonora <noun> <verb>` `` doc comment at the top of
      `internal/cli/clihelp/usage.go`
- [X] T025a [P] Add explicit FR-009 regression coverage: assert `--json`, `--hub-url`, and
      `--verbose` still behave identically on each refactored command. These flags are
      currently only exercised incidentally (integration tests pass `--hub-url` via
      `runCLI`; `--json` and `--verbose` are each asserted for a couple of commands), so
      nothing today would catch one of them being dropped from a rewritten flag set. One
      table-driven test over the `get`/`list` forms plus `play` is sufficient
- [X] T026 Run through `quickstart.md` end-to-end against a local mock hub and fix any
      discrepancy found
- [X] T027 Run `gofmt`, `go vet ./...`, and the project's configured linter per the
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
- **US3 (Phase 5)**: Depends on Foundational only for `respath`'s canonical parsing. FR-007
  and FR-008 hold with canonical names alone, so US3's core is deliverable without US2 — but
  contracts/cli-play.md example #3 (`sonora play <uri> out/office-speaker --display-name …`)
  is an alias invocation, so that documented example is only satisfied once US2 has landed.
  T022a covers it and is gated on T014 accordingly; if US3 ships ahead of US2, T022a ships
  with US2, not with US3.
- **Polish (Phase 6)**: Depends on whichever of US1/US2/US3 are in scope for this release.

### Within Each Phase

- Tests MUST be written and confirmed failing before their corresponding implementation task
  (Constitution Principle VI).
- Implementation tasks within a phase follow the dependency notes inline above.

### Parallel Opportunities

- T002 has no other Foundational task to run alongside (the `[P]` marker is vestigial).
- T004–T008a (US1 tests) touch disjoint files — all `[P]`. T008a is the only one touching
  `tests/unit/cli_*.go`; T004 owns `cmd/sonora/main_test.go` and T005–T008 own
  `tests/integration/*`.
- T012–T013 (US2 tests) touch two disjoint files — both `[P]`.
- T016–T018a (US3 tests) touch four disjoint files — all `[P]`: `tests/unit/cli_play_test.go`,
  `tests/integration/play_test.go`, `tests/contract/play_resolve_test.go`, and
  `tests/unit/hub_client_test.go`.
- T023–T025a (Polish) touch four disjoint files — all `[P]`: README.md, AGENTS.md,
  `internal/cli/clihelp/usage.go`, and the new FR-009 regression test.
- T022a is NOT `[P]` with T016–T018a: it extends `tests/integration/play_test.go`, the file
  T017 rewrites, and it is gated on T014 (US2) besides.

---

## Parallel Example: User Story 1

```bash
Task: "Rewrite cmd/sonora/main_test.go dispatch tests for verb-first grammar"
Task: "Update tests/integration/outputs_list_test.go and outputs_get_test.go"
Task: "Update tests/integration/inputs_list_test.go and inputs_get_test.go"
Task: "Update tests/integration/groups_list_test.go and groups_get_test.go"
Task: "Update tests/integration/routes_list_test.go and routes_get_test.go"
Task: "Assert new-grammar usage lines in tests/unit/cli_{inputs,outputs,groups,routes}*_test.go"
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
- Suffixed ids (`T008a`, `T010a`, `T018a`, `T022a`, `T025a`) were inserted after review
  rather than renumbering T001–T027, so existing cross-references stay valid — matching the
  spec's own `FR-004a` convention. They are ordinary tasks, not optional ones.
- Commit after each task or logical group; stop at any checkpoint to validate that story
  independently before continuing.
