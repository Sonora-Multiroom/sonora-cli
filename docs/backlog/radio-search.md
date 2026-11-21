# Radio: search and browse stations (RADIO-002)

**Status:** Proposed, not started. Written 2026-11-15.

**Origin:** `sonora-radio` RADIO-002 (`specs/002-station-search-api`, extension 0.2.0, PR
`Sonora-Multiroom/sonora-radio#2`, verified live on `multiroom.lan` on 2026-11-15). Its design decision is that
every client gets search from the hub, so the CLI stays thin and holds no catalogue logic.

**Depends on:** `sonora-radio` 0.2.0 merged and deployed on the hub (RADIO-001's `radiobrowser://<uuid>` input
is already what `sonora play` passes through). `sonora-mcp` follows this item (it uses only the `hub` package),
and `sonora-mobile` has its own item, `radio-stations.md`.

**Effort:** Medium.

## Problem

`sonora play` can play a `radiobrowser://<uuid>` input, but the CLI cannot find a UUID. A user has to look the
station up on radio-browser.info and copy it by hand.

## What the hub offers

Read-only requests outside REST v2 (unversioned, like `/api/tts/*`, which the `hub` package already calls for
`speak`). Without the extension every `/api/radio/*` path is 404, which the CLI should report as "the hub has
no radio extension". Contract: the hub's `/api-docs`, and `sonora-radio/specs/002-station-search-api/contracts/radio-rest-api.yaml`.

| Request | Gives |
|---|---|
| `GET /api/radio/stations` | A page of stations. Filters `name`, `tag`, `countryCode`, `language`, `codec`, `bitrateMin`; `order` (`name`, `votes`, `clickCount`, `bitrate`), `direction` (`asc`, `desc`), `limit` (1-100, default 20), `offset`. No parameters: the 20 most voted. |
| `GET /api/radio/stations/{uuid}` | One station, even if the catalogue marks it broken |
| `GET /api/radio/countries`, `/languages` | The whole list, most stations first |
| `GET /api/radio/tags` | A page of tags (`name`, `limit`, `offset`) |

- A station has `uuid`, `uri` (`radiobrowser://<uuid>`), `name`, `favicon`, `homepage`, `countryCode`,
  `languages`, `tags`, `codec`, `bitrate`, `hls`, `lastCheckOk`, `votes`, `clickCount`; nullable fields are
  `null`. A page is `{stations, nextOffset}` and `nextOffset` is null on the last page; a tag page is
  `{tags, nextOffset}`.
- A parameter the request does not define, or one sent twice, is 400, and so is any out-of-range value
  (`limit` over 100 is refused, never reduced). `tag` and `language` match the whole value, so values come
  from the tag and language lists.
- Errors are problem details with `type` `urn:multiroom:radio:error:*` (`invalid-parameter` with a
  `parameter`, `station-not-found` 404, `catalogue-bad-answer` 502, `catalogue-unavailable` 503 with
  `Retry-After: 30`). Their shape matches `ErrorResponseV2` but is a different schema (`RadioProblem`).
- Names, tags and addresses are community-edited text; the hub strips control characters and bounds length,
  but a terminal should still print them as plain text.

## Proposal

- `hub` package: `RadioStation`, `RadioStationPage`, `RadioCountry`, `RadioLanguage`, `RadioTag`,
  `RadioTagPage` and a `RadioProblem` error decoded into the existing error type; calls `SearchStations`,
  `Station`, `Countries`, `Languages`, `Tags`. Optional query fields are pointers sent only when set, as in
  `SpeakRequest`. The 404-on-every-path case becomes a typed "radio not available" error.
- Re-fetch `api/openapi.json` (published as `api.Spec`) from a hub that has `radio` 0.2.0 and check the
  version it was fetched at; the spec then also carries the `Radio*` shapes.
- Commands:
  - `sonora radio search [name] [--tag --country --language --codec --min-bitrate --order --desc|--asc --limit --all]`
    printing name, codec, bitrate, country and the `uuid` (and `uri` with `--output json`); `--all` follows
    `nextOffset` and drops stations whose `uuid` it already printed (the order among equal values is not
    stable).
  - `sonora radio station <uuid>`.
  - `sonora get countries|languages|tags [--name]` (or `sonora radio countries|languages|tags`, decide with
    the existing command layout).
  - `sonora play --station <uuid>` as a shorthand for `radiobrowser://<uuid>`, only if it fits Start's
    current flags.
- Release a tag, then bump `sonora-mcp` (`go get github.com/Sonora-Multiroom/sonora-cli@vX.Y.Z`) and add
  search and browse tools there.

## Open questions

1. Command layout: a `radio` group, or radio nouns under `get`? The CLI's verbs today are `get`, `play`,
   `route`, `speak`; search is a read with filters.
2. Should `play` accept a search result by position or by `name` ("play the first Kiss FM")? It saves a copy
   step but hides which of several stations plays, so a plain `uuid` is the safer first version.
3. Is `sonora radio` hidden when the hub has no extension, or always present and failing with the "no radio
   extension" message? The second needs no extra request.
4. Test fixtures: copy the extension's `search-page.json` shapes, or hand-write them from the YAML?
