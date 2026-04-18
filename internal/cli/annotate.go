package cli

import (
	"errors"

	"github.com/spf13/cobra"
)

// claudeBinary is the name or absolute path of the Claude Code CLI binary.
// Tests override this to point at a fake shell script.
var claudeBinary = "claude"

const annotateDefaultModel = "claude-haiku-4-5"

var annotateCmd = &cobra.Command{
	Use:   "annotate <id|type|query>",
	Short: "Fill empty frontmatter (title, description, tags) using Claude Code CLI",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return errors.New("not implemented")
	},
}

func init() {
	annotateCmd.Flags().String("model", annotateDefaultModel, "Claude model to use")
	rootCmd.AddCommand(annotateCmd)
}
