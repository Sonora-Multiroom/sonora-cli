# Contract: `sonora outputs get`

**Feature**: `002-outputs-get` | **Date**: 2026-08-25

This documents the observable interface of the command — inputs, outputs, and exit codes —
as the contract implementation and tests are written against. The underlying HTTP contract
(`GET /api/v2/outputs/{outputId}`, operationId `getOutput`) is defined by `api/openapi.json`
and is not duplicated here beyond what's needed to explain CLI behavior.

## Invocation

```text
sonora outputs get <output-id> [--json] [--verbose] [--hub-url URL]
```

Exactly one positional argument (`<output-id>`) is required. Zero positional arguments, or
more than one, is a usage error (exit `2`). Any unrecognized flag is also a usage error.

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--json` | bool | `false` | Emit strict JSON instead of the default YAML (FR-007). |
| `--verbose` | bool | `false` | On failure, print the underlying error detail in addition to the friendly message (Principle IV). |
| `--hub-url` | string | `http://localhost:8080`, overridable via `$MULTIROOM_URL`, then `~/.config/sonora/config.json`'s `hubUrl` field, then this flag (highest precedence) | Hub base URL override — identical precedence to `outputs list`. |

Underlying HTTP call: `GET {hub-url}/api/v2/outputs/{output-id}`, single attempt, 5s total
timeout, no retries — same client construction as `outputs list`.

There is no `--include-disabled` flag: the output is always returned regardless of its
enabled state (FR-003) — filtering by enabled state is meaningless for a single-identifier
lookup.

## Example invocations

| # | Command | Result |
|---|---------|--------|
| 1 | `sonora outputs get office-speaker` | That output, YAML, friendly errors only |
| 2 | `sonora outputs get office-speaker --json` | That output, JSON |
| 3 | `sonora outputs get office-speaker --verbose` | YAML; raw error appended on failure |
| 4 | `sonora outputs get missing-speaker` | "output not found" message, exit `5` |
| 5 | `sonora outputs get` | Usage error (missing identifier), exit `2` |
| 6 | `sonora outputs get a b` | Usage error (unexpected argument `b`), exit `2` |

Hub URL override, identical precedence to `outputs list`:

```bash
sonora outputs get office-speaker --hub-url http://192.168.1.50:9090
MULTIROOM_URL=http://hub.local:8080 sonora outputs get office-speaker
```

## Success output

### Default (YAML)

A single record — not wrapped in a list — with every field always present:

```yaml
outputId: "office-speaker"
displayName: "Office Speaker"
volume: 75
muted: false
available: true
enabled: true
```

A disabled output is still returned (FR-003) with its actual state shown, e.g.
`enabled: false`. An enabled-but-unavailable output shows `available: false` explicitly
(FR-005) — never omitted, so its absence can't be mistaken for "available."

### `--json`

Same fields, strict JSON, parseable by any standard JSON parser (SC-004):

```json
{
  "outputId": "office-speaker",
  "displayName": "Office Speaker",
  "volume": 75,
  "muted": false,
  "available": true,
  "enabled": true
}
```

## Failure output & exit codes

All failure messages go to stderr; stdout is reserved for the success payload (YAML/JSON) so
scripts piping stdout never have to distinguish success from failure by parsing error text.

| Condition | Exit code | stderr message shape |
|-----------|-----------|-----------------------|
| Missing `<output-id>` | `2` | Usage message stating an identifier is required. |
| Extra positional argument(s) | `2` | Usage message naming the unexpected argument(s). |
| Bad/unknown flag | `2` | Usage message (flag name + expected form). |
| `~/.config/sonora/config.json` exists but is malformed, or `hubUrl` isn't a string | `2` | Usage message naming the config file path and the problem. |
| Hub unreachable / DNS failure / connection refused | `4` | Clear statement that the hub could not be reached, with the hub URL used. |
| Request timeout (>5s, no response) | `4` | Clear statement that the hub did not respond in time. |
| Hub returned `404` for the given identifier | `5` | Clear "output not found" message naming the identifier (FR-008) — distinct from both a connectivity failure and a generic hub error. |
| Hub returned another non-2xx (e.g. 5xx) | `3` | Statement that the hub reported an error, distinguishing it from connectivity failure and from "not found" (FR-011). |
| Hub returned 2xx but body doesn't match `OutputResponse` shape | `3` | Statement that the hub's response was malformed/unexpected (FR-013) — never partial/garbled output. |

With `--verbose`, each of the above additionally prints the underlying Go error after the
friendly message.

## Out of scope for this contract

- Any mutation (volume/mute/enable) — read-only command, same as `outputs list`.
- Authentication headers — spec Assumptions state the hub's read-only endpoints require
  none, consistent with `api/openapi.json`.
- Fuzzy/partial identifier matching — the identifier is opaque and matched exactly as given
  (spec Assumptions).
