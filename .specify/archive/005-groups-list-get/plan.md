# Implementation Plan: List and Get Output Groups

**Branch**: `005-groups-list-get` | **Date**: 2026-08-26 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `/specs/005-groups-list-get/spec.md`

**Note**: This template is filled in by the `/speckit-plan` command; its definition describes the execution workflow.

## Summary

Add `sonora groups list` and `sonora groups get <group-id>` — read-only `GET /api/v2/groups`
(operationId `listGroups`) and `GET /api/v2/groups/{groupId}` (operationId `getGroup`) calls,
rendered as YAML by default or JSON with `--json`. Both commands are structural mirrors of
`outputs list`/`outputs get`: same boolean `--include-disabled` filter shape, same HTTP client
construction, same hub-URL resolution, same exit-code scheme, same stdout/stderr separation,
and — unlike `004-routes-list-get`'s `Route` — the same five-field set shown identically by
both `list` and `get` (no split view). The one genuinely new piece is the `Group` entity's
`outputIds` member list (`[]string`), which must render explicitly as `[]` when a group has no
member outputs rather than being omitted or rendered as `null`. See [research.md](research.md)
for the decisions specific to this feature.

## Technical Context

**Language/Version**: Go 1.27.0 — same module (`go.mod`) established by `001-list-outputs`;
no new module, no version change.

