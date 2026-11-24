# Main Project Specification

> **Revision**: 2026-11-23 — Seeded from .specify/archive/001-list-outputs (first archival): user stories 1–3, edge cases, FR-001–FR-013, Output entity, SC-001–SC-005, AS-001–AS-005.
> **Revision**: 2026-11-23 — Archived .specify/archive/002-outputs-get: added User Stories 4–6, an edge case, FR-014–FR-017, SC-006–SC-007 and AS-006; folded its equivalent requirements, edge cases, outcomes, assumptions and the Output entity into existing entries (FR-004–FR-013, SC-002–SC-005, AS-001–AS-004), widening FR-004, FR-010, FR-011 and SC-005 to cover the get command and the not-found case.
> **Revision**: 2026-11-23 — Archived .specify/archive/003-inputs-list-get: added User Stories 7–11, FR-018–FR-027, SC-008–SC-011, AS-007 and the Input entity. Its requirements (003 FR-009, FR-010, FR-012–FR-015 and FR-017) were folded into FR-006–FR-011 and FR-013; its outcomes (003 SC-004–SC-006) into SC-002–SC-004; its edge cases into the existing edge cases; and its assumptions into AS-001–AS-006, widening those entries to cover the inputs commands as well as the outputs ones. No existing ID was renumbered. The feature's own FR and SC numbers collided with main-memory IDs, so its new items continue above FR-017 and SC-007.
> **Revision**: 2026-11-23 — Archived .specify/archive/004-routes-list-get (moved from specs/ in the same run): added User Stories 12–16, FR-028–FR-036, SC-012–SC-016, AS-008 and the Route entity. Its FR-008, FR-009, FR-011–FR-014 and FR-016 were folded into FR-006–FR-011 and FR-013, its SC-004–SC-006 into SC-002–SC-004, and its assumptions into AS-001–AS-006, widening those entries to cover the routes commands as well; its invalid-status and playback-not-started edge cases are new. No existing ID was renumbered; the feature's own FR, SC and AS numbers collided with main-memory IDs, so its new items continue above FR-027, SC-011 and AS-007.
> **Revision**: 2026-11-23 — Archived .specify/archive/005-groups-list-get (moved from specs/ in the same run): added User Stories 17–21, FR-037–FR-046, SC-017–SC-021, AS-009 and the Output Group entity. Its FR-009, FR-010, FR-012–FR-015 and FR-017 were folded into FR-006–FR-011 and FR-013, its SC-004–SC-006 into SC-002–SC-004, and its assumptions into AS-001–AS-006, widening those entries to cover the groups commands as well; its no-identifier, unreachable, timeout, hub-error and malformed-response edge cases were folded into the existing edge cases, and its zero-member-outputs edge case is new. No existing ID was renumbered; the feature's own FR and SC numbers collided with main-memory IDs, so its new items continue above FR-036 and SC-016.
> **Revision**: 2026-11-23 — Archived .specify/archive/006-play-command (moved from specs/ in the same run): added User Stories 22–25, FR-047–FR-056, SC-022–SC-025, AS-010–AS-013 and the Playback Request and Playback Result entities. Its FR-007, FR-008, FR-010, FR-011 and FR-012 were folded into FR-006–FR-009, FR-011 and FR-013, widening those entries to cover the play command as well; its SC-005 and SC-006 into SC-003 and SC-004; its assumptions AS-001–AS-004 into the existing entries; and its unreachable, timeout and malformed-response edge cases into the existing ones. Its nine failure edge cases are new. No existing ID was renumbered; the feature's own FR, SC and AS numbers collided with main-memory IDs, so its new items continue above FR-046, SC-021 and AS-009.
> **Revision**: 2026-11-23 — tinyspec help-command (.specify/archive/tiny/2026-08-26-help-command/help-command.md): added User Story 26 and FR-057–FR-058 for the usage help. `help`, `-h`, `--help` and bare invocation print the usage text to stdout and exit `0`; the zero-argument invocation, previously a usage error (exit `2`), is now help. FR-058 keeps the missing-verb and unknown-command usage errors unchanged.
> **Revision**: 2026-11-23 — Archived .specify/archive/007-refactor-cli-commands (moved from specs/ in the same run): added User Stories 27–29, FR-059–FR-070, SC-026–SC-029, AS-014–AS-017, six edge cases and the Resource Kind and Resource Path entities. The feature replaces the noun-verb read commands and `play`'s `--group`/`--output`/`--name` with verb-first `get`/`list <resource>[/<id>]` and `play <uri> <outputs|groups>/<id> --display-name`. On the user's confirmation, FR-015, FR-023, FR-033, FR-042 (get with no identifier is a usage error), FR-049, FR-050, FR-051, FR-052 and AS-012 were removed and their IDs retired (see changelog.md); their replacements are FR-060 (for FR-015, FR-023, FR-033, FR-042) and FR-067 (for FR-050, FR-052); FR-049, FR-051 and AS-012 have none. No incoming item folded into an existing entry. The feature's own FR, SC and AS numbers collided with main-memory IDs, so its new items continue above FR-058, SC-025 and AS-013.
> **Revision**: 2026-11-23 — Archived .specify/archive/008-route-command (moved from specs/ in the same run): added User Stories 30–31, FR-071–FR-079, SC-030–SC-033, AS-018–AS-021, the Routing Request and Routing Result entities and six route edge cases. Its FR-006, FR-007 and FR-009 were folded into FR-006, FR-007 and FR-009, its FR-010 into FR-011 and its FR-011 into FR-013, and its FR-003 into FR-067, widening those entries to cover the route command; its SC-004 into SC-003 and its assumptions AS-001–AS-004 into the existing entries. No existing ID was renumbered. The feature's own FR, SC and AS numbers collided with main-memory IDs, so its new items continue above FR-070, SC-029 and AS-017; retired AS-012 is not reused.
> **Revision**: 2026-11-23 — tinyspec inputs-enable-disable-command (.specify/archive/tiny/2026-08-28-inputs-enable-disable-command/inputs-enable-disable-command.md): added User Story 32 and FR-080–FR-082. Adds `sonora enable inputs/<id>` and `sonora disable inputs/<id>` (`PUT /api/v2/inputs/{id}/enabled`), printing the updated input record. No existing ID was renumbered or rewritten.
> **Revision**: 2026-11-23 — tinyspec route-delete-command (.specify/archive/tiny/2026-08-28-route-delete-command/route-delete-command.md): added User Story 33 and FR-083–FR-084. Adds `sonora delete routes/<id>` and its exact alias `sonora stop routes/<id>` (`DELETE /api/v2/routes/{id}`), stopping playback and removing the route. No existing ID was renumbered or rewritten.
> **Revision**: 2026-11-23 — tinyspec route-transfer-command (.specify/archive/tiny/2026-08-28-route-transfer-command/route-transfer-command.md): added User Story 34 and FR-085–FR-087. Adds `sonora transfer routes/<id> <outputs|groups>/<target-id>` (`POST /api/v2/routes/{id}/transfer`), moving a playing route to a new target and printing the new route. No existing ID was renumbered or rewritten.
> **Revision**: 2026-11-23 — tinyspec groups-enable-disable-command (.specify/archive/tiny/2026-08-29-groups-enable-disable-command/groups-enable-disable-command.md): added User Story 35 and FR-088. Extends `enable`/`disable` to `groups/<id>` (`PUT /api/v2/groups/{id}/enabled`). No existing ID was renumbered or rewritten.
> **Revision**: 2026-11-23 — tinyspec groups-set-volume-command (.specify/archive/tiny/2026-08-29-groups-set-volume-command/groups-set-volume-command.md): added User Story 36 and FR-089. Adds `sonora set groups/<group-id> volume <0-100>` (`PUT /api/v2/groups/{id}/volume`). No existing ID was renumbered or rewritten.
> **Revision**: 2026-11-23 — tinyspec inputs-create-delete-command (.specify/archive/tiny/2026-08-29-inputs-create-delete-command/inputs-create-delete-command.md): added User Story 37 and FR-090–FR-091. Adds `sonora create inputs/<id> <uri> --display-name <name>` (`POST /api/v2/inputs`) and extends `delete` to `inputs/<id>` (`DELETE /api/v2/inputs/{id}`). No existing ID was renumbered or rewritten.
> **Revision**: 2026-11-23 — tinyspec master-mute-command (.specify/archive/tiny/2026-08-29-master-mute-command/master-mute-command.md): added User Story 38 and FR-092–FR-093. Adds the system-wide master mute: `sonora get master-mute`, `sonora mute all` and `sonora unmute all` (`GET`/`PUT /api/v2/master-mute`). No existing ID was renumbered or rewritten.
> **Revision**: 2026-11-23 — tinyspec outputs-enable-disable-command (.specify/archive/tiny/2026-08-29-outputs-enable-disable-command/outputs-enable-disable-command.md): added User Story 39 and FR-094. Extends `enable`/`disable` to `outputs/<id>` (`PUT /api/v2/outputs/{id}/enabled`). No existing ID was renumbered or rewritten.
> **Revision**: 2026-11-23 — tinyspec outputs-groups-mute-command (.specify/archive/tiny/2026-08-29-outputs-groups-mute-command/outputs-groups-mute-command.md): added User Story 40 and FR-095. Adds `sonora mute` / `sonora unmute` for `outputs/<id>` and `groups/<id>` (`PUT /api/v2/{outputs|groups}/{id}/mute`). No existing ID was renumbered or rewritten.
> **Revision**: 2026-11-23 — tinyspec outputs-set-volume-command (.specify/archive/tiny/2026-08-29-outputs-set-volume-command/outputs-set-volume-command.md): added User Story 41 and FR-096. Adds the `set` verb and `sonora set outputs/<output-id> volume <0-100>` (`PUT /api/v2/outputs/{id}/volume`). No existing ID was renumbered or rewritten.
> **Revision**: 2026-11-23 — tinyspec route-pause-resume-command (.specify/archive/tiny/2026-08-29-route-pause-resume-command/route-pause-resume-command.md): added User Story 42 and FR-097. Adds `sonora pause routes/<id>` and `sonora resume routes/<id>` (`PUT /api/v2/routes/{id}/pause`). No existing ID was renumbered or rewritten.
> **Revision**: 2026-11-23 — tinyspec stop-routes-command (.specify/archive/tiny/2026-09-09-stop-routes-command/stop-routes-command.md): added User Story 43 and FR-098–FR-099. Adds bulk stop: `sonora stop routes`, `sonora stop outputs/<id>` and `sonora stop groups/<id>`; `stop` is no longer a routes-only alias. No ID was renumbered. AS-019 was extended in place with a note that stopping, pausing and transferring are separate commands (FR-083, FR-085, FR-097, FR-098); its original text is unchanged.
> **Revision**: 2026-11-23 — Archived .specify/archive/009-tts-commands (moved from specs/ in the same run): added User Stories 44–49, FR-100–FR-117, SC-034–SC-041, AS-023–AS-030, the Announcement Request, Announcement Acceptance, TTS Cache Statistics and TTS Error entities, and sixteen edge cases for the `speak`, `get tts-cache` and `clear tts-cache` commands. Its FR-002 was folded into FR-067 and its FR-014 into FR-009, widening those entries to cover the speak and TTS cache commands (target path rule; no automatic retry); the shared AS-002 and AS-022 were widened to cover the TTS operations (no authentication; no confirmation prompt); and its malformed-body and unreachable-hub cases were folded into the existing edge cases. No existing ID was renumbered; the feature's own FR, SC and AS numbers collided with main-memory IDs, so its new items continue above FR-099, SC-033 and AS-022.
> **Revision**: 2026-11-23 — the TTS "Unknown target output or group" edge case now states exit `12`, matching FR-113 and the spec's clarification; it had said the playback target's class, which is `5`.
> **Revision**: 2026-11-23 — tinyspec list-tts-voices-command (.specify/archive/tiny/2026-09-27-list-tts-voices-command/list-tts-voices-command.md): added User Story 50, three edge cases, FR-118–FR-121, the TTS Voice entity, SC-042 and AS-031. Adds `sonora list tts-voices --provider <name> [--language] [--engine]` (`GET /api/tts/providers/{providerName}/voices`); `tts-voices` is a collection keyword under `list`, and `get tts-voices` is a usage error. The hub lists voices only for `google-cloud` and `google-gemini` entries.
> **Revision**: 2026-11-23 — Archived .specify/archive/010-public-hub-package (moved from specs/ in the same run): added User Stories 51–54, FR-122–FR-136, SC-043–SC-049, AS-032–AS-037, six edge cases, and the Public Hub Client, Error Class, Exit-Code Mapping, Published Hub Spec and Release Tag entities. The feature publishes the Multiroom Audio Hub client as the public `hub/` package (moved from `internal/hub`), publishes the spec as `api.Spec`, and moves exit-code mapping into `internal/cli/exitcode`; the CLI's behaviour is unchanged. No existing entry was folded or retired. The feature's own FR, SC and AS numbers collided with main-memory IDs, so its new items continue above FR-121, SC-042 and AS-031, and its user stories above User Story 50.
> **Revision**: 2026-11-23 — tinyspec decode-route-refusals-409 (.specify/archive/tiny/2026-10-07-decode-route-refusals-409/decode-route-refusals-409.md): added FR-137 (a hub 409 on route create, transfer or play is decoded and exits `14`) and FR-138 (`Route` gains `joinMode`/`outputs`; a conflict error class). Amended three existing entries to match the shipped behaviour: FR-079 (route exit codes gain `14`), FR-085 (a transfer keeps the route identifier) and FR-090 (a duplicate-ID 409 exits `14`, not `3`). No user story, success criterion, edge case or assumption changed.
> **Revision**: 2026-11-23 — Archived .specify/archive/011-hub-0128-alignment (moved from specs/ in the same run): added User Stories 55–57, FR-139–FR-154, SC-050–SC-054, AS-038–AS-044, the Route Refusal entity and eight edge cases. The feature documents what enable/disable do to running playback (hub 0.1.25), shows the play response's `displayName` (hub 0.1.28), and makes `sonora transfer` retry the transient `ROUTE_CHANGING_OUTPUTS` refusal. On the user's confirmation, the Playback Result entity was redefined in place: "renders exactly four" became four fields plus `displayName` second when non-empty (no ID to retire; see changelog.md). No other entry was folded or retired. The feature's own FR, SC and AS numbers collided with main-memory IDs, so its new items continue above FR-138, SC-049 and AS-037, and its user stories above User Story 54.
> **Revision**: 2026-11-23 — Resolved the partial contradictions left by the 011 archival, on the user's request: User Stories 32, 35 and 39 now say what disabling (and enabling an output) does to running playback since hub 0.1.25; FR-137 names `ROUTE_CHANGING_OUTPUTS` as the exception to "the hub's `detail` alone"; AS-022 names `transfer`'s retry as the exception to "one request". Each gained a ref to .specify/archive/011-hub-0128-alignment; no ID changed.
> **Revision**: 2026-11-23 — tinyspec rename-inputs-command (.specify/archive/tiny/2026-10-22-rename-inputs-command/rename-inputs-command.md): added User Story 58, FR-155–FR-160, SC-055 and AS-045. Adds `sonora rename inputs/<id> <name>` and `sonora rename inputs/<id> --reset` (`PUT /api/v2/inputs/{id}/display-name`, hub 0.1.27), and the hub client call that sonora-mcp will use. No existing ID was renumbered or rewritten.
> **Revision**: 2026-11-24 — Archived .specify/archive/012-join-modes (moved from specs/ in the same run): added User Stories 59–61, FR-161–FR-178, SC-056–SC-060, AS-046–AS-051, eleven edge cases and the Join Mode, Route on Output and Output Main Route entities for CLI-012 (`--join-mode` on `play` and `route`, `--default-join-mode` on `create inputs`, `get outputs/<id> --include-routes`; hub 0.1.21 join modes). The Input and Route entities gained the default join mode and the route's join mode and outputs, and the Route entity's "all ten" became "all twelve". Folded: FR-018 into FR-154 (the no-dependency, no-startup-path rule now covers join modes), and AS-038 and AS-044 each gained the join-mode clause and a second source ref. No existing ID was renumbered or retired. SC-054 is partly contradicted by SC-059 (route and input records change their YAML); it is left as written and reported in the archival report.
> **Revision**: 2026-11-24 — tinyspec hub-route-volume-client (.specify/archive/tiny/2026-10-25-hub-route-volume-client/hub-route-volume-client.md): added FR-179–FR-182, SC-061 and AS-052 for the public `hub` package's hub 0.1.29 contract: `outputVolume` replaces `volume` on play requests, route volume and ducking level become optional fields on play, route and input records, `SetRouteVolume` is added, and the `PlaybackRequest.Volume` rename is a breaking change for the public package. No user story: the change has no command of its own (the CLI follows in route-volume-cli). No existing ID was renumbered or rewritten; the Input, Route and Playback Request entities gained the new fields.
> **Revision**: 2026-11-24 — tinyspec route-volume-cli (.specify/archive/tiny/2026-10-25-route-volume-cli/route-volume-cli.md): added User Story 62, FR-183–FR-187, SC-062 and AS-053. `sonora play` replaces `--volume` with `--output-volume` and gains `--route-volume` and `--ducking-level`; `sonora set routes/<id> volume <0-100>` is new; the single-record views show the route and input levels when set. User Story 29 and the join-mode edge case named `--volume` and now say `--output-volume`; User Story 24 and FR-053 call the starting volume generically and are unchanged. FR-096's `set` verb now also takes `routes/<id>` through FR-185. No existing ID was renumbered or rewritten.
> **Revision**: 2026-11-24 — tinyspec hub-tts-unavailable-diagnosis (.specify/archive/tiny/2026-10-28-hub-tts-unavailable-diagnosis/hub-tts-unavailable-diagnosis.md): added FR-188–FR-189 and SC-063. The TTS 404 diagnosis of FR-110 is now public in the `hub` package as `DiagnoseTTSUnavailable`, and a TTS 404 classifies as text-to-speech unavailable rather than a network failure (released in v0.3.3). FR-110 and FR-111 are unchanged: the CLI's output and exit codes do not change. No user story; no existing ID was renumbered or rewritten.
> **Revision**: 2026-11-24 — tinyspec tts-cache-stats-json-tags (.specify/archive/tiny/2026-10-28-tts-cache-stats-json-tags/tts-cache-stats-json-tags.md): added FR-190 and SC-064. The public `hub.TTSCacheStats` now marshals to JSON with the property names of the hub's `CacheStats` schema, so a caller such as sonora-mcp's cache tool returns the spec's names (v0.3.4 tag pending). Decoding and the CLI's output are unchanged. No user story; no existing ID was renumbered or rewritten.
> **Revision**: 2026-11-24 — tinyspec speak-request-fields (.specify/archive/tiny/2026-10-27-speak-request-fields/speak-request-fields.md): added User Story 63, FR-191–FR-193, SC-065 and AS-054. `sonora speak` gains `--playback-mode`, `--route-volume` and `--ducking-level`, and the public `hub.SpeakRequest` gains the seven hub 0.1.29 fields it lacked (sent only when set). AS-048 is amended, on the user's decision: the hub 0.1.29 `SpeakRequest` documents `playbackMode`, so `speak` exposes it; sonora-tts's own playback mode is still its own. Tag pending. No existing ID was renumbered or rewritten.

> **Revision**: 2026-11-24 — Archived .specify/archive/013-info-and-tts-warm (moved from specs/ in the same run): added User Stories 64–66, five edge cases, FR-194–FR-203, SC-066–SC-070, the Hub Info, Warm Request and Warm Result entities, and AS-055–AS-060. No existing entry was folded, superseded or retired. FR-009, AS-002 and AS-022 name the TTS and mutating commands without naming `get info` or `warm tts-cache`; they were left as they stand.

> **Revision**: 2026-11-24 — tinyspec speak-style-prompt-flag (.specify/archive/tiny/2026-11-12-speak-style-prompt-flag/speak-style-prompt-flag.md): `speak` and `warm tts-cache` gain `--style-prompt TEXT` (FR-204, FR-205, AS-061). This supersedes FR-193 and AS-055 for `stylePrompt` only; both are annotated in place, their text kept. The `hub` functions are unchanged.

> **Revision**: 2026-11-24 — tinyspec route-chime (.specify/archive/tiny/2026-11-20-route-chime/route-chime.md): adds User Story 67, FR-206–FR-213, SC-071–SC-073, the Chime entity and AS-062–AS-064: `--chime` on `play` and `route`, `list chimes` and `get chimes[/<id>]`, a `chime` or `defaultChime` field on outputs, groups and inputs, and `get`/`list` with no resource naming every resource the verb accepts. FR-213 extends FR-066's list of names; FR-066 itself is unchanged.
> **Revision**: 2026-11-24 — Archived .specify/archive/014-radio-station-search (moved to .specify/archive in the same run): added User Stories 68–71, FR-214–FR-227, SC-074–SC-079, AS-065–AS-071, the radio edge cases and the Radio Station, Radio Station Page, Radio Country, Radio Language, Radio Tag, Radio Tag Page and Radio Problem entities. Its FR-012 (conventions for new commands) is folded into FR-109; no other entry was folded, superseded or retired.

> **Revision**: 2026-11-24 — tinyspec stored-inputs (.specify/archive/tiny/2026-11-22-stored-inputs/stored-inputs.md): adds User Story 72, FR-228–FR-232, SC-080–SC-081, AS-072, and the origin, playable and unplayable-reason attributes of the Input entity (HUB-041, API 0.1.35): `create inputs` takes `--persistent`, `keep inputs/<id>` is new, and the three input 409 reasons get short messages. FR-090 and AS-064 are annotated in place, their text kept. No existing ID was renumbered.

## User Scenarios & Testing

### User Story 1 - View currently active outputs (Priority: P1)

A person operating the audio system wants to see which speakers/zones are currently
available and controllable, along with their volume and mute state, so they know what
they can act on next (e.g. before adjusting volume or routing audio somewhere).

**Why this priority**: This is the baseline read operation. Without it, a user has no way
to discover output identifiers or current state before issuing any other command — it is
the entry point to every other interaction with outputs.

**Independent Test**: Can be fully tested by running the list command against a hub with
one or more enabled outputs and verifying each enabled output's identifier, display name,
volume, mute state, and availability are shown, and disabled outputs are omitted.

**Acceptance Scenarios**:

1. **Given** the hub has two enabled outputs and one disabled output, **When** the user
   runs the list command with no flags, **Then** only the two enabled outputs are
   displayed, each showing its identifier, display name, volume, mute state, and
   availability.
2. **Given** the hub has zero enabled outputs, **When** the user runs the list command,
   **Then** the user sees a clear indication that no outputs are available rather than an
   empty or ambiguous response.

[Source: .specify/archive/001-list-outputs/spec.md -> User Story 1]

---

### User Story 2 - Include disabled outputs (Priority: P2)

