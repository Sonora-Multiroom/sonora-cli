# Research: Adopt Verb-First Command Landscape

**Feature**: `007-refactor-cli-commands` | **Date**: 2026-08-27

No Technical Context fields were left as `NEEDS CLARIFICATION` (the spec's Clarifications
session already resolved the identifier pattern, the hard-cutover policy, and the error-message
policy). This document instead records the design decisions needed to turn the spec's
functional requirements into a concrete, low-risk implementation.

## 1. A shared `internal/cli/respath` package, not per-caller parsing

**Decision**: Implement resource-path resolution (alias lookup, `resource/id` splitting, id
pattern validation) once, in a new `internal/cli/respath` package, and have both
`cmd/sonora/main.go`'s dispatcher and `internal/cli/play` call it.

**Rationale**: FR-004 (aliases) and FR-004a (id pattern/split rule) must behave identically
everywhere a resource path appears — in `get`/`list` dispatch and in `play`'s target argument.
A single implementation is the only way to guarantee that by construction rather than by
convention; it also gives `route` (spec 008, not yet implemented) a ready-made dependency when
it's built, rather than a third copy of the same alias table.

**Alternatives considered**: Duplicating the alias table and split logic inside `main.go` and
`play` separately — rejected: two hand-maintained copies of the same table are exactly the kind
of "synonym drift" Constitution Principle V warns against, and it doubles the surface area unit
tests need to cover for no benefit.

## 2. Reuse existing `RunList`/`RunGet` functions unchanged; translate at the dispatch boundary

**Decision**: `cmd/sonora/main.go`'s new dispatcher resolves the resource path (kind + id) via
`respath`, then calls the *existing* `inputs.RunList`/`RunGet`, `outputs.RunList`/`RunGet`,
`groups.RunList`/`RunGet`, `routes.RunList`/`RunGet` functions with the same `rest []string`
argument shape they already accept today (id, if any, followed by pass-through flags) — these
functions' signatures and internal logic do not change.

**Rationale**: FR-010 requires byte-for-byte identical displayed data; the safest way to
guarantee that is to not touch the code that produces it. These four functions already own
HTTP calls, rendering, and exit-code classing correctly (and are already tested) — the only
thing genuinely changing is how a command line maps to a call to them.

**Alternatives considered**: Changing each `RunList`/`RunGet` to accept a parsed
`respath.Path` struct directly — rejected for this feature: larger diff and test churn in
already-correct code, for no functional gain, since these functions never need to know about
aliases or path syntax — only the dispatcher does.

## 3. `list` is a pre-dispatch guard, not a second code path

**Decision**: When the verb is `list`, the dispatcher resolves the resource path via `respath`
and returns a usage error immediately if an id was given; otherwise it calls the exact same
`RunList` invocation `get <resource>` (no id) would produce.

**Rationale**: Guarantees `list` and `get` (collection form) are identical by construction
(FR-003), with the "id after `list`" rejection (FR-003's invalid case) enforced in exactly one
place.

**Alternatives considered**: A `list` handler that shells out to the `get` handler at the
command level — rejected as unnecessary indirection within a single process/binary.

## 4. `play`'s target becomes a required second resource path; `--group`/`--output` are deleted from its flag set

**Decision**: `play`'s `flag.FlagSet` no longer registers `--group`/`--output`; its second
positional argument is parsed via `respath` into a resource path (`outputs/<id>` or
`groups/<id>`, aliases allowed). Passing the old `--group`/`--output`/`--name` flags now fails
with the standard "flag provided but not defined" usage error the stdlib `flag` package already
produces for any unregistered flag — no special detection of the old flags is implemented
(spec Clarifications: Option B, standard usage error).

**Rationale**: Matches FR-007/FR-008 directly, and reuses the same parsing path `route` (spec
008) will need, rather than inventing a second convention for "target path" now and a third
one later.

**Consequence for exit codes**: `play`'s existing exit code `7` ("target matches both an
output and a group; use --group or --output to disambiguate", per
`specs/006-play-command/contracts/cli-play.md`) becomes unreachable and is removed — path-style
addressing makes that ambiguity structurally impossible (spec Clarifications, id-collision
question). "Target not found" for a path naming a type+id that doesn't exist continues to use
exit code `5`, now uniformly rather than via `--group <id>`/`--output <id>`-specific branches.

**Alternatives considered**: Keeping `--group`/`--output` internally and silently mapping the
new path syntax onto them — rejected: leaves dead flag-parsing code and a redundant internal
representation for a distinction (group vs. output) the path prefix already encodes.

## 5. Old noun-verb commands are removed from the dispatch switch entirely

**Decision**: `main.go`'s dispatcher no longer special-cases `inputs`/`outputs`/`groups`/
`routes` as a first token followed by a `list`/`get` second token. Running `sonora inputs
list` falls through to the same "unknown command" usage error (exit `2`) any unrecognized
first token already produces.

**Rationale**: Matches the spec's Clarifications answer directly — a hard cutover with a
standard usage error, no tailored detection of old syntax. This also means no code needs to
"remember" the old grammar at all, keeping the dispatcher's surface area equal to the new
grammar only.

**Alternatives considered**: A dedicated error branch that recognizes `inputs`/`outputs`/
`groups`/`routes` as a first token and prints a migration hint — rejected per the spec's
explicit Option B answer.

## 6. Constitution Principle V's example needs a follow-up amendment (non-blocking)

**Decision**: Note in the plan's Constitution Check (not act on it here) that Principle V's
illustrative example (`sonora <noun> <verb> [args]`) is now stale, and recommend a
`/speckit-constitution` update after this feature ships.

**Rationale**: Constitution amendments go through their own governance workflow
(`/speckit-constitution`), not through a feature plan; treating the "e.g." as illustrative
rather than normative avoids blocking this refactor on an unrelated process step, while still
surfacing the staleness rather than silently ignoring it.

**Alternatives considered**: Editing the constitution directly as part of this feature —
rejected: out of process, and this plan is not the place to make governance decisions.
