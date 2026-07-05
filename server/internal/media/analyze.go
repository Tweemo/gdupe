package media

import "os"

// DefaultThreshold is the Hamming-distance cutoff for treating two images'
// difference hashes as "visually similar".
const DefaultThreshold = 10

// AnalyzeResult is the outcome of analysing an upload batch.
type AnalyzeResult struct {
	Kept       []*File           `json:"kept"`       // unique files after exact dedup
	Duplicates []DuplicateSet    `json:"duplicates"` // exact-duplicate sets that were collapsed
	Groups     []SimilarityGroup `json:"groups"`     // proposed similar-image clusters
}

// Analyze populates each file's Type, Hash and (for images) perceptual hash,
// removes exact duplicates, and clusters the remaining images into
// similarity groups. Files are mutated in place with their computed metadata.
//
// Hashing or perceptual-hashing failures for an individual file are tolerated:
// the file keeps an empty hash / HasPHash=false and simply won't match others.
func Analyze(files []*File, threshold int) AnalyzeResult {
	for _, f := range files {
		f.Type = Classify(f.RelPath)

		if h, err := HashFile(f.AbsPath); err == nil {
			f.Hash = h
		}

		if f.Type == Image {
			if ph, err := perceptualHashFile(f.AbsPath); err == nil {
				f.PHash = ph
				f.HasPHash = true
			}
		}
	}

	dedup := Dedup(files)
	groups := Cluster(dedup.Kept, threshold)

	return AnalyzeResult{
		Kept:       dedup.Kept,
		Duplicates: dedup.Duplicates,
		Groups:     groups,
	}
}

func perceptualHashFile(path string) (uint64, error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer f.Close()
	return PerceptualHash(f)
}
