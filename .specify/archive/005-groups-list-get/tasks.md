---

description: "Task list for `sonora groups list` and `sonora groups get`"
---

# Tasks: List and Get Output Groups

**Input**: Design documents from `/specs/005-groups-list-get/`

**Prerequisites**: [plan.md](plan.md), [spec.md](spec.md), [research.md](research.md),
[data-model.md](data-model.md), [contracts/cli-groups-list.md](contracts/cli-groups-list.md),
[contracts/cli-groups-get.md](contracts/cli-groups-get.md), [quickstart.md](quickstart.md)

**Tests**: Tests are **mandatory**, not optional, for this project — constitution Principle
VI ("Test-First Development (NON-NEGOTIABLE)") requires every command and API client method
to have a test written, reviewed, and watched to fail *before* the implementing code is
written. Every implementation task below has a corresponding test task that MUST land first.

**Organization**: Tasks are grouped by user story (from spec.md), in priority order
(P1 → P2 → P3), so each story is independently implementable and testable. Where two stories
share priority P1 (US1, US3), spec order is preserved between them.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies on incomplete tasks)
- **[Story]**: Which user story this task belongs to (US1–US5)
- File paths are exact and relative to the repo root

## Path Conventions

Single Go project (established by `001-list-outputs`, unchanged): `cmd/sonora/`,
`internal/{hub,render,cli/groups,config}/`, `tests/{unit,contract,integration}/`.

---

## Phase 1: Setup

**Purpose**: Confirm the existing `001`–`004` codebase this feature builds on is in a
known-good state before touching it.

- [X] T001 Run `go build ./...` and `go test ./...` from the repo root on branch
  `005-groups-list-get` and confirm everything passes, establishing the baseline this
  feature's changes are layered on (no files modified by this task).

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Define the `Group` entity once, since both P1 user stories (US1 `groups list`,
US3 `groups get`) decode hub responses into it — a shared prerequisite rather than something
either story owns individually (data-model.md). Unlike `004-routes-list-get`'s `Route`, this
one struct is rendered identically by both commands — no split list/get view.

**⚠️ CRITICAL**: No User Story work can begin until this phase is complete.

- [X] T002 Define the `Group` struct in internal/hub/groups.go, mapping field-for-field to
  `#/components/schemas/GroupResponse` in `api/openapi.json`: `GroupID` (`groupId`),
  `DisplayName` (`displayName`), `OutputIDs` (`outputIds`, `[]string`), `Muted` (`muted`,
  bool), `Enabled` (`enabled`, bool) — per data-model.md's field table. No functions yet,
  struct only.

**Checkpoint**: `go build ./...` passes with the new `Group` type in place. User story
implementation can begin.

---

## Phase 3: User Story 1 - View currently enabled output groups (Priority: P1) 🎯 MVP

**Goal**: `sonora groups list` fetches and displays only enabled groups by default, each
showing all five fields: `groupId`/`displayName`/`outputIds`/`muted`/`enabled`, in YAML.

**Independent Test**: Run the list command with no flags against a hub with one or more
enabled groups and one disabled group, and verify only the enabled groups are shown, each
with its identifier, display name, member output identifiers, and mute state; against a hub
with zero enabled groups, verify a clear "no groups" indication instead of an empty/ambiguous
response.

### Tests for User Story 1 (write first; MUST fail before implementation exists)

- [X] T003 [P] [US1] Contract test for `hub.ListGroups`'s success and error paths in
  tests/contract/groups_list_test.go: an `httptest.Server` serving `GET /api/v2/groups` per
  `listGroups` in `api/openapi.json` — assert a default call sends `includeDisabled=false`
  explicitly as a query parameter, a `200` array body decodes into `[]hub.Group` correctly
  (including a group with `outputIds: []` and one with populated member outputs), a non-2xx
  response yields a `*hub.StatusError`, and a malformed body (missing `groupId` or
  `displayName`) yields a `*hub.DecodeError` per FR-017. Also assert the mock server receives
  exactly one request for a single `ListGroups` call (FR-013 — no automatic retry on
  failure).
