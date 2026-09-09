# TinySpec: `stop routes`/`stop outputs/<id>`/`stop groups/<id>` commands

**Branch**: feature/stop-routes-command
**Date**: 2026-09-09
**Status**: done
**Complexity**: small

## What

Wire up the three new bulk stop-routes endpoints already present in
`api/openapi.json`: `DELETE /api/v2/routes` (`stopAllRoutes`), `DELETE
/api/v2/outputs/{outputId}/routes` (`stopRoutesForOutput`), `DELETE
/api/v2/groups/{groupId}/routes` (`stopRoutesForGroup`). All three return
`BulkStopResponse` (`{stoppedCount, stoppedRoutes[]}`). Expose them as
`sonora stop routes` (bare, no id — system-wide), `sonora stop
outputs/<output-id>`, and `sonora stop groups/<group-id>`. The existing
`sonora stop routes/<route-id>` (single-route alias of `delete
routes/<route-id>`) is unchanged; `stop` stops being a routes-only alias and
gains two new resource kinds, so its dispatch moves out of `dispatchDelete`
into a new `dispatchStop`.

## Context

| File | Role |
|------|------|
| `api/openapi.json` | Context only — already updated (uncommitted diff) with the three new operations and `BulkStopResponse`/`StoppedRouteEntry`/`StopReason` schemas |
| `internal/hub/routes.go` | Modified — add `BulkStopResponse{StoppedCount int; StoppedRoutes []StoppedRouteEntry}`, `StoppedRouteEntry{RouteID, TargetType, TargetID string; StopReason *string}`, a shared decode helper (nil `StoppedRoutes` → `[]`), and `StopAllRoutes(ctx, client, baseURL)` — `DELETE /api/v2/routes`, 200-only (no documented 404/400) |
| `internal/hub/outputs.go` | Modified — add `StopRoutesForOutput(ctx, client, baseURL, outputID)` — `DELETE /api/v2/outputs/{id}/routes`, 404 → `NotFoundError{"output"}` |
| `internal/hub/groups.go` | Modified — add `StopRoutesForGroup(ctx, client, baseURL, groupID)` — `DELETE /api/v2/groups/{id}/routes`, 404 → `NotFoundError{"group"}` |
| `internal/render/routes.go` | Modified — add `RenderBulkStopYAML`/`JSON(hub.BulkStopResponse)`, mirroring `RenderRoutesYAML`/`JSON`'s list shape (`stoppedCount` + `stoppedRoutes[]`, with a `writeStopReason` helper mirroring `writeStartedAt` for the nullable field) |
| `internal/cli/routes/stop.go` | New — `RunStopAll`, mirroring `mastermute.RunMute`'s zero-required-positional-argument shape, calling `hub.StopAllRoutes` |
| `internal/cli/outputs/stop.go` | New — `RunStop`, mirroring `outputs.RunEnable`'s one-required-positional-argument shape, calling `hub.StopRoutesForOutput` |
| `internal/cli/groups/stop.go` | New — `RunStop`, mirroring `groups.RunEnable`'s shape, calling `hub.StopRoutesForGroup` |
| `cmd/sonora/main.go` | Modified — remove `"stop"` from `case "delete", "stop":`, add `case "stop": return dispatchStop(...)`; new `dispatchStop` handles `routes` (bare → `RunStopAll`, `/id` → existing `routes.RunDelete` alias), `outputs/<id>` → `outputs.RunStop`, `groups/<id>` → `groups.RunStop`; `inputs` and bare `outputs`/`groups` (no id) are usage errors; update `helpText`'s command table + examples |
| `README.md` | Modified — update the `stop routes/<id>` section; no longer describe `stop` as a routes-only alias |
| `docs/cli-command-landscape.md` | Modified — add three new ✅ rows under `## routes` |
| `tests/contract/routes_stop_test.go`, `outputs_stop_test.go`, `groups_stop_test.go` | New — contract tests for the three hub functions (200, 404 where applicable, malformed body) |
| `tests/unit/cli_routes_stop_test.go`, `cli_outputs_stop_test.go`, `cli_groups_stop_test.go` | New — unit tests for the three `Run*` functions |
| `cmd/sonora/main_test.go` | Modified — dispatch-routing tests for `stop` live alongside the existing `get`/`list` dispatch tests in this file (not a separate `tests/unit` file), matching how `dispatchGetList`/`dispatchEnabled` routing is already tested here |
| `tests/integration/stop_routes_test.go` | New — end-to-end dispatch coverage for `stop routes`, `stop outputs/<id>`, `stop groups/<id>`, plus rejection of `stop inputs/<id>`, `stop outputs` (bare), `stop groups` (bare) |

