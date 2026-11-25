# Data Model: Radio Station Search and Browse

Go types in the public `hub` package, mirroring `api/openapi.json` (hub 0.1.35). JSON tags are the hub's
property names; field order follows the schema.

## RadioStation (`#/components/schemas/RadioStation`)

| Field | JSON | Go type | Required | Notes |
|---|---|---|---|---|
| UUID | `uuid` | `string` | yes | lower case; empty is a decode error |
| URI | `uri` | `string` | yes | `radiobrowser://<uuid>`, passed unchanged to `sonora play`; empty is a decode error |
| Name | `name` | `*string` | no | community text, ≤ 256 |
| Favicon | `favicon` | `*string` | no | address the hub has not fetched |
| Homepage | `homepage` | `*string` | no | as above |
| CountryCode | `countryCode` | `*string` | no | ISO 3166-1 alpha-2 |
| Languages | `languages` | `[]string` | yes | ≤ 10; nil decodes to empty |
| Tags | `tags` | `[]string` | yes | ≤ 30; nil decodes to empty |
| Codec | `codec` | `*string` | no | e.g. MP3, AAC+ |
| Bitrate | `bitrate` | `*int32` | no | kbit/s |
| HLS | `hls` | `bool` | yes | |
| LastCheckOK | `lastCheckOk` | `bool` | yes | false for a station the catalogue marks broken |
| Votes | `votes` | `int64` | yes | |
| ClickCount | `clickCount` | `int64` | yes | plays in the last 24 h |

Nullable fields marshal as `null` (no `omitempty`), so YAML and JSON both show every field (FR-006).
Decoding uses a wire struct with pointers for the required fields, so a missing `uuid`, `uri`, `hls`,
`lastCheckOk`, `votes` or `clickCount` is a `*DecodeError` (spec edge case "body not the declared shape").

## RadioStationPage (`RadioStationPage`)

| Field | JSON | Go type | Notes |
|---|---|---|---|
| Stations | `stations` | `[]RadioStation` | required; missing is a decode error; empty is a valid empty page |
| NextOffset | `nextOffset` | `*int32` | nil on the last page |

## RadioCountry, RadioLanguage, RadioTag

| Type | Fields (JSON) | Required |
|---|---|---|
| `RadioCountry` | `code` string, `name` string, `stationCount` int64 | all |
| `RadioLanguage` | `name` string, `code` *string, `stationCount` int64 | `name`, `stationCount` |
| `RadioTag` | `name` string, `stationCount` int64 | all |

`RadioTagPage` is `{tags []RadioTag (required), nextOffset *int32}`. `Countries` and `Languages` return
plain slices: the hub answers a JSON array.

## StationSearch (request, not a schema)

Query parameters of `searchRadioStations`. Every field is a pointer and sent only when non-nil, each at
most once, in a fixed order, so the hub never sees a repeated parameter (it answers 400 to one).

| Field | Query | Go type |
|---|---|---|
| Name | `name` | `*string` |
| Tag | `tag` | `*string` |
| CountryCode | `countryCode` | `*string` |
| Language | `language` | `*string` |
| Codec | `codec` | `*string` |
| BitrateMin | `bitrateMin` | `*int` |
| Order | `order` | `*string` (`name`, `votes`, `clickCount`, `bitrate`; constants provided) |
| Direction | `direction` | `*string` (`asc`, `desc`; constants provided) |
| Limit | `limit` | `*int` |
| Offset | `offset` | `*int` |

`TagSearch` has `Name`, `Limit`, `Offset` with the same rules. The `hub` package does not range-check
them: the hub's 400 is the contract. The CLI checks paging before calling (research R4).

## Errors

| Type | When | Class (exit) |
|---|---|---|
| `*APIError` + `Type`, `Parameter`, `RetryAfter` | any radio problem body except the two below | by `Type` (research R5) |
| `*NotFoundError{Resource: "radio station"}` | 404 `station-not-found` | `ClassNotFound` (5) |
| `*RadioNotOfferedError` | any other 404 from a radio path | `ClassRadioUnavailable` (15) until diagnosed |
| `*RadioUnavailableError` | built by `DiagnoseRadioUnavailable` | `ClassRadioUnavailable` (15); hub address → `ClassNetwork` (4) |
| `*StatusError` | non-2xx with a body that is not JSON | `ClassHub` (3) |
| `*DecodeError` | 2xx body not the declared shape | `ClassHub` (3) |

`ExtensionDiagnosis` (renamed from `TTSDiagnosis`, which stays as an alias) gives the reason:
`NotInstalled` (with `LoadingDisabled`), `Disabled`, `Rejected` (with `Reason`), `Inert`,
`VersionMismatch` (with the extension's `Version`), `HubAddress`, `Unknown`.
