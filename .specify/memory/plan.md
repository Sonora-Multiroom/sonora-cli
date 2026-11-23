# Main Implementation Plan

> **Revision**: 2026-11-23 — Seeded from specs/001-list-outputs (first archival): summary, Technical Context, project structure, hub-URL configuration, exit codes, testing strategy. Describes the layout as planned by that feature.
> **Revision**: 2026-11-23 — Archived specs/002-outputs-get: added the `outputs get` command to Summary, Technical Context, Project Structure, Routing & Navigation (fifth exit code `5`) and Testing Strategy; `outputs.Run` is now `outputs.RunList`. No Technical Context value conflicted with the 001 values.
> **Revision**: 2026-11-23 — Archived .specify/archive/003-inputs-list-get: added `inputs list` and `inputs get` to Summary, Technical Context (Scale/Scope now composes four commands), Project Structure, Configuration, Routing & Navigation and Testing Strategy. `NotFoundError` is now resource-generic (`Resource`, `ID`), so `outputs get` and `inputs get` share one not-found type. No Technical Context value conflicted; the remaining values are identical to the 001/002 ones and each now carries a 003 ref. The main plan has no legacy per-feature blocks, so the layout is unmixed.
> **Revision**: 2026-11-23 — Archived .specify/archive/004-routes-list-get (moved from specs/ in the same run): added `routes list` and `routes get` to Summary, Technical Context (Scale/Scope now composes six commands; Language/Version, Primary Dependencies, Storage, Target Platform, Project Type, Performance Goals and Constraints each take a 004 ref, stating the same values), Project Structure, Configuration, Routing & Navigation (exit `5` now covers route identifiers) and Testing Strategy. No Technical Context value conflicted with the 001–003 values.

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

`sonora inputs list` and `sonora inputs get <input-id>` mirror the outputs pair: read-only
`GET /api/v2/inputs` (operationId `listInputs`) and `GET /api/v2/inputs/{inputId}`
(operationId `getInput`) calls with the same flags, hub-URL resolution, exit-code scheme and
rendering. The new pieces are the `Input` type, whose attributes differ from `Output` (no
volume, mute or availability; instead `uri`, `source`, `autoRemove`, `pauseable` and a
nullable `createdAt`), the `inputs` noun dispatch, and `NotFoundError` now carrying a
`Resource` name so that `inputs get`'s 404 reuses `ClassNotFound` (exit `5`).
[Source: .specify/archive/003-inputs-list-get/plan.md -> "Summary"]

`sonora routes list` and `sonora routes get <route-id>` are read-only `GET /api/v2/routes`
(operationId `listRoutes`) and `GET /api/v2/routes/{routeId}` (operationId `getRoute`) calls,
rendered as YAML by default or JSON with `--json`, reusing `hub.NewClient()`, hub-URL
resolution, the exit-code scheme and the stdout/stderr separation unchanged. The new pieces are
the `Route` type (the list view shows five fields, the get view ten), three AND-combined list
filters (`--status`, `--input-id`, `--target-id`) that the hub validates and applies, and the
`routes` noun dispatch.
[Source: .specify/archive/004-routes-list-get/plan.md -> "Summary"]

## Technical Context

**Language/Version**: Go 1.27.0 (module-based; `go.mod` created by the list-outputs feature).
[Source: .specify/archive/001-list-outputs/plan.md -> "Language/Version"; .specify/archive/002-outputs-get/plan.md -> "Language/Version"; .specify/archive/003-inputs-list-get/plan.md -> "Language/Version"; .specify/archive/004-routes-list-get/plan.md -> "Language/Version"]

**Primary Dependencies**: Go standard library only — `net/http`, `encoding/json`, `flag`,
`net/url`, `net/http/httptest` (tests). No third-party module; a small fixed-shape YAML emitter in
`internal/render` replaces a YAML library because every output field is a flat scalar.
[Source: .specify/archive/001-list-outputs/plan.md -> "Primary Dependencies"; .specify/archive/002-outputs-get/plan.md -> "Primary Dependencies"; .specify/archive/003-inputs-list-get/plan.md -> "Primary Dependencies"; .specify/archive/004-routes-list-get/plan.md -> "Primary Dependencies"]

**Storage**: A single optional local config file, `~/.config/sonora/config.json` (JSON), read
(never written) as the third-priority source for the hub base URL. No database.
[Source: .specify/archive/001-list-outputs/plan.md -> "Storage"; .specify/archive/002-outputs-get/plan.md -> "Storage"; .specify/archive/003-inputs-list-get/plan.md -> "Storage"; .specify/archive/004-routes-list-get/plan.md -> "Storage"]

