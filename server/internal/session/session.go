// Package session manages per-upload working state: a temp directory holding
// the uploaded files plus the analysis result, kept in memory and addressable
// by an opaque session ID.
package session

import (
	"crypto/rand"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/timl/media-merge/server/internal/media"
)

// Session is one upload batch and its working directory.
type Session struct {
	ID      string
	Dir     string // temp directory holding uploaded files (under <root>/<id>/src)
	Created time.Time

	mu      sync.Mutex
	Files   []*media.File
	Result  *media.AnalyzeResult // nil until analysed
	ZipPath string               // populated after export
	nextSeq int                  // per-session counter for unique on-disk names
}

// Store is an in-memory collection of sessions rooted at a base temp directory.
type Store struct {
	root string
	mu   sync.RWMutex
	byID map[string]*Session
}

// NewStore returns a Store that creates session directories under root.
func NewStore(root string) *Store {
	return &Store{root: root, byID: map[string]*Session{}}
}

// New creates a session with a fresh working directory.
func (st *Store) New() (*Session, error) {
	id, err := randomID()
	if err != nil {
		return nil, err
	}
	dir := filepath.Join(st.root, id, "src")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	s := &Session{ID: id, Dir: dir, Created: time.Now()}

	st.mu.Lock()
	st.byID[id] = s
	st.mu.Unlock()
	return s, nil
}

// Get returns the session with the given ID, if present.
func (st *Store) Get(id string) (*Session, bool) {
	st.mu.RLock()
	defer st.mu.RUnlock()
	s, ok := st.byID[id]
	return s, ok
}

// Delete removes a session and its working directory from disk.
func (st *Store) Delete(id string) error {
	st.mu.Lock()
	s, ok := st.byID[id]
	delete(st.byID, id)
	st.mu.Unlock()
	if !ok {
		return nil
	}
	// Remove the whole <root>/<id> tree, not just src.
	return os.RemoveAll(filepath.Dir(s.Dir))
}

// AddFile streams r into the session's working directory and records a
// media.File describing it. relPath is the browser's webkitRelativePath;
// sourceFolder is the top-level folder the user selected.
func (s *Session) AddFile(relPath, sourceFolder string, r io.Reader) (*media.File, error) {
	s.mu.Lock()
	seq := s.nextSeq
	s.nextSeq++
	s.mu.Unlock()

	id := strconv.Itoa(seq)
	// Store on disk under a flat, collision-free name keyed by sequence,
	// preserving the original extension.
	abs := filepath.Join(s.Dir, id+filepath.Ext(relPath))

	dst, err := os.Create(abs)
	if err != nil {
		return nil, err
	}
	n, err := io.Copy(dst, r)
	if cerr := dst.Close(); err == nil {
		err = cerr
	}
	if err != nil {
		os.Remove(abs)
		return nil, err
	}

	f := &media.File{
		ID:           id,
		RelPath:      relPath,
		SourceFolder: sourceFolder,
		AbsPath:      abs,
		Size:         n,
	}

	s.mu.Lock()
	s.Files = append(s.Files, f)
	s.mu.Unlock()
	return f, nil
}

// FileByID returns the recorded file with the given ID.
func (s *Session) FileByID(id string) (*media.File, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, f := range s.Files {
		if f.ID == id {
			return f, true
		}
	}
	return nil, false
}

func randomID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
