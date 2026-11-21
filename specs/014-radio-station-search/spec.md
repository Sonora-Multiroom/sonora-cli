# Feature Specification: Radio Station Search and Browse

**Feature Branch**: `014-radio-station-search`

**Created**: 2026-11-21

**Status**: Draft

**Input**: User description: "@docs/backlog/radio-search.md" (CLI backlog item `radio-search.md`: search and
browse radio stations through the hub's `radio` extension, RADIO-002)

## Context

`sonora play` can already play a `radiobrowser://<uuid>` input (RADIO-001), but the CLI cannot find a
station's UUID: a user looks it up on radio-browser.info and copies it by hand.

RADIO-002 (`sonora-radio` 0.2.0, `specs/002-station-search-api`, verified live on `multiroom.lan` on
2026-11-15) lets the hub search the radio-browser.info catalogue for every client, so this CLI stays thin and
holds no catalogue logic. The hub offers five read-only operations, unversioned and outside REST v2 (like
`/api/tts/*`): a filtered, paged station search; one station by UUID; the country list; the language list;
and a paged tag list. Without the extension every one of them answers 404. Errors are problem details whose
`type` starts with `urn:multiroom:radio:error:`. `api/openapi.json` (hub 0.1.35) already carries these
paths.

The users are people running `sonora` interactively or from scripts, and developers of sonora-mcp, which
uses this module's public `hub` package and will add search tools after this ships.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Find a station and play it (Priority: P1)

An operator wants to play a station they know by name ("Kiss FM"). They search by name, read the matching
stations with their codec, bitrate and country, pick one, and pass its playable address to `sonora play`
without visiting any website.

**Why this priority**: this is the gap the feature closes; without it radio playback needs a browser.

**Independent Test**: search for a known station name on a hub with the radio extension, then play the
address from the result on a test output.

**Acceptance Scenarios**:

1. **Given** a hub with the radio extension, **When** the operator searches with a name, **Then** the
   matching stations are listed, each with name, codec, bitrate, country, UUID and playable address.
2. **Given** no search terms, **When** the operator searches, **Then** the hub's default page is shown
   (the 20 most voted stations).
3. **Given** filters for tag, country, language, codec and minimum bitrate, **When** the operator
   searches, **Then** each filter that was given is sent to the hub and only those.
4. **Given** a sort field and direction, **When** the operator searches, **Then** results come back in
   that order (name, votes, click count or bitrate; ascending or descending).
5. **Given** a page size, **When** the operator searches, **Then** at most that many stations are shown;
   when more exist the output says how to get the next page.
6. **Given** a request to fetch all results, **When** the operator searches, **Then** the CLI follows
   pages until the last one and prints each station once, even if the catalogue moves a station between
   pages.
7. **Given** `--json`, **When** the operator searches, **Then** the output is JSON holding every station
   field the hub returns.
8. **Given** a search that matches nothing, **When** the operator searches, **Then** the command succeeds
   with an empty result.

---

### User Story 2 - Look up one station (Priority: P2)

An operator has a UUID (from a search, a script or a saved list) and wants to see the station's details
before playing it: name, homepage, tags, codec, bitrate, whether the catalogue's last check succeeded.

**Why this priority**: it confirms what a UUID is, but search already gives most of this.

**Independent Test**: look up a UUID from a search result and compare with the hub's answer.

**Acceptance Scenarios**:

1. **Given** a known UUID, **When** the operator looks it up, **Then** every station field is shown,
   including a station the catalogue marks as broken.
2. **Given** an unknown UUID, **When** the operator looks it up, **Then** the command fails as not found
   and names the UUID.
3. **Given** a value that is not a UUID, **When** the operator looks it up, **Then** the command fails as
   a usage error that shows the hub's message.

---

### User Story 3 - Browse countries, languages and tags (Priority: P2)

Tag and language filters match only a whole value, so an operator needs the lists to know what to filter
by. They list countries or languages (most stations first), or page through tags, optionally narrowed by a
name fragment.

**Why this priority**: it makes the tag and language filters usable, but name search works without it.

**Independent Test**: list tags filtered by a fragment and use one of the results as a search filter.

**Acceptance Scenarios**:

1. **Given** a hub with the radio extension, **When** the operator lists countries or languages,
   **Then** the whole list is shown with each entry's station count, most stations first.
2. **Given** a name fragment, **When** the operator lists tags, **Then** a page of matching tags is
   shown; a page size, an offset and a fetch-all request work as for station search.
3. **Given** `--json`, **When** the operator lists any of the three, **Then** the output is JSON.

---

### User Story 4 - Search from Go code (Priority: P2)

A developer of sonora-mcp (or any Go program) calls the public `hub` package to search stations, fetch one,
and list countries, languages and tags, and gets typed results and the package's usual error
classification.

**Why this priority**: sonora-mcp can add radio tools only once the client has them; a hub client change
goes into this repository first.

**Independent Test**: the contract tests call each function against a test server and check the request
sent and the result decoded.

**Acceptance Scenarios**:

1. **Given** a search with some optional filters set, **When** the function is called, **Then** only the
   filters that were set are sent, each once.
2. **Given** a station page, **When** it is decoded, **Then** every station field and the next-page
   offset are available, and a missing next-page offset means the last page.
3. **Given** a radio problem answer (invalid parameter, station not found, catalogue bad answer,
   catalogue unavailable), **When** the function is called, **Then** the error carries its kind, message,
   the offending parameter where one is named, and the retry delay where one is given, and
   `hub.ClassifyError` places it in the right class.
4. **Given** a 404 without a radio problem body, **When** any radio function is called, **Then** the
   error is a typed "radio is not available" error.

### Edge Cases

- A hub without the radio extension: every radio command fails with "the hub has no radio extension",
  not a generic not-found message.
- The catalogue is unavailable (503 with a retry delay): the command fails, says the catalogue is
  temporarily unavailable and when to retry. The catalogue answering badly (502) is reported as an
  upstream failure.
- A page size over 100 or a negative offset: rejected as a usage error. The hub refuses, never reduces,
  an out-of-range value; the CLI does not quietly clamp it either.
- Fetching all pages stops at the last page and never loops; it also stops at the hub's highest allowed
  offset and says the results were cut off.
- Station names, tags and addresses are community-edited text: they are printed as plain text and cannot
  move the cursor, change colours or clear the terminal.
- Absent station fields (no codec, no country, no homepage) are shown as absent, not as zero or empty
  text.
- A 200 answer whose body is not the declared shape is a malformed-response error.
- The hub is unreachable: the same unreachable-hub error and exit code as every other command.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The public `hub` package MUST offer functions to search stations, fetch one station by UUID,
  and list countries, languages and tags, each returning typed records that marshal to the hub's property
  names.
- **FR-002**: The station search function MUST accept every filter the hub defines (name, tag, country
  code, language, codec, minimum bitrate), the sort field and direction, page size and offset, and MUST
  send only those that are set.
- **FR-003**: Radio error answers MUST be decoded into the package's existing error type, keeping the
  problem kind, message, offending parameter and retry delay, and MUST be classified by
  `hub.ClassifyError`.
- **FR-004**: A 404 from a radio operation that is not a "station not found" problem MUST become a typed
  "radio is not available" error, which the CLI reports as "the hub has no radio extension".
- **FR-005**: The CLI MUST offer a station search command with an optional name argument and options for
  tag, country, language, codec, minimum bitrate, sort field, sort direction, page size, offset and
  fetch-all.
- **FR-006**: Search output MUST show, per station, name, codec, bitrate, country, UUID and the playable
  address that `sonora play` accepts; `--json` MUST give every field the hub returns.
- **FR-007**: The fetch-all option MUST follow next-page offsets to the end, MUST print a station only the
  first time its UUID appears, and MUST stop at the hub's offset limit with a notice.
- **FR-008**: The CLI MUST offer a command that shows one station by UUID.
- **FR-009**: The CLI MUST offer commands that list countries, languages and tags; the tag command MUST
  accept a name filter, page size, offset and fetch-all.
- **FR-010**: Every radio command MUST validate page size (1–100) and offset (not negative) before any
  request, and map hub answers to exit codes: 400 is a usage error with the hub's message, 404 is not
  found (station) or "no radio extension", 502 and 503 are API errors, 503 with its retry delay.
- **FR-011**: Every text value from the catalogue MUST be printed with control and escape characters
  neutralised.
- **FR-012**: The radio commands MUST follow the existing conventions: verb-first command shape, strict
  GNU flag style, YAML by default and JSON with `--json`, errors on standard error, `--verbose`,
  `--hub-url`, an entry in `help`, and the existing exit codes.
- **FR-013**: The radio commands MUST be always present; no request is made at start-up to find out
  whether the hub has the extension.
- **FR-014**: Every request and response shape MUST be validated against `api/openapi.json`, refreshed if
  it lacks any radio shape this feature uses.
- **FR-015**: The README MUST document the radio commands and state that they need the hub's radio
  extension 0.2.0 or later.

### Key Entities

- **Station**: one catalogue entry: UUID, playable address, name, favicon and homepage addresses,
  country code, languages, tags, codec, bitrate, whether it is HLS, whether the catalogue's last check
  succeeded, votes and recent click count. Most fields may be absent.
- **Station page**: a list of stations and the offset of the next page, absent on the last page.
- **Country**, **Language**: a catalogue value and how many stations carry it.
- **Tag** and **Tag page**: a tag with its station count, and a paged list of them.
- **Radio problem**: a radio error's kind, title, detail, offending parameter and retry delay.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: An operator goes from a station name to hearing that station with two commands (search,
  play) and no web browser.
