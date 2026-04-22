package watch_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/dreikanter/notes-cli/note"
	"github.com/dreikanter/notes-cli/note/watch"
)

// debounce is short enough to keep tests fast but long enough to coalesce the
// multi-step file operations fsnotify reports under typical filesystems.
const (
	debounce = 30 * time.Millisecond
	settle   = 300 * time.Millisecond
)

// waitEvent returns true if an event arrives within timeout.
func waitEvent(t *testing.T, w *watch.Watcher, timeout time.Duration) bool {
	t.Helper()
	select {
	case _, ok := <-w.Events():
		return ok
	case <-time.After(timeout):
		return false
	}
}

// drainQuiet asserts no event arrives within timeout.
func drainQuiet(t *testing.T, w *watch.Watcher, timeout time.Duration) {
	t.Helper()
	select {
	case _, ok := <-w.Events():
		if ok {
			t.Fatalf("expected no event, got one")
		}
	case <-time.After(timeout):
	}
}

// TestWatcherStrictCreateModifyDelete verifies that create, modify, and delete
// of a single .md file under YYYY/MM each produce exactly one debounced event.
func TestWatcherStrictCreateModifyDelete(t *testing.T) {
	root := t.TempDir()
	monthDir := filepath.Join(root, "2026", "04")
	if err := os.MkdirAll(monthDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	w, err := watch.New(root, watch.WithDebounce(debounce))
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	t.Cleanup(func() { _ = w.Close() })

	notePath := filepath.Join(monthDir, "20260422_1.md")

	if err := os.WriteFile(notePath, []byte("initial\n"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	if !waitEvent(t, w, settle) {
		t.Fatalf("no event after create")
	}
	drainQuiet(t, w, 2*debounce)

	if err := os.WriteFile(notePath, []byte("updated\n"), 0o644); err != nil {
		t.Fatalf("rewrite: %v", err)
	}
	if !waitEvent(t, w, settle) {
		t.Fatalf("no event after modify")
	}
	drainQuiet(t, w, 2*debounce)

	if err := os.Remove(notePath); err != nil {
		t.Fatalf("remove: %v", err)
	}
	if !waitEvent(t, w, settle) {
		t.Fatalf("no event after delete")
	}
	drainQuiet(t, w, 2*debounce)
}

// TestWatcherDebouncesBurst verifies that a burst of writes to the same file
// produces a single coalesced event rather than one per write.
func TestWatcherDebouncesBurst(t *testing.T) {
	root := t.TempDir()
	monthDir := filepath.Join(root, "2026", "04")
	if err := os.MkdirAll(monthDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	w, err := watch.New(root, watch.WithDebounce(debounce))
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	t.Cleanup(func() { _ = w.Close() })

	notePath := filepath.Join(monthDir, "20260422_1.md")
	for i := 0; i < 5; i++ {
		if err := os.WriteFile(notePath, []byte("x"), 0o644); err != nil {
			t.Fatalf("write %d: %v", i, err)
		}
		time.Sleep(debounce / 4)
	}

	if !waitEvent(t, w, settle) {
		t.Fatalf("no event after burst")
	}
	drainQuiet(t, w, 2*debounce)
}

// TestWatcherStrictIgnoresOutsideLayout verifies that in strict mode, a .md
// file created outside YYYY/MM does not trigger an event.
func TestWatcherStrictIgnoresOutsideLayout(t *testing.T) {
	root := t.TempDir()

	w, err := watch.New(root, watch.WithDebounce(debounce))
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	t.Cleanup(func() { _ = w.Close() })

	// File directly under root: rejected by strict.
	outside := filepath.Join(root, "loose.md")
	if err := os.WriteFile(outside, []byte("x"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	drainQuiet(t, w, settle)

	// File under drafts/ (non-digit parent): rejected by strict.
	drafts := filepath.Join(root, "drafts")
	if err := os.MkdirAll(drafts, 0o755); err != nil {
		t.Fatalf("mkdir drafts: %v", err)
	}
	if err := os.WriteFile(filepath.Join(drafts, "20260422_1.md"), []byte("x"), 0o644); err != nil {
		t.Fatalf("write drafts: %v", err)
	}
	drainQuiet(t, w, settle)
}

// TestWatcherLenientAcceptsAnywhere verifies that Strict=false picks up .md
// files anywhere beneath root, including arbitrary nesting.
func TestWatcherLenientAcceptsAnywhere(t *testing.T) {
	root := t.TempDir()
	drafts := filepath.Join(root, "drafts")
	if err := os.MkdirAll(drafts, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	w, err := watch.New(root,
		watch.WithDebounce(debounce),
		watch.WithScanOptions(note.ScanOptions{Strict: false}),
	)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	t.Cleanup(func() { _ = w.Close() })

	// Pre-existing non-digit subdir.
	if err := os.WriteFile(filepath.Join(drafts, "20260422_1_idea.md"), []byte("x"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	if !waitEvent(t, w, settle) {
		t.Fatalf("no event for lenient create in drafts/")
	}
	drainQuiet(t, w, 2*debounce)

	// Deeply nested subdir created after watcher start.
	deep := filepath.Join(root, "a", "b", "c")
	if err := os.MkdirAll(deep, 0o755); err != nil {
		t.Fatalf("mkdir deep: %v", err)
	}
	// Give fsnotify time to register the new descendants.
	time.Sleep(2 * debounce)
	if err := os.WriteFile(filepath.Join(deep, "20260423_2.md"), []byte("x"), 0o644); err != nil {
		t.Fatalf("write deep: %v", err)
	}
	if !waitEvent(t, w, settle) {
		t.Fatalf("no event for lenient create in deep/")
	}
}

// TestWatcherIgnoresNonMd verifies that non-.md activity does not fire events.
func TestWatcherIgnoresNonMd(t *testing.T) {
	root := t.TempDir()
	monthDir := filepath.Join(root, "2026", "04")
	if err := os.MkdirAll(monthDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	w, err := watch.New(root, watch.WithDebounce(debounce))
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	t.Cleanup(func() { _ = w.Close() })

	if err := os.WriteFile(filepath.Join(monthDir, "scratch.txt"), []byte("x"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	drainQuiet(t, w, settle)
}

// TestWatcherPicksUpNewMonthDir verifies that a newly created YYYY/MM
// directory is observed and its .md contents trigger events.
func TestWatcherPicksUpNewMonthDir(t *testing.T) {
	root := t.TempDir()

	w, err := watch.New(root, watch.WithDebounce(debounce))
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	t.Cleanup(func() { _ = w.Close() })

	monthDir := filepath.Join(root, "2026", "04")
	if err := os.MkdirAll(monthDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	// Let the watcher register the new directories before the file write.
	time.Sleep(2 * debounce)

	if err := os.WriteFile(filepath.Join(monthDir, "20260422_1.md"), []byte("x"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	if !waitEvent(t, w, settle) {
		t.Fatalf("no event after creating note in new month dir")
	}
}

// TestWatcherCloseClosesEvents verifies Close releases resources and closes
// the Events channel.
func TestWatcherCloseClosesEvents(t *testing.T) {
	root := t.TempDir()

	w, err := watch.New(root, watch.WithDebounce(debounce))
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	// Second Close is a no-op.
	if err := w.Close(); err != nil {
		t.Fatalf("Close (second): %v", err)
	}

	select {
	case _, ok := <-w.Events():
		if ok {
			t.Fatalf("expected closed channel, received value")
		}
	case <-time.After(settle):
		t.Fatalf("Events channel not closed after Close")
	}
}

// TestWatcherRejectsMissingRoot ensures New fails fast when root does not exist.
func TestWatcherRejectsMissingRoot(t *testing.T) {
	_, err := watch.New(filepath.Join(t.TempDir(), "does-not-exist"))
	if err == nil {
		t.Fatalf("expected error for missing root")
	}
}
