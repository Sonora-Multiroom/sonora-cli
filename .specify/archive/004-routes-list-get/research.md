# Research: List and Get Audio Routes

**Feature**: `004-routes-list-get` | **Date**: 2026-08-25

This document resolves every open technical question for implementing `sonora routes list`
and `sonora routes get <route-id>` before design (Phase 1) begins. The spec has no unresolved
`NEEDS CLARIFICATION` markers — this feature deliberately reuses every foundational decision
`001-list-outputs`, `002-outputs-get`, and `003-inputs-list-get` already made (language,
dependency policy, HTTP client construction, hub URL resolution, YAML/JSON rendering approach,
`--verbose`/error-translation shape, exit-code scheme, testing strategy, the generalized
`NotFoundError{Resource, ID}`). Only the decisions genuinely new to this feature — driven by
the `Route` entity's shape and its three list filters — are recorded below.

## 1. The `Route` entity and its split list/get field set

- **Decision**: Add a new `Route` struct in `internal/hub/routes.go`, mapping field-for-field
  to `#/components/schemas/RouteResponse` in `api/openapi.json`: `RouteID` (`routeId`),
  `InputID` (`inputId`), `TargetID` (`targetId`), `TargetType` (`targetType`, string enum
  `SINGLE_OUTPUT`/`OUTPUT_GROUP`), `Status` (`status`, string enum
  `STARTING`/`ACTIVE`/`STOPPING`/`STOPPED`/`FAILED`), `CreatedAt` (`createdAt`, plain
  `string` — the schema declares this required and non-nullable, unlike `startedAt`),
  `StartedAt` (`startedAt`, `*string`, nullable — absent until playback begins),
  `Transferable` (`transferable`, bool), `Pauseable` (`pauseable`, bool), `Paused` (`paused`,
  bool). This is one struct decoded from both endpoints, but rendered through two different
  views: `routes list` shows only `routeId`/`inputId`/`targetId`/`targetType`/`status`
  (FR-004), while `routes get` shows all ten fields (FR-007).
- **Rationale**: Principle II requires request/response types to be generated from or
  validated against `openapi.json`; `RouteResponse` is a single schema reused by both
  `listRoutes` and `getRoute`, so one Go struct is the contract-faithful mapping. The
  list/get field-set split is not a Go-side choice — it's dictated directly by FR-004 vs.
  FR-007 in the spec (the first read feature in this codebase where list and get
  legitimately differ in displayed shape, unlike `outputs`/`inputs` where list and get show
  identical field sets).
- **Alternatives considered**: Two separate structs (`RouteSummary` for list, `Route` for
  get) — rejected: both endpoints return the exact same `RouteResponse` JSON shape over the
  wire; a second struct would either duplicate every field or require two independent decode
  paths for one schema, adding maintenance surface with no contract benefit. A single struct
  with two renderers (chosen here) keeps the wire-decoding contract-faithful while letting
  display differ, which is where the spec actually draws the line.

## 2. `StartedAt` is nullable; `CreatedAt` is not

- **Decision**: `StartedAt *string` (RFC 3339 string, matching `openapi.json`'s
  `["string", "null"]` type — no `time.Time` parsing/reformatting, consistent with
  `003-inputs-list-get`'s treatment of `Input.CreatedAt`). `nil` means "playback has not
  started yet." `CreatedAt` is a plain `string`, not a pointer — the schema declares it a
  required, single-typed (`"string"`) field, so it is always present and decodes like
  `Route.RouteID`/`InputID`/etc. Rendering: `routes get`'s YAML emits `startedAt: null`
  (bare, unquoted) when `nil`, and `startedAt: "<value>"` (quoted) otherwise; JSON relies on
  `encoding/json`'s native `null` handling for a nil pointer. This directly satisfies the
  spec's edge case ("playback-started timestamp displayed clearly rather than as an error or
  blank field") using the same explicit-`null` convention `003-inputs-list-get` established
  for `Input.CreatedAt`, rather than the spec edge case's illustrative `"n/a"` string.
- **Rationale**: FR-007 requires "its playback-started timestamp (or an explicit indication
  that playback has not started)" — `null` is the codebase's existing, tested convention for
  exactly this "explicit indication of absence" requirement (research.md §2 of
  `003-inputs-list-get`), and reusing it keeps `--json`'s `null` machine-parseable (SC-005)
  without introducing a second, competing "no value" sentinel (`"n/a"`) into structured
  output.
