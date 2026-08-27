---

description: "Task list for `sonora route`"
---

# Tasks: Route an Existing Input (`route`)

**Input**: Design documents from `/specs/008-route-command/`

**Prerequisites**: [plan.md](plan.md), [spec.md](spec.md), [research.md](research.md),
[data-model.md](data-model.md), [contracts/cli-route.md](contracts/cli-route.md),
[quickstart.md](quickstart.md)

**Tests**: Tests are **mandatory**, not optional, for this project — constitution Principle VI
("Test-First Development (NON-NEGOTIABLE)") requires every command and API client method to
have a test written, reviewed, and watched to fail *before* the implementing code is written.
Routing is additionally named by the constitution as a core flow requiring contract/integration
coverage before merge. Every implementation task below has a corresponding test task that MUST
land first.

**Organization**: Tasks are grouped by user story (from spec.md), in priority order (P1 → P2),
so each story is independently implementable and testable.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies on incomplete tasks)
- **[Story]**: Which user story this task belongs to (US1–US2)
- File paths are exact and relative to the repo root

## Path Conventions

Single Go project (established by `001-list-outputs`, unchanged): `cmd/sonora/`,
`internal/{hub,render,cli/route,config}/`, `tests/{unit,contract,integration}/`.

---

## Phase 1: Setup

**Purpose**: Confirm the existing `001`–`007` codebase this feature builds on is in a
known-good state before touching it.

- [X] T001 Run `go build ./...` and `go test ./...` from the repo root on branch
  `008-route-command` and confirm everything passes, establishing the baseline this feature's
  changes are layered on (no files modified by this task).

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Both user stories depend on the same new hub-layer function (`hub.CreateRoute`)
and the same two new exit-code classes (`ClassInputNotFound`/`ClassTargetNotFound`,
research.md §3) — neither US1 nor US2 can be independently tested without these existing
first. `hub.GetInput` and `hub.ResolveTarget` are reused unchanged (research.md §1–2) and need
no foundational work of their own.

**⚠️ CRITICAL**: No User Story work can begin until this phase is complete.

### Tests for Foundational work (write first; MUST fail before implementation exists)

- [X] T002 [P] Contract test for `hub.CreateRoute` in tests/contract/route_test.go: an
  `httptest.Server` serving `POST /api/v2/routes` per `createRoute` in `api/openapi.json` —
  assert the request body always includes `inputId`/`targetId`/`targetType`; a `201
  RouteResponse` body decodes into the existing `hub.Route` (reusing `validateRoute`'s
  routeId/inputId/targetId-non-empty and targetType/status enum checks); a `400`/`422`
  `ErrorResponse` body decodes into `*hub.APIError{StatusCode, Title, Detail}`; a `404` yields
  `*hub.NotFoundError{Resource: "target", ID: req.TargetID}`; a non-JSON or empty error body on
  `400`/`422` falls back to `*hub.StatusError` rather than failing to classify at all; a
  malformed `201` body (missing `routeId`/`inputId`/`targetId`, or an unrecognized
  `targetType`/`status`) yields `*hub.DecodeError` per FR-011. Also assert the mock server
  receives exactly one request per `CreateRoute` call (FR-009 — no automatic retry).
- [X] T003 [P] Unit test for the two new `hub.ErrorClass` values, extending
  tests/unit/hub_client_test.go: `ClassInputNotFound.ExitCode() == 11` and
  `ClassTargetNotFound.ExitCode() == 12`; assert every exit code in data-model.md's exit code
  table is mutually distinct (FR-010). Also assert `hub.ClassifyError` is unchanged for this
  feature: a `*hub.NotFoundError` with any `Resource` value (including `"input"`) still
  classifies as the existing `ClassNotFound` (exit 5) when routed through `ClassifyError`
  directly — a regression guard for research.md §3's decision to keep the input/target split
  local to `route.Run`, not inside `ClassifyError`.

### Implementation for Foundational work

