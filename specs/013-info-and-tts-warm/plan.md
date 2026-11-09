# Implementation Plan: Hub Info and TTS Cache Warm-up

**Branch**: `013-info-and-tts-warm` | **Date**: 2026-11-06 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `/specs/013-info-and-tts-warm/spec.md`

## Summary

Add two operations from `api/openapi.json` 0.1.33 to the public `hub` package and the CLI: `GetInfo`
(`GET /api/v2/info`) behind `sonora get info`, and `WarmTTSCache` (`POST /api/tts/cache/warm`) behind
`sonora warm tts-cache <text|->`. Both follow patterns already in the repository: `GetInfo` is shaped like
`GetMasterMute`, and `WarmTTSCache` reuses `Speak`'s request building, 404 handling and `TTSError`
decoding. The CLI side reuses `tts.ReportError`, `clihelp`, the YAML/JSON renderers' conventions and the
`speak` flags.

## Technical Context

**Language/Version**: Go (version in `go.mod`)

**Primary Dependencies**: standard library only (Principles III and VII)

**Storage**: N/A

**Testing**: `go test ./...`: contract tests in `tests/contract` (mock server, spec-derived shapes), unit
tests for renderers in `tests/unit`, command tests in `tests/integration`

**Target Platform**: the `sonora` binary on Windows, macOS and Linux; the `hub` package as a library

**Project Type**: CLI plus public Go client library

**Performance Goals**: no change to the startup path (Principle I); no new work before argument parsing

**Constraints**: explicit request timeouts (Principle IV): 5s for info, `hub.SpeakTimeout` (15s, overridable
with `--timeout`) for warm; `hub/` stays protocol-only

**Scale/Scope**: two endpoints, two commands, about 6 new files and 5 touched

## Constitution Check

| Principle | Status | How |
|-----------|--------|-----|
| I. Instant startup | Pass | Two new `case` arms in the dispatcher; no init work. |
| II. Contract fidelity | Pass | Warm types mirror `WarmRequest`/`WarmResponse`; `Info` mirrors `HubInfoResponse` (research R1). |
| III. Minimal dependencies | Pass | stdlib only. |
| IV. Resilient client | Pass | Timeouts set; 404, 400, 503 and malformed bodies each map to a distinct message and exit code; `--verbose` keeps the cause. |
| V. CLI consistency | Pass | `sonora get info` (verb first); `sonora warm tts-cache` adds one verb, justified in research R3. YAML default, `--json`, same flag names as `speak`. |
| VI. Test-first | Pass | Tasks order every test before its code. |
| VII. Public hub package | Pass | New exported identifiers have self-contained godoc; no CLI concerns in `hub/`; additive change, so a minor tag. |

Post-design re-check: unchanged. No violations, so no Complexity Tracking.

## Project Structure

### Documentation (this feature)

```text
specs/013-info-and-tts-warm/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   ├── cli.md           # commands, flags, output, exit codes
│   └── hub-package.md   # new exported identifiers
├── checklists/requirements.md
└── tasks.md             # /speckit-tasks
```

### Source Code (repository root)

```text
hub/
├── info.go                      # new: Info, GetInfo, InfoNotOfferedError
├── tts.go                       # modified: WarmRequest, WarmResult, WarmTTSCache
└── errors.go                    # modified: ClassifyError for InfoNotOfferedError, CACHE_WRITE_FAILED
internal/
├── cli/
│   ├── hubinfo/info.go          # new: RunGet (`get info`)
│   └── tts/warm.go              # new: RunWarm (`warm tts-cache`)
└── render/
    ├── info.go                  # new: YAML and JSON
    └── tts.go                   # modified: warm result YAML and JSON
cmd/sonora/main.go               # modified: `get info`, `warm`, help text
tests/
├── contract/info_test.go        # new
├── contract/tts_test.go         # modified: warm cases
├── unit/render_info_test.go     # new
├── unit/render_tts_warm_test.go # new
└── integration/info_test.go, tts_warm_test.go   # new
README.md                        # modified
```

**Structure Decision**: the existing layout. `get info` is its own small package like `mastermute`; `warm`
joins `internal/cli/tts` because it shares `ReportError`.

## Complexity Tracking

None.
