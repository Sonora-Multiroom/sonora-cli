# Feature Specification: Align the CLI with Hub 0.1.25–0.1.28

**Feature Branch**: `011-hub-0128-alignment`

**Created**: 2026-10-22

**Status**: Draft

**Input**: User description: "docs/backlog/enable-disable-stops-playback-docs.md,
docs/backlog/play-shows-resolved-name.md, docs/backlog/transient-route-changing-outputs.md"

## Context

`api/openapi.json` was refreshed from hub 0.1.24 to 0.1.28 on 2026-10-22. Three changes in that range
affect what this CLI already does, and the CLI has not caught up with them:

- **Hub 0.1.25** (multiroom-ai 027) changed what enabling and disabling does to running playback.
  Disabling an input or a group now stops routes; enabling an output puts it back into running group
  routes. The CLI's documentation still says a disable only blocks new routes.
- **Hub 0.1.25** also added a transient refusal of a transfer, reason `ROUTE_CHANGING_OUTPUTS`, while
  a route is still adding an output that was just enabled. It clears by itself within moments (0.16 s
  measured on the production hub), but the CLI reports it like any other refusal and tells the user to
  change the hub's state.
- **Hub 0.1.28** (multiroom-ai 031) names a SoundCloud or YouTube link played without a name after the
  track it resolves to, and returns the input's final name in the play response. The hub client in this
  module does not read that field, and `sonora play` does not show it.

The users are people running `sonora` interactively or from scripts, and developers of other Sonora
projects (sonora-mcp) that use this module's hub client package.

This feature does not add commands. Renaming inputs (`docs/backlog/rename-inputs.md`) and join modes
(`docs/backlog/join-modes-in-cli.md`) are separate items.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Know that a disable stops playback before running it (Priority: P1)

A user wants to stop new routes from using the kitchen radio input and reads the help for
`sonora disable` (or the README) first. The documentation tells them that disabling an input stops every
route playing it right now, so they know the music will stop.

**Why this priority**: The current documentation is wrong in a way that has an audible consequence: a
user following it stops music they meant to keep. It is also the smallest change.

**Independent Test**: Read `sonora enable --help`, `sonora disable --help` and the README's
enable/disable section, and compare each statement with the hub 0.1.28 contract descriptions of
`setInputEnabled`, `setOutputEnabled` and `setGroupEnabled`.

**Acceptance Scenarios**:

1. **Given** the README, **When** a user reads the enable/disable section, **Then** it states, for each
   of inputs, outputs and groups, what disabling does to routes that are playing, and what enabling does.
2. **Given** `sonora disable inputs/<id> --help` (and the same for outputs and groups), **When** a user
   reads it, **Then** it says in one line what happens to playback that is running on that resource.
3. **Given** `sonora enable outputs/<id> --help`, **When** a user reads it, **Then** it says the output
   rejoins the group routes still playing on its groups, and that a route stopped by the disable does
   not come back.
4. **Given** the documentation of either command, **When** compared with the CLI before this feature,
   **Then** the command's arguments, flags, output and exit codes are unchanged.

---

### User Story 2 - See what `sonora play` started (Priority: P2)

A user plays a SoundCloud link without giving a name. The CLI's output shows the name the hub gave the
input ("Flickermood — Forss"), so the user can tell what is playing and find it later in
`sonora list inputs`. A script reading `--json` output gets the same name in a field.

A developer of sonora-mcp, using this module's hub client, reads the same name from the play result.

**Why this priority**: The hub client no longer mirrors the play response field for field, which this
project treats as a defect, and both the CLI and sonora-mcp lose information the hub provides. Nothing
fails, so it ranks below the misleading documentation.

**Independent Test**: Run `sonora play` against a fake hub that returns a display name, and against one
that omits it, and check the YAML and JSON output and the hub client's result in each case.

**Acceptance Scenarios**:

1. **Given** a hub that answers a play request with a display name, **When** the user runs
   `sonora play <uri> outputs/<id>`, **Then** the output shows that display name alongside the input ID,
   route ID, status and message, in YAML and in JSON.
2. **Given** a hub older than 0.1.28, which does not return a display name, **When** the user runs the
   same command, **Then** the output has no display name field at all (not an empty one), and is
   otherwise identical to the output before this feature.
3. **Given** a developer calling the hub client's play operation, **When** the hub returns a display
   name, **Then** the result carries it; **When** the hub omits it, **Then** the result's display name is
   empty and the call still succeeds.
