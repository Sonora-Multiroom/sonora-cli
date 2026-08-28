# TinySpec: `set outputs/<output-id> volume <0-100>` command

**Branch**: feature/outputs-set-volume-command
**Date**: 2026-08-29
**Status**: accepted
**Complexity**: small

## What

Add `sonora set outputs/<output-id> volume <0-100>`, calling
`PUT /api/v2/outputs/{outputId}/volume` (operationId `setOutputVolume`) with
`{"volume": <int>}`. On success the hub returns `OutputVolumeResponse`
(`outputId`, `volume`, `updatedAt`) — a different shape than `Output`, so it
gets its own hub type and renderer. This also introduces the `set` verb and
its generic `set <resource>/<id> <attribute> <value>` dispatch shape (per
[docs/cli-command-landscape.md](../../docs/cli-command-landscape.md)); only
`outputs`/`volume` is wired up now — `groups`/`volume` and any other
attribute are follow-up work, matching how `enable`/`disable` shipped
`inputs`-only first.

## Context

| File | Role |
|------|------|
| `internal/hub/outputs.go` | Modified — add `OutputVolume` struct (`OutputID`, `Volume`, `UpdatedAt`) mirroring `OutputVolumeResponse`, and `SetOutputVolume(ctx, client, baseURL, outputID string, volume int) (*OutputVolume, error)`, mirroring `SetInputEnabled`'s error handling (200 decode, 404 → `NotFoundError`, 400 → decode `errorResponse` into `APIError` falling back to `StatusError`, other non-2xx → `StatusError`) |
| `internal/render/outputs.go` | Modified — add `RenderOutputVolumeYAML`/`RenderOutputVolumeJSON` for the bare `OutputVolume` record (no list wrapper), mirroring `RenderOutputYAML`/`RenderOutputJSON` |
| `internal/cli/outputs/volume.go` | New — `RunSetVolume`, mirroring `get.go`/`inputs/enabled.go`'s flag-parsing/positional-arg/error-reporting shape; expects exactly 3 positionals (`<output-id>`, literal `volume`, `<value>`), parses `<value>` as an integer 0-100 (usage error if not) |
| `cmd/sonora/main.go` | Modified — add `case "set":` dispatching to a new `dispatchSet` (parses the resource path via `respath`, rejects non-`outputs` kinds as a usage error, passes the rest through to `outputs.RunSetVolume`); update `helpText` |
| `internal/hub/errors.go` | Context only — reuses existing `ClassNotFound`/`ClassValidation`/`ClassUsage`/`ClassHub`, no new exit-code class needed |
| `tests/contract/outputs_volume_test.go` | New — `TestSetOutputVolume_*` contract tests against a mock server (200, 404, 400, boundary values 0/100, malformed body), mirroring `inputs_enabled_test.go` |
| `tests/unit/cli_outputs_volume_test.go` | New — unit tests for `RunSetVolume` (missing id, missing/wrong attribute word, non-numeric value, out-of-range value, unexpected args, not-found, success, `--json`) |
| `tests/unit/render_output_volume_test.go` | New — unit tests for `RenderOutputVolumeYAML`/`RenderOutputVolumeJSON` |
| `tests/integration/outputs_volume_test.go` | New — end-to-end test that `set outputs/<id> volume <n>` dispatches correctly with the right request body |
| `README.md` | Modified — document `sonora set outputs/<id> volume <0-100>` |
| `docs/cli-command-landscape.md` | Modified — mark the `outputs`/`volume` row ✅ |

## Requirements

1. `sonora set outputs/<output-id> volume <n>` sends `PUT /api/v2/outputs/{outputId}/volume`
   with `{"volume": n}`. On success (200) it prints the returned `outputId`/`volume`/`updatedAt`
   record (YAML by default, JSON with `--json`) and exits 0.
2. `<n>` must be an integer between 0 and 100 inclusive; a non-integer or out-of-range value is
   a usage error on stderr (no network call), exit code per `ClassUsage`.
3. A 404 response reports "output not found: `<id>`" on stderr and exits with `ClassNotFound`'s
   code; a 400 reports the hub's error detail (or a generic "the request was rejected as
   invalid" fallback) and exits with `ClassValidation`'s code; other non-2xx/network/malformed-
   body failures reuse the existing `hub.ClassifyError` behavior — no stdout output on any
   failure path.
4. Missing `<output-id>`, a missing/misspelled `volume` attribute word, missing `<value>`, or
   unexpected extra arguments produce a usage error on stderr (exit code per `ClassUsage`),
   matching the pattern in `inputs.RunEnable`/`RunDisable`.
5. `--json`, `--verbose`, and `--hub-url` behave the same as every other command.
6. `set` given a non-`outputs` resource path (e.g. `groups/<id>`) or a non-`volume` attribute is
   a usage error — no other resource/attribute combination is supported yet, so don't silently
   no-op or panic.

## Plan

1. Add contract tests in `tests/contract/outputs_volume_test.go` for `hub.SetOutputVolume`
   (fail first).
2. Implement `hub.OutputVolume` and `hub.SetOutputVolume` in `internal/hub/outputs.go`.
3. Add render unit tests, then implement `RenderOutputVolumeYAML`/`RenderOutputVolumeJSON` in
   `internal/render/outputs.go`.
4. Add CLI unit tests in `tests/unit/cli_outputs_volume_test.go`, then implement
   `outputs.RunSetVolume` in `internal/cli/outputs/volume.go`, validating the `volume` attribute
   word and the 0-100 integer value before calling the hub.
5. Add integration test in `tests/integration/outputs_volume_test.go`, then wire
   `case "set":` in `cmd/sonora/main.go`'s `run()` via a new `dispatchSet` (parse the resource
   path via `respath`, reject non-`outputs` kinds as a usage error, call `outputs.RunSetVolume`
   with the rest of the args). Update `helpText`.
6. Update `README.md`'s command section and `docs/cli-command-landscape.md`'s ✅ marker.

## Tasks

- [ ] Write contract tests for `hub.SetOutputVolume` in `tests/contract/outputs_volume_test.go`
- [ ] Implement `hub.OutputVolume` + `hub.SetOutputVolume` in `internal/hub/outputs.go`
- [ ] Write render tests + implement `RenderOutputVolumeYAML`/`RenderOutputVolumeJSON` in
      `internal/render/outputs.go`
- [ ] Write unit tests + implement `outputs.RunSetVolume` in `internal/cli/outputs/volume.go`
- [ ] Write integration test + wire `set` dispatch (incl. non-outputs/non-volume rejection) in
      `cmd/sonora/main.go`, update `helpText`
- [ ] Update `README.md` and `docs/cli-command-landscape.md`
- [ ] Run `go test ./...`, `go vet ./...`, `gofmt -l .`

## Done When

- [ ] All tasks checked off
- [ ] `go test ./...` passes
- [ ] `go vet ./...` and `gofmt -l .` clean
- [ ] `sonora set outputs/<id> volume <n>` verified end-to-end against a mock hub, including the
      404/400/out-of-range error paths
