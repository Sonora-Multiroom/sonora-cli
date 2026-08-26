# Research: List and Get Output Groups

**Feature**: `005-groups-list-get` | **Date**: 2026-08-26

This document resolves every open technical question for implementing `sonora groups list`
and `sonora groups get <group-id>` before design (Phase 1) begins. The spec has no unresolved
`NEEDS CLARIFICATION` markers — this feature deliberately reuses every foundational decision
`001-list-outputs`, `002-outputs-get`, `003-inputs-list-get`, and `004-routes-list-get` already
made (language, dependency policy, HTTP client construction, hub URL resolution, YAML/JSON
rendering approach, `--verbose`/error-translation shape, exit-code scheme, testing strategy,
the generalized `NotFoundError{Resource, ID}`). Only the decisions genuinely new to this
feature — driven by the `Group` entity's shape — are recorded below.

## 1. The `Group` entity: same list/get field set, structurally closest to `Output`

- **Decision**: Add a new `Group` struct in `internal/hub/groups.go`, mapping field-for-field
  to `#/components/schemas/GroupResponse` in `api/openapi.json`: `GroupID` (`groupId`),
  `DisplayName` (`displayName`), `OutputIDs` (`outputIds`, `[]string`), `Muted` (`muted`,
  bool), `Enabled` (`enabled`, bool). Unlike `004-routes-list-get`'s `Route` (where list and
  get intentionally show different field sets), FR-004 and FR-008 require the exact same five
  fields from both `groups list` and `groups get` — the same shape `001-list-outputs`'s
  `Output` established (`outputId`/`displayName`/.../`enabled`), just with a `[]string` member
  list (`outputIds`) in place of `Output`'s scalar `volume`/`available` fields.
- **Rationale**: Principle II requires request/response types to be generated from or
  validated against `openapi.json`; `GroupResponse` is a single schema reused by both
  `listGroups` and `getGroup`, so one Go struct, rendered identically by both commands, is the
  contract-faithful mapping — mirroring `Output`'s precedent rather than `Route`'s split-view
  one, since the spec draws no list/get distinction here.
- **Alternatives considered**: Following `Route`'s split-struct-view pattern even though the
  field sets are identical — rejected: there is nothing to split; FR-004 and FR-008 list the
  exact same five attributes, so introducing a list-view/get-view distinction would add
  surface area the spec doesn't ask for.

## 2. `OutputIDs` is a plain `[]string`, never `nil` on decode

- **Decision**: `Group.OutputIDs []string`, decoded as-is from the hub's `outputIds` array. If
  the hub returns `[]` (zero member outputs — an explicit edge case in spec.md), it decodes to
  an empty, non-nil `[]string{}` and renders as `outputIds: []` in YAML / `"outputIds": []` in
  JSON — never omitted, never rendered as `null`.
- **Rationale**: The spec's edge case requires a group with zero member outputs to display
  "clearly as empty rather than as an error or omitted field" — an empty slice rendered
  explicitly satisfies this directly and keeps `--json` output strictly parseable (SC-005)
  without introducing a `null`-vs-`[]` ambiguity that `004-routes-list-get`'s nullable
  `StartedAt` (a genuinely optional scalar) needed but a member list does not.
- **Alternatives considered**: Treating an absent/`null` `outputIds` key as a decode error —
  rejected: `GroupResponse`'s schema does not mark `outputIds` nullable, and normalizing a
  JSON `null` to `[]string{}` (the same normalization already applied to the top-level
  list-of-groups array in `ListGroups`, see §5 below) is simpler and consistent with existing
  precedent than rejecting it outright.

## 3. HTTP calls: `GET /api/v2/groups` (`listGroups`) and `GET /api/v2/groups/{groupId}` (`getGroup`)

- **Decision**: Add `hub.ListGroups(ctx, client, baseURL string, includeDisabled bool)
  ([]Group, error)` and `hub.GetGroup(ctx, client, baseURL, groupID string) (*Group, error)` in
  `internal/hub/groups.go`, structurally identical to `hub.ListOutputs`/`hub.GetOutput`:
  `ListGroups` always sends `includeDisabled` as an explicit `true`/`false` query parameter
  (never omitted), matching `listGroups`'s documented boolean-with-default-`false` parameter
  in `api/openapi.json`. Path construction/escaping and non-2xx/404/decode error handling are
  otherwise identical to `ListOutputs`/`GetOutput` and `ListRoutes`/`GetRoute`.
