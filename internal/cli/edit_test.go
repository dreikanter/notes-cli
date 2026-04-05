package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseEditor(t *testing.T) {
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
		bin, args := parseEditor(tt.input)
		if bin != tt.wantBin {
			t.Errorf("parseEditor(%q) bin = %q, want %q", tt.input, bin, tt.wantBin)
		}
		if len(args) != len(tt.wantArgs) {
			t.Errorf("parseEditor(%q) args = %v, want %v", tt.input, args, tt.wantArgs)
			continue
		}
		for i := range args {
			if args[i] != tt.wantArgs[i] {
				t.Errorf("parseEditor(%q) args[%d] = %q, want %q", tt.input, i, args[i], tt.wantArgs[i])
			}
		}
	}
}

func TestIsTerminalEditor(t *testing.T) {
	for _, name := range []string{"vim", "nvim", "nano", "emacs", "vi", "micro"} {
		if !isTerminalEditor(name) {
			t.Errorf("expected %q to be a terminal editor", name)
		}
	}
	if !isTerminalEditor("/usr/bin/vim") {
		t.Error("expected /usr/bin/vim to be a terminal editor")
	}
	for _, name := range []string{"code", "subl", "zed", "gedit"} {
		if isTerminalEditor(name) {
			t.Errorf("expected %q to NOT be a terminal editor", name)
		}
	}
}

// writeFakeEditor creates a shell script named after a known terminal editor
// so the edit command runs it in foreground mode.
func writeFakeEditor(t *testing.T, body string) string {
	t.Helper()
	dir := t.TempDir()
	script := filepath.Join(dir, "vi")
	if err := os.WriteFile(script, []byte("#!/bin/sh\n"+body+"\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	return script
}

func TestEditPassesArgsToEditor(t *testing.T) {
	root := testdataPath(t)

	marker := filepath.Join(t.TempDir(), "edited")
	script := writeFakeEditor(t, `for arg in "$@"; do echo "$arg"; done > `+marker)

	t.Setenv("VISUAL", "")
	t.Setenv("EDITOR", script+" --flag")

	_, err := runEdit(t, root, "8823")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := os.ReadFile(marker)
	if err != nil {
		t.Fatalf("marker file not created: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(got)), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines (flag + path), got %d: %v", len(lines), lines)
	}
	if lines[0] != "--flag" {
		t.Errorf("first arg = %q, want %q", lines[0], "--flag")
	}
	want := filepath.Join(root, "2026/01/20260106_8823_999.md")
	if lines[1] != want {
		t.Errorf("second arg = %q, want %q", lines[1], want)
	}
}

func runEdit(t *testing.T, root string, args ...string) (string, error) {
	t.Helper()

	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	rootCmd.SetOut(stdout)
	rootCmd.SetErr(stderr)
	rootCmd.SetArgs(append([]string{"edit", "--path", root}, args...))

	err := rootCmd.Execute()
	return strings.TrimSpace(stdout.String() + stderr.String()), err
}

func TestEditOpensEditor(t *testing.T) {
	root := testdataPath(t)

	marker := filepath.Join(t.TempDir(), "edited")
	script := writeFakeEditor(t, `echo "$1" > `+marker)

	t.Setenv("VISUAL", "")
	t.Setenv("EDITOR", script)

	_, err := runEdit(t, root, "8823")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := os.ReadFile(marker)
	if err != nil {
		t.Fatalf("marker file not created: %v", err)
	}

	want := filepath.Join(root, "2026/01/20260106_8823_999.md")
	if strings.TrimSpace(string(got)) != want {
		t.Errorf("editor received %q, want %q", strings.TrimSpace(string(got)), want)
	}
}

func TestEditPrefersVisual(t *testing.T) {
	root := testdataPath(t)

	marker := filepath.Join(t.TempDir(), "edited")
	script := writeFakeEditor(t, `echo "$1" > `+marker)

	badScript := writeFakeEditor(t, "exit 1")

	t.Setenv("VISUAL", script)
	t.Setenv("EDITOR", badScript)

	_, err := runEdit(t, root, "8823")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := os.ReadFile(marker)
	if err != nil {
		t.Fatalf("marker file not created: %v", err)
	}

	want := filepath.Join(root, "2026/01/20260106_8823_999.md")
	if strings.TrimSpace(string(got)) != want {
		t.Errorf("editor received %q, want %q", strings.TrimSpace(string(got)), want)
	}
}

func TestEditNoEditorErrors(t *testing.T) {
	root := testdataPath(t)

	t.Setenv("VISUAL", "")
	t.Setenv("EDITOR", "")

	_, err := runEdit(t, root, "8823")
	if err == nil {
		t.Fatal("expected error when no editor configured, got nil")
	}
	if !strings.Contains(err.Error(), "no editor configured") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestEditNonExistentRefErrors(t *testing.T) {
	root := testdataPath(t)

	t.Setenv("EDITOR", "true")

	_, err := runEdit(t, root, "9999")
	if err == nil {
		t.Fatal("expected error for non-existent ref, got nil")
	}
}

func TestEditBySlug(t *testing.T) {
	root := testdataPath(t)

	marker := filepath.Join(t.TempDir(), "edited")
	script := writeFakeEditor(t, `echo "$1" > `+marker)

	t.Setenv("VISUAL", "")
	t.Setenv("EDITOR", script)

	_, err := runEdit(t, root, "meeting")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := os.ReadFile(marker)
	if err != nil {
		t.Fatalf("marker file not created: %v", err)
	}

	want := filepath.Join(root, "2026/01/20260104_8818_meeting.md")
	if strings.TrimSpace(string(got)) != want {
		t.Errorf("editor received %q, want %q", strings.TrimSpace(string(got)), want)
	}
}

func TestEditByType(t *testing.T) {
	root := testdataPath(t)

	marker := filepath.Join(t.TempDir(), "edited")
	script := writeFakeEditor(t, `echo "$1" > `+marker)

	t.Setenv("VISUAL", "")
	t.Setenv("EDITOR", script)

	_, err := runEdit(t, root, "todo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := os.ReadFile(marker)
	if err != nil {
		t.Fatalf("marker file not created: %v", err)
	}

	want := filepath.Join(root, "2026/01/20260102_8814.todo.md")
	if strings.TrimSpace(string(got)) != want {
		t.Errorf("editor received %q, want %q", strings.TrimSpace(string(got)), want)
	}
}
