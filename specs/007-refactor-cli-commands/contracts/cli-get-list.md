# Contract: `sonora get` / `sonora list` (inputs, outputs, groups, routes)

This documents the observable interface of the new verb-first read commands that replace
today's `sonora inputs list`, `sonora inputs get <id>`, `sonora outputs list`, `sonora outputs
get <id>`, `sonora groups list`, `sonora groups get <id>`, `sonora routes list`, and `sonora
routes get <id>`. The underlying HTTP contracts (`listInputs`, `getInput`, `listOutputs`,
`getOutput`, `listGroups`, `getGroup`, `listRoutes`, `getRoute` in `api/openapi.json`) are
unchanged and not duplicated here — see
[001](../../001-list-outputs/contracts), [002](../../002-outputs-get/contracts),
[003](../../003-inputs-list-get/contracts), [004](../../004-routes-list-get/contracts), and
[005](../../005-groups-list-get/contracts) for those.

## Invocation

```text
sonora get <resource>[/<id>] [--include-disabled | --status S --input-id I --target-id T] [--json] [--verbose] [--hub-url URL]
sonora list <resource> [--include-disabled | --status S --input-id I --target-id T] [--json] [--verbose] [--hub-url URL]
```

`<resource>` is one of `inputs`/`in`, `outputs`/`out`, `groups`/`gr`, `routes`/`rt` — full name
and alias are fully interchangeable (FR-004). `--include-disabled` applies only to
`inputs`/`outputs`/`groups`; `--status`/`--input-id`/`--target-id` apply only to `routes` —
unchanged filter semantics from today's equivalent `list` commands (FR-001).

- `sonora get <resource>` (no `/<id>`) — collection form, identical to today's
  `sonora <resource> list`.
- `sonora get <resource>/<id>` — single-item form, identical to today's
  `sonora <resource> get <id>`.
- `sonora list <resource>` — exact synonym of `sonora get <resource>` (FR-003). Collection
  form only.
- `sonora list <resource>/<id>` — usage error, exit `2`: "list" does not take an id.

## Example invocations

| # | Command | Equivalent today | Result |
|---|---|---|---|
| 1 | `sonora get outputs` | `sonora outputs list` | All outputs, YAML |
| 2 | `sonora list outputs` | `sonora outputs list` | Same as #1 (synonym) |
| 3 | `sonora get outputs/office-speaker` | `sonora outputs get office-speaker` | That output, YAML |
| 4 | `sonora get out/office-speaker` | `sonora outputs get office-speaker` | Same as #3 (alias) |
| 5 | `sonora get routes --status active` | `sonora routes list --status active` | Filtered routes |
| 6 | `sonora list outputs/office-speaker` | — | Usage error, exit `2` |
| 7 | `sonora get bogus` | — | Usage error, exit `2` ("unrecognized resource") |
| 8 | `sonora outputs list` | (old form) | Usage error, exit `2` ("unknown command") — no longer recognized |
| 9 | `sonora get outputs/does-not-exist` | `sonora outputs get does-not-exist` | "output not found", exit `5` |

## Failure output & exit codes

| Condition | Exit code |
|---|---|
| Missing `<resource>`/`<resource-path>` argument (`sonora get`, `sonora list`) | `2` — message MUST enumerate the valid resource names (FR-006a) |
| Unrecognized resource name or alias | `2` |
| Malformed resource path (extra `/`, id not matching `^[a-zA-Z0-9_-]{1,255}$`) | `2` |
| `list` given a resource path that includes an id | `2` |
| Old-style invocation (`sonora <resource> list`/`get`) | `2` — standard "unknown command" error; no migration hint (spec Clarifications) |
| Bad/unknown flag | `2` |
| Hub `404` for a single-item lookup | `5` |
| Hub non-2xx other than 404, or malformed response body | `3` |
| Hub unreachable / timeout | `4` |

All exit-code classes and their stderr message conventions otherwise match the
already-shipped `outputs get`/`list` contracts referenced above — this feature changes only
which command lines route to that behavior.

## Out of scope for this contract

- Any mutation (`create`/`delete`/`enable`/`disable`/`set ... volume`/`mute`) — not yet
  implemented; tracked separately per `docs/cli-command-landscape.md`.
- `master-mute` — not yet implemented.
- The `route` command — specified separately (`specs/008-route-command`), not yet implemented.
