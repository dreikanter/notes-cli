package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/dreikanter/notes-cli/internal/editor"
	"github.com/spf13/cobra"
)

var editCmd = &cobra.Command{
	Use:   "edit <id|type|query>",
	Short: "Open a note in your editor",
	Long: `Open a note in your editor. The editor is read from $VISUAL, falling back to $EDITOR.

Terminal editors (vi, vim, nvim, nano, emacs, micro, etc.) run in the foreground with the terminal attached. All other editors are launched as detached processes so control returns to the terminal immediately.

The editor value may include arguments, e.g. EDITOR="subl --wait".`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		root, err := notesRoot()
		if err != nil {
			return err
		}
		n, err := resolveRef(cmd, root, args[0])
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

		bin, extraArgs := editor.Parse(raw)
		path := filepath.Join(root, n.RelPath)
		cmdArgs := make([]string, 0, len(extraArgs)+1)
		cmdArgs = append(cmdArgs, extraArgs...)
		cmdArgs = append(cmdArgs, path)
		ec := exec.Command(bin, cmdArgs...)

		if editor.IsTerminal(bin) {
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
