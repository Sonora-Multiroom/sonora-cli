# Data Model: Text-to-Speech Announcements (`speak`, TTS cache)

**Feature**: `009-tts-commands` | **Plan**: [plan.md](plan.md) | **Research**: [research.md](research.md)

All wire types mirror `api/openapi.json` field-for-field (Principle II). Where the extension's
own contract (`sonora-tts/specs/001-tts-extension/contracts/tts-rest-api.yaml`) is more
precise, the CLI accepts anything that contract allows (FR-016).

## Wire types (`internal/hub`)

### SpeakRequest (`#/components/schemas/SpeakRequest`), sent by `speak`

| Field | JSON | Type | Rule |
| --- | --- | --- | --- |
| Text | `text` | string | Required. Non-empty and not whitespace-only (FR-003). No CLI maximum. |
| TargetName | `targetName` | string | Required. The `<id>` of `outputs/<id>` or `groups/<id>`, already validated by `respath` (`^[a-zA-Z0-9_-]{1,255}$`). |
| TargetType | `targetType` | string enum | Required. `SINGLE_OUTPUT` for `outputs`/`out`, `OUTPUT_GROUP` for `groups`/`gr` (FR-002). |
| ProviderName | `providerName` | *string, omitempty | Sent only when `--provider` is supplied. The value must not be empty or whitespace-only (FR-004). |
| Voice | `voice` | *string, omitempty | Sent only when `--voice` is supplied. Same rule. |
| Language | `language` | *string, omitempty | Sent only when `--language` is supplied. Same rule. No BCP 47 validation client-side. |

### SpeakAccepted (`#/components/schemas/SpeakAcceptedResponse`), 202 from `speak`

| Field | JSON | Type | Required by CLI (FR-005a) |
| --- | --- | --- | --- |
| AnnouncementID | `announcementId` | string (uuid) | Yes, and non-empty. UUID format is not enforced; the value is displayed only. |
| CacheHit | `cacheHit` | bool | Yes. Presence is checked by decoding into `*bool`. |
| QueueDepth | `queueDepth` | int32 | Yes. Presence is checked by decoding into `*int32`. |

A body missing any field → `*DecodeError` → exit 3.

### TTSCacheStats (`#/components/schemas/CacheStats`), 200 from `getCacheStats`

| Field | JSON | Type | Required by CLI (FR-005a) |
| --- | --- | --- | --- |
| TotalEntries | `totalEntries` | int32 | Yes |
| TotalSizeBytes | `totalSizeBytes` | int64 | Yes |
| MaxSizeBytes | `maxSizeBytes` | int64 | Yes |
| EntriesByProvider | `entriesByProvider` | map[string]int32 | No. Absent or `null` is normalised to an empty map, never nil, after decoding. |

### clearCache: `DELETE /api/tts/cache[?providerName=<name>]`

No request body. 204 has no response body. The query parameter is present only when
`--provider` is given, and is URL-encoded with `url.Values`.

### TTS error body (`#/components/schemas/TtsErrorResponse`), 400/503 from any TTS operation

| Field | JSON | Type | Notes |
| --- | --- | --- | --- |
| Code | `error` | string | Known values: `INVALID_REQUEST`, `TARGET_NOT_FOUND`, `PROVIDER_NOT_FOUND`, `PROVIDER_TIMEOUT`, `PROVIDER_RATE_LIMITED`, `PROVIDER_ERROR`, `FORMAT_NORMALIZATION_FAILED`. Unknown values are kept verbatim. |
| Message | `message` | string | Shown to the user, trimmed to one line. |

The decode target also accepts core problem-details `title`/`detail` as a fallback message
source (research §4). Those two fields are never rendered as a code.

### ExtensionInventory (`#/components/schemas/ExtensionInventory`), 200 from `listExtensions`

