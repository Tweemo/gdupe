package media

import (
	"fmt"
	"math/bits"
)

// hammingDistance counts the differing bits between two perceptual hashes.
func hammingDistance(a, b uint64) int {
	return bits.OnesCount64(a ^ b)
}

// Cluster groups eligible images whose perceptual hashes lie within threshold
// Hamming distance of each other. Grouping is transitive (union-find): if a~b
// and b~d, all three land in one group. Only images with HasPHash are
// considered; singletons (no near neighbour) are omitted from the result.
//
// The returned groups are deterministic: members preserve input order and
// groups are ordered by their first member's input position.
func Cluster(files []*File, threshold int) []SimilarityGroup {
	// Collect eligible images in input order.
	var eligible []*File
	for _, f := range files {
		if f.Type == Image && f.HasPHash {
			eligible = append(eligible, f)
		}
	}

	n := len(eligible)
	parent := make([]int, n)
	for i := range parent {
		parent[i] = i
	}
	var find func(int) int
	find = func(x int) int {
		for parent[x] != x {
			parent[x] = parent[parent[x]]
			x = parent[x]
		}
		return x
	}
	union := func(a, b int) {
		ra, rb := find(a), find(b)
		if ra != rb {
			parent[rb] = ra
		}
	}

	for i := 0; i < n; i++ {
		for j := i + 1; j < n; j++ {
			if hammingDistance(eligible[i].PHash, eligible[j].PHash) <= threshold {
				union(i, j)
			}
		}
	}

	// Gather members per root, preserving input order.
	members := map[int][]string{}
	rootOrder := []int{}
	for i := 0; i < n; i++ {
		r := find(i)
		if _, ok := members[r]; !ok {
			rootOrder = append(rootOrder, r)
		}
		members[r] = append(members[r], eligible[i].ID)
	}

	var groups []SimilarityGroup
	idx := 0
	for _, r := range rootOrder {
		if len(members[r]) < 2 {
			continue // singleton: handled as ungrouped by the caller
		}
		idx++
		groups = append(groups, SimilarityGroup{
			ID:      fmt.Sprintf("image-group-%03d", idx),
			FileIDs: members[r],
		})
	}
	return groups
}
