# Changelog

## Merged Features Log

### Chimes on Routes (`play --chime`, `route --chime`, `list chimes`, `get chimes`) — archived 2026-11-24
**Tinyspec:** .specify/archive/tiny/2026-11-20-route-chime/route-chime.md
**Implemented:** 2026-11-20

`sonora play` and `sonora route` take `--chime <id|none>`, sent verbatim and only when given; `sonora list chimes` and `sonora get chimes[/<id>]` read the hub's chimes, and `get outputs`, `get groups` and `get inputs` show a chime only when one is set. `get` and `list` with no resource name now name every resource their verb accepts. The `hub` package gains `Chime`, `ListChimes` and `ChimeNone`, and the `chime` field on play, route, input creation and the output, group and input records (HUB-040, hub 0.1.34; FR-206–FR-213, SC-071–SC-073, AS-062–AS-064, User Story 67). `api/openapi.json` is the hub's 0.1.35 document, of which only the 0.1.34 subset is adopted. The release tag and the `sonora-mcp` bump are still to be done.

### Style Prompt on `speak` and `warm tts-cache` (`--style-prompt`) — archived 2026-11-24
**Tinyspec:** .specify/archive/tiny/2026-11-12-speak-style-prompt-flag/speak-style-prompt-flag.md
**Implemented:** 2026-11-12

`sonora speak` and `sonora warm tts-cache` gain `--style-prompt TEXT`, sent verbatim as `stylePrompt` for `google-gemini` (the hub decides), so a Gemini announcement's delivery can be set from the shell and a styled phrase warmed under the key `speak` looks up. An empty value is sent as given, not a usage error; without the flag no `stylePrompt` key is sent. This supersedes FR-193 and AS-055 for `stylePrompt` only (FR-204, FR-205, AS-061). The release tag is still to be pushed.

### Hub Info and TTS Cache Warm-up (`get info`, `warm tts-cache`) — archived 2026-11-24
**Branch:** 013-info-and-tts-warm
**Spec:** [.specify/archive/013-info-and-tts-warm/spec.md](../archive/013-info-and-tts-warm/spec.md)

**What was added:**
- User Stories 64–66 (P1–P2): `sonora get info` shows the hub's product, version, API version, start time and uptime (read-only; a hub older than 0.1.33 gets its own message); `sonora warm tts-cache <text|->` synthesizes a phrase into the TTS cache without playing it, reporting `cacheHit`; the public `hub` package offers both as typed functions
- FR-194–FR-203, SC-066–SC-070, AS-055–AS-060, five edge cases and the Hub Info, Warm Request and Warm Result entities; no existing entry folded, superseded or retired

**New Components:**
- `hub/info.go` (`GetInfo`, `Info`, `InfoNotOfferedError`); `hub/tts.go` (`WarmTTSCache`, `WarmRequest`, `WarmResult`); `hub/errors.go` (`ClassifyError` for the info 404 and `CACHE_WRITE_FAILED`)
- `internal/cli/hubinfo` (`RunGet`), `internal/cli/tts/warm.go` (`RunWarm`), `internal/render/info.go`, and the warm result in `internal/render/tts.go`
- `tests/{contract,unit,integration}` info and warm tests; the live checks on hub 0.1.33 are in the feature's `verification.md` (2026-11-09)

**Tasks Completed:** 26/27 tasks. T026 (the release tag) is open until the user asks for a release.

### Join Modes in the CLI (`--join-mode`, `--default-join-mode`, `get outputs --include-routes`) — archived 2026-11-24
**Branch:** 012-join-modes
**Spec:** [.specify/archive/012-join-modes/spec.md](../archive/012-join-modes/spec.md)

