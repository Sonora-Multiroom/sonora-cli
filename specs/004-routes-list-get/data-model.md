# Data Model: List and Get Audio Routes

**Feature**: `004-routes-list-get` | **Date**: 2026-08-25

## Route

New entity for this feature — maps field-for-field to
`#/components/schemas/RouteResponse` in `api/openapi.json`. See
[research.md §1](research.md#1-the-route-entity-and-its-split-listget-field-set).

| Field | Type | Source field (openapi.json) | Notes |
|-------|------|------------------------------|-------|
| `RouteID` | string | `routeId` | Unique identifier; required, non-empty. Shown by both `list` and `get`. |
| `InputID` | string | `inputId` | Source input identifier; required, non-empty. Shown by both `list` and `get`. |
| `TargetID` | string | `targetId` | Target output or group identifier; required, non-empty. Shown by both `list` and `get`. |
| `TargetType` | string | `targetType` | Enum `SINGLE_OUTPUT` \| `OUTPUT_GROUP`; required, must be one of these two values. Shown by both `list` and `get`. |
| `Status` | string | `status` | Enum `STARTING` \| `ACTIVE` \| `STOPPING` \| `STOPPED` \| `FAILED`; required, must be one of these five values. Shown by both `list` and `get`. |
| `CreatedAt` | string | `createdAt` | RFC 3339 timestamp; schema declares this required and non-nullable (unlike `Input.CreatedAt`). Display-only, never parsed. `get` only. |
| `StartedAt` | `*string` (nullable) | `startedAt` | RFC 3339 timestamp, or `nil` before playback starts. Display-only, never parsed. See [research.md §2](research.md#2-startedat-is-nullable-createdat-is-not). `get` only. |
| `Transferable` | bool | `transferable` | Whether the route can be transferred to another target. `get` only. |
| `Pauseable` | bool | `pauseable` | Whether the input driving this route supports pause/resume. `get` only. |
| `Paused` | bool | `paused` | Whether playback is currently paused. `get` only. |

**Validation rule (FR-016)**: a decoded body is treated as malformed (Hub-error exit class,
not partial/garbled output) if `RouteID`, `InputID`, or `TargetID` is empty, `TargetType` is
neither `"SINGLE_OUTPUT"` nor `"OUTPUT_GROUP"`, or `Status` is not one of `"STARTING"` /
`"ACTIVE"` / `"STOPPING"` / `"STOPPED"` / `"FAILED"`. See
[research.md §6](research.md#6-validation-rule-for-malformed-responses-fr-016) for why
`CreatedAt`/`StartedAt`/`Transferable`/`Pauseable`/`Paused` need no further validation beyond
their Go type.

**List decode**: identical shape to `001-list-outputs`'s `ListOutputs` — a `nil`/absent JSON
array decodes to an empty `[]Route{}`, not `nil`, so the "zero routes" case (FR-015) renders
consistently rather than as a JSON `null`.

**List vs. get field set**: `routes list` renders only `RouteID`/`InputID`/`TargetID`/
`TargetType`/`Status` (FR-004); `routes get` renders all ten fields (FR-007). This is the
first feature in this codebase where list and get intentionally show different field sets —
see [research.md §1](research.md#1-the-route-entity-and-its-split-listget-field-set).

## RoutesListQuery / RoutesGetQuery (request shaping)

Not persisted entities — the parameters these commands send to the hub.

| Field | Type | Maps to | Default |
|-------|------|---------|---------|
| `Status` | string | `status` query parameter on `GET /api/v2/routes` | `""` (unset — no filter) |
| `InputID` | string | `inputId` query parameter on `GET /api/v2/routes` | `""` (unset — no filter) |
| `TargetID` | string | `targetId` query parameter on `GET /api/v2/routes` | `""` (unset — no filter) |
| `RouteID` | string | `{routeId}` path segment on `GET /api/v2/routes/{routeId}` | none — required, sourced from `routes get`'s positional argument |

An unset filter is never sent as an empty query parameter (e.g. `status=`) — it is simply
omitted from the request, so the hub's own "no filter supplied" behavior applies (see
[research.md §3](research.md#3-http-calls-get-apiv2routes-listroutes-and-get-apiv2routesrouteid-getroute)).
When more than one filter is set, the hub combines them with AND logic (FR-003); the CLI
performs no client-side combination or validation of filter values.

## CLI invocation shape

See [contracts/cli-routes-list.md](contracts/cli-routes-list.md) and
[contracts/cli-routes-get.md](contracts/cli-routes-get.md) for the full contracts.

| Command | Argument/Flag | Type | Default | Effect |
|---|------|------|---------|--------|
| `routes list` | `--status` | string | `""` (no filter) | Only return routes with this status (FR-003). Invalid values are rejected by the hub (`400` → exit `3`), not validated client-side. |
| `routes list` | `--input-id` | string | `""` (no filter) | Only return routes sourced from this input identifier (FR-003). |
| `routes list` | `--target-id` | string | `""` (no filter) | Only return routes pointed at this target identifier (FR-003). |
| `routes list` | `--json` | bool | `false` | Switch rendering from default YAML to JSON (FR-009). |
| `routes list` | `--verbose` | bool | `false` | On failure, additionally print the underlying error detail (Principle IV). |
| `routes list` | `--hub-url` | string | See `outputs list`'s precedence | Overrides the hub base URL. |
| `routes get` | `<route-id>` (positional) | string | none, required | Selects the route to fetch (FR-006). Missing → usage error (exit `2`). |
| `routes get` | `--json` | bool | `false` | Switch rendering from default YAML to JSON (FR-009). |
| `routes get` | `--verbose` | bool | `false` | On failure, additionally print the underlying error detail (Principle IV). |
| `routes get` | `--hub-url` | string | See `outputs get`'s precedence | Overrides the hub base URL. |

Unlike `outputs list`/`inputs list`, `routes list` has no `--include-disabled` equivalent —
routes have no enabled/disabled concept (FR-002); by default all routes are returned
regardless of status, narrowed only by the three optional filters above.

## Config file

Unchanged from prior features; both commands read the same
`~/.config/sonora/config.json` through the existing `config.ResolveHubURL`, with no new
fields.

## Exit code classes (unchanged from `002-outputs-get`/`003-inputs-list-get`)

| Code | Meaning | New in this feature? |
|------|---------|----------------------|
| `0` | Success | No |
| `2` | Usage error (bad/unknown flag, missing/extra positional argument) | No |
| `3` | Hub-reported error (non-2xx other than 404 — including an invalid `--status` filter's `400` — or malformed response body) | No |
| `4` | Network/connectivity error (unreachable host, DNS failure, timeout) | No |
| `5` | Not found — hub responded `404` for the given identifier (`routes get` only) | No (reused, not renumbered — see [research.md §5](research.md#5-exit-code-scheme-unchanged-no-new-class-needed)) |
