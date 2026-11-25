# Data Model: List and Get Output Groups

**Feature**: `005-groups-list-get` | **Date**: 2026-08-26

## Group

New entity for this feature — maps field-for-field to
`#/components/schemas/GroupResponse` in `api/openapi.json`. See
[research.md §1](research.md#1-the-group-entity-same-listget-field-set-structurally-closest-to-output).

| Field | Type | Source field (openapi.json) | Notes |
|-------|------|------------------------------|-------|
| `GroupID` | string | `groupId` | Unique identifier; required, non-empty. Shown by both `list` and `get`. |
| `DisplayName` | string | `displayName` | Human-readable name; required, non-empty. Shown by both `list` and `get`. |
| `OutputIDs` | `[]string` | `outputIds` | Member output identifiers; may be empty (a group with zero member outputs), never `nil` on decode. See [research.md §2](research.md#2-outputids-is-a-plain-string-never-nil-on-decode). Shown by both `list` and `get`. |
| `Muted` | bool | `muted` | Whether all member outputs are currently muted. Shown by both `list` and `get`. |
| `Enabled` | bool | `enabled` | Whether the group is enabled for new route creation. Shown by both `list` and `get`. |

**Validation rule (FR-017)**: a decoded body is treated as malformed (Hub-error exit class,
not partial/garbled output) if `GroupID` or `DisplayName` is empty. `OutputIDs`, `Muted`, and
`Enabled` need no further validation beyond their Go type — `Group` has no enum fields (unlike
`Route`'s `targetType`/`status`). See
[research.md §5](research.md#5-validation-rule-for-malformed-responses-fr-017).

**List decode**: identical shape to `001-list-outputs`'s `ListOutputs` — a `nil`/absent JSON
array decodes to an empty `[]Group{}`, not `nil`, so the "zero groups" case (FR-016) renders
consistently rather than as a JSON `null`.

**List vs. get field set**: unlike `004-routes-list-get`'s `Route`, `groups list` and `groups
get` render the exact same five fields (FR-004, FR-008) — there is no split view for this
entity.

## GroupsListQuery / GroupsGetQuery (request shaping)

Not persisted entities — the parameters these commands send to the hub.

| Field | Type | Maps to | Default |
|-------|------|---------|---------|
| `IncludeDisabled` | bool | `includeDisabled` query parameter on `GET /api/v2/groups` | `false` (always sent explicitly, matching `listGroups`'s documented default) |
| `GroupID` | string | `{groupId}` path segment on `GET /api/v2/groups/{groupId}` | none — required, sourced from `groups get`'s positional argument |

## CLI invocation shape

See [contracts/cli-groups-list.md](contracts/cli-groups-list.md) and
[contracts/cli-groups-get.md](contracts/cli-groups-get.md) for the full contracts.

| Command | Argument/Flag | Type | Default | Effect |
|---|------|------|---------|--------|
| `groups list` | `--include-disabled` | bool | `false` | Include disabled groups in the results (FR-003). |
| `groups list` | `--json` | bool | `false` | Switch rendering from default YAML to JSON (FR-011). |
| `groups list` | `--verbose` | bool | `false` | On failure, additionally print the underlying error detail (Principle IV). |
| `groups list` | `--hub-url` | string | See `outputs list`'s precedence | Overrides the hub base URL. |
| `groups get` | `<group-id>` (positional) | string | none, required | Selects the group to fetch (FR-006). Missing → usage error (exit `2`). |
| `groups get` | `--json` | bool | `false` | Switch rendering from default YAML to JSON (FR-011). |
| `groups get` | `--verbose` | bool | `false` | On failure, additionally print the underlying error detail (Principle IV). |
| `groups get` | `--hub-url` | string | See `outputs get`'s precedence | Overrides the hub base URL. |

## Config file

Unchanged from prior features; both commands read the same
`~/.config/sonora/config.json` through the existing `config.ResolveHubURL`, with no new
fields.

## Exit code classes (unchanged from prior features)

| Code | Meaning | New in this feature? |
|------|---------|----------------------|
| `0` | Success | No |
| `2` | Usage error (bad/unknown flag, missing/extra positional argument) | No |
| `3` | Hub-reported error (non-2xx other than 404, or malformed response body) | No |
| `4` | Network/connectivity error (unreachable host, DNS failure, timeout) | No |
| `5` | Not found — hub responded `404` for the given identifier (`groups get` only) | No (reused, not renumbered — see [research.md §4](research.md#4-exit-code-scheme-unchanged-no-new-class-needed)) |
