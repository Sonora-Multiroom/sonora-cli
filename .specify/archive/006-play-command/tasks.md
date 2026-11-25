---

description: "Task list for `sonora play`"
---

# Tasks: Instant Playback (`play`)

**Input**: Design documents from `/specs/006-play-command/`

**Prerequisites**: [plan.md](plan.md), [spec.md](spec.md), [research.md](research.md),
[data-model.md](data-model.md), [contracts/cli-play.md](contracts/cli-play.md),
[quickstart.md](quickstart.md)

**Tests**: Tests are **mandatory**, not optional, for this project — constitution Principle
VI ("Test-First Development (NON-NEGOTIABLE)") requires every command and API client method
to have a test written, reviewed, and watched to fail *before* the implementing code is
written. `play` is additionally named by the constitution as a core flow requiring
contract/integration coverage before merge. Every implementation task below has a
corresponding test task that MUST land first.

**Organization**: Tasks are grouped by user story (from spec.md), in priority order
(P1 → P2 → P3), so each story is independently implementable and testable. Where two stories
share priority P3 (US3, US4), spec order is preserved between them.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies on incomplete tasks)
- **[Story]**: Which user story this task belongs to (US1–US4)
- File paths are exact and relative to the repo root

## Path Conventions

Single Go project (established by `001-list-outputs`, unchanged): `cmd/sonora/`,
`internal/{hub,render,cli/play,config}/`, `tests/{unit,contract,integration}/`.

---

## Phase 1: Setup

**Purpose**: Confirm the existing `001`–`005` codebase this feature builds on is in a
known-good state before touching it.

- [X] T001 Run `go build ./...` and `go test ./...` from the repo root on branch
  `006-play-command` and confirm everything passes, establishing the baseline this feature's
  changes are layered on (no files modified by this task).

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Unlike prior features (two independent read-only endpoints split cleanly across
two user stories), `play` is a single `POST` endpoint with a client-side target-resolution
step every user story depends on to do anything at all — none of US1–US4 can be independently
tested without `hub.Playback`, `hub.ResolveTarget`, and the new error classes existing first
(data-model.md, research.md §1 and §3).

**⚠️ CRITICAL**: No User Story work can begin until this phase is complete.

### Tests for Foundational work (write first; MUST fail before implementation exists)

- [X] T002 [P] Contract test for `hub.Playback` in tests/contract/play_test.go: an
  `httptest.Server` serving `POST /api/v2/play` per `playback` in `api/openapi.json` — assert
  the request body always includes `uri`/`targetId`/`targetType` and omits `displayName`/
  `volume` when not set (present when set); a `200 PlaybackResponse` body decodes into
  `hub.PlaybackResponse` (`inputId`, nested `route` via the existing `hub.Route`, `message`);
  a `400`/`422`/`502`/`503` `ErrorResponse` body decodes into `*hub.APIError{StatusCode,
  Title, Detail}`; a `404` yields `*hub.NotFoundError{Resource: "target", ID: ...}`; a
  non-JSON or empty error body on any of those statuses falls back to `*hub.StatusError`
  rather than failing to classify at all; a malformed `200` body (missing `inputId`, or
  `route.routeId`/`route.status` empty) yields `*hub.DecodeError` per FR-012. Also assert the
  mock server receives exactly one request per `Playback` call (FR-010 — no automatic retry).
- [X] T003 [P] Contract test for `hub.ResolveTarget` in tests/contract/play_resolve_test.go: a
  fake hub serving `GET /api/v2/outputs/{id}` and `GET /api/v2/groups/{id}` (reusing the
  existing response shapes from `hub.GetOutput`/`hub.GetGroup`) — assert, per
  data-model.md's Target Resolution State table: default mode (`forceGroup=forceOutput=false`)
  returns `SINGLE_OUTPUT` when only the output lookup succeeds, `OUTPUT_GROUP` when only the
  group lookup succeeds, `*hub.AmbiguousTargetError{ID}` when both succeed, and
  `*hub.NotFoundError{Resource: "target"}` when neither succeeds; `forceGroup=true` calls only
  the groups endpoint and returns `*hub.NotFoundError{Resource: "group"}` on a miss even when
  an output with that ID exists (FR-003b); `forceOutput=true` is the symmetric case for
  outputs. Assert request counts confirm the un-needed endpoint is never called in forced
  modes.
