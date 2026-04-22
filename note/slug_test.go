package note

import "testing"

func TestNormalizeSlug(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"", ""},
		{"hello", "hello"},
		{"Hello", "hello"},
		{"HELLO", "hello"},
		{"hello world", "hello-world"},
		{"hello  world", "hello-world"},
		{"hello_world", "hello-world"},
		{"hello-world", "hello-world"},
		{"hello--world", "hello-world"},
		{"hello!@#world", "hello-world"},
		{"---leading", "leading"},
		{"trailing---", "trailing"},
		{"  spaces  ", "spaces"},
		{"café", "caf"},
		{"123abc", "123abc"},
		{"ABC123", "abc123"},
		{"API Redesign (v2)", "api-redesign-v2"},
		{"___", ""},
		{"!!!", ""},
	}
	for _, c := range cases {
		t.Run(c.in, func(t *testing.T) {
			if got := NormalizeSlug(c.in); got != c.want {
				t.Errorf("NormalizeSlug(%q) = %q, want %q", c.in, got, c.want)
			}
		})
	}
}
