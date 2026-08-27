# Phase 1 Data Model: Route an Existing Input (`route`)

## Entities

### CreateRouteRequest (outbound, `internal/hub.CreateRouteRequest`, NEW)

Mirrors `#/components/schemas/CreateRouteRequest` in `api/openapi.json` field-for-field
(constitution Principle II).

| Field | JSON key | Go type | Required | Validation |
|---|---|---|---|---|
| InputID | `inputId` | `string` | yes | non-empty (FR-002); the identifier segment of the parsed `inputs/<id>` path |
| TargetID | `targetId` | `string` | yes | non-empty (FR-002); the identifier segment of the parsed `outputs\|groups/<id>` path |
| TargetType | `targetType` | `string` enum `SINGLE_OUTPUT`\|`OUTPUT_GROUP` | yes | set by the CLI from the target path's prefix (FR-003), never guessed |

### RouteResponse (inbound, reuses existing `internal/hub.Route`, unchanged)

`createRoute`'s 201 body mirrors `#/components/schemas/RouteResponse` — the same schema `routes
get`/`routes list` already decode into `hub.Route` (`internal/hub/routes.go`). No new Go type is
introduced; the existing `validateRoute` helper (routeId/inputId/targetId non-empty, targetType
and status in their enum sets) is reused unchanged to satisfy FR-011 ("reject a routing result
that does not match the expected structure").

### RoutingResult (rendered view, NEW — `internal/render/route.go`)

Not decoded from the hub — combines the decoded `hub.Route` with a CLI-constructed confirmation
message (research.md §4). Rendered field order (both YAML and JSON), per FR-005's "the
resulting route's identifier and status ..., and a human-readable confirmation message":

| Field | JSON key | Source |
|---|---|---|
| RouteID | `routeId` | `hub.Route.RouteID`, exactly as returned by the hub |
| Status | `status` | `hub.Route.Status`, exactly as returned by the hub |
| Message | `message` | constructed by `internal/cli/route`: `fmt.Sprintf("Routed %s to %s.", inputArg, targetArg)` using the two resource-path arguments as typed (e.g. `"Routed inputs/spotify-1 to outputs/office-speaker."`) |

The full `hub.Route` (createdAt/startedAt/transferable/pauseable/paused/targetId/targetType) is
decoded and validated but deliberately not rendered — same reasoning `play`'s
`PlaybackResult` used: the spec does not ask for those fields, and a future `routes get
<routeId>` is the documented follow-up for inspecting them (spec Assumptions).

## Pre-check Sequencing

`route.Run` performs, in order, before ever calling `createRoute`:

1. `hub.GetInput(ctx, client, baseURL, inputID)` — not found → "input not found" (FR-004),
   exit 11 (`ClassInputNotFound`, research.md §3).
2. `hub.ResolveTarget(ctx, client, baseURL, targetID, targetType)` — not found → "target not
   found" (FR-003a), exit 12 (`ClassTargetNotFound`).

Only if both succeed does `route.Run` call `hub.CreateRoute`. This ordering makes the outcome
deterministic when both the input and the target are missing (input is reported first), and
guarantees `createRoute` is never called against a target the CLI already knows doesn't exist
(unlike `play`, which has no analogous "does the input exist" step since `play` mints the input
itself).

## `hub.CreateRoute` Response Handling

`hub.CreateRoute(ctx, client, baseURL, req CreateRouteRequest) (*Route, error)`
(`internal/hub/routes.go`, NEW function, alongside `ListRoutes`/`GetRoute`):

| HTTP status | Outcome |
|---|---|
| 201 | Decode into `hub.Route`, validate via existing `validateRoute`; malformed body → `*DecodeError` (FR-011) |
| 400 | Decode body as `errorResponse` (existing unexported type, `internal/hub/play.go`) → `*APIError{StatusCode:400,...}` |
| 404 | `*NotFoundError{Resource: "target", ID: req.TargetID}` — a same-shape fallback to `Playback`'s own 404 handling for the rare race where the target (or input) vanishes between the pre-checks above and this call; the pre-checks are what give FR-003a/FR-004 their real distinction, this is a backstop |
| 422 | Decode body as `errorResponse` → `*APIError{StatusCode:422,...}` |
| other non-2xx | `*StatusError{StatusCode}` |

## Exit Code Table (full, this feature's additions marked NEW)

| Exit code | Class | Meaning |
|---|---|---|
| 0 | (none) | success |
| 2 | `ClassUsage` | missing `<input-path>`/`<target-path>`, invalid/unrecognized path prefix on either argument, unparseable flags |
| 3 | `ClassHub` | any other unexpected non-2xx hub response, or a malformed 201 body (FR-011) |
| 4 | `ClassNetwork` | hub unreachable or timed out (from any of the three calls) |
| 6 | `ClassValidation` | hub `400` on `createRoute` (bad request field) |
| 8 | `ClassRouteFailed` | hub `422` on `createRoute` (route creation failed, e.g. disabled input, target at capacity) |
| 11 | `ClassInputNotFound` | **NEW** — `GetInput` 404, or `createRoute`'s 404 fallback naming `"input"` |
| 12 | `ClassTargetNotFound` | **NEW** — `ResolveTarget` 404 (`"output"`/`"group"`), or `createRoute`'s 404 fallback |

`ClassInputNotFound`/`ClassTargetNotFound` are resolved locally in `route.Run` (research.md §3)
and are not reachable from any other command; `hub.ClassifyError`'s existing `ClassNotFound`
(exit 5) mapping is unchanged for every other command.
