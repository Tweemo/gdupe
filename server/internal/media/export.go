package media

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
)

// zipRoot is the top-level folder inside the produced archive.
const zipRoot = "merged"

// Layout assigns each file a destination path (relative to the archive root)
// based on the confirmed similarity groups and the file's media type:
//
//   - images in a group  -> <group.ID>/<name>
//   - ungrouped images   -> images/<name>
//   - video / audio      -> video/<name>, audio/<name>
//   - anything else      -> other/<name>
//
// Name collisions within the same destination folder are disambiguated by
// appending "-1", "-2", ... before the extension. The result is deterministic
// for a given input order.
func Layout(files []*File, groups []SimilarityGroup) map[string]string {
	// Map each grouped file ID to its group folder.
	groupFolder := map[string]string{}
	for _, g := range groups {
		for _, id := range g.FileIDs {
			groupFolder[id] = g.ID
		}
	}

	used := map[string]bool{} // destination paths already taken
	layout := map[string]string{}

	for _, f := range files {
		var dir string
		switch {
		case groupFolder[f.ID] != "":
			dir = groupFolder[f.ID]
		case f.Type == Image:
			dir = "images"
		case f.Type == Video:
			dir = "video"
		case f.Type == Audio:
			dir = "audio"
		default:
			dir = "other"
		}
		layout[f.ID] = uniquePath(dir, path.Base(f.RelPath), used)
	}
	return layout
}

// uniquePath returns dir/name, suffixing the base name to avoid collisions.
func uniquePath(dir, name string, used map[string]bool) string {
	candidate := path.Join(dir, name)
	if !used[candidate] {
		used[candidate] = true
		return candidate
	}
	ext := path.Ext(name)
	stem := strings.TrimSuffix(name, ext)
	for i := 1; ; i++ {
		candidate = path.Join(dir, fmt.Sprintf("%s-%d%s", stem, i, ext))
		if !used[candidate] {
			used[candidate] = true
			return candidate
		}
	}
}

// WriteZip streams each file's bytes from disk into a zip archive written to w,
// placing it at merged/<layout path>. Files absent from layout are skipped.
func WriteZip(w io.Writer, files []*File, layout map[string]string) error {
	zw := zip.NewWriter(w)
	for _, f := range files {
		dest, ok := layout[f.ID]
		if !ok {
			continue
		}
		if err := writeOne(zw, f.AbsPath, path.Join(zipRoot, dest)); err != nil {
			zw.Close()
			return err
		}
	}
	return zw.Close()
}

func writeOne(zw *zip.Writer, srcPath, zipPath string) error {
	src, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer src.Close()

	dst, err := zw.Create(zipPath)
	if err != nil {
		return err
	}
	_, err = io.Copy(dst, src)
	return err
}

// Export is a convenience wrapper that builds the layout and writes the zip.
func Export(w io.Writer, files []*File, groups []SimilarityGroup) error {
	return WriteZip(w, files, Layout(files, groups))
}