**What was added:**
- User Stories 59–61 (P1–P3): announcing over music or playing alongside it with `--join-mode` on `play` and `route`, and a default join mode on `create inputs`; seeing an output's routes, including group routes, with `get outputs/<id> --include-routes`; seeing a route's join mode and outputs, and an input's default join mode, in `get routes/<id>` and `get inputs/<id>`
- FR-161–FR-178, SC-056–SC-060, AS-046–AS-051, eleven edge cases, and the Join Mode, Route on Output and Output Main Route entities; FR-154 gained the join-mode rules (012 FR-018), AS-038 and AS-044 gained the join-mode clauses, and the Input and Route entities gained their new fields

**New Components:**
- `hub`: `CreateRouteRequest.JoinMode`, `PlaybackRequest.JoinMode` and `CreateInputRequest.DefaultJoinMode` (`*string`, `omitempty`); `RouteOnOutput` with `Lowered()`; `GetRoutesForOutput`; `JoinMode*` and `DuckState*` constants (additive; the plan targets v0.3.0)
- `internal/cli/joinmode` (the flag check shared by `play`, `route` and `create inputs`); the `--include-routes` second request and its older-hub message in `internal/cli/outputs`; `RenderRouteYAML` and `RenderInputYAML` show the new fields

**Tasks Completed:** 43/43 tasks

### Align the CLI with Hub 0.1.25–0.1.28 — archived 2026-11-23
**Branch:** 011-hub-0128-alignment
**Spec:** [.specify/archive/011-hub-0128-alignment/spec.md](../archive/011-hub-0128-alignment/spec.md)

**What was added:**
- User Stories 55–57 (P1–P3): the `enable`/`disable` help and README say what each command does to running playback (hub 0.1.25); `sonora play` shows the hub's resolved `displayName` (hub 0.1.28), and the hub client's play result carries it; `sonora transfer` retries the transient `ROUTE_CHANGING_OUTPUTS` refusal for about 1.5 s, then exits `14` with a transient message
- FR-139–FR-154, SC-050–SC-054, AS-038–AS-044, the Route Refusal entity and eight edge cases; no new command, flag or dependency

**New Components:**
- `hub.PlaybackResponse.DisplayName`, `hub.ReasonRouteChangingOutputs`, `hub.IsTransientConflict` and the transient-refusal message in `ClassifyError` (additive; next tag is a minor version)
- Retry loop in `internal/cli/route/transfer.go` (100, 200, 400, 800 ms; at most five attempts)
- `--help` description lines in `internal/cli/{inputs,outputs,groups}/enabled.go`; `displayName` rendering in `internal/render/play.go`

**Superseded:**
- Playback Result entity (no ID; from .specify/archive/006-play-command/spec.md -> "Playback Result"; .specify/archive/006-play-command/data-model.md -> "PlaybackResult (rendered view)"): "renders exactly four" fields replaced in place by four fields plus `displayName` second when non-empty, on the user's confirmation. Not a `RETIRED:` line: entities carry no ID to retire.

**Tasks Completed:** 27/27 tasks

### JSON Property Names on the Hub's TTS Cache Statistics (v0.3.4) — archived 2026-11-24
**Tinyspec:** .specify/archive/tiny/2026-10-28-tts-cache-stats-json-tags/tts-cache-stats-json-tags.md
**Implemented:** 2026-10-28

`hub.TTSCacheStats` gains JSON tags, so marshalled it uses `totalEntries`, `totalSizeBytes`, `maxSizeBytes` and `entriesByProvider`, the names of the hub's `CacheStats` schema that sonora-mcp's cache tool needs. Decoding and the CLI's output are unchanged. The v0.3.4 tag is still to be released. (FR-190, SC-064.)

### Diagnose "TTS Not Offered" in the Hub Client (`DiagnoseTTSUnavailable`, v0.3.3) — archived 2026-11-24
**Tinyspec:** .specify/archive/tiny/2026-10-28-hub-tts-unavailable-diagnosis/hub-tts-unavailable-diagnosis.md
**Implemented:** 2026-10-28

