package media

import (
	"sort"
	"testing"
)

// groupContaining returns the FileIDs (sorted) of the group containing id, or nil.
func groupContaining(groups []SimilarityGroup, id string) []string {
	for _, g := range groups {
		for _, fid := range g.FileIDs {
			if fid == id {
				out := append([]string{}, g.FileIDs...)
				sort.Strings(out)
				return out
			}
		}
	}
	return nil
}

func img(id string, phash uint64) *File {
	return &File{ID: id, Type: Image, PHash: phash, HasPHash: true}
}

func TestClusterGroupsNearbyHashes(t *testing.T) {
	files := []*File{
		img("a", 0x0000000000000000),
		img("b", 0x0000000000000001), // distance 1 from a
		img("c", 0xFFFFFFFFFFFFFFFF), // far from both
	}

	groups := Cluster(files, 5)

	ab := groupContaining(groups, "a")
	if len(ab) != 2 || ab[0] != "a" || ab[1] != "b" {
		t.Errorf("expected a and b grouped together, got %v", ab)
	}
	if groupContaining(groups, "c") != nil {
		t.Errorf("c is a singleton and should not appear in any group")
	}
}

func TestClusterTransitiveClosure(t *testing.T) {
	// a-b within threshold, b-d within threshold, a-d beyond threshold.
	files := []*File{
		img("a", 0b0000),
		img("b", 0b0111), // dist 3 from a
		img("d", 0b1110), // dist 3 from b (0b0111 ^ 0b1110 = 0b1001 -> 2 bits)... use clear values below
	}
	// Recompute with explicit values: a=0, b has 3 bits, chained.
	files = []*File{
		img("a", 0x00),
		img("b", 0x07), // dist 3 from a
		img("d", 0x37), // 0x07 ^ 0x37 = 0x30 -> 2 bits from b; 0x00 ^ 0x37 = 6 bits from a
	}

	groups := Cluster(files, 4)

	g := groupContaining(groups, "a")
	if len(g) != 3 {
		t.Errorf("expected a, b, d in one transitive group, got %v", g)
	}
}

func TestClusterIgnoresNonImagesAndMissingPHash(t *testing.T) {
	files := []*File{
		img("a", 0x00),
		{ID: "vid", Type: Video, PHash: 0x00, HasPHash: false},
		{ID: "noph", Type: Image, HasPHash: false},
	}
	groups := Cluster(files, 5)
	if len(groups) != 0 {
		t.Errorf("expected no groups (only one eligible image), got %v", groups)
	}
}
