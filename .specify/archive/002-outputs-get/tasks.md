---

description: "Task list for `sonora outputs get`"
---

# Tasks: Get Single Audio Output

**Input**: Design documents from `/specs/002-outputs-get/`

**Prerequisites**: [plan.md](plan.md), [spec.md](spec.md), [research.md](research.md),
[data-model.md](data-model.md), [contracts/cli-outputs-get.md](contracts/cli-outputs-get.md),
[quickstart.md](quickstart.md)

**Tests**: Tests are **mandatory**, not optional, for this project — constitution Principle
VI ("Test-First Development (NON-NEGOTIABLE)") requires every command and API client method
to have a test written, reviewed, and watched to fail *before* the implementing code is
written. Every implementation task below has a corresponding test task that MUST land first.

**Organization**: Tasks are grouped by user story (from spec.md), in priority order
(P1 → P2 → P3), so each story is independently implementable and testable.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies on incomplete tasks)
- **[Story]**: Which user story this task belongs to (US1, US2, US3)
- File paths are exact and relative to the repo root

## Path Conventions

Single Go project (established by `001-list-outputs`, unchanged): `cmd/sonora/`,
`internal/{hub,render,cli/outputs,config}/`, `tests/{unit,contract,integration}/`.

---

## Phase 1: Setup

**Purpose**: Confirm the existing `001-list-outputs` skeleton this feature builds on is in a
known-good state before touching it.

- [X] T001 Run `go build ./...` and `go test ./...` from the repo root on branch
  `002-outputs-get` and confirm everything passes, establishing the baseline this feature's
  changes are layered on (no files modified by this task).

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: `internal/cli/outputs.Run` must become `outputs.RunList` before a second verb
(`get`) can be added to the same package without an ambiguous exported name (research.md
§3). This is a pure rename with no behavior change, isolated from the `get` command itself,
so it lands first and independently.

**⚠️ CRITICAL**: No User Story work can begin until this phase is complete — `outputs.RunGet`
(added in US1) needs the package's exported-name collision resolved first, and `main.go`'s
dispatch table is edited by both this phase and Phase 3.

- [X] T002 Rename the exported `Run` function to `RunList` in
  internal/cli/outputs/list.go (signature and behavior unchanged).
- [X] T003 [P] Update the `outputs.Run(...)` call sites to `outputs.RunList(...)` in
  tests/unit/cli_outputs_test.go (rename only, no assertion changes).
- [X] T004 [P] Update the `outputs.Run(...)` call site(s) to `outputs.RunList(...)` in
  tests/integration/outputs_list_test.go (rename only, no assertion changes). No call site
  existed — that file invokes the built binary, not `outputs.Run` directly.
- [X] T005 Update the `"list"` case in the verb switch in cmd/sonora/main.go to call
  `outputs.RunList` instead of `outputs.Run`.

**Checkpoint**: `go build ./... && go test ./...` passes with `outputs list` fully
unchanged in behavior, now exposed as `RunList`. User story implementation can begin.

---

## Phase 3: User Story 1 - Look up a specific output by identifier (Priority: P1) 🎯 MVP

**Goal**: `sonora outputs get <output-id>` fetches and displays one output's full state
(identifier, display name, volume, mute, availability, enabled) in YAML by default,
regardless of its enabled/disabled state.

**Independent Test**: Run the get command with the identifier of a known output (try both
an enabled and a disabled one) and verify all six fields are displayed correctly for both.

### Tests for User Story 1 (write first; MUST fail before implementation exists)

- [X] T006 [P] [US1] Contract test for `hub.GetOutput`'s success and hub-error paths in
  tests/contract/outputs_get_test.go: an `httptest.Server` serving `GET
  /api/v2/outputs/{outputId}` per `getOutput` in `api/openapi.json` — assert the request
  path includes the given ID, a `200 OutputResponse` body decodes into `hub.Output`
  correctly, a non-2xx/non-404 (e.g. `500`) response yields a `*hub.StatusError`, and a
  malformed `200` body (wrong field type / missing `outputId`) yields a `*hub.DecodeError`.
- [X] T007 [P] [US1] Integration test for the get command's success path in
  tests/integration/outputs_get_test.go: build/run the `sonora` binary (following the
  `binPath`/`TestMain` pattern already in tests/integration/outputs_list_test.go) against a
  mock server, and assert `sonora outputs get <id>` prints all six fields correctly in YAML
  for both an enabled output and a disabled one (FR-003 — no filtering, unlike `list`), with
  exit code `0`. Additionally wrap the invocation in `time.Now()`/`time.Since()` and assert
  it completes in under 1 second against the local mock server (SC-001).
