# Main Implementation Plan

> **Revision**: 2026-11-23 — Seeded from specs/001-list-outputs (first archival): summary, Technical Context, project structure, hub-URL configuration, exit codes, testing strategy. Describes the layout as planned by that feature.
> **Revision**: 2026-11-23 — Archived specs/002-outputs-get: added the `outputs get` command to Summary, Technical Context, Project Structure, Routing & Navigation (fifth exit code `5`) and Testing Strategy; `outputs.Run` is now `outputs.RunList`. No Technical Context value conflicted with the 001 values.

## Summary

`sonora outputs list` is the CLI's first command: a read-only `GET /api/v2/outputs` call
(optionally with `?includeDisabled=true`) rendered as YAML by default or JSON with
`--json`, each output showing identifier, display name, volume, mute state, enabled state,
and hardware availability. It also establishes the reusable skeleton later commands build
on: a timeout-bound, single-attempt HTTP client (Principle IV), openapi.json-validated
request/response types (Principle II), and dual human/`--json` rendering with
class-distinguishing exit codes (Principle V), all on the Go standard library only.
[Source: .specify/archive/001-list-outputs/plan.md -> "Summary"]

`sonora outputs get <output-id>` is the single-item counterpart: a read-only
`GET /api/v2/outputs/{outputId}` call (operationId `getOutput`) rendered as one YAML record
(or JSON with `--json`), returned regardless of the output's enabled state. It reuses the
`Output` type, `hub.NewClient()`, hub-URL resolution, flag names and rendering approach
unchanged; the new pieces are `GetOutput`, a `NotFoundError`/`ClassNotFound` (exit `5`) class
for 404, and single-record rendering.
[Source: .specify/archive/002-outputs-get/plan.md -> "Summary"]

## Technical Context

**Language/Version**: Go 1.27.0 (module-based; `go.mod` created by the list-outputs feature).
[Source: .specify/archive/001-list-outputs/plan.md -> "Language/Version"; .specify/archive/002-outputs-get/plan.md -> "Language/Version"]

**Primary Dependencies**: Go standard library only — `net/http`, `encoding/json`, `flag`,
`net/url`, `net/http/httptest` (tests). No third-party module; a small fixed-shape YAML emitter in
`internal/render` replaces a YAML library because every output field is a flat scalar.
[Source: .specify/archive/001-list-outputs/plan.md -> "Primary Dependencies"; .specify/archive/002-outputs-get/plan.md -> "Primary Dependencies"]

**Storage**: A single optional local config file, `~/.config/sonora/config.json` (JSON), read
(never written) as the third-priority source for the hub base URL. No database.
[Source: .specify/archive/001-list-outputs/plan.md -> "Storage"; .specify/archive/002-outputs-get/plan.md -> "Storage"]

**Testing**: `go test` (stdlib `testing`) in three layers — unit (`tests/unit`), contract
(`tests/contract`, against an `httptest.Server` shaped from `api/openapi.json`), and
integration (`tests/integration`, full CLI invocation). Written before implementation per
Principle VI. [Source: .specify/archive/001-list-outputs/plan.md -> "Testing"; .specify/archive/002-outputs-get/plan.md -> "Testing"]

**Target Platform**: Cross-platform CLI binary (Linux/macOS/Windows), single static Go binary
via `go build`. [Source: .specify/archive/001-list-outputs/plan.md -> "Target Platform"; .specify/archive/002-outputs-get/plan.md -> "Target Platform"]

**Project Type**: Single project — CLI plus internal packages.
[Source: .specify/archive/001-list-outputs/plan.md -> "Project Type"; .specify/archive/002-outputs-get/plan.md -> "Project Type"]

**Performance Goals**: Cold start to first HTTP request dispatched well under 50ms; full
output list, and a single output, rendered in under 1s under normal network conditions.
[Source: .specify/archive/001-list-outputs/plan.md -> "Performance Goals"; .specify/archive/002-outputs-get/plan.md -> "Performance Goals"]

