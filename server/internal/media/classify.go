package media

import (
	"path/filepath"
	"strings"
)

// MediaType is the broad category a file falls into.
type MediaType string

const (
	Image MediaType = "image"
	Video MediaType = "video"
	Audio MediaType = "audio"
	Other MediaType = "other"
)

var extType = map[string]MediaType{
	".jpg":  Image,
	".jpeg": Image,
	".png":  Image,
	".gif":  Image,
	".bmp":  Image,
	".tiff": Image,
	".tif":  Image,
	".webp": Image,
	".heic": Image,
	".heif": Image,

	".mp4":  Video,
	".mov":  Video,
	".avi":  Video,
	".mkv":  Video,
	".webm": Video,
	".m4v":  Video,
	".wmv":  Video,
	".flv":  Video,

	".mp3":  Audio,
	".flac": Audio,
	".wav":  Audio,
	".m4a":  Audio,
	".aac":  Audio,
	".ogg":  Audio,
	".wma":  Audio,
}

// Classify returns the MediaType for a filename based on its extension.
func Classify(filename string) MediaType {
	ext := strings.ToLower(filepath.Ext(filename))
	if t, ok := extType[ext]; ok {
		return t
	}
	return Other
}
