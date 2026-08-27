# CLI Contract: `sonora route`

## Invocation

```text
sonora route inputs/<input-id> <outputs|groups>/<target-id> [--json] [--verbose] [--hub-url URL]
```

Aliases: `in/<input-id>` for the input path; `out/<target-id>` / `gr/<target-id>` for the target
path (same aliases `internal/cli/respath` already registers for `get`/`list`/`play`).

- `inputs/<input-id>` and `<outputs|groups>/<target-id>` are positional and both required
  (FR-002); either may precede or follow flags (same re-parse-loop handling `play` uses for its
  two positionals).
- No `--volume`/`--display-name`/`--group`/`--output` flags — this command does not create an
  input or resolve an ambiguous identifier; the target's type is read entirely from its path
  prefix (spec Assumptions, FR-003).
- `--json`, `--verbose`, `--hub-url` — same meaning and defaults as every other command.

## Argument validation (before any hub call)

| Condition | Exit code | Example message |
|---|---|---|
| Fewer than 2 positional arguments | 2 | `missing required argument: <input-path>` / `<target-path>` |
| More than 2 positional arguments | 2 | `unexpected argument(s): [...]` |
| Input path's prefix is not `inputs`/`in` | 2 | `input path must start with inputs/ or in/, got "outputs/x"` |
| Input path has no identifier (e.g. bare `inputs`) | 2 | `input path must include an id` |
| Target path's prefix is not `outputs`/`out`/`groups`/`gr` | 2 | `route target must be outputs/<id> or groups/<id>, got "inputs/x"` |
| Target path has no identifier (e.g. bare `outputs`) | 2 | `target path must include an id` |

## Success output (exit 0)

Fields, in order, identical between YAML (default) and JSON (`--json`) per FR-005:

| Field | Description |
|---|---|
| `routeId` | the created route's identifier, exactly as returned by the hub |
| `status` | the route's status exactly as returned by the hub (typically `STARTING`) |
| `message` | a CLI-constructed human-readable confirmation (data-model.md's `RoutingResult`) |

YAML example:

```yaml
routeId: "route_abc123"
status: "STARTING"
message: "Routed inputs/spotify-1 to outputs/office-speaker."
```

JSON example (`--json`):

```json
{"routeId":"route_abc123","status":"STARTING","message":"Routed inputs/spotify-1 to outputs/office-speaker."}
```

## Failure output (exit non-zero)

All failures are written to stderr, never stdout, in the form:

```text
error: <friendly message> (hub URL: <resolved base URL>)
```

with `detail: <raw error>` appended on a second line only when `--verbose` is given. See
[data-model.md](../data-model.md#exit-code-table-full-this-features-additions-marked-new) for
the full exit code table. Representative cases:

| Scenario | Exit code | Example friendly message |
|---|---|---|
| Missing/extra positional arguments, invalid path prefix | 2 | see argument validation table above |
| Input does not exist | 11 | `input not found: <id>` |
| Target (of the stated type) does not exist | 12 | `output not found: <id>` / `group not found: <id>` |
| Hub `400` on route creation (bad request field) | 6 | hub's `detail` message, or a generic fallback if absent |
| Hub `422` on route creation (e.g. disabled input, target at capacity) | 8 | hub's `detail` message, or `route creation failed` |
| Hub unreachable / timeout | 4 | `could not reach the hub` / `hub did not respond in time` |
| Malformed 201 response | 3 | `hub returned an unexpected or malformed response` |
| Any other unexpected non-2xx | 3 | `hub reported an error (HTTP <code>)` |

## Out of scope

Managing the created route afterward (stopping, pausing, transferring) is explicitly not
performed by this command (spec Assumptions) — a future route-management command is the
documented follow-up, addressed by the returned `routeId`.
