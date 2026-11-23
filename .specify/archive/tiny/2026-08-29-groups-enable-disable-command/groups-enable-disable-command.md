# TinySpec: `enable groups` / `disable groups` command

**Branch**: feature/groups-enable-disable-command
**Date**: 2026-08-29
**Status**: done
**Complexity**: small

## What

Extend `enable`/`disable` to `groups`: add `sonora enable groups/<group-id>`
and `sonora disable groups/<group-id>`, both calling
`PUT /api/v2/groups/{groupId}/enabled` (operationId `setGroupEnabled`) with
`{"enabled": true|false}`. On success the hub returns the full updated
`GroupResponse`, rendered with the existing `render.RenderGroupYAML`/
`RenderGroupJSON`. This mirrors
[outputs-enable-disable-command.md](outputs-enable-disable-command.md)
field-for-field, one resource kind later — `dispatchEnabled` in
`cmd/sonora/main.go` currently hard-rejects every kind except `Inputs`/
`Outputs`; this widens it to accept `Groups` too, the same way `dispatchSet`
was already widened to accept `Groups` alongside `Outputs`
([groups-set-volume-command.md](groups-set-volume-command.md)).

## Context

| File | Role |
|------|------|
| `internal/hub/groups.go` | Modified — add `SetGroupEnabled(ctx, client, baseURL, groupID string, enabled bool) (*Group, error)`, mirroring `SetOutputEnabled`'s error handling (200 decode into the existing `Group` struct, 404 → `NotFoundError{Resource: "group", ID: groupID}`, 400 → decode `errorResponse` into `APIError` falling back to `StatusError`, other non-2xx → `StatusError`) |
| `internal/render/groups.go` | Context only — reuses existing `RenderGroupYAML`/`RenderGroupJSON` (response is a full `Group`/`GroupResponse`, no new renderer needed) |
| `internal/cli/groups/enabled.go` | New — `RunEnable`/`RunDisable`, mirroring `internal/cli/outputs/enabled.go`'s flag-parsing/positional-arg/error-reporting shape and shared `runSetEnabled` helper, adapted for `<group-id>`/`hub.SetGroupEnabled`/`render.RenderGroupYAML`/`RenderGroupJSON` |
| `cmd/sonora/main.go` | Modified — extend `dispatchEnabled` to accept `respath.Groups` in addition to `respath.Inputs`/`respath.Outputs`, dispatching to `groups.RunEnable`/`RunDisable`; update the usage string, the unsupported-kind error message, and `helpText` |
| `internal/hub/errors.go` | Context only — reuses existing `ClassNotFound`/`ClassValidation`/`ClassHub`/`ClassUsage`, no new exit-code class needed |
| `tests/contract/groups_enabled_test.go` | New — `TestSetGroupEnabled_*` contract tests against a mock server (200 enable/disable, 404, 400, malformed body), mirroring `tests/contract/outputs_enabled_test.go` |
| `tests/unit/cli_groups_enabled_test.go` | New — unit tests for `groups.RunEnable`/`RunDisable` (missing id, unexpected args, not-found, success, `--json`) |
| `tests/integration/groups_enabled_test.go` | New — end-to-end test that `enable groups/<id>` and `disable groups/<id>` dispatch correctly with the right `enabled` body value, plus that `enable inputs/<id>` and `enable outputs/<id>` still work (regression) |
| `README.md` | Modified — document `sonora enable groups/<id>` / `sonora disable groups/<id>` |
| `docs/cli-command-landscape.md` | Modified — mark the two `groups` enable/disable rows ✅ |

## Requirements

1. `sonora enable groups/<group-id>` sends `PUT /api/v2/groups/{groupId}/enabled`
   with `{"enabled": true}`; `sonora disable groups/<group-id>` sends the same with
   `{"enabled": false}`. On success (200) each prints the updated group record (YAML
   by default, JSON with `--json`) and exits 0.
2. A 404 response reports "group not found: `<id>`" on stderr and exits with
   `ClassNotFound`'s code; a 400 reports the hub's error detail (or a generic
   "the request was rejected as invalid" fallback) and exits with `ClassValidation`'s
   code; other non-2xx/network/malformed-body failures reuse the existing
   `hub.ClassifyError` behavior — no stdout output on any failure path.
3. Missing `<group-id>` or unexpected extra arguments produce a usage error on
   stderr (exit code per `ClassUsage`), matching the pattern in `outputs.RunEnable`/
   `RunDisable`.
4. `--json`, `--verbose`, and `--hub-url` behave the same as every other command.
5. `enable inputs/<id>`/`enable outputs/<id>` and their `disable` counterparts keep
   working exactly as before — this change only widens `dispatchEnabled`'s accepted
   resource kind, it doesn't touch `inputs.RunEnable`/`RunDisable` or
   `outputs.RunEnable`/`RunDisable`.
6. `enable`/`disable` given a resource path that's neither `inputs`, `outputs`, nor
   `groups` (e.g. `routes/<id>`) is a usage error — no silent no-op or panic.

## Plan

1. Add contract tests in `tests/contract/groups_enabled_test.go` for
   `hub.SetGroupEnabled` (fail first).
2. Implement `hub.SetGroupEnabled` in `internal/hub/groups.go`.
3. Add unit tests in `tests/unit/cli_groups_enabled_test.go`, then implement
   `groups.RunEnable`/`RunDisable` in `internal/cli/groups/enabled.go`, mirroring
   `internal/cli/outputs/enabled.go`.
4. Add integration test in `tests/integration/groups_enabled_test.go` covering both
   verbs end-to-end (plus the `inputs`/`outputs` regression cases), then extend
   `dispatchEnabled` in `cmd/sonora/main.go` to route `respath.Groups` to
   `groups.RunEnable`/`RunDisable` alongside the existing `Inputs`/`Outputs` cases,
   and update the rejection message to name all three supported kinds. Update
   `helpText`.
5. Update `README.md`'s command table and `docs/cli-command-landscape.md`'s ✅ markers.

## Tasks

- [x] Write contract tests for `hub.SetGroupEnabled` in `tests/contract/groups_enabled_test.go`
- [x] Implement `hub.SetGroupEnabled` in `internal/hub/groups.go`
- [x] Write unit tests + implement `groups.RunEnable`/`RunDisable` in
      `internal/cli/groups/enabled.go`
- [x] Write integration test + extend `dispatchEnabled` (incl. inputs/outputs
      regression and non-inputs/outputs/groups rejection) in `cmd/sonora/main.go`,
      update `helpText`
- [x] Update `README.md` and `docs/cli-command-landscape.md`
- [x] Run `go test ./...`, `go vet ./...`, `gofmt -l .`

## Done When

- [x] All tasks checked off
- [x] `go test ./...` passes
- [x] `go vet ./...` and `gofmt -l .` clean
- [x] `sonora enable groups/<id>` and `sonora disable groups/<id>` verified
      end-to-end against a mock hub, including the 404/400 error paths, with
      `enable inputs/<id>` and `enable outputs/<id>` still working unchanged
