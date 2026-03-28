package cli

import (
	"os"
	"path/filepath"

	"github.com/dreikanter/notescli/note"
	"github.com/spf13/cobra"
)

var readCmd = &cobra.Command{
	Use:   "read <id|path|basename|slug|type>",
	Short: "Read a note by ID, path, basename, slug, or type",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		root := mustNotesPath()
		n, err := note.ResolveRef(root, args[0])
		if err != nil {
			return err
		}

		data, err := os.ReadFile(filepath.Join(root, n.RelPath))
		if err != nil {
			return err
		}

		noFrontmatter, _ := cmd.Flags().GetBool("no-frontmatter")
		if noFrontmatter {
			data = note.StripFrontmatter(data)
		}

		_, err = os.Stdout.Write(data)
		return err
	},
}

func init() {
	readCmd.Flags().BoolP("no-frontmatter", "F", false, "exclude YAML frontmatter from output")
	rootCmd.AddCommand(readCmd)
}
