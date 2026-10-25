# Adopt hub 034: `outputVolume`, `--route-volume` and `--ducking-level`

- **Status:** Proposed, not started. Written 2026-10-24.
- **Origin:** hub `specs/034-per-route-volume-ducking` (`multiroom-api` 0.1.29).
- **Depends on:** the hub release that carries 034. Re-fetch `api/openapi.json` first.
- **Effort:** Small to medium.

## Problem

The hub's `POST /api/v2/play` now names two volumes: `outputVolume` sets the room and persists (`volume`
is its deprecated alias), `routeVolume` is the volume of this playback only. Routes and inputs also
carry `routeVolume`, `duckingLevelDb`, `defaultRouteVolume` and `defaultDuckingLevelDb`, and
`PUT /api/v2/routes/{id}/volume` changes a playing route. The `hub` package knows none of this, and
`sonora-mcp` is built on it.

## Proposal

1. Re-fetch `api/openapi.json` (published as `api.Spec`) and note the hub version it was fetched at.
2. The `hub` package sends `outputVolume` where it sent `volume`, and gains `RouteVolume` and
   `DuckingLevelDb` on play and route creation, the new response fields, and `SetRouteVolume`.
3. `sonora play` replaces `--volume` with `--output-volume`, with no alias or notice: `--volume` is
   removed and now fails as an unknown flag. It also gains `--route-volume` and `--ducking-level` (the
   latter only with a ducking join mode).
4. `sonora set routes/<route-id> volume <0-100>` changes a playing route's volume through
   `SetRouteVolume`, alongside `set outputs/<id> volume` and `set groups/<id> volume`.
5. The YAML views show the new fields only when the hub sends a non-null value, the way `joinMode`
   and `defaultJoinMode` are handled: `routeVolume` and `duckingLevelDb` in `get routes/<id>` and in
   the routes listed by `get outputs/<id>`, and `defaultRouteVolume` and `defaultDuckingLevelDb` in
   `get inputs/<id>`. List views (`get routes`, `get inputs`) keep their fixed fields. The `--json`
   views carry the new fields as soon as they are on the `hub` types.
6. Release a tag; `sonora-mcp` then bumps it.

## Open questions

- None.
