# Implementation Plan: Text-to-Speech Announcements (`speak`, TTS cache)

**Branch**: `009-tts-commands` | **Date**: 2026-09-21 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `/specs/009-tts-commands/spec.md`

**Note**: This template is filled in by the `/speckit-plan` command; its definition describes the execution workflow.

## Summary

This feature adds three commands. Each one wraps exactly one operation of the hub's optional TTS
extension:

- `sonora speak <text|-> <outputs|groups>/<id> [--provider] [--voice] [--language]` →
  `POST /api/tts/speak`
- `sonora get tts-cache` → `GET /api/tts/cache/stats`
- `sonora clear tts-cache [--provider]` → `DELETE /api/tts/cache`

The TTS operations differ from every existing command in two ways. They can be missing
(404 when the extension isn't active), and they report errors as `{error, message}` rather than
the core problem-details shape. Both are handled by one shared failure path in a new
`internal/cli/tts` package:

- Extension error codes map onto the existing exit-code classes. One new class,
  `ClassTTSUnavailable` = 13, is added.
- A 404 triggers exactly one `listExtensions` lookup to explain *why* TTS is unavailable. If the
  inventory also answers 404, the command reports a hub-address problem with exit 4.

`speak` gets a 15 s request bound so the hub's own 10 s provider-timeout error arrives first. It
can also read its text from standard input (`-`). See [research.md](research.md) for the
decisions.

## Technical Context

**Language/Version**: Go 1.27.0 (same module; no version change).

**Primary Dependencies**: Go standard library only: `net/http`, `net/url`, `encoding/json`,
`flag`, `io`, `errors`, `sort`, `strings`, `time` (tests: `net/http/httptest`, `os/exec`). No new
module.

**Storage**: N/A. The hub URL is resolved through the existing `config.ResolveHubURL`
(`--hub-url` → `MULTIROOM_URL` → `~/.config/sonora/config.json`), and nothing is written.

**Testing**: `go test`, in the three established layers: unit (`tests/unit`), contract
(`tests/contract`, `httptest.Server`s shaped from `speak`/`getCacheStats`/`clearCache`/
`listExtensions` in `api/openapi.json`), and integration (`tests/integration`, the built binary
against a request-counting fake hub, with standard input piped in). Tests are written first;
see [research.md §9](research.md#9-testing-strategy-principle-vi).

**Target Platform**: Cross-platform single static binary (Linux/macOS/Windows), unchanged.

**Project Type**: Single project, CLI plus internal packages.

**Performance Goals**: Cold start to first request stays well under 50 ms (unchanged path).
`speak` with a cache hit is accepted in under 1 s on a LAN (SC-002).

**Constraints**:

- Timeouts: 5 s per request for the cache commands and the inventory lookup; 15 s for `speak`
  (FR-013). No retries (FR-014).
- A successful command makes exactly one request (SC-005a).
- The failure path makes at most one extra request, the inventory lookup (FR-010a).
- No CLI-side text length limit (FR-003).

**Scale/Scope**: Three commands, one new verb (`speak`), one new verb (`clear`) and one
singleton keyword (`tts-cache`). Four hub operations are consumed, one of them internally only.
One exit-code class is added.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Check | Status |
|---|---|---|
| I. Instant Startup | Dispatch adds three string comparisons in `run`. Clients are built inside the `Run*` functions only after flags, positionals, target path and standard input are validated. Standard input is read only for `-`, and only after the target is validated (research §10). | PASS |
| II. API Contract Fidelity | Wire types mirror `SpeakRequest`, `SpeakAcceptedResponse`, `CacheStats`, `TtsErrorResponse`, `ExtensionInventory` and `Extension`. Every documented status (202/200/204/400/503) is handled explicitly. The required-field rules (FR-005a) are *stricter* than the published spec, which marks no field required, and that tightening is a spec decision rather than an invented field. The extension's own contract is honoured where more precise (FR-016). No undocumented endpoint is used; `/actuator/extensions` is explicitly excluded. | PASS |
| III. Minimal Dependencies | No new module. New files: `hub/tts.go`, `hub/extensions.go`, `cli/tts/*`, `render/tts.go`. Existing `respath`, `config`, `clihelp`, `ClassifyError` and exit classes are reused. | PASS |
| IV. Resilient HTTP Client | Every request has an explicit bound: 5 s, or 15 s via the new `NewClientWithTimeout`, with `NewClient` unchanged. Every failure (TTS code, non-TTS body, 404, inventory failure, malformed body, network) maps to a distinguishable exit code with a plain message, and the raw error is available under `--verbose`. Error bodies are read through a size limit. No panics. Connection reuse holds through the shared `DefaultTransport`. | PASS |
| V. CLI UX Consistency | Verb-first: `speak`, `get tts-cache`, `clear tts-cache`. YAML by default and `--json` on all three. `--json`/`--hub-url`/`--verbose` are unchanged; `--provider` means the same thing on both commands that take it. Target addressing is path-style with aliases, as for `play`/`transfer`. The new exit code 13 extends the documented table without moving any existing code. | PASS |
| VI. Test-First | Contract, unit and integration tests are listed per behaviour in research §9 and must be written and failing before implementation. `/speckit-tasks` orders them first. | PASS (planned; enforced at task/implementation time) |

**Post-design re-check** (after Phase 1): still PASS on all six. The design adds no dependency,
no startup work, no unbounded wait and no undocumented endpoint. The only change touching
shared code is additive:

- a new `ErrorClass`;
- two new `ClassifyError` branches, keyed on new error types, so existing errors can't reach
  them;
- `NewClientWithTimeout`, with `NewClient` delegating to it using the same 5 s.

## Project Structure

### Documentation (this feature)

```text
specs/009-tts-commands/
├── plan.md              # This file
├── research.md          # Phase 0: design decisions
├── data-model.md        # Phase 1: wire types, error types, exit codes, rendered views
├── quickstart.md        # Phase 1: validation guide
├── contracts/
│   └── cli-tts.md       # Phase 1: command shapes, output, messages, exit codes
├── checklists/
│   └── requirements.md  # Spec quality checklist (from /speckit-specify)
└── tasks.md             # Phase 2 (/speckit-tasks, not created here)
```

### Source Code (repository root)

```text
cmd/sonora/
├── main.go                  # MODIFIED: `speak` special case → tts.RunSpeak(args[1:], os.Stdin, …);
│                            #   `tts-cache` keyword in dispatchGetList (like master-mute);
│                            #   new `clear` verb → dispatchClear (tts-cache only); helpText entries
└── main_test.go             # MODIFIED: help text lists new commands; clear dispatcher usage errors

internal/hub/
├── client.go                # MODIFIED: NewClientWithTimeout(d); NewClient() delegates (5 s, unchanged)
├── errors.go                # MODIFIED: ClassTTSUnavailable (13); ClassifyError branches for
│                            #   *TTSUnavailableError and *TTSError (existing branches untouched)
├── tts.go                   # NEW: SpeakRequest, SpeakAccepted, TTSCacheStats; Speak, GetTTSCacheStats,
│                            #   ClearTTSCache; TTSError, TTSNotOfferedError, TTSUnavailableError,
│                            #   TTSDiagnosis; SpeakTimeout = 15 s
└── extensions.go            # NEW: ExtensionInventory, Extension; ListExtensions

internal/cli/tts/
├── speak.go                 # NEW: RunSpeak(args, stdin, stdout, stderr): parsing loop with `--` fix,
│                            #   `-` → stdin, target kind → targetType, empty-value checks
├── cache.go                 # NEW: RunGetCache, RunClearCache
└── report.go                # NEW: shared failure path: 404 → one ListExtensions → TTSUnavailableError;
                             #   message formatting (`<code>: <msg>`), --verbose detail, exit code

internal/render/
└── tts.go                   # NEW: RenderSpeak{YAML,JSON}, RenderTTSCacheStats{YAML,JSON},
                             #   RenderTTSCacheCleared{YAML,JSON}

tests/
├── contract/
│   ├── tts_test.go          # NEW: speak / getCacheStats / clearCache request + decode + error contracts
│   └── extensions_test.go   # NEW: listExtensions decode, 404, malformed body
├── integration/
│   └── tts_test.go          # NEW: built binary vs request-counting fake hub; runCLIWithStdin helper;
│                            #   every exit code; slow-provider test (skipped with -short)
└── unit/
    ├── cli_tts_speak_test.go    # NEW
    ├── cli_tts_cache_test.go    # NEW
    ├── tts_report_test.go       # NEW: diagnosis table + code → exit-class table
    └── render_tts_test.go       # NEW

README.md                        # MODIFIED: command table + examples (FR-015)
docs/cli-command-landscape.md    # MODIFIED: new "tts (extension)" section, ✅ rows (FR-015)
```

**Structure Decision**: One new command package, `internal/cli/tts`, because the three commands
share an error model that no other command uses (research §1). There are two new hub files, one
new render file, and additive edits to `hub/client.go`, `hub/errors.go` and `cmd/sonora/main.go`.
There are no new top-level directories and no behaviour changes to existing commands.

## Complexity Tracking

*No entries. The Constitution Check reported no violations requiring justification.*
