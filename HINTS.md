# Hints cheatsheet

Incremental hints for each function, mildest first. Each level is hidden
behind a fold — open them one at a time and go back to coding as soon as
something clicks. Level 3 is close to a spoiler; the true last resort is the
old implementation: `git show faf6a57 -- server/internal/media/`.

General workflow: run `go test ./internal/dedup -run <TestName>` to focus on
one test, and `go doc <pkg>.<Func>` (e.g. `go doc filepath.WalkDir`) to read
docs without leaving the terminal.

---

## `Scan` — walk a directory, collect regular files

<details><summary>Hint 1 (nudge)</summary>

The standard library walks trees for you: look at `path/filepath.WalkDir`.
Your callback runs once per entry — you just decide what to keep. The
`fs.DirEntry` it gives you can tell you whether the entry is a directory
and can produce an `fs.FileInfo` for the size.

</details>

<details><summary>Hint 2 (shape)</summary>

Inside the callback:
- Return the incoming `err` immediately if it's non-nil.
- Skip anything where `d.IsDir()` is true, and anything whose
  `d.Type()` is not a regular file (that filters out symlinks).
- Otherwise `d.Info()` gets you the size; append a `File{Path: path,
  Size: info.Size()}` to a slice you declared before the walk.

To skip a previous `-move` run's folder: when `d.IsDir()` and
`d.Name() == "duplicates"`, return `filepath.SkipDir`.

</details>

<details><summary>Hint 3 (near-solution)</summary>

The whole function is ~15 lines: declare `var files []File`, call
`filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {...})`
with the logic from Hint 2, return `files` and the error from `WalkDir`.
Regular-file check: `d.Type().IsRegular()`.

</details>

---

## `HashFile` — SHA-256 of a file, streamed

<details><summary>Hint 1 (nudge)</summary>

Three pieces: `os.Open`, `crypto/sha256.New()` (it's an `io.Writer`!), and
`io.Copy` to stream one into the other. Then `encoding/hex` turns the raw
digest into a string.

</details>

<details><summary>Hint 2 (shape)</summary>

- Open the file, `defer f.Close()`.
- `h := sha256.New()`, then `io.Copy(h, f)` — check its error.
- The digest is `h.Sum(nil)` (a `[]byte`); hex-encode that.

</details>

<details><summary>Hint 3 (near-solution)</summary>

```
f, err := os.Open(path)        // return "" and err on failure
defer f.Close()
h := sha256.New()
io.Copy(h, f)                   // handle the error
return hex.EncodeToString(h.Sum(nil)), nil
```

</details>

---

## `FindDuplicates` — group identical files

<details><summary>Hint 1 (nudge)</summary>

Classic bucketing problem: a `map[string][]File` keyed by hash. After
filling it, any bucket with 2+ files is a duplicate group. The determinism
test is warning you that map iteration and input order can both vary —
sorting fixes it.

</details>

<details><summary>Hint 2 (shape)</summary>

- Loop over files, `HashFile` each, append to `buckets[hash]`.
- For each bucket with `len >= 2`: sort the bucket by `Path`
  (`slices.SortFunc` or `sort.Slice`), take element 0 as `Keeper`, the
  rest as `Duplicates`.
- `WastedBytes` is just a nested sum over `Groups[i].Duplicates[j].Size`.

</details>

<details><summary>Hint 3 (near-solution)</summary>

Sorting each bucket by path before choosing `bucket[0]` as keeper makes the
keeper independent of input order — that's exactly what
`TestFindDuplicatesIsDeterministic` checks. If you also want stable *group*
ordering in the report, sort the hash keys before building `Groups`.
Optional speed-up (not tested): pre-bucket by `Size` and only hash files
whose size collides.

</details>

---

## `FormatSize` — human-readable bytes

<details><summary>Hint 1 (nudge)</summary>

Loop: while the value is ≥ 1024, divide by 1024 and step through a unit
list (`B`, `KB`, `MB`, `GB`, `TB`). Do the math in `float64` so 1536
becomes 1.5, not 1.

</details>

<details><summary>Hint 2 (shape)</summary>

Bytes under 1024 print with no decimal (`"512 B"`); larger values print
with one decimal: `fmt.Sprintf("%.1f %s", v, unit)`. The test expects
1024-based units labeled KB/MB (adjust the test if you prefer KiB or
1000-based).

</details>

<details><summary>Hint 3 (near-solution)</summary>

```
if bytes < 1024 { return fmt.Sprintf("%d B", bytes) }
v, units := float64(bytes), []string{"KB", "MB", "GB", "TB"}
i := -1
for v >= 1024 && i < len(units)-1 { v /= 1024; i++ }
return fmt.Sprintf("%.1f %s", v, units[i])
```

</details>

---

## `Delete` — remove duplicate copies

<details><summary>Hint 1 (nudge)</summary>

Two nested loops and `os.Remove`. The interesting decision is error
handling: bail on first failure, or try everything and report all
failures at the end?

</details>

<details><summary>Hint 2 (shape)</summary>

`errors.Join` (Go 1.20+) makes "keep going, collect errors" easy:
accumulate into one `err` variable with `err = errors.Join(err, ...)` and
return it at the end — it's nil if nothing failed.

</details>

---

## `Move` — relocate duplicates to `dir/duplicates/`

<details><summary>Hint 1 (nudge)</summary>

`os.MkdirAll` then `os.Rename` per duplicate. The test plants a trap:
`dupe.jpg` and `sub/dupe.jpg` share a base name, and both must survive
the move — the second rename to the same target would silently clobber
the first.

</details>

<details><summary>Hint 2 (shape)</summary>

Before renaming, check whether the target path already exists
(`os.Stat` + `os.IsNotExist`, or `errors.Is(err, fs.ErrNotExist)` with
`os.Lstat`). On collision, derive a new name — e.g. insert a counter
before the extension (`dupe-1.jpg`) and retry until free.

</details>

<details><summary>Hint 3 (near-solution)</summary>

A small helper keeps it clean:

```
func freePath(dir, base string) string  // returns dir/base, or dir/base-N.ext
```

Split `base` with `filepath.Ext` to get name+ext, loop N upward while the
candidate exists. Alternative collision strategy: recreate the duplicate's
relative subpath inside `duplicates/` (`duplicates/sub/dupe.jpg`) — then
adjust the test's flat `Glob` accordingly.

</details>

---

## When all tests pass

Try it for real: `go run ./cmd/media-merge ~/some/folder`, then
`go install ./cmd/media-merge` and run `media-merge` from anywhere.
Next milestones are in the README roadmap.
