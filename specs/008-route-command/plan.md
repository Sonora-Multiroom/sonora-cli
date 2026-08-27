# Implementation Plan: Route an Existing Input (`route`)

**Branch**: `008-route-command` | **Date**: 2026-08-28 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `/specs/008-route-command/spec.md`

**Note**: This template is filled in by the `/speckit-plan` command; its definition describes the execution workflow.

## Summary

Add `sonora route inputs/<input-id> <outputs|groups>/<target-id>` — a single
`POST /api/v2/routes` call (operationId `createRoute`) that connects an existing, already-defined
input to an existing output or output group, without creating any new input. Unlike `play`
(which resolves a possibly-ambiguous bare target identifier against two endpoints and always
creates a new input), `route` never guesses: both arguments are path-style (`inputs/<id>`,
`outputs/<id>`/`groups/<id>`), so the resource type is read directly from the prefix and no
disambiguation logic is needed. The command adds one new pre-check `play` doesn't have — the
input must already exist (`GetInput`) — reuses `hub.ResolveTarget` for the target exactly as-is,
and introduces two new exit-code classes (input-not-found, target-not-found) so the two failure
modes FR-010 requires to be distinguishable don't collide on the existing generic
`ClassNotFound`. See [research.md](research.md) for the decisions specific to this feature.

## Technical Context

**Language/Version**: Go 1.27.0 — same module (`go.mod`) established by `001-list-outputs`; no
new module, no version change.

