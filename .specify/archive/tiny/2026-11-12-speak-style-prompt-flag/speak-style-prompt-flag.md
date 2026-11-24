# TinySpec: `--style-prompt` on `speak` and `warm tts-cache`

**Branch**: feature/speak-style-prompt-flag
**Date**: 2026-11-12
**Status**: done, tag pending
**Complexity**: small

## What

`sonora speak` and `sonora warm tts-cache` gain `--style-prompt TEXT`, which sends the request's
`stylePrompt` (sonora-tts TTS-004, `google-gemini` only), so a Gemini announcement's delivery can be
set from the shell and a styled phrase can be warmed under the cache key `speak` will look up.
`hub.SpeakRequest` and `hub.WarmRequest` already carry the field; only the commands change. This
supersedes requirement 6 of `speak-request-fields.md` and spec 013's deferral for this one field.
Source and rationale: `docs/backlog/speak-style-prompt-flag.md`, which this tinyspec replaces.

## Context

| File | Role |
|------|------|
| `internal/cli/tts/speak.go` | Modified: the flag, filling `req.StylePrompt` |
| `internal/cli/tts/warm.go` | Modified: the same; doc comment |
| `hub/tts.go` | Context: `StylePrompt *string` on both requests (unchanged) |
| `api/openapi.json` | Context: `stylePrompt` on `SpeakRequest` and `WarmRequest` (0.1.33) |
| `tests/unit/cli_tts_speak_test.go` | Modified: request bodies with and without the flag |
| `tests/integration/tts_warm_test.go` | Modified: request bodies with and without the flag |
| `README.md` | Modified: `speak` and `warm tts-cache` synopsis and paragraphs |
| `docs/backlog/speak-style-prompt-flag.md`, `docs/backlog/INDEX.md` | Deleted / row removed when it ships |

## Requirements

1. `--style-prompt TEXT` on both commands sends `stylePrompt` verbatim; whether it was given is read
   with `fs.Visit` (the existing `provided` map), not from the value.
2. Given but empty or blank (`--style-prompt ""`, `"  "`): sent as given, not a usage error. The hub
   reads it as "turn the entry's default off". This differs on purpose from `--provider`, `--voice`,
   `--language` and `--playback-mode`, whose blank values are usage errors.
3. No client-side length or provider check: the hub's 400 decides (over `max-text-length`, or any
   `stylePrompt` sent to a non-`google-gemini` provider) and exits with `ClassValidation`'s code, as
   other rejected requests do.
4. Flag absent: the request body is byte-identical to today's (no `stylePrompt` key).
5. Help text names `google-gemini` and says an empty value turns the configured prompt off, e.g.
   "`google-gemini` only: deliver the announcement as `TEXT` describes; \"\" turns the configured
   style prompt off; without it, the provider's configured prompt". `warm` uses the same wording.

## Plan

1. Tests first (watch them fail): for `speak` and for `warm`, the body has `"stylePrompt":"calmly"`
   with the flag, `"stylePrompt":""` with `--style-prompt ""`, and no `stylePrompt` key without it.
2. Add the flag and the `provided["style-prompt"]` fill in `speak.go` and `warm.go`; keep it out of
   the blank-value loop.
3. README: add `[--style-prompt]` to both synopses and one sentence each. Delete the backlog item and
   its INDEX row. Release a minor tag (no `!`); sonora-mcp needs no bump (`hub` is unchanged).

## Tasks

- [x] Unit tests: `speak` body with value / empty / absent (`tests/unit/cli_tts_speak_test.go`)
- [x] Integration tests: `warm tts-cache` body with value / empty / absent (`tests/integration/tts_warm_test.go`)
- [x] Implement `--style-prompt` in `internal/cli/tts/speak.go`
- [x] Implement `--style-prompt` in `internal/cli/tts/warm.go` and update `RunWarm`'s doc comment
- [x] README `speak` and `warm tts-cache` synopsis and paragraphs
- [x] Delete `docs/backlog/speak-style-prompt-flag.md` and its `INDEX.md` row
- [x] `gofmt -l .`, `go vet ./...`, `go test ./...` clean

## Verification

Production hub `multiroom.lan`, API 0.1.33, 2026-11-12, with the branch's build:

- `warm tts-cache --provider piper-local --style-prompt calmly` and `--style-prompt ""`: 400
  `Field 'stylePrompt' is not supported by provider 'piper-local' of type PIPER`, exit 6; cache
  entry count unchanged (105).
- `warm tts-cache --provider gemini` on one phrase: `--style-prompt "Say this cheerfully and
  quickly."` miss then hit; `--style-prompt ""` miss; no flag (entry default) miss. Three distinct
  cache entries, so the prompt reaches the hub and is part of the key.
- `speak … outputs/office --provider gemini --style-prompt "Say this cheerfully and quickly."
  --route-volume 20`: accepted, `cacheHit: true`. Office volume 30 before and after; no test routes
  or inputs left.

## Done When

- [x] All tasks checked off
- [x] Tests pass
- [x] No lint errors