- **SC-002**: A fetch-all search prints no UUID twice, verified against a test server that repeats
  stations across pages.
- **SC-003**: Against a hub without the radio extension, 100% of radio commands give the same "no radio
  extension" message and exit code.
- **SC-004**: Catalogue text containing escape sequences cannot change the terminal's state, verified by
  an automated test with such text.
- **SC-005**: No radio command produces sound or changes any speaker, output, group, input or route;
  verified by running them against the live hub with nothing new playing afterwards.
- **SC-006**: Each failure listed in the acceptance scenarios and edge cases gives a distinct message and
  a non-zero exit code, covered by an automated test.

## Assumptions

- **Command layout**: the constitution requires verb-first commands, so radio uses nouns under existing
  verbs rather than a `radio` group: for example `sonora search stations [name] [flags]`,
  `sonora get stations/<uuid>`, `sonora get countries|languages|tags`. The exact words are settled in
  planning.
- **Play by UUID only**: the search output prints the playable address, which `sonora play` already
  accepts. Playing a search result by position or by name, and a `play --station` shorthand, are out of
  scope: they hide which of several same-named stations plays, and `play` takes the address as an
  argument, not a flag.
- **Always present**: radio commands exist on every hub and fail with the "no radio extension" message
  where the extension is missing, which costs no extra request (Principle I).
- **Test fixtures** are hand-written from `api/openapi.json` and checked against it, not copied from the
  extension's test resources.
- Countries and languages are short lists returned whole; they take no paging options.
- Releasing a tag and adding the tools to sonora-mcp follow this feature as separate work.
