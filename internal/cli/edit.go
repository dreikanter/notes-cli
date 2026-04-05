package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/dreikanter/notescli/note"
	"github.com/spf13/cobra"
)

// terminalEditors is the set of editors that need a terminal (stdin/stdout).
// Everything else is assumed to be a GUI app and launched detached.
var terminalEditors = map[string]bool{
	"ed":     true,
	"emacs":  true,
	"jed":    true,
	"joe":    true,
	"mcedit": true,
	"micro":  true,
	"nano":   true,
	"ne":     true,
	"nvim":   true,
	"pico":   true,
	"vi":     true,
	"vim":    true,
}

// parseEditor splits an editor string (e.g. "subl --wait") into the binary
// name and any extra arguments.
func parseEditor(raw string) (string, []string) {
	parts := strings.Fields(raw)
	if len(parts) == 0 {
		return "", nil
	}
	return parts[0], parts[1:]
}

// isTerminalEditor returns true if the given binary name is a known terminal editor.
func isTerminalEditor(bin string) bool {
	return terminalEditors[filepath.Base(bin)]
}

var editCmd = &cobra.Command{
	Use:   "edit <id|type|query>",
	Short: "Open a note in your editor",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		root := mustNotesPath()
		n, err := note.ResolveRef(root, args[0])
		if err != nil {
			return err
		}

		raw := os.Getenv("VISUAL")
		if raw == "" {
			raw = os.Getenv("EDITOR")
		}
		if raw == "" {
			return fmt.Errorf("no editor configured: set $EDITOR or $VISUAL")
		}

		bin, extraArgs := parseEditor(raw)
		path := filepath.Join(root, n.RelPath)
		cmdArgs := append(extraArgs, path)
		ec := exec.Command(bin, cmdArgs...)

		if isTerminalEditor(bin) {
			ec.Stdin = os.Stdin
			ec.Stdout = os.Stdout
			ec.Stderr = os.Stderr
			return ec.Run()
		}

		return ec.Start()
	},
}

func init() {
	rootCmd.AddCommand(editCmd)
}