- [X] T008 [P] [US1] Unit test for single-output YAML rendering in
  tests/unit/render_output_get_test.go: `render.RenderOutputYAML(hub.Output{...})` emits all
  six fields as a bare (non-list) record, and an output with `Available: false` shows
  `available: false` explicitly (never omitted) per FR-005.
- [X] T009 [P] [US1] Unit test for `outputs.RunGet`'s argument handling and failure paths in
  tests/unit/cli_outputs_get_test.go: missing `<output-id>` → exit `2`; more than one
  positional argument → exit `2`; unreachable `--hub-url` → exit `4`; and (mirroring the
  existing `TestOutputsRun_VerboseAppendsRawErrorDetail`/`...NonVerboseOmits...` tests for
  `list`) `--verbose` appends raw error detail on failure while its absence omits it. For the
  unreachable-`--hub-url` case, wrap the call in `time.Now()`/`time.Since()` and assert it
  returns in well under 5 seconds — proving the failure is fast/bounded rather than an
  indefinite hang (SC-003); this is a black-box check on top of, not a replacement for,
  `hub.NewClient()`'s own timeout tests in tests/unit/hub_client_test.go.

### Implementation for User Story 1

- [X] T010 [P] [US1] Implement `GetOutput(ctx context.Context, client *http.Client, baseURL,
  outputID string) (*Output, error)` in internal/hub/outputs.go: builds
  `{baseURL}/api/v2/outputs/{outputId}` (path-escaped), issues one GET, decodes a `200` body
  into `Output` (reusing the existing missing-`outputId`/`displayName` validation → wrapped
  as `*DecodeError`), and returns `*StatusError{StatusCode}` for any non-2xx response other
  than `404` (404 handling itself is added in T020, US2 — for now treat 404 as any other
  `*StatusError`, which US1's own tests don't exercise). Makes T006 pass.
- [X] T011 [P] [US1] Implement `RenderOutputYAML(o hub.Output) string` in
  internal/render/outputs.go: emits the same six `outputId`/`displayName`/`volume`/`muted`/
  `available`/`enabled` fields as `RenderYAML`'s per-record shape, but as a single top-level
  record (no `outputs:` wrapper) with every field always explicit. Makes T008 pass.
