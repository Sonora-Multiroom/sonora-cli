---

description: "Task list for List Audio Outputs feature"
---

# Tasks: List Audio Outputs

**Input**: Design documents from `/specs/001-list-outputs/`

**Prerequisites**: [plan.md](plan.md) (required), [spec.md](spec.md) (required for user stories),
[research.md](research.md), [data-model.md](data-model.md), [contracts/cli-outputs-list.md](contracts/cli-outputs-list.md),
[quickstart.md](quickstart.md)

**Tests**: Included and mandatory — Constitution Principle VI (Test-First Development) is
NON-NEGOTIABLE for this project: tests MUST be written before implementation, reviewed, seen
to fail, then implemented against (Red-Green-Refactor).

**Organization**: Tasks are grouped by user story (P1/P2/P3 from [spec.md](spec.md)) to enable
independent implementation and testing of each story. This is also the first feature in the
repo, so Phase 1/2 establish the Go module and the shared skeleton (HTTP client, config
resolution, error classification, rendering) every story and every future command builds on.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1, US2, US3)
- Exact file paths are included in every task description

## Path Conventions

Single Go project per [plan.md](plan.md)'s Project Structure section:

- `cmd/sonora/` — CLI entrypoint
- `internal/{hub,config,cli/outputs,render}/` — implementation packages
- `tests/{unit,contract,integration}/` — test layers

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Initialize the Go module — this is the first code in the repository.

