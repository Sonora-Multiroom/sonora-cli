# Data Model: Hub Info and TTS Cache Warm-up

## Info (`hub.Info`)

Mirrors `#/components/schemas/HubInfoResponse`.

| Field | JSON | Type | Required on decode | Note |
|-------|------|------|--------------------|------|
| Product | `product` | string | yes, non-empty (client rule) | `multiroom` identifies the hub |
| Version | `version` | string | yes, non-empty (client rule) | can be `unknown`; not semver |
| APIVersion | `apiVersion` | string | yes, non-empty (client rule) | can be `unknown`; not semver |
| StartedAt | `startedAt` | string (RFC 3339), omitted when empty | no | equal on every call of one process |
| UptimeSeconds | `uptimeSeconds` | `*int64`, omitted when nil | no | whole seconds |

Validation: the schema has no `required` list; the required column is this client's rule. A 200 body that
is not a JSON object, or lacks a required field, is a `*DecodeError`. `product`, `version` and
`apiVersion` are passed through as received. Marshalling gives the same property names.

## WarmRequest (`hub.WarmRequest`), mirrors `#/components/schemas/WarmRequest`

| Field | JSON | Type | Sent when |
|-------|------|------|-----------|
| Text | `text` | string | always (minLength 1 is the hub's rule; the CLI rejects empty text first) |
| ProviderName | `providerName` | `*string` | non-nil |
| Voice | `voice` | `*string` | non-nil |
| Language | `language` | `*string` | non-nil |
| Engine | `engine` | `*string` | non-nil |
| Pitch | `pitch` | `*float64` | non-nil |
| SpeakingRate | `speakingRate` | `*float64` | non-nil |
| StylePrompt | `stylePrompt` | `*string` | non-nil |

## WarmResult (`hub.WarmResult`), mirrors `#/components/schemas/WarmResponse`

| Field | JSON | Type | Required on decode |
|-------|------|------|--------------------|
| CacheHit | `cacheHit` | bool | yes (a missing field is a `*DecodeError`) |

## Errors

| Situation | Error type | Class | Exit code (existing table) |
|-----------|-----------|-------|----------------------------|
| info 404 | `*InfoNotOfferedError` (new) | `ClassHub` | hub error |
| info other non-2xx | `*StatusError` | `ClassHub` | hub error |
| info bad body | `*DecodeError` | `ClassHub` | hub error |
| warm 404 | `*TTSNotOfferedError`, diagnosed by `ReportError` | `ClassTTSUnavailable` | as `speak` |
| warm 400 | `*TTSError` (`INVALID_REQUEST`, `PROVIDER_NOT_FOUND`) | `ClassValidation`, `ClassNotFound` | as `speak` |
| warm 503 | `*TTSError` (provider codes, `CACHE_WRITE_FAILED`) | `ClassServiceUnavailable` | as `speak` |
| unreachable, timeout | transport error | `ClassNetwork` | network |
