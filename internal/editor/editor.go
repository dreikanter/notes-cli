package editor

import (
	"path/filepath"
	"strings"
)

// terminalEditors is the set of editors that need a terminal (stdin/stdout).
// Everything else is assumed to be a GUI app and launched detached.
var terminalEditors = map[string]bool{
	"ed":     true,
	"emacs":  true,
	"jed":    true,
	"joe":    true,
	"mcedit": true,
	"micro":  true,
	"nano":   true,
	"ne":     true,
	"nvim":   true,
	"pico":   true,
	"vi":     true,
	"vim":    true,
}

// Parse splits an editor string (e.g. "subl --wait") into the binary name
// and any extra arguments.
func Parse(raw string) (string, []string) {
	parts := strings.Fields(raw)
	if len(parts) == 0 {
		return "", nil
	}
	return parts[0], parts[1:]
}

// IsTerminal returns true if the given binary name is a known terminal editor.
func IsTerminal(bin string) bool {
	return terminalEditors[filepath.Base(bin)]
}
