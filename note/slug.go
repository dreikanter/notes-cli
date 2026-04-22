package note

import "strings"

// NormalizeSlug returns an ASCII-lowercase, URL-safe form of s suitable for
// filenames and URL path segments. Rules:
//   - ASCII uppercase letters are folded to lowercase.
//   - ASCII letters and digits pass through unchanged.
//   - Any run of non-alphanumeric bytes (including underscores and non-ASCII
//     bytes) collapses to a single '-'.
//   - Leading dashes are dropped; trailing dashes are trimmed.
//
// Non-ASCII input does not transliterate — it collapses to '-'. Use this for
// the slug portion of paths where predictability matters more than fidelity.
func NormalizeSlug(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	lastDash := false
	for _, r := range s {
		switch {
		case (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9'):
			b.WriteRune(r)
			lastDash = false
		case r >= 'A' && r <= 'Z':
			b.WriteRune(r + ('a' - 'A'))
			lastDash = false
		default:
			if !lastDash && b.Len() > 0 {
				b.WriteByte('-')
				lastDash = true
			}
		}
	}
	return strings.TrimRight(b.String(), "-")
}
