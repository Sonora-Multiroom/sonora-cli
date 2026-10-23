# Feature Specification: Join Modes in the CLI

**Feature Branch**: `012-join-modes`

**Created**: 2026-10-22

**Status**: Draft

**Input**: User description: "docs/backlog/join-modes-in-cli.md"

## Context

Since hub 0.1.21 an output can play several routes at once. A route joins its target with a join mode:

- `REPLACE` stops the routes already playing there, except ducking ones.
- `MIX` plays alongside them.
- `DUCK_OTHERS` plays alongside them and lowers them while it plays.

`api/openapi.json` (refreshed to hub 0.1.28 on 2026-10-22) describes three things the CLI does not use:

- an operation that lists every route playing on one output, in join order, with each route's join mode,
  duck state (`FULL` or `LOWERED`) and whether it is the output's main route;
- an optional join mode on a route request and on a play request, and an optional default join mode on
  a create-input request;
- the join mode and the outputs a route plays on, in the route record, and the default join mode, in
  the input record.

Without a join mode the hub uses the input's default join mode for a route request and `REPLACE` for a
play request. So today `sonora route` and `sonora play` replace whatever is playing, unless the input has
a different default. Announcing over music, or playing a second stream alongside, is impossible from
the CLI. Nor can the CLI show what an output is playing: `sonora list routes --target-id <id>` filters
on what a route was addressed to, so it misses group routes that play on the output.

