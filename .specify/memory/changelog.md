# Changelog

## Merged Features Log

### List and Get Audio Routes — archived 2026-11-23
**Branch:** 004-routes-list-get
**Spec:** [.specify/archive/004-routes-list-get/spec.md](../archive/004-routes-list-get/spec.md)

**What was added:**
- User Stories 12–16 (P1–P3): `sonora routes list` shows all routes (identifier, source input, target, target type, status), `--status`/`--input-id`/`--target-id` narrow it with AND logic, `--json` for scripts
- `sonora routes get <route-id>` shows one route's full state, with a clear "route not found" exit `5`
- FR-028–FR-036, SC-012–SC-016, AS-008 and the Route entity; shared outputs/inputs/routes requirements widened (FR-006–FR-011, FR-013, SC-002–SC-004, AS-001–AS-006)

**New Components:**
- `internal/cli/routes/` (`RunList`, `RunGet`); `internal/hub/routes.go` (`ListRoutes`, `GetRoute`); `internal/render/routes.go`
- `tests/{unit,contract,integration}` routes tests

**Tasks Completed:** 33/33 tasks


### List and Get Audio Inputs — archived 2026-11-23
**Branch:** 003-inputs-list-get
**Spec:** [.specify/archive/003-inputs-list-get/spec.md](../archive/003-inputs-list-get/spec.md)

**What was added:**
- User Stories 7–11 (P1–P3): `sonora inputs list` shows enabled inputs, `--include-disabled` shows disabled ones, `--json` for scripts
- `sonora inputs get <input-id>` shows one input, enabled or disabled, with a clear "input not found" exit `5`
- FR-018–FR-027, SC-008–SC-011, AS-007 and the Input entity; shared outputs/inputs requirements widened (FR-006–FR-011, FR-013, SC-002–SC-004)

**New Components:**
- `internal/cli/inputs/` (`RunList`, `RunGet`); `internal/hub/inputs.go` (`ListInputs`, `GetInput`); `internal/render/inputs.go`
- `hub.NotFoundError` generalized to `Resource`/`ID` (output message text unchanged)
- `tests/{unit,contract,integration}` inputs tests

**Tasks Completed:** 36/36 tasks


### Get Single Audio Output — archived 2026-11-23
**Branch:** 002-outputs-get
**Spec:** [.specify/archive/002-outputs-get/spec.md](../archive/002-outputs-get/spec.md)

**What was added:**
- User Story 4 (P1): `sonora outputs get <output-id>` shows one output, enabled or disabled
- User Story 5 (P2): clear "output not found" message with its own exit code (`5`)
- User Story 6 (P3): `--json` strict machine-readable output for a single output

**New Components:**
- `internal/cli/outputs/get.go` (`RunGet`); `outputs.Run` renamed `RunList`
- `hub.GetOutput`, `NotFoundError` / `ClassNotFound`, single-record `RenderOutputYAML` / `RenderOutputJSON`
- `tests/{unit,contract,integration}` get tests

**Tasks Completed:** 26/26 tasks

### List Audio Outputs — archived 2026-11-23
**Branch:** 001-list-outputs
**Spec:** [.specify/archive/001-list-outputs/spec.md](../archive/001-list-outputs/spec.md)

**What was added:**
- User Story 1 (P1): list enabled outputs with identifier, display name, volume, mute and availability
- User Story 2 (P2): `--include-disabled` shows disabled outputs
- User Story 3 (P3): `--json` strict machine-readable output
- Fail-fast error handling with distinguishable exit codes (0/2/3/4), 5s timeout, no retries

**New Components:**
- `cmd/sonora`, `internal/hub`, `internal/config`, `internal/cli/outputs`, `internal/render`
- `tests/{unit,contract,integration}`

**Tasks Completed:** 33/33 tasks
