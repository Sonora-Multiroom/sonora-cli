# Feature Specification: Text-to-Speech Announcements (`speak`, TTS cache)

**Feature Branch**: `009-tts-commands`

**Created**: 2026-09-21

**Status**: Draft

**Input**: User description: "@api/openapi.json contains new endpoints - extension TTS (text-to-speech) and out CLI should also now provide this; @/d:/projects-sonora/sonora-tts/specs/001-tts-extension/contracts/tts-rest-api.yaml @/d:/projects-sonora/sonora-tts/specs/001-tts-extension/spec.md"

## Overview

The Multiroom Audio Hub can now load an optional Text-to-Speech (TTS) extension. When it is
present, the hub accepts a piece of text and a target (one output or an output group),
synthesizes speech with a configured provider, and plays it on the target — queued behind
any announcement already playing there. The extension also keeps an on-disk cache of
synthesized audio that operators can inspect and clear.

The Sonora Multiroom CLI does not yet expose any of this. This feature adds three
capabilities to the CLI, each mapping to exactly one hub operation:

| Capability | Hub operation |
| --- | --- |
| Speak an announcement on an output or group | `speak` |
| Show TTS cache statistics | `getCacheStats` |
| Clear the TTS cache (all, or one provider's entries) | `clearCache` |

The TTS operations are served by the extension, not by the hub's core v2 API. They differ
from every command the CLI has today in two user-visible ways that this spec must cover:
they may be **absent** on a hub where the extension is not installed or is switched off, and
they report failures in their **own error shape** — a machine-readable code plus a message —
rather than the core API's problem-details shape.

The CLI also calls the hub's extension inventory (`listExtensions`), but only after a TTS
operation answers "not found". It uses the inventory to explain why TTS is unavailable and
never calls it on the success path.

## Clarifications

### Session 2026-09-21

- Q: How does the CLI find out whether the TTS extension is available? → A: Lazily. It
  calls the TTS operation directly. Only when the hub answers 404 does it call
  `listExtensions` and look up the entry with id `tts`, to tell the user *why* TTS is
  unavailable (not installed, disabled, rejected with a reason, inert, or active but
  mismatched). The success path makes no extra request. A 404 is already conclusive,
  because the extension reports unknown targets and providers as 400. So the lookup
  changes only the message, never the exit code, and if the lookup itself fails the
  command falls back to a generic "TTS not available" message.
- Q: What does a 404 from the extension inventory itself mean? → A: The configured hub URL
  is wrong, or the hub's control API (REST) extension is not installed or not loaded.
  Under `019-extension-shared-classloader`, that extension, not core, serves the
  inventory path. The CLI reports this as a hub-address/REST-extension problem naming the
  URL used, not as a TTS problem, and exits with the network/connectivity class, the same
  as an unreachable hub.
- Q: Should the verb be `tts` (e.g. `sonora tts "Hello" out/office`) or `speak`? → A:
  `speak`. The CLI grammar is verb first, and verbs are single words; `tts` is a noun.
- Q: Which exit codes signal "TTS not available on this hub" and "provider not found"? →
  A: TTS not available → new exit code 13; provider not found → the existing generic
  not-found code 5 (distinct from target not found, 12).
- Q: Which fields must a success response contain to be accepted? → A: All of them:
  `announcementId`, `cacheHit`, `queueDepth` for speak; `totalEntries`,
  `totalSizeBytes`, `maxSizeBytes` for cache stats. The one exception is
  `entriesByProvider`: if it is absent or null, the CLI shows an empty breakdown.
- Q: What does `clear tts-cache` print on success, given that the hub returns no body? → A: A
  structured record: `cleared: all`, or `cleared: provider` together with
  `provider: <name>`. YAML by default, JSON with `--json`.
- Q: Should `speak` read the text from standard input? → A: Yes. A lone `-` in place of the
  text means "read the text from standard input", with trailing newlines trimmed.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Announce text on a single output (Priority: P1)

A person (or a script driven by a home automation) wants a spoken announcement — "Dinner is
ready" — to play on one speaker, with a single command.

**Why this priority**: This is the whole point of the TTS extension and the smallest slice
that delivers value. Everything else in this feature is secondary to getting text spoken in
a room.

**Independent Test**: Run the speak command with a short text and the identifier of an
existing output on a hub that has the TTS extension loaded; verify the command succeeds,
prints the announcement identifier, whether the audio came from cache, and the target's
queue depth, and that the announcement is heard on that output.

**Acceptance Scenarios**:

1. **Given** a hub with the TTS extension loaded and an existing output `kitchen`,
   **When** the user runs `sonora speak "Dinner is ready" outputs/kitchen`, **Then** the
   command succeeds and displays the announcement identifier, whether the audio was served
   from cache, and the number of announcements now queued for that target.
2. **Given** the same successful request, **When** the user adds `--json`, **Then** the
   output is strict, parseable JSON carrying the same three fields.
3. **Given** a successful request, **When** the command returns, **Then** it has not waited
   for the announcement to finish playing — it returns as soon as the hub has accepted it
   (by which time the audio already exists and is queued).
4. **Given** empty or whitespace-only text, **When** the user runs the speak command,
   **Then** it fails with a clear usage error and no request is sent to the hub.
5. **Given** text piped on standard input, **When** the user runs
   `echo "Dinner is ready" | sonora speak - outputs/kitchen`, **Then** the hub receives
   `Dinner is ready` (trailing newline removed) and the command succeeds as in scenario 1.
   Empty piped input fails as in scenario 4.

---

### User Story 2 - Announce text on an output group (Priority: P2)

A person wants one announcement — "Motion detected in garden" — to play on every speaker in
a group at once.

**Why this priority**: Whole-house announcements are the next most valuable use case, and
reuse the same command with a different target kind.

**Independent Test**: Run the speak command targeting an existing group and verify the same
success output as User Story 1, with the announcement heard across the group's outputs.

**Acceptance Scenarios**:

1. **Given** an existing group `all-rooms`, **When** the user runs
   `sonora speak "Motion detected in garden" groups/all-rooms`, **Then** the command
   succeeds with the same output fields as for a single output.
2. **Given** a target written with a resource alias (`gr/all-rooms`, `out/kitchen`),
   **When** the user runs the speak command, **Then** it behaves identically to the full
   resource name, as for every other command that takes a target.
3. **Given** a target path whose resource kind is not `outputs` or `groups` (e.g.
   `routes/x`, `inputs/x`) or that has no identifier, **When** the user runs the speak
   command, **Then** it fails with a clear usage error before contacting the hub.

---

### User Story 3 - Choose provider, voice, or language per announcement (Priority: P3)

An operator has several TTS providers configured on the hub (e.g. a cloud one as default
and a local one for offline use) and wants a particular announcement to use a specific
provider, voice, or language.

**Why this priority**: Useful but optional — every announcement works with the hub's
configured defaults.

**Independent Test**: Run the speak command with each override option and confirm the hub
received the chosen value; run it with an unknown provider name and confirm a clear
"provider not found" failure.

**Acceptance Scenarios**:

1. **Given** a hub with a provider named `piper-local`, **When** the user runs the speak
   command with `--provider piper-local`, **Then** the announcement is requested from that
   provider.
2. **Given** the user supplies `--voice` and/or `--language`, **When** the speak command
   runs, **Then** those values are sent to the hub; omitted ones fall back to the
   provider's defaults on the hub.
3. **Given** a provider name the hub does not know, **When** the user runs the speak
   command, **Then** it fails with a clear "provider not found" error, distinct from
   "target not found".

---

### User Story 4 - Inspect the TTS cache (Priority: P3)

An operator wants to see how much disk the TTS cache uses, against its configured limit,
and how the entries split across providers.

**Why this priority**: Observability for the extension; not needed to make announcements.

**Independent Test**: Run the cache-statistics command against a hub with the TTS extension
loaded and verify it prints total entries, total size, maximum size, and per-provider
entry counts, in YAML by default and JSON with `--json`.

**Acceptance Scenarios**:

1. **Given** a hub with the TTS extension loaded, **When** the user runs
   `sonora get tts-cache`, **Then** the command prints the total number of cached entries,
   their total size, the configured maximum size, and the entry count per provider.
2. **Given** an empty cache, **When** the user runs the same command, **Then** it succeeds
   and shows zero entries, zero size, and an empty per-provider breakdown (not an error and
   not a missing field).

---

### User Story 5 - Clear the TTS cache (Priority: P4)

An operator wants to force fresh synthesis — for example after changing a provider's voice
settings — by clearing the cache entirely or only for one provider.

**Why this priority**: An administrative action needed rarely.

**Independent Test**: Run the cache-clear command with and without a provider option and
verify that the hub reports success and that the statistics command afterwards reflects
the cleared entries.

**Acceptance Scenarios**:

1. **Given** a populated cache, **When** the user runs `sonora clear tts-cache`, **Then**
   the whole cache is cleared and the command prints `cleared: all`.
2. **Given** a populated cache, **When** the user runs
   `sonora clear tts-cache --provider openai`, **Then** only that provider's entries are
   cleared and the command prints `cleared: provider` and `provider: openai`.
3. **Given** a provider name the hub does not know, **When** the user runs the clear
   command with that `--provider`, **Then** it fails with a clear "provider not found"
   error and nothing is cleared.
4. **Given** an already-empty cache, **When** the user clears it, **Then** the command
   still succeeds (clearing is idempotent).

---

### Edge Cases

- **TTS extension not installed, disabled, rejected, or inert on the hub**: the TTS
  operations do not exist on that hub and answer 404. Every TTS command MUST fail with a
  clear message saying the hub does not offer text-to-speech and, where the extension
  inventory can tell, why. It MUST NOT show a generic "not found" that could be mistaken
  for an unknown output or group (FR-010).
- **Extension inventory also answers 404**: the hub serves the inventory from its control
  API (REST) extension, not from core. So a 404 there means the configured address is not
  serving the hub API at all: the hub URL points somewhere else, or the REST extension is
  missing or failed to load. The CLI says so, names the URL it used, and exits with the
  network/connectivity class instead of blaming TTS.
- **Extension inventory unreachable or broken for any other reason** (network failure,
  5xx, malformed body): the command still reports "TTS not available on this hub", with no
  reason, and keeps the same exit class.
- **Inventory lists `tts` as active, but the TTS operation answered 404**: the hub's
  extension and this CLI disagree on the TTS API. The CLI reports it as a version
  mismatch, not as "not installed".
- **Unknown target output or group**: the hub rejects the request with its
  `TARGET_NOT_FOUND` code; the CLI reports "target not found", with the same exit class the
  CLI already uses for an unknown playback target.
- **Text longer than the hub allows**: the maximum length is configurable on the hub
  (default 500 characters), so the CLI does not enforce a limit of its own; it reports the
  hub's `INVALID_REQUEST` rejection and message as a validation error.
- **Provider timeout, rate limit, provider error, or unconvertible audio**: the hub answers
  "service unavailable" with one of `PROVIDER_TIMEOUT`, `PROVIDER_RATE_LIMITED`,
  `PROVIDER_ERROR`, `FORMAT_NORMALIZATION_FAILED`. The CLI reports these as a
  service-unavailable failure, shows the hub's message, and exposes the code so scripts
  can tell them apart. The CLI does not retry automatically.
- **Slow synthesis**: the hub synthesizes before it answers, and its default provider
  timeout is 10 seconds. The speak command MUST wait long enough to receive the hub's own
  timeout error rather than giving up first and misreporting it as a network failure — yet
  it MUST remain bounded.
- **Error code the CLI does not recognise** (the hub adds a new one later): the command
  still fails with the correct exit class for the HTTP status and shows the code and
  message verbatim.
- **Error body in the core API's shape, or no parseable body at all** (e.g. a proxy error
  page): the command fails with a clear hub-error message rather than garbled output.