- [X] T004 [P] [US1] Integration test for the list command's default success path in
  tests/integration/groups_list_test.go: build/run the `sonora` binary (following the
  `binPath`/`TestMain` pattern in tests/integration/outputs_list_test.go) against a mock
  server returning a mix of enabled and disabled groups, and assert `sonora groups list` (no
  flags) prints only the enabled groups with all 5 fields in YAML, exit code `0`. Also assert
  a mock server returning zero enabled groups produces the clear "no groups found" output with
  exit code `0` (FR-016) — distinct from every non-zero failure exit code, so "no groups" is
  never mistaken for a failure (SC-006) — and wrap a success invocation in
  `time.Now()`/`time.Since()` to assert completion in under 1 second (SC-001). Additionally,
  mirroring `TestOutputsList_HubNon2xx` in tests/integration/outputs_list_test.go and
  `TestRoutesList_HubNon2xx` in tests/integration/routes_list_test.go, assert that a mock
  server returning `500` causes `sonora groups list` to exit `3` with a clear hub-error
  message and empty stdout (FR-014, FR-017) — this is the full-CLI counterpart to T003's
  contract-level `*hub.StatusError`/`*hub.DecodeError` assertions.
- [X] T005 [P] [US1] Unit test for list YAML rendering in tests/unit/render_groups_test.go:
  `render.RenderGroupsYAML([]hub.Group{...})` emits exactly the 5 fields per group in the
  documented order (`groupId`, `displayName`, `outputIds`, `muted`, `enabled`), a group with
  zero member outputs renders `outputIds: []` explicitly, and a zero-length slice renders the
  `# no groups found` / `groups: []` shape (FR-016).
- [X] T006 [P] [US1] Unit test for `groups.RunList`'s flag parsing/dispatch and failure paths
  in tests/unit/cli_groups_test.go: unreachable `--hub-url` → exit `4`; unknown flag → exit
  `2`; unexpected positional argument → exit `2`; `--verbose` appends raw error detail on
  failure while its absence omits it. For the unreachable-`--hub-url` case, wrap the call in
  `time.Now()`/`time.Since()` and assert it returns well under 5 seconds (SC-004).

### Implementation for User Story 1

- [X] T007 [P] [US1] Implement `ListGroups(ctx context.Context, client *http.Client, baseURL
  string, includeDisabled bool) ([]Group, error)` in internal/hub/groups.go: builds
  `{baseURL}/api/v2/groups`, always sends `includeDisabled` explicitly as `true`/`false`,
  issues one GET, decodes a `200` array body into `[]Group` (nil → `[]Group{}`), and rejects a
  decoded record as malformed (`*DecodeError`) if `GroupID` or `DisplayName` is empty (FR-017,
  data-model.md's validation rule); any other non-2xx status returns `*StatusError`. Depends
  on T002. Makes T003 pass.
- [X] T008 [P] [US1] Implement `RenderGroupsYAML(groups []hub.Group) string` in
  internal/render/groups.go: emits the `groups:` list shape with all 5 fields per group in the
  documented order, `outputIds` as a nested YAML list (or `[]` when empty), and the `# no
  groups found` / `groups: []` shape for a zero-length slice — mirroring
  `internal/render/outputs.go`'s `RenderYAML` structure (FR-004). Makes T005 pass.
- [X] T009 [US1] Implement `RunList(args []string, stdout, stderr io.Writer) int` in
  internal/cli/groups/list.go: `flag.NewFlagSet("groups list", flag.ContinueOnError)` with
  `--verbose` and `--hub-url` (mirroring `outputs.RunList`'s flag definitions; no
  `--include-disabled`/`--json` yet — added in US2/US5), no positional arguments (any → usage
  error, exit `2`), resolves the hub URL via `config.ResolveHubURL`, calls `hub.NewClient()` +
  `hub.ListGroups(ctx, client, baseURL, false)` (`includeDisabled` hardcoded `false`),
  classifies any error via `hub.ClassifyError`, and on success prints
  `render.RenderGroupsYAML(groups)` to stdout and returns `0`. Depends on T007, T008. Makes
  T004, T006 pass.
- [X] T010 [US1] Add a new `case "groups":` to the noun switch in cmd/sonora/main.go, routing
  verb `"list"` to `groups.RunList(rest, stdout, stderr)` (an unrecognized verb under
  `"groups"` returns the existing usage-error shape, exit `2`; `"get"` is added in US3).
  Depends on T009.

**Checkpoint**: `sonora groups list` works end-to-end for the default (enabled-only) case —
first half of the MVP. `go build ./... && go test ./...` passes.

---

## Phase 4: User Story 3 - Look up a specific group by identifier (Priority: P1) 🎯 MVP

**Goal**: `sonora groups get <group-id>` fetches and displays one group's full state (all 5
fields) in YAML by default, regardless of its enabled/disabled state.

