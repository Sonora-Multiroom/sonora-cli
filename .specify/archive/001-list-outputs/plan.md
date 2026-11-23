# Implementation Plan: List Audio Outputs

**Branch**: `001-list-outputs` | **Date**: 2026-08-24 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `/specs/001-list-outputs/spec.md`

**Note**: This template is filled in by the `/speckit-plan` command; its definition describes the execution workflow.

## Summary

Add `sonora outputs list`, the CLI's first command: a read-only `GET /api/v2/outputs` call
(optionally with `?includeDisabled=true`) rendered as YAML by default or JSON with
`--json`, each output showing identifier, display name, volume, mute state, enabled state,
and hardware availability. Because this is the first command in the repo, it also
establishes the reusable skeleton future commands build on: a timeout-bound, single-attempt
HTTP client (Principle IV), openapi.json-validated request/response types (Principle II),
and dual human/`--json` rendering with class-distinguishing exit codes (Principle V) — all
built on the Go standard library only, per the dependency analysis in
[research.md](research.md).

## Technical Context

**Language/Version**: Go 1.27.0 (current stable release, module-based; `go.mod` created by
this feature — see [research.md §1](research.md))

**Primary Dependencies**: Go standard library only — `net/http`, `encoding/json`, `flag`,
`net/http/httptest` (tests). No third-party module is introduced; see
[research.md §2-3](research.md) for why a YAML dependency is not needed even though the
default output format is YAML.

**Storage**: A single local user config file, `~/.config/sonora/config.json` (JSON), read
(not written) by this feature as the third-priority source for the hub base URL — see
[research.md §5](research.md). No database, no other persistence.

**Testing**: `go test` (stdlib `testing` package) across three layers — unit
(`tests/unit`), contract (`tests/contract`, against an `httptest.Server` shaped from
`api/openapi.json`), and integration (`tests/integration`, full CLI invocation). Written
before implementation per Principle VI — see [research.md §8](research.md).

**Target Platform**: Cross-platform CLI binary (Linux/macOS/Windows), single static Go
binary via `go build`.

**Project Type**: Single project — CLI + internal packages (no frontend/backend split, no
mobile component).

**Performance Goals**: Cold start to first HTTP request dispatched well under 50ms
(constitution Performance Standards); full output list rendered in under 1s under normal
network conditions (SC-001).

**Constraints**: 5s total request timeout, single attempt with no automatic retry on
network/timeout failure (FR-009, clarified); no unbounded waits (Principle IV); exit codes
MUST distinguish success / usage error / hub error / network error (FR-011) — see
[research.md §6](research.md) for the concrete scheme.

**Scale/Scope**: One command (`sonora outputs list`) with three behavior-affecting flags
(`--include-disabled`, `--json`, `--verbose`) plus a `--hub-url` override — see
[data-model.md](data-model.md) and [contracts/cli-outputs-list.md](contracts/cli-outputs-list.md).

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Check | Status |
|---|---|---|
| I. Instant Startup & Low-Latency | HTTP client is constructed inside the `outputs list` command handler, after `flag.Parse()` completes — not at package `init()` or CLI dispatch time. The `~/.config/sonora/config.json` read is likewise deferred until hub-URL resolution actually runs (only when needed, and only after `--hub-url`/env are checked first) — no disk/network/plugin work occurs before argument parsing. | PASS |
| II. API Contract Fidelity | `Output` struct fields map 1:1 to `#/components/schemas/OutputResponse` in `api/openapi.json` ([data-model.md](data-model.md)); request path/query (`GET /api/v2/outputs?includeDisabled=`) matches the spec's `listOutputs` operation exactly; no undocumented fields/endpoints are invented. | PASS |
| III. Minimal, Justified Dependencies | No new dependency added. The default-YAML requirement (Principle V) is met with a small fixed-shape emitter instead of a YAML library, since the data shape is flat and fully known — see [research.md §3](research.md) for the alternatives considered. | PASS |
| IV. Resilient, Transparent HTTP Client | `http.Client{Timeout: 5 * time.Second}`, single attempt, no retries; errors classified and translated to plain messages by default with `--verbose` exposing the raw error; no panics — network/decode failures map to distinguishable non-zero exit codes. | PASS |
| V. CLI UX Consistency | `sonora outputs list` follows `<noun> <verb>`; default output is YAML, `--json` switches to JSON; exit codes distinguish success/usage/API/network (see [research.md §6](research.md)); flag names (`--json`, `--verbose`) are chosen to be reused unchanged by future commands. | PASS |
| VI. Test-First Development | Unit, contract, and integration tests ([research.md §8](research.md)) are written and reviewed before any implementation code, per the task breakdown this plan feeds into (`/speckit-tasks`). | PASS (planned; enforced at task/implementation time) |

No violations requiring justification — the Complexity Tracking table below is empty.

## Project Structure

### Documentation (this feature)

```text
specs/001-list-outputs/
├── plan.md              # This file (/speckit-plan command output)
├── research.md          # Phase 0 output (/speckit-plan command)
├── data-model.md         # Phase 1 output (/speckit-plan command)
├── quickstart.md         # Phase 1 output (/speckit-plan command)
├── contracts/            # Phase 1 output (/speckit-plan command)
│   └── cli-outputs-list.md
└── tasks.md              # Phase 2 output (/speckit-tasks command - NOT created by /speckit-plan)
```

### Source Code (repository root)

```text
cmd/
└── sonora/
    └── main.go                    # Entrypoint: top-level dispatch (`outputs` noun → `list` verb), exit code translation

internal/
├── hub/                           # Hub API client — validated against api/openapi.json (Principle II)
│   ├── client.go                  # HTTP client construction (timeout, single attempt) — Principle IV
│   ├── outputs.go                 # ListOutputs(ctx, includeDisabled) — GET /api/v2/outputs
│   └── errors.go                  # Error classification: usage / hub / network (research.md §6-7)
├── config/
│   └── config.go                  # Read-only load of ~/.config/sonora/config.json; hub URL precedence resolution (research.md §5)
├── cli/
│   └── outputs/
│       └── list.go                # `sonora outputs list`: flag definitions, dispatch, exit code mapping
└── render/
    └── outputs.go                 # YAML (default, fixed-shape emitter) / JSON (--json) rendering

tests/
├── contract/
│   └── outputs_list_test.go       # httptest.Server shaped from openapi.json OutputResponse; request+decode contract
├── integration/
│   └── outputs_list_test.go       # Full `sonora outputs list` invocation against the mock server
└── unit/
    ├── hub_client_test.go         # Timeout, single-attempt, error classification
    ├── config_test.go             # Hub URL precedence (flag > env > config file > default), malformed config file handling
    └── render_outputs_test.go     # YAML/JSON rendering, zero-outputs, unavailable-output distinguishability
```

**Structure Decision**: Single Go project (no frontend/backend or mobile split — this is a
CLI talking to one HTTP API). Layout follows Go convention (`cmd/` for the binary
entrypoint, `internal/` for non-exported packages) rather than the generic
`src/models|services|cli|lib` template, since Go's own package conventions already provide
that separation: `internal/hub` is the "service" layer (API client validated against the
spec), `internal/render` is the "lib" layer (output formatting reusable by future
commands), `internal/config` is a small "model+service" pair for the persisted config file
(read-only in this feature), and `internal/cli/outputs` is the "cli" layer (flag parsing
and dispatch for this noun). `tests/{contract,integration,unit}` mirrors the template's
test taxonomy directly.

## Complexity Tracking

*No entries — Constitution Check reported no violations requiring justification.*
