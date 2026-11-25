# Research: List and Get Audio Inputs

**Feature**: `003-inputs-list-get` | **Date**: 2026-08-25

This document resolves every open technical question for implementing `sonora inputs list`
and `sonora inputs get <input-id>` before design (Phase 1) begins. The spec has no
unresolved `NEEDS CLARIFICATION` markers — this feature deliberately reuses every
foundational decision `001-list-outputs` and `002-outputs-get` already made (language,
dependency policy, HTTP client construction, hub URL resolution, YAML/JSON rendering
approach, `--verbose`/error-translation shape, exit-code scheme, testing strategy). Only the
decisions genuinely new to this feature — driven by the `Input` entity's different shape —
are recorded below.

## 1. The `Input` entity has no overlap with `Output`'s fields

- **Decision**: Add a new `Input` struct in `internal/hub/inputs.go`, mapping field-for-field
  to `#/components/schemas/InputResponse` in `api/openapi.json`: `InputID` (`inputId`),
  `DisplayName` (`displayName`), `URI` (`uri`), `Enabled` (`enabled`), `AutoRemove`
  (`autoRemove`), `Source` (`source`, string enum `STATIC`/`EPHEMERAL`), `CreatedAt`
  (`createdAt`, `*string`, nullable — null for static inputs), `Pauseable` (`pauseable`).
- **Rationale**: Principle II requires request/response types to be generated from or
  validated against `openapi.json`. `InputResponse` shares no fields with `OutputResponse`
  (no volume/mute/available) and adds three fields `OutputResponse` doesn't have
  (`uri`, `autoRemove`/`source`/`pauseable`, `createdAt`) — a distinct struct is the only
  contract-faithful option; retrofitting `Output` or a shared "resource" struct would either
  invent fields the hub doesn't return for outputs or bury inputs-only fields behind
  pointers/omitempty on a type meant for a different resource.
- **Alternatives considered**: A shared generic struct with optional pointers for every
  field across both resources — rejected: it would let a caller construct a nonsensical
  half-output/half-input value, and neither existing resource needs this generality;
  `001-list-outputs`'s "one struct per response schema" precedent already covers this case
  cleanly.

## 2. `CreatedAt` is the first nullable field in this codebase

- **Decision**: `CreatedAt *string` (RFC 3339 string, matching `openapi.json`'s
  `date-time` format annotation — no `time.Time` parsing/reformatting, since the CLI only
  displays this value, it never computes with it). `nil` means "no creation timestamp"
  (static inputs). Rendering: YAML emits `createdAt: null` (bare, unquoted) when `nil`, and
  `createdAt: "<value>"` (quoted, like other string fields) otherwise; JSON relies on
  `encoding/json`'s native `null` handling for a nil pointer.
