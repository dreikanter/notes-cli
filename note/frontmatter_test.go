package note

import (
	"testing"
)

func TestBuildFrontmatter(t *testing.T) {
	tests := []struct {
		name        string
		slug        string
		tags        []string
		description string
		want        string
	}{
		{
			name: "empty",
			want: "",
		},
		{
			name: "slug only",
			slug: "todo",
			want: "---\nslug: todo\n---\n\n",
		},
		{
			name: "tags only",
			tags: []string{"journal", "idea"},
			want: "---\ntags: [journal, idea]\n---\n\n",
		},
		{
			name:        "description only",
			description: "Quick thought",
			want:        "---\ndescription: Quick thought\n---\n\n",
		},
		{
			name:        "all fields",
			slug:        "weekly",
			tags:        []string{"review"},
			description: "Week 10",
			want:        "---\nslug: weekly\ntags: [review]\ndescription: Week 10\n---\n\n",
		},
		{
			name: "single tag",
			tags: []string{"journal"},
			want: "---\ntags: [journal]\n---\n\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BuildFrontmatter(tt.slug, tt.tags, tt.description)
			if got != tt.want {
				t.Errorf("BuildFrontmatter(%q, %v, %q) =\n%q\nwant:\n%q", tt.slug, tt.tags, tt.description, got, tt.want)
			}
		})
	}
}

func TestStripFrontmatter(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "no frontmatter",
			input: "# Hello\n\nBody text.\n",
			want:  "# Hello\n\nBody text.\n",
		},
		{
			name:  "with frontmatter",
			input: "---\nslug: todo\ntags: [journal]\n---\n\n# Hello\n\nBody text.\n",
			want:  "# Hello\n\nBody text.\n",
		},
		{
			name:  "frontmatter only",
			input: "---\nslug: todo\n---\n",
			want:  "",
		},
		{
			name:  "empty input",
			input: "",
			want:  "",
		},
		{
			name:  "unclosed frontmatter",
			input: "---\nslug: todo\n# Hello\n",
			want:  "---\nslug: todo\n# Hello\n",
		},
		{
			name:  "triple dash in body not at start",
			input: "# Hello\n\n---\n\nFooter.\n",
			want:  "# Hello\n\n---\n\nFooter.\n",
		},
		{
			name:  "roundtrip with BuildFrontmatter",
			input: BuildFrontmatter("todo", []string{"journal"}, "A note") + "# Content\n",
			want:  "# Content\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := string(StripFrontmatter([]byte(tt.input)))
			if got != tt.want {
				t.Errorf("StripFrontmatter(%q) =\n%q\nwant:\n%q", tt.input, got, tt.want)
			}
		})
	}
}
