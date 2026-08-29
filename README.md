# Sonora Multiroom CLI

A fast, scriptable command-line client for the [Multiroom Audio Hub API](api/openapi.json) —
control real-time audio routing, volume, and mute across your speakers from the terminal.

```
sonora get outputs
```

## Installation

### Scoop (Windows)

```
scoop bucket add sonora https://github.com/Sonora-Multiroom/scoop-bucket
scoop install sonora
```

### Building from source

```
make build
```

Produces `sonora.exe` (or `sonora` on Unix), with the version injected from `git describe`.

No Go toolchain installed? Use `make docker-build` to build the Linux binary inside a
`golang:1.27-alpine` container instead.

## Commands

Usage: `sonora <verb> <resource>[/<id>] [flags]`

Run `sonora help` (or `-h`/`--help`, or `sonora` with no arguments) to print this command
table, common flags, and examples from the terminal.

| Verb | Resources | Aliases |
| --- | --- | --- |
| `get <resource>[/<id>]` | `inputs`, `outputs`, `groups`, `routes` | `in`, `out`, `gr`, `rt` |
| `list <resource>` | same four, collection form only | same |
| `delete routes/<id>` | `routes` only | `rt` |
| `stop routes/<id>` | alias of `delete routes/<id>` | `rt` |
| `pause routes/<id>` | `routes` only | `rt` |
| `resume routes/<id>` | `routes` only | `rt` |
| `enable <resource>/<id>` | `inputs`, `outputs`, `groups` | `in`, `out`, `gr` |
| `disable <resource>/<id>` | `inputs`, `outputs`, `groups` | `in`, `out`, `gr` |
| `set <resource>/<id> volume <0-100>` | `outputs`, `groups` | `out`, `gr` |

`get <resource>` (no id) and `list <resource>` return the collection; `get <resource>/<id>`
returns a single item by id. `list` is an exact synonym of `get` for the collection form —
`list <resource>/<id>` is a usage error, since `list` never takes an id. Aliases are
interchangeable with full resource names everywhere a resource path appears, e.g.
`sonora get out/office-speaker` is identical to `sonora get outputs/office-speaker`.

`delete routes/<id>` stops playback and removes the route; `stop routes/<id>` is an exact
alias. Only `routes` supports deletion today.

`pause routes/<id>` and `resume routes/<id>` pause or resume an active route's playback and
print the updated `routeId`/`paused`/`status` confirmation. Both are idempotent — pausing an
already-paused route (or resuming an already-active one) still succeeds. Only `routes`
supports pause/resume today.

`enable <resource>/<id>` and `disable <resource>/<id>` set the input's, output's, or group's
enabled state; a disabled input remains registered but is unavailable for new route creation.
`inputs`, `outputs`, and `groups` support enable/disable today.

`set outputs/<id> volume <0-100>` and `set groups/<id> volume <0-100>` set the output's or
group's volume level and print the applied `outputId`/`groupId`/`volume`/`updatedAt`
confirmation. Only `outputs`/`volume` and `groups`/`volume` support `set` today.

`transfer routes/<route-id> <outputs|groups>/<target-id>` seamlessly moves an active route's
playback to a new output or group without interruption. The hub replaces the old route with a
new one, so the printed `routeId` is the *new* route's id — update any stored references
accordingly.

`play` wraps a single hub operation: instant playback of an audio URI to an output or output
group, creating the ephemeral input and route in one call —
`sonora play <uri> <outputs|groups>/<id>`. It returns as soon as the hub accepts the request
(no polling for the route to become active). The target's type is given directly by the
`outputs/`/`groups/` path prefix. `--volume N` (0-100) sets the starting volume, and
`--display-name NAME` sets the ephemeral input's display name.

Every command supports `--json` (strict JSON instead of the default YAML), `--hub-url`
(override the hub base URL), and `--verbose` (print underlying error detail on failure).
These are per-command flags, so write them **after** the resource path — `sonora get outputs
--json`, not `sonora --json get outputs`. Only `--version`/`-v` and `--help`/`-h` are
accepted bare, with no verb.

`get`/`list` additionally support `--include-disabled` for inputs/outputs/groups;
`get`/`list routes` also supports `--input-id`, `--target-id`, and `--status` filters
(`--status` takes one of `STARTING`, `ACTIVE`, `STOPPING`, `STOPPED`, `FAILED`, matched
case-insensitively). Run `sonora get <resource> --help` or `sonora list <resource> --help`
for the full flag reference of any command.

```
sonora get outputs --include-disabled
sonora get routes --status active
sonora get groups/<id> --json
sonora play "https://stream.example.com/live.mp3" outputs/office-speaker --volume 40
sonora transfer routes/<route-id> outputs/bedroom-speaker
```

## Configuration

The hub base URL is resolved in order of precedence:

1. `--hub-url` flag
2. `MULTIROOM_URL` environment variable
3. `hubUrl` field in `~/.config/sonora/config.json`
4. default (`http://localhost:8080`)

## Status

Early development. The project follows a spec-first workflow: features are specified,
planned, and implemented one at a time — see [`specs/`](specs/) for what's in progress.

## Design principles

- **Instant startup** — no perceptible lag between launch and the first request.
- **API-contract fidelity** — every request/response is validated against `api/openapi.json`.
- **Resilient HTTP client** — explicit timeouts, no hangs, clear errors on failure.
- **Consistent UX** — YAML output by default, `--json` for scripting, predictable exit codes.

Full rationale lives in the [project constitution](.specify/memory/constitution.md).

## License

[GNU Affero General Public License v3.0](LICENSE)