- [X] T004 [P] Unit test for `hub.ClassifyError`'s new mappings, extending
  tests/unit/hub_client_test.go: `*hub.APIError{StatusCode: 400}` → `ClassValidation`;
  `{StatusCode: 422}` → `ClassRouteFailed`; `{StatusCode: 502}` → `ClassSourceUnreachable`;
  `{StatusCode: 503}` → `ClassServiceUnavailable`; an `*hub.APIError` with any other status
  code → `ClassHub` (fallback); `*hub.AmbiguousTargetError` → `ClassAmbiguous`. Also assert
  every exit code from data-model.md's exit code table is distinct (FR-011) and that the
  pre-existing `NotFoundError`/`StatusError`/network mappings are unchanged.

### Implementation for Foundational work

- [X] T005 [P] Define `PlaybackRequest` and `PlaybackResponse` structs in
  internal/hub/play.go, mapping field-for-field to `#/components/schemas/PlaybackRequest`/
  `PlaybackResponse` in `api/openapi.json` per data-model.md's field tables: `PlaybackRequest`
  has `URI`/`TargetID`/`TargetType` (`string`, required) and `DisplayName *string`/`Volume
  *int` (both `omitempty`); `PlaybackResponse` has `InputID string`, `Route hub.Route`
  (reusing the existing struct), `Message string`. No functions yet, structs only.
- [X] T006 [P] Extend internal/hub/errors.go: add `ClassValidation` (exit 6),
  `ClassAmbiguous` (exit 7), `ClassRouteFailed` (exit 8), `ClassSourceUnreachable` (exit 9),
  `ClassServiceUnavailable` (exit 10) to the `ErrorClass` enum and its `ExitCode()` switch; add
  `APIError{StatusCode int; Title, Detail string}` with an `Error()` method; add
  `AmbiguousTargetError{ID string}` with an `Error()` method; extend `ClassifyError` to
  recognize both new types and route `APIError` by `StatusCode` to the five new classes (with
  any other status code falling back to `ClassHub`), per research.md §3. Existing
  classifications for `NotFoundError`/`StatusError`/`DecodeError`/network errors are
  unchanged. Makes T004 pass.
- [X] T007 Implement `Playback(ctx context.Context, client *http.Client, baseURL string, req
  PlaybackRequest) (*PlaybackResponse, error)` in internal/hub/play.go: `POST
  {baseURL}/api/v2/play` with `req` JSON-encoded as the body; on `200`, decode into
  `PlaybackResponse` and reject as `*DecodeError` if `InputID`, `Route.RouteID`, or
  `Route.Status` is empty; on `404`, return `*NotFoundError{Resource: "target", ID:
  req.TargetID}`; on `400`/`422`/`502`/`503`, attempt to decode the body as
  `#/components/schemas/ErrorResponse` into `*APIError{StatusCode, Title, Detail}`, falling
  back to `*StatusError{StatusCode}` if that decode fails; any other non-2xx status returns
  `*StatusError{StatusCode}`. Depends on T005, T006. Makes T002 pass.
- [X] T008 Implement `ResolveTarget(ctx context.Context, client *http.Client, baseURL,
  targetID string, forceGroup, forceOutput bool) (targetType string, err error)` in
  internal/hub/play.go, per data-model.md's Target Resolution State table: `forceGroup` calls
  only `GetGroup` (returning `"OUTPUT_GROUP"` or its `*NotFoundError`); `forceOutput` calls
  only `GetOutput` (symmetric); neither flag set calls `GetOutput` then `GetGroup`
  sequentially and returns `*AmbiguousTargetError{ID: targetID}` if both succeed,
  `"SINGLE_OUTPUT"`/`"OUTPUT_GROUP"` if exactly one succeeds, or `*NotFoundError{Resource:
  "target", ID: targetID}` if neither succeeds. Depends on T006 (for `AmbiguousTargetError`);
  reuses the existing `GetOutput`/`GetGroup` from internal/hub/outputs.go and
  internal/hub/groups.go unchanged. Makes T003 pass.

**Checkpoint**: `go build ./...` passes with the full hub-layer support for playback and
target resolution in place; no CLI surface exists yet. User story implementation can begin.

---

## Phase 3: User Story 1 - Play a URI to a single output (Priority: P1) 🎯 MVP

