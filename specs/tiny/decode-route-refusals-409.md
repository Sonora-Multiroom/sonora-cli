# TinySpec: Decode the hub's 409 route refusals

**Branch**: feature/decode-route-refusals-409
**Date**: 2026-10-07
**Status**: draft
**Complexity**: small

## What

Since hub 0.1.22 a route the hub refuses (disabled input, output or group, route limit, input already on
the output) answers 409 with a `detail`, a `reason` and an optional `outputId`. The `hub` client drops all
of it, so `sonora play`, `route` and `transfer` print "hub reported an error (HTTP 409)" and exit 3. Decode
the 409, classify it as a conflict with its own exit code, and bring `hub.Route` up to the 0.1.22
contract. Source: `docs/backlog/decode-route-refusals-409.md`.

## Context

| File | Role |
|------|------|
| `api/openapi.json` | Already refreshed to hub 0.1.22: 409 on `createRoute`, `transferRoute`, `playback`; `reason`/`outputId` on `ErrorResponse`; `joinMode`/`outputs` on `RouteResponse` |
| `hub/play.go` | Modified: `errorResponse` gains `Reason`, `OutputID`; `Playback` decodes 409 |
| `hub/routes.go` | Modified: `CreateRoute`, `TransferRoute` decode 409; `Route` gains `JoinMode`, `Outputs`; `TransferRoute` doc says the route ID survives |
| `hub/errors.go` | Modified: `APIError` gains `Reason`, `OutputID`; new `ClassConflict`; `ClassifyError` maps 409 |
| `internal/cli/exitcode/exitcode.go` | Modified: `ClassConflict` → 14 |
| `README.md` | Modified: exit code 14 row; `transfer` paragraph (route ID kept) |
| `tests/contract/{route,play}_test.go`, `tests/unit/{hub_client,exitcode}_test.go` | Modified: tests below |
| `internal/cli/{play,route}/*.go` | Context: already report via `ClassifyError` + `exitcode.For`; no change expected |

## Requirements

1. A 409 from `CreateRoute`, `TransferRoute` or `Playback` with a decodable body returns `*APIError` with
   `StatusCode` 409, `Title`, `Detail`, `Reason` and `OutputID` (empty when the hub omits them).
2. A 409 with an undecodable body returns `*StatusError{409}`, as 400/422 do today.
3. `Reason` is a plain `string`; its doc comment lists the six known values and says an unknown value is a
   generic conflict. No typed constants, so a new hub reason needs no client release.
4. `ClassifyError` returns `ClassConflict` for a 409 `APIError` (message: `Detail`, else `Title`, else a
   generic conflict text) and for a 409 `StatusError` (generic conflict text). `CreateInput`'s duplicate-ID
   409 therefore shows "Input ID already exists".
5. `exitcode.For(hub.ClassConflict)` is 14; README documents it as "the hub refused for its current state
   (409): change that state and retry".
6. `hub.Route` has `JoinMode string` (`json:"joinMode"`) and `Outputs []string` (`json:"outputs"`).
   Existing decoding and validation are unaffected when the hub omits them.
7. The CLI prints only the hub's `detail` for a refusal; `reason`/`outputId` stay available to library
   callers (no `--json` error-output change in this spec).

## Plan

1. Extend `errorResponse` and `APIError`; add a small helper that builds the `*APIError` from a decoded
   body so the three functions and `CreateInput` stay identical.
2. Add `http.StatusConflict` to the 400/422 cases in `CreateRoute`, `TransferRoute` and `Playback`.
3. Add `ClassConflict` at the end of the `ErrorClass` iota (existing values must not shift); add the 409
   cases in `ClassifyError`'s `APIError` and `StatusError` branches.
4. Map the class in `exitcode.For`; update the README exit table and `transfer` paragraph.
5. Add `JoinMode`/`Outputs` to `Route`; fix the `TransferRoute` doc comment.

## Tasks

- [ ] Extend `errorResponse`/`APIError` with `Reason`, `OutputID`, and the shared builder
- [ ] Decode 409 in `CreateRoute`, `TransferRoute`, `Playback`; update their doc comments
- [ ] Add `ClassConflict` and the 409 mapping in `ClassifyError`
- [ ] Map `ClassConflict` → 14 in `exitcode.For`; README exit table row
- [ ] Add `JoinMode`, `Outputs` to `hub.Route`; fix `TransferRoute` doc and README `transfer` paragraph
- [ ] Contract tests: 409 for each of the three functions with a reason and `outputId`; without
      `outputId` (`INPUT_DISABLED`); unknown reason; undecodable body → `StatusError`; `Route` decodes
      `joinMode`/`outputs`
- [ ] Unit tests: `ClassifyError` on 409 `APIError` (detail, title-only) and 409 `StatusError`;
      `exitcode.For(ClassConflict) == 14`; CLI `play` against a 409 stub exits 14 and prints the detail
- [X] Commit the refreshed `api/openapi.json` first, on its own

## Done When

- [ ] All tasks checked off
- [ ] `gofmt -l .` empty, `go vet ./...` and `go test ./...` pass
- [ ] Release notes list the additive `hub` API (`ClassConflict`, `APIError.Reason`/`OutputID`,
      `Route.JoinMode`/`Outputs`); follow-up in sonora-mcp: bump `hub`, map `ClassConflict` in `classify`
