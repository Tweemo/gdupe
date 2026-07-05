package api

import (
	"encoding/json"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/timl/media-merge/server/internal/media"
)

const thumbnailSize = 240

func (s *Server) handleCreateSession(w http.ResponseWriter, r *http.Request) {
	sess, err := s.store.New()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not create session")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"sessionId": sess.ID})
}

// handleUpload streams each multipart file part to the session's working dir.
func (s *Server) handleUpload(w http.ResponseWriter, r *http.Request) {
	sess, ok := s.session(w, r)
	if !ok {
		return
	}

	reader, err := r.MultipartReader()
	if err != nil {
		writeError(w, http.StatusBadRequest, "expected multipart/form-data")
		return
	}

	count := 0
	for {
		part, err := reader.NextPart()
		if err != nil {
			break // io.EOF or end of stream
		}
		if part.FormName() != "files" || part.FileName() == "" {
			part.Close()
			continue
		}
		relPath := part.FileName()
		if _, err := sess.AddFile(relPath, firstSegment(relPath), part); err != nil {
			part.Close()
			writeError(w, http.StatusInternalServerError, "failed to store "+relPath)
			return
		}
		part.Close()
		count++
	}

	writeJSON(w, http.StatusOK, map[string]int{"received": count})
}

func (s *Server) handleAnalyze(w http.ResponseWriter, r *http.Request) {
	sess, ok := s.session(w, r)
	if !ok {
		return
	}
	res := media.Analyze(sess.Files, media.DefaultThreshold)
	sess.Result = &res
	writeJSON(w, http.StatusOK, res)
}

func (s *Server) handleThumbnail(w http.ResponseWriter, r *http.Request) {
	sess, ok := s.session(w, r)
	if !ok {
		return
	}
	file, ok := sess.FileByID(r.PathValue("fileId"))
	if !ok {
		writeError(w, http.StatusNotFound, "file not found")
		return
	}
	if file.Type != media.Image {
		writeError(w, http.StatusUnsupportedMediaType, "not an image")
		return
	}

	f, err := os.Open(file.AbsPath)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "cannot read file")
		return
	}
	defer f.Close()

	thumb, err := media.Thumbnail(f, thumbnailSize)
	if err != nil {
		writeError(w, http.StatusUnsupportedMediaType, "cannot render thumbnail")
		return
	}
	w.Header().Set("Content-Type", "image/jpeg")
	w.Write(thumb)
}

type exportRequest struct {
	Groups []media.SimilarityGroup `json:"groups"`
}

func (s *Server) handleExport(w http.ResponseWriter, r *http.Request) {
	sess, ok := s.session(w, r)
	if !ok {
		return
	}
	if sess.Result == nil {
		writeError(w, http.StatusBadRequest, "analyze the session before exporting")
		return
	}

	var req exportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid export request body")
		return
	}

	zipPath := path.Join(path.Dir(sess.Dir), "merged.zip")
	out, err := os.Create(zipPath)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "cannot create archive")
		return
	}
	err = media.WriteZip(out, sess.Result.Kept, media.Layout(sess.Result.Kept, req.Groups))
	if cerr := out.Close(); err == nil {
		err = cerr
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to build archive")
		return
	}
	sess.ZipPath = zipPath
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (s *Server) handleDownload(w http.ResponseWriter, r *http.Request) {
	sess, ok := s.session(w, r)
	if !ok {
		return
	}
	if sess.ZipPath == "" {
		writeError(w, http.StatusNotFound, "nothing exported yet")
		return
	}
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", `attachment; filename="merged.zip"`)
	http.ServeFile(w, r, sess.ZipPath)
}

func (s *Server) handleDelete(w http.ResponseWriter, r *http.Request) {
	if err := s.store.Delete(r.PathValue("id")); err != nil {
		writeError(w, http.StatusInternalServerError, "cleanup failed")
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// firstSegment returns the top-level folder of a webkitRelativePath.
func firstSegment(relPath string) string {
	if i := strings.IndexByte(relPath, '/'); i >= 0 {
		return relPath[:i]
	}
	return ""
}
