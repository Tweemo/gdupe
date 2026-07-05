package session

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestStoreNewCreatesWorkingDir(t *testing.T) {
	store := NewStore(t.TempDir())
	s, err := store.New()
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if s.ID == "" {
		t.Error("session ID is empty")
	}
	if fi, err := os.Stat(s.Dir); err != nil || !fi.IsDir() {
		t.Errorf("session dir %q not a directory: %v", s.Dir, err)
	}
}

func TestAddFileStreamsToDisk(t *testing.T) {
	store := NewStore(t.TempDir())
	s, _ := store.New()

	f, err := s.AddFile("vacation/beach.jpg", "vacation", strings.NewReader("hello-bytes"))
	if err != nil {
		t.Fatalf("AddFile: %v", err)
	}
	if f.ID == "" {
		t.Error("file ID empty")
	}
	if f.RelPath != "vacation/beach.jpg" {
		t.Errorf("RelPath = %q", f.RelPath)
	}
	if f.SourceFolder != "vacation" {
		t.Errorf("SourceFolder = %q", f.SourceFolder)
	}
	if f.Size != int64(len("hello-bytes")) {
		t.Errorf("Size = %d, want %d", f.Size, len("hello-bytes"))
	}
	// File must live inside the session dir.
	if !strings.HasPrefix(f.AbsPath, s.Dir) {
		t.Errorf("AbsPath %q not under session dir %q", f.AbsPath, s.Dir)
	}
	data, err := os.ReadFile(f.AbsPath)
	if err != nil {
		t.Fatalf("reading stored file: %v", err)
	}
	if string(data) != "hello-bytes" {
		t.Errorf("stored content = %q", string(data))
	}
	// Session tracks the file.
	if len(s.Files) != 1 || s.Files[0].ID != f.ID {
		t.Errorf("session did not record the file")
	}
}

func TestAddFileAssignsUniqueIDsAndPaths(t *testing.T) {
	store := NewStore(t.TempDir())
	s, _ := store.New()
	a, _ := s.AddFile("a/pic.jpg", "a", strings.NewReader("one"))
	b, _ := s.AddFile("b/pic.jpg", "b", strings.NewReader("two"))
	if a.ID == b.ID {
		t.Error("expected unique file IDs")
	}
	if a.AbsPath == b.AbsPath {
		t.Error("expected unique disk paths even for same base name")
	}
}

func TestGetAndDelete(t *testing.T) {
	store := NewStore(t.TempDir())
	s, _ := store.New()
	dir := s.Dir

	got, ok := store.Get(s.ID)
	if !ok || got.ID != s.ID {
		t.Fatalf("Get returned %v, %v", got, ok)
	}

	if _, ok := store.Get("does-not-exist"); ok {
		t.Error("Get of unknown id should return false")
	}

	if err := store.Delete(s.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, ok := store.Get(s.ID); ok {
		t.Error("session still present after Delete")
	}
	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		t.Errorf("session dir %q still exists after Delete", filepath.Clean(dir))
	}
}
