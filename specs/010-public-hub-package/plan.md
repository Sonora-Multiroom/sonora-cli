# Implementation Plan: Public Hub Client Package

**Branch**: `010-public-hub-package` | **Date**: 2026-09-27 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `/specs/010-public-hub-package/spec.md`

## Summary

Make the Multiroom Audio Hub API client importable by other Sonora projects without
changing CLI behavior. Rename the module to `github.com/Sonora-Multiroom/sonora-cli`,
`git mv internal/hub hub`, move exit-code mapping (and the CLI-only `ClassUsage`) into a
new `internal/cli/exitcode` package, rewrite `hub` godoc to be self-contained, publish
`api/openapi.json` as `api.Spec` via `go:embed`, fix all four `-ldflags -X` version sites
(including `.goreleaser.yaml`), add guard tests for the version injection and the godoc
rules, then tag `v0.1.0`.

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

**Scale/Scope**: 111 `.go` files change import path; 77 import the client; 34 CLI files
call `ExitCode()`; ~137 `hub.ClassUsage` references; 57 godoc lines to rewrite across
10 `hub` files; 4 build configs

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

Constitution v1.2.0.

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Instant Startup | Pass | No startup-path work added; package move only. `api` not linked into `sonora`. |
| II. API Contract Fidelity | Pass | `hub/` becomes the single spec→Go place for all consumers; no type or endpoint changes; contract tests kept. |
| III. Minimal Dependencies | Pass | No new modules; `embed` is stdlib. |
| IV. Resilient HTTP | Pass | Timeouts, no-retry, and classification unchanged. |
| V. CLI UX Consistency | Pass | Commands, output, exit codes unchanged; exit-code table preserved in `exitcode.For` and re-asserted by moved tests. |
| VI. Test-First | Pass | New behavior = guards: version-injection test, godoc-rules test, `api.Spec` test, `exitcode` tests are written first and fail before the change (see research §2, §4, §6, §7). Existing tests are the refactor's safety net. |
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
build.sh                       # -X …/internal/version.Version
.goreleaser.yaml               # -X …/internal/version.Version (release binaries)

api/
├── openapi.json               # unchanged, single copy
└── spec.go                    # NEW: package api; //go:embed openapi.json; var Spec []byte

hub/                           # MOVED from internal/hub (git mv), package hub
├── client.go  errors.go  extensions.go  groups.go  inputs.go
├── mastermute.go  outputs.go  play.go  routes.go  tts.go
                               # errors.go: ExitCode() and ClassUsage removed; godoc rewritten

cmd/sonora/
├── main.go
└── main_test.go               # + version-injection config test, + build-and-run test (skipped in -short)

internal/
├── cli/
│   ├── exitcode/              # NEW: Usage const, For(hub.ErrorClass) int
│   └── groups/ inputs/ mastermute/ outputs/ play/ route/ routes/ tts/ clihelp/ respath/
├── config/  render/  version/ # unchanged apart from import paths

tests/
├── contract/                  # import path → …/sonora-cli/hub
├── integration/               # import path
└── unit/
    ├── hub_client_test.go     # ExitCode assertions removed (moved)
    ├── exitcode_test.go       # NEW: mapping table, uniqueness, 7 retired, Usage == 2
    ├── hub_godoc_test.go      # NEW: every exported ident documented; no spec-artifact refs
    └── api_spec_test.go       # NEW: api.Spec == api/openapi.json bytes, valid JSON
```

**Structure Decision**: Single Go module. Two public packages at the root (`hub`, `api`);
everything CLI-specific stays under `internal/`. Existing test layout (`tests/unit`,
`tests/integration`, `tests/contract` as external packages) is kept, which also exercises
`hub` purely through its exported API.

## Implementation Order

1. **Guards first (red)**: `exitcode_test.go`, `api_spec_test.go`, `hub_godoc_test.go`,
   version-injection tests in `cmd/sonora/main_test.go` — they fail to compile or fail
   against the current tree.
2. **Module rename** (`go.mod` + 111 files + 4 build configs) → build, full tests green.
3. **`git mv internal/hub hub`** + 77 import updates → green. Commit separately so history
   shows a pure rename.
4. **`internal/cli/exitcode`**; replace `ExitCode()`/`ClassUsage` call sites; delete them
   from `hub`; move old exit-code tests → green, `exitcode_test.go` green.
5. **Godoc rewrite** in `hub` + package comment → `hub_godoc_test.go` green.
6. **`api/spec.go`** → `api_spec_test.go` green.
7. `git grep` sweep for the old path (quickstart §2); README/CONTRIBUTING currently have
   no references.
8. Quickstart §1–5 on Windows and Linux; merge; `./release.sh --minor` → `v0.1.0`;
   quickstart §5 (post-tag) and §6.

Commits follow Conventional Commits with this repo's types (CONTRIBUTING.md): `refactor!`
for the module rename and exit-code move, `refactor` for the `git mv` and godoc, `feat` for
`api.Spec`, `spec` for release notes. Guard tests ship in the commit they guard.

## Complexity Tracking

No constitution violations.