**Primary Dependencies**: Go standard library only — `net/http`, `encoding/json`, `flag`,
`net/url` (tests: `net/http/httptest`). No third-party module is introduced; this feature adds
a new `internal/cli/groups` package and new files in the existing `internal/hub` and
`internal/render` packages. See [research.md §1](research.md#1-the-group-entity-same-listget-field-set-structurally-closest-to-output).

**Storage**: Same single local user config file, `~/.config/sonora/config.json`, read
(not written) via the existing `config.ResolveHubURL` — unchanged from prior features.

**Testing**: `go test` (stdlib `testing`) across the same three layers established by
`001-list-outputs` — unit (`tests/unit`), contract (`tests/contract`, against
`httptest.Server`s shaped from `listGroups`/`getGroup` in `api/openapi.json`), and integration
(`tests/integration`, full CLI invocation). Written before implementation per Principle VI —
see [research.md §6](research.md#6-testing-strategy--same-three-layers-extended-for-groups).

**Target Platform**: Cross-platform CLI binary (Linux/macOS/Windows), same single static Go
binary via `go build` — this feature adds a noun/verb pair, not a new binary/target.

**Project Type**: Single project — CLI + internal packages (unchanged from prior features).

**Performance Goals**: Cold start to first HTTP request dispatched well under 50ms
(constitution Performance Standards); a list or single group's state rendered in under 1s
under normal network conditions (SC-001, SC-002).

**Constraints**: 5s total request timeout, single attempt with no automatic retry on
network/timeout failure (FR-013), reusing `hub.NewClient()` unchanged; no unbounded waits
(Principle IV); exit codes MUST distinguish success / usage error / not-found (get only) /
hub error / network error (FR-015) — reusing the existing five-class scheme prior features
established, with no sixth class needed, see
[research.md §4](research.md#4-exit-code-scheme-unchanged-no-new-class-needed).

**Scale/Scope**: Two new commands (`sonora groups list`, `sonora groups get <group-id>`)
under a new `groups` noun. Both commands reuse the existing `--json`/`--verbose`/`--hub-url`
flag surface; `list` additionally reuses the existing `--include-disabled` boolean flag
already established by `outputs list`/`inputs list`. See [data-model.md](data-model.md),
[contracts/cli-groups-list.md](contracts/cli-groups-list.md), and
[contracts/cli-groups-get.md](contracts/cli-groups-get.md).

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Check | Status |
|---|---|---|
| I. Instant Startup & Low-Latency | `hub.NewClient()` is constructed inside each command handler, after `flag.Parse()` completes, identical in shape to `outputs`/`inputs`/`routes` — no disk/network work occurs before argument parsing. | PASS |
| II. API Contract Fidelity | `Group` decodes into a new struct mapping 1:1 to `#/components/schemas/GroupResponse`; the request paths/operations (`GET /api/v2/groups` → `listGroups`, `GET /api/v2/groups/{groupId}` → `getGroup`) match `api/openapi.json` exactly, including `listGroups`'s documented `includeDisabled` query parameter and `getGroup`'s documented `404` response. No undocumented fields/endpoints are invented. | PASS |
| III. Minimal, Justified Dependencies | No new dependency added; this feature adds a new package (`internal/cli/groups`) and new files to the existing stdlib-only `internal/hub`/`internal/render` packages. | PASS |
| IV. Resilient, Transparent HTTP Client | Reuses `hub.NewClient()` (5s timeout, single attempt, no retries) unchanged; 404/non-2xx/network/decode failures are all classified and translated to plain messages by default, with `--verbose` exposing the raw error; no panics — all failure paths map to distinguishable non-zero exit codes. | PASS |
| V. CLI UX Consistency | `sonora groups list`/`sonora groups get <id>` follow `<noun> <verb> [args]`; default output is YAML, `--json` switches to JSON; exit codes distinguish success/usage/not-found/hub/network (unchanged scheme); `--json`/`--verbose`/`--hub-url`/`--include-disabled` flag names are reused unchanged from `outputs`/`inputs`, per Principle V's "no synonyms for the same concept" rule. | PASS |
| VI. Test-First Development | Unit, contract, and integration tests (research.md §6) are written and reviewed before any implementation code, per the task breakdown this plan feeds into (`/speckit-tasks`). | PASS (planned; enforced at task/implementation time) |

No violations requiring justification — the Complexity Tracking table below is empty.

## Project Structure

### Documentation (this feature)

```text
specs/005-groups-list-get/
├── plan.md               # This file (/speckit-plan command output)
├── research.md           # Phase 0 output (/speckit-plan command)
├── data-model.md         # Phase 1 output (/speckit-plan command)
├── quickstart.md         # Phase 1 output (/speckit-plan command)
├── contracts/            # Phase 1 output (/speckit-plan command)
│   ├── cli-groups-list.md
│   └── cli-groups-get.md
├── checklists/
│   └── requirements.md   # Pre-existing (speckit-checklist output)
└── tasks.md              # Phase 2 output (/speckit-tasks command - NOT created by /speckit-plan)
```

### Source Code (repository root)

```text
cmd/
└── sonora/
    └── main.go                        # MODIFIED: noun switch gains a "groups" case routing "list"→groups.RunList, "get"→groups.RunGet

internal/
├── hub/                                # Hub API client — validated against api/openapi.json (Principle II)
│   ├── client.go                       # UNCHANGED — hub.NewClient() reused as-is
│   ├── outputs.go                      # UNCHANGED
│   ├── inputs.go                       # UNCHANGED
│   ├── routes.go                       # UNCHANGED
│   ├── groups.go                       # NEW: Group struct, ListGroups(ctx, client, baseURL, includeDisabled), GetGroup(ctx, client, baseURL, groupID)
│   └── errors.go                       # UNCHANGED — NotFoundError{Resource, ID} already generalized; GetGroup passes Resource: "group"
├── config/
│   └── config.go                       # UNCHANGED — config.ResolveHubURL reused as-is
├── cli/
│   ├── outputs/                        # UNCHANGED
│   ├── inputs/                         # UNCHANGED
│   ├── routes/                         # UNCHANGED
│   └── groups/
│       ├── list.go                     # NEW: `sonora groups list` — flag definitions (--include-disabled/--json/--verbose/--hub-url), dispatch, exit code mapping
│       └── get.go                      # NEW: `sonora groups get` — positional <group-id>, flag definitions, dispatch, exit code mapping
└── render/
    ├── outputs.go                      # UNCHANGED
    ├── inputs.go                       # UNCHANGED
    ├── routes.go                       # UNCHANGED
    └── groups.go                       # NEW: RenderGroupsYAML([]hub.Group), RenderGroupsJSON([]hub.Group), RenderGroupYAML(hub.Group), RenderGroupJSON(hub.Group) — identical five-field set for both views

tests/
├── contract/
│   ├── groups_list_test.go             # NEW: httptest.Server shaped from listGroups's 200 response; request query (includeDisabled) + decode contract
│   └── groups_get_test.go              # NEW: httptest.Server shaped from getGroup's 200/404 responses; request path + decode contract
├── integration/
│   ├── groups_list_test.go             # NEW: full `sonora groups list` invocation — default (enabled-only), --include-disabled, --json, exit codes
│   └── groups_get_test.go              # NEW: full `sonora groups get` invocation — found, not-found, missing-id usage error, --json, exit codes
└── unit/
    ├── cli_groups_test.go              # NEW: groups.RunList flag parsing/dispatch
    ├── cli_groups_get_test.go          # NEW: groups.RunGet flag/positional parsing/dispatch
    └── render_groups_test.go           # NEW: list/single-record YAML/JSON (five-field view), zero-groups, zero-member-outputs (outputIds: [])
```

**Structure Decision**: New `internal/cli/groups` package (mirroring the existing
`internal/cli/outputs`/`internal/cli/inputs`/`internal/cli/routes` package structure) plus new
files in the existing `internal/hub` and `internal/render` packages — no new top-level
directories, no changes to existing `outputs`/`inputs`/`routes` code. `errors.go`'s
`NotFoundError{Resource, ID}` is reused unchanged — `GetGroup` simply passes `Resource:
"group"`.

## Complexity Tracking

*No entries — Constitution Check reported no violations requiring justification.*
