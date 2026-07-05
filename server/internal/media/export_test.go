package media

import (
	"archive/zip"
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestLayoutSortsFilesByGroupAndType(t *testing.T) {
	files := []*File{
		{ID: "f1", RelPath: "A/cat.jpg", Type: Image},
		{ID: "f2", RelPath: "B/cat-edit.jpg", Type: Image},
		{ID: "f3", RelPath: "A/lonely.png", Type: Image},
		{ID: "f4", RelPath: "A/clip.mp4", Type: Video},
		{ID: "f5", RelPath: "A/song.mp3", Type: Audio},
		{ID: "f6", RelPath: "A/notes.txt", Type: Other},
	}
	groups := []SimilarityGroup{{ID: "image-group-001", FileIDs: []string{"f1", "f2"}}}

	layout := Layout(files, groups)

	want := map[string]string{
		"f1": "image-group-001/cat.jpg",
		"f2": "image-group-001/cat-edit.jpg",
		"f3": "images/lonely.png",
		"f4": "video/clip.mp4",
		"f5": "audio/song.mp3",
		"f6": "other/notes.txt",
	}
	for id, wantPath := range want {
		if layout[id] != wantPath {
			t.Errorf("layout[%s] = %q, want %q", id, layout[id], wantPath)
		}
	}
}

func TestLayoutDisambiguatesNameCollisions(t *testing.T) {
	files := []*File{
		{ID: "f1", RelPath: "A/photo.jpg", Type: Image},
		{ID: "f2", RelPath: "B/photo.jpg", Type: Image},
	}
	layout := Layout(files, nil)

	if layout["f1"] == layout["f2"] {
		t.Fatalf("colliding names mapped to the same path: %q", layout["f1"])
	}
	if layout["f1"] != "images/photo.jpg" {
		t.Errorf("first file = %q, want images/photo.jpg", layout["f1"])
	}
	if layout["f2"] != "images/photo-1.jpg" {
		t.Errorf("second file = %q, want images/photo-1.jpg", layout["f2"])
	}
}

func TestWriteZipProducesExpectedEntries(t *testing.T) {
	dir := t.TempDir()
	// Two real files on disk.
	p1 := filepath.Join(dir, "one.jpg")
	p2 := filepath.Join(dir, "two.mp4")
	if err := os.WriteFile(p1, []byte("image-bytes"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p2, []byte("video-bytes"), 0o644); err != nil {
		t.Fatal(err)
	}
	files := []*File{
		{ID: "f1", RelPath: "x/one.jpg", AbsPath: p1, Type: Image},
		{ID: "f2", RelPath: "x/two.mp4", AbsPath: p2, Type: Video},
	}
	layout := Layout(files, nil)

	var buf bytes.Buffer
	if err := WriteZip(&buf, files, layout); err != nil {
		t.Fatalf("WriteZip: %v", err)
	}

	zr, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		t.Fatalf("open zip: %v", err)
	}
	got := map[string]string{}
	for _, zf := range zr.File {
		rc, err := zf.Open()
		if err != nil {
			t.Fatal(err)
		}
		b, _ := io.ReadAll(rc)
		rc.Close()
		got[zf.Name] = string(b)
	}

	if got["merged/images/one.jpg"] != "image-bytes" {
		t.Errorf("merged/images/one.jpg = %q", got["merged/images/one.jpg"])
	}
	if got["merged/video/two.mp4"] != "video-bytes" {
		t.Errorf("merged/video/two.mp4 = %q", got["merged/video/two.mp4"])
	}
}
