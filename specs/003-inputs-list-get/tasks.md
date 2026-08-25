---

description: "Task list for `sonora inputs list` and `sonora inputs get`"
---

# Tasks: List and Get Audio Inputs

**Input**: Design documents from `/specs/003-inputs-list-get/`

**Prerequisites**: [plan.md](plan.md), [spec.md](spec.md), [research.md](research.md),
[data-model.md](data-model.md), [contracts/cli-inputs-list.md](contracts/cli-inputs-list.md),
[contracts/cli-inputs-get.md](contracts/cli-inputs-get.md), [quickstart.md](quickstart.md)

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
`internal/{hub,render,cli/inputs,config}/`, `tests/{unit,contract,integration}/`.

---

## Phase 1: Setup

**Purpose**: Confirm the existing `001-list-outputs`/`002-outputs-get` codebase this feature
builds on is in a known-good state before touching it.

- [X] T001 Run `go build ./...` and `go test ./...` from the repo root on branch
  `003-inputs-list-get` and confirm everything passes, establishing the baseline this
  feature's changes are layered on (no files modified by this task).

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Define the `Input` entity once, since both P1 user stories (US1 `inputs list`,
US2 `inputs get`) decode hub responses into it and render it — a shared prerequisite rather
than something either story owns individually (data-model.md).

**⚠️ CRITICAL**: No User Story work can begin until this phase is complete.

- [X] T002 Define the `Input` struct in internal/hub/inputs.go, mapping field-for-field to
  `#/components/schemas/InputResponse` in `api/openapi.json`: `InputID` (`inputId`),
  `DisplayName` (`displayName`), `URI` (`uri`), `Enabled` (`enabled`), `AutoRemove`
  (`autoRemove`), `Source` (`source`, string), `CreatedAt` (`createdAt`, `*string`,
  nullable), `Pauseable` (`pauseable`) — per data-model.md's field table. No functions yet,
  struct only.

**Checkpoint**: `go build ./...` passes with the new `Input` type in place. User story
implementation can begin.

---

## Phase 3: User Story 1 - View currently active inputs (Priority: P1) 🎯 MVP

**Goal**: `sonora inputs list` fetches and displays only enabled inputs by default, each
showing `inputId`/`displayName`/`uri`/`source`/`enabled`/`autoRemove`/`pauseable`/
`createdAt`, in YAML.

**Independent Test**: Run the list command with no flags against a hub with enabled and
disabled inputs and verify only the enabled ones are shown, each with all eight fields.

### Tests for User Story 1 (write first; MUST fail before implementation exists)

- [X] T003 [P] [US1] Contract test for `hub.ListInputs`'s success and error paths in
  tests/contract/inputs_list_test.go: an `httptest.Server` serving `GET /api/v2/inputs` per
  `listInputs` in `api/openapi.json` — assert the request includes `includeDisabled=false`
  by default, a `200` array body decodes into `[]hub.Input` correctly (including a static
  input's `createdAt: null` and an ephemeral input's populated `createdAt`), a non-2xx
  response yields a `*hub.StatusError`, and a malformed body (missing `inputId`, or an
  unrecognized `source` value) yields a `*hub.DecodeError` per FR-017. Also assert the mock
  server receives exactly one request for a single `ListInputs` call (FR-013 — no automatic
  retry on failure).
- [X] T004 [P] [US1] Integration test for the list command's default success path in
  tests/integration/inputs_list_test.go: build/run the `sonora` binary (following the
  `binPath`/`TestMain` pattern in tests/integration/outputs_list_test.go) against a mock
  server returning enabled and disabled inputs, and assert `sonora inputs list` (no flags)
  prints only the enabled inputs with all eight fields in YAML, exit code `0`. Also assert
  a mock server returning zero enabled inputs produces the clear "no inputs found" output
  with exit code `0` (FR-016) — distinct from every non-zero failure exit code, so "no
  inputs" is never mistaken for a failure (SC-006) — and wrap a success invocation in
  `time.Now()`/`time.Since()` to assert completion in under 1 second (SC-001).
