package cli

import (
	"bytes"
	"strings"
	"testing"
)

func runAnnotate(t *testing.T, root string, args ...string) (string, error) {
	t.Helper()

	annotateCmd.ResetFlags()
	annotateCmd.Flags().String("model", annotateDefaultModel, "Claude model to use")

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs(append([]string{"annotate", "--path", root}, args...))

	err := rootCmd.Execute()
	return strings.TrimSpace(buf.String()), err
}

func TestAnnotateCommandRegistered(t *testing.T) {
	cmd, _, err := rootCmd.Find([]string{"annotate"})
	if err != nil {
		t.Fatalf("annotate command not registered: %v", err)
	}
	if cmd.Use == "" || !strings.HasPrefix(cmd.Use, "annotate") {
		t.Errorf("expected annotate Use, got %q", cmd.Use)
	}
}
