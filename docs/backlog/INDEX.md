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
| P3 | [Adopt HUB-040: a chime before a route](adopt-route-chime.md) | API adoption | Small | The hub can start a route with a short sound, but no client can ask for one or list the chimes. `sonora-mcp` follows this item |