**Independent Test**: Run the get command with the identifier of a group known to exist
(whether enabled or disabled) and verify the identifier, display name, member output
identifiers, mute state, and enabled state all display correctly.

### Tests for User Story 3 (write first; MUST fail before implementation exists)

- [X] T011 [P] [US3] Contract test for `hub.GetGroup`'s success and hub-error paths in
  tests/contract/groups_get_test.go: an `httptest.Server` serving `GET
  /api/v2/groups/{groupId}` per `getGroup` in `api/openapi.json` — assert the request path
  includes the given ID, a `200 GroupResponse` body decodes into `hub.Group` correctly for
  both a group with member outputs and one with `outputIds: []`, a non-2xx/non-404 response
  (e.g. `500`) yields a `*hub.StatusError`, and a malformed `200` body yields a
  `*hub.DecodeError`. Also assert the mock server receives exactly one request for a single
  `GetGroup` call (FR-013 — no automatic retry on failure).
- [X] T012 [P] [US3] Integration test for the get command's success path in
  tests/integration/groups_get_test.go: build/run the `sonora` binary against a mock server,
  and assert `sonora groups get <id>` prints all 5 fields correctly in YAML (FR-008) for both
  an enabled and a disabled group (FR-007), with exit code `0`. Wrap a success invocation in
  `time.Now()`/`time.Since()` and assert completion in under 1 second (SC-002).
- [X] T013 [P] [US3] Extend tests/unit/render_groups_test.go with a test for single-group
  YAML rendering: `render.RenderGroupYAML(hub.Group{...})` emits all 5 fields as a bare
  (non-list) record in the documented order, and a group with zero member outputs shows
  `outputIds: []` explicitly (never omitted).
- [X] T014 [P] [US3] Unit test for `groups.RunGet`'s argument handling and failure paths in
  tests/unit/cli_groups_get_test.go: missing `<group-id>` → exit `2`; more than one
  positional argument → exit `2`; unreachable `--hub-url` → exit `4` (wrapped in
  `time.Now()`/`time.Since()` to assert well under 5 seconds, SC-004); `--verbose` appends
  raw error detail on failure while its absence omits it.

### Implementation for User Story 3

- [X] T015 [P] [US3] Implement `GetGroup(ctx context.Context, client *http.Client, baseURL,
  groupID string) (*Group, error)` in internal/hub/groups.go: builds
  `{baseURL}/api/v2/groups/{groupId}` (path-escaped), issues one GET, decodes a `200` body
  into `Group` (same field validation as `ListGroups`, wrapped as `*DecodeError`), and returns
  `*StatusError{StatusCode}` for any non-2xx response including `404` for now (404 gets its
  own `*NotFoundError` handling in US4 — US3's own tests don't exercise the 404 case). Depends
  on T002. Makes T011 pass.
- [X] T016 [P] [US3] Implement `RenderGroupYAML(g hub.Group) string` in
  internal/render/groups.go: emits all 5 fields as a single top-level record (no `groups:`
  wrapper), every field always explicit, including `outputIds: []` when empty. Makes T013
  pass.
- [X] T017 [US3] Implement `RunGet(args []string, stdout, stderr io.Writer) int` in
  internal/cli/groups/get.go: `flag.NewFlagSet("groups get", flag.ContinueOnError)` with
  `--verbose` and `--hub-url` (no `--json` yet — added in US5), one required positional
  `<group-id>` via the same re-parse-loop pattern as `outputs.RunGet`/`routes.RunGet` (so the
  identifier can appear before or after flags; zero or >1 positional args → usage error, exit
  `2`), resolves the hub URL via `config.ResolveHubURL`, calls `hub.NewClient()` +
  `hub.GetGroup`, classifies any error via `hub.ClassifyError`, and on success prints
  `render.RenderGroupYAML(*group)` to stdout and returns `0`. Depends on T015, T016. Makes
  T012, T014 pass.
