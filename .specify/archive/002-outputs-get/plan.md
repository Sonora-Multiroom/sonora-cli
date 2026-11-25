# Implementation Plan: Get Single Audio Output

**Branch**: `002-outputs-get` | **Date**: 2026-08-25 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `/specs/002-outputs-get/spec.md`

**Note**: This template is filled in by the `/speckit-plan` command; its definition describes the execution workflow.

## Summary

Add `sonora outputs get <output-id>`, a read-only `GET /api/v2/outputs/{outputId}` call
(operationId `getOutput`) rendered as a single YAML record by default or JSON with `--json`,
showing identifier, display name, volume, mute state, enabled state, and hardware
availability — regardless of enabled state, unlike `outputs list`'s default filtering. This
is the direct single-item counterpart to `outputs list` and deliberately validates the
reusable skeleton that feature established: it reuses the `Output` type, `hub.NewClient()`,
the config-driven hub-URL resolution, the `--json`/`--verbose`/`--hub-url` flag names, and
the YAML/JSON rendering approach unchanged. The only genuinely new pieces are: the
`GetOutput` hub-client call, a `NotFoundError`/`ClassNotFound` (exit `5`) addition to the
existing error-classification scheme to make "not found" distinguishable from a generic hub
error (FR-012), and single-record (non-list) rendering. See [research.md](research.md) for
the decisions specific to this feature.

## Technical Context

**Language/Version**: Go 1.27.0 — same module (`go.mod`) established by `001-list-outputs`;
no new module, no version change.