**Testing**: `go test` (stdlib `testing`) in three layers — unit (`tests/unit`), contract
(`tests/contract`, against an `httptest.Server` shaped from `api/openapi.json`), and
integration (`tests/integration`, full CLI invocation). Written before implementation per
Principle VI. [Source: .specify/archive/001-list-outputs/plan.md -> "Testing"; .specify/archive/002-outputs-get/plan.md -> "Testing"; .specify/archive/003-inputs-list-get/plan.md -> "Testing"]

**Target Platform**: Cross-platform CLI binary (Linux/macOS/Windows), single static Go binary
via `go build`. [Source: .specify/archive/001-list-outputs/plan.md -> "Target Platform"; .specify/archive/002-outputs-get/plan.md -> "Target Platform"; .specify/archive/003-inputs-list-get/plan.md -> "Target Platform"; .specify/archive/004-routes-list-get/plan.md -> "Target Platform"]

**Project Type**: Single project — CLI plus internal packages.
[Source: .specify/archive/001-list-outputs/plan.md -> "Project Type"; .specify/archive/002-outputs-get/plan.md -> "Project Type"; .specify/archive/003-inputs-list-get/plan.md -> "Project Type"; .specify/archive/004-routes-list-get/plan.md -> "Project Type"]

**Performance Goals**: Cold start to first HTTP request dispatched well under 50ms; full
output, input and route list, and a single output, input or route, rendered in under 1s under normal network conditions.
[Source: .specify/archive/001-list-outputs/plan.md -> "Performance Goals"; .specify/archive/002-outputs-get/plan.md -> "Performance Goals"; .specify/archive/003-inputs-list-get/plan.md -> "Performance Goals"; .specify/archive/004-routes-list-get/plan.md -> "Performance Goals"]

**Constraints**: 5s total request timeout, single attempt, no automatic retry on
network/timeout failure; no unbounded waits; exit codes distinguish success / usage error /
hub error / network error and, for lookups by identifier, not-found (exit `5`).
[Source: .specify/archive/001-list-outputs/plan.md -> "Constraints"; .specify/archive/002-outputs-get/plan.md -> "Constraints"; .specify/archive/003-inputs-list-get/plan.md -> "Constraints"; .specify/archive/004-routes-list-get/plan.md -> "Constraints"]

**Scale/Scope**: Six commands: `sonora outputs list` (`--include-disabled`, `--json`,
`--verbose`, `--hub-url`) and `sonora outputs get <output-id>` (`--json`, `--verbose`,
`--hub-url`; no `--include-disabled`), `sonora inputs list` (the same flags as `outputs list`) and `sonora inputs get <input-id>` (the same flags as `outputs get`), `sonora routes list` (`--status`, `--input-id`, `--target-id`, `--json`, `--verbose`, `--hub-url`) and `sonora routes get <route-id>` (`--json`, `--verbose`, `--hub-url`). [Source: .specify/archive/001-list-outputs/plan.md -> "Scale/Scope"; .specify/archive/002-outputs-get/plan.md -> "Scale/Scope"; .specify/archive/003-inputs-list-get/plan.md -> "Scale/Scope"; .specify/archive/004-routes-list-get/plan.md -> "Scale/Scope"]

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
│   ├── inputs.go                  # NEW (003): ListInputs(ctx, client, baseURL, includeDisabled) — GET /api/v2/inputs; GetInput(ctx, client, baseURL, inputID) — GET /api/v2/inputs/{inputId}
│   ├── routes.go                  # NEW (004): ListRoutes(ctx, client, baseURL, status, inputID, targetID) — GET /api/v2/routes; GetRoute(ctx, client, baseURL, routeID) — GET /api/v2/routes/{routeId}
│   └── errors.go                  # Error classification: usage / hub / network / not found (NotFoundError {Resource, ID}, ClassNotFound)
├── config/
│   └── config.go                  # Read-only load of ~/.config/sonora/config.json; hub URL precedence resolution
├── cli/
│   ├── outputs/
│   │   ├── list.go                # `sonora outputs list` (RunList): flag definitions, dispatch, exit code mapping
│   │   └── get.go                 # `sonora outputs get <output-id>` (RunGet): positional id, flag definitions, dispatch, exit code mapping
│   ├── inputs/                    # NEW (003): same shape as outputs/
│       ├── list.go                # `sonora inputs list` (RunList): flag definitions, dispatch, exit code mapping
│       └── get.go                 # `sonora inputs get <input-id>` (RunGet): positional id, flag definitions, dispatch, exit code mapping
│   └── routes/                    # NEW (004): same shape as outputs/ and inputs/
│       ├── list.go                # `sonora routes list` (RunList): filter flags, dispatch, exit code mapping
│       └── get.go                 # `sonora routes get <route-id>` (RunGet): positional id, flag definitions, dispatch, exit code mapping
└── render/
    ├── outputs.go                 # YAML (default, fixed-shape emitter) / JSON (--json) rendering; RenderOutputYAML/RenderOutputJSON for a single unwrapped record
    ├── inputs.go                  # NEW (003): RenderYAML/RenderJSON for input lists; RenderInputYAML/RenderInputJSON for a single record
    └── routes.go                  # NEW (004): RenderRoutesYAML/RenderRoutesJSON for list views (five fields); RenderRouteYAML/RenderRouteJSON for a single record (ten fields)

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