Makes the CLI's TTS 404 diagnosis public as `hub.DiagnoseTTSUnavailable`, so sonora-mcp can say why TTS is unavailable, and classifies a TTS 404 as text-to-speech unavailable instead of a network failure (released in v0.3.3). The CLI's output and exit codes are unchanged. (FR-188–FR-189, SC-063.)

### Speak Playback Flags and Request Fields (`speak --playback-mode`, `--route-volume`, `--ducking-level`) — archived 2026-11-24
**Tinyspec:** .specify/archive/tiny/2026-10-27-speak-request-fields/speak-request-fields.md
**Implemented:** 2026-10-27

`sonora speak` gains `--playback-mode`, `--route-volume` and `--ducking-level`, so an announcement can duck or mix over music at its own level, and `hub.SpeakRequest` carries the seven hub 0.1.29 fields it lacked, sent only when set. No flags for `engine`, `pitch`, `speakingRate` or `stylePrompt`. AS-048 is amended to match. The tag is still to be released. (User Story 63, FR-191–FR-193, SC-065, AS-054.)

### Route and Output Volume and Ducking Level in the CLI — archived 2026-11-24
**Tinyspec:** .specify/archive/tiny/2026-10-25-route-volume-cli/route-volume-cli.md
**Implemented:** 2026-10-25

`sonora play` replaces `--volume` with `--output-volume` and gains `--route-volume` and `--ducking-level` (which needs `--join-mode DUCK_OTHERS`); `sonora set routes/<id> volume <0-100>` changes a playing route's volume; `get routes/<id>`, `get outputs/<id> --include-routes` and `get inputs/<id>` show the levels when set. Removing `--volume` is a breaking CLI change; the tag is still to be released. (User Story 62, FR-183–FR-187, SC-062, AS-053.)

### Route Volume and Ducking Level in the Hub Client (hub 0.1.29) — archived 2026-11-24
**Tinyspec:** .specify/archive/tiny/2026-10-25-hub-route-volume-client/hub-route-volume-client.md
**Implemented:** 2026-10-25

Brings the public `hub` package to the hub 0.1.29 contract: `PlaybackRequest.Volume` becomes `OutputVolume` (`outputVolume`, a breaking change for the public package), route volume and ducking level are optional on play, route and input records, and `SetRouteVolume` calls `PUT /api/v2/routes/{routeId}/volume`. The release tag is still to be pushed. (FR-179–FR-182, SC-061, AS-052.)

### Rename an Input (`rename`) — archived 2026-11-23
**Tinyspec:** .specify/archive/tiny/2026-10-22-rename-inputs-command/rename-inputs-command.md
**Implemented:** 2026-10-22

Adds `sonora rename inputs/<id> <name>` and `sonora rename inputs/<id> --reset` (`PUT /api/v2/inputs/{id}/display-name`, hub 0.1.27), printing the input record; `hub` gains `SetInputDisplayName` for sonora-mcp. (User Story 58, FR-155–FR-160, SC-055, AS-045.)

### Decode Hub Route Refusals (409) — archived 2026-11-23
**Tinyspec:** .specify/archive/tiny/2026-10-07-decode-route-refusals-409/decode-route-refusals-409.md
**Implemented:** 2026-10-07

A hub 409 on route creation, transfer or play (disabled target, route limit, input already on the output) now prints the hub's detail and exits `14` (conflict) instead of `3`; the public client exposes `APIError.Reason`/`OutputID`, `ClassConflict`, and `Route.JoinMode`/`Outputs`. (FR-137–FR-138; amends FR-079, FR-085, FR-090.)

### Public Hub Client Package (`hub/`, `api.Spec`) — archived 2026-11-23
**Branch:** 010-public-hub-package
**Spec:** [.specify/archive/010-public-hub-package/spec.md](../archive/010-public-hub-package/spec.md)

**What was added:**
- User Stories 51–54 (P1–P2): importing the hub client from another project by its public path at a released tag; existing CLI users noticing no change; consumers verifying against the same published spec; maintainers releasing the shared client safely
- FR-122–FR-136, SC-043–SC-049, AS-032–AS-037, the Public Hub Client, Error Class, Exit-Code Mapping, Published Hub Spec and Release Tag entities, and six edge cases; no existing entry folded or retired

