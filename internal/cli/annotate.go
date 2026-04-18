package cli

// Claude CLI envelope (UNVERIFIED — run-time probe blocked by sandbox,
// shape derived from claude -p --output-format json docs):
//
//	{
//	  "type": "result",
//	  "subtype": "success",
//	  "is_error": false,
//	  "result": "<schema-conforming JSON string>",
//	  "session_id": "...",
//	  "duration_ms": 0,
//	  "total_cost_usd": 0
//	}
//
// The schema-validated payload is the result field (as a JSON string).
// Task 4 tests must be verified against a real invocation before merging.
// If the observed shape differs on another machine, update annotateEnvelope
// and parseAnnotation in Task 4 accordingly.

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
