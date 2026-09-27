---

description: "Task list for 010-public-hub-package"
---

# Tasks: Public Hub Client Package

**Input**: Design documents from `/specs/010-public-hub-package/`

**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/public-packages.md,
quickstart.md

**Tests**: Included and REQUIRED — Constitution Principle VI (Test-First, NON-NEGOTIABLE).
This is a refactor, so the existing suites are the safety net; the new tests are guards
(version injection, godoc/dependency rules, exit-code mapping, published spec) and each is
written and seen failing before the change it guards.

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
`ci` for build/release configs, `spec` for `specs/**` only.

---

## Phase 1: Setup

**Purpose**: Establish a green baseline and a reference binary for behavior comparison.

- [ ] T001 Run `gofmt -l .`, `go vet ./...`, `go test ./...` on branch `010-public-hub-package`
      and confirm all green before any change; note any pre-existing failures in this file
      under "Notes" instead of fixing them here
- [ ] T002 Build the pre-refactor reference binary with `sh build.sh` and copy `sonora.exe`
      to the session scratchpad as `sonora-before.exe` (used by T034); do not commit it

---

## Phase 2: Foundational — module path rename (blocks all stories)

**Purpose**: Give the module its public import path (research §1) and fix every
version-injection site (research §2). Nothing public can be imported until this lands.

### Guard tests (write first)

- [ ] T003 Add `TestBuildConfigsInjectVersionForModulePath` to `cmd/sonora/main_test.go`:
      read the `module` line from `../../go.mod`; for each of `../../Makefile`,
      `../../build.sh`, `../../.goreleaser.yaml`, find every `-X <path>/internal/version.Version=`
      occurrence and assert there is at least one and every `<path>` equals the module path.
      Expect: `Makefile` has 2 occurrences, the others 1. (Passes on the current tree —
      it turns red in T006.)
- [ ] T004 Add `TestVersionFlagReportsInjectedVersion` to `cmd/sonora/main_test.go`: skip when
      `testing.Short()`; `go build -o <t.TempDir()>/sonora[.exe] -ldflags "-X
      <module>/internal/version.Version=test-sentinel-010" .` (module path read from
      `go.mod` as in T003), run the binary with `--version`, assert stdout is exactly
      `test-sentinel-010\n`

### Rename

- [ ] T005 Change `go.mod` line 1 from `module sonora-cli` to
      `module github.com/Sonora-Multiroom/sonora-cli`
- [ ] T006 Run `go test ./cmd/sonora -run TestBuildConfigsInjectVersionForModulePath` and
      confirm it FAILS for all three config files (red)
- [ ] T007 Rewrite every Go import of `"sonora-cli/` to
      `"github.com/Sonora-Multiroom/sonora-cli/` across all `.go` files (111 files:
      `cmd/`, `internal/`, `tests/`), e.g.
      `git grep -l '"sonora-cli/' -- '*.go' | xargs sed -i 's#"sonora-cli/#"github.com/Sonora-Multiroom/sonora-cli/#g'`,
      then `gofmt -w` the changed files (import grouping may reorder)
- [ ] T008 [P] Update both `-X sonora-cli/internal/version.Version=` occurrences in
      `Makefile` (targets `build` and `docker-build`) to
      `-X github.com/Sonora-Multiroom/sonora-cli/internal/version.Version=`
- [ ] T009 [P] Same replacement in `build.sh`
- [ ] T010 [P] Same replacement in `.goreleaser.yaml` (`builds[0].ldflags`) — this is the
      file that builds published release binaries
- [ ] T011 Run `gofmt -l .`, `go vet ./...`, `go test ./...` (without `-short`, so T004
      runs) — all green; `git grep -n 'sonora-cli/' -- ':!specs' ':!docs/reviews'
      ':!.specify'` returns only `github.com/Sonora-Multiroom/sonora-cli/...` matches
- [ ] T012 Commit: `refactor!: rename module to github.com/Sonora-Multiroom/sonora-cli`
      (T003–T011; body notes the four `-ldflags` sites and the new guard tests)

**Checkpoint**: module builds and tests green under its public path.

---

## Phase 3: User Story 1 — Import the hub client from another project (Priority: P1) 🎯 MVP

**Goal**: `github.com/Sonora-Multiroom/sonora-cli/hub` is importable, protocol-only,
stdlib-only, and self-documented (contracts/public-packages.md).

**Independent Test**: quickstart §4 and §5 (pre-release, `go.work` scratch consumer).

### Guard tests (write first)

