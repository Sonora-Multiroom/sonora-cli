# Implementation Plan: Adopt Verb-First Command Landscape

**Branch**: `007-refactor-cli-commands` | **Date**: 2026-08-27 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `/specs/007-refactor-cli-commands/spec.md`

## Summary

Refactor the already-shipped read commands (`inputs`/`outputs`/`groups`/`routes` `list`/`get`)
and `play` off the old `sonora <resource> <verb>` grammar onto the verb-first, path-style
grammar in `docs/cli-command-landscape.md` (`sonora get <resource>[/<id>]`, `sonora list
<resource>`, resource aliases `in`/`out`/`gr`/`rt`, and `play <uri> <outputs|groups>/<id>`).
This is a dispatch/argument-parsing-layer change: the existing `internal/cli/<resource>`
business logic (HTTP calls, rendering, exit codes) is reused unchanged wherever possible —
only how a command line is parsed into a call to that logic changes. `play` additionally
drops its `--group`/`--output`/`--name` flags for a target path and `--display-name`.

## Technical Context

**Language/Version**: Go 1.27.0 (per `go.mod`)

**Primary Dependencies**: Go standard library only (`flag`, `net/http`, `encoding/json`,
`strings`) — `go.sum` is empty; this refactor introduces no third-party dependency.

**Storage**: N/A — stateless CLI; all state lives in the remote Multiroom Audio Hub.

**Testing**: `go test` (stdlib `testing`), following this repo's existing three-tier layout:
`tests/unit` (parsing/rendering logic), `tests/contract` (mock-server-backed HTTP contract
tests), `tests/integration` (end-to-end command behavior). Test-First per Constitution
Principle VI.

**Target Platform**: Cross-platform CLI binary — Windows (Scoop release), Linux (Docker-based
build), macOS.

**Project Type**: Single-project CLI (Go module `sonora-cli`).

**Performance Goals**: No regression against the Constitution's <50ms cold-start-to-first-request
budget — this refactor changes only in-process argument parsing (string splitting/matching),
adding no I/O, reflection, or new dependencies to the startup path.

**Constraints**:
- Hard cutover, no back-compat shim (spec Clarifications) — old command forms are deleted,
  not kept as hidden aliases.
- Byte-for-byte identical displayed data for every refactored read command (FR-010/SC-004) —
  only invocation shape changes.
- No new dependencies (Constitution Principle III).

**Scale/Scope**: 9 existing entry points refactored at the dispatch/parsing layer (4 resources
× `list`+`get`, plus `play`); `internal/cli/<resource>` packages' request/render logic is
expected to remain largely untouched.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Assessment |
|---|---|
| I. Instant Startup & Low-Latency Execution | **Pass.** The new dispatcher does string splitting/matching only, before argument parsing completes, same as today's noun/verb switch. No new I/O or blocking work is introduced. |
| II. API Contract Fidelity | **Pass.** No HTTP request/response handling changes; the same `openapi.json`-derived types and endpoints are used. FR-010 explicitly requires unchanged displayed data. |
| III. Minimal, Justified Dependencies | **Pass.** Implemented with `strings`/`regexp` from the standard library only; no new dependency. |
| IV. Resilient, Transparent HTTP Client Behavior | **Pass.** Not touched by this refactor — timeouts, error translation, and exit-code classing for hub/network failures are reused as-is. |
| V. CLI UX Consistency | **Flagged, not blocking.** This principle's illustrative example — "e.g. `sonora <noun> <verb> [args]`" — textually names the exact grammar this feature replaces. The *substantive* requirements (a consistent, predictable structure; no synonym flag/argument names across commands) are still met: the new verb-first grammar is applied uniformly across all four resources and matches the already-approved `route` command (spec 008), and `list`/`get` is a single, documented, collection-only verb synonym — not a set of divergent flag/argument names for the same concept. Recommend running `/speckit-constitution` after this feature ships to update Principle V's example to the new grammar; not required to unblock this plan. |
| VI. Test-First Development | **Pass, with a process note.** `/speckit-tasks` MUST sequence a failing test (updated `tests/unit`/`tests/contract`/`tests/integration` cases for the new invocation shape) before each corresponding implementation change, per the existing Red-Green-Refactor requirement. |

No unjustified violations — Complexity Tracking is not needed.

**Post-Phase-1 re-check**: Design artifacts (research.md, data-model.md, contracts/,
quickstart.md) introduce no new dependency, no new I/O on the startup path, and no change to
HTTP/error-handling behavior — the table above still holds unchanged after design.

## Project Structure

### Documentation (this feature)

```text
specs/007-refactor-cli-commands/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/           # Phase 1 output
│   ├── cli-get-list.md
│   └── cli-play.md
└── tasks.md             # Phase 2 output (/speckit-tasks — not created here)
```

### Source Code (repository root)

```text
cmd/sonora/
└── main.go                    # Rewritten: verb-first dispatch (get/list/play),
                                # resource-path resolution via internal/cli/respath

internal/cli/
├── respath/                   # NEW: shared resource-path parsing
│   └── respath.go             #   alias table (in/out/gr/rt), id pattern validation,
│                               #   "resource/id" split — used by main.go and play
│                               #   (unit-tested from tests/unit/respath_test.go,
│                               #   matching this repo's existing test-file convention)
├── inputs/                    # RunList/RunGet: unchanged internally; called with the
├── outputs/                   #   same positional-id argv shape as today, now supplied
├── groups/                    #   by the new dispatcher instead of the old noun/verb switch
├── routes/
├── play/
│   └── play.go                # Rewritten: target positional → respath-parsed path;
│                               #   --group/--output removed; --name → --display-name
└── clihelp/
    └── usage.go                # Updated help text for the new grammar

tests/
├── unit/                       # Updated per refactored command + new respath_test.go
├── contract/                   # Updated invocation-shape assertions; HTTP contracts unchanged
└── integration/                # Updated per refactored command
```

**Structure Decision**: Single-project Go CLI (existing layout retained). The refactor is
concentrated in a new `internal/cli/respath` package plus `cmd/sonora/main.go`'s dispatcher
and `internal/cli/play`; the four resource packages' `RunList`/`RunGet` functions are reused
without signature changes wherever the new dispatcher can supply the same argument shape they
already expect.

## Complexity Tracking

*No entries — Constitution Check found no unjustified violations.*
