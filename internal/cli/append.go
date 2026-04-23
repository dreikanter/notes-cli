package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/dreikanter/notes-cli/note"
	"github.com/spf13/cobra"
)

var appendCmd = &cobra.Command{
	Use:   "append [<id|type|query>]",
	Short: "Append text from stdin to a note",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		root, err := notesRoot()
		if err != nil {
			return err
		}

		in := cmd.InOrStdin()
		if stdinIsTerminal(in) {
			return fmt.Errorf("no input: pipe text to stdin (e.g. echo 'text' | notes append <target>)")
		}

		data, err := io.ReadAll(in)
		if err != nil {
			return fmt.Errorf("cannot read stdin: %w", err)
		}
		content := strings.TrimSpace(string(data))
		if content == "" {
			return nil
		}

		f := readFilterFlags(cmd)

		entry, ok, err := resolveOrFilter(cmd, root, args, f)
		if err != nil {
			return err
		}
		if !ok {
			return fmt.Errorf("specify a note by positional argument or filter flags (--type, --slug, --tag, --today)")
		}
		targetPath := filepath.Join(root, entry.RelPath)

		// Read existing file
		existing, err := os.ReadFile(targetPath)
		if err != nil {
			return fmt.Errorf("cannot read note: %w", err)
		}

		// Append: ensure existing ends with \n, then \n + content + \n
		existingStr := string(existing)
		if len(existingStr) > 0 && !strings.HasSuffix(existingStr, "\n") {
			existingStr += "\n"
		}
		result := existingStr + "\n" + content + "\n"

		if err := note.WriteAtomic(targetPath, []byte(result)); err != nil {
			return err
		}

		fmt.Fprintln(cmd.OutOrStdout(), targetPath)
		return nil
	},
}

func registerAppendFlags() {
	addFilterFlags(appendCmd)
}

func init() {
	registerAppendFlags()
	rootCmd.AddCommand(appendCmd)
}