- **Malformed success body**: the command fails clearly rather than printing incomplete
  data, as for every other command. A body missing any required field (FR-005a) counts as
  malformed; an absent or null `entriesByProvider` does not, and is shown as empty.
- **Hub unreachable or not answering**: clear network error, never a hang.
- **Text beginning with a dash**: the user can still speak it (e.g. by placing it after a
  `--` end-of-flags marker) without it being parsed as a flag. A text argument that is
  exactly `-` always means standard input (FR-003a). To speak a lone dash, pipe it in.
- **Standard input empty, whitespace-only, or unreadable when the text is `-`**: usage
  error, and no request is sent.
- **Positional arguments in the wrong order or extra positional arguments**: usage error.
- **`--provider` given with an empty value**: usage error before any request is sent.

## Requirements *(mandatory)*

### Functional Requirements

**Speak**

- **FR-001**: The CLI MUST provide a `speak` command of the form
  `sonora speak <text> <outputs|groups>/<id>` that requests a TTS announcement of the given
  text on the given output or output group.
- **FR-002**: The target's kind (single output vs. group) MUST be taken from the target
  path's resource prefix, and resource aliases (`out`, `gr`) MUST be accepted, exactly as
  for `play` and `transfer`. Any other resource kind, a bare resource with no id, or a
  missing target MUST be a usage error raised before contacting the hub.
