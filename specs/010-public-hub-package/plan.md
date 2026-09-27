# Implementation Plan: Public Hub Client Package

**Branch**: `010-public-hub-package` | **Date**: 2026-09-27 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `/specs/010-public-hub-package/spec.md`

## Summary

Make the Multiroom Audio Hub API client importable by other Sonora projects without
changing CLI behavior. Rename the module to `github.com/Sonora-Multiroom/sonora-cli`,
`git mv internal/hub hub`, move exit-code mapping (and the CLI-only `ClassUsage`) into a
new `internal/cli/exitcode` package, move the CLI-specific hints out of `hub`'s TTS error
messages into `internal/cli/tts` (FR-015), rewrite `hub` godoc to be self-contained,
publish `api/openapi.json` as `api.Spec` via `go:embed`, fix all three tracked `-ldflags -X`
version sites (including `.goreleaser.yaml`) plus the local untracked `build.sh`, add guard
tests for the version injection, the godoc rules, CLI-neutral messages and stray old-path
references, then tag `v0.1.0`.

## Technical Context

**Language/Version**: Go 1.27.0 (`go.mod`)

**Primary Dependencies**: standard library only (unchanged; `embed` added for `api`)

**Storage**: N/A

**Testing**: `go test ./...` — existing `tests/unit`, `tests/integration`, `tests/contract`
(external test packages, `httptest` fakes); `cmd/sonora/main_test.go`

**Target Platform**: Windows and Linux (CI: `ubuntu-latest`); release binaries via
GoReleaser for linux/darwin/windows

**Project Type**: CLI + public Go library packages in the same module

**Performance Goals**: no change to CLI startup (constitution I); `cmd/sonora` does not
import `api`, so the embedded spec does not enter the binary

**Constraints**: zero observable CLI change (FR-010); `hub` standard-library-only
(FR-008); published tags immutable (FR-014)

**Scale/Scope**: 111 `.go` files change import path; 77 import the client; 30 CLI files
(157 calls) call `ExitCode()`; ~137 `hub.ClassUsage` references; 57 godoc lines to rewrite across
10 `hub` files; 3 tracked build-config sites (+ local `build.sh`)

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

Constitution v1.2.0.

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Instant Startup | Pass | No startup-path work added; package move only. `api` not linked into `sonora`. |
| II. API Contract Fidelity | Pass | `hub/` becomes the single spec→Go place for all consumers; no type or endpoint changes; contract tests kept. |
| III. Minimal Dependencies | Pass | No new modules; `embed` is stdlib. |
| IV. Resilient HTTP | Pass | Timeouts, no-retry, and classification unchanged. TTS "hub address"/"version mismatch" messages lose their CLI hints in `hub`; `internal/cli/tts` adds them back, so the CLI's messages are unchanged. |
| V. CLI UX Consistency | Pass | Commands, output, exit codes unchanged; exit-code table preserved in `exitcode.For` and re-asserted by moved tests. |
| VI. Test-First | Pass | New behavior = guards: version-injection test, godoc-rules test, CLI-neutral-message test, stray-reference test, `api.Spec` test, `exitcode` tests are written first and fail before the change (see research §2, §4, §5, §6, §7, §9). Existing tests are the refactor's safety net. |
| VII. Public Hub Client Package | Pass | This feature establishes it: protocol-only `hub/`, CLI concerns (`ExitCode`, `ClassUsage`) moved to `internal/`, self-contained godoc, stdlib-only, contract tests stay here, `v0.1.0` tag with release notes. |

**Post-design re-check**: Pass — design artifacts add no dependencies, no startup work, and
no behavior; the only public-API removals (`ClassUsage`, `ExitCode()`) are CLI concerns
required out by VII.

## Project Structure

### Documentation (this feature)

```text
specs/010-public-hub-package/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   └── public-packages.md
├── checklists/
│   └── requirements.md
└── tasks.md             # /speckit-tasks
```

### Source Code (repository root)

