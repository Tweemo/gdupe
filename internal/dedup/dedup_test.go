package dedup

// These tests describe the behavior the package should have. They all
// fail with "not implemented" until you write the real code — work
// through them roughly in order: Scan → HashFile → FindDuplicates →
// WastedBytes/FormatSize → Delete → Move.

import (
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"
)

// writeTree creates the given files (name → content) under a fresh
// temp dir and returns the dir. Subdirectories in names are created.
func writeTree(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for name, content := range files {
		path := filepath.Join(dir, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func TestScanFindsAllRegularFiles(t *testing.T) {
	dir := writeTree(t, map[string]string{
		"a.jpg":        "aaa",
		"sub/b.jpg":    "bbb",
		"sub/deep/c":   "ccc",
		"empty-ok.txt": "",
	})

	files, err := Scan(dir)
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if len(files) != 4 {
		t.Fatalf("Scan found %d files, want 4: %+v", len(files), files)
	}
	for _, f := range files {
		if f.Path == "" {
			t.Errorf("file has empty Path: %+v", f)
		}
	}
}

func TestScanRecordsSizes(t *testing.T) {
	dir := writeTree(t, map[string]string{"f.bin": "12345"})

	files, err := Scan(dir)
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if len(files) != 1 || files[0].Size != 5 {
		t.Fatalf("got %+v, want one file of size 5", files)
	}
}

func TestHashFileIsContentBased(t *testing.T) {
	dir := writeTree(t, map[string]string{
		"one.txt":   "same content",
		"two.txt":   "same content",
		"other.txt": "different",
	})

	h1, err := HashFile(filepath.Join(dir, "one.txt"))
	if err != nil {
		t.Fatalf("HashFile: %v", err)
	}
	h2, _ := HashFile(filepath.Join(dir, "two.txt"))
	h3, _ := HashFile(filepath.Join(dir, "other.txt"))

	if h1 != h2 {
		t.Errorf("identical contents hashed differently: %q vs %q", h1, h2)
	}
	if h1 == h3 {
		t.Errorf("different contents hashed the same: %q", h1)
	}
	// SHA-256 hex is 64 chars; catches accidental use of a shorter hash.
	if len(h1) != 64 {
		t.Errorf("hash length = %d, want 64 hex chars (SHA-256)", len(h1))
	}
}

func TestFindDuplicatesGroupsIdenticalFiles(t *testing.T) {
	dir := writeTree(t, map[string]string{
		"a.jpg":     "PHOTO-1",
		"copy.jpg":  "PHOTO-1",
		"sub/x.jpg": "PHOTO-1",
		"b.jpg":     "PHOTO-2",
		"c.jpg":     "unique",
	})
	files, err := Scan(dir)
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}

	report, err := FindDuplicates(files)
	if err != nil {
		t.Fatalf("FindDuplicates: %v", err)
	}

	// PHOTO-1 appears 3x → one group with 2 duplicates. Everything else
	// is unique and must not appear in any group.
	if len(report.Groups) != 1 {
		t.Fatalf("got %d groups, want 1: %+v", len(report.Groups), report.Groups)
	}
	g := report.Groups[0]
	if len(g.Duplicates) != 2 {
		t.Fatalf("group has %d duplicates, want 2: %+v", len(g.Duplicates), g)
	}
	if report.DuplicateCount() != 2 {
		t.Errorf("DuplicateCount = %d, want 2", report.DuplicateCount())
	}
	if got, want := report.WastedBytes(), int64(2*len("PHOTO-1")); got != want {
		t.Errorf("WastedBytes = %d, want %d", got, want)
	}
	for _, d := range g.Duplicates {
		if d.Path == g.Keeper.Path {
			t.Errorf("keeper %q also listed as duplicate", d.Path)
		}
	}
}

func TestFindDuplicatesIsDeterministic(t *testing.T) {
	dir := writeTree(t, map[string]string{
		"z.jpg": "PIC",
		"a.jpg": "PIC",
	})
	files, err := Scan(dir)
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}

	first, err := FindDuplicates(files)
	if err != nil {
		t.Fatalf("FindDuplicates: %v", err)
	}
	// Shuffled input order must not change which file is the keeper.
	slices.Reverse(files)
	second, err := FindDuplicates(files)
	if err != nil {
		t.Fatalf("FindDuplicates: %v", err)
	}

	if first.Groups[0].Keeper.Path != second.Groups[0].Keeper.Path {
		t.Errorf("keeper depends on input order: %q vs %q",
			first.Groups[0].Keeper.Path, second.Groups[0].Keeper.Path)
	}
}

func TestFormatSize(t *testing.T) {
	cases := []struct {
		bytes int64
		want  string
	}{
		{0, "0 B"},
		{512, "512 B"},
		{1536, "1.5 KB"}, // adjust if you choose KiB/1000-based units
	}
	for _, c := range cases {
		if got := FormatSize(c.bytes); got != c.want {
			t.Errorf("FormatSize(%d) = %q, want %q", c.bytes, got, c.want)
		}
	}
}

func TestDeleteRemovesOnlyDuplicates(t *testing.T) {
	trash := t.TempDir()
	// a/pic.jpg is the lexicographically-smallest copy, so it is the
	// keeper; b/pic.jpg and c/pic.jpg are duplicates that share a base
	// name → naive Base()-naming in the trash would overwrite one.
	dir := writeTree(t, map[string]string{
		"a/pic.jpg": "SAME",
		"b/pic.jpg": "SAME",
		"c/pic.jpg": "SAME",
		"solo.jpg":  "unique",
	})
	files, err := Scan(dir)
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	report, err := FindDuplicates(files)
	if err != nil {
		t.Fatalf("FindDuplicates: %v", err)
	}

	if err := Delete(report, trash); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	// Keeper and the unique file must be untouched.
	for _, path := range []string{report.Groups[0].Keeper.Path, filepath.Join(dir, "solo.jpg")} {
		if _, err := os.Stat(path); err != nil {
			t.Errorf("non-duplicate %q should still exist: %v", path, err)
		}
	}

	// Every duplicate must be gone from its original location...
	for _, d := range report.Groups[0].Duplicates {
		if _, err := os.Stat(d.Path); !os.IsNotExist(err) {
			t.Errorf("duplicate %q should have been trashed, stat err = %v", d.Path, err)
		}
	}

	// ...and must have arrived in the trash dir without overwriting
	// each other (b/pic.jpg and c/pic.jpg collide on base name).
	trashed, _ := filepath.Glob(filepath.Join(trash, "*"))
	if want := report.DuplicateCount(); len(trashed) != want {
		t.Fatalf("trash holds %d files, want %d (collision must not overwrite): %v",
			len(trashed), want, trashed)
	}
}

func TestTrashDir(t *testing.T) {
	dir, err := TrashDir()
	switch runtime.GOOS {
	case "darwin", "linux":
		if err != nil {
			t.Fatalf("TrashDir on %s: %v", runtime.GOOS, err)
		}
		home, herr := os.UserHomeDir()
		if herr != nil {
			t.Skipf("no home dir available: %v", herr)
		}
		if !strings.HasPrefix(dir, home) {
			t.Errorf("TrashDir = %q, want a path under home %q", dir, home)
		}
	default:
		// Unsupported OS must refuse loudly, never return a usable-looking
		// empty path (Rename into "" would scatter files into the CWD).
		if err == nil {
			t.Errorf("TrashDir on %s = %q, want error", runtime.GOOS, dir)
		}
	}
}

func TestMoveRelocatesDuplicatesIntoSubfolder(t *testing.T) {
	dir := writeTree(t, map[string]string{
		"keep.jpg":     "SAME",
		"dupe.jpg":     "SAME",
		"sub/dupe.jpg": "SAME", // same base name as dupe.jpg → collision case
	})
	files, err := Scan(dir)
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	report, err := FindDuplicates(files)
	if err != nil {
		t.Fatalf("FindDuplicates: %v", err)
	}

	if err := Move(report, dir); err != nil {
		t.Fatalf("Move: %v", err)
	}

	moved, _ := filepath.Glob(filepath.Join(dir, "duplicates", "*"))
	if len(moved) != 2 {
		t.Fatalf("duplicates/ holds %d files, want 2 (collision must not overwrite): %v", len(moved), moved)
	}
}
