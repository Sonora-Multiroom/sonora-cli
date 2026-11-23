# TinySpec: `help` command

**Branch**: main
**Date**: 2026-08-26
**Status**: done
**Complexity**: small

## What

Add a top-level `help` command (plus `-h`/`--help` flags, and bare invocation) to the
hand-rolled dispatcher in `cmd/sonora/main.go` that prints the noun/verb command table,
global flags, and example invocations to stdout, exit code 0.

## Context

| File | Role |
|------|------|
| `cmd/sonora/main.go` | Will be modified — add help text constant, dispatch `help`/`-h`/`--help`/bare-invocation to it |
| `cmd/sonora/main_test.go` | New — unit tests for `run()` help/usage paths (test-first per constitution Principle VI) |
| `internal/hub/errors.go` | Context — `ClassUsage.ExitCode()` = 2, reused for the existing unknown-command/missing-args paths (unchanged) |
| `README.md` | Context — noun/verb table and global flags (`-json`, `-hub-url`, `-verbose`, `-include-disabled`, etc.) to mirror in help text |

## Requirements

1. `sonora help`, `sonora -h`, and `sonora --help` print usage to **stdout** and exit **0**.
2. Bare invocation (`sonora` with no args) prints the same help to **stdout** and exits **0**
   (replacing today's stderr usage-error/exit-2 for the zero-arg case only).
3. Help output includes: top-level usage line, the noun/verb table (`outputs`, `inputs`,
   `routes`, `groups` × `list`/`get`), the verb-less `play <uri> <target-id>` command, the
   global flags (`-json`, `-hub-url`, `-verbose`), and at least one example invocation.
4. `sonora <noun>` with a missing verb, and `sonora <noun> <verb>` with an unknown verb/noun,
   continue to behave exactly as today (stderr message, exit code 2) — unaffected by this
   change.
5. `--version`/`-v` handling is unaffected and still checked before help dispatch.

## Plan

1. In `cmd/sonora/main.go`, add a `helpText` string constant (or small builder func) with the
   usage/table/flags/examples content, written to stdout via `fmt.Fprint(stdout, helpText)`.
2. At the top of `run()`, after the `--version`/`-v` check, add a check: if
   `len(args) == 0 || args[0] == "help" || args[0] == "-h" || args[0] == "--help"`, print
   help to `stdout` and `return 0`.
3. Remove the now-unreachable old zero-arg branch (`len(args) < 2` still guards the
   missing-verb case for `len(args) == 1` non-help nouns; keep that check but adjust so it
   only fires when `len(args) == 1`, since `len(args) == 0` is now handled by step 2).
4. Update `README.md`'s Commands section with a one-line mention of `sonora help`.

## Tasks

- [x] Write `cmd/sonora/main_test.go` covering: `help`, `-h`, `--help`, bare invocation (all →
      stdout output containing key command names, exit 0); unknown noun still exit 2 to
      stderr; missing verb (`sonora outputs`) still exit 2 to stderr
- [x] Implement help dispatch and help text in `cmd/sonora/main.go`
- [x] Run `go test ./...` and confirm new + existing tests pass
- [x] Update `README.md` Commands section to mention `sonora help`

## Done When

- [x] All tasks checked off
- [x] `go test ./...` passes (pre-existing, unrelated `tests/contract` link failure due to
      the environment's C: drive being full — not caused by this change; `cmd/sonora`,
      `tests/unit`, `tests/integration` all pass)
- [x] `go vet ./cmd/...` and `gofmt -l cmd/sonora/` clean
