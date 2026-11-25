# Feature Specification: Radio Station Search and Browse

**Feature Branch**: `014-radio-station-search`

**Created**: 2026-11-21

**Status**: Completed

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

## Clarifications

### Session 2026-11-21

- Q: How should the radio commands be named on the command line? → A: Verb first, with nouns that say
  "radio": `sonora search radio-stations [name]`, `sonora get radio-stations/<uuid>`,
  `sonora get radio-countries`, `sonora get radio-languages`, `sonora get radio-tags`.
- Q: What does a command report when the radio extension is not loaded or not active? → A: As for TTS:
  after a 404 that is not "station not found", one check of the hub's extension inventory names the
  reason (not installed, disabled, failed to load, inactive, too old for search, or wrong hub URL).
- Q: Should default (YAML) search output show every station field or a short summary? → A: Every
  field, the same in YAML and JSON, as other `get` commands do.
- Q: Is a fetch-all option (`--all`) needed? → A: No. Its use, listing everything to `grep` it, is what
  the name filter (any part, any letter case) does in one request; it is dropped from search and tags.

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
   matching stations are listed with every field the hub returns, including name, codec, bitrate,
   country, UUID and playable address.
2. **Given** no search terms, **When** the operator searches, **Then** the hub's default page is shown
   (the 20 most voted stations).
3. **Given** filters for tag, country, language, codec and minimum bitrate, **When** the operator
   searches, **Then** each filter that was given is sent to the hub and only those.
4. **Given** a sort field and direction, **When** the operator searches, **Then** results come back in
   that order (name, votes, click count or bitrate; ascending or descending).
5. **Given** a page size, **When** the operator searches, **Then** at most that many stations are shown;
   when more exist the output carries the offset to pass for the next page.
6. **Given** part of a name in any letter case ("kiss"), **When** the operator searches, **Then** every
   station whose name contains it is matched, as a `grep -i` over a full list would.
7. **Given** `--json`, **When** the operator searches, **Then** the output is JSON with the same fields
   as the default YAML.
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
   a validation error (exit 6) that shows the hub's message.

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
   shown; a page size and an offset work as for station search.
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
   error is a typed "radio not offered" error, and the diagnosis turns it into a "radio is not
   available" error with the reason from the extension inventory.

### Edge Cases

- The radio extension is not serving: every radio command fails with "radio is not available on this
  hub" and the reason the extension inventory gives: not installed (noting when extension loading is
  switched off), installed but disabled, failed to load (with the hub's reason), loaded but inactive, or
  active yet answering 404 (an extension older than 0.2.0, which can play but not search). An inventory
  that also answers 404 means the hub URL is wrong; an inventory that cannot be read leaves the plain
  "radio is not available" message.
- The catalogue is unavailable (503 with a retry delay): the command fails, says the catalogue is
  temporarily unavailable and when to retry, or "retry later" when the hub gives no readable delay. The
  catalogue answering badly (502) is reported as an upstream failure.
- A page size outside 1–100 or an offset outside 0–10000 (the hub's limits): rejected by the CLI as a
  usage error (exit 2) before any request. The hub refuses, never reduces, an out-of-range value; the
  CLI does not quietly clamp it either.
- An offset past the last result gives an empty page, not an error.
- Station names, tags and addresses are community-edited text: they are printed as plain text and cannot
  move the cursor, change colours or clear the terminal.
- Absent station fields (no codec, no country, no homepage) are shown as null, not as zero or empty
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
  "radio not offered" error. The `hub` package MUST offer a diagnosis that makes one extension-inventory
  request, matches the extension by its id `radio`, and returns a "radio is not available" error naming
  the reason listed under Edge Cases; the CLI MUST run it only after such a 404 and print its message.
- **FR-005**: The CLI MUST offer `sonora search radio-stations [name]` with options for
  tag, country, language, codec, minimum bitrate, sort field, sort direction, page size and offset.
- **FR-006**: Search and station output MUST show every station field the hub returns, including the
  UUID and the playable address that `sonora play` accepts, in YAML by default and in JSON with `--json`;
  both formats carry the same fields, and an absent field is shown as null. A single page also carries
  the next-page offset (null on the last page).
- **FR-007**: Every search and list command MUST make exactly one radio request per run; the CLI MUST
  NOT follow next-page offsets by itself.
- **FR-008**: The CLI MUST offer `sonora get radio-stations/<uuid>`, which shows one station.
- **FR-009**: The CLI MUST offer `sonora get radio-countries`, `sonora get radio-languages` and
  `sonora get radio-tags`; the tag command MUST accept a name filter, page size and offset.
- **FR-010**: Every radio command that pages MUST validate page size (1–100) and offset (0–10000)
  before any request, a failure being a usage error (exit 2). Hub answers MUST map to exit codes: 400 is
  a validation error (exit 6) with the hub's message; 404 is not found (exit 5, station) or "radio is
  not available" with its reason (exit 15); 502 is an upstream failure (exit 9); 503 is temporarily
  unavailable (exit 10) with its retry delay.
- **FR-011**: Every text value from the catalogue MUST be printed with control and escape characters
  neutralised.
- **FR-012**: The radio commands MUST follow the existing conventions: verb-first command shape, strict
  GNU flag style, YAML by default and JSON with `--json`, errors on standard error, `--verbose`,
  `--hub-url`, an entry in `help`, and the existing exit codes.
- **FR-013**: The radio commands MUST be always present; no request is made at start-up or before the
  radio request to find out whether the hub has the extension. The inventory check happens only after a
  failed radio request.
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
- **SC-002**: Finding a station by part of its name takes one command and one request to the hub,
  whatever the size of the catalogue.
- **SC-003**: Against a hub where the radio extension is not serving, 100% of radio commands give the
  same exit code and a message naming the reason; each reason listed under Edge Cases has its own
  message, covered by an automated test.
- **SC-004**: Catalogue text containing escape sequences cannot change the terminal's state, verified by
  an automated test with such text.
- **SC-005**: No radio command produces sound or changes any speaker, output, group, input or route;
  verified by running them against the live hub with nothing new playing afterwards.
- **SC-006**: Each other failure listed in the acceptance scenarios and edge cases gives a distinct
  message and a non-zero exit code, covered by an automated test.

## Assumptions

- **Command layout**: the constitution requires verb-first commands, so radio uses `radio-` nouns under
  verbs rather than a `radio` group (see Clarifications). Search gets its own `search` verb because a
  bare `get radio-stations` that returns the 20 most voted would surprise; the prefix makes the
  extension dependency plain in help, so no extra help note is needed.
- **Play by UUID only**: the search output prints the playable address, which `sonora play` already
  accepts. Playing a search result by position or by name, and a `play --station` shorthand, are out of
  scope: they hide which of several same-named stations plays, and `play` takes the address as an
  argument, not a flag.
- **Always present**: radio commands exist on every hub and explain a missing extension only after the
  radio request fails, so a working hub costs one request (Principle I).
- **Test fixtures** are hand-written from `api/openapi.json` and checked against it, not copied from the
  extension's test resources.
- Countries and languages are short lists returned whole; they take no paging options.
- **No fetch-all**: the backlog's `--all` (follow every page, drop repeated UUIDs) is out of scope. Its
  use, listing everything to `grep` it, is what the name filter already does in one request; paging
  stays available through `--offset`, and a Go caller pages with the `hub` package.
- Releasing a tag and adding the tools to sonora-mcp follow this feature as separate work.
