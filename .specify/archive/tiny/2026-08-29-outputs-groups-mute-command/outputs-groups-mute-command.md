# TinySpec: `mute`/`unmute outputs|groups` command

**Branch**: feature/outputs-groups-mute-command
**Date**: 2026-08-29
**Status**: done
**Complexity**: small

## What

Add a new `mute`/`unmute` verb pair for `outputs` and `groups`: `sonora mute
outputs/<output-id>`, `sonora unmute outputs/<output-id>`, `sonora mute
groups/<group-id>`, `sonora unmute groups/<group-id>`. All four call `PUT
/api/v2/{outputs|groups}/{id}/mute` with `{"muted": true|false}`; on success
the hub returns the full updated record, rendered with the existing
`render.RenderOutputYAML`/`JSON` or `RenderGroupYAML`/`JSON` (both already
have a `Muted` field). This mirrors
[groups-enable-disable-command.md](groups-enable-disable-command.md)
field-for-field — `enabled`→`muted`, plus a brand-new top-level verb
(`mute`/`unmute`) instead of widening an existing one. `mute all`/`unmute
all` (`master-mute`) and `inputs` are explicitly out of scope per
`docs/cli-command-landscape.md`.

## Context

| File | Role |
|------|------|
| `internal/hub/outputs.go` | Modified — add `SetOutputMuted(ctx, client, baseURL, outputID string, muted bool) (*Output, error)`, mirroring `SetOutputEnabled`'s error handling (200 decode, 404 → `NotFoundError`, 400 → decode `errorResponse` into `APIError` falling back to `StatusError`, other non-2xx → `StatusError`) against `PUT .../outputs/{outputId}/mute` with body `{"muted": muted}` |
| `internal/hub/groups.go` | Modified — add `SetGroupMuted(ctx, client, baseURL, groupID string, muted bool) (*Group, error)`, mirroring `SetGroupEnabled` against `PUT .../groups/{groupId}/mute` with body `{"muted": muted}` |
| `internal/render/outputs.go`, `internal/render/groups.go` | Context only — reuse existing `RenderOutputYAML`/`JSON` and `RenderGroupYAML`/`JSON` (response is the full record, no new renderer needed) |
| `internal/cli/outputs/muted.go` | New — `RunMute`/`RunUnmute`, mirroring `internal/cli/outputs/enabled.go`'s flag-parsing/positional-arg/error-reporting shape and shared `runSetMuted` helper, adapted for `hub.SetOutputMuted` |
| `internal/cli/groups/muted.go` | New — `RunMute`/`RunUnmute`, mirroring `internal/cli/groups/enabled.go`, adapted for `hub.SetGroupMuted` |
| `cmd/sonora/main.go` | Modified — add `case "mute", "unmute":` to the top-level `run` switch dispatching to a new `dispatchMuted`, mirroring `dispatchEnabled` but restricted to `respath.Outputs`/`respath.Groups` (no `inputs`); update `helpText`'s command table and examples |
| `internal/hub/errors.go` | Context only — reuses existing `ClassNotFound`/`ClassValidation`/`ClassHub`/`ClassUsage`, no new exit-code class needed |
| `tests/contract/outputs_muted_test.go` | New — `TestSetOutputMuted_*` contract tests against a mock server (200 mute/unmute, 404, 400, malformed body), mirroring `outputs_enabled_test.go` |
| `tests/contract/groups_muted_test.go` | New — `TestSetGroupMuted_*` contract tests, mirroring `groups_enabled_test.go` |
| `tests/unit/cli_outputs_muted_test.go`, `tests/unit/cli_groups_muted_test.go` | New — unit tests for `RunMute`/`RunUnmute` (missing id, unexpected args, not-found, success, `--json`) |
| `tests/integration/outputs_groups_muted_test.go` | New — end-to-end test that `mute outputs/<id>`, `unmute outputs/<id>`, `mute groups/<id>`, `unmute groups/<id>` dispatch correctly with the right `muted` body value, plus that `mute inputs/<id>` and `mute routes/<id>` are rejected as usage errors |
| `README.md` | Modified — document the four new commands |
| `docs/cli-command-landscape.md` | Modified — mark the four `outputs`/`groups` mute rows ✅ (rows already exist at lines 47-48, 60-61) |

