# TinySpec: `rename inputs` command

**Branch**: feature/rename-inputs-command
**Date**: 2026-10-22
**Status**: done
**Complexity**: small

## What

Add `sonora rename inputs/<input-id> <name>` and `sonora rename inputs/<input-id> --reset`, calling
`PUT /api/v2/inputs/{inputId}/display-name` (operationId `setInputDisplayName`, hub 0.1.27) with a
`DisplayNameRequest`. The hub returns the full `InputResponse`, rendered like `get inputs/<id>`. The
`hub` package gains the call so sonora-mcp can offer a rename tool afterwards. Source:
the backlog item `docs/backlog/rename-inputs.md` (removed once this shipped). The verb is `rename`, not
`set ... --display-name`: it reads naturally and `--reset` fits it.

## Context

| File | Role |
|------|------|
| `hub/inputs.go` | Modified — add `DisplayNameRequest` and `SetInputDisplayName(ctx, client, baseURL, inputID string, name *string) (*Input, error)`, shaped like `SetInputEnabled` |
| `hub/errors.go` | Context — `apiErrorFromBody` for 400, `NotFoundError` for 404 |
| `internal/cli/inputs/rename.go` | New — `RunRename`, mirroring `enabled.go`'s flag loop, error reporting and rendering |
| `internal/render/inputs.go` | Context — reuses `RenderInputYAML`/`RenderInputJSON` |
| `cmd/sonora/main.go` | Modified — `case "rename"` → `dispatchRename` (inputs only), `helpText` example |
| `tests/contract/inputs_display_name_test.go` | New — contract tests for `hub.SetInputDisplayName` |
| `tests/integration/inputs_rename_test.go` | New — end-to-end tests of the command |
| `README.md`, `docs/cli-command-landscape.md` | Modified — command table row and rename lifetime note |
| `docs/backlog/rename-inputs.md`, `docs/backlog/INDEX.md` | Removed / updated once shipped |

## Requirements

1. `DisplayNameRequest{DisplayName *string}` encodes as `{"displayName": "<name>"}`, or
   `{"displayName": null}` when nil (no `omitempty`). The hub's godoc says null or blank restores the
   original name.
2. `SetInputDisplayName`: 200 → decoded, validated `*Input`; 404 → `*NotFoundError{Resource: "input"}`;
   400 → `*APIError` carrying the hub's detail (fallback `*StatusError`); malformed body →
   `*DecodeError`; other non-2xx → `*StatusError`.
3. `sonora rename inputs/<id> <name>` prints the renamed input (YAML, or JSON with `--json`), exit 0.
   `--reset` sends `null` and prints the input with its original name.
4. A name together with `--reset`, neither of them, a missing `<input-id>`, or extra positional arguments
   is a usage error on stderr (exit 2). A multi-word name is one quoted argument.
5. `rename` with any resource other than `inputs` (e.g. `outputs/<id>`) is a usage error (exit 2).
6. Name length and characters are validated by the hub only; its 400 detail is printed, exit code from
   `ClassifyError`. `--verbose` and `--hub-url` behave as elsewhere; nothing goes to stdout on failure.
7. `--help` prints the usage and says a rename lasts until a restart for a configured input and until
   removal for a runtime input.

## Plan

1. Contract tests first (200 rename, 200 reset with `null` body asserted, 400 with detail, 404,
   malformed body, request path and method), then implement in `hub/inputs.go`.
2. Integration tests first (rename, `--reset`, `--json`, name + `--reset`, neither, non-inputs
   resource, 404), then `internal/cli/inputs/rename.go` and `dispatchRename` in `main.go`.
3. README, landscape doc, `helpText`; remove the backlog item and its `INDEX.md` row.

## Tasks

- [x] Write contract tests in `tests/contract/inputs_display_name_test.go`
- [x] Implement `DisplayNameRequest` and `hub.SetInputDisplayName` in `hub/inputs.go`
- [x] Write integration tests in `tests/integration/inputs_rename_test.go`
- [x] Implement `inputs.RunRename` in `internal/cli/inputs/rename.go`
- [x] Add `dispatchRename` and the `helpText` example to `cmd/sonora/main.go`
- [x] Update `README.md` and `docs/cli-command-landscape.md`
- [x] Delete `docs/backlog/rename-inputs.md`, update `docs/backlog/INDEX.md`
- [x] Run `gofmt -l .`, `go vet ./...`, `go test ./...`

## Done When

- [x] All tasks checked off
- [x] `go test ./...` passes; `go vet ./...` and `gofmt -l .` clean
- [x] Rename, reset and both usage errors verified against a mock hub

## Verification

- 2026-10-22, local: `gofmt -l .` and `go vet ./...` clean; `go test ./...` passes, with 6 new contract
  tests and 10 new integration tests.
- 2026-10-22, production hub `multiroom.lan` (API 0.1.28), on a temporary disabled runtime input
  `rename-test`, never routed and deleted afterwards. 14 of 14 checks passed:
  - rename → new name, exit 0; `get` shows it
  - `--json` with a padded name → name trimmed by the hub
  - `--reset` → original "Rename Test"
  - 65-character name → hub detail "Display name is longer than 64 characters", exit 6;
    `--verbose` adds the detail line
  - unknown input → "input not found", exit 5
  - name with `--reset`, no name, and `outputs/` → usage errors, exit 2
