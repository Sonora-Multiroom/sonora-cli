# Contract: `sonora outputs list`

**Feature**: `001-list-outputs` | **Date**: 2026-08-24

This documents the observable interface of the command — inputs, outputs, and exit codes —
as the contract implementation and tests are written against. The underlying HTTP contract
(`GET /api/v2/outputs`) is defined by `api/openapi.json` and is not duplicated here beyond
what's needed to explain CLI behavior.

## Invocation

```text
sonora outputs list [--include-disabled] [--json] [--verbose] [--hub-url URL]
```

No positional arguments. Any unrecognized flag or extra positional argument is a usage
error (exit `2`).

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--include-disabled` | bool | `false` | Include disabled outputs in the results (FR-003). |
| `--json` | bool | `false` | Emit strict JSON instead of the default YAML (FR-007). |
| `--verbose` | bool | `false` | On failure, print the underlying error detail in addition to the friendly message (Principle IV). |
| `--hub-url` | string | `http://localhost:8080`, overridable via `$MULTIROOM_URL`, then `~/.config/sonora/config.json`'s `hubUrl` field, then this flag (highest precedence) | Hub base URL override (see research.md §5). |

Underlying HTTP call: `GET {hub-url}/api/v2/outputs?includeDisabled={true|false}`, single
attempt, 5s total timeout, no retries.

## Example invocations

The three boolean flags are independent, giving 8 combinations; `--hub-url` (or its
env/config-file equivalents) can be added to any of them.

| # | Command | Result |
|---|---------|--------|
| 1 | `sonora outputs list` | Enabled outputs only, YAML, friendly errors only |
| 2 | `sonora outputs list --include-disabled` | + disabled outputs included |
| 3 | `sonora outputs list --json` | Enabled only, JSON |
| 4 | `sonora outputs list --verbose` | Enabled only, YAML; raw error appended on failure |
| 5 | `sonora outputs list --include-disabled --json` | Disabled included, JSON |
| 6 | `sonora outputs list --include-disabled --verbose` | Disabled included, YAML, raw error on failure |
| 7 | `sonora outputs list --json --verbose` | JSON, raw error on failure |
| 8 | `sonora outputs list --include-disabled --json --verbose` | Everything on |

Hub URL, independent of the above (see [research.md §5](../research.md) for the full
precedence chain):

```bash
# explicit flag (highest precedence)
sonora outputs list --hub-url http://192.168.1.50:9090

# env var (used when no flag given)
MULTIROOM_URL=http://hub.local:8080 sonora outputs list

# ~/.config/sonora/config.json containing {"hubUrl": "http://hub.local:8080"}
# — used when neither flag nor env is given
sonora outputs list

# fully combined
sonora outputs list --include-disabled --json --verbose --hub-url http://hub.local:8080
```

Note the Go `flag` package convention for booleans: `--verbose` or `--verbose=false` work;
`--verbose true` does not (`true` is parsed as a stray positional argument → exit `2`).

## Success output

### Default (YAML)

One record per returned output, in the order the hub returned them (no client-side
sorting — spec Assumptions). Zero outputs is not an error (FR-012): the command prints an
explicit "no outputs" indication and exits `0`.

```yaml
outputs:
  - outputId: "office-speaker"
    displayName: "Office Speaker"
    volume: 75
    muted: false
    available: true
    enabled: true
  - outputId: "garage-speaker"
    displayName: "Garage Speaker"
    volume: 40
    muted: true
    available: false
    enabled: true
```

An output whose `available: false` (FR-005) MUST be visibly distinguishable from an
available one in this default rendering (e.g. an explicit `available: false` line is
always present, never omitted, so its absence can't be mistaken for "available").

Zero-outputs case:

```yaml
outputs: []
```

with the CLI additionally emitting a plain-text human note (outside the YAML document, or
as a comment) that zero outputs were found — exact wording is an implementation detail,
not a contract point, but it MUST be unambiguous that this is success-with-no-data, not
failure (SC-004).

### `--json`

Same fields, strict JSON, parseable by any standard JSON parser (SC-003):

```json
{
  "outputs": [
    {
      "outputId": "office-speaker",
      "displayName": "Office Speaker",
      "volume": 75,
      "muted": false,
      "available": true,
      "enabled": true
    }
  ]
}
```

Zero outputs: `{"outputs": []}`.

## Failure output & exit codes

All failure messages go to stderr; stdout is reserved for the success payload (YAML/JSON)
so scripts piping stdout never have to distinguish success from failure by parsing error
text.

| Condition | Exit code | stderr message shape |
|-----------|-----------|-----------------------|
| Bad/unknown flag | `2` | Usage message (flag name + expected form). |
| `~/.config/sonora/config.json` exists but is malformed JSON, or `hubUrl` isn't a string | `2` | Usage message naming the config file path and the problem. |
| Hub unreachable / DNS failure / connection refused | `4` | Clear statement that the hub could not be reached, with the hub URL used. |
| Request timeout (>5s, no response) | `4` | Clear statement that the hub did not respond in time. |
| Hub returned non-2xx (e.g. 5xx) | `3` | Statement that the hub reported an error, distinguishing it from a connectivity failure (FR-010). |
| Hub returned 2xx but body doesn't match `OutputResponse[]` shape | `3` | Statement that the hub's response was malformed/unexpected (FR-013) — never partial/garbled output. |

With `--verbose`, each of the above additionally prints the underlying Go error (e.g. the
raw `net`/`http`/`json` error) after the friendly message.

## Out of scope for this contract

- Sorting/filtering beyond `--include-disabled` (not required by the spec).
- Any mutation (volume/mute/enable) — read-only command.
- Authentication headers — spec Assumptions state the hub's read-only endpoints require
  none, consistent with `api/openapi.json`.
