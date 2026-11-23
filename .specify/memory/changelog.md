# Changelog

## Merged Features Log

### Adopt Verb-First Command Landscape — archived 2026-11-23
**Branch:** 007-refactor-cli-commands
**Spec:** [.specify/archive/007-refactor-cli-commands/spec.md](../archive/007-refactor-cli-commands/spec.md)

**What was added:**
- User Stories 27–29 (P1–P3): `sonora get <resource>[/<id>]` and `sonora list <resource>` replace the `sonora <resource> list|get` commands for inputs, outputs, groups and routes; aliases `in`/`out`/`gr`/`rt`; `play`'s target is now `outputs/<id>` or `groups/<id>` and its name flag is `--display-name`
- Hard cutover: the old command forms and `play`'s `--group`/`--output`/`--name` fail with a standard usage error
- FR-059–FR-070, SC-026–SC-029, AS-014–AS-017, six edge cases and the Resource Kind and Resource Path entities
- Exit code `7` (ambiguous target) is gone, since a path prefix makes the target type unambiguous

**New Components:**
- `internal/cli/respath/` (alias table, `resource/id` split, id pattern); `cmd/sonora/main.go` verb-first dispatcher (`get`, `list`, `play`); `internal/cli/play` takes a path target and `--display-name`
- `internal/hub/play.go`: `ResolveTarget` replaced by a per-type existence check; `internal/hub/errors.go`: `ClassAmbiguous` and `AmbiguousTargetError` removed
- `internal/cli/clihelp/usage.go`: help text for the new grammar; `tests/{unit,contract,integration}` updated for the new invocation shape

**Superseded:**
- RETIRED: FR-015 (from .specify/archive/002-outputs-get/spec.md -> FR-002) → replaced by FR-060. Reason: `get outputs` with no id is now the collection form, not a usage error.
- RETIRED: FR-023 (from .specify/archive/003-inputs-list-get/spec.md -> FR-006) → replaced by FR-060. Reason: `get inputs` with no id is now the collection form, not a usage error.
- RETIRED: FR-033 (from .specify/archive/004-routes-list-get/spec.md -> FR-006) → replaced by FR-060. Reason: `get routes` with no id is now the collection form, not a usage error.
- RETIRED: FR-042 (from .specify/archive/005-groups-list-get/spec.md -> FR-006) → replaced by FR-060. Reason: `get groups` with no id is now the collection form, not a usage error.
- RETIRED: FR-049 (from .specify/archive/006-play-command/spec.md -> FR-002a) → no replacement. Reason: the `--group` and `--output` flags no longer exist.
- RETIRED: FR-050 (from .specify/archive/006-play-command/spec.md -> FR-003) → replaced by FR-067. Reason: the target type comes from the path prefix, not from auto-detection or forcing flags.
- RETIRED: FR-051 (from .specify/archive/006-play-command/spec.md -> FR-003a) → no replacement. Reason: path-style targets make an ambiguous output/group identifier impossible.
- RETIRED: FR-052 (from .specify/archive/006-play-command/spec.md -> FR-003b) → replaced by FR-067. Reason: the path prefix fixes the type to look up; a missing id of that type is plain "not found".
- RETIRED: AS-012 (from .specify/archive/006-play-command/spec.md -> "Output identifiers and group identifiers may occasionally collide") → no replacement. Reason: a path prefix breaks any output/group identifier tie, so no flags are needed.

**Tasks Completed:** 32/32 tasks

### Help Command (`help`) — archived 2026-11-23
**Tinyspec:** .specify/archive/tiny/2026-08-26-help-command/help-command.md
**Implemented:** 2026-08-26

Adds `sonora help`, `-h`/`--help` and bare `sonora`, which print the command table, global flags and examples to stdout with exit `0` (User Story 26, FR-057–FR-058; `cmd/sonora/main.go`, `cmd/sonora/main_test.go`). The missing-verb and unknown-command usage errors are unchanged.

### Instant Playback (`play`) — archived 2026-11-23
**Branch:** 006-play-command
**Spec:** [.specify/archive/006-play-command/spec.md](../../.specify/archive/006-play-command/spec.md)

**What was added:**
- User Stories 22–25 (P1–P3): `sonora play <uri> <target-id>` starts playback of a URI to one output or a group in one command and reports the created input identifier, route identifier and status as the hub returned them; `--volume` sets a starting volume (0–100), `--name` names the ephemeral input
- `--group`/`--output` force the target type when an output and a group share an identifier; both together is a usage error
- FR-047–FR-056, SC-022–SC-025, AS-010–AS-013 and the Playback Request and Playback Result entities; shared requirements, outcomes and assumptions now cover the play command as well (FR-006–FR-009, FR-011, FR-013, SC-003, SC-004, AS-001–AS-004)
- Five new exit codes (6–10) for validation, ambiguous target, route creation failed, source unreachable and upstream service unavailable; "target not found" reuses exit `5`

**New Components:**
- `internal/cli/play/` (`Run`); `internal/hub/play.go` (`PlaybackRequest`, `PlaybackResponse`, `ResolveTarget`, `Playback`); `internal/render/play.go`
- `internal/hub/errors.go`: five new classes, `APIError`, `AmbiguousTargetError`; `cmd/sonora/main.go`: early `play` special case
- `tests/{unit,contract,integration}` play tests

**Tasks Completed:** 24/24 tasks


## Merged Features Log

### List and Get Output Groups — archived 2026-11-23
**Branch:** 005-groups-list-get
**Spec:** [.specify/archive/005-groups-list-get/spec.md](../archive/005-groups-list-get/spec.md)

**What was added:**
- User Stories 17–21 (P1–P3): `sonora groups list` shows enabled output groups (identifier, display name, member outputs, mute and enabled state), `--include-disabled` shows disabled ones, `--json` for scripts
- `sonora groups get <group-id>` shows one group, enabled or disabled, with a clear "group not found" exit `5`
- FR-037–FR-046, SC-017–SC-021, AS-009 and the Output Group entity; shared outputs/inputs/routes requirements, outcomes and assumptions now cover groups as well (FR-006–FR-011, FR-013, SC-002–SC-004, AS-001–AS-006)

**New Components:**
- `internal/cli/groups/` (`RunList`, `RunGet`); `internal/hub/groups.go` (`ListGroups`, `GetGroup`); `internal/render/groups.go`
- `tests/{unit,contract,integration}` groups tests

**Tasks Completed:** 33/33 tasks


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