- **Rationale**: FR-004 requires "an explicit indication that none exists" rather than an
  omitted or blank field — `null`/absent-value handling must be visible, not silently
  dropped, consistent with the existing "never omit a field" rule `RenderYAML` already
  established for `available: false` (001's FR-005).
- **Alternatives considered**: An empty string (`""`) for static inputs instead of a pointer
  — rejected: it's ambiguous with a hypothetical real empty-string timestamp and loses the
  schema's explicit `null` semantics; a pointer keeps the "absent" case unambiguous and
  matches the schema's declared `["string", "null"]` type exactly (Principle II).

## 3. HTTP calls: `GET /api/v2/inputs` (`listInputs`) and `GET /api/v2/inputs/{inputId}` (`getInput`)

- **Decision**: Add `hub.ListInputs(ctx, client, baseURL string, includeDisabled bool)
  ([]Input, error)` and `hub.GetInput(ctx, client, baseURL, inputID string) (*Input, error)`
  in `internal/hub/inputs.go`, structurally identical to `ListOutputs`/`GetOutput` in
  `internal/hub/outputs.go`: same query-param handling (`includeDisabled`), same path
  construction/escaping, same non-2xx/404/decode error handling.
- **Rationale**: `listInputs`/`getInput` in `api/openapi.json` have the exact same shape
  (list with `includeDisabled` query param; single-item with a 404) as the outputs
  endpoints this codebase already implements correctly and has tested — reusing the proven
  shape is lower-risk than inventing a new one, and Principle II is satisfied identically
  (decode into `Input`, matching `InputResponse`).
- **Alternatives considered**: A generic `fetchList[T]`/`fetchOne[T]` helper parameterized
  over `Output`/`Input` — rejected: Go generics here would save perhaps a dozen lines
  across two small, already-simple functions, at the cost of an extra layer of indirection
  in the one part of the codebase (`internal/hub`) most tied to Principle II's per-endpoint
  fidelity; `001-list-outputs`/`002-outputs-get` did not introduce generics, and duplicating
  the same well-tested shape keeps each endpoint's contract test straightforward to write
  and read independently.

## 4. Not-found error type: generalize rather than duplicate

- **Decision**: Generalize the existing `NotFoundError{OutputID string}` in
  `internal/hub/errors.go` to `NotFoundError{Resource, ID string}`, with
  `Error()` returning `fmt.Sprintf("%s not found: %s", e.Resource, e.ID)`. `GetOutput`
  changes its construction to `&NotFoundError{Resource: "output", ID: outputID}` (message
  text unchanged: `"output not found: <id>"`, so `002-outputs-get`'s existing tests and
  documented contract keep passing untouched). `GetInput` constructs
  `&NotFoundError{Resource: "input", ID: inputID}`, producing `"input not found: <id>"`.
  `ClassifyError`'s existing `errors.As(err, &notFoundErr)` branch needs no change beyond
  reading the new fields.
- **Rationale**: FR-011 (this spec) requires the same "not found" distinguishability
  `002-outputs-get`'s FR-012 already established. Introducing a second, input-specific
  `InputNotFoundError` type would duplicate `NotFoundError`'s entire shape for no behavioral
  difference — the two resources need the same three things (a class, an exit code, a
  templated message), which the generalized struct already provides without changing any
  existing caller-visible string.
- **Alternatives considered**: A second concrete type (`InputNotFoundError{InputID
  string}`) mirroring `NotFoundError` — rejected: pure duplication; every future
  not-found-capable resource (routes, groups) would otherwise add another near-identical
  type and another `errors.As` branch in `ClassifyError` instead of reusing one.

## 5. Exit code scheme: unchanged, no new class needed

- **Decision**: Reuse `ClassNone`/`ClassUsage`/`ClassHub`/`ClassNetwork`/`ClassNotFound`
  (exit `0`/`2`/`3`/`4`/`5`) from `internal/hub/errors.go` exactly as `002-outputs-get`
  defined them. Neither `inputs list` nor `inputs get` needs a sixth class.
- **Rationale**: FR-015 requires the same five-way distinction (success / usage / not-found
  / hub error / network error) `002-outputs-get`'s FR-012 already required and implemented;
  `inputs get`'s 404 case maps onto the existing `ClassNotFound`, and `inputs list` has no
  "not found" concept (same as `outputs list`) so it only ever produces
  none/usage/hub/network — a strict subset of the existing scheme.
- **Alternatives considered**: None — no new failure mode exists that the current scheme
  doesn't already cover.

## 6. CLI dispatch: new `internal/cli/inputs` package, same shape as `internal/cli/outputs`

- **Decision**: New package `internal/cli/inputs` with `RunList` (in `list.go`) and
  `RunGet` (in `get.go`), each a near-verbatim structural copy of
  `internal/cli/outputs`'s `RunList`/`RunGet` — same flag set names (`--json`, `--verbose`,
  `--hub-url`, plus `--include-disabled` on `list` only), same positional-argument
  re-parse loop on `get` (so `<input-id>` can appear before or after flags), same
  stdout/stderr separation, same exit-code mapping via `hub.ClassifyError`. Update
  `cmd/sonora/main.go`'s noun switch to add a new `case "inputs":` branch routing
  `"list"`/`"get"` to `inputs.RunList`/`inputs.RunGet`, alongside the existing `"outputs"`
  case.
