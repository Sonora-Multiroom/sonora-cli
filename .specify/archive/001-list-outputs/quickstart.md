# Quickstart: List Audio Outputs

**Feature**: `001-list-outputs` | **Date**: 2026-08-24

Validates that `sonora outputs list` works end-to-end. Uses a local mock hub for
deterministic validation; the same steps apply against a real Multiroom Audio Hub by
omitting `--hub-url` (or pointing it at the real hub) once one is available.

## Prerequisites

- Go toolchain installed (see `go.mod` at repo root once created).
- Repo built: `go build -o sonora ./cmd/sonora` from repo root.

## 1. Run the automated test suite (primary validation)

```bash
go test ./...
```

Expected: all unit (`tests/unit`), contract (`tests/contract`), and integration
(`tests/integration`) tests pass. The contract and integration tests spin up an
`httptest.Server` shaped from `api/openapi.json`'s `OutputResponse`/`GET /api/v2/outputs`
definitions — no external hub needed to validate correctness.

## 2. Manual smoke test against a mock hub

Start a minimal mock hub (any HTTP server returning the shape below on
`GET /api/v2/outputs`) on `localhost:9090`, then:

```bash
./sonora outputs list --hub-url http://localhost:9090
```

**Expected**: YAML output listing only enabled outputs, each showing
`outputId`/`displayName`/`volume`/`muted`/`available`/`enabled`, with any output that has
`available: false` clearly shown as such — matches [contracts/cli-outputs-list.md](contracts/cli-outputs-list.md).

```bash
./sonora outputs list --hub-url http://localhost:9090 --include-disabled
```

**Expected**: disabled outputs additionally appear, each showing `enabled: false`.

```bash
./sonora outputs list --hub-url http://localhost:9090 --json
```

**Expected**: the same data as strict JSON; pipe through `jq .` (or any JSON parser) to
confirm it's valid (SC-003).

## 3. Manual failure-path smoke tests

```bash
./sonora outputs list --hub-url http://localhost:1        # nothing listening
echo $?   # expect 4
```

```bash
./sonora outputs list --hub-url http://10.255.255.1:8080  # unroutable, exercises the 5s timeout
echo $?   # expect 4, and the command must return within ~5s, not hang
```

```bash
./sonora outputs list --unknown-flag
echo $?   # expect 2
```

**Expected in all three**: a clear one-line message on stderr (no raw Go stack trace or
panic), and with `--verbose` added, the underlying error detail additionally printed.

## 4. Zero-outputs case

Point `--hub-url` at a mock hub configured to return `[]` from `GET /api/v2/outputs`.

**Expected**: exit code `0`, and an unambiguous "no outputs" indication in the output —
not an empty/blank stdout that could be mistaken for a hang or a bug (SC-004).

## Success criteria mapping

| Success Criteria | Validated by |
|---|---|
| SC-001 (list renders <1s) | Step 2, timed manually or via `time ./sonora ...`. |
| SC-002 (failure surfaced within 5s) | Step 3, second command. |
| SC-003 (JSON strictly parseable) | Step 2, third command, piped through a JSON parser. |
| SC-004 (no-outputs vs. failure unambiguous) | Step 4. |
| SC-005 (volume/mute visible without another command) | Step 2, first command. |
