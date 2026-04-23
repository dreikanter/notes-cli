package note

import "time"

// StoreEntry is the single domain object that Store implementations work with.
// It replaces the legacy Entry + Ref pair once the migration reaches Phase 14;
// the temporary Store prefix avoids collision with the existing Entry type
// during the transition.
type StoreEntry struct {
	ID   int
	Meta StoreMeta
	Body string
}

// StoreMeta holds the user-domain metadata for a note. It replaces the public
// Frontmatter type once the migration reaches Phase 14; YAML serialisation
// details live inside OSStore.
//
// CreatedAt maps to the YAML frontmatter "date" field and is both read from
// and written to disk.
//
// UpdatedAt is derived from the file's ModTime on read and is never written
// to YAML. OSStore.Put sets it to time.Now on every write — no file re-read
// needed because the OS updates ModTime when the file is rewritten.
//
// Tags is the merged set of frontmatter "tags" and body "#hashtag" tokens;
// OSStore performs the merge on read and consumers never distinguish between
// the two sources.
//
// Extra replaces the legacy map[string]yaml.Node representation with a
// plain-Go map[string]any. OSStore handles conversion at the serialisation
// boundary.
type StoreMeta struct {
	Title       string
	Slug        string
	Type        string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Tags        []string
	Aliases     []string
	Description string
	Public      bool
	Extra       map[string]any
}
