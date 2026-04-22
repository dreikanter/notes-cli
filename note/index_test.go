package note

import (
	"path/filepath"
	"testing"
)

func TestLoadTestdata(t *testing.T) {
	root := testdataPath(t)
	idx, err := Load(root)
	if err != nil {
		t.Fatalf("Load(%q) error: %v", root, err)
	}
	if idx.Root() != root {
		t.Errorf("Root = %q, want %q", idx.Root(), root)
	}
	entries := idx.Entries()
	if len(entries) != 4 {
		t.Fatalf("Entries = %d, want 4", len(entries))
	}
	if entries[0].ID != "8823" {
		t.Errorf("entries[0].ID = %q, want 8823 (newest first)", entries[0].ID)
	}
	if entries[3].ID != "6973" {
		t.Errorf("entries[3].ID = %q, want 6973 (oldest last)", entries[3].ID)
	}

	// Frontmatter is parsed once during Load.
	e, ok := idx.ByID("8814")
	if !ok {
		t.Fatalf("ByID(8814) not found")
	}
	if len(e.Frontmatter.Tags) != 2 || e.Frontmatter.Tags[0] != "work" {
		t.Errorf("entry 8814 tags = %v, want [work planning]", e.Frontmatter.Tags)
	}
	// Stat fields are populated.
	if e.Size == 0 {
		t.Errorf("entry 8814 Size = 0, want >0")
	}
	if e.ModTime.IsZero() {
		t.Errorf("entry 8814 ModTime is zero")
	}
}

func TestLoadEmpty(t *testing.T) {
	root := t.TempDir()
	idx, err := Load(root)
	if err != nil {
		t.Fatalf("Load empty root error: %v", err)
	}
	if len(idx.Entries()) != 0 {
		t.Errorf("Entries on empty root = %d, want 0", len(idx.Entries()))
	}
	if tags := idx.Tags(); len(tags) != 0 {
		t.Errorf("Tags on empty root = %v, want []", tags)
	}
	if _, ok := idx.ByID("1"); ok {
		t.Errorf("ByID on empty root should miss")
	}
}

func TestLoadWithoutFrontmatter(t *testing.T) {
	root := testdataPath(t)
	idx, err := Load(root, WithFrontmatter(false))
	if err != nil {
		t.Fatalf("Load error: %v", err)
	}
	e, ok := idx.ByID("8814")
	if !ok {
		t.Fatalf("ByID(8814) not found")
	}
	if len(e.Frontmatter.Tags) != 0 {
		t.Errorf("Frontmatter.Tags = %v on WithFrontmatter(false), want empty", e.Frontmatter.Tags)
	}
	if e.Size == 0 || e.ModTime.IsZero() {
		t.Errorf("stat fields unset with WithFrontmatter(false)")
	}
	if tags := idx.Tags(); len(tags) != 0 {
		t.Errorf("Tags = %v with WithFrontmatter(false), want empty", tags)
	}
}

func TestLoadWithScanOptionsLenient(t *testing.T) {
	root := t.TempDir()
	writeNote(t, root, "2026/01/20260101_1.md", "---\ntags: [a]\n---\n\nbody\n")
	writeNote(t, root, "drafts/20260102_2.md", "---\ntags: [b]\n---\n\nbody\n")

	strict, err := Load(root)
	if err != nil {
		t.Fatalf("Load strict error: %v", err)
	}
	if len(strict.Entries()) != 1 {
		t.Fatalf("strict Entries = %d, want 1", len(strict.Entries()))
	}

	lenient, err := Load(root, WithScanOptions(ScanOptions{Strict: false}))
	if err != nil {
		t.Fatalf("Load lenient error: %v", err)
	}
	if len(lenient.Entries()) != 2 {
		t.Fatalf("lenient Entries = %d, want 2", len(lenient.Entries()))
	}
}

func TestIndexByRelAndByID(t *testing.T) {
	root := testdataPath(t)
	idx, err := Load(root)
	if err != nil {
		t.Fatalf("Load error: %v", err)
	}

	rel := filepath.Join("2026", "01", "20260106_8823_999.md")
	e, ok := idx.ByRel(rel)
	if !ok {
		t.Fatalf("ByRel(%q) not found", rel)
	}
	if e.ID != "8823" {
		t.Errorf("ByRel(%q).ID = %q, want 8823", rel, e.ID)
	}

	if _, ok := idx.ByRel("no/such/path.md"); ok {
		t.Errorf("ByRel miss should return false")
	}

	if _, ok := idx.ByID("9999"); ok {
		t.Errorf("ByID(9999) miss should return false")
	}
}

func TestIndexBySlug(t *testing.T) {
	root := testdataPath(t)
	idx, err := Load(root)
	if err != nil {
		t.Fatalf("Load error: %v", err)
	}

	got := idx.BySlug("meeting")
	if len(got) != 1 || got[0].ID != "8818" {
		t.Errorf("BySlug(meeting) = %+v, want [ID=8818]", got)
	}

	if got := idx.BySlug(""); got != nil {
		t.Errorf("BySlug(\"\") = %+v, want nil (empty slugs are not indexed)", got)
	}

	if got := idx.BySlug("nonexistent"); got != nil {
		t.Errorf("BySlug(nonexistent) = %+v, want nil", got)
	}
}

