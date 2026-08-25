---

description: "Task list for `sonora routes list` and `sonora routes get`"
---

# Tasks: List and Get Audio Routes

**Input**: Design documents from `/specs/004-routes-list-get/`

**Prerequisites**: [plan.md](plan.md), [spec.md](spec.md), [research.md](research.md),
[data-model.md](data-model.md), [contracts/cli-routes-list.md](contracts/cli-routes-list.md),
[contracts/cli-routes-get.md](contracts/cli-routes-get.md), [quickstart.md](quickstart.md)

**Tests**: Tests are **mandatory**, not optional, for this project — constitution Principle
VI ("Test-First Development (NON-NEGOTIABLE)") requires every command and API client method
to have a test written, reviewed, and watched to fail *before* the implementing code is
written. Every implementation task below has a corresponding test task that MUST land first.

**Organization**: Tasks are grouped by user story (from spec.md), in priority order
(P1 → P2 → P3), so each story is independently implementable and testable.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies on incomplete tasks)
- **[Story]**: Which user story this task belongs to (US1–US5)
- File paths are exact and relative to the repo root

## Path Conventions

Single Go project (established by `001-list-outputs`, unchanged): `cmd/sonora/`,
`internal/{hub,render,cli/routes,config}/`, `tests/{unit,contract,integration}/`.

---

## Phase 1: Setup

**Purpose**: Confirm the existing `001`–`003` codebase this feature builds on is in a
known-good state before touching it.

- [X] T001 Run `go build ./...` and `go test ./...` from the repo root on branch
  `004-routes-list-get` and confirm everything passes, establishing the baseline this
  feature's changes are layered on (no files modified by this task).

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Define the `Route` entity once, since both P1 user stories (US1 `routes list`,
US2 `routes get`) decode hub responses into it — a shared prerequisite rather than something
either story owns individually (data-model.md).

**⚠️ CRITICAL**: No User Story work can begin until this phase is complete.

- [X] T002 Define the `Route` struct in internal/hub/routes.go, mapping field-for-field to
  `#/components/schemas/RouteResponse` in `api/openapi.json`: `RouteID` (`routeId`),
  `InputID` (`inputId`), `TargetID` (`targetId`), `TargetType` (`targetType`, string),
  `Status` (`status`, string), `CreatedAt` (`createdAt`, plain `string` — required,
  non-nullable), `StartedAt` (`startedAt`, `*string`, nullable), `Transferable`
  (`transferable`, bool), `Pauseable` (`pauseable`, bool), `Paused` (`paused`, bool) — per
  data-model.md's field table. No functions yet, struct only.

**Checkpoint**: `go build ./...` passes with the new `Route` type in place. User story
implementation can begin.

---

## Phase 3: User Story 1 - View currently active routes (Priority: P1) 🎯 MVP

**Goal**: `sonora routes list` fetches and displays all routes regardless of status by
default, each showing the 5-field list view: `routeId`/`inputId`/`targetId`/`targetType`/
`status`, in YAML.

**Independent Test**: Run the list command with no flags against a hub with one or more
routes and verify each route's identifier, source input identifier, target identifier,
target type, and status are shown; against a hub with zero routes, verify a clear "no routes"
indication instead of an empty/ambiguous response.

### Tests for User Story 1 (write first; MUST fail before implementation exists)

- [X] T003 [P] [US1] Contract test for `hub.ListRoutes`'s success and error paths in
  tests/contract/routes_list_test.go: an `httptest.Server` serving `GET /api/v2/routes` per
  `listRoutes` in `api/openapi.json` — assert a call with no filters sends no `status`/
  `inputId`/`targetId` query parameters at all, a `200` array body decodes into `[]hub.Route`
  correctly (including a route with `startedAt: null` and one with `startedAt` populated), a
  non-2xx response yields a `*hub.StatusError`, and a malformed body (missing `routeId`, or
  an unrecognized `targetType`/`status` value) yields a `*hub.DecodeError` per FR-016. Also
  assert the mock server receives exactly one request for a single `ListRoutes` call
  (FR-012 — no automatic retry on failure).
