# TinySpec: Speak request fields and `speak` playback flags

**Branch**: feature/speak-request-fields
**Date**: 2026-10-27
**Status**: done, tag pending
**Complexity**: small

## What

`hub.SpeakRequest` gains the seven `SpeakRequest` properties of hub API 0.1.29 it lacks, and
`sonora speak` gains `--playback-mode`, `--route-volume` and `--ducking-level`, so an announcement can
duck or mix over music at its own volume. Source and rationale: `docs/backlog/speak-request-fields.md`,
which this tinyspec replaces.

## Context

| File | Role |
|------|------|
| `hub/tts.go` | Modified: `SpeakRequest` gets `Engine`, `Pitch`, `SpeakingRate`, `StylePrompt`, `PlaybackMode`, `RouteVolume`, `DuckingLevelDb`; doc comment |
| `internal/cli/tts/speak.go` | Modified: the three flags, their checks, filling the request |
| `internal/cli/play/play.go` | Context: wording and range checks of `--route-volume` / `--ducking-level` |
| `api/openapi.json` | Context: `#/components/schemas/SpeakRequest` |
| `tests/contract/tts_test.go` | Modified: request body omits the seven when nil, sends each when set |
| `tests/unit/cli_tts_speak_test.go` | Modified: flag rules and request bodies |
| `README.md` | Modified: `speak` synopsis and paragraph |
| `docs/backlog/speak-request-fields.md`, `docs/backlog/INDEX.md` | Deleted / row removed when it ships |

## Requirements

1. The seven fields are pointers with `omitempty` (`*string` for `engine`, `stylePrompt`,
   `playbackMode`; `*float64` for `pitch`, `speakingRate`, `duckingLevelDb`; `*int` for `routeVolume`),
   JSON names as in the spec. Unset: absent from the body; set (including `0` / `""`): present.
2. The `SpeakRequest` godoc stays "field-for-field" and says every optional field is sent only when
   non-nil.
3. `--playback-mode MODE` sends `playbackMode` verbatim; no client-side value check (the spec has no
   enum; the hub answers 400). Empty or blank: usage error (exit 2), as `--provider`.
4. `--route-volume N` sends `routeVolume`; outside 0–100: error naming the flag, exit code of
   `ClassValidation`, no request sent (as `play`).
5. `--ducking-level DB` sends `duckingLevelDb`; outside −60–0: same as 4. It is accepted without
   `--playback-mode` (the extension's configured mode may be `duck-others`) and with `mix` (the hub's
   400 decides), unlike `play`, whose join mode defaults to `REPLACE`.
6. No flags for `engine`, `pitch`, `speakingRate`, `stylePrompt`.
7. No flag given: the request body is byte-identical to today's.

## Plan

1. Tests first: contract tests for the body with all seven nil and all seven set. Then the struct.
2. Tests first: unit tests for each flag sent, each out-of-range/blank rejection with no request,
   `--ducking-level` alone accepted. Then `speak.go`, reusing `play`'s flag help wording.
3. README; delete the backlog item and its INDEX row. Release a minor tag (no `!`); then sonora-mcp can
   bump.

## Tasks

- [x] Contract tests: seven fields omitted when nil, each present when set (`tests/contract/tts_test.go`)
- [x] Add the seven fields and fix the godoc (`hub/tts.go`)
- [x] Unit tests for `--playback-mode`, `--route-volume`, `--ducking-level` (`tests/unit/cli_tts_speak_test.go`)
- [x] Implement the flags (`internal/cli/tts/speak.go`)
- [x] README `speak` synopsis and paragraph
- [x] Delete `docs/backlog/speak-request-fields.md` and its `INDEX.md` row
- [x] `gofmt -l .`, `go vet ./...`, `go test ./...` clean

## Done When

- [x] All tasks checked off
- [x] Tests pass
- [x] No lint errors