- **Alternatives considered**: The spec edge case's literal `"n/a"` string — rejected for
  `--json` specifically: a sentinel string is ambiguous with a real value and breaks strict
  machine parsing (SC-005); using `null` in both YAML and JSON keeps one representation for
  both output modes, consistent with how `createdAt: null` already renders for static inputs.

## 3. HTTP calls: `GET /api/v2/routes` (`listRoutes`) and `GET /api/v2/routes/{routeId}` (`getRoute`)

- **Decision**: Add `hub.ListRoutes(ctx, client, baseURL, status, inputID, targetID string)
  ([]Route, error)` and `hub.GetRoute(ctx, client, baseURL, routeID string) (*Route, error)`
  in `internal/hub/routes.go`. `ListRoutes` sets `status`/`inputId`/`targetId` as query
  parameters only when the corresponding argument is non-empty (an empty string means "no
  filter", never sent as `status=`), letting the hub apply AND logic across whichever
  filters are present, per `listRoutes`'s documented behavior in `api/openapi.json`. Path
  construction/escaping and non-2xx/404/decode error handling are otherwise structurally
  identical to `ListOutputs`/`GetOutput` and `ListInputs`/`GetInput`.
- **Rationale**: `listRoutes` takes three independent optional query parameters (`status`,
  `inputId`, `targetId`) instead of the boolean `includeDisabled` `outputs`/`inputs` use —
  the CLI's job is to forward whichever filters the user supplied and let the hub do the
  filtering/validation (including rejecting an invalid `status` enum value with its
  documented `400`), exactly mirroring how the CLI already trusts the hub for 404 semantics
  rather than pre-validating identifiers client-side.
- **Alternatives considered**: Client-side validation of the `--status` value against the
  known enum before sending the request — rejected: the hub already validates and returns a
  documented `400` (`ErrorResponse`) for an invalid status, which the existing `StatusError`
  → `ClassHub` path already handles correctly (see research.md §5); duplicating that
  validation client-side would risk drifting from the hub's authoritative enum if it changes,
  violating Principle II's "hub API is the source of truth" framing.

## 4. Filter flags: `--status`, `--input-id`, `--target-id`, combined with AND logic

- **Decision**: `routes list` gains three new optional string flags —
  `--status`, `--input-id`, `--target-id` — each empty (`""`) by default, meaning "no
  filter." When more than one is supplied, all three are forwarded to the hub as query
  parameters simultaneously; the hub applies AND logic (per `listRoutes`'s documented
  description), so the CLI does no client-side combination logic of its own (FR-003).
- **Rationale**: FR-003 requires "only routes matching all supplied filters" when more than
  one filter is given — the hub's own documented AND semantics already deliver this
  directly, so forwarding raw filter values is both the simplest implementation and the one
  most faithful to Principle II (no client-side reimplementation of hub filtering logic).
- **Alternatives considered**: A single combined `--filter key=value` flag repeated for each
  filter — rejected: three separate, named boolean-shaped-string flags are more discoverable
  (`--help` lists each explicitly) and match Principle V's existing pattern of one flag per
  concept (`--include-disabled` on `outputs`/`inputs list`) rather than inventing a new
  generic filter mini-language for one command.

## 5. Exit code scheme: unchanged, no new class needed

- **Decision**: Reuse `ClassNone`/`ClassUsage`/`ClassHub`/`ClassNetwork`/`ClassNotFound`
  (exit `0`/`2`/`3`/`4`/`5`) from `internal/hub/errors.go` exactly as `002-outputs-get` and
  `003-inputs-list-get` defined them. Neither `routes list` nor `routes get` needs a sixth
  class; the hub's `400` for an invalid `--status` value maps to the existing `StatusError` →
  `ClassHub` path, identical to how any other non-2xx/non-404 status is already classified.
- **Rationale**: FR-014 requires the same five-way distinction (success / usage / not-found /
  hub error / network error) prior features already required and implemented; `routes get`'s
  404 case maps onto the existing `ClassNotFound`, and `routes list` has no "not found"
  concept (same as `outputs list`/`inputs list`) so it only ever produces
  none/usage/hub/network — a strict subset of the existing scheme.
- **Alternatives considered**: A dedicated exit code for "invalid filter value" distinct from
  a generic hub error — rejected: the spec's edge case only requires that an invalid status
  filter produce "a clear usage or hub-reported error rather than a silently empty or
  incorrect result" (not a *new*, distinguishable exit class); the hub's `400` → `ClassHub`
  (exit `3`) already satisfies this without adding a class no other command needs.

## 6. Validation rule for malformed responses (FR-016)

- **Decision**: A decoded `Route` is rejected as malformed (`*DecodeError`, `ClassHub`) if
  `RouteID`, `InputID`, or `TargetID` is empty, `TargetType` is neither `"SINGLE_OUTPUT"` nor
  `"OUTPUT_GROUP"`, or `Status` is not one of `"STARTING"`/`"ACTIVE"`/`"STOPPING"`/
  `"STOPPED"`/`"FAILED"`. `CreatedAt`, `StartedAt`, `Transferable`, `Pauseable`, and `Paused`
  are not further validated beyond their Go type (decode failure for a wrong JSON type
  already surfaces as a `*DecodeError` before this check runs).
- **Rationale**: Mirrors `001-list-outputs`'s `outputId`/`displayName` non-empty check and
  `003-inputs-list-get`'s extension for `source`'s enum — `Route` has two schema-constrained
  enums (`targetType`, `status`) whose values FR-004/FR-007 require the CLI to display
  meaningfully; an unrecognized value in either would silently produce a nonsensical
  displayed value, so it's worth rejecting explicitly (FR-016) rather than passing it
  through.
- **Alternatives considered**: Treating an unrecognized `targetType`/`status` as a pass-through
  string rather than a decode error — rejected: FR-016 explicitly requires rejecting data
  that "does not match the expected structure," and both fields are documented closed enums
  in `api/openapi.json`; silently displaying an unknown value would violate that requirement
  and mask a real hub/CLI version mismatch.

## 7. Testing strategy — same three layers, extended for `routes`

- **Decision**: Mirror `001-list-outputs`/`002-outputs-get`/`003-inputs-list-get`'s
  three-layer approach exactly:
  - **Unit** (`tests/unit`): `tests/unit/cli_routes_test.go` and
    `tests/unit/cli_routes_get_test.go` covering flag/positional parsing and dispatch, plus
    `tests/unit/render_routes_test.go` covering the list view's 5-field YAML/JSON, the get
    view's 10-field YAML/JSON, the zero-routes case, and the null-vs-populated `startedAt`
    display.
  - **Contract** (`tests/contract`): `tests/contract/routes_list_test.go` and
    `tests/contract/routes_get_test.go`, each an `httptest.Server` shaped from
    `listRoutes`/`getRoute` in `api/openapi.json` (200 `RouteResponse`/array, 404 for
    `getRoute`, 400 for an invalid `status` filter on `listRoutes`), asserting request
    path/query construction and both decode paths.
  - **Integration** (`tests/integration`): `tests/integration/routes_list_test.go` and
    `tests/integration/routes_get_test.go`, full `sonora routes list`/`sonora routes get`
    process invocations covering the no-filter default, each filter individually, combined
    filters (AND logic), found/not-found/missing-identifier, `--json`, and exit codes.
  All written before their corresponding implementation code, per Principle VI.
- **Rationale**: Same rationale as prior features' research.md — reusing the established
  three-layer taxonomy rather than inventing a new one, applied to the one genuinely new
  decode/render/query surface (`Route`'s split list/get fields and its three filters).
- **Alternatives considered**: None specific to this feature — the layer choice was already
  settled by `001-list-outputs`.

## Related specs (from memory)

- _none found_ (no memory-tool search was available in this environment; the directly
  relevant sibling specs, `001-list-outputs`, `002-outputs-get`, and `003-inputs-list-get`,
  are already cited throughout this document).