**Goal**: `sonora play <uri> <target-id>` resolves an existing, enabled output as the target
by default, calls the hub, and prints the created input's identifier, the route's identifier
and status, and a confirmation message, in YAML by default or JSON with `--json`.

**Independent Test**: Run the play command with a valid URI and the identifier of a single
existing output, then verify the command immediately reports a created input identifier and
the route's identifier and status exactly as returned by the hub, without waiting for the
route to reach a further state; verify the same result is valid, parseable structured data
under `--json`.

### Tests for User Story 1 (write first; MUST fail before implementation exists)

- [X] T009 [P] [US1] Integration tests for the play command's success and hub-error paths in
  tests/integration/play_test.go: build/run the `sonora` binary (following the
  `binPath`/`TestMain` pattern in tests/integration/outputs_list_test.go) against a fake hub
  serving `/api/v2/play`, `/api/v2/outputs/{id}`, and `/api/v2/groups/{id}`, where `<id>`
  exists only as an output. Assert the success case: `sonora play <uri> <id>` prints
  `inputId`/`routeId`/`status`/`message` in YAML with exit code `0`, and the same fields parse
  via `encoding/json` under `--json`; wrap this invocation in `time.Now()`/`time.Since()` and
  assert completion in under 5 seconds (SC-003). Also assert, for the same `<id>`, that a fake
  `/api/v2/play` returning each of the hub's documented failure statuses causes the real
  binary to exit with the correct, distinct code and a non-empty stderr message: `400` → exit
  `6`, `422` → exit `8`, `502` → exit `9`, `503` → exit `10`, a generic `500` → exit `3`, and a
  malformed `200` body (e.g. missing `inputId`) → exit `3` (FR-009, FR-011, FR-012, SC-004) —
  this is the full-CLI counterpart to T002's contract-level `*hub.APIError`/`*hub.DecodeError`
  assertions, closing the gap where only the hub package's return types, not the compiled
  binary's exit codes, would otherwise be verified end-to-end.
- [X] T010 [P] [US1] Unit test for playback rendering in tests/unit/render_play_test.go:
  `render.RenderPlaybackYAML(hub.PlaybackResponse{...})` and `RenderPlaybackJSON` each emit
  exactly `inputId`, `routeId`, `status`, `message` — in that order — reading `RouteID`/
  `Status` off the nested `Route`, and JSON output round-trips through
  `encoding/json.Unmarshal` without error (FR-006).
- [X] T011 [P] [US1] Unit test for `play.Run`'s flag/positional parsing and failure paths in
  tests/unit/cli_play_test.go: missing `<uri>` or `<target-id>` → exit `2` naming the missing
  argument; more than two positional arguments → exit `2`; an unknown flag → exit `2`; an
  unreachable `--hub-url` → exit `4`, wrapped in `time.Now()`/`time.Since()` to assert well
  under 5 seconds; `--verbose` appends raw error detail on failure while its absence omits it.

### Implementation for User Story 1