- [X] T018 [US3] Add a `case "get":` to the `"groups"` noun's verb switch in
  cmd/sonora/main.go, routing to `groups.RunGet(rest, stdout, stderr)`. Depends on T010, T017.

**Checkpoint**: `sonora groups list` and `sonora groups get <id>` both work end-to-end for
their happy paths — full MVP. `go build ./... && go test ./...` passes.

---

## Phase 5: User Story 2 - Include disabled groups (Priority: P2)

**Goal**: `sonora groups list --include-disabled` returns both enabled and disabled groups,
each showing its enabled/disabled state.

**Independent Test**: Run the list command with the include-disabled option against a hub
with at least one disabled group and verify it appears in the results, distinguishable from
enabled groups via the `enabled` field.

### Tests for User Story 2 (write first; MUST fail before implementation exists)

- [X] T019 [P] [US2] Extend tests/contract/groups_list_test.go with a test asserting: a call
  with `includeDisabled` requested sends `includeDisabled=true` as the query parameter, and a
  mock server returning both enabled and disabled groups decodes all of them correctly (this
  exercises `hub.ListGroups`'s existing `includeDisabled` parameter from T007 — no hub-layer
  code change expected).
- [X] T020 [P] [US2] Extend tests/integration/groups_list_test.go with a test asserting
  `sonora groups list --include-disabled` against a mock server with one enabled and one
  disabled group returns both, each showing its correct `enabled` field value (User Story 2
  acceptance scenario).

### Implementation for User Story 2

- [X] T021 [US2] Add an `--include-disabled` bool flag (default `false`) to `RunList` in
  internal/cli/groups/list.go and pass its value through to `hub.ListGroups` in place of the
  hardcoded `false` from T009 (mirroring `outputs.RunList`'s existing `--include-disabled`
  handling). Makes T019, T020 pass.

**Checkpoint**: Disabled groups are visible on request. `go build ./... && go test ./...`
passes; `groups list`'s default (enabled-only) behavior from US1 is unaffected.

---

## Phase 6: User Story 4 - Handle a nonexistent group identifier (Priority: P2)

**Goal**: `sonora groups get <missing-id>` reports a clear "group not found" message,
distinguishable from every other failure class, and exits with a distinct status.

**Independent Test**: Run the get command with an identifier that does not exist on the mock
hub and verify the user sees an unambiguous "not found" message and a distinct exit code (not
the generic hub-error or network-error code).

### Tests for User Story 4 (write first; MUST fail before implementation exists)

- [X] T022 [P] [US4] Extend tests/contract/groups_get_test.go with a test asserting that a
  `404 ErrorResponse` from the mock server (per `getGroup` in `api/openapi.json`) causes
  `hub.GetGroup` to return a `*hub.NotFoundError{Resource: "group", ID: ...}` (not a generic
  `*hub.StatusError`).
- [X] T023 [P] [US4] Extend tests/integration/groups_get_test.go with a test asserting
  `sonora groups get <missing-id>` against a mock server that 404s prints a clear "group not
  found" message (naming the identifier) on stderr and exits with the not-found exit code
  (`5`), distinct from the exit codes used for hub errors (`3`) and network errors (`4`).

### Implementation for User Story 4

- [X] T024 [US4] Update `GetGroup` in internal/hub/groups.go to check for a `404` status
  before the generic non-2xx fallback added in T015, returning `&NotFoundError{Resource:
  "group", ID: groupID}` in that case (reusing the `NotFoundError{Resource, ID string}` type
  already generalized by `003-inputs-list-get` — no change to `internal/hub/errors.go`
  needed). Depends on T015. Makes T022, T023 pass.

**Checkpoint**: "not found" is a distinct, correctly-classified outcome for `groups get`. `go
build ./... && go test ./...` passes; `outputs get`/`inputs get`/`routes get`'s existing "not
found" behavior is unaffected (no shared code was modified).

---

## Phase 7: User Story 5 - Consume group data from a script (Priority: P3)

**Goal**: `sonora groups list --json` and `sonora groups get <id> --json` emit the same
fields as strict, parseable JSON.