A person managing the audio system (e.g. diagnosing why a speaker isn't selectable) wants
to see outputs that exist but are currently disabled, so they can distinguish "not
available for routing" from "doesn't exist."

**Why this priority**: Secondary to the default view, but necessary for troubleshooting
and administration; it reuses the same read path with one additional option.

**Independent Test**: Can be fully tested by running the list command with the
include-disabled option against a hub with at least one disabled output and verifying it
appears in the results, distinguishable from enabled outputs.

**Acceptance Scenarios**:

1. **Given** the hub has one enabled and one disabled output, **When** the user runs the
   list command with the option to include disabled outputs, **Then** both outputs are
   displayed and each output's enabled/disabled state is visible in the result.

[Source: .specify/archive/001-list-outputs/spec.md -> User Story 2]

---

### User Story 3 - Consume output list from a script (Priority: P3)

An automation or script wants to retrieve the current outputs in a strict, parseable
format so it can act on the data programmatically (e.g. feed it into another tool).

**Why this priority**: Enables automation use cases but is not required for a human
operator to get value from the command; it is an alternate presentation of the same data
as User Story 1.

**Independent Test**: Can be fully tested by running the list command with the
machine-readable output option and verifying the result is strictly parseable and contains
the same fields available in the human-readable view.

**Acceptance Scenarios**:

1. **Given** the hub has at least one enabled output, **When** the user runs the list
   command requesting machine-readable output, **Then** the result is valid, strictly
   parseable structured data containing every output's identifier, display name, volume,
   mute state, availability, and enabled state.

[Source: .specify/archive/001-list-outputs/spec.md -> User Story 3]

---

### User Story 4 - Look up a specific output by identifier (Priority: P1)

A person operating the audio system already knows (e.g. from `outputs list`) the
identifier of a specific output and wants to see its full current state — display name,
volume, mute state, availability, and enabled state — without scanning the full list.

**Why this priority**: This is the sole purpose of the command; without it the feature
delivers no value. It is the direct single-item counterpart to `outputs list`.

**Independent Test**: Can be fully tested by running the get command with the identifier
of an output known to exist on the hub and verifying the identifier, display name,
volume, mute state, availability, and enabled state are all shown correctly.

**Acceptance Scenarios**:

1. **Given** the hub has an enabled output with a known identifier, **When** the user runs
   the get command with that identifier, **Then** the command displays that output's
   identifier, display name, volume, mute state, availability, and enabled state.
2. **Given** the hub has a disabled output with a known identifier, **When** the user runs
   the get command with that identifier, **Then** the command still displays the output
   (unlike the list command's default filtering) with its disabled state clearly shown.

[Source: .specify/archive/002-outputs-get/spec.md -> User Story 1]

---

### User Story 5 - Handle a nonexistent output identifier (Priority: P2)

A person mistypes an output identifier, or refers to one that has since been removed,
and needs to know clearly that it doesn't exist rather than receiving a confusing error
or empty output.

**Why this priority**: Essential for usability once the primary lookup path works, but
secondary to the happy path — the command must be usable correctly before edge cases
matter.

**Independent Test**: Can be fully tested by running the get command with an identifier
that does not exist on the hub and verifying the user sees a clear "not found" message
and a distinct exit status.

**Acceptance Scenarios**:

1. **Given** no output with the given identifier exists on the hub, **When** the user
   runs the get command with that identifier, **Then** the user sees a clear "output not
   found" message rather than an empty or ambiguous result.

[Source: .specify/archive/002-outputs-get/spec.md -> User Story 2]

---

### User Story 6 - Consume a single output's state from a script (Priority: P3)

An automation or script wants to retrieve one output's current state in a strict,
parseable format so it can act on the data programmatically (e.g. check volume before
deciding whether to adjust it).

**Why this priority**: Enables automation use cases consistent with the rest of the CLI,
but is not required for a human operator to get value from the command; it is an
alternate presentation of the same data as User Story 4.

**Independent Test**: Can be fully tested by running the get command with the
machine-readable output option against a known output identifier and verifying the
result is strictly parseable and contains the same fields available in the
human-readable view.

**Acceptance Scenarios**:

1. **Given** the hub has an output with a known identifier, **When** the user runs the
   get command with that identifier requesting machine-readable output, **Then** the
   result is valid, strictly parseable structured data containing the output's
   identifier, display name, volume, mute state, availability, and enabled state.

[Source: .specify/archive/002-outputs-get/spec.md -> User Story 3]

---

### User Story 7 - View currently active inputs (Priority: P1)

A person operating the audio system wants to see which audio inputs (e.g. streaming
sources, ephemeral playback sessions) are currently registered and usable, so they know
what they can route to an output next.

**Why this priority**: This is the baseline read operation for inputs — without it, a
user has no way to discover input identifiers before issuing any other command involving
an input (e.g. creating a route).

**Independent Test**: Can be fully tested by running the list command against a hub with
one or more enabled inputs and verifying each enabled input's identifier, display name,
source URI, enabled state, and source type are shown, and disabled inputs are omitted.

**Acceptance Scenarios**:

1. **Given** the hub has two enabled inputs and one disabled input, **When** the user runs
   the list command with no flags, **Then** only the two enabled inputs are displayed,
   each showing its identifier, display name, source URI, source type, and enabled state.
2. **Given** the hub has zero enabled inputs, **When** the user runs the list command,
   **Then** the user sees a clear indication that no inputs are available rather than an
   empty or ambiguous response.

[Source: .specify/archive/003-inputs-list-get/spec.md -> User Story 1]

---

### User Story 8 - Look up a specific input by identifier (Priority: P1)

A person who already knows an input's identifier (e.g. from `inputs list`) wants to see
its full current state without scanning the full list.

**Why this priority**: This is the sole purpose of the get command and the direct
single-item counterpart to `inputs list`; both are equally core to enabling later
route-management commands.

**Independent Test**: Can be fully tested by running the get command with the identifier
of an input known to exist on the hub and verifying the identifier, display name, source
URI, enabled state, source type, and other attributes are all shown correctly.

**Acceptance Scenarios**:

1. **Given** the hub has an enabled input with a known identifier, **When** the user runs
   the get command with that identifier, **Then** the command displays that input's
   identifier, display name, source URI, enabled state, source type, and remaining
   attributes.
2. **Given** the hub has a disabled input with a known identifier, **When** the user runs
   the get command with that identifier, **Then** the command still displays the input
   (unlike the list command's default filtering) with its disabled state clearly shown.

[Source: .specify/archive/003-inputs-list-get/spec.md -> User Story 2]

---

### User Story 9 - Include disabled inputs in the list (Priority: P2)

A person managing the audio system wants to see inputs that exist but are currently
disabled, so they can distinguish "not currently usable" from "doesn't exist."

**Why this priority**: Secondary to the default view, but necessary for troubleshooting
and administration; it reuses the same read path with one additional option.

**Independent Test**: Can be fully tested by running the list command with the
include-disabled option against a hub with at least one disabled input and verifying it
appears in the results, distinguishable from enabled inputs.

**Acceptance Scenarios**:

1. **Given** the hub has one enabled and one disabled input, **When** the user runs the
   list command with the option to include disabled inputs, **Then** both inputs are
   displayed and each input's enabled/disabled state is visible in the result.

[Source: .specify/archive/003-inputs-list-get/spec.md -> User Story 3]

---

### User Story 10 - Handle a nonexistent input identifier (Priority: P2)

A person mistypes an input identifier, or refers to an ephemeral input that has since
been removed, and needs to know clearly that it doesn't exist rather than receiving a
confusing error or empty output.

**Why this priority**: Essential for usability once the primary lookup path works, but
secondary to the happy path.

**Independent Test**: Can be fully tested by running the get command with an identifier
that does not exist on the hub and verifying the user sees a clear "not found" message
and a distinct exit status.

**Acceptance Scenarios**:

1. **Given** no input with the given identifier exists on the hub, **When** the user runs
   the get command with that identifier, **Then** the user sees a clear "input not found"
   message rather than an empty or ambiguous result.

[Source: .specify/archive/003-inputs-list-get/spec.md -> User Story 4]

---

### User Story 11 - Consume input data from a script (Priority: P3)

An automation or script wants to retrieve the current inputs, or a single input's state,
in a strict, parseable format so it can act on the data programmatically.

**Why this priority**: Enables automation use cases consistent with the rest of the CLI,
but is not required for a human operator to get value from either command; it is an
alternate presentation of the same data as User Stories 7 and 8.

**Independent Test**: Can be fully tested by running each command with the
machine-readable output option and verifying the result is strictly parseable and
contains the same fields available in the human-readable view.

**Acceptance Scenarios**:

1. **Given** the hub has at least one enabled input, **When** the user runs the list
   command requesting machine-readable output, **Then** the result is valid, strictly
   parseable structured data containing every input's identifier, display name, source
   URI, enabled state, source type, and remaining attributes.
2. **Given** the hub has an input with a known identifier, **When** the user runs the get
   command with that identifier requesting machine-readable output, **Then** the result is
   valid, strictly parseable structured data containing that input's full attribute set.

[Source: .specify/archive/003-inputs-list-get/spec.md -> User Story 5]

---

### User Story 12 - View currently active routes (Priority: P1)

A person operating the audio system wants to see which audio routes currently exist —
each connecting a specific input to a specific output or output group — along with each
route's status, so they know what audio is currently flowing where.

**Why this priority**: This is the baseline read operation for routes — without it, a
user has no way to discover route identifiers or see the current routing state before
issuing any other command involving a route (e.g. transferring or stopping one).

**Independent Test**: Can be fully tested by running the list command against a hub with
one or more routes and verifying each route's identifier, source input identifier, target
identifier, target type, and status are shown.

**Acceptance Scenarios**:

1. **Given** the hub has two active routes, **When** the user runs the list command with
   no flags, **Then** both routes are displayed, each showing its identifier, source
   input identifier, target identifier, target type, and status.
2. **Given** the hub has zero routes, **When** the user runs the list command, **Then**
   the user sees a clear indication that no routes are currently active rather than an
   empty or ambiguous response.

[Source: .specify/archive/004-routes-list-get/spec.md -> User Story 1]

---

### User Story 13 - Look up a specific route by identifier (Priority: P1)

A person who already knows a route's identifier (e.g. from `routes list`) wants to see
its full current state without scanning the full list.

**Why this priority**: This is the sole purpose of the get command and the direct
single-item counterpart to `routes list`; both are equally core to giving a user visibility
into the routing state of the system.

**Independent Test**: Can be fully tested by running the get command with the identifier
of a route known to exist on the hub and verifying the identifier, source input
identifier, target identifier, target type, status, and remaining attributes are all
shown correctly.

**Acceptance Scenarios**:

1. **Given** the hub has a route with a known identifier, **When** the user runs the get
   command with that identifier, **Then** the command displays that route's identifier,
   source input identifier, target identifier, target type, status, creation timestamp,
   playback-started timestamp (or an explicit indication that playback has not started),
   whether it can be transferred, whether it supports pause/resume, and whether it is
   currently paused.

[Source: .specify/archive/004-routes-list-get/spec.md -> User Story 2]

---

### User Story 14 - Narrow the route list by status, input, or target (Priority: P2)

A person managing the audio system wants to see only the routes relevant to a specific
input, a specific output/group, or a specific lifecycle status (e.g. only `FAILED`
routes), so they can focus on the subset they care about without scanning the full list.

**Why this priority**: Secondary to the default view, but valuable for troubleshooting
and for scripts that only care about one input or target; it reuses the same read path
with additional optional filters.

**Independent Test**: Can be fully tested by running the list command with each filter
option in turn against a hub with routes in different states, sourced from different
inputs, and pointed at different targets, and verifying only the matching routes are
returned.

**Acceptance Scenarios**:

1. **Given** the hub has routes with different statuses, **When** the user runs the list
   command filtered to a specific status, **Then** only routes with that status are
   displayed.
2. **Given** the hub has routes from different source inputs, **When** the user runs the
   list command filtered to a specific input identifier, **Then** only routes originating
   from that input are displayed.
3. **Given** the hub has routes to different targets, **When** the user runs the list
   command filtered to a specific target identifier, **Then** only routes to that target
   are displayed.
4. **Given** the user supplies more than one filter at once, **When** the user runs the
   list command, **Then** only routes matching all supplied filters are displayed.

[Source: .specify/archive/004-routes-list-get/spec.md -> User Story 3]

---

### User Story 15 - Handle a nonexistent route identifier (Priority: P2)

A person mistypes a route identifier, or refers to a route that has since stopped and
been removed, and needs to know clearly that it doesn't exist rather than receiving a
confusing error or empty output.

**Why this priority**: Essential for usability once the primary lookup path works, but
secondary to the happy path.

**Independent Test**: Can be fully tested by running the get command with an identifier
that does not exist on the hub and verifying the user sees a clear "not found" message
and a distinct exit status.

**Acceptance Scenarios**:

1. **Given** no route with the given identifier exists on the hub, **When** the user runs
   the get command with that identifier, **Then** the user sees a clear "route not found"
   message rather than an empty or ambiguous result.

[Source: .specify/archive/004-routes-list-get/spec.md -> User Story 4]

---

### User Story 16 - Consume route data from a script (Priority: P3)

An automation or script wants to retrieve the current routes, or a single route's state,
in a strict, parseable format so it can act on the data programmatically (e.g. detect
`FAILED` routes and alert on them).

**Why this priority**: Enables automation use cases consistent with the rest of the CLI,
but is not required for a human operator to get value from either command; it is an
alternate presentation of the same data as User Stories 12 and 13.

**Independent Test**: Can be fully tested by running each command with the
machine-readable output option and verifying the result is strictly parseable and
contains the same fields available in the human-readable view.

**Acceptance Scenarios**:

1. **Given** the hub has at least one route, **When** the user runs the list command
   requesting machine-readable output, **Then** the result is valid, strictly parseable
   structured data containing every route's full attribute set.
2. **Given** the hub has a route with a known identifier, **When** the user runs the get
   command with that identifier requesting machine-readable output, **Then** the result is
   valid, strictly parseable structured data containing that route's full attribute set.

[Source: .specify/archive/004-routes-list-get/spec.md -> User Story 5]

---

### User Story 17 - View currently enabled output groups (Priority: P1)

A person operating the audio system wants to see which output groups currently exist —
each a named collection of outputs that can be routed to as a single target — along with
each group's member outputs and mute state, so they know what groups are available for
routing.

**Why this priority**: This is the baseline read operation for groups — without it, a
user has no way to discover group identifiers or see which outputs belong to a group
before issuing any command that routes audio to a group.

**Independent Test**: Can be fully tested by running the list command against a hub with
one or more enabled groups and verifying each group's identifier, display name, member
output identifiers, and mute state are shown, and disabled groups are omitted.

**Acceptance Scenarios**:

1. **Given** the hub has two enabled groups and one disabled group, **When** the user runs
   the list command with no flags, **Then** only the two enabled groups are displayed,
   each showing its identifier, display name, member output identifiers, and mute state.
2. **Given** the hub has zero enabled groups, **When** the user runs the list command,
   **Then** the user sees a clear indication that no groups are currently available rather
   than an empty or ambiguous response.

[Source: .specify/archive/005-groups-list-get/spec.md -> User Story 1]

---

### User Story 18 - Include disabled groups (Priority: P2)

A person managing the audio system wants to see groups that exist but are currently
disabled, so they can distinguish "not available for new routes" from "does not exist"
when troubleshooting.

**Why this priority**: Secondary to the default view; needed for full visibility into the
system but not required for the common case of routing to an available group.

**Independent Test**: Can be fully tested by running the list command with the
include-disabled option against a hub with at least one disabled group and verifying it
appears in the results, distinguishable from enabled groups.

**Acceptance Scenarios**:

1. **Given** the hub has one enabled and one disabled group, **When** the user runs the
   list command with the option to include disabled groups, **Then** both groups are
   displayed and each group's enabled/disabled state is visible in the result.

[Source: .specify/archive/005-groups-list-get/spec.md -> User Story 2]

---

### User Story 19 - Look up a specific group by identifier (Priority: P1)

A person who already knows a group's identifier (e.g. from `groups list`) wants to see
its full current state without scanning the full list.

**Why this priority**: This is the sole purpose of the get command and the direct
single-item counterpart to `groups list`; both are equally core to giving a user
visibility into the groups available for routing.

**Independent Test**: Can be fully tested by running the get command with the identifier
of a group known to exist on the hub, whether enabled or disabled, and verifying the
identifier, display name, member output identifiers, mute state, and enabled state are
all shown correctly.

**Acceptance Scenarios**:

1. **Given** the hub has a group with a known identifier, **When** the user runs the get
   command with that identifier, **Then** the command displays that group's identifier,
   display name, member output identifiers, mute state, and enabled state — regardless of
   whether the group is currently enabled or disabled.

[Source: .specify/archive/005-groups-list-get/spec.md -> User Story 3]

---

### User Story 20 - Handle a nonexistent group identifier (Priority: P2)

A person mistypes a group identifier, or refers to a group that has since been removed,
and needs to know clearly that it doesn't exist rather than receiving a confusing error or
empty output.

**Why this priority**: Essential for usability once the primary lookup path works, but
secondary to the happy path.

**Independent Test**: Can be fully tested by running the get command with an identifier
that does not exist on the hub and verifying the user sees a clear "not found" message and
a distinct exit status.

**Acceptance Scenarios**:

1. **Given** no group with the given identifier exists on the hub, **When** the user runs
   the get command with that identifier, **Then** the user sees a clear "group not found"
   message rather than an empty or ambiguous result.

[Source: .specify/archive/005-groups-list-get/spec.md -> User Story 4]

---

### User Story 21 - Consume group data from a script (Priority: P3)

An automation or script wants to retrieve the current groups, or a single group's state,
in a strict, parseable format so it can act on the data programmatically (e.g. pick a
group's member outputs before issuing a route command).

**Why this priority**: Enables automation use cases consistent with the rest of the CLI,
but is not required for a human operator to get value from either command; it is an
alternate presentation of the same data as User Stories 17 and 19.

**Independent Test**: Can be fully tested by running each command with the
machine-readable output option and verifying the result is strictly parseable and contains
the same fields available in the human-readable view.

**Acceptance Scenarios**:

1. **Given** the hub has at least one group, **When** the user runs the list command
   requesting machine-readable output, **Then** the result is valid, strictly parseable
   structured data containing every group's full attribute set.
2. **Given** the hub has a group with a known identifier, **When** the user runs the get
   command with that identifier requesting machine-readable output, **Then** the result is
   valid, strictly parseable structured data containing that group's full attribute set.

[Source: .specify/archive/005-groups-list-get/spec.md -> User Story 5]

---

### User Story 22 - Play a URI to a single output (Priority: P1)

A person operating the audio system wants to play an audio source (a stream URL, a
YouTube/SoundCloud link, etc.) directly to one speaker without first creating an input and
a route as two separate steps.

**Why this priority**: This is the core value of the feature — instant playback in one
command — and is called out in the API description as core functionality. Without it,
users must fall back to separate input-creation and route-creation steps that don't yet
exist as CLI commands.

**Independent Test**: Can be fully tested by running the play command with a valid URI and
the identifier of a single existing output, then verifying the command immediately reports
a created input identifier and the route's identifier and status exactly as returned by
the hub, without waiting for the route to reach a further state.

**Acceptance Scenarios**:

1. **Given** a reachable audio URI and the identifier of an existing, enabled output,
   **When** the user runs the play command targeting that output, **Then** the command
   succeeds and displays the created input identifier, the route identifier, the route
   status, and a human-readable confirmation message.
2. **Given** the same successful playback request, **When** the user requests
   machine-readable output, **Then** the result is valid, strictly parseable structured
   data containing the same fields as the human-readable view.

[Source: .specify/archive/006-play-command/spec.md -> User Story 1]

---

### User Story 23 - Play a URI to an output group (Priority: P2)

A person wants to play one audio source to a group of outputs at once (e.g. "whole
house") rather than a single speaker.

**Why this priority**: Groups are a first-class routing target already supported by
`groups list`/`groups get`; playback should work the same way against a group as against a
single output, but this is secondary to the single-output happy path.

**Independent Test**: Can be fully tested by running the play command with a valid URI and
the identifier of an existing output group, and verifying the same success output as User
Story 1, with the route's target reflecting the group.

**Acceptance Scenarios**:

1. **Given** a reachable audio URI and the identifier of an existing output group,
   **When** the user runs the play command targeting that group, **Then** the command
   succeeds and displays the created input identifier, the route identifier and status,
   and confirms the target is a group.

[Source: .specify/archive/006-play-command/spec.md -> User Story 2]

---

### User Story 24 - Set the starting volume (Priority: P3)

A person wants playback to start at a specific volume rather than whatever level the
target output or group was last left at.

**Why this priority**: A convenience on top of the core playback flow; playback is fully
usable without it since the target's existing volume applies by default.

**Independent Test**: Can be fully tested by running the play command with a volume option
set to a valid level and verifying the response confirms the requested volume was applied
before playback started.

**Acceptance Scenarios**:

1. **Given** a valid playback request with a volume option set to a value between 0 and
   100, **When** the user runs the play command, **Then** the target's volume is set to
   that value before playback begins.
2. **Given** a volume option set to a value outside 0-100, **When** the user runs the play
   command, **Then** the command fails with a clear usage/validation error before any
   request reaches the hub.

[Source: .specify/archive/006-play-command/spec.md -> User Story 3]

---

### User Story 25 - Name the playback session (Priority: P3)

A person wants the ephemeral input created for this playback to have a recognizable name
(e.g. "Kitchen Radio") instead of an auto-generated identifier, so it's easy to recognize
later in `inputs list`.

**Why this priority**: Cosmetic convenience; playback functions correctly with an
auto-generated name if this is omitted.

**Independent Test**: Can be fully tested by running the play command with a display name
option and verifying the created input's name matches what was supplied.

**Acceptance Scenarios**:

1. **Given** a valid playback request with a display name option, **When** the user runs
   the play command, **Then** the confirmation output reflects that name.

[Source: .specify/archive/006-play-command/spec.md -> User Story 4]

---

### User Story 26 - Show usage help (Priority: P3)

A person who has forgotten the command syntax runs `sonora help`, `sonora -h`, `sonora --help`
or plain `sonora` and sees the command table, the global flags and example invocations, instead
of a usage error.

**Why this priority**: Discoverability aid; every command works without it.

**Independent Test**: Can be fully tested by running each help form and checking that the output
names the noun/verb commands, the `play` command, the global flags and an example, and exits `0`.

**Acceptance Scenarios**:

1. **Given** no arguments, **When** the user runs `sonora`, **Then** the usage text is printed to
   stdout and the exit status is `0`.
2. **Given** `help`, `-h` or `--help`, **When** the user runs it, **Then** the same usage text is
   printed to stdout and the exit status is `0`.

[Source: .specify/archive/tiny/2026-08-26-help-command/help-command.md]

---

### User Story 27 - Read commands follow the verb-first, path-style shape (Priority: P1)

An operator who has learned the new `sonora route inputs/<id> outputs/<id>` command (from the separately specified `route` feature) expects every other resource — inputs, outputs, groups, and routes — to read the same way: verb first, resource path second. Reading those resources previously used the `sonora <resource> <verb>` shape (`sonora inputs list`, `sonora outputs get <id>`), which was inconsistent with `route` and with the documented command landscape.

**Why this priority**: This is the actual inconsistency prompting the refactor — the CLI spoke two different command grammars depending on which resource you were touching. Unifying the already-shipped read commands onto one grammar is the highest-value, lowest-risk piece of the refactor, since these commands don't mutate anything.

**Independent Test**: Can be fully tested by running `sonora get inputs`, `sonora get outputs`, `sonora get groups`, and `sonora get routes` (each with and without a trailing `/<id>`), and verifying each returns exactly the same data as the pre-refactor `sonora <resource> list` / `sonora <resource> get <id>` commands.

**Acceptance Scenarios**:

1. **Given** existing inputs, outputs, groups, or routes, **When** the user runs `sonora get <resource>` for any of the four resources, **Then** the command returns the same collection data as the pre-refactor `sonora <resource> list`, including support for the existing collection filter flags (`--include-disabled` for inputs/outputs/groups; `--status`, `--input-id`, `--target-id` for routes).
2. **Given** an existing resource of a known identifier, **When** the user runs `sonora get <resource>/<id>`, **Then** the command returns the same single-item data as the pre-refactor `sonora <resource> get <id>`.
3. **Given** any of the four resources, **When** the user runs `sonora list <resource>` (no id), **Then** the result is identical to `sonora get <resource>` — `list` behaves as an exact synonym for the collection form only.
4. **Given** any of the four resources, **When** the user runs `sonora list <resource>/<id>` (an id after `list`), **Then** the command fails with a clear usage error, since `list` is valid only for the collection form.
5. **Given** the old command shape, **When** the user runs `sonora inputs list`, `sonora outputs get <id>`, or any other pre-refactor invocation, **Then** the command fails with a clear usage error rather than being silently reinterpreted or producing incorrect output.

[Source: .specify/archive/007-refactor-cli-commands/spec.md -> User Story 1]

---

### User Story 28 - Resource aliases work everywhere a resource path appears (Priority: P2)

An operator who frequently types commands wants to use the short aliases (`in`, `out`, `gr`, `rt`) documented alongside the full resource names, so `sonora get out/<id>` works exactly like `sonora get outputs/<id>`.

**Why this priority**: Aliases are a documented part of the same landscape and a natural companion to User Story 27, but they're a convenience layer on top of the core grammar change, not a prerequisite for it.

**Independent Test**: Can be fully tested by running the collection and single-item forms of `get`/`list` for each resource using its alias instead of its full name, and verifying identical results to using the full name.

**Acceptance Scenarios**:

1. **Given** an existing input, output, group, or route, **When** the user substitutes the resource's short alias (`in`, `out`, `gr`, `rt` respectively) for its full name in any `get`/`list` command, **Then** the command behaves identically to using the full resource name.
2. **Given** an unrecognized resource name or alias, **When** the user runs `sonora get <unrecognized>`, **Then** the command fails with a clear usage error naming the unrecognized resource.

[Source: .specify/archive/007-refactor-cli-commands/spec.md -> User Story 2]

---

### User Story 29 - `play`'s target and flags match the new landscape (Priority: P3)

An operator using `sonora play <uri> <target-id> --output` expects `play` to address its target the same way `route` does — a path like `outputs/<id>` or `groups/<id>` — instead of a bare id plus a separate `--group`/`--output` flag, and expects the display-name flag to be named consistently with the rest of the CLI.

**Why this priority**: `play` is the one already-implemented command that mutates state, so its change carries the most user-facing risk (existing scripts calling it will break). It's sequenced last so the lower-risk read-command grammar lands and stabilizes first.

**Independent Test**: Can be fully tested by running `sonora play <uri> outputs/<id>` and `sonora play <uri> groups/<id>` (including with `--display-name` and `--output-volume`), and verifying the same playback result as the pre-refactor `sonora play <uri> <id> --output` / `sonora play <uri> <id> --group`.

**Acceptance Scenarios**:

1. **Given** a playable URI and an existing output, **When** the user runs `sonora play <uri> outputs/<output-id>`, **Then** playback starts on that output exactly as `sonora play <uri> <output-id> --output` did.
2. **Given** a playable URI and an existing group, **When** the user runs `sonora play <uri> groups/<group-id>`, **Then** playback starts on that group exactly as `sonora play <uri> <group-id> --group` did.
3. **Given** a playable URI and a target, **When** the user supplies `--display-name`, **Then** it sets the new input's display name exactly as the former `--name` flag did.
4. **Given** the old invocation, **When** the user runs `sonora play <uri> <id> --group` or `sonora play <uri> <id> --output` or `sonora play <uri> <id> --name <name>`, **Then** the command fails with a clear usage error rather than silently accepting the old flags.

[Source: .specify/archive/007-refactor-cli-commands/spec.md -> User Story 3]

---

### User Story 30 - Route an existing input to a single output (Priority: P1)

A person operating the audio system has an input that already exists (e.g. a line-in, a
previously created stream, or one created by an earlier `play`) and wants to send it to one
speaker, without creating a new input or a duplicate playback session.

**Why this priority**: This is the core value of the feature — connecting something that's
already playing (or already defined) to a destination in one command. Without it, the only
way to get audio to a target is `play`, which always mints a brand-new input even when a
perfectly good one already exists.

**Independent Test**: Can be fully tested by running the route command with the path of an
existing, enabled input and the path of a single existing output, then verifying the command
reports the resulting route's identifier and status exactly as returned by the hub, without
creating any new input.

**Acceptance Scenarios**:

1. **Given** the path of an existing, enabled input and the path of an existing, enabled
   output, **When** the user runs `sonora route inputs/<input-id> outputs/<output-id>`,
   **Then** the command succeeds and displays the route's identifier, status, and a
   human-readable confirmation message.
2. **Given** the same successful routing request, **When** the user requests
   machine-readable output, **Then** the result is valid, strictly parseable structured data
   containing the same fields as the human-readable view.
3. **Given** a successful routing request, **When** the user subsequently inspects the
   input (e.g. via the existing input inspection command), **Then** the input is unchanged —
   no new input was created as a side effect.

[Source: .specify/archive/008-route-command/spec.md -> User Story 1]

---

### User Story 31 - Route an existing input to an output group (Priority: P2)

A person wants to send an existing input to a group of outputs at once (e.g. "whole house")
rather than a single speaker.

**Why this priority**: Groups are a first-class routing target already supported elsewhere
in the CLI; routing should work the same way against a group as against a single output, but
this is secondary to the single-output happy path.

**Independent Test**: Can be fully tested by running `sonora route inputs/<input-id>
groups/<group-id>` and verifying the same success output as User Story 30, with the route's
target reflecting the group.

**Acceptance Scenarios**:

1. **Given** the path of an existing input and the path of an existing output group,
   **When** the user runs `sonora route inputs/<input-id> groups/<group-id>`, **Then** the
   command succeeds and displays the route's identifier and status, and confirms the target
   is a group.
2. **Given** an output and a group that happen to share the same identifier, **When** the
   user runs `sonora route inputs/<input-id> groups/<id>`, **Then** the command targets the
   group, regardless of an output also existing with that same identifier.
3. **Given** the same colliding identifiers, **When** the user runs `sonora route
   inputs/<input-id> outputs/<id>`, **Then** the command targets the output rather than the
   group.

[Source: .specify/archive/008-route-command/spec.md -> User Story 2]

---

### User Story 32 - Enable or disable an input (Priority: P2)

A person runs `sonora enable inputs/<input-id>` or `sonora disable inputs/<input-id>` to allow or block new playback from an input, and sees the input's updated record. Since hub 0.1.25, disabling an input also stops every route playing it (FR-139); enabling it starts nothing. [Source: .specify/archive/011-hub-0128-alignment/spec.md -> FR-001]

**Why this priority**: Controls the hub from the CLI; every read command works without it.

**Independent Test**: Can be fully tested by: run each verb against a mock hub and check the request body and the printed record.

**Acceptance Scenarios**:

1. **Given** an existing input, **When** the user runs `sonora enable inputs/<input-id>`, **Then** the hub receives `{"enabled": true}`, the updated input is printed and the exit status is `0`.
2. **Given** an existing input, **When** the user runs `sonora disable inputs/<input-id>`, **Then** the hub receives `{"enabled": false}` and the updated input is printed.
3. **Given** an unknown input identifier, **When** the user runs either verb, **Then** an "input not found" message goes to standard error with exit status `5`.

[Source: .specify/archive/tiny/2026-08-28-inputs-enable-disable-command/inputs-enable-disable-command.md]

---

### User Story 33 - Stop a single route (Priority: P1)

A person runs `sonora delete routes/<route-id>`, or its alias `sonora stop routes/<route-id>`, to stop playback and remove the route. (`stop` also has bulk forms; see User Story 43.)

**Why this priority**: Controls the hub from the CLI; every read command works without it.

**Independent Test**: Can be fully tested by: run both verbs against a mock hub and check that the request, the output and the exit codes are identical.

**Acceptance Scenarios**:

1. **Given** a playing route, **When** the user runs `sonora delete routes/<route-id>`, **Then** the hub receives `DELETE /api/v2/routes/<route-id>`, the route identifier, status `stopped` and a message are printed, and the exit status is `0`.
2. **Given** the same route, **When** the user runs `sonora stop routes/<route-id>` instead, **Then** the request, output and exit status are the same.
3. **Given** an unknown route identifier, **When** the user runs either verb, **Then** a "route not found" message goes to standard error with exit status `5`.

[Source: .specify/archive/tiny/2026-08-28-route-delete-command/route-delete-command.md]

---

### User Story 34 - Move a playing route to another target (Priority: P2)

A person runs `sonora transfer routes/<route-id> <outputs|groups>/<target-id>` to move what is playing to another output or group without interruption. The hub replaces the route, so the command prints the new route's identifier.

**Why this priority**: Controls the hub from the CLI; every read command works without it.

**Independent Test**: Can be fully tested by: run the command for an output and for a group against a mock hub and check the request body and the printed identifier.

**Acceptance Scenarios**:

1. **Given** a playing route and an existing output, **When** the user runs `sonora transfer routes/<route-id> outputs/<id>`, **Then** the hub receives the target identifier and type, the new route's identifier, status and message are printed, and the exit status is `0`.
2. **Given** a target that does not exist, **When** the user runs the command, **Then** an "output/group not found" message goes to standard error with exit status `12`, before any transfer request.
3. **Given** an unknown route identifier, **When** the user runs the command, **Then** a "route not found" message goes to standard error with exit status `5`.

[Source: .specify/archive/tiny/2026-08-28-route-transfer-command/route-transfer-command.md]

---

### User Story 35 - Enable or disable an output group (Priority: P2)

A person runs `sonora enable groups/<group-id>` or `sonora disable groups/<group-id>` to allow or block new playback to a group, and sees the group's updated record. Since hub 0.1.25, disabling a group also stops the routes addressed to it, while routes addressed to a member keep playing (FR-139); enabling it starts nothing. [Source: .specify/archive/011-hub-0128-alignment/spec.md -> FR-001]

**Why this priority**: Controls the hub from the CLI; every read command works without it.

**Independent Test**: Can be fully tested by: run each verb against a mock hub and check the request body and the printed group.

**Acceptance Scenarios**:

1. **Given** an existing group, **When** the user runs `sonora enable groups/<group-id>`, **Then** the hub receives `{"enabled": true}` and the updated group is printed with exit status `0`.
2. **Given** an unknown group identifier, **When** the user runs either verb, **Then** a "group not found" message goes to standard error with exit status `5`.

[Source: .specify/archive/tiny/2026-08-29-groups-enable-disable-command/groups-enable-disable-command.md]

---

### User Story 36 - Set a group's volume (Priority: P2)

A person runs `sonora set groups/<group-id> volume <0-100>` to change the volume of an output group and sees the volume the hub applied.

**Why this priority**: Controls the hub from the CLI; every read command works without it.

**Independent Test**: Can be fully tested by: run the command with a valid and an invalid value against a mock hub and check the request and the exit status.

**Acceptance Scenarios**:

1. **Given** an existing group, **When** the user runs `sonora set groups/<group-id> volume 40`, **Then** the hub receives `{"volume": 40}`, the group identifier, volume and update time are printed, and the exit status is `0`.
2. **Given** a value outside 0–100 or not an integer, **When** the user runs the command, **Then** a usage error is printed with exit status `2` and no request is sent.
3. **Given** an unknown group identifier, **When** the user runs the command, **Then** a "group not found" message goes to standard error with exit status `5`.

[Source: .specify/archive/tiny/2026-08-29-groups-set-volume-command/groups-set-volume-command.md]

---

### User Story 37 - Create and delete an input (Priority: P2)

A person runs `sonora create inputs/<input-id> <uri> --display-name <name>` to define a stored input, and `sonora delete inputs/<input-id>` to remove one.

**Why this priority**: Controls the hub from the CLI; every read command works without it.

**Independent Test**: Can be fully tested by: create and then delete an input against a mock hub and check both requests and outputs.

**Acceptance Scenarios**:

1. **Given** a free input identifier, **When** the user runs `sonora create inputs/<id> <uri> --display-name <name>`, **Then** the hub receives the identifier, URI, display name, `autoRemove` false and `enabled` true, the created input is printed, and the exit status is `0`.
2. **Given** `--disabled` or `--auto-remove`, **When** the user runs the command, **Then** `enabled` is false or `autoRemove` is true in the request.
3. **Given** an existing input, **When** the user runs `sonora delete inputs/<id>`, **Then** the input identifier, a status and a message are printed with exit status `0`.

[Source: .specify/archive/tiny/2026-08-29-inputs-create-delete-command/inputs-create-delete-command.md]

---

### User Story 38 - Check and set the master mute (Priority: P2)

A person runs `sonora get master-mute` to see whether the whole system is muted, and `sonora mute all` or `sonora unmute all` to change it.

**Why this priority**: Controls the hub from the CLI; every read command works without it.

**Independent Test**: Can be fully tested by: run the three commands against a mock hub and check the requests and the printed state.

**Acceptance Scenarios**:

1. **Given** a hub, **When** the user runs `sonora get master-mute`, **Then** `muted: true` or `muted: false` is printed with exit status `0`.
2. **Given** a hub, **When** the user runs `sonora mute all`, **Then** the hub receives `{"muted": true}` and the updated state is printed.
3. **Given** `sonora get master-mute/<id>` or `sonora list master-mute`, **Then** a usage error is printed with exit status `2`.

[Source: .specify/archive/tiny/2026-08-29-master-mute-command/master-mute-command.md]

---

### User Story 39 - Enable or disable an output (Priority: P2)

A person runs `sonora enable outputs/<output-id>` or `sonora disable outputs/<output-id>` to allow or block new playback to a speaker, and sees the output's updated record. Since hub 0.1.25, disabling an output also takes it out of every route playing on it, and enabling it puts it back into the live group routes of its enabled groups (FR-139). [Source: .specify/archive/011-hub-0128-alignment/spec.md -> FR-001]

**Why this priority**: Controls the hub from the CLI; every read command works without it.

**Independent Test**: Can be fully tested by: run each verb against a mock hub and check the request body and the printed output.

**Acceptance Scenarios**:

1. **Given** an existing output, **When** the user runs `sonora disable outputs/<output-id>`, **Then** the hub receives `{"enabled": false}` and the updated output is printed with exit status `0`.
2. **Given** an unknown output identifier, **When** the user runs either verb, **Then** an "output not found" message goes to standard error with exit status `5`.

[Source: .specify/archive/tiny/2026-08-29-outputs-enable-disable-command/outputs-enable-disable-command.md]

---

### User Story 40 - Mute or unmute an output or group (Priority: P2)

A person runs `sonora mute outputs/<id>`, `sonora unmute outputs/<id>`, or the same for `groups/<id>`, to silence or restore one speaker or group without changing its volume.

**Why this priority**: Controls the hub from the CLI; every read command works without it.

**Independent Test**: Can be fully tested by: run each verb for an output and a group against a mock hub and check the request body and the printed record.

**Acceptance Scenarios**:

1. **Given** an existing output, **When** the user runs `sonora mute outputs/<id>`, **Then** the hub receives `{"muted": true}` and the updated output is printed with exit status `0`.
2. **Given** an existing group, **When** the user runs `sonora unmute groups/<id>`, **Then** the hub receives `{"muted": false}` and the updated group is printed.
3. **Given** `inputs/<id>` or `routes/<id>`, **When** the user runs `mute`, **Then** a usage error is printed with exit status `2`.

[Source: .specify/archive/tiny/2026-08-29-outputs-groups-mute-command/outputs-groups-mute-command.md]

---

### User Story 41 - Set an output's volume (Priority: P2)

A person runs `sonora set outputs/<output-id> volume <0-100>` to change one speaker's volume and sees the volume the hub applied.

**Why this priority**: Controls the hub from the CLI; every read command works without it.

**Independent Test**: Can be fully tested by: run the command with a valid and an invalid value against a mock hub and check the request and the exit status.

**Acceptance Scenarios**:

1. **Given** an existing output, **When** the user runs `sonora set outputs/<output-id> volume 40`, **Then** the hub receives `{"volume": 40}`, the output identifier, volume and update time are printed, and the exit status is `0`.
2. **Given** a value outside 0–100, a non-integer value, a misspelled attribute word or a non-`volume` attribute, **When** the user runs the command, **Then** a usage error is printed with exit status `2` and no request is sent.
3. **Given** an unknown output identifier, **When** the user runs the command, **Then** an "output not found" message goes to standard error with exit status `5`.

[Source: .specify/archive/tiny/2026-08-29-outputs-set-volume-command/outputs-set-volume-command.md]

---

### User Story 42 - Pause and resume a route (Priority: P2)

A person runs `sonora pause routes/<route-id>` to hold a playing route and `sonora resume routes/<route-id>` to continue it.

**Why this priority**: Controls the hub from the CLI; every read command works without it.

**Independent Test**: Can be fully tested by: run both verbs against a mock hub and check the request body and the printed record.

**Acceptance Scenarios**:

1. **Given** a playing route, **When** the user runs `sonora pause routes/<route-id>`, **Then** the hub receives `{"paused": true}`, the route identifier, paused flag, status and a message are printed, and the exit status is `0`.
2. **Given** an already-paused route, **When** the user runs `sonora pause routes/<route-id>` again, **Then** the command still succeeds, because the hub answers 200 with the current state.
3. **Given** an unknown route identifier, **When** the user runs either verb, **Then** a "route not found" message goes to standard error with exit status `5`.

[Source: .specify/archive/tiny/2026-08-29-route-pause-resume-command/route-pause-resume-command.md]

---

### User Story 43 - Stop many routes at once (Priority: P2)

A person runs `sonora stop routes` to stop everything playing, or `sonora stop outputs/<id>` / `sonora stop groups/<id>` to stop what is playing on one speaker or group, and sees which routes were stopped.

**Why this priority**: Controls the hub from the CLI; every read command works without it.

**Independent Test**: Can be fully tested by: run the three forms against a mock hub, including a case where nothing was playing, and check the requests and the printed list.

**Acceptance Scenarios**:

1. **Given** playing routes, **When** the user runs `sonora stop routes`, **Then** the hub receives `DELETE /api/v2/routes`, the number of stopped routes and each stopped route (identifier, target type, target identifier, stop reason) are printed, and the exit status is `0`.
2. **Given** nothing is playing, **When** the user runs `sonora stop routes`, **Then** a count of zero is printed with exit status `0`.
3. **Given** an unknown output or group identifier, **When** the user runs `sonora stop outputs/<id>`, **Then** a "not found" message goes to standard error with exit status `5`.

[Source: .specify/archive/tiny/2026-09-09-stop-routes-command/stop-routes-command.md]

---

### User Story 44 - Announce text on a single output (Priority: P1)

A person (or a script driven by a home automation) wants a spoken announcement — "Dinner is ready" — to play on one speaker, with a single command.

**Why this priority**: This is the whole point of the TTS extension and the smallest slice that delivers value. Everything else in this feature is secondary to getting text spoken in a room.

**Independent Test**: Run the speak command with a short text and the identifier of an existing output on a hub that has the TTS extension loaded; verify the command succeeds, prints the announcement identifier, whether the audio came from cache, and the target's queue depth, and that the announcement is heard on that output.

**Acceptance Scenarios**:

1. **Given** a hub with the TTS extension loaded and an existing output `kitchen`, **When** the user runs `sonora speak "Dinner is ready" outputs/kitchen`, **Then** the command succeeds and displays the announcement identifier, whether the audio was served from cache, and the number of announcements now queued for that target.
2. **Given** the same successful request, **When** the user adds `--json`, **Then** the output is strict, parseable JSON carrying the same three fields.
3. **Given** a successful request, **When** the command returns, **Then** it has not waited for the announcement to finish playing — it returns as soon as the hub has accepted it (by which time the audio already exists and is queued).
4. **Given** empty or whitespace-only text, **When** the user runs the speak command, **Then** it fails with a clear usage error and no request is sent to the hub.
5. **Given** text piped on standard input, **When** the user runs `echo "Dinner is ready" | sonora speak - outputs/kitchen`, **Then** the hub receives `Dinner is ready` (trailing newline removed) and the command succeeds as in scenario 1. Empty piped input fails as in scenario 4.

[Source: .specify/archive/009-tts-commands/spec.md -> User Story 1]

---

### User Story 45 - Announce text on an output group (Priority: P2)

A person wants one announcement — "Motion detected in garden" — to play on every speaker in a group at once.

**Why this priority**: Whole-house announcements are the next most valuable use case, and reuse the same command with a different target kind.

**Independent Test**: Run the speak command targeting an existing group and verify the same success output as User Story 44, with the announcement heard across the group's outputs.

**Acceptance Scenarios**:

1. **Given** an existing group `all-rooms`, **When** the user runs `sonora speak "Motion detected in garden" groups/all-rooms`, **Then** the command succeeds with the same output fields as for a single output.
2. **Given** a target written with a resource alias (`gr/all-rooms`, `out/kitchen`), **When** the user runs the speak command, **Then** it behaves identically to the full resource name, as for every other command that takes a target.
3. **Given** a target path whose resource kind is not `outputs` or `groups` (e.g. `routes/x`, `inputs/x`) or that has no identifier, **When** the user runs the speak command, **Then** it fails with a clear usage error before contacting the hub.

[Source: .specify/archive/009-tts-commands/spec.md -> User Story 2]

---

### User Story 46 - Choose provider, voice, or language per announcement (Priority: P3)

An operator has several TTS providers configured on the hub (e.g. a cloud one as default and a local one for offline use) and wants a particular announcement to use a specific provider, voice, or language.

**Why this priority**: Useful but optional — every announcement works with the hub's configured defaults.

**Independent Test**: Run the speak command with each override option and confirm the hub received the chosen value; run it with an unknown provider name and confirm a clear "provider not found" failure.

**Acceptance Scenarios**:

1. **Given** a hub with a provider named `piper-local`, **When** the user runs the speak command with `--provider piper-local`, **Then** the announcement is requested from that provider.
2. **Given** the user supplies `--voice` and/or `--language`, **When** the speak command runs, **Then** those values are sent to the hub; omitted ones fall back to the provider's defaults on the hub.
3. **Given** a provider name the hub does not know, **When** the user runs the speak command, **Then** it fails with a clear "provider not found" error, distinct from "target not found".

[Source: .specify/archive/009-tts-commands/spec.md -> User Story 3]

---

### User Story 47 - Inspect the TTS cache (Priority: P3)

An operator wants to see how much disk the TTS cache uses, against its configured limit, and how the entries split across providers.

**Why this priority**: Observability for the extension; not needed to make announcements.

**Independent Test**: Run the cache-statistics command against a hub with the TTS extension loaded and verify it prints total entries, total size, maximum size, and per-provider entry counts, in YAML by default and JSON with `--json`.

**Acceptance Scenarios**:

1. **Given** a hub with the TTS extension loaded, **When** the user runs `sonora get tts-cache`, **Then** the command prints the total number of cached entries, their total size, the configured maximum size, and the entry count per provider.
2. **Given** an empty cache, **When** the user runs the same command, **Then** it succeeds and shows zero entries, zero size, and an empty per-provider breakdown (not an error and not a missing field).

[Source: .specify/archive/009-tts-commands/spec.md -> User Story 4]

---

### User Story 48 - Clear the TTS cache (Priority: P4)

An operator wants to force fresh synthesis — for example after changing a provider's voice settings — by clearing the cache entirely or only for one provider.

**Why this priority**: An administrative action needed rarely.

**Independent Test**: Run the cache-clear command with and without a provider option and verify that the hub reports success and that the statistics command afterwards reflects the cleared entries.

**Acceptance Scenarios**:

1. **Given** a populated cache, **When** the user runs `sonora clear tts-cache`, **Then** the whole cache is cleared and the command prints `cleared: all`.
2. **Given** a populated cache, **When** the user runs `sonora clear tts-cache --provider openai`, **Then** only that provider's entries are cleared and the command prints `cleared: provider` and `provider: openai`.
3. **Given** a provider name the hub does not know, **When** the user runs the clear command with that `--provider`, **Then** it fails with a clear "provider not found" error and nothing is cleared.
4. **Given** an already-empty cache, **When** the user clears it, **Then** the command still succeeds (clearing is idempotent).

[Source: .specify/archive/009-tts-commands/spec.md -> User Story 5]

---

### User Story 49 - Override the response wait for a slow TTS provider (Priority: P4)

An operator whose hub's TTS provider is configured with a longer synthesis timeout than the hub's 10-second default wants `speak` to wait long enough to receive the hub's own timeout error, instead of the CLI giving up first and misreporting a network failure.

**Why this priority**: Only relevant when a hub deviates from the assumed 10-second default; the fixed 15-second bound (FR-114) already covers the common case.

**Independent Test**: Run `speak` against a fake hub that takes longer than 15 seconds to respond, with `--timeout` set high enough to cover it, and confirm the command waits and reports the hub's actual response rather than a CLI-side timeout.

**Acceptance Scenarios**:

1. **Given** a hub whose provider takes longer than 15 seconds, **When** the user runs `speak` with `--timeout 30s`, **Then** the command waits up to 30 seconds and reports the hub's own response (success or provider error) rather than a CLI network timeout.
2. **Given** no `--timeout` is supplied, **When** the user runs `speak`, **Then** the default 15-second bound (FR-114) applies unchanged.
3. **Given** `--timeout` is supplied with a zero, negative, or unparsable value, **When** the user runs the speak command, **Then** it fails with a usage error before contacting the hub.

[Source: .specify/archive/009-tts-commands/spec.md -> User Story 6]

### User Story 50 - Discover a provider's voices (Priority: P3)

An operator wants to see which voices a TTS provider offers, so that the `--voice` value passed to `speak` is one the hub accepts.

**Why this priority**: A discovery aid for `speak --voice`; announcements work without it.

**Independent Test**: Run `sonora list tts-voices --provider <name>` against a hub whose provider is `google-cloud` or `google-gemini`, and verify that every voice is printed in the hub's order with all five fields (absent ones as null), in YAML by default and JSON with `--json`.

**Acceptance Scenarios**:

1. **Given** a hub with the TTS extension loaded and a `google-cloud` provider, **When** the user runs `sonora list tts-voices --provider <name>`, **Then** the command prints each voice in the hub's order with its short name, full name, engine, language and gender.
2. **Given** the same hub, **When** the user adds `--language` or `--engine`, **Then** the request carries that filter as a query parameter; without them the URL has no query string.
3. **Given** a `google-gemini` provider, **When** the user lists its voices, **Then** each voice's `engine` is the provider entry's model and its `language` is shown as null.
4. **Given** a provider whose type does not list voices, or an unknown provider, **When** the user lists its voices, **Then** the command shows the hub's message, exits with the code FR-113 gives, and prints nothing to standard output.
5. **Given** `--provider` is missing or blank, or a positional argument is given, **When** the user runs the command, **Then** it fails with a usage error before contacting the hub.

[Source: .specify/archive/tiny/2026-09-27-list-tts-voices-command/list-tts-voices-command.md -> What; Requirements]

---

### User Story 51 - Import the hub client from another project (Priority: P1)

A developer of another Sonora project (sonora-mcp) adds this module as a dependency by its public import path, pinned to a released tag, and calls hub operations (list outputs, set volume, create a route, speak, …) through the shared client, getting the same timeouts, typed errors and error classification the CLI uses.

**Why this priority**: This is the whole point of the feature. Without an importable, released client, sonora-mcp cannot drop its own client.

**Independent Test**: In a scratch module outside this repository, require this module at the released tag, call one read operation and one write operation against a fake hub, and classify a failure — with no access to anything under this repo's internal packages.

**Acceptance Scenarios**:

1. **Given** a released tag of this module, **When** an external project requires the module by its public path and imports the hub client package, **Then** it resolves and builds with no replace directives, workspace files or private-module settings.
2. **Given** an external project using the client, **When** a hub call fails (hub unreachable, timeout, 404, 4xx/5xx, malformed response), **Then** the project can obtain the error class and friendly message the client's classifier produces — the same classification the CLI starts from — without depending on any CLI-specific concept such as exit codes. (A caller may refine a generic not-found class with its own context, as the CLI's route commands do.)
3. **Given** the client package, **When** a developer reads its package documentation, **Then** every exported identifier is documented in terms that make sense without access to this repo's specs, research notes or constitution.
4. **Given** the client package's dependency list, **When** it is imported, **Then** it adds no third-party dependencies to the consumer beyond the standard library.

[Source: .specify/archive/010-public-hub-package/spec.md -> User Story 1]

---

### User Story 52 - Existing CLI users notice no change (Priority: P1)

A person using `sonora` on Windows or Linux upgrades to the build that contains this refactor and keeps using every command exactly as before.

**Why this priority**: The refactor is only acceptable if it is invisible to CLI users; a behavior regression would break live audio control and scripts that rely on exit codes.

**Independent Test**: Run the full existing unit and contract test suite unchanged in intent (only import paths adjusted) on Windows and Linux, and compare `sonora --help`, command output and exit codes for representative commands before and after.

**Acceptance Scenarios**:

1. **Given** the refactored build, **When** any existing command is run, **Then** its output (YAML and `--json`), error messages and exit codes are identical to the previous build.
2. **Given** a build produced by the project's build scripts with a version supplied, **When** `sonora --version` is run, **Then** it prints the supplied version, not the default placeholder.
3. **Given** a hub error of each class, **When** a command hits it, **Then** the CLI exits with the same code as before, even though the error-class-to-exit-code mapping now lives inside the CLI rather than in the shared client.

[Source: .specify/archive/010-public-hub-package/spec.md -> User Story 2]

---

### User Story 53 - Consumers verify against the same hub spec (Priority: P2)

A developer of a consuming project runs a conformance test against the exact Multiroom Audio Hub API specification that the imported client version was built and tested against, without keeping their own copy of `openapi.json`.

**Why this priority**: Removes the last duplicated artifact between repos, but the client is usable without it.

**Independent Test**: From an external module, import the published spec and confirm it is byte-identical to `api/openapi.json` at the same tag.

**Acceptance Scenarios**:

1. **Given** a released tag, **When** a consumer imports the published spec, **Then** it receives the full contents of `api/openapi.json` from that tag.
2. **Given** `api/openapi.json` is updated in this repo, **When** the module is rebuilt, **Then** the published spec reflects the update with no separate copy to keep in sync.

[Source: .specify/archive/010-public-hub-package/spec.md -> User Story 3]

---

### User Story 54 - Maintainers release the shared client safely (Priority: P2)

A maintainer tags a release of this module that contains the public client so consumers can pin it, following rules that keep already-published versions immutable.

**Why this priority**: Consumers can only depend on a tag; a mistaken or moved tag in a public module is permanent damage.

**Independent Test**: Tag a release, fetch it from a clean environment through the public Go module proxy, and build a consumer against it.

**Acceptance Scenarios**:

1. **Given** the refactor is merged, **When** a new tag is published, **Then** consumers can resolve that tag through the public module proxy.
2. **Given** a published tag has a problem, **When** the maintainer fixes it, **Then** a new tag is published and the existing tag is never moved, deleted or re-created.

[Source: .specify/archive/010-public-hub-package/spec.md -> User Story 4]

---

### User Story 55 - Know that a disable stops playback before running it (Priority: P1)

A user wants to stop new routes from using the kitchen radio input and reads the help for `sonora disable` (or the README) first. The documentation tells them that disabling an input stops every route playing it right now, so they know the music will stop.

**Why this priority**: The current documentation is wrong in a way that has an audible consequence: a user following it stops music they meant to keep. It is also the smallest change.

**Independent Test**: Read `sonora enable --help`, `sonora disable --help` and the README's enable/disable section, and compare each statement with the hub 0.1.28 contract descriptions of `setInputEnabled`, `setOutputEnabled` and `setGroupEnabled`.

**Acceptance Scenarios**:

1. **Given** the README, **When** a user reads the enable/disable section, **Then** it states, for each of inputs, outputs and groups, what disabling does to routes that are playing, and what enabling does.
2. **Given** `sonora disable inputs/<id> --help` (and the same for outputs and groups), **When** a user reads it, **Then** it says in one line what happens to playback that is running on that resource.
3. **Given** `sonora enable outputs/<id> --help`, **When** a user reads it, **Then** it says the output rejoins the group routes still playing on its groups, and that a route stopped by the disable does not come back.
4. **Given** the documentation of either command, **When** compared with the CLI before this feature, **Then** the command's arguments, flags, output and exit codes are unchanged.

[Source: .specify/archive/011-hub-0128-alignment/spec.md -> User Story 1]

---

### User Story 56 - See what `sonora play` started (Priority: P2)

A user plays a SoundCloud link without giving a name. The CLI's output shows the name the hub gave the input ("Flickermood — Forss"), so the user can tell what is playing and find it later in `sonora list inputs`. A script reading `--json` output gets the same name in a field.

A developer of sonora-mcp, using this module's hub client, reads the same name from the play result.

**Why this priority**: The hub client no longer mirrors the play response field for field, which this project treats as a defect, and both the CLI and sonora-mcp lose information the hub provides. Nothing fails, so it ranks below the misleading documentation.

**Independent Test**: Run `sonora play` against a fake hub that returns a display name, and against one that omits it, and check the YAML and JSON output and the hub client's result in each case.

**Acceptance Scenarios**:

1. **Given** a hub that answers a play request with a display name, **When** the user runs `sonora play <uri> outputs/<id>`, **Then** the output shows that display name alongside the input ID, route ID, status and message, in YAML and in JSON.
2. **Given** a hub older than 0.1.28, which does not return a display name, **When** the user runs the same command, **Then** the output has no display name field at all (not an empty one), and is otherwise identical to the output before this feature.
3. **Given** a developer calling the hub client's play operation, **When** the hub returns a display name, **Then** the result carries it; **When** the hub omits it, **Then** the result's display name is empty and the call still succeeds.
4. **Given** `sonora play --help` and the README, **When** a user reads about `--display-name`, **Then** they learn it is optional, that without it a SoundCloud or YouTube link is named after its track and other sources get a generated name, and that a name they give is kept.

[Source: .specify/archive/011-hub-0128-alignment/spec.md -> User Story 2]

---

### User Story 57 - A transfer right after enabling an output just works (Priority: P3)

A user re-enables the office speaker, which rejoins the living-room group route, and immediately moves that route to another group with `sonora transfer`. The hub briefly refuses the transfer because the route is still adding the office speaker. The CLI waits a moment and tries again, and the transfer succeeds without the user noticing. If the hub keeps refusing, the CLI says the route was still adding an output and to try again shortly, instead of telling the user to change the hub's state.

**Why this priority**: It only happens in a narrow window (a transfer within a fraction of a second of an enable), most likely from scripts. When it happens today, the user gets wrong advice, but a manual retry works.

**Independent Test**: Run `sonora transfer` against a fake hub that refuses the first N transfer requests with reason `ROUTE_CHANGING_OUTPUTS` and then accepts, for N within and beyond the retry allowance, and check output, stderr and exit code.

**Acceptance Scenarios**:

1. **Given** a hub that refuses a transfer once with reason `ROUTE_CHANGING_OUTPUTS` and then accepts it, **When** the user runs `sonora transfer routes/<id> groups/<id>`, **Then** the command succeeds with the same output and exit code as a transfer accepted at once, and prints nothing extra.
2. **Given** a hub that keeps refusing with that reason, **When** the user runs the transfer, **Then** the CLI stops retrying within about 2 seconds of the first refusal, exits with the conflict exit code (14), and its error message says the route is still adding an output and to retry shortly.
3. **Given** a hub that refuses a transfer with any other reason (for example `OUTPUT_DISABLED`), or with no reason, **When** the user runs the transfer, **Then** the CLI does not retry and reports the refusal exactly as before this feature.
4. **Given** `--verbose`, **When** the retries run out, **Then** the underlying hub error is shown as it is for any other conflict.
5. **Given** the README's exit-code table, **When** a user reads the row for 14, **Then** it mentions that one kind of refusal clears by itself and the CLI retries it during a transfer.
6. **Given** a developer reading the hub client's documentation of refusal reasons, **When** they look up the reasons and which of them name an output, **Then** `ROUTE_CHANGING_OUTPUTS` is listed, described as transient, and listed among the reasons that name no output.

[Source: .specify/archive/011-hub-0128-alignment/spec.md -> User Story 3]

---

### User Story 58 - Rename an input or restore its original name (Priority: P2)

A person runs `sonora rename inputs/<input-id> <name>` to give an input a new display name, or `sonora rename inputs/<input-id> --reset` to restore the name it had before, and sees the input's updated record. [Source: .specify/archive/tiny/2026-10-22-rename-inputs-command/rename-inputs-command.md]

**Why this priority**: Display names are how inputs are told apart in listings; the command is small and independent of playback.

**Independent Test**: Run each form against a mock hub and check the request body, the printed record, the exit status and the usage errors.

**Acceptance Scenarios**:

1. **Given** an existing input, **When** the user runs `sonora rename inputs/<input-id> "Kitchen radio"`, **Then** the hub receives `{"displayName": "Kitchen radio"}`, the updated input is printed and the exit status is `0`.
2. **Given** an input that has been renamed, **When** the user runs `sonora rename inputs/<input-id> --reset`, **Then** the hub receives `{"displayName": null}` and the input is printed with its original name.
3. **Given** an unknown input identifier, **When** the user runs either form, **Then** an "input not found" message goes to standard error with exit status `5`.
4. **Given** a name together with `--reset`, neither of them, a missing identifier, extra positional arguments, or a resource other than `inputs`, **When** the user runs `rename`, **Then** a usage error goes to standard error with exit status `2` and no request is sent.

[Source: .specify/archive/tiny/2026-10-22-rename-inputs-command/rename-inputs-command.md]

---

### User Story 59 - Announce over music, or play alongside it (Priority: P1)

Music plays in the kitchen. The user plays a doorbell chime there with
`sonora play <uri> outputs/kitchen --join-mode duck_others`. The music gets quieter while the chime
plays and comes back afterwards. With `--join-mode mix` the chime plays over the music at full level.
Without the flag, nothing changes from today.

`sonora route inputs/<id> outputs/<id> --join-mode <mode>` does the same for a configured input. A user
who creates an input meant for announcements can give it a default join mode with
`sonora create inputs/<id> ... --default-join-mode duck_others`, so every route from it ducks unless the
request says otherwise.

A sonora-mcp developer sets the same join mode through the hub client's route, play and create-input
requests.

**Why this priority**: It unlocks something the CLI cannot do at all today, and it is the reason the hub
added join modes. Each request change is one optional field.

**Independent Test**: Run each command against a fake hub, with and without the flag, and check the
request body the hub receives and the command's output and exit code. Run each command with an
invalid mode and check that no request is sent.

**Acceptance Scenarios**:

1. **Given** `sonora play <uri> outputs/<id> --join-mode duck_others`, **When** it runs, **Then** the
   hub receives a play request whose join mode is `DUCK_OTHERS`, and the command's output and exit code
   are the same as for a play without the flag.
2. **Given** `sonora route inputs/<id> groups/<id> --join-mode MIX`, **When** it runs, **Then** the hub
   receives a route request whose join mode is `MIX`.
3. **Given** `sonora create inputs/<id> --display-name ... <uri> --default-join-mode Duck_Others`, **When**
   it runs, **Then** the hub receives a create-input request whose default join mode is `DUCK_OTHERS`.
4. **Given** any of the three commands without its join-mode flag, **When** it runs, **Then** the request
   has no join-mode field at all, so the hub applies its own default, and the request body is the same
   as before this feature.
5. **Given** `--join-mode fade` (or `--default-join-mode fade`), **When** the command runs, **Then** it
   sends no request, prints the usage and an error that lists the valid modes, and exits with the usage
   exit code (2).
6. **Given** `--help` on any of the three commands, **When** a user reads it, **Then** the flag is
   listed with the three valid modes and what the hub does when the flag is left out.
7. **Given** a developer using the hub client, **When** they leave the join mode unset on a route, play
   or create-input request, **Then** the request carries no join-mode field; **When** they set it,
   **Then** the request carries it as given.

[Source: .specify/archive/012-join-modes/spec.md -> User Story 1]

---

### User Story 60 - See what an output is playing (Priority: P2)

The user hears two things in the kitchen and wants to know why. `sonora get outputs/kitchen
--include-routes` shows the kitchen's usual record and below it the list of routes playing on it. The list
is in the order the routes joined and includes group routes. For each route it shows the join mode,
whether the route is lowered, and which one is the output's main route. The user can then stop or move
the right route.

A sonora-mcp developer gets the same list from the hub client.

**Why this priority**: It is the only way to see the effect of User Story 59 and to find group routes
on an output. It needs a new hub client operation and new output, and it depends on the hub's version,
so it is more work than the request flags.

**Independent Test**: Run `sonora get outputs/<id> --include-routes` against a fake hub that returns a
list of routes, an empty list, an unknown duck state, a 404 for the output, and a 404 or 405 for the
routes only (an older hub). Check the YAML, JSON and exit code in each case. Run it without the flag and check
that only one request is sent.

**Acceptance Scenarios**:

1. **Given** an output with a direct route and a group route playing, **When** the user runs
   `sonora get outputs/<id> --include-routes`, **Then** the output shows the output's record followed by
   a `routes` list with both routes in the hub's order. In YAML each entry has its route ID, input,
   target, target type, status, join mode, duck state, whether it is the main route, and whether it is
   paused. In JSON each entry has every field the hub returns for it.
2. **Given** an output with nothing playing, **When** the user runs the same command, **Then** the
   output's record is followed by an empty routes list, in YAML and in JSON.
3. **Given** `sonora get outputs/<id>` without the flag, **When** it runs, **Then** it sends one request,
   and its output is byte-for-byte the same as before this feature.
4. **Given** an unknown output, **When** the user runs the command with the flag, **Then** it reports the
   output as not found, exactly as `get outputs/<id>` does today.
5. **Given** a hub older than 0.1.21, which knows the output but answers 404 or 405 for the output's routes,
   **When** the user runs the command with the flag, **Then** it does not say the output was not
   found. It says that this hub cannot list an output's routes and needs 0.1.21 or later, and exits with
   the exit code for an unexpected hub response (3), not the not-found code (5).
6. **Given** a route whose duck state is a value the CLI does not know, **When** it is shown, **Then** the
   CLI shows the hub's value unchanged and does not fail.
7. **Given** a developer calling the hub client's new operation for an unknown output, **When** the hub
   answers 404, **Then** the client returns the same not-found error it returns for other output
   operations, naming the output.
8. **Given** a developer reading a route entry's duck state, **When** the value is one the client does
   not know, **Then** the client can tell them the route is not lowered, as the contract asks.

[Source: .specify/archive/012-join-modes/spec.md -> User Story 2]

---

### User Story 61 - See a route's join mode and outputs, and an input's default (Priority: P3)

The user runs `sonora get routes/<id>` and sees, in the default YAML, how the route joined its target
and which outputs it plays on right now. A group route may play on fewer members than its group has.
`sonora get inputs/<id>` shows the input's default join mode, or that it has none.

**Why this priority**: The data is already in the hub client and in the JSON output; only the YAML
output lacks it. It is useful, but not needed to use the other two stories.

**Independent Test**: Run `get routes/<id>` and `get inputs/<id>` against a fake hub that returns the
fields and one that omits them, and compare the YAML and JSON output.

**Acceptance Scenarios**:

1. **Given** a route with join mode `MIX` playing on two outputs, **When** the user runs
   `sonora get routes/<id>`, **Then** the YAML shows the join mode and the two outputs in the hub's order,
   after the fields it shows today.
2. **Given** a group route currently playing on none of its members, **When** shown, **Then** the
   outputs field is an empty list.
3. **Given** an input with default join mode `DUCK_OTHERS`, **When** the user runs
   `sonora get inputs/<id>`, **Then** the YAML shows it; **Given** an input without one, **Then** the YAML
   shows the field as null, the way it shows a missing creation time.
4. **Given** `--json` on either command, **When** it runs, **Then** the JSON includes the same fields as
   the YAML and keeps its current shape.
5. **Given** `sonora list routes` and `sonora list inputs`, **When** they run, **Then** their output is
   unchanged.

[Source: .specify/archive/012-join-modes/spec.md -> User Story 3]

---

### User Story 62 - Set output volume, route volume and ducking level from the command line (Priority: P2)

A person plays a URI with `--output-volume`, `--route-volume` and `--ducking-level`, changes a playing route's volume with `sonora set routes/<route-id> volume <0-100>`, and sees a route's levels in `sonora get routes/<route-id>`. `--volume` is removed from `sonora play`.

**Why this priority**: The hub can now set a route's volume and an announcement's ducking level separately from the output's volume; without these flags that capability is unreachable from the CLI.

**Independent Test**: Run each flag and the set command against a fake hub and check the request bodies, the printed route, the exit status and the usage errors.

**Acceptance Scenarios**:

1. **Given** an output, **When** the user runs `sonora play <uri> outputs/<id> --output-volume 30 --route-volume 15 --join-mode DUCK_OTHERS --ducking-level -20`, **Then** the hub receives `outputVolume` 30, `routeVolume` 15 and `duckingLevelDb` -20, and the exit status is `0`.
2. **Given** `--volume 10`, **When** the user runs `play`, **Then** it fails as an unknown flag with exit status `2`, and no request is sent.
3. **Given** `--output-volume` or `--route-volume` outside 0-100, **When** the user runs `play`, **Then** the error names the flag, uses the validation exit code, and no request is sent.
4. **Given** `--ducking-level` without `--join-mode DUCK_OTHERS`, **When** the user runs `play`, **Then** it is a usage error with exit status `2`; **Given** a value outside -60 to 0, **Then** it is a validation error. Neither sends a request.
5. **Given** a playing route, **When** the user runs `sonora set routes/<route-id> volume 40`, **Then** the hub receives `{"volume": 40}`, the returned route is printed with its `routeVolume`, and the exit status is `0`. **Given** an unknown route, **Then** the not-found exit status `5` is returned and nothing is printed to standard output.
6. **Given** a route with `routeVolume` 40 and `duckingLevelDb` -12.5, **When** the user runs `sonora get routes/<route-id>`, **Then** the YAML shows `routeVolume: 40` and `duckingLevelDb: -12.5` after `joinMode`; **Given** either field is unset, **Then** it is left out of the YAML. `get inputs/<id>` shows `defaultRouteVolume` and `defaultDuckingLevelDb` after `defaultJoinMode` in the same way. `--json` carries all of them, null when unset. The list views are unchanged.

[Source: .specify/archive/tiny/2026-10-25-route-volume-cli/route-volume-cli.md]

---

### User Story 63 - Announce at a chosen volume or ducking level over music (Priority: P2)

A person runs `sonora speak` with `--playback-mode`, `--route-volume` or `--ducking-level`, so that an announcement ducks or mixes over what is playing at its own level rather than the output's.

**Why this priority**: An announcement over music needs a level of its own; the flags reach the hub's new fields without changing anything else about `speak`.

**Independent Test**: Run `speak` with each flag against a fake hub and check the request body, the exit status and the usage errors.

**Acceptance Scenarios**:

1. **Given** an output, **When** the user runs `sonora speak "Dinner is ready" outputs/kitchen --playback-mode duck-others --route-volume 80 --ducking-level -20`, **Then** the hub receives `playbackMode`, `routeVolume` and `duckingLevelDb`, and the command exits `0`.
2. **Given** `--ducking-level -20` without `--playback-mode`, **When** the user runs `speak`, **Then** the request carries `duckingLevelDb` only, and the extension's configured mode applies.
3. **Given** a blank `--playback-mode`, or a `--route-volume` or `--ducking-level` out of range, **When** the user runs `speak`, **Then** the command fails before any request is sent: a blank mode with exit `2`, an out-of-range value with the validation exit code.
4. **Given** `speak` with none of the new flags, **When** the user runs it, **Then** the request body is the same as before this change.

[Source: .specify/archive/tiny/2026-10-27-speak-request-fields/speak-request-fields.md]

---

### User Story 64 - See which hub I am talking to (Priority: P1)

An operator runs one command and sees the hub's product, version, API version, start time and uptime. A script uses it to confirm the target is a Sonora Multiroom hub, to check it supports a feature, or to notice that the hub restarted since the last look.

**Why this priority**: it is read-only, needs no TTS extension and is the first thing anyone checks when something is wrong.

**Independent Test**: run it against a hub 0.1.33 and compare the output with the hub's response.

**Acceptance Scenarios**:

1. **Given** a hub 0.1.33, **When** the operator asks for hub info, **Then** product, version, API version, start time and uptime are shown.
2. **Given** `--json`, **When** the operator asks for hub info, **Then** the output is one JSON object with the same fields.
3. **Given** a hub older than 0.1.33, **When** the operator asks for hub info, **Then** the command fails with a message that says the hub does not offer this call, not a generic error.
4. **Given** the hub is unreachable, **When** the operator asks for hub info, **Then** the command fails with the same unreachable-hub error as every other command.

[Source: .specify/archive/013-info-and-tts-warm/spec.md -> "User Story 1 - See which hub I am talking to"]

---

### User Story 65 - Warm the TTS cache before it is needed (Priority: P2)

An operator or script has phrases it will announce often (a doorbell, "dinner is ready"). It asks the hub to synthesize them ahead of time, so the first real announcement starts at once and costs no provider call.

**Why this priority**: saves time and provider cost, but only for installations with the TTS extension.

**Independent Test**: warm a phrase with the free local provider, then speak the same phrase and voice and see `cacheHit: true`.

**Acceptance Scenarios**:

1. **Given** a phrase not in the cache, **When** the operator warms it, **Then** the command succeeds and reports that the phrase was synthesized (`cacheHit: false`), and nothing is played.
2. **Given** a phrase already in the cache, **When** the operator warms it again, **Then** the command succeeds and reports `cacheHit: true`.
3. **Given** provider, voice and language options, **When** the operator warms a phrase with them, **Then** the hub receives them, and a later announcement with the same options is a cache hit.
4. **Given** the text is `-`, **When** the operator warms, **Then** the text is read from standard input, as `speak` does.
5. **Given** `--json`, **When** the operator warms a phrase, **Then** the output is one JSON object with `cacheHit`.
6. **Given** empty text or a request the hub rejects as invalid, **When** the operator warms, **Then** the command fails as a usage error that shows the hub's message. **Given** an unknown provider, **Then** it fails as a not-found error with the hub's message, as `speak` does.
7. **Given** the provider fails or the audio cannot be stored, **When** the operator warms, **Then** the command fails with the hub's message and says that repeating the request is safe.
8. **Given** a hub without the TTS extension, **When** the operator warms, **Then** the command reports TTS as unavailable, as `speak` does.

[Source: .specify/archive/013-info-and-tts-warm/spec.md -> "User Story 2 - Warm the TTS cache before it is needed"]

---

### User Story 66 - Use both from Go code (Priority: P2)

A developer of sonora-mcp (or any Go program) calls the public `hub` package to get hub info and to warm a phrase, and gets typed results and the package's usual error classification.

**Why this priority**: sonora-mcp can only expose these once the client has them. A hub client fix goes into this repository first.

**Independent Test**: the contract tests call each function against a test server and check the request and the decoded result.

**Acceptance Scenarios**:

1. **Given** a server answering with the hub info shape, **When** the function is called, **Then** a typed record with each field is returned, and marshalling it gives the hub's property names.
2. **Given** a warm request with all eight spec fields, **When** the function is called, **Then** only the fields that were set are sent, and `cacheHit` is returned.
3. **Given** an answer the operation declares as an error (info: 404 or any other non-2xx; warm: 400, 404 or 503), **When** the function is called, **Then** the error is one `hub.ClassifyError` already understands.

[Source: .specify/archive/013-info-and-tts-warm/spec.md -> "User Story 3 - Use both from Go code"]

---

### User Story 67 - Play a chime before a route, and see the hub's chimes (Priority: P2)

A person, or a script, wants a short sound to play before an announcement or a stream starts on a speaker, and wants to see which chimes the hub has and which one each output, group or input uses.

**Why this priority**: an additive playback option for every command that starts a route; the existing commands do not change without it.

**Independent Test**: contract tests against a test server for each hub call; unit and integration tests for the flags and renderers; a read-only check of `list chimes` and `get chimes/<id>` against a hub that has chimes, with nothing started.

**Acceptance Scenarios**:

1. **Given** a hub with a chime `cabin`, **When** the user runs `sonora play <uri> outputs/kitchen --chime cabin` or `sonora route inputs/<id> outputs/kitchen --chime cabin`, **Then** the request carries `chime` set to `cabin`; `--chime none` sends `none`, meaning no chime.
2. **Given** `--chime ""` or a blank value, **When** the user runs `play` or `route`, **Then** it fails as a usage error (exit `2`) and no request is sent.
3. **Given** an unknown chime id, **When** the user runs `play` or `route` with it, **Then** the hub's 400 (`Unknown chime: <id>`) is reported; the CLI does not check the id itself.
4. **Given** no `--chime` flag, **When** the user runs `play` or `route`, **Then** the request carries no `chime` field.
5. **Given** a hub with chimes, **When** the user runs `sonora list chimes` or `sonora get chimes`, **Then** each chime shows its `id` and `durationMillis`; with none, the output is `# no chimes found` and `chimes: []`. `list chimes/<id>` is a usage error.
6. **Given** a chime id, **When** the user runs `sonora get chimes/<id>`, **Then** that chime is shown; an unknown id fails with `chime not found: <id>` (exit `5`), and a malformed id is a usage error (exit `2`).
7. **Given** an output, group or input with a chime set, **When** the user gets it, **Then** `get outputs` and `get groups` show `chime:`, and `get inputs` shows `defaultChime:`; with none set, the output is unchanged.
8. **Given** `sonora get` or `sonora list` with no resource name, **Then** the usage error lists every resource that verb accepts.

[Source: .specify/archive/tiny/2026-11-20-route-chime/route-chime.md -> "What"; "Requirements"]

---

### User Story 68 - Find a station and play it (Priority: P1)

An operator wants to play a station they know by name ("Kiss FM"). They search by name, read the matching stations with their codec, bitrate and country, pick one, and pass its playable address to `sonora play` without visiting any website.

**Why this priority**: this is the gap the feature closes; without it radio playback needs a browser.

**Independent Test**: search for a known station name on a hub with the radio extension, then play the address from the result on a test output.

**Acceptance Scenarios**:

1. **Given** a hub with the radio extension, **When** the operator searches with a name, **Then** the matching stations are listed with every field the hub returns, including name, codec, bitrate, country, UUID and playable address.
2. **Given** no search terms, **When** the operator searches, **Then** the hub's default page is shown (the 20 most voted stations).
3. **Given** filters for tag, country, language, codec and minimum bitrate, **When** the operator searches, **Then** each filter that was given is sent to the hub and only those.
4. **Given** a sort field and direction, **When** the operator searches, **Then** results come back in that order (name, votes, click count or bitrate; ascending or descending).
5. **Given** a page size, **When** the operator searches, **Then** at most that many stations are shown; when more exist the output carries the offset to pass for the next page.
6. **Given** part of a name in any letter case ("kiss"), **When** the operator searches, **Then** every station whose name contains it is matched, as a `grep -i` over a full list would.
7. **Given** `--json`, **When** the operator searches, **Then** the output is JSON with the same fields as the default YAML.
8. **Given** a search that matches nothing, **When** the operator searches, **Then** the command succeeds with an empty result.

[Source: .specify/archive/014-radio-station-search/spec.md -> "User Story 1 - Find a station and play it (Priority: P1)"]

---

### User Story 69 - Look up one station (Priority: P2)

An operator has a UUID (from a search, a script or a saved list) and wants to see the station's details before playing it: name, homepage, tags, codec, bitrate, whether the catalogue's last check succeeded.

**Why this priority**: it confirms what a UUID is, but search already gives most of this.

**Independent Test**: look up a UUID from a search result and compare with the hub's answer.

**Acceptance Scenarios**:

1. **Given** a known UUID, **When** the operator looks it up, **Then** every station field is shown, including a station the catalogue marks as broken.
2. **Given** an unknown UUID, **When** the operator looks it up, **Then** the command fails as not found and names the UUID.
3. **Given** a value that is not a UUID, **When** the operator looks it up, **Then** the command fails as a validation error (exit 6) that shows the hub's message.

[Source: .specify/archive/014-radio-station-search/spec.md -> "User Story 2 - Look up one station (Priority: P2)"]

---

### User Story 70 - Browse countries, languages and tags (Priority: P2)

Tag and language filters match only a whole value, so an operator needs the lists to know what to filter by. They list countries or languages (most stations first), or page through tags, optionally narrowed by a name fragment.

**Why this priority**: it makes the tag and language filters usable, but name search works without it.

**Independent Test**: list tags filtered by a fragment and use one of the results as a search filter.

**Acceptance Scenarios**:

1. **Given** a hub with the radio extension, **When** the operator lists countries or languages, **Then** the whole list is shown with each entry's station count, most stations first.
2. **Given** a name fragment, **When** the operator lists tags, **Then** a page of matching tags is shown; a page size and an offset work as for station search.
3. **Given** `--json`, **When** the operator lists any of the three, **Then** the output is JSON.

[Source: .specify/archive/014-radio-station-search/spec.md -> "User Story 3 - Browse countries, languages and tags (Priority: P2)"]

---

### User Story 71 - Search from Go code (Priority: P2)

A developer of sonora-mcp (or any Go program) calls the public `hub` package to search stations, fetch one, and list countries, languages and tags, and gets typed results and the package's usual error classification.

**Why this priority**: sonora-mcp can add radio tools only once the client has them; a hub client change goes into this repository first.

**Independent Test**: the contract tests call each function against a test server and check the request sent and the result decoded.

**Acceptance Scenarios**:

1. **Given** a search with some optional filters set, **When** the function is called, **Then** only the filters that were set are sent, each once.
2. **Given** a station page, **When** it is decoded, **Then** every station field and the next-page offset are available, and a missing next-page offset means the last page.
3. **Given** a radio problem answer (invalid parameter, station not found, catalogue bad answer, catalogue unavailable), **When** the function is called, **Then** the error carries its kind, message, the offending parameter where one is named, and the retry delay where one is given, and `hub.ClassifyError` places it in the right class.
4. **Given** a 404 without a radio problem body, **When** any radio function is called, **Then** the error is a typed "radio not offered" error, and the diagnosis turns it into a "radio is not available" error with the reason from the extension inventory.

[Source: .specify/archive/014-radio-station-search/spec.md -> "User Story 4 - Search from Go code (Priority: P2)"]

---

### User Story 72 - Keep an input across hub restarts (Priority: P2)

An operator wants an input they play often, a stream or a file, to still be there after the hub restarts. They create it as kept, or keep an input that already exists, and can see from the input record which inputs the hub keeps and whether each one can play.

**Why this priority**: without it an input cannot outlive a restart, and the CLI cannot tell a kept input from a runtime one.

**Independent Test**: the contract tests run `create` and `keep` against a test hub and check the request sent and the input printed.

**Acceptance Scenarios**:

1. **Given** no input with the identifier, **When** the user runs `sonora create inputs/<id> <uri> --persistent`, **Then** the request carries `persistent: true`, the created input prints with origin `STORED`, and a create without the flag sends the same body as before.
2. **Given** a runtime input, **When** the user runs `sonora keep inputs/<id>`, **Then** the hub's persistent call is made and the input prints with origin `STORED`; a second `keep` prints it unchanged.
3. **Given** `--persistent` together with `--auto-remove`, **When** the user runs the command, **Then** it is a usage error (exit `2`) and no request is sent.
4. **Given** an identifier the hub does not know, **When** the user runs `sonora keep inputs/<id>`, **Then** the command reports the input as not found and exits `5`.
5. **Given** a 409 with reason `INPUT_ID_TAKEN`, `INPUT_UNPLAYABLE` or `INPUT_STORE_FULL`, **When** the user runs `create` or `keep`, **Then** a short message says what to do and the exit code is `14`.

[Source: .specify/archive/tiny/2026-11-22-stored-inputs/stored-inputs.md -> "What"]

---

### Edge Cases

- What happens when the hub is unreachable (host down, wrong address, network partition)?
  The user must see a clear, actionable error rather than the command hanging indefinitely
  or crashing. [Source: .specify/archive/009-tts-commands/spec.md -> "Hub unreachable or not answering"; .specify/archive/001-list-outputs/spec.md -> "What happens when the hub is unreachable"; .specify/archive/002-outputs-get/spec.md -> "What happens when the hub is unreachable"; .specify/archive/003-inputs-list-get/spec.md -> "What happens when the hub is unreachable"; .specify/archive/004-routes-list-get/spec.md -> "What happens when the hub is unreachable"; .specify/archive/005-groups-list-get/spec.md -> "What happens when the hub is unreachable"; .specify/archive/006-play-command/spec.md -> "What happens when the hub is unreachable or does not respond within a bounded time"]
- What happens when the hub takes too long to respond? The command must fail with a clear
  timeout error rather than waiting indefinitely. [Source: .specify/archive/001-list-outputs/spec.md -> "What happens when the hub takes too long to respond?"; .specify/archive/002-outputs-get/spec.md -> "What happens when the hub takes too long to respond?"; .specify/archive/003-inputs-list-get/spec.md -> "What happens when the hub takes too long to respond?"; .specify/archive/004-routes-list-get/spec.md -> "What happens when the hub takes too long to respond?"; .specify/archive/005-groups-list-get/spec.md -> "What happens when the hub takes too long to respond?"; .specify/archive/006-play-command/spec.md -> "What happens when the hub is unreachable or does not respond within a bounded time"]
- What happens when the hub returns an error response (e.g. 5xx)? The user must see a
  clear message distinguishing "the hub reported a problem" from a plain network problem
  (and, for the get command, from a "not found" result).
  [Source: .specify/archive/001-list-outputs/spec.md -> "What happens when the hub returns an error response"; .specify/archive/002-outputs-get/spec.md -> "What happens when the hub returns an error response other than"; .specify/archive/003-inputs-list-get/spec.md -> "What happens when the hub returns an error response other than"; .specify/archive/004-routes-list-get/spec.md -> "What happens when the hub returns an error response other than"; .specify/archive/005-groups-list-get/spec.md -> "What happens when the hub returns an error response other than"]
- What happens when the hub returns data that doesn't match the expected shape (e.g. a
  malformed or unexpected response)? The command must fail clearly rather than displaying
  garbled or silently incomplete output. For the TTS cache statistics, an absent or null `entriesByProvider` is not malformed and is shown as an empty breakdown (FR-105). [Source: .specify/archive/009-tts-commands/spec.md -> "Malformed success body"]
  [Source: .specify/archive/001-list-outputs/spec.md -> "What happens when the hub returns data that doesn't match the expected shape"; .specify/archive/002-outputs-get/spec.md -> "What happens when the hub returns data that doesn't match the expected shape"; .specify/archive/003-inputs-list-get/spec.md -> "What happens when the hub returns data that doesn't match the expected shape"; .specify/archive/004-routes-list-get/spec.md -> "What happens when the hub returns data that doesn't match the expected shape"; .specify/archive/005-groups-list-get/spec.md -> "What happens when the hub returns data that doesn't match the expected shape"; .specify/archive/006-play-command/spec.md -> "What happens when the hub returns a response that doesn't match the expected shape"]
- What happens when an output has a physically disconnected/unavailable device but is
  still enabled? Its availability state must be visibly distinguishable from an available
  output. [Source: .specify/archive/001-list-outputs/spec.md -> "What happens when an output has a physically disconnected/unavailable device"; .specify/archive/002-outputs-get/spec.md -> "What happens when the requested output exists but is enabled with unavailable hardware?"]
- What happens when a get command is given a missing or malformed identifier? A path with no id (`sonora get outputs`) is the valid collection form, a bare `get` with no resource lists the valid resources (FR-066), and an id outside the identifier pattern is a usage error (FR-063); the user must see a
  clear usage error rather than a confusing failure.
  [Source: .specify/archive/002-outputs-get/spec.md -> "What happens when no identifier is provided?"; .specify/archive/003-inputs-list-get/spec.md -> "What happens when no input identifier is provided to the get command?"; .specify/archive/004-routes-list-get/spec.md -> "What happens when no route identifier is provided to the get command?"; .specify/archive/005-groups-list-get/spec.md -> "What happens when no group identifier is provided to the get command?"]
- What happens when a static (YAML-configured) input is returned? It has no creation
  timestamp; this must be displayed clearly (e.g. "n/a") rather than as an error or blank
  field. [Source: .specify/archive/003-inputs-list-get/spec.md -> "What happens when a static (YAML-configured) input is returned?"]
- What happens when an invalid status value is supplied to the routes list command's status filter? The user must see a clear usage or hub-reported error rather than a silently empty or incorrect result.
  [Source: .specify/archive/004-routes-list-get/spec.md -> "What happens when an invalid status value is supplied to the list command's status filter?"]
- What happens when a route's playback has not yet started? Its playback-started timestamp must be displayed clearly (e.g. `null`) rather than as an error or blank field, consistent with how `inputs get` shows a static input's missing creation timestamp.
  [Source: .specify/archive/004-routes-list-get/spec.md -> "What happens when a route's playback has not yet started?"]
- What happens when a group has zero member outputs? Its member output list must be displayed clearly as empty (`outputIds: []`) rather than as an error or omitted field.
  [Source: .specify/archive/005-groups-list-get/spec.md -> "What happens when a group has zero member outputs?"]
- What happens when the URI is malformed or of an unsupported scheme? The play command must fail with a clear validation error rather than an ambiguous or generic failure. [Source: .specify/archive/006-play-command/spec.md -> "What happens when the URI is malformed or of an unsupported scheme?"]
- What happens when the play target identifier does not exist as an output or a group? The command must fail with a clear "target not found" error distinct from other failure types; an identifier matching neither is the same case. [Source: .specify/archive/006-play-command/spec.md -> "What happens when the target output or group identifier does not exist?"; .specify/archive/006-play-command/spec.md -> "What happens when the target identifier matches neither an existing output nor an existing group?"]
- What happens when neither `uri` nor `targetId` is supplied to the play command? The user must see a clear usage error explaining what is required, not a confusing failure. [Source: .specify/archive/006-play-command/spec.md -> "What happens when neither `uri` nor `targetId` is supplied?"]
- What happens when the hub accepts the request but cannot create the underlying route (e.g. target already at capacity)? The command must fail with a clear "route creation failed" message distinct from "not found" and from validation errors. [Source: .specify/archive/006-play-command/spec.md -> "What happens when the hub accepts the request but cannot create the underlying route"]
- What happens when the URI is well-formed but unreachable (e.g. dead stream link, DNS failure on the hub's side)? The command must fail with a clear error distinguishing "the source could not be reached" from a local network problem or a hub bug. [Source: .specify/archive/006-play-command/spec.md -> "What happens when the URI is well-formed but unreachable"]
- What happens when the hub's upstream playback service is temporarily unavailable? The command must fail with a clear, distinguishable "service unavailable" error. [Source: .specify/archive/006-play-command/spec.md -> "What happens when the hub's upstream playback service is temporarily unavailable?"]
- What happens when a resource path's prefix names a real resource type but the id after the slash doesn't exist (e.g. `sonora get outputs/does-not-exist`)? The command must fail with the same "not found" error the former `sonora outputs get does-not-exist` produced. [Source: .specify/archive/007-refactor-cli-commands/spec.md -> "What happens when a resource path's prefix names a real resource type but the id after the slash doesn't exist"]
- What happens when a resource path has no slash and no id, only the bare resource name (e.g. `sonora get outputs`)? This is the valid collection form, not an error. [Source: .specify/archive/007-refactor-cli-commands/spec.md -> "What happens when a resource path has no slash and no id at all"]
- What happens when the user runs a bare verb with nothing after it (e.g. `sonora get`)? The command must fail with a clear usage error listing the valid resources. [Source: .specify/archive/007-refactor-cli-commands/spec.md -> "What happens when the user runs a bare verb with nothing after it"]
- What happens when a resource path carries a stray slash or extra segment (`sonora get out/foo/bar`)? Since ids match `^[a-zA-Z0-9_-]{1,255}$` and never contain `/`, an extra slash unambiguously means the argument is malformed; the command must fail with a clear usage error rather than guessing which segment is the id. [Source: .specify/archive/007-refactor-cli-commands/spec.md -> "What happens when the user mixes an alias and a full name incorrectly"]
- What happens when the collection filter flags (`--include-disabled`, `--status`, `--input-id`, `--target-id`) are combined with the `get`/`list` verbs? They must continue to filter exactly as before — only the verb/resource-path shape changed. [Source: .specify/archive/007-refactor-cli-commands/spec.md -> "What happens when existing collection filter flags"]
- What happens to `--json`, `--hub-url`, and `--verbose` on the refactored commands? They must keep working, unchanged. [Source: .specify/archive/007-refactor-cli-commands/spec.md -> "What happens to `--json`, `--hub-url`, and `--verbose`?"]
- What happens when the input path (`inputs/<id>` or `in/<id>`) names no existing input? The command must fail with a clear "input not found" error, distinct from "target not found" (exit `11`). [Source: .specify/archive/008-route-command/spec.md -> "What happens when the input path does not correspond to an existing input?"]
- What happens when the target path names an identifier that matches no resource of exactly that type, even though a resource of the other type has that identifier? The command must fail with a clear "target not found" error (exit `12`) and must never retarget the other type. [Source: .specify/archive/008-route-command/spec.md -> "What happens when the target path (`outputs/<id>` or `groups/<id>`) does not correspond"]
- What happens when the input or target path's prefix is not an accepted one (e.g. `sonora route outputs/x outputs/y`)? The command must fail with a usage error that names the invalid prefix, rather than guessing a resource type. [Source: .specify/archive/008-route-command/spec.md -> "What happens when the input path's prefix is not `inputs`/`in`"]
- What happens when neither the input path nor the target path is supplied to the route command? The user must see a usage error explaining what is required, not a confusing failure. [Source: .specify/archive/008-route-command/spec.md -> "What happens when neither the input path nor the target path is supplied?"]
- What happens when the hub accepts the request but cannot create the route (e.g. the input is disabled, or the target is already at capacity)? The command must fail with a "route creation failed" error (exit `8`), distinct from "not found" and from validation errors. [Source: .specify/archive/008-route-command/spec.md -> "What happens when the hub accepts the request but cannot create the underlying route"]
- What happens when the input is already routed to that target, or the input and target identifiers refer to the same underlying entity? The command must surface whatever outcome the hub reports, rather than silently blocking or deduplicating the request. [Source: .specify/archive/008-route-command/spec.md -> "What happens when the input identifier and the target identifier refer to the same underlying entity"]
- **TTS extension not installed, disabled, rejected, or inert on the hub**: the TTS operations do not exist on that hub and answer 404. Every TTS command MUST fail with a clear message saying the hub does not offer text-to-speech and, where the extension inventory can tell, why. It MUST NOT show a generic "not found" that could be mistaken for an unknown output or group (FR-110). [Source: .specify/archive/009-tts-commands/spec.md -> "TTS extension not installed, disabled, rejected, or inert on the hub"]
- **Extension inventory also answers 404**: the hub serves the inventory from its control API (REST) extension, not from core. So a 404 there means the configured address is not serving the hub API at all: the hub URL points somewhere else, or the REST extension is missing or failed to load. The CLI says so, names the URL it used, and exits with the network/connectivity class instead of blaming TTS. [Source: .specify/archive/009-tts-commands/spec.md -> "Extension inventory also answers 404"]
- **Extension inventory unreachable or broken for any other reason** (network failure, 5xx, malformed body): the command still reports "TTS not available on this hub", with no reason, and keeps the same exit class. [Source: .specify/archive/009-tts-commands/spec.md -> "Extension inventory unreachable or broken for any other reason"]
- **Inventory lists `tts` as active, but the TTS operation answered 404**: the hub's extension and this CLI disagree on the TTS API. The CLI reports it as a version mismatch, not as "not installed". [Source: .specify/archive/009-tts-commands/spec.md -> "Inventory lists `tts` as active, but the TTS operation answered 404"]
- **Unknown target output or group**: the hub rejects the request with its `TARGET_NOT_FOUND` code; the CLI reports "target not found" and exits `12`, as FR-113 maps it. [Source: .specify/archive/009-tts-commands/spec.md -> "Unknown target output or group"]
- **Text longer than the hub allows**: the maximum length is configurable on the hub (default 500 characters), so the CLI does not enforce a limit of its own; it reports the hub's `INVALID_REQUEST` rejection and message as a validation error. [Source: .specify/archive/009-tts-commands/spec.md -> "Text longer than the hub allows"]
- **Provider timeout, rate limit, provider error, or unconvertible audio**: the hub answers "service unavailable" with one of `PROVIDER_TIMEOUT`, `PROVIDER_RATE_LIMITED`, `PROVIDER_ERROR`, `FORMAT_NORMALIZATION_FAILED`. The CLI reports these as a service-unavailable failure, shows the hub's message, and exposes the code so scripts can tell them apart. The CLI does not retry automatically. [Source: .specify/archive/009-tts-commands/spec.md -> "Provider timeout, rate limit, provider error, or unconvertible audio"]
- **Slow synthesis**: the hub synthesizes before it answers, and its default provider timeout is 10 seconds. The speak command MUST wait long enough to receive the hub's own timeout error rather than giving up first and misreporting it as a network failure — yet it MUST remain bounded. [Source: .specify/archive/009-tts-commands/spec.md -> "Slow synthesis"]
- **Error code the CLI does not recognise** (the hub adds a new one later): the command still fails with the correct exit class for the HTTP status and shows the code and message verbatim. [Source: .specify/archive/009-tts-commands/spec.md -> "Error code the CLI does not recognise"]
- **Error body in the core API's shape, or no parseable body at all** (e.g. a proxy error page): the command fails with a clear hub-error message rather than garbled output. [Source: .specify/archive/009-tts-commands/spec.md -> "Error body in the core API's shape, or no parseable body at all"]
- **Text beginning with a dash**: the user can still speak it (e.g. by placing it after a `--` end-of-flags marker) without it being parsed as a flag. A text argument that is exactly `-` always means standard input (FR-102). To speak a lone dash, pipe it in. [Source: .specify/archive/009-tts-commands/spec.md -> "Text beginning with a dash"]
- **Standard input empty, whitespace-only, or unreadable when the text is `-`**: usage error, and no request is sent. [Source: .specify/archive/009-tts-commands/spec.md -> "Standard input empty, whitespace-only, or unreadable when the text is `-`"]
- **Positional arguments in the wrong order or extra positional arguments**: usage error. [Source: .specify/archive/009-tts-commands/spec.md -> "Positional arguments in the wrong order or extra positional arguments"]
- **`--provider`, `--voice` or `--language` given an empty or whitespace-only value**: usage error before any request is sent, on every command that takes the flag. [Source: .specify/archive/009-tts-commands/spec.md -> "`--provider`, `--voice` or `--language` given an empty or whitespace-only value"]
- **Hub configured with a provider timeout longer than 10 seconds**: the default 15-second bound (FR-114) is not enough, and the user sees a CLI network timeout instead of the hub's own error. `--timeout` (FR-115) lets the user raise the bound for that invocation; other TTS commands keep their standard 5-second bound, unaffected. [Source: .specify/archive/009-tts-commands/spec.md -> "Hub configured with a provider timeout longer than 10 seconds"]
- **`--timeout` given a zero, negative, or unparsable value**: usage error before any request is sent. [Source: .specify/archive/009-tts-commands/spec.md -> "`--timeout` given a zero, negative, or unparsable value"]

- **`list tts-voices` against a provider type that lists no voices, an unknown provider, a bad filter, or `--engine` on a Gemini entry**: the hub answers 400; the CLI shows the hub's message through the TTS failure path, exits with the class FR-113 gives its code, and prints nothing to standard output. [Source: .specify/archive/tiny/2026-09-27-list-tts-voices-command/list-tts-voices-command.md -> Requirements 4]
- **A voice with absent fields** (for example a Gemini voice's `language`, or a missing `gender`): every field is still printed, an absent one as `null`, never as an empty string; an empty voice list prints an explicit empty list. [Source: .specify/archive/tiny/2026-09-27-list-tts-voices-command/list-tts-voices-command.md -> Requirements 5]
- **`list tts-voices` with a missing or blank `--provider`, a blank `--language` or `--engine`, or a positional argument; or `get tts-voices` or `list tts-voices/<anything>`**: usage error before any request is sent; the last two point to `sonora list tts-voices`. [Source: .specify/archive/tiny/2026-09-27-list-tts-voices-command/list-tts-voices-command.md -> Requirements 2, 3]
- **A build or release script still injects the version under the old module path**: the build succeeds silently but the version stays at the default — this MUST be caught by a test or release check (User Story 52, scenario 2). [Source: .specify/archive/010-public-hub-package/spec.md -> "A build or release script still injects the version under the old module path"]
- **A stray reference to the old module path or the old internal client location remains in code, tests, scripts or docs**: an automated test MUST fail on it. [Source: .specify/archive/010-public-hub-package/spec.md -> "A stray reference to the old module path"]
- **A CLI-only concept (exit codes, flags, rendering, config discovery) is still reachable from the public client**: this is a boundary violation and MUST be removed before release. [Source: .specify/archive/010-public-hub-package/spec.md -> "A CLI-only concept"]
- **A consumer tries to import CLI internals (config discovery, renderers)**: this MUST remain impossible; only the client package and published spec are public. [Source: .specify/archive/010-public-hub-package/spec.md -> "A consumer tries to import CLI internals"]
- **The error-classification messages must stay free of CLI-specific wording**: they must not mention flags, exit codes, environment variables, config files or "this CLI", so they read correctly when surfaced by a non-CLI consumer such as an AI assistant. Today the TTS "hub address" and "version mismatch" messages break this rule; the CLI must add that wording itself instead (FR-136). [Source: .specify/archive/010-public-hub-package/spec.md -> "The error-classification messages must stay free of CLI-specific wording"]
- **Contract tests move with the client**: they MUST keep running in this repo as the conformance gate, not be dropped during the move. [Source: .specify/archive/010-public-hub-package/spec.md -> "Contract tests move with the client"]
- **Play display name with characters outside ASCII** (Cyrillic titles, an em dash): shown as written in YAML and JSON, not escaped into something unreadable. [Source: .specify/archive/011-hub-0128-alignment/spec.md -> "Display name with characters outside ASCII"]
- **Play display name present but empty**: treated like an absent one; no field is shown. [Source: .specify/archive/011-hub-0128-alignment/spec.md -> "Display name present but empty"]
- **The user gave `--display-name`**: the hub returns that name; the CLI shows what the hub returns, not what the user sent. [Source: .specify/archive/011-hub-0128-alignment/spec.md -> "The user gave `--display-name`"]
- **Transfer refused with `ROUTE_CHANGING_OUTPUTS`, then with a different reason on a retry**: the CLI stops retrying and reports the second refusal as it would any other. [Source: .specify/archive/011-hub-0128-alignment/spec.md -> "Transfer refused with `ROUTE_CHANGING_OUTPUTS`, then with a different reason on a retry"]
- **Transfer refused with `ROUTE_CHANGING_OUTPUTS`, then the route is gone (404) on a retry**: reported as the route not being found, as before. [Source: .specify/archive/011-hub-0128-alignment/spec.md -> "Transfer refused with `ROUTE_CHANGING_OUTPUTS`, then the route is gone (404) on a retry"]
- **Network failure or timeout during a transfer retry**: reported as a network failure, as before; the CLI does not retry network failures. [Source: .specify/archive/011-hub-0128-alignment/spec.md -> "Network failure or timeout during a retry"]
- **`ROUTE_CHANGING_OUTPUTS` on a command other than transfer**: the contract documents it only for a transfer. If another command receives it, the CLI does not retry, but the error message still says the refusal is transient. [Source: .specify/archive/011-hub-0128-alignment/spec.md -> "`ROUTE_CHANGING_OUTPUTS` on a command other than transfer"]
- **Unknown future refusal reasons**: still reported as a generic conflict, exit 14, with no retry. [Source: .specify/archive/011-hub-0128-alignment/spec.md -> "Unknown future reasons"]
- **Mixed case**: `--join-mode duck_others`, `Duck_Others` and `DUCK_OTHERS` all mean the same mode and are sent upper-cased, as `--status` on `list routes` does today. The hyphenated spelling that sonora-tts uses (`duck-others`) is not an alias: it is a usage error that lists the valid modes. [Source: .specify/archive/012-join-modes/spec.md -> "Mixed case"]
- **Empty value** (`--join-mode ""`): a usage error, not "use the default". Leaving the flag out is the way to use the default. [Source: .specify/archive/012-join-modes/spec.md -> "Empty value"]
- **A join mode the hub refuses**: a hub older than 0.1.21 ignores or rejects the field. Whatever it answers is reported as any other hub answer is today. The CLI does not check the hub's version first. [Source: .specify/archive/012-join-modes/spec.md -> "A join mode the hub refuses"]
- **Join mode on a group target**: allowed; the hub applies it on every member output. [Source: .specify/archive/012-join-modes/spec.md -> "Join mode on a group target"]
- **`--join-mode` with `--output-volume` and `--display-name` on `play`**: all three are sent together. [Source: .specify/archive/012-join-modes/spec.md -> "`--join-mode` with `--output-volume` and `--display-name` on `play`"]
- **The output disappears between the two requests of `--include-routes`**: the routes request answers 404, and the CLI reports it as an older hub. This window is narrow, and the message tells the user what to check. It is accepted, not handled. [Source: .specify/archive/012-join-modes/spec.md -> "The output disappears between the two requests of `--include-routes`"]
- **The routes request fails for another reason** (network failure, 5xx, malformed body): reported as that failure, with its usual exit code. Nothing is printed on stdout, not even the output's record. [Source: .specify/archive/012-join-modes/spec.md -> "The routes request fails for another reason"]
- **A hub older than 0.1.21 for `get routes/<id>`**: the route has no join mode and no outputs. The YAML leaves those two fields out instead of showing empty values. The JSON keeps printing what it prints today. [Source: .specify/archive/012-join-modes/spec.md -> "A hub older than 0.1.21 for `get routes/<id>`"]
- **`--include-routes` on another resource or on a list** (`get outputs`, `get groups/<id>`, `get inputs/<id>`): a usage error before any request is sent. [Source: .specify/archive/012-join-modes/spec.md -> "`--include-routes` on another resource or on a list"]
- **Many routes on one output**: shown in full, in the hub's order. There is no paging. [Source: .specify/archive/012-join-modes/spec.md -> "Many routes on one output"]
- **Unknown future join mode in a response**: shown as the hub sent it. The flags still accept only the three documented modes, until the contract adds more. [Source: .specify/archive/012-join-modes/spec.md -> "Unknown future join mode in a response"]
- **A 200 body for hub info that is malformed**: a body that is not JSON, or lacks `product`, `version` or `apiVersion`, is a malformed-response error, not an empty record. [Source: .specify/archive/013-info-and-tts-warm/spec.md -> "A 200 body for hub info that is not JSON"]
- **Product other than multiroom**: hub info shows `product` as the hub sent it; a value other than `multiroom` is shown, not hidden or rejected. [Source: .specify/archive/013-info-and-tts-warm/spec.md -> "Hub info shows `product` as the hub sent it"]
- **Empty warm text**: warm text that is empty or only whitespace is rejected by the CLI as a usage error before any request, as `speak` does. [Source: .specify/archive/013-info-and-tts-warm/spec.md -> "Warm text that is empty or only whitespace"]
- **Missing time fields**: a hub whose `startedAt` or `uptimeSeconds` is missing still decodes; the missing field is shown as absent, not as zero. [Source: .specify/archive/013-info-and-tts-warm/spec.md -> "A hub whose `startedAt` or `uptimeSeconds` is missing"]
- **Warm time bound**: warm-up takes as long as the provider takes; the same default time bound and `--timeout` as `speak` apply. [Source: .specify/archive/013-info-and-tts-warm/spec.md -> "Warm-up takes as long as the provider takes"]

- **The radio extension is not serving**: every radio command fails with "radio is not available on this hub" and the reason the extension inventory gives: not installed (noting when extension loading is switched off), installed but disabled, failed to load (with the hub's reason), loaded but inactive, or active yet answering 404 (an extension older than 0.2.0, which can play but not search). An inventory that also answers 404 means the hub URL is wrong; an inventory that cannot be read leaves the plain "radio is not available" message. [Source: .specify/archive/014-radio-station-search/spec.md -> "The radio extension is not serving"]
- **The catalogue is unavailable**: a 503 with a retry delay fails the command, which says the catalogue is temporarily unavailable and when to retry, or "retry later" when the hub gives no readable delay. A catalogue answering badly (502) is reported as an upstream failure. [Source: .specify/archive/014-radio-station-search/spec.md -> "The catalogue is unavailable"]
- **A page size outside 1–100 or an offset outside 0–10000**: rejected by the CLI as a usage error (exit 2) before any request. The hub refuses, never reduces, an out-of-range value; the CLI does not quietly clamp it either. [Source: .specify/archive/014-radio-station-search/spec.md -> "A page size outside 1–100"]
- **An offset past the last result** gives an empty page, not an error. [Source: .specify/archive/014-radio-station-search/spec.md -> "An offset past the last result"]
- **Station names, tags and addresses are community-edited text**: they are printed as plain text and cannot move the cursor, change colours or clear the terminal. [Source: .specify/archive/014-radio-station-search/spec.md -> "Station names, tags and addresses are community-edited text"]
- **Absent station fields** (no codec, no country, no homepage) are shown as null, not as zero or empty text. [Source: .specify/archive/014-radio-station-search/spec.md -> "Absent station fields"]
- **A 200 answer whose body is not the declared shape** is a malformed-response error. [Source: .specify/archive/014-radio-station-search/spec.md -> "A 200 answer whose body is not the declared shape"]
- **The hub is unreachable**: the same unreachable-hub error and exit code as every other command. [Source: .specify/archive/014-radio-station-search/spec.md -> "The hub is unreachable"]

- **Keeping an input that is already kept**: `keep` succeeds and prints the input unchanged, with origin `STORED`. [Source: .specify/archive/tiny/2026-11-22-stored-inputs/stored-inputs.md -> Verification]
- **An input from a hub older than 0.1.35**: the record has no `origin`, so the YAML omits that line; the JSON output is unchanged. [Source: .specify/archive/tiny/2026-11-22-stored-inputs/stored-inputs.md -> Requirement 4]

## Requirements

### Functional Requirements

- **FR-001**: The CLI MUST provide a command that lists the hub's audio outputs. [Source: .specify/archive/001-list-outputs/spec.md -> FR-001]
- **FR-002**: By default, the command MUST show only enabled outputs. [Source: .specify/archive/001-list-outputs/spec.md -> FR-002]
- **FR-003**: The command MUST provide an option for the user to additionally include
  disabled outputs in the results. [Source: .specify/archive/001-list-outputs/spec.md -> FR-003]
- **FR-004**: For each output returned, the command MUST display: its identifier, its
  human-readable display name, its current volume level, its current mute state, whether
  it is enabled, and whether its underlying hardware is currently connected
  (availability) — in both the default human-readable view and the machine-readable
  (--json) view. The same applies to the single output returned by the get command.
  [Source: .specify/archive/001-list-outputs/spec.md -> FR-004; .specify/archive/002-outputs-get/spec.md -> FR-004]
- **FR-005**: An output that is enabled but whose hardware is currently unavailable MUST
  be visibly distinguishable from an available output in the default human-readable view.
  [Source: .specify/archive/001-list-outputs/spec.md -> FR-005; .specify/archive/002-outputs-get/spec.md -> FR-005]
- **FR-006**: Each outputs, inputs, routes and groups command's default output, and the play and route commands' default output, MUST be human-readable and legible without any
  additional flags or tools. [Source: .specify/archive/001-list-outputs/spec.md -> FR-006; .specify/archive/002-outputs-get/spec.md -> FR-006; .specify/archive/003-inputs-list-get/spec.md -> FR-009; .specify/archive/004-routes-list-get/spec.md -> FR-008; .specify/archive/005-groups-list-get/spec.md -> FR-009; .specify/archive/006-play-command/spec.md -> FR-007; .specify/archive/008-route-command/spec.md -> FR-006]
- **FR-007**: Each outputs, inputs, routes and groups command, and the play and route commands, MUST provide an explicit option to switch to strict,
  script-parseable output instead of the human-readable default.
  [Source: .specify/archive/001-list-outputs/spec.md -> FR-007; .specify/archive/002-outputs-get/spec.md -> FR-007; .specify/archive/003-inputs-list-get/spec.md -> FR-010; .specify/archive/004-routes-list-get/spec.md -> FR-009; .specify/archive/005-groups-list-get/spec.md -> FR-010; .specify/archive/006-play-command/spec.md -> FR-008; .specify/archive/008-route-command/spec.md -> FR-007]
- **FR-008**: Each outputs, inputs, routes and groups command MUST fail with a clear, actionable error message — not a raw
  technical error or crash — when the hub is unreachable. [Source: .specify/archive/001-list-outputs/spec.md -> FR-008; .specify/archive/002-outputs-get/spec.md -> FR-009; .specify/archive/003-inputs-list-get/spec.md -> FR-012; .specify/archive/004-routes-list-get/spec.md -> FR-011; .specify/archive/005-groups-list-get/spec.md -> FR-012; .specify/archive/006-play-command/spec.md -> FR-010]
- **FR-009**: Each outputs, inputs, routes and groups command, and the play, route, speak, get tts-cache and clear tts-cache commands, MUST fail with a clear, actionable error message when the hub
  does not respond within a bounded amount of time; it MUST NOT wait indefinitely. Each
  command MUST make a single attempt and MUST NOT automatically retry on network or
  timeout failure. [Source: .specify/archive/009-tts-commands/spec.md -> FR-014; .specify/archive/001-list-outputs/spec.md -> FR-009; .specify/archive/002-outputs-get/spec.md -> FR-010; .specify/archive/003-inputs-list-get/spec.md -> FR-013; .specify/archive/004-routes-list-get/spec.md -> FR-012; .specify/archive/005-groups-list-get/spec.md -> FR-013; .specify/archive/006-play-command/spec.md -> FR-010; .specify/archive/008-route-command/spec.md -> FR-009]
- **FR-010**: Each outputs, inputs, routes and groups command MUST fail with a clear, actionable error message when the hub
  returns an error response, distinguishing that case from a connectivity failure and,
  for the get commands, from a "not found" result.
  [Source: .specify/archive/001-list-outputs/spec.md -> FR-010; .specify/archive/002-outputs-get/spec.md -> FR-011; .specify/archive/003-inputs-list-get/spec.md -> FR-014; .specify/archive/004-routes-list-get/spec.md -> FR-013; .specify/archive/005-groups-list-get/spec.md -> FR-014]
- **FR-011**: Each outputs, inputs, routes and groups command MUST exit with a status that is programmatically distinguishable
  between success, a usage error, a hub-reported error, and a connectivity/network error
  and, for the get commands, "not found". For the play command, a validation error, a route creation failure, an unreachable source and an unavailable upstream service each also have their own exit code (6–10); "target not found" uses the not-found exit code (5). For the route command, the exit codes are those of FR-079.
  [Source: .specify/archive/001-list-outputs/spec.md -> FR-011; .specify/archive/002-outputs-get/spec.md -> FR-012; .specify/archive/003-inputs-list-get/spec.md -> FR-015; .specify/archive/004-routes-list-get/spec.md -> FR-014; .specify/archive/005-groups-list-get/spec.md -> FR-015; .specify/archive/006-play-command/spec.md -> FR-011; .specify/archive/008-route-command/spec.md -> FR-010]
- **FR-012**: The command MUST clearly communicate when there are zero outputs to show,
  distinct from a failure. [Source: .specify/archive/001-list-outputs/spec.md -> FR-012]
- **FR-013**: Each outputs, inputs and routes command, and the play command, MUST reject any output, input, route, group or playback data from the hub that does not match the
  expected structure with a clear error, rather than displaying partial or malformed
  results. [Source: .specify/archive/001-list-outputs/spec.md -> FR-013; .specify/archive/002-outputs-get/spec.md -> FR-013; .specify/archive/003-inputs-list-get/spec.md -> FR-017; .specify/archive/004-routes-list-get/spec.md -> FR-016; .specify/archive/005-groups-list-get/spec.md -> FR-017; .specify/archive/006-play-command/spec.md -> FR-012; .specify/archive/008-route-command/spec.md -> FR-011]
- **FR-014**: The CLI MUST provide a command that retrieves a single audio output from
  the hub by its identifier. [Source: .specify/archive/002-outputs-get/spec.md -> FR-001]
- **FR-016**: The get command MUST return the requested output regardless of its
  enabled/disabled state (unlike the list command's default filtering).
  [Source: .specify/archive/002-outputs-get/spec.md -> FR-003]
- **FR-017**: The get command MUST fail with a clear, actionable "output not found"
  message, distinguishable from other failure types, when no output with the given
  identifier exists on the hub. [Source: .specify/archive/002-outputs-get/spec.md -> FR-008]
- **FR-018**: The CLI MUST provide a command that lists the hub's audio inputs. [Source: .specify/archive/003-inputs-list-get/spec.md -> FR-001]
- **FR-019**: By default, the inputs list command MUST show only enabled inputs. [Source: .specify/archive/003-inputs-list-get/spec.md -> FR-002]
- **FR-020**: The inputs list command MUST provide an option for the user to additionally include disabled inputs in the results. [Source: .specify/archive/003-inputs-list-get/spec.md -> FR-003]
- **FR-021**: For each input returned by the inputs list command, the CLI MUST display: its identifier, its human-readable display name, its source URI, whether it is enabled, its source type (static or ephemeral), whether it auto-removes when its route stops, whether it supports pause/resume, and its creation timestamp (or an explicit indication that none exists) — in both the default human-readable view and the machine-readable (--json) view. [Source: .specify/archive/003-inputs-list-get/spec.md -> FR-004]
- **FR-022**: The CLI MUST provide a command that retrieves a single audio input from the hub by its identifier. [Source: .specify/archive/003-inputs-list-get/spec.md -> FR-005]
- **FR-024**: The inputs get command MUST return the requested input regardless of its enabled/disabled state (unlike the inputs list command's default filtering). [Source: .specify/archive/003-inputs-list-get/spec.md -> FR-007]
- **FR-025**: For the input returned by the inputs get command, the CLI MUST display the same set of attributes specified in FR-021, in both the default human-readable view and the machine-readable (--json) view. [Source: .specify/archive/003-inputs-list-get/spec.md -> FR-008]
- **FR-026**: The inputs get command MUST fail with a clear, actionable "input not found" message, distinguishable from other failure types, when no input with the given identifier exists on the hub. [Source: .specify/archive/003-inputs-list-get/spec.md -> FR-011]
- **FR-027**: The inputs list command MUST clearly communicate when there are zero inputs to show, distinct from a failure. [Source: .specify/archive/003-inputs-list-get/spec.md -> FR-016]
- **FR-028**: The CLI MUST provide a command that lists the hub's audio routes. [Source: .specify/archive/004-routes-list-get/spec.md -> FR-001]
- **FR-029**: By default, the routes list command MUST show all routes regardless of status, since routes (unlike inputs and outputs) have no enabled/disabled concept to filter on. [Source: .specify/archive/004-routes-list-get/spec.md -> FR-002]
- **FR-030**: The routes list command MUST provide optional filters to narrow results by status, by source input identifier, and by target identifier; when more than one filter is supplied, only routes matching all supplied filters MUST be returned. [Source: .specify/archive/004-routes-list-get/spec.md -> FR-003]
- **FR-031**: For each route returned by the routes list command, the CLI MUST display: its identifier, its source input identifier, its target identifier, its target type (single output or output group), and its current status — in both the default human-readable view and the machine-readable (--json) view. [Source: .specify/archive/004-routes-list-get/spec.md -> FR-004]
- **FR-032**: The CLI MUST provide a command that retrieves a single audio route from the hub by its identifier. [Source: .specify/archive/004-routes-list-get/spec.md -> FR-005]
- **FR-034**: For the route returned by the routes get command, the CLI MUST display: its identifier, its source input identifier, its target identifier, its target type, its current status, its creation timestamp, its playback-started timestamp (or an explicit indication that playback has not started), whether it is transferable, whether it supports pause/resume, and whether it is currently paused — in both the default human-readable view and the machine-readable (--json) view. [Source: .specify/archive/004-routes-list-get/spec.md -> FR-007]
- **FR-035**: The routes get command MUST fail with a clear, actionable "route not found" message, distinguishable from other failure types, when no route with the given identifier exists on the hub. [Source: .specify/archive/004-routes-list-get/spec.md -> FR-010]
- **FR-036**: The routes list command MUST clearly communicate when there are zero routes to show, distinct from a failure. [Source: .specify/archive/004-routes-list-get/spec.md -> FR-015]
- **FR-037**: The CLI MUST provide a command that lists the hub's audio output groups. [Source: .specify/archive/005-groups-list-get/spec.md -> FR-001]
- **FR-038**: By default, the groups list command MUST show only enabled groups. [Source: .specify/archive/005-groups-list-get/spec.md -> FR-002]
- **FR-039**: The groups list command MUST provide an option for the user to additionally include disabled groups in the results. [Source: .specify/archive/005-groups-list-get/spec.md -> FR-003]
- **FR-040**: For each group returned by the groups list command, the CLI MUST display: its identifier, its display name, its member output identifiers, its mute state, and its enabled state — in both the default human-readable view and the machine-readable (--json) view. [Source: .specify/archive/005-groups-list-get/spec.md -> FR-004]
- **FR-041**: The CLI MUST provide a command that retrieves a single audio output group from the hub by its identifier. [Source: .specify/archive/005-groups-list-get/spec.md -> FR-005]
- **FR-043**: The groups get command MUST return the requested group regardless of its enabled/disabled state. [Source: .specify/archive/005-groups-list-get/spec.md -> FR-007]
- **FR-044**: For the group returned by the groups get command, the CLI MUST display: its identifier, its display name, its member output identifiers, its mute state, and its enabled state — in both the default human-readable view and the machine-readable (--json) view. [Source: .specify/archive/005-groups-list-get/spec.md -> FR-008]
- **FR-045**: The groups get command MUST fail with a clear, actionable "group not found" message, distinguishable from other failure types, when no group with the given identifier exists on the hub. [Source: .specify/archive/005-groups-list-get/spec.md -> FR-011]
- **FR-046**: The groups list command MUST clearly communicate when there are zero groups to show, distinct from a failure. [Source: .specify/archive/005-groups-list-get/spec.md -> FR-016]
- **FR-047**: The CLI MUST provide a command that starts instant playback of an audio URI to a target output or output group, creating the underlying ephemeral input and route in a single invocation. [Source: .specify/archive/006-play-command/spec.md -> FR-001]
- **FR-048**: The play command MUST require a URI and a target identifier as input, and MUST fail with a clear usage error if either is omitted. [Source: .specify/archive/006-play-command/spec.md -> FR-002]
- **FR-053**: The play command MUST provide an optional way to set the starting volume (0-100) for the target, applied before playback begins, and MUST reject an out-of-range value with a clear usage/validation error before contacting the hub. [Source: .specify/archive/006-play-command/spec.md -> FR-004]
- **FR-054**: The play command MUST provide an optional way to set a display name for the ephemeral input created for this playback session. [Source: .specify/archive/006-play-command/spec.md -> FR-005]
- **FR-055**: On success, the play command MUST return as soon as the hub accepts the request, without waiting or polling for the route to reach any further state, and MUST display the created input's identifier, the resulting route's identifier and status exactly as returned by the hub, and a human-readable confirmation message — in both the default human-readable view and the machine-readable (--json) view. [Source: .specify/archive/006-play-command/spec.md -> FR-006]
- **FR-056**: The play command MUST fail with a clear, actionable error, distinguishable from all other failure types, for each of: a validation error (bad URI or bad request fields), a "target not found" error, a "route creation failed" error, a "source unreachable" error, and an "upstream service unavailable" error. [Source: .specify/archive/006-play-command/spec.md -> FR-009]
- **FR-057**: The CLI MUST print its usage text (the command table, the global flags and example invocations) to standard output and exit with status `0` when invoked as `sonora help`, `sonora -h`, `sonora --help`, or with no arguments. [Source: .specify/archive/tiny/2026-08-26-help-command/help-command.md]
- **FR-058**: Help dispatch MUST NOT change the existing usage errors: an unknown verb (including the removed noun-verb forms such as `sonora outputs list`) still fails with a usage error on standard error and exit status `2` (FR-011). `--version` and `-v` are still checked before help dispatch. [Source: .specify/archive/tiny/2026-08-26-help-command/help-command.md]
- **FR-059**: The CLI MUST provide `sonora get <resource>` for each of inputs, outputs, groups, and routes, returning the same collection data as the pre-refactor `sonora <resource> list`, including the collection filter flags. [Source: .specify/archive/007-refactor-cli-commands/spec.md -> FR-001]
- **FR-060**: The CLI MUST provide `sonora get <resource>/<id>` for each of inputs, outputs, groups, and routes, returning the same single-item data as the pre-refactor `sonora <resource> get <id>`. [Source: .specify/archive/007-refactor-cli-commands/spec.md -> FR-002]
- **FR-061**: The CLI MUST accept `sonora list <resource>` as an exact synonym of `sonora get <resource>` (collection form only), and MUST fail with a clear usage error if `list` is given a resource path that includes an id. [Source: .specify/archive/007-refactor-cli-commands/spec.md -> FR-003]
- **FR-062**: The CLI MUST accept the short aliases `in`, `out`, `gr`, and `rt` interchangeably with `inputs`, `outputs`, `groups`, and `routes` respectively, everywhere a resource name or resource path is accepted. [Source: .specify/archive/007-refactor-cli-commands/spec.md -> FR-004]
- **FR-063**: A resource path MUST be parsed by splitting on the first `/` into a resource-name segment and an identifier segment; the CLI MUST treat identifiers as matching `^[a-zA-Z0-9_-]{1,255}$` and MUST fail with a clear usage error if the identifier segment contains an additional `/` or otherwise doesn't match that pattern. [Source: .specify/archive/007-refactor-cli-commands/spec.md -> FR-004a]
- **FR-064**: The CLI MUST remove the old `sonora <resource> list` and `sonora <resource> get <id>` command forms for inputs, outputs, groups, and routes, rather than keeping them as hidden or deprecated aliases. [Source: .specify/archive/007-refactor-cli-commands/spec.md -> FR-005]
- **FR-065**: The CLI MUST fail with a clear, standard usage error — not a silent misinterpretation or incorrect result — when given an old-style invocation, an unrecognized resource name or alias, or a malformed resource path. The CLI is NOT required to specifically detect old-style syntax in order to name it or point to its new equivalent; a generic "unrecognized command"-style error is sufficient. [Source: .specify/archive/007-refactor-cli-commands/spec.md -> FR-006]
- **FR-066**: When `get` or `list` is given no resource argument at all (`sonora get`, `sonora list`), the CLI MUST fail with a usage error whose message enumerates the valid resource names. This is a deliberate exception to FR-065's "generic error is sufficient" allowance: there the user typed something wrong and the CLI need not diagnose it, whereas here the user typed a valid verb and nothing else, so the set of valid completions is unambiguous and the only useful thing to say. [Source: .specify/archive/007-refactor-cli-commands/spec.md -> FR-006a]
- **FR-067**: The play, route and speak commands' target argument MUST be a resource path (`outputs/<id>`, `groups/<id>`, or their aliases), and the separate `--group` and `--output` disambiguation flags MUST NOT exist on any of these commands. A path naming any other resource (e.g. `inputs/<id>`), or one without an id, is a usage error. [Source: .specify/archive/009-tts-commands/spec.md -> FR-002; .specify/archive/007-refactor-cli-commands/spec.md -> FR-007; .specify/archive/007-refactor-cli-commands/data-model.md -> "ResourcePath"; .specify/archive/008-route-command/spec.md -> FR-003]
- **FR-068**: The play command's display-name flag MUST be `--display-name`, and the command MUST fail with a clear usage error if the old `--name`, `--group`, or `--output` flags are supplied. [Source: .specify/archive/007-refactor-cli-commands/spec.md -> FR-008]
- **FR-069**: The CLI MUST preserve `--json`, `--hub-url`, and `--verbose` unchanged, in meaning and behavior, on every refactored command. [Source: .specify/archive/007-refactor-cli-commands/spec.md -> FR-009]
- **FR-070**: The CLI MUST preserve the exact displayed data (fields and values, in both the default and `--json` views) of every refactored read command — only the invocation shape changes, not what is shown. [Source: .specify/archive/007-refactor-cli-commands/spec.md -> FR-010]
- **FR-071**: The CLI MUST provide a top-level `route` command that starts routing an existing, already-defined input to a target output or output group, without creating any new input as a side effect. [Source: .specify/archive/008-route-command/spec.md -> FR-001]
- **FR-072**: The route command MUST require an input path (`inputs/<input-id>` or `in/<input-id>`) and a target path (`outputs/<target-id>` or `groups/<target-id>`, or `out/` and `gr/`) as arguments, and MUST fail with a clear usage error if either is omitted. [Source: .specify/archive/008-route-command/spec.md -> FR-002]
- **FR-073**: The route command MUST fail with a clear usage error if the input path's prefix is not `inputs` or its alias `in`. [Source: .specify/archive/008-route-command/spec.md -> FR-002a]
- **FR-074**: The route command MUST fail with a clear usage error if the target path's prefix is not one of `outputs`/`out` or `groups`/`gr`. [Source: .specify/archive/008-route-command/spec.md -> FR-002b]
- **FR-075**: When the target path names a type and identifier for which no matching resource exists, the route command MUST fail with a clear "target not found" error, even if a resource with the same identifier exists as the other type. [Source: .specify/archive/008-route-command/spec.md -> FR-003a]
- **FR-076**: When the input identifier matches no existing input, the route command MUST fail with a clear "input not found" error, distinguishable from "target not found". [Source: .specify/archive/008-route-command/spec.md -> FR-004]
- **FR-077**: On success the route command MUST display the resulting route's identifier and status exactly as the hub returned them, with a human-readable confirmation message, in both the default human-readable view and the `--json` view. [Source: .specify/archive/008-route-command/spec.md -> FR-005]
- **FR-078**: The route command MUST fail with a clear, actionable error, distinguishable from every other failure type, for each of: a usage error, "input not found", "target not found", and "route creation failed". [Source: .specify/archive/008-route-command/spec.md -> FR-008]
- **FR-079**: The route command's exit codes MUST be programmatically distinguishable across success (`0`), usage error (`2`), hub-reported error (`3`), connectivity error (`4`), validation error from a hub `400` (`6`), route creation failed (`8`), input not found (`11`), target not found (`12`) and a hub `409` refusal (`14`, FR-137). [Source: .specify/archive/008-route-command/spec.md -> FR-010; .specify/archive/008-route-command/data-model.md -> "Exit Code Table"]
- **FR-080**: The CLI MUST provide top-level `enable` and `disable` commands that take a resource path and set that resource's enabled state with one request, `PUT /api/v2/<resource>/{id}/enabled` carrying `{"enabled": true}` or `{"enabled": false}`. On success (200) the command prints the updated resource record, with the same fields and YAML/`--json` shapes as the matching `get` command, and exits `0`. This requirement covers `inputs/<input-id>`; FR-088 and FR-094 extend it to groups and outputs. Like every command below, it uses the global flags, output separation and exit-code scheme of FR-006, FR-007, FR-009, FR-011 and FR-013. [Source: .specify/archive/tiny/2026-08-28-inputs-enable-disable-command/inputs-enable-disable-command.md]
- **FR-081**: A 404 from an enable or disable request MUST be reported as "<resource> not found: <id>" (for inputs, "input not found") on standard error with the not-found exit code (`5`). A 400 MUST be reported with the hub's error detail, or a generic "the request was rejected as invalid" message when the hub gives none, with the validation exit code (`6`). Any other non-2xx, network or malformed-body failure keeps the generic classification (FR-009). Nothing is written to standard output on any failure. [Source: .specify/archive/tiny/2026-08-28-inputs-enable-disable-command/inputs-enable-disable-command.md]
- **FR-082**: `enable` and `disable` MUST fail with a usage error (exit `2`) when the identifier is missing, when extra arguments are given, or when the resource path names a kind that does not support enabling, without contacting the hub. [Source: .specify/archive/tiny/2026-08-28-inputs-enable-disable-command/inputs-enable-disable-command.md]
- **FR-083**: The CLI MUST provide `sonora delete routes/<route-id>` and `sonora stop routes/<route-id>`, which send `DELETE /api/v2/routes/{id}`; on success (204) they print the route identifier, status `stopped` and a message (YAML, or JSON with `--json`) and exit `0`. `stop routes/<route-id>` is an exact alias: same handler, output and exit codes. `stop` also has bulk forms (FR-098). Like every command below, it uses the global flags, output separation and exit-code scheme of FR-006, FR-007, FR-009, FR-011 and FR-013. [Source: .specify/archive/tiny/2026-08-28-route-delete-command/route-delete-command.md]
- **FR-084**: A 404 from a route delete MUST be reported as "route not found: <id>" with exit `5`; a 422 MUST be reported with the hub's error detail, or a generic "route stop failed" message, with the route-failed exit code (`8`); other failures keep the generic classification. A missing route identifier, extra arguments, or a resource path that `delete` does not support MUST be a usage error (exit `2`) with no request. `delete` accepts `routes` here; FR-091 adds `inputs`, and FR-098 widens `stop`. [Source: .specify/archive/tiny/2026-08-28-route-delete-command/route-delete-command.md]
- **FR-085**: The CLI MUST provide `sonora transfer routes/<route-id> <outputs|groups>/<target-id>`, which sends `POST /api/v2/routes/{id}/transfer` with `{targetId, targetType}`. On success (200) it prints the route's identifier, status and confirmation message, in the fields of FR-077; the identifier is the one the route had before the transfer, which the hub keeps since API 0.1.22 (HUB-024) (YAML, or JSON with `--json`), and exits `0`. Like every command below, it uses the global flags, output separation and exit-code scheme of FR-006, FR-007, FR-009, FR-011 and FR-013. [Source: .specify/archive/tiny/2026-08-28-route-transfer-command/route-transfer-command.md]
- **FR-086**: The first argument MUST be a `routes/<id>` path (alias `rt`) and the second an `outputs/<id>` or `groups/<id>` path (aliases `out`, `gr`); any other kind, or a missing identifier in either, MUST be a usage error (exit `2`). Before the transfer request the target MUST be checked with the per-type existence request of FR-075; a missing target fails with exit `12`. [Source: .specify/archive/tiny/2026-08-28-route-transfer-command/route-transfer-command.md]
- **FR-087**: A 404 from the transfer request MUST be reported as "route not found: <id>" with exit `5`; a 400 with the hub's error detail, or "route is not transferable", and exit `6`; a 422 with the hub's error detail, or "route transfer failed", and exit `8`. Other failures keep the generic classification, and nothing is written to standard output on a failure. [Source: .specify/archive/tiny/2026-08-28-route-transfer-command/route-transfer-command.md]
- **FR-088**: `enable` and `disable` MUST also accept `groups/<group-id>` (alias `gr`), sending `PUT /api/v2/groups/{id}/enabled` as in FR-080 and printing the updated group record. A 404 is reported as "group not found: <id>" (exit `5`); the 400, other-failure and usage-error rules of FR-081 and FR-082 apply. The behaviour of `enable`/`disable` for the other supported kinds is unchanged. [Source: .specify/archive/tiny/2026-08-29-groups-enable-disable-command/groups-enable-disable-command.md]
- **FR-089**: The CLI MUST provide `sonora set groups/<group-id> volume <n>`, which sends `PUT /api/v2/groups/{id}/volume` with `{"volume": n}` and on success (200) prints the group identifier, volume and update time (YAML, or JSON with `--json`) and exits `0`. `<n>` MUST be an integer from 0 to 100; anything else, including a negative integer (read as a value rather than a flag), is a usage error (exit `2`) with no request. A 404 is reported as "group not found: <id>" (exit `5`) and a 400 with the hub's detail (exit `6`). A missing id, a missing or misspelled `volume` word, a missing value or extra arguments is a usage error. `set` accepts `groups` and `outputs` (FR-096); `inputs` and `routes` stay usage errors. Like every command below, it uses the global flags, output separation and exit-code scheme of FR-006, FR-007, FR-009, FR-011 and FR-013. [Source: .specify/archive/tiny/2026-08-29-groups-set-volume-command/groups-set-volume-command.md]
- **FR-090**: The CLI MUST provide `sonora create inputs/<input-id> <uri> --display-name <name> [--auto-remove] [--disabled]`, which sends `POST /api/v2/inputs` with the identifier, URI, display name, `autoRemove` (default false) and `enabled` (false only with `--disabled`). On success (201) it prints the created input record and exits `0`. `--display-name` is required; a missing display name, identifier or URI, extra arguments, or a resource path that is not `inputs/<id>` is a usage error (exit `2`). A 409 (duplicate identifier) is reported with the hub's detail (for example "Input ID already exists") and the conflict exit code (`14`, FR-137); a 400 with the hub's detail and exit `6`. [Annotated by .specify/archive/tiny/2026-11-22-stored-inputs/stored-inputs.md: FR-228 adds `--persistent`; for `INPUT_ID_TAKEN`, FR-230 gives the CLI's own message in place of the hub's detail.] Like every command below, it uses the global flags, output separation and exit-code scheme of FR-006, FR-007, FR-009, FR-011 and FR-013. [Source: .specify/archive/tiny/2026-08-29-inputs-create-delete-command/inputs-create-delete-command.md]
- **FR-091**: `delete` MUST also accept `inputs/<input-id>`, sending `DELETE /api/v2/inputs/{id}`; on success (204) it prints the input identifier, a status and a message (the shape of FR-083) and exits `0`. A 404 is reported as "input not found: <id>" (exit `5`); a 400, for example a static input that cannot be deleted, with the hub's detail (exit `6`). `delete routes/<id>` is unchanged. `delete` given `outputs` or `groups` stays a usage error. [Source: .specify/archive/tiny/2026-08-29-inputs-create-delete-command/inputs-create-delete-command.md]
- **FR-092**: The CLI MUST provide `sonora get master-mute`, which sends `GET /api/v2/master-mute` and on success prints the muted state (YAML, or JSON with `--json`) and exits `0`. Master mute has no identifier and no collection: `get master-mute/<anything>`, `list master-mute` and any extra argument are usage errors (exit `2`). Like every command below, it uses the global flags, output separation and exit-code scheme of FR-006, FR-007, FR-009, FR-011 and FR-013. [Source: .specify/archive/tiny/2026-08-29-master-mute-command/master-mute-command.md]
- **FR-093**: The CLI MUST provide `sonora mute all` and `sonora unmute all`, which send `PUT /api/v2/master-mute` with `{"muted": true}` or `{"muted": false}` and print the updated state. The hub documents no 404 or 400 for either operation, so a non-2xx status or malformed body is a generic hub error (exit `3`) and a network failure is exit `4`; nothing is written to standard output on a failure. `master-mute` and `all` are literal keywords, not resource kinds, and take no further argument. [Source: .specify/archive/tiny/2026-08-29-master-mute-command/master-mute-command.md]
- **FR-094**: `enable` and `disable` MUST also accept `outputs/<output-id>` (alias `out`), sending `PUT /api/v2/outputs/{id}/enabled` as in FR-080 and printing the updated output record. A 404 is reported as "output not found: <id>" (exit `5`); the 400, other-failure and usage-error rules of FR-081 and FR-082 apply. `enable`/`disable` given a kind that is not `inputs`, `outputs` or `groups` is a usage error. [Source: .specify/archive/tiny/2026-08-29-outputs-enable-disable-command/outputs-enable-disable-command.md]
- **FR-095**: The CLI MUST provide `sonora mute` and `sonora unmute` for `outputs/<output-id>` and `groups/<group-id>`, sending `PUT /api/v2/{outputs|groups}/{id}/mute` with `{"muted": true}` or `{"muted": false}`. On success (200) they print the updated output or group record (YAML, or JSON with `--json`) and exit `0`. A 404 is reported as "output not found" or "group not found" (exit `5`) and a 400 with the hub's detail (exit `6`); a missing identifier or extra arguments is a usage error (exit `2`). `mute`/`unmute` given any kind other than `outputs` and `groups` (for example `inputs` or `routes`), other than the literal `all` of FR-093, is a usage error. Like every command below, it uses the global flags, output separation and exit-code scheme of FR-006, FR-007, FR-009, FR-011 and FR-013. [Source: .specify/archive/tiny/2026-08-29-outputs-groups-mute-command/outputs-groups-mute-command.md]
- **FR-096**: The CLI MUST provide a top-level `set` verb of the form `set <resource>/<id> <attribute> <value>`, of which `volume` is the only attribute. `sonora set outputs/<output-id> volume <n>` sends `PUT /api/v2/outputs/{id}/volume` with `{"volume": n}` and prints the output identifier, volume and update time; the value range, error mapping ("output not found", exit `5`; 400, exit `6`) and usage errors are those of FR-089. `set` given a kind other than `outputs` or `groups`, or an attribute other than `volume`, is a usage error. Like every command below, it uses the global flags, output separation and exit-code scheme of FR-006, FR-007, FR-009, FR-011 and FR-013. [Source: .specify/archive/tiny/2026-08-29-outputs-set-volume-command/outputs-set-volume-command.md]
- **FR-097**: The CLI MUST provide `sonora pause routes/<route-id>` and `sonora resume routes/<route-id>`, which send `PUT /api/v2/routes/{id}/pause` with `{"paused": true}` or `{"paused": false}`. On success (200) they print the route identifier, paused flag, status and a message (YAML, or JSON with `--json`) and exit `0`. Pausing a paused route or resuming an active one is a success, since the hub is idempotent; the CLI reads no state first. A 404 is reported as "route not found: <id>" (exit `5`); a 400 (route not active, input not pausable, other validation) with the hub's detail (exit `6`). Only `routes/<id>` (alias `rt`) is accepted; any other kind or a missing identifier is a usage error with no request. Like every command below, it uses the global flags, output separation and exit-code scheme of FR-006, FR-007, FR-009, FR-011 and FR-013. [Source: .specify/archive/tiny/2026-08-29-route-pause-resume-command/route-pause-resume-command.md]
- **FR-098**: The CLI MUST provide three bulk stop forms: bare `sonora stop routes` (`DELETE /api/v2/routes`, every route), `sonora stop outputs/<output-id>` (`DELETE /api/v2/outputs/{id}/routes`) and `sonora stop groups/<group-id>` (`DELETE /api/v2/groups/{id}/routes`). On success (200) each prints the stopped count and the stopped routes, each with its identifier, target type, target identifier and nullable stop reason (YAML, or JSON with `--json`), and exits `0`, including when none was stopped. Like every command below, it uses the global flags, output separation and exit-code scheme of FR-006, FR-007, FR-009, FR-011 and FR-013. [Source: .specify/archive/tiny/2026-09-09-stop-routes-command/stop-routes-command.md]
- **FR-099**: A 404 for an output or group MUST be reported as "output not found" or "group not found" (exit `5`); the system-wide form has no documented 404. `sonora stop routes/<route-id>` is unchanged: it is still the alias of FR-083. Extra arguments, `stop outputs` or `stop groups` without an identifier, and `stop inputs/<id>` are usage errors (exit `2`), so `stop` is no longer limited to routes. [Source: .specify/archive/tiny/2026-09-09-stop-routes-command/stop-routes-command.md]

- **FR-100**: The CLI MUST provide a `speak` command of the form `sonora speak <text> <outputs|groups>/<id>` that requests a TTS announcement of the given text on the given output or output group. [Source: .specify/archive/009-tts-commands/spec.md -> FR-001]
- **FR-101**: The command MUST reject empty or whitespace-only text with a usage error before contacting the hub. It MUST NOT impose its own maximum text length; that limit belongs to the hub. [Source: .specify/archive/009-tts-commands/spec.md -> FR-003]
- **FR-102**: When the text argument is exactly `-`, the command MUST read the text from standard input until end of input and trim trailing newlines (CR/LF). The text is sent as-is otherwise: inner line breaks are kept. FR-101's empty or whitespace-only check applies to the text read. If standard input cannot be read, that is a usage error raised before contacting the hub. [Source: .specify/archive/009-tts-commands/spec.md -> FR-003a]
- **FR-103**: The command MUST accept optional `--provider <name>`, `--voice <id>`, and `--language <tag>` options, each sent to the hub only when supplied. An option supplied with an empty or whitespace-only value MUST be a usage error. [Source: .specify/archive/009-tts-commands/spec.md -> FR-004]
- **FR-104**: On acceptance, the command MUST return immediately — without waiting or polling for playback to start or finish — and MUST display the announcement identifier, whether the audio was served from cache, and the target's queue depth, exactly as the hub reported them. [Source: .specify/archive/009-tts-commands/spec.md -> FR-005]
- **FR-105**: Although `api/openapi.json` marks no success-response field as required, the CLI MUST treat these as required and reject a body missing any of them as malformed (hub-error class, 3): `announcementId`, `cacheHit`, `queueDepth` (speak) and `totalEntries`, `totalSizeBytes`, `maxSizeBytes` (cache stats). An empty `announcementId` counts as missing. An absent or null `entriesByProvider` MUST be shown as an empty breakdown, not rejected. [Source: .specify/archive/009-tts-commands/spec.md -> FR-005a]
- **FR-106**: The CLI MUST provide `sonora get tts-cache`, which displays the TTS cache's total entry count, total size in bytes, configured maximum size in bytes, and entry count per provider. `tts-cache` is a singleton keyword like `master-mute`: it takes no id and has no `list` form. [Source: .specify/archive/009-tts-commands/spec.md -> FR-006]
- **FR-107**: The CLI MUST provide `sonora clear tts-cache [--provider <name>]`, which clears the whole TTS cache, or only the named provider's entries when `--provider` is given. On success (hub 204, no body), it prints a structured confirmation: `cleared: all`, or `cleared: provider` plus `provider: <name>`. The output is YAML by default and the same fields as a JSON object with `--json`. `--provider` follows FR-103's empty-value rule. `clear` is a new verb and applies only to `tts-cache` in this feature; any other resource is a usage error. [Source: .specify/archive/009-tts-commands/spec.md -> FR-007]
- **FR-108**: Clearing an already-empty cache MUST succeed. [Source: .specify/archive/009-tts-commands/spec.md -> FR-008]
- **FR-109**: Every new command MUST follow the verb-first command shape and strict GNU flag style, MUST support `--json`, `--hub-url`, and `--verbose` with the same meaning as on existing commands, MUST default to YAML output, MUST write errors to standard error, MUST use the existing exit codes for the failure classes it shares, and MUST have an entry in `sonora help`. [Source: .specify/archive/009-tts-commands/spec.md -> FR-009; .specify/archive/014-radio-station-search/spec.md -> FR-012]
- **FR-110**: When a TTS operation answers 404, the command MUST treat TTS as unavailable on this hub and fail with a message distinguishable from "target not found" and "provider not found". To explain why, it MUST then query the hub's extension inventory once and look up the entry whose id is `tts`:

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

  [Source: .specify/archive/009-tts-commands/spec.md -> FR-010]
- **FR-111**: The inventory MUST NOT be queried on any path other than a 404 from a TTS operation, so a successful TTS command makes exactly one request. The exit class is "TTS not available" in every row of FR-110's table except one. When the inventory itself answers 404, the fault is the hub address or the REST extension, not TTS, so the exit class is network/connectivity: the same class as an unreachable hub. Any other lookup failure is visible only with `--verbose` and does not change the exit class. [Source: .specify/archive/009-tts-commands/spec.md -> FR-010a]
- **FR-112**: When a TTS operation fails with the extension's error shape, the CLI MUST show the hub's human-readable message and MUST make the machine-readable code (`INVALID_REQUEST`, `TARGET_NOT_FOUND`, `PROVIDER_NOT_FOUND`, `PROVIDER_TIMEOUT`, `PROVIDER_RATE_LIMITED`, `PROVIDER_ERROR`, `FORMAT_NORMALIZATION_FAILED`) visible in the error output. [Source: .specify/archive/009-tts-commands/spec.md -> FR-011]
- **FR-113**: Failures MUST map to exit codes as follows: `TARGET_NOT_FOUND` → target not found (12); `INVALID_REQUEST` → validation error (6); `PROVIDER_NOT_FOUND` → generic not found (5); any provider-side (503) failure → service unavailable (10); TTS not offered by the hub → a new "TTS not available" code, 13; hub unreachable, or the extension inventory answering 404 (FR-111) → network error (4). An unrecognised code MUST fall back to the class implied by the HTTP status. [Source: .specify/archive/009-tts-commands/spec.md -> FR-012]
- **FR-114**: The speak command's wait for a response MUST be bounded, and MUST default to long enough to receive the hub's default 10-second synthesis-timeout error rather than timing out first. Other TTS commands keep the CLI's standard request bound. [Source: .specify/archive/009-tts-commands/spec.md -> FR-013]
- **FR-115**: The speak command MUST accept an optional `--timeout <duration>` flag that overrides FR-114's default bound for that invocation. The value MUST be a positive duration; a zero, negative, or unparsable value MUST be a usage error raised before contacting the hub. Other TTS commands are unaffected by this flag. [Source: .specify/archive/009-tts-commands/spec.md -> FR-013a]
- **FR-116**: The new commands, the `tts-cache` keyword, and the `clear` verb MUST appear in `sonora help`, the README command table, and the CLI command landscape document. The README MUST also list the CLI's exit codes, including the new code 13. [Source: .specify/archive/009-tts-commands/spec.md -> FR-015]
- **FR-117**: All request and response handling MUST follow the TTS operations as published in the hub's API description (`api/openapi.json`); where that description is less precise than the extension's own contract (e.g. error codes, which it leaves as free text), the CLI MUST accept anything the extension's contract allows. [Source: .specify/archive/009-tts-commands/spec.md -> FR-016]
- **FR-118**: The CLI MUST provide `sonora list tts-voices --provider <name> [--language <code>] [--engine <name>]`, which lists the voices a TTS provider offers by calling `GET /api/tts/providers/{providerName}/voices` (operationId `listVoices`), with the provider name path-escaped. `--language` and `--engine` are sent as query parameters only when given, and are passed through verbatim: the hub matches them case-insensitively and accepts aliases. Voices are shown in the hub's order. `tts-voices` is a collection keyword special-cased under `list`, like `tts-cache`; `get tts-voices`, and `list tts-voices/<anything>`, are usage errors that point to `sonora list tts-voices`. [Source: .specify/archive/tiny/2026-09-27-list-tts-voices-command/list-tts-voices-command.md -> What; Requirements 1, 3]
- **FR-119**: A missing or blank `--provider`, a blank `--language` or `--engine`, or any positional argument MUST be a usage error raised before any request is sent. [Source: .specify/archive/tiny/2026-09-27-list-tts-voices-command/list-tts-voices-command.md -> Requirements 2]
- **FR-120**: Every voice MUST show all five fields (`shortName`, `fullName`, `engine`, `language`, `gender`) in YAML and in JSON. An absent field MUST be shown as `null`, never as an empty string, and an empty result MUST print an explicit empty list. [Source: .specify/archive/tiny/2026-09-27-list-tts-voices-command/list-tts-voices-command.md -> Requirements 5]
- **FR-121**: A hub 400 or 503 MUST surface the hub's message through the TTS failure path, with the exit codes FR-113 gives; a 404 MUST get FR-110's diagnosis; nothing MUST go to standard output on any failure. The help text MUST say that for a `google-cloud` voice the value to pass to `speak --voice` is its `fullName`, because `speak` has no `--engine` flag. [Source: .specify/archive/tiny/2026-09-27-list-tts-voices-command/list-tts-voices-command.md -> Requirements 4; What]
- **FR-122**: The module MUST be addressable by its public repository import path (`github.com/Sonora-Multiroom/sonora-cli`), and every internal reference to the old module path MUST be updated, including build-time version injection in `Makefile` and `.goreleaser.yaml` (`release.sh` and `scripts/` must be checked too). The maintainer's local, untracked `build.sh` helper is updated too but is not part of the repository. [Source: .specify/archive/010-public-hub-package/spec.md -> FR-001]
- **FR-123**: The Multiroom Audio Hub API client MUST be moved from its internal location to a public package at the module root (`hub/`), preserving its file history. [Source: .specify/archive/010-public-hub-package/spec.md -> FR-002]
- **FR-124**: The public client MUST contain only hub protocol concerns: request/response types, HTTP calls, timeouts, error types, error classification and friendly error messages. [Source: .specify/archive/010-public-hub-package/spec.md -> FR-003]
- **FR-125**: The mapping from error class to CLI exit code MUST move out of the public client into a CLI-internal package, together with any error class that only the CLI uses (the usage-error class); all CLI call sites MUST use the new mapping and produce the same exit codes as before. [Source: .specify/archive/010-public-hub-package/spec.md -> FR-004]
- **FR-126**: Error classification and its friendly messages MUST remain in the public client so non-CLI consumers can present the same classification. [Source: .specify/archive/010-public-hub-package/spec.md -> FR-005]
- **FR-127**: Helpers the client itself uses to build error messages (e.g. single-line normalization) MUST remain available from the public client. [Source: .specify/archive/010-public-hub-package/spec.md -> FR-006]
- **FR-128**: Every exported identifier of the public client MUST have documentation that is self-contained — no references to this repo's constitution principles, `research.md`, `data-model.md` or other spec artifacts. [Source: .specify/archive/010-public-hub-package/spec.md -> FR-007]
- **FR-129**: The public client MUST keep a standard-library-only dependency footprint. [Source: .specify/archive/010-public-hub-package/spec.md -> FR-008]
- **FR-130**: Existing contract and unit tests for the client MUST remain in this repo and continue to run as the conformance gate against `api/openapi.json`. [Source: .specify/archive/010-public-hub-package/spec.md -> FR-009]
- **FR-131**: The CLI MUST have no observable behavior change: same commands, flags, output (YAML and JSON), error messages and exit codes. [Source: .specify/archive/010-public-hub-package/spec.md -> FR-010]
- **FR-132**: There MUST be an automated check that a build produced with a supplied version reports that version from `sonora --version`. [Source: .specify/archive/010-public-hub-package/spec.md -> FR-011]
- **FR-133**: The hub API specification (`api/openapi.json`) SHOULD be published from the module as importable content, sourced from the single existing file (no second copy). [Source: .specify/archive/010-public-hub-package/spec.md -> FR-012]
- **FR-134**: Hub URL configuration discovery and output rendering MUST stay internal to the CLI and MUST NOT be made public by this feature. [Source: .specify/archive/010-public-hub-package/spec.md -> FR-013]
- **FR-135**: After merge, a release tag containing the public client (and published spec) MUST be created; published tags MUST never be moved or re-created. [Source: .specify/archive/010-public-hub-package/spec.md -> FR-014]
- **FR-136**: Messages produced by the public client (`ClassifyError` and the `Error()` methods of its error types) MUST NOT mention CLI flags, environment variables, config files, exit codes or "the CLI". Where the CLI shows such hints today, the CLI MUST add them itself so its output stays identical (FR-131). [Source: .specify/archive/010-public-hub-package/spec.md -> FR-015]
- **FR-137**: A `409` from route creation (`CreateRoute`), transfer (`TransferRoute`) or play (`Playback`) MUST be decoded into the hub's problem details: `detail`, `title`, `reason` and the optional `outputId` (empty when the hub omits them). The CLI MUST report a decoded refusal with the hub's `detail` alone, on standard error, and exit `14` (the hub refused for its current state: change that state and retry); nothing goes to standard output. The one exception is reason `ROUTE_CHANGING_OUTPUTS` (hub 0.1.25), which clears by itself: it still exits `14`, but its message says the route is still adding an output and to retry in a moment (FR-148, FR-150), and `transfer` retries it first (FR-146). A `409` whose body cannot be decoded also exits `14`, with a generic conflict message. `reason` is an open string: an unknown value is a generic conflict, so a new hub reason needs no client release. [Source: .specify/archive/tiny/2026-10-07-decode-route-refusals-409/decode-route-refusals-409.md; .specify/archive/011-hub-0128-alignment/spec.md -> FR-012]
- **FR-138**: The public `Route` record MUST carry `joinMode` and `outputs` (API 0.1.22; `outputs` lists the outputs a route plays on right now). Both are optional on decode, so a hub that omits them changes nothing. Error classification gains a conflict class for `409` that keeps the public client's messages free of CLI wording (FR-136). [Source: .specify/archive/tiny/2026-10-07-decode-route-refusals-409/decode-route-refusals-409.md]

**Enable and disable documentation (hub 0.1.25)**

- **FR-139**: The README MUST state what each enable and disable does to playback, as the hub 0.1.25 contract describes it: disabling an input stops every route playing it (paused routes and routes sharing its hardware input too) and removes an input registered to remove itself; disabling a group stops the routes addressed to the group while routes addressed to a member keep playing; disabling an output takes it out of every route playing on it, so a route addressed to it stops and a group route keeps playing on its other members; enabling an output puts it back into the live group routes of its enabled groups under the same route ID, and a route stopped by the disable does not come back; enabling an input or a group starts nothing. [Source: .specify/archive/011-hub-0128-alignment/spec.md -> FR-001]
- **FR-140**: The help of `sonora enable` and `sonora disable`, for each of `inputs`, `outputs` and `groups`, MUST include one line on what happens to running playback. [Source: .specify/archive/011-hub-0128-alignment/spec.md -> FR-002]
- **FR-141**: The enable and disable commands' arguments, flags, output, error output and exit codes MUST NOT change; only their `--help` text gains the description line. [Source: .specify/archive/011-hub-0128-alignment/spec.md -> FR-003]

**Play shows the resolved display name (hub 0.1.28)**

- **FR-142**: The hub client's play result MUST carry the display name the hub returns, and MUST still succeed, with an empty name, when the hub omits it. [Source: .specify/archive/011-hub-0128-alignment/spec.md -> FR-004]
- **FR-143**: `sonora play` MUST show the display name in YAML and JSON output, directly after the input ID, when the hub returns a non-empty one. [Source: .specify/archive/011-hub-0128-alignment/spec.md -> FR-005]
- **FR-144**: When the hub returns no display name, or an empty one, `sonora play` output MUST NOT contain a display name field and MUST otherwise be identical to its output before this feature. [Source: .specify/archive/011-hub-0128-alignment/spec.md -> FR-006]
- **FR-145**: The help of `sonora play` and the README MUST describe `--display-name` as optional and say what the hub names an input when it is left out. [Source: .specify/archive/011-hub-0128-alignment/spec.md -> FR-007]

**Transient transfer refusal (hub 0.1.25)**

- **FR-146**: When the hub refuses a transfer with reason `ROUTE_CHANGING_OUTPUTS`, `sonora transfer` MUST retry the same transfer automatically, with a short delay between attempts, for no more than about 2 seconds after the first refusal. [Source: .specify/archive/011-hub-0128-alignment/spec.md -> FR-008]
- **FR-147**: A retry that is accepted MUST produce exactly the output and exit code of a transfer accepted on the first attempt, with nothing extra on stdout or stderr. [Source: .specify/archive/011-hub-0128-alignment/spec.md -> FR-009]
- **FR-148**: When the retries run out, the CLI MUST exit with the conflict exit code (14) and print a message saying the route is still adding an output and the user should retry shortly. [Source: .specify/archive/011-hub-0128-alignment/spec.md -> FR-010]
- **FR-149**: The CLI MUST NOT retry any other refusal reason, a refusal without a reason, a 404, any other hub error or a network failure; each MUST be reported exactly as before this feature. [Source: .specify/archive/011-hub-0128-alignment/spec.md -> FR-011]
- **FR-150**: Wherever else the CLI reports a refusal with reason `ROUTE_CHANGING_OUTPUTS`, its message MUST say the refusal is transient rather than asking the user to change the hub's state. [Source: .specify/archive/011-hub-0128-alignment/spec.md -> FR-012]
- **FR-151**: The hub client's documentation of refusal reasons MUST list `ROUTE_CHANGING_OUTPUTS`, describe it as transient, and include it among the reasons that name no output. The reason MUST stay an open string, so that unknown reasons keep working without a client release. [Source: .specify/archive/011-hub-0128-alignment/spec.md -> FR-013]
- **FR-152**: The README's exit-code row for 14 MUST mention the transient refusal and that a transfer retries it. [Source: .specify/archive/011-hub-0128-alignment/spec.md -> FR-014]

**Hub 0.1.25–0.1.28 alignment: general**

- **FR-153**: Every behaviour changed by the hub 0.1.25–0.1.28 alignment MUST be covered by tests written before the change, against a fake hub whose responses follow `api/openapi.json` at 0.1.28. [Source: .specify/archive/011-hub-0128-alignment/spec.md -> FR-015]
- **FR-154**: The hub 0.1.25–0.1.28 alignment MUST NOT add a third-party dependency, command or flag, nor anything on the CLI's startup path. Join modes (CLI-012) MUST NOT add a third-party dependency or anything on the CLI's startup path, and `--include-routes` MUST send a second request only when the flag is given. [Source: .specify/archive/011-hub-0128-alignment/spec.md -> FR-016] [Source: .specify/archive/012-join-modes/spec.md -> FR-018]

- **FR-155**: The CLI MUST provide `sonora rename inputs/<input-id> <name>` and `sonora rename inputs/<input-id> --reset`, each making one request, `PUT /api/v2/inputs/{inputId}/display-name` (operationId `setInputDisplayName`), with `{"displayName": "<name>"}` or `{"displayName": null}` for `--reset`. A multi-word name is one quoted argument. On success (200) the command prints the returned input record as `get inputs/<input-id>` does (YAML, or JSON with `--json`) and exits `0`. [Source: .specify/archive/tiny/2026-10-22-rename-inputs-command/rename-inputs-command.md]
- **FR-156**: A 404 from a rename request MUST be reported as "input not found: <id>" on standard error with exit `5`. A 400 MUST be reported with the hub's detail (or the generic invalid-request message when the hub gives none) with exit `6`. Any other non-2xx, network or malformed-body failure keeps the generic classification (FR-009). Nothing is written to standard output on any failure. [Source: .specify/archive/tiny/2026-10-22-rename-inputs-command/rename-inputs-command.md]
- **FR-157**: `rename` MUST fail with a usage error (exit `2`), without contacting the hub, when a name and `--reset` are both given, when neither is given, when the `<input-id>` is missing, when extra positional arguments are given, or when the resource is not `inputs` (for example `outputs/<id>`). [Source: .specify/archive/tiny/2026-10-22-rename-inputs-command/rename-inputs-command.md]
- **FR-158**: `rename --help` MUST print the usage and state that a rename of a configured input lasts until a restart, and a rename of a runtime input lasts until the input is removed. [Source: .specify/archive/tiny/2026-10-22-rename-inputs-command/rename-inputs-command.md]
- **FR-159**: The public `hub` package MUST provide `SetInputDisplayName` for the call in FR-155, returning the decoded, validated input. It MUST map a 404 to `NotFoundError` for `input`, a 400 to `APIError` carrying the hub's detail (falling back to `StatusError`), a malformed body to `DecodeError` and other non-2xx responses to `StatusError`, so that other Sonora consumers such as sonora-mcp can offer the same rename. Its godoc MUST state that a null or blank name restores the original name. [Source: .specify/archive/tiny/2026-10-22-rename-inputs-command/rename-inputs-command.md]
- **FR-160**: `DisplayNameRequest` MUST encode a nil name as `{"displayName": null}` (no omit-empty), so that a reset is sent explicitly. Name length and characters are not validated by the CLI; the hub decides and its 400 detail is printed. [Source: .specify/archive/tiny/2026-10-22-rename-inputs-command/rename-inputs-command.md]

**Choosing the join mode (CLI-012, API 0.1.21)**

- **FR-161**: The hub client's route request, play request and create-input request MUST each have an optional join-mode field (a default join mode for the create-input request). The field is left out of the request body when it is unset, and sent as given when it is set. [Source: .specify/archive/012-join-modes/spec.md -> FR-001]

- **FR-162**: `sonora play` and `sonora route` MUST accept `--join-mode <mode>`, and `sonora create inputs/<id>` MUST accept `--default-join-mode <mode>`. The mode is `REPLACE`, `MIX` or `DUCK_OTHERS`, matched without regard to case and sent upper-cased. [Source: .specify/archive/012-join-modes/spec.md -> FR-002]

- **FR-163**: An unknown or empty mode MUST be a usage error (exit 2). The command MUST send no request, and MUST print an error that lists the valid modes. [Source: .specify/archive/012-join-modes/spec.md -> FR-003]

- **FR-164**: Without the flag, each command MUST send the same request it sends today, with no join-mode field, so that the hub's defaults apply. [Source: .specify/archive/012-join-modes/spec.md -> FR-004]

- **FR-165**: The `--help` of each of the three commands MUST list the flag and its modes, and say what happens without it: a route uses the input's default join mode, falling back to `REPLACE`; a play uses `REPLACE`; an input without a default leaves its routes on `REPLACE`. [Source: .specify/archive/012-join-modes/spec.md -> FR-005]


**Showing an output's routes (CLI-012, API 0.1.21)**

- **FR-166**: The hub client MUST offer an operation that returns the routes playing on one output, in the hub's order, with every field of the contract's route-on-output record. A 404 from it MUST be the client's usual not-found error, naming the output. [Source: .specify/archive/012-join-modes/spec.md -> FR-006]

- **FR-167**: The hub client MUST let a consumer tell whether a route on an output is lowered. Any duck state other than `LOWERED` counts as not lowered. [Source: .specify/archive/012-join-modes/spec.md -> FR-007]

- **FR-168**: `sonora get outputs/<id>` MUST accept `--include-routes`, with the help text "also list the routes playing on this output, in join order". With it, the output MUST be the output's record followed by a `routes` list in the hub's order. In YAML each entry MUST hold `routeId`, `inputId`, `targetId`, `targetType`, `status`, `joinMode`, `duckState`, `main` and `paused`, in that order; in JSON each entry MUST hold every field of the contract's route-on-output record. An output with no routes MUST show an empty list. [Source: .specify/archive/012-join-modes/spec.md -> FR-008]

- **FR-169**: `--include-routes` on any other `get` form (a list such as `get outputs`, or a resource other than `outputs/<id>`) MUST be a usage error (exit 2). The command MUST send no request, and the error MUST say the flag applies only to `get outputs/<id>`. [Source: .specify/archive/012-join-modes/spec.md -> FR-008a]

- **FR-170**: Without `--include-routes`, `sonora get outputs/<id>` MUST send exactly one request, and its output MUST be unchanged. [Source: .specify/archive/012-join-modes/spec.md -> FR-009]

- **FR-171**: When the output request succeeds and the routes request answers 404 or 405 (a hub with the stop-routes operation on the same path but not the list answers 405), the CLI MUST report that the hub does not support listing an output's routes and needs 0.1.21 or later. It MUST exit with the unexpected-response exit code (3), and MUST NOT report the output as not found. [Source: .specify/archive/012-join-modes/spec.md -> FR-010]

- **FR-172**: With `--include-routes`, any failure MUST leave stdout empty. The CLI never prints the output's record without its routes. [Source: .specify/archive/012-join-modes/spec.md -> FR-011]


**Showing join modes on routes and inputs (CLI-012, API 0.1.21)**

- **FR-173**: The YAML of `sonora get routes/<id>` MUST show the route's join mode and the list of outputs it plays on, after its current fields. When the hub returns neither field (a hub older than 0.1.21), both MUST be left out. [Source: .specify/archive/012-join-modes/spec.md -> FR-012]

- **FR-174**: The YAML of every command that prints one input record (`get`, `create`, `rename`, `enable` and `disable inputs/<id>`) MUST show the input's default join mode, as null when the input has none. Their JSON already shows it. [Source: .specify/archive/012-join-modes/spec.md -> FR-013]

- **FR-175**: The JSON of `get routes/<id>` and `get inputs/<id>` MUST keep its current shape and contain the same join-mode fields as the YAML. [Source: .specify/archive/012-join-modes/spec.md -> FR-014]

- **FR-176**: `sonora list routes` and `sonora list inputs` output MUST NOT change. [Source: .specify/archive/012-join-modes/spec.md -> FR-015]


**General (CLI-012)**

- **FR-177**: The README MUST describe the three join modes in a short paragraph and document `--join-mode`, `--default-join-mode` and `--include-routes`. [Source: .specify/archive/012-join-modes/spec.md -> FR-016]

- **FR-178**: Every new behaviour MUST be covered by tests written before the code, against a fake hub whose answers follow `api/openapi.json` at 0.1.28. That includes contract tests for the new hub client operation: a list, an empty list, a 404, and an unknown duck state. It also includes contract tests showing that each new request field is sent only when it is set. [Source: .specify/archive/012-join-modes/spec.md -> FR-017]

- **FR-179**: The public `hub` package MUST send `outputVolume` on a play request, never `volume`, and MUST send `routeVolume` and `duckingLevelDb` on play and route requests, and `defaultRouteVolume` and `defaultDuckingLevelDb` on input creation, only when each is set; an unset field MUST be omitted, not sent as null. [Source: .specify/archive/tiny/2026-10-25-hub-route-volume-client/hub-route-volume-client.md]

- **FR-180**: The public `hub` package MUST decode `routeVolume` and `duckingLevelDb` on routes and on routes-per-output records, and `defaultRouteVolume` and `defaultDuckingLevelDb` on inputs, as optional values: null or absent (a pre-0.1.29 hub) MUST decode as unset, and a set value MUST decode as sent, including negative and fractional levels. [Source: .specify/archive/tiny/2026-10-25-hub-route-volume-client/hub-route-volume-client.md]

- **FR-181**: The public `hub` package MUST provide `SetRouteVolume`, which sends `PUT /api/v2/routes/{routeId}/volume` with `{"volume": N}`. A 200 MUST return the decoded, validated route; a 404 MUST be a not-found error for `route`; a 400 MUST carry the hub's detail; a malformed body MUST be a decode error; any other non-success status MUST be a status error. A route's volume MUST NOT be clearable through this call. [Source: .specify/archive/tiny/2026-10-25-hub-route-volume-client/hub-route-volume-client.md]

- **FR-182**: Renaming `PlaybackRequest.Volume` to `OutputVolume` is a breaking change to the public `hub` package. While the module is `v0.x` it is allowed, but the release notes MUST call it out, and every new exported identifier MUST carry godoc that stands without this repository's specs. [Source: .specify/archive/tiny/2026-10-25-hub-route-volume-client/hub-route-volume-client.md]

- **FR-183**: `sonora play` MUST accept `--output-volume N` (0-100), which sends `outputVolume`, and `--route-volume N` (0-100), which sends `routeVolume`. `--volume` MUST NOT exist: it is an unknown flag, with usage exit `2`, and it MUST have no alias or notice. An out-of-range value for either flag MUST be rejected before any request, with an error that names the flag and the validation exit code. [Source: .specify/archive/tiny/2026-10-25-route-volume-cli/route-volume-cli.md]

- **FR-184**: `sonora play` MUST accept `--ducking-level DB` (-60 to 0, decimals allowed), which sends `duckingLevelDb`. Without `--join-mode DUCK_OTHERS` it MUST be a usage error (exit `2`); an out-of-range value MUST be a validation error. Both checks MUST run before any request. [Source: .specify/archive/tiny/2026-10-25-route-volume-cli/route-volume-cli.md]

- **FR-185**: `sonora set routes/<route-id> volume <0-100>` MUST call `SetRouteVolume` and print the returned route with the route renderers, exit `0`. Its missing-identifier, attribute, value and extra-argument errors MUST match `set outputs/<id> volume` in message and exit code, and an unknown route MUST exit with the not-found code and print nothing to standard output. [Source: .specify/archive/tiny/2026-10-25-route-volume-cli/route-volume-cli.md]

- **FR-186**: `get routes/<id>` and the routes shown by `get outputs/<id> --include-routes` MUST print `routeVolume` and `duckingLevelDb` after `joinMode`, and `get inputs/<id>` MUST print `defaultRouteVolume` and `defaultDuckingLevelDb` after `defaultJoinMode`, each only when set. A level MUST print in its shortest form (`-25`, `-12.5`). `--json` MUST carry the fields as the `hub` types encode them, null when unset. The list views MUST NOT change. [Source: .specify/archive/tiny/2026-10-25-route-volume-cli/route-volume-cli.md]

- **FR-187**: Removing `--volume` from `sonora play` is a breaking CLI change, and the release notes MUST say so. [Source: .specify/archive/tiny/2026-10-25-route-volume-cli/route-volume-cli.md]

- **FR-188**: The public `hub` package MUST provide `DiagnoseTTSUnavailable`, which makes one extension-inventory request and returns the reason TTS is unavailable on the hub, using the mapping in FR-110's table, so that a caller outside the CLI can explain a TTS 404 the same way the CLI does. The CLI MUST use it, and its output and exit codes MUST NOT change. [Source: .specify/archive/tiny/2026-10-28-hub-tts-unavailable-diagnosis/hub-tts-unavailable-diagnosis.md]

- **FR-189**: `hub.ClassifyError` MUST classify a TTS 404 (`TTSNotOfferedError`), wrapped or not, as text-to-speech unavailable, with the message "text-to-speech is not available on this hub", and not as a network failure. The v0.3.3 release notes MUST say that the class changed for callers that matched the old one. [Source: .specify/archive/tiny/2026-10-28-hub-tts-unavailable-diagnosis/hub-tts-unavailable-diagnosis.md]

- **FR-190**: The public `hub` package's `TTSCacheStats` MUST marshal to JSON with the property names of `CacheStats` in `api/openapi.json`: `totalEntries`, `totalSizeBytes`, `maxSizeBytes` and `entriesByProvider`. Decoding through `GetTTSCacheStats` and the CLI's output MUST NOT change. [Source: .specify/archive/tiny/2026-10-28-tts-cache-stats-json-tags/tts-cache-stats-json-tags.md]

- **FR-191**: `sonora speak` MUST accept `--playback-mode MODE`, which sends `playbackMode` verbatim with no client-side check of the value; `--route-volume N` (0-100), which sends `routeVolume`; and `--ducking-level DB` (-60 to 0, decimals allowed), which sends `duckingLevelDb`. A blank `--playback-mode` MUST be a usage error (exit `2`). An out-of-range `--route-volume` or `--ducking-level` MUST be a validation error that names the flag. Each check MUST run before any request. `--ducking-level` MUST be accepted without `--playback-mode`, and with `mix`, where the hub decides. [Source: .specify/archive/tiny/2026-10-27-speak-request-fields/speak-request-fields.md]

- **FR-192**: The public `hub` package's `SpeakRequest` MUST carry the seven hub 0.1.29 fields it lacked (`engine`, `pitch`, `speakingRate`, `stylePrompt`, `playbackMode`, `routeVolume`, `duckingLevelDb`) as optional values. Each MUST be omitted from the request when unset, and sent, including zero and empty values, when set. Its godoc MUST say so. [Source: .specify/archive/tiny/2026-10-27-speak-request-fields/speak-request-fields.md]

- **FR-193**: `sonora speak` MUST NOT gain flags for `engine`, `pitch`, `speakingRate` or `stylePrompt` (`stylePrompt` is superseded by FR-204 from 2026-11-12). A `speak` invocation with none of the new flags MUST send the same request body as before this change. [Source: .specify/archive/tiny/2026-10-27-speak-request-fields/speak-request-fields.md]
- **FR-194**: The public `hub` package MUST offer a function that calls `GET /api/v2/info` and returns a typed record of `product`, `version`, `apiVersion`, `startedAt` and `uptimeSeconds`. [Source: .specify/archive/013-info-and-tts-warm/spec.md -> FR-001]
- **FR-195**: The CLI MUST offer a command that shows hub info as YAML by default and as one JSON object with `--json`. [Source: .specify/archive/013-info-and-tts-warm/spec.md -> FR-002]
- **FR-196**: A 404 from the info call MUST be reported as "this hub does not offer hub info (needs hub 0.1.33 or later)". [Source: .specify/archive/013-info-and-tts-warm/spec.md -> FR-003]
- **FR-197**: The public `hub` package MUST offer a function that calls `POST /api/tts/cache/warm` with `text` and the optional `providerName`, `voice`, `language`, `engine`, `pitch`, `speakingRate` and `stylePrompt`, sending only the fields that are set, and returns `cacheHit`. [Source: .specify/archive/013-info-and-tts-warm/spec.md -> FR-004]
- **FR-198**: The CLI MUST offer a command that warms one phrase, taking the text as an argument or `-` for standard input, with `--provider`, `--voice`, `--language` and `--timeout` meaning what they mean for `speak`. [Source: .specify/archive/013-info-and-tts-warm/spec.md -> FR-005]
- **FR-199**: The warm command MUST print `cacheHit` as YAML by default and as one JSON object with `--json`, and MUST NOT play anything. [Source: .specify/archive/013-info-and-tts-warm/spec.md -> FR-006]
- **FR-200**: The warm command MUST map hub answers as `speak` does: 400 is a usage error with the hub's message, a missing TTS extension gives the "text-to-speech is not available" message, and 503 is a failure whose message says that repeating is safe. [Source: .specify/archive/013-info-and-tts-warm/spec.md -> FR-007]
- **FR-201**: Both new commands MUST follow the existing conventions: strict GNU flag style, exit codes, errors on standard error, output on standard output, and an entry in `help`. [Source: .specify/archive/013-info-and-tts-warm/spec.md -> FR-008]
- **FR-202**: Every request and response shape of the two new operations MUST be validated against `api/openapi.json`. [Source: .specify/archive/013-info-and-tts-warm/spec.md -> FR-009]
- **FR-203**: The README MUST document both commands and state that they need hub 0.1.33 or later. [Source: .specify/archive/013-info-and-tts-warm/spec.md -> FR-010]
- **FR-204**: `sonora speak` and `sonora warm tts-cache` MUST accept `--style-prompt TEXT`, which sends `stylePrompt` verbatim, so a `google-gemini` announcement's delivery, or a warmed phrase's, can be set from the shell. A given but empty or blank value MUST be sent as given (the hub turns the entry's default prompt off) and MUST NOT be a usage error. Without the flag the request MUST carry no `stylePrompt` key. The client MUST NOT check the value's length or provider; the hub's 400 decides. [Source: .specify/archive/tiny/2026-11-12-speak-style-prompt-flag/speak-style-prompt-flag.md]
- **FR-205**: The help text of both commands MUST name `google-gemini` as the only provider that takes the prompt and say that an empty value turns the configured prompt off. [Source: .specify/archive/tiny/2026-11-12-speak-style-prompt-flag/speak-style-prompt-flag.md]
- **FR-206**: `sonora play` and `sonora route` MUST accept `--chime <id|none>` and send `chime` with the value verbatim, `none` matched exactly in lower case, and MUST send no `chime` field when the flag is absent. A blank value MUST be a usage error (exit `2`) with no request sent. [Source: .specify/archive/tiny/2026-11-20-route-chime/route-chime.md -> Requirements 2, 4]
- **FR-207**: The public `hub` package MUST send `chime` on play, route and input-creation requests only when it is set, and MUST decode `chime` on `Output`, `Group` and `Input`, where an absent value is nil and `none` is a value, never nil. [Source: .specify/archive/tiny/2026-11-20-route-chime/route-chime.md -> Requirement 2]
- **FR-208**: The public `hub` package MUST offer `ListChimes`, which calls `GET /api/v2/chimes` and returns each chime's `id` and `durationMillis`. An empty list MUST NOT be an error, and a malformed body MUST be a decode error. Its godoc MUST say that `none` is reserved and never listed, without referring to this repository. [Source: .specify/archive/tiny/2026-11-20-route-chime/route-chime.md -> Requirement 3]
- **FR-209**: The CLI MUST NOT check a chime id against the hub's list before sending it. The hub's 400 is reported through the existing error path, so the flag costs no extra request. [Source: .specify/archive/tiny/2026-11-20-route-chime/route-chime.md -> Requirement 5]
- **FR-210**: `sonora list chimes` and `sonora get chimes` MUST print the chimes under `chimes:` as YAML by default or as one JSON object with `--json`. With no chimes they MUST print `# no chimes found` and `chimes: []`. `list chimes/<id>` MUST be a usage error. [Source: .specify/archive/tiny/2026-11-20-route-chime/route-chime.md -> Requirement 6]
- **FR-211**: `sonora get chimes/<id>` MUST print that chime as one record. It MUST check the id against the resource id rule first (a malformed id is a usage error, exit `2`), and an unknown id MUST fail with `chime not found: <id>` (exit `5`). `chimes` MUST NOT be a play or route target and MUST have no alias. [Source: .specify/archive/tiny/2026-11-20-route-chime/route-chime.md -> Requirement 7]
- **FR-212**: `get outputs[/id]` and `get groups[/id]` MUST show `chime:`, and `get inputs[/id]` MUST show `defaultChime:`, only when set, in YAML and JSON. Output for a target with no chime MUST be unchanged. [Source: .specify/archive/tiny/2026-11-20-route-chime/route-chime.md -> Requirement 8]
- **FR-213**: `sonora get` and `sonora list` with no resource name MUST name every resource that verb accepts, as FR-066 requires. `get` adds `chimes`, `info`, `master-mute` and `tts-cache`; `list` adds `chimes` and `tts-voices`. [Source: .specify/archive/tiny/2026-11-20-route-chime/route-chime.md -> Requirement 9]

- **FR-214**: The public `hub` package MUST offer functions to search stations, fetch one station by UUID, and list countries, languages and tags, each returning typed records that marshal to the hub's property names. [Source: .specify/archive/014-radio-station-search/spec.md -> FR-001]
- **FR-215**: The station search function MUST accept every filter the hub defines (name, tag, country code, language, codec, minimum bitrate), the sort field and direction, page size and offset, and MUST send only those that are set. [Source: .specify/archive/014-radio-station-search/spec.md -> FR-002]
- **FR-216**: Radio error answers MUST be decoded into the package's existing error type, keeping the problem kind, message, offending parameter and retry delay, and MUST be classified by `hub.ClassifyError`. [Source: .specify/archive/014-radio-station-search/spec.md -> FR-003]
- **FR-217**: A 404 from a radio operation that is not a "station not found" problem MUST become a typed "radio not offered" error. The `hub` package MUST offer a diagnosis that makes one extension-inventory request, matches the extension by its id `radio`, and returns a "radio is not available" error naming the reason listed in the radio edge cases; the CLI MUST run it only after such a 404 and print its message. [Source: .specify/archive/014-radio-station-search/spec.md -> FR-004]
- **FR-218**: The CLI MUST offer `sonora search radio-stations [name]` with options for tag, country, language, codec, minimum bitrate, sort field, sort direction, page size and offset. [Source: .specify/archive/014-radio-station-search/spec.md -> FR-005]
- **FR-219**: Search and station output MUST show every station field the hub returns, including the UUID and the playable address that `sonora play` accepts, in YAML by default and in JSON with `--json`; both formats carry the same fields, and an absent field is shown as null. A single page also carries the next-page offset (null on the last page). [Source: .specify/archive/014-radio-station-search/spec.md -> FR-006]
- **FR-220**: Every search and list command MUST make exactly one radio request per run; the CLI MUST NOT follow next-page offsets by itself. [Source: .specify/archive/014-radio-station-search/spec.md -> FR-007]
- **FR-221**: The CLI MUST offer `sonora get radio-stations/<uuid>`, which shows one station. [Source: .specify/archive/014-radio-station-search/spec.md -> FR-008]
- **FR-222**: The CLI MUST offer `sonora get radio-countries`, `sonora get radio-languages` and `sonora get radio-tags`; the tag command MUST accept a name filter, page size and offset. [Source: .specify/archive/014-radio-station-search/spec.md -> FR-009]
- **FR-223**: Every radio command that pages MUST validate page size (1–100) and offset (0–10000) before any request, a failure being a usage error (exit 2). Hub answers MUST map to exit codes: 400 is a validation error (exit 6) with the hub's message; 404 is not found (exit 5, station) or "radio is not available" with its reason (exit 15); 502 is an upstream failure (exit 9); 503 is temporarily unavailable (exit 10) with its retry delay. [Source: .specify/archive/014-radio-station-search/spec.md -> FR-010]
- **FR-224**: Every text value from the catalogue MUST be printed with control and escape characters neutralised. [Source: .specify/archive/014-radio-station-search/spec.md -> FR-011]
- **FR-225**: The radio commands MUST be always present; no request is made at start-up or before the radio request to find out whether the hub has the extension. The inventory check happens only after a failed radio request. [Source: .specify/archive/014-radio-station-search/spec.md -> FR-013]
- **FR-226**: Every request and response shape MUST be validated against `api/openapi.json`, refreshed if it lacks any radio shape this feature uses. [Source: .specify/archive/014-radio-station-search/spec.md -> FR-014]
- **FR-227**: The README MUST document the radio commands and state that they need the hub's radio extension 0.2.0 or later. [Source: .specify/archive/014-radio-station-search/spec.md -> FR-015]

- **FR-228**: `sonora create inputs/<input-id> <uri> --persistent` MUST send `persistent: true` on `POST /api/v2/inputs`. Without the flag the request body MUST be unchanged. `--persistent` together with `--auto-remove` is a usage error (exit `2`), reported before any request is sent, since a kept input is never removed automatically. The created input prints as FR-090 describes, with origin `STORED`. [Source: .specify/archive/tiny/2026-11-22-stored-inputs/stored-inputs.md -> Requirement 1]
- **FR-229**: `sonora keep inputs/<input-id>` MUST call `PUT /api/v2/inputs/{inputId}/persistent` and print the input record (YAML, or JSON with `--json`), exiting `0`. Keeping an input the hub already keeps MUST succeed and print it unchanged. A 404 is reported as not found (exit `5`); a 400 with the hub's detail (exit `6`); a 409 as FR-230 says (exit `14`); a 500, a store failure, with the store's `detail` (hub error, exit `3`). A path that is not `inputs/<id>`, extra arguments or a missing identifier are usage errors (exit `2`), with no request sent. [Source: .specify/archive/tiny/2026-11-22-stored-inputs/stored-inputs.md -> Requirement 2]
- **FR-230**: A `409` from `create inputs` or `keep inputs` with reason `INPUT_UNPLAYABLE`, `INPUT_ID_TAKEN` or `INPUT_STORE_FULL` MUST print a short CLI message in place of the hub's detail and keep the conflict exit code (`14`, FR-137): `INPUT_UNPLAYABLE` says the input's link is not playable now, perhaps because its extension is not running, and to retry once it is; `INPUT_ID_TAKEN` says an input with that identifier already exists; `INPUT_STORE_FULL` says the store of kept inputs is full and a kept input must be deleted first. A 409 with any other reason, or none, stays a generic conflict that shows the hub's detail, or a generic conflict message when the hub gives none. [Source: .specify/archive/tiny/2026-11-22-stored-inputs/stored-inputs.md -> Requirement 3]
- **FR-231**: The YAML of an input record, for `get`, `list`, `create`, `rename`, `enable`, `disable` and `keep` alike, MUST show `origin`, `playable`, and `unplayableReason` when the hub gives one. An empty `origin`, from a hub older than 0.1.35, MUST be omitted. The JSON output MUST be unchanged. [Source: .specify/archive/tiny/2026-11-22-stored-inputs/stored-inputs.md -> Requirement 4]
- **FR-232**: The help text MUST list `keep inputs/<input-id>` and the `--persistent` flag of `create`, and the README MUST document both. No other command's behaviour changes. [Source: .specify/archive/tiny/2026-11-22-stored-inputs/stored-inputs.md -> Requirement 5]

### Key Entities

- **Output**: An audio output device/zone managed by the hub. Attributes relevant to this
  feature: unique identifier, human-readable display name, current volume level (a
  bounded numeric range), current mute state (on/off), whether its hardware is currently
  connected (availability), and whether it is enabled for use.
  [Source: .specify/archive/001-list-outputs/spec.md -> Output; .specify/archive/002-outputs-get/spec.md -> Output]
  The get command reuses this entity unchanged; its response is one decoded object, not an array.
  Field mapping to `OutputResponse` in `api/openapi.json` (required fields; wrong JSON type or empty `outputId`/`displayName` is malformed per FR-013; volume is passed through without re-validating bounds):

  | Field | Type | openapi.json field |
  |-------|------|--------------------|
  | `OutputID` | string | `outputId` |
  | `DisplayName` | string | `displayName` |
  | `Volume` | int (0–100 domain range) | `volume` |
  | `Muted` | bool | `muted` |
  | `Available` | bool | `available` |
  | `Enabled` | bool | `enabled` |

  [Source: .specify/archive/001-list-outputs/data-model.md -> Output; .specify/archive/002-outputs-get/data-model.md -> Output]

- **Input**: An audio source registered with the hub, either statically configured or created ephemerally via the API. Attributes relevant to this feature: unique identifier, human-readable display name, source URI, whether it is enabled, source type (STATIC or EPHEMERAL), whether it auto-removes when its route stops, whether it supports pause/resume, its creation timestamp (absent for static inputs), its default join mode (null when the input has none; CLI-012), and its default route volume and default ducking level (null when unset; API 0.1.29), its origin (CONFIGURATION, STORED or RUNTIME; empty from a hub older than 0.1.35), whether it can play (`playable`) and, when it cannot, its `unplayableReason` (HUB-041, FR-231). The inputs get command returns one decoded object, regardless of enabled state. Field mapping to `InputResponse` in `api/openapi.json` (required fields; wrong JSON type, empty `inputId` or `displayName`, or a `source` other than STATIC or EPHEMERAL is malformed per FR-013; `uri` is passed through unvalidated):

  | Field | Type | openapi.json field |
  |-------|------|--------------------|
  | `InputID` | string | `inputId` |
  | `DisplayName` | string | `displayName` |
  | `URI` | string | `uri` |
  | `Enabled` | bool | `enabled` |
  | `AutoRemove` | bool | `autoRemove` |
  | `Source` | string (STATIC \| EPHEMERAL) | `source` |
  | `CreatedAt` | string, nullable (RFC 3339; `nil` for static inputs) | `createdAt` |
  | `Pauseable` | bool | `pauseable` |

  [Source: .specify/archive/003-inputs-list-get/spec.md -> "Input"; .specify/archive/003-inputs-list-get/data-model.md -> "Input"]

- **Route**: A live connection routing audio from one input to one output or output group, managed by the hub. Attributes relevant to this feature: unique identifier, source input identifier, target identifier, target type (single output or output group), current lifecycle status (starting, active, stopping, stopped, or failed), creation timestamp, playback-started timestamp (absent until playback begins), whether it can be transferred to another target, whether it supports pause/resume, and whether it is currently paused, its join mode, the outputs it plays on right now (API 0.1.21; CLI-012), and its route volume and ducking level, each null when unset (API 0.1.29). The routes list command shows the first five of these; the routes get command returns one decoded object with all fourteen. Field mapping to `RouteResponse` in `api/openapi.json` (required fields; wrong JSON type, empty `routeId`, `inputId` or `targetId`, a `targetType` other than SINGLE_OUTPUT or OUTPUT_GROUP, or a `status` outside the five values is malformed per FR-013; timestamps are passed through unparsed):

  | Field | Type | openapi.json field |
  |-------|------|--------------------|
  | `RouteID` | string | `routeId` |
  | `InputID` | string | `inputId` |
  | `TargetID` | string | `targetId` |
  | `TargetType` | string (SINGLE_OUTPUT \| OUTPUT_GROUP) | `targetType` |
  | `Status` | string (STARTING \| ACTIVE \| STOPPING \| STOPPED \| FAILED) | `status` |
  | `CreatedAt` | string (RFC 3339, required, non-nullable) | `createdAt` |
  | `StartedAt` | string, nullable (RFC 3339; `nil` before playback starts) | `startedAt` |
  | `Transferable` | bool | `transferable` |
  | `Pauseable` | bool | `pauseable` |
  | `Paused` | bool | `paused` |

  [Source: .specify/archive/004-routes-list-get/spec.md -> "Route"; .specify/archive/004-routes-list-get/data-model.md -> "Route"]

- **Output Group**: A named collection of audio outputs managed by the hub, which can be used as a single routing target. Attributes relevant to this feature: unique identifier, human-readable display name, member output identifiers (a group may have none; see Edge Cases), whether all member outputs are currently muted, and whether the group is enabled for new route creation. The groups list command shows enabled groups by default; the groups get command returns one decoded object regardless of enabled state. Both commands render the same five fields. Field mapping to `GroupResponse` in `api/openapi.json` (required fields; wrong JSON type, or an empty `groupId` or `displayName`, is malformed per FR-013; an empty `outputIds` array is valid and displayed as `[]`):

  | Field | Type | openapi.json field |
  |-------|------|--------------------|
  | `GroupID` | string | `groupId` |
  | `DisplayName` | string | `displayName` |
  | `OutputIDs` | string list (may be empty, never `nil` on decode) | `outputIds` |
  | `Muted` | bool | `muted` |
  | `Enabled` | bool | `enabled` |

  [Source: .specify/archive/005-groups-list-get/spec.md -> "Output Group"; .specify/archive/005-groups-list-get/data-model.md -> "Group"]

- **Playback Request**: A one-shot instruction to play a given audio source URI to a given target (a single output or an output group), with an optional starting volume and an optional display name for the resulting input. Field mapping to `PlaybackRequest` in `api/openapi.json` (required fields; the target type is set by the CLI from the target path's prefix (`outputs/` or `groups/`), never supplied by the user directly; `outputVolume` is range-checked client-side to 0–100 before any request is sent; `uri` scheme and reachability are validated by the hub, not the CLI):

  | Field | Type | openapi.json field |
  |-------|------|--------------------|
  | `URI` | string | `uri` |
  | `TargetID` | string | `targetId` |
  | `TargetType` | string (SINGLE_OUTPUT \| OUTPUT_GROUP) | `targetType` |
  | `DisplayName` | string, optional | `displayName` |
  | `OutputVolume` | int, optional (0–100) | `outputVolume` |
  | `RouteVolume` | int, optional | `routeVolume` |
  | `DuckingLevelDb` | float, optional | `duckingLevelDb` |

  [Source: .specify/archive/006-play-command/spec.md -> "Playback Request"; .specify/archive/006-play-command/data-model.md -> "PlaybackRequest (outbound, `internal/hub.PlaybackRequest`)"]

- **Playback Result**: The outcome of a playback request — the identifier of the ephemeral input that was created, the identifier and status of the route created to carry audio from that input to the target, and a confirmation message; since hub 0.1.28 also the input's display name once the route exists (the name sent with the request, the title a SoundCloud or YouTube link resolved to, or a generated placeholder). The command renders four of these — `inputId`, `routeId`, `status` and `message` — from the `PlaybackResponse` in `api/openapi.json`, plus `displayName` as the second field, after `inputId`, only when the hub returns a non-empty one; with no display name the record is byte-identical to the four-field record. The response's nested route is decoded in full for contract fidelity but its other fields are not rendered, because managing the route is out of scope. A response missing any of those four required values is malformed (FR-013); a missing or empty `displayName` is not an error and decodes to `""` (a hub before 0.1.28 omits it).
  [Source: .specify/archive/006-play-command/spec.md -> "Playback Result"; .specify/archive/006-play-command/data-model.md -> "PlaybackResult (rendered view)"; .specify/archive/011-hub-0128-alignment/spec.md -> "Play result"; .specify/archive/011-hub-0128-alignment/data-model.md -> "Rendered play record (`sonora play`)"]

- **Resource Kind**: A closed enumeration of the four resources `get`/`list` cover; a CLI-surface concept, not a hub one. Each has a canonical name and a short alias: `inputs`/`in`, `outputs`/`out`, `groups`/`gr`, `routes`/`rt`. An unrecognized name or alias in either position (the `get`/`list` argument, or the prefix of `play`'s target) is a usage error, exit `2`.
  [Source: .specify/archive/007-refactor-cli-commands/data-model.md -> "ResourceKind"]

- **Resource Path**: The parsed form of a command-line argument such as `outputs/office-speaker` or `out/office-speaker`: a Resource Kind (the segment before the first `/`, or the whole argument when there is no `/`) and an optional identifier (the segment after the first `/`) matching `^[a-zA-Z0-9_-]{1,255}$`. A second `/` in the identifier segment is malformed, not a deeper path. `get` accepts a path with or without an id; `list` accepts a bare resource only; `play`'s target must resolve to `outputs` or `groups` and requires an id.
  [Source: .specify/archive/007-refactor-cli-commands/data-model.md -> "ResourcePath"]
- **Routing Request**: A one-shot instruction to connect an existing input (`inputs/<input-id>`) to a target output or output group (`outputs/<target-id>` or `groups/<target-id>`). The command checks that the input and the target exist, then sends the request as one `POST /api/v2/routes` call. [Source: .specify/archive/008-route-command/spec.md -> "Routing Request"; .specify/archive/008-route-command/data-model.md -> "CreateRouteRequest"]
- **Routing Result**: The outcome of a routing request: the identifier and status of the route the hub created, and the confirmation message the CLI builds around them. [Source: .specify/archive/008-route-command/data-model.md -> "RoutingResult"]

- **Master Mute State**: Whether the whole system is muted: a single boolean, `muted`. There is one of it, with no identifier and no collection. [Source: .specify/archive/tiny/2026-08-29-master-mute-command/master-mute-command.md]

- **Bulk Stop Result**: The outcome of a bulk stop: the number of routes stopped and, for each, its route identifier, target type, target identifier and an optional stop reason. The list is empty when nothing was playing. [Source: .specify/archive/tiny/2026-09-09-stop-routes-command/stop-routes-command.md]

- **Announcement Request**: A one-shot request to speak a text on a target (one output or one output group, by id), with optional provider, voice, and language overrides. Sent to the extension's `speak` operation as text, a target id and target type, and the optional fields only when supplied. [Source: .specify/archive/009-tts-commands/spec.md -> "Announcement request"]
- **Announcement Acceptance**: The hub's acceptance of an announcement: the announcement identifier, a flag saying whether the audio was served from cache, and the queue depth for the target, including this announcement. [Source: .specify/archive/009-tts-commands/spec.md -> "Announcement acceptance"]
- **TTS Cache Statistics**: The TTS cache's total entries, total size in bytes, maximum size in bytes, and a provider-name to entry-count breakdown (which may be empty). Its JSON encoding uses the hub's property names `totalEntries`, `totalSizeBytes`, `maxSizeBytes` and `entriesByProvider` (FR-190). [Source: .specify/archive/009-tts-commands/spec.md -> "TTS cache statistics"]
- **TTS Error**: A machine-readable code plus a human-readable message, returned by the extension alongside a 400 (caller-fixable) or 503 (provider-side) status. The CLI keeps its own error shape separate from the core API's problem details. [Source: .specify/archive/009-tts-commands/spec.md -> "TTS error"]

- **TTS Voice**: One voice a provider offers: `shortName`, `fullName`, `engine`, `language` and `gender`. The hub leaves out fields it does not have, so each is shown as `null` when absent; a Gemini voice has no `language`, and its `engine` is the provider entry's model. [Source: .specify/archive/tiny/2026-09-27-list-tts-voices-command/list-tts-voices-command.md -> Context]
- **Public Hub Client**: The importable Multiroom Audio Hub API client — operations, request/response types, error types and error classes. Its exported surface is a semver-governed contract. [Source: .specify/archive/010-public-hub-package/spec.md -> "Public hub client"]
- **Error Class**: The category of a hub failure (e.g. network, timeout, not found, API error) with a friendly message; shared by all consumers. [Source: .specify/archive/010-public-hub-package/spec.md -> "Error class"]
- **Exit-Code Mapping**: CLI-only translation of an error class into a process exit code. [Source: .specify/archive/010-public-hub-package/spec.md -> "Exit-code mapping"]
- **Published Hub Spec**: The contents of `api/openapi.json`, importable at the same version as the client. [Source: .specify/archive/010-public-hub-package/spec.md -> "Published hub spec"]
- **Release Tag**: An immutable module version consumers pin to. [Source: .specify/archive/010-public-hub-package/spec.md -> "Release tag"]
- **Route Refusal**: A 409 answer to a route request, with a human-readable detail, an open-ended reason and, for output-level reasons only, the output that failed (`APIError.OutputID` is empty for `INPUT_DISABLED`, `GROUP_DISABLED` and `ROUTE_CHANGING_OUTPUTS`). `ROUTE_CHANGING_OUTPUTS` is the one reason that clears without the user doing anything; it classifies as a conflict (exit `14`) with the message "the route is still adding an output; retry in a moment", and every other reason, or none, keeps the hub's `detail`, else `title`, else the generic conflict text. [Source: .specify/archive/011-hub-0128-alignment/spec.md -> "Route refusal"; .specify/archive/011-hub-0128-alignment/data-model.md -> "Route refusal (`hub.APIError`, 409)"]
- **Join Mode**: how a route joins the outputs it targets: `REPLACE`, `MIX` or `DUCK_OTHERS`. It is set on a route or play request, or on an input as the default for its routes, and it is reported on each route. [Source: .specify/archive/012-join-modes/spec.md -> "Join mode"]
- **Route on Output**: one route as seen from one output. It has the route's own fields, plus its duck state on that output (`FULL` or `LOWERED`, open to more values) and whether it is the output's main route. A group route appears once under each member output it plays on. [Source: .specify/archive/012-join-modes/spec.md -> "Route on an output"]
- **Output Main Route**: the first route to join the output that is not a ducking route, or, if every route is ducking, the first route. [Source: .specify/archive/012-join-modes/spec.md -> "Output's main route"]
- **Hub Info**: identity and age of a running hub: `product`, `version`, `apiVersion`, `startedAt` and `uptimeSeconds`. `version` and `apiVersion` may read `unknown` and are passed through, never parsed as semver. [Source: .specify/archive/013-info-and-tts-warm/spec.md -> "Hub info"]
- **Warm Request**: the phrase and the voice choices that make up a cache entry's identity. [Source: .specify/archive/013-info-and-tts-warm/spec.md -> "Warm request"]
- **Warm Result**: whether the audio was already stored (`cacheHit`). [Source: .specify/archive/013-info-and-tts-warm/spec.md -> "Warm result"]
- **Chime**: a hub sound addressed by `id`, with its `durationMillis`. `none` is reserved for "no chime" and is never listed. [Source: .specify/archive/tiny/2026-11-20-route-chime/route-chime.md -> "Requirements"]

- **Radio Station**: one catalogue entry: UUID, playable address, name, favicon and homepage addresses, country code, languages, tags, codec, bitrate, whether it is HLS, whether the catalogue's last check succeeded, votes and recent click count. Most fields may be absent. [Source: .specify/archive/014-radio-station-search/spec.md -> "Station"; .specify/archive/014-radio-station-search/data-model.md -> "RadioStation"]
- **Radio Station Page**: a list of stations and the offset of the next page, absent on the last page. [Source: .specify/archive/014-radio-station-search/spec.md -> "Station page"; .specify/archive/014-radio-station-search/data-model.md -> "RadioStationPage"]
- **Radio Country** and **Radio Language**: a catalogue value and how many stations carry it. [Source: .specify/archive/014-radio-station-search/spec.md -> "Country, Language"; .specify/archive/014-radio-station-search/data-model.md -> "RadioCountry, RadioLanguage, RadioTag"]
- **Radio Tag** and **Radio Tag Page**: a tag with its station count, and a paged list of them. [Source: .specify/archive/014-radio-station-search/spec.md -> "Tag and Tag page"; .specify/archive/014-radio-station-search/data-model.md -> "RadioTagPage"]
- **Radio Problem**: a radio error's kind, title, detail, offending parameter and retry delay. [Source: .specify/archive/014-radio-station-search/spec.md -> "Radio problem"; .specify/archive/014-radio-station-search/data-model.md -> "Errors"]

## Success Criteria

### Measurable Outcomes

- **SC-001**: A user can view the full list of currently available outputs in under 1
  second under normal network conditions. [Source: .specify/archive/001-list-outputs/spec.md -> SC-001]
- **SC-002**: When the hub is unreachable or unresponsive, the user receives a clear
  failure message within 5 seconds (not an indefinite hang) in 100% of such cases, for the outputs, inputs, routes and groups commands.
  [Source: .specify/archive/001-list-outputs/spec.md -> SC-002; .specify/archive/002-outputs-get/spec.md -> SC-003; .specify/archive/003-inputs-list-get/spec.md -> SC-004; .specify/archive/004-routes-list-get/spec.md -> SC-004; .specify/archive/005-groups-list-get/spec.md -> SC-004]
- **SC-003**: 100% of machine-readable output produced by the outputs, inputs, routes, groups, play or route commands can be parsed by a
  standard structured-data parser without error. [Source: .specify/archive/001-list-outputs/spec.md -> SC-003; .specify/archive/002-outputs-get/spec.md -> SC-004; .specify/archive/003-inputs-list-get/spec.md -> SC-005; .specify/archive/004-routes-list-get/spec.md -> SC-005; .specify/archive/005-groups-list-get/spec.md -> SC-005; .specify/archive/006-play-command/spec.md -> SC-005; .specify/archive/008-route-command/spec.md -> SC-004]
- **SC-004**: A user can always tell, without ambiguity, whether "no outputs, inputs, routes or groups were found"
  or "the command failed" occurred, and for the play command whether playback was accepted or failed. [Source: .specify/archive/001-list-outputs/spec.md -> SC-004; .specify/archive/003-inputs-list-get/spec.md -> SC-006; .specify/archive/004-routes-list-get/spec.md -> SC-006; .specify/archive/005-groups-list-get/spec.md -> SC-006; .specify/archive/006-play-command/spec.md -> SC-006]
- **SC-005**: A user can identify a specific output's current volume, mute state, and
  availability from the command's output alone (a single invocation), without needing to
  run any other command or cross-reference a list.
  [Source: .specify/archive/001-list-outputs/spec.md -> SC-005; .specify/archive/002-outputs-get/spec.md -> SC-005]
- **SC-006**: A user can retrieve a specific output's full state in under 1 second under
  normal network conditions. [Source: .specify/archive/002-outputs-get/spec.md -> SC-001]
- **SC-007**: When the requested output does not exist, the user receives a clear "not
  found" message — never mistaken for an empty result or a crash — in 100% of such
  cases. [Source: .specify/archive/002-outputs-get/spec.md -> SC-002]
- **SC-008**: A user can view the full list of currently available inputs in under 1 second under normal network conditions. [Source: .specify/archive/003-inputs-list-get/spec.md -> SC-001]
- **SC-009**: A user can retrieve a specific input's full state in under 1 second under normal network conditions. [Source: .specify/archive/003-inputs-list-get/spec.md -> SC-002]
- **SC-010**: When the requested input does not exist, the user receives a clear "input not found" message — never mistaken for an empty result or a crash — in 100% of such cases. [Source: .specify/archive/003-inputs-list-get/spec.md -> SC-003]
- **SC-011**: A user can identify a specific input's source URI, enabled state, and source type from a single command invocation, without needing to run any other command or cross-reference a list. [Source: .specify/archive/003-inputs-list-get/spec.md -> SC-007]
- **SC-012**: A user can view the full list of currently active routes in under 1 second under normal network conditions. [Source: .specify/archive/004-routes-list-get/spec.md -> SC-001]
- **SC-013**: A user can retrieve a specific route's full state in under 1 second under normal network conditions. [Source: .specify/archive/004-routes-list-get/spec.md -> SC-002]
- **SC-014**: When the requested route does not exist, the user receives a clear "route not found" message — never mistaken for an empty result or a crash — in 100% of such cases. [Source: .specify/archive/004-routes-list-get/spec.md -> SC-003]
- **SC-015**: A user can identify a specific route's source input, target, and current status from a single command invocation, without needing to run any other command or cross-reference a list. [Source: .specify/archive/004-routes-list-get/spec.md -> SC-007]
- **SC-016**: A user can narrow a route listing to a single status, input, or target without needing to filter the output themselves after the fact. [Source: .specify/archive/004-routes-list-get/spec.md -> SC-008]
- **SC-017**: A user can view the full list of currently enabled groups in under 1 second under normal network conditions. [Source: .specify/archive/005-groups-list-get/spec.md -> SC-001]
- **SC-018**: A user can retrieve a specific group's full state in under 1 second under normal network conditions. [Source: .specify/archive/005-groups-list-get/spec.md -> SC-002]
- **SC-019**: When the requested group does not exist, the user receives a clear "group not found" message — never mistaken for an empty result or a crash — in 100% of such cases. [Source: .specify/archive/005-groups-list-get/spec.md -> SC-003]
- **SC-020**: A user can identify a specific group's member outputs and mute state from a single command invocation, without needing to run any other command or cross-reference a list. [Source: .specify/archive/005-groups-list-get/spec.md -> SC-007]
- **SC-021**: A user can view groups that exist but are currently disabled without needing to run a separate, differently-named command. [Source: .specify/archive/005-groups-list-get/spec.md -> SC-008]
- **SC-022**: A user can start playback to a single output or a group in one command invocation, without separately creating an input or a route. [Source: .specify/archive/006-play-command/spec.md -> SC-001]
- **SC-023**: A user can start playback with a chosen starting volume and/or a chosen display name in the same single command invocation. [Source: .specify/archive/006-play-command/spec.md -> SC-002]
- **SC-024**: A user receives a response confirming the hub has accepted the playback request (or a clear failure) in under 5 seconds under normal network conditions; this confirms the request was accepted, not that audio has been confirmed playing. [Source: .specify/archive/006-play-command/spec.md -> SC-003]
- **SC-025**: 100% of failure cases of the play command return a message that lets the user tell apart: bad input, target not found, route creation failure, unreachable source, unavailable service, and a connectivity/timeout problem. [Source: .specify/archive/006-play-command/spec.md -> SC-004]
- **SC-026**: 100% of the implemented inputs/outputs/groups/routes read commands are reachable only through the verb-first, path-style shape (`sonora get <resource>[/<id>]`, `sonora list <resource>`) — none remain reachable through the old `sonora <resource> <verb>` shape. [Source: .specify/archive/007-refactor-cli-commands/spec.md -> SC-001]
- **SC-027**: A user substituting any documented resource alias (`in`/`out`/`gr`/`rt`) for its full resource name gets identical output, for every refactored command. [Source: .specify/archive/007-refactor-cli-commands/spec.md -> SC-002]
- **SC-028**: A user running any pre-refactor command form (old noun-verb commands, or `play` with `--group`/`--output`/`--name`) receives a clear usage error in 100% of cases, rather than an incorrect or silently reinterpreted result. [Source: .specify/archive/007-refactor-cli-commands/spec.md -> SC-003]
- **SC-029**: For every refactored command, the data displayed (in both default and `--json` output) is byte-for-byte unchanged from before the refactor, given the same underlying hub state. [Source: .specify/archive/007-refactor-cli-commands/spec.md -> SC-004]
- **SC-030**: A user can connect an existing input to a single output or a group in one command invocation, without creating a duplicate input as a side effect. [Source: .specify/archive/008-route-command/spec.md -> SC-001]
- **SC-031**: A user receives a response confirming the hub has accepted the routing request (or a clear failure) in under 5 seconds under normal network conditions. [Source: .specify/archive/008-route-command/spec.md -> SC-002]
- **SC-032**: 100% of route command failure cases return a message that lets the user tell apart a usage error, input not found, target not found, route creation failure, and a connectivity or timeout problem. [Source: .specify/archive/008-route-command/spec.md -> SC-003]
- **SC-033**: A user can always tell, without ambiguity, whether the route was created successfully or the command failed. [Source: .specify/archive/008-route-command/spec.md -> SC-005]

- **SC-034**: A user can trigger an announcement on any output or group with one command and no more than the text and a target path. [Source: .specify/archive/009-tts-commands/spec.md -> SC-001]
- **SC-035**: With an audio cache hit, `speak` reports acceptance within 1 second of being run on a local network. [Source: .specify/archive/009-tts-commands/spec.md -> SC-002]
- **SC-036**: When the hub's provider times out at its default 10-second limit, the user sees the hub's provider-timeout error — not a CLI network timeout — 100% of the time. [Source: .specify/archive/009-tts-commands/spec.md -> SC-003]
- **SC-037**: A user whose hub is configured with a provider timeout longer than 10 seconds can, by supplying `--timeout`, still receive the hub's own response rather than a CLI network timeout, 100% of the time. [Source: .specify/archive/009-tts-commands/spec.md -> SC-003a]
- **SC-038**: Every failure listed in Edge Cases exits with the exit code class documented for it (FR-113; usage errors → 2) and a message that names the cause, verified by tests; no failure prints a raw program error or stack trace without `--verbose`. [Source: .specify/archive/009-tts-commands/spec.md -> SC-004]
- **SC-039**: Against a hub where the TTS extension is not active, 100% of TTS commands report "TTS not available" rather than a misleading not-found error. Whenever the hub's extension inventory is reachable, they also name the reason (not installed, disabled, rejected with its reason, inert, or version mismatch). [Source: .specify/archive/009-tts-commands/spec.md -> SC-005]
- **SC-040**: A successful TTS command makes exactly one request to the hub. [Source: .specify/archive/009-tts-commands/spec.md -> SC-005a]
- **SC-041**: Existing commands' behaviour and startup time are unchanged by this feature. [Source: .specify/archive/009-tts-commands/spec.md -> SC-006]
- **SC-042**: A `list tts-voices` call makes exactly one request to the hub, and its output parses as YAML or JSON with every voice field present, whatever the hub omitted. [Source: .specify/archive/tiny/2026-09-27-list-tts-voices-command/list-tts-voices-command.md -> Requirements 1, 5]
- **SC-043**: A project outside this repository can depend on a released tag and complete one read call, one write call and one error classification against a fake hub, with zero workarounds (no replace directives, workspace files or private-module settings). [Source: .specify/archive/010-public-hub-package/spec.md -> SC-001]
- **SC-044**: 100% of existing tests pass on both Windows and Linux after the refactor, with no test weakened. Tests that assert exit codes through the client move to the CLI-internal mapping with the same assertions; tests of hub message text change only where FR-136 removes CLI wording, and the CLI tests that check that wording stay unchanged. [Source: .specify/archive/010-public-hub-package/spec.md -> SC-002]
- **SC-045**: For every existing command, output and exit code are identical before and after the refactor across the success path and each error class. Evidence: the existing unit, integration and contract suites pass unchanged in intent (they cover every command and error class), supplemented by a manual before/after comparison of representative commands. [Source: .specify/archive/010-public-hub-package/spec.md -> SC-003]
- **SC-046**: 0 references to the old module path or the old internal client location remain in code, tests, build/release scripts or current docs (README, CONTRIBUTING, AGENTS.md). Historical records — earlier feature specs under `specs/`, `docs/reviews/`, and the constitution's Sync Impact Report in `.specify/` — are excluded. [Source: .specify/archive/010-public-hub-package/spec.md -> SC-004]
- **SC-047**: 100% of the public client's exported identifiers have documentation, and 0 of those docs reference internal spec artifacts or constitution principles. [Source: .specify/archive/010-public-hub-package/spec.md -> SC-005]
- **SC-048**: The public client pulls in 0 third-party dependencies. [Source: .specify/archive/010-public-hub-package/spec.md -> SC-006]
- **SC-049**: A build with a supplied version reports that exact version in 100% of release builds. [Source: .specify/archive/010-public-hub-package/spec.md -> SC-007]
- **SC-050**: Every statement about enabling and disabling in the README and in the six enable/disable help texts agrees with the hub 0.1.28 contract; a review finds zero contradictions. [Source: .specify/archive/011-hub-0128-alignment/spec.md -> SC-001]
- **SC-051**: For a play of a link that the hub names, the user can read the input's name from the command's output on the first try, without a second command. [Source: .specify/archive/011-hub-0128-alignment/spec.md -> SC-002]
- **SC-052**: A transfer refused once as transient completes successfully with no user action, and the whole command takes less than 1 second longer than an immediate success. [Source: .specify/archive/011-hub-0128-alignment/spec.md -> SC-003]
- **SC-053**: A transfer that stays refused as transient ends within 3 seconds of being started, with exit code 14 and a message that names the transient cause. [Source: .specify/archive/011-hub-0128-alignment/spec.md -> SC-004]
- **SC-054**: The output of every command other than `play` and `transfer`, and of `play` against a hub that omits the display name, is byte-for-byte identical to the release before the hub 0.1.25–0.1.28 alignment in the existing test suite. [Source: .specify/archive/011-hub-0128-alignment/spec.md -> SC-005]

- **SC-055**: A rename or `--reset` changes the input's display name on the hub with one request, and the record printed afterwards shows the name the hub stored, so an operator can confirm the change at once. [Source: .specify/archive/tiny/2026-10-22-rename-inputs-command/rename-inputs-command.md]
- **SC-056**: A user can play an announcement over running music without stopping it, in one command. [Source: .specify/archive/012-join-modes/spec.md -> SC-001]
- **SC-057**: A user can find out everything playing on an output, including group routes, in one command, without cross-referencing `list routes` and `get groups`. [Source: .specify/archive/012-join-modes/spec.md -> SC-002]
- **SC-058**: Against a hub older than 0.1.21, `get outputs/<id> --include-routes` never says that an existing output was not found. Its message names the hub version required. [Source: .specify/archive/012-join-modes/spec.md -> SC-003]
- **SC-059**: Every command run without the new flags sends the same requests and prints the same output as the previous release, byte for byte, across the existing test suite, apart from the join-mode fields FR-173 and FR-174 add to the route and input YAML. [Source: .specify/archive/012-join-modes/spec.md -> SC-004]
- **SC-060**: An invalid join mode is rejected before any request reaches the hub, in 100% of the tested cases. [Source: .specify/archive/012-join-modes/spec.md -> SC-005]
- **SC-061**: Contract tests for the hub 0.1.29 volume and ducking fields pass for each response type that carries them (route, route on output, input) with the field set, null and absent, and for each request that sends them, with the field set and unset. [Source: .specify/archive/tiny/2026-10-25-hub-route-volume-client/hub-route-volume-client.md]

- **SC-062**: Each new volume and ducking flag, the set command and the removed `--volume` flag behave as FR-183 to FR-185 state in 100% of the tested cases, and each rejected invocation sends no request. [Source: .specify/archive/tiny/2026-10-25-route-volume-cli/route-volume-cli.md]

- **SC-063**: For each extension status that FR-110's table names, the hub package's diagnosis gives the same reason the CLI gives, with one inventory request per diagnosis. [Source: .specify/archive/tiny/2026-10-28-hub-tts-unavailable-diagnosis/hub-tts-unavailable-diagnosis.md]

- **SC-064**: A marshalled `TTSCacheStats` carries the four property names of the hub's `CacheStats` schema in every contract test, and `sonora get tts-cache` prints the same output as before. [Source: .specify/archive/tiny/2026-10-28-tts-cache-stats-json-tags/tts-cache-stats-json-tags.md]

- **SC-065**: In every contract and unit test, the seven `SpeakRequest` fields are absent from the body when unset and present when set, and each new `speak` flag sends its field. [Source: .specify/archive/tiny/2026-10-27-speak-request-fields/speak-request-fields.md]
- **SC-066**: An operator learns a hub's version and whether it restarted since the last check with one command and no knowledge of the REST API. [Source: .specify/archive/013-info-and-tts-warm/spec.md -> SC-001]
- **SC-067**: After warming a phrase, announcing the same phrase with the same provider, voice and language reports a cache hit in 100% of tries with the local provider. [Source: .specify/archive/013-info-and-tts-warm/spec.md -> SC-002]
- **SC-068**: Warming a phrase that is already stored makes no provider call and succeeds. [Source: .specify/archive/013-info-and-tts-warm/spec.md -> SC-003]
- **SC-069**: Neither command produces sound or changes any speaker, output, group or route; verified by running both against the live hub with nothing playing afterwards. [Source: .specify/archive/013-info-and-tts-warm/spec.md -> SC-004]
- **SC-070**: Each failure listed in the acceptance scenarios gives a distinct message and a non-zero exit code, covered by an automated test. [Source: .specify/archive/013-info-and-tts-warm/spec.md -> SC-005]

- **SC-071**: A route-starting command without `--chime` sends the same request body as before this change, and with it sends `chime` verbatim. [Source: .specify/archive/tiny/2026-11-20-route-chime/route-chime.md -> Requirements 2, 4]
- **SC-072**: `sonora list chimes` and `sonora get chimes` return the hub's chimes read-only, against a hub that has them, with nothing started. [Source: .specify/archive/tiny/2026-11-20-route-chime/route-chime.md -> Verification]
- **SC-073**: `sonora get` and `sonora list` with no resource name print a usage error that names every resource the verb accepts. [Source: .specify/archive/tiny/2026-11-20-route-chime/route-chime.md -> Requirement 9]

- **SC-074**: An operator goes from a station name to hearing that station with two commands (search, play) and no web browser. [Source: .specify/archive/014-radio-station-search/spec.md -> SC-001]
- **SC-075**: Finding a station by part of its name takes one command and one request to the hub, whatever the size of the catalogue. [Source: .specify/archive/014-radio-station-search/spec.md -> SC-002]
- **SC-076**: Against a hub where the radio extension is not serving, 100% of radio commands give the same exit code and a message naming the reason; each reason listed under Edge Cases has its own message, covered by an automated test. [Source: .specify/archive/014-radio-station-search/spec.md -> SC-003]
- **SC-077**: Catalogue text containing escape sequences cannot change the terminal's state, verified by an automated test with such text. [Source: .specify/archive/014-radio-station-search/spec.md -> SC-004]
- **SC-078**: No radio command produces sound or changes any speaker, output, group, input or route; verified by running them against the live hub with nothing new playing afterwards. [Source: .specify/archive/014-radio-station-search/spec.md -> SC-005]
- **SC-079**: Each other failure listed in the acceptance scenarios and edge cases gives a distinct message and a non-zero exit code, covered by an automated test. [Source: .specify/archive/014-radio-station-search/spec.md -> SC-006]

- **SC-080**: A `create` with `--persistent` and `--auto-remove` sends no request to the hub, and a `create` without `--persistent` sends the same body as before; verified by contract tests against a test hub that records each request. [Source: .specify/archive/tiny/2026-11-22-stored-inputs/stored-inputs.md -> Requirement 1]
- **SC-081**: Each of the three new input 409 reasons gives a message that differs from the generic conflict and exits `14`, and an unknown reason still exits `14` as a generic conflict; verified by automated tests. [Source: .specify/archive/tiny/2026-11-22-stored-inputs/stored-inputs.md -> Requirement 3]

## Assumptions

- **AS-001**: The hub's network location (host/port) is available to the CLI through a mechanism
  established once (e.g. a default matching the API's documented default server, overridable
  by the user); this feature consumes that mechanism but does not define it.
  [Source: .specify/archive/001-list-outputs/spec.md -> "The hub's network location (host/port) is available"; .specify/archive/002-outputs-get/spec.md -> "The hub's network location (host/port) is available"; .specify/archive/003-inputs-list-get/spec.md -> "The hub's network location (host/port) is available"; .specify/archive/004-routes-list-get/spec.md -> "The hub's network location (host/port) is available"; .specify/archive/005-groups-list-get/spec.md -> "The hub's network location (host/port) is available"; .specify/archive/006-play-command/spec.md -> "The hub's network location (host/port) is available"; .specify/archive/008-route-command/spec.md -> "The hub's network location (host/port) is available"]
- **AS-002**: The hub API requires no authentication for the read-only endpoints and the creation endpoints (`play`, `route`), consistent with the
  current API specification. The TTS operations (`speak`, `get tts-cache`, `clear tts-cache`) are guarded exactly as the rest of the hub's HTTP surface, so the CLI adds no authentication for them. [Source: .specify/archive/009-tts-commands/spec.md -> "No authentication is added"; .specify/archive/001-list-outputs/spec.md -> "The hub API requires no authentication"; .specify/archive/002-outputs-get/spec.md -> "The hub API requires no authentication"; .specify/archive/003-inputs-list-get/spec.md -> "The hub API requires no authentication"; .specify/archive/004-routes-list-get/spec.md -> "The hub API requires no authentication"; .specify/archive/005-groups-list-get/spec.md -> "The hub API requires no authentication"; .specify/archive/006-play-command/spec.md -> "The hub API requires no authentication"; .specify/archive/008-route-command/spec.md -> "The hub API requires no authentication"]
- **AS-003**: "Human-readable" default output presents the same set of fields as the machine-readable
  output, formatted for readability rather than strict parsing.
  [Source: .specify/archive/001-list-outputs/spec.md -> "\"Human-readable\" default output presents the same set of fields"; .specify/archive/002-outputs-get/spec.md -> "\"Human-readable\" default output presents the same set of fields"; .specify/archive/003-inputs-list-get/spec.md -> "\"Human-readable\" default output presents the same set of fields"; .specify/archive/004-routes-list-get/spec.md -> "\"Human-readable\" default output presents the same set of fields"; .specify/archive/005-groups-list-get/spec.md -> "\"Human-readable\" default output presents the same set of fields"; .specify/archive/006-play-command/spec.md -> "\"Human-readable\" default output presents the same set of fields"; .specify/archive/008-route-command/spec.md -> "\"Human-readable\" default output presents the same set of fields"]
- **AS-004**: The default request timeout is 5 seconds. [Source: .specify/archive/001-list-outputs/spec.md -> "The default request timeout is 5 seconds"; .specify/archive/002-outputs-get/spec.md -> "The default request timeout is 5 seconds"; .specify/archive/003-inputs-list-get/spec.md -> "The default request timeout is 5 seconds"; .specify/archive/004-routes-list-get/spec.md -> "The default request timeout is 5 seconds"; .specify/archive/005-groups-list-get/spec.md -> "The default request timeout is 5 seconds"; .specify/archive/006-play-command/spec.md -> "The default request timeout is 5 seconds"; .specify/archive/008-route-command/spec.md -> "The default request timeout is 5 seconds"]
- **AS-005**: Outputs, inputs, routes and groups are unordered by the hub API; the commands may present them in the order received
  without imposing additional sorting guarantees.
  [Source: .specify/archive/001-list-outputs/spec.md -> "Outputs are unordered by the hub API"; .specify/archive/003-inputs-list-get/spec.md -> "Inputs are unordered by the hub API"; .specify/archive/004-routes-list-get/spec.md -> "Routes are unordered by the hub API"; .specify/archive/005-groups-list-get/spec.md -> "Groups are unordered by the hub API"]
- **AS-006**: The output, input, route and group identifier is an opaque, case-sensitive string as returned by the hub; the
  get commands perform no client-side normalization or fuzzy matching.
  [Source: .specify/archive/002-outputs-get/spec.md -> "The output identifier is an opaque, case-sensitive string"; .specify/archive/003-inputs-list-get/spec.md -> "The input identifier is an opaque, case-sensitive string"; .specify/archive/004-routes-list-get/spec.md -> "The route identifier is an opaque, case-sensitive string"; .specify/archive/005-groups-list-get/spec.md -> "The group identifier is an opaque, case-sensitive string"]
- **AS-007**: Inputs have no volume, mute, or hardware-availability concept in the hub API; the inputs commands display the attributes the API defines for inputs (source URI, source type, auto-remove, pauseable, creation timestamp) rather than an outputs-shaped display.
  [Source: .specify/archive/003-inputs-list-get/spec.md -> "Unlike outputs, inputs have no volume, mute, or hardware-availability concept"]
- **AS-008**: The routes commands are read-only: creating, stopping, transferring, and pausing or resuming routes are out of scope for this feature, mirroring how the outputs and inputs read commands preceded their mutation commands.
  [Source: .specify/archive/004-routes-list-get/spec.md -> "Creating, stopping, transferring, and pausing/resuming routes are out of scope for this feature"]
- **AS-009**: The groups commands are read-only: creating, deleting, and modifying groups (membership, volume, mute, enabled state) are out of scope for this feature, mirroring AS-008.
  [Source: .specify/archive/005-groups-list-get/spec.md -> "Creating, deleting, and modifying groups (membership, volume, mute, enabled state) are out of scope for this feature"]
- **AS-010**: Stopping, pausing, or otherwise managing the route and input created by playback is out of scope for this feature; it covers only starting playback. Managing the resulting route/input is expected to be handled by future route/input management commands. [Source: .specify/archive/006-play-command/spec.md -> "Stopping, pausing, or otherwise managing the route and input created by playback is out of scope for this feature"]
- **AS-011**: The set of supported URI schemes (e.g. YouTube, SoundCloud, direct stream URLs) is determined entirely by the hub; the CLI performs no client-side allow-listing of schemes and passes the URI through as given. [Source: .specify/archive/006-play-command/spec.md -> "The set of supported URI schemes"]
- **AS-013**: The hub's playback response reflects only the route's initial state (typically `STARTING`); it is the caller's responsibility to poll `routes get` afterward if confirmation that audio has actually begun playing is needed — that polling is out of scope for this command. [Source: .specify/archive/006-play-command/spec.md -> "The hub's playback response reflects only the route's initial state"]
- **AS-014**: Only commands already implemented are in scope of the verb-first refactor: the `inputs`/`outputs`/`groups`/`routes` `list`/`get` commands and `play`. The `route` command is out of scope here; it is specified separately.
  [Source: .specify/archive/007-refactor-cli-commands/spec.md -> "Only commands that are already implemented today are in scope"]
- **AS-015**: Commands documented in the command landscape but not yet implemented (`create`, `delete`, `enable`/`disable`, `mute`/`unmute`, `set ... volume`, `transfer`, `stop`/`pause`/`resume`, `master-mute`) are out of scope; the refactor only brought already-shipped commands in line with the documented grammar.
  [Source: .specify/archive/007-refactor-cli-commands/spec.md -> "Commands documented in `docs/cli-command-landscape.md` that aren't implemented yet"]
- **AS-016**: `--json`, `--hub-url`, and `--verbose` are unaffected by the refactor; it changed only command and argument parsing, not request behavior or output rendering.
  [Source: .specify/archive/007-refactor-cli-commands/spec.md -> "`--json`, `--hub-url`, and `--verbose` are unaffected by this refactor"]
- **AS-017**: The hub API and its response shapes are unchanged by the refactor; it is a CLI-surface-only change.
  [Source: .specify/archive/007-refactor-cli-commands/spec.md -> "The underlying hub API and response shapes are unchanged"]
- **AS-018**: The route command accepts no starting volume or display-name option. Those apply only to the ephemeral input `play` creates; an existing input's name and settings are managed elsewhere. [Source: .specify/archive/008-route-command/spec.md -> "This command does not accept a starting volume or a display-name option"]
- **AS-019**: Stopping, pausing, transferring, or otherwise managing the route the route command creates is out of scope; the command covers only creating the route, mirroring AS-010. The route command's own scope is unchanged; stopping, pausing and transferring are separate commands (FR-083, FR-085, FR-097, FR-098). [Source: .specify/archive/008-route-command/spec.md -> "Stopping, pausing, or otherwise managing the resulting route"; .specify/archive/tiny/2026-09-09-stop-routes-command/stop-routes-command.md]
- **AS-020**: Whether an input can be routed to more than one target at once, or whether routing an already-routed input to the same target succeeds, is decided by the hub; the CLI applies no client-side restriction and surfaces the outcome the hub reports. [Source: .specify/archive/008-route-command/spec.md -> "Whether a given input can be routed to more than one target at once"]
- **AS-021**: Output and group identifiers may collide; the target's path prefix decides which one is meant, so the route command needs no disambiguation flag. [Source: .specify/archive/008-route-command/spec.md -> "Output identifiers and group identifiers may occasionally collide"]
- **AS-022**: Each mutating command (`enable`, `disable`, `rename`, `delete`, `stop`, `create`, `transfer`, `set`, `mute`, `unmute`, `pause`, `resume`, `speak`, `clear tts-cache`) sends one request and reports the hub's result; the one exception is `transfer`, which repeats its request after a transient `ROUTE_CHANGING_OUTPUTS` refusal, up to five attempts (FR-146) [Source: .specify/archive/011-hub-0128-alignment/spec.md -> FR-008]. It does not ask for confirmation and does not read the current state first, so repeating a command is whatever the hub makes of it. `clear tts-cache` loses no user data, only costs a re-synthesis on the next request. [Source: .specify/archive/009-tts-commands/spec.md -> "clear does not prompt for confirmation"; .specify/archive/tiny/2026-08-28-inputs-enable-disable-command/inputs-enable-disable-command.md; .specify/archive/tiny/2026-08-29-route-pause-resume-command/route-pause-resume-command.md; .specify/archive/tiny/2026-10-22-rename-inputs-command/rename-inputs-command.md]
- **AS-023**: The hub identifies a TTS target by the same id the CLI already uses for outputs and groups (`outputs/<id>`, `groups/<id>`); the extension's `targetName` field carries that id. [Source: .specify/archive/009-tts-commands/spec.md -> "The hub identifies a TTS target by the same id"]
- **AS-024**: The TTS operations live under the extension's own paths, not the core v2 API, and answer 404 when the extension is not active. The extension never uses 404 itself: it reports unknown targets and providers as 400. So a 404 is conclusive evidence that TTS is unavailable. [Source: .specify/archive/009-tts-commands/spec.md -> "The TTS operations live under the extension's own paths"]
- **AS-025**: The TTS extension's inventory id is `tts`, the `Extension-Id` in its distributable's manifest. Entries are matched on this machine identifier, never on the display name, as the hub's feature `019-extension-shared-classloader` (FR-002, FR-030) requires of consumers. [Source: .specify/archive/009-tts-commands/spec.md -> "The TTS extension's inventory id is `tts`"]
- **AS-026**: Per `019-extension-shared-classloader` FR-030 to FR-032, core contributes no HTTP API path. The inventory operation (`/api/v2/extensions`) is served by the control API (REST) extension, and the TTS operations are served by the TTS extension. That is why the two lookups separate the causes: a 404 from TTS alone means TTS is unavailable, and a 404 from both means the REST extension is absent or the URL is wrong. The CLI MUST NOT fall back to the core-owned operational monitoring endpoint (`/actuator/extensions`), for three reasons: it is not in `api/openapi.json`, its shape is explicitly not a published contract, and as a diagnostic tool it may be reachable only from the hub's own host, so a remote CLI cannot rely on it. [Source: .specify/archive/009-tts-commands/spec.md -> "Per `019-extension-shared-classloader` FR-030 to FR-032"]
- **AS-027**: A speak request's response wait defaults to 15 seconds (the hub's 10-second default synthesis timeout plus margin), overridable per invocation with `--timeout` (FR-115) for hubs configured with a longer provider timeout. [Source: .specify/archive/009-tts-commands/spec.md -> "A speak request's response wait defaults to 15 seconds"]
- **AS-028**: The `speak` verb takes the text as its first positional argument and the target as its second, mirroring `play <uri> <target>`. The text may be `-` to read it from standard input (FR-102). [Source: .specify/archive/009-tts-commands/spec.md -> "The `speak` verb takes the text as its first positional argument"]
- **AS-029**: A user-facing command for the extension inventory (e.g. `get extensions`) and the legacy `/api/**` (non-`v2`) operations present in `api/openapi.json` are out of scope. This feature uses `listExtensions` only internally, for FR-110's diagnosis. [Source: .specify/archive/009-tts-commands/spec.md -> "A user-facing command for the extension inventory"]
- **AS-030**: Queue inspection, cancelling a queued announcement, and waiting for playback to finish are out of scope — the hub exposes no operation for them. [Source: .specify/archive/009-tts-commands/spec.md -> "Queue inspection, cancelling a queued announcement"]
- **AS-032**: **Dependency**: The sonora-cli constitution is amended first (Phase 1 of the source plan, 1.1.3 → 1.2.0) to add the "Public Hub Client Package" rules and to name `hub/` in Principle II as the single place the spec becomes Go types. That amendment is done via `/speckit-constitution`, not by this feature, and must land before or with this feature's merge. [Source: .specify/archive/010-public-hub-package/spec.md -> "The sonora-cli constitution is amended first"]
- **AS-033**: The repository at `github.com/Sonora-Multiroom/sonora-cli` is public, so consumers need no private-module configuration. [Source: .specify/archive/010-public-hub-package/spec.md -> "The repository at"]
- **AS-034**: While the module is `v0.x`, breaking changes to the public client are allowed but must be called out in release notes; the first release with the public client is `v0.1.0` unless an existing `v0.x` tag sequence dictates the next patch/minor number. [Source: .specify/archive/010-public-hub-package/spec.md -> "While the module is `v0.x`"]
- **AS-035**: Publishing the spec (FR-133 / User Story 53) is recommended by the source plan; it is in scope but may be dropped at planning time without blocking User Stories 51 and 52. [Source: .specify/archive/010-public-hub-package/spec.md -> "Publishing the spec"]
- **AS-036**: Counts from the source plan were verified against the repo on 2026-09-27 and bound the scope of mechanical changes: 111 files import the module path and 77 import the internal client; 30 CLI files use the exit-code mapping (the source plan said 34). [Source: .specify/archive/010-public-hub-package/spec.md -> "Counts from the source plan"]
- **AS-037**: Out of scope: the sonora-mcp rewrite, new hub operations, new CLI commands, making config discovery public, and splitting the client into its own repository (option 2b). [Source: .specify/archive/010-public-hub-package/spec.md -> "Out of scope: the sonora-mcp rewrite"]
- **AS-038**: The production hub runs 0.1.28 (checked 2026-10-22); older hubs must still work, so every new response field is optional on the client side; hubs older than 0.1.21 must also keep working for every command that does not use the new join-mode flags (CLI-012). [Source: .specify/archive/011-hub-0128-alignment/spec.md -> "The production hub runs 0.1.28"] [Source: .specify/archive/012-join-modes/spec.md -> "The production hub runs 0.1.28"]
- **AS-039**: A transfer retry window of about 2 seconds is generous: the hub's own measurement of a rejoin was 0.16 s, and the contract says the `ROUTE_CHANGING_OUTPUTS` refusal clears "within moments". The exact delays are a planning decision. [Source: .specify/archive/011-hub-0128-alignment/spec.md -> "A retry window of about 2 seconds is generous"]
- **AS-040**: Retrying belongs to the CLI command, not to the hub client's transfer operation, so each client operation keeps sending exactly one request and other consumers choose their own policy. The client may offer a way to recognise the transient reason. [Source: .specify/archive/011-hub-0128-alignment/spec.md -> "Retrying belongs to the CLI command"]
- **AS-041**: Typed constants for the refusal reasons are not required. The earlier decision (tinyspec `decode-route-refusals-409`) to keep the reason an open string stands; named values may be added only if they keep it open. [Source: .specify/archive/011-hub-0128-alignment/spec.md -> "Typed constants for the refusal reasons are not required"]
- **AS-042**: `sonora transfer` already checks that the target exists before transferring; that check is not repeated on retries. [Source: .specify/archive/011-hub-0128-alignment/spec.md -> "`sonora transfer` already checks that the target exists"]
- **AS-043**: No prompt, warning or `--force` flag is added to `sonora disable`; the backlog item leaves that as an open question for later. [Source: .specify/archive/011-hub-0128-alignment/spec.md -> "No prompt, warning or `--force` flag is added to `sonora disable`"]
- **AS-044**: sonora-mcp picks up the hub client changes of the hub 0.1.25–0.1.28 alignment and of join modes (CLI-012) with its next update of this module; changing sonora-mcp is out of scope. [Source: .specify/archive/011-hub-0128-alignment/spec.md -> "sonora-mcp picks up the client changes"] [Source: .specify/archive/012-join-modes/spec.md -> "sonora-mcp picks up the client changes"]
- **AS-045**: Name length, allowed characters and trimming are the hub's rules, not the CLI's: the CLI sends the name as given, and a name the hub refuses is reported with the hub's detail. A rename does not persist past a restart for a configured input, and a runtime input's rename ends when it is removed (FR-158). [Source: .specify/archive/tiny/2026-10-22-rename-inputs-command/rename-inputs-command.md]
- **AS-046**: **List views stay as they are** (the backlog's open question, settled in Clarifications). `list routes` and `list inputs` keep their fixed fields. The join mode appears in `get routes/<id>`, `get inputs/<id>` and in the routes shown by `--include-routes`, which is where a user asks about one route or one output. Adding columns to the lists would change their shape for every script that reads them, for little gain. [Source: .specify/archive/012-join-modes/spec.md -> "List views stay as they are"]
- **AS-047**: `sonora play` and `sonora route` do not show the join mode in their own output. Their output stays as it is, and `get routes/<id>` shows the join mode afterwards. [Source: .specify/archive/012-join-modes/spec.md -> "sonora play and sonora route do not show the join mode"]
- **AS-048**: sonora-tts's announcement playback mode is out of scope until sonora-tts documents it. Amended by speak-request-fields: the hub 0.1.29 `SpeakRequest` documents `playbackMode`, so `sonora speak --playback-mode` exposes it (FR-191). [Source: .specify/archive/012-join-modes/spec.md -> "sonora-tts's announcement playback mode"]
- **AS-049**: The flag check mirrors `--status` on `list routes`: the valid values are a fixed list in the CLI, taken from the contract's enum. [Source: .specify/archive/012-join-modes/spec.md -> "The flag check mirrors"]
- **AS-050**: The exit code for an older hub on `--include-routes` (3) follows the README's meaning "hub returned an unexpected response". The hub answered 404 to a request for an output it had just returned. [Source: .specify/archive/012-join-modes/spec.md -> "The exit code for an older hub on --include-routes"]
- **AS-051**: If the hub later adds `?includeRoutes=true` to the output request (proposed in `sonora-multiroom/docs/backlog/routes-per-output-in-outputs-list.md`), the flag can switch to one request without changing what users see. That switch is not part of this feature. [Source: .specify/archive/012-join-modes/spec.md -> "If the hub later adds"]
- **AS-052**: The `hub` package's volume and ducking fields follow the hub's 0.1.29 schema in `api/openapi.json`. The `SpeakRequest` route fields are not part of this change and are left to the speak-request-fields tinyspec. [Source: .specify/archive/tiny/2026-10-25-hub-route-volume-client/hub-route-volume-client.md]
- **AS-053**: `--volume` is not kept as an alias or a deprecated notice, so scripts that use it fail at once with exit `2` and must be updated. The removal is released as a breaking change. [Source: .specify/archive/tiny/2026-10-25-route-volume-cli/route-volume-cli.md]
- **AS-054**: `--playback-mode` is passed to the hub unchecked, because the spec has no enum for it; the hub's 400 is the only check on its value. [Source: .specify/archive/tiny/2026-10-27-speak-request-fields/speak-request-fields.md]
- **AS-055**: The CLI warm command takes the same voice options as `speak` (`--provider`, `--voice`, `--language`), so a warmed phrase matches the announcements the CLI makes; `engine`, `pitch` and `speakingRate` are in the `hub` function only; `stylePrompt` is also a CLI flag since FR-204. [Source: .specify/archive/013-info-and-tts-warm/spec.md -> "The CLI warm command takes the same voice options as `speak`"]
- **AS-056**: Warming several phrases is a shell loop for now; a batch form is out of scope. [Source: .specify/archive/013-info-and-tts-warm/spec.md -> "Warming several phrases is a shell loop for now"]
- **AS-057**: Command spelling follows the existing verb-first style, `sonora get info` and `sonora warm tts-cache <text|->`. [Source: .specify/archive/013-info-and-tts-warm/spec.md -> "Command spelling follows the existing verb-first style"]
- **AS-058**: `Group.system` is already mirrored by a separate change and is not part of this feature. [Source: .specify/archive/013-info-and-tts-warm/spec.md -> "`Group.system` is already mirrored"]
- **AS-059**: The hub info schema has no `required` list; the client requires `product`, `version` and `apiVersion` because a body without them does not identify a hub. [Source: .specify/archive/013-info-and-tts-warm/spec.md -> "The hub info schema has no `required` list"]
- **AS-060**: Hubs before 0.1.33 keep working for every other command; only the two new commands need 0.1.33. [Source: .specify/archive/013-info-and-tts-warm/spec.md -> "Hubs before 0.1.33 keep working"]
- **AS-061**: `--style-prompt` is passed to the hub unchecked, as `--playback-mode` is (AS-054); the hub's 400 is the only check on its value, including a `stylePrompt` sent to a provider other than `google-gemini`. [Source: .specify/archive/tiny/2026-11-12-speak-style-prompt-flag/speak-style-prompt-flag.md]
- **AS-062**: `chimes` is addressed like `outputs` but is not a play or route target, so it stays out of the `respath` resource enum; `respath.ValidID` holds the id rule it shares. [Source: .specify/archive/tiny/2026-11-20-route-chime/route-chime.md -> Requirement 7]
- **AS-063**: The hub has no `GET /api/v2/chimes/{id}`, so `get chimes/<id>` lists the chimes and picks the entry. [Source: .specify/archive/tiny/2026-11-20-route-chime/route-chime.md -> Requirement 7]
- **AS-064**: `api/openapi.json` is the hub's 0.1.35 document, a superset of the 0.1.34 this feature needs. Its 0.1.35 additions (persistent inputs, `SpeakRequest.chime`, the radio extension paths) are not adopted here. [Annotated by .specify/archive/tiny/2026-11-22-stored-inputs/stored-inputs.md: persistent inputs were adopted by HUB-041, see FR-228–FR-232 and AS-072; the other additions are unchanged.] [Source: .specify/archive/tiny/2026-11-20-route-chime/route-chime.md -> Verification]
- **AS-065**: the constitution requires verb-first commands, so radio uses `radio-` nouns under verbs rather than a `radio` group. Search gets its own `search` verb because a bare `get radio-stations` that returns the 20 most voted would surprise; the prefix makes the extension dependency plain in help, so no extra help note is needed. [Source: .specify/archive/014-radio-station-search/spec.md -> "Command layout"]
- **AS-066**: the search output prints the playable address, which `sonora play` already accepts. Playing a search result by position or by name, and a `play --station` shorthand, are out of scope: they hide which of several same-named stations plays, and `play` takes the address as an argument, not a flag. [Source: .specify/archive/014-radio-station-search/spec.md -> "Play by UUID only"]
- **AS-067**: radio commands exist on every hub and explain a missing extension only after the radio request fails, so a working hub costs one request (Principle I). [Source: .specify/archive/014-radio-station-search/spec.md -> "Always present"]
- **AS-068**: fixtures are hand-written from `api/openapi.json` and checked against it, not copied from the extension's test resources. [Source: .specify/archive/014-radio-station-search/spec.md -> "Test fixtures"]
- **AS-069**: they take no paging options. [Source: .specify/archive/014-radio-station-search/spec.md -> "Countries and languages are short lists returned whole"]
- **AS-070**: the backlog's `--all` (follow every page, drop repeated UUIDs) is out of scope. Its use, listing everything to `grep` it, is what the name filter already does in one request; paging stays available through `--offset`, and a Go caller pages with the `hub` package. [Source: .specify/archive/014-radio-station-search/spec.md -> "No fetch-all"]
- **AS-071**: follow this feature as separate work. [Source: .specify/archive/014-radio-station-search/spec.md -> "Releasing a tag and adding the tools to sonora-mcp"]
- **AS-072**: the hub decides what is kept and whether a kept input can play. The CLI sends `persistent`, shows `origin`, `playable` and `unplayableReason`, and reports the hub's refusals; it does not check the store, the extension list or the 200-input cap itself. [Source: .specify/archive/tiny/2026-11-22-stored-inputs/stored-inputs.md -> "What"]
