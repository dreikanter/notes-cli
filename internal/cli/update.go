package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/dreikanter/notescli/note"
	"github.com/spf13/cobra"
)

var (
	updateTags        []string
	updateNoTags      bool
	updateTitle       string
	updateDescription string
	updateSlug        string
	updateNoSlug      bool
	updateType        string
	updateNoType      bool
)

var updateCmd = &cobra.Command{
	Use:   "update <ref>",
	Short: "Update frontmatter and/or rename a note",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if updateType != "" && !note.IsKnownType(updateType) {
			return fmt.Errorf("unknown note type %q (valid types: %s)", updateType, strings.Join(note.KnownTypes, ", "))
		}

		root := mustNotesPath()
		notes, err := note.Scan(root)
		if err != nil {
			return err
		}

		n := note.Resolve(notes, args[0])
		if n == nil {
			return fmt.Errorf("note not found: %s", args[0])
		}

		oldPath := filepath.Join(root, n.RelPath)
		data, err := os.ReadFile(oldPath)
		if err != nil {
			return fmt.Errorf("cannot read note: %w", err)
		}

		existing := note.ParseFrontmatterFields(data)
		body := note.StripFrontmatter(data)

		// Merge frontmatter updates.
		updated := existing

		if cmd.Flags().Changed("title") {
			updated.Title = updateTitle
		}
		if cmd.Flags().Changed("description") {
			updated.Description = updateDescription
		}
		if updateNoTags {
			updated.Tags = nil
		} else if cmd.Flags().Changed("tag") {
			updated.Tags = updateTags
		}

		// Determine new slug.
		newSlug := n.Slug
		if updateNoSlug {
			newSlug = ""
		} else if cmd.Flags().Changed("slug") {
			newSlug = updateSlug
		}

		// Determine new type.
		newType := n.Type
		if updateNoType {
			newType = ""
		} else if cmd.Flags().Changed("type") {
			newType = updateType
		}

		id, err := strconv.Atoi(n.ID)
		if err != nil {
			return fmt.Errorf("invalid note id %q: %w", n.ID, err)
		}

		newFilename := note.NoteFilename(n.Date, id, newSlug, newType)
		dir := filepath.Dir(oldPath)
		newPath := filepath.Join(dir, newFilename)

		newContent := note.BuildFrontmatter(updated) + string(body)

		tmpPath := newPath + ".tmp"
		if err := os.WriteFile(tmpPath, []byte(newContent), 0o644); err != nil {
			return fmt.Errorf("cannot write note: %w", err)
		}
		if err := os.Rename(tmpPath, newPath); err != nil {
			os.Remove(tmpPath)
			return fmt.Errorf("cannot rename note: %w", err)
		}
		if newPath != oldPath {
			if err := os.Remove(oldPath); err != nil {
				return fmt.Errorf("cannot remove old note: %w", err)
			}
		}

		fmt.Fprintln(cmd.OutOrStdout(), newPath)
		return nil
	},
}

func init() {
	updateCmd.Flags().StringArrayVar(&updateTags, "tag", nil, "tag for frontmatter (repeatable); replaces existing tags")
	updateCmd.Flags().BoolVar(&updateNoTags, "no-tags", false, "remove all tags from frontmatter")
	updateCmd.Flags().StringVar(&updateTitle, "title", "", "title for frontmatter (empty string clears it)")
	updateCmd.Flags().StringVar(&updateDescription, "description", "", "description for frontmatter (empty string clears it)")
	updateCmd.Flags().StringVar(&updateSlug, "slug", "", "update slug and rename file")
	updateCmd.Flags().BoolVar(&updateNoSlug, "no-slug", false, "remove slug from filename")
	updateCmd.Flags().StringVar(&updateType, "type", "", "update note type and rename file (todo, backlog, weekly)")
	updateCmd.Flags().BoolVar(&updateNoType, "no-type", false, "remove type suffix from filename")
	rootCmd.AddCommand(updateCmd)
}
