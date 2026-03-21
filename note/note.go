package note

import (
	"fmt"
	"path/filepath"
	"strings"
)

// Note represents a single note file in the archive.
type Note struct {
	RelPath  string // relative path from archive root, e.g. "2026/01/20260106_8823.md"
	Date     string // "20260106"
	ID       string // "8823"
	Slug     string // "todo", "disable-letter_opener", or ""
	BaseName string // filename without extension, e.g. "20260106_8823" or "20260102_8814_todo"
}

// ParseFilename parses a note base filename (without .md extension) into its components.
// Expected format: YYYYMMDD_ID[_slug]
func ParseFilename(baseName string) (Note, error) {
	parts := strings.SplitN(baseName, "_", 3)
	if len(parts) < 2 {
		return Note{}, fmt.Errorf("invalid note filename: %s", baseName)
	}

	date := parts[0]
	if len(date) != 8 || !isDigits(date) {
		return Note{}, fmt.Errorf("invalid date in filename: %s", baseName)
	}

	id := parts[1]
	if !isDigits(id) || id == "" {
		return Note{}, fmt.Errorf("invalid id in filename: %s", baseName)
	}

	slug := ""
	if len(parts) == 3 {
		slug = parts[2]
	}

	return Note{
		Date:     date,
		ID:       id,
		Slug:     slug,
		BaseName: baseName,
	}, nil
}

// NoteFilename generates a note filename from date, id, and optional slug.
func NoteFilename(date string, id int, slug string) string {
	if slug != "" {
		return fmt.Sprintf("%s_%d_%s.md", date, id, slug)
	}
	return fmt.Sprintf("%s_%d.md", date, id)
}

// NoteDirPath returns the YYYY/MM directory path for a given date string (YYYYMMDD).
func NoteDirPath(root, date string) string {
	year := date[:4]
	month := date[4:6]
	return filepath.Join(root, year, month)
}

func isDigits(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}
