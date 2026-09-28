# Research: Public Hub Client Package

**Feature**: 010-public-hub-package | **Date**: 2026-09-27

All findings were checked against the repository on 2026-09-27 (branch
`010-public-hub-package`, base `a194faa`).

## §1 Module path

- **Decision**: `module github.com/Sonora-Multiroom/sonora-cli` in `go.mod`, matching the
  `origin` remote (`git@github.com:Sonora-Multiroom/sonora-cli.git`). Rewrite every
  `"sonora-cli/…"` import (111 `.go` files) with a mechanical search/replace, then
  `gofmt -l .` and `go build ./...`.
- **Rationale**: Go resolves a module by its path; only the repo URL form lets the module
  proxy and `go get` find it without `replace` directives (SC-001). It also enables
  `go install github.com/Sonora-Multiroom/sonora-cli/cmd/sonora@<tag>` as a side benefit.
- **Alternatives considered**: a vanity path (e.g. `sonora.dev/cli`) — needs a hosted
  `go-import` meta page, which nothing else requires; rejected.

## §2 Version injection sites

- **Decision**: Update all three tracked places that inject the version with
  `-X sonora-cli/internal/version.Version=…`:
  1. `Makefile` target `build`
  2. `Makefile` target `docker-build`
  3. `.goreleaser.yaml` `builds[0].ldflags` (**not listed in the source plan** — this is the
     one that produces published release binaries)

  The maintainer's `build.sh` has the same `-X` flag but is a local, untracked helper
  (excluded via `.git/info/exclude`): it is updated locally, never committed, and not
  checked by any test, since it does not exist in CI checkouts.

  `release.sh`, `scripts/update-openapi.sh` and `.github/workflows/*.yml` do not contain the
  module path (checked).
- **Guard (FR-011)**: a unit test in `cmd/sonora` that reads the module path from `go.mod`
  and asserts each of the tracked files above contains `-X <module>/internal/version.Version=`,
  and contains no `-X` with any other module prefix. It runs in milliseconds, needs no
  toolchain invocation, and fails the next time the module path and a build config drift.
  A second test builds the binary with `-ldflags -X <module>/internal/version.Version=
  test-sentinel` and asserts `sonora --version` prints `test-sentinel`; it is skipped under
  `go test -short` because it shells out to `go build`.
- **Rationale**: `-X` with an unknown symbol path is silently ignored by the linker, so the
  build succeeds and only the version string is wrong — nothing else would catch it.
- **Alternatives considered**: a `release.sh` pre-flight grep — only runs for maintainers
  who release locally, not in CI or GoReleaser; rejected as the only guard.

## §3 Moving the package

- **Decision**: `git mv internal/hub hub`; package name stays `hub`. Update the 77 files
  importing `…/internal/hub` (CLI packages, `tests/contract`, `tests/integration`,
  `tests/unit`; `cmd/sonora` does not import it).
- **Rationale**: `git mv` keeps `git log --follow` history (FR-002). The package name is
  unchanged, so call sites only change their import line.
- **Alternatives considered**: `hub/` under `pkg/` — adds a path segment with no meaning in
  Go; rejected. A separate module in the same repo (`hub/go.mod`) — two tag namespaces
  (`hub/v0.x.y`) and a `replace` in the CLI; not needed while the CLI and client release
  together (option 2a).

## §4 Exit codes and `ClassUsage` out of `hub`

