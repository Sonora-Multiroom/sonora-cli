# CLI Contract: `speak`, `get tts-cache`, `clear tts-cache`

**Feature**: `009-tts-commands` | **Data model**: [../data-model.md](../data-model.md)

This contract binds the user-facing interface: command shapes, output, stderr messages and
exit codes. Hub-side request and response shapes are in [data-model.md](../data-model.md) and
`api/openapi.json`.

Common to all three commands:

- Flags `--json`, `--hub-url URL` and `--verbose` behave as on every existing command
  (FR-009). `--help`/`-h` prints the command's usage and flags to stdout and exits 0.
- Flags may appear before, between or after positional arguments. `--` ends flag parsing.
- Results go to stdout. Every failure goes to stderr only, as `error: <message> (hub URL:
  <url>)`, followed by `detail: <underlying error>` when `--verbose` is set.
- Usage errors print the usage line and then `error: <reason>` to stderr, exit 2, and send
  no request.

---

## `sonora speak <text> <outputs|groups>/<id> [flags]`

```text
usage: sonora speak <text|-> <outputs|groups>/<id> [flags]

Flags:
  --hub-url URL      hub base URL override
  --json             emit strict JSON instead of the default YAML
  --language TAG     BCP 47 language tag override
  --provider NAME    TTS provider override
  --verbose          print the underlying error detail on failure
  --voice ID         provider-specific voice override
```

Requests: exactly one `POST /api/tts/speak`, bounded at 15 s. After a 404, exactly one
additional `GET /api/v2/extensions`, bounded at 5 s.

### Arguments

| Input | Result |
| --- | --- |
| `<text>` is `-` | Text is read from standard input to end of input, and trailing CR/LF are trimmed. |
| `<text>` empty or whitespace-only (after the trim when `-`) | `error: text must not be empty` → 2 |
| Standard input unreadable | `error: could not read text from standard input: <err>` → 2 |
| Target `outputs/<id>` / `out/<id>` | `targetType: SINGLE_OUTPUT` |
| Target `groups/<id>` / `gr/<id>` | `targetType: OUTPUT_GROUP` |
| Target of another kind (`inputs/x`, `routes/x`) | `error: speak target must be outputs/<id> or groups/<id>, got "<arg>"` → 2 |
| Target with no id (`outputs`) | `error: missing required argument: <target-path> must include an id` → 2 |
| Unparseable target | `sonora: <respath error>` → 2 |
| 0 / 1 / 3+ positionals | `error: missing required argument: <text>` / `… <target-path>` / `error: unexpected argument(s): [...]` → 2 |
| `--provider`, `--voice` or `--language` given an empty value | `error: --<flag> must not be empty` → 2 |

To speak text that starts with a dash, put it after `--`:
`sonora speak -- "-5 degrees outside" outputs/porch`. A lone `-` always means standard input.

### Success (202)

YAML (default):

```yaml
announcementId: "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
cacheHit: true
queueDepth: 1
```

JSON (`--json`), one line:

```json
{"announcementId":"a1b2c3d4-e5f6-7890-abcd-ef1234567890","cacheHit":true,"queueDepth":1}
```

The command returns as soon as the 202 arrives and never waits for playback (FR-005).

---

## `sonora get tts-cache [flags]`

```text
usage: sonora get tts-cache [flags]
```

Flags: `--hub-url`, `--json`, `--verbose`. Any positional argument → 2.
`sonora list tts-cache` prints `error: list does not support tts-cache; use 'sonora get
tts-cache' instead` → 2. `tts-cache/<anything>` is not recognised as a resource → 2.

Requests: one `GET /api/tts/cache/stats` (5 s), plus one inventory call only after a 404.

YAML:

```yaml
totalEntries: 42
totalSizeBytes: 15728640
maxSizeBytes: 524288000
entriesByProvider:
  "openai": 35
  "piper-local": 7
```

Empty cache:

```yaml
totalEntries: 0
totalSizeBytes: 0
maxSizeBytes: 524288000
entriesByProvider: {}
```

