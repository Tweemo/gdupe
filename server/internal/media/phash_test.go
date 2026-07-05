package media

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"testing"
)

// horizontalGradient builds an image whose brightness increases left to right.
func horizontalGradient(w, h int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for x := 0; x < w; x++ {
		v := uint8(x * 255 / (w - 1))
		for y := 0; y < h; y++ {
			img.Set(x, y, color.RGBA{v, v, v, 255})
		}
	}
	return img
}

// verticalGradient builds an image whose brightness increases top to bottom.
func verticalGradient(w, h int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		v := uint8(y * 255 / (h - 1))
		for x := 0; x < w; x++ {
			img.Set(x, y, color.RGBA{v, v, v, 255})
		}
	}
	return img
}

func pngBytes(t *testing.T, img image.Image) []byte {
	t.Helper()
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("encode png: %v", err)
	}
	return buf.Bytes()
}

func TestPerceptualHashIdenticalImagesMatch(t *testing.T) {
	data := pngBytes(t, horizontalGradient(64, 64))
	a, err := PerceptualHash(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("PerceptualHash a: %v", err)
	}
	b, err := PerceptualHash(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("PerceptualHash b: %v", err)
	}
	if d := hammingDistance(a, b); d != 0 {
		t.Errorf("identical images had distance %d, want 0", d)
	}
}

func TestPerceptualHashSimilarImagesAreClose(t *testing.T) {
	// Same gradient at two different resolutions should be perceptually close.
	a, _ := PerceptualHash(bytes.NewReader(pngBytes(t, horizontalGradient(64, 64))))
	b, _ := PerceptualHash(bytes.NewReader(pngBytes(t, horizontalGradient(200, 200))))
	if d := hammingDistance(a, b); d > 8 {
		t.Errorf("resized similar images had distance %d, want <= 8", d)
	}
}

func TestPerceptualHashDifferentImagesAreFar(t *testing.T) {
	a, _ := PerceptualHash(bytes.NewReader(pngBytes(t, horizontalGradient(64, 64))))
	b, _ := PerceptualHash(bytes.NewReader(pngBytes(t, verticalGradient(64, 64))))
	if d := hammingDistance(a, b); d < 10 {
		t.Errorf("structurally different images had distance %d, want >= 10", d)
	}
}

func TestPerceptualHashRejectsNonImage(t *testing.T) {
	_, err := PerceptualHash(bytes.NewReader([]byte("not an image")))
	if err == nil {
		t.Errorf("expected error for non-image input, got nil")
	}
}
