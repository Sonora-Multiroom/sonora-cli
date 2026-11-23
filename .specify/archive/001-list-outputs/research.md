# Research: List Audio Outputs

**Feature**: `001-list-outputs` | **Date**: 2026-08-24

This document resolves every open technical question for implementing `sonora outputs list`
before design (Phase 1) begins. There are no unresolved `NEEDS CLARIFICATION` markers
remaining after this phase.

## 1. Language & toolchain

- **Decision**: Go (module-based, `go.mod` at repo root; this is the first code in the
  repo, so `go mod init` will target the current stable Go release, Go 1.27.0 (released
  2026-08-19)).
- **Rationale**: The constitution's Principle III examples (`net/http`, `encoding/json`,
  `flag`) and Principle I's `go test -bench` requirement presuppose Go. No other language
  is referenced anywhere in the constitution, spec, or `AGENTS.md`.
- **Alternatives considered**: None — the constitution does not leave this open.

## 2. Dependency policy: stdlib-first (Principle III)

- **Decision**: Build the entire feature — HTTP client, argument parsing, JSON decode,
  and output rendering — on the Go standard library only (`net/http`, `encoding/json`,
  `flag`, `os`, `fmt`, `time`, `net/http/httptest` for tests). No third-party module is
  added by this feature.
- **Rationale**: Principle III requires a new dependency to provide value "that cannot be
  reasonably achieved with the standard library," with impact on binary size/startup noted
  in the PR. `flag` covers CLI parsing; `encoding/json` covers decode/validation; a bounded,
  fixed-schema YAML emitter (see §3) is small enough to hand-write correctly.
- **Alternatives considered**: A CLI framework (e.g. `cobra`) — rejected: Principle I
  disallows reflection/DI-heavy frameworks without a benchmark proving negligible startup
  cost, and `flag` is sufficient for a single `outputs list` command with three flags.

## 3. Default human-readable output format (Principle V: YAML default, `--json` opt-in)