- [X] T004 [P] [US1] Integration test for the list command's default success path in
  tests/integration/routes_list_test.go: build/run the `sonora` binary (following the
  `binPath`/`TestMain` pattern in tests/integration/outputs_list_test.go) against a mock
  server returning routes of mixed status, and assert `sonora routes list` (no flags) prints
  every route with all 5 list-view fields in YAML, exit code `0`. Also assert a mock server
  returning zero routes produces the clear "no routes found" output with exit code `0`
  (FR-015) — distinct from every non-zero failure exit code, so "no routes" is never mistaken
  for a failure (SC-006) — and wrap a success invocation in `time.Now()`/`time.Since()` to
  assert completion in under 1 second (SC-001).
- [X] T005 [P] [US1] Unit test for list YAML rendering in tests/unit/render_routes_test.go:
  `render.RenderRoutesYAML([]hub.Route{...})` emits exactly the 5 list-view fields per route
  in the documented order (`routeId`, `inputId`, `targetId`, `targetType`, `status` — no
  `createdAt`/`startedAt`/`transferable`/`pauseable`/`paused`), and a zero-length slice
  renders the `# no routes found` / `routes: []` shape (FR-015).
- [X] T006 [P] [US1] Unit test for `routes.RunList`'s flag parsing/dispatch and failure paths
  in tests/unit/cli_routes_test.go: unreachable `--hub-url` → exit `4`; unknown flag → exit
  `2`; unexpected positional argument → exit `2`; `--verbose` appends raw error detail on
  failure while its absence omits it. For the unreachable-`--hub-url` case, wrap the call in
  `time.Now()`/`time.Since()` and assert it returns well under 5 seconds (SC-004).

### Implementation for User Story 1

- [X] T007 [P] [US1] Implement `ListRoutes(ctx context.Context, client *http.Client, baseURL,
  status, inputID, targetID string) ([]Route, error)` in internal/hub/routes.go: builds
  `{baseURL}/api/v2/routes`, adding `status`/`inputId`/`targetId` as query parameters only
  when each argument is non-empty, issues one GET, decodes a `200` array body into
  `[]Route` (nil → `[]Route{}`), and rejects a decoded record as malformed (`*DecodeError`)
  if `RouteID`/`InputID`/`TargetID` is empty, `TargetType` isn't `"SINGLE_OUTPUT"` or
  `"OUTPUT_GROUP"`, or `Status` isn't one of the five documented enum values (FR-016,
  data-model.md's validation rule); any other non-2xx status returns `*StatusError`. Depends
  on T002. Makes T003 pass.
- [X] T008 [P] [US1] Implement `RenderRoutesYAML(routes []hub.Route) string` in
  internal/render/routes.go: emits the `routes:` list shape with only the 5 list-view fields
  per route in the documented order, and the `# no routes found` / `routes: []` shape for a
  zero-length slice — mirroring `internal/render/outputs.go`'s `RenderYAML` structure but
  with a narrower field set (FR-004). Makes T005 pass.
- [X] T009 [US1] Implement `RunList(args []string, stdout, stderr io.Writer) int` in
  internal/cli/routes/list.go: `flag.NewFlagSet("routes list", flag.ContinueOnError)` with
  `--verbose` and `--hub-url` (mirroring `outputs.RunList`'s flag definitions; no
  `--status`/`--input-id`/`--target-id`/`--json` yet — added in US3/US5), no positional
  arguments (any → usage error, exit `2`), resolves the hub URL via `config.ResolveHubURL`,
  calls `hub.NewClient()` + `hub.ListRoutes(ctx, client, baseURL, "", "", "")` (all filters
  empty), classifies any error via `hub.ClassifyError`, and on success prints
  `render.RenderRoutesYAML(routes)` to stdout and returns `0`. Depends on T007, T008. Makes
  T004, T006 pass.
- [X] T010 [US1] Add a new `case "routes":` to the noun switch in cmd/sonora/main.go, routing
  verb `"list"` to `routes.RunList(rest, stdout, stderr)` (an unrecognized verb under
  `"routes"` returns the existing usage-error shape, exit `2`; `"get"` is added in US2).
  Depends on T009.

**Checkpoint**: `sonora routes list` works end-to-end for the default (all-routes) case —
first half of the MVP. `go build ./... && go test ./...` passes.

---

## Phase 4: User Story 2 - Look up a specific route by identifier (Priority: P1) 🎯 MVP

**Goal**: `sonora routes get <route-id>` fetches and displays one route's full state (all 10
fields) in YAML by default.

