# Implementation Plan: List and Get Audio Routes

**Branch**: `004-routes-list-get` | **Date**: 2026-08-25 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `/specs/004-routes-list-get/spec.md`

**Note**: This template is filled in by the `/speckit-plan` command; its definition describes the execution workflow.

## Summary

Add `sonora routes list` and `sonora routes get <route-id>` — read-only `GET /api/v2/routes`
(operationId `listRoutes`) and `GET /api/v2/routes/{routeId}` (operationId `getRoute`) calls,
rendered as YAML by default or JSON with `--json`. Both commands are structural mirrors of
`outputs list`/`outputs get` and `inputs list`/`inputs get`: same HTTP client construction,
same hub-URL resolution, same exit-code scheme, same stdout/stderr separation. Two things are
genuinely new: (1) `routes list` has no enabled/disabled concept to filter on — instead it
takes three independent, AND-combined optional filters (`--status`, `--input-id`,
`--target-id`) that the hub itself validates and applies; and (2) `routes get` returns five
fields (`createdAt`, `startedAt`, `transferable`, `pauseable`, `paused`) that `routes list`
does not show — the first read command in this codebase where the list and get views
genuinely differ in shape rather than being the same record set. See
[research.md](research.md) for the decisions specific to this feature.

## Technical Context

**Language/Version**: Go 1.27.0 — same module (`go.mod`) established by `001-list-outputs`;
no new module, no version change.