- [ ] T013 [P] [US1] Create `tests/unit/exitcode_test.go` (package `unit`) importing
      `github.com/Sonora-Multiroom/sonora-cli/internal/cli/exitcode` and `.../hub`:
      - `TestExitcodeUsage`: `exitcode.Usage == 2`
      - `TestExitcodeFor_Table`: `exitcode.For` returns ClassNone 0, ClassHub 3,
        ClassNetwork 4, ClassNotFound 5, ClassValidation 6, ClassRouteFailed 8,
        ClassSourceUnreachable 9, ClassServiceUnavailable 10, ClassInputNotFound 11,
        ClassTargetNotFound 12, ClassTTSUnavailable 13 (data-model.md)
      - `TestExitcodeFor_Distinct`: codes of all non-None classes plus `exitcode.Usage` are
        pairwise distinct and none equals 7 (retired)
      Red: does not compile (package missing).
- [ ] T014 [P] [US1] Create `tests/unit/hub_godoc_test.go` that parses every non-test
      `.go` file in `../../hub` with `go/parser` (`parser.ParseComments`) and asserts:
      - the package has a package doc comment;
      - every exported top-level func, type, method on an exported type, and exported
        const/var (a doc on the enclosing `const (...)`/`var (...)` group counts for its
        members) has a non-empty doc comment;
      - no comment in the package matches
        `research\.md|data-model\.md|spec\.md|\bFR-\d|\bSC-\d|§|[Pp]rinciple [IVX]+|constitution|\b0\d\d-[a-z]`;
      - every import path is standard library (first path element contains no `.`).
      Report each violation with `file:line`. Red: `../../hub` does not exist yet.

### Implementation

- [ ] T015 [US1] `git mv internal/hub hub`; rewrite all imports of
      `"github.com/Sonora-Multiroom/sonora-cli/internal/hub"` to
      `"github.com/Sonora-Multiroom/sonora-cli/hub"` (77 files incl. `tests/contract`,
      `tests/integration`, `tests/unit`, `internal/cli/**`); `go build ./... && go vet ./... && go test ./...` —
      the only failures allowed are T013/T014 (red by design)
- [ ] T016 [US1] Commit the move on its own: `refactor: move internal/hub to public hub
      package` (no content edits besides import lines, so `git log --follow hub/` shows
      history)
