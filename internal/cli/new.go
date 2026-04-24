package cli

import (
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/dreikanter/notes-cli/note"
	"github.com/spf13/cobra"
)

var newCmd = &cobra.Command{
	Use:   "new",
	Short: "Create a new note",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		slug, _ := cmd.Flags().GetString("slug")
		noteType, _ := cmd.Flags().GetString("type")
		tags, _ := cmd.Flags().GetStringSlice("tag")
		description, _ := cmd.Flags().GetString("description")
		title, _ := cmd.Flags().GetString("title")
		publicFlag, _ := cmd.Flags().GetBool("public")
		upsert, _ := cmd.Flags().GetBool("upsert")

		if err := note.ValidateSlug(slug); err != nil {
			return err
		}

		if upsert && noteType == "" && slug == "" {
			return fmt.Errorf("--upsert requires --type or --slug")
		}

		store, err := notesStore()
		if err != nil {
			return err
		}

		if upsert {
			if existing, found, err := findUpsertEntry(store, noteType, slug); err != nil {
				return err
			} else if found {
				fmt.Fprintln(cmd.OutOrStdout(), store.AbsPath(existing))
				return nil
			}
		}

		body, err := readStdinBody(cmd)
		if err != nil {
			return err
		}

		entry := note.Entry{
			Meta: note.Meta{
				Title:       title,
				Slug:        slug,
				Type:        noteType,
				Tags:        tags,
				Description: description,
				Public:      publicFlag,
			},
			Body: body,
		}
		saved, err := store.Put(entry)
		if err != nil {
			return err
		}

		fmt.Fprintln(cmd.OutOrStdout(), store.AbsPath(saved))
		return nil
	},
}

// stdinIsTerminal reports whether in looks like an interactive terminal. Only
// *os.File readers are heuristically inspected; any other reader (a pipe,
// buffer, or other io.Reader injected via cmd.SetIn) is treated as non-terminal
// so tests and piped invocations read the provided bytes.
func stdinIsTerminal(in io.Reader) bool {
	f, ok := in.(*os.File)
	if !ok {
		return false
	}
	fi, err := f.Stat()
	if err != nil {
		return true
	}
	return fi.Mode()&os.ModeCharDevice != 0
}

// findUpsertEntry looks for today's note matching noteType and slug.
// Returns (entry, true, nil) on hit, (zero, false, nil) on clean miss, and
// a non-nil error only for I/O failures.
func findUpsertEntry(store note.Store, noteType, slug string) (note.Entry, bool, error) {
	opts := []note.QueryOpt{note.WithExactDate(time.Now())}
	if noteType != "" {
		opts = append(opts, note.WithType(noteType))
	}
	if slug != "" {
		opts = append(opts, note.WithSlug(slug))
	}
	entry, err := store.Find(opts...)
	if err != nil {
		if errors.Is(err, note.ErrNotFound) {
			return note.Entry{}, false, nil
		}
		return note.Entry{}, false, err
	}
	return entry, true, nil
}

// readStdinBody reads stdin when it is not a terminal and returns its content.
// Returns ("", nil) when stdin is a terminal (no piped input).
func readStdinBody(cmd *cobra.Command) (string, error) {
	in := cmd.InOrStdin()
	if stdinIsTerminal(in) {
		return "", nil
	}
	data, err := io.ReadAll(in)
	if err != nil {
		return "", fmt.Errorf("cannot read stdin: %w", err)
	}
	return string(data), nil
}

func registerNewFlags() {
	newCmd.Flags().String("slug", "", "descriptive slug appended to filename")
	newCmd.Flags().String("type", "", "note type (free-form; todo/backlog/weekly get special behavior)")
	newCmd.Flags().StringSlice("tag", nil, "tag for frontmatter (repeatable)")
	newCmd.Flags().String("description", "", "description for frontmatter")
	newCmd.Flags().String("title", "", "title for frontmatter")
	newCmd.Flags().Bool("public", false, "mark note as public in frontmatter")
	newCmd.Flags().Bool("private", false, "mark note as private in frontmatter (default)")
	newCmd.Flags().Bool("upsert", false, "return existing note if today already has one matching --type/--slug")
	newCmd.MarkFlagsMutuallyExclusive("public", "private")
}

func init() {
	registerNewFlags()
	rootCmd.AddCommand(newCmd)
}
