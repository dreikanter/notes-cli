// migrate-types scans a notes archive and outputs rename commands for notes
// whose slug matches a known type (todo, backlog, weekly), converting them
// to the new secondary-extension format (e.g. _todo.md → .todo.md).
//
// Usage:
//
//	go run ./cmd/migrate-types            # preview renames
//	go run ./cmd/migrate-types | bash     # apply renames
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/dreikanter/notescli/note"
)

func main() {
	root := os.Getenv("NOTES_PATH")
	if root == "" {
		fmt.Fprintln(os.Stderr, "NOTES_PATH is not set")
		os.Exit(1)
	}

	entries, err := scanRaw(root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "scan error: %v\n", err)
		os.Exit(1)
	}

	count := 0
	for _, e := range entries {
		if e.noteType == "" {
			continue
		}
		oldAbs := filepath.Join(root, e.relPath)
		dir := filepath.Dir(oldAbs)

		// Build new filename: strip _<type> from basename, add .<type>.md
		newBase := strings.TrimSuffix(e.baseName, "_"+e.noteType)
		newFilename := newBase + "." + e.noteType + ".md"
		newAbs := filepath.Join(dir, newFilename)

		fmt.Printf("mv %q %q\n", oldAbs, newAbs)
		count++
	}

	fmt.Fprintf(os.Stderr, "# %d file(s) to rename\n", count)
}

type entry struct {
	relPath  string
	baseName string
	noteType string
}

// scanRaw walks the archive using the OLD naming convention (type encoded as slug).
func scanRaw(root string) ([]entry, error) {
	var entries []entry

	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		if filepath.Ext(path) != ".md" {
			return nil
		}

		base := strings.TrimSuffix(filepath.Base(path), ".md")
		parts := strings.SplitN(base, "_", 3)
		if len(parts) < 2 {
			return nil
		}

		date := parts[0]
		if len(date) != 8 || !isDigits(date) {
			return nil
		}
		id := parts[1]
		if !isDigits(id) || id == "" {
			return nil
		}

		slug := ""
		if len(parts) == 3 {
			slug = parts[2]
		}

		rel, _ := filepath.Rel(root, path)

		noteType := ""
		if note.IsKnownType(slug) {
			noteType = slug
		}

		entries = append(entries, entry{
			relPath:  rel,
			baseName: base,
			noteType: noteType,
		})
		return nil
	})

	return entries, err
}

func isDigits(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}