- **FR-003**: The command MUST reject empty or whitespace-only text with a usage error
  before contacting the hub. It MUST NOT impose its own maximum text length; that limit
  belongs to the hub.
- **FR-003a**: When the text argument is exactly `-`, the command MUST read the text from
  standard input until end of input and trim trailing newlines (CR/LF). The text is sent
  as-is otherwise: inner line breaks are kept. FR-003's empty or whitespace-only check
  applies to the text read. If standard input cannot be read, that is a usage error
  raised before contacting the hub.
- **FR-004**: The command MUST accept optional `--provider <name>`, `--voice <id>`, and
  `--language <tag>` options, each sent to the hub only when supplied. An option supplied
  with an empty value MUST be a usage error.
- **FR-005**: On acceptance, the command MUST return immediately — without waiting or
  polling for playback to start or finish — and MUST display the announcement identifier,
  whether the audio was served from cache, and the target's queue depth, exactly as the
  hub reported them.
- **FR-005a**: Although `api/openapi.json` marks no success-response field as required,
  the CLI MUST treat these as required and reject a body missing any of them as malformed
  (hub-error class, 3): `announcementId`, `cacheHit`, `queueDepth` (speak) and
  `totalEntries`, `totalSizeBytes`, `maxSizeBytes` (cache stats). An empty
  `announcementId` counts as missing. An absent or null `entriesByProvider` MUST be shown
  as an empty breakdown, not rejected.