- **Rationale**: `listGroups` takes the same boolean `includeDisabled` shape `listOutputs`
  and `listInputs` already use (not `routes`'s three independent string filters), so the
  existing `outputs`/`inputs`-style boolean-flag pattern is the direct, contract-faithful
  match — no new query-parameter-handling approach is needed.
- **Alternatives considered**: Reusing `routes`'s "omit unless non-empty" filter-forwarding
  pattern — not applicable: `includeDisabled` is a boolean with a hub-documented default,
  identical in shape to `outputs`/`inputs`'s existing `includeDisabled`, not an optional string
  filter; `001-list-outputs`'s always-send-the-bool approach is the correct precedent to reuse
  here, not `004-routes-list-get`'s.

## 4. Exit code scheme: unchanged, no new class needed

- **Decision**: Reuse `ClassNone`/`ClassUsage`/`ClassHub`/`ClassNetwork`/`ClassNotFound`
  (exit `0`/`2`/`3`/`4`/`5`) from `internal/hub/errors.go` exactly as prior features defined
  them. `groups get`'s 404 case maps onto `ClassNotFound` via `NotFoundError{Resource:
  "group", ID: groupID}`; `groups list` has no "not found" concept (same as `outputs
  list`/`inputs list`/`routes list`) so it only ever produces none/usage/hub/network.
- **Rationale**: FR-015 requires the same five-way distinction every prior read feature
  already required and implemented; no group-specific failure mode needs a sixth class.
- **Alternatives considered**: None — this feature introduces no new failure mode beyond what
  `outputs`/`inputs`/`routes` already cover.

## 5. Validation rule for malformed responses (FR-017)

- **Decision**: A decoded `Group` is rejected as malformed (`*DecodeError`, `ClassHub`) if
  `GroupID` or `DisplayName` is empty. `OutputIDs`, `Muted`, and `Enabled` are not further
  validated beyond their Go type (a decode failure for a wrong JSON type already surfaces as a
  `*DecodeError` before this check runs; `Group` has no enum fields the way `Route` does, so
  no enum-membership check is needed). A `nil`/absent top-level array from `listGroups`
  decodes to an empty `[]Group{}`, not `nil` (identical to `ListOutputs`/`ListRoutes`'s
  existing normalization), so the "zero groups" case (FR-016) renders consistently rather than
  as a JSON `null`.
- **Rationale**: Mirrors `001-list-outputs`'s `outputId`/`displayName` non-empty check exactly
  — `Group` shares the same two required-identifier-and-label fields, with no additional enum
  surface (unlike `Route`'s `targetType`/`status`) that would need its own validation rule.
- **Alternatives considered**: Also requiring `OutputIDs` to be non-empty — rejected: the
  spec's edge case explicitly requires a group with zero member outputs to be accepted and
  displayed as an empty list, not rejected as malformed.

## 6. Testing strategy — same three layers, extended for `groups`

- **Decision**: Mirror `001-list-outputs`/`002-outputs-get`/`003-inputs-list-get`/
  `004-routes-list-get`'s three-layer approach exactly:
  - **Unit** (`tests/unit`): `tests/unit/cli_groups_test.go` and
    `tests/unit/cli_groups_get_test.go` covering flag/positional parsing and dispatch, plus
    `tests/unit/render_groups_test.go` covering the five-field YAML/JSON view (list and get,
    identical field set), the zero-groups case, and the zero-member-outputs (`outputIds: []`)
    case.
  - **Contract** (`tests/contract`): `tests/contract/groups_list_test.go` and
    `tests/contract/groups_get_test.go`, each an `httptest.Server` shaped from
    `listGroups`/`getGroup` in `api/openapi.json` (200 `GroupResponse`/array, 404 for
    `getGroup`), asserting request path/query construction (including the always-sent
    `includeDisabled` parameter) and both decode paths.
  - **Integration** (`tests/integration`): `tests/integration/groups_list_test.go` and
    `tests/integration/groups_get_test.go`, full `sonora groups list`/`sonora groups get`
    process invocations covering the default (enabled-only) view, `--include-disabled`,
    found/not-found/missing-identifier, `--json`, and exit codes.
  All written before their corresponding implementation code, per Principle VI.
- **Rationale**: Same rationale as prior features' research.md — reusing the established
  three-layer taxonomy rather than inventing a new one, applied to the one genuinely new
  decode/render surface (`Group`'s `outputIds` member list).
- **Alternatives considered**: None specific to this feature — the layer choice was already
  settled by `001-list-outputs`.

## Related specs (from memory)

- _none found_ (no memory-tool search was available in this environment; the directly
  relevant sibling specs, `001-list-outputs`, `002-outputs-get`, `003-inputs-list-get`, and
  `004-routes-list-get`, are already cited throughout this document).
