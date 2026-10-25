# TinySpec: `hub` package support for route volume and ducking level (hub 0.1.29)

**Branch**: feature/hub-route-volume-client
**Date**: 2026-10-25
**Status**: implemented, tag pending
**Complexity**: small

## What

Hub 034 (`multiroom-api` 0.1.29, live on `multiroom.lan` since 2026-10-24) adds `outputVolume`,
`routeVolume` and `duckingLevelDb` to play and route creation, matching response fields, and
`PUT /api/v2/routes/{routeId}/volume`. Bring `api/openapi.json` and the `hub` package up to that contract
and release a tag so `sonora-mcp` can bump it. The CLI changes follow in
[route-volume-cli.md](route-volume-cli.md). Source: `docs/backlog/adopt-route-volume.md` steps 1, 2, 6.

## Context

| File | Role |
|------|------|
| `api/openapi.json` | Re-fetched with `scripts/update-openapi.sh` (0.1.28 → 0.1.29) |
| `hub/play.go` | Modified: `PlaybackRequest.Volume` → `OutputVolume` (`outputVolume`); add `RouteVolume *int`, `DuckingLevelDb *float64` |
| `hub/routes.go` | Modified: `Route` and `CreateRouteRequest` gain `RouteVolume`, `DuckingLevelDb`; add `SetRouteVolume` |
| `hub/outputs.go` | Modified: `RouteOnOutput` gains `RouteVolume`, `DuckingLevelDb`. `SetOutputVolume` is the model for `SetRouteVolume` |
| `hub/inputs.go` | Modified: `Input` and `CreateInputRequest` gain `DefaultRouteVolume *int`, `DefaultDuckingLevelDb *float64` |
| `tests/contract/{play,route,outputs_routes,inputs_get,inputs_create}_test.go` | Modified: new fields encoded and decoded |
| `tests/contract/routes_volume_test.go` | New: `SetRouteVolume`, mirroring `outputs_volume_test.go` |
| `internal/cli/play/play.go` | Modified only so the build holds: assigns `req.OutputVolume` instead of `req.Volume` |
| `tests/integration/play_test.go`, `tests/unit/{render_routes,render_output_get}_test.go` | Modified: `--volume` now sends `outputVolume`; the JSON views carry the two new keys (null when unset) |

## Requirements

1. `api/openapi.json` is the hub's 0.1.29 document, LF line endings; `go test ./tests/unit/` (`api.Spec`)
   passes.
2. `PlaybackRequest` sends `outputVolume`, never `volume`; `routeVolume` and `duckingLevelDb` are sent
   only when non-nil. `CreateRouteRequest` and `CreateInputRequest` likewise omit nil new fields.
3. `Route`, `RouteOnOutput` and `Input` decode the new fields as pointers: null or absent (a pre-0.1.29
   hub) → nil, `-25.0` → `-25`.
4. `SetRouteVolume(ctx, client, baseURL, routeID string, volume int) (*Route, error)` sends
   `PUT /api/v2/routes/{routeId}/volume` with `{"volume": N}`. 200 → decoded `*Route` (validated like
   `GetRoute`); 404 → `*NotFoundError{Resource: "route", ID: routeID}`; 400 → `*APIError` with the hub's
   detail (fallback `*StatusError`); malformed body → `*DecodeError`; other non-2xx → `*StatusError`.
5. Every new exported identifier has godoc that stands without this repo, including that a route
   volume cannot be cleared and that `duckingLevelDb` is only for `DUCK_OTHERS`.
6. Renaming `Volume` is a breaking change to the public package; the commit and tag notes say so.

## Plan

1. Run `scripts/update-openapi.sh`; check the diff covers only 034's additions.
2. Contract tests first for the request encoding, response decoding and `SetRouteVolume`; then the
   type changes and `SetRouteVolume`.
3. Fix the one `req.Volume` reference in `internal/cli/play/play.go` (no behaviour change: `--volume`
   still exists until the CLI tinyspec).
4. Release: tag `v0.x.(y+1)` with the breaking-change note; `sonora-mcp` bumps it in its own repo.

## Tasks

- [x] Refresh `api/openapi.json` to 0.1.29
- [x] Contract tests: `outputVolume`/`routeVolume`/`duckingLevelDb` on play, route volume and ducking level on route creation, input defaults on input creation
- [x] Contract tests: new response fields on `Route`, `RouteOnOutput`, `Input` (set, null, absent)
- [x] Contract tests in `tests/contract/routes_volume_test.go` (200, 400, 404, malformed body, path and method)
- [x] Update `hub/play.go`, `hub/routes.go`, `hub/outputs.go`, `hub/inputs.go`; add `SetRouteVolume`
- [x] Rename `req.Volume` in `internal/cli/play/play.go`
- [x] `gofmt -l .`, `go vet ./...`, `go test ./...`

## Done When

- [x] All tasks checked off; tests pass; no vet or gofmt findings
- [ ] Tag released with the `PlaybackRequest.Volume` → `OutputVolume` breaking change in its notes

## Verification (2026-10-25)

- `scripts/update-openapi.sh`: 0.1.28 → 0.1.29. Changed: the new `setRouteVolume` path, and the new
  properties on `PlaybackRequest`, `CreateRouteRequest`, `CreateInputRequest`, `RouteResponse`,
  `RouteOnOutputResponse`, `InputResponse` and `SpeakRequest`. `SpeakRequest`'s `routeVolume` and
  `duckingLevelDb` are not adopted here: `hub.SpeakRequest` already lags that schema.
- New contract tests failed to compile before the `hub` changes (`unknown field DefaultRouteVolume`).
- `gofmt -l .`: no output. `go vet ./...`: exit 0. `go test ./...`: exit 0.
