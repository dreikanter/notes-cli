package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func runConfig(t *testing.T, args ...string) (string, error) {
	t.Helper()

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs(append([]string{"config"}, args...))

	err := rootCmd.Execute()
	return buf.String(), err
}

// resetNotesPath restores the global notesPath flag value after a test, so
// other tests in the package (which rely on an unset state) are unaffected.
func resetNotesPath(t *testing.T) {
	t.Helper()
	t.Cleanup(func() { notesPath = "" })
}

func TestConfigPathFromFlag(t *testing.T) {
	resetNotesPath(t)
	t.Setenv("NOTES_PATH", "")
	t.Setenv("ANTHROPIC_API_KEY", "")

	root := t.TempDir()
	out, err := runConfig(t, "--path", root)
	require.NoError(t, err)

	assert.Contains(t, out, "Notes store:")
	assert.Contains(t, out, root)
	assert.Contains(t, out, "source:  flag (--path)")
	assert.Contains(t, out, "status:  ok")
	assert.Contains(t, out, "NOTES_PATH")
	assert.Contains(t, out, "unset")
	assert.Contains(t, out, "ANTHROPIC_API_KEY")
}

func TestConfigPathFromEnv(t *testing.T) {
	resetNotesPath(t)
	root := t.TempDir()
	t.Setenv("NOTES_PATH", root)
	t.Setenv("ANTHROPIC_API_KEY", "")

	out, err := runConfig(t)
	require.NoError(t, err)
	assert.Contains(t, out, "source:  env (NOTES_PATH)")
	assert.Contains(t, out, "status:  ok")
	assert.Contains(t, out, root)
	assert.Regexp(t, `NOTES_PATH +set`, out)
}

func TestConfigPathUnset(t *testing.T) {
	resetNotesPath(t)
	t.Setenv("NOTES_PATH", "")
	t.Setenv("ANTHROPIC_API_KEY", "")

	out, err := runConfig(t)
	require.NoError(t, err)
	assert.Contains(t, out, "path:    (unset)")
	assert.Contains(t, out, "source:  unset")
	assert.Contains(t, out, "set NOTES_PATH or pass --path")
}

func TestConfigPathMissingDirectory(t *testing.T) {
	resetNotesPath(t)
	missing := filepath.Join(t.TempDir(), "does-not-exist")
	t.Setenv("NOTES_PATH", "")
	t.Setenv("ANTHROPIC_API_KEY", "")

	out, err := runConfig(t, "--path", missing)
	require.NoError(t, err, "config must not error on a broken path; should report inline")
	assert.Contains(t, out, "directory does not exist")
}

func TestConfigPathNotADirectory(t *testing.T) {
	resetNotesPath(t)
	dir := t.TempDir()
	file := filepath.Join(dir, "regular.txt")
	require.NoError(t, os.WriteFile(file, []byte("x"), 0o644))
	t.Setenv("NOTES_PATH", "")
	t.Setenv("ANTHROPIC_API_KEY", "")

	out, err := runConfig(t, "--path", file)
	require.NoError(t, err)
	assert.Contains(t, out, "not a directory")
}

func TestConfigOptionsAndDefaults(t *testing.T) {
	resetNotesPath(t)
	root := t.TempDir()
	t.Setenv("NOTES_PATH", root)
	t.Setenv("ANTHROPIC_API_KEY", "")

	out, err := runConfig(t)
	require.NoError(t, err)

	assert.Contains(t, out, "Options:")
	assert.Contains(t, out, "annotate --model")
	assert.Contains(t, out, annotateDefaultModel)
	assert.Contains(t, out, "annotate --max-chars")
	assert.Contains(t, out, "annotate --timeout")
	assert.Contains(t, out, annotateDefaultTimeout.String())
	for _, line := range strings.Split(out, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "annotate ") {
			assert.Contains(t, line, "default")
		}
	}
}

func TestConfigEnvPresenceMarks(t *testing.T) {
	resetNotesPath(t)
	root := t.TempDir()
	t.Setenv("NOTES_PATH", root)
	t.Setenv("ANTHROPIC_API_KEY", "secret-not-printed")

	out, err := runConfig(t)
	require.NoError(t, err)

	assert.Contains(t, out, "ANTHROPIC_API_KEY")
	assert.Contains(t, out, "set")
	assert.NotContains(t, out, "secret-not-printed", "config must never print env values")
}
