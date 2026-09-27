---

description: "Task list for 010-public-hub-package"
---

# Tasks: Public Hub Client Package

**Input**: Design documents from `/specs/010-public-hub-package/`

**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/public-packages.md,
quickstart.md

**Tests**: Included and REQUIRED — Constitution Principle VI (Test-First, NON-NEGOTIABLE).
This is a refactor, so the existing suites are the safety net; the new tests are guards
(version injection, godoc/dependency rules, exit-code mapping, CLI-neutral messages, stray
old-path references, published spec) and each is written and seen failing before the change
it guards — except two regression guards that pin behavior which must not change: T004
(`--version` wiring; made to fail once on purpose) and T015's characterization test.

`build.sh` is the maintainer's local, untracked helper (excluded via `.git/info/exclude`):
it is updated locally (T009) but never committed and never read by tests.

**Organization**: Tasks are grouped by user story (spec.md) so each can be checked
independently.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependency on an incomplete task)
- **[Story]**: Which user story this task belongs to (US1–US4)

## Path Conventions

Single Go module `github.com/Sonora-Multiroom/sonora-cli` (after T005). Source under
`cmd/`, `hub/`, `api/`, `internal/`; tests under `tests/unit`, `tests/contract`,
`tests/integration` (external test packages — repo convention), plus
`cmd/sonora/main_test.go`. No `make` on the Windows dev machine: run `gofmt -l .`,
`go vet ./...`, `go test ./...` directly.

**Commit types** (CONTRIBUTING.md): `refactor` for moves/renames, `feat` for `api.Spec`,
`ci` for build/release configs, `spec` for `specs/**` only. Exception: the `Makefile` and
`.goreleaser.yaml` edits go into T012's `refactor!` commit because they must land
atomically with the `go.mod` rename (otherwise T003 is red in between).

---

## Phase 1: Setup

**Purpose**: Establish a green baseline and a reference binary for behavior comparison.

- [ ] T001 Run `gofmt -l .`, `go vet ./...`, `go test ./...` on branch `010-public-hub-package`
      and confirm all green before any change; note any pre-existing failures in this file
      under "Notes" instead of fixing them here
- [ ] T002 Build the pre-refactor reference binary with `sh build.sh` and copy `sonora.exe`
      to the session scratchpad as `sonora-before.exe` (used by T038); do not commit it

---

## Phase 2: Foundational — module path rename (blocks all stories)

**Purpose**: Give the module its public import path (research §1) and fix every
version-injection site (research §2). Nothing public can be imported until this lands.

### Guard tests (write first)

- [ ] T003 Add `TestBuildConfigsInjectVersionForModulePath` to `cmd/sonora/main_test.go`:
      read the `module` line from `../../go.mod`; for each of `../../Makefile` and
      `../../.goreleaser.yaml` (tracked files only — not the untracked `build.sh`), find every
      `-X <path>/internal/version.Version=` occurrence and assert there is at least one and
      every `<path>` equals the module path. Expect: `Makefile` has 2 occurrences,
      `.goreleaser.yaml` 1. (Passes on the current tree —
      it turns red in T006.)
- [ ] T004 Add `TestVersionFlagReportsInjectedVersion` to `cmd/sonora/main_test.go`: skip when
      `testing.Short()`; `go build -o <t.TempDir()>/sonora[.exe] -ldflags "-X
      <module>/internal/version.Version=test-sentinel-010" .` (module path read from
      `go.mod` as in T003), run the binary with `--version`, assert stdout is exactly
      `test-sentinel-010\n`. This is a regression guard (it passes before the rename too):
      see it fail once by temporarily injecting `-X wrong/internal/version.Version=…` and
      confirming the assertion reports `dev`, then restore

### Rename

- [ ] T005 Change `go.mod` line 1 from `module sonora-cli` to
      `module github.com/Sonora-Multiroom/sonora-cli`
- [ ] T006 Run `go test ./cmd/sonora -run TestBuildConfigsInjectVersionForModulePath` and
      confirm it FAILS for both config files (red)
