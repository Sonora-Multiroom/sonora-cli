# Main Implementation Plan

> **Revision**: 2026-11-23 — Seeded from specs/001-list-outputs (first archival): summary, Technical Context, project structure, hub-URL configuration, exit codes, testing strategy. Describes the layout as planned by that feature.

## Summary

`sonora outputs list` is the CLI's first command: a read-only `GET /api/v2/outputs` call
(optionally with `?includeDisabled=true`) rendered as YAML by default or JSON with
`--json`, each output showing identifier, display name, volume, mute state, enabled state,
and hardware availability. It also establishes the reusable skeleton later commands build
on: a timeout-bound, single-attempt HTTP client (Principle IV), openapi.json-validated
request/response types (Principle II), and dual human/`--json` rendering with
class-distinguishing exit codes (Principle V), all on the Go standard library only.
[Source: .specify/archive/001-list-outputs/plan.md -> "Summary"]

## Technical Context

**Language/Version**: Go 1.27.0 (module-based; `go.mod` created by the list-outputs feature).
[Source: .specify/archive/001-list-outputs/plan.md -> "Language/Version"]

**Primary Dependencies**: Go standard library only — `net/http`, `encoding/json`, `flag`,
`net/http/httptest` (tests). No third-party module; a small fixed-shape YAML emitter in
`internal/render` replaces a YAML library because every output field is a flat scalar.
[Source: .specify/archive/001-list-outputs/plan.md -> "Primary Dependencies"]

**Storage**: A single optional local config file, `~/.config/sonora/config.json` (JSON), read
(never written) as the third-priority source for the hub base URL. No database.
[Source: .specify/archive/001-list-outputs/plan.md -> "Storage"]

**Testing**: `go test` (stdlib `testing`) in three layers — unit (`tests/unit`), contract
(`tests/contract`, against an `httptest.Server` shaped from `api/openapi.json`), and
integration (`tests/integration`, full CLI invocation). Written before implementation per
Principle VI. [Source: .specify/archive/001-list-outputs/plan.md -> "Testing"]

**Target Platform**: Cross-platform CLI binary (Linux/macOS/Windows), single static Go binary
via `go build`. [Source: .specify/archive/001-list-outputs/plan.md -> "Target Platform"]

**Project Type**: Single project — CLI plus internal packages.
[Source: .specify/archive/001-list-outputs/plan.md -> "Project Type"]

**Performance Goals**: Cold start to first HTTP request dispatched well under 50ms; full
output list rendered in under 1s under normal network conditions.
[Source: .specify/archive/001-list-outputs/plan.md -> "Performance Goals"]

**Constraints**: 5s total request timeout, single attempt, no automatic retry on
network/timeout failure; no unbounded waits; exit codes distinguish success / usage error /
hub error / network error. [Source: .specify/archive/001-list-outputs/plan.md -> "Constraints"]

**Scale/Scope**: One command (`sonora outputs list`) with `--include-disabled`, `--json`,
`--verbose` and a `--hub-url` override. [Source: .specify/archive/001-list-outputs/plan.md -> "Scale/Scope"]

## Project Structure

### Source Code (repository root)

```text
cmd/
└── sonora/
    └── main.go                    # Entrypoint: top-level dispatch (`outputs` noun → `list` verb), exit code translation

internal/
├── hub/                           # Hub API client — validated against api/openapi.json (Principle II)
│   ├── client.go                  # HTTP client construction (timeout, single attempt) — Principle IV
│   ├── outputs.go                 # ListOutputs(ctx, includeDisabled) — GET /api/v2/outputs
│   └── errors.go                  # Error classification: usage / hub / network
├── config/
│   └── config.go                  # Read-only load of ~/.config/sonora/config.json; hub URL precedence resolution
├── cli/
│   └── outputs/
│       └── list.go                # `sonora outputs list`: flag definitions, dispatch, exit code mapping
└── render/
    └── outputs.go                 # YAML (default, fixed-shape emitter) / JSON (--json) rendering

tests/
├── contract/
│   └── outputs_list_test.go       # httptest.Server shaped from openapi.json OutputResponse
├── integration/
│   └── outputs_list_test.go       # Full `sonora outputs list` invocation against the mock server
└── unit/
    ├── hub_client_test.go         # Timeout, single-attempt, error classification
    ├── config_test.go             # Hub URL precedence, malformed config file handling
    └── render_outputs_test.go     # YAML/JSON rendering, zero-outputs, unavailable-output distinguishability
```

Go layout: `cmd/` for the entrypoint, `internal/` for non-exported packages. `internal/hub`
is the service layer, `internal/render` the reusable formatting layer, `internal/config`
the read-only config pair, `internal/cli/outputs` the per-noun flag parsing and dispatch.
[Source: .specify/archive/001-list-outputs/plan.md -> "Structure Decision"]

## Configuration

- **Hub base URL**, highest precedence first: `--hub-url` flag; `MULTIROOM_URL` environment
  variable; `hubUrl` field of `~/.config/sonora/config.json`; built-in default
  `http://localhost:8080` (the `servers[0].url` in `api/openapi.json`). The config file is
  read lazily, only when the URL must be resolved, after flag parsing. A missing file falls
  through silently; malformed JSON or a non-string `hubUrl` is a usage error (exit `2`)
  naming the file. Writing the file is out of scope. [Source: .specify/archive/001-list-outputs/research.md -> "5. Hub location resolution"]
- **Config file** `~/.config/sonora/config.json`: optional; field `hubUrl` (string).
  [Source: .specify/archive/001-list-outputs/data-model.md -> "Config file"]

## Routing & Navigation

- `sonora outputs list [--include-disabled] [--json] [--verbose] [--hub-url URL]` → `GET {hub-url}/api/v2/outputs?includeDisabled={true|false}`; single attempt, 5s timeout, no retries. No positional arguments; unknown flags or extra positionals are a usage error (exit `2`).
  [Source: .specify/archive/001-list-outputs/contracts/cli-outputs-list.md -> "Invocation"]
- **Output**: YAML by default (a top-level `outputs:` list; string scalars always double-quoted; `available` always printed; zero outputs prints `outputs: []` plus a plain-text note), strict JSON `{"outputs": [...]}` with `--json`. Success payload on stdout, failures on stderr; `--verbose` appends the raw underlying error.
  [Source: .specify/archive/001-list-outputs/contracts/cli-outputs-list.md -> "Success output"]
- **Exit codes**: `0` success (including zero outputs), `2` usage error, `3` hub-reported error (non-2xx, or 2xx with a malformed body), `4` network/connectivity error (unreachable, DNS, timeout).
  [Source: .specify/archive/001-list-outputs/research.md -> "6. Exit code scheme"]
- **Errors**: `internal/hub` classifies stdlib errors into the three failure classes with short actionable messages; raw Go errors appear only with `--verbose`.
  [Source: .specify/archive/001-list-outputs/research.md -> "7. Error message translation"]

## Testing Strategy

- **Unit** (`tests/unit`): `internal/hub` request construction and decoding; `internal/render` YAML/JSON output including the zero-outputs and unavailable-output cases.
- **Contract** (`tests/contract`): an `httptest.Server` serving `GET /api/v2/outputs` shaped from `api/openapi.json` (including `includeDisabled` and the `OutputResponse` fields); asserts the request and rejection of a malformed body.
- **Integration** (`tests/integration`): full `sonora outputs list` invocation against the same mock server, asserting stdout format, exit codes and include-disabled behavior.

All written before implementation (Principle VI). `gofmt` and `go vet` are the baseline checks.
[Source: .specify/archive/001-list-outputs/research.md -> "8. Testing strategy"]