The `inputs` noun (003) mirrors `outputs` in its own packages: `internal/cli/inputs` (`RunList`,
`RunGet`), `internal/hub/inputs.go` (`ListInputs`, `GetInput`) and `internal/render/inputs.go`.
Its `NotFoundError` change is additive: `outputs get`'s message text and exit code are unchanged.
[Source: .specify/archive/003-inputs-list-get/research.md -> "6. CLI dispatch: new `internal/cli/inputs` package, same shape as `internal/cli/outputs`"; .specify/archive/003-inputs-list-get/research.md -> "4. Not-found error type: generalize rather than duplicate"]

The `routes` noun (004) follows the same layout: `internal/cli/routes` (`RunList`, `RunGet`),
`internal/hub/routes.go` (`ListRoutes`, `GetRoute`) and `internal/render/routes.go`. Its
`NotFoundError` use passes `Resource: "route"`, so no shared error type changed.
[Source: .specify/archive/004-routes-list-get/plan.md -> "Structure Decision"]

## Configuration

- **Hub base URL**, highest precedence first: `--hub-url` flag; `MULTIROOM_URL` environment
  variable; `hubUrl` field of `~/.config/sonora/config.json`; built-in default
  `http://localhost:8080` (the `servers[0].url` in `api/openapi.json`). The config file is
  read lazily, only when the URL must be resolved, after flag parsing. A missing file falls
  through silently; malformed JSON or a non-string `hubUrl` is a usage error (exit `2`)
  naming the file. Writing the file is out of scope. `outputs get`, `inputs list`, `inputs get`, `routes list` and `routes get` use the same precedence.
  [Source: .specify/archive/001-list-outputs/research.md -> "5. Hub location resolution"; .specify/archive/002-outputs-get/data-model.md -> "--hub-url"; .specify/archive/003-inputs-list-get/data-model.md -> "Config file"; .specify/archive/004-routes-list-get/data-model.md -> "Config file"]
- **Config file** `~/.config/sonora/config.json`: optional; field `hubUrl` (string).
  [Source: .specify/archive/001-list-outputs/data-model.md -> "Config file"]

## Routing & Navigation

- `sonora outputs list [--include-disabled] [--json] [--verbose] [--hub-url URL]` → `GET {hub-url}/api/v2/outputs?includeDisabled={true|false}`; single attempt, 5s timeout, no retries. No positional arguments; unknown flags or extra positionals are a usage error (exit `2`).
  [Source: .specify/archive/001-list-outputs/contracts/cli-outputs-list.md -> "Invocation"]
- `sonora outputs get <output-id> [--json] [--verbose] [--hub-url URL]` → `GET {hub-url}/api/v2/outputs/{output-id}` (path-escaped; operationId `getOutput`); single attempt, 5s timeout, no retries. Exactly one positional argument is required (none or more than one is a usage error, exit `2`); no `--include-disabled`, since the output is returned regardless of enabled state. Success output is a single unwrapped record (YAML by default, strict JSON with `--json`), every field always present.
  [Source: .specify/archive/002-outputs-get/contracts/cli-outputs-get.md -> "Invocation"]
- `sonora inputs list [--include-disabled] [--json] [--verbose] [--hub-url URL]` → `GET {hub-url}/api/v2/inputs?includeDisabled={true|false}` (operationId `listInputs`); single attempt, 5s timeout, no retries. Same flags, usage errors and exit codes as `outputs list`. Success output is a top-level `inputs:` list (YAML) or `{"inputs": [...]}` (`--json`); `createdAt` is `null` for static inputs.
  [Source: .specify/archive/003-inputs-list-get/contracts/cli-inputs-list.md -> "Invocation"]
- `sonora inputs get <input-id> [--json] [--verbose] [--hub-url URL]` → `GET {hub-url}/api/v2/inputs/{input-id}` (path-escaped; operationId `getInput`); single attempt, 5s timeout, no retries. Exactly one positional argument is required (exit `2` otherwise); no `--include-disabled`, since the input is returned regardless of its enabled state. Success output is one unwrapped record, as for `outputs get`; a 404 exits `5` with an "input not found" message.
  [Source: .specify/archive/003-inputs-list-get/contracts/cli-inputs-get.md -> "Invocation"]