**Independent Test**: Run the get command with the identifier of a route known to exist and
verify the identifier, source input identifier, target identifier, target type, status,
creation timestamp, playback-started timestamp (or explicit absence), transferable,
pauseable, and paused all display correctly.

### Tests for User Story 2 (write first; MUST fail before implementation exists)

- [X] T011 [P] [US2] Contract test for `hub.GetRoute`'s success and hub-error paths in
  tests/contract/routes_get_test.go: an `httptest.Server` serving `GET
  /api/v2/routes/{routeId}` per `getRoute` in `api/openapi.json` — assert the request path
  includes the given ID, a `200 RouteResponse` body decodes into `hub.Route` correctly for
  both a route where playback has started (`startedAt` populated) and one where it hasn't
  (`startedAt: null`), a non-2xx/non-404 response (e.g. `500`) yields a `*hub.StatusError`,
  and a malformed `200` body yields a `*hub.DecodeError`. Also assert the mock server
  receives exactly one request for a single `GetRoute` call (FR-012 — no automatic retry on
  failure).
- [X] T012 [P] [US2] Integration test for the get command's success path in
  tests/integration/routes_get_test.go: build/run the `sonora` binary against a mock server,
  and assert `sonora routes get <id>` prints all 10 fields correctly in YAML (FR-007), with
  exit code `0`. Wrap a success invocation in `time.Now()`/`time.Since()` and assert
  completion in under 1 second (SC-002).
- [X] T013 [P] [US2] Extend tests/unit/render_routes_test.go with a test for single-route
  YAML rendering: `render.RenderRouteYAML(hub.Route{...})` emits all 10 fields as a bare
  (non-list) record in the documented order, and a route whose playback hasn't started shows
  `startedAt: null` explicitly (never omitted).
- [X] T014 [P] [US2] Unit test for `routes.RunGet`'s argument handling and failure paths in
  tests/unit/cli_routes_get_test.go: missing `<route-id>` → exit `2`; more than one
  positional argument → exit `2`; unreachable `--hub-url` → exit `4` (wrapped in
  `time.Now()`/`time.Since()` to assert well under 5 seconds, SC-004); `--verbose` appends
  raw error detail on failure while its absence omits it.

### Implementation for User Story 2

- [X] T015 [P] [US2] Implement `GetRoute(ctx context.Context, client *http.Client, baseURL,
  routeID string) (*Route, error)` in internal/hub/routes.go: builds
  `{baseURL}/api/v2/routes/{routeId}` (path-escaped), issues one GET, decodes a `200` body
  into `Route` (same field validation as `ListRoutes`, wrapped as `*DecodeError`), and
  returns `*StatusError{StatusCode}` for any non-2xx response including `404` for now (404
  gets its own `*NotFoundError` handling in US4 — US2's own tests don't exercise the 404
  case). Depends on T002. Makes T011 pass.
- [X] T016 [P] [US2] Implement `RenderRouteYAML(r hub.Route) string` in
  internal/render/routes.go: emits all 10 fields (the 5 list-view fields plus `createdAt`,
  `startedAt`, `transferable`, `pauseable`, `paused`) as a single top-level record (no
  `routes:` wrapper), every field always explicit, including bare `startedAt: null`. Makes
  T013 pass.
- [X] T017 [US2] Implement `RunGet(args []string, stdout, stderr io.Writer) int` in
  internal/cli/routes/get.go: `flag.NewFlagSet("routes get", flag.ContinueOnError)` with
  `--verbose` and `--hub-url` (no `--json` yet — added in US5), one required positional
  `<route-id>` via the same re-parse-loop pattern as `outputs.RunGet`/`inputs.RunGet` (so the
  identifier can appear before or after flags; zero or >1 positional args → usage error, exit
  `2`), resolves the hub URL via `config.ResolveHubURL`, calls `hub.NewClient()` +
  `hub.GetRoute`, classifies any error via `hub.ClassifyError`, and on success prints
  `render.RenderRouteYAML(*route)` to stdout and returns `0`. Depends on T015, T016. Makes
  T012, T014 pass.
- [X] T018 [US2] Add a `case "get":` to the `"routes"` noun's verb switch in
  cmd/sonora/main.go, routing to `routes.RunGet(rest, stdout, stderr)`. Depends on T010,
  T017.

