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
| P3 | [Support join modes: list the routes on an output and choose how a route joins](join-modes-in-cli.md) | Feature | Small to medium | The hub has supported several routes per output since 0.1.21, but the CLI can only replace what is playing and cannot show what an output plays; nothing is broken. An output's routes are shown with `get outputs/<id> --include-routes` |