**New Components:**
- `hub/` (public Multiroom Audio Hub client, moved from `internal/hub` with `git mv`): protocol concerns only, self-contained godoc, standard library only
- `api/` (`api.Spec`, `go:embed` of `api/openapi.json`, one copy of the spec)
- `internal/cli/exitcode/` (`Usage`, `For`): exit-code mapping moved out of `hub`
- Module path `github.com/Sonora-Multiroom/sonora-cli`, with the `-X` version injection sites in `Makefile`, `.goreleaser.yaml` and the local `build.sh`
- Guard tests: exit-code table, `api.Spec`, godoc rules, CLI-neutral messages, stray old-path references, version injection

**Tasks Completed:** 47/50 tasks. The three open checkboxes (T045–T047) are the PR, release-tag and post-tag steps; `main` carries the merge commit `b87cf4d`.

### List TTS Voices (`list tts-voices`) — archived 2026-11-23
**Tinyspec:** .specify/archive/tiny/2026-09-27-list-tts-voices-command/list-tts-voices-command.md
**Implemented:** 2026-09-27

Adds `sonora list tts-voices --provider <name> [--language] [--engine]` (`GET /api/tts/providers/{providerName}/voices`) so users can find valid `speak --voice` values; `get tts-voices` is a usage error. (User Story 50, FR-118–FR-121, SC-042, AS-031.)

### Text-to-Speech Announcements (`speak`, `get tts-cache`, `clear tts-cache`) — archived 2026-11-23
**Branch:** 009-tts-commands
**Spec:** [.specify/archive/009-tts-commands/spec.md](../archive/009-tts-commands/spec.md)

**What was added:**
- User Stories 44–49 (P1–P4): `sonora speak <text|-> <outputs|groups>/<id>` announces text on one output or group, with `--provider`, `--voice`, `--language`, `--timeout` and standard input (`-`); `sonora get tts-cache` shows the TTS cache statistics; `sonora clear tts-cache [--provider]` clears the whole cache or one provider's entries
- Exit code `13` (TTS not available); a 404 from a TTS operation triggers one extension-inventory lookup that explains why TTS is unavailable
- FR-100–FR-117, SC-034–SC-041, AS-023–AS-030, the Announcement Request, Announcement Acceptance, TTS Cache Statistics and TTS Error entities and sixteen edge cases; shared FR-009, FR-067, AS-002 and AS-022 widened to cover the TTS commands

**New Components:**
- `internal/cli/tts/` (`RunSpeak`, `RunGetCache`, `RunClearCache`, shared failure path); `internal/hub/tts.go`, `internal/hub/extensions.go`; `internal/render/tts.go`; `internal/hub/client.go` (`NewClientWithTimeout`); `internal/hub/errors.go` (`ClassTTSUnavailable`); `cmd/sonora/main.go`: `speak` special case, `tts-cache` keyword, `clear` verb
- `tests/{unit,contract,integration}` TTS tests

**Tasks Completed:** 55/55 tasks

### Bulk Stop Routes (`stop routes`, `stop outputs/<id>`, `stop groups/<id>`) — archived 2026-11-23
**Tinyspec:** .specify/archive/tiny/2026-09-09-stop-routes-command/stop-routes-command.md
**Implemented:** 2026-09-09

Adds bulk stop: `sonora stop routes`, `sonora stop outputs/<id>` and `sonora stop groups/<id>`; `stop` is no longer a routes-only alias. (User Story 43, FR-098–FR-099.)

### Pause and Resume a Route (`pause`, `resume`) — archived 2026-11-23
**Tinyspec:** .specify/archive/tiny/2026-08-29-route-pause-resume-command/route-pause-resume-command.md
**Implemented:** 2026-08-29