- [ ] T007 Rewrite every Go import of `"sonora-cli/` to
      `"github.com/Sonora-Multiroom/sonora-cli/` across all `.go` files (111 files:
      `cmd/`, `internal/`, `tests/`), e.g.
      `git grep -l '"sonora-cli/' -- '*.go' | xargs sed -i 's#"sonora-cli/#"github.com/Sonora-Multiroom/sonora-cli/#g'`,
      then `gofmt -w` the changed files (import grouping may reorder)
- [ ] T008 [P] Update both `-X sonora-cli/internal/version.Version=` occurrences in
      `Makefile` (targets `build` and `docker-build`) to
      `-X github.com/Sonora-Multiroom/sonora-cli/internal/version.Version=`
- [ ] T009 [P] Same replacement in the local, untracked `build.sh` (not committed; keeps
      local builds such as T035 reporting the right version)
- [ ] T010 [P] Same replacement in `.goreleaser.yaml` (`builds[0].ldflags`) — this is the
      file that builds published release binaries
- [ ] T011 Run `gofmt -l .`, `go vet ./...`, `go test ./...` (without `-short`, so T004
      runs) — all green; `git grep -n 'sonora-cli/' -- ':!specs' ':!docs/reviews'
      ':!.specify'` returns only `github.com/Sonora-Multiroom/sonora-cli/...` matches
- [ ] T012 Commit: `refactor!: rename module to github.com/Sonora-Multiroom/sonora-cli`
      (T003–T011 except the untracked `build.sh`; body notes the three `-ldflags` sites and the
      new guard tests)

**Checkpoint**: module builds and tests green under its public path.

---

## Phase 3: User Story 1 — Import the hub client from another project (Priority: P1) 🎯 MVP

**Goal**: `github.com/Sonora-Multiroom/sonora-cli/hub` is importable, protocol-only,
stdlib-only, and self-documented (contracts/public-packages.md).

**Independent Test**: quickstart §4 and §5 (pre-release, `go.work` scratch consumer).

### Guard tests (write first)

- [ ] T013 [P] [US1] Create a compile-only stub `internal/cli/exitcode/exitcode.go`
      (`package exitcode`; `const Usage = 0`; `func For(c hub.ErrorClass) int { return -1 }`,
      importing the client at its current path `.../internal/hub` — T016's import rewrite
      updates it) so the test below compiles and the rest of `tests/unit` keeps running.
      Then create `tests/unit/exitcode_test.go` (package `unit`) importing
      `github.com/Sonora-Multiroom/sonora-cli/internal/cli/exitcode` and the hub package:
      - `TestExitcodeUsage`: `exitcode.Usage == 2`
      - `TestExitcodeFor_Table`: `exitcode.For` returns ClassNone 0, ClassHub 3,
        ClassNetwork 4, ClassNotFound 5, ClassValidation 6, ClassRouteFailed 8,
        ClassSourceUnreachable 9, ClassServiceUnavailable 10, ClassInputNotFound 11,
        ClassTargetNotFound 12, ClassTTSUnavailable 13 (data-model.md)
      - `TestExitcodeFor_Distinct`: codes of all non-None classes plus `exitcode.Usage` are
        pairwise distinct and none equals 7 (retired)
      Red: compiles, fails on assertions against the stub. `go test ./tests/unit` must
      still compile and run every other unit test.
- [ ] T014 [P] [US1] Create `tests/unit/hub_godoc_test.go` that parses every non-test
      `.go` file in `../../hub` with `go/parser` (`parser.ParseComments`) and asserts:
      - the package has a package doc comment;
      - every exported top-level func, type, method on an exported type, and exported
        const/var (a doc on the enclosing `const (...)`/`var (...)` group counts for its
        members) has a non-empty doc comment — except `Error`, `Unwrap` and `String`
        methods, which implement standard interfaces and follow Go convention of no doc;
      - no comment in the package matches
        `research\.md|data-model\.md|spec\.md|\bFR-\d|\bSC-\d|§|[Pp]rinciple [IVX]+|[Cc]onstitution|\b0\d\d-[a-z]`;
      - every import path is standard library (first path element contains no `.`).
      Report each violation with `file:line`. Red: `../../hub` does not exist yet (the test
      fails at runtime, not at compile time, so the package still builds); after T016 it
      fails on the godoc violations until T029.
