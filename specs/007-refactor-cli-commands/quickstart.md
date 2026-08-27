# Quickstart: Adopt Verb-First Command Landscape

**Feature**: `007-refactor-cli-commands` | **Date**: 2026-08-27

Validates that the new verb-first grammar (`get`/`list` + resource path, and `play`'s new
target path) works end-to-end and that the old grammar is genuinely gone, not silently kept
alive. Uses the same mock-hub approach as prior features.

## Prerequisites

- Go toolchain installed.
- Repo built: `go build -o sonora ./cmd/sonora` from repo root.

## 1. Run the automated test suite (primary validation)

```bash
go test ./...
```

Expected: `tests/unit` (including the new `internal/cli/respath` tests), `tests/contract`,
and `tests/integration` all pass, updated for the new invocation shape. Per-resource HTTP
contract tests are unchanged in substance — only the CLI invocation used to trigger them
changes.

## 2. New grammar — collection and single-item forms

```bash
./sonora get outputs --hub-url http://localhost:9090
./sonora list outputs --hub-url http://localhost:9090
```

**Expected**: identical output, YAML, both equivalent to today's `sonora outputs list`.

```bash
./sonora get outputs/office-speaker --hub-url http://localhost:9090
./sonora get out/office-speaker --hub-url http://localhost:9090
```

**Expected**: identical output, both equivalent to today's `sonora outputs get
office-speaker` (full name and alias interchangeable, FR-004).

Repeat with `inputs`/`in`, `groups`/`gr`, `routes`/`rt` — including `routes`'
`--status`/`--input-id`/`--target-id` filters, which must behave exactly as they do today.

## 3. `list` rejects an id

```bash
./sonora list outputs/office-speaker --hub-url http://localhost:9090
echo $?   # expect 2
```

**Expected**: a clear usage error — `list` is collection-only (FR-003).

## 4. Old grammar is gone

```bash
./sonora outputs list --hub-url http://localhost:9090
echo $?   # expect 2
./sonora outputs get office-speaker --hub-url http://localhost:9090
echo $?   # expect 2
```

**Expected**: a standard "unknown command" usage error, not a migration hint and not the old
behavior (spec Clarifications, hard cutover with no tailored detection).

## 5. `play`'s new target path

```bash
./sonora play "https://stream.example.com/live.mp3" outputs/office-speaker --display-name "Radio" --hub-url http://localhost:9090
```

**Expected**: playback starts, same result shape as today's
`sonora play <uri> office-speaker --output --name "Radio"`.

```bash
./sonora play "https://stream.example.com/live.mp3" office-speaker --output --hub-url http://localhost:9090
echo $?   # expect 2 — --output is no longer a defined flag
```

## 6. Malformed resource path

```bash
./sonora get out/foo/bar --hub-url http://localhost:9090
echo $?   # expect 2
./sonora get bogus --hub-url http://localhost:9090
echo $?   # expect 2
```

**Expected**: clear usage errors — an id with an extra `/`, and an unrecognized resource
name, are both rejected before any hub request is made.

## Success criteria mapping

| Success Criteria | Validated by |
|---|---|
| SC-001 (only new grammar reachable) | Steps 2–3 (works) and Step 4 (old form fails). |
| SC-002 (aliases identical to full names) | Step 2 (`out`/`gr`/`in`/`rt` variants). |
| SC-003 (pre-refactor forms fail clearly) | Steps 4–6. |
| SC-004 (displayed data unchanged) | Step 2, diffed against pre-refactor output for the same mock-hub fixture. |
