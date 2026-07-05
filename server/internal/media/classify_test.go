package media

import "testing"

func TestClassify(t *testing.T) {
	cases := []struct {
		name string
		want MediaType
	}{
		{"photo.jpg", Image},
		{"photo.JPG", Image},
		{"image.jpeg", Image},
		{"pic.png", Image},
		{"anim.gif", Image},
		{"shot.heic", Image},
		{"raw.webp", Image},
		{"clip.mp4", Video},
		{"movie.MOV", Video},
		{"old.avi", Video},
		{"web.webm", Video},
		{"song.mp3", Audio},
		{"track.flac", Audio},
		{"voice.wav", Audio},
		{"sound.m4a", Audio},
		{"notes.txt", Other},
		{"archive.zip", Other},
		{"noextension", Other},
	}
	for _, c := range cases {
		if got := Classify(c.name); got != c.want {
			t.Errorf("Classify(%q) = %v, want %v", c.name, got, c.want)
		}
	}
}
