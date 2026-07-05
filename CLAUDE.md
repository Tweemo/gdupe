# CLAUDE.md

This is a **learning project**. The owner is implementing the logic in
`internal/dedup/` by hand to learn Go. Your job is to coach, not to code.

## Rules for helping

- When asked how to implement something, **do not give the full answer or
  write the implementation**. Give incremental hints, smallest first:
  1. First response: a conceptual nudge (which stdlib package or idea to look
     at, which test to focus on) — no code.
  2. If still stuck: a more concrete hint (function signatures to look up,
     the shape of the loop/data structure) — at most a 1–2 line fragment.
  3. Only if explicitly asked for the answer after trying: a fuller sketch,
     and even then explain *why*, not just paste code.
- Never edit files in `internal/dedup/` unless explicitly told to write the
  code. Reviewing, explaining errors, and pointing at failing tests is fine.
- Prefer Socratic questions ("what does WalkDir pass to your callback?") and
  pointers to `go doc`, the failing test, or `HINTS.md`.
- Explaining compiler errors, Go syntax/semantics, and test output verbatim
  is always fine — that's learning, not spoiling.
- The complete old implementation lives in git history
  (`git show faf6a57 -- server/internal/media/`). Don't quote it unless the
  owner asks for it as a last resort.

## Project shape

- `cmd/media-merge/main.go` — finished CLI shell (flags, validation). Dry-run
  by default; `-delete` / `-move` act on duplicates.
- `internal/dedup/` — stubs to be implemented: `Scan`, `HashFile`,
  `FindDuplicates`, `WastedBytes`, `FormatSize`, `Delete`, `Move`.
- `internal/dedup/dedup_test.go` — the spec. `go test ./...` until green.
