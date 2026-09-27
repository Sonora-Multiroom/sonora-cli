# Data Model: Public Hub Client Package

**Feature**: 010-public-hub-package | **Date**: 2026-09-27

This is a refactor; no hub request/response types change. The model below covers the
entities whose location or shape changes.

## ErrorClass (public, `hub`)

Category of a hub-call failure, produced by `hub.ClassifyError(err) (ErrorClass, string)`.

| Class | Produced when | CLI exit code (via `exitcode.For`) |
|-------|---------------|------------------------------------|
| `ClassNone` | `err == nil` | 0 |
| `ClassHub` | non-2xx without a more specific type; malformed response; other API status | 3 |
| `ClassNetwork` | unreachable hub, timeout, TTS unavailable because of the hub address | 4 |
| `ClassNotFound` | 404 for a resource; TTS `PROVIDER_NOT_FOUND` | 5 |
| `ClassValidation` | 400 / TTS `INVALID_REQUEST` | 6 |
| `ClassRouteFailed` | 422 on route creation / playback | 8 |
| `ClassSourceUnreachable` | 502 on playback | 9 |
| `ClassServiceUnavailable` | 503; TTS provider timeout / rate limit / error | 10 |
| `ClassInputNotFound` | not produced by `ClassifyError`; assigned by callers (route create) when a `NotFoundError` refers to the input | 11 |
| `ClassTargetNotFound` | TTS `TARGET_NOT_FOUND`; also assigned by callers (route create/transfer) when a `NotFoundError` refers to the target | 12 |
| `ClassTTSUnavailable` | TTS not offered / extension not ready | 13 |

- **Removed from `hub`**: `ClassUsage` (never produced by `ClassifyError`) and the method
  `ErrorClass.ExitCode()`. See [research.md §4](research.md).
- **Validation rules**: `ClassifyError` messages MUST NOT mention flags, commands, or exit
  codes. Exit code 7 stays retired and is never returned. No two classes share an exit
  code.

## Exit-code mapping (internal, `internal/cli/exitcode`)

- `Usage = 2` — CLI argument/usage errors (replaces `hub.ClassUsage.ExitCode()`).
- `For(hub.ErrorClass) int` — the table above; unknown classes map to 0 (same as today's
  `default` branch).
- Relationship: depends on `hub`; `hub` never depends on it.

## Published spec (public, `api`)

- `api.Spec []byte` — full bytes of `api/openapi.json` at the same module version.
- Validation: byte-identical to the file; non-empty; valid JSON with a top-level `openapi`
  field.
- Callers MUST NOT modify it.

## Release tag

- Form `vMAJOR.MINOR.PATCH`; first tag with the public packages is `v0.1.0`.
- State: created once → published to the module proxy → immutable. A fix is a new tag.

## Package dependency direction

```text
cmd/sonora ──► internal/cli/* ──► internal/cli/exitcode ──► hub
                    │                                      ▲
                    ├──► internal/render, internal/config  │
                    └──────────────────────────────────────┘
api  (leaf, no imports beyond embed)
hub  (standard library only; imports nothing from internal/ or api)
```