**Constraints**: 5s total request timeout, single attempt, no automatic retry on
network/timeout failure; no unbounded waits; exit codes distinguish success / usage error /
hub error / network error and, for lookups by identifier, not-found (exit `5`).
[Source: .specify/archive/001-list-outputs/plan.md -> "Constraints"; .specify/archive/002-outputs-get/plan.md -> "Constraints"]

**Scale/Scope**: Two commands: `sonora outputs list` (`--include-disabled`, `--json`,
`--verbose`, `--hub-url`) and `sonora outputs get <output-id>` (`--json`, `--verbose`,
`--hub-url`; no `--include-disabled`). [Source: .specify/archive/001-list-outputs/plan.md -> "Scale/Scope"; .specify/archive/002-outputs-get/plan.md -> "Scale/Scope"]

## Project Structure

### Source Code (repository root)

```text
cmd/
└── sonora/
    └── main.go                    # Entrypoint: top-level dispatch (`outputs` noun → `list`/`get` verbs), exit code translation

internal/
├── hub/                           # Hub API client — validated against api/openapi.json (Principle II)
│   ├── client.go                  # HTTP client construction (timeout, single attempt) — Principle IV
│   ├── outputs.go                 # ListOutputs(ctx, includeDisabled) — GET /api/v2/outputs; GetOutput(ctx, client, baseURL, outputID) — GET /api/v2/outputs/{outputId}
│   └── errors.go                  # Error classification: usage / hub / network / not found (NotFoundError, ClassNotFound)
├── config/
│   └── config.go                  # Read-only load of ~/.config/sonora/config.json; hub URL precedence resolution
├── cli/
│   └── outputs/
│       ├── list.go                # `sonora outputs list` (RunList): flag definitions, dispatch, exit code mapping
│       └── get.go                 # `sonora outputs get <output-id>` (RunGet): positional id, flag definitions, dispatch, exit code mapping
└── render/
    └── outputs.go                 # YAML (default, fixed-shape emitter) / JSON (--json) rendering; RenderOutputYAML/RenderOutputJSON for a single unwrapped record

tests/
├── contract/
│   ├── outputs_list_test.go       # httptest.Server shaped from openapi.json OutputResponse
│   └── outputs_get_test.go        # httptest.Server shaped from getOutput's 200/404 responses; request path + decode contract
├── integration/
│   ├── outputs_list_test.go       # Full `sonora outputs list` invocation against the mock server
│   └── outputs_get_test.go        # Full `sonora outputs get` invocation: found, not-found, missing-id usage error, --json, exit codes
└── unit/
    ├── cli_outputs_test.go        # Command-level tests (call sites use outputs.RunList)
    ├── hub_client_test.go         # Timeout, single-attempt, error classification
    ├── config_test.go             # Hub URL precedence, malformed config file handling
    ├── render_outputs_test.go     # YAML/JSON rendering, zero-outputs, unavailable-output distinguishability
    └── render_output_get_test.go  # Single-output YAML/JSON rendering, unavailable-output distinguishability
```

Go layout: `cmd/` for the entrypoint, `internal/` for non-exported packages. `internal/hub`
is the service layer, `internal/render` the reusable formatting layer, `internal/config`
the read-only config pair, `internal/cli/outputs` the per-noun flag parsing and dispatch.
[Source: .specify/archive/001-list-outputs/plan.md -> "Structure Decision"]

Adding a second verb to the `outputs` noun required a distinct exported name per verb, so
`outputs.Run` became `outputs.RunList` (a pure rename) beside `outputs.RunGet`.
[Source: .specify/archive/002-outputs-get/plan.md -> "Structure Decision"]

## Configuration

