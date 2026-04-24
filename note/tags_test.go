package note

import (
	"reflect"
	"testing"
)

func TestExtractHashtagsBasic(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want []string
	}{
		{"empty", "", nil},
		{"simple", "hello #world", []string{"world"}},
		{"multiple", "#alpha and #beta here", []string{"alpha", "beta"}},
		{"digits and dashes", "#a-b_c #123 #x1", []string{"a-b_c", "123", "x1"}},
		{"slash terminates", "see #foo/bar", []string{"foo"}},
		{"punctuation after", "ok #tag, next.", []string{"tag"}},
		{"parens", "(#tag)", []string{"tag"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := ExtractHashtags([]byte(c.in))
			if !reflect.DeepEqual(got, c.want) {
				t.Fatalf("got %v, want %v", got, c.want)
			}
		})
	}
}

func TestExtractHashtagsNegative(t *testing.T) {
	cases := []struct {
		name string
		in   string
	}{
		{"heading h1", "# heading not a tag"},
		{"heading h2", "## another heading"},
		{"indented heading", "   # still a heading"},
		{"word-prefixed", "foo#bar baz"},
		{"bare hash", "look here: # not-tag"},
		{"lone hash", "just # alone"},
		{"url anchor", "https://www.teamviewer.com/en/#screenshotsAnchor"},
		{"url anchor bare", "see example.com/path/#section for more"},
		{"backticked tag", "prose `#hashtag` continues"},
		{"chained hashes", "#one#two"},
		{"chained three", "prefix #one#two#three suffix"},
		{"domain anchor", "visit foo.bar/#frag here"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := ExtractHashtags([]byte(c.in))
			if len(got) != 0 {
				t.Fatalf("expected no tags, got %v", got)
			}
		})
	}
}

func TestExtractHashtagsInlineCode(t *testing.T) {
	in := "real #out and `inline #in` and #back"
	want := []string{"out", "back"}
	got := ExtractHashtags([]byte(in))
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestExtractHashtagsFencedBlock(t *testing.T) {
	in := "before #a\n```\n#hidden\n#also-hidden\n```\nafter #b\n"
	want := []string{"a", "b"}
	got := ExtractHashtags([]byte(in))
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestExtractHashtagsFencedBlockWithInfoString(t *testing.T) {
	in := "top #ok\n```go\n// #comment\n```\nend #done\n"
	want := []string{"ok", "done"}
	got := ExtractHashtags([]byte(in))
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestExtractHashtagsCRLF(t *testing.T) {
	in := "before #a\r\n```\r\n#hidden\r\n```\r\nafter #b\r\n"
	want := []string{"a", "b"}
	got := ExtractHashtags([]byte(in))
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestExtractHashtagsBareHash(t *testing.T) {
	cases := []string{"#", "text # and #", "line #\nnext #"}
	for _, in := range cases {
		got := ExtractHashtags([]byte(in))
		if len(got) != 0 {
			t.Errorf("input %q: expected no tags, got %v", in, got)
		}
	}
}
