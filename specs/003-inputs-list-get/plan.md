# Implementation Plan: List and Get Audio Inputs

**Branch**: `003-inputs-list-get` | **Date**: 2026-08-25 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `/specs/003-inputs-list-get/spec.md`

**Note**: This template is filled in by the `/speckit-plan` command; its definition describes the execution workflow.

## Summary

Add `sonora inputs list` and `sonora inputs get <input-id>` — read-only `GET /api/v2/inputs`
(operationId `listInputs`) and `GET /api/v2/inputs/{inputId}` (operationId `getInput`) calls,
rendered as YAML by default or JSON with `--json`. Both commands are structural mirrors of
`outputs list`/`outputs get`: same flag names, same HTTP client construction, same hub-URL
resolution, same exit-code scheme, same stdout/stderr separation. The only genuinely new
work is driven by the `Input` entity having a different shape than `Output` — no volume,
mute, or hardware-availability concept; instead `uri`, `source` (STATIC/EPHEMERAL),
`autoRemove`, `pauseable`, and a nullable `createdAt` (the codebase's first nullable field).
This also generalizes the existing `NotFoundError` (previously output-specific) to carry a
`Resource` name, so `inputs get`'s 404 case reuses the same type and exit-code class
(`ClassNotFound`, exit `5`) `outputs get` already established, without duplicating it. See
[research.md](research.md) for the decisions specific to this feature.

## Technical Context

**Language/Version**: Go 1.27.0 — same module (`go.mod`) established by `001-list-outputs`;
no new module, no version change.

**Primary Dependencies**: Go standard library only — `net/http`, `encoding/json`, `flag`,
`net/url`, `net/http/httptest` (tests). No third-party module is introduced; this feature
adds a new `internal/cli/inputs` package and new files in the existing `internal/hub` and
`internal/render` packages. See
[research.md §1, §3](research.md#1-the-input-entity-has-no-overlap-with-outputs-fields).

**Storage**: Same single local user config file, `~/.config/sonora/config.json`, read
(not written) via the existing `config.ResolveHubURL` — unchanged from
`001-list-outputs`/`002-outputs-get`.

**Testing**: `go test` (stdlib `testing`) across the same three layers established by
`001-list-outputs` — unit (`tests/unit`), contract (`tests/contract`, against
`httptest.Server`s shaped from `listInputs`/`getInput` in `api/openapi.json`), and
integration (`tests/integration`, full CLI invocation). Written before implementation per
Principle VI — see [research.md §9](research.md#9-testing-strategy--same-three-layers-extended-for-inputs).

**Target Platform**: Cross-platform CLI binary (Linux/macOS/Windows), same single static Go
binary via `go build` — this feature adds a noun/verb pair, not a new binary/target.

**Project Type**: Single project — CLI + internal packages (unchanged from
`001-list-outputs`/`002-outputs-get`).

**Performance Goals**: Cold start to first HTTP request dispatched well under 50ms
(constitution Performance Standards); a list or single input's state rendered in under 1s
under normal network conditions (SC-001, SC-002).

**Constraints**: 5s total request timeout, single attempt with no automatic retry on
network/timeout failure (FR-013), reusing `hub.NewClient()` unchanged; no unbounded waits
(Principle IV); exit codes MUST distinguish success / usage error / not-found (get only) /
hub error / network error (FR-015) — reusing the existing five-class scheme
`002-outputs-get` established, with no sixth class needed, see
[research.md §5](research.md#5-exit-code-scheme-unchanged-no-new-class-needed).

**Scale/Scope**: Two new commands (`sonora inputs list`, `sonora inputs get <input-id>`)
under a new `inputs` noun, with the same flag surface as their `outputs` counterparts
(`--include-disabled` on `list` only, `--json`, `--verbose`, `--hub-url`). See
[data-model.md](data-model.md), [contracts/cli-inputs-list.md](contracts/cli-inputs-list.md),
and [contracts/cli-inputs-get.md](contracts/cli-inputs-get.md).

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Check | Status |
|---|---|---|
| I. Instant Startup & Low-Latency | `hub.NewClient()` is constructed inside each command handler, after `flag.Parse()` completes, identical in shape to `outputs list`/`outputs get` — no disk/network work occurs before argument parsing. | PASS |
| II. API Contract Fidelity | `Input` decodes into a new struct mapping 1:1 to `#/components/schemas/InputResponse`; the request paths/operations (`GET /api/v2/inputs` → `listInputs`, `GET /api/v2/inputs/{inputId}` → `getInput`) match `api/openapi.json` exactly, including `getInput`'s documented `404` response. No undocumented fields/endpoints are invented. | PASS |
| III. Minimal, Justified Dependencies | No new dependency added; this feature adds a new package (`internal/cli/inputs`) and new files to the existing stdlib-only `internal/hub`/`internal/render` packages. | PASS |
| IV. Resilient, Transparent HTTP Client | Reuses `hub.NewClient()` (5s timeout, single attempt, no retries) unchanged; 404/non-2xx/network/decode failures are all classified and translated to plain messages by default, with `--verbose` exposing the raw error; no panics — all failure paths map to distinguishable non-zero exit codes. | PASS |
| V. CLI UX Consistency | `sonora inputs list`/`sonora inputs get <id>` follow `<noun> <verb> [args]`; default output is YAML, `--json` switches to JSON; exit codes distinguish success/usage/not-found/hub/network (unchanged scheme); `--json`/`--verbose`/`--hub-url`/`--include-disabled` flag names are reused unchanged from the `outputs` commands, per Principle V's "no synonyms for the same concept" rule. | PASS |
| VI. Test-First Development | Unit, contract, and integration tests (research.md §9) are written and reviewed before any implementation code, per the task breakdown this plan feeds into (`/speckit-tasks`). | PASS (planned; enforced at task/implementation time) |

No violations requiring justification — the Complexity Tracking table below is empty.

## Project Structure

### Documentation (this feature)

```text
specs/003-inputs-list-get/
├── plan.md               # This file (/speckit-plan command output)
├── research.md           # Phase 0 output (/speckit-plan command)
├── data-model.md         # Phase 1 output (/speckit-plan command)
├── quickstart.md         # Phase 1 output (/speckit-plan command)
├── contracts/            # Phase 1 output (/speckit-plan command)
│   ├── cli-inputs-list.md
│   └── cli-inputs-get.md
└── tasks.md              # Phase 2 output (/speckit-tasks command - NOT created by /speckit-plan)
```

### Source Code (repository root)

```text
cmd/
└── sonora/
    └── main.go                        # MODIFIED: noun switch gains a "inputs" case routing "list"→inputs.RunList, "get"→inputs.RunGet

internal/
├── hub/                                # Hub API client — validated against api/openapi.json (Principle II)
│   ├── client.go                       # UNCHANGED — hub.NewClient() reused as-is
│   ├── outputs.go                      # UNCHANGED
│   ├── inputs.go                       # NEW: Input struct, ListInputs(ctx, client, baseURL, includeDisabled), GetInput(ctx, client, baseURL, inputID)
│   └── errors.go                       # MODIFIED: generalize NotFoundError{OutputID string} → NotFoundError{Resource, ID string} (message text unchanged for outputs); GetOutput updated to pass Resource: "output"
├── config/
│   └── config.go                       # UNCHANGED — config.ResolveHubURL reused as-is
├── cli/
│   ├── outputs/                        # UNCHANGED
│   └── inputs/
│       ├── list.go                     # NEW: `sonora inputs list` — flag definitions, dispatch, exit code mapping
│       └── get.go                      # NEW: `sonora inputs get` — positional <input-id>, flag definitions, dispatch, exit code mapping
└── render/
    ├── outputs.go                      # UNCHANGED
    └── inputs.go                       # NEW: RenderYAML([]hub.Input), RenderJSON([]hub.Input), RenderInputYAML(hub.Input), RenderInputJSON(hub.Input)

tests/
├── contract/
│   ├── outputs_list_test.go            # UNCHANGED
│   ├── outputs_get_test.go             # UNCHANGED
│   ├── inputs_list_test.go             # NEW: httptest.Server shaped from listInputs's 200 response; request path/query + decode contract
│   └── inputs_get_test.go              # NEW: httptest.Server shaped from getInput's 200/404 responses; request path + decode contract
├── integration/
│   ├── outputs_list_test.go            # UNCHANGED
│   ├── outputs_get_test.go             # UNCHANGED
│   ├── inputs_list_test.go             # NEW: full `sonora inputs list` invocation — enabled-only default, --include-disabled, --json, exit codes
│   └── inputs_get_test.go              # NEW: full `sonora inputs get` invocation — found, not-found, missing-id usage error, --json, exit codes
└── unit/
    ├── cli_outputs_test.go             # UNCHANGED
    ├── cli_outputs_get_test.go         # UNCHANGED
    ├── hub_client_test.go              # MODIFIED: extend NotFoundError coverage for the generalized Resource/ID fields
    ├── config_test.go                  # UNCHANGED
    ├── render_outputs_test.go          # UNCHANGED
    ├── render_output_get_test.go       # UNCHANGED
    ├── cli_inputs_test.go              # NEW: inputs.RunList flag parsing/dispatch
    ├── cli_inputs_get_test.go          # NEW: inputs.RunGet flag/positional parsing/dispatch
    └── render_inputs_test.go           # NEW: list YAML/JSON, single-record YAML/JSON, zero-inputs, disabled-included, null-vs-populated createdAt
```

**Structure Decision**: New `internal/cli/inputs` package (mirroring the existing
`internal/cli/outputs` package structure) plus new files in the existing `internal/hub` and
`internal/render` packages — no new top-level directories. The one shared-code change is
generalizing `NotFoundError` in `internal/hub/errors.go` from an output-specific
`{OutputID string}` to a resource-generic `{Resource, ID string}` (research.md §4); this is
additive to its behavior — `outputs get`'s existing "output not found: `<id>`" message text
and exit code are unchanged — so no existing test's expected output changes.

## Complexity Tracking

*No entries — Constitution Check reported no violations requiring justification.*