- **Decision**: New package `internal/cli/exitcode`:
  - `const Usage = 2` for CLI usage errors.
  - `func For(c hub.ErrorClass) int` with today's mapping for every other class (3, 4, 5,
    6, 8, 9, 10, 11, 12, 13; 0 for `ClassNone`; 7 stays retired).
  - Delete `ErrorClass.ExitCode()` **and** `hub.ClassUsage` from `hub`.
  - Call sites: `hub.ClassUsage.ExitCode()` → `exitcode.Usage`; `class.ExitCode()` →
    `exitcode.For(class)` (30 CLI files, 157 calls, 128 of them
    `hub.ClassUsage.ExitCode()`; the source plan said 34 files).
  - Existing exit-code tests in `tests/unit/hub_client_test.go` (mapping table, "no code
    reused", "7 retired") move to `tests/unit/exitcode_test.go` and assert the same numbers
    against `exitcode.For`/`exitcode.Usage`.
- **Rationale**: `ClassUsage` is never returned by `ClassifyError`; it exists only so CLI
  argument-parsing code can reach exit code 2 through the class type. It is a CLI concern
  (constitution VII) and would be a confusing, never-produced value in sonora-mcp. The
  first public release is the cheapest moment to drop it; after `v0.1.0` it becomes a
  release-note-worthy break.
- **Numeric values**: `ErrorClass` is an `iota` enum; removing `ClassUsage` shifts later
  constants' integer values. No code or output depends on the integer (exit codes come
  from the explicit switch, and classes are never serialized), so this is safe. The shift
  is recorded in the release notes because it is the public type's first published shape.
