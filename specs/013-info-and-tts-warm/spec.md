# Feature Specification: Hub Info and TTS Cache Warm-up

**Feature Branch**: `013-info-and-tts-warm`

**Created**: 2026-11-06

**Status**: Draft

**Input**: User description: "Add two hub client features for hub 0.1.33 (API 0.1.33): hub info (`GET /api/v2/info`) and TTS cache warm-up (`POST /api/tts/cache/warm`), each as a function in the public `hub` package and as a CLI command."

## Context

`api/openapi.json` was refreshed to hub 0.1.33 on 2026-11-06. It adds two operations the CLI does not use:

- **Hub info.** A read-only call that touches no device. It identifies the hub (`product` is
  `multiroom`), reports the hub `version` and `apiVersion`, when it started (`startedAt`) and how long it
  has been up (`uptimeSeconds`). A changed `startedAt` means the hub restarted, which drops every playing
  route and the on/off overrides. A hub older than 0.1.33 answers 404. The spec declares the response as
  `HubInfoResponse` with these five fields.
- **TTS cache warm-up.** Synthesizes a phrase and stores the audio without playing, queueing or pinning
  it, so a later announcement of the same text and voice is served from the cache with no provider call.
  A phrase already stored costs nothing. The result says whether the audio was already stored
  (`cacheHit`). It is part of the hub's optional TTS extension and is not rate limited.

The users are people running `sonora` interactively or from scripts, and developers of sonora-mcp, which
uses this module's public `hub` package.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - See which hub I am talking to (Priority: P1)

An operator runs one command and sees the hub's product, version, API version, start time and uptime.
A script uses it to confirm the target is a Sonora Multiroom hub, to check it supports a feature, or to
notice that the hub restarted since the last look.

**Why this priority**: it is read-only, needs no TTS extension and is the first thing anyone checks when
something is wrong.

**Independent Test**: run it against a hub 0.1.33 and compare the output with the hub's response.

**Acceptance Scenarios**:

1. **Given** a hub 0.1.33, **When** the operator asks for hub info, **Then** product, version, API
   version, start time and uptime are shown.
2. **Given** `--json`, **When** the operator asks for hub info, **Then** the output is one JSON object
   with the same fields.
3. **Given** a hub older than 0.1.33, **When** the operator asks for hub info, **Then** the command fails
   with a message that says the hub does not offer this call, not a generic error.
4. **Given** the hub is unreachable, **When** the operator asks for hub info, **Then** the command fails
   with the same unreachable-hub error as every other command.

---

### User Story 2 - Warm the TTS cache before it is needed (Priority: P2)

An operator or script has phrases it will announce often (a doorbell, "dinner is ready"). It asks the hub
to synthesize them ahead of time, so the first real announcement starts at once and costs no provider
call.

**Why this priority**: saves time and provider cost, but only for installations with the TTS extension.

**Independent Test**: warm a phrase with the free local provider, then speak the same phrase and voice
and see `cacheHit: true`.

**Acceptance Scenarios**:

1. **Given** a phrase not in the cache, **When** the operator warms it, **Then** the command succeeds and
   reports that the phrase was synthesized (`cacheHit: false`), and nothing is played.
2. **Given** a phrase already in the cache, **When** the operator warms it again, **Then** the command
   succeeds and reports `cacheHit: true`.
3. **Given** provider, voice and language options, **When** the operator warms a phrase with them,
   **Then** the hub receives them, and a later announcement with the same options is a cache hit.
4. **Given** the text is `-`, **When** the operator warms, **Then** the text is read from standard input,
   as `speak` does.
5. **Given** `--json`, **When** the operator warms a phrase, **Then** the output is one JSON object with
   `cacheHit`.
6. **Given** an unknown provider or empty text, **When** the operator warms, **Then** the command fails as
   a usage error that shows the hub's message.
7. **Given** the provider fails or the audio cannot be stored, **When** the operator warms, **Then** the
   command fails with the hub's message and says that repeating the request is safe.
8. **Given** a hub without the TTS extension, **When** the operator warms, **Then** the command reports
   TTS as unavailable, as `speak` does.

---

### User Story 3 - Use both from Go code (Priority: P2)

A developer of sonora-mcp (or any Go program) calls the public `hub` package to get hub info and to warm
a phrase, and gets typed results and the package's usual error classification.

**Why this priority**: sonora-mcp can only expose these once the client has them. A hub client fix goes
into this repository first.

