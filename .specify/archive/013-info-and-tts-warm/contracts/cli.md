# CLI contract

## `sonora get info [--json] [--verbose] [--hub-url URL]`

Takes no arguments. `list info` is a usage error pointing at `get info`, as for `master-mute`.

YAML (default); `startedAt` and `uptimeSeconds` are left out when the hub did not send them:

```yaml
product: "multiroom"
version: "0.1.33"
apiVersion: "0.1.33"
startedAt: "2026-11-06T10:06:59.512Z"
uptimeSeconds: 16981
```

`--json`: one object with the same five property names.

Failures (stderr, `error: <message> (hub URL: <url>)`):

| Case | Message | Exit |
|------|---------|------|
| hub older than 0.1.33 | `this hub does not offer hub info (needs hub 0.1.33 or later)` | hub error |
| unreachable or timeout | the usual network message | network |
| malformed body | `hub returned an unexpected or malformed response` | hub error |
| extra argument | usage | usage |

## `sonora warm tts-cache <text|-> [--provider NAME] [--voice ID] [--language TAG] [--timeout DURATION] [--json] [--verbose] [--hub-url URL]`

`<text>` is `-` to read standard input (trailing line breaks trimmed, as `speak`). Flags may come before or
after the text. Plays nothing.

YAML: `cacheHit: false` (just synthesized) or `cacheHit: true` (already stored). `--json`: `{"cacheHit":false}`.

| Case | Message | Exit |
|------|---------|------|
| no text, or empty or blank text | `error: text must not be empty` (or missing argument) | usage |
| an empty `--provider`, `--voice`, `--language`; bad `--timeout` | as `speak` | usage |
| hub 400 | the hub's message | validation, as `speak` |
| unknown provider | the hub's message | not found, as `speak` |
| TTS extension absent | the diagnosed "text-to-speech is not available" message | as `speak` |
| 503 | the hub's message, then `; nothing was stored, repeating the request is safe` | service unavailable |
| unreachable or timeout | the usual network message | network |

Help text lists both commands in the top-level `help`, with an example for each.