- **Alternatives considered**: keep `hub.ClassUsage` (plan's minimal option) — leaves a CLI
  concept in the public API permanently; rejected. `exitcode.For(hub.ClassUsage)` with a
  CLI-local class type — more indirection for one constant; rejected.

## §5 What stays in `hub`

- **Decision**: `ErrorClass` (minus `ClassUsage`), `ClassifyError`, all error types
  (`StatusError`, `DecodeError`, `NotFoundError`, `APIError`, `TTSError`,
  `TTSUnavailableError`, `TTSNotOfferedError`, `TTSDiagnosis`), `SingleLine`,
  `NewClient`, `NewClientWithTimeout`, `SpeakTimeout`, and every operation function and
  type. No signature changes.
- **Rationale**: sonora-mcp needs `ClassifyError` messages for `isError` results.
  `SingleLine` is used by `hub/tts.go` itself and by `internal/cli/tts/report.go`.
  Keeping signatures unchanged keeps this a pure move (FR-010) and keeps the diff
  reviewable.
- **Exception found in analysis — CLI wording in TTS messages (FR-015)**:
  `TTSUnavailableError.Error()`, which `ClassifyError` returns as its message, has two
  CLI-specific strings:
  - `DiagnosisHubAddress`: ends with "; set the correct address with --hub-url,
    MULTIROOM_URL, or the config file".
  - `DiagnosisVersionMismatch`: "the hub's TTS API does not match this CLI version".

  **Decision**: in `hub`, drop the hint suffix from the hub-address message and say "this
  client version" in the version-mismatch message. In `internal/cli/tts.ReportError`, the
  CLI builds its own text for these two diagnoses: it appends the same hint suffix and
  uses "this CLI version", so stderr stays byte-identical (FR-010). The hub-level
  expectations in `tests/unit/hub_client_test.go` (`TestClassifyError_TTSUnavailableError_ByDiagnosis`)
  change to the neutral text. The CLI-level tests (`tests/unit/tts_report_test.go`,
  `tests/unit/cli_tts_voices_test.go`) stay unchanged and prove the CLI output did not
  change. Today only the hub-level test pins the full hub-address text, so a CLI-level
  characterization test in `tts_report_test.go` must first assert the exact current
  `ReportError` stderr for `DiagnosisHubAddress`. It passes before the change and must
  still pass after.
  **Guard**: a unit test runs `ClassifyError` over one error of every kind (each typed
  error, each `TTSDiagnosis`, network, timeout, decode) and fails if a message contains
  `--`, `MULTIROOM_URL`, `config file`, `exit code` or the word `CLI`.
  **Alternative considered**: leave the hints and document them. Rejected — sonora-mcp
  would show an AI assistant flags it cannot use.
- **Alternatives considered**: reshaping the API into a `Client` struct with methods
  (`c.ListOutputs(ctx)`) instead of `(ctx, *http.Client, baseURL, …)` functions — nicer for
  consumers, but a large behavior-neutral churn across every CLI call site. Out of scope;
  possible in a later `v0.x` with release notes.

## §6 Godoc cleanup

- **Decision**: Rewrite the 57 comment lines in `internal/hub` that cite
  `research.md`, `data-model.md`, `FR-…`, `§…`, feature numbers (`009-tts-commands`) or
  "constitution Principle …" so they state the behavior directly. References to
  `api/openapi.json` schemas (e.g. `#/components/schemas/OutputResponse`) stay — they name
  the public spec, which ships in the same module (§7). Add a package comment that
  describes the call shape (`NewClient` + base URL + context), error classification, and
  the no-retry/timeout behavior.
- **Guard**: a test (in `tests/unit`) that parses `hub/*.go` comments with `go/parser` and
  fails on the patterns above, plus asserts every exported identifier has a doc comment
  (SC-005), exempting `Error`/`Unwrap`/`String` methods. Runs in milliseconds.
- **Distribution**: client.go 4, errors.go 8, extensions.go 4, groups.go 3, inputs.go 2,
  mastermute.go 1, outputs.go 2, play.go 4, routes.go 9, tts.go 20.
- **Alternatives considered**: manual review only — regresses silently the next time a
  feature adds a `research.md §N` reference; rejected.

## §7 Publishing the spec (FR-012)

- **Decision**: `api/spec.go`, package `api`, with `//go:embed openapi.json` into an
  exported `var Spec []byte` and a godoc comment stating it is the exact spec the `hub`
  package at the same module version is tested against.
- **Rationale**: embed reads the existing file, so there is no second copy (US3 scenario 2);
  standard library only; no cost for the CLI binary since `cmd/sonora` never imports
  `api` (the linker drops unreferenced packages — the ~spec size is not added to `sonora`).
- **Test**: a unit test asserting `api.Spec` is byte-identical to `os.ReadFile("../../api/
  openapi.json")` (from `tests/unit`) and parses as JSON with an `openapi` field.
- **Alternatives considered**: `var Spec string` — `[]byte` is what `json.Unmarshal` and
  OpenAPI loaders take; chosen. A function `Spec() []byte` returning a copy — guards
  against mutation but is unusual for embedded data; the godoc says "MUST NOT be
  modified" instead.

## §8 Release

- **Decision**: After merge, run `./release.sh --minor` → `v0.1.0` (current latest tag is
  `v0.0.17`). Release notes list: new module path, public `hub` and `api` packages,
  `ClassUsage`/`ExitCode()` removal from the public surface, Go ≥ 1.27 required (from
  `go.mod`). Never move or re-create a pushed tag.
- **Rationale**: a minor bump signals the new public surface; `release.sh` already
  refuses to reuse an existing tag.
- **Consumer check**: after the tag is pushed, run quickstart §5 post-tag steps (scratch
  module, `rm go.work`, `GOWORK=off go get …@v0.1.0`, `GOWORK=off go run .`).

## §9 Out-of-repo references

- **Decision**: Update live references only: `Makefile`, `.goreleaser.yaml` (plus the
  local untracked `build.sh`),
  and README/CONTRIBUTING if they mention the old path (they currently don't). The
  constitution's Sync Impact Report mentions `internal/hub` as the pre-merge location;
  that is a dated record and stays as written.
  Leave `specs/001–009/**` and `docs/reviews/**` untouched — they are historical records
  of earlier features. `.idea/` and `sonora.exe` are untracked local files.
- **Guard**: `tests/unit/stray_refs_test.go` walks the live files — every `.go` file
  outside `specs/`, plus `Makefile`, `release.sh`, `.goreleaser.yaml`,
  `scripts/*`, `.github/workflows/*`, `README.md`, `CONTRIBUTING.md`, `AGENTS.md` — and
  fails on `"sonora-cli/`, `X sonora-cli/` or `internal/hub`, reporting `file:line`. It
  builds these patterns by string concatenation so it does not flag itself. It runs in CI
  with the rest of `go test ./...`, so a regression fails the build rather than depending
  on someone remembering the grep. Leftover non-import comments (e.g.
  `internal/cli/groups/volume.go:24` naming `internal/hub/errors.go`) are expected first
  hits.
- **Rationale**: SC-004 targets code, tests, build/release scripts, and current docs;
  rewriting history documents would misstate what those features did.
