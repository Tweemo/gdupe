package media

// File is one media file under analysis within a session.
type File struct {
	ID           string    `json:"id"`
	RelPath      string    `json:"relPath"`      // path relative to its source folder (from webkitRelativePath)
	SourceFolder string    `json:"sourceFolder"` // top-level folder the user selected
	AbsPath      string    `json:"-"`            // location on the server's temp disk
	Size         int64     `json:"size"`
	Hash         string    `json:"hash"`  // SHA-256 hex; empty until hashed
	Type         MediaType `json:"type"`  // image/video/audio/other
	PHash        uint64    `json:"-"`     // perceptual hash; images only
	HasPHash     bool      `json:"-"`     // false when perceptual hashing was skipped/failed
}

// DuplicateSet is a group of files that share the same content hash.
type DuplicateSet struct {
	Hash       string   `json:"hash"`
	KeptFileID string   `json:"keptFileId"`
	FileIDs    []string `json:"fileIds"` // every file in the set, including the kept one
}

// SimilarityGroup is a cluster of visually-similar images.
type SimilarityGroup struct {
	ID      string   `json:"id"`
	FileIDs []string `json:"fileIds"`
}

// DedupResult is the outcome of exact-duplicate detection.
type DedupResult struct {
	Kept       []*File        // one keeper per unique hash, in first-seen order
	Duplicates []DuplicateSet // only hashes with more than one file
}
