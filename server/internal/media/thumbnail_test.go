package media

import (
	"bytes"
	"image"
	_ "image/jpeg"
	"testing"
)

func TestThumbnailFitsWithinMaxSize(t *testing.T) {
	src := pngBytes(t, horizontalGradient(400, 200))

	out, err := Thumbnail(bytes.NewReader(src), 100)
	if err != nil {
		t.Fatalf("Thumbnail: %v", err)
	}

	cfg, format, err := image.DecodeConfig(bytes.NewReader(out))
	if err != nil {
		t.Fatalf("thumbnail is not a decodable image: %v", err)
	}
	if format != "jpeg" {
		t.Errorf("thumbnail format = %q, want jpeg", format)
	}
	if cfg.Width > 100 || cfg.Height > 100 {
		t.Errorf("thumbnail %dx%d exceeds max dimension 100", cfg.Width, cfg.Height)
	}
	// Aspect ratio (2:1) should be preserved: width should be the limiting side.
	if cfg.Width != 100 {
		t.Errorf("expected width pinned to 100, got %d", cfg.Width)
	}
}

func TestThumbnailRejectsNonImage(t *testing.T) {
	if _, err := Thumbnail(bytes.NewReader([]byte("nope")), 100); err == nil {
		t.Errorf("expected error for non-image input")
	}
}
