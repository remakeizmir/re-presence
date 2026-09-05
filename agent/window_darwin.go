//go:build darwin

package main

import (
	"os/exec"
	"strings"
)

// Reading the focused window on macOS.
//
// AppleScript through System Events, which needs no accessibility permission
// for the frontmost application's name and its window title — the first run
// raises a one-time "allow this to control System Events" prompt and nothing
// after that.

const darwinScript = `
tell application "System Events"
	set frontApp to name of first application process whose frontmost is true
	set windowTitle to ""
	try
		tell process frontApp
			set windowTitle to name of front window
		end try
	end try
	return frontApp & "\n" & windowTitle
end tell`

func focusedWindow() (string, string, error) {
	out, err := exec.Command("osascript", "-e", darwinScript).Output()
	if err != nil {
		return "", "", err
	}
	lines := strings.SplitN(strings.TrimSpace(string(out)), "\n", 2)
	if len(lines) == 1 {
		return strings.TrimSpace(lines[0]), "", nil
	}
	return strings.TrimSpace(lines[0]), strings.TrimSpace(lines[1]), nil
}
