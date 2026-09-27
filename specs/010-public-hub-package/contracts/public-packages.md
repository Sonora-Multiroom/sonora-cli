# Contract: Public Packages of `github.com/Sonora-Multiroom/sonora-cli`

**Feature**: 010-public-hub-package | **Date**: 2026-09-27

The module exposes exactly two importable packages. Everything else is under `internal/`
(or is `package main`) and is unimportable by other modules.

## `github.com/Sonora-Multiroom/sonora-cli/hub`

Multiroom Audio Hub API client. Standard library only.

### Construction

| Identifier | Contract |
|------------|----------|
| `NewClient() *http.Client` | 5s overall request timeout, default transport (keep-alive). No I/O. |
| `NewClientWithTimeout(d time.Duration) *http.Client` | Same with timeout `d`. |
| `SpeakTimeout` | Recommended timeout for `Speak` (15s). |

### Operations

Every operation has the shape `Op(ctx context.Context, client *http.Client, baseURL string,
…) (…, error)`, performs exactly one HTTP request (no retries), and returns either a
decoded value matching `api/openapi.json` or an error classifiable by `ClassifyError`.
The set of operations is unchanged by this feature:

- Outputs: `ListOutputs`, `GetOutput`, `SetOutputEnabled`, `SetOutputMuted`,
  `SetOutputVolume`, `StopRoutesForOutput`
- Groups: `ListGroups`, `GetGroup`, `SetGroupEnabled`, `SetGroupMuted`, `SetGroupVolume`,
  `StopRoutesForGroup`
- Inputs: `ListInputs`, `GetInput`, `CreateInput`, `DeleteInput`, `SetInputEnabled`
- Routes: `ListRoutes`, `GetRoute`, `CreateRoute`, `DeleteRoute`, `TransferRoute`,
  `SetPauseState`, `StopAllRoutes`
- Playback: `Playback`, `ResolveTarget`
- Master mute: `GetMasterMute`, `SetMasterMute`
- TTS & extensions: `Speak`, `ListTTSVoices`, `GetTTSCacheStats`, `ClearTTSCache`,
  `ListExtensions`

Request/response types (`Output`, `Group`, `Input`, `Route`, `PlaybackRequest`, …) are
unchanged.

### Errors

| Identifier | Contract |
|------------|----------|
| `ErrorClass`, `Class*` constants | See [data-model.md](../data-model.md). `ClassUsage` is **not** part of the public API. |
| `ClassifyError(err) (ErrorClass, string)` | Total: every non-nil error maps to a class and a one-line, CLI-agnostic message. `nil` → `ClassNone, ""`. |
| `StatusError`, `DecodeError`, `NotFoundError`, `APIError`, `TTSError`, `TTSUnavailableError`, `TTSNotOfferedError`, `TTSDiagnosis` | Typed errors usable with `errors.As`. |
| `SingleLine(s string) string` | Collapses `\r\n`, `\n`, `\r` into single spaces. |

**Removed** (was in `internal/hub`, never public): `ErrorClass.ExitCode()`, `ClassUsage`.

### Documentation

Every exported identifier has a doc comment. No doc comment references `research.md`,
`data-model.md`, `FR-`/`SC-` IDs, `§` sections, feature numbers, or constitution
principles.

## `github.com/Sonora-Multiroom/sonora-cli/api`

| Identifier | Contract |
|------------|----------|
| `Spec []byte` | Exact bytes of `api/openapi.json` at this module version; the spec the `hub` package is tested against. Read-only. |

## Versioning

- Governed by the module's `vX.Y.Z` tags. While `v0.x`, breaking changes are allowed and
  are listed in release notes (constitution VII).
- Consumers need Go ≥ the `go` directive in `go.mod` (currently 1.27.0).

## Not public (stays internal)

`internal/cli/**` (commands, flags, `exitcode`), `internal/config` (hub URL discovery),
`internal/render` (YAML/JSON output), `internal/version`.
