# Implementation Plan: Align the CLI with Hub 0.1.25–0.1.28

**Branch**: `011-hub-0128-alignment` | **Date**: 2026-10-22 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `specs/011-hub-0128-alignment/spec.md`

## Summary

Three small corrections after the `api/openapi.json` refresh to hub 0.1.28, with no new command or flag:

1. **Docs**: the enable/disable help texts and README say what each command does to running playback
   (hub 0.1.25 semantics). The help gets a description line under the usage line, which `clihelp`
   already prints verbatim.
2. **Play name**: `hub.PlaybackResponse` gains `DisplayName`; `sonora play` renders `displayName` after
   `inputId` when non-empty and omits it otherwise; `--display-name` help and README describe the hub's
   naming.
3. **Transient refusal**: `hub` exports `ReasonRouteChangingOutputs` and `IsTransientConflict(err)`,
   and `ClassifyError` gives that refusal its own message. `sonora transfer` retries only that refusal
   on a fixed backoff (100, 200, 400, 800 ms; at most 4 retries, 1.5 s of waiting), then exits 14 with
   the transient message.

## Technical Context

**Language/Version**: Go 1.27 (`go.mod`)

**Primary Dependencies**: standard library only (`net/http`, `encoding/json`, `flag`, `time`)

**Storage**: N/A

**Testing**: `go test ./...`: `tests/unit` (render, classification, help), `tests/contract` (`hub` against
`httptest` servers shaped by `api/openapi.json`), `tests/integration` (whole commands against a mock hub)

**Target Platform**: Windows and Linux (amd64, arm64 for the Pi)

**Project Type**: CLI plus a public Go client package (`hub/`)

**Performance Goals**: startup unchanged (< 50 ms to first request); an accepted retry adds at most
the 100 ms first delay plus one round trip

**Constraints**: each `hub` call stays one HTTP request; no change to any other command's output; older
hubs (< 0.1.28 without `displayName`, < 0.1.25 without the reason) keep working

**Scale/Scope**: about 10 source files, 1 README, 4 test files extended

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Check | Status |
|---|---|---|
| I. Instant startup | Nothing added before argument parsing; the retry runs only after a 409 | Pass |
| II. Contract fidelity | `PlaybackResponse` brought back to field-for-field; reason handling follows the 0.1.28 `ErrorResponse` description; tests use 0.1.28 shapes | Pass |
| III. Minimal dependencies | Standard library only | Pass |
| IV. Resilient HTTP | Each attempt keeps the 5 s client timeout; retries bounded (1.5 s waiting); network errors not retried; `--verbose` still shows the underlying error | Pass |
| V. CLI UX consistency | No new verbs or flags; exit code 14 kept for the conflict class; YAML default and `--json` both updated | Pass |
| VI. Test-first | Every task pair starts with a failing test (render, contract, integration, help) | Pass (enforced in tasks) |
| VII. Public `hub/` package | Only protocol concerns added (a reason constant, an error predicate, a message); retry policy stays in `internal/cli/route`; godoc self-contained; additive, so a minor tag | Pass |

Post-design re-check: unchanged, all pass. No Complexity Tracking entries.

## Project Structure

### Documentation (this feature)

```text
specs/011-hub-0128-alignment/
├── plan.md              # This file
├── research.md          # Phase 0: decisions
├── data-model.md        # Phase 1: changed types and rendered records
├── quickstart.md        # Phase 1: validation scenarios
├── contracts/
│   ├── cli.md           # Help text, play output, transfer behaviour, exit codes
│   └── hub-package.md   # Additions to the public hub/ API
└── tasks.md             # Phase 2 (/speckit-tasks)
```

### Source Code (repository root)

```text
hub/
├── play.go              # PlaybackResponse.DisplayName
└── errors.go            # ReasonRouteChangingOutputs, IsTransientConflict, ClassifyError message, doc comments

internal/
├── render/play.go       # displayName after inputId, omitted when empty
└── cli/
    ├── play/play.go     # --display-name help text
    ├── route/transfer.go# retry loop on the transient refusal
    ├── inputs/enabled.go   # description line per verb
    ├── outputs/enabled.go  # description line per verb
    └── groups/enabled.go   # description line per verb

README.md                # enable/disable paragraph, play paragraph, exit-code 14 row

tests/
├── unit/render_play_test.go      # displayName present / absent / non-ASCII
├── unit/hub_client_test.go       # ClassifyError message, IsTransientConflict
├── unit/cli_*_enabled_test.go    # help description lines (or one table test)
├── contract/play_test.go         # displayName decodes; omitted decodes to ""
├── contract/route_refusal_test.go# ROUTE_CHANGING_OUTPUTS decodes with no outputId
└── integration/route_test.go     # transfer: retry succeeds, retries run out, other reasons not retried
```

**Structure Decision**: existing single-module layout; no new packages or files outside `specs/`.

## Complexity Tracking

None.
