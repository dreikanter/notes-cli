package note

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

var slugRe = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)

// ValidateSlug returns an error if the slug cannot safely appear in a note
// filename. Empty slugs are accepted (they just omit the slug segment).
// All-digit slugs are rejected because they conflict with numeric ID lookup.
// Anything outside [A-Za-z0-9_-] is rejected to keep filenames portable and to
// avoid confusing the filename cache suffix.
func ValidateSlug(slug string) error {
	if slug == "" {
		return nil
	}
	if IsDigits(slug) {
		return fmt.Errorf("slug %q is all digits, which conflicts with note ID resolution", slug)
	}
	if !slugRe.MatchString(slug) {
		return fmt.Errorf("slug %q contains invalid characters; only [A-Za-z0-9_-] are allowed", slug)
	}
	return nil
}

// hasAllTags reports whether every entry in required appears in noteTags,
// case-insensitively. Used by both MemStore and OSStore for WithTag filtering.
func hasAllTags(noteTags []string, required []string) bool {
	set := make(map[string]struct{}, len(noteTags))
	for _, t := range noteTags {
		set[strings.ToLower(t)] = struct{}{}
	}
	for _, r := range required {
		if _, ok := set[strings.ToLower(r)]; !ok {
			return false
		}
	}
	return true
}

// computeMergedTags builds the sorted, lowercased, deduplicated union of
// frontmatter tags and body hashtags. bodyHashtags is assumed already
// lowercased (as produced by normalizeHashtags). Returns nil when the
// union is empty.
func computeMergedTags(fmTags, bodyHashtags []string) []string {
	set := make(map[string]struct{}, len(fmTags)+len(bodyHashtags))
	for _, t := range fmTags {
		if t == "" {
			continue
		}
		set[strings.ToLower(t)] = struct{}{}
	}
	for _, t := range bodyHashtags {
		set[t] = struct{}{}
	}
	if len(set) == 0 {
		return nil
	}
	out := make([]string, 0, len(set))
	for t := range set {
		out = append(out, t)
	}
	sort.Strings(out)
	return out
}

// normalizeHashtags lowercases and deduplicates a hashtag list from
// ExtractHashtags into the canonical form merged into Meta.Tags by
// OSStore.
func normalizeHashtags(raw []string) []string {
	if len(raw) == 0 {
		return nil
	}
	set := make(map[string]struct{}, len(raw))
	for _, t := range raw {
		if t == "" {
			continue
		}
		set[strings.ToLower(t)] = struct{}{}
	}
	if len(set) == 0 {
		return nil
	}
	out := make([]string, 0, len(set))
	for t := range set {
		out = append(out, t)
	}
	sort.Strings(out)
	return out
}
