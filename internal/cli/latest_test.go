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

func runLatest(t *testing.T, args ...string) (string, error) {
	t.Helper()

	root := testdataPath(t)

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs(append([]string{"latest", "--path", root}, args...))

	err := rootCmd.Execute()
	return strings.TrimSpace(buf.String()), err
}

func TestLatestNoArgs(t *testing.T) {
	out, err := runLatest(t)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	root := testdataPath(t)
	want := filepath.Join(root, "2026/01/20260106_8823.md")
	if out != want {
		t.Errorf("got %q, want %q", out, want)
	}
}

func TestLatestWithType(t *testing.T) {
	out, err := runLatest(t, "todo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	root := testdataPath(t)
	want := filepath.Join(root, "2026/01/20260102_8814.todo.md")
	if out != want {
		t.Errorf("got %q, want %q", out, want)
	}
}

func TestLatestNotFound(t *testing.T) {
	_, err := runLatest(t, "nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent type, got nil")
	}
}
