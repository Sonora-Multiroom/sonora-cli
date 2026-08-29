# TinySpec: `get master-mute` / `mute`/`unmute all` command

**Branch**: feature/master-mute-command
**Date**: 2026-08-29
**Status**: done
**Complexity**: small

## What

Add the system-wide master-mute singleton: `sonora get master-mute` (GET
`/api/v2/master-mute`), `sonora mute all` and `sonora unmute all` (both PUT
`/api/v2/master-mute` with `{"muted": true|false}`). Response is always
`{"muted": bool}` (`MasterMuteResponse`), rendered by new
`render.RenderMasterMuteYAML`/`JSON`. Mirrors
[outputs-groups-mute-command.md](outputs-groups-mute-command.md)'s shape but
singular: master-mute has no id, no collection, and (per `api/openapi.json`)
neither operation documents a 404 or 400 response, so no `NotFoundError`/
`APIError` handling is needed. `master-mute` and `all` are literal keywords,
not `resource/id` paths, so they're special-cased directly in
`cmd/sonora/main.go`'s dispatchers rather than added to `respath`'s
`ResourceKind` enum (which models addressable, id-bearing collections shared
with `play`'s target argument).

## Context

| File | Role |
|------|------|
| `api/openapi.json` | Context only — `getMasterMute`/`setMasterMute` operations (lines 312-363), `MasterMuteResponse` schema (line 1653): only a 200 response is documented for either operation |
| `internal/hub/mastermute.go` | New — `MasterMute{Muted bool}`, `GetMasterMute(ctx, client, baseURL) (*MasterMute, error)` (GET, mirrors `ListOutputs`'s 2xx-only handling, no id), `SetMasterMute(ctx, client, baseURL string, muted bool) (*MasterMute, error)` (PUT `{"muted": muted}`, same 2xx-only handling — non-2xx → `StatusError`, malformed body → `DecodeError`) |
| `internal/render/mastermute.go` | New — `RenderMasterMuteYAML(hub.MasterMute) string` and `RenderMasterMuteJSON(hub.MasterMute) string`, mirroring `RenderOutputVolumeYAML`/`JSON`'s bare-record shape |
| `internal/cli/mastermute/mastermute.go` | New — `RunGet`, `RunMute`, `RunUnmute`, mirroring `outputs.RunGet`/`RunMute`'s flag-parsing/error-reporting shape but with **zero** required positional arguments (any positional argument is "unexpected") |
| `cmd/sonora/main.go` | Modified — `dispatchGetList` special-cases `args[0] == "master-mute"`: `list master-mute` and `get master-mute/<anything>` are usage errors, `get master-mute` calls `mastermute.RunGet`; `dispatchMuted` special-cases `args[0] == "all"` (exact match) to call `mastermute.RunMute`/`RunUnmute`, ahead of its existing `respath.Parse` path for `outputs`/`groups`; update `helpText`'s command table and examples |
| `internal/cli/respath/respath.go` | Context only — **not modified**; `master-mute`/`all` stay outside `ResourceKind` since they carry no id and aren't valid `play` targets |
| `internal/hub/errors.go` | Context only — only `ClassHub`/`ClassNetwork` are reachable (no id to 404 on, no documented 400) |
| `tests/contract/master_mute_test.go` | New — `TestGetMasterMute_*`/`TestSetMasterMute_*` contract tests against a mock server (200 get, 200 mute/unmute, non-2xx, malformed body) |
| `tests/unit/cli_master_mute_test.go` | New — unit tests for `RunGet`/`RunMute`/`RunUnmute` (unexpected args, hub failure, success, `--json`) |
| `tests/integration/master_mute_test.go` | New — end-to-end test that `get master-mute`, `mute all`, `unmute all` dispatch correctly, plus that `get master-mute/x`, `list master-mute`, and `mute all/foo` are rejected as usage errors |
| `README.md` | Modified — document the three new commands |
| `docs/cli-command-landscape.md` | Modified — mark the three `master-mute` rows (lines 91-93) ✅ |

## Requirements

1. `sonora get master-mute` sends `GET /api/v2/master-mute`; on success (200)
   it prints `{muted: bool}` (YAML by default, JSON with `--json`) and exits
   0. `master-mute` has no id and no collection: `sonora get
   master-mute/<anything>` and `sonora list master-mute` are usage errors
   (exit per `ClassUsage`), not silently treated as `get`/an empty list.
2. `sonora mute all` sends `PUT /api/v2/master-mute` with `{"muted": true}`;
   `sonora unmute all` sends the same with `{"muted": false}`. On success
   (200) each prints the updated `{muted: bool}` record and exits 0.
3. A non-2xx status or malformed body reuses `hub.ClassifyError` (`ClassHub`);
   a network/timeout failure maps to `ClassNetwork`. No stdout output on any
   failure path.
4. Any positional argument after `master-mute` (for `get`) or `all` (for
   `mute`/`unmute`) is a usage error on stderr (`ClassUsage`), matching the
   "unexpected argument(s)" pattern in `outputs.RunGet`.
5. `--json`, `--verbose`, and `--hub-url` behave the same as every other
   command.
6. No existing dispatch changes: `mute`/`unmute outputs|groups/<id>`
   ([outputs-groups-mute-command.md](outputs-groups-mute-command.md)) and
   `get`/`list inputs|outputs|groups|routes` behave exactly as before.

## Plan

1. Add contract tests in `tests/contract/master_mute_test.go` for
   `hub.GetMasterMute`/`SetMasterMute` (fail first).
2. Implement `hub.GetMasterMute`/`SetMasterMute` in
   `internal/hub/mastermute.go` and `render.RenderMasterMuteYAML`/`JSON` in
   `internal/render/mastermute.go`.
3. Add unit tests, then implement `mastermute.RunGet`/`RunMute`/`RunUnmute`
   in `internal/cli/mastermute/mastermute.go`, mirroring `outputs/get.go`
   and `outputs/muted.go` with zero required positional arguments.
4. Add integration test in `tests/integration/master_mute_test.go` covering
   all three commands plus the `master-mute/<id>`, `list master-mute`, and
   `all/foo` rejection cases, then wire the `master-mute`/`all` literals into
   `dispatchGetList`/`dispatchMuted` in `cmd/sonora/main.go`. Update
   `helpText`.
5. Update `README.md`'s command table and flip the three `master-mute` rows
   to ✅ in `docs/cli-command-landscape.md`.

## Tasks

- [x] Write contract tests for `hub.GetMasterMute`/`SetMasterMute`
- [x] Implement `hub.GetMasterMute`/`SetMasterMute` and
      `render.RenderMasterMuteYAML`/`JSON`
- [x] Write unit tests + implement `mastermute.RunGet`/`RunMute`/`RunUnmute`
- [x] Write integration test + wire `master-mute`/`all` into
      `dispatchGetList`/`dispatchMuted` in `cmd/sonora/main.go`, update
      `helpText`
- [x] Update `README.md` and `docs/cli-command-landscape.md`
- [x] Run `go test ./...`, `go vet ./...`, `gofmt -l .`

## Done When

- [x] All tasks checked off
- [x] `go test ./...` passes
- [x] `go vet ./...` and `gofmt -l .` clean
- [x] `sonora get master-mute`, `mute all`, `unmute all` verified end-to-end
      against a mock hub, including the non-2xx/malformed-body error path and
      the `master-mute/<id>`/`list master-mute`/`all/foo` rejection cases
