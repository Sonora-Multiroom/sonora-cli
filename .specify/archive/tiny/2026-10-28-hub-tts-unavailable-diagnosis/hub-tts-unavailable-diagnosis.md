# TinySpec: Diagnose "TTS not offered" in `hub`

**Branch**: fix/hub-tts-unavailable-diagnosis
**Date**: 2026-10-28
**Status**: released in v0.3.3
**Complexity**: small

## What

`hub.ClassifyError` reports a TTS 404 (`*TTSNotOfferedError`) as `ClassNetwork`, "could not reach the
hub", and the mapping from the `tts` extension's status to a diagnosis is private to the CLI. Make the
diagnosis public as `hub.DiagnoseTTSUnavailable` and classify `*TTSNotOfferedError` correctly, so
sonora-mcp's TTS tools can report why TTS is unavailable. Source: an uncommitted backlog item
(`docs/backlog/tts-unavailable-diagnosis.md`, 2026-10-28) that this tinyspec replaces.

## Context

| File | Role |
|------|------|
| `hub/tts.go` | Modified: add `DiagnoseTTSUnavailable` beside `TTSNotOfferedError` and `TTSUnavailableError` |
| `hub/errors.go` | Modified: `ClassifyError` branch for `*TTSNotOfferedError` |
| `hub/extensions.go` | Context: `ListExtensions`, `ExtensionInventory` |
| `internal/cli/tts/report.go` | Modified: `diagnoseUnavailable` becomes a call to `hub.DiagnoseTTSUnavailable` |
| `tests/unit/hub_client_test.go` | Modified: diagnosis cases and the new `ClassifyError` case |
| `tests/unit/tts_report_test.go` | Unchanged: guards the CLI's hint, wording and single inventory call |

## Requirements

1. `DiagnoseTTSUnavailable(ctx context.Context, client *http.Client, baseURL string) *TTSUnavailableError`
   makes exactly one `ListExtensions` call with the given context and client, and never returns nil.
2. It maps, matching on the extension id `tts` only: absent → `DiagnosisNotInstalled`, with
   `LoadingDisabled` true only when `loadingEnabled` is `false`; `DISABLED` → `DiagnosisDisabled`;
   `REJECTED` → `DiagnosisRejected` with `Reason` = `SingleLine(rejectionReason)` (empty when null);
   `INERT` → `DiagnosisInert`; `ACTIVE` → `DiagnosisVersionMismatch`; any other status →
   `DiagnosisUnknown`; inventory 404 → `DiagnosisHubAddress` with `BaseURL`; any other inventory
   failure → `DiagnosisUnknown` with `Cause` set to that error.
3. `ClassifyError(&TTSNotOfferedError{})`, wrapped or not, returns `ClassTTSUnavailable`,
   "text-to-speech is not available on this hub".
4. The CLI's output and exit codes are unchanged: `tests/unit/tts_report_test.go` passes unmodified.
   The CLI still diagnoses with `context.Background()` and a fresh `hub.NewClient()` (5 s), not the
   speak client.
5. The new function's godoc stands without this repo's specs (`tests/unit/hub_godoc_test.go`).

## Plan

1. Tests first in `tests/unit/hub_client_test.go`: a table over the requirement 2 cases against an
   `httptest` extensions server, a one-call count, and the `ClassifyError` case. They fail to compile
   until the function exists; the `ClassifyError` case fails on `ClassNetwork`.
2. Move the body of `diagnoseUnavailable` into `hub/tts.go`, taking `ctx` and `client`.
3. Add the `*TTSNotOfferedError` branch to `ClassifyError` next to the other TTS branches.
4. Replace the CLI's `diagnoseUnavailable` body with the `hub` call.
5. Release `v0.3.3`; its notes say `ClassifyError` now returns `ClassTTSUnavailable` instead of
   `ClassNetwork` for `*TTSNotOfferedError`. sonora-mcp bumps it in its own repo.

## Tasks

- [x] Failing tests for `DiagnoseTTSUnavailable` (every mapping, one call) and `ClassifyError`
- [x] Add `hub.DiagnoseTTSUnavailable` with standalone godoc
- [x] Add the `ClassifyError` branch for `*TTSNotOfferedError`
- [x] Switch `internal/cli/tts/report.go` to the `hub` function
- [x] `gofmt -l .`, `go vet ./...`, `go test ./...`
- [x] Delete the backlog item and its `INDEX.md` row

## Done When

- [x] All tasks checked off; tests pass; no vet or gofmt findings
- [x] Tag `v0.3.3` released with the reclassification in its notes

## Verification (2026-10-28)

- Red first: `go test ./tests/unit/ -run 'TTSNotOffered|DiagnoseTTS'` failed to build
  (`undefined: hub.DiagnoseTTSUnavailable`); with the function added, the 14 diagnosis subtests passed
  and both `TestClassifyError_TTSNotOfferedError` subtests failed on `ClassNetwork`, "could not reach
  the hub".
- After the `ClassifyError` branch and the CLI switch: `gofmt -l .` no output; `go vet ./...` exit 0;
  `go test ./...` exit 0 (4 packages ok). `tests/unit/tts_report_test.go` unchanged and passing.
- No live hub calls: the change is client-side and covered by `httptest` servers.