- **Rationale**: Principle V requires consistent `<noun> <verb> [args]` structure and no
  flag-name synonyms across commands; reusing the exact flag names and argument-parsing
  shape `outputs` already established is what "consistent" means in practice here. A
  distinct package per noun (rather than a shared generic command builder) matches
  `001`/`002`'s existing precedent and keeps each noun's flag set (e.g. `get`'s positional
  ID, `list`'s `--include-disabled`) declared locally and simply.
- **Alternatives considered**: A single generic `runList`/`runGet` helper shared by both
  `outputs` and `inputs` packages, parameterized by the hub-call and render functions —
  rejected: the two commands' bodies are almost entirely flag/parse/dispatch boilerplate
  around a two-line hub call and a two-line render call; extracting that boilerplate now,
  before a third noun exists to confirm the shape actually generalizes, risks the wrong
  abstraction more than it saves — consistent with the project's "don't design for
  hypothetical future requirements" guidance. If a third noun's commands turn out
  identical in shape, that is the point to extract a helper, informed by three concrete
  examples instead of two.

## 7. Rendering: new `internal/render/inputs.go`, same pattern as `internal/render/outputs.go`

- **Decision**: Add `RenderYAML([]hub.Input) string`, `RenderJSON([]hub.Input) string`,
  `RenderInputYAML(hub.Input) string`, and `RenderInputJSON(hub.Input) string` in a new
  `internal/render/inputs.go` file (same package `render`, so the exported names stay
  `RenderYAML`/`RenderJSON`/etc. — Go's package-qualified call sites, e.g.
  `render.RenderYAML(inputs)` vs `render.RenderYAML(outputs)`, are already disambiguated by
  argument type, matching how `001`/`002` structured `internal/render/outputs.go`). Field
  emission order for both YAML and JSON: `inputId`, `displayName`, `uri`, `source`,
  `enabled`, `autoRemove`, `pauseable`, `createdAt` — identifier and display name first
  (matching the outputs renderer's convention), then the remaining attributes grouped by
  how central they are to a user distinguishing one input from another. Every field is
  always emitted explicitly, including `createdAt: null`, reusing the never-omit rule
  `001-list-outputs`'s `RenderYAML` established for `available: false`.
- **Rationale**: FR-004/FR-008/FR-009/FR-010 require the same fields in both the
  human-readable default and `--json`, in a legible, always-fully-populated form — directly
  reusing the proven per-record/list vs. single-record split `002-outputs-get`'s research.md
  §5 already established for `outputs`.
- **Alternatives considered**: Overloading `render.RenderYAML`/`RenderJSON` via a shared
  interface (e.g. a `Renderable` type both `Output` and `Input` implement) — rejected: Go's
  static overload-by-argument-type already resolves the two `RenderYAML` functions
  correctly at each call site without runtime dispatch or an interface that exists solely
  to satisfy two call sites; adding one would be complexity with no behavioral benefit.

## 8. Validation rule for malformed responses (FR-017)

- **Decision**: A decoded `Input` is rejected as malformed (`*DecodeError`, `ClassHub`) if
  `InputID == ""`, `DisplayName == ""`, or `Source` is neither `"STATIC"` nor `"EPHEMERAL"`.
  `URI`, `AutoRemove`, `Pauseable`, and `CreatedAt` are not further validated beyond their
  Go type (bool/pointer decode failure already surfaces as a JSON decode error before this
  check runs).
- **Rationale**: Mirrors `001-list-outputs`'s existing `outputId`/`displayName`
  non-empty check (its two required, always-non-empty identity fields) and extends it with
  the one new field this schema constrains to an enum (`source`) — an unrecognized `source`
  value would silently break FR-004's "source type (static or ephemeral)" display and the
  edge-case handling for `createdAt` absence (which is conditioned on knowing whether an
  input is static), so it's worth rejecting explicitly rather than displaying a bogus third
  value.
- **Alternatives considered**: Validating every field strictly (e.g. rejecting a negative
  or nonsensical `createdAt` string) — rejected: `createdAt` is opaque, display-only data
  the CLI never parses or computes with (see §2); there is no meaningful "invalid" value
  for a string the CLI just echoes, so validating it would only add surface area without
  catching real integration bugs.

## 9. Testing strategy — same three layers, extended for `inputs`

- **Decision**: Mirror `001-list-outputs`/`002-outputs-get`'s three-layer approach exactly:
  - **Unit** (`tests/unit`): `internal/hub/inputs_test.go`-equivalent request/decode
    coverage (or added to existing hub client test file) plus
    `tests/unit/render_inputs_test.go` covering list YAML/JSON, single-record YAML/JSON,
    the zero-inputs case, the disabled-input-included case, and the null-vs-populated
    `createdAt` display.
  - **Contract** (`tests/contract`): `tests/contract/inputs_list_test.go` and
    `tests/contract/inputs_get_test.go`, each an `httptest.Server` shaped from
    `listInputs`/`getInput` in `api/openapi.json` (200 `InputResponse`/array, 404 for
    `getInput`), asserting request path/query and both decode paths.
  - **Integration** (`tests/integration`): `tests/integration/inputs_list_test.go` and
    `tests/integration/inputs_get_test.go`, full `sonora inputs list`/`sonora inputs get`
    process invocations covering found/not-found/missing-identifier/`--json`/exit codes.
  All written before their corresponding implementation code, per Principle VI.
- **Rationale**: Same rationale as `001-list-outputs` research.md §8 and `002-outputs-get`
  research.md §6 — reusing the established three-layer taxonomy rather than inventing a new
  one, applied to the one genuinely new decode/render surface (`Input`'s extra fields and
  nullable `createdAt`).
- **Alternatives considered**: None specific to this feature — the layer choice was already
  settled by `001-list-outputs`.

## Related specs (from memory)

- _none found_ (no memory-tool search was available in this environment; the directly
  relevant sibling specs, `001-list-outputs` and `002-outputs-get`, are already cited
  throughout this document).