**Independent Test**: Run each command with `--json` and verify the result parses with a
standard JSON parser and contains the full 5-field attribute set documented for that command.

### Tests for User Story 5 (write first; MUST fail before implementation exists)

- [X] T025 [US5] Extend tests/unit/render_groups_test.go with tests asserting
  `render.RenderGroupsJSON([]hub.Group{...})` produces `{"groups": [...]}` with all 5 fields
  per group and round-trips through `encoding/json.Unmarshal` without error, and
  `render.RenderGroupJSON(hub.Group{...})` produces a single JSON object (no list wrapper)
  with all 5 fields that likewise round-trips.
- [X] T026 [P] [US5] Integration test in tests/integration/groups_list_test.go asserting
  `sonora groups list --json` exits `0` and its stdout parses via `encoding/json` into the
  documented `{"groups": [...]}` shape with the same fields and values as the YAML view.
- [X] T027 [P] [US5] Integration test in tests/integration/groups_get_test.go asserting
  `sonora groups get <id> --json` exits `0` and its stdout parses via `encoding/json` into an
  object with the same fields and values as the YAML view.

### Implementation for User Story 5

- [X] T028 [P] [US5] Implement `RenderGroupsJSON(groups []hub.Group) string` in
  internal/render/groups.go: strict JSON `{"groups": [...]}` containing all 5 fields per
  group (nil → `[]hub.Group{}`), mirroring `internal/render/outputs.go`'s `RenderJSON`. Makes
  T025's list-JSON assertion pass.
- [X] T029 [P] [US5] Implement `RenderGroupJSON(g hub.Group) string` in
  internal/render/groups.go: strict JSON of the single record with all 5 fields (no list
  wrapper), mirroring `RenderOutputJSON`/`RenderRouteJSON`. Makes T025's single-JSON assertion
  pass.
- [X] T030 [US5] Add a `--json` bool flag to `RunList` in internal/cli/groups/list.go and
  switch rendering to `render.RenderGroupsJSON` when set (mirroring `outputs.RunList`'s
  existing `--json` handling). Depends on T028. Makes T026 pass.
- [X] T031 [US5] Add a `--json` bool flag to `RunGet` in internal/cli/groups/get.go and switch
  rendering to `render.RenderGroupJSON` when set (mirroring `outputs.RunGet`'s existing
  `--json` handling). Depends on T029. Makes T027 pass.

**Checkpoint**: All five user stories are independently functional. `go build ./... && go
test ./...` passes.

---

## Phase 8: Polish & Cross-Cutting Concerns

**Purpose**: Final quality gates spanning all five stories.

- [X] T032 [P] Run `gofmt -l .`, `go vet ./...`, and the project's configured linter across
  all files touched by this feature (internal/hub/, internal/render/, internal/cli/groups/,
  cmd/sonora/main.go, tests/); fix any findings (constitution Development Workflow).
- [X] T033 Execute every step in quickstart.md end-to-end (`go test ./...`, the manual
  success/include-disabled/not-found/failure-path smoke tests against a mock hub) and confirm
  each mapped Success Criterion (SC-001…SC-008) holds.

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — start immediately.
- **Foundational (Phase 2)**: Depends on Setup — BLOCKS all user stories (both P1 stories
  decode into the `Group` type T002 defines).
