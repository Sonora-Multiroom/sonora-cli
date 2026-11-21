# Remove the deprecated `TTSDiagnosis` alias

**Status:** Proposed, not started. Written 2026-11-21.

**Origin:** CLI-014 (radio station search). The diagnosis type was renamed `ExtensionDiagnosis` because
radio uses it too, and `hub.TTSDiagnosis` stayed as a deprecated alias so existing callers kept compiling.

**Depends on:** sonora-mcp's backlog item `adopt-radio-search.md` (it names only the `Diagnosis*` constants
and the `Diagnosis` field today, and must use `hub.ExtensionDiagnosis` in anything new), and a release of
this repository that contains `ExtensionDiagnosis`.

**Effort:** Small.

## Problem

`hub.TTSDiagnosis` is now a second name for `hub.ExtensionDiagnosis`. It keeps a name that says "TTS" on a
type that also explains radio failures.

## Proposal

- Delete the `TTSDiagnosis` alias from `hub/tts.go`.
- Do it only after sonora-mcp has moved to a release with `ExtensionDiagnosis` and no consumer names the old
  type. Check with a search of the Sonora repositories for `TTSDiagnosis`.
- It is a breaking change in a public package: name it in the release notes (the module is `v0.x`, where
  that is allowed).

## Open questions

1. Is there any consumer outside the Sonora repositories? None is known.