- [X] T004 [P] Define `CreateRouteRequest` struct in internal/hub/routes.go, mapping
  field-for-field to `#/components/schemas/CreateRouteRequest` in `api/openapi.json` per
  data-model.md: `InputID`, `TargetID`, `TargetType` (`string`, all required, json tags
  `inputId`/`targetId`/`targetType`). No function yet, struct only.
- [X] T005 [P] Extend internal/hub/errors.go: add `ClassInputNotFound` (exit 11) and
  `ClassTargetNotFound` (exit 12) to the `ErrorClass` enum and its `ExitCode()` switch.
  `ClassifyError` itself is **not** modified — its existing `*NotFoundError` branch continues
  to return `ClassNotFound` regardless of `.Resource`, per research.md §3. Makes T003 pass.
- [X] T006 Implement `CreateRoute(ctx context.Context, client *http.Client, baseURL string, req
  CreateRouteRequest) (*Route, error)` in internal/hub/routes.go: `POST
  {baseURL}/api/v2/routes` with `req` JSON-encoded as the body; on `201`, decode into `Route`
  and validate via the existing `validateRoute` helper, returning `*DecodeError` on failure; on
  `404`, return `*NotFoundError{Resource: "target", ID: req.TargetID}`; on `400`/`422`, attempt
  to decode the body as `#/components/schemas/ErrorResponse` (reusing the existing unexported
  `errorResponse` type from internal/hub/play.go) into `*APIError{StatusCode, Title, Detail}`,
  falling back to `*StatusError{StatusCode}` if that decode fails; any other non-2xx status
  returns `*StatusError{StatusCode}`. Depends on T004. Makes T002 pass.

**Checkpoint**: `go build ./...` passes with `hub.CreateRoute` and the two new exit-code
classes in place; no CLI surface exists yet. User story implementation can begin.

---

## Phase 3: User Story 1 - Route an existing input to a single output (Priority: P1) 🎯 MVP

**Goal**: `sonora route inputs/<input-id> outputs/<target-id>` verifies the input and the
output each already exist, calls the hub to create a route between them, and prints the
route's identifier and status plus a confirmation message, in YAML by default or JSON with
`--json` — without creating any new input.

**Independent Test**: Run the route command with the path of an existing, enabled input and
the path of a single existing output, then verify the command reports the resulting route's
identifier and status exactly as returned by the hub, that the same result is valid, parseable
structured data under `--json`, and that the input is unchanged afterward (no new input
created).

### Tests for User Story 1 (write first; MUST fail before implementation exists)

- [X] T007 [P] [US1] Integration tests for the route command's success and failure paths in
  tests/integration/route_test.go: build/run the `sonora` binary (following the
  `binPath`/`TestMain` pattern in tests/integration/outputs_list_test.go) against a fake hub
  serving `/api/v2/inputs/{id}`, `/api/v2/outputs/{id}`, `/api/v2/groups/{id}`, and
  `/api/v2/routes`, where `<input-id>` exists as an enabled input and `<target-id>` exists only
  as an output. Assert the success case: `sonora route inputs/<input-id> outputs/<target-id>`
  prints `routeId`/`status`/`message` in YAML with exit code `0`, the same fields parse via
  `encoding/json` under `--json`, and a follow-up fake-hub request log shows the input was
  never mutated; wrap the invocation in `time.Now()`/`time.Since()` and assert completion in
  under 5 seconds (SC-002). Also assert failure paths: an unknown `<input-id>` → exit `11` with
  an "input not found" message and zero calls to `/api/v2/routes`; an unknown `<target-id>` →
  exit `12` with an "output not found" message and zero calls to `/api/v2/routes`; a fake
  `/api/v2/routes` returning `400` → exit `6`, `422` → exit `8`, a generic `500` → exit `3`,
  and a malformed `201` body (e.g. missing `routeId`) → exit `3` (FR-008, FR-010, FR-011,
  SC-003).