**Checkpoint**: `sonora routes list` and `sonora routes get <id>` both work end-to-end for
their happy paths — full MVP. `go build ./... && go test ./...` passes.

---

## Phase 5: User Story 3 - Narrow the route list by status, input, or target (Priority: P2)

**Goal**: `sonora routes list --status/--input-id/--target-id` narrows results to only the
routes matching every supplied filter (AND logic).

**Independent Test**: Run the list command with each filter option in turn, and with more
than one at once, against a hub with routes in different states/inputs/targets, and verify
only the matching routes are returned in each case.

### Tests for User Story 3 (write first; MUST fail before implementation exists)

- [X] T019 [P] [US3] Extend tests/contract/routes_list_test.go with tests asserting: a call
  with `status`/`inputID`/`targetID` set sends exactly those query parameters (and omits any
  left empty) to the hub; a `400 ErrorResponse` from the mock server (invalid `status` value,
  per `listRoutes` in `api/openapi.json`) causes `hub.ListRoutes` to return a
  `*hub.StatusError` (this exercises `hub.ListRoutes`'s existing filter parameters from T007
  — no hub-layer code change expected).
- [X] T020 [P] [US3] Extend tests/integration/routes_list_test.go with tests asserting:
  `sonora routes list --status FAILED` against a mock server with routes of mixed status
  returns only `FAILED` routes; `--input-id`/`--target-id` individually narrow results the
  same way; `--status ACTIVE --target-id kitchen-speaker` together return only routes
  matching both (AND logic, FR-003); and `--status NOT_A_REAL_STATUS` against a mock server
  that 400s on an unrecognized status exits `3` with a clear hub-error message (spec.md edge
  case).

### Implementation for User Story 3

- [X] T021 [US3] Add `--status`, `--input-id`, and `--target-id` string flags (each default
  `""`) to `RunList` in internal/cli/routes/list.go and pass their values through to
  `hub.ListRoutes` in place of the three hardcoded empty strings from T009 (mirroring
  `outputs.RunList`'s existing `--include-disabled` handling, extended to three independent
  filters). Makes T019, T020 pass.

**Checkpoint**: All three filters work individually and combined. `go build ./... && go test
./...` passes; `routes list`'s default (no-filter, all-routes) behavior from US1 is
unaffected.

---

## Phase 6: User Story 4 - Handle a nonexistent route identifier (Priority: P2)

**Goal**: `sonora routes get <missing-id>` reports a clear "route not found" message,
distinguishable from every other failure class, and exits with a distinct status.

**Independent Test**: Run the get command with an identifier that does not exist on the mock
hub and verify the user sees an unambiguous "not found" message and a distinct exit code (not
the generic hub-error or network-error code).

### Tests for User Story 4 (write first; MUST fail before implementation exists)

- [X] T022 [P] [US4] Extend tests/contract/routes_get_test.go with a test asserting that a
  `404 ErrorResponse` from the mock server (per `getRoute` in `api/openapi.json`) causes
  `hub.GetRoute` to return a `*hub.NotFoundError{Resource: "route", ID: ...}` (not a generic
  `*hub.StatusError`).
- [X] T023 [P] [US4] Extend tests/integration/routes_get_test.go with a test asserting
  `sonora routes get <missing-id>` against a mock server that 404s prints a clear "route not
  found" message (naming the identifier) on stderr and exits with the not-found exit code
  (`5`), distinct from the exit codes used for hub errors (`3`) and network errors (`4`).

### Implementation for User Story 4

- [X] T024 [US4] Update `GetRoute` in internal/hub/routes.go to check for a `404` status
  before the generic non-2xx fallback added in T015, returning `&NotFoundError{Resource:
  "route", ID: routeID}` in that case (reusing the `NotFoundError{Resource, ID string}` type
  already generalized by `003-inputs-list-get` — no change to `internal/hub/errors.go`
  needed). Depends on T015. Makes T022, T023 pass.

**Checkpoint**: "not found" is a distinct, correctly-classified outcome for `routes get`. `go
build ./... && go test ./...` passes; `outputs get`/`inputs get`'s existing "not found"
behavior is unaffected (no shared code was modified).

---

## Phase 7: User Story 5 - Consume route data from a script (Priority: P3)

**Goal**: `sonora routes list --json` and `sonora routes get <id> --json` emit the same
fields as strict, parseable JSON.

