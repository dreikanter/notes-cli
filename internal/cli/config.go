package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Print the effective runtime configuration",
	Long: `Print the effective runtime configuration: the resolved notes store
path, root flags, per-command default values, and the status of external
environment variables consumed by subcommands.

Each option is shown with its resolution source — flag, env var, or built-in
default. The store path is validated inline; problems are surfaced in the
output without aborting.`,
	Args: cobra.NoArgs,
	RunE: configRunE,
}

func configRunE(cmd *cobra.Command, _ []string) error {
	printConfig(cmd.OutOrStdout())
	return nil
}

func printConfig(out io.Writer) {
	path, source := resolveStorePathSource()
	abs, status := validateStorePath(path)

	fmt.Fprintln(out, "Notes store:")
	switch {
	case path == "":
		fmt.Fprintln(out, "  path:    (unset)")
	case abs != "" && abs != path:
		fmt.Fprintf(out, "  path:    %s (from %s)\n", abs, path)
	default:
		fmt.Fprintf(out, "  path:    %s\n", path)
	}
	fmt.Fprintf(out, "  source:  %s\n", source)
	fmt.Fprintf(out, "  status:  %s\n", status)
	fmt.Fprintln(out)

	w := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Options:")
	fmt.Fprintf(w, "  --path\t%s\t%s\n", displayValue(notesPath), pathFlagSource())
	fmt.Fprintf(w, "  annotate --model\t%s\tdefault\n", annotateDefaultModel)
	fmt.Fprintf(w, "  annotate --max-chars\t0\tdefault\n")
	fmt.Fprintf(w, "  annotate --timeout\t%s\tdefault\n", annotateDefaultTimeout)
	if err := w.Flush(); err != nil {
		return
	}
	fmt.Fprintln(out)

	w = tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Environment:")
	fmt.Fprintf(w, "  NOTES_PATH\t%s\tnotes store path fallback for --path\n", presenceMark("NOTES_PATH"))
	fmt.Fprintf(w, "  ANTHROPIC_API_KEY\t%s\trequired by annotate (consumed by claude CLI)\n", presenceMark("ANTHROPIC_API_KEY"))
	if err := w.Flush(); err != nil {
		return
	}
}

// resolveStorePathSource returns the raw notes store path and the source it
// came from (flag, env, or unset). Unlike notesRoot, this does not validate
// the path, so config can report issues without aborting.
func resolveStorePathSource() (string, string) {
	if notesPath != "" {
		return notesPath, "flag (--path)"
	}
	if env := os.Getenv("NOTES_PATH"); env != "" {
		return env, "env (NOTES_PATH)"
	}
	return "", "unset"
}

// validateStorePath resolves the absolute path and reports its status.
// Returns an empty abs string when path is empty.
func validateStorePath(path string) (string, string) {
	if path == "" {
		return "", "unset (set NOTES_PATH or pass --path)"
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Sprintf("error: %s", err)
	}
	info, err := os.Stat(abs)
	switch {
	case os.IsNotExist(err):
		return abs, "error: directory does not exist"
	case err != nil:
		return abs, fmt.Sprintf("error: %s", err)
	case !info.IsDir():
		return abs, "error: not a directory"
	}
	return abs, "ok"
}

// pathFlagSource reports where the --path flag value came from. The flag
// itself is the only thing we can directly observe; if it is empty the value
// is the built-in default ("").
func pathFlagSource() string {
	if notesPath != "" {
		return "flag"
	}
	return "default"
}

func displayValue(s string) string {
	if s == "" {
		return `""`
	}
	return s
}

func presenceMark(name string) string {
	if os.Getenv(name) != "" {
		return "set"
	}
	return "unset"
}

func init() {
	rootCmd.AddCommand(configCmd)
}
