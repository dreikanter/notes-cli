package cli

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

func testdataPath(t *testing.T) string {
	t.Helper()
	abs, err := filepath.Abs("../../testdata")
	if err != nil {
		t.Fatalf("cannot resolve testdata path: %v", err)
	}
	return abs
}

func runResolve(t *testing.T, root string, args ...string) (string, error) {
	t.Helper()

	resolveCmd.ResetFlags()
	registerResolveFlags()

	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	rootCmd.SetOut(stdout)
	rootCmd.SetErr(stderr)
	rootCmd.SetArgs(append([]string{"resolve", "--path", root}, args...))

	err := rootCmd.Execute()
	return strings.TrimSpace(stdout.String()), err
}

func TestResolveNewestNoArgs(t *testing.T) {
	root := testdataPath(t)
	out, err := runResolve(t, root)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := filepath.Join(root, "2026/01/20260106_8823_999.md")
	if out != want {
		t.Errorf("got %q, want %q", out, want)
	}
}

func TestResolveByID(t *testing.T) {
	root := testdataPath(t)
	out, err := runResolve(t, root, "--id", "8823")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := filepath.Join(root, "2026/01/20260106_8823_999.md")
	if out != want {
		t.Errorf("got %q, want %q", out, want)
	}
}

func TestResolveByIDNotFound(t *testing.T) {
	root := testdataPath(t)
	_, err := runResolve(t, root, "--id", "99999")
	if err == nil {
		t.Fatal("expected error for missing ID")
	}
}

func TestResolveByIDNonInteger(t *testing.T) {
	root := testdataPath(t)
	_, err := runResolve(t, root, "--id", "notnumber")
	if err == nil {
		t.Fatal("expected error for non-integer id")
	}
}

func TestResolveBySlug(t *testing.T) {
	root := testdataPath(t)
	out, err := runResolve(t, root, "--slug", "meeting")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := filepath.Join(root, "2026/01/20260104_8818_meeting.md")
	if out != want {
		t.Errorf("got %q, want %q", out, want)
	}
}

func TestResolveByType(t *testing.T) {
	root := testdataPath(t)
	out, err := runResolve(t, root, "--type", "todo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := filepath.Join(root, "2026/01/20260102_8814.todo.md")
	if out != want {
		t.Errorf("got %q, want %q", out, want)
	}
}

func TestResolveByTag(t *testing.T) {
	root := testdataPath(t)
	out, err := runResolve(t, root, "--tag", "meeting")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := filepath.Join(root, "2026/01/20260104_8818_meeting.md")
	if out != want {
		t.Errorf("got %q, want %q", out, want)
	}
}

func TestResolveNoMatchErrors(t *testing.T) {
	root := testdataPath(t)
	_, err := runResolve(t, root, "--slug", "nonexistent-slug-xyz")
	if err == nil {
		t.Fatal("expected error for no match")
	}
}

func TestResolveMultipleFlagsError(t *testing.T) {
	root := testdataPath(t)
	_, err := runResolve(t, root, "--id", "1", "--slug", "x")
	if err == nil {
		t.Fatal("expected error when combining --id and --slug")
	}
}