- [X] T012 [P] [US1] Implement `RenderPlaybackYAML(p hub.PlaybackResponse) string` and
  `RenderPlaybackJSON(p hub.PlaybackResponse) string` in internal/render/play.go: both read
  `InputID`, `Route.RouteID`, `Route.Status`, `Message` off `p` and render exactly those four
  fields (YAML as a bare record, JSON as a flat object via a small local payload struct — see
  data-model.md's "PlaybackResult (rendered view)" section), never the full nested `Route`.
  Makes T010 pass.
- [X] T013 [US1] Implement `Run(args []string, stdout, stderr io.Writer) int` in
  internal/cli/play/play.go: `flag.NewFlagSet("play", flag.ContinueOnError)` with `--json`,
  `--verbose`, `--hub-url` (no `--group`/`--output`/`--volume`/`--name` yet — added in
  US2–US4); two required positionals `<uri>` and `<target-id>` via the same re-parse-loop
  pattern as `routes.RunGet`/`groups.RunGet`, extended to collect two positionals instead of
  one (usage error, exit `2`, if the count is not exactly two); resolves the hub URL via
  `config.ResolveHubURL`; calls `hub.NewClient()`, then `hub.ResolveTarget(ctx, client,
  baseURL, targetID, false, false)`, then builds a `hub.PlaybackRequest{URI: uri, TargetID:
  targetID, TargetType: resolvedType}` (no `DisplayName`/`Volume` yet) and calls
  `hub.Playback`; classifies any error via `hub.ClassifyError` (from either call); on success
  renders via `render.RenderPlaybackYAML`/`RenderPlaybackJSON`. Depends on T007, T008, T012.
  Makes T009, T011 pass. **Note**: implemented with the full US2–US4 flag surface
  (`--group`/`--output`/`--volume`/`--name`) in place from the start rather than incrementally
  — see T017/T020/T022 notes below.
- [X] T014 [US1] Add a special case for `play` in cmd/sonora/main.go's `run()` function,
  checked *before* the existing `len(args) < 2` usage-error gate and the `noun, verb, rest :=
  args[0], args[1], args[2:]` split: both of those assume every command has a verb token, so
  for `sonora play <uri> <target-id>` they would bind `<uri>` (`args[1]`) to the unused `verb`
  variable and exclude it from `rest`, silently dropping it before it ever reaches
  `play.Run`. Immediately after the existing `--version`/`-v` check (and before the
  `len(args) < 2` gate), add: `if len(args) >= 1 && args[0] == "play" { return
  play.Run(args[1:], stdout, stderr) }`, so `play.Run` receives every token after `play`
  unchanged — including the zero/one/two-positional cases its own parsing (T013) validates.
  Depends on T013.

**Checkpoint**: `sonora play <uri> <output-id>` works end-to-end for the default,
single-output case — MVP core. `go build ./... && go test ./...` passes.

---

## Phase 4: User Story 2 - Play a URI to an output group (Priority: P2)

**Goal**: `sonora play <uri> <target-id>` also resolves an existing group as the target by
default, and `--group`/`--output` let the caller force either type when an identifier
collides across both, failing clearly when the forced type doesn't match or when both flags
are given together.

**Independent Test**: Run the play command with a valid URI and the identifier of an existing
output group, and verify the same success output as User Story 1, with the route's target
reflecting the group; verify `--group`/`--output` each force their stated type regardless of
what auto-detection would have picked when an identifier exists as both.

### Tests for User Story 2 (write first; MUST fail before implementation exists)

- [X] T015 [P] [US2] Extend tests/integration/play_test.go: `sonora play <uri> <id>` where
  `<id>` exists only as a group succeeds identically to US1's shape (targeting resolves to
  `OUTPUT_GROUP`); where `<id>` exists as both an output and a group and neither `--group` nor
  `--output` is given, the command exits `7` with a message naming both possibilities
  (FR-003a); `--group <id>` forces the group even when an output of the same ID also exists,
  and `--output <id>` is the symmetric case; `--group <id>` where `<id>` exists only as an
  output exits `5` with a "group not found" message (FR-003b), and `--output` is the symmetric
  case; `--group` and `--output` given together exits `2` with a usage error (FR-002a).
- [X] T016 [P] [US2] Extend tests/unit/cli_play_test.go: asserting `--group` and `--output`
  given together is rejected as a usage error (exit `2`) *before* any HTTP request is made —
  verify via a fake hub request counter that stays at zero.

### Implementation for User Story 2

- [X] T017 [US2] Add `--group` and `--output` bool flags to `Run` in
  internal/cli/play/play.go: immediately after parsing, if both are set, print a usage error
  and return `hub.ClassUsage.ExitCode()` without calling `ResolveTarget`/`Playback`; otherwise
  pass their values through as `ResolveTarget`'s `forceGroup`/`forceOutput` arguments in place
  of the two hardcoded `false`s from T013. Depends on T013. Makes T015, T016 pass (the
  ambiguous/not-found exit codes themselves are already correct from Foundational's T006/T008
  — this task only wires the new flags through to `ResolveTarget`). Delivered as part of T013.

**Checkpoint**: `sonora play` supports both single-output and group targets, plus explicit
disambiguation. `go build ./... && go test ./...` passes; User Story 1's default behavior is
unaffected.

---

## Phase 5: User Story 3 - Set the starting volume (Priority: P3)

**Goal**: `sonora play <uri> <target-id> --volume N` sets the target's volume before playback
starts, rejecting an out-of-range value before any request reaches the hub.

**Independent Test**: Run the play command with a volume option set to a valid level and
verify the response confirms the requested volume was applied before playback started; run it
with a value outside 0-100 and verify a clear usage/validation error before any request
reaches the hub.

### Tests for User Story 3 (write first; MUST fail before implementation exists)

