package editor_test

import (
	"testing"

	"github.com/dreikanter/notes-cli/internal/editor"
)

func TestParse(t *testing.T) {
	tests := []struct {
		input    string
		wantBin  string
		wantArgs []string
	}{
		{"vim", "vim", nil},
		{"subl --wait", "subl", []string{"--wait"}},
		{"/usr/bin/code -w --new-window", "/usr/bin/code", []string{"-w", "--new-window"}},
		{"", "", nil},
	}
	for _, tt := range tests {
		bin, args := editor.Parse(tt.input)
		if bin != tt.wantBin {
			t.Errorf("Parse(%q) bin = %q, want %q", tt.input, bin, tt.wantBin)
		}
		if len(args) != len(tt.wantArgs) {
			t.Errorf("Parse(%q) args = %v, want %v", tt.input, args, tt.wantArgs)
			continue
		}
		for i := range args {
			if args[i] != tt.wantArgs[i] {
				t.Errorf("Parse(%q) args[%d] = %q, want %q", tt.input, i, args[i], tt.wantArgs[i])
			}
		}
	}
}

func TestIsTerminal(t *testing.T) {
	for _, name := range []string{"vim", "nvim", "nano", "emacs", "vi", "micro"} {
		if !editor.IsTerminal(name) {
			t.Errorf("expected %q to be a terminal editor", name)
		}
	}
	if !editor.IsTerminal("/usr/bin/vim") {
		t.Error("expected /usr/bin/vim to be a terminal editor")
	}
	for _, name := range []string{"code", "subl", "zed", "gedit"} {
		if editor.IsTerminal(name) {
			t.Errorf("expected %q to NOT be a terminal editor", name)
		}
	}
}
