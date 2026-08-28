# TinySpec: `delete routes` / `stop routes` command

**Branch**: main
**Date**: 2026-08-28
**Status**: done
**Complexity**: small

## What

Add `sonora delete routes/<route-id>` (and `sonora stop routes/<route-id>` as an exact alias)
calling `DELETE /api/v2/routes/{routeId}` (operationId `deleteRoute`), which stops playback and
removes the route. Follows the existing verb-dispatch pattern already used by `get`/`list`.

## Context

| File | Role |
|------|------|
| `internal/hub/routes.go` | Modified — add `DeleteRoute`, mirroring `GetRoute`/`CreateRoute`'s error handling (204 success, 404 → `NotFoundError`, 422 → `APIError` w/ `StatusError` fallback, other non-2xx → `StatusError`) |
| `internal/render/route.go` | Modified — add `RenderRouteDeletedYAML`/`RenderRouteDeletedJSON` (routeId/status/message, mirroring `RenderRouteCreated*`) |
| `internal/cli/routes/delete.go` | New — `RunDelete`, mirroring `get.go`'s flag-parsing/positional-arg/error-reporting shape |
| `cmd/sonora/main.go` | Modified — dispatch top-level `delete`/`stop` verbs (routes only) to `routes.RunDelete`; update `helpText` |
| `internal/hub/errors.go` | Context only — reuses existing `ClassNotFound`/`ClassRouteFailed`/`ClassHub` (no new exit-code class needed) |
| `tests/contract/route_test.go` | Modified — add `TestDeleteRoute_*` contract tests against a mock server (204/404/422/malformed-body/non-JSON-error-body), mirroring the existing `TestCreateRoute_*` tests |
| `tests/unit/render_route_test.go` | Modified — unit tests for `RenderRouteDeletedYAML`/`JSON` |
| `tests/unit/cli_routes_delete_test.go` | New — unit tests for `RunDelete` (missing id, unexpected args, not-found, success, `--json`) |
| `tests/integration/route_test.go` | Modified — end-to-end test that `delete routes/<id>` and `stop routes/<id>` both dispatch to the same behavior |
| `README.md` | Modified — document `sonora delete routes/<id>` / `sonora stop routes/<id>` |
| `docs/cli-command-landscape.md` | Modified — mark the `stop`/`delete routes` rows ✅ |

## Requirements

1. `sonora delete routes/<route-id>` calls `DELETE /api/v2/routes/{routeId}`; on success (204)
   it prints a routeId/status/message record (YAML by default, JSON with `--json`) and exits 0.
2. `sonora stop routes/<route-id>` behaves identically to `delete routes/<route-id>` in every
   respect (same handler, same output, same exit codes) — a true alias, not a separate path.
3. A 404 response reports "route not found: `<id>`" on stderr and exits with `ClassNotFound`'s
   code; a 422 reports the hub's error detail (or a generic "route stop failed" fallback) and
   exits with `ClassRouteFailed`'s code; other non-2xx/network/malformed-body failures reuse
   the existing `hub.ClassifyError` behavior — no stdout output on any failure path.
4. Missing `<route-id>` or unexpected extra arguments produce a usage error on stderr (exit
   code per `ClassUsage`), matching the pattern in `routes.RunGet`.
5. `--json`, `--verbose`, and `--hub-url` behave the same as every other routes command.
6. `delete`/`stop` given a non-`routes` resource path (e.g. `inputs/<id>`) is a usage error —
   no other resource supports delete yet, so don't silently no-op or panic.

## Plan

1. Add contract tests in `tests/contract/route_test.go` for `hub.DeleteRoute` (fail first).
2. Implement `hub.DeleteRoute(ctx, client, baseURL, routeID) error` in `internal/hub/routes.go`.
3. Add render unit tests, then implement `RenderRouteDeletedYAML`/`JSON` in
   `internal/render/route.go` (payload: routeId, status="stopped", message).
4. Add unit tests in `tests/unit/cli_routes_delete_test.go`, then implement
   `routes.RunDelete` in `internal/cli/routes/delete.go`, mirroring `get.go`'s structure.
5. Add integration test in `tests/integration/route_test.go` covering both `delete` and `stop`
   verbs end-to-end, then wire `case "delete", "stop":` in `cmd/sonora/main.go`'s `run()` —
   parse the resource path via `respath`, reject non-`routes` kinds as a usage error, and call
   `routes.RunDelete`. Update `helpText` with the new commands.
6. Update `README.md`'s command table and `docs/cli-command-landscape.md`'s ✅ markers.

## Tasks

- [x] Write contract tests for `hub.DeleteRoute` in `tests/contract/route_test.go`
- [x] Implement `hub.DeleteRoute` in `internal/hub/routes.go`
- [x] Write + implement `RenderRouteDeletedYAML`/`JSON` in `internal/render/route.go`
- [x] Write unit tests + implement `routes.RunDelete` in `internal/cli/routes/delete.go`
- [x] Write integration test + wire `delete`/`stop` dispatch (incl. non-routes rejection) in
      `cmd/sonora/main.go`, update `helpText`
- [x] Update `README.md` and `docs/cli-command-landscape.md`
- [x] Run `go test ./...`, `go vet ./...`, `gofmt -l .`

## Done When

- [x] All tasks checked off
- [x] `go test ./...` passes
- [x] `go vet ./...` and `gofmt -l .` clean
- [x] `sonora delete routes/<id>` and `sonora stop routes/<id>` verified identical via
      `TestRouteStop_IsIdenticalAliasOfDelete`, which runs the compiled binary end-to-end
      against a mock hub (`tests/integration/route_test.go`)
