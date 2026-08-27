# Data Model: Adopt Verb-First Command Landscape

**Feature**: `007-refactor-cli-commands` | **Date**: 2026-08-27

This feature introduces no new hub-facing entities or endpoints — `Input`, `Output`, `Group`,
and `Route` (and their list/get query parameters) are unchanged from
[001-list-outputs](../001-list-outputs/data-model.md),
[002-outputs-get](../002-outputs-get/data-model.md),
[003-inputs-list-get](../003-inputs-list-get/data-model.md),
[004-routes-list-get](../004-routes-list-get/data-model.md), and
[005-groups-list-get](../005-groups-list-get/data-model.md). What's new is purely CLI-surface:
how a command line resolves into a call to that unchanged logic.

## ResourceKind

A closed enum of the four resources this feature's `get`/`list` cover. Not a hub concept — an
`internal/cli/respath` type.

| Canonical name | Alias | Backing package |
|---|---|---|
| `inputs` | `in` | `internal/cli/inputs` |
| `outputs` | `out` | `internal/cli/outputs` |
| `groups` | `gr` | `internal/cli/groups` |
| `routes` | `rt` | `internal/cli/routes` |

An unrecognized name/alias in either position (`sonora get <unrecognized>`, or as the target
prefix in `play`) is a usage error (FR-006), exit code `2`.

## ResourcePath

The parsed form of a command-line argument like `outputs/office-speaker` or `out/office-speaker`.

| Field | Type | Rule |
|---|---|---|
| `Kind` | `ResourceKind` | Resolved from the segment before the first `/` (or the whole argument, if there is no `/`). |
| `ID` | `string` (optional) | The segment after the first `/`, if present. Must match `^[a-zA-Z0-9_-]{1,255}$` (FR-004a, per spec Clarifications). |

**Parsing rule (FR-004a)**: split the argument on the *first* `/` only. Ids are guaranteed
slash-free by the pattern above, so a second `/` in the identifier segment (e.g.
`out/foo/bar`) is unambiguously malformed, not a deeper path — usage error, exit `2`.

**Used by**:
- `sonora get <resource-path>` / `sonora list <resource>` — `Kind` selects which of
  `inputs`/`outputs`/`groups`/`routes`' `RunList`/`RunGet` to call; `ID`, if present, is
  passed through as that function's existing positional-id argument.
- `sonora play <uri> <resource-path>` — `Kind` MUST be `outputs` or `groups` (a `ResourcePath`
  resolving to `inputs`/`routes` here is a usage error); `ID` is required.

## CLI invocation shape — `get` / `list`

See [contracts/cli-get-list.md](contracts/cli-get-list.md) for the full contract.

| Argument/Flag | Applies to | Type | Default | Effect |
|---|---|---|---|---|
| `<resource-path>` (positional) | `get` | `ResourcePath` | none, required | `inputs`/`in`, `outputs`/`out`, `groups`/`gr` accept an optional `/<id>`; `routes`/`rt` likewise. No id → collection form; `/<id>` → single-item form. |
| `<resource>` (positional, no `/`) | `list` | `ResourceKind` | none, required | Collection form only (FR-003) — a `list <resource>/<id>` invocation is a usage error, exit `2`. |
| `--include-disabled` | `get`/`list inputs`, `outputs`, `groups` (collection form only) | bool | `false` | Unchanged from today's `inputs list`/`outputs list`/`groups list` (FR-001). |
| `--status`, `--input-id`, `--target-id` | `get`/`list routes` (collection form only) | string | `""` (unset) | Unchanged from today's `routes list` (FR-001). |
| `--json`, `--verbose`, `--hub-url` | all | — | — | Unchanged (FR-009). |

## CLI invocation shape — `play` (changed fields only)

Full contract: [contracts/cli-play.md](contracts/cli-play.md). Unchanged fields (`<uri>`,
`--volume`, `--json`, `--verbose`, `--hub-url`) are not repeated here — see
[006-play-command/data-model.md](../006-play-command/data-model.md).

| Argument/Flag | Type | Default | Effect |
|---|---|---|---|
| `<target-path>` (positional, replaces old `<target-id>` + `--group`/`--output`) | `ResourcePath` (kind restricted to `outputs`\|`groups`) | none, required | Selects the target and its type in one argument (FR-007). |
| `--display-name` (replaces `--name`) | string | `""` | Sets the new ephemeral input's display name (FR-008). Identical behavior to the old `--name`, renamed only. |

`--group`, `--output`, and `--name` are no longer registered flags; supplying any of them
produces the standard library `flag` package's own "flag provided but not defined" usage
error (exit `2`) — no tailored detection (spec Clarifications).

## Config file

Unchanged — all refactored commands continue to read `~/.config/sonora/config.json` via the
existing `config.ResolveHubURL`, with the same three-layer precedence.

## Exit code classes

| Code | Meaning | Change in this feature? |
|---|---|---|
| `0` | Success | No |
| `2` | Usage error (bad/unknown flag, missing/malformed positional argument, unrecognized resource name/alias, old-style invocation) | Scope widened to cover resource-path/alias errors and the now-removed old command forms — same class, no renumbering. |
| `3` | Hub-reported error (non-2xx other than 404, or malformed response body) | No |
| `4` | Network/connectivity error | No |
| `5` | Not found (hub `404` for a `get <resource>/<id>` lookup, or `play`'s target path naming a type+id that doesn't exist) | No new class — `play`'s per-type "group not found"/"output not found" cases (previously reached via `--group <id>`/`--output <id>`) still land here, just via the target path instead. |
| `6` | `play` validation error (e.g. `--volume` out of range, hub `400`) | No |
| `7` | *(removed)* `play`'s old "target matches both an output and a group; use --group or --output to disambiguate" | **Removed** — path-style addressing makes this case structurally unreachable (spec Clarifications). |
| `8`–`10` | `play`'s other hub-error cases (route creation failed, source unreachable, service unavailable) | No |
