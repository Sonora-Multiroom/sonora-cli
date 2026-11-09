# Research: Hub Info and TTS Cache Warm-up

## R1. Hub info response shape

- **Decision**: `Info{Product, Version, APIVersion string; StartedAt string (omitempty); UptimeSeconds
  *int64 (omitempty)}`, mirroring `#/components/schemas/HubInfoResponse`. `product`, `version` and
  `apiVersion` are required on decode; the other two are optional.
- **Rationale**: the spec declares `HubInfoResponse` for the `200` answer (added to the hub after 0.1.33
  shipped without it, and refreshed into `api/openapi.json` on 2026-11-09). The schema has five optional
  properties and no `required` list, so requiring `product`, `version` and `apiVersion` is the client's
  choice: they identify a hub, and a body without them is not a hub info answer (spec edge case). The time
  fields stay optional so a hub that drops one still decodes. `version` and `apiVersion` can be the string
  `unknown`; they are passed through, never parsed as semver. The live hub 0.1.33 answered (2026-11-06,
  read-only), and the contract test uses the same shape:
  `{"product":"multiroom","version":"0.1.33","apiVersion":"0.1.33","startedAt":"2026-11-06T10:06:59.512Z","uptimeSeconds":16981}`.
- **Alternatives**: `map[string]any` (rejects nothing, useless to typed consumers); all fields required
  (breaks on a hub that omits `uptimeSeconds`).

## R2. 404 on info

- **Decision**: `GetInfo` returns a new `*InfoNotOfferedError` for 404; `ClassifyError` maps it to
  `ClassHub` with "this hub does not offer hub info (needs hub 0.1.33 or later)".
- **Rationale**: a plain `*StatusError` would print "hub reported an error (HTTP 404)", which tells the
  user nothing. `ClassNotFound` is for an unknown output/input/group, so it would mislead.
- **Alternative**: handle 404 in the command only; rejected because sonora-mcp would need the same
  mapping (Principle VII).

## R3. Command spelling

- **Decision**: `sonora get info`, `sonora warm tts-cache <text|->`.
- **Rationale**: Principle V wants verb then resource. `get` already serves singletons (`master-mute`,
  `tts-cache`). Warming is not creating a resource the user names, nor setting a value, so a new verb
  reads best and sits beside `clear tts-cache`. `tts-cache` is the noun both commands share.
- **Alternatives**: `create tts-cache/<text>` (the text is not an id); `speak --warm` (changes `speak`'s
  contract: no target, no playback flags); `set tts-cache` (nothing is set).

## R4. Warm reuse of speak code

- **Decision**: `WarmRequest` has `Text string` and pointer-typed optional fields with `omitempty`, as
  `SpeakRequest` does. `WarmTTSCache` copies `Speak`'s flow: 404 gives `*TTSNotOfferedError`, other non-2xx
  go through `decodeTTSError`. A 2xx decodes through a wire type with `*bool` so a missing `cacheHit` is a
  `*DecodeError`.
- **Rationale**: same extension, same error body (`TtsErrorResponse`), same diagnosis of "TTS
  unavailable". `ReportError` in the CLI already handles all of it.
- **Alternative**: default a missing `cacheHit` to false; rejected because it would claim "synthesized"
  on a malformed reply.

## R5. 503 and CACHE_WRITE_FAILED

- **Decision**: add `CACHE_WRITE_FAILED` to the `ClassServiceUnavailable` codes in `ClassifyError`. The
  warm command appends "; nothing was stored, repeating the request is safe" on that class.
- **Rationale**: the spec states a 503 stores nothing and a repeat is safe. The status fallback already
  gives class 503, but only the code identifies the cause. The hint is CLI wording, so it stays in
  `internal/` (Principle VII).

## R6. Warm CLI flags

- **Decision**: `--provider`, `--voice`, `--language`, `--timeout`, `--json`, `--verbose`, `--hub-url`,
  validated as in `speak` (a given flag must not be empty; `--timeout` a positive duration).
- **Rationale**: a cache entry is keyed by text and voice; using `speak`'s options means warmed entries
  match what the CLI later announces. The other four request fields stay `hub`-only until needed.
- **Alternative**: expose all eight; rejected as scope growth with no stated user.

## R7. Live checks

- **Decision**: `get info` against the live hub is read-only and allowed. `warm tts-cache` writes to the
  production TTS cache (no sound, no speaker change). It needs the user's permission each time, uses
  `--provider piper-local` (free, runs on the Pi), and leaves one cache entry; `clear tts-cache` is not
  used, because it would also remove entries the user already has.
- **Rationale**: the workspace rules for production require permission for anything that changes state.