Adds `sonora pause routes/<id>` and `sonora resume routes/<id>` (`PUT /api/v2/routes/{id}/pause`). (User Story 42, FR-097.)

### Set Output Volume (`set outputs/<id> volume`) — archived 2026-11-23
**Tinyspec:** .specify/archive/tiny/2026-08-29-outputs-set-volume-command/outputs-set-volume-command.md
**Implemented:** 2026-08-29

Adds the `set` verb and `sonora set outputs/<output-id> volume <0-100>` (`PUT /api/v2/outputs/{id}/volume`). (User Story 41, FR-096.)

### Mute and Unmute Outputs and Groups (`mute`, `unmute`) — archived 2026-11-23
**Tinyspec:** .specify/archive/tiny/2026-08-29-outputs-groups-mute-command/outputs-groups-mute-command.md
**Implemented:** 2026-08-29

Adds `sonora mute` / `sonora unmute` for `outputs/<id>` and `groups/<id>` (`PUT /api/v2/{outputs|groups}/{id}/mute`). (User Story 40, FR-095.)

### Enable and Disable Outputs — archived 2026-11-23
**Tinyspec:** .specify/archive/tiny/2026-08-29-outputs-enable-disable-command/outputs-enable-disable-command.md
**Implemented:** 2026-08-29

Extends `enable`/`disable` to `outputs/<id>` (`PUT /api/v2/outputs/{id}/enabled`). (User Story 39, FR-094.)

### Master Mute (`get master-mute`, `mute all`, `unmute all`) — archived 2026-11-23
**Tinyspec:** .specify/archive/tiny/2026-08-29-master-mute-command/master-mute-command.md
**Implemented:** 2026-08-29

Adds the system-wide master mute: `sonora get master-mute`, `sonora mute all` and `sonora unmute all` (`GET`/`PUT /api/v2/master-mute`). (User Story 38, FR-092–FR-093.)

### Create and Delete Inputs — archived 2026-11-23
**Tinyspec:** .specify/archive/tiny/2026-08-29-inputs-create-delete-command/inputs-create-delete-command.md
**Implemented:** 2026-08-29

Adds `sonora create inputs/<id> <uri> --display-name <name>` (`POST /api/v2/inputs`) and extends `delete` to `inputs/<id>` (`DELETE /api/v2/inputs/{id}`). (User Story 37, FR-090–FR-091.)

### Set Group Volume (`set groups/<id> volume`) — archived 2026-11-23
**Tinyspec:** .specify/archive/tiny/2026-08-29-groups-set-volume-command/groups-set-volume-command.md
**Implemented:** 2026-08-29

Adds `sonora set groups/<group-id> volume <0-100>` (`PUT /api/v2/groups/{id}/volume`). (User Story 36, FR-089.)

### Enable and Disable Output Groups — archived 2026-11-23
**Tinyspec:** .specify/archive/tiny/2026-08-29-groups-enable-disable-command/groups-enable-disable-command.md
**Implemented:** 2026-08-29

Extends `enable`/`disable` to `groups/<id>` (`PUT /api/v2/groups/{id}/enabled`). (User Story 35, FR-088.)

### Transfer a Route (`transfer`) — archived 2026-11-23
**Tinyspec:** .specify/archive/tiny/2026-08-28-route-transfer-command/route-transfer-command.md
**Implemented:** 2026-08-28

Adds `sonora transfer routes/<id> <outputs|groups>/<target-id>` (`POST /api/v2/routes/{id}/transfer`), moving a playing route to a new target and printing the new route. (User Story 34, FR-085–FR-087.)

### Delete or Stop a Route (`delete routes`, `stop routes/<id>`) — archived 2026-11-23
**Tinyspec:** .specify/archive/tiny/2026-08-28-route-delete-command/route-delete-command.md
**Implemented:** 2026-08-28

Adds `sonora delete routes/<id>` and its exact alias `sonora stop routes/<id>` (`DELETE /api/v2/routes/{id}`), stopping playback and removing the route. (User Story 33, FR-083–FR-084.)

