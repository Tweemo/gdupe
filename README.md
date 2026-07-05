# media-merge

A CLI tool that finds exact duplicate files in a directory. By default it
only reports — nothing on disk changes unless you explicitly ask.

> **Status:** learning project. The CLI shell, package layout, and a full
> test suite are in place; the logic in `internal/dedup` is intentionally
> unimplemented so it can be written by hand, test by test.

## Install

```bash
go install github.com/timl/media-merge/cmd/media-merge@latest
```

Or from a checkout:

```bash
go install ./cmd/media-merge
```

## Usage

```bash
media-merge ~/Pictures            # dry-run: report duplicates + wasted space
media-merge -delete ~/Pictures    # delete duplicate copies (keeps one of each)
media-merge -move ~/Pictures      # move duplicates into ~/Pictures/duplicates/
```

`-delete` and `-move` are mutually exclusive. A dry-run prints something like:

```
37 duplicate file(s), 412.6 MB wasted
```

## How it works

1. **Scan** — walk the directory tree and collect every regular file
   (`internal/dedup/scan.go`).
2. **Hash** — SHA-256 each file, streaming so large media files don't
   blow up memory (`internal/dedup/hash.go`).
3. **Group** — files with identical hashes form a group: one keeper,
   the rest are duplicates (`internal/dedup/dedup.go`).
4. **Apply** — only with a flag: delete the duplicates, or move them to a
   `duplicates/` subfolder (`internal/dedup/apply.go`).

## Working on it

The tests describe the intended behavior and currently fail on the
unimplemented stubs — implement until they pass:

```bash
go test ./...
```

Suggested order: `Scan` → `HashFile` → `FindDuplicates` →
`WastedBytes`/`FormatSize` → `Delete` → `Move`. Each stub's comment lists
hints and edge cases (symlinks, name collisions on move, error strategy).

Stuck? Open **[HINTS.md](HINTS.md)** — a cheatsheet with three levels of
folded hints per function, from gentle nudge to near-solution. And
`CLAUDE.md` instructs Claude to answer implementation questions with hints
rather than full solutions, so asking for help won't spoil the exercise.

## Roadmap

1. **Exact dedup** — the current skeleton: hash, group, report, delete/move.
2. **Better reporting** — per-group listings, largest-first sorting, a
   `-verbose` flag.
3. **Similar-image grouping** — perceptual hashing to catch resized or
   re-encoded copies of the same photo. A previous version of this project
   implemented this as a web app; that code (perceptual hashing with
   `goimagehash`, clustering, thumbnailing) is preserved in git history:
   `git show <first-commit>:server/internal/media/`.