| Field | JSON | Type | Used for |
| --- | --- | --- | --- |
| LoadingEnabled | `loadingEnabled` | *bool | Refines the "not installed" message when `false` (research §5). |
| Extensions | `extensions` | []Extension | Scanned for `id == "tts"`. A missing array is treated as empty. |

`extensionsDirectory` is not decoded, because nothing uses it.

### Extension (`#/components/schemas/Extension`)

| Field | JSON | Type | Used for |
| --- | --- | --- | --- |
| ID | `id` | string | Matching; the only match key (never `name`). |
| Status | `status` | string enum | `ACTIVE` \| `REJECTED` \| `DISABLED` \| `INERT`. An unknown value → generic diagnosis. |
| RejectionReason | `rejectionReason` | *string | Appended for `REJECTED` when non-empty. |

`name`, `version`, `requiredApiVersion` and `connectionState` are not decoded.

A 200 body that isn't a JSON object → diagnosis `Unknown` (never a separate failure).

## Error types (`internal/hub/errors.go`, `internal/hub/tts.go`)

| Type | Fields | Classified as |
| --- | --- | --- |
| `TTSError` | `StatusCode int`, `Code string`, `Message string` | By `Code`, then by `StatusCode` (see exit codes below) |
| `TTSNotOfferedError` | none | Internal only. Always converted to `TTSUnavailableError` before classification. |
| `TTSUnavailableError` | `Diagnosis TTSDiagnosis`, `Reason string`, `LoadingDisabled bool`, `BaseURL string`, `Cause error` | 13, or 4 when `Diagnosis == DiagnosisHubAddress` |

`TTSDiagnosis` values: `DiagnosisUnknown`, `DiagnosisNotInstalled` (and the
`loadingEnabled == false` wording), `DiagnosisDisabled`, `DiagnosisRejected`,
`DiagnosisInert`, `DiagnosisVersionMismatch`, `DiagnosisHubAddress`. `Cause` carries the
lookup's own error for `--verbose` when the diagnosis is `Unknown` because the lookup failed.

## Exit codes

The existing classes are unchanged, and one class is added.

| Code | Class | TTS sources |
| --- | --- | --- |
| 0 | success | 202 / 200 / 204 |
| 2 | usage | bad args, empty text, empty flag value, wrong target kind, unreadable standard input, `clear` of anything but `tts-cache`, `list tts-cache` |
| 3 | hub | malformed success body; non-TTS-shaped error with a status other than 400/503; unrecognised code with a status other than 400/503 |
| 4 | network | hub unreachable or timed out (5 s standard, 15 s for `speak`); inventory answers 404 (`HubAddress`) |
| 5 | not found | `PROVIDER_NOT_FOUND` |
| 6 | validation | `INVALID_REQUEST`; unrecognised or absent code on a 400 |
| 10 | service unavailable | `PROVIDER_TIMEOUT`, `PROVIDER_RATE_LIMITED`, `PROVIDER_ERROR`, `FORMAT_NORMALIZATION_FAILED`; unrecognised or absent code on a 503 |
| 12 | target not found | `TARGET_NOT_FOUND` |
| **13** | **TTS not available** (new: `ClassTTSUnavailable`) | a TTS operation answered 404, for every diagnosis except `HubAddress` |

## Rendered views (`internal/render/tts.go`)

| View | YAML fields, in order | JSON |
| --- | --- | --- |
| Speak accepted | `announcementId`, `cacheHit`, `queueDepth` | same keys |
| Cache stats | `totalEntries`, `totalSizeBytes`, `maxSizeBytes`, `entriesByProvider` (providers sorted; `{}` when empty) | same keys; `entriesByProvider` is always an object |
| Cache cleared | `cleared: all` or `cleared: provider` + `provider` | `{"cleared":"all"}` / `{"cleared":"provider","provider":"<name>"}` |

## State

The CLI holds no state. Each invocation makes one TTS request, plus at most one inventory
request after a 404. Nothing is cached or written locally.