- [ ] T015 [P] [US1] CLI-neutral message guards (FR-015, research §5):
      - In `tests/unit/tts_report_test.go` add `TestReportError_HubAddress_ExactStderr`:
        `ReportError` with `&hub.TTSUnavailableError{Diagnosis: hub.DiagnosisHubAddress,
        BaseURL: u}` writes exactly `error: <u> is not serving the Multiroom Audio Hub API:
        the hub URL is wrong, or the hub's control API (REST) extension is not installed or
        not loaded; set the correct address with --hub-url, MULTIROOM_URL, or the config
        file\n` and returns exit code 4. This is a characterization test: it passes now and
        must still pass after T024. Existing tests already pin the version-mismatch CLI text.
      - In `tests/unit/hub_client_test.go` add `TestClassifyError_MessagesAreCLIAgnostic`:
        run `hub.ClassifyError` over one error of every kind — `StatusError` 400/404/422/
        500/502/503 as the existing tests build them, `APIError`, `DecodeError`,
        `NotFoundError`, a `TTSError` per extension error code, `TTSNotOfferedError`,
        `TTSUnavailableError` for every `TTSDiagnosis` (with `BaseURL` set), a network
        error and `context.DeadlineExceeded`. Fail with the class and message when a
        message matches `--|MULTIROOM_URL|config file|exit code|\bCLI\b`.
      - In `TestClassifyError_TTSUnavailableError_ByDiagnosis`, change the expected text for
        `DiagnosisHubAddress` (drop the "; set the correct address …" suffix) and
        `DiagnosisVersionMismatch` ("this client version").
      Red: the new hub test fails on those two diagnoses, and `ByDiagnosis` fails on the two
      changed expectations.

### Implementation

- [ ] T016 [US1] `git mv internal/hub hub`; rewrite all imports of
      `"github.com/Sonora-Multiroom/sonora-cli/internal/hub"` to
      `"github.com/Sonora-Multiroom/sonora-cli/hub"` (77 files incl. `tests/contract`,
      `tests/integration`, `tests/unit`, `internal/cli/**`, plus the T013 stub = 78 files);
      `go build ./... && go vet ./... && go test ./...` — every test compiles; the only
      failing tests allowed are T013's `TestExitcode*`, T014's `TestHubGodoc*`, T015's two
      hub message tests, and T039's test if US3 is in progress (red by design)
- [ ] T017 [US1] Commit the move on its own: `refactor: move internal/hub to public hub
      package` (no content edits besides import lines, so `git log --follow hub/` shows
      history)
