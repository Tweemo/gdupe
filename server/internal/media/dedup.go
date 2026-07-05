package media

// Dedup groups files by their content Hash. For each distinct hash the
// first file encountered (in input order) is kept; the rest are recorded as
// duplicates. Hashes seen on only one file produce no DuplicateSet.
func Dedup(files []*File) DedupResult {
	order := []string{}            // distinct hashes in first-seen order
	byHash := map[string][]*File{} // hash -> files sharing it

	for _, f := range files {
		if _, seen := byHash[f.Hash]; !seen {
			order = append(order, f.Hash)
		}
		byHash[f.Hash] = append(byHash[f.Hash], f)
	}

	res := DedupResult{}
	for _, h := range order {
		group := byHash[h]
		res.Kept = append(res.Kept, group[0])
		if len(group) > 1 {
			ids := make([]string, len(group))
			for i, f := range group {
				ids[i] = f.ID
			}
			res.Duplicates = append(res.Duplicates, DuplicateSet{
				Hash:       h,
				KeptFileID: group[0].ID,
				FileIDs:    ids,
			})
		}
	}
	return res
}
