package note

import "testing"

func TestMatches_WithPublic(t *testing.T) {
	pub := Entry{ID: 1, Meta: Meta{Public: true}}
	priv := Entry{ID: 2, Meta: Meta{Public: false}}

	tests := []struct {
		name  string
		opt   QueryOpt
		entry Entry
		want  bool
	}{
		{"public=true matches public entry", WithPublic(true), pub, true},
		{"public=true rejects private entry", WithPublic(true), priv, false},
		{"public=false rejects public entry", WithPublic(false), pub, false},
		{"public=false matches private entry", WithPublic(false), priv, true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			q := buildQuery([]QueryOpt{tc.opt})
			if got := matches(tc.entry, q); got != tc.want {
				t.Fatalf("matches = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestMatches_WithPublicNotSetMatchesAny(t *testing.T) {
	q := buildQuery(nil)
	for _, e := range []Entry{
		{Meta: Meta{Public: true}},
		{Meta: Meta{Public: false}},
	} {
		if !matches(e, q) {
			t.Fatalf("matches with no opts should accept %+v", e)
		}
	}
}

func TestMatches_WithPublicAndTagAreAND(t *testing.T) {
	q := buildQuery([]QueryOpt{WithPublic(true), WithTag("x")})

	cases := []struct {
		entry Entry
		want  bool
	}{
		{Entry{Meta: Meta{Public: true, Tags: []string{"x"}}}, true},
		{Entry{Meta: Meta{Public: true, Tags: []string{"y"}}}, false},
		{Entry{Meta: Meta{Public: false, Tags: []string{"x"}}}, false},
	}
	for i, c := range cases {
		if got := matches(c.entry, q); got != c.want {
			t.Fatalf("case %d: matches = %v, want %v", i, got, c.want)
		}
	}
}
