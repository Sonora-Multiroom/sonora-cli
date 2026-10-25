# TinySpec: CLI for output volume, route volume and ducking level

**Branch**: feature/route-volume-cli
**Date**: 2026-10-25
**Status**: implemented, tag pending
**Complexity**: small

## What

Expose hub 034 in the CLI: `sonora play` replaces `--volume` with `--output-volume` and gains
`--route-volume` and `--ducking-level`; `sonora set routes/<id> volume N` changes a playing route; the
single-record YAML views show the new fields when set. Depends on
[hub-route-volume-client.md](hub-route-volume-client.md). Source: hub `specs/034-per-route-volume-ducking`;
this tinyspec replaces the backlog item it came from.

## Context

| File | Role |
|------|------|
| `internal/cli/play/play.go` | Modified: flags `--output-volume`, `--route-volume`, `--ducking-level`; `--volume` removed |
| `internal/cli/routes/volume.go` | New: `RunSetVolume`, copying `internal/cli/outputs/volume.go` with `hub.SetRouteVolume` |
| `cmd/sonora/main.go` | Modified: `dispatchSet` accepts `respath.Routes`; usage, error message, `helpText` line and example |
| `internal/render/routes.go` | Modified: `RenderRouteYAML` prints `routeVolume`, `duckingLevelDb` when non-nil |
| `internal/render/outputs.go` | Modified: `RenderOutputWithRoutesYAML` likewise for each route |
| `internal/render/inputs.go` | Modified: `RenderInputYAML` prints `defaultRouteVolume`, `defaultDuckingLevelDb` when non-nil |
| `tests/unit/cli_play_test.go`, `tests/integration/play_test.go` | Modified: the new flags and the removed `--volume` |
| `tests/unit/render_route_volume_test.go` | New: the YAML fields for routes, output routes and inputs |
| `tests/integration/groups_volume_test.go` | Modified: `routes/<id>` is no longer an unsupported `set` resource |
| `tests/unit/cli_routes_volume_test.go`, `tests/integration/routes_volume_test.go` | New, mirroring the outputs volume tests |
| `README.md`, `docs/cli-command-landscape.md` | Modified: play flags, `set routes` row |

## Requirements

1. `--output-volume N` (0–100) sends `outputVolume`. `--volume` is an unknown flag (exit 2), with no
   alias or notice.
2. `--route-volume N` (0–100) sends `routeVolume`. Out of range for either volume flag: the error names
   the flag, exit code of `ClassValidation`, no request sent.
3. `--ducking-level DB` (−60–0, decimals allowed, e.g. `--ducking-level -25`) sends `duckingLevelDb`.
   Without `--join-mode DUCK_OTHERS` it is a usage error (exit 2); out of range is `ClassValidation`.
   Both are checked before any request.
4. `sonora set routes/<route-id> volume <0-100>` calls `SetRouteVolume` and prints the returned route
   with `RenderRouteYAML`/`RenderRouteJSON`, exit 0. Missing id, missing or wrong attribute word,
   non-numeric or out-of-range value, extra arguments: same messages and exit codes as
   `set outputs/<id> volume`. Unknown route: not-found exit code, nothing on stdout.
5. `get routes/<id>` and the routes in `get outputs/<id>` print `routeVolume: N` and
   `duckingLevelDb: X` after `joinMode`, each only when non-nil; `get inputs/<id>` prints
   `defaultRouteVolume` and `defaultDuckingLevelDb` after `defaultJoinMode` likewise. A level prints
   in its shortest form (`-25`, `-12.5`). List views are unchanged.
6. `--json` output carries the fields as the `hub` types encode them (null when unset).

## Plan

1. Tests first: `cli_play_test.go` and `integration/play_test.go` for the flag rules and request bodies.
   Then `play.go`.
2. Tests first: `cli_routes_volume_test.go` and `integration/routes_volume_test.go`. Then
   `routes/volume.go` and `dispatchSet`.
3. Tests first: render tests with set, null and absent fields. Then the three YAML renderers.
4. README (play paragraph and example, `set` table row and paragraph), landscape doc, `helpText`.
   Release a tag noting the removed `--volume`.

## Tasks

- [x] Tests, then `--output-volume`, `--route-volume`, `--ducking-level` in `play.go`
- [x] Tests, then `internal/cli/routes/volume.go` and `routes` in `dispatchSet`
- [x] Tests, then the YAML renderers for routes, output routes and inputs
- [x] README, `docs/cli-command-landscape.md`, `helpText`
- [x] `gofmt -l .`, `go vet ./...`, `go test ./...`

## Done When

- [x] All tasks checked off; tests pass; no vet or gofmt findings
- [ ] Tag released with the `--volume` removal as a breaking change in its notes

## Verification (2026-10-25)

- Each group of tests failed before its code existed: the `play` flags (7 failing tests), the missing
  `routes.RunSetVolume` (build failure), and the YAML fields (4 failing tests).
- `gofmt -l .`: no output. `go vet ./...`: exit 0. `go test ./...`: exit 0, 1265 passing tests and
  subtests (1223 before).
- Built binary: `play --help` lists `--output-volume`, `--route-volume`, `--ducking-level` and no
  `--volume`; `play … --volume 10` is an unknown flag, exit 2, with no request sent; `set --help`
  names `routes`. `get outputs/office --include-routes` against `multiroom.lan` (read-only) prints the
  user's route without `routeVolume` or `duckingLevelDb`, which are null.
- `defaultJoinMode` keeps printing `null` when unset; the four new fields are left out when null, as
  this spec asks.