**Independent Test**: Run each command with `--json` and verify the result parses with a
standard JSON parser and contains the full field set documented for that command (5 fields
for list, 10 for get).

### Tests for User Story 5 (write first; MUST fail before implementation exists)

- [X] T025 [US5] Extend tests/unit/render_routes_test.go with tests asserting
  `render.RenderRoutesJSON([]hub.Route{...})` produces `{"routes": [...]}` with only the 5
  list-view fields per route and round-trips through `encoding/json.Unmarshal` without error,
  and `render.RenderRouteJSON(hub.Route{...})` produces a single JSON object (no list
  wrapper) with all 10 fields that likewise round-trips.
- [X] T026 [P] [US5] Integration test in tests/integration/routes_list_test.go asserting
  `sonora routes list --json` exits `0` and its stdout parses via `encoding/json` into the
  documented `{"routes": [...]}` shape with the same 5 fields and values as the YAML view.
- [X] T027 [P] [US5] Integration test in tests/integration/routes_get_test.go asserting
  `sonora routes get <id> --json` exits `0` and its stdout parses via `encoding/json` into an
  object with the same 10 fields and values as the YAML view.

### Implementation for User Story 5

- [X] T028 [P] [US5] Implement `RenderRoutesJSON(routes []hub.Route) string` in
  internal/render/routes.go: strict JSON `{"routes": [...]}` containing only the 5 list-view
  fields per route (nil → `[]hub.Route{}`), mirroring `internal/render/outputs.go`'s
  `RenderJSON`. Makes T025's list-JSON assertion pass.
- [X] T029 [P] [US5] Implement `RenderRouteJSON(r hub.Route) string` in
  internal/render/routes.go: strict JSON of the single record with all 10 fields (no list
  wrapper), mirroring `RenderOutputJSON`/`RenderInputJSON`. Makes T025's single-JSON
  assertion pass.
- [X] T030 [US5] Add a `--json` bool flag to `RunList` in internal/cli/routes/list.go and
  switch rendering to `render.RenderRoutesJSON` when set (mirroring `outputs.RunList`'s
  existing `--json` handling). Depends on T028. Makes T026 pass.
- [X] T031 [US5] Add a `--json` bool flag to `RunGet` in internal/cli/routes/get.go and
  switch rendering to `render.RenderRouteJSON` when set (mirroring `outputs.RunGet`'s
  existing `--json` handling). Depends on T029. Makes T027 pass.

**Checkpoint**: All five user stories are independently functional. `go build ./... && go
test ./...` passes.

---

## Phase 8: Polish & Cross-Cutting Concerns

**Purpose**: Final quality gates spanning all five stories.

- [X] T032 [P] Run `gofmt -l .`, `go vet ./...`, and the project's configured linter across
  all files touched by this feature (internal/hub/, internal/render/, internal/cli/routes/,
  cmd/sonora/main.go, tests/); fix any findings (constitution Development Workflow).
- [X] T033 Execute every step in quickstart.md end-to-end (`go test ./...`, the manual
  success/filter/not-found/failure-path smoke tests against a mock hub) and confirm each
  mapped Success Criterion (SC-001…SC-008) holds.

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — start immediately.
- **Foundational (Phase 2)**: Depends on Setup — BLOCKS all user stories (both P1 stories
  decode into the `Route` type T002 defines).