### Enable and Disable Inputs (`enable`, `disable`) — archived 2026-11-23
**Tinyspec:** .specify/archive/tiny/2026-08-28-inputs-enable-disable-command/inputs-enable-disable-command.md
**Implemented:** 2026-08-28

Adds `sonora enable inputs/<id>` and `sonora disable inputs/<id>` (`PUT /api/v2/inputs/{id}/enabled`), printing the updated input record. (User Story 32, FR-080–FR-082.)

### Route an Existing Input (`route`) — archived 2026-11-23
**Branch:** 008-route-command
**Spec:** [.specify/archive/008-route-command/spec.md](../archive/008-route-command/spec.md)

**What was added:**
- User Stories 30–31 (P1–P2): `sonora route inputs/<input-id> <outputs|groups>/<target-id>` connects an existing input to one output or one output group in one command, without creating a new input; the target type comes from the path prefix, so an output and a group sharing an identifier need no flag
- Exit codes `11` (input not found) and `12` (target not found); route creation failed reuses `8`
- FR-071–FR-079, SC-030–SC-033, AS-018–AS-021, six route edge cases and the Routing Request and Routing Result entities; shared FR-006, FR-007, FR-009, FR-011 and FR-013 widened to cover the route command, and FR-067 to cover its target path

**New Components:**
- `internal/cli/route/` (`Run`); `internal/hub/routes.go` (`CreateRouteRequest`, `CreateRoute`); `internal/hub/errors.go` (`ClassInputNotFound`, `ClassTargetNotFound`); `internal/render/route.go`; `cmd/sonora/main.go`: early `route` special case beside `play`
- `tests/{unit,contract,integration}` route tests

**Tasks Completed:** 17/17 tasks

### Adopt Verb-First Command Landscape — archived 2026-11-23
**Branch:** 007-refactor-cli-commands
**Spec:** [.specify/archive/007-refactor-cli-commands/spec.md](../archive/007-refactor-cli-commands/spec.md)

**What was added:**
- User Stories 27–29 (P1–P3): `sonora get <resource>[/<id>]` and `sonora list <resource>` replace the `sonora <resource> list|get` commands for inputs, outputs, groups and routes; aliases `in`/`out`/`gr`/`rt`; `play`'s target is now `outputs/<id>` or `groups/<id>` and its name flag is `--display-name`
- Hard cutover: the old command forms and `play`'s `--group`/`--output`/`--name` fail with a standard usage error
- FR-059–FR-070, SC-026–SC-029, AS-014–AS-017, six edge cases and the Resource Kind and Resource Path entities
- Exit code `7` (ambiguous target) is gone, since a path prefix makes the target type unambiguous

**New Components:**
- `internal/cli/respath/` (alias table, `resource/id` split, id pattern); `cmd/sonora/main.go` verb-first dispatcher (`get`, `list`, `play`); `internal/cli/play` takes a path target and `--display-name`
- `internal/hub/play.go`: `ResolveTarget` replaced by a per-type existence check; `internal/hub/errors.go`: `ClassAmbiguous` and `AmbiguousTargetError` removed
- `internal/cli/clihelp/usage.go`: help text for the new grammar; `tests/{unit,contract,integration}` updated for the new invocation shape

