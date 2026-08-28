# Contributing

## Commit messages

This project follows [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/):

```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

Types used in this repo's history:

| Type | Use for |
| --- | --- |
| `feat` | A new user-facing capability (e.g. a new command or flag) |
| `fix` | A bug fix |
| `refactor` | A code change that neither fixes a bug nor adds a feature |
| `docs` | Documentation-only changes (README, AGENTS.md, etc.) |
| `spec` | Spec Kit artifacts (`specs/**`) with no corresponding code change |
| `ci` | Build, release, or CI/workflow changes |
| `chore` | Routine maintenance with no user-facing or CI effect (deps, tooling config, etc.) |

Breaking changes: append `!` after the type/scope (`feat!:`) and/or add a
`BREAKING CHANGE:` footer, per the spec.

Keep the description short and in the imperative mood ("add", not "added"/"adds"). Do not
add a `Co-Authored-By` trailer — see [AGENTS.md](AGENTS.md).

## Pull requests

Do not add "Generated with Claude Code" or similar tool-attribution text to PR
descriptions — see [AGENTS.md](AGENTS.md).
