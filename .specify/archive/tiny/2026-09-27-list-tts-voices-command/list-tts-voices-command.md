# TinySpec: `list tts-voices` command

**Branch**: feature/list-tts-voices-command
**Date**: 2026-09-27
**Status**: done
**Complexity**: small

## What

Add `sonora list tts-voices --provider <name> [--language <code>] [--engine <name>]`, wrapping
`GET /api/tts/providers/{providerName}/voices` (operationId `listVoices`), so users can discover
valid `speak --voice` values. Like `tts-cache`, `tts-voices` is a literal keyword special-cased in
`dispatchGetList`, not a `respath` resource; it is a collection, so `get tts-voices` is the usage
error (inverse of `tts-cache`). The provider is a required path parameter, taken as `--provider`
to match `speak`/`clear tts-cache`. Voices are listed in hub order; the hub already sorts them.

Hub-side behaviour (sonora-tts 0.1.3, archived features 002 and 004,
`.specify/archive/004-gemini-tts-provider/contracts/tts-rest-api.yaml`): only `google-cloud` and
`google-gemini` entries list voices, and every other type returns 400 `INVALID_REQUEST`. A Gemini voice
has no `language`, its `engine` is the entry's model, and `--engine` on a Gemini entry is a 400.
`gender` may be absent, and the hub leaves out null fields. Filters are case-insensitive and
accept aliases on the hub, so the CLI passes them through verbatim and does no local checks.
`speak` has no `--engine` flag, so for `google-cloud` the usable `speak --voice` value is
`fullName`; the help text says so. Adding `speak --engine` is out of scope.

## Context

| File | Role |
|------|------|
| `api/openapi.json` | Context — `listVoices` (200 `TtsVoiceListResponse`, 400/503 `TtsErrorResponse`); `TtsVoiceDescriptor{shortName, fullName, engine, language, gender}`, all optional strings, so none is required when decoding. Response media type is `*/*`: decode the body whatever its `Content-Type` |
| `internal/hub/tts.go` | Modified — `TTSVoice` (`*string` fields, so absent ≠ empty), `TTSVoiceList{ProviderName, Voices}`, `ListTTSVoices(ctx, client, baseURL, provider string, language, engine *string)`: path-escaped provider, query params only when non-nil, 404 → `*TTSNotOfferedError`, other non-2xx → `decodeTTSError`, bad body → `*DecodeError`, nil `voices` → empty slice |
| `internal/render/tts.go` | Modified — `RenderTTSVoicesYAML`/`JSON`: `providerName` plus a `voices:` list with every field emitted and an absent one as a bare `null`, as `writeStartedAt`/`writeCreatedAt` do; empty list gets the `# no voices found` + `voices: []` form used by `RenderYAML` |
| `internal/cli/tts/voices.go` | New — `RunListVoices`, mirroring `RunClearCache`: `--provider` required and non-blank, `--language`/`--engine` optional but non-blank when given, no positional args, errors via `ReportError` |
| `internal/cli/tts/report.go` | Context — 404 diagnosis; an older hub with an ACTIVE TTS extension but no voices endpoint already maps to `DiagnosisVersionMismatch` |
| `cmd/sonora/main.go` | Modified — `tts-voices` branch in `dispatchGetList`; add it to `helpText` table + an example |
| `tests/contract/tts_test.go` | Modified — `ListTTSVoices` against a mock server: 200 google-cloud and Gemini-shaped (no `language`/`gender`), with/without filters, escaped provider, nil voices, 400, 404, 503, malformed body |
| `tests/unit/cli_tts_voices_test.go`, `tests/unit/render_tts_test.go` | New / modified — flag validation, success YAML/JSON, empty list, hub error |
| `tests/integration/tts_test.go` | Modified — `list tts-voices` dispatches; `get tts-voices` and `list tts-voices/x` are usage errors |
| `README.md`, `docs/cli-command-landscape.md` | Modified — document the command; add a ✅ row under `## tts (extension)` |

## Requirements

1. `sonora list tts-voices --provider P` sends `GET /api/tts/providers/P/voices` (P path-escaped),
   prints the list (YAML by default, JSON with `--json`) and exits 0; `--language`/`--engine` add
   the matching query parameters and are omitted from the URL when not given.
2. A missing or blank `--provider`, a blank `--language`/`--engine`, or any positional argument is a
   usage error (`ClassUsage`) with no request sent.
3. `get tts-voices` and `list tts-voices/<anything>` are usage errors pointing to
   `sonora list tts-voices`.
4. 400 (unknown provider / type without voices / bad filter / `--engine` on Gemini) and 503
   (`VOICE_CATALOGUE_UNAVAILABLE`, `PROVIDER_ERROR`) surface the hub's `message`
   through `decodeTTSError` with the existing exit codes; a 404 goes through `ReportError`'s
   extension diagnosis. Nothing goes to stdout on any failure.
5. Every field of every voice is always emitted, with absent fields (such as a Gemini voice's
   `language`) as `null`, never `""`; an empty result prints an explicit empty list.
6. Existing `get`/`list`, `speak` and `tts-cache` behaviour is unchanged.

## Tasks

- [x] Contract tests for `hub.ListTTSVoices` (fail first), then implement it in `internal/hub/tts.go`
- [x] Render tests, then `RenderTTSVoicesYAML`/`JSON` in `internal/render/tts.go`
- [x] Unit tests, then `RunListVoices` in `internal/cli/tts/voices.go`
- [x] Integration tests, then wire `tts-voices` into `dispatchGetList` and `helpText`
- [x] Update `README.md` and `docs/cli-command-landscape.md`
- [x] Run `gofmt -l .`, `go vet ./...`, `go test ./...`; smoke-test against `multiroom.lan`

## Done When

- [x] All tasks checked off
- [x] `go test ./...` passes; `go vet ./...` and `gofmt -l .` clean
- [x] `sonora list tts-voices --provider <real provider>` returns voices from the live hub
