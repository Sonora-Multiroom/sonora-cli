# Contract: `sonora groups list`

**Feature**: `005-groups-list-get` | **Date**: 2026-08-26

This documents the observable interface of the command — inputs, outputs, and exit codes —
as the contract implementation and tests are written against. The underlying HTTP contract
(`GET /api/v2/groups`, operationId `listGroups`) is defined by `api/openapi.json` and is not
duplicated here beyond what's needed to explain CLI behavior.

## Invocation

```text
sonora groups list [--include-disabled] [--json] [--verbose] [--hub-url URL]
```

No positional arguments. Any unrecognized flag or unexpected positional argument is a usage
error (exit `2`).

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--include-disabled` | bool | `false` | Include disabled groups in the results (FR-002, FR-003). |
| `--json` | bool | `false` | Emit strict JSON instead of the default YAML (FR-011). |
| `--verbose` | bool | `false` | On failure, print the underlying error detail in addition to the friendly message (Principle IV). |
| `--hub-url` | string | `http://localhost:8080`, overridable via `$MULTIROOM_URL`, then `~/.config/sonora/config.json`'s `hubUrl` field, then this flag (highest precedence) | Hub base URL override — identical precedence to `outputs list`. |

Underlying HTTP call: `GET {hub-url}/api/v2/groups?includeDisabled=true|false` — the parameter
is always sent explicitly — single attempt, 5s total timeout, no retries — same client
construction as `outputs list`/`inputs list`/`routes list`.

## Example invocations

| # | Command | Result |
|---|---------|--------|
| 1 | `sonora groups list` | Only enabled groups, YAML |
| 2 | `sonora groups list --include-disabled` | All groups regardless of enabled state, YAML |
| 3 | `sonora groups list --json` | Only enabled groups, JSON |
| 4 | `sonora groups list --verbose` | YAML; raw error appended on failure |

Hub URL override, identical precedence to `outputs list`:

```bash
sonora groups list --hub-url http://192.168.1.50:9090
MULTIROOM_URL=http://hub.local:8080 sonora groups list
```

## Success output

### Default (YAML)

All five fields are shown per group — the same field set `groups get` shows (FR-004, FR-008;
see [research.md §1](../research.md#1-the-group-entity-same-listget-field-set-structurally-closest-to-output)):

```yaml
groups:
  - groupId: "living-room"
    displayName: "Living Room Speakers"
    outputIds:
      - "office-speaker"
      - "bedroom-speaker"
    muted: false
    enabled: true
  - groupId: "whole-house"
    displayName: "Whole House"
    outputIds: []
    muted: true
    enabled: false
```

Zero groups (FR-016):

```yaml
# no groups found
groups: []
```

### `--json`

Same fields, strict JSON, parseable by any standard JSON parser (SC-005):

```json
{
  "groups": [
    {
      "groupId": "living-room",
      "displayName": "Living Room Speakers",
      "outputIds": ["office-speaker", "bedroom-speaker"],
      "muted": false,
      "enabled": true
    }
  ]
}
```

## Failure output & exit codes

All failure messages go to stderr; stdout is reserved for the success payload (YAML/JSON) so
scripts piping stdout never have to distinguish success from failure by parsing error text.

| Condition | Exit code | stderr message shape |
|-----------|-----------|-----------------------|
| Unexpected positional argument(s) | `2` | Usage message naming the unexpected argument(s). |
| Bad/unknown flag | `2` | Usage message (flag name + expected form). |
| `~/.config/sonora/config.json` exists but is malformed, or `hubUrl` isn't a string | `2` | Usage message naming the config file path and the problem. |
| Hub unreachable / DNS failure / connection refused | `4` | Clear statement that the hub could not be reached, with the hub URL used. |
| Request timeout (>5s, no response) | `4` | Clear statement that the hub did not respond in time. |
| Hub returned non-2xx (e.g. 5xx) | `3` | Statement that the hub reported an error, distinguishing it from a connectivity failure. |
| Hub returned 2xx but body doesn't match `GroupResponse[]` shape | `3` | Statement that the hub's response was malformed/unexpected (FR-017) — never partial/garbled output. |

With `--verbose`, each of the above additionally prints the underlying Go error after the
friendly message.

## Out of scope for this contract

- Any mutation (create/delete/volume/mute/enabled) — read-only command.
- Authentication headers — spec Assumptions state the hub's read-only endpoints require none,
  consistent with `api/openapi.json`.
- Sort ordering — groups are presented in the order the hub returns them (spec Assumptions).
