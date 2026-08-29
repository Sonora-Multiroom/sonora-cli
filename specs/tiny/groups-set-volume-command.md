# TinySpec: `set groups/<group-id> volume <0-100>` command

**Branch**: feature/groups-set-volume-command
**Date**: 2026-08-29
**Status**: done
**Complexity**: small

## What

Extend `set` to support groups: `sonora set groups/<group-id> volume <0-100>`,
calling `PUT /api/v2/groups/{groupId}/volume` (operationId `setGroupVolume`)
with `{"volume": <int>}`. On success the hub returns `GroupVolumeResponse`
(`groupId`, `volume`, `updatedAt`), mirroring the existing
`set outputs/<id> volume <0-100>` command
([outputs-set-volume-command.md](outputs-set-volume-command.md)) field-for-field
but for the `groups` resource. `dispatchSet` in `cmd/sonora/main.go` currently
hard-rejects every resource kind except `Outputs`; this widens it to accept
`Groups` too and route to a new `groups.RunSetVolume`.

## Context

| File | Role |
|------|------|
| `internal/hub/groups.go` | Modified — add `GroupVolume` struct (`GroupID`, `Volume`, `UpdatedAt`) mirroring `GroupVolumeResponse`, and `SetGroupVolume(ctx, client, baseURL, groupID string, volume int) (*GroupVolume, error)`, mirroring `hub.SetOutputVolume`'s error handling (200 decode, 404 → `NotFoundError{Resource: "group", ID: groupID}`, 400 → decode `errorResponse` into `APIError` falling back to `StatusError`, other non-2xx → `StatusError`) |
| `internal/render/groups.go` | Modified — add `RenderGroupVolumeYAML`/`RenderGroupVolumeJSON` for the bare `GroupVolume` record (no list wrapper), mirroring `RenderOutputVolumeYAML`/`RenderOutputVolumeJSON` |
| `internal/cli/groups/volume.go` | New — `RunSetVolume`, copying `outputs.RunSetVolume`'s flag-parsing/positional-arg/error-reporting shape (including the negative-value-peeling loop and `looksNumeric` helper) verbatim except for the hub/render calls and usage string |
| `cmd/sonora/main.go` | Modified — `dispatchSet` currently returns a usage error for any `path.Kind != respath.Outputs`; change the check to accept `respath.Outputs` or `respath.Groups` and call `outputs.RunSetVolume` or `groups.RunSetVolume` respectively (still rejecting `inputs`/`routes`); update the usage string, `helpText`, and the `set` help line to read `set <outputs\|groups>/<id> volume <0-100>` |
| `internal/hub/errors.go` | Context only — reuses existing `ClassNotFound`/`ClassValidation`/`ClassUsage`/`ClassHub`, no new exit-code class needed |
| `tests/contract/groups_volume_test.go` | New — `TestSetGroupVolume_*` contract tests against a mock server (200, 404, 400, boundary values 0/100, malformed body), mirroring `tests/contract/outputs_volume_test.go` |
| `tests/unit/cli_groups_volume_test.go` | New — unit tests for `groups.RunSetVolume` (missing id, missing/wrong attribute word, non-numeric value, out-of-range value, unexpected args, not-found, success, `--json`), mirroring `tests/unit/cli_outputs_volume_test.go` |
| `tests/unit/render_group_volume_test.go` | New — unit tests for `RenderGroupVolumeYAML`/`RenderGroupVolumeJSON` |
| `tests/integration/groups_volume_test.go` | New — end-to-end test that `set groups/<id> volume <n>` dispatches correctly with the right request body; extend/add a dispatch test confirming `set outputs/<id> ...` still works and `set inputs/<id> ...`/`set routes/<id> ...` still fail as usage errors |
| `README.md` | Modified — update the `set` row and prose (lines 45, 60-61) to cover both `outputs` and `groups` |
| `docs/cli-command-landscape.md` | Modified — mark the `groups`/`volume` row (line 59) ✅ |

## Requirements

1. `sonora set groups/<group-id> volume <n>` sends `PUT /api/v2/groups/{groupId}/volume`
   with `{"volume": n}`. On success (200) it prints the returned `groupId`/`volume`/`updatedAt`
   record (YAML by default, JSON with `--json`) and exits 0.
2. `<n>` must be an integer between 0 and 100 inclusive; a non-integer or out-of-range value is
   a usage error on stderr (no network call), exit code per `ClassUsage`.
3. A 404 response reports "group not found: `<id>`" on stderr and exits with `ClassNotFound`'s
   code; a 400 reports the hub's error detail (or a generic fallback) and exits with
   `ClassValidation`'s code; other non-2xx/network/malformed-body failures reuse the existing
   `hub.ClassifyError` behavior — no stdout output on any failure path.
4. Missing `<group-id>`, a missing/misspelled `volume` attribute word, missing `<value>`, or
   unexpected extra arguments produce a usage error on stderr (exit code per `ClassUsage`),
   matching `outputs.RunSetVolume`.
5. `--json`, `--verbose`, and `--hub-url` behave the same as every other command.
6. `set outputs/<id> volume <n>` continues to work unchanged; `set` given `inputs/<id>` or
   `routes/<id>` remains a usage error — only `outputs` and `groups` support `set` now.

## Plan

1. Add contract tests in `tests/contract/groups_volume_test.go` for `hub.SetGroupVolume`
   (fail first).
2. Implement `hub.GroupVolume` and `hub.SetGroupVolume` in `internal/hub/groups.go`.
3. Add render unit tests, then implement `RenderGroupVolumeYAML`/`RenderGroupVolumeJSON` in
   `internal/render/groups.go`.
4. Add CLI unit tests in `tests/unit/cli_groups_volume_test.go`, then implement
   `groups.RunSetVolume` in `internal/cli/groups/volume.go` by copying `outputs.RunSetVolume`'s
   shape.
5. Add/extend integration tests in `tests/integration/groups_volume_test.go`, then widen
   `dispatchSet` in `cmd/sonora/main.go` to accept `respath.Groups` alongside
   `respath.Outputs`, dispatching to the matching package. Update the usage string and
   `helpText`.
6. Update `README.md`'s `set` row/prose and `docs/cli-command-landscape.md`'s ✅ marker.

## Tasks

- [x] Write contract tests for `hub.SetGroupVolume` in `tests/contract/groups_volume_test.go`
- [x] Implement `hub.GroupVolume` + `hub.SetGroupVolume` in `internal/hub/groups.go`
- [x] Write render tests + implement `RenderGroupVolumeYAML`/`RenderGroupVolumeJSON` in
      `internal/render/groups.go`
- [x] Write unit tests + implement `groups.RunSetVolume` in `internal/cli/groups/volume.go`
- [x] Write/extend integration tests + widen `dispatchSet` to accept `groups` (incl. confirming
      `inputs`/`routes` still rejected) in `cmd/sonora/main.go`, update `helpText`
- [x] Update `README.md` and `docs/cli-command-landscape.md`
- [x] Run `go test ./...`, `go vet ./...`, `gofmt -l .`

## Done When

- [x] All tasks checked off
- [x] `go test ./...` passes
- [x] `go vet ./...` and `gofmt -l .` clean
- [x] `sonora set groups/<id> volume <n>` verified end-to-end against a mock hub, including the
      404/400/out-of-range error paths, and `set outputs/<id> volume <n>` still works