- [X] T018 [P] [US3] Extend tests/unit/cli_play_test.go: `--volume -1` and `--volume 150`
  each exit `6` with a clear range-error message, verified via a fake hub request counter that
  stays at zero (FR-004 — checked before any request); `--volume 0` and `--volume 100`
  (boundary values) are accepted and proceed to call the hub.
- [X] T019 [P] [US3] Extend tests/integration/play_test.go: `--volume 50` results in a
  `POST /api/v2/play` request body whose `volume` field is `50` (captured via the fake hub),
  and omitting `--volume` results in a request body with no `volume` field at all (`null`/
  absent, matching `PlaybackRequest.Volume`'s `omitempty` pointer semantics).

### Implementation for User Story 3

- [X] T020 [US3] Add a `--volume` int flag to `Run` in internal/cli/play/play.go, using `-1`
  as the "not provided" sentinel (the valid range 0-100 never produces `-1`): if the parsed
  value is not `-1`, validate it is within `[0, 100]` and return
  `hub.ClassValidation.ExitCode()` immediately (before `ResolveTarget`/`Playback`) if not;
  when valid and provided, set `PlaybackRequest.Volume` to a pointer to that value; when not
  provided, leave `Volume` `nil`. Depends on T017. Makes T018, T019 pass. Delivered as part of
  T013, with one deviation from this task's original design: T018 requires `--volume -1`
  itself (explicitly passed) to be rejected as out-of-range (exit 6), which is impossible to
  distinguish from "not provided" using `-1` as a bare sentinel value. Implemented instead by
  tracking whether `--volume` was explicitly set via `fs.Visit`, so an explicit `--volume -1`
  is correctly rejected while an omitted `--volume` still leaves `Volume` `nil`.

**Checkpoint**: Starting volume is configurable and validated client-side. `go build ./... &&
go test ./...` passes; US1/US2 behavior without `--volume` is unaffected.

---

## Phase 6: User Story 4 - Name the playback session (Priority: P3)

**Goal**: `sonora play <uri> <target-id> --name NAME` gives the ephemeral input created for
this playback session a recognizable display name.

**Independent Test**: Run the play command with a display name option and verify the created
input's name matches what was supplied, reflected in the confirmation output.

### Tests for User Story 4 (write first; MUST fail before implementation exists)

- [X] T021 [P] [US4] Extend tests/integration/play_test.go: `--name "Kitchen Radio"` results
  in a `POST /api/v2/play` request body whose `displayName` field is `"Kitchen Radio"`
  (captured via the fake hub), and the command's rendered `message` field reflects whatever
  the fake hub's response echoes back for that name; omitting `--name` results in a request
  body with no `displayName` field at all.

### Implementation for User Story 4

- [X] T022 [US4] Add a `--name` string flag to `Run` in internal/cli/play/play.go: when
  non-empty, set `PlaybackRequest.DisplayName` to a pointer to its value; when empty (not
  given), leave `DisplayName` `nil`. Depends on T020. Makes T021 pass. Delivered as part of
  T013.

**Checkpoint**: All four user stories are independently functional — `sonora play` now
matches the full contract in [contracts/cli-play.md](contracts/cli-play.md). `go build ./...
&& go test ./...` passes.

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Final quality gates spanning all four stories.

- [X] T023 [P] Run `gofmt -l .`, `go vet ./...`, and the project's configured linter across
  all files touched by this feature (internal/hub/play.go, internal/hub/errors.go,
  internal/render/play.go, internal/cli/play/, cmd/sonora/main.go, tests/); fix any findings
  (constitution Development Workflow).
- [X] T024 Execute every step in quickstart.md end-to-end (`go test ./...`, the manual
  success/group/ambiguous/volume/name/failure-path smoke tests against a fake hub) and
  confirm each mapped Success Criterion (SC-001…SC-006) holds, including that all nine
  failure-class exit codes in data-model.md's exit code table are mutually distinct (FR-011,
  SC-004).

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — start immediately.
- **Foundational (Phase 2)**: Depends on Setup — BLOCKS all user stories (every story issues
  `hub.Playback` and, for its default path, `hub.ResolveTarget`).
- **User Story 1 (Phase 3)**: Depends on Foundational. No dependency on US2–US4.
- **User Story 2 (Phase 4)**: Depends on Foundational and on US1's `Run` (T013) existing to
  add `--group`/`--output` to. Independent of US3/US4.
