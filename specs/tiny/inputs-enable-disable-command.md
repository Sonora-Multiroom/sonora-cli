# TinySpec: `enable inputs` / `disable inputs` command

**Branch**: feature/inputs-enable-disable-command
**Date**: 2026-08-28
**Status**: done
**Complexity**: small

## What

Add `sonora enable inputs/<input-id>` and `sonora disable inputs/<input-id>`, both calling
`PUT /api/v2/inputs/{inputId}/enabled` (operationId `setInputEnabled`) with
`{"enabled": true|false}`. On success the hub returns the full updated `InputResponse`, which
is rendered with the existing `RenderInputYAML`/`RenderInputJSON`. Only `inputs` is in scope —
`outputs`/`groups` enable/disable (same `/enabled` shape, per
[docs/cli-command-landscape.md](../../docs/cli-command-landscape.md)) are follow-up work.

## Context

| File | Role |
|------|------|
| `internal/hub/inputs.go` | Modified — add `SetInputEnabled(ctx, client, baseURL, inputID string, enabled bool) (*Input, error)`, mirroring `GetInput`'s error handling (200 decode, 404 → `NotFoundError`, 400 → `APIError`, other non-2xx → `StatusError`) |
| `internal/render/inputs.go` | Context only — reuses existing `RenderInputYAML`/`RenderInputJSON` (response is a full `InputResponse`, no new renderer needed) |
| `internal/cli/inputs/enabled.go` | New — `RunEnable`/`RunDisable`, mirroring `routes/delete.go`'s flag-parsing/positional-arg/error-reporting shape, sharing one internal `runSetEnabled(enabled bool, ...)` helper |
| `cmd/sonora/main.go` | Modified — dispatch top-level `enable`/`disable` verbs (inputs only) to `inputs.RunEnable`/`RunDisable`; update `helpText` |
| `internal/hub/errors.go` | Context only — reuses existing `ClassNotFound`/`ClassValidation`/`ClassHub` (400 → `ClassValidation`, no new exit-code class needed) |
| `tests/contract/inputs_enabled_test.go` | New — `TestSetInputEnabled_*` contract tests against a mock server (200 enable/disable, 404, 400, malformed body), mirroring `inputs_get_test.go` |
| `tests/unit/cli_inputs_enabled_test.go` | New — unit tests for `RunEnable`/`RunDisable` (missing id, unexpected args, not-found, success, `--json`) |
| `tests/integration/inputs_enabled_test.go` | New — end-to-end test that `enable inputs/<id>` and `disable inputs/<id>` dispatch correctly with the right `enabled` body value |
| `README.md` | Modified — document `sonora enable inputs/<id>` / `sonora disable inputs/<id>` |
| `docs/cli-command-landscape.md` | Modified — mark the two `inputs` enable/disable rows ✅ |

## Requirements

1. `sonora enable inputs/<input-id>` sends `PUT /api/v2/inputs/{inputId}/enabled` with
   `{"enabled": true}`; `sonora disable inputs/<input-id>` sends the same with
   `{"enabled": false}`. On success (200) each prints the updated input record (YAML by
   default, JSON with `--json`) and exits 0.
2. A 404 response reports "input not found: `<id>`" on stderr and exits with
   `ClassNotFound`'s code; a 400 reports the hub's error detail (or a generic
   "the request was rejected as invalid" fallback) and exits with `ClassValidation`'s code;
   other non-2xx/network/malformed-body failures reuse the existing `hub.ClassifyError`
   behavior — no stdout output on any failure path.
3. Missing `<input-id>` or unexpected extra arguments produce a usage error on stderr (exit
   code per `ClassUsage`), matching the pattern in `routes.RunDelete`.
4. `--json`, `--verbose`, and `--hub-url` behave the same as every other command.
5. `enable`/`disable` given a non-`inputs` resource path (e.g. `outputs/<id>`) is a usage
   error — no other resource supports enable/disable yet, so don't silently no-op or panic.

## Plan

1. Add contract tests in `tests/contract/inputs_enabled_test.go` for `hub.SetInputEnabled`
   (fail first).
2. Implement `hub.SetInputEnabled` in `internal/hub/inputs.go`.
3. Add unit tests in `tests/unit/cli_inputs_enabled_test.go`, then implement
   `inputs.RunEnable`/`RunDisable` in `internal/cli/inputs/enabled.go`, mirroring
   `routes/delete.go`'s structure (shared helper parameterized by the `enabled` bool and by
   verb name for usage strings).
4. Add integration test in `tests/integration/inputs_enabled_test.go` covering both verbs
   end-to-end, then wire `case "enable", "disable":` in `cmd/sonora/main.go`'s `run()` —
   parse the resource path via `respath`, reject non-`inputs` kinds as a usage error, and
   call the matching `inputs.Run*`. Update `helpText`.
5. Update `README.md`'s command table and `docs/cli-command-landscape.md`'s ✅ markers.

## Tasks

- [x] Write contract tests for `hub.SetInputEnabled` in `tests/contract/inputs_enabled_test.go`
- [x] Implement `hub.SetInputEnabled` in `internal/hub/inputs.go`
- [x] Write unit tests + implement `inputs.RunEnable`/`RunDisable` in
      `internal/cli/inputs/enabled.go`
- [x] Write integration test + wire `enable`/`disable` dispatch (incl. non-inputs rejection)
      in `cmd/sonora/main.go`, update `helpText`
- [x] Update `README.md` and `docs/cli-command-landscape.md`
- [x] Run `go test ./...`, `go vet ./...`, `gofmt -l .`

## Done When

- [x] All tasks checked off
- [x] `go test ./...` passes
- [x] `go vet ./...` and `gofmt -l .` clean
- [x] `sonora enable inputs/<id>` and `sonora disable inputs/<id>` verified end-to-end against
      a mock hub, including the 404/400 error paths
