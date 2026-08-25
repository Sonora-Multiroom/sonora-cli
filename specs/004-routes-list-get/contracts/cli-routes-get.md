# Contract: `sonora routes get`

**Feature**: `004-routes-list-get` | **Date**: 2026-08-25

This documents the observable interface of the command — inputs, outputs, and exit codes —
as the contract implementation and tests are written against. The underlying HTTP contract
(`GET /api/v2/routes/{routeId}`, operationId `getRoute`) is defined by `api/openapi.json` and
is not duplicated here beyond what's needed to explain CLI behavior.

## Invocation

```text
sonora routes get <route-id> [--json] [--verbose] [--hub-url URL]
```

Exactly one positional argument (`<route-id>`) is required. Zero positional arguments, or
more than one, is a usage error (exit `2`). Any unrecognized flag is also a usage error.

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--json` | bool | `false` | Emit strict JSON instead of the default YAML (FR-009). |
| `--verbose` | bool | `false` | On failure, print the underlying error detail in addition to the friendly message (Principle IV). |
| `--hub-url` | string | `http://localhost:8080`, overridable via `$MULTIROOM_URL`, then `~/.config/sonora/config.json`'s `hubUrl` field, then this flag (highest precedence) | Hub base URL override — identical precedence to `outputs get`. |

Underlying HTTP call: `GET {hub-url}/api/v2/routes/{route-id}`, single attempt, 5s total
timeout, no retries — same client construction as `outputs get`/`inputs get`.

## Example invocations

| # | Command | Result |
|---|---------|--------|
| 1 | `sonora routes get route-abc-123` | That route, YAML, friendly errors only |
| 2 | `sonora routes get route-abc-123 --json` | That route, JSON |
| 3 | `sonora routes get route-abc-123 --verbose` | YAML; raw error appended on failure |
| 4 | `sonora routes get missing-route` | "route not found" message, exit `5` |
| 5 | `sonora routes get` | Usage error (missing identifier), exit `2` |
| 6 | `sonora routes get a b` | Usage error (unexpected argument `b`), exit `2` |

Hub URL override, identical precedence to `outputs get`:

```bash
sonora routes get route-abc-123 --hub-url http://192.168.1.50:9090
MULTIROOM_URL=http://hub.local:8080 sonora routes get route-abc-123
```

## Success output

### Default (YAML)

A single record — not wrapped in a list — with all ten fields always present. Unlike
`routes list`'s five-field view, `routes get` additionally shows `createdAt`, `startedAt`,
`transferable`, `pauseable`, and `paused` (FR-007; see
[research.md §1](../research.md#1-the-route-entity-and-its-split-listget-field-set)):

```yaml
routeId: "route-abc-123"
inputId: "spotify-1"
targetId: "kitchen-speaker"
targetType: "SINGLE_OUTPUT"
status: "ACTIVE"
createdAt: "2026-06-22T14:30:00Z"
startedAt: "2026-06-22T14:30:01Z"
transferable: true
pauseable: true
paused: false
```

A route whose playback has not yet started shows `startedAt: null` explicitly — never
omitted, so its absence can't be mistaken for a decode bug (same convention as `inputs get`'s
`createdAt: null` for a static input):

```yaml
routeId: "route-def-456"
inputId: "spotify-1"
targetId: "whole-house"
targetType: "OUTPUT_GROUP"
status: "STARTING"
createdAt: "2026-06-22T14:35:00Z"
startedAt: null
transferable: false
pauseable: true
paused: false
```

### `--json`

Same fields, strict JSON, parseable by any standard JSON parser (SC-005):

```json
{
  "routeId": "route-abc-123",
  "inputId": "spotify-1",
  "targetId": "kitchen-speaker",
  "targetType": "SINGLE_OUTPUT",
  "status": "ACTIVE",
  "createdAt": "2026-06-22T14:30:00Z",
  "startedAt": "2026-06-22T14:30:01Z",
  "transferable": true,
  "pauseable": true,
  "paused": false
}
```

## Failure output & exit codes

All failure messages go to stderr; stdout is reserved for the success payload (YAML/JSON) so
scripts piping stdout never have to distinguish success from failure by parsing error text.

| Condition | Exit code | stderr message shape |
|-----------|-----------|-----------------------|
| Missing `<route-id>` | `2` | Usage message stating an identifier is required. |
| Extra positional argument(s) | `2` | Usage message naming the unexpected argument(s). |
| Bad/unknown flag | `2` | Usage message (flag name + expected form). |
| `~/.config/sonora/config.json` exists but is malformed, or `hubUrl` isn't a string | `2` | Usage message naming the config file path and the problem. |
| Hub unreachable / DNS failure / connection refused | `4` | Clear statement that the hub could not be reached, with the hub URL used. |
| Request timeout (>5s, no response) | `4` | Clear statement that the hub did not respond in time. |
| Hub returned `404` for the given identifier | `5` | Clear "route not found" message naming the identifier (FR-010) — distinct from both a connectivity failure and a generic hub error. |
| Hub returned another non-2xx (e.g. 5xx) | `3` | Statement that the hub reported an error, distinguishing it from connectivity failure and from "not found" (FR-013). |
| Hub returned 2xx but body doesn't match `RouteResponse` shape (including an unrecognized `targetType`/`status` value) | `3` | Statement that the hub's response was malformed/unexpected (FR-016) — never partial/garbled output. |

With `--verbose`, each of the above additionally prints the underlying Go error after the
friendly message.

## Out of scope for this contract

- Any mutation (create/stop/transfer/pause) — read-only command.
- Authentication headers — spec Assumptions state the hub's read-only endpoints require none,
  consistent with `api/openapi.json`.
- Fuzzy/partial identifier matching — the identifier is opaque and matched exactly as given
  (spec Assumptions).
