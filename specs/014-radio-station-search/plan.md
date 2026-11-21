# Implementation Plan: Radio Station Search and Browse

**Branch**: `014-radio-station-search` | **Date**: 2026-11-21 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `/specs/014-radio-station-search/spec.md`

## Summary

Add the five read-only operations of the hub's `radio` extension (RADIO-002, radio 0.2.0, already in
`api/openapi.json` 0.1.35) to the public `hub` package: `SearchStations`, `Station`, `Countries`,
`Languages`, `Tags`. Put them behind a new `search` verb (`sonora search radio-stations [name]`) and four
`get`/`list` nouns (`radio-stations/<uuid>`, `radio-countries`, `radio-languages`, `radio-tags`). Radio
errors decode into the existing `*hub.APIError`, which gains `Type`, `Parameter` and `RetryAfter`. A 404
that is not "station not found" is explained by the TTS diagnosis, generalised to any extension id, and
exits with a new code 15. Output is every field in YAML or JSON, with catalogue text escaped.

## Technical Context

**Language/Version**: Go 1.27 (`go.mod`)

**Primary Dependencies**: standard library only (Principles III and VII)

**Storage**: N/A

**Testing**: `go test ./...`: contract tests in `tests/contract` (test servers, fixture keys checked
against `api.Spec`), renderer and helper tests in `tests/unit`, command tests in `tests/integration`

**Target Platform**: the `sonora` binary on Windows, macOS and Linux; the `hub` package as a library for
sonora-mcp

**Project Type**: CLI plus public Go client library

**Performance Goals**: no change to the startup path; one radio request per command, the inventory
request only after a failure (FR-007, FR-013)

**Constraints**: `hub.RadioTimeout` 10 s, above the hub's 5 s catalogue budget (research R2); inventory
diagnosis 5 s; `hub/` stays protocol-only; TTS messages and behaviour unchanged

**Scale/Scope**: five endpoints, five commands, about 8 new files and 8 touched

## Constitution Check

| Principle | Status | How |
|-----------|--------|-----|
| I. Instant startup | Pass | New `case` arms and name checks in the dispatcher; no init work, no probe request (FR-013). |
| II. Contract fidelity | Pass | Types mirror the `Radio*` schemas; fixtures checked against `api.Spec`; spec already current (research R1). Only described parameters are sent; enum and pattern checks are left to the hub (research R4). |
| III. Minimal dependencies | Pass | stdlib only. |
| IV. Resilient client | Pass | Explicit timeouts (R2); every hub answer maps to its own message and exit code, including a new class for "radio not available" (R5, R6); `--verbose` keeps the cause. |
| V. CLI consistency | Pass | Verb first; `search` is a new verb, justified in research R3 (a search is not a `get` of a collection). YAML default, `--json`, `--verbose`, `--hub-url` as everywhere; the new flags (`--name`, `--tag`, `--limit`, `--offset`, …) have no existing synonym (research R4); new exit code 15 is a distinct failure class. |
| VI. Test-first | Pass | Tasks put every test before its code; the existing TTS diagnosis tests guard the shared refactor. |
| VII. Public hub package | Pass | Additive only: new identifiers, new fields, `TTSDiagnosis` kept as an alias of `ExtensionDiagnosis`, `ClassRadioUnavailable` appended so existing values keep their numbers. Godoc self-contained; rendering, escaping and exit codes stay in `internal/`. Minor tag. |

Post-design re-check: unchanged. No violations, so no Complexity Tracking.

## Project Structure

### Documentation (this feature)

```text
specs/014-radio-station-search/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   ├── cli.md           # commands, flags, output, exit codes, help lines
│   └── hub-package.md   # new and changed exported identifiers
├── checklists/requirements.md
└── tasks.md             # /speckit-tasks
```

### Source Code (repository root)

```text
hub/
├── radio.go                     # new: Radio* types, StationSearch, TagSearch, RadioTimeout,
│                                #      SearchStations, Station, Countries, Languages, Tags,
│                                #      radio error decoding, RadioNotOfferedError
├── extensions.go                # modified: Extension.Version, ExtensionDiagnosis, diagnoseExtension,
│                                #           RadioUnavailableError, DiagnoseRadioUnavailable
├── tts.go                       # modified: TTSDiagnosis becomes an alias; DiagnoseTTSUnavailable
│                                #           uses diagnoseExtension
└── errors.go                    # modified: APIError.Type/Parameter/RetryAfter,
                                 #           ClassRadioUnavailable, radio branches in ClassifyError
internal/
├── cli/
│   ├── radio/                   # new package
│   │   ├── search.go            # RunSearch (`search radio-stations`)
│   │   ├── station.go           # RunGetStation (`get radio-stations/<uuid>`)
│   │   ├── lists.go             # RunCountries, RunLanguages, RunTags
│   │   └── report.go            # ReportError: diagnosis, hints, exit code
│   └── exitcode/exitcode.go     # modified: ClassRadioUnavailable → 15
└── render/
    └── radio.go                 # new: YAML and JSON for stations, pages, countries, languages, tags;
                                 #      terminal-safe escaping
cmd/sonora/main.go               # modified: `search` verb, radio nouns in dispatchGetList, help text
tests/
├── contract/radio_test.go       # new
├── contract/extensions_test.go  # modified: radio diagnoses, Version
├── unit/render_radio_test.go    # new
├── unit/radio_report_test.go    # new
├── unit/exitcode_test.go        # modified
└── integration/radio_test.go    # new
README.md                        # modified: radio commands, needs radio 0.2.0, exit code 15
```

**Structure Decision**: the existing layout. The radio commands share one failure path, so they live in
one `internal/cli/radio` package, as the TTS commands do in `internal/cli/tts`.

## Complexity Tracking

None.
