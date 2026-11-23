# Contract: `sonora play` (updated target/flag shape)

This documents only what changes from the existing
[006-play-command/contracts/cli-play.md](../../006-play-command/contracts/cli-play.md)
contract. Everything not mentioned here (`<uri>`, `--volume`, `--json`, `--verbose`,
`--hub-url`, success output fields, hub-error exit codes `3`/`4`/`6`/`8`/`9`/`10`) is
unchanged.

## Invocation

```text
sonora play <uri> <outputs|groups>/<target-id> [--display-name NAME] [--volume N]
            [--json] [--verbose] [--hub-url URL]
```

- `<outputs|groups>/<target-id>` replaces the old `<target-id>` positional plus
  `--group`/`--output` flags (FR-007). Aliases `out/`/`gr/` are accepted (FR-004).
- `--display-name` replaces `--name` (FR-008); identical behavior, new name only.
- `--group`, `--output`, `--name` are no longer registered flags.

## Example invocations

| # | Command | Equivalent today | Result |
|---|---|---|---|
| 1 | `sonora play <uri> outputs/office-speaker` | `sonora play <uri> office-speaker --output` | Playback starts on that output |
| 2 | `sonora play <uri> groups/whole-house` | `sonora play <uri> whole-house --group` | Playback starts on that group |
| 3 | `sonora play <uri> out/office-speaker --display-name "Radio"` | `sonora play <uri> office-speaker --output --name "Radio"` | Same, with display name set |
| 4 | `sonora play <uri> office-speaker --output` | (old form) | Usage error, exit `2` — `--output` is not a defined flag |
| 5 | `sonora play <uri> office-speaker --name Radio` | (old form) | Usage error, exit `2` — `--name` is not a defined flag |
| 6 | `sonora play <uri> inputs/some-input` | — | Usage error, exit `2` — target must be `outputs`/`groups` |
| 7 | `sonora play <uri> groups/does-not-exist` | `sonora play <uri> does-not-exist --group` | "group not found", exit `5` |

## Removed case: ambiguous target (exit code `7`)

The old contract's "target matches both an output and a group; use --group or --output to
disambiguate" case (exit `7`) is removed. Path-style addressing makes the ambiguity it
handled structurally impossible: `outputs/<id>` and `groups/<id>` are always distinct
arguments, so an identifier shared by an output and a group is resolved by which prefix the
caller wrote, not by inspecting both and detecting a collision (spec Clarifications).

## Out of scope

Unchanged from the original contract — no polling for a terminal route state; see
[006-play-command/contracts/cli-play.md § Out of scope](../../006-play-command/contracts/cli-play.md#out-of-scope).
