# Sonora Multiroom CLI

A fast, scriptable command-line client for the [Multiroom Audio Hub API](api/openapi.json) —
control real-time audio routing, volume, and mute across your speakers from the terminal.

```
sonora outputs list
```

## Status

Early development. The project follows a spec-first workflow: features are specified,
planned, and implemented one at a time — see [`specs/`](specs/) for what's in progress.

## Design principles

- **Instant startup** — no perceptible lag between launch and the first request.
- **API-contract fidelity** — every request/response is validated against `api/openapi.json`.
- **Resilient HTTP client** — explicit timeouts, no hangs, clear errors on failure.
- **Consistent UX** — YAML output by default, `--json` for scripting, predictable exit codes.

Full rationale lives in the [project constitution](.specify/memory/constitution.md).

## License

[GNU Affero General Public License v3.0](LICENSE)
