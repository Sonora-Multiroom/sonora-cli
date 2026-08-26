# Sonora Multiroom CLI

A fast, scriptable command-line client for the [Multiroom Audio Hub API](api/openapi.json) —
control real-time audio routing, volume, and mute across your speakers from the terminal.

```
sonora outputs list
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

## Commands

Usage: `sonora <noun> <verb> [flags]`

Run `sonora help` (or `-h`/`--help`, or `sonora` with no arguments) to print this command
table, global flags, and examples from the terminal.

| Noun | Verbs |
| --- | --- |
| `outputs` | `list`, `get` |
| `inputs` | `list`, `get` |
| `routes` | `list`, `get` |
| `groups` | `list`, `get` |

`play` is verb-less — `sonora play <uri> <target-id>` — since it wraps a single hub
operation: instant playback of an audio URI to an output or output group, creating the
ephemeral input and route in one call. It returns as soon as the hub accepts the request
(no polling for the route to become active). The target type is auto-detected by default;
`--group`/`--output` force it explicitly when an identifier collides across both. `--volume N`
(0-100) sets the starting volume, and `--name NAME` sets the ephemeral input's display name.

Every command supports `--json` (strict JSON instead of the default YAML), `--hub-url`
(override the hub base URL), and `--verbose` (print underlying error detail on failure).
`list` commands additionally support `--include-disabled`; `routes list` also supports
`--input-id`, `--target-id`, and `--status` filters. Run `sonora <noun> <verb> --help` for
the full flag reference of any command.

```
sonora outputs list --include-disabled
sonora routes list --status active
sonora groups get <id> --json
sonora play "https://stream.example.com/live.mp3" office-speaker --volume 40
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