- [X] T005 [P] [US1] Unit test for list YAML rendering in tests/unit/render_inputs_test.go:
  `render.RenderYAML([]hub.Input{...})` emits all eight fields per input in the documented
  order (`inputId`, `displayName`, `uri`, `source`, `enabled`, `autoRemove`, `pauseable`,
  `createdAt`), a static input's `createdAt` renders as bare `null` (never omitted, never
  quoted), an ephemeral input's `createdAt` renders as a quoted string, and a zero-length
  slice renders the `# no inputs found` / `inputs: []` shape (FR-016).
- [X] T006 [P] [US1] Unit test for `inputs.RunList`'s flag parsing/dispatch and failure
  paths in tests/unit/cli_inputs_test.go: unreachable `--hub-url` → exit `4`; unknown flag →
  exit `2`; unexpected positional argument → exit `2`; `--verbose` appends raw error detail
  on failure while its absence omits it. For the unreachable-`--hub-url` case, wrap the call
  in `time.Now()`/`time.Since()` and assert it returns well under 5 seconds (SC-004).

### Implementation for User Story 1

- [X] T007 [P] [US1] Implement `ListInputs(ctx context.Context, client *http.Client,
  baseURL string, includeDisabled bool) ([]Input, error)` in internal/hub/inputs.go: builds
  `{baseURL}/api/v2/inputs?includeDisabled={bool}`, issues one GET, decodes a `200` array
  body into `[]Input` (nil → `[]Input{}`), and rejects a decoded record as malformed
  (`*DecodeError`) if `InputID`/`DisplayName` is empty or `Source` isn't `"STATIC"` or
  `"EPHEMERAL"` (FR-017, data-model.md's validation rule); any other non-2xx status returns
  `*StatusError`. Depends on T002. Makes T003 pass.
- [X] T008 [P] [US1] Implement `RenderYAML(inputs []hub.Input) string` in
  internal/render/inputs.go: emits the `inputs:` list shape with all eight fields per input
  in the documented order, `createdAt: null` bare and unquoted when `nil`, `createdAt:
  "<value>"` quoted otherwise, and the `# no inputs found` / `inputs: []` shape for a
  zero-length slice — mirroring `internal/render/outputs.go`'s `RenderYAML` structure. Makes
  T005 pass.
- [X] T009 [US1] Implement `RunList(args []string, stdout, stderr io.Writer) int` in
  internal/cli/inputs/list.go: `flag.NewFlagSet("inputs list", flag.ContinueOnError)` with
  `--verbose` and `--hub-url` (mirroring `outputs.RunList`'s flag definitions; no
  `--include-disabled` or `--json` yet — added in US3/US5), no positional arguments (any
  → usage error, exit `2`), resolves the hub URL via `config.ResolveHubURL`, calls
  `hub.NewClient()` + `hub.ListInputs(ctx, client, baseURL, false)`, classifies any error via
  `hub.ClassifyError`, and on success prints `render.RenderYAML(inputs)` to stdout and
  returns `0`. Depends on T007, T008. Makes T004, T006 pass.
- [X] T010 [US1] Add a new `case "inputs":` to the noun switch in cmd/sonora/main.go,
  routing verb `"list"` to `inputs.RunList(rest, stdout, stderr)` (an unrecognized verb
  under `"inputs"` returns the existing usage-error shape, exit `2`; `"get"` is added in
  US2). Depends on T009.

**Checkpoint**: `sonora inputs list` works end-to-end for the default (enabled-only) case —
first half of the MVP. `go build ./... && go test ./...` passes.

---

## Phase 4: User Story 2 - Look up a specific input by identifier (Priority: P1) 🎯 MVP

**Goal**: `sonora inputs get <input-id>` fetches and displays one input's full state
(all eight fields) in YAML by default, regardless of its enabled/disabled state.

**Independent Test**: Run the get command with the identifier of a known input (try both an
enabled and a disabled one, and both a static and an ephemeral one) and verify all eight
fields display correctly in every case.

### Tests for User Story 2 (write first; MUST fail before implementation exists)

- [X] T011 [P] [US2] Contract test for `hub.GetInput`'s success and hub-error paths in
  tests/contract/inputs_get_test.go: an `httptest.Server` serving `GET
  /api/v2/inputs/{inputId}` per `getInput` in `api/openapi.json` — assert the request path
  includes the given ID, a `200 InputResponse` body decodes into `hub.Input` correctly for
  both a static input (`source: STATIC`, `createdAt: null`) and an ephemeral one (`source:
  EPHEMERAL`, `createdAt` populated), a non-2xx/non-404 response (e.g. `500`) yields a
  `*hub.StatusError`, and a malformed `200` body yields a `*hub.DecodeError`. Also assert
  the mock server receives exactly one request for a single `GetInput` call (FR-013 — no
  automatic retry on failure).
- [X] T012 [P] [US2] Integration test for the get command's success path in
  tests/integration/inputs_get_test.go: build/run the `sonora` binary against a mock server,
  and assert `sonora inputs get <id>` prints all eight fields correctly in YAML for both an
  enabled input and a disabled one (FR-007 — no filtering, unlike `list`), with exit code
  `0`. Wrap a success invocation in `time.Now()`/`time.Since()` and assert completion in
  under 1 second (SC-002).
- [X] T013 [P] [US2] Extend tests/unit/render_inputs_test.go with a test for single-input
  YAML rendering: `render.RenderInputYAML(hub.Input{...})` emits all eight fields as a bare
  (non-list) record in the documented order, and a static input shows `createdAt: null`
  explicitly (never omitted).
- [X] T014 [P] [US2] Unit test for `inputs.RunGet`'s argument handling and failure paths in
  tests/unit/cli_inputs_get_test.go: missing `<input-id>` → exit `2`; more than one
  positional argument → exit `2`; unreachable `--hub-url` → exit `4` (wrapped in
  `time.Now()`/`time.Since()` to assert well under 5 seconds, SC-004); `--verbose` appends
  raw error detail on failure while its absence omits it.

### Implementation for User Story 2

- [X] T015 [P] [US2] Implement `GetInput(ctx context.Context, client *http.Client, baseURL,
  inputID string) (*Input, error)` in internal/hub/inputs.go: builds
  `{baseURL}/api/v2/inputs/{inputId}` (path-escaped), issues one GET, decodes a `200` body
  into `Input` (same `InputID`/`DisplayName`/`Source` validation as `ListInputs`, wrapped as
  `*DecodeError`), and returns `*StatusError{StatusCode}` for any non-2xx response other
  than `404` (404 handling is added in US4 — for now treat it as any other `*StatusError`,
  which US2's own tests don't exercise). Depends on T002. Makes T011 pass.
- [X] T016 [P] [US2] Implement `RenderInputYAML(i hub.Input) string` in
  internal/render/inputs.go: emits the same eight fields as `RenderYAML`'s per-record shape,
  but as a single top-level record (no `inputs:` wrapper), every field always explicit,
  including bare `createdAt: null`. Makes T013 pass.
- [X] T017 [US2] Implement `RunGet(args []string, stdout, stderr io.Writer) int` in
  internal/cli/inputs/get.go: `flag.NewFlagSet("inputs get", flag.ContinueOnError)` with
  `--verbose` and `--hub-url` (no `--json` yet — added in US5), one required positional
  `<input-id>` via the same re-parse-loop pattern as `outputs.RunGet` (so the identifier can
  appear before or after flags; zero or >1 positional args → usage error, exit `2`),
  resolves the hub URL via `config.ResolveHubURL`, calls `hub.NewClient()` +
  `hub.GetInput`, classifies any error via `hub.ClassifyError`, and on success prints
  `render.RenderInputYAML(*input)` to stdout and returns `0`. Depends on T015, T016. Makes
  T012, T014 pass.
- [X] T018 [US2] Add a `case "get":` to the `"inputs"` noun's verb switch in
  cmd/sonora/main.go, routing to `inputs.RunGet(rest, stdout, stderr)`. Depends on T010,
  T017.

**Checkpoint**: `sonora inputs list` and `sonora inputs get <id>` both work end-to-end for
their happy paths — full MVP. `go build ./... && go test ./...` passes.

---

## Phase 5: User Story 3 - Include disabled inputs in the list (Priority: P2)

**Goal**: `sonora inputs list --include-disabled` shows both enabled and disabled inputs,
each with its `enabled` state visible.

**Independent Test**: Run the list command with `--include-disabled` against a hub with at
least one disabled input and verify it appears in the results, distinguishable from enabled
inputs.

### Tests for User Story 3 (write first; MUST fail before implementation exists)

- [X] T019 [P] [US3] Extend tests/contract/inputs_list_test.go with a test asserting that
  requesting `includeDisabled=true` sends that query value to the hub, and that a mixed
  enabled/disabled response body decodes correctly (this exercises `hub.ListInputs`'s
  existing `includeDisabled` parameter from T007 — no hub-layer code change expected).
- [X] T020 [P] [US3] Extend tests/integration/inputs_list_test.go with a test asserting
  `sonora inputs list --include-disabled` against a mock server with one enabled and one
  disabled input prints both, each showing its `enabled` state correctly.

### Implementation for User Story 3

- [X] T021 [US3] Add an `--include-disabled` bool flag to `RunList` in
  internal/cli/inputs/list.go and pass its value through to `hub.ListInputs` in place of the
  hardcoded `false` from T009 (mirroring `outputs.RunList`'s existing
  `--include-disabled` handling). Makes T019, T020 pass.

**Checkpoint**: `--include-disabled` works. `go build ./... && go test ./...` passes;
`inputs list`'s default (enabled-only) behavior from US1 is unaffected.

---

## Phase 6: User Story 4 - Handle a nonexistent input identifier (Priority: P2)

**Goal**: `sonora inputs get <missing-id>` reports a clear "input not found" message,
distinguishable from every other failure class, and exits with a distinct status.

**Independent Test**: Run the get command with an identifier that does not exist on the mock
hub and verify the user sees an unambiguous "not found" message and a distinct exit code
(not the generic hub-error or network-error code).

### Tests for User Story 4 (write first; MUST fail before implementation exists)

- [X] T022 [P] [US4] Extend tests/contract/inputs_get_test.go with a test asserting that a
  `404 ErrorResponse` from the mock server (per `getInput` in `api/openapi.json`) causes
  `hub.GetInput` to return a `*hub.NotFoundError` (not a generic `*hub.StatusError`).
- [X] T023 [P] [US4] Extend tests/integration/inputs_get_test.go with a test asserting
  `sonora inputs get <missing-id>` against a mock server that 404s prints a clear "input not
  found" message (naming the identifier) on stderr and exits with the not-found exit code
  (`5`), distinct from the exit codes used for hub errors (`3`) and network errors (`4`).
- [X] T024 [P] [US4] Extend tests/unit/hub_client_test.go asserting
  `hub.ClassifyError(&hub.NotFoundError{Resource: "input", ID: "x"})` returns
  `hub.ClassNotFound` and the message `"input not found: x"`, and that the existing
  `&hub.NotFoundError{Resource: "output", ID: "x"}` case still produces
  `"output not found: x"` (regression check on the generalized `Resource`/`ID` fields
  replacing the old `OutputID`-only shape).

### Implementation for User Story 4

- [X] T025 [US4] Generalize `NotFoundError` in internal/hub/errors.go from
  `{OutputID string}` to `{Resource, ID string}`, with `Error() string` returning
  `fmt.Sprintf("%s not found: %s", e.Resource, e.ID)`; update the `errors.As(err,
  &notFoundErr)` branch in `ClassifyError` to build its message from `notFoundErr.Resource`
  and `notFoundErr.ID` instead of the old `notFoundErr.OutputID`.
- [X] T026 [US4] Update `GetOutput` in internal/hub/outputs.go to construct
  `&NotFoundError{Resource: "output", ID: outputID}` in place of the old
  `&NotFoundError{OutputID: outputID}` (message text and exit code unchanged — existing
  `002-outputs-get` tests keep passing). Depends on T025.
- [X] T027 [US4] Update `GetInput` in internal/hub/inputs.go to return
  `&NotFoundError{Resource: "input", ID: inputID}` specifically when the response status is
  `404` (checked before the generic non-2xx `StatusError` fallback added in T015). Depends
  on T025. Makes T022, T023 pass.

**Checkpoint**: "not found" is a distinct, correctly-classified outcome for both `inputs`
and `outputs`. `go build ./... && go test ./...` passes; `outputs get`'s existing "not
found" behavior and exit code are unaffected (T024's regression assertion, T026).

---

## Phase 7: User Story 5 - Consume input data from a script (Priority: P3)

**Goal**: `sonora inputs list --json` and `sonora inputs get <id> --json` emit the same
fields as strict, parseable JSON.

**Independent Test**: Run each command with `--json` and verify the result parses with a
standard JSON parser and contains all eight fields per input.

### Tests for User Story 5 (write first; MUST fail before implementation exists)

- [X] T028 [US5] Extend tests/unit/render_inputs_test.go with tests asserting
  `render.RenderJSON([]hub.Input{...})` produces `{"inputs": [...]}` with all eight fields
  per input and round-trips through `encoding/json.Unmarshal` without error, and
  `render.RenderInputJSON(hub.Input{...})` produces a single JSON object (no list wrapper)
  that likewise round-trips.
- [X] T029 [P] [US5] Integration test in tests/integration/inputs_list_test.go asserting
  `sonora inputs list --json` exits `0` and its stdout parses via `encoding/json` into the
  documented `{"inputs": [...]}` shape with the same fields and values as the YAML view.
- [X] T030 [P] [US5] Integration test in tests/integration/inputs_get_test.go asserting
  `sonora inputs get <id> --json` exits `0` and its stdout parses via `encoding/json` into
  an object with the same eight fields and values as the YAML view.

### Implementation for User Story 5

- [X] T031 [P] [US5] Implement `RenderJSON(inputs []hub.Input) string` in
  internal/render/inputs.go: strict JSON `{"inputs": [...]}` (nil → `[]hub.Input{}`),
  mirroring `internal/render/outputs.go`'s `RenderJSON`. Makes T028's list-JSON assertion
  pass.
- [X] T032 [P] [US5] Implement `RenderInputJSON(i hub.Input) string` in
  internal/render/inputs.go: strict JSON of the single record (no list wrapper), mirroring
  `RenderOutputJSON`. Makes T028's single-JSON assertion pass.
- [X] T033 [US5] Add a `--json` bool flag to `RunList` in internal/cli/inputs/list.go and
  switch rendering to `render.RenderJSON` when set (mirroring `outputs.RunList`'s existing
  `--json` handling). Depends on T031. Makes T029 pass.
- [X] T034 [US5] Add a `--json` bool flag to `RunGet` in internal/cli/inputs/get.go and
  switch rendering to `render.RenderInputJSON` when set (mirroring `outputs.RunGet`'s
  existing `--json` handling). Depends on T032. Makes T030 pass.

**Checkpoint**: All five user stories are independently functional. `go build ./... && go
test ./...` passes.

---

## Phase 8: Polish & Cross-Cutting Concerns

**Purpose**: Final quality gates spanning all five stories.

- [X] T035 [P] Run `gofmt -l .`, `go vet ./...`, and the project's configured linter across
  all files touched by this feature (internal/hub/, internal/render/, internal/cli/inputs/,
  cmd/sonora/main.go, tests/); fix any findings (constitution Development Workflow).
- [X] T036 Execute every step in quickstart.md end-to-end (`go test ./...`, the manual
  success/not-found/failure-path smoke tests against a mock hub) and confirm each mapped
  Success Criterion (SC-001…SC-007) holds.

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — start immediately.
- **Foundational (Phase 2)**: Depends on Setup — BLOCKS all user stories (both P1 stories
  decode into the `Input` type T002 defines).
- **User Story 1 (Phase 3)**: Depends on Foundational. No dependency on US2–US5.
- **User Story 2 (Phase 4)**: Depends on Foundational. T018 (main.go `"get"` case) depends
  on T010 (US1's `"inputs"` case) existing in the same switch statement; otherwise
  independent of US1/US3/US4/US5.
- **User Story 3 (Phase 5)**: Depends on Foundational and on US1's `RunList`/`ListInputs`
  (T007, T009) existing to extend. Independent of US2/US4/US5.
- **User Story 4 (Phase 6)**: Depends on Foundational and on US2's `GetInput`/`RunGet`
  (T015, T017) existing to extend; also touches `GetOutput` (T026), so should land after
  `002-outputs-get`'s code exists (already true on this branch). Independent of US1/US3/US5.
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
- Within US4: T025 before T026 and T027 (both edit callers of the type T025 changes).
- Within US5: T031 before T033; T032 before T034 (each `RunX` needs its render function to
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
- US4 tests: T022, T023, T024 — three different files — can all run in parallel.
- US5: T029 and T030 (different files) can run in parallel; T031 and T032 (different
  functions, same file) can be authored in parallel and merged.
- T035 (lint/format) can run in parallel with T036 (quickstart validation).

---

## Parallel Example: User Story 1

```bash
# After Phase 2 (Foundational) completes, launch all US1 tests together:
Task: "Contract test for hub.ListInputs in tests/contract/inputs_list_test.go"
Task: "Integration test for list command default success path in tests/integration/inputs_list_test.go"
Task: "Unit test for RenderYAML in tests/unit/render_inputs_test.go"
Task: "Unit test for RunList flag parsing in tests/unit/cli_inputs_test.go"

# Once those tests fail as expected, launch the two independent implementation tasks:
Task: "Implement hub.ListInputs in internal/hub/inputs.go"
Task: "Implement render.RenderYAML for inputs in internal/render/inputs.go"
```

---

## Implementation Strategy

### MVP First (User Stories 1 and 2 only)

1. Complete Phase 1: Setup.
2. Complete Phase 2: Foundational (the `Input` struct — blocks everything else).
3. Complete Phase 3: User Story 1 (`inputs list` default).
4. Complete Phase 4: User Story 2 (`inputs get <id>`).
5. **STOP and VALIDATE**: run T003–T014's tests plus manual `sonora inputs list`/`sonora
   inputs get <id>` smoke tests against a mock hub (quickstart.md §2–§3).
6. This is a shippable MVP: operators can discover input identifiers and look up any single
   input's full state.

### Incremental Delivery

1. Setup + Foundational → foundation ready.
2. Add User Story 1 → validate independently.
3. Add User Story 2 → validate independently → MVP complete.
4. Add User Story 3 → validate `--include-disabled` independently → doesn't touch US1/US2
   default behavior.
5. Add User Story 4 → validate the not-found path independently (exit code `5`, distinct
   message) → doesn't touch US1/US2/US3 behavior; regression-checks `outputs get`'s
   unchanged not-found message.
6. Add User Story 5 → validate `--json` for both commands independently → commands now
   match the full contracts in [contracts/cli-inputs-list.md](contracts/cli-inputs-list.md)
   and [contracts/cli-inputs-get.md](contracts/cli-inputs-get.md).
7. Polish: lint/format pass, full quickstart.md run.

---

## Notes

- [P] tasks touch different files with no dependency on an incomplete task.
- [Story] labels trace every task back to spec.md's user stories.
- Tests are mandatory (Principle VI) — write and confirm each test fails before starting its
  paired implementation task.
- This feature deliberately touches very little of `001-list-outputs`/`002-outputs-get`'s
  existing code: the `NotFoundError` generalization in Phase 6 (T025–T026), plus additive
  functions/types everywhere else — `outputs list`/`outputs get`'s behavior, exit codes, and
  passing tests are expected to be unaffected end to end.
</content>