The hub client already reads the route's join mode and outputs (#21) and the input's default join mode
(#22). `get routes/<id> --json` and `get inputs/<id> --json` already include them, because those
commands print the client's record unchanged. The YAML output, which is the default, does not show
them.

The users are people running `sonora` interactively or from scripts, and developers of sonora-mcp, which
uses this module's hub client.

How to show an output's routes was decided in the backlog item on 2026-10-22: an opt-in
`--include-routes` flag on `sonora get outputs/<id>`. The alternatives considered and rejected there
(`--routes`, `list routes --playing-on`, `list routes --output`, `get outputs/<id>/routes`, always
including the routes) are not reopened here.

## Clarifications

### Session 2026-10-22

- Q: Should `list routes` and `list inputs` also show the join mode, or only the single-item views? → A:
  Only the single-item views (`get routes/<id>`, `get inputs/<id>`, `get outputs/<id> --include-routes`);
  both lists stay unchanged.
- Q: Should `--join-mode` also accept the hyphenated spelling `duck-others` used by sonora-tts? → A: No;
  only the hub's spelling, in any case. `duck-others` is a usage error that lists the valid modes.
- Q: With `--include-routes`, should each route show every field or a compact set? → A: YAML shows a
  compact set (`routeId`, `inputId`, `targetId`, `targetType`, `status`, `joinMode`, `duckState`, `main`,
  `paused`); JSON shows every field.

### Session 2026-10-23

- Q: What should `--include-routes` do when used with anything other than `get outputs/<id>`? → A: Usage
  error (exit 2), no request sent, and the message says the flag applies only to `get outputs/<id>`.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Announce over music, or play alongside it (Priority: P1)

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

---

### User Story 2 - See what an output is playing (Priority: P2)

The user hears two things in the kitchen and wants to know why. `sonora get outputs/kitchen
--include-routes` shows the kitchen's usual record and below it the list of routes playing on it. The list
is in the order the routes joined and includes group routes. For each route it shows the join mode,
whether the route is lowered, and which one is the output's main route. The user can then stop or move
the right route.

A sonora-mcp developer gets the same list from the hub client.

**Why this priority**: It is the only way to see the effect of User Story 1 and to find group routes
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

---

### User Story 3 - See a route's join mode and outputs, and an input's default (Priority: P3)

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

---

### Edge Cases

- **Mixed case**: `--join-mode duck_others`, `Duck_Others` and `DUCK_OTHERS` all mean the same mode and
  are sent upper-cased, as `--status` on `list routes` does today. The hyphenated spelling that
  sonora-tts uses (`duck-others`) is not an alias: it is a usage error that lists the valid modes.
- **Empty value** (`--join-mode ""`): a usage error, not "use the default". Leaving the flag out is the
  way to use the default.
- **A join mode the hub refuses**: a hub older than 0.1.21 ignores or rejects the field. Whatever it
  answers is reported as any other hub answer is today. The CLI does not check the hub's version first.
- **Join mode on a group target**: allowed; the hub applies it on every member output.
- **`--join-mode` with `--volume` and `--display-name` on `play`**: all three are sent together.
- **The output disappears between the two requests of `--include-routes`**: the routes request answers
  404, and the CLI reports it as an older hub. This window is narrow, and the message tells the user what
  to check. It is accepted, not handled.
- **The routes request fails for another reason** (network failure, 5xx, malformed body): reported as
  that failure, with its usual exit code. Nothing is printed on stdout, not even the output's record.
- **A hub older than 0.1.21 for `get routes/<id>`**: the route has no join mode and no outputs. The
  YAML leaves those two fields out instead of showing empty values. The JSON keeps printing what it
  prints today.
- **`--include-routes` on another resource or on a list** (`get outputs`, `get groups/<id>`,
  `get inputs/<id>`): a usage error before any request is sent.
- **Many routes on one output**: shown in full, in the hub's order. There is no paging.
- **Unknown future join mode in a response**: shown as the hub sent it. The flags still accept only the
  three documented modes, until the contract adds more.

## Requirements *(mandatory)*

### Functional Requirements

**Choosing the join mode**

- **FR-001**: The hub client's route request, play request and create-input request MUST each have an
  optional join-mode field (a default join mode for the create-input request). The field is left out
  of the request body when it is unset, and sent as given when it is set.
- **FR-002**: `sonora play` and `sonora route` MUST accept `--join-mode <mode>`, and
  `sonora create inputs/<id>` MUST accept `--default-join-mode <mode>`. The mode is `REPLACE`, `MIX` or
  `DUCK_OTHERS`, matched without regard to case and sent upper-cased.
- **FR-003**: An unknown or empty mode MUST be a usage error (exit 2). The command MUST send no request,
  and MUST print an error that lists the valid modes.
- **FR-004**: Without the flag, each command MUST send the same request it sends today, with no
  join-mode field, so that the hub's defaults apply.
- **FR-005**: The `--help` of each of the three commands MUST list the flag and its modes, and say what
  happens without it: a route uses the input's default join mode, falling back to `REPLACE`; a play
  uses `REPLACE`; an input without a default leaves its routes on `REPLACE`.

**Showing an output's routes**

- **FR-006**: The hub client MUST offer an operation that returns the routes playing on one output, in
  the hub's order, with every field of the contract's route-on-output record. A 404 from it MUST be the
  client's usual not-found error, naming the output.
- **FR-007**: The hub client MUST let a consumer tell whether a route on an output is lowered. Any duck
  state other than `LOWERED` counts as not lowered.
- **FR-008**: `sonora get outputs/<id>` MUST accept `--include-routes`, with the help text "also list
  the routes playing on this output, in join order". With it, the output MUST be the output's record
  followed by a `routes` list in the hub's order. In YAML each entry MUST hold `routeId`, `inputId`,
  `targetId`, `targetType`, `status`, `joinMode`, `duckState`, `main` and `paused`, in that order; in JSON
  each entry MUST hold every field of the contract's route-on-output record. An output with no routes
  MUST show an empty list.
- **FR-008a**: `--include-routes` on any other `get` form (a list such as `get outputs`, or a resource
  other than `outputs/<id>`) MUST be a usage error (exit 2). The command MUST send no request, and the
  error MUST say the flag applies only to `get outputs/<id>`.
- **FR-009**: Without `--include-routes`, `sonora get outputs/<id>` MUST send exactly one request, and
  its output MUST be unchanged.
- **FR-010**: When the output request succeeds and the routes request answers 404 or 405 (a hub with
  the stop-routes operation on the same path but not the list answers 405), the CLI MUST report
  that the hub does not support listing an output's routes and needs 0.1.21 or later. It MUST exit with
  the unexpected-response exit code (3), and MUST NOT report the output as not found.
- **FR-011**: With `--include-routes`, any failure MUST leave stdout empty. The CLI never prints the
  output's record without its routes.

**Showing join modes on routes and inputs**

- **FR-012**: The YAML of `sonora get routes/<id>` MUST show the route's join mode and the list of outputs
  it plays on, after its current fields. When the hub returns neither field (a hub older than 0.1.21),
  both MUST be left out.
- **FR-013**: The YAML of every command that prints one input record (`get`, `create`, `rename`,
  `enable` and `disable inputs/<id>`) MUST show the input's default join mode, as null when the input
  has none. Their JSON already shows it.
- **FR-014**: The JSON of `get routes/<id>` and `get inputs/<id>` MUST keep its current shape and contain
  the same join-mode fields as the YAML.
- **FR-015**: `sonora list routes` and `sonora list inputs` output MUST NOT change.

**General**

- **FR-016**: The README MUST describe the three join modes in a short paragraph and document
  `--join-mode`, `--default-join-mode` and `--include-routes`.
- **FR-017**: Every new behaviour MUST be covered by tests written before the code, against a fake hub
  whose answers follow `api/openapi.json` at 0.1.28. That includes contract tests for the new hub
  client operation: a list, an empty list, a 404, and an unknown duck state. It also includes contract
  tests showing that each new request field is sent only when it is set.
- **FR-018**: No new third-party dependency MAY be added, and nothing MAY be added to the CLI's startup
  path. Only `--include-routes` adds a second request, and only when it is given.

### Key Entities

- **Join mode**: how a route joins the outputs it targets: `REPLACE`, `MIX` or `DUCK_OTHERS`. It is
  set on a route or play request, or on an input as the default for its routes, and it is reported on
  each route.
- **Route on an output**: one route as seen from one output. It has the route's own fields, plus its
  duck state on that output (`FULL` or `LOWERED`, open to more values) and whether it is the output's
  main route. A group route appears once under each member output it plays on.
- **Output's main route**: the first route to join the output that is not a ducking route, or, if every
  route is ducking, the first route.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A user can play an announcement over running music without stopping it, in one command.
- **SC-002**: A user can find out everything playing on an output, including group routes, in one
  command, without cross-referencing `list routes` and `get groups`.
- **SC-003**: Against a hub older than 0.1.21, `get outputs/<id> --include-routes` never says that an
  existing output was not found. Its message names the hub version required.
- **SC-004**: Every command run without the new flags sends the same requests and prints the same
  output as the previous release, byte for byte, across the existing test suite, apart from the
  join-mode fields FR-012 and FR-013 add to the route and input YAML.
- **SC-005**: An invalid join mode is rejected before any request reaches the hub, in 100% of the tested
  cases.

## Assumptions

- The production hub runs 0.1.28. Hubs older than 0.1.21 must still work for every command that does
  not use the new flags.
- **List views stay as they are** (the backlog's open question, settled in Clarifications). `list routes`
  and `list inputs` keep their fixed fields. The join mode appears in `get routes/<id>`,
  `get inputs/<id>` and in the routes shown by `--include-routes`, which is where a user asks about one
  route or one output. Adding columns to the lists would change their shape for every script that reads
  them, for little gain.
- `sonora play` and `sonora route` do not show the join mode in their own output. Their output stays as
  it is, and `get routes/<id>` shows the join mode afterwards.
- sonora-tts's announcement playback mode is out of scope until sonora-tts documents it.
- The flag check mirrors `--status` on `list routes`: the valid values are a fixed list in the CLI, taken
  from the contract's enum.
- The exit code for an older hub on `--include-routes` (3) follows the README's meaning "hub returned an
  unexpected response". The hub answered 404 to a request for an output it had just returned.
- sonora-mcp picks up the client changes with its next update of this module; changing sonora-mcp is
  out of scope.
- If the hub later adds `?includeRoutes=true` to the output request (proposed in
  `sonora-multiroom/docs/backlog/routes-per-output-in-outputs-list.md`), the flag can switch to one request
  without changing what users see. That switch is not part of this feature.