- **User Story 1 (Phase 3)**: Depends on Foundational. No dependency on US2–US5.
- **User Story 3 (Phase 4)**: Depends on Foundational. T018 (main.go `"get"` case) depends
  on T010 (US1's `"groups"` case) existing in the same switch statement; otherwise
  independent of US1/US2/US4/US5.
- **User Story 2 (Phase 5)**: Depends on Foundational and on US1's `RunList`/`ListGroups`
  (T007, T009) existing to extend. Independent of US3/US4/US5.
- **User Story 4 (Phase 6)**: Depends on Foundational and on US3's `GetGroup`/`RunGet`
  (T015, T017) existing to extend. Independent of US1/US2/US5. Touches no shared file
  (`errors.go` is unchanged — the `NotFoundError{Resource, ID}` type already exists from
  `003-inputs-list-get`), so it carries no regression risk to `outputs`/`inputs`/`routes`.
- **User Story 5 (Phase 7)**: Depends on Foundational and on US1's `RunList` (T009) and
  US3's `RunGet` (T017) existing to add `--json` to. Independent of US2/US4.
- **Polish (Phase 8)**: Depends on all five user stories being complete.

### Within Each User Story

- Tests are written first and MUST fail before their corresponding implementation task
  starts (Principle VI).
- Within US1: T007 and T008 (different files, no shared dependency) before T009 (uses
  both); T009 before T010 (main.go needs `RunList` to exist).
- Within US3: T015 and T016 (different files) before T017 (uses both); T017 before T018.
- Within US2: T021 only (no new hub/render code — reuses T007/T008 as-is).
- Within US4: T024 only (no new type needed — reuses the existing `NotFoundError`).
- Within US5: T028 before T030; T029 before T031 (each `RunX` needs its render function to
  call).

### Parallel Opportunities

- US1 tests: T003, T004, T005, T006 — four different files, no shared dependency — can all
  run in parallel.
- US1 implementation: T007 and T008 — different files — can run in parallel; both must land
  before T009.
- US3 tests: T011, T012, T013, T014 — four different files (T013 extends a file US1 already
  created, but as an independent test function) — can all run in parallel.
- US3 implementation: T015 and T016 — different files — can run in parallel; both must land
  before T017.
- US2 tests: T019 and T020 — different files — can run in parallel.
- US4 tests: T022 and T023 — different files — can run in parallel.
- US5: T026 and T027 (different files) can run in parallel; T028 and T029 (different
  functions, same file) can be authored in parallel and merged.
- T032 (lint/format) can run in parallel with T033 (quickstart validation).

---

## Parallel Example: User Story 1

```bash
# After Phase 2 (Foundational) completes, launch all US1 tests together:
Task: "Contract test for hub.ListGroups in tests/contract/groups_list_test.go"
Task: "Integration test for list command default success path in tests/integration/groups_list_test.go"
Task: "Unit test for RenderGroupsYAML in tests/unit/render_groups_test.go"
Task: "Unit test for RunList flag parsing in tests/unit/cli_groups_test.go"

# Once those tests fail as expected, launch the two independent implementation tasks:
Task: "Implement hub.ListGroups in internal/hub/groups.go"
Task: "Implement render.RenderGroupsYAML in internal/render/groups.go"
```

---

## Implementation Strategy

### MVP First (User Stories 1 and 3 only)

1. Complete Phase 1: Setup.
2. Complete Phase 2: Foundational (the `Group` struct — blocks everything else).
3. Complete Phase 3: User Story 1 (`groups list` default, enabled-only).
4. Complete Phase 4: User Story 3 (`groups get <id>`, full 5-field view).
5. **STOP and VALIDATE**: run T003–T014's tests plus manual `sonora groups list`/`sonora
   groups get <id>` smoke tests against a mock hub (quickstart.md §2–§3).
6. This is a shippable MVP: operators can discover group identifiers and look up any single
   group's full state.

### Incremental Delivery

1. Setup + Foundational → foundation ready.
2. Add User Story 1 → validate independently.
3. Add User Story 3 → validate independently → MVP complete.
4. Add User Story 2 → validate `--include-disabled` independently → doesn't touch US1/US3
   default behavior.
5. Add User Story 4 → validate the not-found path independently (exit code `5`, distinct
   message) → doesn't touch US1/US2/US3 behavior or any shared error type.
6. Add User Story 5 → validate `--json` for both commands independently → commands now match
   the full contracts in [contracts/cli-groups-list.md](contracts/cli-groups-list.md) and
   [contracts/cli-groups-get.md](contracts/cli-groups-get.md).
7. Polish: lint/format pass, full quickstart.md run.

---

## Notes

- [P] tasks touch different files with no dependency on an incomplete task.
- [Story] labels trace every task back to spec.md's user stories.
- Tests are mandatory (Principle VI) — write and confirm each test fails before starting its
  paired implementation task.
- This feature touches no existing `outputs`/`inputs`/`routes` code at all —
  `NotFoundError{Resource, ID}` is reused unchanged from `003-inputs-list-get`, and every new
  type/function/flag is additive in new files (`internal/hub/groups.go`,
  `internal/render/groups.go`, `internal/cli/groups/`) plus one new `case "groups":` branch in
  `cmd/sonora/main.go`.
