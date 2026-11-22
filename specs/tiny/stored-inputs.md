# TinySpec: Adopt stored inputs (HUB-041, API 0.1.35)

**Branch**: feat/stored-inputs-hub-041
**Date**: 2026-11-22
**Status**: implemented, tag pending
**Complexity**: small

## What

The hub can keep an input across restarts (origin `STORED`). The CLI decodes the new fields but cannot
create a kept input, promote a runtime one, or show origin, playability and the new 409 reasons.
`api/openapi.json` is already at 0.1.35. The release tag and the sonora-mcp bump follow this change.

## Context

| File | Role |
|------|------|
| `hub/inputs.go` | Modified: `CreateInputRequest.Persistent`, `PersistInput` |
| `hub/errors.go` | Modified: reason constants, friendly 409 text, 500 detail |
| `internal/cli/inputs/create.go` | Modified: `--persistent` |
| `internal/cli/inputs/keep.go` | New: `keep inputs/<id>` |
| `internal/render/inputs.go` | Modified: origin, playable, unplayableReason in YAML |
| `cmd/sonora/main.go` | Modified: `keep` verb, help |
| `tests/contract/` | New tests for create, keep, errors, rendering |
| `README.md` | Modified: commands |

## Requirements

1. `create inputs/<id> <uri> --persistent` sends `persistent: true`; with `--auto-remove` it is a usage
   error before any request. Without the flag the body is unchanged.
2. `keep inputs/<id>` calls `PUT /api/v2/inputs/{id}/persistent` and prints the input; 404 is not
   found, 400/409 carry the hub's detail, 500 shows the store's `detail`.
3. 409 reasons `INPUT_UNPLAYABLE`, `INPUT_ID_TAKEN`, `INPUT_STORE_FULL` give a short message and stay
   `ClassConflict`; an unknown reason stays a generic conflict.
4. YAML for one input and for lists shows `origin`, `playable` and, when set, `unplayableReason`;
   JSON is unchanged. An empty origin (hub older than 0.1.35) is omitted.
5. Help text lists `keep` and `--persistent`; other commands behave as before.

## Plan

1. Tests first (contract tests with an `httptest` hub), then the hub package, the commands, the renderer.
2. Live check on `multiroom.lan` (below), then docs.

## Tasks

- [x] Failing tests: create body, `PersistInput` statuses, reason messages, render
- [x] `hub`: `Persistent`, `PersistInput`, reason constants, `ClassifyError` messages, 500 detail
- [x] `create --persistent` with the `--auto-remove` conflict
- [x] `keep` command and dispatch, help text
- [x] Render origin / playable / unplayableReason
- [x] README and help
- [x] `gofmt -l .`, `go vet ./...`, `go test ./...` (no `make` here)
- [x] Live check on the Pi: create a persistent input (no route, silent), `get`, `keep` on a runtime
      input, `keep` again (no change), `create` the same id (409 `INPUT_ID_TAKEN`), delete both;
      confirm `list` is back to the starting state

## Done When

- [x] All tasks checked off, results recorded here
- [x] Tests pass, no vet or gofmt findings
- [x] Live check left no input behind

## Verification (2026-11-22)

- `gofmt -l .` empty, `go vet ./...` clean, `go test ./...` all packages ok.
- Live, `multiroom.lan` (hub 0.1.35), no route started: `create --persistent` gave `origin: STORED`;
  a plain `create` gave `RUNTIME`; `keep` on it gave `STORED` (and again, unchanged); a repeated id gave
  409 `INPUT_ID_TAKEN` (exit 14); `keep` on an unknown id gave 404 (exit 5); `--persistent
  --auto-remove` was a usage error (exit 2). Both test inputs deleted; 17 inputs before and after.
- Not exercised live: `INPUT_STORE_FULL`, `INPUT_UNPLAYABLE` and the store 500 (covered by contract tests).
