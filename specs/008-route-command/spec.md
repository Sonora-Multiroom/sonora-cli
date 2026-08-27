# Feature Specification: Route an Existing Input (`route`)

**Feature Branch**: `008-route-command`

**Created**: 2026-08-27

**Status**: Draft

**Input**: User description: "`sonora route inputs/<input-id> <outputs|groups>/<target-id>`
(singular, top-level, no subcommand, path-style resource addressing) — mirrors `sonora play
<uri> <outputs|groups>/<target-id>` exactly: both are \"start something flowing to a target,\"
with the target's resource type given by its path prefix rather than a separate flag."

## Clarifications

### Session 2026-08-27

- Q: What shape should the command take — a `routes create` subcommand, a `routes connect`
  subcommand, or something else? → A: A new top-level command, `sonora route <input-id>
  <target-id>`, singular and un-nested, mirroring the existing top-level `sonora play <uri>
  <target-id>` command exactly.
- Q: Should the input and target arguments be bare identifiers with explicit `--group`/
  `--output` flags to disambiguate a colliding identifier, or path-style resource addressing?
  → A: Path-style addressing, per `docs/cli-command-landscape.md` — the target is
  written as `outputs/<id>` or `groups/<id>` (short aliases `out/<id>` / `gr/<id>` also
  accepted), and the input as `inputs/<id>` (alias `in/<id>`). The resource type is read
  directly from the path prefix, so `--group`/`--output` flags are unnecessary and the
  "ambiguous target" failure mode no longer exists — an identifier shared by an output and a
  group is resolved deterministically by which prefix the user typed.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Route an existing input to a single output (Priority: P1)

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

---

### User Story 2 - Route an existing input to an output group (Priority: P2)

A person wants to send an existing input to a group of outputs at once (e.g. "whole house")
rather than a single speaker.

**Why this priority**: Groups are a first-class routing target already supported elsewhere
in the CLI; routing should work the same way against a group as against a single output, but
this is secondary to the single-output happy path.

**Independent Test**: Can be fully tested by running `sonora route inputs/<input-id>
groups/<group-id>` and verifying the same success output as User Story 1, with the route's
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

---

### Edge Cases

- What happens when the input path does not correspond to an existing input? The command
  must fail with a clear "input not found" error distinct from "target not found."
- What happens when the target path (`outputs/<id>` or `groups/<id>`) does not correspond to
  an existing resource of that exact type? The command must fail with a clear "target not
  found" error, even if a resource with the same identifier exists as the other type.
- What happens when the hub accepts the request but cannot create the underlying route (e.g.
  the input is disabled, or the target is already at capacity)? The command must fail with a
  clear "route creation failed" message distinct from "not found" and from validation
  errors.
- What happens when neither the input path nor the target path is supplied? The user must
  see a clear usage error explaining what's required, not a confusing failure.
- What happens when the input path's prefix is not `inputs`/`in`, or the target path's prefix
  is neither `outputs`/`out` nor `groups`/`gr`? The command must fail with a clear usage
  error identifying the invalid prefix, rather than guessing a resource type.
- What happens when the input identifier and the target identifier refer to the same
  underlying entity, or the input is already routed to that exact target? The command must
  surface whatever outcome the hub reports (success or a "route creation failed" error)
  rather than the CLI silently blocking or silently deduplicating the request.
- What happens when the hub is unreachable or does not respond within a bounded time? The
  command must fail with a clear, actionable error rather than hanging indefinitely.
- What happens when the hub returns a response that doesn't match the expected shape? The
  command must fail clearly rather than displaying garbled or silently incomplete output.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The CLI MUST provide a top-level command that starts routing an existing,
  already-defined input to a target output or output group, without creating any new input
  as a side effect.
- **FR-002**: The command MUST require an input path (`inputs/<input-id>`, alias
  `in/<input-id>`) and a target path (`outputs/<target-id>` or `groups/<target-id>`, aliases
  `out/<target-id>` / `gr/<target-id>`) as arguments, and MUST fail with a clear usage error
  if either is omitted.
- **FR-002a**: The command MUST fail with a clear usage error if the input path's prefix is
  not `inputs` or its alias `in`.
- **FR-002b**: The command MUST fail with a clear usage error if the target path's prefix is
  not one of `outputs`/`out` or `groups`/`gr`.
