# TinySpec: `pause`/`resume routes/<route-id>` commands

**Branch**: feature/route-pause-resume-command
**Date**: 2026-08-29
**Status**: draft
**Complexity**: small

## What

Add `sonora pause routes/<route-id>` and `sonora resume routes/<route-id>`, both calling
`PUT /api/v2/routes/{routeId}/pause` (operationId `setPauseState`) with `{"paused": true}` or
`{"paused": false}` respectively. The hub is idempotent — setting the same state twice still
returns 200 with the current state — so no client-side pre-check of the current `paused` value is
needed. Follows the existing single-resource-path dispatch pattern already used by
`delete routes/<id>` and `enable`/`disable inputs/<id>`.

## Context

| File | Role |
|------|------|
| `internal/hub/routes.go` | Modified — add `PauseRequest` (`paused bool`) and `SetPauseState(ctx, client, baseURL, routeID string, paused bool) (*Route, error)`, mirroring `TransferRoute`'s error handling but with only the two status codes the spec defines: 200 → decoded+validated `Route`, 404 → `NotFoundError{Resource: "route", ID: routeID}`, 400 → decode `errorResponse` into `APIError` falling back to `StatusError`, other non-2xx → `StatusError` |
| `internal/render/route.go` | Modified — add `RenderRoutePauseYAML`/`RenderRoutePauseJSON`, a new flat payload exposing `routeId`/`paused`/`status`/`message` (paused matters here, unlike `RenderRouteCreatedYAML`'s routeId/status/message), shared by both `pause` and `resume` (message text differs) |
| `internal/cli/route/pause.go` | New — `RunPause`/`RunResume` (each a thin wrapper setting `paused` true/false) over a shared `runSetPause(verb string, paused bool, args, stdout, stderr) int`, copying `inputs.runSetEnabled`'s flag-parsing/single-positional-arg/error-reporting shape verbatim except for the hub/render calls and usage string |
| `cmd/sonora/main.go` | Modified — add a `dispatchPause(verb string, args, stdout, stderr)` mirroring `dispatchEnabled`, restricted to `respath.Routes`, dispatching `pause`/`resume` to `route.RunPause`/`route.RunResume`; wire `case "pause", "resume":` in `run()`; update `helpText` and its Examples |
| `internal/hub/errors.go` | Context only — reuses existing `ClassNotFound`/`ClassValidation`/`ClassHub`/`ClassNetwork`, no new exit-code class needed (pause has no 422 case, unlike transfer) |
| `tests/contract/route_test.go` | Modified — add `TestSetPauseState_*` contract tests against a mock server (200 pause, 200 resume, 404, 400, malformed body, non-JSON error body), mirroring the existing `TestTransferRoute_*` tests |
| `tests/unit/cli_route_test.go` | Modified — unit tests for `RunPause`/`RunResume` (missing route-id, non-routes resource, unexpected args, not-found, success, `--json`) |
| `tests/integration/route_test.go` | Modified — end-to-end tests that `pause routes/<id>` and `resume routes/<id>` dispatch correctly, send the right request body (`{"paused":true}`/`{"paused":false}`), and reject non-routes resources as a usage error |
| `README.md` | Modified — document `sonora pause routes/<route-id>` and `sonora resume routes/<route-id>` |
| `docs/cli-command-landscape.md` | Modified — mark the `pause routes`/`resume routes` rows (lines 74-75) done |

## Requirements

1. `sonora pause routes/<route-id>` sends `PUT /api/v2/routes/{routeId}/pause` with
   `{"paused": true}`; `sonora resume routes/<route-id>` sends the same endpoint with
   `{"paused": false}`. On success (200) each prints the returned route's
   `routeId`/`paused`/`status` plus a message (YAML by default, JSON with `--json`) and exits 0.
2. Only `routes/<id>` (or `rt/<id>`) is accepted; any other resource kind, or a path missing an
   id, is a usage error on stderr (no network call), exit code per `ClassUsage` — mirroring
   `dispatchEnabled`'s restriction to `inputs`.
3. A 404 response reports "route not found: `<id>`" and exits with `ClassNotFound`'s code; a 400
   (route not active, input not pauseable, or other validation failure) reports the hub's error
   detail (or a generic fallback) and exits with `ClassValidation`'s code; other non-2xx/network/
   malformed-body failures reuse the existing `hub.ClassifyError` behavior — no stdout output on
   any failure path.
4. Calling `pause` on an already-paused route, or `resume` on an already-active route, is not
   treated as an error — the hub returns 200 with the current state (idempotent per the OpenAPI
   description) and the CLI reports success either way.
5. `--json`, `--verbose`, and `--hub-url` behave the same as every other command.

## Plan

1. Add contract tests in `tests/contract/route_test.go` for `hub.SetPauseState` (fail first).
2. Implement `hub.PauseRequest` and `hub.SetPauseState` in `internal/hub/routes.go`.
3. Add render unit tests, then implement `RenderRoutePauseYAML`/`RenderRoutePauseJSON` in
   `internal/render/route.go`.
4. Add CLI unit tests in `tests/unit/cli_route_test.go`, then implement `RunPause`/`RunResume` +
   shared `runSetPause` in `internal/cli/route/pause.go`, copying `inputs.runSetEnabled`'s shape.
5. Add/extend integration tests in `tests/integration/route_test.go`, then add `dispatchPause` in
   `cmd/sonora/main.go` and wire `case "pause", "resume":` in `run()`. Update `helpText`.
6. Update `README.md`'s command table and `docs/cli-command-landscape.md`'s status markers.

## Tasks

- [ ] Write contract tests for `hub.SetPauseState` in `tests/contract/route_test.go`
- [ ] Implement `hub.PauseRequest`/`hub.SetPauseState` in `internal/hub/routes.go`
- [ ] Write render tests + implement `RenderRoutePauseYAML`/`RenderRoutePauseJSON` in
      `internal/render/route.go`
- [ ] Write unit tests + implement `RunPause`/`RunResume`/`runSetPause` in
      `internal/cli/route/pause.go`
- [ ] Write/extend integration tests + add `dispatchPause` + wire `pause`/`resume` dispatch in
      `cmd/sonora/main.go`, update `helpText`
- [ ] Update `README.md` and `docs/cli-command-landscape.md`
- [ ] Run `go test ./...`, `go vet ./...`, `gofmt -l .`

## Done When

- [ ] All tasks checked off
- [ ] `go test ./...` passes
- [ ] `go vet ./...` and `gofmt -l .` clean
- [ ] `sonora pause routes/<id>` and `sonora resume routes/<id>` verified end-to-end against a
      mock hub, including not-found/validation error paths and the idempotent-repeat case