```text
go.mod                         # module github.com/Sonora-Multiroom/sonora-cli
Makefile                       # -X …/internal/version.Version (build, docker-build)
build.sh                       # local, untracked helper (never committed): -X …/internal/version.Version
.goreleaser.yaml               # -X …/internal/version.Version (release binaries)

api/
├── openapi.json               # unchanged, single copy
└── spec.go                    # NEW: package api; //go:embed openapi.json; var Spec []byte

hub/                           # MOVED from internal/hub (git mv), package hub
├── client.go  errors.go  extensions.go  groups.go  inputs.go
├── mastermute.go  outputs.go  play.go  routes.go  tts.go
                               # errors.go: ExitCode() and ClassUsage removed; godoc rewritten
                               # tts.go: TTSUnavailableError messages made CLI-neutral

cmd/sonora/
├── main.go
└── main_test.go               # + version-injection config test, + build-and-run test (skipped in -short)

internal/
├── cli/
│   ├── exitcode/              # NEW: Usage const, For(hub.ErrorClass) int
│   └── groups/ inputs/ mastermute/ outputs/ play/ route/ routes/ tts/ clihelp/ respath/
│                              # tts/report.go: adds back the CLI hints hub no longer carries
├── config/  render/  version/ # unchanged apart from import paths

tests/
├── contract/                  # import path → …/sonora-cli/hub
├── integration/               # import path
└── unit/
    ├── hub_client_test.go     # ExitCode assertions removed (moved); + CLI-neutral message test
    ├── exitcode_test.go       # NEW: mapping table, uniqueness, 7 retired, Usage == 2
    ├── hub_godoc_test.go      # NEW: every exported ident documented; no spec-artifact refs
    ├── api_spec_test.go       # NEW: api.Spec == api/openapi.json bytes, valid JSON
    └── stray_refs_test.go     # NEW: no old module path or internal/hub in live files
```

**Structure Decision**: Single Go module. Two public packages at the root (`hub`, `api`);
everything CLI-specific stays under `internal/`. Existing test layout (`tests/unit`,
`tests/integration`, `tests/contract` as external packages) is kept, which also exercises
`hub` purely through its exported API.

## Implementation Order

1. **Guards first (red)**: `exitcode_test.go`, `api_spec_test.go`, `hub_godoc_test.go`,
   the CLI-neutral-message test, version-injection tests in `cmd/sonora/main_test.go` —
   they compile (via small stubs where needed) and fail on their assertions, so the rest
   of the suite keeps running.
2. **Module rename** (`go.mod` + 111 files + 3 build-config sites + local `build.sh`) → build, full tests green.
3. **`git mv internal/hub hub`** + 77 import updates → green. Commit separately so history
   shows a pure rename.
4. **`internal/cli/exitcode`**; replace `ExitCode()`/`ClassUsage` call sites; delete them
   from `hub`; move old exit-code tests → green, `exitcode_test.go` green.
5. **CLI-neutral messages**: `hub` TTS messages drop CLI hints; `internal/cli/tts` adds
   them back → CLI-neutral-message test green, CLI TTS tests unchanged and green.
6. **Godoc rewrite** in `hub` + package comment → `hub_godoc_test.go` green.
7. **`api/spec.go`** → `api_spec_test.go` green.
8. **`stray_refs_test.go`** (red on leftover comments such as
   `internal/cli/groups/volume.go:24`) → fix hits → green; quickstart §2 grep as a
   cross-check.
9. Quickstart §1–5 on Windows and Linux; merge; `./release.sh --minor` → `v0.1.0`;
   quickstart §5 (post-tag) and §6.

Commits follow Conventional Commits with this repo's types (CONTRIBUTING.md): `refactor!`
for the module rename (including its `Makefile`/`.goreleaser.yaml` edits, which must land
atomically with `go.mod`) and exit-code move, `refactor` for the `git mv` and godoc, `feat` for
`api.Spec`, `spec` for release notes. Guard tests ship in the commit they guard.

## Complexity Tracking

No constitution violations.
