# Quickstart: Get Single Audio Output

**Feature**: `002-outputs-get` | **Date**: 2026-08-25

Validates that `sonora outputs get <output-id>` works end-to-end. Uses a local mock hub for
deterministic validation; the same steps apply against a real Multiroom Audio Hub by
pointing `--hub-url` at it once one is available.

## Prerequisites

- Go toolchain installed (repo's `go.mod` already exists, from `001-list-outputs`).
- Repo built: `go build -o sonora ./cmd/sonora` from repo root.

## 1. Run the automated test suite (primary validation)

```bash
go test ./...
```

Expected: all unit (`tests/unit`), contract (`tests/contract`), and integration
(`tests/integration`) tests pass, including the new `outputs get` tests alongside the
existing `outputs list` ones. The contract and integration tests spin up an
`httptest.Server` shaped from `api/openapi.json`'s `getOutput` operation (`200
OutputResponse` / `404 ErrorResponse`) — no external hub needed to validate correctness.

## 2. Manual smoke test against a mock hub

Start a minimal mock hub (any HTTP server returning the shape below on
`GET /api/v2/outputs/{outputId}`) on `localhost:9090`, then:

```bash
./sonora outputs get office-speaker --hub-url http://localhost:9090
```

**Expected**: YAML output for that single output, showing
`outputId`/`displayName`/`volume`/`muted`/`available`/`enabled` — matches
[contracts/cli-outputs-get.md](contracts/cli-outputs-get.md). Try this against both an
enabled and a disabled output on the mock hub — both must be returned (unlike `outputs
list`'s default filtering, FR-003).

```bash
./sonora outputs get office-speaker --hub-url http://localhost:9090 --json
```

**Expected**: the same data as strict JSON; pipe through `jq .` (or any JSON parser) to
confirm it's valid (SC-004).

## 3. Not-found smoke test

```bash
./sonora outputs get does-not-exist --hub-url http://localhost:9090
echo $?   # expect 5
```

**Expected**: a clear "output not found" message on stderr naming the identifier — not an
empty result, not a generic hub-error message (SC-002).

## 4. Manual failure-path smoke tests

```bash
./sonora outputs get office-speaker --hub-url http://127.0.0.1:1   # nothing listening
echo $?   # expect 4
```

```bash
./sonora outputs get office-speaker --hub-url http://10.255.255.1:8080  # unroutable, exercises the 5s timeout
echo $?   # expect 4, and the command must return within ~5s, not hang (SC-003)
```

```bash
./sonora outputs get
echo $?   # expect 2 — missing identifier
```

```bash
./sonora outputs get office-speaker --unknown-flag
echo $?   # expect 2
```

**Expected in all four**: a clear one-line message on stderr (no raw Go stack trace or
panic), and with `--verbose` added, the underlying error detail additionally printed.

## Success criteria mapping

| Success Criteria | Validated by |
|---|---|
| SC-001 (single output retrieved <1s) | Step 2, timed manually or via `time ./sonora ...`. |
| SC-002 (not-found always clear, never empty/crash) | Step 3. |
| SC-003 (unreachable/unresponsive hub fails within 5s) | Step 4, second command. |
| SC-004 (`--json` output strictly parseable) | Step 2, second command, piped through a JSON parser. |
| SC-005 (volume/mute/availability visible from one invocation) | Step 2, first command. |
