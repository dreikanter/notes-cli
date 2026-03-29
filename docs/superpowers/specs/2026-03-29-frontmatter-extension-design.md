# notescli Frontmatter Refactor Design

Date: 2026-03-29

## Context

`notescli/note` is the canonical source of note-format knowledge, shared with `notespub` as a Go module dependency. Currently all YAML frontmatter handling is hand-rolled:

- `ParseFrontmatterFields` — line-by-line string matching, only reads `title`, `tags`, `description`
- `BuildFrontmatter` — manual string concatenation to produce YAML output

Both are fragile: special YAML characters (colons, quotes) in titles or tags would produce invalid output or silently misparse. `notespub` also needs fields not covered (`public`, `slug`), and any future consumer would need to duplicate knowledge of the format.

## Goal

Replace all manual YAML handling with `gopkg.in/yaml.v3`. No mix of manual and library approaches — all frontmatter reading and writing goes through the library.

## Approach

### Parsing

Replace hand-rolled parser with a new function using `yaml.v3`:

```go
// ParseFrontmatter returns all frontmatter fields as a map.
// Returns nil if no valid frontmatter is present.
func ParseFrontmatter(data []byte) map[string]any
```

- Extracts the YAML block using the existing delimiter logic
- Unmarshals into `map[string]any` via `yaml.v3` — supports any field, any value type
- `ParseFrontmatterFields` is reimplemented on top of `ParseFrontmatter` (no duplication, backward compatible)
- `StripFrontmatter` is unchanged — delimiter logic stays as-is

### Writing

Replace manual string builder with `yaml.v3` marshaling:

```go
// BuildFrontmatter generates YAML frontmatter from the given fields.
func BuildFrontmatter(f FrontmatterFields) string
```

- Same signature and behavior, now backed by `yaml.v3` marshal
- Correctly handles special characters in all field values

## Fields notespub requires

| Field | Type | Purpose |
|---|---|---|
| `public` | bool | Whether note is included in the build |
| `title` | string | Page title |
| `slug` | string | URL slug override (falls back to slugified title) |
| `tags` | []string | Tag pages + related notes |
| `description` | string | Meta description |

## Implementation Notes

- Promote `gopkg.in/yaml.v3` from indirect to direct dependency (already in module graph via golangci-lint)
- No changes to `Note` struct, store scanning, or `StripFrontmatter`
- All changes in `frontmatter.go` and `frontmatter_test.go`
- Existing tests must continue to pass; add cases for special characters and new fields

## Out of Scope

- Validating frontmatter schema
- Supporting TOML or JSON frontmatter
- Block-style tag syntax (already supported via yaml.v3 unmarshal)