**Primary Dependencies**: Go standard library only — `net/http`, `encoding/json`, `flag`,
`errors` (tests: `net/http/httptest`). No third-party module is introduced; this feature adds a
new `internal/cli/route` package, one new function (`CreateRoute`) plus one new request struct
in the existing `internal/hub/routes.go`, and a new `internal/render/route.go`. See
[research.md §1-2](research.md#1-target-resolution-reuse-hubresolvetarget-unchanged--no-new-resolution-code).

**Storage**: Same single local user config file, `~/.config/sonora/config.json`, read (not
written) via the existing `config.ResolveHubURL` — unchanged from prior features.

**Testing**: `go test` (stdlib `testing`) across the same three layers established by
`001-list-outputs` — unit (`tests/unit`), contract (`tests/contract`, against `httptest.Server`s
shaped from `createRoute`'s request/response/error bodies in `api/openapi.json`), and
integration (`tests/integration`, full CLI invocation, using a fake hub serving
`/api/v2/inputs/{id}`, `/api/v2/outputs/{id}`, `/api/v2/groups/{id}`, and `/api/v2/routes`).
Written before implementation per Principle VI — see
[research.md §6](research.md#6-testing-strategy-contract-test-for-createroute-alone-integration-test-for-the-full-chain).

**Target Platform**: Cross-platform CLI binary (Linux/macOS/Windows), same single static Go
binary via `go build` — this feature adds a new top-level noun, not a new binary/target.

**Project Type**: Single project — CLI + internal packages (unchanged from prior features).

**Performance Goals**: Cold start to first HTTP request dispatched well under 50ms (constitution
Performance Standards); a full routing round trip (up to two pre-check lookups plus the
`createRoute` call) reported in under 5s under normal network conditions (SC-002).

**Constraints**: 5s total request timeout per HTTP call (reusing `hub.NewClient()` unchanged),
single attempt with no automatic retry on network/timeout failure (FR-009); no unbounded waits
(Principle IV); exit codes MUST distinguish at least: success, usage error, input not found,
target not found, route creation failed, and a connectivity/network error (FR-010) — two new
classes beyond the eight already established by `006-play-command`, see
[research.md §3](research.md#3-two-new-exit-code-classes-applied-locally-in-routerun--hubclassifyerror-unchanged).

**Scale/Scope**: One new command (`sonora route inputs/<id> <outputs|groups>/<id>`) under a new
`route` noun (verb-less, like `play`). No new flags beyond the existing `--json`/`--verbose`/
`--hub-url` surface — no `--volume`/`--display-name`/`--group`/`--output` (spec Assumptions).
See [data-model.md](data-model.md) and [contracts/cli-route.md](contracts/cli-route.md).

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Check | Status |
|---|---|---|
| I. Instant Startup & Low-Latency | `hub.NewClient()` is constructed inside `route.Run`, after `flag.Parse`/positional/prefix validation completes, identical in shape to `play`/`get`/`list` — no disk/network work occurs before argument parsing; the three hub calls happen only after both paths are validated. | PASS |
| II. API Contract Fidelity | `CreateRouteRequest` mirrors `#/components/schemas/CreateRouteRequest` field-for-field; the 201 body decodes into the existing `hub.Route` (`#/components/schemas/RouteResponse`, already validated by `validateRoute`); `POST /api/v2/routes`'s documented 400/404/422 `ErrorResponse` bodies are all mapped explicitly. The confirmation `message` FR-005 requires is deliberately *not* added as an invented field on `hub.Route` — it's constructed in the CLI layer, keeping the hub-facing type a faithful mirror of the API (research.md §4). | PASS |
| III. Minimal, Justified Dependencies | No new dependency added; this feature adds one new package (`internal/cli/route`), one new file (`internal/render/route.go`), and extends two existing stdlib-only files (`internal/hub/routes.go`, `internal/hub/errors.go`). `hub.ResolveTarget` and `hub.GetInput` are reused unchanged rather than reimplemented (research.md §1-2). | PASS |
| IV. Resilient, Transparent HTTP Client | Reuses `hub.NewClient()` (5s timeout, single attempt, no retries) unchanged for all three possible requests (input check, target check, create call); every non-2xx/network/decode failure is classified and translated to a plain message by default, with `--verbose` exposing the raw error; no panics — all failure paths map to distinguishable non-zero exit codes. | PASS |
| V. CLI UX Consistency | `sonora route inputs/<id> <outputs\|groups>/<id>` follows the same verb-less top-level noun shape `play` established; default output is YAML, `--json` switches to JSON; exit codes distinguish at least six classes (data-model.md's exit code table); flag names (`--json`/`--verbose`/`--hub-url`) are reused unchanged, per Principle V's "no synonyms" rule; the target's type is read solely from its path prefix, never a flag, matching the CLI command landscape's path-style addressing (spec clarifications). | PASS |
| VI. Test-First Development | Unit, contract, and integration tests (research.md §6) are written and reviewed before any implementation code, per the task breakdown this plan feeds into (`/speckit-tasks`); routing is called out by the constitution by name as a core flow requiring contract/integration coverage before merge. | PASS (planned; enforced at task/implementation time) |

No violations requiring justification — the Complexity Tracking table below is empty.

## Project Structure

### Documentation (this feature)

```text
specs/008-route-command/
├── plan.md               # This file (/speckit-plan command output)
├── research.md            # Phase 0 output (/speckit-plan command)
├── data-model.md          # Phase 1 output (/speckit-plan command)
├── quickstart.md          # Phase 1 output (/speckit-plan command)
├── contracts/              # Phase 1 output (/speckit-plan command)
│   └── cli-route.md
├── checklists/
│   └── requirements.md   # Pre-existing (speckit-checklist output)
└── tasks.md               # Phase 2 output (/speckit-tasks command - NOT created by /speckit-plan)
```

### Source Code (repository root)

```text
cmd/
└── sonora/
    └── main.go                        # MODIFIED: adds an early special case for `route`, alongside the existing `play` check, routing directly to route.Run(args[1:], ...); helpText gains a `route` entry

internal/
├── hub/
│   ├── client.go                       # UNCHANGED — hub.NewClient() reused as-is
│   ├── inputs.go                       # UNCHANGED — GetInput reused for the input pre-check
│   ├── outputs.go                      # UNCHANGED — GetOutput reused (via ResolveTarget) for output targets
│   ├── groups.go                       # UNCHANGED — GetGroup reused (via ResolveTarget) for group targets
│   ├── play.go                         # UNCHANGED — ResolveTarget reused as-is (research.md §1); errorResponse type reused by CreateRoute
│   ├── routes.go                       # MODIFIED: adds CreateRouteRequest struct and CreateRoute(ctx, client, baseURL, req) (*Route, error), reusing existing Route/validateRoute/errorResponse
│   └── errors.go                       # MODIFIED: adds ClassInputNotFound=11/ClassTargetNotFound=12 with their ExitCode() cases; ClassifyError itself is unchanged (research.md §3)
├── config/
│   └── config.go                       # UNCHANGED — config.ResolveHubURL reused as-is
├── cli/
│   ├── outputs/                        # UNCHANGED
│   ├── inputs/                         # UNCHANGED
│   ├── routes/                         # UNCHANGED
│   ├── groups/                         # UNCHANGED
│   ├── play/                           # UNCHANGED
│   ├── respath/                        # UNCHANGED — Inputs/Outputs/Groups kinds and aliases already registered
│   └── route/
│       └── route.go                    # NEW: `sonora route inputs/<id> <outputs|groups>/<id>` — flag definitions (--json/--verbose/--hub-url), two-positional parsing (play's re-parse-loop pattern), path-prefix validation (FR-002a/FR-002b), GetInput pre-check, ResolveTarget pre-check, CreateRoute call, local NotFoundError→ClassInputNotFound/ClassTargetNotFound dispatch, render call
└── render/
    ├── outputs.go                      # UNCHANGED
    ├── inputs.go                       # UNCHANGED
    ├── routes.go                       # UNCHANGED — RenderRouteYAML/JSON (full 10-field view) not reused here; route creation uses its own leaner view
    ├── groups.go                       # UNCHANGED
    ├── play.go                         # UNCHANGED
    └── route.go                        # NEW: RenderRouteCreatedYAML(hub.Route, message string), RenderRouteCreatedJSON(hub.Route, message string) — expose exactly routeId/status/message (FR-005)

tests/
├── contract/
│   └── route_test.go                   # NEW: httptest.Server shaped from createRoute's 201/400/404/422 response bodies; request body + decode contract
├── integration/
│   └── route_test.go                   # NEW: full `sonora route` invocation against a fake hub serving /api/v2/inputs/{id} + /api/v2/outputs/{id} + /api/v2/groups/{id} + /api/v2/routes — single-output target, group target, colliding identifiers, missing input, missing target, missing args, invalid prefixes, --json, every exit code class
└── unit/
    ├── cli_route_test.go                 # NEW: route.Run flag/positional parsing, prefix validation usage errors, dispatch, NotFoundError→exit-code mapping
    └── render_route_test.go              # NEW: YAML/JSON rendering of a route-creation result, field set matches FR-005 exactly
```

**Structure Decision**: New `internal/cli/route` package (mirroring `play`'s single-`Run`-entry-
point shape, since the API exposes one operation) plus one new file in `internal/render`, plus
targeted, additive extensions of `internal/hub/routes.go` (one new function) and
`internal/hub/errors.go` (two new classes). No new top-level directories; no changes to
`outputs`/`inputs`/`groups`/`play`/`respath` beyond reusing their existing exported functions
unchanged, and no change to `hub.ClassifyError`'s existing behavior for any other command.

## Complexity Tracking

*No entries — Constitution Check reported no violations requiring justification.*