**Superseded:**
- RETIRED: FR-015 (from .specify/archive/002-outputs-get/spec.md -> FR-002) → replaced by FR-060. Reason: `get outputs` with no id is now the collection form, not a usage error.
- RETIRED: FR-023 (from .specify/archive/003-inputs-list-get/spec.md -> FR-006) → replaced by FR-060. Reason: `get inputs` with no id is now the collection form, not a usage error.
- RETIRED: FR-033 (from .specify/archive/004-routes-list-get/spec.md -> FR-006) → replaced by FR-060. Reason: `get routes` with no id is now the collection form, not a usage error.
- RETIRED: FR-042 (from .specify/archive/005-groups-list-get/spec.md -> FR-006) → replaced by FR-060. Reason: `get groups` with no id is now the collection form, not a usage error.
- RETIRED: FR-049 (from .specify/archive/006-play-command/spec.md -> FR-002a) → no replacement. Reason: the `--group` and `--output` flags no longer exist.
- RETIRED: FR-050 (from .specify/archive/006-play-command/spec.md -> FR-003) → replaced by FR-067. Reason: the target type comes from the path prefix, not from auto-detection or forcing flags.
- RETIRED: FR-051 (from .specify/archive/006-play-command/spec.md -> FR-003a) → no replacement. Reason: path-style targets make an ambiguous output/group identifier impossible.
- RETIRED: FR-052 (from .specify/archive/006-play-command/spec.md -> FR-003b) → replaced by FR-067. Reason: the path prefix fixes the type to look up; a missing id of that type is plain "not found".
- RETIRED: AS-012 (from .specify/archive/006-play-command/spec.md -> "Output identifiers and group identifiers may occasionally collide") → no replacement. Reason: a path prefix breaks any output/group identifier tie, so no flags are needed.

**Tasks Completed:** 32/32 tasks

### Help Command (`help`) — archived 2026-11-23
**Tinyspec:** .specify/archive/tiny/2026-08-26-help-command/help-command.md
**Implemented:** 2026-08-26

Adds `sonora help`, `-h`/`--help` and bare `sonora`, which print the command table, global flags and examples to stdout with exit `0` (User Story 26, FR-057–FR-058; `cmd/sonora/main.go`, `cmd/sonora/main_test.go`). The missing-verb and unknown-command usage errors are unchanged.

### Instant Playback (`play`) — archived 2026-11-23
**Branch:** 006-play-command
**Spec:** [.specify/archive/006-play-command/spec.md](../../.specify/archive/006-play-command/spec.md)

**What was added:**
- User Stories 22–25 (P1–P3): `sonora play <uri> <target-id>` starts playback of a URI to one output or a group in one command and reports the created input identifier, route identifier and status as the hub returned them; `--volume` sets a starting volume (0–100), `--name` names the ephemeral input
- `--group`/`--output` force the target type when an output and a group share an identifier; both together is a usage error
- FR-047–FR-056, SC-022–SC-025, AS-010–AS-013 and the Playback Request and Playback Result entities; shared requirements, outcomes and assumptions now cover the play command as well (FR-006–FR-009, FR-011, FR-013, SC-003, SC-004, AS-001–AS-004)
- Five new exit codes (6–10) for validation, ambiguous target, route creation failed, source unreachable and upstream service unavailable; "target not found" reuses exit `5`

**New Components:**
- `internal/cli/play/` (`Run`); `internal/hub/play.go` (`PlaybackRequest`, `PlaybackResponse`, `ResolveTarget`, `Playback`); `internal/render/play.go`
- `internal/hub/errors.go`: five new classes, `APIError`, `AmbiguousTargetError`; `cmd/sonora/main.go`: early `play` special case
- `tests/{unit,contract,integration}` play tests

**Tasks Completed:** 24/24 tasks


### List and Get Output Groups — archived 2026-11-23
**Branch:** 005-groups-list-get
**Spec:** [.specify/archive/005-groups-list-get/spec.md](../archive/005-groups-list-get/spec.md)

**What was added:**
- User Stories 17–21 (P1–P3): `sonora groups list` shows enabled output groups (identifier, display name, member outputs, mute and enabled state), `--include-disabled` shows disabled ones, `--json` for scripts
- `sonora groups get <group-id>` shows one group, enabled or disabled, with a clear "group not found" exit `5`
- FR-037–FR-046, SC-017–SC-021, AS-009 and the Output Group entity; shared outputs/inputs/routes requirements, outcomes and assumptions now cover groups as well (FR-006–FR-011, FR-013, SC-002–SC-004, AS-001–AS-006)

