# Data Model: Get Single Audio Output

**Feature**: `002-outputs-get` | **Date**: 2026-08-25

## Output

Same entity, same shape as [001-list-outputs/data-model.md § Output](../001-list-outputs/data-model.md#output)
— this feature reuses `hub.Output` unchanged (`#/components/schemas/OutputResponse`); it
adds no new fields and no new struct. Reproduced here for reference:

| Field | Type | Source field (openapi.json) | Notes |
|-------|------|------------------------------|-------|
| `OutputID` | string | `outputId` | Unique identifier; required, non-empty. |
| `DisplayName` | string | `displayName` | Human-readable name; required, non-empty. |
| `Volume` | int (0–100) | `volume` | Current volume level; passed through as-is. |
| `Muted` | bool | `muted` | Current mute state. |
| `Available` | bool | `available` | Whether the underlying hardware is currently connected. Drives FR-005's visible distinction. |
| `Enabled` | bool | `enabled` | Unlike `outputs list`'s default, `outputs get` always returns the output regardless of this value (FR-003). |

**Validation rule (FR-013)**: identical to `001-list-outputs` — a decoded body missing
`outputId`/`displayName` (empty string), or with a field of the wrong JSON type, is treated
as malformed and fails with the Hub-error exit class rather than rendering partial data.

**Difference from `001-list-outputs`**: that feature's decode path validates each element of
an array; this feature validates a single decoded object (there is no `[]Output` step, and
`outputs == nil`'s "treat as empty array" case does not apply — a `getOutput` 200 response
is always exactly one object).

## OutputsGetQuery (request shaping)

Not a persisted entity — the parameter this command sends to the hub.

| Field | Type | Maps to | Default |
|-------|------|---------|---------|
| `OutputID` | string | `{outputId}` path segment on `GET /api/v2/outputs/{outputId}` | none — required, sourced from the command's positional argument |

## CLI invocation shape

The flags and argument accepted by `sonora outputs get`, and how they affect behavior — see
[contracts/cli-outputs-get.md](contracts/cli-outputs-get.md) for the full contract.

| Argument/Flag | Type | Default | Effect |
|------|------|---------|--------|
| `<output-id>` (positional) | string | none, required | Selects the output to fetch (FR-002). Missing → usage error (exit `2`). |
| `--json` | bool | `false` | Switches rendering from default YAML to JSON; FR-007. |
| `--verbose` | bool | `false` | On failure, additionally prints the underlying error detail; Principle IV. |
| `--hub-url` | string | Same three-layer precedence as `outputs list` — see [001-list-outputs/research.md §5](../001-list-outputs/research.md#5-hub-location-resolution-spec-assumption-established-once-this-feature-consumes-but-does-not-define) (`--hub-url` flag → `MULTIROOM_URL` env → `~/.config/sonora/config.json`'s `hubUrl` → `http://localhost:8080`) | Overrides the hub base URL. |

`--include-disabled` does not exist on this command: FR-003 makes the enabled/disabled
filter moot for a single-identifier lookup — the output is always returned regardless of
its `enabled` state.

## Config file

Unchanged from `001-list-outputs`; this feature reads the same
`~/.config/sonora/config.json` through the existing `config.ResolveHubURL`, with no new
fields.

## Exit code classes (extends 001-list-outputs research.md §6)

| Code | Meaning | New in this feature? |
|------|---------|----------------------|
| `0` | Success | No |
| `2` | Usage error (bad/unknown flag, missing/extra positional argument) | No (extended to cover missing identifier) |
| `3` | Hub-reported error (non-2xx other than 404, or malformed response body) | No |
| `4` | Network/connectivity error (unreachable host, DNS failure, timeout) | No |
| `5` | Not found — hub responded `404` for the given identifier (FR-008, FR-012) | **Yes** |

See [research.md §2](research.md#2-the-404-not-found-case-needs-its-own-exit-code-class-fr-012)
for why this is an additive fifth class rather than a renumbering.