- [X] T001 Run `go mod init` at repo root targeting Go 1.27.0 per [research.md §1](research.md), creating `go.mod`; add a minimal `.gitignore` entry for the built `sonora`/`sonora.exe` binary (repo already has a root `.gitignore` per git status — extend it, don't replace it)
- [X] T002 Create the package skeleton directories with placeholder `doc.go` files (or empty package files) for `cmd/sonora/`, `internal/hub/`, `internal/config/`, `internal/cli/outputs/`, `internal/render/`, `tests/unit/`, `tests/contract/`, `tests/integration/` per [plan.md](plan.md)'s Project Structure
- [X] T003 [P] Add a `Makefile` or documented `go vet ./...` / `gofmt -l .` check (Development Workflow in the constitution) — no linter config exists yet, so wire `gofmt` and `go vet` as the baseline; note `golangci-lint` as optional/future if not installed

**Checkpoint**: `go build ./...` succeeds (even with empty packages) before any test/implementation work begins.

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: The shared skeleton — HTTP client, hub URL resolution, error classification, and
output rendering — that every user story's implementation depends on. Per Principle VI, each
piece is test-first: write the failing unit test, then the minimum code to pass it.

**⚠️ CRITICAL**: No user story implementation can begin until this phase is complete.

- [X] T004 [P] Write failing unit tests for hub URL precedence (flag > `MULTIROOM_URL` env > `~/.config/sonora/config.json`'s `hubUrl` > built-in default `http://localhost:8080`) and malformed-config-file handling (exit-2 usage error naming the file) in `tests/unit/config_test.go`, per [research.md §5](research.md) and [data-model.md](data-model.md)'s Config file section
- [X] T005 Implement `internal/config/config.go`: `ResolveHubURL(flagVal string) (string, error)` reading `~/.config/sonora/config.json` lazily (only when needed, after flags are known — Principle I), applying the four-layer precedence, to make T004 pass
- [X] T006 [P] Write failing unit tests for the HTTP client in `tests/unit/hub_client_test.go`, per [research.md §4](research.md): (a) timeout enforcement — a `httptest.Server` handler that sleeps past 5s, asserting the client aborts the request at ~5s rather than hanging, rather than merely asserting `Timeout: 5 * time.Second` is set; (b) single-attempt/no-retry behavior — a request-counting handler that fails the first call, asserting the client makes exactly one request and does not retry; (c) the client is constructed only when the command handler runs, not at package `init()`
- [X] T007 Implement `internal/hub/client.go`: `NewClient() *http.Client` with the 5s timeout and default `http.Transport`, to make T006 pass
- [X] T008 [P] Write failing unit tests for error classification — `context.DeadlineExceeded`/`*net.OpError` → network error class, non-2xx status → hub error class, JSON decode/shape mismatch → hub error class, each mapped to the exit codes `0`/`2`/`3`/`4` from [research.md §6](research.md) — in `tests/unit/hub_client_test.go` (extend T006's file)
- [X] T009 Implement `internal/hub/errors.go`: error classification types/functions (`ClassifyError(err error) (class ErrorClass, friendlyMsg string)`) covering usage/hub/network classes, to make T008 pass, per [research.md §6-7](research.md)
- [X] T010 Define the `Output` domain struct (per [data-model.md](data-model.md)'s field table: `OutputID`, `DisplayName`, `Volume`, `Muted`, `Available`, `Enabled`) and implement `ListOutputs(ctx context.Context, client *http.Client, baseURL string, includeDisabled bool) ([]Output, error)` in `internal/hub/outputs.go` — builds `GET {baseURL}/api/v2/outputs?includeDisabled={true|false}`, decodes with `json.Decoder` (no `omitempty`/pointer fields, per [data-model.md](data-model.md)'s validation rule), checks non-empty `outputId`/`displayName` post-decode, returns a classified error via T009's classifier on any failure (depends on T007, T009). `Output` lives in `internal/hub`, not `internal/render`, since it is the service/domain layer's data — `internal/render` depends on it, not the other way around.
- [X] T011 [P] Write failing unit tests for YAML/JSON output rendering of `hub.Output` records (type defined in T010) — including the zero-outputs case (`outputs: []` / `{"outputs": []}` with an unambiguous "no outputs" note) and the unavailable-output distinguishability case (`available: false` always explicit, never omitted, FR-005) — in `tests/unit/render_outputs_test.go`, per [research.md §3](research.md) and [contracts/cli-outputs-list.md](contracts/cli-outputs-list.md)
- [X] T012 Implement `internal/render/outputs.go`: `RenderYAML(outputs []hub.Output) string`, `RenderJSON(outputs []hub.Output) string`, operating on the `hub.Output` type from T010, to make T011 pass (depends on T010, T011)

**Checkpoint**: Foundation ready — `internal/hub`, `internal/config`, `internal/render` all compile and pass their unit tests. User story implementation can now begin.

---

## Phase 3: User Story 1 - View currently active outputs (Priority: P1) 🎯 MVP

**Goal**: `sonora outputs list` with no flags calls `GET /api/v2/outputs`, shows only enabled
outputs (identifier, display name, volume, mute state, availability) as default YAML, and
handles the zero-outputs, unreachable-hub, timeout, non-2xx, and malformed-response cases per
FR-001/002/004-006/008-013 and SC-001/002/004/005.

**Independent Test**: Run the command with no flags against a mock hub with two enabled
outputs and one disabled output; verify only the two enabled outputs are printed with all
required fields, and disabled outputs are omitted. Separately verify the zero-outputs,
unreachable-hub, and malformed-response failure paths.

### Tests for User Story 1 (write first, confirm failing)

- [X] T013 [P] [US1] Write failing contract test in `tests/contract/outputs_list_test.go`: `httptest.Server` serving `GET /api/v2/outputs` shaped from `api/openapi.json`'s `OutputResponse` (2 enabled outputs), asserting the CLI's HTTP client sends the right request (no `includeDisabled` query value implying `false`, or `includeDisabled=false` explicitly per implementation choice) and decodes correctly; also serve a malformed body (wrong JSON type on `volume`) and assert the client rejects it (FR-013)
- [X] T014 [P] [US1] Write failing integration test in `tests/integration/outputs_list_test.go`: full `sonora outputs list` process/command invocation (via `os/exec` or direct `main`-equivalent function call) against an `httptest.Server` with 2 enabled + 1 disabled output, asserting stdout is YAML containing only the 2 enabled outputs with all six fields, and exit code `0`
- [X] T015 [P] [US1] Extend `tests/integration/outputs_list_test.go` with the zero-outputs case (mock server returns `[]`), asserting exit code `0` and an unambiguous "no outputs" indication on stdout (FR-012, SC-004)
- [X] T016 [P] [US1] Extend `tests/integration/outputs_list_test.go` with failure-path cases: unreachable hub URL (exit `4`, stderr mentions hub URL), non-2xx hub response (exit `3`, stderr distinguishes hub error from connectivity failure per FR-010), and malformed response body (exit `3`, FR-013) — all asserting stderr-only error output and stdout left empty
- [X] T017 [P] [US1] Extend `tests/integration/outputs_list_test.go` with an unknown-flag case, asserting exit code `2` and a usage message on stderr

### Implementation for User Story 1

- [X] T018 [US1] Implement `internal/cli/outputs/list.go`: `Run(args []string, stdout, stderr io.Writer) int` — defines the `flag.FlagSet` for `outputs list` (`--include-disabled`, `--json`, `--verbose`, `--hub-url`), resolves the hub URL via `internal/config` (T005), constructs the HTTP client via `internal/hub` (T007), calls `ListOutputs` (T010) with `includeDisabled=false` for this story, renders via `internal/render` (T012) defaulting to YAML, writes success payload to `stdout` and any error (friendly message, plus raw detail if `--verbose`) to `stderr`, and returns the exit code per [research.md §6](research.md)'s table (depends on T005, T007, T010, T012)
- [X] T019 [US1] Implement `cmd/sonora/main.go`: top-level dispatch parsing the first two positional args as `<noun> <verb>` (`outputs list` → `internal/cli/outputs.Run`), unrecognized noun/verb → exit `2`, and `os.Exit` with the code `Run` returns (depends on T018)
- [X] T020 [US1] Run `go test ./tests/contract/... ./tests/integration/... -run TestOutputsList` and fix implementation until T013-T017 all pass; confirm the zero-outputs and malformed-response cases are unambiguous per FR-012/FR-013

**Checkpoint**: `sonora outputs list` is fully functional and independently testable — enabled-only listing, zero-outputs handling, and all failure/exit-code paths work end-to-end. This is the shippable MVP.

---

## Phase 4: User Story 2 - Include disabled outputs (Priority: P2)

**Goal**: `sonora outputs list --include-disabled` additionally returns disabled outputs,
each visibly showing `enabled: false`, reusing the same read path (FR-003).

**Independent Test**: Run the command with `--include-disabled` against a mock hub with one
enabled and one disabled output; verify both appear and each output's enabled/disabled state
is visible in the result.

### Tests for User Story 2 (write first, confirm failing)

- [X] T021 [P] [US2] Write failing contract test in `tests/contract/outputs_list_test.go`: assert that when the CLI is invoked with `--include-disabled`, the request sent to the mock `httptest.Server` includes `includeDisabled=true` as a query parameter
- [X] T022 [P] [US2] Extend `tests/integration/outputs_list_test.go` with a `--include-disabled` case against a mock hub returning one enabled + one disabled output, asserting both appear in stdout, each with its `enabled` field visible and correct (true/false)

### Implementation for User Story 2

- [X] T023 [US2] Wire `--include-disabled` in `internal/cli/outputs/list.go` (T018) through to `ListOutputs`'s `includeDisabled` parameter (T010 already accepts it — this is a flag-plumbing change, not new client logic)
- [X] T024 [US2] Run `go test ./tests/contract/... ./tests/integration/... -run TestOutputsList` and fix implementation until T021-T022 pass, confirming User Story 1's default (enabled-only) behavior is unchanged

**Checkpoint**: User Stories 1 AND 2 both work independently — default view stays enabled-only; `--include-disabled` surfaces both states correctly.

---

## Phase 5: User Story 3 - Consume output list from a script (Priority: P3)

**Goal**: `sonora outputs list --json` emits strict, parseable JSON with the same fields as
the human-readable view (FR-007, SC-003).

**Independent Test**: Run the command with `--json` against a mock hub with at least one
enabled output; verify the result is valid JSON (parseable by a standard JSON parser) and
contains identifier, display name, volume, mute state, availability, and enabled state.

### Tests for User Story 3 (write first, confirm failing)

- [X] T025 [P] [US3] Extend `tests/integration/outputs_list_test.go` with a `--json` case: invoke `sonora outputs list --json` against a mock hub with 1+ enabled outputs, parse stdout with `encoding/json` in the test itself, and assert it unmarshals cleanly into `{"outputs": [...]}` with all six fields per output present and correctly typed
- [X] T026 [P] [US3] Extend `tests/integration/outputs_list_test.go` with the `--json` zero-outputs case, asserting stdout is exactly `{"outputs": []}` (or equivalent strictly-parseable empty form) and exit code `0`

### Implementation for User Story 3

- [X] T027 [US3] Wire `--json` in `internal/cli/outputs/list.go` (T018) to select `render.RenderJSON` (T012) instead of `RenderYAML` — `internal/render` already implements both; this is a dispatch change only
- [X] T028 [US3] Run `go test ./tests/integration/... -run TestOutputsList` and fix implementation until T025-T026 pass, confirming User Stories 1 and 2's YAML default output is unchanged when `--json` is absent

**Checkpoint**: All three user stories are independently functional — enabled-only YAML by default, `--include-disabled` for administration, `--json` for automation, all combinable per the contract's 8-combination table.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Verification and cleanup that spans all three stories.

- [X] T029 [P] Run `gofmt -l .` and `go vet ./...` across the repo and fix any findings (Development Workflow)
- [X] T030 [P] Write and confirm failing-then-passing unit tests for `--verbose` behavior (raw error detail appended on failure, absent by default) in `tests/unit/hub_client_test.go` or a dedicated `tests/unit/cli_outputs_test.go`, then wire `--verbose` through `internal/cli/outputs/list.go` if not already covered by T018's implementation
- [X] T031 Execute [quickstart.md](quickstart.md) end-to-end: `go test ./...` (step 1), then the manual smoke tests (steps 2-4) against a locally started mock hub, confirming SC-001 through SC-005 all hold
- [X] T032 [P] Verify all 8 flag combinations from [contracts/cli-outputs-list.md](contracts/cli-outputs-list.md)'s example table behave as documented (can be folded into existing integration tests as targeted subtests rather than new files)
- [X] T033 Self-review the full diff against Constitution Principles I, III, IV, and VI (Development Workflow requirement for any PR touching startup path, dependency list, or HTTP client construction) before requesting merge

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — start immediately.
- **Foundational (Phase 2)**: Depends on Setup completion — BLOCKS all user stories. Internally: T004→T005, T006→T007, T008→T009; T010 (Output struct + ListOutputs) depends on T007+T009; T011→T012 (render), with T012 depending on T010's `Output` type as well as T011.
- **User Story 1 (Phase 3)**: Depends on Foundational (Phase 2) completion. Tests (T013-T017) before implementation (T018-T020).
- **User Story 2 (Phase 4)**: Depends on Foundational (Phase 2) and, practically, on User Story 1's `list.go`/`main.go` existing (T018-T019) since it extends the same flag set — implemented after US1 rather than in parallel, though its own tests (T021-T022) could be drafted earlier.
- **User Story 3 (Phase 5)**: Same as User Story 2 — depends on Foundational and extends `list.go` from US1.
- **Polish (Phase 6)**: Depends on all three user stories being complete.

### User Story Dependencies

- **User Story 1 (P1)**: No dependencies on other stories — this is the MVP.
- **User Story 2 (P2)**: Reuses US1's `internal/cli/outputs/list.go` and `internal/hub.ListOutputs` (already parameterized by `includeDisabled` from Phase 2) — additive flag-plumbing only, does not modify US1 behavior.
- **User Story 3 (P3)**: Reuses US1's `list.go` and Phase 2's `render.RenderJSON` — additive flag-plumbing only, does not modify US1/US2 behavior.

### Within Each User Story

- Tests MUST be written and confirmed failing before implementation (Principle VI, NON-NEGOTIABLE).
- Foundational client/render/config code (Phase 2) before any story's CLI wiring.
- Story complete (its own tests passing, prior stories' tests still passing) before moving to the next priority.

### Parallel Opportunities

- T004, T006, T008, T011 (all foundational unit test-writing tasks, different files/sections) can be written in parallel.
- T013-T017 (US1 test-writing, all in the same two files but independent test functions) can be drafted in parallel by different people, though they land in shared files — coordinate merge order.
- T021-T022 and T025-T026 can similarly be parallel-drafted once Phase 2 is done, even though implementation is sequenced after US1.
- T029 and T032 in Polish can run in parallel with each other.

---

## Parallel Example: Foundational Phase

```bash
# Launch all foundational unit-test-writing tasks together:
Task: "Write failing unit tests for hub URL precedence in tests/unit/config_test.go"
Task: "Write failing unit tests for the HTTP client timeout/single-attempt behavior in tests/unit/hub_client_test.go"
Task: "Write failing unit tests for error classification in tests/unit/hub_client_test.go"
Task: "Write failing unit tests for YAML/JSON rendering in tests/unit/render_outputs_test.go"
```

## Parallel Example: User Story 1

```bash
# Launch all US1 test-writing tasks together (coordinate on shared files):
Task: "Contract test for GET /api/v2/outputs in tests/contract/outputs_list_test.go"
Task: "Integration test for default enabled-only listing in tests/integration/outputs_list_test.go"
Task: "Integration test for zero-outputs case in tests/integration/outputs_list_test.go"
Task: "Integration test for failure paths (unreachable/non-2xx/malformed) in tests/integration/outputs_list_test.go"
Task: "Integration test for unknown-flag usage error in tests/integration/outputs_list_test.go"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (`go mod init`, package skeleton).
2. Complete Phase 2: Foundational — HTTP client, config resolution, error classification, rendering (all test-first).
3. Complete Phase 3: User Story 1 — default enabled-only listing, YAML output, full error handling.
4. **STOP and VALIDATE**: run `go test ./...` and the quickstart's manual smoke tests for User Story 1's scope.
5. Ship — `sonora outputs list` is a complete, useful command on its own.

### Incremental Delivery

1. Setup + Foundational → shared skeleton ready (reused by every future command, not just this one).
2. Add User Story 1 → test independently → MVP shippable.
3. Add User Story 2 (`--include-disabled`) → test independently → administrative/troubleshooting use case unlocked.
4. Add User Story 3 (`--json`) → test independently → automation/scripting use case unlocked.
5. Polish → format/vet, quickstart validation, constitution self-review.

### Single-Developer Strategy

Given this is a small, single-command feature building the first skeleton of the repo, stories
are best done sequentially (P1 → P2 → P3) rather than parallelized across people, since US2 and
US3 both extend the same two files (`list.go`'s flag set) that US1 creates. The parallel
opportunities above apply to *test-writing within* a phase, not to running whole stories
concurrently.

## Notes

- [P] tasks target different files, or independent test functions appended to a shared file with no ordering dependency between them.
- [Story] labels map tasks to spec.md's user stories (US1/US2/US3) for traceability; Setup/Foundational/Polish carry no story label.
- Per Constitution Principle VI, every implementation task in Phases 2-5 has a preceding test-writing task that MUST be confirmed failing first — do not skip ahead to implementation tasks.
- Commit after each task or logical group.
- Stop at each phase checkpoint to validate before proceeding to the next.