- [X] T008 [P] [US1] Unit test for route-creation rendering in
  tests/unit/render_route_test.go: `render.RenderRouteCreatedYAML(hub.Route{...}, message)` and
  `RenderRouteCreatedJSON` each emit exactly `routeId`, `status`, `message` — in that order —
  and JSON output round-trips through `encoding/json.Unmarshal` without error (FR-005).
- [X] T009 [P] [US1] Unit test for `route.Run`'s flag/positional parsing and validation
  failures in tests/unit/cli_route_test.go: missing `<input-path>` or `<target-path>` → exit
  `2` naming the missing argument; more than two positional arguments → exit `2`; an input path
  whose prefix is not `inputs`/`in` (e.g. `outputs/x`) → exit `2` identifying the invalid
  prefix; a target path whose prefix is not `outputs`/`out`/`groups`/`gr` (e.g. `inputs/x`) →
  exit `2`; a bare resource name with no id (e.g. `inputs`, `outputs`) on either argument →
  exit `2`; an unknown flag → exit `2`; an unreachable `--hub-url` → exit `4`, wrapped in
  `time.Now()`/`time.Since()` to assert well under 5 seconds; `--verbose` appends raw error
  detail on failure while its absence omits it. Assert every case that fails before a hub call
  is reached does so via a fake-hub request counter that stays at zero.

### Implementation for User Story 1

