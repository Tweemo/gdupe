package media

import (
	"os"
	"path/filepath"
	"testing"
)

func writeFile(t *testing.T, dir, name string, data []byte) string {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, data, 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestAnalyzeEndToEnd(t *testing.T) {
	dir := t.TempDir()

	aBytes := pngBytes(t, horizontalGradient(64, 64))
	cBytes := pngBytes(t, horizontalGradient(200, 200)) // similar to a, different bytes
	dBytes := pngBytes(t, verticalGradient(64, 64))     // structurally different
	vid := []byte("fake-video-bytes")

	files := []*File{
		{ID: "a", RelPath: "f1/a.png", AbsPath: writeFile(t, dir, "a.png", aBytes)},
		{ID: "b", RelPath: "f2/b.png", AbsPath: writeFile(t, dir, "b.png", aBytes)}, // exact dup of a
		{ID: "c", RelPath: "f1/c.png", AbsPath: writeFile(t, dir, "c.png", cBytes)},
		{ID: "d", RelPath: "f1/d.png", AbsPath: writeFile(t, dir, "d.png", dBytes)},
		{ID: "e", RelPath: "f1/e.mp4", AbsPath: writeFile(t, dir, "e.mp4", vid)},
	}

	res := Analyze(files, 10)

	// Exact duplicate a/b collapses to one set.
	if len(res.Duplicates) != 1 {
		t.Fatalf("expected 1 duplicate set, got %d", len(res.Duplicates))
	}
	if len(res.Duplicates[0].FileIDs) != 2 {
		t.Errorf("duplicate set should hold 2 files, got %v", res.Duplicates[0].FileIDs)
	}

	// Kept = a, c, d, e (b removed as duplicate).
	if len(res.Kept) != 4 {
		t.Fatalf("expected 4 kept files, got %d", len(res.Kept))
	}

	// Types were classified.
	for _, f := range files {
		want := Image
		if f.ID == "e" {
			want = Video
		}
		if f.Type != want {
			t.Errorf("file %s type = %v, want %v", f.ID, f.Type, want)
		}
	}

	// a and c should cluster as a similar-image group; d and e should not.
	g := groupContaining(res.Groups, "a")
	if len(g) != 2 {
		t.Errorf("expected a and c grouped, got %v", g)
	}
	if groupContaining(res.Groups, "d") != nil {
		t.Errorf("d should be an ungrouped singleton")
	}
	if groupContaining(res.Groups, "e") != nil {
		t.Errorf("video e should never be grouped")
	}
}