- **FR-003**: The command MUST determine the target's resource type solely from its path
  prefix (or alias) — an identifier that exists as both an output and a group MUST NOT cause
  any ambiguity, since only the type named by the prefix is checked.
- **FR-003a**: When the target path names a type and identifier for which no matching
  resource exists, the command MUST fail with a clear "target not found" error, even if a
  resource with the same identifier exists as the other type.
- **FR-004**: When the input path's identifier does not correspond to an existing input, the
  command MUST fail with a clear "input not found" error, distinguishable from "target not
  found."
- **FR-005**: On success, the command MUST display the resulting route's identifier and
  status exactly as returned by the hub, and a human-readable confirmation message — in both
  the default human-readable view and the machine-readable (--json) view.
- **FR-006**: The command's default output MUST be human-readable and legible without any
  additional flags or tools.
- **FR-007**: The command MUST provide an explicit option to switch to strict,
  script-parseable output instead of the human-readable default.
- **FR-008**: The command MUST fail with a clear, actionable error, distinguishable from all
  other failure types, for each of: a usage error (missing argument or invalid path prefix),
  an "input not found" error, a "target not found" error, and a "route creation failed"
  error.
- **FR-009**: The command MUST fail with a clear, actionable error message when the hub is
  unreachable, and MUST fail with a clear timeout error when the hub does not respond within
  a bounded amount of time; the command MUST make a single attempt and MUST NOT
  automatically retry on network or timeout failure.
- **FR-010**: The command MUST exit with a status that is programmatically distinguishable
  across: success, a usage error, "input not found", "target not found", "route creation
  failed", and a connectivity/network error.
- **FR-011**: The command MUST reject a routing result from the hub that does not match the
  expected structure with a clear error, rather than displaying partial or malformed
  results.

### Key Entities

- **Routing Request**: A one-shot instruction to connect a given, already-existing input to
  a given target (a single output or an output group), addressed as `inputs/<input-id>` and
  `outputs/<target-id>` / `groups/<target-id>` respectively.
- **Routing Result**: The outcome of a routing request — the identifier and status of the
  route created to carry audio from the input to the target, and a confirmation message.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A user can connect an existing input to a single output or a group in one
  command invocation, without creating a duplicate input as a side effect.
- **SC-002**: A user receives a response confirming the hub has accepted the routing request
  (or a clear failure) in under 5 seconds under normal network conditions.
- **SC-003**: 100% of failure cases return a message that lets the user tell apart: a usage
  error (missing argument or invalid resource-path prefix), input not found, target not
  found, route creation failure, and a connectivity/timeout problem.
- **SC-004**: 100% of machine-readable output produced by the command can be parsed by a
  standard structured-data parser without error.
- **SC-005**: A user can always tell, without ambiguity, whether the route was created
  successfully or the command failed.

## Assumptions

- The hub's network location (host/port) is available to the CLI through the same mechanism
  established by the existing `outputs`/`inputs`/`routes`/`groups`/`play` features; this
  feature consumes that mechanism but does not redefine it.
- The hub API requires no authentication for this endpoint, consistent with the current API
  specification.
- "Human-readable" default output presents the same set of fields as the machine-readable
  output, formatted for readability rather than strict parsing, consistent with the other
  commands.
- The default request timeout is 5 seconds, consistent with the other commands.
- This command does not accept a starting volume or a display-name option — those apply only
  to the ephemeral input `play` creates; an existing input's name and settings are managed
  elsewhere and are out of scope here.
- Stopping, pausing, or otherwise managing the resulting route (or the input) is out of
  scope for this feature; it covers only creating the route. Managing it afterward is
  expected to be handled by future route-management commands, not by this command.
- Whether a given input can be routed to more than one target at once, or whether routing an
  already-routed input to the same target again succeeds or fails, is decided entirely by
  the hub; the CLI applies no additional client-side restriction and surfaces whatever
  outcome the hub reports.
- Output identifiers and group identifiers may occasionally collide (the hub does not
  guarantee a shared namespace between the two); this is resolved entirely by the target
  argument's path prefix (`outputs/<id>` vs `groups/<id>`, or the `out`/`gr` aliases) rather
  than by a separate disambiguation flag, consistent with the `play` command and the broader
  CLI command landscape (see `docs/cli-command-landscape.md`).
