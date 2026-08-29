# TinySpec: `create inputs` / `delete inputs` command

**Branch**: feature/inputs-create-delete-command
**Date**: 2026-08-29
**Status**: draft
**Complexity**: small

## What

Add `sonora create inputs/<input-id> <uri> --display-name <name> [--auto-remove] [--disabled]`
(`POST /api/v2/inputs`, operationId `createInput`) and `sonora delete inputs/<input-id>`
(`DELETE /api/v2/inputs/{inputId}`, operationId `deleteInput`). `delete` already exists as a
verb (routes-only today); this extends its dispatcher to also accept `inputs`. `create` is a
brand-new top-level verb. Both follow the existing per-resource `Run*` + top-level `dispatch*`
pattern.

## Context

| File | Role |
|------|------|
| `internal/hub/inputs.go` | Modified — add `CreateInputRequest` struct (mirrors `CreateInputRequest` schema: inputId/displayName/uri/enabled/autoRemove) and `CreateInput`/`DeleteInput` funcs, mirroring `CreateRoute`/`DeleteRoute`'s error handling |
| `internal/render/inputs.go` | Context only — `create` renders the returned `InputResponse` with existing `RenderInputYAML`/`JSON`; `delete` needs a new `RenderInputDeletedYAML`/`JSON`, mirroring `RenderRouteDeletedYAML`/`JSON` in `internal/render/route.go` |
| `internal/cli/inputs/create.go` | New — `RunCreate`, mirroring `route.Run`'s two-positional-arg parsing (`<input-id>` from the resource path, `<uri>` positional) plus `--display-name` (required string), `--auto-remove`/`--disabled` (bool flags) |
| `internal/cli/inputs/delete.go` | New — `RunDelete`, mirroring `routes/delete.go`'s single-positional-arg shape |
| `cmd/sonora/main.go` | Modified — add `case "create":` dispatching only `inputs` to `inputs.RunCreate`; extend `dispatchDelete` to also accept `respath.Inputs` → `inputs.RunDelete`; update `helpText` |
| `internal/hub/errors.go` | Context only — reuses `ClassValidation` (400, duplicate-id `409` falls through `APIError`'s switch to the existing `ClassHub` default, since no dedicated "already exists" class exists yet) and `ClassNotFound` |
| `tests/contract/inputs_create_test.go` | New — `TestCreateInput_*` against a mock server (201, 400, 409, malformed body), mirroring `tests/contract/route_test.go`'s `TestCreateRoute_*` |
| `tests/contract/inputs_delete_test.go` | New — `TestDeleteInput_*` (204, 400 static-input, 404, malformed error body), mirroring `TestDeleteRoute_*` |
| `tests/unit/cli_inputs_create_test.go` | New — unit tests for `RunCreate` (missing id/uri, missing `--display-name`, 409, success, `--json`) |
| `tests/unit/cli_inputs_delete_test.go` | New — unit tests for `RunDelete` (missing id, unexpected args, not-found, 400 static-input, success, `--json`) |
| `tests/unit/render_inputs_test.go` | Modified — unit tests for `RenderInputDeletedYAML`/`JSON` |
| `tests/integration/inputs_create_test.go` | New — end-to-end `create inputs/<id> <uri>` dispatch test |
| `tests/integration/inputs_delete_test.go` | New — end-to-end `delete inputs/<id>` dispatch test, plus a case confirming `delete routes/<id>` still works unaffected |
| `README.md` | Modified — document `create`/`delete` for `inputs` in the command table |
| `docs/cli-command-landscape.md` | Modified — mark the `create inputs` and `delete inputs` rows ✅ |

## Requirements

1. `sonora create inputs/<input-id> <uri> --display-name <name> [--auto-remove] [--disabled]`
   sends `POST /api/v2/inputs` with `{inputId, uri, displayName, autoRemove, enabled}` where
   `enabled` is `false` when `--disabled` is set, else `true` (auto-remove defaults `false`).
   On success (201) it prints the created input record (YAML by default, JSON with `--json`)
   and exits 0. `--display-name` is required; its absence is a usage error.
2. `sonora delete inputs/<input-id>` sends `DELETE /api/v2/inputs/{inputId}`; on success (204)
   it prints an inputId/status/message record (mirroring `RenderRouteDeletedYAML`/`JSON`'s
   shape) and exits 0. `sonora delete routes/<id>` continues to work unchanged.
3. A 404 on create's target lookup n/a; for create, a 409 (duplicate input id) reports the
   hub's error detail (or a generic "hub reported an error" fallback) and exits with
   `ClassHub`'s code; a 400 (validation) reports the detail and exits with `ClassValidation`'s
   code. For delete, a 404 reports "input not found: `<id>`" and exits with `ClassNotFound`'s
   code; a 400 (static input, cannot be deleted) reports the detail and exits with
   `ClassValidation`'s code.
4. Missing `<input-id>`/`<uri>` (create) or `<input-id>` (delete), or unexpected extra
   arguments, produce a usage error on stderr (`ClassUsage`'s exit code).
5. `--json`, `--verbose`, and `--hub-url` behave the same as every other command.
6. `create` given a non-`inputs` resource path is a usage error (no other resource supports
   create yet). `delete`/`stop` given `outputs`/`groups` remains a usage error — only `routes`
   and (after this change) `inputs` support delete; `stop` stays a routes-only alias, not
   extended to `inputs`.

## Plan

1. Add contract tests in `tests/contract/inputs_create_test.go` and
   `tests/contract/inputs_delete_test.go` (fail first).
2. Implement `hub.CreateInputRequest`, `hub.CreateInput`, `hub.DeleteInput` in
   `internal/hub/inputs.go`.
3. Add render unit tests, then implement `RenderInputDeletedYAML`/`JSON` in
   `internal/render/inputs.go`.
4. Add unit tests, then implement `inputs.RunCreate` (`internal/cli/inputs/create.go`) and
   `inputs.RunDelete` (`internal/cli/inputs/delete.go`), mirroring `route.Run`/
   `routes/delete.go`'s structure.
5. Add integration tests, then wire `case "create":` (inputs-only, new dispatcher) and extend
   `dispatchDelete` to accept `respath.Inputs` in `cmd/sonora/main.go`; update `helpText`.
6. Update `README.md`'s command table and `docs/cli-command-landscape.md`'s ✅ markers.

## Tasks

- [ ] Write contract tests for `hub.CreateInput`/`hub.DeleteInput`
- [ ] Implement `hub.CreateInputRequest`, `hub.CreateInput`, `hub.DeleteInput` in `internal/hub/inputs.go`
- [ ] Write + implement `RenderInputDeletedYAML`/`JSON` in `internal/render/inputs.go`
- [ ] Write unit tests + implement `inputs.RunCreate` in `internal/cli/inputs/create.go`
- [ ] Write unit tests + implement `inputs.RunDelete` in `internal/cli/inputs/delete.go`
- [ ] Write integration tests + wire `create` dispatch and extend `dispatchDelete` for `inputs`
      in `cmd/sonora/main.go`, update `helpText`
- [ ] Update `README.md` and `docs/cli-command-landscape.md`
- [ ] Run `go test ./...`, `go vet ./...`, `gofmt -l .`

## Done When

- [ ] All tasks checked off
- [ ] `go test ./...` passes
- [ ] `go vet ./...` and `gofmt -l .` clean
- [ ] `sonora create inputs/<id> <uri> --display-name <name>` and `sonora delete inputs/<id>`
      verified end-to-end against a mock hub, including 400/404/409 error paths, and
      `sonora delete routes/<id>` confirmed unaffected
