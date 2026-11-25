# Quickstart: List and Get Output Groups

**Feature**: `005-groups-list-get` | **Date**: 2026-08-26

Validates that `sonora groups list` and `sonora groups get <group-id>` work end-to-end.
Uses a local mock hub for deterministic validation; the same steps apply against a real
Multiroom Audio Hub by pointing `--hub-url` at it once one is available.

## Prerequisites

- Go toolchain installed (repo's `go.mod` already exists, from `001-list-outputs`).
- Repo built: `go build -o sonora ./cmd/sonora` from repo root.

## 1. Run the automated test suite (primary validation)

```bash
go test ./...
```

Expected: all unit (`tests/unit`), contract (`tests/contract`), and integration
(`tests/integration`) tests pass, including the new `groups list`/`groups get` tests
alongside the existing `outputs`/`inputs`/`routes` ones. The contract and integration tests
spin up an `httptest.Server` shaped from `api/openapi.json`'s `listGroups`/`getGroup`
operations — no external hub needed to validate correctness.

## 2. Manual smoke test: list

Start a minimal mock hub (any HTTP server returning the shape below on `GET
/api/v2/groups`) on `localhost:9090`, then:

```bash
./sonora groups list --hub-url http://localhost:9090
```

**Expected**: YAML output listing only enabled groups, showing
`groupId`/`displayName`/`outputIds`/`muted`/`enabled` per group — matches
[contracts/cli-groups-list.md](contracts/cli-groups-list.md).

```bash
./sonora groups list --hub-url http://localhost:9090 --include-disabled
```

**Expected**: both enabled and disabled groups are returned, each showing its `enabled`
state (User Story 2).

```bash
./sonora groups list --hub-url http://localhost:9090 --json | jq .
```

**Expected**: valid JSON (SC-005), same fields as the YAML view.

## 3. Manual smoke test: get

```bash
./sonora groups get living-room --hub-url http://localhost:9090
```

**Expected**: YAML output for that single group, showing all five fields — matches
[contracts/cli-groups-get.md](contracts/cli-groups-get.md). Try this against both a group
with member outputs and one with zero member outputs to confirm `outputIds: []` displays
correctly (edge case in spec.md).

```bash
./sonora groups get living-room --hub-url http://localhost:9090 --json | jq .
```

**Expected**: the same data as strict JSON.

## 4. Not-found smoke test

```bash
./sonora groups get does-not-exist --hub-url http://localhost:9090
echo $?   # expect 5
```

**Expected**: a clear "group not found" message on stderr naming the identifier — not an
empty result, not a generic hub-error message (SC-003).

## 5. Manual failure-path smoke tests

```bash
./sonora groups list --hub-url http://127.0.0.1:1   # nothing listening
echo $?   # expect 4
```

```bash
./sonora groups get living-room --hub-url http://10.255.255.1:8080  # unroutable, exercises the 5s timeout
echo $?   # expect 4, and the command must return within ~5s, not hang (SC-004)
```

```bash
./sonora groups get
echo $?   # expect 2 — missing identifier
```

```bash
./sonora groups list --unknown-flag
echo $?   # expect 2
```

**Expected in all four**: a clear one-line message on stderr (no raw Go stack trace or
panic), and with `--verbose` added, the underlying error detail additionally printed.

## Success criteria mapping

| Success Criteria | Validated by |
|---|---|
| SC-001 (list retrieved <1s) | Step 2, timed manually or via `time ./sonora ...`. |
| SC-002 (get retrieved <1s) | Step 3, timed manually. |
| SC-003 (not-found always clear, never empty/crash) | Step 4. |
| SC-004 (unreachable/unresponsive hub fails within 5s, both commands) | Step 5, first two commands. |
| SC-005 (`--json` output strictly parseable, both commands) | Steps 2 and 3, JSON invocations piped through `jq`. |
| SC-006 (zero groups vs. failure always distinguishable) | Step 2, first command, run against a mock hub returning an empty list. |
| SC-007 (member outputs and mute state visible from one invocation) | Step 3, first command. |
| SC-008 (view disabled groups without a separate command) | Step 2, second command. |
