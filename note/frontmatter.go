package note

import (
	"bytes"
	"strings"

	"gopkg.in/yaml.v3"
)

var frontmatterDelim = []byte("---")

// FrontmatterFields holds optional fields for note frontmatter.
type FrontmatterFields struct {
	Title       string   `yaml:"title,omitempty"`
	Slug        string   `yaml:"slug,omitempty"`
	Tags        []string `yaml:"tags,omitempty"`
	Description string   `yaml:"description,omitempty"`
	Public      bool     `yaml:"public,omitempty"`
}

// BuildFrontmatter generates YAML frontmatter from the given fields.
// Returns empty string if no fields are provided. Tags are emitted in
// flow style (`[a, b]`) to minimize diffs against existing notes.
func BuildFrontmatter(f FrontmatterFields) string {
	root := &yaml.Node{Kind: yaml.MappingNode}

	addScalar := func(key, value string) {
		root.Content = append(root.Content,
			&yaml.Node{Kind: yaml.ScalarNode, Value: key},
			&yaml.Node{Kind: yaml.ScalarNode, Value: value})
	}

	if f.Title != "" {
		addScalar("title", f.Title)
	}
	if f.Slug != "" {
		addScalar("slug", f.Slug)
	}
	if len(f.Tags) > 0 {
		tags := &yaml.Node{Kind: yaml.SequenceNode, Style: yaml.FlowStyle}
		for _, t := range f.Tags {
			tags.Content = append(tags.Content, &yaml.Node{Kind: yaml.ScalarNode, Value: t})
		}
		root.Content = append(root.Content,
			&yaml.Node{Kind: yaml.ScalarNode, Value: "tags"},
			tags)
	}
	if f.Description != "" {
		addScalar("description", f.Description)
	}
	if f.Public {
		addScalar("public", "true")
	}

	if len(root.Content) == 0 {
		return ""
	}

	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(root); err != nil {
		return ""
	}
	if err := enc.Close(); err != nil {
		return ""
	}

	// Normalize trailing whitespace so the wrapped output is stable even if
	// yaml.v3 ever changes its trailing-newline emission.
	body := strings.TrimRight(buf.String(), "\n") + "\n"
	return "---\n" + body + "---\n\n"
}

// ParseFrontmatterFields extracts all frontmatter fields from data.
// Returns zero-value FrontmatterFields if no valid frontmatter block is
// present or if the YAML inside the block is not a mapping.
//
// Per-field errors are tolerated: a single field whose value cannot be
// decoded into its target type (e.g. `public: maybe`) is skipped, and the
// remaining fields still parse. This matches the old line-based parser's
// graceful degradation on partially malformed notes.
func ParseFrontmatterFields(data []byte) FrontmatterFields {
	body, _, ok := findFrontmatterBlock(data)
	if !ok {
		return FrontmatterFields{}
	}

	var doc yaml.Node
	if err := yaml.Unmarshal(body, &doc); err != nil {
		return FrontmatterFields{}
	}
	if doc.Kind != yaml.DocumentNode || len(doc.Content) == 0 {
		return FrontmatterFields{}
	}
	mapping := doc.Content[0]
	if mapping.Kind != yaml.MappingNode {
		return FrontmatterFields{}
	}

	var f FrontmatterFields
	for i := 0; i+1 < len(mapping.Content); i += 2 {
		key, value := mapping.Content[i], mapping.Content[i+1]
		switch key.Value {
		case "title":
			_ = value.Decode(&f.Title)
		case "slug":
			_ = value.Decode(&f.Slug)
		case "tags":
			_ = value.Decode(&f.Tags)
		case "description":
			_ = value.Decode(&f.Description)
		case "public":
			_ = value.Decode(&f.Public)
		}
	}
	return f
}

// StripFrontmatter removes YAML frontmatter from the beginning of data.
// Frontmatter must start on the first line with "---" and end with a
// subsequent "---" line. Exactly one blank line after the closing
// delimiter is also consumed.
func StripFrontmatter(data []byte) []byte {
	_, after, ok := findFrontmatterBlock(data)
	if !ok {
		return data
	}
	if len(after) > 0 && after[0] == '\n' {
		return after[1:]
	}
	if len(after) > 1 && after[0] == '\r' && after[1] == '\n' {
		return after[2:]
	}
	return after
}

// findFrontmatterBlock locates the YAML frontmatter block at the start of data.
// Returns the body between the opening/closing "---" delimiter lines and the
// remaining data after the closing delimiter's newline.
func findFrontmatterBlock(data []byte) (body, after []byte, ok bool) {
	if !bytes.HasPrefix(data, frontmatterDelim) {
		return nil, nil, false
	}
	rest := data[len(frontmatterDelim):]
	idx := bytes.IndexByte(rest, '\n')
	if idx < 0 {
		return nil, nil, false
	}
	if len(bytes.TrimRight(rest[:idx], "\r")) > 0 {
		return nil, nil, false
	}
	rest = rest[idx+1:]

	var bodyBuf []byte
	for {
		line, remainder, found := bytes.Cut(rest, []byte("\n"))
		if bytes.Equal(bytes.TrimRight(line, "\r"), frontmatterDelim) {
			return bodyBuf, remainder, true
		}
		if !found {
			return nil, nil, false
		}
		bodyBuf = append(bodyBuf, line...)
		bodyBuf = append(bodyBuf, '\n')
		rest = remainder
	}
}