**New Components:**
- `internal/cli/groups/` (`RunList`, `RunGet`); `internal/hub/groups.go` (`ListGroups`, `GetGroup`); `internal/render/groups.go`
- `tests/{unit,contract,integration}` groups tests

**Tasks Completed:** 33/33 tasks


### List and Get Audio Routes — archived 2026-11-23
**Branch:** 004-routes-list-get
**Spec:** [.specify/archive/004-routes-list-get/spec.md](../archive/004-routes-list-get/spec.md)

**What was added:**
- User Stories 12–16 (P1–P3): `sonora routes list` shows all routes (identifier, source input, target, target type, status), `--status`/`--input-id`/`--target-id` narrow it with AND logic, `--json` for scripts
- `sonora routes get <route-id>` shows one route's full state, with a clear "route not found" exit `5`
- FR-028–FR-036, SC-012–SC-016, AS-008 and the Route entity; shared outputs/inputs/routes requirements widened (FR-006–FR-011, FR-013, SC-002–SC-004, AS-001–AS-006)

**New Components:**
- `internal/cli/routes/` (`RunList`, `RunGet`); `internal/hub/routes.go` (`ListRoutes`, `GetRoute`); `internal/render/routes.go`
- `tests/{unit,contract,integration}` routes tests

**Tasks Completed:** 33/33 tasks


### List and Get Audio Inputs — archived 2026-11-23
**Branch:** 003-inputs-list-get
**Spec:** [.specify/archive/003-inputs-list-get/spec.md](../archive/003-inputs-list-get/spec.md)

**What was added:**
- User Stories 7–11 (P1–P3): `sonora inputs list` shows enabled inputs, `--include-disabled` shows disabled ones, `--json` for scripts
- `sonora inputs get <input-id>` shows one input, enabled or disabled, with a clear "input not found" exit `5`
- FR-018–FR-027, SC-008–SC-011, AS-007 and the Input entity; shared outputs/inputs requirements widened (FR-006–FR-011, FR-013, SC-002–SC-004)

**New Components:**
- `internal/cli/inputs/` (`RunList`, `RunGet`); `internal/hub/inputs.go` (`ListInputs`, `GetInput`); `internal/render/inputs.go`
- `hub.NotFoundError` generalized to `Resource`/`ID` (output message text unchanged)
- `tests/{unit,contract,integration}` inputs tests

**Tasks Completed:** 36/36 tasks


### Get Single Audio Output — archived 2026-11-23
**Branch:** 002-outputs-get
**Spec:** [.specify/archive/002-outputs-get/spec.md](../archive/002-outputs-get/spec.md)

**What was added:**
- User Story 4 (P1): `sonora outputs get <output-id>` shows one output, enabled or disabled
- User Story 5 (P2): clear "output not found" message with its own exit code (`5`)
- User Story 6 (P3): `--json` strict machine-readable output for a single output

**New Components:**
- `internal/cli/outputs/get.go` (`RunGet`); `outputs.Run` renamed `RunList`
- `hub.GetOutput`, `NotFoundError` / `ClassNotFound`, single-record `RenderOutputYAML` / `RenderOutputJSON`
- `tests/{unit,contract,integration}` get tests

**Tasks Completed:** 26/26 tasks

### List Audio Outputs — archived 2026-11-23
**Branch:** 001-list-outputs
**Spec:** [.specify/archive/001-list-outputs/spec.md](../archive/001-list-outputs/spec.md)

**What was added:**
- User Story 1 (P1): list enabled outputs with identifier, display name, volume, mute and availability
- User Story 2 (P2): `--include-disabled` shows disabled outputs
- User Story 3 (P3): `--json` strict machine-readable output
- Fail-fast error handling with distinguishable exit codes (0/2/3/4), 5s timeout, no retries

**New Components:**
- `cmd/sonora`, `internal/hub`, `internal/config`, `internal/cli/outputs`, `internal/render`
- `tests/{unit,contract,integration}`

**Tasks Completed:** 33/33 tasks