**Primary Dependencies**: Go standard library only — `net/http`, `encoding/json`, `flag`,
`net/url`, `net/http/httptest` (tests). No third-party module is introduced; this feature
adds functions to the packages `001-list-outputs` already created rather than new
dependencies. See [research.md §1](research.md#1-http-call-get-apiv2outputsoutputid-operationid-getoutput).

**Storage**: Same single local user config file, `~/.config/sonora/config.json`, read
(not written) via the existing `config.ResolveHubURL` — unchanged from `001-list-outputs`.

**Testing**: `go test` (stdlib `testing`) across the same three layers established by
`001-list-outputs` — unit (`tests/unit`), contract (`tests/contract`, against an
`httptest.Server` shaped from `getOutput` in `api/openapi.json`), and integration
(`tests/integration`, full CLI invocation). Written before implementation per Principle VI —
see [research.md §6](research.md#6-testing-strategy--same-three-layers-extended-for-get).

**Target Platform**: Cross-platform CLI binary (Linux/macOS/Windows), same single static Go
binary via `go build` — this feature adds a verb, not a new binary/target.

**Project Type**: Single project — CLI + internal packages (unchanged from `001-list-outputs`).

**Performance Goals**: Cold start to first HTTP request dispatched well under 50ms
(constitution Performance Standards); a single output's state rendered in under 1s under
normal network conditions (SC-001).

**Constraints**: 5s total request timeout, single attempt with no automatic retry on
network/timeout failure (FR-010), reusing `hub.NewClient()` unchanged; no unbounded waits
(Principle IV); exit codes MUST distinguish success / usage error / not-found / hub error /
network error (FR-012) — a fifth exit-code class beyond `001-list-outputs`'s four, see
[research.md §2](research.md#2-the-404-not-found-case-needs-its-own-exit-code-class-fr-012).

**Scale/Scope**: One new command (`sonora outputs get <output-id>`) with two
behavior-affecting flags (`--json`, `--verbose`) plus the shared `--hub-url` override — no
`--include-disabled` (FR-003 makes it moot for a single lookup). See
[data-model.md](data-model.md) and [contracts/cli-outputs-get.md](contracts/cli-outputs-get.md).

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Check | Status |
|---|---|---|
| I. Instant Startup & Low-Latency | `GetOutput`'s HTTP client is constructed inside the `outputs get` command handler, after `flag.Parse()` completes, identical in shape to `outputs list` — no disk/network work occurs before argument parsing. | PASS |
| II. API Contract Fidelity | `GetOutput` decodes into the existing `Output` struct, which already maps 1:1 to `#/components/schemas/OutputResponse`; the request path/operation (`GET /api/v2/outputs/{outputId}` → `getOutput`) matches `api/openapi.json` exactly, including its documented `404` response. No undocumented fields/endpoints are invented. | PASS |
| III. Minimal, Justified Dependencies | No new dependency added; this feature extends the existing stdlib-only `internal/hub` and `internal/render` packages. | PASS |
| IV. Resilient, Transparent HTTP Client | Reuses `hub.NewClient()` (5s timeout, single attempt, no retries) unchanged; `404`/non-2xx/network/decode failures are all classified and translated to plain messages by default, with `--verbose` exposing the raw error; no panics — all failure paths map to distinguishable non-zero exit codes. | PASS |
| V. CLI UX Consistency | `sonora outputs get <id>` follows `<noun> <verb> [args]`; default output is YAML, `--json` switches to JSON; exit codes distinguish success/usage/not-found/hub/network (research.md §2); `--json`/`--verbose`/`--hub-url` flag names are reused unchanged from `outputs list`, per Principle V's "no synonyms for the same concept" rule. | PASS |
| VI. Test-First Development | Unit, contract, and integration tests (research.md §6) are written and reviewed before any implementation code, per the task breakdown this plan feeds into (`/speckit-tasks`). | PASS (planned; enforced at task/implementation time) |

No violations requiring justification — the Complexity Tracking table below is empty.

## Project Structure

### Documentation (this feature)

```text
specs/002-outputs-get/
├── plan.md              # This file (/speckit-plan command output)
├── research.md          # Phase 0 output (/speckit-plan command)
├── data-model.md         # Phase 1 output (/speckit-plan command)
├── quickstart.md         # Phase 1 output (/speckit-plan command)
├── contracts/            # Phase 1 output (/speckit-plan command)
│   └── cli-outputs-get.md
└── tasks.md              # Phase 2 output (/speckit-tasks command - NOT created by /speckit-plan)
```

### Source Code (repository root)

```text
cmd/
└── sonora/
    └── main.go                        # MODIFIED: verb switch routes "list"→outputs.RunList, "get"→outputs.RunGet

internal/
├── hub/                               # Hub API client — validated against api/openapi.json (Principle II)
│   ├── client.go                      # UNCHANGED — hub.NewClient() reused as-is
│   ├── outputs.go                     # MODIFIED: add GetOutput(ctx, client, baseURL, outputID) (*Output, error)
│   └── errors.go                      # MODIFIED: add NotFoundError, ClassNotFound (exit 5), ClassifyError branch
├── config/
│   └── config.go                      # UNCHANGED — config.ResolveHubURL reused as-is
├── cli/
│   └── outputs/
│       ├── list.go                    # MODIFIED: exported Run renamed to RunList (no behavior change)
│       └── get.go                     # NEW: `sonora outputs get` — positional <output-id>, flag definitions, dispatch, exit code mapping
└── render/
    └── outputs.go                     # MODIFIED: add RenderOutputYAML(hub.Output), RenderOutputJSON(hub.Output) — single-record, not list-wrapped

tests/
├── contract/
│   ├── outputs_list_test.go           # UNCHANGED
│   └── outputs_get_test.go            # NEW: httptest.Server shaped from getOutput's 200/404 responses; request path + decode contract
├── integration/
│   ├── outputs_list_test.go           # MODIFIED: outputs.Run call sites → outputs.RunList (rename only)
│   └── outputs_get_test.go            # NEW: full `sonora outputs get` invocation — found, not-found, missing-id usage error, --json, exit codes
└── unit/
    ├── cli_outputs_test.go            # MODIFIED: outputs.Run call sites → outputs.RunList (rename only)
    ├── hub_client_test.go             # UNCHANGED
    ├── config_test.go                 # UNCHANGED
    └── render_output_get_test.go      # NEW: single-output YAML/JSON rendering, unavailable-output distinguishability
```

**Structure Decision**: No structural change from `001-list-outputs` — this feature adds a
second verb to the existing `internal/cli/outputs` package and extends the existing
`internal/hub`/`internal/render` packages rather than introducing new top-level directories.
The one rename (`outputs.Run` → `outputs.RunList`) is necessary because `Run` was only
unambiguous while `list` was the package's sole verb (research.md §3); it is a pure rename
with no behavior change, applied consistently at its two existing call sites plus `main.go`.

## Complexity Tracking

*No entries — Constitution Check reported no violations requiring justification.*
