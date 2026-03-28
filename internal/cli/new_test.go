package cli

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func runNew(t *testing.T, root string, stdin string, args ...string) (string, error) {
	t.Helper()

	newCmd.ResetFlags()
	newCmd.Flags().String("slug", "", "descriptive slug appended to filename")
	newCmd.Flags().String("type", "", "note type (todo, backlog, weekly)")
	newCmd.Flags().StringArray("tag", nil, "tag for frontmatter (repeatable)")
	newCmd.Flags().String("description", "", "description for frontmatter")
	newCmd.Flags().String("title", "", "title for frontmatter")

	rootCmd.SetArgs(append([]string{"new", "--path", root}, args...))

	// Capture os.Stdout (new prints via fmt.Println, not cmd.OutOrStdout).
	oldStdout := os.Stdout
	pr, pw, err := os.Pipe()
	if err != nil {
		t.Fatalf("cannot create stdout pipe: %v", err)
	}
	os.Stdout = pw
	defer func() { os.Stdout = oldStdout }()

	// Replace stdin.
	oldStdin := os.Stdin
	sr, sw, err := os.Pipe()
	if err != nil {
		t.Fatalf("cannot create stdin pipe: %v", err)
	}
	os.Stdin = sr
	defer func() { os.Stdin = oldStdin }()

	go func() {
		_, _ = io.WriteString(sw, stdin)
		sw.Close()
	}()

	execErr := rootCmd.Execute()
	pw.Close()
	out, _ := io.ReadAll(pr)
	return strings.TrimSpace(string(out)), execErr
}

func TestNewDefault(t *testing.T) {
	root := copyTestdata(t)
	out, err := runNew(t, root, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasPrefix(out, root) {
		t.Errorf("expected path under root, got %q", out)
	}
	if _, err := os.Stat(out); err != nil {
		t.Errorf("created file does not exist: %v", err)
	}
}

func TestNewWithSlug(t *testing.T) {
	root := copyTestdata(t)
	out, err := runNew(t, root, "", "--slug", "myslug")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(filepath.Base(out), "_myslug") {
		t.Errorf("expected slug in filename, got %q", filepath.Base(out))
	}
}

func TestNewWithType(t *testing.T) {
	root := copyTestdata(t)
	out, err := runNew(t, root, "", "--type", "todo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(filepath.Base(out), ".todo.") {
		t.Errorf("expected type in filename, got %q", filepath.Base(out))
	}
}

func TestNewInvalidTypeErrors(t *testing.T) {
	root := copyTestdata(t)
	_, err := runNew(t, root, "", "--type", "invalid")
	if err == nil {
		t.Fatal("expected error for unknown type, got nil")
	}
}

func TestNewWithTags(t *testing.T) {
	root := copyTestdata(t)
	out, err := runNew(t, root, "", "--tag", "work", "--tag", "daily")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	data, _ := os.ReadFile(out)
	if !strings.Contains(string(data), "tags: [work, daily]") {
		t.Errorf("expected tags in frontmatter, got:\n%s", string(data))
	}
}

func TestNewWithTitleAndDescription(t *testing.T) {
	root := copyTestdata(t)
	out, err := runNew(t, root, "", "--title", "My Note", "--description", "A description")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	data, _ := os.ReadFile(out)
	content := string(data)
	if !strings.Contains(content, "title: My Note") {
		t.Errorf("expected title in frontmatter, got:\n%s", content)
	}
	if !strings.Contains(content, "description: A description") {
		t.Errorf("expected description in frontmatter, got:\n%s", content)
	}
}

func TestNewWithBody(t *testing.T) {
	root := copyTestdata(t)
	out, err := runNew(t, root, "hello world\n")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	data, _ := os.ReadFile(out)
	if !strings.Contains(string(data), "hello world") {
		t.Errorf("expected body content in file, got:\n%s", string(data))
	}
}
