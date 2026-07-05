package media

import (
	"sort"
	"testing"
)

func TestDedupCollapsesIdenticalHashes(t *testing.T) {
	files := []*File{
		{ID: "1", RelPath: "a/one.jpg", Hash: "aaa"},
		{ID: "2", RelPath: "b/copy.jpg", Hash: "aaa"},
		{ID: "3", RelPath: "c/unique.jpg", Hash: "bbb"},
	}

	res := Dedup(files)

	if len(res.Kept) != 2 {
		t.Fatalf("expected 2 kept files, got %d", len(res.Kept))
	}
	if len(res.Duplicates) != 1 {
		t.Fatalf("expected 1 duplicate set, got %d", len(res.Duplicates))
	}
	set := res.Duplicates[0]
	if set.Hash != "aaa" {
		t.Errorf("duplicate set hash = %q, want aaa", set.Hash)
	}
	if len(set.FileIDs) != 2 {
		t.Errorf("duplicate set should contain 2 files, got %d", len(set.FileIDs))
	}
	if set.KeptFileID != "1" {
		t.Errorf("KeptFileID = %q, want 1 (first encountered)", set.KeptFileID)
	}
}

func TestDedupNoDuplicates(t *testing.T) {
	files := []*File{
		{ID: "1", RelPath: "a.jpg", Hash: "x"},
		{ID: "2", RelPath: "b.jpg", Hash: "y"},
	}
	res := Dedup(files)
	if len(res.Kept) != 2 {
		t.Errorf("expected 2 kept, got %d", len(res.Kept))
	}
	if len(res.Duplicates) != 0 {
		t.Errorf("expected 0 duplicate sets, got %d", len(res.Duplicates))
	}
}

func TestDedupKeptIsDeterministic(t *testing.T) {
	files := []*File{
		{ID: "1", RelPath: "a.jpg", Hash: "h"},
		{ID: "2", RelPath: "b.jpg", Hash: "h"},
		{ID: "3", RelPath: "c.jpg", Hash: "h"},
	}
	res := Dedup(files)
	if len(res.Kept) != 1 {
		t.Fatalf("expected 1 kept, got %d", len(res.Kept))
	}
	ids := []string{}
	for _, f := range res.Kept {
		ids = append(ids, f.ID)
	}
	sort.Strings(ids)
	if ids[0] != "1" {
		t.Errorf("kept file = %q, want 1", ids[0])
	}
}