## Requirements

1. `sonora stop routes` sends `DELETE /api/v2/routes`; on 200 it prints
   `{stoppedCount, stoppedRoutes[]}` (YAML by default, JSON with `--json`)
   and exits 0, including the zero-stopped case.
2. `sonora stop outputs/<output-id>` sends `DELETE
   /api/v2/outputs/<output-id>/routes`; a 404 is `ClassNotFound` (exit 5);
   success prints the same `BulkStopResponse` shape.
3. `sonora stop groups/<group-id>` sends `DELETE
   /api/v2/groups/<group-id>/routes`; same 404/success handling as outputs.
4. `sonora stop routes/<route-id>` keeps behaving exactly as today (alias of
   `delete routes/<route-id>`, single-route stop) — unaffected by this
   change.
5. Any unexpected positional argument is a usage error (`ClassUsage`, exit
   2); `stop outputs`/`stop groups` with no id, and `stop inputs/<id>`, are
   also usage errors.
6. `--json`, `--verbose`, and `--hub-url` behave the same as every other
   command.

## Plan

1. Add contract tests, then implement `hub.StopAllRoutes`/
   `StopRoutesForOutput`/`StopRoutesForGroup` plus `BulkStopResponse`/
   `StoppedRouteEntry` in `internal/hub/{routes,outputs,groups}.go`.
2. Implement `render.RenderBulkStopYAML`/`JSON` in `internal/render/routes.go`.
3. Add unit tests, then implement `routes.RunStopAll`, `outputs.RunStop`,
   `groups.RunStop`.
4. Add unit tests for main.go routing, then implement `dispatchStop` in
   `cmd/sonora/main.go`, removing `"stop"` from `dispatchDelete`'s case and
   updating `helpText`.
5. Add integration test covering all three new commands plus their rejection
   cases.
6. Update `README.md` and `docs/cli-command-landscape.md`.
7. Run `go test ./...`, `go vet ./...`, `gofmt -l .`.

## Tasks

- [x] Contract tests + `hub.StopAllRoutes`/`StopRoutesForOutput`/`StopRoutesForGroup` + `BulkStopResponse`/`StoppedRouteEntry`
- [x] `render.RenderBulkStopYAML`/`JSON`
- [x] Unit tests + `routes.RunStopAll`, `outputs.RunStop`, `groups.RunStop`
- [x] `dispatchStop` in `cmd/sonora/main.go` (+ `helpText`) + its dispatch tests in `cmd/sonora/main_test.go`
- [x] Integration test for `stop routes`/`stop outputs/<id>`/`stop groups/<id>` and rejection cases
- [x] Update `README.md` and `docs/cli-command-landscape.md`
- [x] Run `go test ./...`, `go vet ./...`, `gofmt -l .`

## Done When

- [x] All tasks checked off
- [x] `go test ./...` passes
- [x] `go vet ./...` and `gofmt -l .` clean
- [x] `stop routes`, `stop outputs/<id>`, `stop groups/<id>` verified
      end-to-end against a mock hub, including 404 and zero-stopped paths,
      and existing `stop routes/<id>` still works unchanged
