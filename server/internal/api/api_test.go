package api

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"image"
	"image/color"
	"image/png"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/timl/media-merge/server/internal/session"
)

// ---- test helpers ----

func newTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	store := session.NewStore(t.TempDir())
	srv := New(store)
	ts := httptest.NewServer(srv.Handler())
	t.Cleanup(ts.Close)
	return ts
}

func horizGradientPNG(t *testing.T, w, h int) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for x := 0; x < w; x++ {
		v := uint8(x * 255 / (w - 1))
		for y := 0; y < h; y++ {
			img.Set(x, y, color.RGBA{v, v, v, 255})
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func createSession(t *testing.T, ts *httptest.Server) string {
	t.Helper()
	resp, err := http.Post(ts.URL+"/api/sessions", "application/json", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("create session status = %d", resp.StatusCode)
	}
	var body struct {
		SessionID string `json:"sessionId"`
	}
	json.NewDecoder(resp.Body).Decode(&body)
	if body.SessionID == "" {
		t.Fatal("empty sessionId")
	}
	return body.SessionID
}

type upload struct {
	relPath string
	data    []byte
}

func uploadFiles(t *testing.T, ts *httptest.Server, id string, ups []upload) {
	t.Helper()
	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	for _, u := range ups {
		// The part filename carries the webkitRelativePath.
		pw, err := mw.CreateFormFile("files", u.relPath)
		if err != nil {
			t.Fatal(err)
		}
		pw.Write(u.data)
	}
	mw.Close()

	resp, err := http.Post(ts.URL+"/api/sessions/"+id+"/upload", mw.FormDataContentType(), &body)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("upload status = %d: %s", resp.StatusCode, b)
	}
}

// ---- tests ----

func TestCreateSessionReturnsID(t *testing.T) {
	ts := newTestServer(t)
	if id := createSession(t, ts); id == "" {
		t.Fatal("expected a session id")
	}
}

type analyzeResponse struct {
	Kept []struct {
		ID      string `json:"id"`
		RelPath string `json:"relPath"`
		Type    string `json:"type"`
	} `json:"kept"`
	Duplicates []struct {
		FileIDs []string `json:"fileIds"`
	} `json:"duplicates"`
	Groups []struct {
		ID      string   `json:"id"`
		FileIDs []string `json:"fileIds"`
	} `json:"groups"`
}

func analyze(t *testing.T, ts *httptest.Server, id string) analyzeResponse {
	t.Helper()
	resp, err := http.Post(ts.URL+"/api/sessions/"+id+"/analyze", "application/json", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("analyze status = %d: %s", resp.StatusCode, b)
	}
	var out analyzeResponse
	json.NewDecoder(resp.Body).Decode(&out)
	return out
}

func TestUploadAnalyzeDedupAndGroup(t *testing.T) {
	ts := newTestServer(t)
	id := createSession(t, ts)

	a := horizGradientPNG(t, 64, 64)
	c := horizGradientPNG(t, 200, 200) // similar to a, different bytes
	uploadFiles(t, ts, id, []upload{
		{"trip/a.png", a},
		{"trip/b.png", a}, // exact duplicate of a
		{"trip/c.png", c},
		{"trip/clip.mp4", []byte("fake-video")},
	})

	res := analyze(t, ts, id)

	if len(res.Duplicates) != 1 || len(res.Duplicates[0].FileIDs) != 2 {
		t.Errorf("expected one duplicate set of 2, got %+v", res.Duplicates)
	}
	if len(res.Kept) != 3 {
		t.Errorf("expected 3 kept files, got %d", len(res.Kept))
	}
	// a and c should be proposed as a similar group.
	if len(res.Groups) != 1 || len(res.Groups[0].FileIDs) != 2 {
		t.Errorf("expected one similar group of 2, got %+v", res.Groups)
	}
}

func TestThumbnailServesJPEG(t *testing.T) {
	ts := newTestServer(t)
	id := createSession(t, ts)
	uploadFiles(t, ts, id, []upload{{"a/pic.png", horizGradientPNG(t, 120, 120)}})
	analyze(t, ts, id)

	resp, err := http.Get(ts.URL + "/api/sessions/" + id + "/thumbnail/0")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("thumbnail status = %d", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	if _, format, err := image.DecodeConfig(bytes.NewReader(body)); err != nil || format != "jpeg" {
		t.Errorf("thumbnail not jpeg (format=%q, err=%v)", format, err)
	}
}

func TestExportAndDownloadProducesZip(t *testing.T) {
	ts := newTestServer(t)
	id := createSession(t, ts)
	a := horizGradientPNG(t, 64, 64)
	c := horizGradientPNG(t, 200, 200)
	uploadFiles(t, ts, id, []upload{
		{"trip/a.png", a},
		{"trip/c.png", c},
		{"trip/song.mp3", []byte("audio")},
	})
	res := analyze(t, ts, id)

	// Confirm the proposed groups as-is.
	exportBody, _ := json.Marshal(map[string]any{"groups": res.Groups})
	resp, err := http.Post(ts.URL+"/api/sessions/"+id+"/export", "application/json", bytes.NewReader(exportBody))
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("export status = %d", resp.StatusCode)
	}

	dl, err := http.Get(ts.URL + "/api/sessions/" + id + "/download")
	if err != nil {
		t.Fatal(err)
	}
	defer dl.Body.Close()
	if dl.StatusCode != http.StatusOK {
		t.Fatalf("download status = %d", dl.StatusCode)
	}
	zbytes, _ := io.ReadAll(dl.Body)
	zr, err := zip.NewReader(bytes.NewReader(zbytes), int64(len(zbytes)))
	if err != nil {
		t.Fatalf("response is not a valid zip: %v", err)
	}
	var names []string
	for _, f := range zr.File {
		names = append(names, f.Name)
	}
	joined := strings.Join(names, "\n")
	if !strings.Contains(joined, "merged/") {
		t.Errorf("zip missing merged/ root, names:\n%s", joined)
	}
	if !strings.Contains(joined, "audio/song.mp3") {
		t.Errorf("zip missing audio file, names:\n%s", joined)
	}
}

func TestDeleteSession(t *testing.T) {
	ts := newTestServer(t)
	id := createSession(t, ts)

	req, _ := http.NewRequest(http.MethodDelete, ts.URL+"/api/sessions/"+id, nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("delete status = %d", resp.StatusCode)
	}
	// Analyzing a deleted session should now fail.
	resp2, err := http.Post(ts.URL+"/api/sessions/"+id+"/analyze", "application/json", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer resp2.Body.Close()
	if resp2.StatusCode == http.StatusOK {
		t.Errorf("expected non-200 analyzing a deleted session, got 200")
	}
}