- `sonora routes list [--status STATUS] [--input-id ID] [--target-id ID] [--json] [--verbose] [--hub-url URL]` → `GET {hub-url}/api/v2/routes`, sending `status`, `inputId` and `targetId` as query parameters only when set (operationId `listRoutes`); the hub applies them with AND logic and answers an invalid `status` with `400`, which exits `3`. Single attempt, 5s timeout, no retries. No positional arguments. Success output is a top-level `routes:` list (YAML) or `{"routes": [...]}` (`--json`) with five fields per route; zero routes exits `0`.
  [Source: .specify/archive/004-routes-list-get/data-model.md -> "CLI invocation shape"]
- `sonora routes get <route-id> [--json] [--verbose] [--hub-url URL]` → `GET {hub-url}/api/v2/routes/{route-id}` (path-escaped; operationId `getRoute`); single attempt, 5s timeout, no retries. Exactly one positional argument is required (exit `2` otherwise); a 404 exits `5` with a "route not found" message. Success output is one unwrapped record with all ten fields; `startedAt` is `null` until playback starts.
  [Source: .specify/archive/004-routes-list-get/data-model.md -> "CLI invocation shape"]
- **Output**: YAML by default (a top-level `outputs:` list; string scalars always double-quoted; `available` always printed; zero outputs prints `outputs: []` plus a plain-text note), strict JSON `{"outputs": [...]}` with `--json`. Success payload on stdout, failures on stderr; `--verbose` appends the raw underlying error.
  [Source: .specify/archive/001-list-outputs/contracts/cli-outputs-list.md -> "Success output"]
- **Exit codes**: `0` success (including zero outputs), `2` usage error, `3` hub-reported error (non-2xx other than a get's 404, or 2xx with a malformed body), `4` network/connectivity error (unreachable, DNS, timeout), `5` not found (hub answered `404` for the given output, input or route identifier; appended, nothing renumbered).
  [Source: .specify/archive/001-list-outputs/research.md -> "6. Exit code scheme"; .specify/archive/002-outputs-get/research.md -> "2. The 404 \"not found\" case needs its own exit-code class"; .specify/archive/004-routes-list-get/data-model.md -> "Exit code classes"]
- **Errors**: `internal/hub` classifies stdlib errors into the failure classes (usage, hub, network, and for `get` not-found) with short actionable messages; raw Go errors appear only with `--verbose`.
  [Source: .specify/archive/001-list-outputs/research.md -> "7. Error message translation"]

## Testing Strategy

- **Unit** (`tests/unit`): `internal/hub` request construction and decoding; `internal/render` YAML/JSON output including the zero-outputs and unavailable-output cases.
- **Contract** (`tests/contract`): an `httptest.Server` serving `GET /api/v2/outputs` shaped from `api/openapi.json` (including `includeDisabled` and the `OutputResponse` fields); asserts the request and rejection of a malformed body.
- **Integration** (`tests/integration`): full `sonora outputs list` invocation against the same mock server, asserting stdout format, exit codes and include-disabled behavior.
- **Get command**: a unit test for single-output YAML/JSON rendering (including the unavailable-output case); a contract test whose mock server serves `getOutput`'s `200 OutputResponse` and `404 ErrorResponse` shapes and asserts the request path and both decode paths; an integration test of `sonora outputs get <id>` covering found, not-found, missing-identifier usage error, `--json` and exit codes.
  [Source: .specify/archive/002-outputs-get/research.md -> "6. Testing strategy — same three layers, extended for get"]

- **Inputs** (003): unit `tests/unit/cli_inputs_test.go`, `cli_inputs_get_test.go` and `render_inputs_test.go` (list and single-record YAML/JSON, zero inputs, disabled included, null-vs-populated `createdAt`), with `hub_client_test.go` extended for `NotFoundError`'s `Resource`/`ID`; contract `tests/contract/inputs_list_test.go` and `inputs_get_test.go` (`httptest.Server`s shaped from `listInputs`/`getInput`); integration `tests/integration/inputs_list_test.go` and `inputs_get_test.go`.
  [Source: .specify/archive/003-inputs-list-get/research.md -> "9. Testing strategy — same three layers, extended for `inputs`"]

- **Routes** (004): unit `tests/unit/cli_routes_test.go`, `cli_routes_get_test.go` and `render_routes_test.go` (list and single-record YAML/JSON, zero routes, the nullable `startedAt`); contract `tests/contract/routes_list_test.go` and `routes_get_test.go` (`httptest.Server`s shaped from `listRoutes`/`getRoute`, including the filter query parameters and the `400` for an invalid status); integration `tests/integration/routes_list_test.go` and `routes_get_test.go`.
  [Source: .specify/archive/004-routes-list-get/research.md -> "7. Testing strategy — same three layers, extended for `routes`"]

All written before implementation (Principle VI). `gofmt` and `go vet` are the baseline checks.
[Source: .specify/archive/001-list-outputs/research.md -> "8. Testing strategy"]
