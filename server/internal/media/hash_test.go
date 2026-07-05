package media

import (
	"strings"
	"testing"
)

func TestHashReaderKnownValue(t *testing.T) {
	// SHA-256 of "hello"
	const want = "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824"
	got, err := HashReader(strings.NewReader("hello"))
	if err != nil {
		t.Fatalf("HashReader returned error: %v", err)
	}
	if got != want {
		t.Errorf("HashReader = %q, want %q", got, want)
	}
}

func TestHashReaderIdenticalContentMatches(t *testing.T) {
	a, _ := HashReader(strings.NewReader("the same bytes"))
	b, _ := HashReader(strings.NewReader("the same bytes"))
	if a != b {
		t.Errorf("identical content produced different hashes: %q vs %q", a, b)
	}
}

func TestHashReaderDifferentContentDiffers(t *testing.T) {
	a, _ := HashReader(strings.NewReader("content one"))
	b, _ := HashReader(strings.NewReader("content two"))
	if a == b {
		t.Errorf("different content produced same hash: %q", a)
	}
}