4. **Given** `sonora play --help` and the README, **When** a user reads about `--display-name`, **Then**
   they learn it is optional, that without it a SoundCloud or YouTube link is named after its track and
   other sources get a generated name, and that a name they give is kept.

---

### User Story 3 - A transfer right after enabling an output just works (Priority: P3)

A user re-enables the office speaker, which rejoins the living-room group route, and immediately moves
that route to another group with `sonora transfer`. The hub briefly refuses the transfer because the
route is still adding the office speaker. The CLI waits a moment and tries again, and the transfer
succeeds without the user noticing. If the hub keeps refusing, the CLI says the route was still adding an
output and to try again shortly, instead of telling the user to change the hub's state.

**Why this priority**: It only happens in a narrow window (a transfer within a fraction of a second of an
enable), most likely from scripts. When it happens today, the user gets wrong advice, but a manual retry
works.

**Independent Test**: Run `sonora transfer` against a fake hub that refuses the first N transfer
requests with reason `ROUTE_CHANGING_OUTPUTS` and then accepts, for N within and beyond the retry
allowance, and check output, stderr and exit code.

**Acceptance Scenarios**:

1. **Given** a hub that refuses a transfer once with reason `ROUTE_CHANGING_OUTPUTS` and then accepts
   it, **When** the user runs `sonora transfer routes/<id> groups/<id>`, **Then** the command succeeds
   with the same output and exit code as a transfer accepted at once, and prints nothing extra.
2. **Given** a hub that keeps refusing with that reason, **When** the user runs the transfer, **Then**
   the CLI stops retrying within about 2 seconds of the first refusal, exits with the conflict exit code
   (14), and its error message says the route is still adding an output and to retry shortly.
3. **Given** a hub that refuses a transfer with any other reason (for example `OUTPUT_DISABLED`), or
   with no reason, **When** the user runs the transfer, **Then** the CLI does not retry and reports the
   refusal exactly as before this feature.
4. **Given** `--verbose`, **When** the retries run out, **Then** the underlying hub error is shown as it
   is for any other conflict.
5. **Given** the README's exit-code table, **When** a user reads the row for 14, **Then** it mentions
   that one kind of refusal clears by itself and the CLI retries it during a transfer.
6. **Given** a developer reading the hub client's documentation of refusal reasons, **When** they look up
   the reasons and which of them name an output, **Then** `ROUTE_CHANGING_OUTPUTS` is listed, described
   as transient, and listed among the reasons that name no output.

---

### Edge Cases

- **Display name with characters outside ASCII** (Cyrillic titles, an em dash): shown as written in
  YAML and JSON, not escaped into something unreadable.
- **Display name present but empty**: treated like an absent one; no field is shown.
- **The user gave `--display-name`**: the hub returns that name; the CLI shows what the hub returns, not
  what the user sent.
- **Transfer refused with `ROUTE_CHANGING_OUTPUTS`, then with a different reason on a retry**: the CLI
  stops retrying and reports the second refusal as it would any other.
- **Transfer refused with `ROUTE_CHANGING_OUTPUTS`, then the route is gone (404) on a retry**: reported
  as the route not being found, as today.
- **Network failure or timeout during a retry**: reported as a network failure, as today; the CLI does
  not retry network failures.
- **`ROUTE_CHANGING_OUTPUTS` on a command other than transfer**: the contract documents it only for a
  transfer. If another command receives it, the CLI does not retry, but the error message still says the
  refusal is transient.
- **Unknown future reasons**: still reported as a generic conflict, exit 14, with no retry.

## Requirements *(mandatory)*

### Functional Requirements

**Enable and disable documentation**

- **FR-001**: The README MUST state what each enable and disable does to playback, as the hub 0.1.25
  contract describes it: disabling an input stops every route playing it (paused and shared ones too)
  and removes an input registered to remove itself; disabling a group stops the routes addressed to the
  group while routes addressed to a member keep playing; disabling an output takes it out of every route
  playing on it; enabling an output puts it back into the live group routes of its enabled groups under
  the same route ID; enabling an input or a group starts nothing.
- **FR-002**: The help of `sonora enable` and `sonora disable`, for each of `inputs`, `outputs` and
  `groups`, MUST include one line on what happens to running playback.
- **FR-003**: The enable and disable commands' arguments, flags, output and exit codes MUST NOT change.

