# Data Model: List Audio Outputs

**Feature**: `001-list-outputs` | **Date**: 2026-08-24

## Output

Mirrors `#/components/schemas/OutputResponse` in `api/openapi.json` field-for-field
(Principle II: request/response types MUST be validated against the spec). This is the
only domain entity this feature reads.

| Field | Type | Source field (openapi.json) | Notes |
|-------|------|------------------------------|-------|
| `OutputID` | string | `outputId` | Unique identifier; required, non-empty. |
| `DisplayName` | string | `displayName` | Human-readable name; required, non-empty. |
| `Volume` | int (0–100) | `volume` | Current volume level. Spec doesn't declare min/max on the response object itself, but the domain range (0-100) matches `VolumeRequest`; the CLI does not re-validate bounds — it passes through whatever the hub returns and displays it as-is (Principle II: don't invent stricter behavior than the spec). |
| `Muted` | bool | `muted` | Current mute state. |
| `Available` | bool | `available` | Whether the underlying hardware is currently connected. Drives FR-005's visible distinction in the default view. |
| `Enabled` | bool | `enabled` | Whether the output is enabled for routing. Present in both default and `--include-disabled` results; used to label disabled entries when included. |

**Validation rule (FR-013)**: a decoded element that is missing any of the above fields, or
where a field has the wrong JSON type (e.g. `volume` is a string), is treated as a
malformed response — the whole command fails with the Hub-error exit class (see
[research.md §6](research.md)) rather than rendering partial data. Decoding uses
`json.Decoder` with a struct that has no `omitempty`/pointer fields for these six
properties, so `encoding/json`'s type mismatch errors surface naturally; presence of the
required identifier/name fields is checked explicitly after decode (empty string is
treated as "missing" for `outputId`/`displayName`).

**State/lifecycle**: none — this feature only reads outputs; it defines no transitions,
creates nothing, and holds no state between invocations.

## OutputsListQuery (request shaping)

Not a persisted entity — the parameters this command sends to the hub.

| Field | Type | Maps to | Default |
|-------|------|---------|---------|
| `IncludeDisabled` | bool | `includeDisabled` query param on `GET /api/v2/outputs` | `false` |

## CLI invocation shape

The flags accepted by `sonora outputs list`, and how they affect behavior — see
[contracts/cli-outputs-list.md](contracts/cli-outputs-list.md) for the full contract.

| Flag | Type | Default | Effect |
|------|------|---------|--------|
| `--include-disabled` | bool | `false` | Sets `OutputsListQuery.IncludeDisabled = true`; FR-003. |
| `--json` | bool | `false` | Switches rendering from default YAML to JSON; FR-007. |
| `--verbose` | bool | `false` | On failure, additionally prints the underlying error detail; Principle IV. |
| `--hub-url` | string | `http://localhost:8080`, falling back through `MULTIROOM_URL` env and `~/.config/sonora/config.json`'s `hubUrl` field before that default applies | Overrides the hub base URL; see [research.md §5](research.md). |

## Config file (read-only in this feature)

`~/.config/sonora/config.json` — optional; absence is not an error.

| Field | Type | Required | Notes |
|-------|------|----------|-------|
| `hubUrl` | string | No | Used as the third-priority source for the hub base URL (below `--hub-url` flag and `MULTIROOM_URL` env, above the built-in default). Malformed JSON or a non-string `hubUrl` is a usage error (exit `2`) naming the file. |

This feature reads the file; it does not create, migrate, or write to it. Writing/managing
it (e.g. a future `sonora config set` command) is out of scope here — see
[research.md §5](research.md).

## Exit code classes

See [research.md §6](research.md) for the full table (`0` success, `2` usage error, `3`
hub error, `4` network error). Not a data entity, but part of this feature's observable
contract (FR-011) and reused by every future command.
