package media

import (
	"bytes"
	"image"
	"io"

	"github.com/disintegration/imaging"
)

// Thumbnail decodes an image from r, scales it so its longest side is at most
// maxSize pixels (preserving aspect ratio, never upscaling), and returns the
// result encoded as JPEG. It returns an error if r is not a decodable image.
func Thumbnail(r io.Reader, maxSize int) ([]byte, error) {
	img, _, err := image.Decode(r)
	if err != nil {
		return nil, err
	}

	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	if w > maxSize || h > maxSize {
		if w >= h {
			img = imaging.Resize(img, maxSize, 0, imaging.Lanczos)
		} else {
			img = imaging.Resize(img, 0, maxSize, imaging.Lanczos)
		}
	}

	var buf bytes.Buffer
	if err := imaging.Encode(&buf, img, imaging.JPEG, imaging.JPEGQuality(80)); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
