# media-merge

A CLI tool that finds exact duplicate files in a directory. By default it
only reports — nothing on disk changes unless you explicitly ask.

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
media-merge -delete ~/Pictures    # move duplicate copies to the system trash (keeps one of each)
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
4. **Apply** — only with a flag: trash the duplicates, or move them to a
   `duplicates/` subfolder (`internal/dedup/apply.go`).

## Working on it

This started as a learning project: the package layout, CLI shell, and test
suite were scaffolded first, and the logic in `internal/dedup` was then
written by hand, test by test. The tests describe the intended behavior:

```bash
go test ./...
```

**[HINTS.md](HINTS.md)** remains as a cheatsheet from that phase — three
levels of folded hints per function, from gentle nudge to near-solution.
`CLAUDE.md` instructs Claude to answer implementation questions with hints
rather than full solutions, which still applies to roadmap work.

## Roadmap

Done: exact dedup (hash, group, report, move) with a file-size pre-filter
so files with a unique size are never hashed, and trash-instead-of-delete —
`-delete` moves duplicates to the OS trash (recoverable) rather than
unlinking them permanently (macOS `~/.Trash` and Linux XDG trash only;
same-filesystem moves only).

1. **Concurrent hashing** — hash files in parallel with a worker pool
   (goroutines, channels, `-race`-clean map access).
2. **Hash cache across runs** — persist `path, size, mtime → hash` so
   re-scans skip unchanged files.
3. **Partial-hash tier** — hash the first 64KB first; full-hash only files
   whose prefix also matches, so differing large files are read once, cheaply.
4. **Better reporting** — per-group listings, largest-first sorting, a
   `-verbose` flag.
5. **Similar-image grouping** — perceptual hashing to catch resized or
   re-encoded copies of the same photo. A previous version of this project
   implemented this as a web app; that code (perceptual hashing with
   `goimagehash`, clustering, thumbnailing) is preserved in git history:
   `git show <first-commit>:server/internal/media/`.
6. **Windows Recycle Bin support** — the Recycle Bin isn't a plain folder;
   needs a shell API wrapper or a cross-platform trash library.
