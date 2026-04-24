package cli

import (
	"fmt"

	"github.com/dreikanter/notes-cli/note"
	"github.com/spf13/cobra"
)

// stderrLogger returns a note.Logger that writes non-fatal warnings from
// Load/Scan/Reload to cmd's stderr. The note package itself no longer writes
// to os.Stderr — CLI commands wire this at the edge.
//
// Kept while the `new` and `new-todo` commands still call `note.Load`; it
// disappears once those migrate to the Store in later phases.
func stderrLogger(cmd *cobra.Command) note.Logger {
	return func(err error) {
		fmt.Fprintf(cmd.ErrOrStderr(), "warn: %v\n", err)
	}
}
