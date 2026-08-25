# Contract: `sonora inputs list`

**Feature**: `003-inputs-list-get` | **Date**: 2026-08-25

This documents the observable interface of the command — inputs, outputs, and exit codes —
as the contract implementation and tests are written against. The underlying HTTP contract
(`GET /api/v2/inputs`, operationId `listInputs`) is defined by `api/openapi.json` and is not
duplicated here beyond what's needed to explain CLI behavior.

## Invocation

```text
sonora inputs list [--include-disabled] [--json] [--verbose] [--hub-url URL]
```

No positional arguments. Any unrecognized flag or unexpected positional argument is a usage
error (exit `2`).

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--include-disabled` | bool | `false` | Include disabled inputs in the results (FR-003). |
| `--json` | bool | `false` | Emit strict JSON instead of the default YAML (FR-010). |
| `--verbose` | bool | `false` | On failure, print the underlying error detail in addition to the friendly message (Principle IV). |
| `--hub-url` | string | `http://localhost:8080`, overridable via `$MULTIROOM_URL`, then `~/.config/sonora/config.json`'s `hubUrl` field, then this flag (highest precedence) | Hub base URL override — identical precedence to `outputs list`. |

Underlying HTTP call: `GET {hub-url}/api/v2/inputs?includeDisabled={bool}`, single attempt,
5s total timeout, no retries — same client construction as `outputs list`.

## Example invocations

| # | Command | Result |
|---|---------|--------|
| 1 | `sonora inputs list` | Enabled inputs only, YAML |
| 2 | `sonora inputs list --include-disabled` | All inputs, YAML, `enabled` field shown per input |
| 3 | `sonora inputs list --json` | Enabled inputs only, JSON |
| 4 | `sonora inputs list --verbose` | YAML; raw error appended on failure |

Hub URL override, identical precedence to `outputs list`:

```bash
sonora inputs list --hub-url http://192.168.1.50:9090
MULTIROOM_URL=http://hub.local:8080 sonora inputs list
```

## Success output

### Default (YAML)

```yaml
inputs:
  - inputId: "spotify-1"
    displayName: "Spotify Stream"
    uri: "https://stream.example.com/live.mp3"
    source: "EPHEMERAL"
    enabled: true
    autoRemove: true
    pauseable: true
    createdAt: "2026-06-22T14:30:00Z"
  - inputId: "line-in-1"
    displayName: "Line In"
    uri: "line:///1"
    source: "STATIC"
    enabled: true
    autoRemove: false
    pauseable: false
    createdAt: null
```

Zero inputs (FR-016):

```yaml
# no inputs found
inputs: []
```

A static input's `createdAt` is always shown as explicit `null` (FR-004; never omitted, per
the same rule `outputs`'s `available: false` follows).

### `--json`

Same fields, strict JSON, parseable by any standard JSON parser (SC-005):

```json
{
  "inputs": [
    {
      "inputId": "spotify-1",
      "displayName": "Spotify Stream",
      "uri": "https://stream.example.com/live.mp3",
      "source": "EPHEMERAL",
      "enabled": true,
      "autoRemove": true,
      "pauseable": true,
      "createdAt": "2026-06-22T14:30:00Z"
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
| Hub returned a non-2xx (e.g. 5xx) | `3` | Statement that the hub reported an error, distinguishing it from a connectivity failure. |
| Hub returned 2xx but body doesn't match `InputResponse[]` shape (including an unrecognized `source` value) | `3` | Statement that the hub's response was malformed/unexpected (FR-017) — never partial/garbled output. |

With `--verbose`, each of the above additionally prints the underlying Go error after the
friendly message.

## Out of scope for this contract

- Any mutation (create/delete/enable) — read-only command, same as `outputs list`.
- Authentication headers — spec Assumptions state the hub's read-only endpoints require
  none, consistent with `api/openapi.json`.
- Sort ordering — inputs are presented in the order the hub returns them (spec Assumptions).