- [X] T012 [US1] Implement `RunGet(args []string, stdout, stderr io.Writer) int` in
  internal/cli/outputs/get.go: `flag.NewFlagSet("outputs get", flag.ContinueOnError)` with
  `--verbose` and `--hub-url` (mirroring `RunList`'s flag definitions), one required
  positional `<output-id>` (`fs.Arg(0)`; zero or >1 positional args → usage error, exit `2`,
  matching `RunList`'s "unexpected argument(s)" message shape), resolves the hub URL via the
  existing `config.ResolveHubURL`, calls `hub.NewClient()` + `hub.GetOutput`, classifies any
  error via `hub.ClassifyError` (exit code from `class.ExitCode()`), and on success prints
  `render.RenderOutputYAML(*output)` to stdout and returns `0`. Depends on T010, T011. Makes
  T007, T009 pass (for their non-not-found, non-JSON assertions).
- [X] T013 [US1] Add the `"get"` case to the verb switch in cmd/sonora/main.go, routing
  `outputs get` to `outputs.RunGet(rest, stdout, stderr)`. Depends on T012.

**Checkpoint**: `sonora outputs get <id>` works end-to-end for the happy path — MVP
deliverable. `go build ./... && go test ./...` passes.

---

## Phase 4: User Story 2 - Handle a nonexistent output identifier (Priority: P2)

**Goal**: `sonora outputs get <missing-id>` reports a clear "output not found" message,
distinguishable from every other failure class, and exits with a distinct status.

**Independent Test**: Run the get command with an identifier that does not exist on the mock
hub and verify the user sees an unambiguous "not found" message and a distinct exit code
(not the generic hub-error or network-error code).

### Tests for User Story 2 (write first; MUST fail before implementation exists)

- [X] T014 [P] [US2] Extend tests/contract/outputs_get_test.go with a test asserting that a
  `404 ErrorResponse` from the mock server (per `getOutput` in `api/openapi.json`) causes
  `hub.GetOutput` to return a `*hub.NotFoundError` (not a generic `*hub.StatusError`).
- [X] T015 [P] [US2] Extend tests/integration/outputs_get_test.go with a test asserting
  `sonora outputs get <missing-id>` against a mock server that 404s prints a clear "output
  not found" message (naming the identifier) on stderr and exits with the not-found exit
  code, and that this code differs from the exit codes used for hub errors (`3`) and network
  errors (`4`).
- [X] T016 [P] [US2] Unit test in tests/unit/hub_client_test.go asserting
  `hub.ClassifyError(&hub.NotFoundError{OutputID: "x"})` returns `hub.ClassNotFound` and a
  friendly message that names the identifier, and that `hub.ClassNotFound.ExitCode()` is
  distinct from `ClassUsage`, `ClassHub`, and `ClassNetwork`'s exit codes.

### Implementation for User Story 2

- [X] T017 [US2] Add `NotFoundError{OutputID string}` (with an `Error() string` message) in
  internal/hub/errors.go, distinct from `StatusError`.
- [X] T018 [US2] Add `ClassNotFound` to the `ErrorClass` enum in internal/hub/errors.go, with
  `ExitCode()` returning `5` (additive — `ClassUsage`=2, `ClassHub`=3, `ClassNetwork`=4 stay
  unchanged, per research.md §2). Depends on T017.
- [X] T019 [US2] Add a branch to `ClassifyError` in internal/hub/errors.go: `errors.As(err,
  &notFoundErr)` → `(ClassNotFound, "output not found: "+notFoundErr.OutputID)`, checked
  before the existing `StatusError` branch. Depends on T017, T018. Makes T016 pass.
- [X] T020 [US2] Update `GetOutput` in internal/hub/outputs.go to return
  `&NotFoundError{OutputID: outputID}` specifically when the response status is `404`
  (checked before the generic non-2xx `StatusError` fallback added in T010). Depends on
  T017. Makes T014, T015 pass.

**Checkpoint**: "not found" is now a distinct, correctly-classified outcome. `go build ./...
&& go test ./...` passes; `outputs list`'s exit codes (`0/2/3/4`) are unaffected.

---

## Phase 5: User Story 3 - Consume a single output's state from a script (Priority: P3)

**Goal**: `sonora outputs get <id> --json` emits the same fields as strict, parseable JSON.

**Independent Test**: Run the get command with `--json` against a known identifier and
verify the result parses with a standard JSON parser and contains all six fields.

### Tests for User Story 3 (write first; MUST fail before implementation exists)

- [X] T021 [P] [US3] Unit test in tests/unit/render_output_get_test.go asserting
  `render.RenderOutputJSON(hub.Output{...})` produces a single JSON object (not wrapped in an
  `outputs` array) containing all six fields, and that the output round-trips through
  `encoding/json.Unmarshal` without error.
- [X] T022 [P] [US3] Integration test in tests/integration/outputs_get_test.go asserting
  `sonora outputs get <id> --json` exits `0`, and its stdout parses via `encoding/json` into
  an object with the same six fields and values as the YAML view from T007.

### Implementation for User Story 3

- [X] T023 [US3] Implement `RenderOutputJSON(o hub.Output) string` in
  internal/render/outputs.go: strict JSON of the single record (no list wrapper), mirroring
  `RenderJSON`'s marshal approach. Makes T021 pass.
- [X] T024 [US3] Add a `--json` bool flag to `RunGet` in internal/cli/outputs/get.go and
  switch rendering to `render.RenderOutputJSON` when set (mirroring `RunList`'s existing
  `--json` handling). Depends on T023. Makes T022 pass.

**Checkpoint**: All three user stories are independently functional. `go build ./... && go
test ./...` passes.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Final quality gates spanning all three stories.

- [X] T025 [P] Run `gofmt -l .`, `go vet ./...`, and the project's configured linter across
  all files touched by this feature (internal/hub/, internal/render/, internal/cli/outputs/,
  cmd/sonora/main.go, tests/); fix any findings (constitution Development Workflow). No
  linter is configured in this repo beyond gofmt/go vet; one gofmt finding (the new
  outputs_get_test.go) was fixed.
- [X] T026 Execute every step in quickstart.md end-to-end (`go test ./...`, the manual
  success/not-found/failure-path smoke tests against a mock hub) and confirm each mapped
  Success Criterion (SC-001…SC-005) holds. Verified: `go test ./...` passes; manual runs
  against a live mock hub on 127.0.0.1:9090 confirmed YAML success (enabled + disabled
  outputs), `--json` output, not-found (exit 5), unreachable-host (exit 4, fast), the
  ~5s-bounded timeout against an unroutable host (exit 4 at ~5.0s), missing-identifier usage
  error (exit 2), and unknown-flag usage error (exit 2).

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — start immediately.
- **Foundational (Phase 2)**: Depends on Setup — BLOCKS all user stories (T005 touches the
  same `main.go` dispatch switch that US1's T013 extends next).
- **User Story 1 (Phase 3)**: Depends on Foundational. No dependency on US2/US3.
- **User Story 2 (Phase 4)**: Depends on Foundational. Builds on US1's `GetOutput`/`RunGet`
  (T010, T012) by adding the 404 branch — implement US1 first in practice, though US2's own
  tests (T014-T016) can be written in parallel with US1.
- **User Story 3 (Phase 5)**: Depends on Foundational and on US1's `RunGet` (T012) existing
  to add `--json` to. Independent of US2.
- **Polish (Phase 6)**: Depends on all three user stories being complete.

### Within Each User Story

- Tests (T006-T009, T014-T016, T021-T022) are written first and MUST fail before their
  corresponding implementation task starts (Principle VI).
- Within US1: T010 and T011 (different files, no shared dependency) before T012 (uses both);
  T012 before T013 (main.go needs `RunGet` to exist).
- Within US2: T017 before T018 before T019 (same file, `internal/hub/errors.go`, additive
  edits); T020 depends only on T017 (the `NotFoundError` type).
- Within US3: T023 before T024 (`RunGet` needs `RenderOutputJSON` to call).

### Parallel Opportunities

- Foundational: T003 and T004 (different test files) can run in parallel, after T002.
- US1 tests: T006, T007, T008, T009 — four different files, no shared dependency — can all
  run in parallel.
- US1 implementation: T010 and T011 — different files, no shared dependency — can run in
  parallel; both must land before T012.
- US2 tests: T014, T015, T016 — three different files — can all run in parallel.
- US3 tests: T021 and T022 — different files — can run in parallel.
- T025 (lint/format) can run in parallel with T026 (quickstart validation) since neither
  edits the other's inputs, though both should follow all implementation tasks.

---

## Parallel Example: User Story 1

```bash
# After Phase 2 (Foundational) completes, launch all US1 tests together:
Task: "Contract test for hub.GetOutput in tests/contract/outputs_get_test.go"
Task: "Integration test for get command success path in tests/integration/outputs_get_test.go"
Task: "Unit test for RenderOutputYAML in tests/unit/render_output_get_test.go"
Task: "Unit test for RunGet argument handling in tests/unit/cli_outputs_get_test.go"

# Once those tests fail as expected, launch the two independent implementation tasks:
Task: "Implement hub.GetOutput in internal/hub/outputs.go"
Task: "Implement render.RenderOutputYAML in internal/render/outputs.go"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup.
2. Complete Phase 2: Foundational (the `RunList` rename — blocks everything else).
3. Complete Phase 3: User Story 1.
4. **STOP and VALIDATE**: run T006-T009's tests plus a manual `sonora outputs get <id>`
   smoke test against a mock hub (quickstart.md §2).
5. This is a shippable MVP: operators can already look up any single output's state.

### Incremental Delivery

1. Setup + Foundational → foundation ready.
2. Add User Story 1 → validate independently → MVP.
3. Add User Story 2 → validate the not-found path independently (exit code `5`, distinct
   message) → still doesn't touch US1's behavior.
4. Add User Story 3 → validate `--json` independently → command now matches the full
   contract in [contracts/cli-outputs-get.md](contracts/cli-outputs-get.md).
5. Polish: lint/format pass, full quickstart.md run.

---

## Notes

- [P] tasks touch different files with no dependency on an incomplete task.
- [Story] labels trace every task back to spec.md's user stories.
- Tests are mandatory (Principle VI) — write and confirm each test fails before starting its
  paired implementation task.
- This feature deliberately touches very little of `001-list-outputs`'s existing code: the
  one rename in Phase 2, plus additive functions/types everywhere else — `outputs list`'s
  behavior, exit codes, and passing tests are expected to be unaffected end to end.
