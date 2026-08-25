# Contract: `sonora routes list`

**Feature**: `004-routes-list-get` | **Date**: 2026-08-25

This documents the observable interface of the command — inputs, outputs, and exit codes —
as the contract implementation and tests are written against. The underlying HTTP contract
(`GET /api/v2/routes`, operationId `listRoutes`) is defined by `api/openapi.json` and is not
duplicated here beyond what's needed to explain CLI behavior.

## Invocation

```text
sonora routes list [--status STATUS] [--input-id ID] [--target-id ID] [--json] [--verbose] [--hub-url URL]
```

No positional arguments. Any unrecognized flag or unexpected positional argument is a usage
error (exit `2`).

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--status` | string | `""` (no filter) | Only return routes with this status (`STARTING`/`ACTIVE`/`STOPPING`/`STOPPED`/`FAILED`). Not validated client-side — an unrecognized value is forwarded to the hub, which rejects it with `400` (FR-003). |
| `--input-id` | string | `""` (no filter) | Only return routes sourced from this input identifier (FR-003). |
| `--target-id` | string | `""` (no filter) | Only return routes pointed at this target identifier (FR-003). |
| `--json` | bool | `false` | Emit strict JSON instead of the default YAML (FR-009). |
| `--verbose` | bool | `false` | On failure, print the underlying error detail in addition to the friendly message (Principle IV). |
| `--hub-url` | string | `http://localhost:8080`, overridable via `$MULTIROOM_URL`, then `~/.config/sonora/config.json`'s `hubUrl` field, then this flag (highest precedence) | Hub base URL override — identical precedence to `outputs list`. |

Underlying HTTP call: `GET {hub-url}/api/v2/routes[?status=...][&inputId=...][&targetId=...]`
— only the filters actually supplied are included as query parameters — single attempt, 5s
total timeout, no retries — same client construction as `outputs list`/`inputs list`.

## Example invocations

| # | Command | Result |
|---|---------|--------|
| 1 | `sonora routes list` | All routes regardless of status, YAML |
| 2 | `sonora routes list --status FAILED` | Only `FAILED` routes, YAML |
| 3 | `sonora routes list --input-id spotify-1` | Only routes sourced from `spotify-1`, YAML |
| 4 | `sonora routes list --target-id kitchen-speaker` | Only routes targeting `kitchen-speaker`, YAML |
| 5 | `sonora routes list --status ACTIVE --target-id kitchen-speaker` | Only routes matching both filters (AND logic), YAML |
| 6 | `sonora routes list --json` | All routes, JSON |
| 7 | `sonora routes list --verbose` | YAML; raw error appended on failure |

Hub URL override, identical precedence to `outputs list`:

```bash
sonora routes list --hub-url http://192.168.1.50:9090
MULTIROOM_URL=http://hub.local:8080 sonora routes list
```

## Success output

### Default (YAML)

Unlike `outputs list`/`inputs list`, only five fields are shown per route — the fields
`routes list` displays are a strict subset of what `routes get` shows (FR-004 vs. FR-007; see
[research.md §1](../research.md#1-the-route-entity-and-its-split-listget-field-set)):

```yaml
routes:
  - routeId: "route-abc-123"
    inputId: "spotify-1"
    targetId: "kitchen-speaker"
    targetType: "SINGLE_OUTPUT"
    status: "ACTIVE"
  - routeId: "route-def-456"
    inputId: "spotify-1"
    targetId: "whole-house"
    targetType: "OUTPUT_GROUP"
    status: "FAILED"
```

Zero routes (FR-015):

```yaml
# no routes found
routes: []
```

### `--json`

Same fields, strict JSON, parseable by any standard JSON parser (SC-005):

```json
{
  "routes": [
    {
      "routeId": "route-abc-123",
      "inputId": "spotify-1",
      "targetId": "kitchen-speaker",
      "targetType": "SINGLE_OUTPUT",
      "status": "ACTIVE"
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
| Hub returned `400` (e.g. invalid `--status` value) | `3` | Statement that the hub reported an error, distinguishing it from a connectivity failure. |
| Hub returned another non-2xx (e.g. 5xx) | `3` | Statement that the hub reported an error, distinguishing it from a connectivity failure. |
| Hub returned 2xx but body doesn't match `RouteResponse[]` shape (including an unrecognized `targetType`/`status` value) | `3` | Statement that the hub's response was malformed/unexpected (FR-016) — never partial/garbled output. |

With `--verbose`, each of the above additionally prints the underlying Go error after the
friendly message.

## Out of scope for this contract

- Any mutation (create/stop/transfer/pause) — read-only command.
- Authentication headers — spec Assumptions state the hub's read-only endpoints require none,
  consistent with `api/openapi.json`.
- Sort ordering — routes are presented in the order the hub returns them (spec Assumptions).