**Cache**

- **FR-006**: The CLI MUST provide `sonora get tts-cache`, which displays the TTS cache's
  total entry count, total size in bytes, configured maximum size in bytes, and entry
  count per provider. `tts-cache` is a singleton keyword like `master-mute`: it takes no
  id and has no `list` form.
- **FR-007**: The CLI MUST provide `sonora clear tts-cache [--provider <name>]`, which
  clears the whole TTS cache, or only the named provider's entries when `--provider` is
  given. On success (hub 204, no body), it prints a structured confirmation: `cleared: all`,
  or `cleared: provider` plus `provider: <name>`. The output is YAML by default and the same
  fields as a JSON object with `--json`. `clear` is a new verb and
  applies only to `tts-cache` in this feature; any other resource is a usage error.
- **FR-008**: Clearing an already-empty cache MUST succeed.

**Cross-cutting**

- **FR-009**: Every new command MUST support `--json`, `--hub-url`, and `--verbose` with the
  same meaning as on existing commands, and MUST default to YAML output.
- **FR-010**: When a TTS operation answers 404, the command MUST treat TTS as unavailable
  on this hub and fail with a message distinguishable from "target not found" and
  "provider not found". To explain why, it MUST then query the hub's extension inventory
  once and look up the entry whose id is `tts`:

  | Inventory result | Reported reason |
  | --- | --- |
  | no `tts` entry | TTS extension is not installed on this hub |
  | no `tts` entry, and the inventory reports `loadingEnabled: false` | TTS extension is not installed on this hub, noting that extension loading is switched off in the hub's configuration |
  | status `DISABLED` | TTS extension is installed but disabled in the hub's configuration |
  | status `REJECTED` | TTS extension failed to load, followed by the hub's rejection reason |
  | status `INERT` | TTS extension is loaded but inactive |
  | status `ACTIVE` | hub's TTS API does not match this CLI (version mismatch) |
  | any other status value | none: generic "TTS not available on this hub" |
  | inventory itself answers 404 | the address is not a Multiroom Audio Hub API: either the hub URL is wrong, or the hub's control API (REST) extension is not installed or not loaded. The message names the URL that was used and how to change it (`--hub-url`, `MULTIROOM_URL`, config file) |
  | inventory lookup fails any other way (network, 5xx, malformed body) | none: generic "TTS not available on this hub" |

- **FR-010a**: The inventory MUST NOT be queried on any path other than a 404 from a TTS
  operation, so a successful TTS command makes exactly one request. The exit class is
  "TTS not available" in every row above except one. When the inventory itself answers
  404, the fault is the hub address or the REST extension, not TTS, so the exit class is
  network/connectivity: the same class as an unreachable hub. Any other lookup failure is
  visible only with `--verbose` and does not change the exit class.
- **FR-011**: When a TTS operation fails with the extension's error shape, the CLI MUST
  show the hub's human-readable message and MUST make the machine-readable code
  (`INVALID_REQUEST`, `TARGET_NOT_FOUND`, `PROVIDER_NOT_FOUND`, `PROVIDER_TIMEOUT`,
  `PROVIDER_RATE_LIMITED`, `PROVIDER_ERROR`, `FORMAT_NORMALIZATION_FAILED`) visible in the
  error output.
- **FR-012**: Failures MUST map to exit codes as follows: `TARGET_NOT_FOUND` → target not
  found (12); `INVALID_REQUEST` → validation error (6); `PROVIDER_NOT_FOUND` → generic
  not found (5); any provider-side (503) failure → service unavailable (10); TTS not
  offered by the hub → a new "TTS not available" code, 13; hub unreachable, or the
  extension inventory answering 404 (FR-010a) → network error (4). An unrecognised code MUST fall back to the class
  implied by the HTTP status.