- [ ] T017 [US1] Create `internal/cli/exitcode/exitcode.go`: package doc ("maps hub error
      classes to sonora CLI process exit codes"); `const Usage = 2`; `func For(c
      hub.ErrorClass) int` with the switch from `hub/errors.go` `ExitCode()` minus the
      `ClassUsage` case; keep the explanatory comment about retired code 7 (internal code
      may cite spec artifacts)
- [ ] T018 [US1] In `internal/cli/**`, replace every `hub.ClassUsage.ExitCode()` with
      `exitcode.Usage` and add the `exitcode` import (sed across the ~34 files, then
      `goimports`-style fix by hand or `gofmt`); also replace any other `hub.ClassUsage`
      use (e.g. passed to a helper that later calls `.ExitCode()`) with the equivalent
      `exitcode.Usage` flow — `git grep -n 'ClassUsage' internal/` must be empty afterward
- [ ] T019 [US1] In `internal/cli/**`, replace every remaining `<expr>.ExitCode()` on a
      `hub.ErrorClass` value with `exitcode.For(<expr>)` (e.g. `class.ExitCode()` →
      `exitcode.For(class)`, `hub.ClassNotFound.ExitCode()` → `exitcode.For(hub.ClassNotFound)`);
      `git grep -n '\.ExitCode()' internal/` must be empty afterward
- [ ] T020 [US1] In `tests/unit/hub_client_test.go`, delete `TestErrorClass_ExitCodes`,
      `TestErrorClass_NewExitCodes`, `TestErrorClass_AllExitCodesDistinct`,
      `TestErrorClass_RouteExitCodes` (and any other test calling `.ExitCode()` on a class —
      15 call sites today; now covered by T013), and in `TestClassifyError_NotFound` drop
      the `ClassUsage`/exit-code-uniqueness block at its end. Replace `hub.ClassUsage` in any
      other file under `tests/` with the matching `exitcode.Usage` assertion. Do NOT touch
      `exec.ExitError.ExitCode()` calls in `tests/integration/outputs_list_test.go` and
      `tests/integration/tts_test.go`
- [ ] T021 [US1] In `hub/errors.go`, delete the `ExitCode()` method and the `ClassUsage`
      constant (research §4); run `go build ./... && go vet ./... && go test ./...` — all
      green except T014 (godoc) which stays red until T026
- [ ] T022 [US1] Commit: `refactor!: move exit-code mapping out of hub into internal/cli/exitcode`
      (body: `hub.ClassUsage` and `ErrorClass.ExitCode()` removed; later `ErrorClass`
      integer values shift; no exit code changes)
- [ ] T023 [P] [US1] Create `hub/doc.go` with the package comment: what the package is
      (Multiroom Audio Hub API client for `api/openapi.json`), the call shape
      (`NewClient()` or `NewClientWithTimeout`, then `Op(ctx, client, baseURL, …)`), one
      HTTP request per call with no retries, typed errors and `ClassifyError` for a
      class plus one-line user-facing message, standard library only; remove the
      one-line package comment from `hub/client.go`
- [ ] T024 [P] [US1] Rewrite godoc in `hub/tts.go` (20 flagged lines): replace citations
      of `research.md`, `data-model.md`, `FR-…`, `§…`, `009-tts-commands`, constitution
      principles with a direct statement of the behavior; keep `#/components/schemas/...`
      and `api/openapi.json` references
- [ ] T025 [P] [US1] Same godoc rewrite in `hub/routes.go` (9), `hub/play.go` (4),
      `hub/extensions.go` (4)
- [ ] T026 [P] [US1] Same godoc rewrite in `hub/errors.go` (8), `hub/client.go` (4),
      `hub/groups.go` (3), `hub/inputs.go` (2), `hub/outputs.go` (2), `hub/mastermute.go`
      (1); add doc comments to any exported identifier T014 reports as undocumented
      (e.g. the `Class*` const group, `DiagnosisUnknown` group, `Error()` methods if
      flagged); then `go test ./tests/unit -run HubGodoc` green and `go test ./...` green
- [ ] T027 [US1] Review `go doc -all ./hub` end to end: it must read as a self-contained
      client API (quickstart §4); fix wording that assumes CLI context (e.g. "command",
      "--verbose", "exit") in exported docs
- [ ] T028 [US1] Commit: `refactor: make hub godoc self-contained for external consumers`
- [ ] T029 [US1] External-consumer check, pre-release (quickstart §5): in the session
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

- [ ] T030 [US2] Run quickstart §2 stray-reference sweep:
      `git grep -n '"sonora-cli/' -- ':!specs' ':!docs/reviews'`,
      `git grep -n 'internal/hub' -- ':!specs' ':!docs/reviews' ':!.specify'`,
      `git grep -n 'X sonora-cli/' -- ':!specs'` — all empty; fix any hit in code, tests,
      scripts, README or CONTRIBUTING
- [ ] T031 [US2] Quickstart §3: `sh build.sh && ./sonora.exe --version` prints
      `git describe --tags --always --dirty` output, not `dev`
- [ ] T032 [US2] Linux check: `go test ./...` on Linux — push the branch and confirm the
      `test.yml` workflow is green, or run `docker run --rm -v "$PWD":/app -w /app
      golang:1.27-alpine go test ./...` locally
- [ ] T033 [US2] GoReleaser config check: if `goreleaser` is available run
      `goreleaser build --snapshot --clean --single-target` and confirm the built binary's
      `--version` is not `dev`; otherwise rely on T003 and note it here
- [ ] T034 [US2] Behavior spot check (quickstart §7) against a reachable hub if one is
      available: compare `sonora-before.exe` (T002) and the new `sonora.exe` for
      `get outputs`, `get outputs --json`, `get routes`, `get outputs/no-such-id`,
      `set outputs/no-such-id volume 10`, and a usage error (`sonora get`) — identical
      stdout/stderr and exit codes. If no hub is reachable, run the usage-error and
      unreachable-hub cases (`--hub-url` / config pointing at a closed port) only and record
      that here

**Checkpoint**: no observable CLI change on Windows and Linux.

---

## Phase 5: User Story 3 — Consumers verify against the same hub spec (Priority: P2)

**Goal**: `github.com/Sonora-Multiroom/sonora-cli/api.Spec` exposes `api/openapi.json`
with no second copy (research §7).

**Independent Test**: T035 green; quickstart §5 consumer prints `len(api.Spec) > 0`.

- [ ] T035 [P] [US3] Create `tests/unit/api_spec_test.go`: `api.Spec` is byte-identical to
      `os.ReadFile("../../api/openapi.json")`, non-empty, and `json.Unmarshal` into
      `map[string]any` yields a string `openapi` field. Red: package missing
- [ ] T036 [US3] Create `api/spec.go`: `package api` with package doc, `import _ "embed"`,
      `//go:embed openapi.json` above `var Spec []byte`, godoc stating it is the exact
      Multiroom Audio Hub API spec the `hub` package at the same module version is
      tested against and that callers must not modify it; T035 green
- [ ] T037 [US3] Confirm the CLI binary does not link `api`:
      `go list -deps ./cmd/sonora | grep -x 'github.com/Sonora-Multiroom/sonora-cli/api'`
      prints nothing
- [ ] T038 [US3] Extend the T029 scratch consumer to print `len(api.Spec) > 0` (true)
- [ ] T039 [US3] Commit: `feat: publish hub OpenAPI spec as api.Spec`

**Checkpoint**: consumers can import the spec at the client's version.

---

## Phase 6: User Story 4 — Maintainers release the shared client safely (Priority: P2)

**Goal**: A `v0.1.0` tag consumers can resolve through the public module proxy.

**Independent Test**: quickstart §5 (post-tag, `GOWORK=off`) and §6.

- [ ] T040 [US4] Write release notes in `specs/010-public-hub-package/release-notes.md`
      for `v0.1.0`: module path is now `github.com/Sonora-Multiroom/sonora-cli`; new public
      packages `hub` and `api` (link contracts/public-packages.md); not in the public API:
      `ErrorClass.ExitCode()`, `ClassUsage`; `ErrorClass` integer values are not stable —
      compare against constants; requires Go ≥ 1.27; no CLI behavior change; `go install
      github.com/Sonora-Multiroom/sonora-cli/cmd/sonora@v0.1.0` now works. Commit with
      `spec:` type
- [ ] T041 [US4] Open the PR for `010-public-hub-package` and merge after review and green
      CI (outward-facing — confirm with the maintainer before pushing/merging)
- [ ] T042 [US4] On `main` after merge, run `./release.sh --minor` (expects `v0.0.17` →
      `v0.1.0`); paste T040's notes into the GitHub release body. Confirm with the
      maintainer before pushing the tag. Never move, delete or re-create a pushed tag —
      fix forward with `v0.1.1`
- [ ] T043 [US4] Post-tag consumer check (quickstart §5): in the scratch consumer, `rm
      go.work`, `GOWORK=off go get github.com/Sonora-Multiroom/sonora-cli@v0.1.0`,
      `GOWORK=off go run .` — same output as T029/T038, no `replace` directives

**Checkpoint**: sonora-mcp can `go get` the tag (unblocks its Phase 3).

---

## Phase 7: Polish & Cross-Cutting

- [ ] T044 [P] Add a short "Using the hub client from Go" section to `README.md`
      (`go get …@v0.1.0`, import `…/hub`, 5-line example, link to `go doc`); commit `docs:`
- [ ] T045 Final gate: `gofmt -l .` (no output), `go vet ./...`, `go test ./...` green;
      mark all tasks above `[X]`
- [ ] T046 [P] Tick the Phase 2 checklist items in
      `D:\projects-sonora\sonora-mcp\docs\future\go-rewrite-shared-hub-client.md`
      (sonora-mcp repo; separate commit there, confirm with the maintainer)

---

## Dependencies & Execution Order

- **Phase 1** → **Phase 2 (T003–T012)** → everything else. The module path is in every
  import line, so no story can start before T012.
- **US1 (Phase 3)**: T013/T014 [P] first (red) → T015 → T016 → T017 → T018 → T019 → T020
  → T021 → T022 → T023–T026 [P] → T027 → T028 → T029.
- **US3 (Phase 5)**: depends only on Phase 2; can run in parallel with US1 (touches
  `api/` and a new test file only). T038 needs T029's scratch consumer.
- **US2 (Phase 4)**: verification — run after US1 and US3 are committed (it validates the
  final tree).
- **US4 (Phase 6)**: after US1–US3 and US2 verification; T042 after merge; T043 after tag.
- **Polish**: T044 any time after T040; T045 last before PR; T046 after T043.

## Parallel Opportunities

- Phase 2: T008, T009, T010 (three different config files).
- US1: T013 ∥ T014 (new test files); T023 ∥ T024 ∥ T025 ∥ T026 (disjoint `hub/*.go` files).
- US1 ∥ US3: T035–T037 can proceed while T015–T028 are in progress.

```text
# After T012:
Agent A: T013, T014 → T015 … T022            (hub move + exitcode)
Agent B: T035 → T036 → T037                 (api.Spec)
# After T022:
Agents A–D: T023 | T024 | T025 | T026         (godoc, one file group each)
```

## Implementation Strategy

- **MVP**: Phase 1 + Phase 2 + US1 (T001–T029). At that point `hub` is importable via a
  workspace and the CLI is green — enough to start sonora-mcp prototyping with `go.work`.
- **Increment 2**: US3 (`api.Spec`) — small, independent.
- **Increment 3**: US2 verification on both OSes, then US4 release `v0.1.0`.
- Keep commits small and in the order above so the `git mv` commit (T016) is a pure
  rename and each `refactor!` commit is reviewable on its own.

## Notes

- `[P]` = different files, no dependency on an incomplete task.
- Internal code (e.g. `internal/cli/exitcode`) may keep references to spec artifacts;
  only `hub/` public docs are restricted (constitution VII).
- Historical records (`specs/001–009/**`, `docs/reviews/**`, the constitution's Sync
  Impact Report) are not rewritten.
