# Quickstart: List and Get Audio Routes

**Feature**: `004-routes-list-get` | **Date**: 2026-08-25

Validates that `sonora routes list` and `sonora routes get <route-id>` work end-to-end.
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
(`tests/integration`) tests pass, including the new `routes list`/`routes get` tests
alongside the existing `outputs`/`inputs` ones. The contract and integration tests spin up an
`httptest.Server` shaped from `api/openapi.json`'s `listRoutes`/`getRoute` operations — no
external hub needed to validate correctness.

## 2. Manual smoke test: list

Start a minimal mock hub (any HTTP server returning the shape below on `GET
/api/v2/routes`) on `localhost:9090`, then:

```bash
./sonora routes list --hub-url http://localhost:9090
```

**Expected**: YAML output listing every route regardless of status, showing
`routeId`/`inputId`/`targetId`/`targetType`/`status` per route — matches
[contracts/cli-routes-list.md](contracts/cli-routes-list.md).

```bash
./sonora routes list --hub-url http://localhost:9090 --status FAILED
./sonora routes list --hub-url http://localhost:9090 --input-id spotify-1
./sonora routes list --hub-url http://localhost:9090 --target-id kitchen-speaker
./sonora routes list --hub-url http://localhost:9090 --status ACTIVE --target-id kitchen-speaker
```

**Expected**: each command returns only the routes matching the supplied filter(s); the last
command (two filters) returns only routes matching both (AND logic, FR-003).

```bash
./sonora routes list --hub-url http://localhost:9090 --json | jq .
```

**Expected**: valid JSON (SC-005), same fields as the YAML view.

## 3. Manual smoke test: get

```bash
./sonora routes get route-abc-123 --hub-url http://localhost:9090
```

**Expected**: YAML output for that single route, showing all ten fields — matches
[contracts/cli-routes-get.md](contracts/cli-routes-get.md). Try this against both a route
whose playback has started (`startedAt` populated) and one that hasn't (`startedAt: null`) to
confirm both display correctly (edge case in spec.md).

```bash
./sonora routes get route-abc-123 --hub-url http://localhost:9090 --json | jq .
```

**Expected**: the same data as strict JSON.

## 4. Not-found smoke test

```bash
./sonora routes get does-not-exist --hub-url http://localhost:9090
echo $?   # expect 5
```

**Expected**: a clear "route not found" message on stderr naming the identifier — not an
empty result, not a generic hub-error message (SC-003).

## 5. Manual failure-path smoke tests

```bash
./sonora routes list --hub-url http://127.0.0.1:1   # nothing listening
echo $?   # expect 4
```

```bash
./sonora routes get route-abc-123 --hub-url http://10.255.255.1:8080  # unroutable, exercises the 5s timeout
echo $?   # expect 4, and the command must return within ~5s, not hang (SC-004)
```

```bash
./sonora routes get
echo $?   # expect 2 — missing identifier
```

```bash
./sonora routes list --unknown-flag
echo $?   # expect 2
```

```bash
./sonora routes list --hub-url http://localhost:9090 --status NOT_A_REAL_STATUS
echo $?   # expect 3 — hub rejects the invalid status filter with 400
```

**Expected in all five**: a clear one-line message on stderr (no raw Go stack trace or
panic), and with `--verbose` added, the underlying error detail additionally printed.

## Success criteria mapping

| Success Criteria | Validated by |
|---|---|
| SC-001 (list retrieved <1s) | Step 2, timed manually or via `time ./sonora ...`. |
| SC-002 (get retrieved <1s) | Step 3, timed manually. |
| SC-003 (not-found always clear, never empty/crash) | Step 4. |
| SC-004 (unreachable/unresponsive hub fails within 5s, both commands) | Step 5, first two commands. |
| SC-005 (`--json` output strictly parseable, both commands) | Steps 2 and 3, JSON invocations piped through `jq`. |
| SC-006 (zero routes vs. failure always distinguishable) | Step 2, first command, run against a mock hub returning an empty list. |
| SC-007 (source input, target, and status visible from one invocation) | Step 3, first command. |
| SC-008 (narrow by status/input/target without post-filtering) | Step 2, filter commands. |