- **FR-013**: The speak command's wait for a response MUST be bounded, and MUST be long
  enough to receive the hub's default 10-second synthesis-timeout error rather than timing
  out first. Other TTS commands keep the CLI's standard request bound.
- **FR-014**: The CLI MUST NOT retry any TTS request automatically.
- **FR-015**: The new commands, the `tts-cache` keyword, and the `clear` verb MUST appear in
  `sonora help`, the README command table, and the CLI command landscape document.
- **FR-016**: All request and response handling MUST follow the TTS operations as published
  in the hub's API description (`api/openapi.json`); where that description is less precise
  than the extension's own contract (e.g. error codes, which it leaves as free text), the
  CLI MUST accept anything the extension's contract allows.

### Key Entities

- **Announcement request**: text, a target (output or group, by id), and optional
  provider, voice, and language overrides.
- **Announcement acceptance**: announcement identifier, cache-hit flag, and queue depth for
  the target, including this announcement.
- **TTS cache statistics**: total entries, total size, maximum size, and a
  provider-name → entry-count breakdown.
- **TTS error**: a machine-readable code plus a human-readable message, returned alongside
  a 400 (caller-fixable) or 503 (provider-side) status.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A user can trigger an announcement on any output or group with one command
  and no more than the text and a target path.
- **SC-002**: With an audio cache hit, `speak` reports acceptance within 1 second of being
  run on a local network.
- **SC-003**: When the hub's provider times out at its default 10-second limit, the user
  sees the hub's provider-timeout error — not a CLI network timeout — 100% of the time.
- **SC-004**: Every failure listed in Edge Cases produces a distinct, documented exit code
  class and a message that names the cause, verified by tests; no failure prints a raw
  program error or stack trace without `--verbose`.
- **SC-005**: Against a hub where the TTS extension is not active, 100% of TTS commands
  report "TTS not available" rather than a misleading not-found error. Whenever the hub's
  extension inventory is reachable, they also name the reason (not installed, disabled,
  rejected with its reason, inert, or version mismatch).
- **SC-005a**: A successful TTS command makes exactly one request to the hub.
- **SC-006**: Existing commands' behaviour and startup time are unchanged by this feature.

## Assumptions

- The hub identifies a TTS target by the same id the CLI already uses for outputs and
  groups (`outputs/<id>`, `groups/<id>`); the extension's `targetName` field carries that
  id.
- The TTS operations live under the extension's own paths, not the core v2 API, and
  answer 404 when the extension is not active. The extension never uses 404 itself: it
  reports unknown targets and providers as 400. So a 404 is conclusive evidence that TTS
  is unavailable.
- The TTS extension's inventory id is `tts`, the `Extension-Id` in its distributable's
  manifest. Entries are matched on this machine identifier, never on the display name,
  as the hub's feature `019-extension-shared-classloader` (FR-002, FR-030) requires of
  consumers.
- Per `019-extension-shared-classloader` FR-030 to FR-032, core contributes no HTTP API
  path. The inventory operation (`/api/v2/extensions`) is served by the control API (REST)
  extension, and the TTS operations are served by the TTS extension. That is why the two
  lookups separate the causes: a 404 from TTS alone means TTS is unavailable, and a 404
  from both means the REST extension is absent or the URL is wrong. The CLI MUST NOT fall
  back to the core-owned operational monitoring endpoint (`/actuator/extensions`), for
  three reasons: it is not in `api/openapi.json`, its shape is explicitly not a published
  contract, and as a diagnostic tool it may be reachable only from the hub's own host, so
  a remote CLI cannot rely on it.
- A speak request's response wait is bounded at 15 seconds (hub's 10-second default
  synthesis timeout plus margin). Hubs configured with a slower provider may exceed this;
  the user then gets a network-timeout failure, which is acceptable for this feature.
- The `speak` verb takes the text as its first positional argument and the target as its
  second, mirroring `play <uri> <target>`. The text may be `-` to read it from standard
  input (FR-003a).
- `clear` does not prompt for confirmation: clearing the cache loses no user data, only
  costs a re-synthesis on the next request.
- A user-facing command for the extension inventory (e.g. `get extensions`) and the
  legacy `/api/**` (non-`v2`) operations present in `api/openapi.json` are out of scope.
  This feature uses `listExtensions` only internally, for FR-010's diagnosis.
- Queue inspection, cancelling a queued announcement, and waiting for playback to finish
  are out of scope — the hub exposes no operation for them.
- No authentication is added: the TTS operations are guarded exactly as the rest of the
  hub's HTTP surface, which the CLI already handles.
