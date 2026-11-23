# CLI Contract: `sonora play`

## Invocation

```text
sonora play <uri> <target-id> [--group | --output] [--volume N] [--name NAME]
            [--json] [--verbose] [--hub-url URL]
```

- `<uri>` and `<target-id>` are positional and both required (FR-002); either may precede or
  follow flags (same re-parse-loop handling as `routes get`/`groups get`'s single positional,
  extended to two).
- `--group` and `--output` are boolean, mutually exclusive (FR-002a). Neither given → default
  auto-detection (FR-003).
- `--volume N` — optional integer, `0`-`100` inclusive (FR-004). Out of range → usage/validation
  failure before any request is sent.
- `--name NAME` — optional string, becomes the ephemeral input's display name (FR-005).
- `--json`, `--verbose`, `--hub-url` — same meaning and defaults as every other command.

## Success output (exit 0)

Fields, in order, identical between YAML (default) and JSON (`--json`) per FR-006:

| Field | Description |
|---|---|
| `inputId` | the created ephemeral input's identifier |
| `routeId` | the created route's identifier |
| `status` | the route's status exactly as returned by the hub (typically `STARTING`) |
| `message` | the hub's human-readable confirmation message |

YAML example:

```yaml
inputId: "playback_1782345678"
routeId: "route_abc123"
status: "STARTING"
message: "Playback started: Radio Stream → office-speaker"
```

JSON example (`--json`):

```json
{"inputId":"playback_1782345678","routeId":"route_abc123","status":"STARTING","message":"Playback started: Radio Stream → office-speaker"}
```

## Failure output (exit non-zero)

All failures are written to stderr, never stdout (so `--json` piping never mixes error text
into parsed output), in the form:

```text
error: <friendly message> (hub URL: <resolved base URL>)
```

with `detail: <raw error>` appended on a second line only when `--verbose` is given. See
[data-model.md](../data-model.md#exit-code-table-full-this-features-additions-marked-new) for
the full exit code table. Representative cases:

| Scenario | Exit code | Example friendly message |
|---|---|---|
| Missing `<uri>` or `<target-id>` | 2 | `missing required argument: <uri>` / `<target-id>` |
| Both `--group` and `--output` given | 2 | `--group and --output are mutually exclusive` |
| `--volume` outside 0-100 | 6 | `volume must be between 0 and 100` |
| Hub `400` (bad URI/request field) | 6 | hub's `detail` message, or a generic fallback if absent |
| Target matches neither an output nor a group | 5 | `target not found: <id>` |
| `--group <id>` where `<id>` is only an output | 5 | `group not found: <id>` |
| `--output <id>` where `<id>` is only a group | 5 | `output not found: <id>` |
| Target matches both an output and a group, no `--group`/`--output` | 7 | `target "<id>" matches both an output and a group; use --group or --output to disambiguate` |
| Hub `422` (route creation failed) | 8 | hub's `detail` message, or a generic fallback |
| Hub `502` (URI unreachable) | 9 | `the audio source could not be reached` |
| Hub `503` (upstream service unavailable) | 10 | `the hub's playback service is temporarily unavailable` |
| Hub unreachable / timeout | 4 | `could not reach the hub` / `hub did not respond in time` |
| Malformed 200 response | 3 | `hub returned an unexpected or malformed response` |

## Out of scope

Polling the created route for a terminal state (`ACTIVE`/`FAILED`) is explicitly not
performed by this command (FR-006, SC-003); callers who need that must use a future `routes
get` follow-up against the returned `routeId`.