- **Decision**: Implement a small, purpose-built YAML emitter scoped to exactly the shapes
  this feature needs — a top-level list of flat records with known field names and known
  scalar types (`string`, `int`, `bool`). String scalars are always double-quoted (side-
  stepping YAML's plain-scalar ambiguity rules for colons, leading indicators, etc.);
  bools/ints are emitted bare. This lives in `internal/render` so later list commands
  (routes, groups, inputs) can reuse the same primitive.
- **Rationale**: Principle V mandates YAML as the default format for data-returning
  commands. Go's stdlib has no YAML encoder, so *some* YAML-producing code is required.
  Because every field in `OutputResponse` (openapi.json) is a flat scalar — no nested
  maps, no multi-line strings — a general-purpose YAML library is not needed to produce
  correct output; a small fixed-shape emitter is fully sufficient and keeps Principle III's
  "no dependency unless the stdlib can't reasonably do it" intact.
- **Alternatives considered**:
  - `gopkg.in/yaml.v3` (or similar) — rejected for *this* feature: it would be the first
    third-party dependency in the whole project, solely to marshal flat records whose
    shape is already fully known and bounded. If a future command needs genuinely
    general-purpose YAML (nested structures, anchors, etc.), that command's plan can
    introduce the dependency then, justified by that command's actual need — not
    speculatively here.
  - Reusing `encoding/json` with indentation as the "human-readable" default — rejected:
    the constitution is explicit that the default must be YAML, not indented JSON.

## 4. HTTP client construction & timeout behavior (Principle IV; clarified: 5s, no retries)

- **Decision**: One `*http.Client` per CLI invocation, constructed inside the command
  handler (not at package `init()`/global scope, to respect Principle I's "no blocking
  work before argument parsing"), with `Timeout: 5 * time.Second` covering the full
  request/response round trip. A single request attempt is made — no retry loop, no
  backoff. The default `http.Transport` (connection reuse/keep-alive) is used as-is,
  satisfying the "reuse connections" guidance even though this feature issues only one
  request per invocation; this keeps the client shape consistent for future multi-request
  commands.
- **Rationale**: Directly implements the spec's clarified answers (fail fast, single
  attempt, 5s timeout) and Principle IV ("explicit connect and request timeouts...
  unbounded waits are not permitted").
- **Alternatives considered**: Separate connect-timeout vs. total-timeout (via
  `context.WithTimeout` + custom `DialContext`) — rejected as unnecessary complexity for a
  single-request command; a single `http.Client.Timeout` already bounds the whole
  operation, which is sufficient to satisfy FR-009/SC-002.

## 5. Hub location resolution (spec Assumption: "established once... this feature consumes but does not define")

- **Decision**: Resolve the hub base URL through four layers, highest precedence first:
  1. `--hub-url` flag
  2. `MULTIROOM_URL` environment variable
  3. `hubUrl` field in a persisted user config file at `~/.config/sonora/config.json`
     (JSON; resolved via `os.UserHomeDir()` + `.config/sonora/config.json`, so it lands
     under `%USERPROFILE%\.config\sonora\config.json` on Windows too)
  4. Built-in default `http://localhost:8080` (the `servers[0].url` documented in
     `api/openapi.json`)

  (`MULTIROOM_URL` names the hub/service being addressed, consistent with the product name
  "Multiroom Audio Hub" rather than the CLI's own name — avoids the `SONORA_`-prefix
  implying it's a setting *of* the CLI binary itself.)

  This feature only *reads* that file — if it doesn't exist, resolution silently falls
  through to the next layer; if it exists but isn't valid JSON or `hubUrl` isn't a string,
  that's a usage error (exit `2`) with a message naming the file, since it's a local
  misconfiguration, not a hub problem. *Writing* the file (e.g. a `sonora config set
  hub-url ...` command) is explicitly out of scope for this feature — introducing it here
  would pull a general config-management command into what's meant to be the read-only
  "list outputs" feature; it belongs to whatever future feature defines the config
  mechanism per the spec's Assumptions. Until that exists, the file is populated by hand.
  The read happens lazily, only when `outputs list` actually needs to resolve a hub URL —
  after `flag.Parse()` — so Principle I's "no blocking work before argument parsing" and
  "config discovery... scoped to what the invoked command actually needs" both hold.
- **Rationale**: The spec scopes full config/discovery out of this feature but requires
  *some* mechanism to exist for the command to run against a real or test hub; the user has
  additionally directed that config persist in `~/.config/sonora/config.json`. Layering it
  under the flag/env override (rather than replacing them) keeps every prior use case
  working — CI/tests can still pin a URL via flag or env without touching the filesystem —
  while giving an interactive user a way to set a hub URL once and not repeat it on every
  invocation.
- **Alternatives considered**:
  - Hardcoding `http://localhost:8080` with no override — rejected: it would make
    contract/integration tests (Principle VI, against a mock `httptest.Server` on a random
    port) impossible without process-level environment hacking or monkey-patching, which is
    worse than a documented flag/env/config layering.
  - Respecting `$XDG_CONFIG_HOME` instead of hardcoding `~/.config` — rejected for now:
    the user specified the literal path `~/.config/sonora/config.json`; honoring
    `XDG_CONFIG_HOME` can be added later without breaking this shape if ever requested.
  - Building a full `sonora config` read/write command as part of this feature — rejected:
    the spec explicitly scopes the config *mechanism* itself out of "list outputs"; this
    feature only needs to *consume* a hub URL, so it adds read support for the file and
    nothing more.

## 6. Exit code scheme (Principle V minimum classes; FR-011)

- **Decision**:
  | Code | Meaning |
  |------|---------|
  | `0` | Success (including the zero-outputs case, per FR-012) |
  | `2` | Usage error (bad/unknown flags) |
  | `3` | Hub-reported error (non-2xx HTTP response, or a 2xx response whose body does not match the expected `OutputResponse` shape — FR-013) |
  | `4` | Network/connectivity error (unreachable host, DNS failure, timeout) |
- **Rationale**: FR-011 requires success, usage error, hub error, and network error to be
  programmatically distinguishable; `2` matches Go's own `flag` package convention
  (`flag.ExitOnError` calls `os.Exit(2)` on parse failure), so a manual usage-error path
  stays consistent with `flag`'s built-in behavior rather than colliding with it.
- **Alternatives considered**: Reusing exit code `1` for all failures — rejected: fails
  FR-011's requirement that failure classes be distinguishable, not just "failed."

## 7. Error message translation (Principle IV: no raw errors by default, `--verbose` debug mode)

- **Decision**: A small error-classification layer in `internal/hub` maps Go/stdlib errors
  (`context.DeadlineExceeded`, `*net.OpError`, non-2xx status, JSON decode failure) to one
  of the three failure classes above, each with a short, actionable, user-facing message.
  A `--verbose` flag on the command prints the underlying Go error (e.g. via `%+v`) in
  addition to the friendly message; without `--verbose`, only the friendly message is
  shown.
- **Rationale**: Principle IV requires exactly this behavior ("raw Go errors... MUST NOT
  be the default output, but the underlying error MUST remain available via a
  `--verbose`/debug mode"). This is a whole-CLI mandate, so it applies to the very first
  command even though the spec's FRs don't name `--verbose` explicitly.
- **Alternatives considered**: A global `--debug` env var instead of a per-command flag —
  rejected: flags are the CLI-idiomatic, discoverable mechanism (Principle V), and keeps
  the option consistent with `--json`/`--include-disabled`.

## 8. Testing strategy (Principle VI: TDD, contract tests against a mock server)

- **Decision**: Three test layers, all written before implementation:
  - **Unit** (`tests/unit`): `internal/hub` request construction and response decoding;
    `internal/render` YAML/JSON output formatting including the "zero outputs" and
    "unavailable output" display cases.
  - **Contract** (`tests/contract`): an `httptest.Server` that serves responses shaped
    exactly like `GET /api/v2/outputs` in `api/openapi.json` (including the
    `includeDisabled` query parameter and the `OutputResponse` schema fields), asserting
    the CLI's HTTP client sends the right request and correctly rejects a malformed body.
  - **Integration** (`tests/integration`): full `sonora outputs list` process/command
    invocation against the same mock server, asserting stdout format (YAML/JSON), exit
    codes, and the include-disabled behavior end-to-end.
- **Rationale**: Principle VI requires unit tests for command logic and the API client
  wrapper, and mock-server contract tests before the calling code is merged; per Principle
  II the mock server's response shapes must trace back to `api/openapi.json`.
- **Alternatives considered**: Skipping contract tests since `outputs list` isn't listed
  among Principle VI's named "critical" flows (playback/routing/volume/mute) — rejected:
  this feature is explicitly the shared skeleton every future command reuses, so locking
  down the HTTP/decode contract now is high-leverage and cheap while the surface is small.