## Requirements

1. `sonora mute outputs/<output-id>` sends `PUT /api/v2/outputs/{outputId}/mute`
   with `{"muted": true}`; `sonora unmute outputs/<output-id>` sends the same with
   `{"muted": false}`. Same pair for `groups/<group-id>` against
   `/api/v2/groups/{groupId}/mute`. On success (200) each prints the updated record
   (YAML by default, JSON with `--json`) and exits 0.
2. A 404 response reports "output/group not found: `<id>`" on stderr and exits with
   `ClassNotFound`'s code; a 400 reports the hub's error detail (or a generic
   fallback) and exits with `ClassValidation`'s code; other non-2xx/network/
   malformed-body failures reuse `hub.ClassifyError` — no stdout output on any
   failure path.
3. Missing `<id>` or unexpected extra arguments produce a usage error on stderr
   (exit code per `ClassUsage`), matching the pattern in `outputs.RunEnable`/
   `RunDisable`.
4. `--json`, `--verbose`, and `--hub-url` behave the same as every other command.
5. `mute`/`unmute` given a resource path that's neither `outputs` nor `groups`
   (e.g. `inputs/<id>`, `routes/<id>`, or the bare `all`) is a usage error — no
   silent no-op or panic. `mute all`/`master-mute` stays unimplemented follow-up
   work.
6. No existing verb's behavior changes — `enable`/`disable`/`pause`/`resume`/`set`
   dispatch is untouched.

## Plan

1. Add contract tests in `tests/contract/outputs_muted_test.go` and
   `tests/contract/groups_muted_test.go` for `hub.SetOutputMuted`/`SetGroupMuted`
   (fail first).
2. Implement `hub.SetOutputMuted` in `internal/hub/outputs.go` and
   `hub.SetGroupMuted` in `internal/hub/groups.go`.
3. Add unit tests, then implement `outputs.RunMute`/`RunUnmute` in
   `internal/cli/outputs/muted.go` and `groups.RunMute`/`RunUnmute` in
   `internal/cli/groups/muted.go`, mirroring the `enabled.go` files.
4. Add integration test in `tests/integration/outputs_groups_muted_test.go`
   covering all four verb/resource combos plus the inputs/routes/`all` rejection
   cases, then add `dispatchMuted` to `cmd/sonora/main.go` and wire `case "mute",
   "unmute":` into the top-level `run` switch. Update `helpText`.
5. Update `README.md`'s command table and flip the four `outputs`/`groups` mute
   rows to ✅ in `docs/cli-command-landscape.md`.

## Tasks

- [x] Write contract tests for `hub.SetOutputMuted`/`SetGroupMuted`
- [x] Implement `hub.SetOutputMuted` and `hub.SetGroupMuted`
- [x] Write unit tests + implement `outputs.RunMute`/`RunUnmute` and
      `groups.RunMute`/`RunUnmute`
- [x] Write integration test + add `dispatchMuted` and wire it into `run` in
      `cmd/sonora/main.go`, update `helpText`
- [x] Update `README.md` and `docs/cli-command-landscape.md`
- [x] Run `go test ./...`, `go vet ./...`, `gofmt -l .`

## Done When

- [x] All tasks checked off
- [x] `go test ./...` passes
- [x] `go vet ./...` and `gofmt -l .` clean
- [x] `sonora mute outputs/<id>`, `unmute outputs/<id>`, `mute groups/<id>`,
      `unmute groups/<id>` verified end-to-end against a mock hub, including the
      404/400 error paths and the inputs/routes/`all` rejection cases
