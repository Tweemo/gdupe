// Command media-merge finds exact duplicate files in a directory.
//
// By default it runs as a dry-run: it reports how many duplicate files
// exist and how much space they waste, without touching anything.
// Pass -delete to remove the duplicate copies, or -move to relocate
// them into a subfolder instead.
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/timl/media-merge/internal/dedup"
)

func main() {
	deleteDupes := flag.Bool("delete", false, "delete duplicate copies in place (keeps one of each)")
	moveDupes := flag.Bool("move", false, "move duplicate copies into a 'duplicates' subfolder")
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: media-merge [flags] <directory>\n\n")
		fmt.Fprintf(flag.CommandLine.Output(), "Finds exact duplicate files. Dry-run by default: nothing is\nmodified unless -delete or -move is given.\n\nFlags:\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(2)
	}
	if *deleteDupes && *moveDupes {
		fmt.Fprintln(os.Stderr, "media-merge: -delete and -move are mutually exclusive")
		os.Exit(2)
	}
	dir := flag.Arg(0)

	info, err := os.Stat(dir)
	if err != nil || !info.IsDir() {
		fmt.Fprintf(os.Stderr, "media-merge: %q is not a directory\n", dir)
		os.Exit(2)
	}

	if err := run(dir, *deleteDupes, *moveDupes); err != nil {
		fmt.Fprintf(os.Stderr, "media-merge: %v\n", err)
		os.Exit(1)
	}
}

// run drives the whole workflow: scan the directory, group duplicates,
// print the report, and (only when asked) delete or move the extra copies.
func run(dir string, deleteDupes, moveDupes bool) error {
	files, err := dedup.Scan(dir)
	if err != nil {
		return err
	}

	report, err := dedup.FindDuplicates(files)
	if err != nil {
		return err
	}

	fmt.Printf("%d duplicate file(s), %s wasted\n",
		report.DuplicateCount(), dedup.FormatSize(report.WastedBytes()))

	switch {
	case deleteDupes:
		trashDir, err := dedup.TrashDir()
		if err != nil {
			return err
		}

		return dedup.Delete(report, trashDir)
	case moveDupes:
		return dedup.Move(report, dir)
	}
	return nil
}
