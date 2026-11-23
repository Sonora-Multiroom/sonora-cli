# Contract: `sonora groups get`

**Feature**: `005-groups-list-get` | **Date**: 2026-08-26

This documents the observable interface of the command — inputs, outputs, and exit codes —
as the contract implementation and tests are written against. The underlying HTTP contract
(`GET /api/v2/groups/{groupId}`, operationId `getGroup`) is defined by `api/openapi.json` and
is not duplicated here beyond what's needed to explain CLI behavior.

## Invocation

```text
sonora groups get <group-id> [--json] [--verbose] [--hub-url URL]
```

Exactly one positional argument (`<group-id>`) is required. Zero positional arguments, or
more than one, is a usage error (exit `2`). Any unrecognized flag is also a usage error.

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--json` | bool | `false` | Emit strict JSON instead of the default YAML (FR-011). |
| `--verbose` | bool | `false` | On failure, print the underlying error detail in addition to the friendly message (Principle IV). |
| `--hub-url` | string | `http://localhost:8080`, overridable via `$MULTIROOM_URL`, then `~/.config/sonora/config.json`'s `hubUrl` field, then this flag (highest precedence) | Hub base URL override — identical precedence to `outputs get`. |

Underlying HTTP call: `GET {hub-url}/api/v2/groups/{group-id}`, single attempt, 5s total
timeout, no retries — same client construction as `outputs get`/`inputs get`/`routes get`.

## Example invocations

| # | Command | Result |
|---|---------|--------|
| 1 | `sonora groups get living-room` | That group, YAML, friendly errors only |
| 2 | `sonora groups get living-room --json` | That group, JSON |
| 3 | `sonora groups get living-room --verbose` | YAML; raw error appended on failure |
| 4 | `sonora groups get missing-group` | "group not found" message, exit `5` |
| 5 | `sonora groups get` | Usage error (missing identifier), exit `2` |
| 6 | `sonora groups get a b` | Usage error (unexpected argument `b`), exit `2` |

Hub URL override, identical precedence to `outputs get`:

```bash
sonora groups get living-room --hub-url http://192.168.1.50:9090
MULTIROOM_URL=http://hub.local:8080 sonora groups get living-room
```

## Success output

### Default (YAML)

A single record — not wrapped in a list — with all five fields always present, returned
regardless of the group's enabled/disabled state (FR-007):

```yaml
groupId: "living-room"
displayName: "Living Room Speakers"
outputIds:
  - "office-speaker"
  - "bedroom-speaker"
muted: false
enabled: true
```

A group with zero member outputs shows `outputIds: []` explicitly — never omitted, never
`null` (spec edge case):

```yaml
groupId: "empty-group"
displayName: "Empty Group"
outputIds: []
muted: false
enabled: false
```

### `--json`

Same fields, strict JSON, parseable by any standard JSON parser (SC-005):

```json
{
  "groupId": "living-room",
  "displayName": "Living Room Speakers",
  "outputIds": ["office-speaker", "bedroom-speaker"],
  "muted": false,
  "enabled": true
}
```

## Failure output & exit codes

All failure messages go to stderr; stdout is reserved for the success payload (YAML/JSON) so
scripts piping stdout never have to distinguish success from failure by parsing error text.

| Condition | Exit code | stderr message shape |
|-----------|-----------|-----------------------|
| Missing `<group-id>` | `2` | Usage message stating an identifier is required. |
| Extra positional argument(s) | `2` | Usage message naming the unexpected argument(s). |
| Bad/unknown flag | `2` | Usage message (flag name + expected form). |
| `~/.config/sonora/config.json` exists but is malformed, or `hubUrl` isn't a string | `2` | Usage message naming the config file path and the problem. |
| Hub unreachable / DNS failure / connection refused | `4` | Clear statement that the hub could not be reached, with the hub URL used. |
| Request timeout (>5s, no response) | `4` | Clear statement that the hub did not respond in time. |
| Hub returned `404` for the given identifier | `5` | Clear "group not found" message naming the identifier (FR-011) — distinct from both a connectivity failure and a generic hub error. |
| Hub returned another non-2xx (e.g. 5xx) | `3` | Statement that the hub reported an error, distinguishing it from connectivity failure and from "not found" (FR-014). |
| Hub returned 2xx but body doesn't match `GroupResponse` shape | `3` | Statement that the hub's response was malformed/unexpected (FR-017) — never partial/garbled output. |

With `--verbose`, each of the above additionally prints the underlying Go error after the
friendly message.

## Out of scope for this contract

- Any mutation (create/delete/volume/mute/enabled) — read-only command.
- Authentication headers — spec Assumptions state the hub's read-only endpoints require none,
  consistent with `api/openapi.json`.
- Fuzzy/partial identifier matching — the identifier is opaque and matched exactly as given
  (spec Assumptions).