- [ ] T018 [US1] Replace the T013 stub in `internal/cli/exitcode/exitcode.go`: package doc ("maps hub error
      classes to the process exit codes of the `sonora` command"); `const Usage = 2`; `func For(c
      hub.ErrorClass) int` with the switch from `hub/errors.go` `ExitCode()` minus the
      `ClassUsage` case; T013 turns green; keep the explanatory comment about retired code 7 (internal code
      may cite spec artifacts)
- [ ] T019 [US1] In `internal/cli/**`, replace every `hub.ClassUsage.ExitCode()` with
      `exitcode.Usage` and add the `exitcode` import (sed across the 128 calls, then
      `goimports`-style fix by hand or `gofmt`); also replace any other `hub.ClassUsage`
      use (e.g. passed to a helper that later calls `.ExitCode()`) with the equivalent
      `exitcode.Usage` flow — `git grep -n 'ClassUsage' internal/` must be empty afterward
- [ ] T020 [US1] In `internal/cli/**`, replace every remaining `<expr>.ExitCode()` on a
      `hub.ErrorClass` value with `exitcode.For(<expr>)` (e.g. `class.ExitCode()` →
      `exitcode.For(class)`, `hub.ClassNotFound.ExitCode()` → `exitcode.For(hub.ClassNotFound)`);
      `git grep -n '\.ExitCode()' internal/` must be empty afterward (30 files, 157 calls in
      total across T019 and this task)
- [ ] T021 [US1] In `tests/unit/hub_client_test.go`, delete `TestErrorClass_ExitCodes`,
      `TestErrorClass_NewExitCodes`, `TestErrorClass_AllExitCodesDistinct`,
      `TestErrorClass_RouteExitCodes` (and any other test calling `.ExitCode()` on a class —
      15 call sites today; now covered by T013), and in `TestClassifyError_NotFound` drop
      the `ClassUsage`/exit-code-uniqueness block at its end. Replace `hub.ClassUsage` in any
      other file under `tests/` with the matching `exitcode.Usage` assertion. Do NOT touch
      `exec.ExitError.ExitCode()` calls in `tests/integration/outputs_list_test.go` and
      `tests/integration/tts_test.go`
- [ ] T022 [US1] In `hub/errors.go`, delete the `ExitCode()` method and the `ClassUsage`
      constant (research §4); run `go build ./... && go vet ./... && go test ./...` — all
      green except T014 (godoc), which stays red until T029, and T015's hub message tests,
      which stay red until T024
- [ ] T023 [US1] Commit: `refactor!: move exit-code mapping out of hub into internal/cli/exitcode`
      (body: `hub.ClassUsage` and `ErrorClass.ExitCode()` removed; later `ErrorClass`
      integer values shift; no exit code changes)
- [ ] T024 [US1] Make hub TTS messages CLI-neutral (FR-015, research §5):
      - In `hub/tts.go` `TTSUnavailableError.Error()`, end the `DiagnosisHubAddress`
        message after "…is not installed or not loaded", and change `DiagnosisVersionMismatch`
        to "…does not match this client version".
      - In `internal/cli/tts/report.go` `ReportError`, after `hub.ClassifyError`: for
        `DiagnosisHubAddress` append `"; set the correct address with --hub-url,
        MULTIROOM_URL, or the config file"`; for `DiagnosisVersionMismatch` use the CLI text
        `"text-to-speech is not available on this hub: the hub's TTS API does not match this
        CLI version"`. Build these from constants, not `strings.Replace`.
      - `git grep -n 'TTSUnavailableError\|ClassifyError' -- internal/cli` to confirm no
        other CLI path prints these messages.

      T015 goes green. `tests/unit/tts_report_test.go` and `tests/unit/cli_tts_voices_test.go`
      pass without edits. `go test ./...` is green except T014.
- [ ] T025 [US1] Commit: `refactor: keep CLI hints out of hub TTS error messages` (body:
      `hub` messages are CLI-neutral; `internal/cli/tts` adds the hints back, so CLI
      output is unchanged)
- [ ] T026 [P] [US1] Create `hub/doc.go` with the package comment: what the package is
      (Multiroom Audio Hub API client for `api/openapi.json`), the call shape
      (`NewClient()` or `NewClientWithTimeout`, then `Op(ctx, client, baseURL, …)`), one
      HTTP request per call with no retries, typed errors and `ClassifyError` for a
      class plus one-line user-facing message, standard library only. Touch only
      `hub/doc.go`; T029 removes the old package comment from `hub/client.go`
- [ ] T027 [P] [US1] Rewrite godoc in `hub/tts.go` (20 flagged lines): replace citations
      of `research.md`, `data-model.md`, `FR-…`, `§…`, `009-tts-commands`, constitution
      principles with a direct statement of the behavior; keep `#/components/schemas/...`
      and `api/openapi.json` references
- [ ] T028 [P] [US1] Same godoc rewrite in `hub/routes.go` (9), `hub/play.go` (4),
      `hub/extensions.go` (4)
- [ ] T029 [P] [US1] Same godoc rewrite in `hub/errors.go` (8), `hub/client.go` (4),
      `hub/groups.go` (3), `hub/inputs.go` (2), `hub/outputs.go` (2), `hub/mastermute.go`
      (1); remove the one-line package comment from `hub/client.go` (T026's `hub/doc.go`
      replaces it); add doc comments to any exported identifier T014 reports as undocumented
      (e.g. the `Class*` const group, `DiagnosisUnknown` group; `Error`/`Unwrap`/`String`
      methods are exempt); then `go test ./tests/unit -run HubGodoc` green and `go test ./...` green
- [ ] T030 [US1] Review `go doc -all ./hub` end to end: it must read as a self-contained
      client API (quickstart §4); fix wording that assumes CLI context (e.g. "command",
      "--verbose", "exit") in exported docs
- [ ] T031 [US1] Commit: `refactor: make hub godoc self-contained for external consumers`
- [ ] T032 [US1] External-consumer check, pre-release (quickstart §5): in the session
      scratchpad create module `example.com/hubconsumer` with `go.work` pointing at it and
      this repo; `main.go` uses `httptest.Server` to serve `GET /api/v2/outputs` and
      `PUT /api/v2/master-mute` (check paths in `hub/outputs.go`, `hub/mastermute.go`),
      calls `hub.ListOutputs`, `hub.SetMasterMute`, then a call to a closed port and
      prints `hub.ClassifyError(err)` (expect `ClassNetwork`, "could not reach the hub");
      confirm importing `.../internal/config` fails with "use of internal package … not
      allowed". Scratch files are not committed

**Checkpoint**: `hub` is public, protocol-only, documented; US1 verified via workspace.

---

## Phase 4: User Story 2 — Existing CLI users notice no change (Priority: P1)

**Goal**: Identical commands, output, exit codes; correct `--version` in every build path.

**Independent Test**: full suites on Windows and Linux; quickstart §2, §3, §7.

(Guard tests T003/T004 and T013 already cover version injection and the exit-code table.)

- [ ] T033 [US2] Create `tests/unit/stray_refs_test.go` (`TestStrayRefs`, research §9).
      Walk `../..`, skipping a hard-coded list of directories: `.git`, `specs`,
      `docs/reviews`, `.specify`, `.idea`, `.claude`, `trash`. Scan every `.go` file, plus
      `Makefile`, `release.sh`, `.goreleaser.yaml`, `scripts/*`, `.github/workflows/*`, `README.md`,
      `CONTRIBUTING.md` and `AGENTS.md`. Fail with `file:line` on `"sonora-cli/`,
      `X sonora-cli/` or `internal/hub`. Build the patterns by concatenation (e.g.
      `"internal/" + "hub"`) so the test does not match itself.
      Red: expected hits are non-import comments such as
      `internal/cli/groups/volume.go:24`, which names `internal/hub/errors.go`.
- [ ] T034 [US2] Fix every hit T033 reports (code, tests, scripts, README, CONTRIBUTING)
      until `go test ./tests/unit -run StrayRefs` is green. Cross-check with the
      quickstart §2 greps (`git grep -n '"sonora-cli/' -- ':!specs' ':!docs/reviews'`,
      `git grep -n 'internal/hub' -- ':!specs' ':!docs/reviews' ':!.specify'`,
      `git grep -n 'X sonora-cli/' -- ':!specs'`) — all empty. Commit T033 and T034
      together: `refactor: drop leftover internal/hub references and guard against them`
- [ ] T035 [US2] Quickstart §3: `sh build.sh && ./sonora.exe --version` prints
      `git describe --tags --always --dirty` output, not `dev`
- [ ] T036 [US2] Linux check: `go test ./...` on Linux — push the branch and confirm the
      `test.yml` workflow is green, or run `docker run --rm -v "$PWD":/app -w /app
      golang:1.27-alpine go test ./...` locally
- [ ] T037 [US2] GoReleaser config check: if `goreleaser` is available run
      `goreleaser build --snapshot --clean --single-target` and confirm the built binary's
      `--version` is not `dev`; otherwise rely on T003 and note it here
- [ ] T038 [US2] Behavior spot check (quickstart §7) against a reachable hub if one is
      available: compare `sonora-before.exe` (T002) and the new `sonora.exe` for
      `--help`, `--version` (only the version string may differ), `get outputs`, `get outputs --json`, `get routes`, `get outputs/no-such-id`,
      `set outputs/no-such-id volume 10`, and a usage error (`sonora get`) — identical
      stdout/stderr and exit codes. If no hub is reachable, run the usage-error and
      unreachable-hub cases (`--hub-url` / config pointing at a closed port) only and record
      that here

**Checkpoint**: no observable CLI change on Windows and Linux.

---

## Phase 5: User Story 3 — Consumers verify against the same hub spec (Priority: P2)

**Goal**: `github.com/Sonora-Multiroom/sonora-cli/api.Spec` exposes `api/openapi.json`
with no second copy (research §7).

**Independent Test**: T039 green; quickstart §5 consumer prints `len(api.Spec) > 0`.

- [ ] T039 [P] [US3] Create a compile-only stub `api/spec.go` (`package api`;
      `var Spec []byte`, no embed yet), then `tests/unit/api_spec_test.go`: `api.Spec` is
      byte-identical to `os.ReadFile("../../api/openapi.json")`, non-empty, and
      `json.Unmarshal` into `map[string]any` yields a string `openapi` field. Red: compiles,
      fails on the empty `Spec`
- [ ] T040 [US3] Complete `api/spec.go`: package doc, `import _ "embed"`,
      `//go:embed openapi.json` above `var Spec []byte`, godoc stating it is the exact
      Multiroom Audio Hub API spec the `hub` package at the same module version is
      tested against and that callers must not modify it; T039 green
- [ ] T041 [US3] Confirm the CLI binary does not link `api`:
      `go list -deps ./cmd/sonora | grep -x 'github.com/Sonora-Multiroom/sonora-cli/api'`
      prints nothing
- [ ] T042 [US3] Extend the T032 scratch consumer to print `len(api.Spec) > 0` (true)
- [ ] T043 [US3] Commit: `feat: publish hub OpenAPI spec as api.Spec`

**Checkpoint**: consumers can import the spec at the client's version.

---

## Phase 6: User Story 4 — Maintainers release the shared client safely (Priority: P2)

**Goal**: A `v0.1.0` tag consumers can resolve through the public module proxy.

**Independent Test**: quickstart §5 (post-tag, `GOWORK=off`) and §6.

- [ ] T044 [US4] Write release notes in `specs/010-public-hub-package/release-notes.md`
      for `v0.1.0`: module path is now `github.com/Sonora-Multiroom/sonora-cli`; new public
      packages `hub` and `api` (link contracts/public-packages.md); not in the public API:
      `ErrorClass.ExitCode()`, `ClassUsage`; `TTSUnavailableError` messages are
      CLI-neutral (hub-address hint dropped, "this client version"); `ErrorClass` integer values are not stable —
      compare against constants; requires Go ≥ 1.27; no CLI behavior change; `go install
      github.com/Sonora-Multiroom/sonora-cli/cmd/sonora@v0.1.0` now works. Commit with
      `spec:` type
- [ ] T045 [US4] Open the PR for `010-public-hub-package` and merge after review and green
      CI (outward-facing — confirm with the maintainer before pushing/merging). The PR
      touches `go.mod` and moves HTTP client construction (`hub.NewClient`), so its
      description MUST include the constitution's Development Workflow self-review against
      Principles I, III, IV and VI (adapt plan.md's Constitution Check rows: no startup
      work added, no new dependencies, timeouts/no-retry unchanged, guard tests written
      first). Do not add AI attribution to the PR description
- [ ] T046 [US4] On `main` after merge, run `./release.sh --minor` (expects `v0.0.17` →
      `v0.1.0`; `--minor` skips the version prompt, but the script still asks
      "Continue? [y/N]" before tagging and pushing — answer only after maintainer approval); paste T044's notes into the GitHub release body. Confirm with the
      maintainer before pushing the tag. Never move, delete or re-create a pushed tag —
      fix forward with `v0.1.1`
- [ ] T047 [US4] Post-tag consumer check (quickstart §5): in the scratch consumer, `rm
      go.work`, `GOWORK=off go get github.com/Sonora-Multiroom/sonora-cli@v0.1.0`,
      `GOWORK=off go run .` — same output as T032/T042, no `replace` directives

**Checkpoint**: sonora-mcp can `go get` the tag (unblocks its Phase 3).

---

## Phase 7: Polish & Cross-Cutting

- [ ] T048 [P] Add a short "Using the hub client from Go" section to `README.md`
      (`go get …@v0.1.0`, import `…/hub`, 5-line example, link to `go doc`); commit `docs:`
- [ ] T049 Final gate: `gofmt -l .` (no output), `go vet ./...`, `go test ./...` green;
      mark all tasks above `[X]`
- [ ] T050 [P] Tick the Phase 2 checklist items in
      `D:\projects-sonora\sonora-mcp\docs\future\go-rewrite-shared-hub-client.md`
      (sonora-mcp repo; separate commit there, confirm with the maintainer)

---

## Dependencies & Execution Order

- **Phase 1** → **Phase 2 (T003–T012)** → everything else. The module path is in every
  import line, so no story can start before T012.
- **US1 (Phase 3)**: T013/T014/T015 [P] first (red on assertions, never on compilation, so
  the unit suite keeps guarding the move) → T016 → T017 → T018 → T019 → T020 → T021
  → T022 → T023 → T024 → T025 → T026–T029 [P] → T030 → T031 → T032.
- **US3 (Phase 5)**: depends only on Phase 2; can run in parallel with US1 (touches
  `api/` and a new test file only). T042 needs T032's scratch consumer.
- **US2 (Phase 4)**: verification — run after US1 and US3 are committed (it validates the
  final tree). T033 before T034.
- **US4 (Phase 6)**: after US1–US3 and US2 verification; T046 after merge; T047 after tag.
- **Polish**: T048 any time after T044; T049 last before PR; T050 after T047.

## Parallel Opportunities

- Phase 2: T008, T009, T010 (three different config files).
- US1: T013 ∥ T014 ∥ T015 (disjoint test files); T026 ∥ T027 ∥ T028 ∥ T029 (disjoint
  `hub/*.go` files — only T029 touches `hub/client.go`).
- US1 ∥ US3: T039–T041 can proceed while T016–T031 are in progress.

```text
# After T012:
Agent A: T013, T014, T015 → T016 … T023 → T024 → T025  (hub move, exitcode, messages)
Agent B: T039 → T040 → T041                 (api.Spec)
# After T025:
Agents A–D: T026 | T027 | T028 | T029         (godoc, one file group each)
```

## Implementation Strategy

- **MVP**: Phase 1 + Phase 2 + US1 (T001–T032). At that point `hub` is importable via a
  workspace and the CLI is green — enough to start sonora-mcp prototyping with `go.work`.
- **Increment 2**: US3 (`api.Spec`) — small, independent.
- **Increment 3**: US2 verification on both OSes, then US4 release `v0.1.0`.
- Keep commits small and in the order above so the `git mv` commit (T017) is a pure
  rename and each `refactor!` commit is reviewable on its own.

## Notes

- `[P]` = different files, no dependency on an incomplete task.
- Internal code (e.g. `internal/cli/exitcode`) may keep references to spec artifacts;
  only `hub/` public docs are restricted (constitution VII).
- Historical records (`specs/001–009/**`, `docs/reviews/**`, the constitution's Sync
  Impact Report) are not rewritten.
