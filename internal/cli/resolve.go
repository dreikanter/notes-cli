package cli

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/dreikanter/notescli/note"
	"github.com/spf13/cobra"
)

var resolveCmd = &cobra.Command{
	Use:   "resolve <id|path|basename|slug|type>",
	Short: "Resolve a note reference and print its absolute path",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		today, _ := cmd.Flags().GetBool("today")

		root := mustNotesPath()

		var date string
		if today {
			date = time.Now().Format("20060102")
		}

		n, err := note.ResolveRefDate(root, args[0], date)
		if err != nil {
			return err
		}

		fmt.Fprintln(cmd.OutOrStdout(), filepath.Join(root, n.RelPath))
		return nil
	},
}

func init() {
	resolveCmd.Flags().Bool("today", false, "only match notes created today")
	rootCmd.AddCommand(resolveCmd)
}