- **User Story 1 (Phase 3)**: Depends on Foundational. No dependency on US2–US5.
- **User Story 2 (Phase 4)**: Depends on Foundational. T018 (main.go `"get"` case) depends
  on T010 (US1's `"routes"` case) existing in the same switch statement; otherwise
  independent of US1/US3/US4/US5.
- **User Story 3 (Phase 5)**: Depends on Foundational and on US1's `RunList`/`ListRoutes`
  (T007, T009) existing to extend. Independent of US2/US4/US5.
- **User Story 4 (Phase 6)**: Depends on Foundational and on US2's `GetRoute`/`RunGet`
  (T015, T017) existing to extend. Independent of US1/US3/US5. Touches no shared file
  (`errors.go` is unchanged — the `NotFoundError{Resource, ID}` type already exists from
  `003-inputs-list-get`), so it carries no regression risk to `outputs`/`inputs`.
- **User Story 5 (Phase 7)**: Depends on Foundational and on US1's `RunList` (T009) and
  US2's `RunGet` (T017) existing to add `--json` to. Independent of US3/US4.
- **Polish (Phase 8)**: Depends on all five user stories being complete.

### Within Each User Story

- Tests are written first and MUST fail before their corresponding implementation task
  starts (Principle VI).
- Within US1: T007 and T008 (different files, no shared dependency) before T009 (uses
  both); T009 before T010 (main.go needs `RunList` to exist).
- Within US2: T015 and T016 (different files) before T017 (uses both); T017 before T018.
- Within US3: T021 only (no new hub/render code — reuses T007/T008 as-is).
- Within US4: T024 only (no new type needed — reuses the existing `NotFoundError`).
- Within US5: T028 before T030; T029 before T031 (each `RunX` needs its render function to
  call).

### Parallel Opportunities

- US1 tests: T003, T004, T005, T006 — four different files, no shared dependency — can all
  run in parallel.
- US1 implementation: T007 and T008 — different files — can run in parallel; both must land
  before T009.
- US2 tests: T011, T012, T013, T014 — four different files (T013 extends a file US1 already
  created, but as an independent test function) — can all run in parallel.
- US2 implementation: T015 and T016 — different files — can run in parallel; both must land
  before T017.
- US3 tests: T019 and T020 — different files — can run in parallel.
- US4 tests: T022 and T023 — different files — can run in parallel.
- US5: T026 and T027 (different files) can run in parallel; T028 and T029 (different
  functions, same file) can be authored in parallel and merged.
- T032 (lint/format) can run in parallel with T033 (quickstart validation).

---

## Parallel Example: User Story 1

```bash
# After Phase 2 (Foundational) completes, launch all US1 tests together:
Task: "Contract test for hub.ListRoutes in tests/contract/routes_list_test.go"
Task: "Integration test for list command default success path in tests/integration/routes_list_test.go"
Task: "Unit test for RenderRoutesYAML in tests/unit/render_routes_test.go"
Task: "Unit test for RunList flag parsing in tests/unit/cli_routes_test.go"

# Once those tests fail as expected, launch the two independent implementation tasks:
Task: "Implement hub.ListRoutes in internal/hub/routes.go"
Task: "Implement render.RenderRoutesYAML in internal/render/routes.go"
```

---

## Implementation Strategy

### MVP First (User Stories 1 and 2 only)

1. Complete Phase 1: Setup.
2. Complete Phase 2: Foundational (the `Route` struct — blocks everything else).
3. Complete Phase 3: User Story 1 (`routes list` default, all statuses).
4. Complete Phase 4: User Story 2 (`routes get <id>`, full 10-field view).
5. **STOP and VALIDATE**: run T003–T014's tests plus manual `sonora routes list`/`sonora
   routes get <id>` smoke tests against a mock hub (quickstart.md §2–§3).
6. This is a shippable MVP: operators can discover route identifiers and look up any single
   route's full state.

### Incremental Delivery

1. Setup + Foundational → foundation ready.
2. Add User Story 1 → validate independently.
3. Add User Story 2 → validate independently → MVP complete.
4. Add User Story 3 → validate `--status`/`--input-id`/`--target-id` (individually and
   combined) independently → doesn't touch US1/US2 default behavior.
5. Add User Story 4 → validate the not-found path independently (exit code `5`, distinct
   message) → doesn't touch US1/US2/US3 behavior or any shared error type.
6. Add User Story 5 → validate `--json` for both commands independently → commands now
   match the full contracts in [contracts/cli-routes-list.md](contracts/cli-routes-list.md)
   and [contracts/cli-routes-get.md](contracts/cli-routes-get.md).
7. Polish: lint/format pass, full quickstart.md run.

---

## Notes

- [P] tasks touch different files with no dependency on an incomplete task.
- [Story] labels trace every task back to spec.md's user stories.
- Tests are mandatory (Principle VI) — write and confirm each test fails before starting its
  paired implementation task.
- This feature touches no existing `outputs`/`inputs` code at all — `NotFoundError{Resource,
  ID}` is reused unchanged from `003-inputs-list-get`, and every new type/function/flag is
  additive in new files (`internal/hub/routes.go`, `internal/render/routes.go`,
  `internal/cli/routes/`) plus one new `case "routes":` branch in `cmd/sonora/main.go`.