- [X] T010 [P] [US1] Implement `RenderRouteCreatedYAML(r hub.Route, message string) string` and
  `RenderRouteCreatedJSON(r hub.Route, message string) string` in internal/render/route.go:
  both read `RouteID`, `Status` off `r` plus the given `message` and render exactly those three
  fields (YAML as a bare record, JSON as a flat object via a small local payload struct — see
  data-model.md's "RoutingResult" section), never the full `hub.Route`. Makes T008 pass.
- [X] T011 [US1] Implement `Run(args []string, stdout, stderr io.Writer) int` in
  internal/cli/route/route.go: `flag.NewFlagSet("route", flag.ContinueOnError)` with `--json`,
  `--verbose`, `--hub-url` only; two required positionals `<input-path>` and `<target-path>`
  via the same re-parse-loop pattern `play.Run` uses; parse both via `respath.Parse`, returning
  a usage error (exit `2`) if either fails to parse, if the input path's `Kind` is not
  `respath.Inputs`, if the target path's `Kind` is not `respath.Outputs`/`respath.Groups`, or
  if either path's `ID` is empty (FR-002a, FR-002b); map the target `Kind` to `"SINGLE_OUTPUT"`/
  `"OUTPUT_GROUP"`; resolve the hub URL via `config.ResolveHubURL` and construct
  `hub.NewClient()`; call `hub.GetInput(ctx, client, baseURL, inputID)` — on a
  `*hub.NotFoundError`, report its message and exit `hub.ClassInputNotFound.ExitCode()`; on any
  other error, classify via `hub.ClassifyError` (FR-004); call
  `hub.ResolveTarget(ctx, client, baseURL, targetID, targetType)` — on a `*hub.NotFoundError`,
  report its message and exit `hub.ClassTargetNotFound.ExitCode()`; on any other error,
  classify via `hub.ClassifyError` (FR-003a); build `hub.CreateRouteRequest{InputID, TargetID,
  TargetType}` and call `hub.CreateRoute` — on a `*hub.NotFoundError`, exit
  `ClassTargetNotFound` (per data-model.md, `CreateRoute`'s own 404 fallback always names the
  target, mirroring `Playback`'s precedent; input absence is already caught earlier by the
  dedicated `GetInput` pre-check, so no `.Resource` branch is needed here); on any other
  error, classify via `hub.ClassifyError`; on success, construct the confirmation message
  `fmt.Sprintf("Routed %s to %s.", inputArg, targetArg)` (the two arguments exactly as typed)
  and render via `render.RenderRouteCreatedYAML`/`RenderRouteCreatedJSON`. Depends on T006,
  T010. Makes T007, T009 pass.
- [X] T012 [US1] Add a special case for `route` in cmd/sonora/main.go's `run()` function,
  alongside the existing `play` special case: `if args[0] == "route" { return
  route.Run(args[1:], stdout, stderr) }`, placed before the `get`/`list` verb switch (same
  reasoning as `play`'s placement — `route` has no verb token for the switch to dispatch on).
  Update `helpText` to add a `route inputs/<id> <outputs|groups>/<id>` entry and a matching
  example line. Depends on T011.

**Checkpoint**: `sonora route inputs/<id> outputs/<id>` works end-to-end for the single-output
case — MVP core. `go build ./... && go test ./...` passes.

---

## Phase 4: User Story 2 - Route an existing input to an output group (Priority: P2)

**Goal**: `sonora route inputs/<input-id> groups/<target-id>` works identically to US1 against
a group target, and a target identifier that collides between an output and a group is always
resolved by the path prefix the user typed, never ambiguously.

**Independent Test**: Run `sonora route inputs/<input-id> groups/<group-id>` and verify the
same success output as User Story 1, with the route's target reflecting the group; run it
against a colliding identifier that exists as both an output and a group, once with each
prefix, and verify each invocation targets only the stated type.

### Tests for User Story 2 (write first; MUST fail before implementation exists)

- [X] T013 [P] [US2] Extend tests/integration/route_test.go: `sonora route inputs/<input-id>
  groups/<target-id>` where `<target-id>` exists only as a group succeeds with the same
  three-field shape as US1 (FR-002, FR-005). Add a fake-hub scenario where one identifier
  exists as both an output and a group with distinct underlying resources: `sonora route
  inputs/<input-id> groups/<id>` targets the group — assert via a per-endpoint request counter
  that `/api/v2/groups/{id}` was called and `/api/v2/outputs/{id}` was not; `sonora route
  inputs/<input-id> outputs/<id>` with the same colliding `<id>` targets the output — assert
  the reverse (FR-003, acceptance scenarios 2–3). Also add the exact-type-mismatch case FR-003a
  calls out explicitly: an identifier that exists as an output but has no group of that same
  identifier — `sonora route inputs/<input-id> groups/<output-only-id>` → exit `12` with a
  "group not found" message, confirming the wrong-type lookup does not fall back to the other
  type's match; assert the symmetric case (`outputs/<group-only-id>` → exit `12`, "output not
  found") too.
- [X] T014 [P] [US2] Extend tests/unit/cli_route_test.go: target paths using the `groups`/`gr`
  prefixes both map to `targetType` `"OUTPUT_GROUP"`, and `outputs`/`out` both map to
  `"SINGLE_OUTPUT"`; the input path's `inputs`/`in` aliases are both accepted.

### Implementation for User Story 2

- [X] T015 [US2] No new production code is expected: `route.Run` (T011) already derives
  `targetType` solely from the target path's `respath.Kind` and calls the already-generic
  `hub.ResolveTarget` (research.md §1), so a `groups/<id>` target and a same-identifier
  collision are handled by the exact same code path as US1's `outputs/<id>` case. Run T013 and
  T014 against the existing implementation; if either reveals a gap, fix the target-type
  mapping or alias handling in internal/cli/route/route.go. Depends on T011.

**Checkpoint**: `sonora route` supports both single-output and group targets, with
identifier collisions resolved deterministically by prefix. `go build ./... && go test ./...`
passes; User Story 1's behavior is unaffected.

---

## Phase 5: Polish & Cross-Cutting Concerns

**Purpose**: Final quality gates spanning both stories.

- [X] T016 [P] Run `gofmt -l .`, `go vet ./...`, and the project's configured linter across all
  files touched by this feature (internal/hub/routes.go, internal/hub/errors.go,
  internal/render/route.go, internal/cli/route/, cmd/sonora/main.go, tests/); fix any findings
  (constitution Development Workflow).
- [X] T017 Execute every step in quickstart.md end-to-end (`go test ./...`, the manual
  single-output/group/collision/not-found/failure-path smoke tests against a fake hub) and
  confirm each mapped Success Criterion (SC-001…SC-005) holds, including that all exit codes in
  data-model.md's exit code table are mutually distinct (FR-010).

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — start immediately.
- **Foundational (Phase 2)**: Depends on Setup — BLOCKS both user stories (both issue
  `hub.GetInput`, `hub.ResolveTarget`, and `hub.CreateRoute`).
- **User Story 1 (Phase 3)**: Depends on Foundational. No dependency on US2.
- **User Story 2 (Phase 4)**: Depends on Foundational and on US1's `Run` (T011) existing —
  US2 adds no new flags or branches, only test coverage of behavior T011 already implements
  generically (see T015).
- **Polish (Phase 5)**: Depends on both user stories being complete.

### Within Each User Story

- Tests are written first and MUST fail before their corresponding implementation task starts
  (Principle VI).
- Within Foundational: T004 and T005 (different files) before T006 (which uses T004).
- Within US1: T010 (render) can proceed in parallel with Foundational's T006 having already
  landed; T011 depends on T006 and T010; T012 depends on T011.
- Within US2: T015 depends on T011 and is verification-only — no new implementation task is
  expected to precede it.

### Parallel Opportunities

- Foundational tests: T002, T003 — different files — can run in parallel.
- Foundational implementation: T004 and T005 — different files — can run in parallel; both
  must land before T006.
- US1 tests: T007, T008, T009 — three different files — can all run in parallel.
- US1 implementation: T010 can run in parallel with Foundational's T006 having already landed;
  T011 waits on both T006 and T010.
- US2 tests: T013 and T014 — different files — can run in parallel.
- T016 (lint/format) can run in parallel with T017 (quickstart validation).

---

## Parallel Example: Foundational Phase

```bash
# Launch both Foundational tests together:
Task: "Contract test for hub.CreateRoute in tests/contract/route_test.go"
Task: "Unit test for the two new ErrorClass values in tests/unit/hub_client_test.go"

# Once those tests fail as expected, launch the two independent struct/error tasks:
Task: "Define CreateRouteRequest in internal/hub/routes.go"
Task: "Extend internal/hub/errors.go with ClassInputNotFound/ClassTargetNotFound"

# Then implement hub.CreateRoute, which depends on both:
Task: "Implement hub.CreateRoute in internal/hub/routes.go"
```

---

## Implementation Strategy

### MVP First (User Story 1 only)

1. Complete Phase 1: Setup.
2. Complete Phase 2: Foundational (`hub.CreateRoute`, the two new error classes — blocks
   everything else).
3. Complete Phase 3: User Story 1 (`sonora route inputs/<id> outputs/<id>`).
4. **STOP and VALIDATE**: run T002–T009's tests plus the manual single-output smoke test
   against a fake hub (quickstart.md Scenario 1).
5. This is a shippable MVP: operators can connect an existing input to a single output in one
   command, without minting a duplicate input.

### Incremental Delivery

1. Setup + Foundational → foundation ready.
2. Add User Story 1 → validate independently → MVP complete.
3. Add User Story 2 → validate group targeting and identifier-collision resolution
   independently → doesn't change US1's single-output behavior.
4. Polish: lint/format pass, full quickstart.md run.

---

## Notes

- [P] tasks touch different files with no dependency on an incomplete task.
- [Story] labels trace every task back to spec.md's user stories.
- Tests are mandatory (Principle VI) — write and confirm each test fails before starting its
  paired implementation task.
- This feature touches no existing `outputs`/`inputs`/`groups`/`play`/`respath` code beyond
  reusing `GetInput`/`GetOutput`/`GetGroup`/`ResolveTarget` unchanged — every new
  type/function/flag is additive in new files (`internal/cli/route/`,
  `internal/render/route.go`) plus a targeted, purely-additive extension of
  `internal/hub/routes.go`, `internal/hub/errors.go`, and one new special case in
  `cmd/sonora/main.go`.
