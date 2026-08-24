# Agent Instructions

## Product naming

- The product is **Sonora Multiroom**. This CLI's full name is **Sonora Multiroom CLI**.
- "Sonora CLI" is informal shorthand — do not use it in specs, docs, or headings; use
  "Sonora Multiroom CLI" (or "Sonora Multiroom" when referring to the product/hub as a
  whole) instead.
- The invoked binary/command name is `sonora` (e.g. `sonora outputs list`) and is
  unaffected by this — only prose naming changes.
- The backend service this CLI talks to is the **Multiroom Audio Hub API**
  (`api/openapi.json`), a separate component from the CLI itself.

## Project context

- Governance and non-negotiable engineering principles live in
  `.specify/memory/constitution.md` — read it before planning or implementing.
- API contract fidelity: all HTTP request/response handling MUST be validated against
  `api/openapi.json`, the single source of truth for the hub API.
- Feature work follows the Spec Kit workflow (`specs/<NNN-feature>/spec.md` →
  `plan.md` → `tasks.md`), driven by the `/speckit-*` skills.