JSON: `{"totalEntries":42,"totalSizeBytes":15728640,"maxSizeBytes":524288000,"entriesByProvider":{"openai":35,"piper-local":7}}`.
`entriesByProvider` is always an object, `{}` when empty.

---

## `sonora clear tts-cache [--provider NAME] [flags]`

```text
usage: sonora clear tts-cache [flags]

Flags:
  --hub-url URL      hub base URL override
  --json             emit strict JSON instead of the default YAML
  --provider NAME    clear only this provider's cache entries
  --verbose          print the underlying error detail on failure
```

Requests: one `DELETE /api/tts/cache`, with `?providerName=<name>` only when `--provider` is
given (5 s). The command does not prompt for confirmation.

| Invocation | YAML | JSON |
| --- | --- | --- |
| `clear tts-cache` | `cleared: all` | `{"cleared":"all"}` |
| `clear tts-cache --provider openai` | `cleared: provider`<br>`provider: "openai"` | `{"cleared":"provider","provider":"openai"}` |

`--provider ""` → 2. Clearing an already-empty cache succeeds (FR-008).

`clear` with no argument, or with any resource other than `tts-cache`:
`usage: sonora clear tts-cache [flags]` then `error: clear supports only tts-cache` → 2.

---

## Failure messages (all three commands)

`<code>` is the hub's machine-readable code and `<msg>` is the hub's message, printed
verbatim on one line.

| Condition | stderr `error:` message | Exit |
| --- | --- | --- |
| TTS error with a code | `<code>: <msg>` | per code (below) |
| `TARGET_NOT_FOUND` | `TARGET_NOT_FOUND: <msg>` | 12 |
| `INVALID_REQUEST` | `INVALID_REQUEST: <msg>` | 6 |
| `PROVIDER_NOT_FOUND` | `PROVIDER_NOT_FOUND: <msg>` | 5 |
| `PROVIDER_TIMEOUT` / `PROVIDER_RATE_LIMITED` / `PROVIDER_ERROR` / `FORMAT_NORMALIZATION_FAILED` | `<code>: <msg>` | 10 |
| Unrecognised code | `<code>: <msg>` | 6 on 400, 10 on 503, 3 otherwise |
| Error body not TTS-shaped | `hub rejected the request (HTTP <status>): <detail or title>`, or `hub reported an error (HTTP <status>)` when there's nothing to show | 6 on 400, 10 on 503, 3 otherwise |
| 404, `tts` not in the inventory | `text-to-speech is not available on this hub: the TTS extension is not installed on this hub` | 13 |
| 404, not in the inventory, `loadingEnabled: false` | `… not installed on this hub (extension loading is switched off in the hub's configuration)` | 13 |
| 404, `DISABLED` | `text-to-speech is not available on this hub: the TTS extension is installed but disabled in the hub's configuration` | 13 |
| 404, `REJECTED` | `text-to-speech is not available on this hub: the TTS extension failed to load: <rejectionReason>` (reason omitted when null or empty) | 13 |
| 404, `INERT` | `text-to-speech is not available on this hub: the TTS extension is loaded but inactive` | 13 |
| 404, `ACTIVE` | `text-to-speech is not available on this hub: the hub's TTS API does not match this CLI version` | 13 |
| 404, inventory lookup fails any other way, or an unknown status | `text-to-speech is not available on this hub` (lookup error only under `--verbose`) | 13 |
| 404, inventory also 404 | `<url> is not serving the Multiroom Audio Hub API: the hub URL is wrong, or the hub's control API (REST) extension is not installed or not loaded; set the correct address with --hub-url, MULTIROOM_URL, or the config file` | 4 |
| Malformed success body (FR-005a) | `hub returned an unexpected or malformed response` | 3 |
| Hub unreachable | `could not reach the hub` | 4 |
| No response within the bound (5 s; 15 s for `speak`) | `hub did not respond in time` | 4 |

The CLI never retries a TTS request (FR-014).