**Independent Test**: the contract tests call each function against a test server and check the request
and the decoded result.

**Acceptance Scenarios**:

1. **Given** a server answering with the hub info shape, **When** the function is called, **Then** a typed
   record with each field is returned, and marshalling it gives the hub's property names.
2. **Given** a warm request with all eight spec fields, **When** the function is called, **Then** only
   the fields that were set are sent, and `cacheHit` is returned.
3. **Given** a 400, 404 or 503 answer, **When** either function is called, **Then** the error is one
   `hub.ClassifyError` already understands.

### Edge Cases

- A 200 body for hub info that is not JSON or lacks `product` is a malformed-response error, not an empty
  record.
- Hub info shows `product` as the hub sent it; a value other than `multiroom` is shown, not hidden or
  rejected.
- Warm text that is only whitespace: the hub decides; the CLI rejects only empty text.
- A hub whose `startedAt` or `uptimeSeconds` is missing still decodes; the missing field is shown as
  absent, not as zero.
- Warm-up takes as long as the provider takes; the same default time bound and `--timeout` as `speak`
  apply.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The public `hub` package MUST offer a function that calls `GET /api/v2/info` and returns a
  typed record of `product`, `version`, `apiVersion`, `startedAt` and `uptimeSeconds`.
- **FR-002**: The CLI MUST offer a command that shows hub info as YAML by default and as one JSON
  object with `--json`.
- **FR-003**: A 404 from the info call MUST be reported as "this hub does not offer hub info (needs hub
  0.1.33 or later)".
- **FR-004**: The public `hub` package MUST offer a function that calls `POST /api/tts/cache/warm` with
  `text` and the optional `providerName`, `voice`, `language`, `engine`, `pitch`, `speakingRate` and
  `stylePrompt`, sending only the fields that are set, and returns `cacheHit`.
- **FR-005**: The CLI MUST offer a command that warms one phrase, taking the text as an argument or `-`
  for standard input, with `--provider`, `--voice`, `--language` and `--timeout` meaning what they mean
  for `speak`.
- **FR-006**: The warm command MUST print `cacheHit` as YAML by default and as one JSON object with
  `--json`, and MUST NOT play anything.
- **FR-007**: The warm command MUST map hub answers as `speak` does: 400 is a usage error with the hub's
  message, a missing TTS extension is "TTS unavailable", and 503 is a failure whose message says that
  repeating is safe.
- **FR-008**: Both new commands MUST follow the existing conventions: strict GNU flag style, exit codes,
  errors on standard error, output on standard output, and an entry in `help`.
- **FR-009**: Every request and response shape MUST be validated against `api/openapi.json`.
- **FR-010**: The README MUST document both commands and state that they need hub 0.1.33 or later.

### Key Entities

- **Hub info**: identity and age of a running hub: `product`, `version`, `apiVersion`, `startedAt`,
  `uptimeSeconds`.
- **Warm request**: the phrase and the voice choices that make up a cache entry's identity.
- **Warm result**: whether the audio was already stored (`cacheHit`).

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: An operator learns a hub's version and whether it restarted since the last check with one
  command and no knowledge of the REST API.
- **SC-002**: After warming a phrase, announcing the same phrase with the same provider, voice and
  language reports a cache hit in 100% of tries with the local provider.
- **SC-003**: Warming a phrase that is already stored makes no provider call and succeeds.
- **SC-004**: Neither command produces sound or changes any speaker, output, group or route; verified by
  running both against the live hub with nothing playing afterwards.
- **SC-005**: Each failure listed in the acceptance scenarios gives a distinct message and a non-zero
  exit code, covered by an automated test.

## Assumptions

- The CLI warm command takes the same voice options as `speak` (`--provider`, `--voice`, `--language`),
  so a warmed phrase matches the announcements the CLI makes. `engine`, `pitch`, `speakingRate` and
  `stylePrompt` are in the `hub` function only; a CLI flag for each can follow when someone needs it.
- Warming several phrases is a shell loop for now; a batch form is out of scope.
- Command spelling follows the existing verb-first style, for example `sonora get info` and
  `sonora warm tts-cache <text|->`; the exact words are settled in planning.
- `Group.system` is already mirrored (separate change) and is not part of this feature.
- The hub info schema has no `required` list; the client requires `product`, `version` and `apiVersion`
  because a body without them does not identify a hub.
- Hubs before 0.1.33 keep working for every other command; only the two new commands need 0.1.33.