- **User Story 3 (Phase 5)**: Depends on Foundational and on US2's `Run` (T017) existing to
  add `--volume` to. Independent of US4 (order chosen to match spec.md's story sequence, not a
  functional requirement — `--volume` and `--name` touch disjoint fields).
- **User Story 4 (Phase 6)**: Depends on Foundational and on US3's `Run` (T020) existing to
  add `--name` to. Independent of US2/US3's actual behavior.
- **Polish (Phase 7)**: Depends on all four user stories being complete.

### Within Each User Story

- Tests are written first and MUST fail before their corresponding implementation task
  starts (Principle VI).
- Within Foundational: T005 and T006 (different files) before T007 and T008 (both use T006;
  T007 also uses T005).
- Within US1: T012 (render) can proceed in parallel with T007/T008 having already landed;
  T013 depends on T007, T008, T012; T014 depends on T013 and must special-case `play` ahead
  of `main.go`'s existing verb-assuming dispatch logic (see T014's description) rather than
  simply adding a switch case, since `play` has no verb token to split on.
- Within US2: T017 only (extends T013's flag set; no new hub/render code).
- Within US3: T020 only (extends T017's flag set).
- Within US4: T022 only (extends T020's flag set).

### Parallel Opportunities

- Foundational tests: T002, T003, T004 — three different files, no shared dependency — can
  all run in parallel.
- Foundational implementation: T005 and T006 — different files — can run in parallel; both
  must land before T007/T008 (which can then also run in parallel with each other).
- US1 tests: T009, T010, T011 — three different files — can all run in parallel.
- US1 implementation: T012 can run in parallel with the Foundational T007/T008 work; T013
  waits on all three.
- US2 tests: T015 and T016 — different files — can run in parallel.
- US3 tests: T018 and T019 — different files — can run in parallel.
- T023 (lint/format) can run in parallel with T024 (quickstart validation).

---

## Parallel Example: Foundational Phase

```bash
# Launch all Foundational tests together:
Task: "Contract test for hub.Playback in tests/contract/play_test.go"
Task: "Contract test for hub.ResolveTarget in tests/contract/play_resolve_test.go"
Task: "Unit test for ClassifyError's new mappings in tests/unit/hub_client_test.go"

# Once those tests fail as expected, launch the two independent struct/error tasks:
Task: "Define PlaybackRequest/PlaybackResponse in internal/hub/play.go"
Task: "Extend internal/hub/errors.go with the five new error classes and types"

# Then, once both land, launch the two API-call implementations in parallel:
Task: "Implement hub.Playback in internal/hub/play.go"
Task: "Implement hub.ResolveTarget in internal/hub/play.go"
```

---

## Implementation Strategy

### MVP First (User Story 1 only)

1. Complete Phase 1: Setup.
2. Complete Phase 2: Foundational (`hub.Playback`, `hub.ResolveTarget`, new error classes —
   blocks everything else).
3. Complete Phase 3: User Story 1 (`sonora play <uri> <output-id>`, default detection).
4. **STOP and VALIDATE**: run T002–T011's tests plus the manual single-output smoke test
   against a fake hub (quickstart.md Scenario 1).
5. This is a shippable MVP: operators can start instant playback to a single output in one
   command.

### Incremental Delivery

1. Setup + Foundational → foundation ready.
2. Add User Story 1 → validate independently → MVP complete.
3. Add User Story 2 → validate group targeting and disambiguation independently → doesn't
   change US1's default single-output behavior.
4. Add User Story 3 → validate `--volume` independently → doesn't change behavior when the
   flag is omitted.
5. Add User Story 4 → validate `--name` independently → command now matches the full contract
   in [contracts/cli-play.md](contracts/cli-play.md).
6. Polish: lint/format pass, full quickstart.md run.

---

## Notes

- [P] tasks touch different files with no dependency on an incomplete task.
- [Story] labels trace every task back to spec.md's user stories.
- Tests are mandatory (Principle VI) — write and confirm each test fails before starting its
  paired implementation task.
- This feature touches no existing `outputs`/`inputs`/`routes`/`groups` code beyond reusing
  `GetOutput`/`GetGroup` unchanged — every new type/function/flag is additive in new files
  (`internal/hub/play.go`, `internal/render/play.go`, `internal/cli/play/`) plus a targeted,
  purely-additive extension of `internal/hub/errors.go` and one new `case "play":` branch in
  `cmd/sonora/main.go`.
