# Data Model: List and Get Audio Inputs

**Feature**: `003-inputs-list-get` | **Date**: 2026-08-25

## Input

New entity for this feature — maps field-for-field to
`#/components/schemas/InputResponse` in `api/openapi.json`. See
[research.md §1](research.md#1-the-input-entity-has-no-overlap-with-outputs-fields).

| Field | Type | Source field (openapi.json) | Notes |
|-------|------|------------------------------|-------|
| `InputID` | string | `inputId` | Unique identifier; required, non-empty. |
| `DisplayName` | string | `displayName` | Human-readable name; required, non-empty. |
| `URI` | string | `uri` | Source URI; passed through as-is, no validation. |
| `Enabled` | bool | `enabled` | Unlike `inputs list`'s default, `inputs get` always returns the input regardless of this value (FR-007). |
| `AutoRemove` | bool | `autoRemove` | Whether the input is removed automatically when its route stops. |
| `Source` | string | `source` | Enum `STATIC` \| `EPHEMERAL`; required, must be one of these two values. |
| `CreatedAt` | `*string` (nullable) | `createdAt` | RFC 3339 timestamp string, or `nil` for static inputs. Never re-parsed/reformatted — display-only. See [research.md §2](research.md#2-createdat-is-the-first-nullable-field-in-this-codebase). |
| `Pauseable` | bool | `pauseable` | Whether this input supports pause/resume. |

**Validation rule (FR-017)**: a decoded body is treated as malformed (Hub-error exit class,
not partial/garbled output) if `InputID` or `DisplayName` is empty, or if `Source` is
neither `"STATIC"` nor `"EPHEMERAL"`. See
[research.md §8](research.md#8-validation-rule-for-malformed-responses-fr-017) for why
`URI`/`AutoRemove`/`Pauseable`/`CreatedAt` need no further validation beyond their Go type.

**List decode**: identical shape to `001-list-outputs`'s `ListOutputs` — a `nil`/absent JSON
array decodes to an empty `[]Input{}`, not `nil`, so the "zero inputs" case (FR-016) renders
consistently rather than as a JSON `null`.

## InputsListQuery / InputsGetQuery (request shaping)

Not persisted entities — the parameters these commands send to the hub.

| Field | Type | Maps to | Default |
|-------|------|---------|---------|
| `IncludeDisabled` | bool | `includeDisabled` query parameter on `GET /api/v2/inputs` | `false` |
| `InputID` | string | `{inputId}` path segment on `GET /api/v2/inputs/{inputId}` | none — required, sourced from `inputs get`'s positional argument |

## CLI invocation shape

See [contracts/cli-inputs-list.md](contracts/cli-inputs-list.md) and
[contracts/cli-inputs-get.md](contracts/cli-inputs-get.md) for the full contracts.

| Command | Argument/Flag | Type | Default | Effect |
|---|------|------|---------|--------|
| `inputs list` | `--include-disabled` | bool | `false` | Include disabled inputs in the results (FR-003). |
| `inputs list` | `--json` | bool | `false` | Switch rendering from default YAML to JSON (FR-010). |
| `inputs list` | `--verbose` | bool | `false` | On failure, additionally print the underlying error detail (Principle IV). |
| `inputs list` | `--hub-url` | string | See `outputs list`'s precedence | Overrides the hub base URL. |
| `inputs get` | `<input-id>` (positional) | string | none, required | Selects the input to fetch (FR-006). Missing → usage error (exit `2`). |
| `inputs get` | `--json` | bool | `false` | Switch rendering from default YAML to JSON (FR-010). |
| `inputs get` | `--verbose` | bool | `false` | On failure, additionally print the underlying error detail (Principle IV). |
| `inputs get` | `--hub-url` | string | See `outputs get`'s precedence | Overrides the hub base URL. |

`inputs get` has no `--include-disabled` flag: FR-007 makes the enabled/disabled filter
moot for a single-identifier lookup — the input is always returned regardless of its
`enabled` state, exactly mirroring `outputs get`.

## Config file

Unchanged from `001-list-outputs`/`002-outputs-get`; both commands read the same
`~/.config/sonora/config.json` through the existing `config.ResolveHubURL`, with no new
fields.

## Exit code classes (unchanged from `002-outputs-get`)

| Code | Meaning | New in this feature? |
|------|---------|----------------------|
| `0` | Success | No |
| `2` | Usage error (bad/unknown flag, missing/extra positional argument) | No |
| `3` | Hub-reported error (non-2xx other than 404, or malformed response body) | No |
| `4` | Network/connectivity error (unreachable host, DNS failure, timeout) | No |
| `5` | Not found — hub responded `404` for the given identifier (`inputs get` only) | No (reused, not renumbered — see [research.md §5](research.md#5-exit-code-scheme-unchanged-no-new-class-needed)) |
