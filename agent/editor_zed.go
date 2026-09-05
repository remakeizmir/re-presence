//go:build darwin || linux

package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

// Zed does not hand out a window title.
//
// Every other editor writes the file and the project into the title bar, and
// the desktop will read that back to anyone who asks. Zed draws its own
// windows with GPUI, and macOS's System Events cannot even see one: asking
// for its title answers "invalid index".
//
// So for Zed the answer comes from Zed. It keeps a small SQLite database of
// which workspace is open and which files were visited in it, which is enough
// for the card — the project and the current file. Nothing else is read from
// it, and only the last segment of each path ever leaves the machine.
//
// This needs the sqlite3 command, which macOS ships and most Linux
// distributions have. Without it the card falls back to the app name alone.

const zedQuery = `
SELECT w.paths, h.path
FROM recent_navigation_history h
JOIN workspaces w ON w.workspace_id = h.workspace_id
WHERE h.position = 0
ORDER BY w.timestamp DESC
LIMIT 1;`

// zedState returns the project and file Zed has open, or empty strings.
func zedState() (project, file string) {
	db := zedDatabase()
	if db == "" {
		return "", ""
	}

	// Read-only, so nothing can be written to someone's editor state by a
	// status reporter — and so a running Zed is never blocked by this.
	out, err := exec.Command("sqlite3", "-readonly", db, zedQuery).Output()
	if err != nil {
		return "", ""
	}

	parts := strings.SplitN(strings.TrimSpace(string(out)), "|", 2)
	if len(parts) == 0 || parts[0] == "" {
		return "", ""
	}

	project = baseOrEmpty(parts[0])
	if len(parts) == 2 {
		file = baseOrEmpty(parts[1])
	}
	return project, file
}

// zedDatabase finds the most recently touched channel database — someone may
// have stable and preview installed, and the newest one is the one they are
// working in.
func zedDatabase() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	roots := []string{
		filepath.Join(home, "Library", "Application Support", "Zed", "db"),
		filepath.Join(home, ".local", "share", "zed", "db"),
	}

	var newest string
	var newestAt int64

	for _, root := range roots {
		entries, err := os.ReadDir(root)
		if err != nil {
			continue
		}
		names := make([]string, 0, len(entries))
		for _, entry := range entries {
			// 0-global holds settings shared across channels, not workspaces.
			if entry.IsDir() && entry.Name() != "0-global" {
				names = append(names, entry.Name())
			}
		}
		sort.Strings(names)

		for _, name := range names {
			path := filepath.Join(root, name, "db.sqlite")
			info, err := os.Stat(path)
			if err != nil {
				continue
			}
			// The write-ahead log is what moves while Zed is running; the
			// database file itself can sit still for hours.
			at := info.ModTime().Unix()
			if wal, err := os.Stat(path + "-wal"); err == nil && wal.ModTime().Unix() > at {
				at = wal.ModTime().Unix()
			}
			if at > newestAt {
				newest, newestAt = path, at
			}
		}
	}
	return newest
}
