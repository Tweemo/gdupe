package media

import (
	"image"
	"io"

	// Register image format decoders for image.Decode.
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	_ "golang.org/x/image/bmp"
	_ "golang.org/x/image/tiff"
	_ "golang.org/x/image/webp"

	"github.com/corona10/goimagehash"
)

// PerceptualHash decodes an image from r and returns its 64-bit difference
// hash. Visually similar images (including resized or re-encoded copies)
// produce hashes within a small Hamming distance of each other. It returns an
// error if r does not contain a decodable image (e.g. HEIC, or a corrupt file).
func PerceptualHash(r io.Reader) (uint64, error) {
	img, _, err := image.Decode(r)
	if err != nil {
		return 0, err
	}
	hash, err := goimagehash.DifferenceHash(img)
	if err != nil {
		return 0, err
	}
	return hash.GetHash(), nil
}
