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

// The active tab of the active pane, for every open workspace — one row per
// Zed window. Which of those windows is in front is a question macOS will not
// answer for Zed: it publishes no accessibility windows, so System Events
// cannot see one, and reading a window's title otherwise needs the screen
// recording permission, which is far too much to ask for a status line.
//
// So the rows come back for all of them and the choice is made below.
const zedQuery = `
SELECT w.paths, e.path
FROM items i
JOIN panes p ON p.pane_id = i.pane_id AND p.workspace_id = i.workspace_id
JOIN editors e ON e.item_id = i.item_id AND e.workspace_id = i.workspace_id
JOIN workspaces w ON w.workspace_id = i.workspace_id
WHERE i.active = 1 AND p.active = 1
ORDER BY w.timestamp DESC
LIMIT 8;`

// If the tables above tell us nothing — an empty pane, a workspace that has
// only just opened — the navigation history still knows where someone was.
const zedFallbackQuery = `
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

	if project, file = pickOpenWorkspace(zedRows(db, zedQuery)); project != "" {
		return project, file
	}
	return zedQueryRow(db, zedFallbackQuery)
}

// pickOpenWorkspace chooses between several open Zed windows.
//
// Zed writes every window's state in one go, so their timestamps are
// identical and say nothing about which one someone is sitting in — and macOS
// will not say either: Zed publishes no accessibility windows, and reading a
// window title otherwise needs the screen recording permission, which is far
// too much to ask for a status line.
//
// What does say something is movement. The agent runs continuously, so it can
// remember what each window had open a moment ago; the window whose tab
// changed is the window someone is in. Between changes the last answer sticks,
// which is right — you are still in that project while you read.
//
// The first sample has nothing to compare against and takes the database's own
// order. It corrects itself the moment a tab is switched.
func pickOpenWorkspace(rows [][2]string) (project, file string) {
	current := make(map[string]string, len(rows))
	for _, row := range rows {
		if row[0] != "" {
			current[row[0]] = row[1]
		}
	}
	if len(current) == 0 {
		return "", ""
	}

	chosen := ""
	for workspace, openFile := range current {
		if previous, seen := zedSeen[workspace]; seen && previous != openFile {
			chosen = workspace
			break
		}
	}

	// Nothing moved: stay where we were, as long as that window is still open.
	if chosen == "" {
		if _, stillOpen := current[zedChoice]; stillOpen {
			chosen = zedChoice
		}
	}

	// First run, or the window we were watching has closed.
	if chosen == "" {
		for _, row := range rows {
			if row[0] != "" {
				chosen = row[0]
				break
			}
		}
	}

	zedSeen = current
	zedChoice = chosen
	return baseOrEmpty(chosen), baseOrEmpty(current[chosen])
}

// What each Zed window had open at the previous sample, and which one was
// chosen from it.
var (
	zedSeen   = map[string]string{}
	zedChoice string
)

// zedRows runs a query that returns "workspace path | file path" lines.
func zedRows(db, query string) [][2]string {
	out, err := exec.Command("sqlite3", "-readonly", db, query).Output()
	if err != nil {
		return nil
	}

	var rows [][2]string
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "|", 2)
		if len(parts) == 2 {
			rows = append(rows, [2]string{parts[0], parts[1]})
		}
	}
	return rows
}

func zedQueryRow(db, query string) (project, file string) {
	// Read-only, so nothing can be written to someone's editor state by a
	// status reporter — and so a running Zed is never blocked by this.
	out, err := exec.Command("sqlite3", "-readonly", db, query).Output()
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
