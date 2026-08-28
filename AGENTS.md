# Agent Instructions

## Product naming

- The product is **Sonora Multiroom**. This CLI's full name is **Sonora Multiroom CLI**.
- "Sonora CLI" is informal shorthand — do not use it in specs, docs, or headings; use
  "Sonora Multiroom CLI" (or "Sonora Multiroom" when referring to the product/hub as a
  whole) instead.
- The invoked binary/command name is `sonora` (e.g. `sonora get outputs`) and is
  unaffected by this — only prose naming changes.
- The backend service this CLI talks to is the **Multiroom Audio Hub API**
  (`api/openapi.json`), a separate component from the CLI itself.

## CLI flag conventions

- Strict GNU/POSIX style (like `kubectl`/`gh`): long flags require `--` (e.g.
  `--version`, `--help`), short flags are single-letter with `-` (e.g. `-v`, `-h`).
  Single-dash multi-letter forms (`-version`, `-help`) are invalid and must not be
  added.

## Git / PR conventions

- Commit messages follow [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/)
  — see [CONTRIBUTING.md](CONTRIBUTING.md) for the types this repo uses.
- Never add a `Co-Authored-By` trailer (or similar attribution) to git commit messages.
- Never add "Generated with Claude Code" (or similar attribution) to pull request
  descriptions.

## Project context

- Governance and non-negotiable engineering principles live in
  `.specify/memory/constitution.md` — read it before planning or implementing.
- API contract fidelity: all HTTP request/response handling MUST be validated against
  `api/openapi.json`, the single source of truth for the hub API.
- Feature work follows the Spec Kit workflow (`specs/<NNN-feature>/spec.md` →
  `plan.md` → `tasks.md`), driven by the `/speckit-*` skills.