func TestIndexByTagAndTags(t *testing.T) {
	root := t.TempDir()
	writeNote(t, root, "2026/01/20260101_1.md", "---\ntags: [Work, Meeting]\n---\n\nbody\n")
	writeNote(t, root, "2026/01/20260102_2.md", "---\ntags: [work, planning]\n---\n\nbody\n")
	writeNote(t, root, "2026/01/20260103_3.md", "body with #inline only\n")

	idx, err := Load(root)
	if err != nil {
		t.Fatalf("Load error: %v", err)
	}

	tags := idx.Tags()
	want := []string{"meeting", "planning", "work"}
	if len(tags) != len(want) {
		t.Fatalf("Tags = %v, want %v", tags, want)
	}
	for i, w := range want {
		if tags[i] != w {
			t.Errorf("Tags[%d] = %q, want %q", i, tags[i], w)
		}
	}

	// ByTag is case-insensitive.
	work := idx.ByTag("WORK")
	if len(work) != 2 {
		t.Fatalf("ByTag(WORK) = %d entries, want 2", len(work))
	}
	// Sorted newest-first.
	if work[0].ID != "2" || work[1].ID != "1" {
		t.Errorf("ByTag(WORK) order = [%s,%s], want [2,1]", work[0].ID, work[1].ID)
	}

	if got := idx.ByTag("inline"); got != nil {
		t.Errorf("ByTag(inline) = %+v, want nil (body hashtags not indexed)", got)
	}
}

func TestIndexResolve(t *testing.T) {
	root := testdataPath(t)
	idx, err := Load(root)
	if err != nil {
		t.Fatalf("Load error: %v", err)
	}

	tests := []struct {
		name    string
		query   string
		wantID  string
		wantOK  bool
		wantErr bool
	}{
		{"empty returns most recent", "", "8823", true, false},
		{"numeric id", "8823", "8823", true, false},
		{"numeric not found is miss not error", "9999", "", false, false},
		{"type with special behavior", "todo", "8814", true, false},
		{"slug substring", "letter_opener", "6973", true, false},
		{"slug miss", "nonexistent", "", false, false},
		{"numeric date-ish string is ID miss, not slug match", "202601", "", false, false},
		{"absolute path hit", filepath.Join(root, "2026", "01", "20260106_8823_999.md"), "8823", true, false},
		{"path outside root is error", "/tmp", "", false, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e, ok, err := idx.Resolve(tt.query)
			if tt.wantErr {
				if err == nil {
					t.Errorf("Resolve(%q) expected error", tt.query)
				}
				return
			}
			if err != nil {
				t.Fatalf("Resolve(%q) unexpected error: %v", tt.query, err)
			}
			if ok != tt.wantOK {
				t.Fatalf("Resolve(%q) ok = %v, want %v", tt.query, ok, tt.wantOK)
			}
			if ok && e.ID != tt.wantID {
				t.Errorf("Resolve(%q).ID = %q, want %q", tt.query, e.ID, tt.wantID)
			}
		})
	}
}

func TestIndexResolveEmptyStore(t *testing.T) {
	root := t.TempDir()
	idx, err := Load(root)
	if err != nil {
		t.Fatalf("Load error: %v", err)
	}
	_, ok, err := idx.Resolve("")
	if err != nil {
		t.Fatalf("Resolve empty on empty store error: %v", err)
	}
	if ok {
		t.Errorf("Resolve empty on empty store should miss")
	}
}

// TestEntriesDefensiveCopy verifies mutating the returned slice or its
// frontmatter tag slices does not bleed back into the index.
func TestEntriesDefensiveCopy(t *testing.T) {
	root := testdataPath(t)
	idx, err := Load(root)
	if err != nil {
		t.Fatalf("Load error: %v", err)
	}
	e, ok := idx.ByID("8814")
	if !ok {
		t.Fatalf("ByID(8814) not found")
	}
	if len(e.Frontmatter.Tags) < 2 {
		t.Fatalf("need at least 2 tags to mutate")
	}
	e.Frontmatter.Tags[0] = "tampered"

	again, _ := idx.ByID("8814")
	if again.Frontmatter.Tags[0] == "tampered" {
		t.Errorf("mutating returned Entry's Tags changed the index")
	}
}

func TestIndexByIDKeepsNewestOnCollision(t *testing.T) {
	root := t.TempDir()
	// Two notes with the same ID in different months; the newer RelPath wins.
	writeNote(t, root, "2026/01/20260101_1.md", "---\ntags: [old]\n---\n\n")
	writeNote(t, root, "2026/02/20260201_1_newer.md", "---\ntags: [new]\n---\n\n")

	idx, err := Load(root)
	if err != nil {
		t.Fatalf("Load error: %v", err)
	}
	e, ok := idx.ByID("1")
	if !ok {
		t.Fatalf("ByID(1) not found")
	}
	if e.Slug != "newer" {
		t.Errorf("ByID(1).Slug = %q, want \"newer\" (newest entry)", e.Slug)
	}
}
