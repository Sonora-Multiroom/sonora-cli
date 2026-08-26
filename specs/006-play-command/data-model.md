# Phase 1 Data Model: Instant Playback (`play`)

## Entities

### PlaybackRequest (outbound, `internal/hub.PlaybackRequest`)

Mirrors `#/components/schemas/PlaybackRequest` in `api/openapi.json` field-for-field
(constitution Principle II).

| Field | JSON key | Go type | Required | Validation |
|---|---|---|---|---|
| URI | `uri` | `string` | yes | non-empty (FR-002); scheme/reachability validated by the hub, not the CLI (Assumptions) |
| TargetID | `targetId` | `string` | yes | non-empty (FR-002); resolved type, not raw user input alone, determines `TargetType` |
| TargetType | `targetType` | `string` enum `SINGLE_OUTPUT`\|`OUTPUT_GROUP` | yes | set by the CLI from target resolution (FR-003), never user-supplied directly |
| DisplayName | `displayName` | `*string` | no | omitted (`nil`) when `--name` not given; sent as given otherwise (FR-005) |
| Volume | `volume` | `*int` | no | omitted (`nil`) when `--volume` not given; range-checked `[0,100]` client-side before send (FR-004) |

### PlaybackResponse (inbound, `internal/hub.PlaybackResponse`)

Mirrors `#/components/schemas/PlaybackResponse` field-for-field. The nested `route` field
decodes into the existing `hub.Route` struct (`internal/hub/routes.go`) unchanged — full
ten-field validation (`validateRoute`-style) is out of scope for this shallow nested decode;
only `RouteID` and `Status` are required non-empty for the response to be considered
well-formed (FR-012), consistent with FR-006 only surfacing those two route fields.

| Field | JSON key | Go type | Required for well-formed response |
|---|---|---|---|
| InputID | `inputId` | `string` | non-empty |
| Route | `route` | `hub.Route` | `Route.RouteID` and `Route.Status` non-empty |
| Message | `message` | `string` | non-empty |

A response missing any required field is rejected as a `*hub.DecodeError` (FR-012), same
pattern as `validateRoute`/`ListGroups`'s inline shape checks.

### PlaybackResult (rendered view)

Not a distinct Go struct — `internal/render/play.go` reads `InputID`, `Route.RouteID`,
`Route.Status`, and `Message` directly off `hub.PlaybackResponse` and renders exactly those
four fields, per FR-006 ("the created input's identifier, the resulting route's identifier and
status ..., and a human-readable confirmation message") — the full nested `Route` (its other
eight fields: `InputID`, `TargetID`, `TargetType`, `CreatedAt`, `StartedAt`, `Transferable`,
`Pauseable`, `Paused`) is decoded (for contract fidelity and to catch a malformed response) but
deliberately not rendered, since the spec does not ask for it and rendering it would leak
internal route-management surface this feature explicitly excludes (Assumptions: "Managing the
resulting route/input is expected to be handled by future route/input management commands").

Rendered field order (both YAML and JSON): `inputId`, `routeId`, `status`, `message`.

### AmbiguousTargetError (client-side, `internal/hub.AmbiguousTargetError`)

| Field | Type | Purpose |
|---|---|---|
| ID | `string` | the target identifier that matched both an output and a group |

Not decoded from any hub response — constructed by `hub.ResolveTarget` when both `GetOutput`
and `GetGroup` succeed for the same ID and neither `--group` nor `--output` was given (FR-003a).

### APIError (client-side, `internal/hub.APIError`)

Decoded from `#/components/schemas/ErrorResponse` for `POST /api/v2/play`'s 400/422/502/503
responses (404 continues to use the existing `NotFoundError`, since it names a resource by ID
the same way `outputs`/`groups` 404s already do).

| Field | Source |
|---|---|
| StatusCode | HTTP response status |
| Title | `ErrorResponse.title` |
| Detail | `ErrorResponse.detail` |

If the body fails to decode as `ErrorResponse`, `Playback` falls back to the existing generic
`*StatusError{StatusCode}` so a malformed error body never crashes error handling.

## Target Resolution State

`hub.ResolveTarget(ctx, client, baseURL, targetID, forceGroup, forceOutput bool) (targetType
string, err error)`:

| Mode | Outputs lookup | Groups lookup | Both found | Only output found | Only group found | Neither found |
|---|---|---|---|---|---|---|
| default (`forceGroup=forceOutput=false`) | yes | yes | `AmbiguousTargetError` (FR-003a) | `SINGLE_OUTPUT` | `OUTPUT_GROUP` | `NotFoundError{Resource:"target"}` |
| `--group` (`forceGroup=true`) | no | yes | n/a | n/a | `OUTPUT_GROUP` | `NotFoundError{Resource:"group"}` (FR-003b, even if it exists as an output) |
| `--output` (`forceOutput=true`) | yes | no | n/a | `SINGLE_OUTPUT` | n/a | `NotFoundError{Resource:"output"}` (FR-003b, even if it exists as a group) |

`--group` and `--output` given together is rejected as `ClassUsage` before `ResolveTarget` is
ever called (FR-002a) — `play.Run` checks this immediately after `flag.Parse`.

## Exit Code Table (full, this feature's additions marked NEW)

| Exit code | Class | Meaning |
|---|---|---|
| 0 | (none) | success |
| 2 | `ClassUsage` | missing `<uri>`/`<target-id>`, both `--group` and `--output` given, unparseable flags |
| 3 | `ClassHub` | any other unexpected non-2xx hub response |
| 4 | `ClassNetwork` | hub unreachable or timed out |
| 5 | `ClassNotFound` | target (or forced type) does not exist |
| 6 | `ClassValidation` | **NEW** — out-of-range `--volume`, or hub `400` (bad URI/request field) |
| 7 | `ClassAmbiguous` | **NEW** — target ID matches both an output and a group |
| 8 | `ClassRouteFailed` | **NEW** — hub `422` (route creation failed) |
| 9 | `ClassSourceUnreachable` | **NEW** — hub `502` (URI unreachable) |
| 10 | `ClassServiceUnavailable` | **NEW** — hub `503` (upstream service unavailable) |
