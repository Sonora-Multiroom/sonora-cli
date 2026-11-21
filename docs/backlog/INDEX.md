# Backlog index

Every item in `docs/backlog/`, highest priority first. Each one is a proposal, not a requirement. When
you add, ship or delete an item, update this list.

| Priority | Meaning |
|---|---|
| **P1** | A crash, data loss or other user-visible fault. Do it next. |
| **P2** | Wrong or confusing behaviour, or something users need now. |
| **P3** | A worthwhile feature or cleanup with no pressure on it. |
| **P4** | Only worth doing once a concrete need appears. |

| Priority | Item | Kind | Effort | Why this priority |
|---|---|---|---|---|
| P3 | [Radio: search and browse stations](radio-search.md) | Feature | Medium | The hub's `radio` extension now has search and browse (RADIO-002, `sonora-radio` 0.2.0, verified live on the Pi; PR #2 pending merge), but the CLI cannot find a station UUID to play. `sonora-mcp` and `sonora-mobile` follow this item |