- **Hub base URL**, highest precedence first: `--hub-url` flag; `MULTIROOM_URL` environment
  variable; `hubUrl` field of `~/.config/sonora/config.json`; built-in default
  `http://localhost:8080` (the `servers[0].url` in `api/openapi.json`). The config file is
  read lazily, only when the URL must be resolved, after flag parsing. A missing file falls
  through silently; malformed JSON or a non-string `hubUrl` is a usage error (exit `2`)
  naming the file. Writing the file is out of scope. `outputs get` uses the same precedence.
  [Source: .specify/archive/001-list-outputs/research.md -> "5. Hub location resolution"; .specify/archive/002-outputs-get/data-model.md -> "--hub-url"]
- **Config file** `~/.config/sonora/config.json`: optional; field `hubUrl` (string).
  [Source: .specify/archive/001-list-outputs/data-model.md -> "Config file"]

## Routing & Navigation

- `sonora outputs list [--include-disabled] [--json] [--verbose] [--hub-url URL]` → `GET {hub-url}/api/v2/outputs?includeDisabled={true|false}`; single attempt, 5s timeout, no retries. No positional arguments; unknown flags or extra positionals are a usage error (exit `2`).
  [Source: .specify/archive/001-list-outputs/contracts/cli-outputs-list.md -> "Invocation"]
- `sonora outputs get <output-id> [--json] [--verbose] [--hub-url URL]` → `GET {hub-url}/api/v2/outputs/{output-id}` (path-escaped; operationId `getOutput`); single attempt, 5s timeout, no retries. Exactly one positional argument is required (none or more than one is a usage error, exit `2`); no `--include-disabled`, since the output is returned regardless of enabled state. Success output is a single unwrapped record (YAML by default, strict JSON with `--json`), every field always present.
  [Source: .specify/archive/002-outputs-get/contracts/cli-outputs-get.md -> "Invocation"]
- **Output**: YAML by default (a top-level `outputs:` list; string scalars always double-quoted; `available` always printed; zero outputs prints `outputs: []` plus a plain-text note), strict JSON `{"outputs": [...]}` with `--json`. Success payload on stdout, failures on stderr; `--verbose` appends the raw underlying error.
  [Source: .specify/archive/001-list-outputs/contracts/cli-outputs-list.md -> "Success output"]
- **Exit codes**: `0` success (including zero outputs), `2` usage error, `3` hub-reported error (non-2xx other than a get's 404, or 2xx with a malformed body), `4` network/connectivity error (unreachable, DNS, timeout), `5` not found (hub answered `404` for the given output identifier; appended, nothing renumbered).
  [Source: .specify/archive/001-list-outputs/research.md -> "6. Exit code scheme"; .specify/archive/002-outputs-get/research.md -> "2. The 404 \"not found\" case needs its own exit-code class"]
- **Errors**: `internal/hub` classifies stdlib errors into the failure classes (usage, hub, network, and for `get` not-found) with short actionable messages; raw Go errors appear only with `--verbose`.
  [Source: .specify/archive/001-list-outputs/research.md -> "7. Error message translation"]

## Testing Strategy

- **Unit** (`tests/unit`): `internal/hub` request construction and decoding; `internal/render` YAML/JSON output including the zero-outputs and unavailable-output cases.
- **Contract** (`tests/contract`): an `httptest.Server` serving `GET /api/v2/outputs` shaped from `api/openapi.json` (including `includeDisabled` and the `OutputResponse` fields); asserts the request and rejection of a malformed body.
- **Integration** (`tests/integration`): full `sonora outputs list` invocation against the same mock server, asserting stdout format, exit codes and include-disabled behavior.
- **Get command**: a unit test for single-output YAML/JSON rendering (including the unavailable-output case); a contract test whose mock server serves `getOutput`'s `200 OutputResponse` and `404 ErrorResponse` shapes and asserts the request path and both decode paths; an integration test of `sonora outputs get <id>` covering found, not-found, missing-identifier usage error, `--json` and exit codes.
  [Source: .specify/archive/002-outputs-get/research.md -> "6. Testing strategy — same three layers, extended for get"]

All written before implementation (Principle VI). `gofmt` and `go vet` are the baseline checks.
[Source: .specify/archive/001-list-outputs/research.md -> "8. Testing strategy"]