**Primary Dependencies**: Go standard library only — `net/http`, `encoding/json`, `flag`,
`net/url` (tests: `net/http/httptest`). No third-party module is introduced; this feature adds
a new `internal/cli/routes` package and new files in the existing `internal/hub` and
`internal/render` packages. See [research.md §1](research.md#1-the-route-entity-and-its-splitlistget-field-set).

**Storage**: Same single local user config file, `~/.config/sonora/config.json`, read
(not written) via the existing `config.ResolveHubURL` — unchanged from
`001-list-outputs`/`002-outputs-get`/`003-inputs-list-get`.

**Testing**: `go test` (stdlib `testing`) across the same three layers established by
`001-list-outputs` — unit (`tests/unit`), contract (`tests/contract`, against
`httptest.Server`s shaped from `listRoutes`/`getRoute` in `api/openapi.json`), and integration
(`tests/integration`, full CLI invocation). Written before implementation per Principle VI —
see [research.md §7](research.md#7-testing-strategy--same-three-layers-extended-for-routes).

**Target Platform**: Cross-platform CLI binary (Linux/macOS/Windows), same single static Go
binary via `go build` — this feature adds a noun/verb pair, not a new binary/target.

**Project Type**: Single project — CLI + internal packages (unchanged from prior features).

**Performance Goals**: Cold start to first HTTP request dispatched well under 50ms
(constitution Performance Standards); a list or single route's state rendered in under 1s
under normal network conditions (SC-001, SC-002).

**Constraints**: 5s total request timeout, single attempt with no automatic retry on
network/timeout failure (FR-012), reusing `hub.NewClient()` unchanged; no unbounded waits
(Principle IV); exit codes MUST distinguish success / usage error / not-found (get only) /
hub error / network error (FR-014) — reusing the existing five-class scheme `002-outputs-get`
established, with no sixth class needed, see
[research.md §5](research.md#5-exit-code-scheme-unchanged-no-new-class-needed).

**Scale/Scope**: Two new commands (`sonora routes list`, `sonora routes get <route-id>`)
under a new `routes` noun. `list` gains three new optional filter flags not present on any
existing command (`--status`, `--input-id`, `--target-id`); both commands otherwise reuse the
existing `--json`/`--verbose`/`--hub-url` flag surface. See [data-model.md](data-model.md),
[contracts/cli-routes-list.md](contracts/cli-routes-list.md), and
[contracts/cli-routes-get.md](contracts/cli-routes-get.md).

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Check | Status |
|---|---|---|
| I. Instant Startup & Low-Latency | `hub.NewClient()` is constructed inside each command handler, after `flag.Parse()` completes, identical in shape to `outputs`/`inputs` — no disk/network work occurs before argument parsing. | PASS |
| II. API Contract Fidelity | `Route` decodes into a new struct mapping 1:1 to `#/components/schemas/RouteResponse`; the request paths/operations (`GET /api/v2/routes` → `listRoutes`, `GET /api/v2/routes/{routeId}` → `getRoute`) match `api/openapi.json` exactly, including `listRoutes`'s documented `status`/`inputId`/`targetId` query parameters and `getRoute`'s documented `404` response. No undocumented fields/endpoints are invented. | PASS |
| III. Minimal, Justified Dependencies | No new dependency added; this feature adds a new package (`internal/cli/routes`) and new files to the existing stdlib-only `internal/hub`/`internal/render` packages. | PASS |
| IV. Resilient, Transparent HTTP Client | Reuses `hub.NewClient()` (5s timeout, single attempt, no retries) unchanged; 404/non-2xx (including the hub's own 400 for an invalid `status` filter)/network/decode failures are all classified and translated to plain messages by default, with `--verbose` exposing the raw error; no panics — all failure paths map to distinguishable non-zero exit codes. | PASS |
| V. CLI UX Consistency | `sonora routes list`/`sonora routes get <id>` follow `<noun> <verb> [args]`; default output is YAML, `--json` switches to JSON; exit codes distinguish success/usage/not-found/hub/network (unchanged scheme); `--json`/`--verbose`/`--hub-url` flag names are reused unchanged, and the three new filter flags (`--status`, `--input-id`, `--target-id`) follow the same hyphenated-lowercase naming convention as `--include-disabled`/`--hub-url`, per Principle V's "no synonyms for the same concept" rule. | PASS |
| VI. Test-First Development | Unit, contract, and integration tests (research.md §7) are written and reviewed before any implementation code, per the task breakdown this plan feeds into (`/speckit-tasks`). | PASS (planned; enforced at task/implementation time) |

No violations requiring justification — the Complexity Tracking table below is empty.

## Project Structure

### Documentation (this feature)

```text
specs/004-routes-list-get/
├── plan.md               # This file (/speckit-plan command output)
├── research.md           # Phase 0 output (/speckit-plan command)
├── data-model.md         # Phase 1 output (/speckit-plan command)
├── quickstart.md         # Phase 1 output (/speckit-plan command)
├── contracts/            # Phase 1 output (/speckit-plan command)
│   ├── cli-routes-list.md
│   └── cli-routes-get.md
└── tasks.md              # Phase 2 output (/speckit-tasks command - NOT created by /speckit-plan)
```

### Source Code (repository root)

```text
cmd/
└── sonora/
    └── main.go                        # MODIFIED: noun switch gains a "routes" case routing "list"→routes.RunList, "get"→routes.RunGet

internal/
├── hub/                                # Hub API client — validated against api/openapi.json (Principle II)
│   ├── client.go                       # UNCHANGED — hub.NewClient() reused as-is
│   ├── outputs.go                      # UNCHANGED
│   ├── inputs.go                       # UNCHANGED
│   ├── routes.go                       # NEW: Route struct, ListRoutes(ctx, client, baseURL, status, inputID, targetID), GetRoute(ctx, client, baseURL, routeID)
│   └── errors.go                       # UNCHANGED — NotFoundError{Resource, ID} already generalized by 003-inputs-list-get; GetRoute passes Resource: "route"
├── config/
│   └── config.go                       # UNCHANGED — config.ResolveHubURL reused as-is
├── cli/
│   ├── outputs/                        # UNCHANGED
│   ├── inputs/                         # UNCHANGED
│   └── routes/
│       ├── list.go                     # NEW: `sonora routes list` — flag definitions (--status/--input-id/--target-id/--json/--verbose/--hub-url), dispatch, exit code mapping
│       └── get.go                      # NEW: `sonora routes get` — positional <route-id>, flag definitions, dispatch, exit code mapping
└── render/
    ├── outputs.go                      # UNCHANGED
    ├── inputs.go                       # UNCHANGED
    └── routes.go                       # NEW: RenderRoutesYAML([]hub.Route), RenderRoutesJSON([]hub.Route) (list-view fields only), RenderRouteYAML(hub.Route), RenderRouteJSON(hub.Route) (full get-view fields)

tests/
├── contract/
│   ├── routes_list_test.go             # NEW: httptest.Server shaped from listRoutes's 200 response; request path/query (status/inputId/targetId) + decode contract
│   └── routes_get_test.go              # NEW: httptest.Server shaped from getRoute's 200/404 responses; request path + decode contract
├── integration/
│   ├── routes_list_test.go             # NEW: full `sonora routes list` invocation — no-filter default, each filter individually, combined filters, --json, exit codes
│   └── routes_get_test.go              # NEW: full `sonora routes get` invocation — found, not-found, missing-id usage error, --json, exit codes
└── unit/
    ├── cli_routes_test.go              # NEW: routes.RunList flag parsing/dispatch
    ├── cli_routes_get_test.go          # NEW: routes.RunGet flag/positional parsing/dispatch
    └── render_routes_test.go           # NEW: list YAML/JSON (5-field view), single-record YAML/JSON (10-field view), zero-routes, null-vs-populated startedAt
```

**Structure Decision**: New `internal/cli/routes` package (mirroring the existing
`internal/cli/outputs`/`internal/cli/inputs` package structure) plus new files in the existing
`internal/hub` and `internal/render` packages — no new top-level directories, no changes to
existing `outputs`/`inputs` code. `errors.go`'s `NotFoundError{Resource, ID}` (generalized by
`003-inputs-list-get`) is reused unchanged — `GetRoute` simply passes `Resource: "route"`.

## Complexity Tracking

*No entries — Constitution Check reported no violations requiring justification.*
