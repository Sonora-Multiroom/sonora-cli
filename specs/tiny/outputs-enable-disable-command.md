# TinySpec: `enable outputs` / `disable outputs` command

**Branch**: feature/outputs-enable-disable-command
**Date**: 2026-08-29
**Status**: done
**Complexity**: small

## What

Extend the existing `enable`/`disable` verbs to `outputs`: add
`sonora enable outputs/<output-id>` and `sonora disable outputs/<output-id>`,
both calling `PUT /api/v2/outputs/{outputId}/enabled` (operationId
`setOutputEnabled`) with `{"enabled": true|false}`. On success the hub
returns the full updated `OutputResponse`, rendered with the existing
`RenderOutputYAML`/`RenderOutputJSON`. This mirrors
[inputs-enable-disable-command.md](inputs-enable-disable-command.md) exactly,
one resource kind later — `groups` enable/disable is still follow-up work.

## Context

| File | Role |
|------|------|
| `internal/hub/outputs.go` | Modified — add `SetOutputEnabled(ctx, client, baseURL, outputID string, enabled bool) (*Output, error)`, mirroring `SetInputEnabled`'s error handling (200 decode, 404 → `NotFoundError`, 400 → decode `errorResponse` into `APIError` falling back to `StatusError`, other non-2xx → `StatusError`) |
| `internal/render/outputs.go` | Context only — reuses existing `RenderOutputYAML`/`RenderOutputJSON` (response is a full `OutputResponse`, no new renderer needed) |
| `internal/cli/outputs/enabled.go` | New — `RunEnable`/`RunDisable`, mirroring `internal/cli/inputs/enabled.go`'s flag-parsing/positional-arg/error-reporting shape and shared `runSetEnabled` helper, adapted for `<output-id>`/`hub.SetOutputEnabled`/`render.RenderOutputYAML`/`RenderOutputJSON` |
| `cmd/sonora/main.go` | Modified — extend `dispatchEnabled` to accept `respath.Outputs` in addition to `respath.Inputs`, dispatching to `outputs.RunEnable`/`RunDisable`; update the usage string and `helpText` |
| `internal/hub/errors.go` | Context only — reuses existing `ClassNotFound`/`ClassValidation`/`ClassHub`, no new exit-code class needed |
| `tests/contract/outputs_enabled_test.go` | New — `TestSetOutputEnabled_*` contract tests against a mock server (200 enable/disable, 404, 400, malformed body), mirroring `inputs_enabled_test.go` |
| `tests/unit/cli_outputs_enabled_test.go` | New — unit tests for `RunEnable`/`RunDisable` (missing id, unexpected args, not-found, success, `--json`) |
| `tests/integration/outputs_enabled_test.go` | New — end-to-end test that `enable outputs/<id>` and `disable outputs/<id>` dispatch correctly with the right `enabled` body value, plus that `enable inputs/<id>` still works (regression) |
| `README.md` | Modified — document `sonora enable outputs/<id>` / `sonora disable outputs/<id>` |
| `docs/cli-command-landscape.md` | Modified — mark the two `outputs` enable/disable rows ✅ |

## Requirements

1. `sonora enable outputs/<output-id>` sends `PUT /api/v2/outputs/{outputId}/enabled`
   with `{"enabled": true}`; `sonora disable outputs/<output-id>` sends the same with
   `{"enabled": false}`. On success (200) each prints the updated output record (YAML
   by default, JSON with `--json`) and exits 0.
2. A 404 response reports "output not found: `<id>`" on stderr and exits with
   `ClassNotFound`'s code; a 400 reports the hub's error detail (or a generic
   "the request was rejected as invalid" fallback) and exits with `ClassValidation`'s
   code; other non-2xx/network/malformed-body failures reuse the existing
   `hub.ClassifyError` behavior — no stdout output on any failure path.
3. Missing `<output-id>` or unexpected extra arguments produce a usage error on
   stderr (exit code per `ClassUsage`), matching the pattern in `outputs.RunEnable`/
   `RunDisable`.
4. `--json`, `--verbose`, and `--hub-url` behave the same as every other command.
5. `enable inputs/<id>` / `disable inputs/<id>` keep working exactly as before — this
   change only widens `dispatchEnabled`'s accepted resource kind, it doesn't touch
   `inputs.RunEnable`/`RunDisable`.
6. `enable`/`disable` given a resource path that's neither `inputs` nor `outputs`
   (e.g. `groups/<id>`) is a usage error — no silent no-op or panic.

## Plan

1. Add contract tests in `tests/contract/outputs_enabled_test.go` for
   `hub.SetOutputEnabled` (fail first).
2. Implement `hub.SetOutputEnabled` in `internal/hub/outputs.go`.
3. Add unit tests in `tests/unit/cli_outputs_enabled_test.go`, then implement
   `outputs.RunEnable`/`RunDisable` in `internal/cli/outputs/enabled.go`, mirroring
   `internal/cli/inputs/enabled.go`.
4. Add integration test in `tests/integration/outputs_enabled_test.go` covering both
   verbs end-to-end (plus the `inputs` regression case), then extend
   `dispatchEnabled` in `cmd/sonora/main.go` to route `respath.Outputs` to
   `outputs.RunEnable`/`RunDisable` alongside the existing `respath.Inputs` case, and
   update the rejection message to name both supported kinds. Update `helpText`.
5. Update `README.md`'s command table and `docs/cli-command-landscape.md`'s ✅ markers.

## Tasks

- [x] Write contract tests for `hub.SetOutputEnabled` in `tests/contract/outputs_enabled_test.go`
- [x] Implement `hub.SetOutputEnabled` in `internal/hub/outputs.go`
- [x] Write unit tests + implement `outputs.RunEnable`/`RunDisable` in
      `internal/cli/outputs/enabled.go`
- [x] Write integration test + extend `dispatchEnabled` (incl. inputs regression and
      non-inputs/non-outputs rejection) in `cmd/sonora/main.go`, update `helpText`
- [x] Update `README.md` and `docs/cli-command-landscape.md`
- [x] Run `go test ./...`, `go vet ./...`, `gofmt -l .`

## Done When

- [x] All tasks checked off
- [x] `go test ./...` passes
- [x] `go vet ./...` and `gofmt -l .` clean
- [x] `sonora enable outputs/<id>` and `sonora disable outputs/<id>` verified
      end-to-end against a mock hub, including the 404/400 error paths, with
      `enable inputs/<id>` still working unchanged
