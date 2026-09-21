# Quickstart: validating `speak` and the TTS cache commands

**Feature**: `009-tts-commands` | **Contract**: [contracts/cli-tts.md](contracts/cli-tts.md)

## Prerequisites

- Go toolchain matching `go.mod` (1.27).
- For the manual checks: a Multiroom Audio Hub reachable at `$MULTIROOM_URL`, with the TTS
  extension loaded, at least one provider configured, an output (here `kitchen`) and a group
  (here `all-rooms`).

## 1. Automated checks

```sh
make check                              # gofmt + go vet + go test ./...
go test ./tests/contract/ -run 'TTS|Extensions' -v
go test ./tests/unit/ -run 'TTS|Speak|Cache' -v
go test ./tests/integration/ -run 'TTS|Speak|Cache' -v   # includes the ~11 s slow-provider test
go test -short ./...                                      # same, without the slow test
```

Expected: all pass. The integration suite's request counters confirm SC-005a: one request per
successful command, and only one inventory call after a 404.

## 2. Build

```sh
make build        # produces ./sonora.exe (or ./sonora via docker-build)
```

## 3. Manual scenarios (against a real hub)

| # | Command | Expect |
| --- | --- | --- |
| 1 | `sonora speak "Dinner is ready" outputs/kitchen` | Exit 0, YAML with `announcementId`, `cacheHit`, `queueDepth`, and the phrase is heard in the kitchen (US1) |
| 2 | Repeat #1 | `cacheHit: true`, returns in under 1 s (SC-002) |
| 3 | `sonora speak "Motion detected in garden" gr/all-rooms --json` | One-line JSON, heard on every output in the group (US2) |
| 4 | `echo "Piped hello" \| sonora speak - out/kitchen` | Exit 0; hub receives `Piped hello` (FR-003a) |
| 5 | `sonora speak -- "-5 degrees outside" out/kitchen` | Exit 0; text spoken as-is |
| 6 | `sonora speak "Hi" out/kitchen --provider no-such-provider` | `PROVIDER_NOT_FOUND: …`, exit 5 (US3) |
| 7 | `sonora speak "Hi" out/no-such-output` | `TARGET_NOT_FOUND: …`, exit 12 |
| 8 | `sonora speak "   " out/kitchen` | Usage error, exit 2, no request sent |
| 9 | `sonora speak "Hi" routes/x` | Usage error, exit 2 |
| 10 | `sonora get tts-cache` then `--json` | Totals and sorted per-provider counts (US4) |
| 11 | `sonora clear tts-cache --provider <name>` then `get tts-cache` | `cleared: provider`, then that provider's count is gone (US5) |
| 12 | `sonora clear tts-cache`, run twice | `cleared: all` both times, exit 0 (FR-008) |
| 13 | Disable the TTS extension in the hub config, restart, run #1 | `text-to-speech is not available on this hub: the TTS extension is installed but disabled …`, exit 13 |
| 14 | `sonora speak "Hi" out/kitchen --hub-url http://localhost:9` | `could not reach the hub`, exit 4 |
| 15 | `sonora get tts-cache --hub-url <some non-hub web server>` returning 404 for everything | Hub-address message naming the URL, exit 4 |

Check each exit code with `echo $?` (POSIX shells) or `$LASTEXITCODE` (PowerShell).

## 4. Regression check

```sh
go test ./cmd/... ./tests/...
```

Existing commands' outputs and exit codes are unchanged (SC-006). `sonora help` now lists
`speak`, `get tts-cache` and `clear tts-cache` (FR-015).