**Play shows the resolved display name**

- **FR-004**: The hub client's play result MUST carry the display name the hub returns, and MUST still
  succeed, with an empty name, when the hub omits it.
- **FR-005**: `sonora play` MUST show the display name in YAML and JSON output, directly after the input
  ID, when the hub returns a non-empty one.
- **FR-006**: When the hub returns no display name, or an empty one, `sonora play` output MUST NOT
  contain a display name field and MUST otherwise be identical to its output before this feature.
- **FR-007**: The help of `sonora play` and the README MUST describe `--display-name` as optional and
  say what the hub names an input when it is left out.

**Transient transfer refusal**

- **FR-008**: When the hub refuses a transfer with reason `ROUTE_CHANGING_OUTPUTS`, `sonora transfer`
  MUST retry the same transfer automatically, with a short delay between attempts, for no more than about
  2 seconds after the first refusal.
- **FR-009**: A retry that is accepted MUST produce exactly the output and exit code of a transfer
  accepted on the first attempt, with nothing extra on stdout or stderr.
- **FR-010**: When the retries run out, the CLI MUST exit with the conflict exit code (14) and print a
  message saying the route is still adding an output and the user should retry shortly.
- **FR-011**: The CLI MUST NOT retry any other refusal reason, a refusal without a reason, a 404, any
  other hub error or a network failure; each MUST be reported exactly as before this feature.
- **FR-012**: Wherever else the CLI reports a refusal with reason `ROUTE_CHANGING_OUTPUTS`, its message
  MUST say the refusal is transient rather than asking the user to change the hub's state.
- **FR-013**: The hub client's documentation of refusal reasons MUST list `ROUTE_CHANGING_OUTPUTS`,
  describe it as transient, and include it among the reasons that name no output. The reason MUST stay an
  open string, so that unknown reasons keep working without a client release.
- **FR-014**: The README's exit-code row for 14 MUST mention the transient refusal and that a transfer
  retries it.

**General**

- **FR-015**: Every changed behaviour MUST be covered by tests written before the change, against a fake
  hub whose responses follow `api/openapi.json` at 0.1.28.
- **FR-016**: No new third-party dependency, command or flag MAY be added, and nothing MAY be added to
  the CLI's startup path.

### Key Entities

- **Play result**: what the hub answers to a play request: the input ID, the route, a message and, since
  0.1.28, the input's display name once the route exists.
- **Route refusal**: a 409 answer to a route request, with a human-readable detail, an open-ended reason
  and, for output-level reasons only, the output that failed. `ROUTE_CHANGING_OUTPUTS` is the one reason
  that clears without the user doing anything.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Every statement about enabling and disabling in the README and in the six enable/disable
  help texts agrees with the hub 0.1.28 contract; a review finds zero contradictions.
- **SC-002**: For a play of a link that the hub names, the user can read the input's name from the
  command's output on the first try, without a second command.
- **SC-003**: A transfer refused once as transient completes successfully with no user action, and the
  whole command takes less than 1 second longer than an immediate success.
- **SC-004**: A transfer that stays refused as transient ends within 3 seconds of being started, with
  exit code 14 and a message that names the transient cause.
- **SC-005**: The output of every command other than `play` and `transfer`, and of `play` against a hub
  that omits the display name, is byte-for-byte identical to the previous release in the existing test
  suite.

## Assumptions

- The production hub runs 0.1.28 (checked 2026-10-22); older hubs must still work, so every new response
  field is optional on the client side.
- A retry window of about 2 seconds is generous: the hub's own measurement of a rejoin was 0.16 s, and
  the contract says the refusal clears "within moments". The exact delays are a planning decision.
- Retrying belongs to the CLI command, not to the hub client's transfer operation, so each client
  operation keeps sending exactly one request and other consumers choose their own policy. The client
  may offer a way to recognise the transient reason.
- Typed constants for the refusal reasons are not required. The earlier decision (tinyspec
  `decode-route-refusals-409`) to keep the reason an open string stands; named values may be added only
  if they keep it open.
- `sonora transfer` already checks that the target exists before transferring; that check is not
  repeated on retries.
- No prompt, warning or `--force` flag is added to `sonora disable`; the backlog item leaves that as an
  open question for later.
- sonora-mcp picks up the client changes with its next update of this module; changing sonora-mcp is out
  of scope.
