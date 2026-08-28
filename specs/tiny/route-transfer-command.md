# TinySpec: `transfer routes` command

**Branch**: feature/route-transfer-command
**Date**: 2026-08-28
**Status**: draft
**Complexity**: small

## What

Add `sonora transfer routes/<route-id> <outputs|groups>/<target-id>`, calling
`POST /api/v2/routes/{routeId}/transfer` (operationId `transferRoute`) to seamlessly move an
active route's playback to a new output/group without interruption. The hub replaces the old
route with a new one, so the response (and rendered output) carries the *new* routeId. Follows
the existing two-positional-arg dispatch pattern already used by `sonora route`.

## Context

| File | Role |
|------|------|
| `internal/hub/routes.go` | Modified — add `TransferRequest` (targetId/targetType) and `TransferRoute`, mirroring `CreateRoute`'s error handling (200 success → decoded+validated `Route`, 404 → `NotFoundError`, 400/422 → `APIError` w/ `StatusError` fallback, other non-2xx → `StatusError`) |
| `internal/render/route.go` | Modified — reuse `RenderRouteCreatedYAML`/`RenderRouteCreatedJSON` (routeId/status/message) for the transfer result, same as `route`'s own output |
| `internal/cli/route/transfer.go` | New — `RunTransfer`, mirroring `route.go`'s two-positional-arg flag-parsing loop, resource-path validation, and error reporting |
| `cmd/sonora/main.go` | Modified — dispatch top-level `transfer` verb to `route.RunTransfer`; update `helpText` |
| `internal/hub/play.go` | Context only — reuses existing `ResolveTarget(ctx, client, baseURL, targetID, targetType)` to pre-check the target exists before calling the hub |
| `internal/hub/errors.go` | Context only — reuses existing `ClassNotFound`/`ClassValidation`/`ClassRouteFailed`/`ClassHub` (no new exit-code class needed); 400 → `ClassValidation` ("route not transferable"), 422 → `ClassRouteFailed` |
| `tests/contract/route_test.go` | Modified — add `TestTransferRoute_*` contract tests against a mock server (200/404/400/422/malformed-body/non-JSON-error-body), mirroring the existing `TestCreateRoute_*` tests |
| `tests/unit/cli_route_test.go` | Modified — unit tests for `RunTransfer` (missing route-id, missing target, bad target kind, unexpected args, not-found, success, `--json`) |
| `tests/integration/route_test.go` | Modified — end-to-end test that `transfer routes/<id> outputs/<id>` (and `groups/<id>`) dispatches correctly and renders the new routeId |
| `README.md` | Modified — document `sonora transfer routes/<route-id> <outputs\|groups>/<target-id>` |
| `docs/cli-command-landscape.md` | Modified — mark the `transfer routes` row done |

## Requirements

1. `sonora transfer routes/<route-id> outputs/<target-id>` (and `groups/<target-id>`) calls
   `POST /api/v2/routes/{routeId}/transfer` with body `{targetId, targetType}`; on success (200)
   it prints the *new* route's routeId/status/message (YAML by default, JSON with `--json`) and
   exits 0.
2. The first positional argument must be `routes/<id>` (or `rt/<id>`) and the second must be
   `outputs/<id>` or `groups/<id>` (`out`/`gr` aliases ok) — no auto-detect. A route path for
   the second argument, an inputs path for either, or a missing id in either segment is a usage
   error (exit per `ClassUsage`), matching `route.Run`'s target validation.
3. Before calling transfer, the target is resolved via the existing `hub.ResolveTarget` helper;
   a target-not-found 404 there reports "output/group not found: `<id>`" and exits with
   `ClassTargetNotFound`'s code (mirrors `route.Run`'s pre-check, not the hub's own 404 for an
   unknown routeId).
4. A 404 from the transfer call itself (unknown route-id) reports "route not found: `<id>`" and
   exits with `ClassNotFound`'s code; a 400 reports the hub's error detail (or "route is not
   transferable" fallback) and exits with `ClassValidation`'s code; a 422 reports the hub's
   error detail (or "route transfer failed" fallback) and exits with `ClassRouteFailed`'s code;
   other non-2xx/network/malformed-body failures reuse `hub.ClassifyError` — no stdout output on
   any failure path.
5. `--json`, `--verbose`, and `--hub-url` behave the same as every other route command.

## Plan

1. Add contract tests in `tests/contract/route_test.go` for `hub.TransferRoute` (fail first).
2. Implement `hub.TransferRequest` and `hub.TransferRoute(ctx, client, baseURL, routeID, req)
   (*Route, error)` in `internal/hub/routes.go`.
3. Add unit tests in `tests/unit/cli_route_test.go`, then implement `route.RunTransfer` in
   `internal/cli/route/transfer.go`, mirroring `route.go`'s two-arg parsing loop and reusing
   `hub.ResolveTarget` + `render.RenderRouteCreatedYAML`/`JSON`.
4. Add integration test in `tests/integration/route_test.go`, then wire `case "transfer":` in
   `cmd/sonora/main.go`'s `run()` calling `route.RunTransfer(args[1:], stdout, stderr)` directly
   (no respath pre-dispatch needed — `RunTransfer` parses both positional paths itself, same as
   `route.Run` does today). Update `helpText`.
5. Update `README.md`'s command table and `docs/cli-command-landscape.md`'s status marker.

## Tasks

- [ ] Write contract tests for `hub.TransferRoute` in `tests/contract/route_test.go`
- [ ] Implement `hub.TransferRequest`/`hub.TransferRoute` in `internal/hub/routes.go`
- [ ] Write unit tests + implement `route.RunTransfer` in `internal/cli/route/transfer.go`
- [ ] Write integration test + wire `transfer` dispatch in `cmd/sonora/main.go`, update `helpText`
- [ ] Update `README.md` and `docs/cli-command-landscape.md`
- [ ] Run `go test ./...`, `go vet ./...`, `gofmt -l .`

## Done When

- [ ] All tasks checked off
- [ ] `go test ./...` passes
- [ ] `go vet ./...` and `gofmt -l .` clean
- [ ] `sonora transfer routes/<id> outputs/<id>` and `.../groups/<id>` verified end-to-end
      against a mock hub, including target-not-found and route-not-found error paths
